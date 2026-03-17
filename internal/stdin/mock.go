package stdin

import (
	"os"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

// MockStdinData returns a realistic StdinData suitable for preview rendering.
// All top-level fields are populated so every widget has data to display.
//
// The context window is set to ~68% used (116k+12k+8k / 200k * 100),
// which falls in the normal green range and exercises the context widget.
//
// transcriptPath is used verbatim for TranscriptPath; Cwd is os.Getwd()
// falling back to "/tmp" on error.
func MockStdinData(transcriptPath string) *model.StdinData {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "/tmp"
	}

	usedPct := float64(68) // (116000+12000+8000) / 200000 * 100 = 68

	return &model.StdinData{
		TranscriptPath: transcriptPath,
		Cwd:            cwd,
		Model: &struct {
			ID          string `json:"id"`
			DisplayName string `json:"display_name"`
		}{
			ID:          "claude-sonnet-4-20250514",
			DisplayName: "Sonnet",
		},
		ContextWindow: &struct {
			Size         int      `json:"context_window_size"`
			UsedPercent  *float64 `json:"used_percentage"`
			CurrentUsage *struct {
				InputTokens              int `json:"input_tokens"`
				CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
				CacheReadInputTokens     int `json:"cache_read_input_tokens"`
			} `json:"current_usage"`
		}{
			Size:        200000,
			UsedPercent: &usedPct,
			CurrentUsage: &struct {
				InputTokens              int `json:"input_tokens"`
				CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
				CacheReadInputTokens     int `json:"cache_read_input_tokens"`
			}{
				InputTokens:              116000,
				CacheCreationInputTokens: 12000,
				CacheReadInputTokens:     8000,
			},
		},
		Cost: &model.Cost{
			TotalCostUSD:       2.47,
			TotalDurationMs:    425000,
			TotalAPIDurationMs: 38000,
			TotalLinesAdded:    187,
			TotalLinesRemoved:  42,
		},
		OutputStyle: &model.OutputStyle{
			Name: "concise",
		},
		ContextPercent: 68,
	}
}
