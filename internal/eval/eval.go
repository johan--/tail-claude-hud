// Package eval provides a graded design report for rendered statusline output.
// It wires together the color math (color.go), terminal palettes (palette.go),
// and ANSI parser (parse.go) to score four dimensions of visual quality:
// contrast, coherence, hierarchy, and adaptability.
package eval

import (
	"fmt"
	"strings"
)

// Grade is a letter grade for a scored dimension.
type Grade string

const (
	GradeA Grade = "A"
	GradeB Grade = "B"
	GradeC Grade = "C"
	GradeD Grade = "D"
	GradeF Grade = "F"
)

// DimensionResult holds the outcome of scoring one design dimension.
type DimensionResult struct {
	Name     string
	Grade    Grade
	Findings []string
}

// Report aggregates all dimension results and an overall grade.
type Report struct {
	Dimensions []DimensionResult
	Overall    Grade
}

// Evaluate parses rendered ANSI output, scores each design dimension, and
// returns a Report. The mode parameter is accepted for future use but does
// not currently affect scoring.
func Evaluate(rendered string, mode string) Report {
	segments := Parse(rendered)

	dims := []DimensionResult{
		evalContrast(segments),
		evalCoherence(segments),
		evalHierarchy(segments),
		evalAdaptability(segments),
	}

	// Compute overall grade as the average of dimension grades mapped back to a letter.
	total := 0
	for _, d := range dims {
		total += gradeToNum(d.Grade)
	}
	avg := total / len(dims)

	return Report{
		Dimensions: dims,
		Overall:    numToGrade(avg),
	}
}

// resolveColor converts a Color to an RGB value using the supplied palette for
// ColorDefault and ColorANSI16 lookups.
func resolveColor(c Color, palette TerminalPalette, isFg bool) RGB {
	switch c.Type {
	case ColorDefault:
		if isFg {
			return palette.DefaultFg
		}
		return palette.DefaultBg
	case ColorANSI16:
		idx := c.Index
		if idx < 0 {
			idx = 0
		}
		if idx > 15 {
			idx = 15
		}
		return palette.Colors[idx]
	case ColorXterm256:
		return Xterm256ToRGB(c.Index)
	case ColorTruecolor:
		return RGB{c.R, c.G, c.B}
	default:
		if isFg {
			return palette.DefaultFg
		}
		return palette.DefaultBg
	}
}

// gradeToNum converts a Grade to a numeric score for averaging (A=4, F=0).
func gradeToNum(g Grade) int {
	switch g {
	case GradeA:
		return 4
	case GradeB:
		return 3
	case GradeC:
		return 2
	case GradeD:
		return 1
	default: // GradeF
		return 0
	}
}

// numToGrade converts a numeric average back to a Grade letter.
func numToGrade(n int) Grade {
	switch {
	case n >= 4:
		return GradeA
	case n >= 3:
		return GradeB
	case n >= 2:
		return GradeC
	case n >= 1:
		return GradeD
	default:
		return GradeF
	}
}

// evalContrast scores WCAG contrast ratios for each segment across all palettes.
// It tests every fg/bg combination against every palette in AllPalettes() and
// grades based on the worst-case ratio observed.
//
// Grade thresholds:
//   - A: all ratios >= 4.5
//   - B: all ratios >= 3.0
//   - C: minimum ratio >= 2.0
//   - D: minimum ratio >= 1.5
//   - F: any ratio < 1.5
func evalContrast(segments []StyledSegment) DimensionResult {
	palettes := AllPalettes()
	var findings []string
	overallMin := 21.0 // max possible contrast ratio

	for _, seg := range segments {
		if strings.TrimSpace(seg.Text) == "" {
			continue
		}
		segMin := 21.0
		var worstPalette string
		for _, pal := range palettes {
			fg := resolveColor(seg.Fg, pal, true)
			bg := resolveColor(seg.Bg, pal, false)
			ratio := ContrastRatio(fg, bg)
			if ratio < segMin {
				segMin = ratio
				worstPalette = pal.Name
			}
		}

		if segMin < overallMin {
			overallMin = segMin
		}

		label := segLabel(seg.Text)
		var status string
		switch {
		case segMin >= 4.5:
			status = fmt.Sprintf("PASS  %s  ratio=%.2f  worst-palette=%s", label, segMin, worstPalette)
		case segMin >= 3.0:
			status = fmt.Sprintf("WARN  %s  ratio=%.2f  worst-palette=%s", label, segMin, worstPalette)
		default:
			status = fmt.Sprintf("FAIL  %s  ratio=%.2f  worst-palette=%s", label, segMin, worstPalette)
		}
		findings = append(findings, status)
	}

	if len(findings) == 0 {
		return DimensionResult{
			Name:     "Contrast",
			Grade:    GradeF,
			Findings: []string{"no visible segments found"},
		}
	}

	var grade Grade
	switch {
	case overallMin >= 4.5:
		grade = GradeA
	case overallMin >= 3.0:
		grade = GradeB
	case overallMin >= 2.0:
		grade = GradeC
	case overallMin >= 1.5:
		grade = GradeD
	default:
		grade = GradeF
	}

	return DimensionResult{
		Name:     "Contrast",
		Grade:    grade,
		Findings: findings,
	}
}

