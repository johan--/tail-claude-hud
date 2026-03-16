package widget

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

var envStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

// Env renders a compact summary of the active Claude Code environment.
// Each non-zero category is shown with a letter suffix:
//
//	NM = MCP servers, NC = CLAUDE.md files, NR = rule files, NH = hooks
//
// Example: "3M 2C 4R 3H". Returns "" when ctx.EnvCounts is nil or all zeros.
func Env(ctx *model.RenderContext, cfg *config.Config) string {
	if ctx.EnvCounts == nil {
		return ""
	}

	ec := ctx.EnvCounts
	var parts []string

	if ec.MCPServers > 0 {
		parts = append(parts, fmt.Sprintf("%dM", ec.MCPServers))
	}
	if ec.ClaudeMdFiles > 0 {
		parts = append(parts, fmt.Sprintf("%dC", ec.ClaudeMdFiles))
	}
	if ec.RuleFiles > 0 {
		parts = append(parts, fmt.Sprintf("%dR", ec.RuleFiles))
	}
	if ec.Hooks > 0 {
		parts = append(parts, fmt.Sprintf("%dH", ec.Hooks))
	}

	if len(parts) == 0 {
		return ""
	}

	return envStyle.Render(strings.Join(parts, " "))
}
