package threads

import (
	"context"
	"fmt"
)

// Follow follows the given user. Returns nil on success even if the request
// became a follow request for a private account.
//
// Requires Bearer auth (NewWithAuth or NewFull).
func (c *Client) Follow(ctx context.Context, userID string) error {
	return c.simpleFriendshipAction(ctx, "create", userID)
}

// Unfollow unfollows the given user.
func (c *Client) Unfollow(ctx context.Context, userID string) error {
	return c.simpleFriendshipAction(ctx, "destroy", userID)
}

// Block blocks the given user.
func (c *Client) Block(ctx context.Context, userID string) error {
	return c.simpleFriendshipAction(ctx, "block", userID)
}

// Unblock removes a block on the given user.
func (c *Client) Unblock(ctx context.Context, userID string) error {
	return c.simpleFriendshipAction(ctx, "unblock", userID)
}

// Mute mutes the given user (their posts won't appear in the timeline).
// scope must be "post" or "story" — Threads currently only supports "post".
func (c *Client) Mute(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("%w: userID must not be empty", ErrInvalidParams)
	}
	body, err := c.writeSignedPOST(ctx, "/api/v1/friendships/mute_posts_or_story_from_follow/", map[string]interface{}{
		"target_posts_author_id": userID,
		"container_module":       "ig_text_feed_timeline",
	})
	if err != nil {
		return err
	}
	return checkOK(body)
}

// Unmute removes a mute on the given user.
func (c *Client) Unmute(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("%w: userID must not be empty", ErrInvalidParams)
	}
	body, err := c.writeSignedPOST(ctx, "/api/v1/friendships/unmute_posts_or_story_from_follow/", map[string]interface{}{
		"target_posts_author_id": userID,
		"container_module":       "ig_text_feed_timeline",
	})
	if err != nil {
		return err
	}
	return checkOK(body)
}

// Restrict restricts the given user. Their replies won't appear publicly.
func (c *Client) Restrict(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("%w: userID must not be empty", ErrInvalidParams)
	}
	body, err := c.writeSignedPOST(ctx, "/api/v1/restrict_action/restrict_many/", map[string]interface{}{
		"target_user_ids":  "[" + userID + "]",
		"container_module": "ig_text_feed_timeline",
	})
	if err != nil {
		return err
	}
	return checkOK(body)
}

// Unrestrict removes a restrict on the given user.
func (c *Client) Unrestrict(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("%w: userID must not be empty", ErrInvalidParams)
	}
	body, err := c.writeSignedPOST(ctx, "/api/v1/restrict_action/unrestrict/", map[string]interface{}{
		"target_user_id":   userID,
		"container_module": "ig_text_feed_timeline",
	})
	if err != nil {
		return err
	}
	return checkOK(body)
}

// simpleFriendshipAction wraps the create/destroy/block/unblock pattern.
func (c *Client) simpleFriendshipAction(ctx context.Context, action, userID string) error {
	if userID == "" {
		return fmt.Errorf("%w: userID must not be empty", ErrInvalidParams)
	}
	body, err := c.writeSignedPOST(ctx, "/api/v1/friendships/"+action+"/"+userID+"/", map[string]interface{}{
		"user_id":          userID,
		"container_module": "ig_text_feed_timeline",
	})
	if err != nil {
		return err
	}
	return checkOK(body)
}
