package shared_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/shared"
)

func TestConfigDir(t *testing.T) {
	t.Run("returns XDG_CONFIG_HOME when set", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/custom/config")
		got := shared.ConfigDir()
		assert.Equal(t, filepath.Join("/custom/config", "cc-tools"), got)
	})

	t.Run("defaults to ~/.config/cc-tools", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "")
		home, err := os.UserHomeDir()
		require.NoError(t, err)
		got := shared.ConfigDir()
		assert.Equal(t, filepath.Join(home, ".config", "cc-tools"), got)
	})
}
