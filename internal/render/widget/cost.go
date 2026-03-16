package widget

import (
	"fmt"

	"charm.land/lipgloss/v2"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

var costStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("71")) // muted green

// Cost renders the total session cost in USD.
// Format: "$1.23" — dollar sign followed by amount with 2 decimal places.
// Returns "" when cost is zero or unavailable.
func Cost(ctx *model.RenderContext, cfg *config.Config) string {
	if ctx.SessionCostUSD == 0 {
		return ""
	}
	return costStyle.Render(fmt.Sprintf("$%.2f", ctx.SessionCostUSD))
}
