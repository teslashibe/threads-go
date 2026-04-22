package threads

import (
	"context"
	"encoding/json"
	"fmt"
)

// Like marks the given post as liked by the authenticated user.
// Requires Bearer auth (NewWithAuth or NewFull).
func (c *Client) Like(ctx context.Context, postID string) error {
	if postID == "" {
		return fmt.Errorf("%w: postID must not be empty", ErrInvalidParams)
	}
	body, err := c.writeSignedPOST(ctx, "/api/v1/media/"+postID+"/like/", map[string]interface{}{
		"media_id":          postID,
		"container_module":  "threads_post_view",
	})
	if err != nil {
		return err
	}
	return checkOK(body)
}

// Unlike removes the authenticated user's like from a post.
func (c *Client) Unlike(ctx context.Context, postID string) error {
	if postID == "" {
		return fmt.Errorf("%w: postID must not be empty", ErrInvalidParams)
	}
	body, err := c.writeSignedPOST(ctx, "/api/v1/media/"+postID+"/unlike/", map[string]interface{}{
		"media_id":         postID,
		"container_module": "threads_post_view",
	})
	if err != nil {
		return err
	}
	return checkOK(body)
}

// Repost reshares a post to the authenticated user's profile.
func (c *Client) Repost(ctx context.Context, postID string) error {
	if postID == "" {
		return fmt.Errorf("%w: postID must not be empty", ErrInvalidParams)
	}
	body, err := c.writeSignedPOST(ctx, "/api/v1/repost/create_repost/", map[string]interface{}{
		"media_id": postID,
	})
	if err != nil {
		return err
	}
	return checkOK(body)
}

// DeleteRepost removes the authenticated user's repost of a post.
func (c *Client) DeleteRepost(ctx context.Context, postID string) error {
	if postID == "" {
		return fmt.Errorf("%w: postID must not be empty", ErrInvalidParams)
	}
	body, err := c.writeSignedPOST(ctx, "/api/v1/repost/delete_text_app_repost/", map[string]interface{}{
		"original_media_id": postID,
	})
	if err != nil {
		return err
	}
	return checkOK(body)
}

// DeletePost permanently deletes a post owned by the authenticated user.
// mediaType selects the delete endpoint variant; for text posts use 19 (default).
func (c *Client) DeletePost(ctx context.Context, postID string) error {
	if postID == "" {
		return fmt.Errorf("%w: postID must not be empty", ErrInvalidParams)
	}
	body, err := c.writeSignedPOST(ctx, "/api/v1/media/"+postID+"/delete/?media_type=TEXT_POST", map[string]interface{}{
		"media_id": postID,
	})
	if err != nil {
		return err
	}
	return checkOK(body)
}

// checkOK inspects a write response body for {"status": "ok"}. Any other
// value is surfaced as a FailStatusError. Bodies without a status field
// are treated as success.
func checkOK(body []byte) error {
	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil
	}
	if resp.Status == "" || resp.Status == "ok" {
		return nil
	}
	return &FailStatusError{Message: resp.Message}
}
