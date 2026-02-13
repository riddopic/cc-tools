package hooks_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/riddopic/cc-tools/internal/hooks"
)

func TestValidateWithSkipCheck_RealIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name         string
		setupSkip    func(t *testing.T, dir string)
		input        map[string]any
		debug        bool
		wantExitCode int
		wantInStderr []string
	}{
		{
			name: "no skip runs validation",
			setupSkip: func(_ *testing.T, _ string) {
				// No skip setup - should run validation
			},
			input: map[string]any{
				"hook_event_name": "PostToolUse",
				"tool_name":       "Edit",
				"tool_input": map[string]any{
					"file_path": "main.go",
				},
			},
			debug:        false,
			wantExitCode: 2,
			wantInStderr: nil,
		},
		{
			name: "skip both with debug messages",
			setupSkip: func(_ *testing.T, _ string) {
				// Create a skip registry file in test location
				// This would need actual registry setup
			},
			input: map[string]any{
				"hook_event_name": "PostToolUse",
				"tool_name":       "Edit",
				"tool_input": map[string]any{
					"file_path": "main.go",
				},
			},
			debug:        true,
			wantExitCode: 2,
			wantInStderr: []string{
				"Checking skips for project root",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			t.Chdir(tmpDir)

			if tt.setupSkip != nil {
				tt.setupSkip(t, tmpDir)
			}

			updateToolInputPathIntegration(tt.input, tmpDir)

			inputJSON, _ := json.Marshal(tt.input)
			stdin := bytes.NewReader(inputJSON)
			var stdout, stderr bytes.Buffer

			exitCode := hooks.ValidateWithSkipCheck(
				context.Background(),
				stdin, &stdout, &stderr,
				tt.debug, 5, 0,
			)

			assertExitCode(t, exitCode, tt.wantExitCode)
			assertStderrStringsIntegration(t, stderr.String(), tt.wantInStderr)
		})
	}
}

// updateToolInputPathIntegration updates the file_path in tool_input to use an absolute path.
func updateToolInputPathIntegration(input map[string]any, baseDir string) {
	toolInput, ok := input["tool_input"].(map[string]any)
	if !ok {
		return
	}
	filePath, fpOk := toolInput["file_path"].(string)
	if !fpOk {
		return
	}
	toolInput["file_path"] = filepath.Join(baseDir, filePath)
}

// assertStderrStringsIntegration checks that stderr contains all expected substrings.
func assertStderrStringsIntegration(t *testing.T, stderrStr string, expected []string) {
	t.Helper()
	for _, exp := range expected {
		if !strings.Contains(stderrStr, exp) {
			t.Errorf("Expected stderr to contain %q, got: %s", exp, stderrStr)
		}
	}
}

func TestCheckSkipsFromInput_Unit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		debug    bool
		wantLogs []string
	}{
		{
			name: "valid JSON with file path",
			input: `{
				"hook_event_name": "PostToolUse",
				"tool_name": "Edit",
				"tool_input": {
					"file_path": "/tmp/test.go"
				}
			}`,
			debug: true,
			wantLogs: []string{
				"Checking skips for project root: /tmp",
			},
		},
		{
			name:  "invalid JSON",
			input: `{invalid json}`,
			debug: true,
			wantLogs: []string{
				"Failed to parse JSON input",
			},
		},
		{
			name: "no file path",
			input: `{
				"hook_event_name": "PostToolUse",
				"tool_name": "Edit",
				"tool_input": {}
			}`,
			debug: true,
			wantLogs: []string{
				"No file path found in input",
			},
		},
		{
			name: "nested tool_input structure",
			input: `{
				"tool_input": {
					"file_path": "/home/user/project/src/main.go"
				}
			}`,
			debug: true,
			wantLogs: []string{
				"Checking skips for project root: /home/user/project/src",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var stderr hooks.MockOutputWriter

			hooks.CheckSkipsFromInputForTest(ctx, []byte(tt.input), tt.debug, &stderr)

			assertStderrStringsIntegration(t, stderr.String(), tt.wantLogs)
		})
	}
}

func TestValidateWithSkipCheck_ErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		stdin        io.Reader
		wantExitCode int
	}{
		{
			name:         "empty reader",
			stdin:        bytes.NewReader([]byte{}),
			wantExitCode: 0,
		},
		{
			name:         "nil bytes",
			stdin:        bytes.NewReader(nil),
			wantExitCode: 0,
		},
		{
			name:         "reader that returns error",
			stdin:        &errorReaderIntegration{err: context.DeadlineExceeded},
			wantExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			exitCode := hooks.ValidateWithSkipCheck(
				context.Background(),
				tt.stdin, &stdout, &stderr,
				false, 1, 0,
			)

			assertExitCode(t, exitCode, tt.wantExitCode)
		})
	}
}

// errorReaderIntegration is a reader that always returns an error.
type errorReaderIntegration struct {
	err error
}

func (r *errorReaderIntegration) Read(_ []byte) (int, error) {
	return 0, r.err
}
