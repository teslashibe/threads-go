package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// SearchPosts is not currently supported via the Threads web read API.
//
// Meta has retired the REST text-search endpoint
// (/api/v1/text_feed/text_search/) and the web SERP page now drives all
// post search through opaque, rotating GraphQL doc_ids. Mapping that
// surface from a Go SDK is brittle and out of scope for the v1 of
// threads-go. SearchHashtags + SearchUsers remain available; for ad-hoc
// post discovery, walk a known user/hashtag and filter client-side.
//
// _ = ctx, _ = query, _ = count, _ = cursor — kept for forward-compatible signature.
func (c *Client) SearchPosts(ctx context.Context, query string, count int, cursor string) (PostPage, error) {
	_ = ctx
	_ = count
	_ = cursor
	if query == "" {
		return PostPage{}, fmt.Errorf("%w: query must not be empty", ErrInvalidParams)
	}
	return PostPage{}, fmt.Errorf("%w: text-post search is not currently exposed by the Threads web REST API", ErrNotFound)
}

// RecommendedUsers returns suggested users to follow. This endpoint is
// served by i.instagram.com and requires Bearer auth.
func (c *Client) RecommendedUsers(ctx context.Context, count int) (UserPage, error) {
	if count <= 0 {
		count = 30
	}
	params := url.Values{
		"phone_id":           []string{c.auth.DeviceID},
		"module":             []string{"discover_people"},
		"paginate":           []string{"true"},
		"max_id":             []string{""},
		"count":              []string{strconv.Itoa(count)},
	}
	body, err := c.writeGET(ctx, "/api/v1/text_feed/recommended_users/", params)
	if err != nil {
		return UserPage{}, err
	}
	var resp struct {
		Users []struct {
			User userPayload `json:"user"`
		} `json:"users"`
		NextMaxID json.RawMessage `json:"next_max_id"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return UserPage{}, fmt.Errorf("%w: decoding recommended users: %v", ErrRequestFailed, err)
	}
	page := UserPage{}
	for _, u := range resp.Users {
		page.Users = append(page.Users, toUser(u.User))
	}
	page.NextCursor = decodeMaxID(resp.NextMaxID)
	page.HasNext = page.NextCursor != ""
	return page, nil
}
