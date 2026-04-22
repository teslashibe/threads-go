package threads

import (
	"context"
	"encoding/json"
	"time"
)

// Checkpoint is a serialisable position in a paginated read iteration.
// Save it between runs and pass it to WithCheckpoint to resume.
type Checkpoint struct {
	Cursor      string    `json:"cursor,omitempty"`
	LastPostID  string    `json:"lastPostId,omitempty"`
	PostsSeen   int       `json:"postsSeen"`
	CreatedAt   time.Time `json:"createdAt"`
	Subject     string    `json:"subject,omitempty"` // userID, hashtag, etc.
}

// Marshal serialises a Checkpoint to JSON bytes.
func (cp Checkpoint) Marshal() ([]byte, error) { return json.Marshal(cp) }

// UnmarshalCheckpoint deserialises a Checkpoint from JSON bytes.
func UnmarshalCheckpoint(data []byte) (Checkpoint, error) {
	var cp Checkpoint
	err := json.Unmarshal(data, &cp)
	return cp, err
}

// PostIterator walks through paginated post results one page at a time.
// Use the factory functions: NewUserThreadsIterator, NewUserRepliesIterator,
// NewLikedPostsIterator, NewHashtagIterator, NewSearchPostsIterator.
type PostIterator struct {
	client    *Client
	fetchFn   func(ctx context.Context, cursor string) (PostPage, error)
	cursor    string
	lastID    string
	seen      int
	maxPosts  int
	stopAtID  string
	subject   string
	page      []Post
	threads   []Thread
	done      bool
	err       error
}

// IteratorOption configures a PostIterator.
type IteratorOption func(*PostIterator)

// WithMaxPosts caps total posts returned across all pages.
func WithMaxPosts(n int) IteratorOption {
	return func(it *PostIterator) { it.maxPosts = n }
}

// WithStopAtID stops iteration when a post with this ID is encountered.
// Useful for incremental scraping (resume to a known-good last post).
func WithStopAtID(postID string) IteratorOption {
	return func(it *PostIterator) { it.stopAtID = postID }
}

// WithCheckpoint resumes iteration from a saved Checkpoint.
func WithCheckpoint(cp Checkpoint) IteratorOption {
	return func(it *PostIterator) {
		it.cursor = cp.Cursor
		it.lastID = cp.LastPostID
		it.seen = cp.PostsSeen
		if cp.Subject != "" {
			it.subject = cp.Subject
		}
	}
}

// Next advances to the next page. Returns false when iteration is complete.
// Call Page() to read the page's posts (flattened from threads).
func (it *PostIterator) Next(ctx context.Context) bool {
	if it.done {
		return false
	}
	if ctx.Err() != nil {
		it.err = ctx.Err()
		it.done = true
		return false
	}
	if it.maxPosts > 0 && it.seen >= it.maxPosts {
		it.done = true
		return false
	}

	page, err := it.fetchFn(ctx, it.cursor)
	if err != nil {
		it.err = err
		it.done = true
		return false
	}

	posts := flattenThreads(page.Threads)
	if len(posts) == 0 {
		it.done = true
		return false
	}

	if it.stopAtID != "" {
		trimmed := make([]Post, 0, len(posts))
		for _, p := range posts {
			if p.ID == it.stopAtID {
				it.done = true
				break
			}
			trimmed = append(trimmed, p)
		}
		posts = trimmed
		if len(posts) == 0 {
			return false
		}
	}

	if it.maxPosts > 0 {
		remaining := it.maxPosts - it.seen
		if remaining < len(posts) {
			posts = posts[:remaining]
			it.done = true
		}
	}

	it.page = posts
	it.threads = page.Threads
	it.seen += len(posts)
	if len(posts) > 0 {
		it.lastID = posts[len(posts)-1].ID
	}
	if page.HasNext && page.NextCursor != "" && !it.done {
		it.cursor = page.NextCursor
	} else {
		it.done = true
	}
	return len(it.page) > 0
}

// Page returns the flattened posts from the most recent Next() call.
func (it *PostIterator) Page() []Post { return it.page }

// Threads returns the raw Thread groups from the most recent Next() call.
// Useful when you need to preserve thread relationships (a single Thread
// may contain a chain of Posts).
func (it *PostIterator) Threads() []Thread { return it.threads }

// Err returns the first error encountered during iteration, if any.
func (it *PostIterator) Err() error { return it.err }

// Seen returns the total number of posts returned so far.
func (it *PostIterator) Seen() int { return it.seen }

// Checkpoint returns a serialisable position that can be used to resume.
func (it *PostIterator) Checkpoint() Checkpoint {
	return Checkpoint{
		Cursor:     it.cursor,
		LastPostID: it.lastID,
		PostsSeen:  it.seen,
		CreatedAt:  time.Now(),
		Subject:    it.subject,
	}
}

// flattenThreads collapses a slice of Threads into a flat slice of Posts,
// preserving order and including all chain items.
func flattenThreads(threads []Thread) []Post {
	var out []Post
	for _, t := range threads {
		out = append(out, t.ThreadItems...)
	}
	return out
}

// NewUserThreadsIterator iterates a user's profile thread feed.
func NewUserThreadsIterator(c *Client, userID string, pageSize int, opts ...IteratorOption) *PostIterator {
	it := &PostIterator{client: c, subject: userID}
	for _, o := range opts {
		o(it)
	}
	it.fetchFn = func(ctx context.Context, cursor string) (PostPage, error) {
		return c.UserThreads(ctx, userID, pageSize, cursor)
	}
	return it
}

// NewUserRepliesIterator iterates a user's reply feed.
func NewUserRepliesIterator(c *Client, userID string, pageSize int, opts ...IteratorOption) *PostIterator {
	it := &PostIterator{client: c, subject: userID}
	for _, o := range opts {
		o(it)
	}
	it.fetchFn = func(ctx context.Context, cursor string) (PostPage, error) {
		return c.UserReplies(ctx, userID, pageSize, cursor)
	}
	return it
}

