package config

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFile writes content to path, creating parent directories as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

// TestCountEnv_EmptyCwd verifies that an empty cwd and empty home produces zero
// counts without errors.
func TestCountEnv_EmptyCwd(t *testing.T) {
	home := t.TempDir()
	counts := countEnvWithHome("", home)
	if counts == nil {
		t.Fatal("countEnvWithHome returned nil")
	}
	if counts.MCPServers != 0 {
		t.Errorf("MCPServers = %d, want 0", counts.MCPServers)
	}
	if counts.ClaudeMdFiles != 0 {
		t.Errorf("ClaudeMdFiles = %d, want 0", counts.ClaudeMdFiles)
	}
	if counts.RuleFiles != 0 {
		t.Errorf("RuleFiles = %d, want 0", counts.RuleFiles)
	}
	if counts.Hooks != 0 {
		t.Errorf("Hooks = %d, want 0", counts.Hooks)
	}
}

// TestCountEnv_ClaudeMdFiles verifies CLAUDE.md files are counted from the
// standard cwd locations as a separate ClaudeMdFiles category.
func TestCountEnv_ClaudeMdFiles(t *testing.T) {
	home := t.TempDir()
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "CLAUDE.md"), "# project instructions")
	writeFile(t, filepath.Join(dir, "CLAUDE.local.md"), "# local overrides")
	writeFile(t, filepath.Join(dir, ".claude", "CLAUDE.md"), "# claude dir")

	counts := countEnvWithHome(dir, home)
	if counts.ClaudeMdFiles != 3 {
		t.Errorf("ClaudeMdFiles = %d, want 3 (3 CLAUDE.md files)", counts.ClaudeMdFiles)
	}
	// Other categories must remain zero.
	if counts.RuleFiles != 0 {
		t.Errorf("RuleFiles = %d, want 0", counts.RuleFiles)
	}
	if counts.Hooks != 0 {
		t.Errorf("Hooks = %d, want 0", counts.Hooks)
	}
}

// TestCountEnv_HomeScopeClaudeMd verifies the home-scope CLAUDE.md is counted
// in ClaudeMdFiles.
func TestCountEnv_HomeScopeClaudeMd(t *testing.T) {
	home := t.TempDir()
	dir := t.TempDir()

	writeFile(t, filepath.Join(home, ".claude", "CLAUDE.md"), "# user instructions")

	counts := countEnvWithHome(dir, home)
	if counts.ClaudeMdFiles != 1 {
		t.Errorf("ClaudeMdFiles = %d, want 1 (home CLAUDE.md)", counts.ClaudeMdFiles)
	}
}

// TestCountEnv_RulesFiles verifies .md files in rules directories are counted
// in RuleFiles and not mixed into ClaudeMdFiles.
func TestCountEnv_RulesFiles(t *testing.T) {
	home := t.TempDir()
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, ".claude", "rules", "coding.md"), "# coding rules")
	writeFile(t, filepath.Join(dir, ".claude", "rules", "writing.md"), "# writing rules")
	writeFile(t, filepath.Join(dir, ".claude", "rules", "nested", "security.md"), "# security")

	counts := countEnvWithHome(dir, home)
	if counts.RuleFiles != 3 {
		t.Errorf("RuleFiles = %d, want 3 (3 rule files)", counts.RuleFiles)
	}
	if counts.ClaudeMdFiles != 0 {
		t.Errorf("ClaudeMdFiles = %d, want 0 (rules are not CLAUDE.md files)", counts.ClaudeMdFiles)
	}
}

// TestCountEnv_RulesIgnoresNonMd verifies non-.md files in rules dirs are skipped.
func TestCountEnv_RulesIgnoresNonMd(t *testing.T) {
	home := t.TempDir()
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, ".claude", "rules", "coding.md"), "# rule")
	writeFile(t, filepath.Join(dir, ".claude", "rules", "notes.txt"), "ignored")
	writeFile(t, filepath.Join(dir, ".claude", "rules", "config.json"), "{}")

	counts := countEnvWithHome(dir, home)
	if counts.RuleFiles != 1 {
		t.Errorf("RuleFiles = %d, want 1 (only .md files count)", counts.RuleFiles)
	}
}

