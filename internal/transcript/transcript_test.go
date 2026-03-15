package transcript_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/transcript"
)

// --- ParseEntry ---

func TestParseEntry_BasicFields(t *testing.T) {
	line := []byte(`{
		"type": "assistant",
		"uuid": "abc-123",
		"timestamp": "2024-01-15T12:00:00Z",
		"message": {
			"role": "assistant",
			"model": "claude-opus-4-5",
			"stop_reason": "end_turn"
		}
	}`)

	e, err := transcript.ParseEntry(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Type != "assistant" {
		t.Errorf("Type = %q, want %q", e.Type, "assistant")
	}
	if e.UUID != "abc-123" {
		t.Errorf("UUID = %q, want %q", e.UUID, "abc-123")
	}
	if e.Message.Role != "assistant" {
		t.Errorf("Message.Role = %q, want %q", e.Message.Role, "assistant")
	}
	if e.Message.Model != "claude-opus-4-5" {
		t.Errorf("Message.Model = %q, want %q", e.Message.Model, "claude-opus-4-5")
	}
	if e.Message.StopReason == nil || *e.Message.StopReason != "end_turn" {
		t.Errorf("Message.StopReason = %v, want %q", e.Message.StopReason, "end_turn")
	}
}

func TestParseEntry_InvalidJSON(t *testing.T) {
	_, err := transcript.ParseEntry([]byte(`{not valid json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	// Must not panic — the test reaching here without panicking is the assertion.
}

func TestParseEntry_EmptyJSON(t *testing.T) {
	_, err := transcript.ParseEntry([]byte(`{}`))
	if err != nil {
		t.Fatalf("unexpected error for empty JSON object: %v", err)
	}
}

func TestParseEntry_CustomTitleEntry(t *testing.T) {
	line := []byte(`{"type":"custom-title","customTitle":"My Session","slug":"my-session"}`)
	e, err := transcript.ParseEntry(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Type != "custom-title" {
		t.Errorf("Type = %q, want %q", e.Type, "custom-title")
	}
	if e.CustomTitle != "My Session" {
		t.Errorf("CustomTitle = %q, want %q", e.CustomTitle, "My Session")
	}
	if e.Slug != "my-session" {
		t.Errorf("Slug = %q, want %q", e.Slug, "my-session")
	}
}

// --- ExtractContentBlocks: tool_use ---

func TestExtractContentBlocks_ToolUse(t *testing.T) {
	input := json.RawMessage(`{"key":"value"}`)
	line := buildEntryWithContent(t, "assistant", json.RawMessage(`[
		{"type":"tool_use","id":"tu-1","name":"Bash","input":{"key":"value"}}
	]`))

	e, err := transcript.ParseEntry(line)
	if err != nil {
		t.Fatalf("ParseEntry: %v", err)
	}

	blocks := transcript.ExtractContentBlocks(e)
	if len(blocks.ToolUse) != 1 {
		t.Fatalf("ToolUse count = %d, want 1", len(blocks.ToolUse))
	}
	tu := blocks.ToolUse[0]
	if tu.ID != "tu-1" {
		t.Errorf("ToolUse.ID = %q, want %q", tu.ID, "tu-1")
	}
	if tu.Name != "Bash" {
		t.Errorf("ToolUse.Name = %q, want %q", tu.Name, "Bash")
	}

	var gotInput map[string]string
	if err := json.Unmarshal(tu.Input, &gotInput); err != nil {
		t.Fatalf("unmarshal ToolUse.Input: %v", err)
	}
	var wantInput map[string]string
	if err := json.Unmarshal(input, &wantInput); err != nil {
		t.Fatalf("unmarshal expected input: %v", err)
	}
	if gotInput["key"] != wantInput["key"] {
		t.Errorf("ToolUse.Input[key] = %q, want %q", gotInput["key"], wantInput["key"])
	}
}

func TestExtractContentBlocks_MultipleToolUse(t *testing.T) {
	line := buildEntryWithContent(t, "assistant", json.RawMessage(`[
		{"type":"tool_use","id":"tu-1","name":"Read","input":{"file_path":"/a"}},
		{"type":"text","text":"some text"},
		{"type":"tool_use","id":"tu-2","name":"Write","input":{"file_path":"/b"}}
	]`))

	e, _ := transcript.ParseEntry(line)
	blocks := transcript.ExtractContentBlocks(e)

	if len(blocks.ToolUse) != 2 {
		t.Fatalf("ToolUse count = %d, want 2", len(blocks.ToolUse))
	}
	if blocks.ToolUse[0].Name != "Read" {
		t.Errorf("blocks.ToolUse[0].Name = %q, want Read", blocks.ToolUse[0].Name)
	}
	if blocks.ToolUse[1].Name != "Write" {
		t.Errorf("blocks.ToolUse[1].Name = %q, want Write", blocks.ToolUse[1].Name)
	}
}

// --- ExtractContentBlocks: tool_result ---

func TestExtractContentBlocks_ToolResult(t *testing.T) {
	line := buildEntryWithContent(t, "user", json.RawMessage(`[
		{"type":"tool_result","tool_use_id":"tu-1","content":"output text","is_error":false}
	]`))

	e, err := transcript.ParseEntry(line)
	if err != nil {
		t.Fatalf("ParseEntry: %v", err)
	}

	blocks := transcript.ExtractContentBlocks(e)
	if len(blocks.ToolResult) != 1 {
		t.Fatalf("ToolResult count = %d, want 1", len(blocks.ToolResult))
	}
	tr := blocks.ToolResult[0]
	if tr.ToolUseID != "tu-1" {
		t.Errorf("ToolResult.ToolUseID = %q, want %q", tr.ToolUseID, "tu-1")
	}
	if tr.IsError {
		t.Errorf("ToolResult.IsError = true, want false")
	}
}

func TestExtractContentBlocks_ToolResultIsError(t *testing.T) {
	line := buildEntryWithContent(t, "user", json.RawMessage(`[
		{"type":"tool_result","tool_use_id":"tu-err","content":"error output","is_error":true}
	]`))

	e, _ := transcript.ParseEntry(line)
	blocks := transcript.ExtractContentBlocks(e)

	if len(blocks.ToolResult) != 1 {
		t.Fatalf("ToolResult count = %d, want 1", len(blocks.ToolResult))
	}
	if !blocks.ToolResult[0].IsError {
		t.Error("ToolResult.IsError = false, want true")
	}
}

// --- ExtractContentBlocks: edge cases ---

func TestExtractContentBlocks_StringContent(t *testing.T) {
	// Plain string content — no blocks to extract.
	line := buildEntryWithContent(t, "user", json.RawMessage(`"plain text message"`))
	e, _ := transcript.ParseEntry(line)
	blocks := transcript.ExtractContentBlocks(e)

	if len(blocks.ToolUse) != 0 || len(blocks.ToolResult) != 0 {
		t.Errorf("expected empty blocks for string content, got %+v", blocks)
	}
}

func TestExtractContentBlocks_NoContent(t *testing.T) {
	line := []byte(`{"type":"user","uuid":"x","message":{"role":"user"}}`)
	e, _ := transcript.ParseEntry(line)
	blocks := transcript.ExtractContentBlocks(e)

	if len(blocks.ToolUse) != 0 || len(blocks.ToolResult) != 0 {
		t.Errorf("expected empty blocks for absent content, got %+v", blocks)
	}
}

// --- ParsedTimestamp ---

func TestParsedTimestamp_RFC3339Nano(t *testing.T) {
	line := []byte(`{"uuid":"x","timestamp":"2024-03-15T14:22:33.123456789Z"}`)
	e, _ := transcript.ParseEntry(line)
	ts := e.ParsedTimestamp()
	if ts.IsZero() {
		t.Fatal("expected non-zero timestamp for RFC3339Nano format")
	}
	if ts.Year() != 2024 || ts.Month() != time.March || ts.Day() != 15 {
		t.Errorf("unexpected date: %v", ts)
	}
}

func TestParsedTimestamp_RFC3339(t *testing.T) {
	line := []byte(`{"uuid":"x","timestamp":"2024-03-15T14:22:33Z"}`)
	e, _ := transcript.ParseEntry(line)
	ts := e.ParsedTimestamp()
	if ts.IsZero() {
		t.Fatal("expected non-zero timestamp for RFC3339 format")
	}
}

func TestParsedTimestamp_NoTimezone(t *testing.T) {
	// Claude sometimes emits timestamps without a timezone suffix.
	line := []byte(`{"uuid":"x","timestamp":"2024-03-15T14:22:33.123456789"}`)
	e, _ := transcript.ParseEntry(line)
	ts := e.ParsedTimestamp()
	if ts.IsZero() {
		t.Fatal("expected non-zero timestamp for no-timezone variant")
	}
	if ts.Hour() != 14 || ts.Minute() != 22 {
		t.Errorf("unexpected time: %v", ts)
	}
}

func TestParsedTimestamp_WithOffset(t *testing.T) {
	// Timezone offset variant, e.g. +12:00 (New Zealand).
	line := []byte(`{"uuid":"x","timestamp":"2024-03-15T14:22:33+12:00"}`)
	e, _ := transcript.ParseEntry(line)
	ts := e.ParsedTimestamp()
	if ts.IsZero() {
		t.Fatal("expected non-zero timestamp for offset timezone format")
	}
}

func TestParsedTimestamp_Missing(t *testing.T) {
	line := []byte(`{"uuid":"x"}`)
	e, _ := transcript.ParseEntry(line)
	ts := e.ParsedTimestamp()
	if !ts.IsZero() {
		t.Errorf("expected zero timestamp for missing field, got %v", ts)
	}
}

func TestParsedTimestamp_Invalid(t *testing.T) {
	line := []byte(`{"uuid":"x","timestamp":"not-a-timestamp"}`)
	e, _ := transcript.ParseEntry(line)
	ts := e.ParsedTimestamp()
	if !ts.IsZero() {
		t.Errorf("expected zero timestamp for invalid format, got %v", ts)
	}
}

// --- ParseTranscriptFile ---

func TestParseTranscriptFile_MultipleLines(t *testing.T) {
	data := []byte(
		`{"type":"user","uuid":"u1","timestamp":"2024-01-01T00:00:00Z"}` + "\n" +
			`{"type":"assistant","uuid":"u2","timestamp":"2024-01-01T00:00:01Z"}` + "\n" +
			`{"type":"custom-title","customTitle":"Test"}` + "\n",
	)

	entries := transcript.ParseTranscriptFile(data)
	if len(entries) != 3 {
		t.Fatalf("entry count = %d, want 3", len(entries))
	}
	if entries[0].UUID != "u1" {
		t.Errorf("entries[0].UUID = %q, want u1", entries[0].UUID)
	}
	if entries[2].CustomTitle != "Test" {
		t.Errorf("entries[2].CustomTitle = %q, want Test", entries[2].CustomTitle)
	}
}

func TestParseTranscriptFile_SkipsInvalidLines(t *testing.T) {
	data := []byte(
		`{"type":"user","uuid":"u1"}` + "\n" +
			`{invalid json}` + "\n" +
			`{"type":"assistant","uuid":"u2"}` + "\n",
	)

	entries := transcript.ParseTranscriptFile(data)
	// Invalid line is skipped; two valid lines remain.
	if len(entries) != 2 {
		t.Fatalf("entry count = %d, want 2 (invalid line skipped)", len(entries))
	}
}

func TestParseTranscriptFile_Empty(t *testing.T) {
	entries := transcript.ParseTranscriptFile([]byte{})
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for empty input, got %d", len(entries))
	}
}

// --- helpers ---

// buildEntryWithContent constructs a JSONL line with a given role and content.
func buildEntryWithContent(t *testing.T, role string, content json.RawMessage) []byte {
	t.Helper()
	type msg struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}
	type entry struct {
		Type    string `json:"type"`
		UUID    string `json:"uuid"`
		Message msg    `json:"message"`
	}
	e := entry{
		Type: role,
		UUID: "test-uuid",
		Message: msg{
			Role:    role,
			Content: content,
		},
	}
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("buildEntryWithContent: marshal failed: %v", err)
	}
	return b
}
