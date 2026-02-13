package hooks

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/riddopic/cc-tools/internal/shared"
)

// CommandOutput contains the output from a command execution.
type CommandOutput struct {
	Stdout []byte
	Stderr []byte
}

// CommandRunner executes external commands.
type CommandRunner interface {
	RunContext(ctx context.Context, dir, name string, args ...string) (*CommandOutput, error)
	LookPath(file string) (string, error)
}

// ProcessManager manages system processes.
type ProcessManager interface {
	GetPID() int
	FindProcess(pid int) (*os.Process, error)
	ProcessExists(pid int) bool
}

// Clock provides time operations.
type Clock interface {
	Now() time.Time
}

// InputReader reads input from various sources.
type InputReader interface {
	ReadAll() ([]byte, error)
	IsTerminal() bool
}

// OutputWriter writes output to various destinations.
type OutputWriter interface {
	io.Writer
}

// Dependencies holds all external dependencies.
type Dependencies struct {
	FS      shared.HooksFS
	Runner  CommandRunner
	Process ProcessManager
	Clock   Clock
	Input   InputReader
	Stdout  OutputWriter
	Stderr  OutputWriter
}

// Production implementations

type realCommandRunner struct{}

func (r *realCommandRunner) RunContext(ctx context.Context, dir, name string, args ...string) (*CommandOutput, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	// Capture stdout and stderr separately
	var stdout, stderr []byte
	var err error

	// Get stdout
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdout pipe: %w", err)
	}

	// Get stderr
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("create stderr pipe: %w", err)
	}

	// Start the command
	if startErr := cmd.Start(); startErr != nil {
		return nil, fmt.Errorf("start command %s: %w", name, startErr)
	}

	// Read outputs
	stdout, _ = io.ReadAll(stdoutPipe)
	stderr, _ = io.ReadAll(stderrPipe)

	// Wait for completion
	err = cmd.Wait()

	output := &CommandOutput{
		Stdout: stdout,
		Stderr: stderr,
	}

	if err != nil {
		return output, fmt.Errorf("run command %s: %w", name, err)
	}

	return output, nil
}

func (r *realCommandRunner) LookPath(file string) (string, error) {
	path, err := exec.LookPath(file)
	if err != nil {
		return "", fmt.Errorf("look path %s: %w", file, err)
	}
	return path, nil
}

type realProcessManager struct{}

func (r *realProcessManager) GetPID() int {
	return os.Getpid()
}

func (r *realProcessManager) FindProcess(pid int) (*os.Process, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("find process %d: %w", pid, err)
	}
	return process, nil
}

func (r *realProcessManager) ProcessExists(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, FindProcess always succeeds, so we need to send signal 0
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

type realClock struct{}

func (r *realClock) Now() time.Time {
	return time.Now()
}

type stdinReader struct{}

func (s *stdinReader) ReadAll() ([]byte, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("read stdin: %w", err)
	}
	return data, nil
}

func (s *stdinReader) IsTerminal() bool {
	fileInfo, _ := os.Stdin.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// NewDefaultDependencies creates production dependencies.
func NewDefaultDependencies() *Dependencies {
	return &Dependencies{
		FS:      &shared.RealFS{},
		Runner:  &realCommandRunner{},
		Process: &realProcessManager{},
		Clock:   &realClock{},
		Input:   &stdinReader{},
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}
}
