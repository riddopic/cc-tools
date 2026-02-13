package hooks_test

import (
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/riddopic/cc-tools/internal/hooks"
)

// requireReadSuccess calls ReadHookInput and asserts it succeeds with a non-nil result.
func requireReadSuccess(t *testing.T, reader hooks.InputReader) *hooks.HookInput {
	t.Helper()

	input, err := hooks.ReadHookInput(reader)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if input == nil {
		t.Fatal("Expected input, got nil")
	}

	return input
}

// requireReadFailure calls ReadHookInput and asserts it fails with a nil result.
func requireReadFailure(t *testing.T, reader hooks.InputReader) error {
	t.Helper()

	input, err := hooks.ReadHookInput(reader)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if input != nil {
		t.Error("Expected nil input on failure")
	}

	return err
}

// assertStringField asserts that a string field on HookInput matches the expected value.
func assertStringField(t *testing.T, got, want, fieldName string) {
	t.Helper()

	if got != want {
		t.Errorf("Expected %s %q, got %q", fieldName, want, got)
	}
}

// assertToolInputField unmarshals ToolInput and checks that the given key matches the expected value.
func assertToolInputField(t *testing.T, raw json.RawMessage, key string, want any) {
	t.Helper()

	var toolInput map[string]any
	if err := json.Unmarshal(raw, &toolInput); err != nil {
		t.Fatalf("Failed to unmarshal ToolInput: %v", err)
	}

	if toolInput[key] != want {
		t.Errorf("Expected %s %v, got %v", key, want, toolInput[key])
	}
}

// newPipeReader creates a MockInputReader that provides the given data via a non-terminal stdin.
func newPipeReader(data []byte, err error) *hooks.MockInputReader {
	return &hooks.MockInputReader{
		IsTerminalFunc: func() bool { return false },
		ReadAllFunc: func() ([]byte, error) {
			return data, err
		},
	}
}

