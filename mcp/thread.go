package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	threads "github.com/teslashibe/threads-go"
)

// GetThreadInput is the typed input for threads_get_thread.
type GetThreadInput struct {
	PostID string `json:"post_id" jsonschema:"description=numeric post ID of the focal Threads post,required"`
}

func getThread(ctx context.Context, c *threads.Client, in GetThreadInput) (any, error) {
	return c.GetThread(ctx, in.PostID)
}

// GetThreadRepliesInput is the typed input for threads_get_thread_replies.
type GetThreadRepliesInput struct {
	PostID         string `json:"post_id" jsonschema:"description=numeric post ID of the focal Threads post,required"`
	Count          int    `json:"count,omitempty" jsonschema:"description=approximate number of replies to fetch,minimum=1,maximum=50,default=20"`
	DownwardCursor string `json:"downward_cursor,omitempty" jsonschema:"description=opaque pagination cursor from a previous response (downward_cursor)"`
}

func getThreadReplies(ctx context.Context, c *threads.Client, in GetThreadRepliesInput) (any, error) {
	return c.GetThreadReplies(ctx, in.PostID, in.Count, in.DownwardCursor)
}

// GetPostInput is the typed input for threads_get_post.
type GetPostInput struct {
	PostID string `json:"post_id" jsonschema:"description=numeric post ID of the Threads post,required"`
}

func getPost(ctx context.Context, c *threads.Client, in GetPostInput) (any, error) {
	return c.GetPost(ctx, in.PostID)
}

// GetLikersInput is the typed input for threads_get_likers.
type GetLikersInput struct {
	PostID string `json:"post_id" jsonschema:"description=numeric post ID of the Threads post,required"`
}

func getLikers(ctx context.Context, c *threads.Client, in GetLikersInput) (any, error) {
	res, err := c.GetLikers(ctx, in.PostID)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(res.Users, res.NextCursor, len(res.Users)), nil
}

// GetRepostersInput is the typed input for threads_get_reposters.
type GetRepostersInput struct {
	PostID string `json:"post_id" jsonschema:"description=numeric post ID of the Threads post,required"`
}

func getReposters(ctx context.Context, c *threads.Client, in GetRepostersInput) (any, error) {
	res, err := c.GetReposters(ctx, in.PostID)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(res.Users, res.NextCursor, len(res.Users)), nil
}

// GetQuotersInput is the typed input for threads_get_quoters.
type GetQuotersInput struct {
	PostID string `json:"post_id" jsonschema:"description=numeric post ID of the Threads post,required"`
}

func getQuoters(ctx context.Context, c *threads.Client, in GetQuotersInput) (any, error) {
	res, err := c.GetQuoters(ctx, in.PostID)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(res.Users, res.NextCursor, len(res.Users)), nil
}

var threadTools = []mcptool.Tool{
	mcptool.Define[*threads.Client, GetThreadInput](
		"threads_get_thread",
		"Fetch a Threads post along with its containing thread and top-level replies",
		"GetThread",
		getThread,
	),
	mcptool.Define[*threads.Client, GetThreadRepliesInput](
		"threads_get_thread_replies",
		"Paginate through additional replies for a Threads post",
		"GetThreadReplies",
		getThreadReplies,
	),
	mcptool.Define[*threads.Client, GetPostInput](
		"threads_get_post",
		"Fetch a single Threads post by ID",
		"GetPost",
		getPost,
	),
	mcptool.Define[*threads.Client, GetLikersInput](
		"threads_get_likers",
		"List the users who liked a given Threads post",
		"GetLikers",
		getLikers,
	),
	mcptool.Define[*threads.Client, GetRepostersInput](
		"threads_get_reposters",
		"List the users who reposted a given Threads post",
		"GetReposters",
		getReposters,
	),
	mcptool.Define[*threads.Client, GetQuotersInput](
		"threads_get_quoters",
		"List the users who quote-reposted a given Threads post",
		"GetQuoters",
		getQuoters,
	),
}
