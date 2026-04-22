// Package threads provides a Go client for the Threads (Meta) private API.
//
// It covers reads (profiles, thread posts, replies, likers, social graph,
// hashtags, search, notifications, home/user feeds) via the cookie-based
// www.threads.com REST endpoints, and writes (post, like, follow, repost,
// delete, block, mute, restrict) via the Bearer-token i.instagram.com
// REST endpoints.
//
// No API keys, no developer-app registration, zero dependencies.
//
// Quick start (reads, cookie auth):
//
//	c, err := threads.New(threads.Cookies{
//		SessionID: os.Getenv("THREADS_SESSIONID"),
//		CSRFToken: os.Getenv("THREADS_CSRFTOKEN"),
//		DSUserID:  os.Getenv("THREADS_DS_USER_ID"),
//		Mid:       os.Getenv("THREADS_MID"),
//		IgDid:     os.Getenv("THREADS_IG_DID"),
//	})
//
//	me, err      := c.Me(ctx)
//	user, err    := c.GetProfile(ctx, "zuck")
//	page, err    := c.UserThreads(ctx, user.ID, 20, "")
//	thread, err  := c.GetThread(ctx, postID)
//	hashtag, err := c.GetHashtag(ctx, "golang")
//
// Quick start (writes, Bearer auth):
//
//	c, err := threads.NewWithAuth(threads.Auth{
//		Token:    bearerToken,    // IGT:2:... from Bloks login
//		UserID:   "75472043478",
//		DeviceID: "android-abc123def4567",
//	})
//
//	post, err := c.CreatePost(ctx, "Hello from threads-go!")
//	err = c.Like(ctx, post.ID)
//	err = c.Follow(ctx, userID)
//	err = c.Repost(ctx, post.ID)
package threads
