package widget

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

var (
	dimStyle    = lipgloss.NewStyle().Faint(true)
	greenStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	yellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	redStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

// Context renders a filled/empty progress bar representing context window usage,
// followed by a value label. The bar color shifts from green to yellow at 70%
// and from yellow to red at 85%.
//
// The label format is controlled by cfg.Context.Value:
//   - "percent" (default): "42%"
//   - "tokens": "84k/200k"
//   - "remaining": "116k left"
//
// When context exceeds 85% and cfg.Context.ShowBreakdown is true, a token
// breakdown is appended: " in:84k cr:12k rd:8k".
//
// Returns "" when both ContextPercent and ContextWindowSize are zero.
func Context(ctx *model.RenderContext, cfg *config.Config) string {
	if ctx.ContextPercent == 0 && ctx.ContextWindowSize == 0 {
		return ""
	}

	barWidth := cfg.Context.BarWidth
	if barWidth <= 0 {
		barWidth = 10
	}

	pct := ctx.ContextPercent

	// Select color based on usage thresholds.
	colorStyle := greenStyle
	if pct >= 85 {
		colorStyle = redStyle
	} else if pct >= 70 {
		colorStyle = yellowStyle
	}

	filled := (pct * barWidth) / 100
	if filled > barWidth {
		filled = barWidth
	}
	empty := barWidth - filled

	bar := colorStyle.Render(strings.Repeat("█", filled)) +
		dimStyle.Render(strings.Repeat("░", empty))

	// Compute token totals used by both "tokens" and "remaining" modes.
	used := ctx.InputTokens + ctx.CacheCreation + ctx.CacheRead
	total := ctx.ContextWindowSize

	// Build the value label based on the configured mode.
	var label string
	switch cfg.Context.Value {
	case "tokens":
		label = fmt.Sprintf("%s/%s", formatTokenCount(used), formatTokenCount(total))
	case "remaining":
		remaining := total - used
		label = fmt.Sprintf("%s left", formatTokenCount(remaining))
	default: // "percent" or empty
		label = fmt.Sprintf("%d%%", pct)
	}

	result := bar + " " + colorStyle.Render(label)

	// Append token breakdown when context is high and breakdown is enabled.
	if pct > 85 && cfg.Context.ShowBreakdown {
		breakdown := fmt.Sprintf(" in:%s cr:%s rd:%s",
			formatTokenCount(ctx.InputTokens),
			formatTokenCount(ctx.CacheCreation),
			formatTokenCount(ctx.CacheRead),
		)
		result += dimStyle.Render(breakdown)
	}

	return result
}

// formatTokenCount formats a token count into a compact human-readable string:
//   - < 1000: "123"
//   - < 100000: "12.3k" (one decimal place)
//   - >= 100000: "123k" (no decimal)
func formatTokenCount(n int) string {
	switch {
	case n < 1000:
		return fmt.Sprintf("%d", n)
	case n < 100000:
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	default:
		return fmt.Sprintf("%dk", n/1000)
	}
}
