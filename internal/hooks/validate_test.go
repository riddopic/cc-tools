package hooks_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/riddopic/cc-tools/internal/hooks"
)

func TestValidateResult_FormatMessage(t *testing.T) {
	tests := []struct {
		name         string
		result       *hooks.ValidateResult
		wantEmpty    bool
		wantContains []string
	}{
		{
			name: "both passed",
			result: &hooks.ValidateResult{
				LintResult: nil,
				TestResult: nil,
				BothPassed: true,
			},
			wantEmpty:    false,
			wantContains: []string{"Validations pass"},
		},
		{
			name: "lint failed only",
			result: &hooks.ValidateResult{
				LintResult: &hooks.ValidationResult{
					Type:     hooks.CommandTypeLint,
					Success:  false,
					ExitCode: 1,
					Message:  "",
					Command: &hooks.DiscoveredCommand{
						Type:       hooks.CommandTypeLint,
						Command:    "make",
						Args:       []string{"lint"},
						WorkingDir: "/project",
						Source:     "",
					},
					Error: nil,
				},
				TestResult: &hooks.ValidationResult{
					Type:     hooks.CommandTypeTest,
					Success:  true,
					ExitCode: 0,
					Message:  "",
					Command: &hooks.DiscoveredCommand{
						Type:       hooks.CommandTypeTest,
						Command:    "make",
						Args:       []string{"test"},
						WorkingDir: "",
						Source:     "",
					},
					Error: nil,
				},
				BothPassed: false,
			},
			wantEmpty:    false,
			wantContains: []string{"BLOCKING", "lint failures", "make lint"},
		},
		{
			name: "test failed only",
			result: &hooks.ValidateResult{
				LintResult: &hooks.ValidationResult{
					Type:     hooks.CommandTypeLint,
					Success:  true,
					ExitCode: 0,
					Message:  "",
					Command: &hooks.DiscoveredCommand{
						Type:       hooks.CommandTypeLint,
						Command:    "make",
						Args:       []string{"lint"},
						WorkingDir: "",
						Source:     "",
					},
					Error: nil,
				},
				TestResult: &hooks.ValidationResult{
					Type:     hooks.CommandTypeTest,
					Success:  false,
					ExitCode: 1,
					Message:  "",
					Command: &hooks.DiscoveredCommand{
						Type:       hooks.CommandTypeTest,
						Command:    "make",
						Args:       []string{"test"},
						WorkingDir: "/project",
						Source:     "",
					},
					Error: nil,
				},
				BothPassed: false,
			},
			wantEmpty:    false,
			wantContains: []string{"BLOCKING", "test failures", "make test"},
		},
		{
			name: "both failed",
			result: &hooks.ValidateResult{
				LintResult: &hooks.ValidationResult{
					Type:     hooks.CommandTypeLint,
					Success:  false,
					ExitCode: 1,
					Message:  "",
					Command: &hooks.DiscoveredCommand{
						Type:       hooks.CommandTypeLint,
						Command:    "make",
						Args:       []string{"lint"},
						WorkingDir: "/project",
						Source:     "",
					},
					Error: nil,
				},
				TestResult: &hooks.ValidationResult{
					Type:     hooks.CommandTypeTest,
					Success:  false,
					ExitCode: 1,
					Message:  "",
					Command: &hooks.DiscoveredCommand{
						Type:       hooks.CommandTypeTest,
						Command:    "make",
						Args:       []string{"test"},
						WorkingDir: "/project",
						Source:     "",
					},
					Error: nil,
				},
				BothPassed: false,
			},
			wantEmpty:    false,
			wantContains: []string{"BLOCKING", "Lint and test failures", "make lint", "make test"},
		},
		{
			name: "no commands found",
			result: &hooks.ValidateResult{
				LintResult: nil,
				TestResult: nil,
				BothPassed: true,
			},
			wantEmpty:    false,
			wantContains: []string{"Validations pass"},
		},
		{
			name: "only lint found and passed",
			result: &hooks.ValidateResult{
				LintResult: &hooks.ValidationResult{
					Type:     hooks.CommandTypeLint,
					Success:  true,
					ExitCode: 0,
					Message:  "",
					Command: &hooks.DiscoveredCommand{
						Type:       hooks.CommandTypeLint,
						Command:    "make",
						Args:       []string{"lint"},
						WorkingDir: "",
						Source:     "",
					},
					Error: nil,
				},
				TestResult: nil,
				BothPassed: true,
			},
			wantEmpty:    false,
			wantContains: []string{"Validations pass"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := tt.result.FormatMessage()
			message = stripANSI(message)

			assertFormatMessageResult(t, message, tt.wantEmpty, tt.wantContains)
		})
	}
}

