package preset

import (
	"os"
	"path/filepath"
	"testing"
)

// writePreset creates a .toml preset file in dir with the given content and
// returns its full path.
func writePreset(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name+".toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writePreset: %v", err)
	}
	return path
}

func TestLoadFromFile_BasicFields(t *testing.T) {
	dir := t.TempDir()
	content := `
name = "my-preset"

[[line]]
widgets = ["model", "context", "duration"]

[style]
separator = " | "
icons = "nerdfont"
mode = "plain"
theme = "default"

[directory]
style = "fish"
`
	path := writePreset(t, dir, "my-preset", content)

	p, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile error: %v", err)
	}

	if p.Name != "my-preset" {
		t.Errorf("Name: want %q, got %q", "my-preset", p.Name)
	}
	if len(p.Lines) != 1 {
		t.Fatalf("Lines: want 1, got %d", len(p.Lines))
	}
	if len(p.Lines[0].Widgets) != 3 {
		t.Errorf("Lines[0].Widgets: want 3, got %d", len(p.Lines[0].Widgets))
	}
	if p.Separator != " | " {
		t.Errorf("Separator: want %q, got %q", " | ", p.Separator)
	}
	if p.Icons != "nerdfont" {
		t.Errorf("Icons: want %q, got %q", "nerdfont", p.Icons)
	}
	if p.Mode != "plain" {
		t.Errorf("Mode: want %q, got %q", "plain", p.Mode)
	}
	if p.Theme != "default" {
		t.Errorf("Theme: want %q, got %q", "default", p.Theme)
	}
	if p.DirectoryStyle != "fish" {
		t.Errorf("DirectoryStyle: want %q, got %q", "fish", p.DirectoryStyle)
	}
}

func TestLoadFromFile_NameDefaultsToFilename(t *testing.T) {
	dir := t.TempDir()
	// No name field in the TOML — should default to filename without extension.
	content := `
[[line]]
widgets = ["model"]
`
	path := writePreset(t, dir, "compact", content)

	p, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile error: %v", err)
	}
	if p.Name != "compact" {
		t.Errorf("Name: want %q, got %q", "compact", p.Name)
	}
}

func TestLoadFromFile_MissingFileReturnsError(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/does-not-exist.toml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestListCustom_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// Create an empty presets dir.
	presetsDir := filepath.Join(dir, ".config", "tail-claude-hud", "presets")
	if err := os.MkdirAll(presetsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	names := ListCustom()
	if len(names) != 0 {
		t.Errorf("want empty slice, got %v", names)
	}
}

func TestListCustom_DirectoryDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	// Do NOT create the presets directory.

	names := ListCustom()
	if names == nil {
		t.Error("want empty slice, got nil")
	}
	if len(names) != 0 {
		t.Errorf("want empty slice, got %v", names)
	}
}

func TestListCustom_WithFiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	presetsDir := filepath.Join(dir, ".config", "tail-claude-hud", "presets")
	if err := os.MkdirAll(presetsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Write three preset files in non-alphabetical order.
	for _, name := range []string{"zebra", "alpha", "beta"} {
		writePreset(t, presetsDir, name, `[[line]]`+"\n"+"widgets = []\n")
	}
	// Write a non-.toml file that should be ignored.
	if err := os.WriteFile(filepath.Join(presetsDir, "readme.txt"), []byte("ignore me"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	names := ListCustom()
	want := []string{"alpha", "beta", "zebra"}
	if len(names) != len(want) {
		t.Fatalf("want %v, got %v", want, names)
	}
	for i, w := range want {
		if names[i] != w {
			t.Errorf("names[%d]: want %q, got %q", i, w, names[i])
		}
	}
}

func TestListAll_BuiltinsBeforeCustom(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	presetsDir := filepath.Join(dir, ".config", "tail-claude-hud", "presets")
	if err := os.MkdirAll(presetsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	writePreset(t, presetsDir, "my-custom", `[[line]]`+"\n"+"widgets = []\n")

	all := ListAll()
	// BuiltinNames is empty for now, so ListAll == ListCustom.
	if len(all) < 1 {
		t.Fatalf("expected at least 1 entry, got %v", all)
	}
	if all[len(all)-1] != "my-custom" {
		t.Errorf("expected my-custom at end, got %v", all)
	}
}

func TestListAll_NoDuplicates(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	presetsDir := filepath.Join(dir, ".config", "tail-claude-hud", "presets")
	if err := os.MkdirAll(presetsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	writePreset(t, presetsDir, "alpha", `[[line]]`+"\n"+"widgets = []\n")
	writePreset(t, presetsDir, "beta", `[[line]]`+"\n"+"widgets = []\n")

	all := ListAll()
	seen := make(map[string]int)
	for _, n := range all {
		seen[n]++
		if seen[n] > 1 {
			t.Errorf("duplicate entry %q in ListAll", n)
		}
	}
}
