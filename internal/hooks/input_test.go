package hooks_test

import (
	"encoding/json"
	"testing"

	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/riddopic/cc-tools/internal/hooks"
)

// assertStringField asserts that a string field on HookInput matches the expected value.
func assertStringField(t *testing.T, got, want, fieldName string) {
	t.Helper()

	if got != want {
		t.Errorf("Expected %s %q, got %q", fieldName, want, got)
	}
}

func TestGetFilePath(t *testing.T) {
	tests := []struct {
		name       string
		input      *hookcmd.HookInput
		expectPath string
	}{
		{
			name: "Edit tool with file_path",
			input: &hookcmd.HookInput{
				ToolName: "Edit",
				ToolInput: hooks.MustMarshalJSON(map[string]any{
					"file_path": "/project/main.go",
				}),
			},
			expectPath: "/project/main.go",
		},
		{
			name: "MultiEdit tool with file_path",
			input: &hookcmd.HookInput{
				ToolName: "MultiEdit",
				ToolInput: hooks.MustMarshalJSON(map[string]any{
					"file_path": "/project/test.py",
				}),
			},
			expectPath: "/project/test.py",
		},
		{
			name: "Write tool with file_path",
			input: &hookcmd.HookInput{
				ToolName: "Write",
				ToolInput: hooks.MustMarshalJSON(map[string]any{
					"file_path": "/project/new.js",
					"content":   "console.log('hello');",
				}),
			},
			expectPath: "/project/new.js",
		},
		{
			name: "NotebookEdit tool with notebook_path",
			input: &hookcmd.HookInput{
				ToolName: "NotebookEdit",
				ToolInput: hooks.MustMarshalJSON(map[string]any{
					"notebook_path": "/project/analysis.ipynb",
					"cell_id":       "cell123",
				}),
			},
			expectPath: "/project/analysis.ipynb",
		},
		{
			name: "NotebookEdit with both paths prefers notebook_path",
			input: &hookcmd.HookInput{
				ToolName: "NotebookEdit",
				ToolInput: hooks.MustMarshalJSON(map[string]any{
					"notebook_path": "/project/notebook.ipynb",
					"file_path":     "/project/wrong.ipynb",
				}),
			},
			expectPath: "/project/notebook.ipynb",
		},
		{
			name: "nil tool input",
			input: &hookcmd.HookInput{
				ToolName:  "Edit",
				ToolInput: nil,
			},
			expectPath: "",
		},
		{
			name: "empty tool input",
			input: &hookcmd.HookInput{
				ToolName:  "Edit",
				ToolInput: hooks.MustMarshalJSON(map[string]any{}),
			},
			expectPath: "",
		},
		{
			name: "file_path is not a string",
			input: &hookcmd.HookInput{
				ToolName: "Edit",
				ToolInput: hooks.MustMarshalJSON(map[string]any{
					"file_path": 123,
				}),
			},
			expectPath: "",
		},
		{
			name: "file_path is null",
			input: &hookcmd.HookInput{
				ToolName: "Edit",
				ToolInput: hooks.MustMarshalJSON(map[string]any{
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
		input      *hookcmd.HookInput
		expectEdit bool
	}{
		{
			name:       "Edit is an edit tool",
			input:      &hookcmd.HookInput{ToolName: "Edit"},
			expectEdit: true,
		},
		{
			name:       "MultiEdit is an edit tool",
			input:      &hookcmd.HookInput{ToolName: "MultiEdit"},
			expectEdit: true,
		},
		{
			name:       "Write is an edit tool",
			input:      &hookcmd.HookInput{ToolName: "Write"},
			expectEdit: true,
		},
		{
			name:       "NotebookEdit is an edit tool",
			input:      &hookcmd.HookInput{ToolName: "NotebookEdit"},
			expectEdit: true,
		},
		{
			name:       "Bash is not an edit tool",
			input:      &hookcmd.HookInput{ToolName: "Bash"},
			expectEdit: false,
		},
		{
			name:       "Read is not an edit tool",
			input:      &hookcmd.HookInput{ToolName: "Read"},
			expectEdit: false,
		},
		{
			name:       "Grep is not an edit tool",
			input:      &hookcmd.HookInput{ToolName: "Grep"},
			expectEdit: false,
		},
		{
			name:       "empty tool name is not an edit tool",
			input:      &hookcmd.HookInput{ToolName: ""},
			expectEdit: false,
		},
		{
			name:       "case sensitive - edit is not Edit",
			input:      &hookcmd.HookInput{ToolName: "edit"},
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
		input := &hookcmd.HookInput{
			HookEventName:  "PostToolUse",
			SessionID:      "test123",
			TranscriptPath: "/path/transcript",
			Cwd:            "/project",
			ToolName:       "Edit",
			ToolInput: hooks.MustMarshalJSON(map[string]any{
				"file_path": "/file.go",
			}),
			ToolOutput: hooks.MustMarshalJSON(map[string]any{
				"success": true,
			}),
		}

		data, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}

		var decoded hookcmd.HookInput
		if unmarshalErr := json.Unmarshal(data, &decoded); unmarshalErr != nil {
			t.Fatalf("Failed to unmarshal: %v", unmarshalErr)
		}

		assertStringField(t, decoded.HookEventName, input.HookEventName, "HookEventName")
		assertStringField(t, decoded.SessionID, input.SessionID, "SessionID")
	})
}

func BenchmarkGetFilePath(b *testing.B) {
	input := &hookcmd.HookInput{
		ToolName: "Edit",
		ToolInput: hooks.MustMarshalJSON(map[string]any{
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