// stripANSI removes common ANSI color codes from a string.
func stripANSI(s string) string {
	s = strings.ReplaceAll(s, "\033[0;33m", "")
	s = strings.ReplaceAll(s, "\033[0;31m", "")
	s = strings.ReplaceAll(s, "\033[0;32m", "")
	s = strings.ReplaceAll(s, "\033[0m", "")
	return s
}

// assertFormatMessageResult verifies the FormatMessage output against expected conditions.
func assertFormatMessageResult(t *testing.T, message string, wantEmpty bool, wantContains []string) {
	t.Helper()
	if wantEmpty && message != "" {
		t.Errorf("FormatMessage() = %q, want empty", message)
	}
	for _, want := range wantContains {
		assertStringContains(t, message, want)
	}
}

// --- Validate executor mock setup helpers ---

// makeDiscoveryAndExecRunner creates a mock runner that handles both make dry-run discovery
// and actual execution for lint and test targets.
func makeDiscoveryAndExecRunner(
	lintResult func() (*hooks.CommandOutput, error),
	testResult func() (*hooks.CommandOutput, error),
) func(context.Context, string, string, ...string) (*hooks.CommandOutput, error) {
	return func(_ context.Context, _, name string, args ...string) (*hooks.CommandOutput, error) {
		if name != "make" {
			return nil, errors.New("command failed")
		}
		return handleMakeCommand(args, lintResult, testResult)
	}
}

// handleMakeCommand routes make commands to the appropriate handler.
func handleMakeCommand(
	args []string,
	lintResult func() (*hooks.CommandOutput, error),
	testResult func() (*hooks.CommandOutput, error),
) (*hooks.CommandOutput, error) {
	isDryRun := len(args) >= 3 && args[len(args)-2] == "-n"
	target := ""
	if len(args) > 0 {
		target = args[len(args)-1]
	}

	if isDryRun {
		return handleDryRun(target, lintResult, testResult)
	}
	return handleExecution(target, lintResult, testResult)
}

// handleDryRun handles make -n (dry-run) discovery calls.
func handleDryRun(
	target string,
	lintResult func() (*hooks.CommandOutput, error),
	testResult func() (*hooks.CommandOutput, error),
) (*hooks.CommandOutput, error) {
	switch target {
	case "lint":
		if lintResult != nil {
			return &hooks.CommandOutput{Stdout: []byte("echo lint"), Stderr: nil}, nil
		}
	case "test":
		if testResult != nil {
			return &hooks.CommandOutput{Stdout: []byte("echo test"), Stderr: nil}, nil
		}
	}
	return nil, errors.New("target not found")
}

// handleExecution handles actual make execution calls.
func handleExecution(
	target string,
	lintResult func() (*hooks.CommandOutput, error),
	testResult func() (*hooks.CommandOutput, error),
) (*hooks.CommandOutput, error) {
	switch target {
	case "lint":
		if lintResult != nil {
			return lintResult()
		}
	case "test":
		if testResult != nil {
			return testResult()
		}
	}
	return nil, errors.New("command failed")
}

// successOutput returns a factory for successful command output.
func successOutput(msg string) func() (*hooks.CommandOutput, error) {
	return func() (*hooks.CommandOutput, error) {
		return &hooks.CommandOutput{Stdout: []byte(msg), Stderr: nil}, nil
	}
}

