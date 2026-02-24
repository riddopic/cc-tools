package hookcmd_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/hookcmd"
)

func TestParseInput(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantErr        bool
		errContains    string
		wantHookEvent  string
		wantSessionID  string
		wantToolName   string
		wantSource     string
		wantStopActive bool
	}{
		{
			name: "PreToolUse event parses correctly",
			input: `{
				"hook_event_name": "PreToolUse",
				"session_id": "sess-123",
				"transcript_path": "/tmp/transcript.json",
				"cwd": "/project",
				"permission_mode": "default",
				"tool_name": "Bash",
				"tool_input": {"command": "ls -la"},
				"tool_use_id": "tu-456"
			}`,
			wantErr:        false,
			errContains:    "",
			wantHookEvent:  "PreToolUse",
			wantSessionID:  "sess-123",
			wantToolName:   "Bash",
			wantSource:     "",
			wantStopActive: false,
		},
		{
			name: "SessionStart event parses correctly",
			input: `{
				"hook_event_name": "SessionStart",
				"session_id": "sess-789",
				"transcript_path": "/tmp/t.json",
				"cwd": "/home",
				"permission_mode": "plan",
				"source": "vscode"
			}`,
			wantErr:        false,
			errContains:    "",
			wantHookEvent:  "SessionStart",
			wantSessionID:  "sess-789",
			wantToolName:   "",
			wantSource:     "vscode",
			wantStopActive: false,
		},
		{
			name: "Stop event with stop_hook_active true",
			input: `{
				"hook_event_name": "Stop",
				"session_id": "sess-stop",
				"transcript_path": "/tmp/t.json",
				"cwd": "/project",
				"permission_mode": "default",
				"stop_hook_active": true
			}`,
			wantErr:        false,
			errContains:    "",
			wantHookEvent:  "Stop",
			wantSessionID:  "sess-stop",
			wantToolName:   "",
			wantSource:     "",
			wantStopActive: true,
		},
		{
			name: "lenient parsing ignores unknown fields",
			input: `{
				"hook_event_name": "PreToolUse",
				"session_id": "sess-lenient",
				"transcript_path": "/tmp/t.json",
				"cwd": "/project",
				"permission_mode": "default",
				"unknown_field": "should be ignored",
				"another_unknown": 42
			}`,
			wantErr:        false,
			errContains:    "",
			wantHookEvent:  "PreToolUse",
			wantSessionID:  "sess-lenient",
			wantToolName:   "",
			wantSource:     "",
			wantStopActive: false,
		},
		{
			name:           "empty input returns empty HookInput",
			input:          "",
			wantErr:        false,
			errContains:    "",
			wantHookEvent:  "",
			wantSessionID:  "",
			wantToolName:   "",
			wantSource:     "",
			wantStopActive: false,
		},
		{
			name:           "invalid JSON returns error",
			input:          `{invalid json}`,
			wantErr:        true,
			errContains:    "parsing hook input JSON",
			wantHookEvent:  "",
			wantSessionID:  "",
			wantToolName:   "",
			wantSource:     "",
			wantStopActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			got, err := hookcmd.ParseInput(reader)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.wantHookEvent, got.HookEventName)
			assert.Equal(t, tt.wantSessionID, got.SessionID)

			if tt.wantToolName != "" {
				assert.Equal(t, tt.wantToolName, got.ToolName)
			}
			if tt.wantSource != "" {
				assert.Equal(t, tt.wantSource, got.Source)
			}
			if tt.wantStopActive {
				assert.True(t, got.StopHookActive)
			}
		})
	}
}

