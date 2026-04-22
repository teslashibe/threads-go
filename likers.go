package threads

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetLikers fetches users who liked a given post.
//
// Meta returns the full list in one shot (no cursor) up to a server-imposed
// limit; HasNext is therefore always false on the returned page.
//
// Note on visibility: Meta increasingly hides identities of likers from
// non-author callers. When that happens, the response is HTTP 200 with
// Users=[] and NumResults set to the public like count. Treat NumResults
// as authoritative and Users as best-effort.
func (c *Client) GetLikers(ctx context.Context, postID string) (UserPage, error) {
	if postID == "" {
		return UserPage{}, fmt.Errorf("%w: postID must not be empty", ErrInvalidParams)
	}
	body, err := c.readGET(ctx, "/api/v1/media/"+postID+"/likers/", nil)
	if err != nil {
		return UserPage{}, err
	}
	var resp struct {
		Users     []userPayload `json:"users"`
		UserCount int           `json:"user_count"`
		Status    string        `json:"status"`
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

// GetReposters is not supported via the Threads web read API.
//
// Meta has not exposed a public reposters-list endpoint on www.threads.com;
// only the aggregate repost count is returned with the post itself
// (see Post.RepostCount). This stub is kept so callers get a typed error
// rather than a 404 from the underlying transport.
func (c *Client) GetReposters(ctx context.Context, postID string) (UserPage, error) {
	if postID == "" {
		return UserPage{}, fmt.Errorf("%w: postID must not be empty", ErrInvalidParams)
	}
	return UserPage{}, fmt.Errorf("%w: reposters list is not exposed by the Threads web API; use Post.RepostCount", ErrNotFound)
}

// GetQuoters is not supported via the Threads web read API.
//
// Meta has not exposed a public quoters-list endpoint on www.threads.com;
// only the aggregate quote count is returned with the post itself.
// Use the repost feed of an individual user to discover quote-reposts they
// have authored.
func (c *Client) GetQuoters(ctx context.Context, postID string) (UserPage, error) {
	if postID == "" {
		return UserPage{}, fmt.Errorf("%w: postID must not be empty", ErrInvalidParams)
	}
	return UserPage{}, fmt.Errorf("%w: quoters list is not exposed by the Threads web API", ErrNotFound)
}
