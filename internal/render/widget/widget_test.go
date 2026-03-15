package widget_test

import (
	"strings"
	"testing"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/render/widget"
)

// newConfig returns a config with sensible test defaults.
func newConfig() *config.Config {
	cfg := &config.Config{}
	cfg.Model.ShowContextSize = true
	cfg.Context.BarWidth = 10
	cfg.Directory.Levels = 1
	return cfg
}

// ---- Model widget ----

func TestModel_BasicRender(t *testing.T) {
	ctx := &model.RenderContext{
		ModelDisplayName:  "Claude Opus 4",
		ContextWindowSize: 200_000,
	}
	cfg := newConfig()

	out := widget.Model(ctx, cfg)
	if out == "" {
		t.Fatal("expected non-empty output")
	}
	// ANSI escape sequences wrap the content; check the visible text is present.
	if !strings.Contains(out, "Claude Opus 4") {
		t.Errorf("output missing model name: %q", out)
	}
	if !strings.Contains(out, "200k context") {
		t.Errorf("output missing context size: %q", out)
	}
}

func TestModel_HideContextSize(t *testing.T) {
	ctx := &model.RenderContext{
		ModelDisplayName:  "Claude Haiku",
		ContextWindowSize: 100_000,
	}
	cfg := newConfig()
	cfg.Model.ShowContextSize = false

	out := widget.Model(ctx, cfg)
	if strings.Contains(out, "context") {
		t.Errorf("context size should be hidden but found in output: %q", out)
	}
	if !strings.Contains(out, "Claude Haiku") {
		t.Errorf("model name missing from output: %q", out)
	}
}

func TestModel_EmptyDisplayName(t *testing.T) {
	ctx := &model.RenderContext{ModelDisplayName: ""}
	cfg := newConfig()

	if got := widget.Model(ctx, cfg); got != "" {
		t.Errorf("expected empty string for empty ModelDisplayName, got %q", got)
	}
}

func TestModel_LargeContextMillions(t *testing.T) {
	ctx := &model.RenderContext{
		ModelDisplayName:  "Opus",
		ContextWindowSize: 1_000_000,
	}
	cfg := newConfig()

	out := widget.Model(ctx, cfg)
	if !strings.Contains(out, "1M context") {
		t.Errorf("expected 1M context size formatting, got %q", out)
	}
}

// ---- Context widget ----

func TestContext_GreenBelow70(t *testing.T) {
	ctx := &model.RenderContext{ContextPercent: 50, ContextWindowSize: 100_000}
	cfg := newConfig()

	out := widget.Context(ctx, cfg)
	if out == "" {
		t.Fatal("expected non-empty output for 50%")
	}
	if !strings.Contains(out, "50%") {
		t.Errorf("expected percentage in output, got %q", out)
	}
}

func TestContext_YellowAt70(t *testing.T) {
	ctx := &model.RenderContext{ContextPercent: 70, ContextWindowSize: 100_000}
	cfg := newConfig()

	out := widget.Context(ctx, cfg)
	if out == "" {
		t.Fatal("expected non-empty output for 70%")
	}
	if !strings.Contains(out, "70%") {
		t.Errorf("expected 70%% in output, got %q", out)
	}
	// Yellow = ANSI color 220; the rendered string should contain 220 somewhere.
	if !strings.Contains(out, "220") {
		t.Errorf("expected yellow (color 220) styling at 70%%, got %q", out)
	}
}

func TestContext_YellowBelow85(t *testing.T) {
	ctx := &model.RenderContext{ContextPercent: 84, ContextWindowSize: 100_000}
	cfg := newConfig()

	out := widget.Context(ctx, cfg)
	if !strings.Contains(out, "220") {
		t.Errorf("expected yellow (color 220) at 84%%, got %q", out)
	}
}

func TestContext_RedAt85(t *testing.T) {
	ctx := &model.RenderContext{ContextPercent: 85, ContextWindowSize: 100_000}
	cfg := newConfig()

	out := widget.Context(ctx, cfg)
	if !strings.Contains(out, "85%") {
		t.Errorf("expected 85%% in output, got %q", out)
	}
	// Red = ANSI color 196.
	if !strings.Contains(out, "196") {
		t.Errorf("expected red (color 196) styling at 85%%, got %q", out)
	}
}

