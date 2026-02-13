package hooks

import (
	"encoding/json"
	"errors"
	"io"
	"testing"
)

func TestReadHookInput(t *testing.T) {
	t.Run("successful parsing of complete input", func(t *testing.T) {
		reader := &mockInputReader{
			isTerminalFunc: func() bool { return false },
			readAllFunc: func() ([]byte, error) {
				return []byte(`{
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
				}`), nil
			},
		}

		input, err := ReadHookInput(reader)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if input == nil {
			t.Fatal("Expected input, got nil")
		}
		if input.HookEventName != "PostToolUse" {
			t.Errorf("Expected HookEventName 'PostToolUse', got %s", input.HookEventName)
		}
		if input.SessionID != "session123" {
			t.Errorf("Expected SessionID 'session123', got %s", input.SessionID)
		}
		if input.ToolName != "Edit" {
			t.Errorf("Expected ToolName 'Edit', got %s", input.ToolName)
		}
		// Parse ToolInput to verify contents
		var toolInput map[string]any
		if unmarshalErr := json.Unmarshal(input.ToolInput, &toolInput); unmarshalErr != nil {
			t.Fatalf("Failed to unmarshal ToolInput: %v", unmarshalErr)
		}
		if toolInput["file_path"] != "/project/main.go" {
			t.Errorf("Expected file_path '/project/main.go', got %v", toolInput["file_path"])
		}
	})

	t.Run("returns error when terminal", func(t *testing.T) {
		reader := &mockInputReader{
			isTerminalFunc: func() bool { return true },
		}

		input, err := ReadHookInput(reader)
		if err == nil {
			t.Fatal("Expected error for terminal input")
		}
		if !errors.Is(err, ErrNoInput) {
			t.Errorf("Expected ErrNoInput, got %v", err)
		}
		if input != nil {
			t.Error("Expected nil input")
		}
	})

	t.Run("returns error on read failure", func(t *testing.T) {
		reader := &mockInputReader{
			isTerminalFunc: func() bool { return false },
			readAllFunc: func() ([]byte, error) {
				return nil, io.ErrUnexpectedEOF
			},
		}

		input, err := ReadHookInput(reader)
		if err == nil {
			t.Fatal("Expected error for read failure")
		}
		if input != nil {
			t.Error("Expected nil input")
		}
	})

	t.Run("returns error for empty input", func(t *testing.T) {
		reader := &mockInputReader{
			isTerminalFunc: func() bool { return false },
			readAllFunc: func() ([]byte, error) {
				return []byte{}, nil
			},
		}

		input, err := ReadHookInput(reader)
		if err == nil {
			t.Fatal("Expected error for empty input")
		}
		if !errors.Is(err, ErrNoInput) {
			t.Errorf("Expected ErrNoInput, got %v", err)
		}
		if input != nil {
			t.Error("Expected nil input")
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		reader := &mockInputReader{
			isTerminalFunc: func() bool { return false },
			readAllFunc: func() ([]byte, error) {
				return []byte(`{invalid json}`), nil
			},
		}

		input, err := ReadHookInput(reader)
		if err == nil {
			t.Fatal("Expected error for invalid JSON")
		}
		if input != nil {
			t.Error("Expected nil input")
		}
	})

	t.Run("handles minimal valid input", func(t *testing.T) {
		reader := &mockInputReader{
			isTerminalFunc: func() bool { return false },
			readAllFunc: func() ([]byte, error) {
				return []byte(`{"hook_event_name": "PreToolUse"}`), nil
			},
		}

		input, err := ReadHookInput(reader)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if input == nil {
			t.Fatal("Expected input")
		}
		if input.HookEventName != "PreToolUse" {
			t.Errorf("Expected HookEventName 'PreToolUse', got %s", input.HookEventName)
		}
	})

	t.Run("handles complex tool_input types", func(t *testing.T) {
		reader := &mockInputReader{
			isTerminalFunc: func() bool { return false },
			readAllFunc: func() ([]byte, error) {
				return []byte(`{
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
				}`), nil
			},
		}

		input, err := ReadHookInput(reader)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(input.ToolInput) == 0 {
			t.Fatal("Expected tool_input")
		}
		// Parse ToolInput to verify contents
		var toolInput map[string]any
		if unmarshalErr := json.Unmarshal(input.ToolInput, &toolInput); unmarshalErr != nil {
			t.Fatalf("Failed to unmarshal ToolInput: %v", unmarshalErr)
		}
		if toolInput["count"] != float64(42) { // JSON numbers are float64
			t.Errorf("Expected count 42, got %v", toolInput["count"])
		}
		if toolInput["enabled"] != true {
			t.Errorf("Expected enabled true, got %v", toolInput["enabled"])
		}
	})
}

func TestGetFilePath(t *testing.T) {
	tests := []struct {
		name       string
		input      *HookInput
		expectPath string
	}{
		{
			name: "Edit tool with file_path",
			input: &HookInput{
				ToolName: "Edit",
				ToolInput: mustMarshalJSON(map[string]any{
					"file_path": "/project/main.go",
				}),
			},
			expectPath: "/project/main.go",
		},
		{
			name: "MultiEdit tool with file_path",
			input: &HookInput{
				ToolName: "MultiEdit",
				ToolInput: mustMarshalJSON(map[string]any{
					"file_path": "/project/test.py",
				}),
			},
			expectPath: "/project/test.py",
		},
		{
			name: "Write tool with file_path",
			input: &HookInput{
				ToolName: "Write",
				ToolInput: mustMarshalJSON(map[string]any{
					"file_path": "/project/new.js",
					"content":   "console.log('hello');",
				}),
			},
			expectPath: "/project/new.js",
		},
		{
			name: "NotebookEdit tool with notebook_path",
			input: &HookInput{
				ToolName: "NotebookEdit",
				ToolInput: mustMarshalJSON(map[string]any{
					"notebook_path": "/project/analysis.ipynb",
					"cell_id":       "cell123",
				}),
			},
			expectPath: "/project/analysis.ipynb",
		},
		{
			name: "NotebookEdit with both paths prefers notebook_path",
			input: &HookInput{
				ToolName: "NotebookEdit",
				ToolInput: mustMarshalJSON(map[string]any{
					"notebook_path": "/project/notebook.ipynb",
					"file_path":     "/project/wrong.ipynb",
				}),
			},
			expectPath: "/project/notebook.ipynb",
		},
		{
			name: "nil tool input",
			input: &HookInput{
				ToolName:  "Edit",
				ToolInput: nil,
			},
			expectPath: "",
		},
		{
			name: "empty tool input",
			input: &HookInput{
				ToolName:  "Edit",
				ToolInput: mustMarshalJSON(map[string]any{}),
			},
			expectPath: "",
		},
		{
			name: "file_path is not a string",
			input: &HookInput{
				ToolName: "Edit",
				ToolInput: mustMarshalJSON(map[string]any{
					"file_path": 123, // number instead of string
				}),
			},
			expectPath: "",
		},
		{
			name: "file_path is null",
			input: &HookInput{
				ToolName: "Edit",
				ToolInput: mustMarshalJSON(map[string]any{
					"file_path": nil,
				}),
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
		input      *HookInput
		expectEdit bool
	}{
		{
			name:       "Edit is an edit tool",
			input:      &HookInput{ToolName: "Edit"},
			expectEdit: true,
		},
		{
			name:       "MultiEdit is an edit tool",
			input:      &HookInput{ToolName: "MultiEdit"},
			expectEdit: true,
		},
		{
			name:       "Write is an edit tool",
			input:      &HookInput{ToolName: "Write"},
			expectEdit: true,
		},
		{
			name:       "NotebookEdit is an edit tool",
			input:      &HookInput{ToolName: "NotebookEdit"},
			expectEdit: true,
		},
		{
			name:       "Bash is not an edit tool",
			input:      &HookInput{ToolName: "Bash"},
			expectEdit: false,
		},
		{
			name:       "Read is not an edit tool",
			input:      &HookInput{ToolName: "Read"},
			expectEdit: false,
		},
		{
			name:       "Grep is not an edit tool",
			input:      &HookInput{ToolName: "Grep"},
			expectEdit: false,
		},
		{
			name:       "empty tool name is not an edit tool",
			input:      &HookInput{ToolName: ""},
			expectEdit: false,
		},
		{
			name:       "case sensitive - edit is not Edit",
			input:      &HookInput{ToolName: "edit"},
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

func TestJSONMarshaling(t *testing.T) {
	t.Run("HookInput marshals correctly", func(t *testing.T) {
		input := &HookInput{
			HookEventName:  "PostToolUse",
			SessionID:      "test123",
			TranscriptPath: "/path/transcript",
			CWD:            "/project",
			ToolName:       "Edit",
			ToolInput: mustMarshalJSON(map[string]any{
				"file_path": "/file.go",
			}),
			ToolResponse: mustMarshalJSON(map[string]any{
				"success": true,
			}),
		}

		data, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}

		var decoded HookInput
		if unmarshalErr := json.Unmarshal(data, &decoded); unmarshalErr != nil {
			t.Fatalf("Failed to unmarshal: %v", unmarshalErr)
		}

		if decoded.HookEventName != input.HookEventName {
			t.Errorf("HookEventName mismatch: got %s, want %s", decoded.HookEventName, input.HookEventName)
		}
		if decoded.SessionID != input.SessionID {
			t.Errorf("SessionID mismatch: got %s, want %s", decoded.SessionID, input.SessionID)
		}
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

	reader := &mockInputReader{
		isTerminalFunc: func() bool { return false },
		readAllFunc: func() ([]byte, error) {
			return jsonData, nil
		},
	}

	b.ResetTimer()
	for range b.N {
		ReadHookInput(reader)
	}
}

func BenchmarkGetFilePath(b *testing.B) {
	input := &HookInput{
		ToolName: "Edit",
		ToolInput: mustMarshalJSON(map[string]any{
			"file_path":  "/project/main.go",
			"old_string": "foo",
			"new_string": "bar",
		}),
	}

	b.ResetTimer()
	for range b.N {
		_ = input.GetFilePath()
	}
}
