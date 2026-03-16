// Package render walks config lines, calls widget functions, joins non-empty
// results with the configured separator, and writes each line to an io.Writer.
package render

import (
	"fmt"
	"io"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/logging"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/render/widget"
)

// truncateSuffix is appended when a line is truncated to fit terminal width.
const truncateSuffix = "..."

// ansiReset is prepended to every output line so that our ANSI color codes
// render correctly even when Claude Code applies dim styling to plugin output.
// Without this, Claude Code's dim setting bleeds into the statusline colors.
const ansiReset = "\x1b[0m"

// minTruncateWidth is the smallest terminal width at which truncation is
// applied. Below this threshold the suffix itself would consume most of the
// available space and produce output that is less useful than the raw text.
const minTruncateWidth = 20

// Render walks config lines, looks up widgets in the registry, joins non-empty
// results with the configured separator, and writes each line to w.
//
// Unknown widget names are skipped silently (logged at Debug level).
// Lines where all widgets return empty strings are skipped entirely.
//
// When ctx.TerminalWidth is at least minTruncateWidth (20), each output line
// is truncated to that width using ANSI-aware grapheme counting so that escape
// sequences and wide characters are measured correctly. Truncated lines gain a
// "..." suffix. Below the minimum, truncation is skipped so that very narrow
// terminals still receive content rather than collapsing to "...".
//
// The caller is expected to populate ctx.TerminalWidth before calling Render
// (the gather stage does this via terminalWidth() in gather.go).
func Render(w io.Writer, ctx *model.RenderContext, cfg *config.Config) {
	sep := cfg.Style.Separator

	for _, line := range cfg.Lines {
		var parts []string
		for _, name := range line.Widgets {
			fn, ok := widget.Registry[name]
			if !ok {
				logging.Debug("render: unknown widget %q, skipping", name)
				continue
			}
			result := fn(ctx, cfg)
			if result.IsEmpty() {
				continue
			}
			s := applyWidgetStyle(result, name, cfg)
			parts = append(parts, s)
		}
		if len(parts) == 0 {
			continue // skip lines where every widget returned empty
		}

		output := strings.Join(parts, sep)

		if ctx.TerminalWidth >= minTruncateWidth {
			output = ansi.Truncate(output, ctx.TerminalWidth, truncateSuffix)
		}

		// Prepend reset so our colors override Claude Code's dim styling.
		// Then replace spaces with non-breaking spaces (U+00A0) to prevent
		// VS Code's integrated terminal from trimming trailing whitespace.
		// ANSI escape sequences do not contain spaces, so this replacement
		// is safe to apply to the full line including escape codes.
		outLine := strings.ReplaceAll(ansiReset+output, " ", "\u00a0")
		fmt.Fprintln(w, outLine)
	}
}

// applyWidgetStyle converts a WidgetResult to a styled string, incorporating
// theme colors from the resolved config theme map.
//
// Color precedence (highest to lowest):
//  1. WidgetResult.FgColor / WidgetResult.BgColor — explicit per-render override
//  2. cfg.ResolvedTheme[widgetName].Fg / .Bg — theme default for this widget
//  3. Widget's own pre-styled ANSI output (FgColor == "" and no theme bg)
//
// When FgColor is empty the Text is returned as-is (the widget pre-styled it
// internally), unless a theme BgColor applies in which case the text is wrapped
// with that background. When FgColor is set, a fresh lipgloss.Style is built
// from FgColor and the resolved BgColor (widget > theme) and applied to Text.
func applyWidgetStyle(r widget.WidgetResult, widgetName string, cfg *config.Config) string {
	// Resolve background: widget result takes precedence over theme.
	bgColor := r.BgColor
	if bgColor == "" {
		if colors, ok := cfg.ResolvedTheme[widgetName]; ok {
			bgColor = colors.Bg
		}
	}

	if r.FgColor == "" {
		// Pre-styled output: only apply bg if theme provides one.
		if bgColor == "" {
			return r.Text
		}
		return lipgloss.NewStyle().Background(lipgloss.Color(bgColor)).Render(r.Text)
	}

	// Structured output: build full style from fg + resolved bg.
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(r.FgColor))
	if bgColor != "" {
		style = style.Background(lipgloss.Color(bgColor))
	}
	return style.Render(r.Text)
}
