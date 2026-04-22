package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// UserThreads fetches a single page of a user's thread posts. Pass an empty
// cursor for the first page; use NextCursor from the result to paginate.
//
//	page, _ := c.UserThreads(ctx, userID, 25, "")
//	for page.HasNext {
//	    page, _ = c.UserThreads(ctx, userID, 25, page.NextCursor)
//	}
func (c *Client) UserThreads(ctx context.Context, userID string, count int, cursor string) (PostPage, error) {
	if userID == "" {
		return PostPage{}, fmt.Errorf("%w: userID must not be empty", ErrInvalidParams)
	}
	params := url.Values{}
	if count > 0 {
		params.Set("count", strconv.Itoa(count))
	}
	if cursor != "" {
		params.Set("max_id", cursor)
	}
	body, err := c.readGET(ctx, "/api/v1/text_feed/"+userID+"/profile/", params)
	if err != nil {
		return PostPage{}, err
	}
	return parseThreadFeed(body)
}

// UserReplies fetches a single page of a user's replies (posts they've made
// in reply to other threads). Cursor uses next_max_id.
func (c *Client) UserReplies(ctx context.Context, userID string, count int, cursor string) (PostPage, error) {
	if userID == "" {
		return PostPage{}, fmt.Errorf("%w: userID must not be empty", ErrInvalidParams)
	}
	params := url.Values{}
	if count > 0 {
		params.Set("count", strconv.Itoa(count))
	}
	if cursor != "" {
		params.Set("max_id", cursor)
	}
	body, err := c.readGET(ctx, "/api/v1/text_feed/"+userID+"/profile/replies/", params)
	if err != nil {
		return PostPage{}, err
	}
	return parseThreadFeed(body)
}

// LikedPosts is not currently supported via the Threads web read API.
//
// Meta has removed the dedicated Threads-only liked-feed endpoint
// (/api/v1/text_feed/text_app_liked_feed/) from www.threads.com. The only
// remaining liked-content surface, /api/v1/feed/liked/, returns Instagram
// media (not Threads posts) and is intentionally not exposed by this SDK
// to avoid mixing payload shapes.
//
// As of the current Meta web surface, there is no clean way to list the
// authenticated user's Threads likes. Track a user's liked-tab manually
// (e.g. via Threads' own UI) until Meta restores a usable endpoint.
//
// _ = ctx, _ = count, _ = cursor — kept for forward-compatible signature.
func (c *Client) LikedPosts(ctx context.Context, count int, cursor string) (PostPage, error) {
	_ = ctx
	_ = count
	_ = cursor
	return PostPage{}, fmt.Errorf("%w: liked-posts feed is not currently exposed by the Threads web API", ErrNotFound)
}

// HomeTimeline fetches the For You / Following home timeline. This endpoint
// is served by i.instagram.com and requires Bearer token auth (NewWithAuth
// or NewFull). Use cursor "" to start from the top.
func (c *Client) HomeTimeline(ctx context.Context, count int, cursor string) (PostPage, error) {
	params := url.Values{
		"feed_type":         []string{"for_you"},
		"feed_view_info":    []string{"[]"},
		"reason":            []string{"cold_start_fetch"},
		"client_session_id": []string{c.auth.DeviceID},
	}
	if count > 0 {
		params.Set("pagination_source_module", "feed_unit")
	}
	if cursor != "" {
		params.Set("max_id", cursor)
	}
	body, err := c.writeGET(ctx, "/api/v1/feed/text_post_app_timeline/", params)
	if err != nil {
		return PostPage{}, err
	}
	return parseThreadFeed(body)
}

// parseThreadFeed decodes the standard {threads, next_max_id} response shape.
func parseThreadFeed(body []byte) (PostPage, error) {
	var resp struct {
		Threads    []threadPayload `json:"threads"`
		NextMaxID  json.RawMessage `json:"next_max_id"`
		Status     string          `json:"status"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return PostPage{}, fmt.Errorf("%w: decoding feed: %v", ErrRequestFailed, err)
	}
	page := PostPage{Threads: toThreads(resp.Threads)}
	page.NextCursor = decodeMaxID(resp.NextMaxID)
	page.HasNext = page.NextCursor != ""
	return page, nil
}

// decodeMaxID handles next_max_id values which may be string, int, or null.
func decodeMaxID(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var n json.Number
	if err := json.Unmarshal(raw, &n); err == nil {
		return n.String()
	}
	return ""
}
