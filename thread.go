package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// GetThread fetches a single post by its numeric ID, returning the post
// together with its reply context (root post, top-level replies, sibling
// threads, downward cursor).
//
// Pagination of further replies uses the DownwardCursor field. Pass it to
// GetThreadReplies for the next page.
func (c *Client) GetThread(ctx context.Context, postID string) (*ThreadContext, error) {
	return c.GetThreadReplies(ctx, postID, 0, "")
}

// GetThreadReplies fetches a thread's reply context with explicit count
// and downward cursor for pagination.
func (c *Client) GetThreadReplies(ctx context.Context, postID string, count int, downwardCursor string) (*ThreadContext, error) {
	if postID == "" {
		return nil, fmt.Errorf("%w: postID must not be empty", ErrInvalidParams)
	}
	params := url.Values{}
	if count > 0 {
		params.Set("count", strconv.Itoa(count))
	}
	if downwardCursor != "" {
		params.Set("paging_token", downwardCursor)
	}
	body, err := c.readGET(ctx, "/api/v1/text_feed/"+postID+"/replies/", params)
	if err != nil {
		return nil, err
	}
	var resp struct {
		TargetPostID                string          `json:"target_post_id"`
		ContainingThread            threadPayload   `json:"containing_thread"`
		ReplyThreads                []threadPayload `json:"reply_threads"`
		SiblingThreads              []threadPayload `json:"sibling_threads"`
		PagingTokens                struct {
			Downwards string `json:"downwards"`
		} `json:"paging_tokens"`
		DownwardsThreadWillContinue bool `json:"downwards_thread_will_continue"`
		IsSubscribedToTargetPost    bool `json:"is_subscribed_to_target_post"`
		IsAuthorOfRootPost          bool `json:"is_author_of_root_post"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("%w: decoding thread context: %v", ErrRequestFailed, err)
	}
	tc := &ThreadContext{
		TargetPostID:                resp.TargetPostID,
		ContainingThread:            toThread(resp.ContainingThread),
		ReplyThreads:                toThreads(resp.ReplyThreads),
		SiblingThreads:              toThreads(resp.SiblingThreads),
		DownwardCursor:              resp.PagingTokens.Downwards,
		DownwardsThreadWillContinue: resp.DownwardsThreadWillContinue,
		IsSubscribedToTargetPost:    resp.IsSubscribedToTargetPost,
		IsAuthorOfRootPost:          resp.IsAuthorOfRootPost,
	}
	return tc, nil
}

// GetPost fetches just the focal post of a thread (no replies). Convenience
// wrapper around GetThread that returns only the first thread item.
func (c *Client) GetPost(ctx context.Context, postID string) (*Post, error) {
	tc, err := c.GetThread(ctx, postID)
	if err != nil {
		return nil, err
	}
	if len(tc.ContainingThread.ThreadItems) == 0 {
		return nil, ErrNotFound
	}
	// The focal post matches TargetPostID; default to last item if the
	// thread is a chain (each preceding item is a parent context post).
	for i := len(tc.ContainingThread.ThreadItems) - 1; i >= 0; i-- {
		if tc.ContainingThread.ThreadItems[i].ID == tc.TargetPostID {
			p := tc.ContainingThread.ThreadItems[i]
			return &p, nil
		}
	}
	p := tc.ContainingThread.ThreadItems[len(tc.ContainingThread.ThreadItems)-1]
	return &p, nil
}
