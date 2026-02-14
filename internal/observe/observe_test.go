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
					Timestamp: fixedTime,
					Phase:     "pre",
					ToolName:  "Bash",
					ToolInput: json.RawMessage(`{"command":"ls"}`),
					SessionID: "sess-001",
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
					Timestamp: fixedTime,
					Phase:     "post",
					ToolName:  "Edit",
					ToolInput: nil,
					SessionID: "sess-002",
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
					Timestamp: fixedTime,
					Phase:     "pre",
					ToolName:  "Read",
					ToolInput: nil,
					SessionID: "sess-003",
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
					Timestamp: fixedTime,
					Phase:     "pre",
					ToolName:  "Bash",
					ToolInput: json.RawMessage(`{"command":"echo hello"}`),
					SessionID: "sess-004",
				},
				{
					Timestamp: fixedTime.Add(time.Second),
					Phase:     "post",
					ToolName:  "Bash",
					ToolInput: json.RawMessage(`{"output":"hello"}`),
					SessionID: "sess-004",
				},
				{
					Timestamp: fixedTime.Add(2 * time.Second),
					Phase:     "pre",
					ToolName:  "Edit",
					ToolInput: nil,
					SessionID: "sess-004",
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

			filePath := filepath.Join(dir, "observations.jsonl")

			if tt.wantLines == 0 {
				_, err := os.Stat(filePath)
				assert.True(t, os.IsNotExist(err), "file should not exist when disabled")
				return
			}

			data, err := os.ReadFile(filePath)
			require.NoError(t, err)

			lines := strings.Split(strings.TrimSpace(string(data)), "\n")
			assert.Len(t, lines, tt.wantLines)

			// Verify each line is valid JSON with expected fields.
			for i, line := range lines {
				var parsed observe.Event
				require.NoError(t, json.Unmarshal([]byte(line), &parsed),
					"line %d should be valid JSON", i)
				assert.Equal(t, tt.events[i].ToolName, parsed.ToolName)
				assert.Equal(t, tt.events[i].Phase, parsed.Phase)
				assert.Equal(t, tt.events[i].SessionID, parsed.SessionID)
			}
		})
	}
}
