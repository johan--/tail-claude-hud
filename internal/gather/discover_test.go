package gather

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

// writeSubagentFile creates a minimal subagent JSONL file with the given
// message content. Returns the path to the created file.
func writeSubagentFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	entry := map[string]interface{}{
		"type":        "user",
		"uuid":        "test-uuid",
		"timestamp":   time.Now().Format(time.RFC3339Nano),
		"isSidechain": true,
		"agentId":     name,
		"message": map[string]interface{}{
			"role":    "user",
			"content": content,
		},
	}
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal subagent entry: %v", err)
	}
	data = append(data, '\n')

	path := filepath.Join(dir, "agent-"+name+".jsonl")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write subagent file: %v", err)
	}
	return path
}

// setupSubagentsDir creates a session directory with a subagents/ subdirectory
// and returns the transcript path and subagents directory path.
func setupSubagentsDir(t *testing.T) (transcriptPath, subagentsDir string) {
	t.Helper()
	tmp := t.TempDir()
	sessionID := "test-session-id"
	transcriptPath = filepath.Join(tmp, sessionID+".jsonl")
	// Create the transcript file so it exists.
	if err := os.WriteFile(transcriptPath, []byte{}, 0o644); err != nil {
		t.Fatalf("write transcript: %v", err)
	}
	subagentsDir = filepath.Join(tmp, sessionID, "subagents")
	if err := os.MkdirAll(subagentsDir, 0o755); err != nil {
		t.Fatalf("mkdir subagents: %v", err)
	}
	return transcriptPath, subagentsDir
}

func TestDiscoverSubagents_NoSubagentsDir(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "no-session.jsonl")

	agents := discoverSubagents(path)
	if len(agents) != 0 {
		t.Errorf("expected 0 agents for missing dir, got %d", len(agents))
	}
}

func TestDiscoverSubagents_FiltersWarmupAgents(t *testing.T) {
	transcriptPath, subagentsDir := setupSubagentsDir(t)

	writeSubagentFile(t, subagentsDir, "a1b2c3d", "Warmup")
	writeSubagentFile(t, subagentsDir, "e4f5g6h", "Implement the feature")

	agents := discoverSubagents(transcriptPath)
	if len(agents) != 1 {
		t.Fatalf("expected 1 agent (warmup filtered), got %d", len(agents))
	}
	if agents[0].Name != "e4f5g6h" {
		t.Errorf("expected agent name 'e4f5g6h', got %q", agents[0].Name)
	}
}

func TestDiscoverSubagents_FiltersCompactAgents(t *testing.T) {
	transcriptPath, subagentsDir := setupSubagentsDir(t)

	writeSubagentFile(t, subagentsDir, "acompact123", "some content")
	writeSubagentFile(t, subagentsDir, "abc1234", "real task")

	agents := discoverSubagents(transcriptPath)
	if len(agents) != 1 {
		t.Fatalf("expected 1 agent (compact filtered), got %d", len(agents))
	}
	if agents[0].Name != "abc1234" {
		t.Errorf("expected agent name 'abc1234', got %q", agents[0].Name)
	}
}

func TestDiscoverSubagents_FiltersEmptyFiles(t *testing.T) {
	transcriptPath, subagentsDir := setupSubagentsDir(t)

	// Create an empty file.
	emptyPath := filepath.Join(subagentsDir, "agent-empty1.jsonl")
	if err := os.WriteFile(emptyPath, []byte{}, 0o644); err != nil {
		t.Fatalf("write empty file: %v", err)
	}

	writeSubagentFile(t, subagentsDir, "real1", "do the thing")

	agents := discoverSubagents(transcriptPath)
	if len(agents) != 1 {
		t.Fatalf("expected 1 agent (empty filtered), got %d", len(agents))
	}
}

func TestDiscoverSubagents_RunningVsCompleted(t *testing.T) {
	transcriptPath, subagentsDir := setupSubagentsDir(t)

	// Write a "recently modified" agent.
	writeSubagentFile(t, subagentsDir, "running1", "active task")

	// Write an "old" agent by backdating its modtime.
	oldPath := writeSubagentFile(t, subagentsDir, "done1", "finished task")
	oldTime := time.Now().Add(-5 * time.Minute)
	if err := os.Chtimes(oldPath, oldTime, oldTime); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	agents := discoverSubagents(transcriptPath)
	if len(agents) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(agents))
	}

	// Find agents by name.
	byName := make(map[string]model.AgentEntry, len(agents))
	for _, a := range agents {
		byName[a.Name] = a
	}

	running, ok := byName["running1"]
	if !ok {
		t.Fatal("missing agent 'running1'")
	}
	if running.Status != "running" {
		t.Errorf("expected running1 status 'running', got %q", running.Status)
	}

	done, ok := byName["done1"]
	if !ok {
		t.Fatal("missing agent 'done1'")
	}
	if done.Status != "completed" {
		t.Errorf("expected done1 status 'completed', got %q", done.Status)
	}
}

func TestDiscoverSubagents_IgnoresNonAgentFiles(t *testing.T) {
	transcriptPath, subagentsDir := setupSubagentsDir(t)

	// Non-agent files should be ignored.
	if err := os.WriteFile(filepath.Join(subagentsDir, "other.jsonl"), []byte("data\n"), 0o644); err != nil {
		t.Fatalf("write other file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subagentsDir, "agent-abc.txt"), []byte("data\n"), 0o644); err != nil {
		t.Fatalf("write txt file: %v", err)
	}

	writeSubagentFile(t, subagentsDir, "real1", "task content")

	agents := discoverSubagents(transcriptPath)
	if len(agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(agents))
	}
}

