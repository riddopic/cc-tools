// Package pkgmanager detects the preferred JavaScript package manager for a project.
package pkgmanager

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// DetectWithPreferred returns the preferred package manager, using the config
// value if set, otherwise falling back to Detect (env var → lock file → default).
func DetectWithPreferred(projectDir, preferred string) string {
	if preferred != "" {
		return preferred
	}
	return Detect(projectDir)
}

// WriteToEnvFile writes the PREFERRED_PACKAGE_MANAGER to the specified env file
// so it persists across Bash commands in the Claude Code session.
// If the file already contains a PREFERRED_PACKAGE_MANAGER line, the existing
// value is preserved to respect the user's choice.
func WriteToEnvFile(envFilePath, manager string) error {
	prefix := envVarName + "="

	data, err := os.ReadFile(envFilePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read env file %s: %w", envFilePath, err)
	}

	if err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), prefix) {
				return nil // already set — respect existing value
			}
		}
	}

	var content string
	if len(data) > 0 && !strings.HasSuffix(string(data), "\n") {
		content = string(data) + "\n" + prefix + manager + "\n"
	} else {
		content = string(data) + prefix + manager + "\n"
	}

	//nolint:gosec // File permissions 0644 are appropriate for env files
	if writeErr := os.WriteFile(envFilePath, []byte(content), 0o644); writeErr != nil {
		return fmt.Errorf("write env file %s: %w", envFilePath, writeErr)
	}

	return nil
}
