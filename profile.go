package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// GetProfile retrieves a user's full profile by numeric user ID.
//
//	user, err := c.GetProfile(ctx, "314216")
func (c *Client) GetProfile(ctx context.Context, userID string) (*User, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: userID must not be empty", ErrInvalidParams)
	}
	body, err := c.readGET(ctx, "/api/v1/users/"+userID+"/info/", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		User userPayload `json:"user"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("%w: decoding profile: %v", ErrRequestFailed, err)
	}
	if stringID(resp.User.PK) == "" && resp.User.PKID == "" {
		return nil, ErrNotFound
	}
	u := toUser(resp.User)
	return &u, nil
}

// GetProfileExtended retrieves a user's profile with the extra fields
// returned when ?entry_point=profile&from_module=profile_page is set.
// Use this when you need fields like text_app_is_low_like, etc.
func (c *Client) GetProfileExtended(ctx context.Context, userID string) (*User, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: userID must not be empty", ErrInvalidParams)
	}
	params := url.Values{
		"entry_point": []string{"profile"},
		"from_module": []string{"profile_page"},
	}
	body, err := c.readGET(ctx, "/api/v1/users/"+userID+"/info/", params)
	if err != nil {
		return nil, err
	}
	var resp struct {
		User userPayload `json:"user"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("%w: decoding profile: %v", ErrRequestFailed, err)
	}
	if stringID(resp.User.PK) == "" && resp.User.PKID == "" {
		return nil, ErrNotFound
	}
	u := toUser(resp.User)
	return &u, nil
}

// GetProfileByUsername resolves a username to a numeric user ID via search,
// then fetches the full profile. Returns ErrNotFound if no exact match.
//
// Threads has no /users/lookup/?username= endpoint (returns 405) so this
// goes through search and exact-matches the username field.
func (c *Client) GetProfileByUsername(ctx context.Context, username string) (*User, error) {
	if username == "" {
		return nil, fmt.Errorf("%w: username must not be empty", ErrInvalidParams)
	}
	page, err := c.SearchUsers(ctx, username, 10)
	if err != nil {
		return nil, err
	}
	for _, u := range page.Users {
		if u.Username == username {
			return c.GetProfile(ctx, u.ID)
		}
	}
	return nil, ErrNotFound
}

// SearchUsers searches Threads / Instagram users by query. count is capped
// server-side at ~50.
//
//	page, err := c.SearchUsers(ctx, "zuck", 20)
func (c *Client) SearchUsers(ctx context.Context, query string, count int) (UserPage, error) {
	if query == "" {
		return UserPage{}, fmt.Errorf("%w: query must not be empty", ErrInvalidParams)
	}
	if count <= 0 {
		count = 20
	}
	params := url.Values{
		"q":     []string{query},
		"count": []string{strconv.Itoa(count)},
	}
	body, err := c.readGET(ctx, "/api/v1/users/search/", params)
	if err != nil {
		return UserPage{}, err
	}
	var resp struct {
		NumResults int           `json:"num_results"`
		Users      []userPayload `json:"users"`
		HasMore    bool          `json:"has_more"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return UserPage{}, fmt.Errorf("%w: decoding search: %v", ErrRequestFailed, err)
	}
	page := UserPage{NumResults: resp.NumResults, HasNext: resp.HasMore}
	for _, u := range resp.Users {
		page.Users = append(page.Users, toUser(u))
	}
	return page, nil
}
