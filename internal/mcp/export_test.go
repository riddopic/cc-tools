package mcp

import (
	"context"

	"github.com/riddopic/cc-tools/internal/output"
)

// NewTestManager creates a Manager with explicit fields for use in external tests.
func NewTestManager(settingsPath string, out *output.Terminal, executor CommandExecutor) *Manager {
	return &Manager{
		settingsPath: settingsPath,
		output:       out,
		executor:     executor,
	}
}

// ManagerSettingsPath returns the settingsPath field for test assertions.
func ManagerSettingsPath(m *Manager) string {
	return m.settingsPath
}

// ManagerOutput returns the output field for test assertions.
func ManagerOutput(m *Manager) *output.Terminal {
	return m.output
}

// ManagerHasExecutor reports whether the Manager has a non-nil executor.
func ManagerHasExecutor(m *Manager) bool {
	return m.executor != nil
}

// ManagerExecutorIs reports whether the Manager executor matches the given instance.
func ManagerExecutorIs(m *Manager, executor CommandExecutor) bool {
	return m.executor == executor
}

// ManagerLoadSettings exposes the unexported loadSettings method for testing.
func ManagerLoadSettings(m *Manager) (*Settings, error) {
	return m.loadSettings()
}

// ManagerFindMCPByName exposes the unexported findMCPByName method for testing.
func ManagerFindMCPByName(m *Manager, settings *Settings, name string) (string, *Server, error) {
	return m.findMCPByName(settings, name)
}

// ManagerRemoveMCP exposes the unexported removeMCP method for testing.
func ManagerRemoveMCP(ctx context.Context, m *Manager, name string) error {
	return m.removeMCP(ctx, name)
}
