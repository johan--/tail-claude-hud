package widget

import (
	"strings"
	"testing"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

func TestSkillsWidget_NilTranscript_ReturnsEmpty(t *testing.T) {
	ctx := &model.RenderContext{}
	cfg := defaultCfg()

	if got := Skills(ctx, cfg); got != "" {
		t.Errorf("expected empty string for nil transcript, got %q", got)
	}
}

func TestSkillsWidget_NoSkills_ReturnsEmpty(t *testing.T) {
	ctx := &model.RenderContext{
		Transcript: &model.TranscriptData{},
	}
	cfg := defaultCfg()

	if got := Skills(ctx, cfg); got != "" {
		t.Errorf("expected empty string when no skills, got %q", got)
	}
}

func TestSkillsWidget_SingleSkill_DisplaysName(t *testing.T) {
	ctx := &model.RenderContext{
		Transcript: &model.TranscriptData{
			SkillNames: []string{"commit"},
		},
	}
	cfg := defaultCfg()

	got := Skills(ctx, cfg)
	if !strings.Contains(got, "commit") {
		t.Errorf("expected output to contain 'commit', got %q", got)
	}
}

func TestSkillsWidget_MultipleSkills_DisplaysNewestFirst(t *testing.T) {
	// SkillNames is ordered oldest-first; the widget reverses to newest-first
	// for display.
	ctx := &model.RenderContext{
		Transcript: &model.TranscriptData{
			SkillNames: []string{"commit", "review-pr", "lint"},
		},
	}
	cfg := defaultCfg()

	got := Skills(ctx, cfg)
	// All names should appear.
	for _, name := range []string{"commit", "review-pr", "lint"} {
		if !strings.Contains(got, name) {
			t.Errorf("expected output to contain %q, got %q", name, got)
		}
	}

	// "lint" is the newest (last in slice) so it should appear first in output.
	lintIdx := strings.Index(got, "lint")
	reviewIdx := strings.Index(got, "review-pr")
	if lintIdx > reviewIdx {
		t.Errorf("expected 'lint' (newest) to appear before 'review-pr', got %q", got)
	}
}

func TestSkillsWidget_DuplicateSkills_DeduplicatesNewestFirst(t *testing.T) {
	// "commit" appears twice; the widget should deduplicate and keep only the
	// newest occurrence (position closest to end of slice).
	ctx := &model.RenderContext{
		Transcript: &model.TranscriptData{
			SkillNames: []string{"commit", "lint", "commit"},
		},
	}
	cfg := defaultCfg()

	got := Skills(ctx, cfg)

	// "commit" should appear exactly once.
	count := strings.Count(got, "commit")
	if count != 1 {
		t.Errorf("expected 'commit' to appear once after dedup, got %d occurrences in %q", count, got)
	}
	// "commit" is the newest (last), so it should appear before "lint".
	commitIdx := strings.Index(got, "commit")
	lintIdx := strings.Index(got, "lint")
	if commitIdx > lintIdx {
		t.Errorf("expected newest 'commit' before 'lint', got %q", got)
	}
}