// failOutput returns a factory for failed command output.
func failOutput(msg string) func() (*hooks.CommandOutput, error) {
	return func() (*hooks.CommandOutput, error) {
		return &hooks.CommandOutput{Stdout: nil, Stderr: []byte(msg)}, errors.New("exit status 1")
	}
}

// setupMakefileFS configures the mock filesystem to find a Makefile.
func setupMakefileFS(deps *hooks.TestDependencies) {
	deps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "Makefile") {
			return hooks.NewMockFileInfo("Makefile", 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}
}

// assertValidateResults checks BothPassed and presence of LintResult/TestResult.
func assertValidateResults(t *testing.T, result *hooks.ValidateResult, wantBothPass, wantLint, wantTest bool) {
	t.Helper()
	if result.BothPassed != wantBothPass {
		t.Errorf("BothPassed = %v, want %v", result.BothPassed, wantBothPass)
	}
	if (result.LintResult != nil) != wantLint {
		t.Errorf("LintResult found = %v, want %v", result.LintResult != nil, wantLint)
	}
	if (result.TestResult != nil) != wantTest {
		t.Errorf("TestResult found = %v, want %v", result.TestResult != nil, wantTest)
	}
}

func TestParallelValidateExecutor_ExecuteValidations(t *testing.T) {
	tests := []struct {
		name          string
		setupDeps     func(*hooks.TestDependencies)
		wantBothPass  bool
		wantLintFound bool
		wantTestFound bool
	}{
		{
			name: "both commands succeed",
			setupDeps: func(deps *hooks.TestDependencies) {
				setupMakefileFS(deps)
				deps.MockRunner.RunContextFunc = makeDiscoveryAndExecRunner(
					successOutput("Linting..."),
					successOutput("Testing..."),
				)
			},
			wantBothPass:  true,
			wantLintFound: true,
			wantTestFound: true,
		},
		{
			name: "lint fails test passes",
			setupDeps: func(deps *hooks.TestDependencies) {
				setupMakefileFS(deps)
				deps.MockRunner.RunContextFunc = makeDiscoveryAndExecRunner(
					failOutput("Lint error"),
					successOutput("Tests pass"),
				)
			},
			wantBothPass:  false,
			wantLintFound: true,
			wantTestFound: true,
		},
		{
			name: "no commands found",
			setupDeps: func(deps *hooks.TestDependencies) {
				deps.MockFS.StatFunc = func(_ string) (os.FileInfo, error) {
					return nil, os.ErrNotExist
				}
				deps.MockRunner.RunContextFunc = func(_ context.Context, _, _ string, _ ...string) (*hooks.CommandOutput, error) {
					return nil, errors.New("command failed")
				}
			},
			wantBothPass:  true,
			wantLintFound: false,
			wantTestFound: false,
		},
		{
			name: "only lint command found",
			setupDeps: func(deps *hooks.TestDependencies) {
				setupMakefileFS(deps)
				deps.MockRunner.RunContextFunc = makeDiscoveryAndExecRunner(
					successOutput("Linting..."),
					nil, // no test target
				)
			},
			wantBothPass:  true,
			wantLintFound: true,
			wantTestFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDeps := hooks.CreateTestDependencies()
			tt.setupDeps(testDeps)

			executor := hooks.NewParallelValidateExecutor("/project", 10, false, nil, testDeps.Dependencies)
			result, err := executor.ExecuteValidations(context.Background(), "/project", "/project")
			if err != nil {
				t.Fatalf("ExecuteValidations() error = %v", err)
			}

			assertValidateResults(t, result, tt.wantBothPass, tt.wantLintFound, tt.wantTestFound)
		})
	}
}

// --- Validate hook test helpers ---

// setupValidateHookDeps configures mock dependencies for RunValidateHook tests with
// filesystem and runner mocks.
func setupValidateHookDeps(deps *hooks.TestDependencies, input string) {
	deps.MockInput.ReadAllFunc = func() ([]byte, error) {
		return []byte(input), nil
	}
}

