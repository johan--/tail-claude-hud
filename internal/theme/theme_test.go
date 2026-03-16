package theme

import (
	"testing"
)

func TestLoad_knownTheme(t *testing.T) {
	for _, name := range BuiltinNames() {
		t.Run(name, func(t *testing.T) {
			th := Load(name)
			if th == nil {
				t.Fatalf("Load(%q) returned nil", name)
			}
			// Every built-in theme must have at least a model and context entry.
			if _, ok := th["model"]; !ok {
				t.Errorf("theme %q missing 'model' entry", name)
			}
			if _, ok := th["context"]; !ok {
				t.Errorf("theme %q missing 'context' entry", name)
			}
		})
	}
}

func TestLoad_unknownFallsBackToDefault(t *testing.T) {
	th := Load("nonexistent-theme")
	def := Load("default")

	if len(th) != len(def) {
		t.Errorf("fallback theme len=%d, want %d (default)", len(th), len(def))
	}

	// Spot-check a key entry.
	if th["model"] != def["model"] {
		t.Errorf("fallback model colors %+v, want %+v", th["model"], def["model"])
	}
}

func TestLoad_emptyNameFallsBackToDefault(t *testing.T) {
	th := Load("")
	def := Load("default")

	if th["context"] != def["context"] {
		t.Errorf("empty name: context colors %+v, want %+v", th["context"], def["context"])
	}
}

func TestMergeOverrides_appliesOverrides(t *testing.T) {
	base := Theme{
		"model":   {Fg: "#ffffff", Bg: "#000000"},
		"context": {Fg: "#aaaaaa", Bg: "#111111"},
		"git":     {Fg: "#bbbbbb", Bg: "#222222"},
	}

	overrides := map[string]WidgetColors{
		"model": {Fg: "#ff0000", Bg: "#0000ff"},
	}

	merged := MergeOverrides(base, overrides)

	if merged["model"].Fg != "#ff0000" {
		t.Errorf("override: model Fg = %q, want %q", merged["model"].Fg, "#ff0000")
	}
	if merged["model"].Bg != "#0000ff" {
		t.Errorf("override: model Bg = %q, want %q", merged["model"].Bg, "#0000ff")
	}

	// Non-overridden entries must be unchanged.
	if merged["context"] != base["context"] {
		t.Errorf("non-overridden context changed: got %+v, want %+v", merged["context"], base["context"])
	}
	if merged["git"] != base["git"] {
		t.Errorf("non-overridden git changed: got %+v, want %+v", merged["git"], base["git"])
	}
}

func TestMergeOverrides_addsNewWidgetEntry(t *testing.T) {
	base := Theme{
		"model": {Fg: "#ffffff", Bg: "#000000"},
	}
	overrides := map[string]WidgetColors{
		"custom-widget": {Fg: "#123456", Bg: "#abcdef"},
	}

	merged := MergeOverrides(base, overrides)
	if merged["custom-widget"].Fg != "#123456" {
		t.Errorf("new entry Fg = %q, want %q", merged["custom-widget"].Fg, "#123456")
	}
}

func TestMergeOverrides_doesNotMutateBase(t *testing.T) {
	base := Theme{
		"model": {Fg: "#ffffff", Bg: "#000000"},
	}
	overrides := map[string]WidgetColors{
		"model": {Fg: "#ff0000", Bg: "#0000ff"},
	}

	_ = MergeOverrides(base, overrides)

	// base must be unchanged after merge.
	if base["model"].Fg != "#ffffff" {
		t.Errorf("base mutated: model Fg = %q, want %q", base["model"].Fg, "#ffffff")
	}
}

func TestMergeOverrides_emptyOverrides(t *testing.T) {
	base := Theme{
		"model": {Fg: "#ffffff", Bg: "#000000"},
	}

	merged := MergeOverrides(base, nil)

	if merged["model"] != base["model"] {
		t.Errorf("empty overrides changed result: got %+v, want %+v", merged["model"], base["model"])
	}
}

func TestBuiltinNames_complete(t *testing.T) {
	names := BuiltinNames()
	if len(names) < 6 {
		t.Errorf("expected at least 6 built-in themes, got %d", len(names))
	}

	expected := []string{"default", "dark", "gruvbox", "nord", "rose-pine", "tokyo-night"}
	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}
	for _, want := range expected {
		if !nameSet[want] {
			t.Errorf("expected built-in theme %q not found in BuiltinNames()", want)
		}
	}
}

func TestAllBuiltinThemesHaveAllWidgets(t *testing.T) {
	widgets := []string{"model", "context", "directory", "git", "project", "env",
		"duration", "tools", "agents", "todos", "session", "thinking"}

	for _, name := range BuiltinNames() {
		th := Load(name)
		for _, w := range widgets {
			if _, ok := th[w]; !ok {
				t.Errorf("theme %q missing widget entry %q", name, w)
			}
		}
	}
}