// evalCoherence scores the lightness distribution and hue consistency of
// foreground colors across all segments.
//
// Grade thresholds:
//   - A: 4 distinct lightness buckets present and no adjacent hue collisions
//   - B: 3+ distinct lightness buckets
//   - C: 2 distinct lightness buckets
//   - D/F: only 1 or 0 distinct lightness buckets
func evalCoherence(segments []StyledSegment) DimensionResult {
	var findings []string

	// Collect all non-default fg colors and convert to HSL.
	type hslColor struct {
		h, s, l float64
	}
	var fgColors []hslColor
	buckets := make(map[int]bool)

	for _, seg := range segments {
		if strings.TrimSpace(seg.Text) == "" {
			continue
		}
		if seg.Fg.Type == ColorDefault {
			continue
		}
		// Resolve to RGB using the first palette (XtermDefault) as a reference.
		rgb := resolveColor(seg.Fg, XtermDefault, true)
		h, s, l := RGBToHSL(rgb)
		fgColors = append(fgColors, hslColor{h, s, l})

		// Bucket lightness into 4 ranges.
		bucket := lightnessTooBucket(l)
		buckets[bucket] = true
	}

	distinctLevels := len(buckets)
	findings = append(findings, fmt.Sprintf("distinct lightness buckets: %d/4", distinctLevels))

	// Check adjacent segments for hue collisions (delta < 20 degrees).
	collisions := 0
	var prevHSL *hslColor
	for i, seg := range segments {
		if strings.TrimSpace(seg.Text) == "" {
			continue
		}
		if seg.Fg.Type == ColorDefault {
			prevHSL = nil
			continue
		}
		if i < len(fgColors) {
			cur := &fgColors[len(fgColors)-1] // placeholder; recompute inline
			rgb := resolveColor(seg.Fg, XtermDefault, true)
			h, _, _ := RGBToHSL(rgb)
			curHSL := &hslColor{h: h}
			cur = curHSL

			if prevHSL != nil {
				delta := HueDelta(prevHSL.h, cur.h)
				if delta < 20 {
					collisions++
					findings = append(findings, fmt.Sprintf(
						"adjacent hue collision: delta=%.1f°  near %q", delta, segLabel(seg.Text),
					))
				}
			}
			prevHSL = cur
		}
	}
	if collisions == 0 {
		findings = append(findings, "no adjacent hue collisions")
	}

	// Count dim mechanisms.
	faintCount := 0
	dimColorCount := 0
	for _, seg := range segments {
		if seg.Faint {
			faintCount++
		}
		if seg.Fg.Type != ColorDefault {
			rgb := resolveColor(seg.Fg, XtermDefault, true)
			_, _, l := RGBToHSL(rgb)
			if l < 0.25 {
				dimColorCount++
			}
		}
	}
	findings = append(findings, fmt.Sprintf("dim mechanisms: faint-attr=%d  low-lightness-colors=%d", faintCount, dimColorCount))

	var grade Grade
	switch {
	case distinctLevels >= 4 && collisions == 0:
		grade = GradeA
	case distinctLevels >= 3:
		grade = GradeB
	case distinctLevels >= 2:
		grade = GradeC
	case distinctLevels == 1:
		grade = GradeD
	default:
		grade = GradeF
	}

	return DimensionResult{
		Name:     "Coherence",
		Grade:    grade,
		Findings: findings,
	}
}

// lightnessTooBucket assigns a lightness value to one of four buckets:
// 0: <0.25, 1: 0.25-0.5, 2: 0.5-0.75, 3: >0.75
func lightnessTooBucket(l float64) int {
	switch {
	case l < 0.25:
		return 0
	case l < 0.5:
		return 1
	case l < 0.75:
		return 2
	default:
		return 3
	}
}

