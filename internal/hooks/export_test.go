package hooks

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"time"

	"github.com/riddopic/cc-tools/internal/hookcmd"
)

// MustMarshalJSON creates a [json.RawMessage] from a map (test helper).
func MustMarshalJSON(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return json.RawMessage(data)
}

// --- Mock implementations (exported for hooks_test package) ---

// MockFileSystem implements shared.HooksFS for testing.
type MockFileSystem struct {
	StatFunc            func(string) (os.FileInfo, error)
	ReadFileFunc        func(string) ([]byte, error)
	WriteFileFunc       func(string, []byte, os.FileMode) error
	TempDirFunc         func() string
	CreateExclusiveFunc func(string, []byte, os.FileMode) error
	RemoveFunc          func(string) error
}

func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	if m.StatFunc != nil {
		return m.StatFunc(name)
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) ReadFile(name string) ([]byte, error) {
	if m.ReadFileFunc != nil {
		return m.ReadFileFunc(name)
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	if m.WriteFileFunc != nil {
		return m.WriteFileFunc(name, data, perm)
	}
	return nil
}

func (m *MockFileSystem) TempDir() string {
	if m.TempDirFunc != nil {
		return m.TempDirFunc()
	}
	return "/tmp"
}

func (m *MockFileSystem) CreateExclusive(name string, data []byte, perm os.FileMode) error {
	if m.CreateExclusiveFunc != nil {
		return m.CreateExclusiveFunc(name, data, perm)
	}
	return nil
}

func (m *MockFileSystem) Remove(name string) error {
	if m.RemoveFunc != nil {
		return m.RemoveFunc(name)
	}
	return nil
}

// MockCommandRunner implements CommandRunner for testing.
type MockCommandRunner struct {
	RunContextFunc func(ctx context.Context, dir, name string, args ...string) (*CommandOutput, error)
	LookPathFunc   func(file string) (string, error)
}

func (m *MockCommandRunner) RunContext(
	ctx context.Context,
	dir, name string,
	args ...string,
) (*CommandOutput, error) {
	if m.RunContextFunc != nil {
		return m.RunContextFunc(ctx, dir, name, args...)
	}
	return nil, errors.New("command not found")
}

func (m *MockCommandRunner) LookPath(file string) (string, error) {
	if m.LookPathFunc != nil {
		return m.LookPathFunc(file)
	}
	return "", errors.New("command not found")
}

// MockProcessManager implements ProcessManager for testing.
type MockProcessManager struct {
	GetPIDFunc        func() int
	FindProcessFunc   func(pid int) (*os.Process, error)
	ProcessExistsFunc func(pid int) bool
}

func (m *MockProcessManager) GetPID() int {
	if m.GetPIDFunc != nil {
		return m.GetPIDFunc()
	}
	return 12345
}

func (m *MockProcessManager) FindProcess(pid int) (*os.Process, error) {
	if m.FindProcessFunc != nil {
		return m.FindProcessFunc(pid)
	}
	return nil, errors.New("process not found")
}

func (m *MockProcessManager) ProcessExists(pid int) bool {
	if m.ProcessExistsFunc != nil {
		return m.ProcessExistsFunc(pid)
	}
	return false
}

// MockClock implements Clock for testing.
type MockClock struct {
	NowFunc func() time.Time
}

func (m *MockClock) Now() time.Time {
	if m.NowFunc != nil {
		return m.NowFunc()
	}
	return time.Unix(1700000000, 0)
}

// MockOutputWriter implements OutputWriter for testing.
type MockOutputWriter struct {
	WrittenData []byte
}

func (m *MockOutputWriter) Write(p []byte) (int, error) {
	m.WrittenData = append(m.WrittenData, p...)
	return len(p), nil
}

func (m *MockOutputWriter) String() string {
	return string(m.WrittenData)
}

// TestDependencies wraps Dependencies with direct access to mock implementations.
type TestDependencies struct {
	*Dependencies

	MockFS      *MockFileSystem
	MockRunner  *MockCommandRunner
	MockProcess *MockProcessManager
	MockClock   *MockClock
	MockStdout  *MockOutputWriter
	MockStderr  *MockOutputWriter
}

// CreateTestDependencies creates test dependencies with mock implementations.
func CreateTestDependencies() *TestDependencies {
	fs := &MockFileSystem{
		StatFunc:            nil,
		ReadFileFunc:        nil,
		WriteFileFunc:       nil,
		TempDirFunc:         nil,
		CreateExclusiveFunc: nil,
		RemoveFunc:          nil,
	}
	runner := &MockCommandRunner{
		RunContextFunc: nil,
		LookPathFunc:   nil,
	}
	process := &MockProcessManager{
		GetPIDFunc:        nil,
		FindProcessFunc:   nil,
		ProcessExistsFunc: nil,
	}
	clock := &MockClock{
		NowFunc: nil,
	}
	stdout := &MockOutputWriter{
		WrittenData: nil,
	}
	stderr := &MockOutputWriter{
		WrittenData: nil,
	}

	return &TestDependencies{
		Dependencies: &Dependencies{
			FS:      fs,
			Runner:  runner,
			Process: process,
			Clock:   clock,
			Stdout:  stdout,
			Stderr:  stderr,
		},
		MockFS:      fs,
		MockRunner:  runner,
		MockProcess: process,
		MockClock:   clock,
		MockStdout:  stdout,
		MockStderr:  stderr,
	}
}

// MockFileInfo implements [os.FileInfo] for testing.
type MockFileInfo struct {
	FName    string
	FSize    int64
	FMode    os.FileMode
	FModTime time.Time
	FIsDir   bool
}

func (m MockFileInfo) Name() string       { return m.FName }
func (m MockFileInfo) Size() int64        { return m.FSize }
func (m MockFileInfo) Mode() os.FileMode  { return m.FMode }
func (m MockFileInfo) ModTime() time.Time { return m.FModTime }
func (m MockFileInfo) IsDir() bool        { return m.FIsDir }

func (m MockFileInfo) Sys() any { return nil }

// NewMockFileInfo creates a MockFileInfo for use in external test packages.
func NewMockFileInfo(name string, size int64, mode os.FileMode, modTime time.Time, isDir bool) os.FileInfo {
	return MockFileInfo{
		FName:    name,
		FSize:    size,
		FMode:    mode,
		FModTime: modTime,
		FIsDir:   isDir,
	}
}

// DefaultTime returns a zero-value time for use in test struct literals.
func DefaultTime() time.Time { return time.Time{} }

// DetectProjectTypesForTest exposes the unexported detectProjectTypes method.
func (cd *CommandDiscovery) DetectProjectTypesForTest(dir string) []string {
	return cd.detectProjectTypes(dir)
}

// DetectPackageManagerForTest exposes the unexported detectPackageManager method.
func (cd *CommandDiscovery) DetectPackageManagerForTest(dir string) string {
	return cd.detectPackageManager(dir)
}

// HandleInputErrorForTest exposes handleInputError for external test packages.
func HandleInputErrorForTest(err error, debug bool, stderr OutputWriter) {
	handleInputError(err, debug, stderr)
}

// ValidateHookEventForTest exposes validateHookEvent for external test packages.
func ValidateHookEventForTest(input *hookcmd.HookInput, debug bool, stderr OutputWriter) (string, bool) {
	return validateHookEvent(input, debug, stderr)
}

// SplitLinesForTest exposes splitLines for external test packages.
func SplitLinesForTest(s string) []string {
	return splitLines(s)
}

// CheckSkipsFromInputForTest exposes checkSkipsFromInput for external test packages.
func CheckSkipsFromInputForTest(
	ctx context.Context,
	input *hookcmd.HookInput,
	debug bool,
	stderr io.Writer,
) (bool, bool) {
	return checkSkipsFromInput(ctx, input, debug, stderr)
}

// SetCleanupOnExit sets the cleanupOnExit field on a LockManager for testing.
func (l *LockManager) SetCleanupOnExit(v bool) {
	l.cleanupOnExit = v
}

// LockFileForTest returns the lock file path for testing.
func (l *LockManager) LockFileForTest() string {
	return l.lockFile
}
