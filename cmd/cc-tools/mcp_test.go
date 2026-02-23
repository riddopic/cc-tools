//go:build testmode

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/mcp"
	"github.com/riddopic/cc-tools/internal/output"
)

// testCommandExecutor is a mock that captures commands without running them.
type testCommandExecutor struct {
	// output is returned by CombinedOutput / Output calls.
	output []byte
	err    error
}

// CommandContext returns a command that runs "echo" with the mock output,
// or "false" to simulate an error.
func (e *testCommandExecutor) CommandContext(_ context.Context, _ string, _ ...string) *exec.Cmd {
	if e.err != nil {
		// Use a command guaranteed to fail.
		return exec.Command("false")
	}
	if len(e.output) > 0 {
		return exec.Command("echo", "-n", string(e.output))
	}
	return exec.Command("true")
}

// newTestMCPManager creates an isolated MCP manager rooted in a temp directory
// with a mock command executor to avoid running real `claude` CLI commands.
func newTestMCPManager(t *testing.T, executor mcp.CommandExecutor) (*mcp.Manager, string) {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	claudeDir := filepath.Join(tmpDir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0o750))

	out := output.NewTerminal(&bytes.Buffer{}, &bytes.Buffer{})
	return mcp.NewManagerWithExecutor(out, executor), claudeDir
}

func writeSettings(t *testing.T, claudeDir string, settings *mcp.Settings) {
	t.Helper()
	data, err := json.MarshalIndent(settings, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0o600))
}

func TestListMCPServers(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		executor := &testCommandExecutor{output: []byte("server1: connected\n")}
		mgr, _ := newTestMCPManager(t, executor)
		ctx := context.Background()

		err := listMCPServers(ctx, mgr)
		require.NoError(t, err)
	})

	t.Run("command failure", func(t *testing.T) {
		executor := &testCommandExecutor{err: assert.AnError}
		mgr, _ := newTestMCPManager(t, executor)
		ctx := context.Background()

		err := listMCPServers(ctx, mgr)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "listing MCPs")
	})
}

func TestEnableMCPServer(t *testing.T) {
	t.Run("server found in settings", func(t *testing.T) {
		executor := &testCommandExecutor{output: []byte("enabled")}
		mgr, claudeDir := newTestMCPManager(t, executor)
		writeSettings(t, claudeDir, &mcp.Settings{
			MCPServers: map[string]mcp.Server{
				"jira": {
					Type:    "stdio",
					Command: "jira-mcp",
					Args:    []string{},
				},
			},
		})
		ctx := context.Background()

		err := enableMCPServer(ctx, mgr, "jira")
		require.NoError(t, err)
	})

	t.Run("server not found in settings", func(t *testing.T) {
		executor := &testCommandExecutor{}
		mgr, claudeDir := newTestMCPManager(t, executor)
		writeSettings(t, claudeDir, &mcp.Settings{
			MCPServers: map[string]mcp.Server{},
		})
		ctx := context.Background()

		err := enableMCPServer(ctx, mgr, "nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nonexistent")
	})

	t.Run("no settings file", func(t *testing.T) {
		executor := &testCommandExecutor{}
		mgr, _ := newTestMCPManager(t, executor)
		ctx := context.Background()

		err := enableMCPServer(ctx, mgr, "anything")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reading settings")
	})
}

func TestDisableMCPServer(t *testing.T) {
	t.Run("server found in settings", func(t *testing.T) {
		executor := &testCommandExecutor{output: []byte("disabled")}
		mgr, claudeDir := newTestMCPManager(t, executor)
		writeSettings(t, claudeDir, &mcp.Settings{
			MCPServers: map[string]mcp.Server{
				"jira": {
					Type:    "stdio",
					Command: "jira-mcp",
					Args:    []string{},
				},
			},
		})
		ctx := context.Background()

		err := disableMCPServer(ctx, mgr, "jira")
		require.NoError(t, err)
	})

	t.Run("no settings file falls back to name", func(t *testing.T) {
		executor := &testCommandExecutor{output: []byte("removed")}
		mgr, _ := newTestMCPManager(t, executor)
		ctx := context.Background()

		// Disable with no settings file uses the raw name.
		err := disableMCPServer(ctx, mgr, "some-server")
		require.NoError(t, err)
	})
}

func TestEnableAllMCPServers(t *testing.T) {
	t.Run("no settings file", func(t *testing.T) {
		executor := &testCommandExecutor{}
		mgr, _ := newTestMCPManager(t, executor)
		ctx := context.Background()

		err := enableAllMCPServers(ctx, mgr)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reading settings")
	})

	t.Run("empty servers", func(t *testing.T) {
		executor := &testCommandExecutor{}
		mgr, claudeDir := newTestMCPManager(t, executor)
		writeSettings(t, claudeDir, &mcp.Settings{
			MCPServers: map[string]mcp.Server{},
		})
		ctx := context.Background()

		err := enableAllMCPServers(ctx, mgr)
		require.NoError(t, err)
	})

	t.Run("with servers", func(t *testing.T) {
		executor := &testCommandExecutor{output: []byte("enabled")}
		mgr, claudeDir := newTestMCPManager(t, executor)
		writeSettings(t, claudeDir, &mcp.Settings{
			MCPServers: map[string]mcp.Server{
				"server-a": {Type: "stdio", Command: "a-mcp", Args: []string{}},
				"server-b": {Type: "stdio", Command: "b-mcp", Args: []string{}},
			},
		})
		ctx := context.Background()

		err := enableAllMCPServers(ctx, mgr)
		require.NoError(t, err)
	})
}

func TestDisableAllMCPServers(t *testing.T) {
	t.Run("no servers listed", func(t *testing.T) {
		// "claude mcp list" returns empty output → no servers to disable.
		executor := &testCommandExecutor{output: []byte("")}
		mgr, _ := newTestMCPManager(t, executor)
		ctx := context.Background()

		err := disableAllMCPServers(ctx, mgr)
		require.NoError(t, err)
	})

	t.Run("command failure", func(t *testing.T) {
		executor := &testCommandExecutor{err: assert.AnError}
		mgr, _ := newTestMCPManager(t, executor)
		ctx := context.Background()

		err := disableAllMCPServers(ctx, mgr)
		require.Error(t, err)
	})
}
