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

// WriteToEnvFile appends an exported PREFERRED_PACKAGE_MANAGER to the given
// env file. Claude Code sources the file named by CLAUDE_ENV_FILE before each
// Bash command, so the variable must be exported to reach child processes.
// If the file already sets the variable, the existing value is preserved to
// respect the user's choice. The write is an append because other SessionStart
// hooks may add to the same file concurrently.
func WriteToEnvFile(envFilePath, manager string) error {
	data, err := os.ReadFile(envFilePath) // #nosec G304 -- path supplied by Claude Code
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read env file %s: %w", envFilePath, err)
	}

	if setsEnvVar(string(data)) {
		return nil
	}

	line := "export " + envVarName + "=" + manager + "\n"
	if len(data) > 0 && !strings.HasSuffix(string(data), "\n") {
		line = "\n" + line
	}

	//nolint:gosec // File permissions 0644 are appropriate for env files
	f, err := os.OpenFile(envFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open env file %s: %w", envFilePath, err)
	}

	if _, writeErr := f.WriteString(line); writeErr != nil {
		_ = f.Close()
		return fmt.Errorf("write env file %s: %w", envFilePath, writeErr)
	}

	if closeErr := f.Close(); closeErr != nil {
		return fmt.Errorf("close env file %s: %w", envFilePath, closeErr)
	}

	return nil
}

// setsEnvVar reports whether env file content already assigns the variable,
// with or without an export prefix.
func setsEnvVar(content string) bool {
	prefix := envVarName + "="
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimPrefix(strings.TrimSpace(scanner.Text()), "export ")
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}

	return false
}
