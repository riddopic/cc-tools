package hooks_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

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

// assertStderrContains checks that the mock stderr output contains the expected substring.
func assertStderrContains(t *testing.T, deps *hooks.TestDependencies, expected string) {
	t.Helper()
	output := deps.MockStderr.String()
	if !strings.Contains(output, expected) {
		t.Errorf("Expected stderr to contain %q, got: %s", expected, output)
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

// setupEditInput configures the mock to provide a PostToolUse Edit event for the given file.
func setupEditInput(deps *hooks.TestDependencies, filePath string) {
	deps.MockInput.IsTerminalFunc = func() bool { return false }
	deps.MockInput.ReadAllFunc = func() ([]byte, error) {
		return []byte(`{
			"hook_event_name": "PostToolUse",
			"tool_name": "Edit",
			"tool_input": {"file_path": "` + filePath + `"}
		}`), nil
	}
}

// setupGitProjectFS configures the mock filesystem so that .git is found (project detection).
func setupGitProjectFS(deps *hooks.TestDependencies) {
	deps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.Contains(path, ".git") {
			return hooks.NewMockFileInfo(".git", 0, 0, time.Time{}, true), nil
		}
		return nil, errors.New("not found")
	}
}

// setupGitMakefileFS configures the mock filesystem for .git and Makefile detection.
func setupGitMakefileFS(deps *hooks.TestDependencies) {
	deps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.Contains(path, ".git") || strings.Contains(path, "Makefile") {
			return hooks.NewMockFileInfo("", 0, 0, time.Time{}, strings.Contains(path, ".git")), nil
		}
		return nil, errors.New("not found")
	}
}

// setupLockAvailable configures mocks so that lock acquisition succeeds.
func setupLockAvailable(deps *hooks.TestDependencies) {
	deps.MockFS.TempDirFunc = func() string { return "/tmp" }
	deps.MockFS.ReadFileFunc = func(_ string) ([]byte, error) {
		return nil, errors.New("not found")
	}
	deps.MockFS.WriteFileFunc = func(_ string, _ []byte, _ os.FileMode) error {
		return nil
	}
	deps.MockProcess.GetPIDFunc = func() int { return 99999 }
	deps.MockProcess.ProcessExistsFunc = func(_ int) bool { return false }
	deps.MockClock.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
}

