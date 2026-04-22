package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Notifications fetches the authenticated user's inbox notifications.
// Requires Bearer auth (NewWithAuth or NewFull).
//
// The Notification.Args field carries story-type-specific JSON; consumers
// should switch on StoryType and decode accordingly.
func (c *Client) Notifications(ctx context.Context) (NotificationPage, error) {
	body, err := c.writeGET(ctx, "/api/v1/text_feed/notifications/", nil)
	if err != nil {
		return NotificationPage{}, err
	}
	var resp struct {
		NewStories []struct {
			StoryType int             `json:"story_type"`
			Args      json.RawMessage `json:"args"`
			Timestamp float64         `json:"timestamp"`
		} `json:"new_stories"`
		OldStories []struct {
			StoryType int             `json:"story_type"`
			Args      json.RawMessage `json:"args"`
			Timestamp float64         `json:"timestamp"`
		} `json:"old_stories"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return NotificationPage{}, fmt.Errorf("%w: decoding notifications: %v", ErrRequestFailed, err)
	}
	page := NotificationPage{}
	for _, s := range resp.NewStories {
		n := Notification{StoryType: s.StoryType, Args: s.Args}
		if s.Timestamp > 0 {
			n.Timestamp = time.Unix(int64(s.Timestamp), 0).UTC()
		}
		page.Notifications = append(page.Notifications, n)
	}
	for _, s := range resp.OldStories {
		n := Notification{StoryType: s.StoryType, Args: s.Args}
		if s.Timestamp > 0 {
			n.Timestamp = time.Unix(int64(s.Timestamp), 0).UTC()
		}
		page.Notifications = append(page.Notifications, n)
	}
	return page, nil
}
