package widget

import (
	"fmt"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

// Messages renders the number of conversational turns in the current session.
// Tool_result entries are excluded because they carry tool output back to the
// model rather than representing a human or assistant turn.
// Returns "" when ctx.Transcript is nil or no turns have been counted yet.
func Messages(ctx *model.RenderContext, cfg *config.Config) string {
	if ctx.Transcript == nil || ctx.Transcript.MessageCount == 0 {
		return ""
	}
	return dimStyle.Render(fmt.Sprintf("%d msgs", ctx.Transcript.MessageCount))
}
