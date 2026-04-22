// Example: search Threads users by query and print top hits.
//
// Required env vars: THREADS_SESSIONID, THREADS_CSRFTOKEN,
// THREADS_DS_USER_ID, THREADS_MID, THREADS_IG_DID.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	threads "github.com/teslashibe/threads-go"
)

func main() {
	q := flag.String("q", "golang", "search query")
	n := flag.Int("n", 10, "max results")
	flag.Parse()

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

	page, err := c.SearchUsers(context.Background(), *q, *n)
	if err != nil {
		log.Fatalf("SearchUsers: %v", err)
	}
	fmt.Printf("Found %d users for %q (showing %d)\n", page.NumResults, *q, len(page.Users))
	for _, u := range page.Users {
		fmt.Printf("  @%-20s  %s  (followers=%d, verified=%v)\n",
			u.Username, u.FullName, u.FollowerCount, u.IsVerified)
	}
}