func TestContext_RedAbove85(t *testing.T) {
	ctx := &model.RenderContext{ContextPercent: 95, ContextWindowSize: 100_000}
	cfg := newConfig()

	out := widget.Context(ctx, cfg)
	if !strings.Contains(out, "196") {
		t.Errorf("expected red (color 196) at 95%%, got %q", out)
	}
}

func TestContext_EmptyWhenZero(t *testing.T) {
	ctx := &model.RenderContext{ContextPercent: 0, ContextWindowSize: 0}
	cfg := newConfig()

	if got := widget.Context(ctx, cfg); got != "" {
		t.Errorf("expected empty string when both percent and size are zero, got %q", got)
	}
}

func TestContext_BarContainsFilled(t *testing.T) {
	ctx := &model.RenderContext{ContextPercent: 50, ContextWindowSize: 100_000}
	cfg := newConfig()

	out := widget.Context(ctx, cfg)
	// Filled blocks should appear in styled output (inside ANSI sequences).
	if !strings.Contains(out, "█") {
		t.Errorf("expected filled block characters in context bar, got %q", out)
	}
	if !strings.Contains(out, "░") {
		t.Errorf("expected empty block characters in context bar, got %q", out)
	}
}

// ---- Directory widget ----

func TestDirectory_LastOneSegment(t *testing.T) {
	ctx := &model.RenderContext{Cwd: "/home/user/projects/my-app"}
	cfg := newConfig()
	cfg.Directory.Levels = 1

	out := widget.Directory(ctx, cfg)
	if !strings.Contains(out, "my-app") {
		t.Errorf("expected last segment 'my-app', got %q", out)
	}
	if strings.Contains(out, "projects") {
		t.Errorf("should not include parent segment 'projects', got %q", out)
	}
}

func TestDirectory_LastTwoSegments(t *testing.T) {
	ctx := &model.RenderContext{Cwd: "/home/user/projects/my-app"}
	cfg := newConfig()
	cfg.Directory.Levels = 2

	out := widget.Directory(ctx, cfg)
	if !strings.Contains(out, "projects/my-app") {
		t.Errorf("expected 'projects/my-app', got %q", out)
	}
}

func TestDirectory_LevelsExceedDepth(t *testing.T) {
	ctx := &model.RenderContext{Cwd: "/a/b"}
	cfg := newConfig()
	cfg.Directory.Levels = 10

	out := widget.Directory(ctx, cfg)
	// When levels > path depth, return full path (without leading slash).
	if !strings.Contains(out, "a/b") {
		t.Errorf("expected full path 'a/b', got %q", out)
	}
}

func TestDirectory_EmptyCwd(t *testing.T) {
	ctx := &model.RenderContext{Cwd: ""}
	cfg := newConfig()

	if got := widget.Directory(ctx, cfg); got != "" {
		t.Errorf("expected empty string for empty Cwd, got %q", got)
	}
}

func TestDirectory_RootPath(t *testing.T) {
	ctx := &model.RenderContext{Cwd: "/"}
	cfg := newConfig()

	// Root path with levels=1 should return empty (no meaningful segment).
	out := widget.Directory(ctx, cfg)
	// It shouldn't panic; the exact output is implementation-defined for "/".
	_ = out
}

// ---- Registry ----

func TestRegistry_ContainsAllTenWidgets(t *testing.T) {
	expected := []string{
		"model", "context", "directory",
		"git", "env", "duration", "usage",
		"tools", "agents", "todos",
	}
	for _, name := range expected {
		if _, ok := widget.Registry[name]; !ok {
			t.Errorf("Registry missing widget %q", name)
		}
	}
}

func TestRegistry_PlaceholdersReturnEmpty(t *testing.T) {
	placeholders := []string{"git", "env", "duration", "usage", "tools", "agents", "todos"}
	ctx := &model.RenderContext{}
	cfg := newConfig()

	for _, name := range placeholders {
		fn := widget.Registry[name]
		if got := fn(ctx, cfg); got != "" {
			t.Errorf("placeholder widget %q should return empty string, got %q", name, got)
		}
	}
}
