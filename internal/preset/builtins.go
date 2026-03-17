package preset

import "github.com/kylesnowschwartz/tail-claude-hud/internal/config"

// builtins is the registry of named built-in presets.
var builtins = map[string]Preset{
	"default": {
		Name: "default",
		Lines: []config.Line{
			{Widgets: []string{"thinking", "model", "context", "project", "todos", "duration"}},
			{Widgets: []string{"agents"}},
			{Widgets: []string{"tools"}},
		},
		Separator:      " | ",
		Icons:          "nerdfont",
		Mode:           "plain",
		Theme:          "default",
		DirectoryStyle: "full",
	},
	"compact": {
		Name: "compact",
		Lines: []config.Line{
			{Widgets: []string{"model", "context", "cost", "duration"}},
		},
		Separator:      " | ",
		Icons:          "nerdfont",
		Mode:           "plain",
		Theme:          "default",
		DirectoryStyle: "basename",
	},
	"detailed": {
		Name: "detailed",
		Lines: []config.Line{
			{Widgets: []string{"model", "context", "cost", "duration", "speed"}},
			{Widgets: []string{"directory", "git", "lines", "outputstyle"}},
			{Widgets: []string{"agents", "messages", "skills"}},
			{Widgets: []string{"thinking", "tools"}},
		},
		Separator:      " | ",
		Icons:          "nerdfont",
		Mode:           "plain",
		Theme:          "default",
		DirectoryStyle: "fish",
	},
	"powerline": {
		Name: "powerline",
		Lines: []config.Line{
			{Widgets: []string{"model", "context", "git", "cost", "duration"}},
		},
		Separator:      "",
		Icons:          "nerdfont",
		Mode:           "powerline",
		Theme:          "dark",
		DirectoryStyle: "basename",
	},
	"capsule": {
		Name: "capsule",
		Lines: []config.Line{
			{Widgets: []string{"model", "context", "git", "cost", "duration"}},
		},
		Separator:      "",
		Icons:          "nerdfont",
		Mode:           "capsule",
		Theme:          "default",
		DirectoryStyle: "basename",
	},
	"minimal": {
		Name: "minimal",
		Lines: []config.Line{
			{Widgets: []string{"model", "context", "duration"}},
		},
		Separator:      " ",
		Icons:          "nerdfont",
		Mode:           "minimal",
		Theme:          "default",
		DirectoryStyle: "basename",
	},
}
