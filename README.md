# threads-go

Go client for the [Threads](https://www.threads.com) (Meta) private API. Zero dependencies, cookie-based reads + Bearer-token writes.

```bash
go get github.com/teslashibe/threads-go
```

## Quick start

```go
import threads "github.com/teslashibe/threads-go"

// Reads (cookie auth)
c, _ := threads.New(threads.Cookies{
    SessionID: os.Getenv("THREADS_SESSIONID"),
    CSRFToken: os.Getenv("THREADS_CSRFTOKEN"),
    DSUserID:  os.Getenv("THREADS_DS_USER_ID"),
    Mid:       os.Getenv("THREADS_MID"),
    IgDid:     os.Getenv("THREADS_IG_DID"),
})
ctx := context.Background()

me,    _ := c.Me(ctx)
zuck,  _ := c.GetProfileByUsername(ctx, "zuck")
page,  _ := c.UserThreads(ctx, zuck.ID, 25, "")
ctx2,  _ := c.GetThread(ctx, page.Threads[0].ThreadItems[0].ID)
likers,_ := c.GetLikers(ctx, page.Threads[0].ThreadItems[0].ID)
feed,  _ := c.GetHashtag(ctx, "golang")

// Writes (Bearer auth from Bloks login)
login, _ := threads.Login(ctx, "user@example.com", "password", "")
w, _ := threads.NewWithAuth(threads.Auth{
    Token:    login.Token,
    UserID:   login.UserID,
    DeviceID: login.DeviceID,
})
post, _ := w.CreatePost(ctx, "Hello from threads-go!")
_ = w.Like(ctx, post.ID)
_ = w.Follow(ctx, zuck.ID)
_ = w.Repost(ctx, post.ID)
```

## Authentication

`threads-go` uses **two distinct auth modes** because Threads does too: reads go through `www.threads.com` with browser session cookies, writes go through `i.instagram.com` with a Bearer token from the Bloks login flow.

| Mode | Host | Constructor | Required credentials |
|---|---|---|---|
| Cookies (read) | `www.threads.com` | `New(Cookies)` | `sessionid`, `csrftoken`, `ds_user_id`, `mid`, `ig_did` |
| Bearer (write) | `i.instagram.com` | `NewWithAuth(Auth)` | `Token` (IGT:2:…), `UserID`, `DeviceID` |
| Both | both | `NewFull(Cookies, Auth)` | both sets |

### Read auth — export browser cookies

Open `https://www.threads.com` in a logged-in browser, open DevTools → Application → Cookies → `.threads.com`, and copy these five values:

```bash
export THREADS_SESSIONID="75472043478%3A...%3A11%3A..."
export THREADS_CSRFTOKEN="od4SlgoL4bFP0YTLqg89O8XlHyAYgFux"
export THREADS_DS_USER_ID="75472043478"
export THREADS_MID="aekP4AAEAAFB_7rh0M8tO5Hh244o"
export THREADS_IG_DID="F0FFA972-40C0-4CB3-9DFC-CF5BFA0A2BF8"
```

`sessionid` is URL-encoded (note the `%3A`'s); pass it through verbatim.

### Write auth — Bloks login

```go
res, err := threads.Login(ctx, "user@example.com", "password", "" /* device id */)
// or pass a stable device ID generated once:
//   deviceID := threads.GenerateDeviceID()
```

The token is a ~160-char `IGT:2:…` string and is valid for hours to days. Persist it together with the device ID; refresh by re-calling `Login` when you start getting `ErrUnauthorized`.

> The Bloks login is best-effort and may be challenged with checkpoints by Meta. For production use, hold a long-lived device ID and run the login from the same egress IP each time.

### Combined client

```go
c, _ := threads.NewFull(cookies, auth)   // can both read and write
me,   _ := c.Me(ctx)                      // uses cookies
post, _ := c.CreatePost(ctx, "hi")        // uses bearer
```

Calling a write method on a cookie-only client returns `ErrWriteAuthRequired`; calling a read method on a Bearer-only client returns `ErrUnauthorized`.

## Features

### Profiles

```go
me,    _ := c.Me(ctx)                                    // authenticated user
user,  _ := c.GetProfile(ctx, "314216")                  // by numeric ID
user,  _  = c.GetProfileExtended(ctx, "314216")          // adds extended fields
user,  _  = c.GetProfileByUsername(ctx, "zuck")          // resolve handle → profile
page,  _ := c.SearchUsers(ctx, "golang", 20)
```

`User` exposes `ID`, `Username`, `FullName`, `Biography`, `IsPrivate`, `IsVerified`, `FollowerCount`, `FollowingCount`, `MediaCount`, `ProfilePicURL`, `HDProfilePicURL`, plus Threads-specific badge / link fields.

### Threads (posts)

```go
page, _ := c.UserThreads(ctx, userID, 25, "")            // a user's posts
page, _  = c.UserReplies(ctx, userID, 25, "")            // a user's replies
page, _  = c.LikedPosts(ctx, 25, "")                     // self only
page, _  = c.SearchPosts(ctx, "golang", 25, "")          // text search
page, _  = c.HomeTimeline(ctx, 25, "")                   // For You — Bearer
```

Response: `PostPage{ Threads []Thread; NextCursor string; HasNext bool }`. Each `Thread` is a chain of one or more `Post`s (`ThreadItems`). Use `flattenThreads` (or the iterator helpers) to walk every post.

### Single thread / replies

```go
ctx2, _ := c.GetThread(ctx, postID)                      // post + reply context
ctx2, _  = c.GetThreadReplies(ctx, postID, 25, cursor)   // paginate replies
post, _ := c.GetPost(ctx, postID)                        // focal post only
```

`ThreadContext` returns the `ContainingThread` (root + parent posts), `ReplyThreads`, `SiblingThreads`, and `DownwardCursor` for paginating further replies.

### Engagement / likers

```go
likers,    _ := c.GetLikers(ctx, postID)                 // who liked
reposters, _ := c.GetReposters(ctx, postID)              // who reposted
quoters,   _ := c.GetQuoters(ctx, postID)                // who quoted
```

Likers/reposters/quoters return up to a server-imposed cap (typically 200) in a single response — `HasNext` is always false.

### Social graph

```go
followers, _ := c.GetFollowers(ctx, userID, 100, "")
following, _ := c.GetFollowing(ctx, userID, 100, "")
fs,        _ := c.GetFriendship(ctx, userID)             // single relationship
many,      _ := c.GetFriendships(ctx, []string{id1, id2})// bulk
pending,   _ := c.PendingRequests(ctx)                   // private-account requests
```

### Hashtags

```go
tags,  _ := c.SearchHashtags(ctx, "golang", 20)
feed,  _ := c.GetHashtag(ctx, "golang")
feed,  _  = c.GetHashtagPage(ctx, "golang", 25, cursor)
```

### Notifications & discovery (Bearer)

```go
notifs, _   := c.Notifications(ctx)
suggested,_ := c.RecommendedUsers(ctx, 30)
```

### Iterators

Long results paginate cleanly via the iterator helpers, with built-in
`MaxPosts` / `MaxUsers` caps, `StopAtID` for incremental scraping, and
serialisable `Checkpoint`s for resume across runs.

```go
it := threads.NewUserThreadsIterator(c, userID, 25,
    threads.WithMaxPosts(500),
    threads.WithStopAtID(lastSeenID),     // optional incremental boundary
)
for it.Next(ctx) {
    for _, p := range it.Page() { process(p) }
}
if err := it.Err(); err != nil { ... }
cp := it.Checkpoint()                      // save for next run
```

Available iterators:

| Iterator | Source |
|---|---|
| `NewUserThreadsIterator` | `UserThreads` |
| `NewUserRepliesIterator` | `UserReplies` |
| `NewLikedPostsIterator`  | `LikedPosts` |
| `NewHashtagIterator`     | `GetHashtagPage` |
| `NewSearchPostsIterator` | `SearchPosts` |
| `NewHomeTimelineIterator`| `HomeTimeline` (Bearer) |
| `NewFollowersIterator`   | `GetFollowers` |
| `NewFollowingIterator`   | `GetFollowing` |

### Writes (Bearer required)

```go
post, _ := c.CreatePost(ctx, "Hello world!")
post, _  = c.CreatePost(ctx, "Look at this cat",
    threads.WithImage("/tmp/cat.jpg"),
    threads.WithReplyControl("accounts_you_follow"),
)
post, _  = c.CreatePost(ctx, "Multi-pic",
    threads.WithImage("a.jpg"), threads.WithImage("b.jpg"),
)

reply, _ := c.Reply(ctx, postID, "Great point!")
quote, _ := c.Quote(ctx, postID, "Hot take →")

_ = c.Like(ctx, postID)
_ = c.Unlike(ctx, postID)
_ = c.Repost(ctx, postID)
_ = c.DeleteRepost(ctx, postID)
_ = c.DeletePost(ctx, postID)

_ = c.Follow(ctx, userID)
_ = c.Unfollow(ctx, userID)
_ = c.Block(ctx, userID)
_ = c.Unblock(ctx, userID)
_ = c.Mute(ctx, userID)
_ = c.Unmute(ctx, userID)
_ = c.Restrict(ctx, userID)
_ = c.Unrestrict(ctx, userID)

mediaID, _ := c.UploadImage(ctx, "/tmp/cat.jpg")        // or call directly
post, _   = c.CreatePost(ctx, "x", threads.WithMediaIDs(mediaID))
```

`CreatePost` chooses between the text-only, single-photo, and carousel
configure endpoints automatically based on attached images.

### Configuration options

```go
c, _ := threads.New(cookies,
    threads.WithUserAgent("Barcelona 289.0.0.14.109 Android"),
    threads.WithReadUserAgent(...),                // override only reads
    threads.WithWriteUserAgent(...),               // override only writes
    threads.WithMinRequestGap(2 * time.Second),    // pacing (default 1.5s)
    threads.WithRetry(3, 750*time.Millisecond),    // attempts, base
    threads.WithProxy("http://user:pass@proxy:8080"),
    threads.WithHTTPClient(myCustomClient),
)
```

## Errors

All errors are `errors.Is`-comparable to the sentinels:

| Sentinel | Meaning |
|---|---|
| `ErrInvalidAuth` | Required credentials missing from `Cookies` / `Auth` |
| `ErrUnauthorized` | Server rejected session (401, expired) |
| `ErrSessionSuspended` | 403 with `logout_reason: 8` — temporary device-fingerprint anomaly. Wait 15-30 min and retry. |
| `ErrForbidden` | Genuine access denial (private profile, blocked viewer) |
| `ErrNotFound` | Resource doesn't exist or was deleted |
| `ErrRateLimited` | HTTP 429 |
| `ErrUserAgentMismatch` | Backend rejected the User-Agent — use a Barcelona/Instagram UA |
| `ErrWriteAuthRequired` | Write method called on a cookie-only client |
| `ErrInvalidParams` | Caller passed empty / bad arguments |
| `ErrRequestFailed` | Generic transport / decode failure (use `errors.As` to inspect `*FailStatusError` for application-level errors) |

```go
post, err := c.CreatePost(ctx, "hi")
switch {
case errors.Is(err, threads.ErrSessionSuspended):
    time.Sleep(20 * time.Minute)
case errors.Is(err, threads.ErrWriteAuthRequired):
    log.Fatal("need Bearer client; call NewWithAuth or NewFull")
case errors.Is(err, threads.ErrRateLimited):
    backoff()
}
```

## Rate limiting

Threads uses **behavioural** rate limiting — there are no `X-RateLimit-*`
headers. After roughly 20–30 consecutive API calls, the session is parked
with `403 {"message":"login_required","logout_reason":8}` and recovers on
its own in 15–30 minutes if you back off.

The default `WithMinRequestGap` of **1.5s** keeps a single client well clear
of the limiter for sustained use. For batch scraping, increase the gap or
plug in a proxy pool via `WithProxy`. The client retries transient `5xx`
and network errors with exponential backoff but **does not retry**
`ErrSessionSuspended` — surface it to your scheduler so you can sleep.

## Endpoints covered

### Reads (cookie auth, `www.threads.com`)

| Method | Threads endpoint |
|---|---|
| `Me` | `GET /api/v1/accounts/current_user/?edit=true` |
| `GetProfile` | `GET /api/v1/users/{id}/info/` |
| `GetProfileExtended` | `GET /api/v1/users/{id}/info/?entry_point=profile&from_module=profile_page` |
| `SearchUsers` | `GET /api/v1/users/search/?q=&count=` |
| `UserThreads` | `GET /api/v1/text_feed/{id}/profile/` |
| `UserReplies` | `GET /api/v1/text_feed/{id}/profile/replies/` |
| `LikedPosts` | `GET /api/v1/text_feed/text_app_liked_feed/` |
| `SearchPosts` | `GET /api/v1/text_feed/text_search/?q=` |
| `GetThread` / `GetThreadReplies` | `GET /api/v1/text_feed/{post_id}/replies/` |
| `GetLikers` | `GET /api/v1/text_feed/{post_id}/likers/` |
| `GetReposters` | `GET /api/v1/text_feed/{post_id}/reposters/` |
| `GetQuoters` | `GET /api/v1/text_feed/{post_id}/quoters/` |
| `GetFollowers` | `GET /api/v1/friendships/{id}/followers/` |
| `GetFollowing` | `GET /api/v1/friendships/{id}/following/` |
| `GetFriendship` | `GET /api/v1/friendships/show/{id}/` |
| `GetFriendships` | `POST /api/v1/friendships/show_many/` |
| `PendingRequests` | `GET /api/v1/friendships/pending/` |
| `SearchHashtags` | `GET /api/v1/text_app/tags/search/` |
| `GetHashtag` | `GET /api/v1/text_app/tags/{name}/feed/` |

### Reads (Bearer auth, `i.instagram.com`)

| Method | Endpoint |
|---|---|
| `HomeTimeline` | `GET /api/v1/feed/text_post_app_timeline/` |
| `Notifications` | `GET /api/v1/text_feed/notifications/` |
| `RecommendedUsers` | `GET /api/v1/text_feed/recommended_users/` |

### Writes (Bearer auth, `i.instagram.com`)

| Method | Endpoint |
|---|---|
| `Login` | `POST /api/v1/bloks/apps/com.bloks.www.bloks.caa.login.async.send_login_request/` |
| `CreatePost` (text) | `POST /api/v1/media/configure_text_only_post/` |
| `CreatePost` (single image) | `POST /api/v1/media/configure_text_post_app_feed/` |
| `CreatePost` (carousel) | `POST /api/v1/media/configure_text_post_app_carousel/` |
| `UploadImage` | `POST /rupload_igphoto/{name}` |
| `Like` | `POST /api/v1/media/{id}/like/` |
| `Unlike` | `POST /api/v1/media/{id}/unlike/` |
| `Repost` | `POST /api/v1/repost/create_repost/` |
| `DeleteRepost` | `POST /api/v1/repost/delete_text_app_repost/` |
| `DeletePost` | `POST /api/v1/media/{id}/delete/?media_type=TEXT_POST` |
| `Follow` | `POST /api/v1/friendships/create/{id}/` |
| `Unfollow` | `POST /api/v1/friendships/destroy/{id}/` |
| `Block` | `POST /api/v1/friendships/block/{id}/` |
| `Unblock` | `POST /api/v1/friendships/unblock/{id}/` |
| `Mute` | `POST /api/v1/friendships/mute_posts_or_story_from_follow/` |
| `Unmute` | `POST /api/v1/friendships/unmute_posts_or_story_from_follow/` |
| `Restrict` | `POST /api/v1/restrict_action/restrict_many/` |
| `Unrestrict` | `POST /api/v1/restrict_action/unrestrict/` |

All Bearer write endpoints take a `signed_body` envelope of the form
`SIGNATURE.{url-encoded JSON}` — the literal string `SIGNATURE` is used
as the signing prefix; no real cryptographic signing is required.

## Examples

Runnable examples in `examples/`:

```bash
go run ./examples/get_profile          # fetch self + @zuck profile
go run ./examples/get_threads          # paginate @zuck's posts
go run ./examples/search_users -q golang -n 20
```

## Integration tests

Read-only, against the live API, gated behind a build tag:

```bash
export THREADS_SESSIONID=...
export THREADS_CSRFTOKEN=...
export THREADS_DS_USER_ID=...
export THREADS_MID=...
export THREADS_IG_DID=...
go test -tags=integration -v ./...
```

The tests are paced at 2.5s between requests to stay under the limiter.

## Out of scope (V1)

- DMs / inbox messaging
- Insights & analytics endpoints
- Story posting (Threads doesn't surface stories yet)
- Live video and audio rooms
- Edit a published post (no public endpoint)
- Public-data API (the official Threads Graph API — different surface)

## License

MIT (matches the rest of the `teslashibe/*` SDKs).
