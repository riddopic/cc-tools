package debug_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/riddopic/cc-tools/internal/debug"
)

// setupEnabledLogger creates a test logger with file output for writing tests.
func setupEnabledLogger(t *testing.T) (*debug.Logger, string) {
	t.Helper()

	tmpFile := filepath.Join(t.TempDir(), "test.log")

	file, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	logger := debug.NewTestLogger(file, tmpFile, true)

	return logger, tmpFile
}

// readLogContent reads and returns the log file content as a string.
func readLogContent(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	return string(content)
}

// assertLogContains is a test helper that checks log content for a substring.
func assertLogContains(t *testing.T, content, substr string) {
	t.Helper()

	if !strings.Contains(content, substr) {
		t.Errorf("Log should contain %q, got %q", substr, content)
	}
}

// assertLogEmpty is a test helper that checks the log file is empty.
func assertLogEmpty(t *testing.T, content string) {
	t.Helper()

	if content != "" {
		t.Errorf("Log should be empty, got %q", content)
	}
}

// setupNewLoggerTest sets up the debug config so NewLogger reads from a temp dir.
func setupNewLoggerTest(t *testing.T, enabledDirs map[string]bool) {
	t.Helper()

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	configDir := filepath.Join(tmpHome, ".claude")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	config := &debug.Config{EnabledDirs: enabledDirs}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	configPath := filepath.Join(configDir, "debug-config.json")
	if writeErr := os.WriteFile(configPath, data, 0o600); writeErr != nil {
		t.Fatalf("Failed to write config: %v", writeErr)
	}
}

func TestNewLogger(t *testing.T) {
	ctx := context.Background()

	t.Run("creates enabled logger when debug is on", func(t *testing.T) {
		workDir, _ := os.Getwd()
		absDir, _ := filepath.Abs(workDir)

		setupNewLoggerTest(t, map[string]bool{absDir: true})

		logger, err := debug.NewLogger(ctx, workDir)
		if err != nil {
			t.Fatalf("NewLogger() error = %v", err)
		}
		defer logger.Close()

		if !logger.LoggerEnabled() {
			t.Error("logger.enabled = false, want true")
		}

		if logger.LoggerFile() == nil {
			t.Error("Enabled logger should have file handle")
		}

		if logger.LoggerFilePath() == "" {
			t.Error("Enabled logger should have file path")
		}
	})

	t.Run("creates disabled logger when debug is off", func(t *testing.T) {
		setupNewLoggerTest(t, map[string]bool{})

		logger, err := debug.NewLogger(ctx, ".")
		if err != nil {
			t.Fatalf("NewLogger() error = %v", err)
		}
		defer logger.Close()

		if logger.LoggerEnabled() {
			t.Error("logger.enabled = true, want false")
		}
	})

	t.Run("creates enabled logger when parent directory is enabled", func(t *testing.T) {
		parentDir, _ := filepath.Abs(".")

		setupNewLoggerTest(t, map[string]bool{parentDir: true})

		logger, err := debug.NewLogger(ctx, "./subdir")
		if err != nil {
			t.Fatalf("NewLogger() error = %v", err)
		}
		defer logger.Close()

		if !logger.LoggerEnabled() {
			t.Error("logger.enabled = false, want true")
		}
	})
}

func TestLoggerLogf(t *testing.T) {
	t.Run("logs message when enabled", func(t *testing.T) {
		logger, tmpFile := setupEnabledLogger(t)

		logger.Logf("Test message: %s", "value")
		logger.Close()

		content := readLogContent(t, tmpFile)
		assertLogContains(t, content, "Test message: value")
		assertLogContains(t, content, "[20")
		assertLogContains(t, content, "]")
	})

	t.Run("does not log when disabled", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "test.log")
		logger := debug.NewTestLogger(nil, tmpFile, false)

		logger.Logf("Should not appear")

		content, _ := os.ReadFile(tmpFile)
		assertLogEmpty(t, string(content))
	})

	t.Run("handles multiple arguments", func(t *testing.T) {
		logger, tmpFile := setupEnabledLogger(t)

		logger.Logf("Values: %d, %s, %v", 42, "test", true)
		logger.Close()

		content := readLogContent(t, tmpFile)
		assertLogContains(t, content, "Values: 42, test, true")
	})
}

func TestLoggerLogSection(t *testing.T) {
	logger, tmpFile := setupEnabledLogger(t)

	logger.LogSection("TEST SECTION")
	logger.Close()

	content := readLogContent(t, tmpFile)
	assertLogContains(t, content, "========== TEST SECTION ==========")
}

func TestLoggerLogError(t *testing.T) {
	t.Run("logs error with context", func(t *testing.T) {
		logger, tmpFile := setupEnabledLogger(t)

		logger.LogError(os.ErrNotExist, "file operation")
		logger.Close()

		content := readLogContent(t, tmpFile)
		assertLogContains(t, content, "ERROR in file operation:")
	})

	t.Run("handles nil error", func(t *testing.T) {
		logger, tmpFile := setupEnabledLogger(t)

		logger.LogError(nil, "operation")
		logger.Close()

		content := readLogContent(t, tmpFile)
		assertLogEmpty(t, content)
	})
}

