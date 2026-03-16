package render

import (
	"bytes"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/render/widget"
)

func TestRender_ProducesOutput(t *testing.T) {
	ctx := &model.RenderContext{
		ModelDisplayName:  "Sonnet",
		ContextWindowSize: 200000,
		ContextPercent:    50,
		Cwd:               "/Users/kyle/Code/project",
	}
	cfg := config.LoadHud()

	var buf bytes.Buffer
	Render(&buf, ctx, cfg)

	out := buf.String()
	if out == "" {
		t.Fatal("Render produced no output")
	}

	// Default config line 1 has model, context, directory
	if !strings.Contains(out, "Sonnet") {
		t.Errorf("expected 'Sonnet' in output, got %q", out)
	}
	if !strings.Contains(out, "50%") {
		t.Errorf("expected '50%%' in output, got %q", out)
	}
	if !strings.Contains(out, "project") {
		t.Errorf("expected 'project' in output, got %q", out)
	}
}

func TestRender_SkipsUnknownWidgets(t *testing.T) {
	ctx := &model.RenderContext{ModelDisplayName: "Opus"}
	cfg := config.LoadHud()
	cfg.Lines = []config.Line{
		{Widgets: []string{"nonexistent", "model"}},
	}

	var buf bytes.Buffer
	Render(&buf, ctx, cfg)

	out := buf.String()
	if !strings.Contains(out, "Opus") {
		t.Errorf("expected 'Opus' after skipping unknown widget, got %q", out)
	}
}

func TestRender_SkipsEmptyLines(t *testing.T) {
	ctx := &model.RenderContext{} // no data -> all widgets return ""
	cfg := config.LoadHud()

	var buf bytes.Buffer
	Render(&buf, ctx, cfg)

	// With no transcript or model data all widgets return "". All three lines
	// are empty and skipped, so output should be empty.
	out := buf.String()
	if out != "" {
		t.Errorf("expected empty output when all widgets have no data, got %q", out)
	}
}

func TestRender_TruncatesLongLines(t *testing.T) {
	// Build a model name long enough that the rendered line will exceed a narrow width.
	longName := strings.Repeat("X", 100)
	ctx := &model.RenderContext{
		ModelDisplayName: longName,
		TerminalWidth:    20,
	}
	cfg := config.LoadHud()
	cfg.Lines = []config.Line{
		{Widgets: []string{"model"}},
	}

	var buf bytes.Buffer
	Render(&buf, ctx, cfg)

	line := strings.TrimRight(buf.String(), "\n")
	// Visual width must not exceed TerminalWidth.
	w := ansi.StringWidth(line)
	if w > 20 {
		t.Errorf("expected visual width <= 20, got %d for %q", w, line)
	}
	// Truncated lines must contain the suffix (ANSI reset codes may follow it).
	if !strings.Contains(line, truncateSuffix) {
		t.Errorf("expected %q in truncated line, got %q", truncateSuffix, line)
	}
}

func TestRender_NoTruncationWhenWidthZero(t *testing.T) {
	// When TerminalWidth is 0, output should not be truncated regardless of length.
	longName := strings.Repeat("Y", 200)
	ctx := &model.RenderContext{
		ModelDisplayName: longName,
		TerminalWidth:    0,
	}
	cfg := config.LoadHud()
	cfg.Lines = []config.Line{
		{Widgets: []string{"model"}},
	}

	var buf bytes.Buffer
	Render(&buf, ctx, cfg)

	line := strings.TrimRight(buf.String(), "\n")
	if !strings.Contains(line, longName) {
		t.Errorf("expected full name in output when no width limit, got %q", line)
	}
}

func TestRender_NoTruncationWhenLineShortEnough(t *testing.T) {
	ctx := &model.RenderContext{
		ModelDisplayName: "Sonnet",
		TerminalWidth:    200,
	}
	cfg := config.LoadHud()
	cfg.Lines = []config.Line{
		{Widgets: []string{"model"}},
	}

	var buf bytes.Buffer
	Render(&buf, ctx, cfg)

	line := strings.TrimRight(buf.String(), "\n")
	if strings.HasSuffix(line, truncateSuffix) {
		t.Errorf("expected no truncation for short line, got %q", line)
	}
	if !strings.Contains(line, "Sonnet") {
		t.Errorf("expected 'Sonnet' in untruncated output, got %q", line)
	}
}

