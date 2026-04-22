package threads

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetLikers fetches users who liked a given post. The Threads likers
// endpoint returns the full set in one shot (no cursor) up to a server-imposed
// limit. The HasNext field will always be false for this endpoint.
func (c *Client) GetLikers(ctx context.Context, postID string) (UserPage, error) {
	if postID == "" {
		return UserPage{}, fmt.Errorf("%w: postID must not be empty", ErrInvalidParams)
	}
	body, err := c.readGET(ctx, "/api/v1/text_feed/"+postID+"/likers/", nil)
	if err != nil {
		return UserPage{}, err
	}
	var resp struct {
		Users      []userPayload `json:"users"`
		UserCount  int           `json:"user_count"`
		Status     string        `json:"status"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return UserPage{}, fmt.Errorf("%w: decoding likers: %v", ErrRequestFailed, err)
	}
	page := UserPage{NumResults: resp.UserCount}
	for _, u := range resp.Users {
		page.Users = append(page.Users, toUser(u))
	}
	return page, nil
}

// GetReposters fetches users who reposted a given post. Same shape as GetLikers.
func (c *Client) GetReposters(ctx context.Context, postID string) (UserPage, error) {
	if postID == "" {
		return UserPage{}, fmt.Errorf("%w: postID must not be empty", ErrInvalidParams)
	}
	body, err := c.readGET(ctx, "/api/v1/text_feed/"+postID+"/reposters/", nil)
	if err != nil {
		return UserPage{}, err
	}
	var resp struct {
		Users      []userPayload `json:"users"`
		UserCount  int           `json:"user_count"`
		Status     string        `json:"status"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return UserPage{}, fmt.Errorf("%w: decoding reposters: %v", ErrRequestFailed, err)
	}
	page := UserPage{NumResults: resp.UserCount}
	for _, u := range resp.Users {
		page.Users = append(page.Users, toUser(u))
	}
	return page, nil
}

// GetQuoters fetches users who quote-reposted a given post.
func (c *Client) GetQuoters(ctx context.Context, postID string) (UserPage, error) {
	if postID == "" {
		return UserPage{}, fmt.Errorf("%w: postID must not be empty", ErrInvalidParams)
	}
	body, err := c.readGET(ctx, "/api/v1/text_feed/"+postID+"/quoters/", nil)
	if err != nil {
		return UserPage{}, err
	}
	var resp struct {
		Users      []userPayload `json:"users"`
		UserCount  int           `json:"user_count"`
		Status     string        `json:"status"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return UserPage{}, fmt.Errorf("%w: decoding quoters: %v", ErrRequestFailed, err)
	}
	page := UserPage{NumResults: resp.UserCount}
	for _, u := range resp.Users {
		page.Users = append(page.Users, toUser(u))
	}
	return page, nil
}
