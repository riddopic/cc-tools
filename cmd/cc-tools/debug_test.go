//go:build testmode

package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/debug"
	"github.com/riddopic/cc-tools/internal/output"
)

func newDebugTestTerminal(t *testing.T) (*output.Terminal, *bytes.Buffer) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	out := output.NewTerminal(&stdout, &stderr)
	return out, &stdout
}

func newIsolatedDebugManager(t *testing.T) *debug.Manager {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	return debug.NewManager()
}

func TestEnableDebug(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	mgr := newIsolatedDebugManager(t)
	out, stdout := newDebugTestTerminal(t)
	ctx := context.Background()

	err := enableDebug(ctx, out, mgr)
	require.NoError(t, err)

	outputStr := stdout.String()
	assert.Contains(t, outputStr, "Debug logging enabled")
	assert.Contains(t, outputStr, "Log file")
}

func TestDisableDebug(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	mgr := newIsolatedDebugManager(t)
	ctx := context.Background()

	// Enable first so there is something to disable.
	enableOut, _ := newDebugTestTerminal(t)
	require.NoError(t, enableDebug(ctx, enableOut, mgr))

	out, stdout := newDebugTestTerminal(t)
	err := disableDebug(ctx, out, mgr)
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Debug logging disabled")
}

func TestShowDebugStatus(t *testing.T) {
	t.Run("disabled state", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		mgr := newIsolatedDebugManager(t)
		out, stdout := newDebugTestTerminal(t)
		ctx := context.Background()

		err := showDebugStatus(ctx, out, mgr)
		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "DISABLED")
	})

	t.Run("enabled state", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		mgr := newIsolatedDebugManager(t)
		ctx := context.Background()

		// Enable debug first.
		enableOut, _ := newDebugTestTerminal(t)
		require.NoError(t, enableDebug(ctx, enableOut, mgr))

		out, stdout := newDebugTestTerminal(t)
		err := showDebugStatus(ctx, out, mgr)
		require.NoError(t, err)

		outputStr := stdout.String()
		assert.Contains(t, outputStr, "ENABLED")
		assert.Contains(t, outputStr, "Log file")
	})
}

func TestListDebugDirs(t *testing.T) {
	t.Run("no directories enabled", func(t *testing.T) {
		mgr := newIsolatedDebugManager(t)
		out, stdout := newDebugTestTerminal(t)
		ctx := context.Background()

		err := listDebugDirs(ctx, out, mgr)
		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "No directories have debug logging enabled")
	})

	t.Run("with enabled directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		mgr := newIsolatedDebugManager(t)
		ctx := context.Background()

		// Enable debug for a directory.
		enableOut, _ := newDebugTestTerminal(t)
		require.NoError(t, enableDebug(ctx, enableOut, mgr))

		out, stdout := newDebugTestTerminal(t)
		err := listDebugDirs(ctx, out, mgr)
		require.NoError(t, err)

		outputStr := stdout.String()
		assert.Contains(t, outputStr, "Directories with debug logging enabled")
		assert.Contains(t, outputStr, "Directory")
	})
}

func TestShowDebugFilename(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	out, stdout := newDebugTestTerminal(t)
	err := showDebugFilename(out)
	require.NoError(t, err)

	outputStr := stdout.String()
	assert.NotEmpty(t, outputStr)
	assert.Contains(t, outputStr, "cc-tools")
}
