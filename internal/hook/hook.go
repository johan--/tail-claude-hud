// Package hook handles Claude Code hook events dispatched via the CLI
// subcommand "tail-claude-hud hook <event>".
//
// Each handler reads JSON from stdin (the hook payload), performs a breadcrumb
// operation, and exits. Handlers always succeed (exit 0) because a hook failure
// would block Claude Code.
package hook

import (
	"encoding/json"
	"io"
	"path/filepath"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/breadcrumb"
)

// payload is the common subset of fields from Claude Code's hook stdin JSON.
type payload struct {
	SessionID string `json:"session_id"`
	CWD       string `json:"cwd"`
	ToolName  string `json:"tool_name"`
}

// HandlePermissionRequest reads the hook payload and writes a breadcrumb
// indicating this session is waiting for permission approval.
func HandlePermissionRequest(r io.Reader) error {
	var p payload
	if err := json.NewDecoder(r).Decode(&p); err != nil {
		return err
	}
	if p.SessionID == "" {
		return nil // no session ID — can't write a meaningful breadcrumb
	}

	project := filepath.Base(p.CWD)
	if project == "." || project == "/" {
		project = ""
	}

	return breadcrumb.Write(breadcrumb.Breadcrumb{
		SessionID: p.SessionID,
		Project:   project,
		ToolName:  p.ToolName,
	})
}

// HandleCleanup reads the hook payload and removes any breadcrumb for the
// session. Called by both PostToolUse and Stop hooks. Removing a breadcrumb
// that doesn't exist is a no-op.
func HandleCleanup(r io.Reader) error {
	var p payload
	if err := json.NewDecoder(r).Decode(&p); err != nil {
		return err
	}
	if p.SessionID == "" {
		return nil
	}
	return breadcrumb.Remove(p.SessionID)
}
