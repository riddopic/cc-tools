// Package pkgmanager detects the preferred JavaScript package manager for a project.
package pkgmanager

import (
	"fmt"
	"os"
	"path/filepath"
)

// lockFileEntry maps a lock file name to its corresponding package manager.
type lockFileEntry struct {
	filename string
	manager  string
}

// lockFilePriority returns the lock file detection order. First match wins.
func lockFilePriority() []lockFileEntry {
	return []lockFileEntry{
		{filename: "bun.lock", manager: "bun"},
		{filename: "bun.lockb", manager: "bun"},
		{filename: "pnpm-lock.yaml", manager: "pnpm"},
		{filename: "yarn.lock", manager: "yarn"},
		{filename: "package-lock.json", manager: "npm"},
	}
}

// defaultManager is returned when no lock file is found and no env var is set.
const defaultManager = "npm"

// envVarName is the environment variable that overrides lock file detection.
const envVarName = "PREFERRED_PACKAGE_MANAGER"

// Detect returns the preferred package manager for the given project directory.
// Detection priority: PREFERRED_PACKAGE_MANAGER env var, then lock file, then default "npm".
func Detect(projectDir string) string {
	if envVal := os.Getenv(envVarName); envVal != "" {
		return envVal
	}

	for _, entry := range lockFilePriority() {
		lockPath := filepath.Join(projectDir, entry.filename)
		if _, err := os.Stat(lockPath); err == nil {
			return entry.manager
		}
	}

	return defaultManager
}

// WriteToEnvFile writes the PREFERRED_PACKAGE_MANAGER to the specified env file
// so it persists across Bash commands in the Claude Code session.
func WriteToEnvFile(envFilePath, manager string) error {
	//nolint:gosec // File permissions 0644 are appropriate for env files
	f, err := os.OpenFile(envFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open env file %s: %w", envFilePath, err)
	}
	defer f.Close()

	line := envVarName + "=" + manager + "\n"

	_, writeErr := f.WriteString(line)
	if writeErr != nil {
		return fmt.Errorf("write to env file %s: %w", envFilePath, writeErr)
	}

	return nil
}
