package hooks_test

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/riddopic/cc-tools/internal/hooks"
)

// --- Test helpers for executor tests ---

// assertExitCode fails if got != want.
func assertExitCode(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("Expected exit code %d, got %d", want, got)
	}
}

// assertStringContains checks that s contains the expected substring.
func assertStringContains(t *testing.T, s, expected string) {
	t.Helper()
	if !strings.Contains(s, expected) {
		t.Errorf("Expected string to contain %q, got: %s", expected, s)
	}
}

// assertExecutorSuccess verifies the executor result indicates success.
func assertExecutorSuccess(t *testing.T, result *hooks.ExecutorResult) {
	t.Helper()
	if !result.Success {
		t.Errorf("Expected success, got error: %v", result.Error)
	}
}

// assertExecutorFailure verifies the executor result indicates failure.
func assertExecutorFailure(t *testing.T, result *hooks.ExecutorResult) {
	t.Helper()
	if result.Success {
		t.Error("Expected failure")
	}
	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

// newTestDiscoveredCommand creates a DiscoveredCommand with all fields for lint compliance.
func newTestDiscoveredCommand(
	cmdType hooks.CommandType,
	command string,
	args []string,
	workDir string,
) *hooks.DiscoveredCommand {
	return &hooks.DiscoveredCommand{
		Type:       cmdType,
		Command:    command,
		Args:       args,
		WorkingDir: workDir,
		Source:     "",
	}
}

// --- TestCommandExecutor ---

func TestCommandExecutor(t *testing.T) {
	t.Run("successful command execution", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, name string, args ...string) (*hooks.CommandOutput, error) {
			if name == "echo" && len(args) == 1 && args[0] == "hello" {
				return &hooks.CommandOutput{Stdout: []byte("hello\n"), Stderr: nil}, nil
			}
			return nil, errors.New("unexpected command")
		}

		executor := hooks.NewCommandExecutor(5, false, testDeps.Dependencies)
		cmd := newTestDiscoveredCommand(hooks.CommandTypeLint, "echo", []string{"hello"}, ".")

		result := executor.Execute(context.Background(), cmd)
		assertExecutorSuccess(t, result)
		assertExitCode(t, result.ExitCode, 0)
		assertStringContains(t, result.Stdout, "hello")
	})

	t.Run("command failure", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, name string, _ ...string) (*hooks.CommandOutput, error) {
			if name == "false" {
				return &hooks.CommandOutput{Stdout: nil, Stderr: nil}, &exec.ExitError{}
			}
			return nil, errors.New("unexpected command")
		}

		executor := hooks.NewCommandExecutor(5, false, testDeps.Dependencies)
		cmd := newTestDiscoveredCommand(hooks.CommandTypeLint, "false", []string{}, ".")

		result := executor.Execute(context.Background(), cmd)
		assertExecutorFailure(t, result)
	})
}

// --- TestValidateHookEvent ---

// newTestHookInput creates a HookInput with all fields for lint compliance.
func newTestHookInput(eventName, toolName string, toolInput map[string]any) *hooks.HookInput {
	var rawInput json.RawMessage
	if toolInput != nil {
		rawInput = hooks.MustMarshalJSON(toolInput)
	}
	return &hooks.HookInput{
		HookEventName:  eventName,
		SessionID:      "",
		TranscriptPath: "",
		CWD:            "",
		ToolName:       toolName,
		ToolInput:      rawInput,
		ToolResponse:   nil,
	}
}

// newMockStderr creates a MockOutputWriter with all fields for lint compliance.
func newMockStderr() *hooks.MockOutputWriter {
	return &hooks.MockOutputWriter{WrittenData: nil}
}

func TestValidateHookEvent(t *testing.T) {
	t.Run("valid PostToolUse Edit event", func(t *testing.T) {
		input := newTestHookInput("PostToolUse", "Edit", map[string]any{"file_path": "/project/main.go"})
		stderr := newMockStderr()

		filePath, shouldProcess := hooks.ValidateHookEventForTest(input, false, stderr)
		if !shouldProcess {
			t.Error("Expected event to be processed")
		}
		if filePath != "/project/main.go" {
			t.Errorf("Expected file path /project/main.go, got %s", filePath)
		}
	})

	t.Run("wrong event name", func(t *testing.T) {
		input := newTestHookInput("PreToolUse", "Edit", map[string]any{"file_path": "/project/main.go"})
		stderr := newMockStderr()

		_, shouldProcess := hooks.ValidateHookEventForTest(input, false, stderr)
		if shouldProcess {
			t.Error("Expected event not to be processed")
		}
	})

	t.Run("non-edit tool", func(t *testing.T) {
		input := newTestHookInput("PostToolUse", "Bash", nil)
		stderr := newMockStderr()

		_, shouldProcess := hooks.ValidateHookEventForTest(input, false, stderr)
		if shouldProcess {
			t.Error("Expected event not to be processed")
		}
	})

	t.Run("missing file path", func(t *testing.T) {
		input := newTestHookInput("PostToolUse", "Edit", map[string]any{})
		stderr := newMockStderr()

		_, shouldProcess := hooks.ValidateHookEventForTest(input, false, stderr)
		if shouldProcess {
			t.Error("Expected event not to be processed")
		}
	})

	t.Run("nil input", func(t *testing.T) {
		stderr := newMockStderr()

		_, shouldProcess := hooks.ValidateHookEventForTest(nil, false, stderr)
		if shouldProcess {
			t.Error("Expected event not to be processed")
		}
	})

	t.Run("debug output", func(t *testing.T) {
		input := newTestHookInput("PreToolUse", "Bash", nil)
		stderr := newMockStderr()

		hooks.ValidateHookEventForTest(input, true, stderr)

		output := stderr.String()
		assertStringContains(t, output, "Ignoring event")
	})
}
