package widget

import (
	"strings"
	"testing"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

func defaultCfg() *config.Config {
	return config.LoadHud()
}

func TestModelWidget_DisplaysNameInBrackets(t *testing.T) {
	ctx := &model.RenderContext{ModelDisplayName: "Opus", ContextWindowSize: 200000}
	cfg := defaultCfg()

	got := Model(ctx, cfg)
	if !strings.Contains(got, "Opus") {
		t.Errorf("expected output to contain 'Opus', got %q", got)
	}
	if !strings.Contains(got, "200k context") {
		t.Errorf("expected context size '200k context', got %q", got)
	}
}

func TestModelWidget_HidesContextSize(t *testing.T) {
	ctx := &model.RenderContext{ModelDisplayName: "Sonnet", ContextWindowSize: 200000}
	cfg := defaultCfg()
	cfg.Model.ShowContextSize = false

	got := Model(ctx, cfg)
	if strings.Contains(got, "context") {
		t.Errorf("expected no context size when disabled, got %q", got)
	}
	if !strings.Contains(got, "Sonnet") {
		t.Errorf("expected 'Sonnet' in output, got %q", got)
	}
}

func TestModelWidget_EmptyName(t *testing.T) {
	ctx := &model.RenderContext{}
	cfg := defaultCfg()

	if got := Model(ctx, cfg); got != "" {
		t.Errorf("expected empty string for empty model name, got %q", got)
	}
}

func TestContextWidget_GreenUnder70(t *testing.T) {
	ctx := &model.RenderContext{ContextPercent: 50, ContextWindowSize: 200000}
	cfg := defaultCfg()

	got := Context(ctx, cfg)
	if !strings.Contains(got, "50%") {
		t.Errorf("expected '50%%' in output, got %q", got)
	}
}

func TestContextWidget_YellowAt70(t *testing.T) {
	ctx := &model.RenderContext{ContextPercent: 75, ContextWindowSize: 200000}
	cfg := defaultCfg()

	got := Context(ctx, cfg)
	if !strings.Contains(got, "75%") {
		t.Errorf("expected '75%%' in output, got %q", got)
	}
}

func TestContextWidget_RedAt85(t *testing.T) {
	ctx := &model.RenderContext{ContextPercent: 90, ContextWindowSize: 200000}
	cfg := defaultCfg()

	got := Context(ctx, cfg)
	if !strings.Contains(got, "90%") {
		t.Errorf("expected '90%%' in output, got %q", got)
	}
}

func TestContextWidget_EmptyWhenZero(t *testing.T) {
	ctx := &model.RenderContext{}
	cfg := defaultCfg()

	if got := Context(ctx, cfg); got != "" {
		t.Errorf("expected empty string for zero context, got %q", got)
	}
}

func TestDirectoryWidget_SingleSegment(t *testing.T) {
	ctx := &model.RenderContext{Cwd: "/Users/kyle/Code/my-projects/tail-claude-hud"}
	cfg := defaultCfg()
	cfg.Directory.Levels = 1

	got := Directory(ctx, cfg)
	if !strings.Contains(got, "tail-claude-hud") {
		t.Errorf("expected 'tail-claude-hud', got %q", got)
	}
}

func TestDirectoryWidget_MultipleSegments(t *testing.T) {
	ctx := &model.RenderContext{Cwd: "/Users/kyle/Code/my-projects/tail-claude-hud"}
	cfg := defaultCfg()
	cfg.Directory.Levels = 2

	got := Directory(ctx, cfg)
	if !strings.Contains(got, "my-projects/tail-claude-hud") {
		t.Errorf("expected 2 segments, got %q", got)
	}
}

func TestDirectoryWidget_EmptyCwd(t *testing.T) {
	ctx := &model.RenderContext{}
	cfg := defaultCfg()

	if got := Directory(ctx, cfg); got != "" {
		t.Errorf("expected empty string for empty cwd, got %q", got)
	}
}

func TestRegistryHasAllWidgets(t *testing.T) {
	expected := []string{"model", "context", "directory", "git", "env", "duration", "usage", "tools", "agents", "todos"}
	for _, name := range expected {
		if _, ok := Registry[name]; !ok {
			t.Errorf("Registry missing widget %q", name)
		}
	}
	if len(Registry) != len(expected) {
		t.Errorf("Registry has %d entries, expected %d", len(Registry), len(expected))
	}
}

func TestPlaceholderReturnsEmpty(t *testing.T) {
	ctx := &model.RenderContext{}
	cfg := defaultCfg()

	placeholders := []string{"git", "env", "duration", "usage", "tools", "agents", "todos"}
	for _, name := range placeholders {
		fn := Registry[name]
		if got := fn(ctx, cfg); got != "" {
			t.Errorf("placeholder widget %q returned %q, expected empty", name, got)
		}
	}
}

func TestIconsFor_Modes(t *testing.T) {
	tests := []struct {
		mode      string
		wantCheck string
	}{
		{"unicode", "✓"},
		{"ascii", "v"},
	}
	for _, tt := range tests {
		icons := IconsFor(tt.mode)
		if icons.Check != tt.wantCheck {
			t.Errorf("IconsFor(%q).Check = %q, want %q", tt.mode, icons.Check, tt.wantCheck)
		}
	}

	// Nerdfont should return non-empty
	nf := IconsFor("nerdfont")
	if nf.Check == "" {
		t.Error("nerdfont Check icon is empty")
	}

	// Unknown mode falls back to ascii
	unk := IconsFor("unknown")
	if unk.Check != "v" {
		t.Errorf("unknown mode should fall back to ascii, got Check=%q", unk.Check)
	}
}

func TestLastNSegments(t *testing.T) {
	tests := []struct {
		path string
		n    int
		want string
	}{
		{"/Users/kyle/Code", 1, "Code"},
		{"/Users/kyle/Code", 2, "kyle/Code"},
		{"/Users/kyle/Code", 5, "Users/kyle/Code"},
		{"relative/path", 1, "path"},
		{"/trailing/slash/", 1, "slash"},
		{"", 1, ""},
		{"/", 1, ""},
	}

	for _, tt := range tests {
		got := lastNSegments(tt.path, tt.n)
		if got != tt.want {
			t.Errorf("lastNSegments(%q, %d) = %q, want %q", tt.path, tt.n, got, tt.want)
		}
	}
}
