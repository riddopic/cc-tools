package debug

import (
	"os"
	"sync"
)

// NewTestManager creates a Manager with a custom config path for testing.
func NewTestManager(configPath string) *Manager {
	return &Manager{
		mu:       sync.RWMutex{},
		config:   &Config{EnabledDirs: make(map[string]bool)},
		filepath: configPath,
	}
}

// NewTestManagerWithConfig creates a Manager with a custom config and path for testing.
func NewTestManagerWithConfig(configPath string, config *Config) *Manager {
	return &Manager{
		mu:       sync.RWMutex{},
		config:   config,
		filepath: configPath,
	}
}

// ManagerConfig returns the manager's config for test assertions.
func (m *Manager) ManagerConfig() *Config {
	return m.config
}

// NewTestLogger creates a Logger with explicit fields for testing.
func NewTestLogger(file *os.File, filePath string, enabled bool) *Logger {
	return &Logger{
		mu:       sync.Mutex{},
		file:     file,
		filePath: filePath,
		enabled:  enabled,
	}
}

// LoggerEnabled returns the logger's enabled state for test assertions.
func (l *Logger) LoggerEnabled() bool {
	return l.enabled
}

// LoggerFile returns the logger's file handle for test assertions.
func (l *Logger) LoggerFile() *os.File {
	return l.file
}

// LoggerFilePath returns the logger's file path for test assertions.
func (l *Logger) LoggerFilePath() string {
	return l.filePath
}

// ExportGetConfigDir exposes getConfigDir for testing.
func ExportGetConfigDir() string {
	return getConfigDir()
}
