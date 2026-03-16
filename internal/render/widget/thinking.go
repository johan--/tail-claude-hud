package widget

import (
	"fmt"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

// Thinking renders a peripheral signal for active or completed thinking blocks.
//
// When ThinkingActive is true it shows the thinking icon in yellow — a live
// signal that Claude is currently reasoning. When thinking has completed
// (ThinkingCount > 0 but not active) it shows the icon in dim with the total
// count, giving a quick audit trail. Returns "" when no thinking has occurred.
func Thinking(ctx *model.RenderContext, cfg *config.Config) string {
	if ctx.Transcript == nil {
		return ""
	}

	icons := IconsFor(cfg.Style.Icons)

	if ctx.Transcript.ThinkingActive {
		return yellowStyle.Render(icons.Thinking)
	}

	if ctx.Transcript.ThinkingCount > 0 {
		return dimStyle.Render(fmt.Sprintf("%s%d", icons.Thinking, ctx.Transcript.ThinkingCount))
	}

	return ""
}