func TestDiscoverSubagents_ColorIndexWraps(t *testing.T) {
	transcriptPath, subagentsDir := setupSubagentsDir(t)

	// Create 10 agents to verify color index wraps at 8.
	for i := 0; i < 10; i++ {
		name := "agent" + string(rune('a'+i))
		writeSubagentFile(t, subagentsDir, name, "task")
	}

	agents := discoverSubagents(transcriptPath)
	if len(agents) != 10 {
		t.Fatalf("expected 10 agents, got %d", len(agents))
	}

	for i, a := range agents {
		expected := i % 8
		if a.ColorIndex != expected {
			t.Errorf("agent %d: expected ColorIndex %d, got %d", i, expected, a.ColorIndex)
		}
	}
}

func TestMergeSubagents_AppendsNew(t *testing.T) {
	td := &model.TranscriptData{
		Agents: []model.AgentEntry{
			{Name: "existing", Status: "running"},
		},
	}
	fsAgents := []model.AgentEntry{
		{Name: "new-agent", Status: "running"},
	}

	mergeSubagents(td, fsAgents)

	if len(td.Agents) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(td.Agents))
	}
	if td.Agents[1].Name != "new-agent" {
		t.Errorf("expected second agent 'new-agent', got %q", td.Agents[1].Name)
	}
}

func TestMergeSubagents_OverridesCompletedToRunning(t *testing.T) {
	td := &model.TranscriptData{
		Agents: []model.AgentEntry{
			{Name: "agent1", Status: "completed", DurationMs: 6},
		},
	}
	fsAgents := []model.AgentEntry{
		{Name: "agent1", Status: "running"},
	}

	mergeSubagents(td, fsAgents)

	if len(td.Agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(td.Agents))
	}
	if td.Agents[0].Status != "running" {
		t.Errorf("expected status 'running', got %q", td.Agents[0].Status)
	}
	if td.Agents[0].DurationMs != 0 {
		t.Errorf("expected DurationMs 0, got %d", td.Agents[0].DurationMs)
	}
}

func TestMergeSubagents_NoOverrideWhenBothCompleted(t *testing.T) {
	td := &model.TranscriptData{
		Agents: []model.AgentEntry{
			{Name: "agent1", Status: "completed", DurationMs: 5000},
		},
	}
	fsAgents := []model.AgentEntry{
		{Name: "agent1", Status: "completed"},
	}

	mergeSubagents(td, fsAgents)

	if td.Agents[0].DurationMs != 5000 {
		t.Errorf("expected DurationMs preserved at 5000, got %d", td.Agents[0].DurationMs)
	}
}

func TestMergeSubagents_Empty(t *testing.T) {
	td := &model.TranscriptData{
		Agents: []model.AgentEntry{
			{Name: "existing", Status: "running"},
		},
	}

	mergeSubagents(td, nil)

	if len(td.Agents) != 1 {
		t.Fatalf("expected 1 agent unchanged, got %d", len(td.Agents))
	}
}

func TestIsWarmupAgent_True(t *testing.T) {
	dir := t.TempDir()
	path := writeSubagentFile(t, dir, "warmup", "Warmup")
	if !isWarmupAgent(path) {
		t.Error("expected true for warmup agent")
	}
}

func TestIsWarmupAgent_False(t *testing.T) {
	dir := t.TempDir()
	path := writeSubagentFile(t, dir, "real", "Implement feature X")
	if isWarmupAgent(path) {
		t.Error("expected false for non-warmup agent")
	}
}

func TestIsWarmupAgent_MissingFile(t *testing.T) {
	if isWarmupAgent("/nonexistent/path.jsonl") {
		t.Error("expected false for missing file")
	}
}

func BenchmarkDiscoverSubagents(b *testing.B) {
	b.ReportAllocs()

	tmp := b.TempDir()
	sessionID := "bench-session"
	transcriptPath := filepath.Join(tmp, sessionID+".jsonl")
	if err := os.WriteFile(transcriptPath, []byte{}, 0o644); err != nil {
		b.Fatalf("write transcript: %v", err)
	}
	subagentsDir := filepath.Join(tmp, sessionID, "subagents")
	if err := os.MkdirAll(subagentsDir, 0o755); err != nil {
		b.Fatalf("mkdir: %v", err)
	}

	// Create 5 real agents and 3 warmup agents.
	for i := 0; i < 5; i++ {
		entry := map[string]interface{}{
			"type": "user", "uuid": "u",
			"timestamp": time.Now().Format(time.RFC3339Nano),
			"message":   map[string]interface{}{"role": "user", "content": "real task"},
		}
		data, _ := json.Marshal(entry)
		data = append(data, '\n')
		name := filepath.Join(subagentsDir, "agent-"+string(rune('a'+i))+".jsonl")
		if err := os.WriteFile(name, data, 0o644); err != nil {
			b.Fatalf("write: %v", err)
		}
	}
	for i := 0; i < 3; i++ {
		entry := map[string]interface{}{
			"type": "user", "uuid": "u",
			"timestamp": time.Now().Format(time.RFC3339Nano),
			"message":   map[string]interface{}{"role": "user", "content": "Warmup"},
		}
		data, _ := json.Marshal(entry)
		data = append(data, '\n')
		name := filepath.Join(subagentsDir, "agent-warmup"+string(rune('a'+i))+".jsonl")
		if err := os.WriteFile(name, data, 0o644); err != nil {
			b.Fatalf("write: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		discoverSubagents(transcriptPath)
	}
}