func TestGetToolInputString(t *testing.T) {
	tests := []struct {
		name      string
		toolInput json.RawMessage
		key       string
		want      string
	}{
		{
			name:      "extracts string field correctly",
			toolInput: json.RawMessage(`{"command": "ls -la", "timeout": "30"}`),
			key:       "command",
			want:      "ls -la",
		},
		{
			name:      "returns empty for missing field",
			toolInput: json.RawMessage(`{"command": "ls"}`),
			key:       "nonexistent",
			want:      "",
		},
		{
			name:      "returns empty for nil ToolInput",
			toolInput: nil,
			key:       "command",
			want:      "",
		},
		{
			name:      "returns empty for non-string field",
			toolInput: json.RawMessage(`{"count": 42}`),
			key:       "count",
			want:      "",
		},
		{
			name:      "returns empty for empty JSON object",
			toolInput: json.RawMessage(`{}`),
			key:       "command",
			want:      "",
		},
		{
			name:      "returns empty for invalid JSON",
			toolInput: json.RawMessage(`{invalid`),
			key:       "command",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &hookcmd.HookInput{
				ToolInput: tt.toolInput,
			}
			got := input.GetToolInputString(tt.key)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHookInput_IsEditTool(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		want     bool
	}{
		{name: "Edit is an edit tool", toolName: "Edit", want: true},
		{name: "MultiEdit is an edit tool", toolName: "MultiEdit", want: true},
		{name: "Write is an edit tool", toolName: "Write", want: true},
		{name: "NotebookEdit is an edit tool", toolName: "NotebookEdit", want: true},
		{name: "Read is not an edit tool", toolName: "Read", want: false},
		{name: "Bash is not an edit tool", toolName: "Bash", want: false},
		{name: "empty tool name is not an edit tool", toolName: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &hookcmd.HookInput{ToolName: tt.toolName}
			assert.Equal(t, tt.want, input.IsEditTool())
		})
	}
}

func TestHookInput_GetFilePath(t *testing.T) {
	tests := []struct {
		name      string
		toolName  string
		toolInput json.RawMessage
		want      string
	}{
		{
			name:      "extracts file_path from Edit tool",
			toolName:  "Edit",
			toolInput: json.RawMessage(`{"file_path": "/tmp/main.go", "old_string": "foo"}`),
			want:      "/tmp/main.go",
		},
		{
			name:      "extracts notebook_path from NotebookEdit tool",
			toolName:  "NotebookEdit",
			toolInput: json.RawMessage(`{"notebook_path": "/tmp/notebook.ipynb", "cell_number": 0}`),
			want:      "/tmp/notebook.ipynb",
		},
		{
			name:      "returns empty for missing file_path",
			toolName:  "Edit",
			toolInput: json.RawMessage(`{"old_string": "foo"}`),
			want:      "",
		},
		{
			name:      "returns empty for empty tool_input",
			toolName:  "Edit",
			toolInput: nil,
			want:      "",
		},
		{
			name:      "returns empty for invalid JSON",
			toolName:  "Edit",
			toolInput: json.RawMessage(`{invalid`),
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &hookcmd.HookInput{
				ToolName:  tt.toolName,
				ToolInput: tt.toolInput,
			}
			assert.Equal(t, tt.want, input.GetFilePath())
		})
	}
}

func TestFileSafeSessionKey(t *testing.T) {
	// hashPrefix returns the first 16 hex chars of SHA-256 for a given input.
	hashPrefix := func(s string) string {
		h := sha256.Sum256([]byte(s))
		return hex.EncodeToString(h[:])[:16]
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "valid alphanumeric ID passes through",
			input: "abc123",
			want:  "abc123",
		},
		{
			name:  "valid UUID-like ID with hyphens passes through",
			input: "abc-123-def",
			want:  "abc-123-def",
		},
		{
			name:  "ID with dots and underscores passes through",
			input: "session_01.v2",
			want:  "session_01.v2",
		},
		{
			name:  "empty string returns empty",
			input: "",
			want:  "",
		},
		{
			name:  "ID with slash gets hashed",
			input: "sess/123",
			want:  hashPrefix("sess/123"),
		},
		{
			name:  "ID with path traversal gets hashed",
			input: "../etc/passwd",
			want:  hashPrefix("../etc/passwd"),
		},
		{
			name:  "ID with asterisk glob gets hashed",
			input: "hello*world",
			want:  hashPrefix("hello*world"),
		},
		{
			name:  "ID with question mark glob gets hashed",
			input: "what?",
			want:  hashPrefix("what?"),
		},
		{
			name:  "ID with bracket glob gets hashed",
			input: "foo[bar",
			want:  hashPrefix("foo[bar"),
		},
		{
			name:  "ID with null byte gets hashed",
			input: "hello\x00world",
			want:  hashPrefix("hello\x00world"),
		},
		{
			name:  "very long valid ID still passes through",
			input: strings.Repeat("a", 500),
			want:  strings.Repeat("a", 500),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hookcmd.FileSafeSessionKey(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
