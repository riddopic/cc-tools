package observe_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/observe"
)

// verifyJSONLLines reads the observations file, checks the line count, and
// validates that each line round-trips correctly against the source events.
func verifyJSONLLines(t *testing.T, filePath string, events []observe.Event, wantLines int) {
	t.Helper()

	if wantLines == 0 {
		_, err := os.Stat(filePath)
		assert.True(t, os.IsNotExist(err), "file should not exist when disabled")

		return
	}

	data, err := os.ReadFile(filePath)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	assert.Len(t, lines, wantLines)

	for i, line := range lines {
		var parsed observe.Event
		require.NoError(t, json.Unmarshal([]byte(line), &parsed),
			"line %d should be valid JSON", i)
		assert.Equal(t, events[i].ToolName, parsed.ToolName)
		assert.Equal(t, events[i].Phase, parsed.Phase)
		assert.Equal(t, events[i].SessionID, parsed.SessionID)

		if events[i].ToolOutput != nil {
			assert.JSONEq(t, string(events[i].ToolOutput), string(parsed.ToolOutput),
				"line %d ToolOutput mismatch", i)
		} else {
			assert.Nil(t, parsed.ToolOutput, "line %d ToolOutput should be nil", i)
		}

		if events[i].Error != "" {
			assert.Equal(t, events[i].Error, parsed.Error,
				"line %d Error mismatch", i)
		} else {
			assert.Empty(t, parsed.Error, "line %d Error should be empty", i)
		}
	}
}

func TestRecord(t *testing.T) {
	fixedTime := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		setupDir  func(t *testing.T) string
		events    []observe.Event
		wantLines int
		wantErr   bool
	}{
		{
			name: "writes valid JSONL with correct fields",
			setupDir: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			events: []observe.Event{
				{
					Timestamp:  fixedTime,
					Phase:      "pre",
					ToolName:   "Bash",
					ToolInput:  json.RawMessage(`{"command":"ls"}`),
					ToolOutput: nil,
					Error:      "",
					SessionID:  "sess-001",
				},
			},
			wantLines: 1,
			wantErr:   false,
		},
		{
			name: "creates directory if not exists",
			setupDir: func(t *testing.T) string {
				t.Helper()
				return filepath.Join(t.TempDir(), "nested", "observe")
			},
			events: []observe.Event{
				{
					Timestamp:  fixedTime,
					Phase:      "post",
					ToolName:   "Edit",
					ToolInput:  nil,
					ToolOutput: nil,
					Error:      "",
					SessionID:  "sess-002",
				},
			},
			wantLines: 1,
			wantErr:   false,
		},
		{
			name: "skips when disabled file exists",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				require.NoError(t, os.WriteFile(
					filepath.Join(dir, ".disabled"), []byte(""), 0o600,
				))
				return dir
			},
			events: []observe.Event{
				{
					Timestamp:  fixedTime,
					Phase:      "pre",
					ToolName:   "Read",
					ToolInput:  nil,
					ToolOutput: nil,
					Error:      "",
					SessionID:  "sess-003",
				},
			},
			wantLines: 0,
			wantErr:   false,
		},
		{
			name: "appends multiple events as multi-line JSONL",
			setupDir: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			events: []observe.Event{
				{
					Timestamp:  fixedTime,
					Phase:      "pre",
					ToolName:   "Bash",
					ToolInput:  json.RawMessage(`{"command":"echo hello"}`),
					ToolOutput: nil,
					Error:      "",
					SessionID:  "sess-004",
				},
				{
					Timestamp:  fixedTime.Add(time.Second),
					Phase:      "post",
					ToolName:   "Bash",
					ToolInput:  json.RawMessage(`{"output":"hello"}`),
					ToolOutput: nil,
					Error:      "",
					SessionID:  "sess-004",
				},
				{
					Timestamp:  fixedTime.Add(2 * time.Second),
					Phase:      "pre",
					ToolName:   "Edit",
					ToolInput:  nil,
					ToolOutput: nil,
					Error:      "",
					SessionID:  "sess-004",
				},
			},
			wantLines: 3,
			wantErr:   false,
		},
		{
			name: "round-trips ToolOutput and Error fields through JSONL",
			setupDir: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			events: []observe.Event{
				{
					Timestamp:  fixedTime,
					Phase:      "post",
					ToolName:   "Bash",
					ToolInput:  json.RawMessage(`{"command":"ls"}`),
					ToolOutput: json.RawMessage(`{"stdout":"file1.go\nfile2.go"}`),
					Error:      "",
					SessionID:  "sess-005",
				},
				{
					Timestamp:  fixedTime.Add(time.Second),
					Phase:      "failure",
					ToolName:   "Bash",
					ToolInput:  json.RawMessage(`{"command":"rm /protected"}`),
					ToolOutput: nil,
					Error:      "permission denied",
					SessionID:  "sess-005",
				},
				{
					Timestamp:  fixedTime.Add(2 * time.Second),
					Phase:      "pre",
					ToolName:   "Read",
					ToolInput:  json.RawMessage(`{"file_path":"/tmp/test"}`),
					ToolOutput: nil,
					Error:      "",
					SessionID:  "sess-005",
				},
			},
			wantLines: 3,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupDir(t)
			obs := observe.NewObserver(dir, 10)

			for _, event := range tt.events {
				err := obs.Record(event)
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
			}

			verifyJSONLLines(t, filepath.Join(dir, "observations.jsonl"), tt.events, tt.wantLines)
		})
	}
}