// TestCountEnv_McpServersFromSettingsJson verifies MCP servers are counted from
// settings.json files.
func TestCountEnv_McpServersFromSettingsJson(t *testing.T) {
	home := t.TempDir()
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, ".claude", "settings.json"), `{
		"mcpServers": {
			"filesystem": {},
			"github": {}
		}
	}`)

	counts := countEnvWithHome(dir, home)
	if counts.MCPServers != 2 {
		t.Errorf("MCPServers = %d, want 2", counts.MCPServers)
	}
}

// TestCountEnv_McpServersFromMcpJson verifies MCP servers are counted from .mcp.json.
func TestCountEnv_McpServersFromMcpJson(t *testing.T) {
	home := t.TempDir()
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, ".mcp.json"), `{
		"mcpServers": {
			"local-tool": {},
			"dev-tool": {}
		}
	}`)

	counts := countEnvWithHome(dir, home)
	if counts.MCPServers != 2 {
		t.Errorf("MCPServers = %d, want 2", counts.MCPServers)
	}
}

// TestCountEnv_McpServersMergedAcrossFiles verifies that server names are
// deduplicated across settings files (same name in two files counts once).
func TestCountEnv_McpServersMergedAcrossFiles(t *testing.T) {
	home := t.TempDir()
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, ".claude", "settings.json"), `{
		"mcpServers": { "shared": {}, "project-only": {} }
	}`)
	writeFile(t, filepath.Join(dir, ".claude", "settings.local.json"), `{
		"mcpServers": { "shared": {}, "local-only": {} }
	}`)

	counts := countEnvWithHome(dir, home)
	// "shared" deduped, "project-only" and "local-only" each count once → 3 total
	if counts.MCPServers != 3 {
		t.Errorf("MCPServers = %d, want 3 (deduplicated across files)", counts.MCPServers)
	}
}

// TestCountEnv_McpServersDeduplicatedWithHome verifies home and cwd server names
// are also deduplicated against each other.
func TestCountEnv_McpServersDeduplicatedWithHome(t *testing.T) {
	home := t.TempDir()
	dir := t.TempDir()

	writeFile(t, filepath.Join(home, ".claude", "settings.json"), `{
		"mcpServers": { "shared": {}, "home-only": {} }
	}`)
	writeFile(t, filepath.Join(dir, ".claude", "settings.json"), `{
		"mcpServers": { "shared": {}, "project-only": {} }
	}`)

	counts := countEnvWithHome(dir, home)
	// shared (deduped), home-only, project-only = 3
	if counts.MCPServers != 3 {
		t.Errorf("MCPServers = %d, want 3", counts.MCPServers)
	}
}

// TestCountEnv_HooksCountedSeparately verifies that non-empty hooks arrays
// in settings.json are tracked in the Hooks field, not mixed with other counts.
func TestCountEnv_HooksCountedSeparately(t *testing.T) {
	home := t.TempDir()
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, ".claude", "settings.json"), `{
		"hooks": {
			"PreToolUse": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "lint"}]}],
			"PostToolUse": [{"matcher": "*", "hooks": [{"type": "command", "command": "notify"}]}],
			"Stop": []
		}
	}`)

	counts := countEnvWithHome(dir, home)
	// 2 non-empty hook arrays (Stop is empty so excluded)
	if counts.Hooks != 2 {
		t.Errorf("Hooks = %d, want 2 (2 non-empty hook arrays)", counts.Hooks)
	}
	// Hooks must not bleed into other categories.
	if counts.ClaudeMdFiles != 0 {
		t.Errorf("ClaudeMdFiles = %d, want 0", counts.ClaudeMdFiles)
	}
	if counts.RuleFiles != 0 {
		t.Errorf("RuleFiles = %d, want 0", counts.RuleFiles)
	}
}

// TestCountEnv_MissingFilesSkippedWithoutError verifies that a cwd with no
// config files at all produces zero counts and does not panic.
func TestCountEnv_MissingFilesSkippedWithoutError(t *testing.T) {
	home := t.TempDir()
	dir := t.TempDir()
	// dir and home exist but have no files inside

	counts := countEnvWithHome(dir, home)
	if counts == nil {
		t.Fatal("countEnvWithHome returned nil for empty dir")
	}
	if counts.MCPServers != 0 {
		t.Errorf("MCPServers = %d, want 0", counts.MCPServers)
	}
	if counts.ClaudeMdFiles != 0 {
		t.Errorf("ClaudeMdFiles = %d, want 0", counts.ClaudeMdFiles)
	}
	if counts.RuleFiles != 0 {
		t.Errorf("RuleFiles = %d, want 0", counts.RuleFiles)
	}
	if counts.Hooks != 0 {
		t.Errorf("Hooks = %d, want 0", counts.Hooks)
	}
}

