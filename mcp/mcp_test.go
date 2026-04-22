package mcp_test

import (
	"reflect"
	"testing"

	"github.com/teslashibe/mcptool"
	threads "github.com/teslashibe/threads-go"
	threadsmcp "github.com/teslashibe/threads-go/mcp"
)

// TestEveryClientMethodIsWrappedOrExcluded fails when a new exported method
// is added to *threads.Client without either being wrapped by an MCP tool
// or being added to threadsmcp.Excluded with a reason. This is the drift-
// prevention mechanism: keeping the MCP surface in lockstep with the
// package API is enforced by CI rather than convention.
func TestEveryClientMethodIsWrappedOrExcluded(t *testing.T) {
	rep := mcptool.Coverage(
		reflect.TypeOf(&threads.Client{}),
		threadsmcp.Provider{}.Tools(),
		threadsmcp.Excluded,
	)
	if len(rep.Missing) > 0 {
		t.Fatalf("methods missing MCP exposure (add a tool or list in excluded.go): %v", rep.Missing)
	}
	if len(rep.UnknownExclusions) > 0 {
		t.Fatalf("excluded.go references methods that don't exist on *Client (rename?): %v", rep.UnknownExclusions)
	}
	if len(rep.Wrapped)+len(rep.Excluded) == 0 {
		t.Fatal("no wrapped or excluded methods detected — coverage helper is mis-configured")
	}
}

// TestToolsValidate verifies every tool has a non-empty name in canonical
// snake_case form, a description within length limits, and a non-nil
// Invoke + InputSchema.
func TestToolsValidate(t *testing.T) {
	if err := mcptool.ValidateTools(threadsmcp.Provider{}.Tools()); err != nil {
		t.Fatal(err)
	}
}

// TestPlatformName guards against accidental rebrands.
func TestPlatformName(t *testing.T) {
	if got := (threadsmcp.Provider{}).Platform(); got != "threads" {
		t.Errorf("Platform() = %q, want threads", got)
	}
}

// TestToolsHaveThreadsPrefix encodes the per-platform naming convention.
func TestToolsHaveThreadsPrefix(t *testing.T) {
	for _, tool := range (threadsmcp.Provider{}).Tools() {
		if len(tool.Name) < len("threads_") || tool.Name[:len("threads_")] != "threads_" {
			t.Errorf("tool %q lacks threads_ prefix", tool.Name)
		}
	}
}
