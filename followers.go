package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// GetFollowers fetches a single page of users following the given user.
// Cursor uses max_id (string).
func (c *Client) GetFollowers(ctx context.Context, userID string, count int, cursor string) (UserPage, error) {
	if userID == "" {
		return UserPage{}, fmt.Errorf("%w: userID must not be empty", ErrInvalidParams)
	}
	params := url.Values{}
	if count > 0 {
		params.Set("count", strconv.Itoa(count))
	}
	if cursor != "" {
		params.Set("max_id", cursor)
	}
	body, err := c.readGET(ctx, "/api/v1/friendships/"+userID+"/followers/", params)
	if err != nil {
		return UserPage{}, err
	}
	return parseUserList(body)
}

// GetFollowing fetches a single page of users the given user follows.
func (c *Client) GetFollowing(ctx context.Context, userID string, count int, cursor string) (UserPage, error) {
	if userID == "" {
		return UserPage{}, fmt.Errorf("%w: userID must not be empty", ErrInvalidParams)
	}
	params := url.Values{}
	if count > 0 {
		params.Set("count", strconv.Itoa(count))
	}
	if cursor != "" {
		params.Set("max_id", cursor)
	}
	body, err := c.readGET(ctx, "/api/v1/friendships/"+userID+"/following/", params)
	if err != nil {
		return UserPage{}, err
	}
	return parseUserList(body)
}

// GetFriendship returns the relationship between the authenticated user
// and the target user (following, blocking, muting, restricted, etc.).
func (c *Client) GetFriendship(ctx context.Context, userID string) (*FriendshipStatus, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: userID must not be empty", ErrInvalidParams)
	}
	body, err := c.readGET(ctx, "/api/v1/friendships/show/"+userID+"/", nil)
	if err != nil {
		return nil, err
	}
	var p friendshipPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, fmt.Errorf("%w: decoding friendship: %v", ErrRequestFailed, err)
	}
	fs := toFriendship(&p)
	if fs.UserID == "" {
		fs.UserID = userID
	}
	return &fs, nil
}

// GetFriendships returns relationships in bulk for multiple user IDs.
func (c *Client) GetFriendships(ctx context.Context, userIDs []string) (map[string]FriendshipStatus, error) {
	if len(userIDs) == 0 {
		return map[string]FriendshipStatus{}, nil
	}
	form := url.Values{}
	form.Set("user_ids", joinIDs(userIDs))
	body, err := c.readPOSTForm(ctx, "/api/v1/friendships/show_many/", form)
	if err != nil {
		return nil, err
	}
	var resp struct {
		FriendshipStatuses map[string]friendshipPayload `json:"friendship_statuses"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("%w: decoding bulk friendships: %v", ErrRequestFailed, err)
	}
	out := make(map[string]FriendshipStatus, len(resp.FriendshipStatuses))
	for id, p := range resp.FriendshipStatuses {
		fs := toFriendship(&p)
		if fs.UserID == "" {
			fs.UserID = id
		}
		out[id] = fs
	}
	return out, nil
}

// PendingRequests returns users with pending follow requests to the
// authenticated user (only meaningful for private accounts).
func (c *Client) PendingRequests(ctx context.Context) (UserPage, error) {
	body, err := c.readGET(ctx, "/api/v1/friendships/pending/", nil)
	if err != nil {
		return UserPage{}, err
	}
	return parseUserList(body)
}

// parseUserList decodes the {users: [...]} envelope used by friendships,
// followers, and following endpoints.
func parseUserList(body []byte) (UserPage, error) {
	var resp struct {
		Users     []userPayload   `json:"users"`
		BigList   bool            `json:"big_list"`
		PageSize  int             `json:"page_size"`
		NextMaxID json.RawMessage `json:"next_max_id"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return UserPage{}, fmt.Errorf("%w: decoding user list: %v", ErrRequestFailed, err)
	}
	page := UserPage{}
	for _, u := range resp.Users {
		page.Users = append(page.Users, toUser(u))
	}
	page.NextCursor = decodeMaxID(resp.NextMaxID)
	page.HasNext = page.NextCursor != ""
	return page, nil
}

func joinIDs(ids []string) string {
	out := ""
	for i, id := range ids {
		if i > 0 {
			out += ","
		}
		out += id
	}
	return out
}
