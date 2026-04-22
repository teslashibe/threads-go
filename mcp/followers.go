package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	threads "github.com/teslashibe/threads-go"
)

// GetFollowersInput is the typed input for threads_get_followers.
type GetFollowersInput struct {
	UserID string `json:"user_id" jsonschema:"description=numeric Threads user ID,required"`
	Count  int    `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	Cursor string `json:"cursor,omitempty" jsonschema:"description=opaque pagination cursor from a previous response (next_cursor)"`
}

func getFollowers(ctx context.Context, c *threads.Client, in GetFollowersInput) (any, error) {
	res, err := c.GetFollowers(ctx, in.UserID, in.Count, in.Cursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 20
	}
	return mcptool.PageOf(res.Users, res.NextCursor, limit), nil
}

// GetFollowingInput is the typed input for threads_get_following.
type GetFollowingInput struct {
	UserID string `json:"user_id" jsonschema:"description=numeric Threads user ID,required"`
	Count  int    `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	Cursor string `json:"cursor,omitempty" jsonschema:"description=opaque pagination cursor from a previous response (next_cursor)"`
}

func getFollowing(ctx context.Context, c *threads.Client, in GetFollowingInput) (any, error) {
	res, err := c.GetFollowing(ctx, in.UserID, in.Count, in.Cursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 20
	}
	return mcptool.PageOf(res.Users, res.NextCursor, limit), nil
}

// GetFriendshipInput is the typed input for threads_get_friendship.
type GetFriendshipInput struct {
	UserID string `json:"user_id" jsonschema:"description=numeric Threads user ID of the target user,required"`
}

func getFriendship(ctx context.Context, c *threads.Client, in GetFriendshipInput) (any, error) {
	return c.GetFriendship(ctx, in.UserID)
}

// GetFriendshipsInput is the typed input for threads_get_friendships.
type GetFriendshipsInput struct {
	UserIDs []string `json:"user_ids" jsonschema:"description=list of numeric Threads user IDs to fetch relationship state for,required"`
}

func getFriendships(ctx context.Context, c *threads.Client, in GetFriendshipsInput) (any, error) {
	return c.GetFriendships(ctx, in.UserIDs)
}

// PendingRequestsInput is the typed input for threads_pending_requests.
type PendingRequestsInput struct{}

func pendingRequests(ctx context.Context, c *threads.Client, _ PendingRequestsInput) (any, error) {
	res, err := c.PendingRequests(ctx)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(res.Users, res.NextCursor, len(res.Users)), nil
}

var followerTools = []mcptool.Tool{
	mcptool.Define[*threads.Client, GetFollowersInput](
		"threads_get_followers",
		"List the followers of a Threads user",
		"GetFollowers",
		getFollowers,
	),
	mcptool.Define[*threads.Client, GetFollowingInput](
		"threads_get_following",
		"List the accounts a Threads user follows",
		"GetFollowing",
		getFollowing,
	),
	mcptool.Define[*threads.Client, GetFriendshipInput](
		"threads_get_friendship",
		"Fetch the relationship state between the viewer and a single user",
		"GetFriendship",
		getFriendship,
	),
	mcptool.Define[*threads.Client, GetFriendshipsInput](
		"threads_get_friendships",
		"Fetch the relationship state between the viewer and many users at once",
		"GetFriendships",
		getFriendships,
	),
	mcptool.Define[*threads.Client, PendingRequestsInput](
		"threads_pending_requests",
		"List incoming follow requests awaiting the viewer's approval",
		"PendingRequests",
		pendingRequests,
	),
}