func TestRender_NoTruncationBelowMinWidth(t *testing.T) {
	// When TerminalWidth is below the minimum (20), truncation is skipped
	// entirely so the HUD does not collapse to "..." in very narrow terminals.
	ctx := &model.RenderContext{
		ModelDisplayName: "claude-sonnet-4-5",
		TerminalWidth:    10, // below minTruncateWidth
	}
	cfg := config.LoadHud()
	cfg.Lines = []config.Line{
		{Widgets: []string{"model"}},
	}

	var buf bytes.Buffer
	Render(&buf, ctx, cfg)

	line := strings.TrimRight(buf.String(), "\n")
	// Line must not be just "..." — some real content must appear.
	if line == truncateSuffix {
		t.Errorf("expected real content at narrow width, got only %q", line)
	}
	// The full model name may or may not fit, but the suffix must NOT appear
	// since truncation is skipped below minTruncateWidth.
	if strings.HasSuffix(line, truncateSuffix) {
		t.Errorf("expected no truncation suffix below min width, got %q", line)
	}
}

func TestRender_AnsiResetPrefixPerLine(t *testing.T) {
	// Each rendered line must start with the ANSI reset escape so that Claude
	// Code's dim styling does not bleed into our statusline colors.
	ctx := &model.RenderContext{
		ModelDisplayName:  "Sonnet",
		ContextWindowSize: 200000,
		ContextPercent:    50,
	}
	cfg := config.LoadHud()
	cfg.Lines = []config.Line{
		{Widgets: []string{"model"}},
		{Widgets: []string{"context"}},
	}

	var buf bytes.Buffer
	Render(&buf, ctx, cfg)

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) == 0 {
		t.Fatal("Render produced no lines")
	}
	for i, line := range lines {
		if !strings.HasPrefix(line, "\x1b[0m") {
			t.Errorf("line %d does not start with ANSI reset, got %q", i, line)
		}
	}
}

func TestRender_UsesSeparator(t *testing.T) {
	ctx := &model.RenderContext{
		ModelDisplayName:  "Opus",
		ContextWindowSize: 200000,
		ContextPercent:    42,
	}
	cfg := config.LoadHud()
	cfg.Style.Separator = " :: "
	cfg.Lines = []config.Line{
		{Widgets: []string{"model", "context"}},
	}

	var buf bytes.Buffer
	Render(&buf, ctx, cfg)

	out := buf.String()
	// Spaces are replaced with NBSP in final output, so the separator
	// " :: " becomes "\u00a0::\u00a0".
	nbspSep := strings.ReplaceAll(" :: ", " ", "\u00a0")
	if !strings.Contains(out, nbspSep) {
		t.Errorf("expected NBSP separator %q in output, got %q", nbspSep, out)
	}
}

func TestRender_ReplacesSpacesWithNBSP(t *testing.T) {
	// VS Code's integrated terminal trims trailing spaces from lines, which
	// collapses padded statusline content. All regular spaces in the final
	// output are replaced with non-breaking spaces (U+00A0) to prevent this.
	ctx := &model.RenderContext{
		ModelDisplayName:  "Sonnet",
		ContextWindowSize: 200000,
		ContextPercent:    50,
	}
	cfg := config.LoadHud()
	cfg.Style.Separator = " | "
	cfg.Lines = []config.Line{
		{Widgets: []string{"model", "context"}},
	}

	var buf bytes.Buffer
	Render(&buf, ctx, cfg)

	out := buf.String()
	// No regular space must appear in any output line.
	if strings.Contains(out, " ") {
		t.Errorf("expected no regular spaces in output (should be NBSP), got %q", out)
	}
	// Non-breaking spaces must be present (the separator has spaces).
	if !strings.Contains(out, "\u00a0") {
		t.Errorf("expected NBSP (U+00A0) in output, got %q", out)
	}
}

