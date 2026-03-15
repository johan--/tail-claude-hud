package widget

import (
	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

// Session renders the current session name with dim styling.
// Returns "" when ctx.Transcript is nil or SessionName is empty.
func Session(ctx *model.RenderContext, cfg *config.Config) string {
	if ctx.Transcript == nil || ctx.Transcript.SessionName == "" {
		return ""
	}
	return dimStyle.Render(ctx.Transcript.SessionName)
}