func TestLoggerLogCommand(t *testing.T) {
	logger, tmpFile := setupEnabledLogger(t)

	logger.LogCommand("make", []string{"test", "-v"}, "/project/dir")
	logger.Close()

	content := readLogContent(t, tmpFile)
	assertLogContains(t, content, "Executing command: make")
	assertLogContains(t, content, "Args: [test -v]")
	assertLogContains(t, content, "Working dir: /project/dir")
}

func TestLoggerLogDiscovery(t *testing.T) {
	t.Run("logs successful discovery", func(t *testing.T) {
		logger, tmpFile := setupEnabledLogger(t)

		logger.LogDiscovery("lint", "make lint", "/project")
		logger.Close()

		content := readLogContent(t, tmpFile)
		assertLogContains(t, content, "Discovery for lint in /project")
		assertLogContains(t, content, "Found: make lint")
	})

	t.Run("logs failed discovery", func(t *testing.T) {
		logger, tmpFile := setupEnabledLogger(t)

		logger.LogDiscovery("test", "", "/project")
		logger.Close()

		content := readLogContent(t, tmpFile)
		assertLogContains(t, content, "Discovery for test in /project")
		assertLogContains(t, content, "Not found")
	})
}

func TestLoggerClose(t *testing.T) {
	t.Run("closes file successfully", func(t *testing.T) {
		logger, _ := setupEnabledLogger(t)

		err := logger.Close()
		if err != nil {
			t.Errorf("Close() error = %v", err)
		}

		if logger.LoggerFile() != nil {
			t.Error("File handle should be nil after close")
		}

		err2 := logger.Close()
		if err2 != nil {
			t.Error("Second close should not return error")
		}
	})

	t.Run("handles nil file", func(t *testing.T) {
		logger := debug.NewTestLogger(nil, "", false)

		err := logger.Close()
		if err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
}

func TestLoggerIsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		want    bool
	}{
		{
			name:    "returns true when enabled",
			enabled: true,
			want:    true,
		},
		{
			name:    "returns false when disabled",
			enabled: false,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := debug.NewTestLogger(nil, "", tt.enabled)
			if got := logger.IsEnabled(); got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoggerConcurrency(t *testing.T) {
	logger, tmpFile := setupEnabledLogger(t)
	defer logger.Close()

	done := make(chan bool, 4)

	go func() {
		for i := range 100 {
			logger.Logf("Message %d", i)
		}
		done <- true
	}()

	go func() {
		for range 50 {
			logger.LogSection("Section")
		}
		done <- true
	}()

	go func() {
		for range 50 {
			logger.LogError(os.ErrExist, "test")
		}
		done <- true
	}()

	go func() {
		for range 50 {
			logger.LogCommand("cmd", []string{"arg"}, "/dir")
		}
		done <- true
	}()

	for range 4 {
		<-done
	}

	content := readLogContent(t, tmpFile)
	if len(content) == 0 {
		t.Error("Log file should have content after concurrent writes")
	}
}

func TestLoggerTimestampFormat(t *testing.T) {
	logger, tmpFile := setupEnabledLogger(t)

	beforeLog := time.Now()
	logger.Logf("Test message")
	logger.Close()

	content := readLogContent(t, tmpFile)

	start := strings.Index(content, "[")
	end := strings.Index(content, "]")

	if start == -1 || end == -1 || start >= end {
		t.Fatal("Could not find timestamp in log")
	}

	timestamp := content[start+1 : end]

	parsed, err := time.ParseInLocation("2006-01-02 15:04:05.000", timestamp, time.Local)
	if err != nil {
		t.Errorf("Failed to parse timestamp %q: %v", timestamp, err)
	}

	diff := parsed.Sub(beforeLog).Abs()
	if diff > time.Second {
		t.Errorf("Timestamp seems incorrect, diff = %v", diff)
	}
}

func TestLoggerFilePermissions(t *testing.T) {
	ctx := context.Background()

	workDir, _ := os.Getwd()
	absDir, _ := filepath.Abs(workDir)

	setupNewLoggerTest(t, map[string]bool{absDir: true})

	logger, err := debug.NewLogger(ctx, workDir)
	if err != nil {
		t.Skipf("Could not create logger: %v", err)
	}
	defer logger.Close()

	if !logger.LoggerEnabled() {
		t.Skip("Logger not enabled")
	}

	info, statErr := os.Stat(logger.LoggerFilePath())
	if statErr != nil {
		t.Fatalf("Failed to stat log file: %v", statErr)
	}

	mode := info.Mode()
	if mode.Perm() != 0o600 {
		t.Errorf("Log file permissions = %v, want 0600", mode.Perm())
	}
}

func TestLoggerDisabledOperations(t *testing.T) {
	logger := debug.NewTestLogger(nil, "", false)

	logger.Logf("test")
	logger.LogSection("section")
	logger.LogError(os.ErrNotExist, "context")
	logger.LogCommand("cmd", []string{"arg"}, "/dir")
	logger.LogDiscovery("type", "result", "/dir")

	if logger.IsEnabled() {
		t.Error("Disabled logger should report as disabled")
	}

	err := logger.Close()
	if err != nil {
		t.Errorf("Close on disabled logger should not error: %v", err)
	}
}
