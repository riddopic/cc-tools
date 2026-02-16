//go:build testmode

package main

import (
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
