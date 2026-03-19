package widget

import (
	"charm.land/lipgloss/v2"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

var permissionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

// Permission renders a red alert icon when another Claude Code session is
// waiting for permission approval. Returns an empty WidgetResult when no
// session needs attention, so the widget occupies zero space in normal operation.
func Permission(ctx *model.RenderContext, cfg *config.Config) WidgetResult {
	if !ctx.PermissionWaiting {
		return WidgetResult{}
	}

	icons := IconsFor(cfg.Style.Icons)
	icon := permissionIcon(icons)

	return WidgetResult{
		Text:      permissionStyle.Render(icon),
		PlainText: icon,
		FgColor:   "1", // red
	}
}

// permissionIcon returns the icon for the permission-waiting state.
func permissionIcon(icons Icons) string {
	switch {
	case icons.Error != "":
		// Nerdfont/unicode: use the error icon (✗ or nf cross)
		return icons.Error
	default:
		return "!"
	}
}
