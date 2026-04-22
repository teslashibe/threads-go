// Example: fetch the authenticated user's profile and look up @zuck.
//
// Required env vars:
//   THREADS_SESSIONID, THREADS_CSRFTOKEN, THREADS_DS_USER_ID,
//   THREADS_MID, THREADS_IG_DID
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

	me, err := c.Me(ctx)
	if err != nil {
		log.Fatalf("Me: %v", err)
	}
	fmt.Printf("Logged in as @%s (%s) — followers=%d following=%d\n",
		me.Username, me.FullName, me.FollowerCount, me.FollowingCount)

	zuck, err := c.GetProfileByUsername(ctx, "zuck")
	if err != nil {
		log.Fatalf("GetProfileByUsername(zuck): %v", err)
	}
	fmt.Printf("@zuck — id=%s followers=%d verified=%v\n",
		zuck.ID, zuck.FollowerCount, zuck.IsVerified)

	if zuck.Biography != "" {
		fmt.Printf("Bio: %s\n", zuck.Biography)
	}
}
