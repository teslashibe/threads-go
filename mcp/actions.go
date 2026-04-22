package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	threads "github.com/teslashibe/threads-go"
)

// PostActionInput is the shared typed input for the post-targeted actions
// (like / repost / delete and their inverses).
type PostActionInput struct {
	PostID string `json:"post_id" jsonschema:"description=numeric post ID of the target Threads post,required"`
}

func like(ctx context.Context, c *threads.Client, in PostActionInput) (any, error) {
	if err := c.Like(ctx, in.PostID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "post_id": in.PostID}, nil
}

func unlike(ctx context.Context, c *threads.Client, in PostActionInput) (any, error) {
	if err := c.Unlike(ctx, in.PostID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "post_id": in.PostID}, nil
}

func repost(ctx context.Context, c *threads.Client, in PostActionInput) (any, error) {
	if err := c.Repost(ctx, in.PostID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "post_id": in.PostID}, nil
}

func deleteRepost(ctx context.Context, c *threads.Client, in PostActionInput) (any, error) {
	if err := c.DeleteRepost(ctx, in.PostID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "post_id": in.PostID}, nil
}

func deletePost(ctx context.Context, c *threads.Client, in PostActionInput) (any, error) {
	if err := c.DeletePost(ctx, in.PostID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "post_id": in.PostID}, nil
}

var actionTools = []mcptool.Tool{
	mcptool.Define[*threads.Client, PostActionInput](
		"threads_like",
		"Like a Threads post",
		"Like",
		like,
	),
	mcptool.Define[*threads.Client, PostActionInput](
		"threads_unlike",
		"Remove a like from a Threads post",
		"Unlike",
		unlike,
	),
	mcptool.Define[*threads.Client, PostActionInput](
		"threads_repost",
		"Repost a Threads post to the viewer's own feed",
		"Repost",
		repost,
	),
	mcptool.Define[*threads.Client, PostActionInput](
		"threads_delete_repost",
		"Undo a previous repost of a Threads post",
		"DeleteRepost",
		deleteRepost,
	),
	mcptool.Define[*threads.Client, PostActionInput](
		"threads_delete_post",
		"Permanently delete a Threads post owned by the authenticated viewer",
		"DeletePost",
		deletePost,
	),
}
