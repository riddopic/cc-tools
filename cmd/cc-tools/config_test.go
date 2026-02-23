//go:build testmode

package main

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/output"
)

func newTestConfigManager(t *testing.T) *config.Manager {
	t.Helper()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	return config.NewManagerWithPath(configPath)
}

func newTestTerminal(t *testing.T) (*output.Terminal, *bytes.Buffer) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	out := output.NewTerminal(&stdout, &stderr)
	return out, &stdout
}

func TestHandleConfigGet(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "valid key returns value",
			key:        "validate.timeout",
			wantErr:    false,
			wantOutput: "60",
		},
		{
			name:       "unknown key returns error",
			key:        "nonexistent.key",
			wantErr:    true,
			wantOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := newTestConfigManager(t)
			out, stdout := newTestTerminal(t)
			ctx := context.Background()

			err := handleConfigGet(ctx, out, mgr, tt.key)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Contains(t, stdout.String(), tt.wantOutput)
		})
	}
}

func TestHandleConfigSet(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
	}{
		{
			name:    "set valid integer key",
			key:     "validate.timeout",
			value:   "120",
			wantErr: false,
		},
		{
			name:    "set valid string key",
			key:     "notifications.ntfy_topic",
			value:   "my-topic",
			wantErr: false,
		},
		{
			name:    "set unknown key returns error",
			key:     "unknown.key",
			value:   "value",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := newTestConfigManager(t)
			out, stdout := newTestTerminal(t)
			ctx := context.Background()

			err := handleConfigSet(ctx, out, mgr, tt.key, tt.value)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Contains(t, stdout.String(), tt.key)

			// Verify the value was persisted by reading it back.
			getOut, getStdout := newTestTerminal(t)
			getErr := handleConfigGet(ctx, getOut, mgr, tt.key)
			require.NoError(t, getErr)
			assert.Contains(t, getStdout.String(), tt.value)
		})
	}
}

func TestHandleConfigList(t *testing.T) {
	mgr := newTestConfigManager(t)
	out, stdout := newTestTerminal(t)
	ctx := context.Background()

	err := handleConfigList(ctx, out, mgr)
	require.NoError(t, err)

	result := stdout.String()
	assert.Contains(t, result, "validate.timeout")
	assert.Contains(t, result, "validate.cooldown")
	assert.Contains(t, result, "notifications.ntfy_topic")
}

func TestHandleConfigReset(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "reset specific key",
			key:     "validate.timeout",
			wantErr: false,
		},
		{
			name:    "reset all keys",
			key:     "",
			wantErr: false,
		},
		{
			name:    "reset unknown key returns error",
			key:     "unknown.key",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := newTestConfigManager(t)
			ctx := context.Background()

			// Set a non-default value first.
			if tt.key == "validate.timeout" || tt.key == "" {
				setOut, _ := newTestTerminal(t)
				setErr := handleConfigSet(ctx, setOut, mgr, "validate.timeout", "999")
				require.NoError(t, setErr)
			}

			out, stdout := newTestTerminal(t)
			err := handleConfigReset(ctx, out, mgr, tt.key)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Contains(t, stdout.String(), "Reset")

			// Verify reset restored default for specific key.
			if tt.key == "validate.timeout" {
				getOut, getStdout := newTestTerminal(t)
				getErr := handleConfigGet(ctx, getOut, mgr, "validate.timeout")
				require.NoError(t, getErr)
				assert.Contains(t, getStdout.String(), "60")
			}
		})
	}
}

// Command-execution tests exercise the Cobra RunE wrappers to cover
// the newTerminal → newConfigManager → handler delegation path.

func TestConfigGetCmd(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cmd := newConfigGetCmd()
	require.NoError(t, cmd.RunE(cmd, []string{"validate.timeout"}))
}

func TestConfigSetCmd(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cmd := newConfigSetCmd()
	require.NoError(t, cmd.RunE(cmd, []string{"validate.timeout", "90"}))
}

func TestConfigListCmd(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cmd := newConfigListCmd()
	require.NoError(t, cmd.RunE(cmd, nil))
}

func TestConfigResetCmd(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cmd := newConfigResetCmd()
	require.NoError(t, cmd.RunE(cmd, nil))
}

func TestConfigResetCmd_WithKey(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cmd := newConfigResetCmd()
	require.NoError(t, cmd.RunE(cmd, []string{"validate.timeout"}))
}
