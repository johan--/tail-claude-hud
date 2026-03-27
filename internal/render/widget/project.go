package widget

import (
	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

// Project composes the Directory and Git widgets into a single segment
// without a separator between them.
// Format: '{directory} {branch}{dirty}{ahead}{behind}'
// e.g. 'tail-claude-hud main*' or 'tail-claude-hud feat/auth↑2'
// Returns an empty WidgetResult when both sub-widgets are empty.
// When Git has no data, renders directory only.
// When in a worktree, the git branch is hidden (the worktree widget shows
// "wt:<branch>"), so only the project directory name is rendered.
// FgColor is left empty because the sub-widgets compose multiple styles;
// the renderer passes the pre-styled Text through as-is.
func Project(ctx *model.RenderContext, cfg *config.Config) WidgetResult {
	dir := Directory(ctx, cfg)

	// In a worktree the branch is shown by the worktree widget — just show
	// the project directory so the user keeps their project identity.
	if ctx.WorktreeName != "" {
		return dir
	}

	if dir.IsEmpty() {
		return WidgetResult{}
	}

	git := Git(ctx, cfg)
	if git.IsEmpty() {
		return dir
	}

	return WidgetResult{
		Text:      dir.Text + " " + git.Text,
		PlainText: dir.PlainText + " " + git.PlainText,
		FgColor:   "13", // inherit directory's dominant color
	}
}
