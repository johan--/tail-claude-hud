package eval

import (
	"bytes"
	"testing"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/render"
)

// TestDesignEval renders the statusline with a realistic fixture context and
// runs the full design evaluator. The test always passes — it is informational.
// Use `just eval` to see the formatted report.
func TestDesignEval(t *testing.T) {
	ctx := &model.RenderContext{
		ModelDisplayName:  "Claude Sonnet 4",
		ContextPercent:    42,
		ContextWindowSize: 200000,
		InputTokens:       45000,
		CacheCreation:     12000,
		CacheRead:         8000,
		Cwd:               "/Users/test/project",
		Git: &model.GitStatus{
			Branch:  "main",
			AheadBy: 3,
		},
		TotalDurationMs: 185000,
	}
	cfg := config.LoadHud()

	var buf bytes.Buffer
	render.Render(&buf, ctx, cfg)

	rendered := buf.String()
	if rendered == "" {
		t.Log("Render produced no output — evaluation skipped")
		return
	}

	report := Evaluate(rendered, cfg.Style.Mode)

	// Print the full formatted report for inspection.
	t.Log("\n" + FormatReport(report))

	// Basic structural assertions — these are the only hard requirements.
	if len(report.Dimensions) != 4 {
		t.Errorf("expected 4 dimensions, got %d", len(report.Dimensions))
	}

	validGrades := map[Grade]bool{
		GradeA: true,
		GradeB: true,
		GradeC: true,
		GradeD: true,
		GradeF: true,
	}

	for _, dim := range report.Dimensions {
		if !validGrades[dim.Grade] {
			t.Errorf("dimension %q has invalid grade %q", dim.Name, dim.Grade)
		}
	}

	if !validGrades[report.Overall] {
		t.Errorf("overall grade %q is not a valid letter", report.Overall)
	}
}

// TestEvaluate_EmptyString ensures Evaluate handles empty input gracefully
// without panicking and still returns a 4-dimension report.
func TestEvaluate_EmptyString(t *testing.T) {
	report := Evaluate("", "plain")

	if len(report.Dimensions) != 4 {
		t.Fatalf("expected 4 dimensions, got %d", len(report.Dimensions))
	}
	for _, dim := range report.Dimensions {
		if dim.Grade == "" {
			t.Errorf("dimension %q has empty grade", dim.Name)
		}
	}
}

// TestEvaluate_PlainText ensures Evaluate handles input without any ANSI codes.
func TestEvaluate_PlainText(t *testing.T) {
	report := Evaluate("Hello World | 42% | main", "plain")
	if len(report.Dimensions) != 4 {
		t.Fatalf("expected 4 dimensions, got %d", len(report.Dimensions))
	}
}

// TestEvalContrast_HighContrast verifies that black-on-white ANSI text grades well.
func TestEvalContrast_HighContrast(t *testing.T) {
	// ESC[30m = ANSI black fg, ESC[47m = ANSI white bg
	ansiBlackOnWhite := "\x1b[30;47mHello\x1b[0m"
	result := evalContrast(Parse(ansiBlackOnWhite))

	if result.Name != "Contrast" {
		t.Errorf("expected dimension name 'Contrast', got %q", result.Name)
	}
	// Black on white should produce high contrast (A or B) on most palettes.
	if result.Grade == GradeF {
		t.Errorf("expected better than F for black-on-white, got %s; findings: %v", result.Grade, result.Findings)
	}
}

// TestEvalAdaptability_ANSI16 verifies that ANSI16 colors score well.
func TestEvalAdaptability_ANSI16(t *testing.T) {
	// All standard ANSI16 colors.
	ansi16Text := "\x1b[31mred\x1b[0m \x1b[32mgreen\x1b[0m \x1b[33myellow\x1b[0m"
	result := evalAdaptability(Parse(ansi16Text))

	if result.Grade == GradeD || result.Grade == GradeF {
		t.Errorf("expected A/B/C for all-ANSI16, got %s; findings: %v", result.Grade, result.Findings)
	}
}

// TestEvalHierarchy_ThreeTiers verifies that a mix of bold, normal, and faint
// segments produces an A-grade hierarchy.
func TestEvalHierarchy_ThreeTiers(t *testing.T) {
	// Tier 1: bold attribute
	// Tier 2: medium truecolor (lightness ~0.5, not bold, not faint) -> RGB(128,128,128)
	// Tier 3: faint attribute
	mixed := "\x1b[1;32mBold\x1b[0m \x1b[38;2;128;128;128mNormal\x1b[0m \x1b[2;32mFaint\x1b[0m"
	result := evalHierarchy(Parse(mixed))

	if result.Grade != GradeA {
		t.Errorf("expected A for three-tier mix, got %s; findings: %v", result.Grade, result.Findings)
	}
}

// TestFormatReport_Structure verifies that FormatReport produces the expected
// header, dimension lines, and footer.
func TestFormatReport_Structure(t *testing.T) {
	report := Report{
		Dimensions: []DimensionResult{
			{Name: "Contrast", Grade: GradeA, Findings: []string{"PASS everything"}},
			{Name: "Coherence", Grade: GradeB, Findings: []string{"3 lightness levels"}},
			{Name: "Hierarchy", Grade: GradeC, Findings: []string{"2 tiers"}},
			{Name: "Adaptability", Grade: GradeD, Findings: []string{"no ANSI16"}},
		},
		Overall: GradeB,
	}

	out := FormatReport(report)

	if !bytes.Contains([]byte(out), []byte("=== Statusline Design Evaluation ===")) {
		t.Error("missing header")
	}
	if !bytes.Contains([]byte(out), []byte("Overall: B")) {
		t.Error("missing overall grade")
	}
	if !bytes.Contains([]byte(out), []byte("Dimension 1: Contrast")) {
		t.Error("missing dimension 1 label")
	}
}

// TestGradeToNum_RoundTrip verifies the grade<->number conversion is consistent.
func TestGradeToNum_RoundTrip(t *testing.T) {
	cases := []struct {
		grade Grade
		num   int
	}{
		{GradeA, 4},
		{GradeB, 3},
		{GradeC, 2},
		{GradeD, 1},
		{GradeF, 0},
	}
	for _, tc := range cases {
		if got := gradeToNum(tc.grade); got != tc.num {
			t.Errorf("gradeToNum(%s) = %d, want %d", tc.grade, got, tc.num)
		}
		if got := numToGrade(tc.num); got != tc.grade {
			t.Errorf("numToGrade(%d) = %s, want %s", tc.num, got, tc.grade)
		}
	}
}
