package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	threads "github.com/teslashibe/threads-go"
)

// SearchHashtagsInput is the typed input for threads_search_hashtags.
type SearchHashtagsInput struct {
	Query string `json:"query" jsonschema:"description=hashtag fragment to search (with or without leading #),required"`
	Count int    `json:"count,omitempty" jsonschema:"description=maximum number of hashtags to return,minimum=1,maximum=50,default=20"`
}

func searchHashtags(ctx context.Context, c *threads.Client, in SearchHashtagsInput) (any, error) {
	res, err := c.SearchHashtags(ctx, in.Query, in.Count)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 20
	}
	return mcptool.PageOf(res, "", limit), nil
}

// GetHashtagInput is the typed input for threads_get_hashtag.
type GetHashtagInput struct {
	Name string `json:"name" jsonschema:"description=hashtag name without the leading #,required"`
}

func getHashtag(ctx context.Context, c *threads.Client, in GetHashtagInput) (any, error) {
	return c.GetHashtag(ctx, in.Name)
}

// GetHashtagPageInput is the typed input for threads_get_hashtag_page.
type GetHashtagPageInput struct {
	Name   string `json:"name" jsonschema:"description=hashtag name without the leading #,required"`
	Count  int    `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	Cursor string `json:"cursor,omitempty" jsonschema:"description=opaque pagination cursor from a previous response (next_cursor)"`
}

func getHashtagPage(ctx context.Context, c *threads.Client, in GetHashtagPageInput) (any, error) {
	return c.GetHashtagPage(ctx, in.Name, in.Count, in.Cursor)
}

var hashtagTools = []mcptool.Tool{
	mcptool.Define[*threads.Client, SearchHashtagsInput](
		"threads_search_hashtags",
		"Search Threads hashtags by name fragment",
		"SearchHashtags",
		searchHashtags,
	),
	mcptool.Define[*threads.Client, GetHashtagInput](
		"threads_get_hashtag",
		"Fetch the first page of posts for a hashtag along with its metadata",
		"GetHashtag",
		getHashtag,
	),
	mcptool.Define[*threads.Client, GetHashtagPageInput](
		"threads_get_hashtag_page",
		"Fetch a paginated page of posts for a hashtag",
		"GetHashtagPage",
		getHashtagPage,
	),
}
