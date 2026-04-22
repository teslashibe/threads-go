// Package mcp exposes the threads-go [threads.Client] surface as a set of
// MCP (Model Context Protocol) tools that any host application can mount on
// its own MCP server.
//
// All tools wrap exported methods on *threads.Client. Each tool is defined
// via [mcptool.Define] so the JSON input schema is reflected from the typed
// input struct — no hand-maintained schemas, no drift.
//
// Usage from a host application:
//
//	import (
//	    "github.com/teslashibe/mcptool"
//	    threads "github.com/teslashibe/threads-go"
//	    threadsmcp "github.com/teslashibe/threads-go/mcp"
//	)
//
//	client, _ := threads.NewFull(cookies, auth)
//	for _, tool := range threadsmcp.Provider{}.Tools() {
//	    // register tool with your MCP server, passing client as the client
//	    // arg when invoking
//	}
//
// The threads-go package exposes three constructors —
// [threads.New] (cookies, read-only), [threads.NewWithAuth] (Bearer auth,
// write-only), and [threads.NewFull] (both). The MCP tools defined here
// cover the union; calling a write-only tool against a read-only client
// will surface the underlying [threads.ErrWriteAuthRequired] back to the
// agent verbatim.
//
// The [Excluded] map documents methods on *Client that are intentionally
// not exposed via MCP, with a one-line reason. The coverage test in
// mcp_test.go fails if a new exported method is added without either being
// wrapped by a tool or appearing in [Excluded].
package mcp

import "github.com/teslashibe/mcptool"

// Provider implements [mcptool.Provider] for threads-go. The zero value is
// ready to use.
type Provider struct{}

// Platform returns "threads".
func (Provider) Platform() string { return "threads" }

// Tools returns every threads-go MCP tool, in registration order.
func (Provider) Tools() []mcptool.Tool {
	out := make([]mcptool.Tool, 0,
		len(searchTools)+
			len(feedTools)+
			len(profileTools)+
			len(hashtagTools)+
			len(threadTools)+
			len(followerTools)+
			len(notificationTools)+
			len(socialTools)+
			len(postTools)+
			len(actionTools),
	)
	out = append(out, searchTools...)
	out = append(out, feedTools...)
	out = append(out, profileTools...)
	out = append(out, hashtagTools...)
	out = append(out, threadTools...)
	out = append(out, followerTools...)
	out = append(out, notificationTools...)
	out = append(out, socialTools...)
	out = append(out, postTools...)
	out = append(out, actionTools...)
	return out
}
