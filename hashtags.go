package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// SearchHashtags searches Threads hashtags by query.
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
	body, err := c.readGET(ctx, "/api/v1/text_app/tags/search/", params)
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

// GetHashtag fetches metadata + the recent feed for a hashtag.
//
//	feed, err := c.GetHashtag(ctx, "golang")
func (c *Client) GetHashtag(ctx context.Context, name string) (*HashtagFeed, error) {
	return c.GetHashtagPage(ctx, name, 0, "")
}

// GetHashtagPage fetches a paginated page of a hashtag's feed.
func (c *Client) GetHashtagPage(ctx context.Context, name string, count int, cursor string) (*HashtagFeed, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: hashtag name must not be empty", ErrInvalidParams)
	}
	params := url.Values{}
	if count > 0 {
		params.Set("count", strconv.Itoa(count))
	}
	if cursor != "" {
		params.Set("max_id", cursor)
	}
	body, err := c.readGET(ctx, "/api/v1/text_app/tags/"+url.PathEscape(name)+"/feed/", params)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Hashtag    hashtagPayload  `json:"hashtag"`
		Threads    []threadPayload `json:"threads"`
		NextMaxID  json.RawMessage `json:"next_max_id"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("%w: decoding hashtag feed: %v", ErrRequestFailed, err)
	}
	feed := &HashtagFeed{
		Hashtag: toHashtag(resp.Hashtag),
		Threads: toThreads(resp.Threads),
	}
	feed.NextCursor = decodeMaxID(resp.NextMaxID)
	feed.HasNext = feed.NextCursor != ""
	return feed, nil
}
