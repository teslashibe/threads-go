package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	threads "github.com/teslashibe/threads-go"
)

// SearchPostsInput is the typed input for threads_search_posts.
type SearchPostsInput struct {
	Query  string `json:"query" jsonschema:"description=keywords to search Threads posts for,required"`
	Count  int    `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	Cursor string `json:"cursor,omitempty" jsonschema:"description=opaque pagination cursor from a previous response (next_cursor)"`
}

func searchPosts(ctx context.Context, c *threads.Client, in SearchPostsInput) (any, error) {
	res, err := c.SearchPosts(ctx, in.Query, in.Count, in.Cursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 20
	}
	return mcptool.PageOf(res.Threads, res.NextCursor, limit), nil
}

// RecommendedUsersInput is the typed input for threads_recommended_users.
type RecommendedUsersInput struct {
	Count int `json:"count,omitempty" jsonschema:"description=number of recommended users to return,minimum=1,maximum=50,default=20"`
}

func recommendedUsers(ctx context.Context, c *threads.Client, in RecommendedUsersInput) (any, error) {
	res, err := c.RecommendedUsers(ctx, in.Count)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 20
	}
	return mcptool.PageOf(res.Users, res.NextCursor, limit), nil
}

// SearchUsersInput is the typed input for threads_search_users.
type SearchUsersInput struct {
	Query string `json:"query" jsonschema:"description=keywords or username fragment to search Threads users for,required"`
	Count int    `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
}

func searchUsers(ctx context.Context, c *threads.Client, in SearchUsersInput) (any, error) {
	res, err := c.SearchUsers(ctx, in.Query, in.Count)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 20
	}
	return mcptool.PageOf(res.Users, res.NextCursor, limit), nil
}

var searchTools = []mcptool.Tool{
	mcptool.Define[*threads.Client, SearchPostsInput](
		"threads_search_posts",
		"Search Threads posts by keyword",
		"SearchPosts",
		searchPosts,
	),
	mcptool.Define[*threads.Client, RecommendedUsersInput](
		"threads_recommended_users",
		"Fetch users recommended to the authenticated viewer",
		"RecommendedUsers",
		recommendedUsers,
	),
	mcptool.Define[*threads.Client, SearchUsersInput](
		"threads_search_users",
		"Search Threads users by username or display-name fragment",
		"SearchUsers",
		searchUsers,
	),
}
