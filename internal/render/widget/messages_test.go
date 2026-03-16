package widget

import (
	"strings"
	"testing"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

func TestMessagesWidget_NilTranscriptReturnsEmpty(t *testing.T) {
	ctx := &model.RenderContext{Transcript: nil}
	cfg := defaultCfg()

	if got := Messages(ctx, cfg); got != "" {
		t.Errorf("Messages with nil Transcript: expected empty, got %q", got)
	}
}

func TestMessagesWidget_ZeroCountReturnsEmpty(t *testing.T) {
	ctx := &model.RenderContext{Transcript: &model.TranscriptData{MessageCount: 0}}
	cfg := defaultCfg()

	if got := Messages(ctx, cfg); got != "" {
		t.Errorf("Messages with zero count: expected empty, got %q", got)
	}
}

func TestMessagesWidget_NonZeroCountRendersCount(t *testing.T) {
	ctx := &model.RenderContext{Transcript: &model.TranscriptData{MessageCount: 7}}
	cfg := defaultCfg()

	got := Messages(ctx, cfg)
	if !strings.Contains(got, "7") {
		t.Errorf("Messages: expected '7' in output, got %q", got)
	}
	if !strings.Contains(got, "msgs") {
		t.Errorf("Messages: expected 'msgs' in output, got %q", got)
	}
}

func TestMessagesWidget_ExactFormat(t *testing.T) {
	ctx := &model.RenderContext{Transcript: &model.TranscriptData{MessageCount: 3}}
	cfg := defaultCfg()

	got := Messages(ctx, cfg)
	want := dimStyle.Render("3 msgs")
	if got != want {
		t.Errorf("Messages: expected %q, got %q", want, got)
	}
}

func TestMessagesWidget_RegisteredInRegistry(t *testing.T) {
	if _, ok := Registry["messages"]; !ok {
		t.Error("Registry missing 'messages' widget")
	}
}
