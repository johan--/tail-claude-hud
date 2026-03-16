package widget

import (
	"testing"

	"charm.land/lipgloss/v2"
)

// -- AgentColorStyle ----------------------------------------------------------

func TestAgentColorStyle_IndicesReturnDistinctColors(t *testing.T) {
	// Each index 0-7 must produce a style that renders text differently,
	// confirming the palette entries are all distinct.
	probe := "X"
	rendered := make([]string, 8)
	for i := 0; i < 8; i++ {
		rendered[i] = AgentColorStyle(i).Render(probe)
	}
	for i := 0; i < 8; i++ {
		for j := i + 1; j < 8; j++ {
			if rendered[i] == rendered[j] {
				t.Errorf("AgentColorStyle(%d) and AgentColorStyle(%d) produced identical output %q", i, j, rendered[i])
			}
		}
	}
}

func TestAgentColorStyle_IndexEightWrapsToIndexZero(t *testing.T) {
	probe := "X"
	got := AgentColorStyle(8).Render(probe)
	want := AgentColorStyle(0).Render(probe)
	if got != want {
		t.Errorf("AgentColorStyle(8) = %q, want same as AgentColorStyle(0) = %q", got, want)
	}
}

func TestAgentColorStyle_LargeIndexWraps(t *testing.T) {
	// Index 16 should wrap to index 0 (16 % 8 == 0).
	probe := "X"
	got := AgentColorStyle(16).Render(probe)
	want := AgentColorStyle(0).Render(probe)
	if got != want {
		t.Errorf("AgentColorStyle(16) = %q, want same as AgentColorStyle(0) = %q", got, want)
	}
}

// -- ModelFamilyColor ---------------------------------------------------------

func TestModelFamilyColor_OpusDetectedCaseInsensitive(t *testing.T) {
	cases := []string{"opus", "Opus", "OPUS", "claude-opus-4", "Claude Opus 4.6"}
	want := lipgloss.NewStyle().Foreground(lipgloss.Color("204")).Render("X")
	for _, name := range cases {
		got := ModelFamilyColor(name).Render("X")
		if got != want {
			t.Errorf("ModelFamilyColor(%q) did not return coral(204); got %q", name, got)
		}
	}
}

func TestModelFamilyColor_SonnetDetectedCaseInsensitive(t *testing.T) {
	cases := []string{"sonnet", "Sonnet", "SONNET", "claude-sonnet-4-6", "Claude Sonnet 4.6"}
	want := lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Render("X")
	for _, name := range cases {
		got := ModelFamilyColor(name).Render("X")
		if got != want {
			t.Errorf("ModelFamilyColor(%q) did not return blue(75); got %q", name, got)
		}
	}
}

func TestModelFamilyColor_HaikuDetectedCaseInsensitive(t *testing.T) {
	cases := []string{"haiku", "Haiku", "HAIKU", "claude-haiku-3-5", "Claude Haiku 4.5"}
	want := lipgloss.NewStyle().Foreground(lipgloss.Color("114")).Render("X")
	for _, name := range cases {
		got := ModelFamilyColor(name).Render("X")
		if got != want {
			t.Errorf("ModelFamilyColor(%q) did not return green(114); got %q", name, got)
		}
	}
}

func TestModelFamilyColor_DefaultReturnsCyan(t *testing.T) {
	cases := []string{"", "gpt-4o", "gemini-pro", "unknown-model"}
	want := lipgloss.NewStyle().Foreground(lipgloss.Color("87")).Render("X")
	for _, name := range cases {
		got := ModelFamilyColor(name).Render("X")
		if got != want {
			t.Errorf("ModelFamilyColor(%q) did not return cyan(87); got %q", name, got)
		}
	}
}
