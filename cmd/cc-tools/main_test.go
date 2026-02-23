//go:build testmode

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRootCmd(t *testing.T) {
	cmd := newRootCmd()

	assert.Equal(t, "cc-tools", cmd.Use)
	assert.Equal(t, "Claude Code Tools", cmd.Short)
	assert.Equal(t, version, cmd.Version)
	assert.True(t, cmd.SilenceUsage)
	assert.True(t, cmd.SilenceErrors)

	expectedSubcommands := []string{
		"hook", "session", "config", "skip", "unskip",
		"debug", "mcp", "validate",
	}

	subcommandNames := make([]string, 0, len(cmd.Commands()))
	for _, sub := range cmd.Commands() {
		subcommandNames = append(subcommandNames, sub.Name())
	}

	for _, expected := range expectedSubcommands {
		assert.Contains(t, subcommandNames, expected, "root command should have %q subcommand", expected)
	}
}

func TestNewRootCmd_UnknownCommand(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"nonexistent"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}

func TestNewRootCmd_VersionFlag(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestWriteDebugLog(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// writeDebugLog uses getDebugLogPath() which derives the path from cwd.
	writeDebugLog([]string{"cc-tools", "hook"}, nil)

	logPath := getDebugLogPath()
	data, err := os.ReadFile(logPath)
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "cc-tools invoked")
	assert.Contains(t, content, "cc-tools hook")
	assert.Contains(t, content, "Stdin: (no data)")
}

func TestWriteDebugLog_WithStdin(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	writeDebugLog([]string{"cc-tools", "validate"}, []byte(`{"tool_input":{}}`))

	logPath := getDebugLogPath()
	data, err := os.ReadFile(logPath)
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "cc-tools validate")
	assert.Contains(t, content, `{"tool_input":{}}`)
}

func TestGetDebugLogPath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	path := getDebugLogPath()
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "cc-tools")
	// The path should be based on the current working directory.
	assert.True(t, filepath.IsAbs(path))
}
