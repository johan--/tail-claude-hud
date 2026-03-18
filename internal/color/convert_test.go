package color

import (
	"testing"
)

func TestHexToRGB(t *testing.T) {
	tests := []struct {
		hex     string
		r, g, b uint8
		wantErr bool
	}{
		{"#FF0000", 255, 0, 0, false},
		{"#00FF00", 0, 255, 0, false},
		{"#0000FF", 0, 0, 255, false},
		{"#FFFFFF", 255, 255, 255, false},
		{"#000000", 0, 0, 0, false},
		{"#1a2b3c", 0x1a, 0x2b, 0x3c, false},
		// Shorthand #RGB expands to #RRGGBB.
		{"#F00", 255, 0, 0, false},
		{"#0F0", 0, 255, 0, false},
		{"#00F", 0, 0, 255, false},
		// Error cases.
		{"", 0, 0, 0, true},
		{"#ZZZ", 0, 0, 0, true},
		{"#12345", 0, 0, 0, true}, // wrong length
	}

	for _, tc := range tests {
		r, g, b, err := HexToRGB(tc.hex)
		if tc.wantErr {
			if err == nil {
				t.Errorf("HexToRGB(%q): expected error, got nil", tc.hex)
			}
			continue
		}
		if err != nil {
			t.Errorf("HexToRGB(%q): unexpected error: %v", tc.hex, err)
			continue
		}
		if r != tc.r || g != tc.g || b != tc.b {
			t.Errorf("HexToRGB(%q) = (%d,%d,%d), want (%d,%d,%d)",
				tc.hex, r, g, b, tc.r, tc.g, tc.b)
		}
	}
}

func TestHexToAnsi256(t *testing.T) {
	tests := []struct {
		hex  string
		want int
	}{
		// Pure black → cube floor.
		{"#000000", 16},
		// Pure white → cube cap.
		{"#FFFFFF", 231},
		// Pure red in the cube.
		{"#FF0000", 196},
		// Pure green in the cube.
		{"#00FF00", 46},
		// Pure blue in the cube.
		{"#0000FF", 21},
		// A mid-grey falls in the grayscale ramp.
		{"#808080", 244},
		// Near-black grey stays at 16 (cube floor via grayscale path).
		{"#030303", 16},
		// Near-white grey: 248 is not > 248, so falls through to ramp: ((248-8)/247)*24 ≈ 23 → index 255.
		{"#F8F8F8", 255},
		// Invalid input → -1.
		{"notacolor", -1},
	}

	for _, tc := range tests {
		got := HexToAnsi256(tc.hex)
		if got != tc.want {
			t.Errorf("HexToAnsi256(%q) = %d, want %d", tc.hex, got, tc.want)
		}
	}
}

func TestAnsi256ToBasic(t *testing.T) {
	tests := []struct {
		idx  int
		want int
	}{
		// Indices 0-15 pass through.
		{0, 0},
		{7, 7},
		{15, 15},
		// Grayscale ramp: step = idx-232, thresholds at 6/12/18.
		{232, 0},  // step 0: near-black → black
		{237, 0},  // step 5: still < 6 → black
		{238, 8},  // step 6: dark gray
		{243, 8},  // step 11: < 12 → dark gray
		{244, 7},  // step 12: light gray
		{249, 7},  // step 17: < 18 → light gray
		{250, 15}, // step 18: bright white
		{255, 15}, // step 23: bright white
		// Color cube red → basic red.
		{196, 9}, // bright red (full red in cube)
		// Negative index → 0.
		{-1, 0},
	}

	for _, tc := range tests {
		got := Ansi256ToBasic(tc.idx)
		if got != tc.want {
			t.Errorf("Ansi256ToBasic(%d) = %d, want %d", tc.idx, got, tc.want)
		}
	}
}

func TestHexToBasic(t *testing.T) {
	tests := []struct {
		hex  string
		want int
	}{
		// Saturated primaries.
		{"#FF0000", 9},  // bright red
		{"#00FF00", 10}, // bright green
		{"#0000FF", 12}, // bright blue
		{"#FFFF00", 11}, // bright yellow (red+green dominant)
		// Dark versions stay in normal range.
		{"#800000", 1}, // red
		{"#008000", 2}, // green
		// White and black.
		{"#FFFFFF", 15},
		{"#000000", 0},
		// Invalid → -1.
		{"bad", -1},
	}

	for _, tc := range tests {
		got := HexToBasic(tc.hex)
		if got != tc.want {
			t.Errorf("HexToBasic(%q) = %d, want %d", tc.hex, got, tc.want)
		}
	}
}

// TestRgbToAnsi256RoundTrip verifies that non-grayscale 6x6x6 cube entries
// round-trip correctly. Grayscale entries (r==g==b) are routed to the 24-step
// ramp instead of the cube, so they are skipped here.
func TestRgbToAnsi256RoundTrip(t *testing.T) {
	// The six discrete steps on a single axis in the 6x6x6 cube: 0,51,102,153,204,255.
	steps := []uint8{0, 51, 102, 153, 204, 255}

	for ri := 0; ri < 6; ri++ {
		for gi := 0; gi < 6; gi++ {
			for bi := 0; bi < 6; bi++ {
				if ri == gi && gi == bi {
					// Grayscale diagonal goes to the ramp, not the cube.
					continue
				}
				r, g, b := steps[ri], steps[gi], steps[bi]
				want := 16 + 36*ri + 6*gi + bi
				got := rgbToAnsi256(r, g, b)
				if got != want {
					t.Errorf("rgbToAnsi256(%d,%d,%d) = %d, want %d",
						r, g, b, got, want)
				}
			}
		}
	}
}

// TestResolveColorName_NamedColorsMapToANSINumbers verifies that the 16
// standard ANSI color names and common aliases resolve to their numeric
// equivalents, which lipgloss.Color() can parse.
func TestResolveColorName_NamedColorsMapToANSINumbers(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"black", "0"},
		{"red", "1"},
		{"green", "2"},
		{"yellow", "3"},
		{"blue", "4"},
		{"magenta", "5"},
		{"cyan", "6"},
		{"white", "7"},
		{"brightblack", "8"},
		{"gray", "8"},
		{"grey", "8"},
		{"brightred", "9"},
		{"brightgreen", "10"},
		{"brightyellow", "11"},
		{"brightblue", "12"},
		{"brightmagenta", "13"},
		{"brightcyan", "14"},
		{"brightwhite", "15"},
	}
	for _, tc := range cases {
		got := ResolveColorName(tc.input)
		if got != tc.want {
			t.Errorf("ResolveColorName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// TestResolveColorName_CaseInsensitive verifies that mixed-case named colors
// resolve to the same numeric string as their lowercase equivalents.
func TestResolveColorName_CaseInsensitive(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"Green", "2"},
		{"GREEN", "2"},
		{"Red", "1"},
		{"YELLOW", "3"},
		{"BrightCyan", "14"},
	}
	for _, tc := range cases {
		got := ResolveColorName(tc.input)
		if got != tc.want {
			t.Errorf("ResolveColorName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// TestResolveColorName_PassThrough verifies that hex strings, numeric strings,
// and unrecognised names are returned unchanged.
func TestResolveColorName_PassThrough(t *testing.T) {
	cases := []string{
		"#ff0000",
		"#abc",
		"114",
		"2",
		"",
		"unknown-color",
	}
	for _, input := range cases {
		got := ResolveColorName(input)
		if got != input {
			t.Errorf("ResolveColorName(%q) = %q, want pass-through %q", input, got, input)
		}
	}
}
