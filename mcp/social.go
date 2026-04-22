package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	threads "github.com/teslashibe/threads-go"
)

// SocialInput is the shared typed input for the user-targeted social
// actions (follow / block / mute / restrict and their inverses).
type SocialInput struct {
	UserID string `json:"user_id" jsonschema:"description=numeric Threads user ID of the target user,required"`
}

func follow(ctx context.Context, c *threads.Client, in SocialInput) (any, error) {
	if err := c.Follow(ctx, in.UserID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "user_id": in.UserID}, nil
}

func unfollow(ctx context.Context, c *threads.Client, in SocialInput) (any, error) {
	if err := c.Unfollow(ctx, in.UserID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "user_id": in.UserID}, nil
}

func block(ctx context.Context, c *threads.Client, in SocialInput) (any, error) {
	if err := c.Block(ctx, in.UserID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "user_id": in.UserID}, nil
}

func unblock(ctx context.Context, c *threads.Client, in SocialInput) (any, error) {
	if err := c.Unblock(ctx, in.UserID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "user_id": in.UserID}, nil
}

func mute(ctx context.Context, c *threads.Client, in SocialInput) (any, error) {
	if err := c.Mute(ctx, in.UserID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "user_id": in.UserID}, nil
}

func unmute(ctx context.Context, c *threads.Client, in SocialInput) (any, error) {
	if err := c.Unmute(ctx, in.UserID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "user_id": in.UserID}, nil
}

func restrict(ctx context.Context, c *threads.Client, in SocialInput) (any, error) {
	if err := c.Restrict(ctx, in.UserID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "user_id": in.UserID}, nil
}

func unrestrict(ctx context.Context, c *threads.Client, in SocialInput) (any, error) {
	if err := c.Unrestrict(ctx, in.UserID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "user_id": in.UserID}, nil
}

var socialTools = []mcptool.Tool{
	mcptool.Define[*threads.Client, SocialInput](
		"threads_follow",
		"Follow a Threads user",
		"Follow",
		follow,
	),
	mcptool.Define[*threads.Client, SocialInput](
		"threads_unfollow",
		"Unfollow a Threads user",
		"Unfollow",
		unfollow,
	),
	mcptool.Define[*threads.Client, SocialInput](
		"threads_block",
		"Block a Threads user (also unfollows them)",
		"Block",
		block,
	),
	mcptool.Define[*threads.Client, SocialInput](
		"threads_unblock",
		"Unblock a previously blocked Threads user",
		"Unblock",
		unblock,
	),
	mcptool.Define[*threads.Client, SocialInput](
		"threads_mute",
		"Mute a Threads user (their posts no longer appear in the viewer's feed)",
		"Mute",
		mute,
	),
	mcptool.Define[*threads.Client, SocialInput](
		"threads_unmute",
		"Unmute a previously muted Threads user",
		"Unmute",
		unmute,
	),
	mcptool.Define[*threads.Client, SocialInput](
		"threads_restrict",
		"Restrict a Threads user (their replies are hidden from others by default)",
		"Restrict",
		restrict,
	),
	mcptool.Define[*threads.Client, SocialInput](
		"threads_unrestrict",
		"Unrestrict a previously restricted Threads user",
		"Unrestrict",
		unrestrict,
	),
}
