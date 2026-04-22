package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	threads "github.com/teslashibe/threads-go"
)

// UserThreadsInput is the typed input for threads_user_threads.
type UserThreadsInput struct {
	UserID string `json:"user_id" jsonschema:"description=numeric Threads user ID (use threads_get_profile_by_username to resolve a handle),required"`
	Count  int    `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	Cursor string `json:"cursor,omitempty" jsonschema:"description=opaque pagination cursor from a previous response (next_cursor)"`
}

func userThreads(ctx context.Context, c *threads.Client, in UserThreadsInput) (any, error) {
	res, err := c.UserThreads(ctx, in.UserID, in.Count, in.Cursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 20
	}
	return mcptool.PageOf(res.Threads, res.NextCursor, limit), nil
}

// UserRepliesInput is the typed input for threads_user_replies.
type UserRepliesInput struct {
	UserID string `json:"user_id" jsonschema:"description=numeric Threads user ID,required"`
	Count  int    `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	Cursor string `json:"cursor,omitempty" jsonschema:"description=opaque pagination cursor from a previous response (next_cursor)"`
}

func userReplies(ctx context.Context, c *threads.Client, in UserRepliesInput) (any, error) {
	res, err := c.UserReplies(ctx, in.UserID, in.Count, in.Cursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 20
	}
	return mcptool.PageOf(res.Threads, res.NextCursor, limit), nil
}

// LikedPostsInput is the typed input for threads_liked_posts.
type LikedPostsInput struct {
	Count  int    `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	Cursor string `json:"cursor,omitempty" jsonschema:"description=opaque pagination cursor from a previous response (next_cursor)"`
}

func likedPosts(ctx context.Context, c *threads.Client, in LikedPostsInput) (any, error) {
	res, err := c.LikedPosts(ctx, in.Count, in.Cursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 20
	}
	return mcptool.PageOf(res.Threads, res.NextCursor, limit), nil
}

// HomeTimelineInput is the typed input for threads_home_timeline.
type HomeTimelineInput struct {
	Count  int    `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	Cursor string `json:"cursor,omitempty" jsonschema:"description=opaque pagination cursor from a previous response (next_cursor)"`
}

func homeTimeline(ctx context.Context, c *threads.Client, in HomeTimelineInput) (any, error) {
	res, err := c.HomeTimeline(ctx, in.Count, in.Cursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 20
	}
	return mcptool.PageOf(res.Threads, res.NextCursor, limit), nil
}

var feedTools = []mcptool.Tool{
	mcptool.Define[*threads.Client, UserThreadsInput](
		"threads_user_threads",
		"Fetch the top-level posts authored by a Threads user",
		"UserThreads",
		userThreads,
	),
	mcptool.Define[*threads.Client, UserRepliesInput](
		"threads_user_replies",
		"Fetch the reply posts authored by a Threads user",
		"UserReplies",
		userReplies,
	),
	mcptool.Define[*threads.Client, LikedPostsInput](
		"threads_liked_posts",
		"Fetch posts the authenticated viewer has liked",
		"LikedPosts",
		likedPosts,
	),
	mcptool.Define[*threads.Client, HomeTimelineInput](
		"threads_home_timeline",
		"Fetch the authenticated viewer's For You / following home timeline",
		"HomeTimeline",
		homeTimeline,
	),
}
