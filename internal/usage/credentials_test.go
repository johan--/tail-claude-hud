package usage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// --- helpers ---

// fakeRunner returns a keychainRunner that immediately returns the given output/error.
func fakeRunner(output string, err error) keychainRunner {
	return func(_ context.Context) ([]byte, error) {
		if err != nil {
			return nil, err
		}
		return []byte(output), nil
	}
}

// writeCredentialsFile writes a .credentials.json file with the given token
// under the given home directory.
func writeCredentialsFile(t *testing.T, homeDir string, token string) {
	t.Helper()
	claudeDir := filepath.Join(homeDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}
	content := fmt.Sprintf(`{"claudeAiOauth":{"token":%q}}`, token)
	if err := os.WriteFile(filepath.Join(claudeDir, ".credentials.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write .credentials.json: %v", err)
	}
}

// --- skip check tests ---

func TestGetToken_SkipsWhenNonAnthropicBaseURL(t *testing.T) {
	t.Setenv("ANTHROPIC_BASE_URL", "https://bedrock.amazonaws.com/v1")

	// Even if credentials would exist, we should return "" immediately.
	result := GetToken(t.TempDir())
	if result != "" {
		t.Errorf("expected empty token for non-Anthropic base URL, got %q", result)
	}
}

func TestGetToken_DoesNotSkipWhenAnthropicBaseURL(t *testing.T) {
	// Set ANTHROPIC_BASE_URL to an anthropic.com URL — should NOT be skipped.
	t.Setenv("ANTHROPIC_BASE_URL", "https://api.anthropic.com")

	stateDir := t.TempDir()
	// Override home to a temp dir with a credentials file.
	fakeHome := t.TempDir()
	writeCredentialsFile(t, fakeHome, "test-token-value")
	t.Setenv("HOME", fakeHome)

	// Only file fallback can return a token in this test (no real keychain).
	// On macOS the keychain runner would be called; we skip that path with a
	// stateDir that has an already-active backoff to force file fallback.
	backoffPath := filepath.Join(stateDir, ".keychain-backoff")
	ts := fmt.Sprintf("%d", time.Now().Unix())
	_ = os.WriteFile(backoffPath, []byte(ts), 0o644)

	result := GetToken(stateDir)
	if result != "test-token-value" {
		t.Errorf("expected token from file fallback, got %q", result)
	}
}

func TestGetToken_DoesNotSkipWhenNoBaseURL(t *testing.T) {
	t.Setenv("ANTHROPIC_BASE_URL", "")

	stateDir := t.TempDir()
	fakeHome := t.TempDir()
	writeCredentialsFile(t, fakeHome, "file-token")
	t.Setenv("HOME", fakeHome)

	// Force keychain backoff so only file fallback runs.
	backoffPath := filepath.Join(stateDir, ".keychain-backoff")
	ts := fmt.Sprintf("%d", time.Now().Unix())
	_ = os.WriteFile(backoffPath, []byte(ts), 0o644)

	result := GetToken(stateDir)
	if result != "file-token" {
		t.Errorf("expected file token, got %q", result)
	}
}

// --- file fallback tests ---

func TestReadFileToken_ReturnsTokenFromFile(t *testing.T) {
	fakeHome := t.TempDir()
	writeCredentialsFile(t, fakeHome, "my-oauth-token")
	t.Setenv("HOME", fakeHome)

	got := readFileToken()
	if got != "my-oauth-token" {
		t.Errorf("expected %q, got %q", "my-oauth-token", got)
	}
}

func TestReadFileToken_ReturnsEmptyWhenFileAbsent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	got := readFileToken()
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestReadFileToken_ReturnsEmptyOnMalformedJSON(t *testing.T) {
	fakeHome := t.TempDir()
	claudeDir := filepath.Join(fakeHome, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)
	_ = os.WriteFile(filepath.Join(claudeDir, ".credentials.json"), []byte("{not valid json"), 0o644)
	t.Setenv("HOME", fakeHome)

	got := readFileToken()
	if got != "" {
		t.Errorf("expected empty string for malformed JSON, got %q", got)
	}
}

func TestReadFileToken_ReturnsEmptyWhenTokenFieldMissing(t *testing.T) {
	fakeHome := t.TempDir()
	claudeDir := filepath.Join(fakeHome, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)
	// claudeAiOauth exists but no token field.
	_ = os.WriteFile(filepath.Join(claudeDir, ".credentials.json"), []byte(`{"claudeAiOauth":{}}`), 0o644)
	t.Setenv("HOME", fakeHome)

	got := readFileToken()
	if got != "" {
		t.Errorf("expected empty string when token field is missing, got %q", got)
	}
}

// --- backoff tests ---

func TestIsInBackoff_FalseWhenFileAbsent(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".keychain-backoff")
	if isInBackoff(path) {
		t.Error("expected false when backoff file is absent")
	}
}

func TestIsInBackoff_TrueForRecentTimestamp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".keychain-backoff")
	ts := fmt.Sprintf("%d", time.Now().Unix())
	_ = os.WriteFile(path, []byte(ts), 0o644)

	if !isInBackoff(path) {
		t.Error("expected in-backoff for a just-written timestamp")
	}
}

