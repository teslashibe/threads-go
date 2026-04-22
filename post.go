package threads

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// CreatePost publishes a new text post. Use options to add images, reply to,
// quote, or restrict reply scope.
//
//	post, _ := c.CreatePost(ctx, "Hello from threads-go!")
//	post, _ = c.CreatePost(ctx, "Look!", threads.WithImage("/tmp/cat.jpg"))
//	post, _ = c.CreatePost(ctx, "Replying", threads.WithReplyTo(parentID))
func (c *Client) CreatePost(ctx context.Context, text string, opts ...PostOption) (*Post, error) {
	if !c.hasBearer {
		return nil, ErrWriteAuthRequired
	}
	o := &postOptions{publishMode: "text_post"}
	for _, fn := range opts {
		fn(o)
	}

	// Upload any local image paths first to obtain media IDs.
	for _, path := range o.imagePaths {
		id, err := c.UploadImage(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("upload %s: %w", path, err)
		}
		o.mediaIDs = append(o.mediaIDs, id)
	}

	payload := map[string]interface{}{
		"caption":             text,
		"publish_mode":        o.publishMode,
		"text_post_app_info":  buildTextPostAppInfo(o),
		"upload_id":           strconv.FormatInt(time.Now().UnixNano(), 10),
		"timezone_offset":     "0",
		"source_type":         "4",
		"_uid":                c.auth.UserID,
		"device_id":           c.auth.DeviceID,
	}

	var path string
	switch len(o.mediaIDs) {
	case 0:
		path = "/api/v1/media/configure_text_only_post/"
	case 1:
		path = "/api/v1/media/configure_text_post_app_feed/"
		payload["upload_id"] = o.mediaIDs[0]
	default:
		path = "/api/v1/media/configure_text_post_app_carousel/"
		children := make([]map[string]interface{}, 0, len(o.mediaIDs))
		for _, id := range o.mediaIDs {
			children = append(children, map[string]interface{}{"upload_id": id})
		}
		payload["children_metadata"] = children
	}

	body, err := c.writeSignedPOST(ctx, path, payload)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Media  postPayload `json:"media"`
		Status string      `json:"status"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("%w: decoding create post: %v", ErrRequestFailed, err)
	}
	if resp.Status != "" && resp.Status != "ok" {
		return nil, &FailStatusError{Message: "create post failed"}
	}
	p := toPost(resp.Media)
	return &p, nil
}

// Reply is a convenience wrapper around CreatePost(WithReplyTo).
func (c *Client) Reply(ctx context.Context, postID, text string, opts ...PostOption) (*Post, error) {
	all := append([]PostOption{WithReplyTo(postID)}, opts...)
	return c.CreatePost(ctx, text, all...)
}

// Quote is a convenience wrapper around CreatePost(WithQuote).
func (c *Client) Quote(ctx context.Context, postID, text string, opts ...PostOption) (*Post, error) {
	all := append([]PostOption{WithQuote(postID)}, opts...)
	return c.CreatePost(ctx, text, all...)
}

// UploadImage uploads a local image file to Instagram's rupload endpoint
// and returns the upload_id that can later be passed to CreatePost via
// WithMediaIDs.
func (c *Client) UploadImage(ctx context.Context, path string) (string, error) {
	if !c.hasBearer {
		return "", ErrWriteAuthRequired
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("%w: reading %s: %v", ErrInvalidParams, path, err)
	}
	uploadID := strconv.FormatInt(time.Now().UnixNano(), 10)
	name := uploadID + "_0_" + randomToken(10)
	endpoint := writeBaseURL + "/rupload_igphoto/" + name

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(data)))
	if err != nil {
		return "", fmt.Errorf("%w: building upload request: %v", ErrRequestFailed, err)
	}
	uploadParams, _ := json.Marshal(map[string]interface{}{
		"media_type":     "1",
		"upload_id":      uploadID,
		"image_compression": map[string]interface{}{
			"lib_name":    "moz",
			"lib_version": "3.1.m",
			"quality":     "80",
		},
		"xsharing_user_ids": "[]",
	})
	c.setWriteHeaders(req)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Entity-Type", "image/jpeg")
	req.Header.Set("X-Entity-Name", name)
	req.Header.Set("X-Entity-Length", strconv.Itoa(len(data)))
	req.Header.Set("X-Instagram-Rupload-Params", string(uploadParams))
	req.Header.Set("Offset", "0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: upload: %v", ErrRequestFailed, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return "", fmt.Errorf("%w: reading upload response: %v", ErrRequestFailed, err)
	}
	if err := mapStatus(resp.StatusCode, body); err != nil {
		return "", err
	}
	var r struct {
		UploadID string `json:"upload_id"`
		Status   string `json:"status"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return "", fmt.Errorf("%w: decoding upload response: %v (body: %s)", ErrRequestFailed, err, truncate(string(body), 200))
	}
	if r.UploadID == "" {
		return "", fmt.Errorf("%w: upload returned no upload_id (body: %s)", ErrRequestFailed, truncate(string(body), 200))
	}
	return r.UploadID, nil
}

// buildTextPostAppInfo composes the text_post_app_info object expected by
// the configure endpoints. It encodes reply, quote, and reply-control
// settings in the shape the server expects.
func buildTextPostAppInfo(o *postOptions) map[string]interface{} {
	info := map[string]interface{}{
		"reply_control": coalesceReplyControl(o.replyControl),
	}
	if o.replyToID != "" {
		info["reply_id"] = o.replyToID
	}
	if o.quotePostID != "" {
		info["quoted_post_id"] = o.quotePostID
	}
	return info
}

func coalesceReplyControl(s string) string {
	switch s {
	case "everyone", "accounts_you_follow", "mentioned_only":
		return s
	default:
		return "everyone"
	}
}

// randomToken returns a hex-encoded random string of the given length.
func randomToken(nChars int) string {
	if nChars <= 0 {
		nChars = 10
	}
	b := make([]byte, (nChars+1)/2)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)[:nChars]
}
