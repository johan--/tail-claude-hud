//go:build !darwin

// Package sessions provides a no-op stub on non-macOS platforms.
// The sysctl-based process detection is macOS-specific.
package sessions

// AnyWaitingForPermission always returns false on non-macOS platforms.
func AnyWaitingForPermission(ownTranscript string) bool {
	return false
}