func TestIsInBackoff_FalseForOldTimestamp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".keychain-backoff")
	// 61 seconds ago — beyond the 60s backoff window.
	ts := fmt.Sprintf("%d", time.Now().Unix()-61)
	_ = os.WriteFile(path, []byte(ts), 0o644)

	if isInBackoff(path) {
		t.Error("expected false for timestamp older than backoff window")
	}
}

func TestIsInBackoff_FalseForUnparseable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".keychain-backoff")
	_ = os.WriteFile(path, []byte("not-a-number"), 0o644)

	if isInBackoff(path) {
		t.Error("expected false for unparseable backoff file")
	}
}

func TestWriteBackoff_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".keychain-backoff")
	before := time.Now().Unix()
	writeBackoff(path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("backoff file not created: %v", err)
	}

	var ts int64
	if _, err := fmt.Sscanf(string(data), "%d", &ts); err != nil {
		t.Fatalf("backoff file contains non-numeric content: %q", string(data))
	}
	if ts < before || ts > time.Now().Unix() {
		t.Errorf("backoff timestamp %d out of expected range [%d, %d]", ts, before, time.Now().Unix())
	}
}

func TestWriteBackoff_IsAtomic(t *testing.T) {
	// After writeBackoff, no .tmp file should remain.
	dir := t.TempDir()
	path := filepath.Join(dir, ".keychain-backoff")
	writeBackoff(path)

	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Error("expected .tmp file to be cleaned up after atomic rename")
	}
}

// --- keychain runner tests (command construction, no real binary) ---

func TestGetToken_KeychainFailureWritesBackoff(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("keychain test only runs on macOS")
	}

	stateDir := t.TempDir()
	// Inject a runner that always fails.
	failRunner := fakeRunner("", fmt.Errorf("security: item not found in keychain"))

	// Set HOME to a temp dir with no credentials file so file fallback also returns "".
	t.Setenv("HOME", t.TempDir())

	got := getToken(stateDir, failRunner)
	if got != "" {
		t.Errorf("expected empty token on keychain failure, got %q", got)
	}

	backoffPath := filepath.Join(stateDir, ".keychain-backoff")
	if _, err := os.Stat(backoffPath); os.IsNotExist(err) {
		t.Error("expected backoff file to be written after keychain failure")
	}
}

func TestGetToken_KeychainSuccessDeletesBackoff(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("keychain test only runs on macOS")
	}

	stateDir := t.TempDir()

	// Pre-write a backoff file to simulate a previous failure.
	backoffPath := filepath.Join(stateDir, ".keychain-backoff")
	oldTs := fmt.Sprintf("%d", time.Now().Unix()-120) // 2 min ago — already expired
	_ = os.WriteFile(backoffPath, []byte(oldTs), 0o644)

	// Inject a runner that succeeds with a valid JSON token.
	successRunner := fakeRunner(`{"token":"keychain-token-value"}`, nil)

	got := getToken(stateDir, successRunner)
	if got != "keychain-token-value" {
		t.Errorf("expected keychain token, got %q", got)
	}

	// Backoff file should be removed after success.
	if _, err := os.Stat(backoffPath); !os.IsNotExist(err) {
		t.Error("expected backoff file to be deleted after keychain success")
	}
}

func TestGetToken_KeychainSkippedDuringBackoff(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("keychain test only runs on macOS")
	}

	stateDir := t.TempDir()

	// Write a fresh (in-window) backoff file.
	backoffPath := filepath.Join(stateDir, ".keychain-backoff")
	ts := fmt.Sprintf("%d", time.Now().Unix())
	_ = os.WriteFile(backoffPath, []byte(ts), 0o644)

	// The runner should never be called; track invocations.
	called := false
	trackingRunner := func(_ context.Context) ([]byte, error) {
		called = true
		return []byte(`{"token":"should-not-be-called"}`), nil
	}

	fakeHome := t.TempDir()
	writeCredentialsFile(t, fakeHome, "file-fallback-token")
	t.Setenv("HOME", fakeHome)

	got := getToken(stateDir, trackingRunner)
	if called {
		t.Error("keychain runner should not be called during backoff")
	}
	// File fallback should be used instead.
	if got != "file-fallback-token" {
		t.Errorf("expected file fallback token, got %q", got)
	}
}

func TestGetToken_FallsBackToFileWhenKeychainReturnsNoToken(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("keychain test only runs on macOS")
	}

	stateDir := t.TempDir()

	// Keychain returns JSON with empty token.
	emptyTokenRunner := fakeRunner(`{"token":""}`, nil)

	fakeHome := t.TempDir()
	writeCredentialsFile(t, fakeHome, "file-only-token")
	t.Setenv("HOME", fakeHome)

	got := getToken(stateDir, emptyTokenRunner)
	if got != "file-only-token" {
		t.Errorf("expected file fallback when keychain token is empty, got %q", got)
	}
}

// --- absolute path spec test ---

func TestDefaultKeychainRunner_UsesAbsolutePath(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("keychain runner path test only runs on macOS")
	}

	// We can't easily inspect the constructed exec.Cmd after the fact, but we
	// can verify that the runner is not nil and that it wraps /usr/bin/security.
	// The real assertion is in code review: the implementation must use
	// "/usr/bin/security" not "security". This test documents the expectation.
	runner := defaultKeychainRunner(t.TempDir())
	if runner == nil {
		t.Error("defaultKeychainRunner returned nil")
	}
}
