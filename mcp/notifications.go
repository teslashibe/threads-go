package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	threads "github.com/teslashibe/threads-go"
)

// NotificationsInput is the typed input for threads_notifications.
type NotificationsInput struct{}

func notifications(ctx context.Context, c *threads.Client, _ NotificationsInput) (any, error) {
	res, err := c.Notifications(ctx)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(res.Notifications, res.NextCursor, len(res.Notifications)), nil
}

var notificationTools = []mcptool.Tool{
	mcptool.Define[*threads.Client, NotificationsInput](
		"threads_notifications",
		"Fetch the authenticated viewer's recent inbox notifications",
		"Notifications",
		notifications,
	),
}
