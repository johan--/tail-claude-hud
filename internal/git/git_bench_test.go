package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/git"
)

// BenchmarkGetStatus_GitRepo measures the wall-clock cost of git.GetStatus
// against a real (but minimal) git repository. This shells out to three git
// subprocesses: rev-parse, status --porcelain, and rev-list --left-right.
//
// OVER 10ms THRESHOLD: Measured ~55ms on Apple M3 Max (darwin/arm64). Each call
// forks up to three child processes. This is the dominant bottleneck in the
// Gather pipeline. Follow-up card needed to optimize: cache git status across
// ticks or reduce subprocess count (e.g., single `git status --branch --porcelain=v2`
// call instead of three separate invocations).
func BenchmarkGetStatus_GitRepo(b *testing.B) {
	b.ReportAllocs()

	// Create a minimal git repository in a temp directory.
	dir := b.TempDir()
	initGitRepo(b, dir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = git.GetStatus(dir)
	}
}

// BenchmarkGetStatus_GitRepo_WithChanges benchmarks git.GetStatus against a
// repository that has staged and unstaged changes — a more realistic scenario
// for an active working session.
func BenchmarkGetStatus_GitRepo_WithChanges(b *testing.B) {
	b.ReportAllocs()

	dir := b.TempDir()
	initGitRepo(b, dir)

	// Stage a file change so porcelain output is non-empty.
	writeFile(b, filepath.Join(dir, "staged.txt"), "staged content\n")
	gitCmd(b, dir, "add", "staged.txt")

	// Add an untracked file.
	writeFile(b, filepath.Join(dir, "untracked.txt"), "untracked\n")

	// Modify the original file in the worktree without staging.
	writeFile(b, filepath.Join(dir, "hello.txt"), "modified content\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = git.GetStatus(dir)
	}
}

// BenchmarkGetStatus_NonGitDir measures the fast-fail path: GetStatus called
// on a directory that is not a git repository. Only one subprocess is spawned
// (rev-parse fails immediately).
//
// OVER 10ms THRESHOLD: Measured ~17ms on Apple M3 Max (darwin/arm64). Even a
// single failing git subprocess costs ~17ms due to process spawn overhead.
// The optimization card for BenchmarkGetStatus_GitRepo should also cover this path.
func BenchmarkGetStatus_NonGitDir(b *testing.B) {
	b.ReportAllocs()

	dir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = git.GetStatus(dir)
	}
}

// initGitRepo initialises a minimal git repository at dir with a single commit.
func initGitRepo(b *testing.B, dir string) {
	b.Helper()

	gitCmd(b, dir, "init", "-b", "main")
	gitCmd(b, dir, "config", "user.email", "bench@example.com")
	gitCmd(b, dir, "config", "user.name", "Benchmarker")

	// Write a file and commit so HEAD is valid (rev-parse returns "main").
	writeFile(b, filepath.Join(dir, "hello.txt"), "hello\n")
	gitCmd(b, dir, "add", "hello.txt")
	gitCmd(b, dir, "commit", "-m", "init")
}

// gitCmd runs a git command in dir, failing the benchmark on error.
func gitCmd(b *testing.B, dir string, args ...string) {
	b.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		b.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// writeFile writes content to path, failing the benchmark on error.
func writeFile(b *testing.B, path, content string) {
	b.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		b.Fatalf("writeFile %s: %v", path, err)
	}
}
