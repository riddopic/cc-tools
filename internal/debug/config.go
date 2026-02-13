// Package debug provides debug configuration management for cc-tools.
package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/riddopic/cc-tools/internal/shared"
)

// Config represents debug configuration settings.
type Config struct {
	EnabledDirs map[string]bool `json:"enabled_dirs"`
}

// Manager handles debug configuration persistence.
type Manager struct {
	mu       sync.RWMutex
	config   *Config
	filepath string
}

// NewManager creates a new debug configuration manager.
func NewManager() *Manager {
	migrateDebugConfigIfNeeded()

	configPath := filepath.Join(getConfigDir(), "debug-config.json")
	return &Manager{
		mu:       sync.RWMutex{},
		config:   &Config{EnabledDirs: make(map[string]bool)},
		filepath: configPath,
	}
}

// Load reads debug configuration from disk.
func (m *Manager) Load(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read debug config: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	var config Config
	if unmarshalErr := json.Unmarshal(data, &config); unmarshalErr != nil {
		return fmt.Errorf("parse debug config: %w", unmarshalErr)
	}

	if config.EnabledDirs == nil {
		config.EnabledDirs = make(map[string]bool)
	}

	m.config = &config
	return nil
}

// Save writes debug configuration to disk.
func (m *Manager) Save(_ context.Context) error {
	m.mu.RLock()
	data, err := json.MarshalIndent(m.config, "", "  ")
	m.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("marshal debug config: %w", err)
	}

	dir := filepath.Dir(m.filepath)
	if mkdirErr := os.MkdirAll(dir, 0o750); mkdirErr != nil {
		return fmt.Errorf("create config dir: %w", mkdirErr)
	}

	data = append(data, '\n')

	tempFile := m.filepath + ".tmp"
	if writeErr := os.WriteFile(tempFile, data, 0o600); writeErr != nil {
		return fmt.Errorf("write temp file: %w", writeErr)
	}

	if renameErr := os.Rename(tempFile, m.filepath); renameErr != nil {
		_ = os.Remove(tempFile)
		return fmt.Errorf("rename config file: %w", renameErr)
	}

	return nil
}

// Enable turns on debug logging for a directory and returns the log file path.
func (m *Manager) Enable(ctx context.Context, dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("get absolute path: %w", err)
	}

	if loadErr := m.Load(ctx); loadErr != nil {
		return "", loadErr
	}

	m.mu.Lock()
	m.config.EnabledDirs[absDir] = true
	m.mu.Unlock()

	if saveErr := m.Save(ctx); saveErr != nil {
		return "", saveErr
	}

	logFile := GetLogFilePath(absDir)
	return logFile, nil
}

// Disable turns off debug logging for a directory.
func (m *Manager) Disable(ctx context.Context, dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("get absolute path: %w", err)
	}

	if loadErr := m.Load(ctx); loadErr != nil {
		return loadErr
	}

	m.mu.Lock()
	delete(m.config.EnabledDirs, absDir)
	m.mu.Unlock()

	return m.Save(ctx)
}

// IsEnabled checks if debug logging is enabled for a directory or any parent.
func (m *Manager) IsEnabled(ctx context.Context, dir string) (bool, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false, fmt.Errorf("get absolute path: %w", err)
	}

	if loadErr := m.Load(ctx); loadErr != nil {
		return false, loadErr
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for enabledDir := range m.config.EnabledDirs {
		if strings.HasPrefix(absDir, enabledDir) {
			return true, nil
		}
	}

	return false, nil
}

// GetEnabledDirs returns all directories with debug logging enabled.
func (m *Manager) GetEnabledDirs(ctx context.Context) ([]string, error) {
	if loadErr := m.Load(ctx); loadErr != nil {
		return nil, loadErr
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	dirs := make([]string, 0, len(m.config.EnabledDirs))
	for dir := range m.config.EnabledDirs {
		dirs = append(dirs, dir)
	}

	return dirs, nil
}

// GetLogFilePath generates a log file path for a directory.
// Delegates to shared.GetDebugLogPathForDir for consistent naming.
func GetLogFilePath(dir string) string {
	return shared.GetDebugLogPathForDir(dir)
}

func getConfigDir() string {
	return shared.ConfigDir()
}

// migrateDebugConfigIfNeeded copies debug-config.json from ~/.claude/ to the
// new config dir if the old file exists and the new one does not.
func migrateDebugConfigIfNeeded() {
	newPath := filepath.Join(shared.ConfigDir(), "debug-config.json")
	if _, err := os.Stat(newPath); err == nil {
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	oldPath := filepath.Join(home, ".claude", "debug-config.json")

	data, err := os.ReadFile(oldPath) // #nosec G304 - file path is constructed from home dir
	if err != nil {
		return
	}

	dir := filepath.Dir(newPath)
	_ = os.MkdirAll(dir, 0o750)
	_ = os.WriteFile(newPath, data, 0o600)
}
