package widget

import (
	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

// OutputStyle renders the current Claude Code output style name.
// Returns "" when ctx.OutputStyle is empty (data not present in stdin).
func OutputStyle(ctx *model.RenderContext, cfg *config.Config) string {
	if ctx.OutputStyle == "" {
		return ""
	}
	return ctx.OutputStyle
}