func TestReadHookInput(t *testing.T) {
	t.Run("successful parsing of complete input", func(t *testing.T) {
		reader := newPipeReader([]byte(`{
			"hook_event_name": "PostToolUse",
			"session_id": "session123",
			"transcript_path": "/path/to/transcript",
			"cwd": "/project",
			"tool_name": "Edit",
			"tool_input": {
				"file_path": "/project/main.go",
				"old_string": "foo",
				"new_string": "bar"
			},
			"tool_response": {
				"success": true
			}
		}`), nil)

		input := requireReadSuccess(t, reader)
		assertStringField(t, input.HookEventName, "PostToolUse", "HookEventName")
		assertStringField(t, input.SessionID, "session123", "SessionID")
		assertStringField(t, input.ToolName, "Edit", "ToolName")
		assertToolInputField(t, input.ToolInput, "file_path", "/project/main.go")
	})

	t.Run("returns error when terminal", func(t *testing.T) {
		reader := &hooks.MockInputReader{
			IsTerminalFunc: func() bool { return true },
			ReadAllFunc:    nil,
		}

		err := requireReadFailure(t, reader)
		if !errors.Is(err, hooks.ErrNoInput) {
			t.Errorf("Expected ErrNoInput, got %v", err)
		}
	})

	t.Run("returns error on read failure", func(t *testing.T) {
		reader := newPipeReader(nil, io.ErrUnexpectedEOF)
		requireReadFailure(t, reader)
	})

	t.Run("returns error for empty input", func(t *testing.T) {
		reader := newPipeReader([]byte{}, nil)

		err := requireReadFailure(t, reader)
		if !errors.Is(err, hooks.ErrNoInput) {
			t.Errorf("Expected ErrNoInput, got %v", err)
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		reader := newPipeReader([]byte(`{invalid json}`), nil)
		requireReadFailure(t, reader)
	})

	t.Run("handles minimal valid input", func(t *testing.T) {
		reader := newPipeReader([]byte(`{"hook_event_name": "PreToolUse"}`), nil)

		input := requireReadSuccess(t, reader)
		assertStringField(t, input.HookEventName, "PreToolUse", "HookEventName")
	})

	t.Run("handles complex tool_input types", func(t *testing.T) {
		reader := newPipeReader([]byte(`{
			"hook_event_name": "PostToolUse",
			"tool_name": "MultiEdit",
			"tool_input": {
				"file_path": "/project/main.go",
				"edits": [
					{"old": "foo", "new": "bar"},
					{"old": "baz", "new": "qux"}
				],
				"count": 42,
				"enabled": true
			}
		}`), nil)

		input := requireReadSuccess(t, reader)
		if len(input.ToolInput) == 0 {
			t.Fatal("Expected tool_input")
		}
		assertToolInputField(t, input.ToolInput, "count", float64(42))
		assertToolInputField(t, input.ToolInput, "enabled", true)
	})
}

func TestGetFilePath(t *testing.T) {
	tests := []struct {
		name       string
		input      *hooks.HookInput
		expectPath string
	}{
		{
			name: "Edit tool with file_path",
			input: &hooks.HookInput{
				HookEventName:  "",
				SessionID:      "",
				TranscriptPath: "",
				CWD:            "",
				ToolName:       "Edit",
				ToolInput: hooks.MustMarshalJSON(map[string]any{
					"file_path": "/project/main.go",
				}),
				ToolResponse: nil,
			},
			expectPath: "/project/main.go",
		},
		{
			name: "MultiEdit tool with file_path",
			input: &hooks.HookInput{
				HookEventName:  "",
				SessionID:      "",
				TranscriptPath: "",
				CWD:            "",
				ToolName:       "MultiEdit",
				ToolInput: hooks.MustMarshalJSON(map[string]any{
					"file_path": "/project/test.py",
				}),
				ToolResponse: nil,
			},
			expectPath: "/project/test.py",
		},
		{
			name: "Write tool with file_path",
			input: &hooks.HookInput{
				HookEventName:  "",
				SessionID:      "",
				TranscriptPath: "",
				CWD:            "",
				ToolName:       "Write",
				ToolInput: hooks.MustMarshalJSON(map[string]any{
					"file_path": "/project/new.js",
					"content":   "console.log('hello');",
				}),
				ToolResponse: nil,
			},
			expectPath: "/project/new.js",
		},
		{
			name: "NotebookEdit tool with notebook_path",
			input: &hooks.HookInput{
				HookEventName:  "",
				SessionID:      "",
				TranscriptPath: "",
				CWD:            "",
				ToolName:       "NotebookEdit",
				ToolInput: hooks.MustMarshalJSON(map[string]any{
					"notebook_path": "/project/analysis.ipynb",
					"cell_id":       "cell123",
				}),
				ToolResponse: nil,
			},
			expectPath: "/project/analysis.ipynb",
		},
		{
			name: "NotebookEdit with both paths prefers notebook_path",
			input: &hooks.HookInput{
				HookEventName:  "",
				SessionID:      "",
				TranscriptPath: "",
				CWD:            "",
				ToolName:       "NotebookEdit",
				ToolInput: hooks.MustMarshalJSON(map[string]any{
					"notebook_path": "/project/notebook.ipynb",
					"file_path":     "/project/wrong.ipynb",
				}),
				ToolResponse: nil,
			},
			expectPath: "/project/notebook.ipynb",
		},
		{
			name: "nil tool input",
			input: &hooks.HookInput{
				HookEventName:  "",
				SessionID:      "",
				TranscriptPath: "",
				CWD:            "",
				ToolName:       "Edit",
				ToolInput:      nil,
				ToolResponse:   nil,
			},
			expectPath: "",
		},
		{
			name: "empty tool input",
			input: &hooks.HookInput{
				HookEventName:  "",
				SessionID:      "",
				TranscriptPath: "",
				CWD:            "",
				ToolName:       "Edit",
				ToolInput:      hooks.MustMarshalJSON(map[string]any{}),
				ToolResponse:   nil,
			},
			expectPath: "",
		},
		{
			name: "file_path is not a string",
			input: &hooks.HookInput{
				HookEventName:  "",
				SessionID:      "",
				TranscriptPath: "",
				CWD:            "",
				ToolName:       "Edit",
				ToolInput: hooks.MustMarshalJSON(map[string]any{
					"file_path": 123, // number instead of string
				}),
				ToolResponse: nil,
			},
			expectPath: "",
		},
		{
			name: "file_path is null",
			input: &hooks.HookInput{
				HookEventName:  "",
				SessionID:      "",
				TranscriptPath: "",
				CWD:            "",
				ToolName:       "Edit",
				ToolInput: hooks.MustMarshalJSON(map[string]any{
					"file_path": nil,
				}),
				ToolResponse: nil,
			},
			expectPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.input.GetFilePath()
			if path != tt.expectPath {
				t.Errorf("GetFilePath() = %v, want %v", path, tt.expectPath)
			}
		})
	}
}