// TestRender_PlainModeOutputIdentical verifies spec 5: plain mode output is
// identical before and after the WidgetResult restructure. Simple widgets that
// return FgColor must produce the same ANSI output as the old style.Render call.
func TestRender_PlainModeOutputIdentical(t *testing.T) {
	ctx := &model.RenderContext{
		EnvCounts: &model.EnvCounts{MCPServers: 3, Hooks: 2},
	}
	cfg := config.LoadHud()
	cfg.Lines = []config.Line{
		{Widgets: []string{"env"}},
	}

	var buf bytes.Buffer
	Render(&buf, ctx, cfg)

	rendered := strings.TrimRight(buf.String(), "\n")

	// The Env widget returns FgColor="245" with Text="3M 2H".
	// applyWidgetStyle must produce the same ANSI color codes as the old envStyle.Render.
	// The renderer also prepends ansiReset and converts spaces to NBSP, so we match that here.
	styled := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("3M 2H")
	want := strings.ReplaceAll("\x1b[0m"+styled, " ", "\u00a0")
	if rendered != want {
		t.Errorf("plain mode output mismatch: got %q, want %q", rendered, want)
	}

	// Cross-check: verify the WidgetResult fields themselves.
	result := widget.Registry["env"](ctx, cfg)
	if result.FgColor != "245" {
		t.Errorf("Env FgColor: expected '245', got %q", result.FgColor)
	}
	if result.Text != "3M 2H" {
		t.Errorf("Env Text: expected '3M 2H', got %q", result.Text)
	}
}

// --- Powerline renderer unit tests ---

// TestRenderPowerline_TwoSegmentTransition verifies that the arrow between two
// segments uses the left segment's bg as fg and the right segment's bg as bg.
func TestRenderPowerline_TwoSegmentTransition(t *testing.T) {
	segs := []widget.WidgetResult{
		{Text: "A", FgColor: "255", BgColor: "75"},
		{Text: "B", FgColor: "255", BgColor: "114"},
	}

	out := renderPowerline(segs)

	// The arrow between A→B must have:
	//   bg = right segment bg (114): \x1b[48;5;114m
	//   fg = left segment bg (75):   \x1b[38;5;75m
	if !strings.Contains(out, "\x1b[48;5;114m") {
		t.Errorf("expected bg=114 (right segment bg) in transition, got %q", out)
	}
	if !strings.Contains(out, "\x1b[38;5;75m") {
		t.Errorf("expected fg=75 (left segment bg) in transition, got %q", out)
	}
	// Arrow character must be present.
	if !strings.Contains(out, powerlineArrow) {
		t.Errorf("expected powerline arrow character in output, got %q", out)
	}
	// Start cap must be present.
	if !strings.Contains(out, powerlineStartCap) {
		t.Errorf("expected start cap character in output, got %q", out)
	}
	// Both segment texts must appear.
	if !strings.Contains(out, "A") {
		t.Errorf("expected segment text 'A' in output, got %q", out)
	}
	if !strings.Contains(out, "B") {
		t.Errorf("expected segment text 'B' in output, got %q", out)
	}
}

// TestRenderPowerline_EmptySegmentsSkipped verifies that segments with empty
// Text are not rendered and do not produce spurious arrows.
func TestRenderPowerline_EmptySegmentsSkipped(t *testing.T) {
	segs := []widget.WidgetResult{
		{Text: "A", BgColor: "75"},
		{Text: "", BgColor: "114"}, // empty — must be skipped
		{Text: "C", BgColor: "196"},
	}

	out := renderPowerline(segs)

	// "B" was never part of the input, but 114 is the bg of the empty segment.
	// The transition should jump from 75→196, so bg=114 should NOT appear for
	// any arrow (it would appear if the empty segment were mistakenly included).
	if strings.Contains(out, "\x1b[48;5;114m") {
		t.Errorf("expected empty segment (bg=114) to be skipped, got %q", out)
	}
	// A and C must still appear.
	if !strings.Contains(out, "A") {
		t.Errorf("expected 'A' in output, got %q", out)
	}
	if !strings.Contains(out, "C") {
		t.Errorf("expected 'C' in output, got %q", out)
	}
}

