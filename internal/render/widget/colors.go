package widget

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// agentColors is the 8-color palette for agent identity, using ANSI 16 codes so colors
// adapt to the terminal's own palette.
// Indexed as: bright blue, bright green, bright red, bright yellow, bright magenta, bright cyan, yellow, magenta.
var agentColors = [8]string{"12", "10", "9", "11", "13", "14", "3", "5"}

// circleSliceIcons are Nerd Font circle-slice characters representing fill levels 1/8 through
// 8/8 (U+F0A9E–U+F0AA5). Index 0 is nearly empty, index 7 is fully filled.
var circleSliceIcons = [8]string{
	"\U000F0A9E", // circle_slice_1 — 1/8 filled
	"\U000F0A9F", // circle_slice_2 — 2/8 filled
	"\U000F0AA0", // circle_slice_3 — 3/8 filled
	"\U000F0AA1", // circle_slice_4 — 4/8 filled
	"\U000F0AA2", // circle_slice_5 — 5/8 filled
	"\U000F0AA3", // circle_slice_6 — 6/8 filled
	"\U000F0AA4", // circle_slice_7 — 7/8 filled
	"\U000F0AA5", // circle_slice_8 — fully filled
}

// percentToIcon maps a percentage (0–100) to one of 8 circle-slice Nerd Font icons.
// The mapping divides the range into equal eighths: index = (percent * 8) / 100,
// clamped to [0, 7] so that 100% yields the fully-filled icon.
func percentToIcon(percent int) string {
	if percent <= 0 {
		return circleSliceIcons[0]
	}
	if percent >= 100 {
		return circleSliceIcons[7]
	}
	idx := (percent * 8) / 100
	if idx > 7 {
		idx = 7
	}
	return circleSliceIcons[idx]
}

// AgentColorStyle returns a foreground lipgloss.Style for the given color index.
// The index wraps around the 8-color palette, so any non-negative integer is valid.
func AgentColorStyle(colorIndex int) lipgloss.Style {
	color := agentColors[colorIndex%8]
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color))
}

var (
	opusStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	sonnetStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	haikuStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	defaultModelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
)

// ModelFamily returns the model family name ("opus", "sonnet", "haiku") from
// a model identifier string. Returns "" for unrecognized models.
func ModelFamily(modelName string) string {
	lower := strings.ToLower(modelName)
	switch {
	case strings.Contains(lower, "opus"):
		return "opus"
	case strings.Contains(lower, "sonnet"):
		return "sonnet"
	case strings.Contains(lower, "haiku"):
		return "haiku"
	default:
		return ""
	}
}

// ModelFamilyColor returns a foreground lipgloss.Style based on the Claude model family.
// Detection is case-insensitive via strings.Contains on the lowercased model name:
//   - "opus"   → bright red (ANSI 9)
//   - "sonnet" → bright blue (ANSI 12)
//   - "haiku"  → bright green (ANSI 10)
//   - default  → bright cyan (ANSI 14)
func ModelFamilyColor(modelName string) lipgloss.Style {
	switch ModelFamily(modelName) {
	case "opus":
		return opusStyle
	case "sonnet":
		return sonnetStyle
	case "haiku":
		return haikuStyle
	default:
		return defaultModelStyle
	}
}
