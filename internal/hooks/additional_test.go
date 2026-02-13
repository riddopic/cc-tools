package hooks_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/riddopic/cc-tools/internal/hooks"
)

// Additional tests to cover remaining edge cases

func TestExecutorEdgeCases(t *testing.T) {
	t.Run("Execute with nil command", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		executor := hooks.NewCommandExecutor(5, false, testDeps.Dependencies)

		result := executor.Execute(context.Background(), nil)
		if result.Success {
			t.Error("Expected failure for nil command")
		}
		if result.Error == nil {
			t.Error("Expected error for nil command")
		}
	})

	t.Run("ExecuteForHook with timeout", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()

		// Mock command that times out
		testDeps.MockRunner.RunContextFunc = func(ctx context.Context, _, _ string, _ ...string) (*hooks.CommandOutput, error) {
			<-ctx.Done()
			return nil, context.DeadlineExceeded
		}

		executor := hooks.NewCommandExecutor(1, false, testDeps.Dependencies)
		cmd := &hooks.DiscoveredCommand{
			Type:       hooks.CommandTypeLint,
			Command:    "sleep",
			Args:       []string{"10"},
			WorkingDir: "/project",
			Source:     "",
		}

		exitCode, message := executor.ExecuteForHook(context.Background(), cmd, hooks.CommandTypeLint)
		if exitCode != 2 {
			t.Errorf("Expected exit code 2, got %d", exitCode)
		}
		if !strings.Contains(message, "timed out") {
			t.Errorf("Expected timeout message, got: %s", message)
		}
	})

	t.Run("ExecuteForHook with unknown command type", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()

		testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, _ string, _ ...string) (*hooks.CommandOutput, error) {
			return nil, &exec.ExitError{}
		}

		executor := hooks.NewCommandExecutor(5, false, testDeps.Dependencies)
		cmd := &hooks.DiscoveredCommand{
			Type:       hooks.CommandType("unknown"),
			Command:    "test",
			Args:       []string{},
			WorkingDir: "/project",
			Source:     "",
		}

		exitCode, message := executor.ExecuteForHook(context.Background(), cmd, hooks.CommandType("unknown"))
		if exitCode != 2 {
			t.Errorf("Expected exit code 2, got %d", exitCode)
		}
		if !strings.Contains(message, "BLOCKING") {
			t.Errorf("Expected BLOCKING message, got: %s", message)
		}
	})

	t.Run("ExecuteForHook success always shows message", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()

		testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, _ string, _ ...string) (*hooks.CommandOutput, error) {
			return &hooks.CommandOutput{
				Stdout: []byte("success"),
				Stderr: nil,
			}, nil
		}

		executor := hooks.NewCommandExecutor(5, false, testDeps.Dependencies)
		cmd := &hooks.DiscoveredCommand{
			Type:       hooks.CommandType("unknown"),
			Command:    "echo",
			Args:       []string{"test"},
			WorkingDir: "/project",
			Source:     "",
		}

		exitCode, message := executor.ExecuteForHook(context.Background(), cmd, hooks.CommandType("unknown"))
		if exitCode != 2 {
			t.Errorf("Expected exit code 2, got %d", exitCode)
		}
		if !strings.Contains(message, "\u2713") {
			t.Errorf("Expected success message with checkmark, got: %s", message)
		}
	})
}

func TestDiscoveryEdgeCases(t *testing.T) {
	t.Run("checkPackageJSON with jq error", testDiscoveryJQError)
	t.Run("detects bun from lock file", testDiscoveryBunLockFile)
	t.Run("Python pylint fallback", testDiscoveryPylintFallback)
	t.Run("no Python linters available", testDiscoveryNoPythonLinters)
	t.Run("stops at filesystem root", testDiscoveryStopsAtRoot)
}

