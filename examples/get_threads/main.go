// Example: paginate through @zuck's recent thread posts.
//
// Required env vars: THREADS_SESSIONID, THREADS_CSRFTOKEN,
// THREADS_DS_USER_ID, THREADS_MID, THREADS_IG_DID.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	threads "github.com/teslashibe/threads-go"
)

func main() {
	c, err := threads.New(threads.Cookies{
		SessionID: os.Getenv("THREADS_SESSIONID"),
		CSRFToken: os.Getenv("THREADS_CSRFTOKEN"),
		DSUserID:  os.Getenv("THREADS_DS_USER_ID"),
		Mid:       os.Getenv("THREADS_MID"),
		IgDid:     os.Getenv("THREADS_IG_DID"),
	})
	if err != nil {
		log.Fatalf("New: %v", err)
	}

	ctx := context.Background()
	zuck, err := c.GetProfileByUsername(ctx, "zuck")
	if err != nil {
		log.Fatalf("lookup: %v", err)
	}
	fmt.Printf("Fetching threads for @%s (id=%s)\n", zuck.Username, zuck.ID)

	it := threads.NewUserThreadsIterator(c, zuck.ID, 25, threads.WithMaxPosts(50))
	for it.Next(ctx) {
		for _, p := range it.Page() {
			text := p.Text
			if len(text) > 80 {
				text = text[:80] + "…"
			}
			fmt.Printf("[%s] likes=%d replies=%d reposts=%d  %s\n",
				p.TakenAt.Format("2006-01-02 15:04"), p.LikeCount, p.ReplyCount, p.RepostCount, text)
		}
	}
	if err := it.Err(); err != nil {
		log.Fatalf("iterator: %v", err)
	}
	fmt.Printf("done — %d posts seen\n", it.Seen())
}
