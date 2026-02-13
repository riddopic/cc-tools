package shared

import (
	"os"
	"path/filepath"
)

// ConfigDir returns the cc-tools configuration directory.
// Respects $XDG_CONFIG_HOME; defaults to ~/.config/cc-tools.
func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "cc-tools")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("/tmp", ".config", "cc-tools")
	}

	return filepath.Join(home, ".config", "cc-tools")
}