// setupGitMakefileProjectFS configures the mock filesystem to detect .git and Makefile for
// project root resolution.
func setupGitMakefileProjectFS(deps *hooks.TestDependencies) {
	deps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, ".git") || strings.HasSuffix(path, "Makefile") {
			return hooks.NewMockFileInfo(filepath.Base(path), 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}
}

func TestRunValidateHook(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		setupDeps    func(*hooks.TestDependencies)
		wantExitCode int
	}{
		{
			name: "successful validation",
			input: `{
				"hook_event_name": "PostToolUse",
				"tool_name": "Edit",
				"tool_input": {"file_path": "/project/main.go"}
			}`,
			setupDeps: func(deps *hooks.TestDependencies) {
				setupGitMakefileProjectFS(deps)
				deps.MockRunner.RunContextFunc = makeDiscoveryAndExecRunner(
					successOutput("OK"),
					successOutput("OK"),
				)
			},
			wantExitCode: 2,
		},
		{
			name: "validation failures",
			input: `{
				"hook_event_name": "PostToolUse",
				"tool_name": "Edit",
				"tool_input": {"file_path": "/project/main.go"}
			}`,
			setupDeps: func(deps *hooks.TestDependencies) {
				setupGitMakefileProjectFS(deps)
				deps.MockRunner.RunContextFunc = makeDiscoveryAndExecRunner(
					failOutput("lint errors"),
					successOutput("OK"),
				)
			},
			wantExitCode: 2,
		},
		{
			name:         "invalid input",
			input:        "not json",
			setupDeps:    func(_ *hooks.TestDependencies) {},
			wantExitCode: 0,
		},
		{
			name: "wrong event type",
			input: `{
				"hook_event_name": "PreToolUse",
				"tool_name": "Edit",
				"tool_input": {"file_path": "/project/main.go"}
			}`,
			setupDeps:    func(_ *hooks.TestDependencies) {},
			wantExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDeps := hooks.CreateTestDependencies()
			tt.setupDeps(testDeps)
			setupValidateHookDeps(testDeps, tt.input)

			exitCode := hooks.RunValidateHook(
				context.Background(),
				false, 10, 2,
				testDeps.Dependencies,
			)

			assertExitCode(t, exitCode, tt.wantExitCode)
		})
	}
}

func TestValidateExecutor_Parallelism(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()
	setupMakefileFS(testDeps)

	var executionOrder []string
	var mu sync.Mutex
	testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, name string, args ...string) (*hooks.CommandOutput, error) {
		fullCmd := fmt.Sprintf("%s %s", name, strings.Join(args, " "))

		if name == "make" && len(args) >= 3 && args[len(args)-2] == "-n" {
			return &hooks.CommandOutput{Stdout: []byte("echo cmd"), Stderr: nil}, nil
		}

		if name == "make" && len(args) == 1 {
			mu.Lock()
			executionOrder = append(executionOrder, fullCmd)
			mu.Unlock()
			return &hooks.CommandOutput{Stdout: []byte("OK"), Stderr: nil}, nil
		}

		return nil, fmt.Errorf("unknown command: %s", fullCmd)
	}

	executor := hooks.NewParallelValidateExecutor("/project", 10, false, nil, testDeps.Dependencies)
	result, err := executor.ExecuteValidations(context.Background(), "/project", "/project")
	if err != nil {
		t.Fatalf("ExecuteValidations() error = %v", err)
	}

	if !result.BothPassed {
		t.Error("Expected both validations to pass")
	}

	assertParallelExecution(t, executionOrder)
}

// assertParallelExecution verifies that both lint and test commands were executed.
func assertParallelExecution(t *testing.T, executionOrder []string) {
	t.Helper()
	if len(executionOrder) < 2 {
		t.Errorf("Expected at least 2 commands to be executed, got %d", len(executionOrder))
	}

	hasLint := false
	hasTest := false
	for _, cmd := range executionOrder {
		if strings.Contains(cmd, "lint") {
			hasLint = true
		}
		if strings.Contains(cmd, "test") {
			hasTest = true
		}
	}

	if !hasLint || !hasTest {
		t.Errorf("Expected both lint and test to be executed. Commands: %v", executionOrder)
	}
}