// TestRenderPowerline_SingleSegment verifies that a single segment renders with
// a start cap and an end cap arrow, but no mid-segment transition arrows.
func TestRenderPowerline_SingleSegment(t *testing.T) {
	segs := []widget.WidgetResult{
		{Text: "solo", FgColor: "255", BgColor: "75"},
	}

	out := renderPowerline(segs)

	if !strings.Contains(out, powerlineStartCap) {
		t.Errorf("expected start cap in single-segment output, got %q", out)
	}
	if !strings.Contains(out, powerlineArrow) {
		t.Errorf("expected end cap arrow in single-segment output, got %q", out)
	}
	if !strings.Contains(out, "solo") {
		t.Errorf("expected segment text in single-segment output, got %q", out)
	}
	// Count arrow characters: start cap (×1) + end arrow (×1) = 2 total
	// but they use different codepoints, so just verify both appear once each.
	capCount := strings.Count(out, powerlineStartCap)
	arrowCount := strings.Count(out, powerlineArrow)
	if capCount != 1 {
		t.Errorf("expected exactly 1 start cap, got %d in %q", capCount, out)
	}
	if arrowCount != 1 {
		t.Errorf("expected exactly 1 end arrow, got %d in %q", arrowCount, out)
	}
}

// TestRenderPowerline_AllEmptyReturnsEmpty verifies that all-empty results
// produce an empty string (so the line is skipped).
func TestRenderPowerline_AllEmptyReturnsEmpty(t *testing.T) {
	segs := []widget.WidgetResult{
		{Text: ""},
		{Text: ""},
	}
	out := renderPowerline(segs)
	if out != "" {
		t.Errorf("expected empty string for all-empty segments, got %q", out)
	}
}

// TestRenderPowerline_DefaultBgFallback verifies that segments without a BgColor
// use the defaultPowerlineBg constant.
func TestRenderPowerline_DefaultBgFallback(t *testing.T) {
	segs := []widget.WidgetResult{
		{Text: "X"}, // no BgColor
	}
	out := renderPowerline(segs)

	// Start cap fg should be the default bg.
	expectedStartFg := ansiSetFg(defaultPowerlineBg)
	if !strings.Contains(out, expectedStartFg) {
		t.Errorf("expected default bg %q as start cap fg color, got %q", defaultPowerlineBg, out)
	}
}

// TestRender_PowerlineMode verifies that the Render function routes to the
// powerline renderer when style.mode = "powerline".
func TestRender_PowerlineMode(t *testing.T) {
	ctx := &model.RenderContext{
		ModelDisplayName: "Sonnet",
	}
	cfg := config.LoadHud()
	cfg.Style.Mode = "powerline"
	cfg.Lines = []config.Line{
		{Widgets: []string{"model"}},
	}

	var buf bytes.Buffer
	Render(&buf, ctx, cfg)

	out := buf.String()
	// Powerline output must contain the start cap or the arrow — plain mode would not.
	if !strings.Contains(out, powerlineArrow) && !strings.Contains(out, powerlineStartCap) {
		t.Errorf("expected powerline characters in powerline mode output, got %q", out)
	}
	if !strings.Contains(out, "Sonnet") {
		t.Errorf("expected 'Sonnet' in powerline output, got %q", out)
	}
}

// TestRender_PerLinePowerlineMode verifies that a per-line mode override works
// independently of the global style.mode.
func TestRender_PerLinePowerlineMode(t *testing.T) {
	ctx := &model.RenderContext{
		ModelDisplayName:  "Opus",
		ContextWindowSize: 200000,
		ContextPercent:    42,
	}
	cfg := config.LoadHud()
	cfg.Style.Mode = "plain" // global: plain
	cfg.Style.Separator = " | "
	cfg.Lines = []config.Line{
		{Widgets: []string{"model"}, Mode: "powerline"}, // line 0: powerline override
		{Widgets: []string{"context"}},                  // line 1: inherits plain
	}

	var buf bytes.Buffer
	Render(&buf, ctx, cfg)

	out := buf.String()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected 2 output lines, got %d: %q", len(lines), out)
	}

	// Line 0 (model, powerline mode) must have powerline characters.
	if !strings.Contains(lines[0], powerlineArrow) && !strings.Contains(lines[0], powerlineStartCap) {
		t.Errorf("line 0 (powerline): expected powerline chars, got %q", lines[0])
	}
	// Line 1 (context, plain mode) must not have start cap character.
	if strings.Contains(lines[1], powerlineStartCap) {
		t.Errorf("line 1 (plain): unexpected start cap char, got %q", lines[1])
	}
}
