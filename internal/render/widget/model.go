package widget

import (
	"fmt"
	"regexp"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

var cyanStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("87"))

// reBedrockDate matches Bedrock date suffixes like "-20250514".
var reBedrockDate = regexp.MustCompile(`-\d{8}$`)

// reBedrockVersion matches Bedrock version suffixes like "-v1:0".
var reBedrockVersion = regexp.MustCompile(`-v\d+:\d+$`)

// knownModelNames maps normalized Claude slugs to human-readable display names.
var knownModelNames = map[string]string{
	"claude-opus-4-5":    "Claude Opus 4.5",
	"claude-opus-4":      "Claude Opus 4",
	"claude-sonnet-4-5":  "Claude Sonnet 4.5",
	"claude-sonnet-4":    "Claude Sonnet 4",
	"claude-haiku-4":     "Claude Haiku 4",
	"claude-haiku-3-5":   "Claude Haiku 3.5",
	"claude-haiku-3":     "Claude Haiku 3",
	"claude-sonnet-3-7":  "Claude Sonnet 3.7",
	"claude-sonnet-3-5":  "Claude Sonnet 3.5",
	"claude-sonnet-3":    "Claude Sonnet 3",
	"claude-opus-3":      "Claude Opus 3",
}

// normalizeModelName cleans up a raw model ID (including Bedrock-style IDs)
// into a human-readable display name.
//
// Steps applied in order:
//  1. Strip "anthropic." prefix (Bedrock vendor namespace)
//  2. Strip date suffix like "-20250514"
//  3. Strip version suffix like "-v1:0"
//  4. Map to a known display name; fall back to the cleaned slug
func normalizeModelName(raw string) string {
	slug := raw

	// Strip Bedrock vendor prefix.
	slug = strings.TrimPrefix(slug, "anthropic.")

	// Strip date and version suffixes (apply repeatedly in case both are present).
	slug = reBedrockVersion.ReplaceAllString(slug, "")
	slug = reBedrockDate.ReplaceAllString(slug, "")

	if name, ok := knownModelNames[slug]; ok {
		return name
	}
	return slug
}

// Model renders the model display name in cyan, optionally suffixed with the
// context window size when cfg.Model.ShowContextSize is true.
// Returns "" when ctx.ModelDisplayName is empty.
//
// Raw Bedrock model IDs (e.g. "anthropic.claude-sonnet-4-20250514-v1:0") are
// normalized to a human-readable name before rendering.
func Model(ctx *model.RenderContext, cfg *config.Config) string {
	if ctx.ModelDisplayName == "" {
		return ""
	}

	name := normalizeModelName(ctx.ModelDisplayName)

	if cfg.Model.ShowContextSize && ctx.ContextWindowSize > 0 {
		name = fmt.Sprintf("%s (%s context)", name, formatTokens(ctx.ContextWindowSize))
	}

	return cyanStyle.Render(fmt.Sprintf("[%s]", name))
}

// formatTokens converts a raw token count to a human-readable string.
func formatTokens(n int) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.0fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%dk", n/1_000)
	}
	return fmt.Sprintf("%d", n)
}