// evalHierarchy classifies each segment into brightness tiers and reports the
// tier distribution.
//
// Tiers:
//   - Tier 1: Bold=true or lightness > 0.6 (prominent)
//   - Tier 2: Lightness 0.3-0.6 (normal)
//   - Tier 3: Faint=true or lightness < 0.3 or ColorDefault (receded)
//
// Grade:
//   - A: all 3 tiers present
//   - B: 2 tiers present
//   - C/D/F: 1 or 0 tiers
func evalHierarchy(segments []StyledSegment) DimensionResult {
	tierCounts := [4]int{} // index 1-3 used; 0 unused

	for _, seg := range segments {
		if strings.TrimSpace(seg.Text) == "" {
			continue
		}
		tier := segmentTier(seg)
		tierCounts[tier]++
	}

	var findings []string
	findings = append(findings, fmt.Sprintf(
		"tier distribution: tier1(prominent)=%d  tier2(normal)=%d  tier3(receded)=%d",
		tierCounts[1], tierCounts[2], tierCounts[3],
	))

	tiersPresent := 0
	for t := 1; t <= 3; t++ {
		if tierCounts[t] > 0 {
			tiersPresent++
		}
	}
	findings = append(findings, fmt.Sprintf("distinct tiers present: %d/3", tiersPresent))

	var grade Grade
	switch tiersPresent {
	case 3:
		grade = GradeA
	case 2:
		grade = GradeB
	case 1:
		grade = GradeC
	default:
		grade = GradeF
	}

	return DimensionResult{
		Name:     "Hierarchy",
		Grade:    grade,
		Findings: findings,
	}
}

// segmentTier returns the brightness tier for a segment (1=prominent, 2=normal, 3=receded).
func segmentTier(seg StyledSegment) int {
	if seg.Faint || seg.Fg.Type == ColorDefault {
		return 3
	}

	rgb := resolveColor(seg.Fg, XtermDefault, true)
	_, _, l := RGBToHSL(rgb)

	switch {
	case seg.Bold || l > 0.6:
		return 1
	case l < 0.3:
		return 3
	default:
		return 2
	}
}

// evalAdaptability scores how well the statusline adapts to different terminal
// palettes by measuring the proportion of ANSI16 colors (which follow the
// user's palette) vs xterm-256 or truecolor (which are fixed).
//
// Grade:
//   - A: >50% of fg colors are ANSI16
//   - B: >25% ANSI16
//   - C: any ANSI16 present
//   - D: all non-default colors are xterm-256 or truecolor
//   - F: zero non-empty segments
func evalAdaptability(segments []StyledSegment) DimensionResult {
	var ansi16, xterm256, truecolor, def int

	for _, seg := range segments {
		if strings.TrimSpace(seg.Text) == "" {
			continue
		}
		switch seg.Fg.Type {
		case ColorDefault:
			def++
		case ColorANSI16:
			ansi16++
		case ColorXterm256:
			xterm256++
		case ColorTruecolor:
			truecolor++
		}
	}

	total := ansi16 + xterm256 + truecolor + def
	var findings []string
	findings = append(findings, fmt.Sprintf(
		"fg color types: ansi16=%d  xterm256=%d  truecolor=%d  default=%d  total=%d",
		ansi16, xterm256, truecolor, def, total,
	))

	if total == 0 {
		return DimensionResult{
			Name:     "Adaptability",
			Grade:    GradeF,
			Findings: findings,
		}
	}

	nonDefault := ansi16 + xterm256 + truecolor
	var grade Grade
	if nonDefault == 0 {
		// All default colors — maximally adaptive but nothing to score.
		grade = GradeA
		findings = append(findings, "all segments use terminal default colors (maximally adaptive)")
	} else {
		pct := float64(ansi16) / float64(nonDefault)
		findings = append(findings, fmt.Sprintf("ANSI16 proportion of non-default fg: %.0f%%", pct*100))
		switch {
		case pct > 0.50:
			grade = GradeA
		case pct > 0.25:
			grade = GradeB
		case ansi16 > 0:
			grade = GradeC
		default:
			grade = GradeD
		}
	}

	return DimensionResult{
		Name:     "Adaptability",
		Grade:    grade,
		Findings: findings,
	}
}

// FormatReport formats a Report as human-readable text suitable for logging or
// terminal output.
func FormatReport(report Report) string {
	var sb strings.Builder
	sb.WriteString("=== Statusline Design Evaluation ===\n")

	for i, dim := range report.Dimensions {
		// Right-align the grade in a fixed-width column.
		label := fmt.Sprintf("Dimension %d: %s", i+1, dim.Name)
		dots := strings.Repeat(".", 40-len(label))
		if len(dots) < 1 {
			dots = " "
		}
		sb.WriteString(fmt.Sprintf("%s%s %s\n", label, dots, dim.Grade))
		for _, f := range dim.Findings {
			sb.WriteString(fmt.Sprintf("  %s\n", f))
		}
	}

	sb.WriteString(fmt.Sprintf("Overall: %s\n", report.Overall))
	return sb.String()
}

// segLabel returns a short printable label for a segment (trimmed, max 30 chars).
func segLabel(text string) string {
	t := strings.TrimSpace(text)
	// Replace non-breaking spaces (U+00A0) with regular spaces for readability.
	t = strings.ReplaceAll(t, "\u00a0", " ")
	if len(t) > 30 {
		t = t[:27] + "..."
	}
	if t == "" {
		return "(space)"
	}
	return fmt.Sprintf("%q", t)
}
