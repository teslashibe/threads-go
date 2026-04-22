package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// SearchPosts performs a unified text search and returns matching posts.
// Threads' search endpoint accepts an optional rank_token; pass "" for the
// first request. To paginate use the returned NextCursor.
func (c *Client) SearchPosts(ctx context.Context, query string, count int, cursor string) (PostPage, error) {
	if query == "" {
		return PostPage{}, fmt.Errorf("%w: query must not be empty", ErrInvalidParams)
	}
	if count <= 0 {
		count = 25
	}
	params := url.Values{
		"q":     []string{query},
		"count": []string{strconv.Itoa(count)},
	}
	if cursor != "" {
		params.Set("max_id", cursor)
	}
	body, err := c.readGET(ctx, "/api/v1/text_feed/text_search/", params)
	if err != nil {
		return PostPage{}, err
	}
	return parseThreadFeed(body)
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
