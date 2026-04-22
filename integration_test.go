//go:build integration

// integration_test.go runs read-only end-to-end tests against the live
// Threads API using session cookies from the environment.
//
// Required env vars:
//   THREADS_SESSIONID, THREADS_CSRFTOKEN, THREADS_DS_USER_ID,
//   THREADS_MID, THREADS_IG_DID
//
// Optional:
//   THREADS_TEST_USERNAME (default: "zuck")  — public account used for lookups
//   THREADS_TEST_HASHTAG  (default: "golang") — hashtag used for feed test
//
// Run with:
//   go test -tags=integration -v -run TestIntegration_ ./...
package threads

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

func cookiesFromEnv() Cookies {
	return Cookies{
		SessionID: os.Getenv("THREADS_SESSIONID"),
		CSRFToken: os.Getenv("THREADS_CSRFTOKEN"),
		DSUserID:  os.Getenv("THREADS_DS_USER_ID"),
		Mid:       os.Getenv("THREADS_MID"),
		IgDid:     os.Getenv("THREADS_IG_DID"),
	}
}

func mustClient(t *testing.T) *Client {
	t.Helper()
	cs := cookiesFromEnv()
	if cs.SessionID == "" || cs.CSRFToken == "" {
		t.Skip("THREADS_SESSIONID / THREADS_CSRFTOKEN not set; skipping integration test")
	}
	c, err := New(cs, WithMinRequestGap(4*time.Second))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func testUsername() string {
	if u := os.Getenv("THREADS_TEST_USERNAME"); u != "" {
		return u
	}
	return "zuck"
}

func testHashtag() string {
	if h := os.Getenv("THREADS_TEST_HASHTAG"); h != "" {
		return h
	}
	return "golang"
}

func TestIntegration_Session(t *testing.T) {
	c := mustClient(t)

	me, err := c.Me(context.Background())
	if err != nil {
		t.Fatalf("Me: %v", err)
	}
	t.Logf("Logged in as @%s id=%s followers=%d following=%d verified=%v",
		me.Username, me.ID, me.FollowerCount, me.FollowingCount, me.IsVerified)
	if me.ID == "" {
		t.Error("viewer ID is empty")
	}
	if me.Username == "" {
		t.Error("viewer username is empty")
	}
}

func TestIntegration_GetProfileByUsername(t *testing.T) {
	c := mustClient(t)
	ctx := context.Background()

	u, err := c.GetProfileByUsername(ctx, testUsername())
	if err != nil {
		t.Fatalf("GetProfileByUsername(%q): %v", testUsername(), err)
	}
	t.Logf("@%s id=%s followers=%d following=%d verified=%v posts=%d",
		u.Username, u.ID, u.FollowerCount, u.FollowingCount, u.IsVerified, u.MediaCount)
	if u.Username != testUsername() {
		t.Errorf("expected username=%q, got %q", testUsername(), u.Username)
	}
	if u.ID == "" {
		t.Error("profile ID is empty")
	}
}

func TestIntegration_SearchUsers(t *testing.T) {
	c := mustClient(t)
	ctx := context.Background()

	page, err := c.SearchUsers(ctx, testUsername(), 10)
	if err != nil {
		t.Fatalf("SearchUsers: %v", err)
	}
	t.Logf("SearchUsers(%q): %d users (numResults=%d)", testUsername(), len(page.Users), page.NumResults)
	if len(page.Users) == 0 {
		t.Errorf("expected at least one user for query %q", testUsername())
	}
	for i, u := range page.Users {
		if i >= 3 {
			break
		}
		t.Logf("  %d. @%s — %s", i+1, u.Username, u.FullName)
	}
}

func TestIntegration_UserThreads(t *testing.T) {
	c := mustClient(t)
	ctx := context.Background()

	u, err := c.GetProfileByUsername(ctx, testUsername())
	if err != nil {
		t.Fatalf("lookup %q: %v", testUsername(), err)
	}
	page, err := c.UserThreads(ctx, u.ID, 10, "")
	if err != nil {
		t.Fatalf("UserThreads(%s): %v", u.ID, err)
	}
	t.Logf("UserThreads(@%s): %d threads, hasNext=%v cursor=%q",
		u.Username, len(page.Threads), page.HasNext, page.NextCursor)
	if len(page.Threads) == 0 {
		t.Skipf("@%s has no public threads — try a different account", u.Username)
	}
	posts := flattenThreads(page.Threads)
	if len(posts) == 0 {
		t.Error("threads contained no posts")
	}
	for i, p := range posts {
		if i >= 3 {
			break
		}
		t.Logf("  post %s likes=%d replies=%d at %s — %.80s",
			p.ID, p.LikeCount, p.ReplyCount, p.TakenAt.Format("2006-01-02"), p.Text)
	}
}

func TestIntegration_UserThreadsPagination(t *testing.T) {
	c := mustClient(t)
	ctx := context.Background()

	u, err := c.GetProfileByUsername(ctx, testUsername())
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	it := NewUserThreadsIterator(c, u.ID, 10, WithMaxPosts(20))
	pages := 0
	for it.Next(ctx) {
		pages++
		t.Logf("page %d: %d posts (cumulative: %d)", pages, len(it.Page()), it.Seen())
	}
	if err := it.Err(); err != nil {
		t.Fatalf("iterator: %v", err)
	}
	t.Logf("done — pages=%d total=%d", pages, it.Seen())
}

func TestIntegration_GetThread(t *testing.T) {
	c := mustClient(t)
	ctx := context.Background()

	u, err := c.GetProfileByUsername(ctx, testUsername())
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	page, err := c.UserThreads(ctx, u.ID, 5, "")
	if err != nil {
		t.Fatalf("UserThreads: %v", err)
	}
	if len(page.Threads) == 0 || len(page.Threads[0].ThreadItems) == 0 {
		t.Skipf("no threads to inspect")
	}
	postID := page.Threads[0].ThreadItems[0].ID

	tc, err := c.GetThread(ctx, postID)
	if err != nil {
		t.Fatalf("GetThread(%s): %v", postID, err)
	}
	t.Logf("Thread %s: containingItems=%d replies=%d siblings=%d cursor=%q",
		tc.TargetPostID, len(tc.ContainingThread.ThreadItems), len(tc.ReplyThreads),
		len(tc.SiblingThreads), tc.DownwardCursor)
}

func TestIntegration_GetLikers(t *testing.T) {
	c := mustClient(t)
	ctx := context.Background()

	u, err := c.GetProfileByUsername(ctx, testUsername())
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	page, err := c.UserThreads(ctx, u.ID, 5, "")
	if err != nil {
		t.Fatalf("UserThreads: %v", err)
	}
	if len(page.Threads) == 0 || len(page.Threads[0].ThreadItems) == 0 {
		t.Skipf("no threads")
	}
	postID := page.Threads[0].ThreadItems[0].ID

	likers, err := c.GetLikers(ctx, postID)
	if err != nil {
		t.Fatalf("GetLikers(%s): %v", postID, err)
	}
	t.Logf("Likers of %s: %d users (numResults=%d)", postID, len(likers.Users), likers.NumResults)
	for i, u := range likers.Users {
		if i >= 3 {
			break
		}
		t.Logf("  @%s (%s)", u.Username, u.FullName)
	}
}

func TestIntegration_GetFriendship(t *testing.T) {
	c := mustClient(t)
	ctx := context.Background()

	u, err := c.GetProfileByUsername(ctx, testUsername())
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	fs, err := c.GetFriendship(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetFriendship(%s): %v", u.ID, err)
	}
	t.Logf("Friendship with @%s: following=%v followedBy=%v blocking=%v muting=%v",
		u.Username, fs.Following, fs.FollowedBy, fs.Blocking, fs.Muting)
}

func TestIntegration_SearchHashtags(t *testing.T) {
	c := mustClient(t)
	ctx := context.Background()

	tags, err := c.SearchHashtags(ctx, testHashtag(), 10)
	if err != nil {
		t.Fatalf("SearchHashtags(%q): %v", testHashtag(), err)
	}
	t.Logf("Hashtag search %q: %d tags", testHashtag(), len(tags))
	for i, h := range tags {
		if i >= 5 {
			break
		}
		t.Logf("  #%s (media=%d)", h.Name, h.MediaCount)
	}
}

func TestIntegration_GetHashtag(t *testing.T) {
	c := mustClient(t)
	ctx := context.Background()

	feed, err := c.GetHashtag(ctx, testHashtag())
	if err != nil {
		t.Fatalf("GetHashtag(%q): %v", testHashtag(), err)
	}
	t.Logf("#%s feed: %d threads, hasNext=%v media=%d",
		feed.Hashtag.Name, len(feed.Threads), feed.HasNext, feed.Hashtag.MediaCount)
	posts := flattenThreads(feed.Threads)
	for i, p := range posts {
		if i >= 3 {
			break
		}
		t.Logf("  @%s — likes=%d  %.80s", p.User.Username, p.LikeCount, p.Text)
	}
}

func TestIntegration_GetFollowing(t *testing.T) {
	c := mustClient(t)
	ctx := context.Background()

	me, err := c.Me(ctx)
	if err != nil {
		t.Fatalf("Me: %v", err)
	}
	page, err := c.GetFollowing(ctx, me.ID, 10, "")
	if err != nil {
		t.Fatalf("GetFollowing(self): %v", err)
	}
	t.Logf("I follow %d users (page 1; hasNext=%v)", len(page.Users), page.HasNext)
}

func TestIntegration_LikedPosts(t *testing.T) {
	c := mustClient(t)
	ctx := context.Background()

	// Meta no longer exposes a Threads-only liked feed via the web API.
	// LikedPosts is documented to return ErrNotFound; we assert that
	// behaviour so a future Meta change forces us to revisit.
	_, err := c.LikedPosts(ctx, 10, "")
	if err == nil {
		t.Fatalf("LikedPosts: expected ErrNotFound, got nil — Meta may have restored the endpoint")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("LikedPosts: expected ErrNotFound, got %v", err)
	}
	t.Logf("LikedPosts correctly returns ErrNotFound: %v", err)
}

func TestIntegration_SearchPosts(t *testing.T) {
	c := mustClient(t)
	ctx := context.Background()

	page, err := c.SearchPosts(ctx, testHashtag(), 10, "")
	if err != nil {
		// Search endpoint may be gated; treat as soft-fail.
		t.Logf("SearchPosts(%q): %v (endpoint may be gated)", testHashtag(), err)
		return
	}
	posts := flattenThreads(page.Threads)
	t.Logf("SearchPosts(%q): %d threads / %d posts", testHashtag(), len(page.Threads), len(posts))
}
