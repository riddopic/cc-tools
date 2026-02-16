//go:build testmode

package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/handler"
)

func TestWriteHookResponse(t *testing.T) {
	tests := []struct {
		name       string
		response   *handler.Response
		wantStdout string
		wantStderr string
		wantErr    bool
		wantCode   int
	}{
		{
			name: "empty response produces no output",
			response: &handler.Response{
				ExitCode: 0,
				Stdout:   nil,
				Stderr:   "",
			},
			wantStdout: "",
			wantStderr: "",
			wantErr:    false,
			wantCode:   0,
		},
		{
			name: "stderr text is written to stderr writer",
			response: &handler.Response{
				ExitCode: 0,
				Stdout:   nil,
				Stderr:   "something went wrong",
			},
			wantStdout: "",
			wantStderr: "something went wrong",
			wantErr:    false,
			wantCode:   0,
		},
		{
			name: "stdout with HookOutput is JSON marshaled",
			response: &handler.Response{
				ExitCode: 0,
				Stdout: &handler.HookOutput{
					Continue:           true,
					StopReason:         "",
					SuppressOutput:     false,
					SystemMessage:      "",
					HookSpecificOutput: nil,
					AdditionalContext:  nil,
					PermissionDecision: "",
					UpdatedInput:       nil,
				},
				Stderr: "",
			},
			wantStdout: "",
			wantStderr: "",
			wantErr:    false,
			wantCode:   0,
		},
		{
			name: "non-zero exit code returns exitError",
			response: &handler.Response{
				ExitCode: 2,
				Stdout:   nil,
				Stderr:   "",
			},
			wantStdout: "",
			wantStderr: "",
			wantErr:    true,
			wantCode:   2,
		},
		{
			name: "combined stderr stdout and non-zero exit",
			response: &handler.Response{
				ExitCode: 1,
				Stdout: &handler.HookOutput{
					Continue:           false,
					StopReason:         "",
					SuppressOutput:     true,
					SystemMessage:      "",
					HookSpecificOutput: nil,
					AdditionalContext:  nil,
					PermissionDecision: "",
					UpdatedInput:       nil,
				},
				Stderr: "blocked",
			},
			wantStdout: "",
			wantStderr: "blocked",
			wantErr:    true,
			wantCode:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			err := writeHookResponse(&stdout, &stderr, tt.response)

			assert.Equal(t, tt.wantStderr, stderr.String())

			if tt.wantErr {
				require.Error(t, err)
				var exitErr *exitError
				require.ErrorAs(t, err, &exitErr)
				assert.Equal(t, tt.wantCode, exitErr.code)
			} else {
				require.NoError(t, err)
			}

			if tt.response.Stdout != nil {
				// Verify stdout contains valid JSON matching the HookOutput.
				outputLine := stdout.String()
				assert.NotEmpty(t, outputLine)

				var parsed handler.HookOutput
				unmarshalErr := json.Unmarshal([]byte(outputLine), &parsed)
				require.NoError(t, unmarshalErr)
				assert.Equal(t, tt.response.Stdout.Continue, parsed.Continue)
				assert.Equal(t, tt.response.Stdout.SuppressOutput, parsed.SuppressOutput)
			} else {
				assert.Equal(t, tt.wantStdout, stdout.String())
			}
		})
	}
}

func TestExitError(t *testing.T) {
	err := &exitError{code: 42}
	assert.Equal(t, "exit code 42", err.Error())
}