func TestIsEditTool(t *testing.T) {
	tests := []struct {
		name       string
		input      *hooks.HookInput
		expectEdit bool
	}{
		{
			name:       "Edit is an edit tool",
			input:      newTestHookInputMinimal("Edit"),
			expectEdit: true,
		},
		{
			name:       "MultiEdit is an edit tool",
			input:      newTestHookInputMinimal("MultiEdit"),
			expectEdit: true,
		},
		{
			name:       "Write is an edit tool",
			input:      newTestHookInputMinimal("Write"),
			expectEdit: true,
		},
		{
			name:       "NotebookEdit is an edit tool",
			input:      newTestHookInputMinimal("NotebookEdit"),
			expectEdit: true,
		},
		{
			name:       "Bash is not an edit tool",
			input:      newTestHookInputMinimal("Bash"),
			expectEdit: false,
		},
		{
			name:       "Read is not an edit tool",
			input:      newTestHookInputMinimal("Read"),
			expectEdit: false,
		},
		{
			name:       "Grep is not an edit tool",
			input:      newTestHookInputMinimal("Grep"),
			expectEdit: false,
		},
		{
			name:       "empty tool name is not an edit tool",
			input:      newTestHookInputMinimal(""),
			expectEdit: false,
		},
		{
			name:       "case sensitive - edit is not Edit",
			input:      newTestHookInputMinimal("edit"),
			expectEdit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEdit := tt.input.IsEditTool()
			if isEdit != tt.expectEdit {
				t.Errorf("IsEditTool() = %v, want %v", isEdit, tt.expectEdit)
			}
		})
	}
}

// newTestHookInputMinimal creates a HookInput with only ToolName set and all other fields at zero.
func newTestHookInputMinimal(toolName string) *hooks.HookInput {
	return &hooks.HookInput{
		HookEventName:  "",
		SessionID:      "",
		TranscriptPath: "",
		CWD:            "",
		ToolName:       toolName,
		ToolInput:      nil,
		ToolResponse:   nil,
	}
}

func TestJSONMarshaling(t *testing.T) {
	t.Run("HookInput marshals correctly", func(t *testing.T) {
		input := &hooks.HookInput{
			HookEventName:  "PostToolUse",
			SessionID:      "test123",
			TranscriptPath: "/path/transcript",
			CWD:            "/project",
			ToolName:       "Edit",
			ToolInput: hooks.MustMarshalJSON(map[string]any{
				"file_path": "/file.go",
			}),
			ToolResponse: hooks.MustMarshalJSON(map[string]any{
				"success": true,
			}),
		}

		data, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}

		var decoded hooks.HookInput
		if unmarshalErr := json.Unmarshal(data, &decoded); unmarshalErr != nil {
			t.Fatalf("Failed to unmarshal: %v", unmarshalErr)
		}

		assertStringField(t, decoded.HookEventName, input.HookEventName, "HookEventName")
		assertStringField(t, decoded.SessionID, input.SessionID, "SessionID")
	})
}

func BenchmarkReadHookInput(b *testing.B) {
	jsonData := []byte(`{
		"hook_event_name": "PostToolUse",
		"session_id": "session123",
		"transcript_path": "/path/to/transcript",
		"cwd": "/project",
		"tool_name": "Edit",
		"tool_input": {
			"file_path": "/project/main.go",
			"old_string": "foo",
			"new_string": "bar"
		}
	}`)

	reader := &hooks.MockInputReader{
		IsTerminalFunc: func() bool { return false },
		ReadAllFunc: func() ([]byte, error) {
			return jsonData, nil
		},
	}

	b.ResetTimer()
	for range b.N {
		hooks.ReadHookInput(reader)
	}
}

func BenchmarkGetFilePath(b *testing.B) {
	input := &hooks.HookInput{
		HookEventName:  "",
		SessionID:      "",
		TranscriptPath: "",
		CWD:            "",
		ToolName:       "Edit",
		ToolInput: hooks.MustMarshalJSON(map[string]any{
			"file_path":  "/project/main.go",
			"old_string": "foo",
			"new_string": "bar",
		}),
		ToolResponse: nil,
	}

	b.ResetTimer()
	for range b.N {
		_ = input.GetFilePath()
	}
}
