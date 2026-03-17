// Package preset defines the Preset data model and built-in presets.
// A preset bundles visual configuration only: widget layout, separator,
// render mode, theme, icon set, and directory style. It does not include
// data-source settings such as thresholds, git options, or speed windows.
package preset

import (
	"sort"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
)

// Preset holds the visual configuration for a named layout.
type Preset struct {
	Name           string
	Lines          []config.Line
	Separator      string
	Icons          string
	Mode           string // plain, powerline, capsule, minimal
	Theme          string
	DirectoryStyle string
}

// Load returns the named built-in preset. Returns a zero-value Preset and
// false when name is not a known built-in.
func Load(name string) (Preset, bool) {
	p, ok := builtins[name]
	return p, ok
}

// BuiltinNames returns the names of all built-in presets in sorted order.
func BuiltinNames() []string {
	names := make([]string, 0, len(builtins))
	for name := range builtins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