// makeLintDryRunRunner returns a mock runner function that handles Makefile lint dry-run
// discovery and delegates actual execution to execFn.
func makeLintDryRunRunner(
	execFn func(ctx context.Context) (*hooks.CommandOutput, error),
) func(context.Context, string, string, ...string) (*hooks.CommandOutput, error) {
	return func(ctx context.Context, _, name string, args ...string) (*hooks.CommandOutput, error) {
		if name == "make" && len(args) > 0 && args[len(args)-1] == "lint" {
			if len(args) > 1 && args[len(args)-2] == "-n" {
				return &hooks.CommandOutput{Stdout: []byte("golangci-lint run"), Stderr: nil}, nil
			}
			return execFn(ctx)
		}
		return nil, errors.New("command not found")
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

// --- TestRunSmartHook ---

func TestRunSmartHook(t *testing.T) {
	t.Run("exit code 0 when no input", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		testDeps.MockInput.IsTerminalFunc = func() bool { return true }

		exitCode := hooks.RunSmartHook(context.Background(), hooks.CommandTypeLint, false, 20, 5, testDeps.Dependencies)
		assertExitCode(t, exitCode, 0)
	})

	t.Run("exit code 0 when wrong event type", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		testDeps.MockInput.IsTerminalFunc = func() bool { return false }
		testDeps.MockInput.ReadAllFunc = func() ([]byte, error) {
			return []byte(`{"hook_event_name": "PreToolUse", "tool_name": "Edit"}`), nil
		}

		exitCode := hooks.RunSmartHook(context.Background(), hooks.CommandTypeLint, false, 20, 5, testDeps.Dependencies)
		assertExitCode(t, exitCode, 0)
	})

	t.Run("exit code 0 when non-edit tool", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		testDeps.MockInput.IsTerminalFunc = func() bool { return false }
		testDeps.MockInput.ReadAllFunc = func() ([]byte, error) {
			return []byte(`{"hook_event_name": "PostToolUse", "tool_name": "Bash"}`), nil
		}

		exitCode := hooks.RunSmartHook(context.Background(), hooks.CommandTypeLint, false, 20, 5, testDeps.Dependencies)
		assertExitCode(t, exitCode, 0)
	})

	t.Run("exit code 0 when file should be skipped", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		setupEditInput(testDeps, "/project/main_test.go")

		exitCode := hooks.RunSmartHook(context.Background(), hooks.CommandTypeLint, false, 20, 5, testDeps.Dependencies)
		assertExitCode(t, exitCode, 0)
	})

	t.Run("exit code 0 when lock cannot be acquired", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		setupEditInput(testDeps, "/project/main.go")
		setupGitProjectFS(testDeps)

		testDeps.MockFS.TempDirFunc = func() string { return "/tmp" }
		testDeps.MockFS.ReadFileFunc = func(path string) ([]byte, error) {
			if strings.Contains(path, "lock") {
				return []byte("12345\n"), nil
			}
			return nil, errors.New("not found")
		}
		testDeps.MockProcess.ProcessExistsFunc = func(pid int) bool {
			return pid == 12345
		}

		exitCode := hooks.RunSmartHook(context.Background(), hooks.CommandTypeLint, false, 20, 5, testDeps.Dependencies)
		assertExitCode(t, exitCode, 0)
	})

	t.Run("exit code 0 when no command found", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		setupEditInput(testDeps, "/project/main.go")
		setupGitProjectFS(testDeps)
		setupLockAvailable(testDeps)

		testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, _ string, _ ...string) (*hooks.CommandOutput, error) {
			return nil, errors.New("command not found")
		}
		testDeps.MockRunner.LookPathFunc = func(_ string) (string, error) {
			return "", errors.New("not found")
		}

		exitCode := hooks.RunSmartHook(context.Background(), hooks.CommandTypeLint, false, 20, 5, testDeps.Dependencies)
		assertExitCode(t, exitCode, 0)
	})

	t.Run("exit code 2 on lint failure", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		setupEditInput(testDeps, "/project/main.go")
		setupGitMakefileFS(testDeps)
		setupLockAvailable(testDeps)

		testDeps.MockRunner.RunContextFunc = makeLintDryRunRunner(
			func(_ context.Context) (*hooks.CommandOutput, error) {
				return &hooks.CommandOutput{Stdout: nil, Stderr: []byte("lint errors")}, &exec.ExitError{}
			},
		)

		exitCode := hooks.RunSmartHook(context.Background(), hooks.CommandTypeLint, false, 20, 5, testDeps.Dependencies)
		assertExitCode(t, exitCode, 2)
		assertStderrContains(t, testDeps, "BLOCKING")
		assertStderrContains(t, testDeps, "make lint")
	})

	t.Run("exit code 2 on lint success", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		setupEditInput(testDeps, "/project/main.go")
		setupGitMakefileFS(testDeps)
		setupLockAvailable(testDeps)

		testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, name string, args ...string) (*hooks.CommandOutput, error) {
			if name == "make" && len(args) > 0 && args[len(args)-1] == "lint" {
				return &hooks.CommandOutput{Stdout: []byte("lint passed"), Stderr: nil}, nil
			}
			return nil, errors.New("command not found")
		}

		exitCode := hooks.RunSmartHook(context.Background(), hooks.CommandTypeLint, false, 20, 5, testDeps.Dependencies)
		assertExitCode(t, exitCode, 2)
		assertStderrContains(t, testDeps, "Lints pass")
	})

	t.Run("exit code 2 on test success", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		setupEditInput(testDeps, "/project/main.go")
		setupGitMakefileFS(testDeps)
		setupLockAvailable(testDeps)

		testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, name string, args ...string) (*hooks.CommandOutput, error) {
			if name == "make" && len(args) > 0 && args[len(args)-1] == "test" {
				return &hooks.CommandOutput{Stdout: []byte("test passed"), Stderr: nil}, nil
			}
			return nil, errors.New("command not found")
		}

		exitCode := hooks.RunSmartHook(context.Background(), hooks.CommandTypeTest, false, 20, 5, testDeps.Dependencies)
		assertExitCode(t, exitCode, 2)
		assertStderrContains(t, testDeps, "Tests pass")
	})

	t.Run("command timeout", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		setupEditInput(testDeps, "/project/main.go")
		setupGitMakefileFS(testDeps)
		setupLockAvailable(testDeps)

		testDeps.MockRunner.RunContextFunc = makeLintDryRunRunner(
			func(ctx context.Context) (*hooks.CommandOutput, error) {
				<-ctx.Done()
				return nil, context.DeadlineExceeded
			},
		)

		exitCode := hooks.RunSmartHook(context.Background(), hooks.CommandTypeLint, false, 1, 5, testDeps.Dependencies)
		assertExitCode(t, exitCode, 2)
		assertStderrContains(t, testDeps, "timed out")
	})
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

	t.Run("ExecuteForHook formats lint failure message", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, _ string, _ ...string) (*hooks.CommandOutput, error) {
			return &hooks.CommandOutput{Stdout: nil, Stderr: []byte("lint errors")}, &exec.ExitError{}
		}

		executor := hooks.NewCommandExecutor(5, false, testDeps.Dependencies)
		cmd := newTestDiscoveredCommand(hooks.CommandTypeLint, "make", []string{"lint"}, "/project")

		exitCode, message := executor.ExecuteForHook(context.Background(), cmd, hooks.CommandTypeLint)
		assertExitCode(t, exitCode, 2)
		assertStringContains(t, message, "BLOCKING")
		assertStringContains(t, message, "make lint")
	})

	t.Run("ExecuteForHook formats test failure message", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, _ string, _ ...string) (*hooks.CommandOutput, error) {
			return &hooks.CommandOutput{Stdout: nil, Stderr: []byte("test failures")}, &exec.ExitError{}
		}

		executor := hooks.NewCommandExecutor(5, false, testDeps.Dependencies)
		cmd := newTestDiscoveredCommand(hooks.CommandTypeTest, "go", []string{"test", "./..."}, "/project")

		exitCode, message := executor.ExecuteForHook(context.Background(), cmd, hooks.CommandTypeTest)
		assertExitCode(t, exitCode, 2)
		assertStringContains(t, message, "test failures")
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