func testDiscoveryJQError(t *testing.T) {
	t.Helper()

	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "package.json") {
			return hooks.NewMockFileInfo("package.json", 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	// jq fails
	testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, _ string, _ ...string) (*hooks.CommandOutput, error) {
		return nil, errors.New("jq error")
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(context.Background(), hooks.CommandTypeLint, "/project")

	if err == nil {
		t.Fatal("Expected error when package.json script not found")
	}
	if cmd != nil {
		t.Fatal("Expected no command when jq fails")
	}
}

func testDiscoveryBunLockFile(t *testing.T) {
	t.Helper()

	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "package.json") || strings.HasSuffix(path, "bun.lockb") {
			return hooks.NewMockFileInfo(path, 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, name string, _ ...string) (*hooks.CommandOutput, error) {
		if name == "jq" {
			return &hooks.CommandOutput{
				Stdout: []byte(`"test script"`),
				Stderr: nil,
			}, nil
		}
		return nil, errors.New("command failed")
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(context.Background(), hooks.CommandTypeTest, "/project")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd.Command != "bun" {
		t.Errorf("Expected bun, got %s", cmd.Command)
	}
}

func testDiscoveryPylintFallback(t *testing.T) {
	t.Helper()

	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "setup.py") {
			return hooks.NewMockFileInfo("setup.py", 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	testDeps.MockRunner.LookPathFunc = func(file string) (string, error) {
		if file == "pylint" {
			return "/usr/local/bin/pylint", nil
		}
		return "", errors.New("not found")
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(context.Background(), hooks.CommandTypeLint, "/project")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd.Command != "pylint" {
		t.Errorf("Expected pylint, got %s", cmd.Command)
	}
}

func testDiscoveryNoPythonLinters(t *testing.T) {
	t.Helper()

	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "requirements.txt") {
			return hooks.NewMockFileInfo("requirements.txt", 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	testDeps.MockRunner.LookPathFunc = func(_ string) (string, error) {
		return "", errors.New("not found")
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(context.Background(), hooks.CommandTypeLint, "/project")

	// Should return nil when no linters are found
	if err == nil {
		t.Fatal("Expected error when no Python linters found")
	}
	if cmd != nil {
		t.Fatal("Expected no command when no linters available")
	}
}

func testDiscoveryStopsAtRoot(t *testing.T) {
	t.Helper()

	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(_ string) (os.FileInfo, error) {
		return nil, os.ErrNotExist
	}

	discovery := hooks.NewCommandDiscovery("/", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(context.Background(), hooks.CommandTypeLint, "/")

	if err == nil {
		t.Fatal("Expected error at filesystem root")
	}
	if cmd != nil {
		t.Fatal("Expected no command at filesystem root")
	}
}

func TestRunSmartHookEdgeCases(t *testing.T) {
	t.Run("handles error finding project root", testSmartHookErrorFindingRoot)
	t.Run("handles lock acquisition error", testSmartHookLockAcquisitionError)
	t.Run("handles missing file path in input", testSmartHookMissingFilePath)
	t.Run("handles discovery error with debug", testSmartHookDiscoveryError)
}

func testSmartHookErrorFindingRoot(t *testing.T) {
	t.Helper()

	testDeps := hooks.CreateTestDependencies()

	testDeps.MockInput.IsTerminalFunc = func() bool { return false }
	testDeps.MockInput.ReadAllFunc = func() ([]byte, error) {
		return []byte(`{
			"hook_event_name": "PostToolUse",
			"tool_name": "Edit",
			"tool_input": {"file_path": "/nonexistent/file.go"}
		}`), nil
	}

	// No .git directory found anywhere
	testDeps.MockFS.StatFunc = func(_ string) (os.FileInfo, error) {
		return nil, os.ErrNotExist
	}

	exitCode := hooks.RunSmartHook(context.Background(), hooks.CommandTypeLint, false, 20, 5, testDeps.Dependencies)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func testSmartHookLockAcquisitionError(t *testing.T) {
	t.Helper()

	testDeps := hooks.CreateTestDependencies()

	testDeps.MockInput.IsTerminalFunc = func() bool { return false }
	testDeps.MockInput.ReadAllFunc = func() ([]byte, error) {
		return []byte(`{
			"hook_event_name": "PostToolUse",
			"tool_name": "Edit",
			"tool_input": {"file_path": "/project/main.go"}
		}`), nil
	}

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.Contains(path, ".git") {
			return hooks.NewMockFileInfo(".git", 0, 0, time.Time{}, true), nil
		}
		return nil, os.ErrNotExist
	}

	// Lock file operations fail
	testDeps.MockFS.TempDirFunc = func() string { return "/tmp" }
	testDeps.MockFS.ReadFileFunc = func(path string) ([]byte, error) {
		if strings.Contains(path, "lock") {
			return nil, errors.New("permission denied")
		}
		return nil, os.ErrNotExist
	}
	testDeps.MockFS.WriteFileFunc = func(path string, _ []byte, _ os.FileMode) error {
		if strings.Contains(path, "lock") {
			return errors.New("permission denied")
		}
		return nil
	}

	exitCode := hooks.RunSmartHook(
		context.Background(),
		hooks.CommandTypeLint,
		true, 20, 5,
		testDeps.Dependencies,
	)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func testSmartHookMissingFilePath(t *testing.T) {
	t.Helper()

	testDeps := hooks.CreateTestDependencies()

	testDeps.MockInput.IsTerminalFunc = func() bool { return false }
	testDeps.MockInput.ReadAllFunc = func() ([]byte, error) {
		return []byte(`{
			"hook_event_name": "PostToolUse",
			"tool_name": "Edit",
			"tool_input": {}
		}`), nil
	}

	exitCode := hooks.RunSmartHook(
		context.Background(),
		hooks.CommandTypeLint,
		true, 20, 5,
		testDeps.Dependencies,
	)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Check debug message
	output := testDeps.MockStderr.String()
	if !strings.Contains(output, "No file path") {
		t.Errorf("Expected 'No file path' in debug output, got: %s", output)
	}
}

func testSmartHookDiscoveryError(t *testing.T) {
	t.Helper()

	testDeps := hooks.CreateTestDependencies()

	testDeps.MockInput.IsTerminalFunc = func() bool { return false }
	testDeps.MockInput.ReadAllFunc = func() ([]byte, error) {
		return []byte(`{
			"hook_event_name": "PostToolUse",
			"tool_name": "Edit",
			"tool_input": {"file_path": "/project/main.go"}
		}`), nil
	}

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.Contains(path, ".git") {
			return hooks.NewMockFileInfo(".git", 0, 0, time.Time{}, true), nil
		}
		// All other files don't exist
		return nil, os.ErrNotExist
	}
	testDeps.MockFS.TempDirFunc = func() string { return "/tmp" }
	testDeps.MockFS.ReadFileFunc = func(_ string) ([]byte, error) {
		return nil, os.ErrNotExist
	}
	testDeps.MockFS.WriteFileFunc = func(_ string, _ []byte, _ os.FileMode) error {
		return nil
	}

	testDeps.MockProcess.GetPIDFunc = func() int { return 99999 }

	exitCode := hooks.RunSmartHook(
		context.Background(),
		hooks.CommandTypeLint,
		true, 20, 5,
		testDeps.Dependencies,
	)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Check debug message
	output := testDeps.MockStderr.String()
	if !strings.Contains(output, "Error discovering command") {
		t.Errorf("Expected 'Error discovering command' in debug output, got: %s", output)
	}
}

func TestHandleInputError(t *testing.T) {
	t.Run("logs non-ErrNoInput errors in debug mode", func(t *testing.T) {
		stderr := &hooks.MockOutputWriter{WrittenData: nil}
		err := errors.New("unexpected error")

		hooks.HandleInputErrorForTest(err, true, stderr)

		output := stderr.String()
		if !strings.Contains(output, "Error reading input") {
			t.Errorf("Expected error log in debug mode, got: %s", output)
		}
	})

	t.Run("silent for ErrNoInput even in debug mode", func(t *testing.T) {
		stderr := &hooks.MockOutputWriter{WrittenData: nil}

		hooks.HandleInputErrorForTest(hooks.ErrNoInput, true, stderr)

		output := stderr.String()
		if output != "" {
			t.Errorf("Expected no output for ErrNoInput, got: %s", output)
		}
	})

	t.Run("silent when not in debug mode", func(t *testing.T) {
		stderr := &hooks.MockOutputWriter{WrittenData: nil}
		err := errors.New("some error")

		hooks.HandleInputErrorForTest(err, false, stderr)

		output := stderr.String()
		if output != "" {
			t.Errorf("Expected no output when not in debug, got: %s", output)
		}
	})
}

func TestLockManagerCleanupOnExit(t *testing.T) {
	t.Run("Release respects cleanupOnExit flag", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()

		var writeCount int
		testDeps.MockFS.TempDirFunc = func() string { return "/tmp" }
		testDeps.MockFS.WriteFileFunc = func(_ string, _ []byte, _ os.FileMode) error {
			writeCount++
			return nil
		}

		lm := hooks.NewLockManager("/project", "test", 5, testDeps.Dependencies)
		lm.SetCleanupOnExit(false) // Disable cleanup

		err := lm.Release()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if writeCount != 0 {
			t.Error("Expected no write when cleanupOnExit is false")
		}
	})

	t.Run("Release handles write error", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()

		testDeps.MockFS.TempDirFunc = func() string { return "/tmp" }
		testDeps.MockFS.WriteFileFunc = func(_ string, _ []byte, _ os.FileMode) error {
			return errors.New("disk full")
		}
		testDeps.MockClock.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }

		lm := hooks.NewLockManager("/project", "test", 5, testDeps.Dependencies)

		err := lm.Release()
		if err == nil {
			t.Fatal("Expected error on write failure")
		}
		if !strings.Contains(err.Error(), "disk full") {
			t.Errorf("Expected 'disk full' in error, got: %v", err)
		}
	})
}
