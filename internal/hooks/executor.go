package hooks

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	debuglog "github.com/riddopic/cc-tools/internal/debug"
)

const (
	// ExitCodeShowMessage is used to signal that a message should be shown to Claude.
	ExitCodeShowMessage = 2
)

// ExecutorResult represents the result of executing a command.
type ExecutorResult struct {
	Success  bool
	ExitCode int
	Stdout   string
	Stderr   string
	Error    error
	TimedOut bool
}

// CommandExecutor handles executing discovered commands.
type CommandExecutor struct {
	timeout time.Duration
	debug   bool
	deps    *Dependencies
}

// NewCommandExecutor creates a new command executor.
func NewCommandExecutor(timeoutSecs int, debug bool, deps *Dependencies) *CommandExecutor {
	if deps == nil {
		deps = NewDefaultDependencies()
	}
	return &CommandExecutor{
		timeout: time.Duration(timeoutSecs) * time.Second,
		debug:   debug,
		deps:    deps,
	}
}

// Execute runs the discovered command with the given context and timeout.
func (ce *CommandExecutor) Execute(ctx context.Context, cmd *DiscoveredCommand) *ExecutorResult {
	if cmd == nil {
		return &ExecutorResult{
			Success:  false,
			ExitCode: 0,
			Stdout:   "",
			Stderr:   "",
			Error:    errors.New("no command to execute"),
			TimedOut: false,
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, ce.timeout)
	defer cancel()

	// Run the command through dependencies
	output, err := ce.deps.Runner.RunContext(ctx, cmd.WorkingDir, cmd.Command, cmd.Args...)

	// Check if context timed out
	if ctx.Err() == context.DeadlineExceeded {
		var stdout, stderr string
		if output != nil {
			stdout = string(output.Stdout)
			stderr = string(output.Stderr)
		}
		return &ExecutorResult{
			Success:  false,
			ExitCode: -1,
			Stdout:   stdout,
			Stderr:   stderr,
			Error:    fmt.Errorf("command timed out after %v", ce.timeout),
			TimedOut: true,
		}
	}

	// Get exit code
	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	var stdout, stderr string
	if output != nil {
		stdout = string(output.Stdout)
		stderr = string(output.Stderr)
	}

	return &ExecutorResult{
		Success:  err == nil,
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
		Error:    err,
		TimedOut: false,
	}
}

// handleInputError handles errors from reading hook input.
func handleInputError(err error, debug bool, stderr OutputWriter) {
	if !errors.Is(err, ErrNoInput) && debug {
		// Only log if it's not the expected "no input" error
		_, _ = fmt.Fprintf(stderr, "Error reading input: %v\n", err)
	}
}

// validateHookEvent checks if the event should be processed.
func validateHookEvent(input *HookInput, debug bool, stderr OutputWriter) (string, bool) {
	if input == nil || input.HookEventName != "PostToolUse" || !input.IsEditTool() {
		if debug && input != nil {
			_, _ = fmt.Fprintf(stderr, "Ignoring event: %s, tool: %s\n",
				input.HookEventName, input.ToolName)
		}
		return "", false
	}

	filePath := input.GetFilePath()
	if filePath == "" {
		if debug {
			_, _ = fmt.Fprintf(stderr, "No file path found in input\n")
		}
		return "", false
	}

	return filePath, true
}

// acquireLock tries to acquire the lock for the hook.
func acquireLock(lockMgr *LockManager, debug bool, stderr OutputWriter, logger *debuglog.Logger) bool {
	acquired, err := lockMgr.TryAcquire()
	if err != nil {
		if logger != nil && logger.IsEnabled() {
			logger.LogError(err, "acquiring lock")
		}
		if debug {
			_, _ = fmt.Fprintf(stderr, "Error acquiring lock: %v\n", err)
		}
		return false
	}
	if !acquired {
		if logger != nil && logger.IsEnabled() {
			logger.Logf("Another instance is running or in cooldown")
		}
		if debug {
			_, _ = fmt.Fprintf(stderr, "Another instance is running or in cooldown\n")
		}
		return false
	}
	if logger != nil && logger.IsEnabled() {
		logger.Logf("Lock acquired successfully")
	}
	return true
}