// TestCountEnv_InvalidJsonSkipped verifies that an invalid JSON file is skipped
// without returning an error and without contributing to counts.
func TestCountEnv_InvalidJsonSkipped(t *testing.T) {
	home := t.TempDir()
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, ".claude", "settings.json"), `{ this is not valid json`)
	writeFile(t, filepath.Join(dir, ".mcp.json"), `not json at all`)

	counts := countEnvWithHome(dir, home)
	if counts.MCPServers != 0 {
		t.Errorf("MCPServers = %d, want 0 (invalid JSON skipped)", counts.MCPServers)
	}
}

// TestCountEnv_AllSourcesCombined verifies that all sources contribute correctly
// to the separate count categories when used together.
func TestCountEnv_AllSourcesCombined(t *testing.T) {
	home := t.TempDir()
	dir := t.TempDir()

	// Home CLAUDE.md: 1 ClaudeMdFile
	writeFile(t, filepath.Join(home, ".claude", "CLAUDE.md"), "# user")

	// Home rules: 1 RuleFile
	writeFile(t, filepath.Join(home, ".claude", "rules", "global.md"), "rule")

	// cwd CLAUDE.md files: 2 ClaudeMdFiles
	writeFile(t, filepath.Join(dir, "CLAUDE.md"), "# project")
	writeFile(t, filepath.Join(dir, "CLAUDE.local.md"), "# local")

	// cwd rule files: 2 RuleFiles
	writeFile(t, filepath.Join(dir, ".claude", "rules", "a.md"), "rule a")
	writeFile(t, filepath.Join(dir, ".claude", "rules", "b.md"), "rule b")

	// Hooks: 1 non-empty hook array
	writeFile(t, filepath.Join(dir, ".claude", "settings.json"), `{
		"mcpServers": { "alpha": {}, "beta": {} },
		"hooks": {
			"PreToolUse": [{"matcher": "Bash", "hooks": []}]
		}
	}`)

	// MCP from .mcp.json: 1 new, 1 duplicate of alpha
	writeFile(t, filepath.Join(dir, ".mcp.json"), `{
		"mcpServers": { "alpha": {}, "gamma": {} }
	}`)

	counts := countEnvWithHome(dir, home)

	// MCPServers: alpha (deduped), beta, gamma = 3
	if counts.MCPServers != 3 {
		t.Errorf("MCPServers = %d, want 3", counts.MCPServers)
	}

	// ClaudeMdFiles: 1 home + 2 cwd = 3
	if counts.ClaudeMdFiles != 3 {
		t.Errorf("ClaudeMdFiles = %d, want 3", counts.ClaudeMdFiles)
	}

	// RuleFiles: 1 home + 2 cwd = 3
	if counts.RuleFiles != 3 {
		t.Errorf("RuleFiles = %d, want 3", counts.RuleFiles)
	}

	// Hooks: 1 non-empty hook array
	if counts.Hooks != 1 {
		t.Errorf("Hooks = %d, want 1", counts.Hooks)
	}
}

// TestCountMdFilesRecursive verifies that countMdFilesRecursive handles nested
// directories correctly.
func TestCountMdFilesRecursive(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "a.md"), "")
	writeFile(t, filepath.Join(dir, "b.md"), "")
	writeFile(t, filepath.Join(dir, "sub", "c.md"), "")
	writeFile(t, filepath.Join(dir, "sub", "deep", "d.md"), "")
	writeFile(t, filepath.Join(dir, "ignored.txt"), "")

	got := countMdFilesRecursive(dir)
	if got != 4 {
		t.Errorf("countMdFilesRecursive = %d, want 4", got)
	}
}

// TestCountMdFilesRecursive_NonExistentDir verifies that a missing directory
// returns 0 without error.
func TestCountMdFilesRecursive_NonExistentDir(t *testing.T) {
	got := countMdFilesRecursive("/does/not/exist")
	if got != 0 {
		t.Errorf("countMdFilesRecursive = %d, want 0 for missing dir", got)
	}
}
