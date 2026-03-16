// Package color handles terminal color capability detection and color conversion.
// Detection is auto-derived from environment variables and can be overridden
// via the config color_level field.
package color

import (
	"os"
	"strings"
)

// Level represents the color depth supported by the terminal.
type Level int

const (
	// LevelBasic supports the 16 standard ANSI colors (30-37, 90-97).
	LevelBasic Level = iota
	// Level256 supports 256-color mode via ANSI escape sequences.
	Level256
	// LevelTruecolor supports 24-bit RGB color.
	LevelTruecolor
)

// String returns the canonical name for a Level.
func (l Level) String() string {
	switch l {
	case LevelTruecolor:
		return "truecolor"
	case Level256:
		return "256"
	default:
		return "basic"
	}
}

// truecolorTermPrograms are terminal emulators whose TERM_PROGRAM value
// indicates truecolor support.
var truecolorTermPrograms = map[string]bool{
	"iTerm.app": true,
	"iTerm2":    true,
	"kitty":     true,
	"WezTerm":   true,
	"Hyper":     true,
	"Alacritty": true,
	"Rio":       true,
	"vscode":    true,
	"Tabby":     true,
}

// truecolorTermValues are values for the TERM variable that imply truecolor.
var truecolorTermValues = map[string]bool{
	"xterm-kitty":   true,
	"xterm-ghostty": true,
	"wezterm":       true,
	"alacritty":     true,
	"foot":          true,
	"contour":       true,
}

// DetectLevel returns the color level from the environment.
// Detection order:
//  1. COLORTERM=truecolor or 24bit → LevelTruecolor
//  2. TERM is a known truecolor terminal → LevelTruecolor
//  3. TERM_PROGRAM is a known truecolor application → LevelTruecolor
//  4. TERM contains "256color" → Level256
//  5. Default → LevelBasic
//
// The caller is responsible for applying any config override before
// calling this function (see ParseLevel / LevelFromConfig).
func DetectLevel() Level {
	colorterm := os.Getenv("COLORTERM")
	if colorterm == "truecolor" || colorterm == "24bit" {
		return LevelTruecolor
	}

	term := os.Getenv("TERM")
	if truecolorTermValues[term] {
		return LevelTruecolor
	}

	termProgram := os.Getenv("TERM_PROGRAM")
	if truecolorTermPrograms[termProgram] {
		return LevelTruecolor
	}

	if strings.Contains(strings.ToLower(term), "256color") {
		return Level256
	}

	// COLORTERM set to anything else (e.g. "1") means at least basic color.
	// Absence leaves us at basic too, which is the safe fallback.
	return LevelBasic
}

// ParseLevel converts a config string to a Level.
// Accepted values: "truecolor", "256", "basic".
// Unknown values return LevelBasic and false.
func ParseLevel(s string) (Level, bool) {
	switch strings.ToLower(s) {
	case "truecolor":
		return LevelTruecolor, true
	case "256":
		return Level256, true
	case "basic":
		return LevelBasic, true
	default:
		return LevelBasic, false
	}
}

// LevelFromConfig resolves the effective Level given a config override string.
// When override is "auto" or empty, DetectLevel is called. Otherwise ParseLevel
// is attempted; unrecognised values fall back to DetectLevel.
func LevelFromConfig(override string) Level {
	if override == "" || strings.ToLower(override) == "auto" {
		return DetectLevel()
	}
	if level, ok := ParseLevel(override); ok {
		return level
	}
	return DetectLevel()
}
