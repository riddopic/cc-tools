package hooks_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/riddopic/cc-tools/internal/hooks"
)

// requireAcquireSuccess calls TryAcquire and asserts it succeeds without error.
func requireAcquireSuccess(t *testing.T, lm *hooks.LockManager) {
	t.Helper()

	acquired, err := lm.TryAcquire()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !acquired {
		t.Fatal("Expected to acquire lock")
	}
}

// requireAcquireBlocked calls TryAcquire and asserts that the lock was not acquired, without error.
func requireAcquireBlocked(t *testing.T, lm *hooks.LockManager) {
	t.Helper()

	acquired, err := lm.TryAcquire()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if acquired {
		t.Fatal("Should not have acquired lock")
	}
}

// requireRelease calls Release and asserts no error.
func requireRelease(t *testing.T, lm *hooks.LockManager) {
	t.Helper()

	if err := lm.Release(); err != nil {
		t.Fatalf("Unexpected error releasing lock: %v", err)
	}
}

// assertTwoCreateExclusiveCalls verifies that createExclusive was called exactly twice
// (first attempt fails on existing file, second attempt succeeds after stale lock removal).
func assertTwoCreateExclusiveCalls(t *testing.T, got int) {
	t.Helper()

	const expected = 2
	if got != expected {
		t.Errorf("Expected %d createExclusive calls, got %d", expected, got)
	}
}

// setBasicFSMocks configures common filesystem mocks needed for lock operations.
func setBasicFSMocks(td *hooks.TestDependencies) {
	td.MockFS.TempDirFunc = func() string { return "/tmp" }
	td.MockProcess.GetPIDFunc = func() int { return 99999 }
	td.MockClock.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
}

