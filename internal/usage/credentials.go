// Package usage handles OAuth credential reading and the Anthropic usage API.
package usage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/logging"
)

const keychainBackoffSeconds = 60

// keychainRunner is a function type for running the security command.
// Using a function parameter makes keychain execution testable via injection.
type keychainRunner func(ctx context.Context) ([]byte, error)

// keychainJSON is the JSON structure returned by the macOS keychain for
// claude-api-credentials.
type keychainJSON struct {
	Token string `json:"token"`
}

// credentialsFileJSON is the structure of ~/.claude/.credentials.json.
type credentialsFileJSON struct {
	ClaudeAiOauth *struct {
		Token string `json:"token"`
	} `json:"claudeAiOauth"`
}

// GetToken returns an OAuth Bearer token string, or empty string if unavailable.
// Never returns an error — callers treat missing credentials as 'usage unavailable'.
//
// stateDir is the directory for state files (e.g. ~/.claude/plugins/tail-claude-hud/).
func GetToken(stateDir string) string {
	// Skip check: if ANTHROPIC_BASE_URL is set and does not contain "anthropic.com",
	// this is a Bedrock/Vertex/custom endpoint — OAuth usage API does not apply.
	if baseURL := strings.TrimSpace(os.Getenv("ANTHROPIC_BASE_URL")); baseURL != "" {
		if !strings.Contains(baseURL, "anthropic.com") {
			logging.Debug("credentials: skipping — ANTHROPIC_BASE_URL is non-Anthropic endpoint")
			return ""
		}
	}

	runner := defaultKeychainRunner(stateDir)
	return getToken(stateDir, runner)
}

// getToken is the internal implementation, accepting an injected keychainRunner
// to allow tests to override keychain execution.
func getToken(stateDir string, runner keychainRunner) string {
	// Try macOS keychain first.
	if runtime.GOOS == "darwin" {
		if token := readKeychainToken(stateDir, runner); token != "" {
			return token
		}
	}

	// Fall back to file-based credentials.
	return readFileToken()
}

// readKeychainToken reads an OAuth token from the macOS keychain.
// Returns "" on failure, applying and recording backoff state as needed.
func readKeychainToken(stateDir string, runner keychainRunner) string {
	backoffPath := filepath.Join(stateDir, ".keychain-backoff")

	// Check backoff: if a recent failure was recorded, skip the keychain call.
	if isInBackoff(backoffPath) {
		logging.Debug("credentials: keychain in backoff period, skipping")
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := runner(ctx)
	if err != nil {
		logging.Debug("credentials: keychain command failed: %v", err)
		writeBackoff(backoffPath)
		return ""
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		logging.Debug("credentials: keychain returned empty output")
		writeBackoff(backoffPath)
		return ""
	}

	var parsed keychainJSON
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		logging.Debug("credentials: keychain JSON parse failed: %v", err)
		writeBackoff(backoffPath)
		return ""
	}

	if parsed.Token == "" {
		logging.Debug("credentials: keychain JSON missing token field")
		writeBackoff(backoffPath)
		return ""
	}

	// Success — clear any existing backoff file.
	_ = os.Remove(backoffPath)
	return parsed.Token
}

// readFileToken reads an OAuth token from ~/.claude/.credentials.json.
func readFileToken() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	credPath := filepath.Join(home, ".claude", ".credentials.json")
	data, err := os.ReadFile(credPath)
	if err != nil {
		// File not found is normal for API users; don't log.
		return ""
	}

	var parsed credentialsFileJSON
	if err := json.Unmarshal(data, &parsed); err != nil {
		logging.Debug("credentials: .credentials.json parse failed: %v", err)
		return ""
	}

	if parsed.ClaudeAiOauth == nil || parsed.ClaudeAiOauth.Token == "" {
		return ""
	}

	return parsed.ClaudeAiOauth.Token
}

// isInBackoff returns true if the backoff file exists and was written within
// the last keychainBackoffSeconds seconds.
func isInBackoff(backoffPath string) bool {
	data, err := os.ReadFile(backoffPath)
	if err != nil {
		// File absent — not in backoff.
		return false
	}

	var ts int64
	if _, err := fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &ts); err != nil {
		// Unparseable — not in backoff.
		return false
	}

	elapsed := time.Now().Unix() - ts
	return elapsed >= 0 && elapsed < keychainBackoffSeconds
}

// writeBackoff writes the current Unix timestamp to the backoff file atomically.
func writeBackoff(backoffPath string) {
	if err := os.MkdirAll(filepath.Dir(backoffPath), 0o755); err != nil {
		return
	}

	ts := fmt.Sprintf("%d", time.Now().Unix())
	tmp := backoffPath + ".tmp"
	if err := os.WriteFile(tmp, []byte(ts), 0o644); err != nil {
		return
	}
	// Atomic rename: readers see either the old complete file or the new one.
	_ = os.Rename(tmp, backoffPath)
}

// defaultKeychainRunner returns a keychainRunner that calls
// /usr/bin/security find-generic-password -s 'claude-api-credentials' -w.
// Using an absolute path prevents PATH hijacking.
func defaultKeychainRunner(stateDir string) keychainRunner {
	_ = stateDir // kept for future service-name derivation if needed
	return func(ctx context.Context) ([]byte, error) {
		cmd := exec.CommandContext(
			ctx,
			"/usr/bin/security",
			"find-generic-password",
			"-s", "claude-api-credentials",
			"-w",
		)
		return cmd.Output()
	}
}