// NewLikedPostsIterator iterates the authenticated user's liked posts.
func NewLikedPostsIterator(c *Client, pageSize int, opts ...IteratorOption) *PostIterator {
	it := &PostIterator{client: c, subject: "liked"}
	for _, o := range opts {
		o(it)
	}
	it.fetchFn = func(ctx context.Context, cursor string) (PostPage, error) {
		return c.LikedPosts(ctx, pageSize, cursor)
	}
	return it
}

// NewHashtagIterator iterates a hashtag's recent feed.
func NewHashtagIterator(c *Client, name string, pageSize int, opts ...IteratorOption) *PostIterator {
	it := &PostIterator{client: c, subject: "#" + name}
	for _, o := range opts {
		o(it)
	}
	it.fetchFn = func(ctx context.Context, cursor string) (PostPage, error) {
		feed, err := c.GetHashtagPage(ctx, name, pageSize, cursor)
		if err != nil {
			return PostPage{}, err
		}
		return PostPage{Threads: feed.Threads, NextCursor: feed.NextCursor, HasNext: feed.HasNext}, nil
	}
	return it
}

// NewSearchPostsIterator iterates Threads search results for a query.
func NewSearchPostsIterator(c *Client, query string, pageSize int, opts ...IteratorOption) *PostIterator {
	it := &PostIterator{client: c, subject: "q:" + query}
	for _, o := range opts {
		o(it)
	}
	it.fetchFn = func(ctx context.Context, cursor string) (PostPage, error) {
		return c.SearchPosts(ctx, query, pageSize, cursor)
	}
	return it
}

// NewHomeTimelineIterator iterates the authenticated user's home timeline.
// Requires Bearer auth.
func NewHomeTimelineIterator(c *Client, pageSize int, opts ...IteratorOption) *PostIterator {
	it := &PostIterator{client: c, subject: "home"}
	for _, o := range opts {
		o(it)
	}
	it.fetchFn = func(ctx context.Context, cursor string) (PostPage, error) {
		return c.HomeTimeline(ctx, pageSize, cursor)
	}
	return it
}

// UserIterator walks through paginated user results (followers, following).
type UserIterator struct {
	client    *Client
	fetchFn   func(ctx context.Context, cursor string) (UserPage, error)
	cursor    string
	lastID    string
	seen      int
	maxUsers  int
	subject   string
	page      []User
	done      bool
	err       error
}

// UserIteratorOption configures a UserIterator.
type UserIteratorOption func(*UserIterator)

// WithMaxUsers caps total users returned across all pages.
func WithMaxUsers(n int) UserIteratorOption {
	return func(it *UserIterator) { it.maxUsers = n }
}

// WithUserCheckpoint resumes from a Checkpoint.
func WithUserCheckpoint(cp Checkpoint) UserIteratorOption {
	return func(it *UserIterator) {
		it.cursor = cp.Cursor
		it.seen = cp.PostsSeen
		if cp.Subject != "" {
			it.subject = cp.Subject
		}
	}
}

// Next advances to the next page.
func (it *UserIterator) Next(ctx context.Context) bool {
	if it.done {
		return false
	}
	if ctx.Err() != nil {
		it.err = ctx.Err()
		it.done = true
		return false
	}
	if it.maxUsers > 0 && it.seen >= it.maxUsers {
		it.done = true
		return false
	}
	page, err := it.fetchFn(ctx, it.cursor)
	if err != nil {
		it.err = err
		it.done = true
		return false
	}
	if len(page.Users) == 0 {
		it.done = true
		return false
	}
	if it.maxUsers > 0 {
		remaining := it.maxUsers - it.seen
		if remaining < len(page.Users) {
			page.Users = page.Users[:remaining]
			it.done = true
		}
	}
	it.page = page.Users
	it.seen += len(page.Users)
	if len(page.Users) > 0 {
		it.lastID = page.Users[len(page.Users)-1].ID
	}
	if page.HasNext && page.NextCursor != "" && !it.done {
		it.cursor = page.NextCursor
	} else {
		it.done = true
	}
	return len(it.page) > 0
}

// Page returns the users from the most recent Next() call.
func (it *UserIterator) Page() []User { return it.page }

// Err returns the first error encountered.
func (it *UserIterator) Err() error { return it.err }

// Seen returns total users returned so far.
func (it *UserIterator) Seen() int { return it.seen }

// Checkpoint returns a saveable position.
func (it *UserIterator) Checkpoint() Checkpoint {
	return Checkpoint{
		Cursor:    it.cursor,
		PostsSeen: it.seen,
		CreatedAt: time.Now(),
		Subject:   it.subject,
	}
}

// NewFollowersIterator iterates a user's followers list.
func NewFollowersIterator(c *Client, userID string, pageSize int, opts ...UserIteratorOption) *UserIterator {
	it := &UserIterator{client: c, subject: "followers:" + userID}
	for _, o := range opts {
		o(it)
	}
	it.fetchFn = func(ctx context.Context, cursor string) (UserPage, error) {
		return c.GetFollowers(ctx, userID, pageSize, cursor)
	}
	return it
}

// NewFollowingIterator iterates a user's following list.
func NewFollowingIterator(c *Client, userID string, pageSize int, opts ...UserIteratorOption) *UserIterator {
	it := &UserIterator{client: c, subject: "following:" + userID}
	for _, o := range opts {
		o(it)
	}
	it.fetchFn = func(ctx context.Context, cursor string) (UserPage, error) {
		return c.GetFollowing(ctx, userID, pageSize, cursor)
	}
	return it
}
