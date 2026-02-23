//go:build testmode

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/config"
)

func TestResolveValidateConfig(t *testing.T) {
	defaults := config.GetDefaultConfig()

	tests := []struct {
		name        string
		timeout     int
		cooldown    int
		envTimeout  string
		envCooldown string
		setupConfig func(t *testing.T)
		wantTimeout int
		wantCool    int
	}{
		{
			name:        "defaults pass through",
			timeout:     defaults.Validate.Timeout,
			cooldown:    defaults.Validate.Cooldown,
			wantTimeout: defaults.Validate.Timeout,
			wantCool:    defaults.Validate.Cooldown,
		},
		{
			name:        "env var overrides timeout",
			timeout:     defaults.Validate.Timeout,
			cooldown:    defaults.Validate.Cooldown,
			envTimeout:  "300",
			wantTimeout: 300,
			wantCool:    defaults.Validate.Cooldown,
		},
		{
			name:        "env var overrides cooldown",
			timeout:     defaults.Validate.Timeout,
			cooldown:    defaults.Validate.Cooldown,
			envCooldown: "15",
			wantTimeout: defaults.Validate.Timeout,
			wantCool:    15,
		},
		{
			name:        "invalid env timeout ignored",
			timeout:     defaults.Validate.Timeout,
			cooldown:    defaults.Validate.Cooldown,
			envTimeout:  "notanumber",
			wantTimeout: defaults.Validate.Timeout,
			wantCool:    defaults.Validate.Cooldown,
		},
		{
			name:        "zero env cooldown accepted",
			timeout:     defaults.Validate.Timeout,
			cooldown:    defaults.Validate.Cooldown,
			envCooldown: "0",
			wantTimeout: defaults.Validate.Timeout,
			wantCool:    0,
		},
		{
			name:        "negative env timeout ignored",
			timeout:     defaults.Validate.Timeout,
			cooldown:    defaults.Validate.Cooldown,
			envTimeout:  "-5",
			wantTimeout: defaults.Validate.Timeout,
			wantCool:    defaults.Validate.Cooldown,
		},
		{
			name:     "config file overrides defaults",
			timeout:  defaults.Validate.Timeout,
			cooldown: defaults.Validate.Cooldown,
			setupConfig: func(t *testing.T) {
				t.Helper()
				tmpDir := t.TempDir()
				t.Setenv("XDG_CONFIG_HOME", tmpDir)

				configDir := filepath.Join(tmpDir, "cc-tools")
				require.NoError(t, os.MkdirAll(configDir, 0o750))

				cfg := &config.Values{
					Validate: config.ValidateValues{
						Timeout:  120,
						Cooldown: 30,
					},
				}
				data, err := json.Marshal(cfg)
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(
					filepath.Join(configDir, "config.json"), data, 0o600,
				))
			},
			wantTimeout: 120,
			wantCool:    30,
		},
		{
			name:     "env var overrides config file",
			timeout:  defaults.Validate.Timeout,
			cooldown: defaults.Validate.Cooldown,
			setupConfig: func(t *testing.T) {
				t.Helper()
				tmpDir := t.TempDir()
				t.Setenv("XDG_CONFIG_HOME", tmpDir)

				configDir := filepath.Join(tmpDir, "cc-tools")
				require.NoError(t, os.MkdirAll(configDir, 0o750))

				cfg := &config.Values{
					Validate: config.ValidateValues{
						Timeout:  120,
						Cooldown: 30,
					},
				}
				data, err := json.Marshal(cfg)
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(
					filepath.Join(configDir, "config.json"), data, 0o600,
				))
			},
			envTimeout:  "500",
			wantTimeout: 500,
			wantCool:    30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Isolate config path so tests don't read the real config.
			if tt.setupConfig == nil {
				t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			} else {
				tt.setupConfig(t)
			}

			if tt.envTimeout != "" {
				t.Setenv("CC_TOOLS_HOOKS_VALIDATE_TIMEOUT_SECONDS", tt.envTimeout)
			}
			if tt.envCooldown != "" {
				t.Setenv("CC_TOOLS_HOOKS_VALIDATE_COOLDOWN_SECONDS", tt.envCooldown)
			}

			gotTimeout, gotCooldown := resolveValidateConfig(defaults, tt.timeout, tt.cooldown)
			assert.Equal(t, tt.wantTimeout, gotTimeout)
			assert.Equal(t, tt.wantCool, gotCooldown)
		})
	}
}
