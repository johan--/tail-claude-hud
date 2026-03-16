package color

import (
	"testing"
)


func TestDetectLevel_COLORTERM(t *testing.T) {
	tests := []struct {
		name      string
		colorterm string
		want      Level
	}{
		{"truecolor", "truecolor", LevelTruecolor},
		{"24bit", "24bit", LevelTruecolor},
		{"other value", "1", LevelBasic},
		{"empty", "", LevelBasic},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("COLORTERM", tc.colorterm)
			t.Setenv("TERM", "")
			t.Setenv("TERM_PROGRAM", "")

			got := DetectLevel()
			if got != tc.want {
				t.Errorf("DetectLevel() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDetectLevel_TERM(t *testing.T) {
	tests := []struct {
		name string
		term string
		want Level
	}{
		{"xterm-kitty", "xterm-kitty", LevelTruecolor},
		{"xterm-ghostty", "xterm-ghostty", LevelTruecolor},
		{"wezterm", "wezterm", LevelTruecolor},
		{"alacritty", "alacritty", LevelTruecolor},
		{"foot", "foot", LevelTruecolor},
		{"contour", "contour", LevelTruecolor},
		{"xterm-256color", "xterm-256color", Level256},
		{"screen-256color", "screen-256color", Level256},
		{"xterm", "xterm", LevelBasic},
		{"dumb", "dumb", LevelBasic},
		{"empty", "", LevelBasic},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("COLORTERM", "")
			t.Setenv("TERM", tc.term)
			t.Setenv("TERM_PROGRAM", "")

			got := DetectLevel()
			if got != tc.want {
				t.Errorf("DetectLevel() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDetectLevel_TERM_PROGRAM(t *testing.T) {
	tests := []struct {
		name        string
		termProgram string
		want        Level
	}{
		{"iTerm.app", "iTerm.app", LevelTruecolor},
		{"iTerm2", "iTerm2", LevelTruecolor},
		{"kitty", "kitty", LevelTruecolor},
		{"WezTerm", "WezTerm", LevelTruecolor},
		{"Hyper", "Hyper", LevelTruecolor},
		{"Alacritty", "Alacritty", LevelTruecolor},
		{"Rio", "Rio", LevelTruecolor},
		{"vscode", "vscode", LevelTruecolor},
		{"Tabby", "Tabby", LevelTruecolor},
		{"Apple_Terminal", "Apple_Terminal", LevelBasic},
		{"unknown", "SomeUnknownTerm", LevelBasic},
		{"empty", "", LevelBasic},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("COLORTERM", "")
			t.Setenv("TERM", "")
			t.Setenv("TERM_PROGRAM", tc.termProgram)

			got := DetectLevel()
			if got != tc.want {
				t.Errorf("DetectLevel() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDetectLevel_Priority(t *testing.T) {
	// COLORTERM takes precedence over TERM and TERM_PROGRAM.
	t.Setenv("COLORTERM", "truecolor")
	t.Setenv("TERM", "xterm-256color") // would be Level256 alone
	t.Setenv("TERM_PROGRAM", "Apple_Terminal")

	got := DetectLevel()
	if got != LevelTruecolor {
		t.Errorf("DetectLevel() = %v, want LevelTruecolor (COLORTERM should win)", got)
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input   string
		want    Level
		wantOK  bool
	}{
		{"truecolor", LevelTruecolor, true},
		{"TRUECOLOR", LevelTruecolor, true},
		{"256", Level256, true},
		{"basic", LevelBasic, true},
		{"BASIC", LevelBasic, true},
		{"auto", LevelBasic, false},
		{"", LevelBasic, false},
		{"garbage", LevelBasic, false},
	}

	for _, tc := range tests {
		got, ok := ParseLevel(tc.input)
		if ok != tc.wantOK || got != tc.want {
			t.Errorf("ParseLevel(%q) = (%v, %v), want (%v, %v)",
				tc.input, got, ok, tc.want, tc.wantOK)
		}
	}
}

func TestLevelFromConfig(t *testing.T) {
	// Force a known environment for deterministic auto-detect in these tests.
	t.Setenv("COLORTERM", "")
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "")

	tests := []struct {
		override string
		want     Level
	}{
		{"truecolor", LevelTruecolor},
		{"256", Level256},
		{"basic", LevelBasic},
		{"auto", Level256},  // auto → DetectLevel() → Level256 given xterm-256color
		{"", Level256},      // empty → DetectLevel()
		{"garbage", Level256}, // unknown → DetectLevel()
	}

	for _, tc := range tests {
		got := LevelFromConfig(tc.override)
		if got != tc.want {
			t.Errorf("LevelFromConfig(%q) = %v, want %v", tc.override, got, tc.want)
		}
	}
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{LevelTruecolor, "truecolor"},
		{Level256, "256"},
		{LevelBasic, "basic"},
	}
	for _, tc := range tests {
		if got := tc.level.String(); got != tc.want {
			t.Errorf("Level(%d).String() = %q, want %q", tc.level, got, tc.want)
		}
	}
}
