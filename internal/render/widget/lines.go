package widget

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

var (
	linesAddedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // ANSI bright green
	linesRemovedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // ANSI bright red
)

// Lines renders the lines added and removed during the current session.
// Format: "+N -M" with green for additions and red for removals.
// Returns an empty WidgetResult when both counts are zero or no cost data was provided.
// FgColor is left empty because the widget composes two different styles;
// the renderer passes the pre-styled Text through as-is.
func Lines(ctx *model.RenderContext, cfg *config.Config) WidgetResult {
	if ctx.LinesAdded == 0 && ctx.LinesRemoved == 0 {
		return WidgetResult{}
	}

	var parts []string

	if ctx.LinesAdded > 0 {
		parts = append(parts, linesAddedStyle.Render(fmt.Sprintf("+%d", ctx.LinesAdded)))
	}
	if ctx.LinesRemoved > 0 {
		parts = append(parts, linesRemovedStyle.Render(fmt.Sprintf("-%d", ctx.LinesRemoved)))
	}

	return WidgetResult{Text: strings.Join(parts, " ")}
}
