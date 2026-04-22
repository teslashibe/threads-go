package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// SearchHashtags searches Threads hashtags by query.
//
// Endpoint: GET /api/v1/tags/search/
func (c *Client) SearchHashtags(ctx context.Context, query string, count int) ([]Hashtag, error) {
	if query == "" {
		return nil, fmt.Errorf("%w: query must not be empty", ErrInvalidParams)
	}
	if count <= 0 {
		count = 30
	}
	params := url.Values{
		"q":     []string{query},
		"count": []string{strconv.Itoa(count)},
	}
	body, err := c.readGET(ctx, "/api/v1/tags/search/", params)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Results []hashtagPayload `json:"results"`
		Status  string           `json:"status"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("%w: decoding hashtag search: %v", ErrRequestFailed, err)
	}
	out := make([]Hashtag, 0, len(resp.Results))
	for _, h := range resp.Results {
		out = append(out, toHashtag(h))
	}
	return out, nil
}

// GetHashtag fetches metadata (name, id, post counts, follow status) for a hashtag.
//
// This returns metadata only. Meta does not currently expose the full
// hashtag post feed via the public www.threads.com API; only the sections
// navigation endpoint is reachable, and it returns an empty list for
// non-mobile clients. For end-to-end tag discovery use SearchPosts with
// "#name" as the query instead.
//
// Endpoint: GET /api/v1/tags/web_info/?tag_name={name}
func (c *Client) GetHashtag(ctx context.Context, name string) (*HashtagFeed, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: hashtag name must not be empty", ErrInvalidParams)
	}
	params := url.Values{"tag_name": []string{name}}
	body, err := c.readGET(ctx, "/api/v1/tags/web_info/", params)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data hashtagPayload `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("%w: decoding hashtag info: %v", ErrRequestFailed, err)
	}
	return &HashtagFeed{Hashtag: toHashtag(resp.Data)}, nil
}

// GetHashtagPage is intentionally not implemented.
//
// Meta has gated the per-tag post feed behind the sections endpoint
// (POST /api/v1/tags/{name}/sections/), which currently returns only
// a tab-navigation payload (no post data) for cookie-authenticated web
// callers. Use SearchPosts(ctx, "#"+name, …) as a workaround until
// Meta restores a usable surface.
func (c *Client) GetHashtagPage(ctx context.Context, name string, count int, cursor string) (*HashtagFeed, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: hashtag name must not be empty", ErrInvalidParams)
	}
	return nil, fmt.Errorf("%w: hashtag post feed is not currently exposed by the Threads web API; use SearchPosts(\"#%s\", ...)", ErrNotFound, name)
}
