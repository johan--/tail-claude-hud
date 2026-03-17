package widget

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

var (
	gitBranchStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("87"))
	gitDimStyle    = lipgloss.NewStyle().Faint(true)
)

// Git renders branch name, dirty indicator, and optionally ahead/behind counts.
// Branch name is rendered in cyan. Dirty state uses the nerdfont dirty icon when
// cfg.Git.Dirty is true. Ahead/behind counts appear when cfg.Git.AheadBehind is true.
// Returns an empty WidgetResult when ctx.Git is nil.
// FgColor is left empty because the widget composes multiple styles internally;
// the renderer passes the pre-styled Text through as-is.
func Git(ctx *model.RenderContext, cfg *config.Config) WidgetResult {
	if ctx.Git == nil {
		return WidgetResult{}
	}

	icons := IconsFor(cfg.Style.Icons)
	g := ctx.Git

	var parts []string

	// Branch icon + name in cyan.
	branch := gitBranchStyle.Render(fmt.Sprintf("%s%s", icons.Branch, g.Branch))
	parts = append(parts, branch)

	// Dirty indicator (modified, staged, or untracked files).
	if cfg.Git.Dirty && g.IsDirty() {
		parts = append(parts, gitDimStyle.Render("*"))
	}

	// Ahead/behind counts.
	if cfg.Git.AheadBehind {
		if g.AheadBy > 0 {
			parts = append(parts, gitDimStyle.Render(fmt.Sprintf("↑%d", g.AheadBy)))
		}
		if g.BehindBy > 0 {
			parts = append(parts, gitDimStyle.Render(fmt.Sprintf("↓%d", g.BehindBy)))
		}
	}

	return WidgetResult{Text: strings.Join(parts, "")}
}