func TestLockManagerWithMocks(t *testing.T) {
	t.Run("successful lock acquisition", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		setBasicFSMocks(testDeps)
		testDeps.MockFS.CreateExclusiveFunc = func(_ string, _ []byte, _ os.FileMode) error {
			return nil
		}

		lm := hooks.NewLockManager("/project", "test", 5, testDeps.Dependencies)
		requireAcquireSuccess(t, lm)
	})

	t.Run("lock held by running process", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		setBasicFSMocks(testDeps)
		testDeps.MockFS.CreateExclusiveFunc = func(_ string, _ []byte, _ os.FileMode) error {
			return errors.New("file exists")
		}
		testDeps.MockFS.ReadFileFunc = func(_ string) ([]byte, error) {
			return []byte("12345\n"), nil
		}
		testDeps.MockProcess.ProcessExistsFunc = func(pid int) bool {
			return pid == 12345
		}

		lm := hooks.NewLockManager("/project", "test", 5, testDeps.Dependencies)
		requireAcquireBlocked(t, lm)
	})

	t.Run("lock held by dead process", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		setBasicFSMocks(testDeps)

		var createExclusiveCallCount int

		testDeps.MockFS.CreateExclusiveFunc = func(_ string, _ []byte, _ os.FileMode) error {
			createExclusiveCallCount++
			if createExclusiveCallCount == 1 {
				return errors.New("file exists")
			}
			return nil
		}
		testDeps.MockFS.ReadFileFunc = func(_ string) ([]byte, error) {
			return []byte("12345\n"), nil
		}
		testDeps.MockFS.RemoveFunc = func(_ string) error {
			return nil
		}
		testDeps.MockProcess.ProcessExistsFunc = func(_ int) bool {
			return false
		}

		lm := hooks.NewLockManager("/project", "test", 5, testDeps.Dependencies)
		requireAcquireSuccess(t, lm)
		assertTwoCreateExclusiveCalls(t, createExclusiveCallCount)
	})

	t.Run("respects cooldown period", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		setBasicFSMocks(testDeps)
		testDeps.MockFS.CreateExclusiveFunc = func(_ string, _ []byte, _ os.FileMode) error {
			return errors.New("file exists")
		}
		testDeps.MockFS.ReadFileFunc = func(_ string) ([]byte, error) {
			return []byte("\n1700000099\n"), nil
		}
		testDeps.MockClock.NowFunc = func() time.Time {
			return time.Unix(1700000100, 0)
		}

		lm := hooks.NewLockManager("/project", "test", 5, testDeps.Dependencies)
		requireAcquireBlocked(t, lm)
	})

	t.Run("acquires after cooldown expires", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		setBasicFSMocks(testDeps)

		var createExclusiveCallCount int

		testDeps.MockFS.CreateExclusiveFunc = func(_ string, _ []byte, _ os.FileMode) error {
			createExclusiveCallCount++
			if createExclusiveCallCount == 1 {
				return errors.New("file exists")
			}
			return nil
		}
		testDeps.MockFS.ReadFileFunc = func(_ string) ([]byte, error) {
			return []byte("\n1700000094\n"), nil
		}
		testDeps.MockFS.RemoveFunc = func(_ string) error {
			return nil
		}
		testDeps.MockClock.NowFunc = func() time.Time {
			return time.Unix(1700000100, 0)
		}

		lm := hooks.NewLockManager("/project", "test", 5, testDeps.Dependencies)
		requireAcquireSuccess(t, lm)
		assertTwoCreateExclusiveCalls(t, createExclusiveCallCount)
	})

	t.Run("release writes timestamp", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		setBasicFSMocks(testDeps)

		var writtenData []byte

		testDeps.MockFS.WriteFileFunc = func(_ string, data []byte, _ os.FileMode) error {
			writtenData = data
			return nil
		}
		testDeps.MockClock.NowFunc = func() time.Time {
			return time.Unix(1700000200, 0)
		}

		lm := hooks.NewLockManager("/project", "test", 5, testDeps.Dependencies)
		requireRelease(t, lm)

		expected := "\n1700000200\n"
		if string(writtenData) != expected {
			t.Errorf("Expected written data %q, got %q", expected, string(writtenData))
		}
	})

	t.Run("handles write error on acquire", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		setBasicFSMocks(testDeps)
		testDeps.MockFS.CreateExclusiveFunc = func(_ string, _ []byte, _ os.FileMode) error {
			return errors.New("permission denied")
		}
		testDeps.MockFS.ReadFileFunc = func(_ string) ([]byte, error) {
			return nil, errors.New("file not found")
		}

		lm := hooks.NewLockManager("/project", "test", 5, testDeps.Dependencies)
		requireAcquireBlocked(t, lm)
	})

	t.Run("handles malformed lock file", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		setBasicFSMocks(testDeps)

		var createExclusiveCallCount int

		testDeps.MockFS.CreateExclusiveFunc = func(_ string, _ []byte, _ os.FileMode) error {
			createExclusiveCallCount++
			if createExclusiveCallCount == 1 {
				return errors.New("file exists")
			}
			return nil
		}
		testDeps.MockFS.ReadFileFunc = func(_ string) ([]byte, error) {
			return []byte("not-a-number\n"), nil
		}
		testDeps.MockFS.RemoveFunc = func(_ string) error {
			return nil
		}

		lm := hooks.NewLockManager("/project", "test", 5, testDeps.Dependencies)
		requireAcquireSuccess(t, lm)
		assertTwoCreateExclusiveCalls(t, createExclusiveCallCount)
	})

	t.Run("handles malformed timestamp", func(t *testing.T) {
		testDeps := hooks.CreateTestDependencies()
		setBasicFSMocks(testDeps)

		var createExclusiveCallCount int

		testDeps.MockFS.CreateExclusiveFunc = func(_ string, _ []byte, _ os.FileMode) error {
			createExclusiveCallCount++
			if createExclusiveCallCount == 1 {
				return errors.New("file exists")
			}
			return nil
		}
		testDeps.MockFS.ReadFileFunc = func(_ string) ([]byte, error) {
			return []byte("\nnot-a-timestamp\n"), nil
		}
		testDeps.MockFS.RemoveFunc = func(_ string) error {
			return nil
		}

		lm := hooks.NewLockManager("/project", "test", 5, testDeps.Dependencies)
		requireAcquireSuccess(t, lm)
		assertTwoCreateExclusiveCalls(t, createExclusiveCallCount)
	})
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "unix line endings",
			input:    "line1\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "windows line endings",
			input:    "line1\r\nline2\r\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "mixed line endings",
			input:    "line1\nline2\r\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "empty lines",
			input:    "\n\n",
			expected: []string{"", ""},
		},
		{
			name:     "no newline at end",
			input:    "line1\nline2",
			expected: []string{"line1", "line2"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hooks.SplitLinesForTest(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d", len(tt.expected), len(result))
				return
			}
			for i, line := range result {
				if line != tt.expected[i] {
					t.Errorf("Line %d: expected %q, got %q", i, tt.expected[i], line)
				}
			}
		})
	}
}
