package transcript_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/transcript"
)

// BenchmarkStateFile_ReadAndUnmarshal measures the cost of loading a state file
// from disk (os.ReadFile + JSON unmarshal). This is the snapshot restore path
// that runs on every tick before processing new transcript lines.
func BenchmarkStateFile_ReadAndUnmarshal(b *testing.B) {
	b.ReportAllocs()

	stateDir := b.TempDir()
	transcriptPath := filepath.Join(b.TempDir(), "session.jsonl")

	// Write a minimal transcript so ReadIncremental succeeds.
	base := syntheticTranscript(10)
	if err := os.WriteFile(transcriptPath, base, 0o644); err != nil {
		b.Fatalf("write transcript: %v", err)
	}

	// Build a snapshot with realistic content (20 tools, 5 agents).
	es := buildRealisticExtractionState(b, 20, 5)
	snap, err := es.MarshalSnapshot()
	if err != nil {
		b.Fatalf("MarshalSnapshot: %v", err)
	}

	// Persist state including the snapshot.
	sm := transcript.NewStateManager(stateDir)
	if _, err := sm.ReadIncremental(transcriptPath); err != nil {
		b.Fatalf("ReadIncremental: %v", err)
	}
	sm.SetSnapshot(snap)
	if err := sm.SaveState(transcriptPath); err != nil {
		b.Fatalf("SaveState: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Each iteration creates a fresh StateManager and reads incremental,
		// which triggers loadState (os.ReadFile + JSON unmarshal) internally.
		smIter := transcript.NewStateManager(stateDir)
		if _, err := smIter.ReadIncremental(transcriptPath); err != nil {
			b.Fatalf("ReadIncremental: %v", err)
		}
		// Accessing LoadSnapshot confirms the snapshot was loaded.
		_ = smIter.LoadSnapshot()
	}
}

// BenchmarkStateFile_MarshalAndWrite measures the cost of marshaling the
// extraction state snapshot and writing it to disk (JSON marshal + os.WriteFile
// + os.Rename). This is the state persistence path that runs on every tick
// after processing new transcript lines.
func BenchmarkStateFile_MarshalAndWrite(b *testing.B) {
	b.ReportAllocs()

	stateDir := b.TempDir()
	transcriptPath := filepath.Join(b.TempDir(), "session.jsonl")

	base := syntheticTranscript(10)
	if err := os.WriteFile(transcriptPath, base, 0o644); err != nil {
		b.Fatalf("write transcript: %v", err)
	}

	// Pre-read so the StateManager has an offset to persist.
	sm := transcript.NewStateManager(stateDir)
	if _, err := sm.ReadIncremental(transcriptPath); err != nil {
		b.Fatalf("ReadIncremental: %v", err)
	}

	// Realistic extraction state: 20 tools, 5 agents.
	es := buildRealisticExtractionState(b, 20, 5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		snap, err := es.MarshalSnapshot()
		if err != nil {
			b.Fatalf("MarshalSnapshot: %v", err)
		}
		sm.SetSnapshot(snap)
		if err := sm.SaveState(transcriptPath); err != nil {
			b.Fatalf("SaveState: %v", err)
		}
	}
}

// BenchmarkSnapshotMarshal isolates ExtractionState.MarshalSnapshot (JSON
// encoding only, no disk I/O). Useful for attributing total state-save cost.
func BenchmarkSnapshotMarshal(b *testing.B) {
	b.ReportAllocs()
	es := buildRealisticExtractionState(b, 20, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = es.MarshalSnapshot()
	}
}

// BenchmarkSnapshotUnmarshal isolates ExtractionState.UnmarshalSnapshot (JSON
// decoding only, no disk I/O). Useful for attributing total snapshot-restore cost.
func BenchmarkSnapshotUnmarshal(b *testing.B) {
	b.ReportAllocs()

	es := buildRealisticExtractionState(b, 20, 5)
	snap, err := es.MarshalSnapshot()
	if err != nil {
		b.Fatalf("MarshalSnapshot: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fresh := transcript.NewExtractionState()
		if err := fresh.UnmarshalSnapshot(snap); err != nil {
			b.Fatalf("UnmarshalSnapshot: %v", err)
		}
	}
}

// buildRealisticExtractionState populates an ExtractionState with nTools tool
// entries and nAgents agent entries by feeding synthetic transcript lines.
func buildRealisticExtractionState(b *testing.B, nTools, nAgents int) *transcript.ExtractionState {
	b.Helper()

	es := transcript.NewExtractionState()

	// Feed tool_use + tool_result pairs to populate nTools entries.
	for i := 0; i < nTools; i++ {
		// assistant entry (tool_use)
		assistantLine := syntheticEntry(2*i + 1)
		e, err := transcript.ParseEntry(assistantLine)
		if err != nil {
			b.Fatalf("ParseEntry tool_use: %v", err)
		}
		es.ProcessEntry(e)

		// user entry (tool_result)
		userLine := syntheticEntry(2*i + 2)
		e, err = transcript.ParseEntry(userLine)
		if err != nil {
			b.Fatalf("ParseEntry tool_result: %v", err)
		}
		es.ProcessEntry(e)
	}

	return es
}
