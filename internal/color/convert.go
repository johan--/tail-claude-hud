package color

import (
	"fmt"
	"strconv"
	"strings"
)

// namedANSIColors maps CSS/ANSI color names to their ANSI numeric string
// equivalents. lipgloss.Color() only accepts hex strings or numeric strings;
// named colors like "green" return noColor and produce no styling output.
// This table covers the 16 standard ANSI names plus common aliases advertised
// in the config template and theme override examples.
var namedANSIColors = map[string]string{
	"black":         "0",
	"red":           "1",
	"green":         "2",
	"yellow":        "3",
	"blue":          "4",
	"magenta":       "5",
	"cyan":          "6",
	"white":         "7",
	"brightblack":   "8",
	"gray":          "8",
	"grey":          "8",
	"brightred":     "9",
	"brightgreen":   "10",
	"brightyellow":  "11",
	"brightblue":    "12",
	"brightmagenta": "13",
	"brightcyan":    "14",
	"brightwhite":   "15",
}

// ResolveColorName converts a CSS/ANSI color name to the numeric string
// equivalent expected by lipgloss.Color(). Hex strings ("#rrggbb") and
// numeric strings ("2", "114") pass through unchanged. Unrecognised names
// also pass through, which lets lipgloss handle them (it returns noColor for
// unknown values, which is the same behaviour as before this function existed).
//
// This is the canonical resolver for all color strings in the config system.
// Both the widget and render packages call this before passing config color
// values to lipgloss.Color().
func ResolveColorName(colorName string) string {
	if num, ok := namedANSIColors[strings.ToLower(colorName)]; ok {
		return num
	}
	return colorName
}

// HexToRGB parses a CSS hex color string (#RRGGBB or #RGB) into r, g, b
// components in the range [0, 255]. Returns an error if the input is not a
// valid hex color. The leading '#' is required.
func HexToRGB(hex string) (r, g, b uint8, err error) {
	if len(hex) == 0 || hex[0] != '#' {
		return 0, 0, 0, fmt.Errorf("invalid hex color %q: must start with '#'", hex)
	}
	hex = hex[1:]

	switch len(hex) {
	case 3:
		// Expand shorthand: #RGB → #RRGGBB
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	case 6:
		// already full form
	default:
		return 0, 0, 0, fmt.Errorf("invalid hex color %q: must be #RGB or #RRGGBB", "#"+hex)
	}

	rv, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid hex color: %w", err)
	}
	gv, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid hex color: %w", err)
	}
	bv, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid hex color: %w", err)
	}

	return uint8(rv), uint8(gv), uint8(bv), nil
}

// HexToAnsi256 converts a #RRGGBB hex color to the nearest xterm-256 palette
// index. The mapping uses the standard 6x6x6 color cube (indices 16-231) and
// the 24-step grayscale ramp (indices 232-255).
//
// Returns -1 if hex is not a valid color string.
func HexToAnsi256(hex string) int {
	r, g, b, err := HexToRGB(hex)
	if err != nil {
		return -1
	}
	return rgbToAnsi256(r, g, b)
}

// rgbToAnsi256 maps an RGB triple to the nearest xterm-256 palette index.
// Greyscale inputs (r==g==b) are routed to the 24-step ramp (232-255) when
// the value falls in the interior; extreme values snap to the cube floor/cap.
func rgbToAnsi256(r, g, b uint8) int {
	// Check for grayscale: only enter the ramp when all three channels are equal.
	if r == g && g == b {
		if r < 8 {
			return 16
		}
		if r > 248 {
			return 231
		}
		return int(((float64(r)-8)/247)*24+0.5) + 232
	}

	ri := int(float64(r)/255*5 + 0.5)
	gi := int(float64(g)/255*5 + 0.5)
	bi := int(float64(b)/255*5 + 0.5)
	return 16 + 36*ri + 6*gi + bi
}

// Ansi256ToBasic maps a 256-color index to the nearest of the 16 basic ANSI
// colors (0-15). Indices 0-15 pass through unchanged. Indices in the 6x6x6
// cube (16-231) are projected back to RGB and matched by brightness and hue.
// Grayscale ramp entries (232-255) map to black, dark gray, light gray, or
// white based on brightness.
func Ansi256ToBasic(idx int) int {
	if idx < 0 {
		return 0
	}
	if idx < 16 {
		return idx
	}

	if idx >= 232 {
		// Grayscale ramp: 232 is near-black, 255 is near-white.
		step := idx - 232 // 0-23
		switch {
		case step < 6:
			return 0 // black
		case step < 12:
			return 8 // dark gray (bright black)
		case step < 18:
			return 7 // light gray (white)
		default:
			return 15 // bright white
		}
	}

	// 6x6x6 cube: recover the RGB indices.
	idx -= 16
	bi := idx % 6
	gi := (idx / 6) % 6
	ri := idx / 36

	// Scale to 0-255.
	r := uint8(ri * 255 / 5)
	g := uint8(gi * 255 / 5)
	b := uint8(bi * 255 / 5)

	return rgbToBasic(r, g, b)
}

// HexToBasic converts a hex color directly to the nearest basic 16-color
// ANSI index (0-15). Returns -1 for invalid input.
func HexToBasic(hex string) int {
	r, g, b, err := HexToRGB(hex)
	if err != nil {
		return -1
	}
	return rgbToBasic(r, g, b)
}

// rgbToBasic maps an RGB value to one of the 16 basic ANSI colors (0-15).
// The heuristic prioritises the dominant channel to choose a named color
// family, then uses the dominant channel value to select normal vs. bright.
// "Bright" is chosen when the dominant channel exceeds 170 (2/3 of 255).
func rgbToBasic(r, g, b uint8) int {
	maxChannel := max3(int(r), int(g), int(b))
	minChannel := min3(int(r), int(g), int(b))
	brightness := (int(r) + int(g) + int(b)) / 3

	bright := maxChannel > 170

	// Achromatic: route to black/gray/white.
	if maxChannel-minChannel < 30 {
		switch {
		case brightness < 64:
			return 0 // black
		case brightness < 128:
			return 8 // bright black (dark gray)
		case brightness < 192:
			return 7 // white
		default:
			return 15 // bright white
		}
	}

	// Chromatic: identify dominant hue.
	if int(r) >= int(g) && int(r) >= int(b) {
		if int(g) > int(b)+30 {
			// Red and green together → yellow.
			if bright {
				return 11 // bright yellow
			}
			return 3 // yellow
		}
		if bright {
			return 9 // bright red
		}
		return 1 // red
	}

	if int(g) >= int(r) && int(g) >= int(b) {
		if bright {
			return 10 // bright green
		}
		return 2 // green
	}

	// Blue dominant.
	if int(r) > int(b)-30 && int(r) > 80 {
		// Blue and red → magenta.
		if bright {
			return 13 // bright magenta
		}
		return 5 // magenta
	}

	// Check for cyan (blue + green).
	if int(g) > int(b)-30 && int(g) > 80 {
		if bright {
			return 14 // bright cyan
		}
		return 6 // cyan
	}

	// Pure blue.
	if bright {
		return 12 // bright blue
	}
	return 4 // blue
}

func max3(a, b, c int) int {
	if a >= b && a >= c {
		return a
	}
	if b >= c {
		return b
	}
	return c
}

func min3(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= c {
		return b
	}
	return c
}
