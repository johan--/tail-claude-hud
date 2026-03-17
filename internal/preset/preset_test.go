package preset_test

import (
	"testing"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/preset"
)

func TestLoadValidPreset(t *testing.T) {
	p, ok := preset.Load("default")
	if !ok {
		t.Fatal("Load(\"default\") returned false, want true")
	}
	if p.Name != "default" {
		t.Errorf("Name = %q, want %q", p.Name, "default")
	}
}

func TestLoadUnknownPreset(t *testing.T) {
	p, ok := preset.Load("nonexistent")
	if ok {
		t.Fatal("Load(\"nonexistent\") returned true, want false")
	}
	if p.Name != "" || p.Lines != nil {
		t.Errorf("Load(\"nonexistent\") returned non-zero Preset: %+v", p)
	}
}

func TestBuiltinNamesReturnsSixSorted(t *testing.T) {
	names := preset.BuiltinNames()
	if len(names) != 6 {
		t.Errorf("BuiltinNames() returned %d names, want 6: %v", len(names), names)
	}
	for i := 1; i < len(names); i++ {
		if names[i] <= names[i-1] {
			t.Errorf("BuiltinNames() not sorted at index %d: %v", i, names)
		}
	}
}

func TestAllBuiltinsLoadSuccessfully(t *testing.T) {
	for _, name := range preset.BuiltinNames() {
		t.Run(name, func(t *testing.T) {
			p, ok := preset.Load(name)
			if !ok {
				t.Fatalf("Load(%q) returned false", name)
			}
			if p.Name == "" {
				t.Errorf("preset %q has empty Name", name)
			}
			if len(p.Lines) == 0 {
				t.Errorf("preset %q has no Lines", name)
			}
			for i, line := range p.Lines {
				if len(line.Widgets) == 0 {
					t.Errorf("preset %q line %d has no widgets", name, i)
				}
			}
		})
	}
}

func TestDefaultPresetMatchesConfigDefaults(t *testing.T) {
	p, ok := preset.Load("default")
	if !ok {
		t.Fatal("Load(\"default\") returned false")
	}

	// Spec: 3 lines matching config.defaults() layout
	if len(p.Lines) != 3 {
		t.Fatalf("default preset has %d lines, want 3", len(p.Lines))
	}

	wantLine0 := []string{"thinking", "model", "context", "project", "todos", "duration"}
	wantLine1 := []string{"agents"}
	wantLine2 := []string{"tools"}

	assertWidgets(t, "line 0", p.Lines[0].Widgets, wantLine0)
	assertWidgets(t, "line 1", p.Lines[1].Widgets, wantLine1)
	assertWidgets(t, "line 2", p.Lines[2].Widgets, wantLine2)
}

func assertWidgets(t *testing.T, label string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: got %v (len %d), want %v (len %d)", label, got, len(got), want, len(want))
		return
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s widget[%d]: got %q, want %q", label, i, got[i], want[i])
		}
	}
}
