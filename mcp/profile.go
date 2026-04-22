package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	threads "github.com/teslashibe/threads-go"
)

// GetProfileInput is the typed input for threads_get_profile.
type GetProfileInput struct {
	UserID string `json:"user_id" jsonschema:"description=numeric Threads user ID,required"`
}

func getProfile(ctx context.Context, c *threads.Client, in GetProfileInput) (any, error) {
	return c.GetProfile(ctx, in.UserID)
}

// GetProfileExtendedInput is the typed input for threads_get_profile_extended.
type GetProfileExtendedInput struct {
	UserID string `json:"user_id" jsonschema:"description=numeric Threads user ID,required"`
}

func getProfileExtended(ctx context.Context, c *threads.Client, in GetProfileExtendedInput) (any, error) {
	return c.GetProfileExtended(ctx, in.UserID)
}

// GetProfileByUsernameInput is the typed input for threads_get_profile_by_username.
type GetProfileByUsernameInput struct {
	Username string `json:"username" jsonschema:"description=Threads handle without the leading @,required"`
}

func getProfileByUsername(ctx context.Context, c *threads.Client, in GetProfileByUsernameInput) (any, error) {
	return c.GetProfileByUsername(ctx, in.Username)
}

// MeInput is the typed input for threads_me.
type MeInput struct{}

func me(ctx context.Context, c *threads.Client, _ MeInput) (any, error) {
	return c.Me(ctx)
}

var profileTools = []mcptool.Tool{
	mcptool.Define[*threads.Client, GetProfileInput](
		"threads_get_profile",
		"Fetch a Threads user profile by numeric user ID",
		"GetProfile",
		getProfile,
	),
	mcptool.Define[*threads.Client, GetProfileExtendedInput](
		"threads_get_profile_extended",
		"Fetch a Threads user profile with extended fields (links, badges, counts)",
		"GetProfileExtended",
		getProfileExtended,
	),
	mcptool.Define[*threads.Client, GetProfileByUsernameInput](
		"threads_get_profile_by_username",
		"Resolve a Threads username to a profile (use to obtain a numeric user_id)",
		"GetProfileByUsername",
		getProfileByUsername,
	),
	mcptool.Define[*threads.Client, MeInput](
		"threads_me",
		"Return the authenticated viewer's own profile",
		"Me",
		me,
	),
}
