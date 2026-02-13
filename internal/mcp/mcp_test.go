package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/riddopic/cc-tools/internal/mcp"
	"github.com/riddopic/cc-tools/internal/output"
)

// mockCommandExecutor is a mock implementation of [mcp.CommandExecutor] for testing.
type mockCommandExecutor struct {
	capturedCmd  string
	capturedArgs []string
	mockOutput   string
	shouldFail   bool
	// commandHandler provides varying responses for complex tests.
	commandHandler func(_ string, args []string) *exec.Cmd
}

// CommandContext captures the command and returns a mock [exec.Cmd].
func (m *mockCommandExecutor) CommandContext(_ context.Context, name string, args ...string) *exec.Cmd {
	m.capturedCmd = name
	m.capturedArgs = args

	if m.commandHandler != nil {
		return m.commandHandler(name, args)
	}

	if m.shouldFail {
		if m.mockOutput != "" {
			return exec.Command("sh", "-c", "echo '"+m.mockOutput+"' && false")
		}
		return exec.Command("false")
	}
	if m.mockOutput != "" {
		return exec.Command("echo", m.mockOutput)
	}
	return exec.Command("echo", "success")
}

func TestNewManager(t *testing.T) {
	out := output.NewTerminal(&bytes.Buffer{}, &bytes.Buffer{})
	m := mcp.NewManager(out)

	settingsPath := mcp.ManagerSettingsPath(m)
	if settingsPath == "" {
		t.Error("settingsPath should not be empty")
	}
	if !strings.Contains(settingsPath, "settings.json") {
		t.Errorf("settingsPath should contain settings.json, got %s", settingsPath)
	}
	if mcp.ManagerOutput(m) == nil {
		t.Error("output should be set")
	}
	if !mcp.ManagerHasExecutor(m) {
		t.Error("executor should be set")
	}
}

func TestNewManagerWithExecutor(t *testing.T) {
	out := output.NewTerminal(&bytes.Buffer{}, &bytes.Buffer{})
	mockExec := &mockCommandExecutor{
		capturedCmd:    "",
		capturedArgs:   nil,
		mockOutput:     "",
		shouldFail:     false,
		commandHandler: nil,
	}
	m := mcp.NewManagerWithExecutor(out, mockExec)

	if !mcp.ManagerExecutorIs(m, mockExec) {
		t.Error("custom executor should be set")
	}
}

// assertSettingsLoaded verifies that loadSettings returned valid settings.
func assertSettingsLoaded(t *testing.T, settings *mcp.Settings, err error, wantErr bool) bool {
	t.Helper()
	if (err != nil) != wantErr {
		t.Errorf("loadSettings() error = %v, wantErr %v", err, wantErr)
		return false
	}
	return err == nil && settings != nil
}

func TestLoadSettings(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(_ *testing.T, settingsPath string)
		checkFunc func(t *testing.T, settings *mcp.Settings)
		wantErr   bool
	}{
		{
			name: "loads valid settings",
			setupFunc: func(_ *testing.T, settingsPath string) {
				settings := &mcp.Settings{
					MCPServers: map[string]mcp.Server{
						"targetprocess": {
							Type:    "local",
							Command: "node",
							Args:    []string{"server.js"},
							Env: map[string]any{
								"API_KEY": "test-key",
							},
						},
						"jira": {
							Type:    "local",
							Command: "python",
							Args:    []string{"jira_mcp.py"},
							Env:     nil,
						},
					},
				}
				data, _ := json.MarshalIndent(settings, "", "  ")
				os.MkdirAll(filepath.Dir(settingsPath), 0o755)
				os.WriteFile(settingsPath, data, 0o600)
			},
			checkFunc: func(t *testing.T, settings *mcp.Settings) {
				t.Helper()
				assertMCPServerCount(t, settings, 2)
				assertTargetprocessServer(t, settings)
			},
			wantErr: false,
		},
		{
			name:      "handles missing file",
			setupFunc: nil,
			checkFunc: nil,
			wantErr:   true,
		},
		{
			name: "handles corrupt JSON",
			setupFunc: func(_ *testing.T, settingsPath string) {
				os.MkdirAll(filepath.Dir(settingsPath), 0o755)
				os.WriteFile(settingsPath, []byte("{invalid json}"), 0o600)
			},
			checkFunc: nil,
			wantErr:   true,
		},
		{
			name: "handles empty MCP servers",
			setupFunc: func(_ *testing.T, settingsPath string) {
				settings := &mcp.Settings{
					MCPServers: map[string]mcp.Server{},
				}
				data, _ := json.MarshalIndent(settings, "", "  ")
				os.MkdirAll(filepath.Dir(settingsPath), 0o755)
				os.WriteFile(settingsPath, data, 0o600)
			},
			checkFunc: func(t *testing.T, settings *mcp.Settings) {
				t.Helper()
				if settings.MCPServers == nil {
					t.Error("MCPServers should be initialized")
				}
				if len(settings.MCPServers) != 0 {
					t.Error("MCPServers should be empty")
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			settingsPath := filepath.Join(tmpDir, "settings.json")

			if tt.setupFunc != nil {
				tt.setupFunc(t, settingsPath)
			}

			m := mcp.NewTestManager(settingsPath, nil, nil)
			settings, err := mcp.ManagerLoadSettings(m)
			if assertSettingsLoaded(t, settings, err, tt.wantErr) && tt.checkFunc != nil {
				tt.checkFunc(t, settings)
			}
		})
	}
}

// assertMCPServerCount verifies the number of MCP servers in settings.
func assertMCPServerCount(t *testing.T, settings *mcp.Settings, expected int) {
	t.Helper()
	if len(settings.MCPServers) != expected {
		t.Errorf("expected %d MCP servers, got %d", expected, len(settings.MCPServers))
	}
}

// assertTargetprocessServer verifies the targetprocess server config.
func assertTargetprocessServer(t *testing.T, settings *mcp.Settings) {
	t.Helper()
	tp, ok := settings.MCPServers["targetprocess"]
	if !ok {
		t.Error("targetprocess server should exist")
		return
	}
	if tp.Type != "local" {
		t.Errorf("targetprocess type = %s, want local", tp.Type)
	}
	if tp.Command != "node" {
		t.Errorf("targetprocess command = %s, want node", tp.Command)
	}
	if len(tp.Args) != 1 || tp.Args[0] != "server.js" {
		t.Errorf("targetprocess args = %v, want [server.js]", tp.Args)
	}
}

// assertServerFound verifies that findMCPByName returned the expected key and server.
func assertServerFound(t *testing.T, key string, server *mcp.Server, err error, wantKey string) {
	t.Helper()
	if err != nil {
		t.Errorf("findMCPByName() error = %v, want found", err)
		return
	}
	if key != wantKey {
		t.Errorf("findMCPByName() key = %s, want %s", key, wantKey)
	}
	if server == nil {
		t.Error("findMCPByName() server should not be nil")
	}
}

// assertServerNotFound verifies that findMCPByName returned a not-found error.
func assertServerNotFound(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Error("findMCPByName() should return error for not found")
		return
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got %v", err)
	}
}

func TestFindMCPByName(t *testing.T) {
	settings := &mcp.Settings{
		MCPServers: map[string]mcp.Server{
			"targetprocess": {
				Type:    "local",
				Command: "node",
				Args:    nil,
				Env:     nil,
			},
			"jira-mcp": {
				Type:    "local",
				Command: "python",
				Args:    nil,
				Env:     nil,
			},
			"GitHub": {
				Type:    "local",
				Command: "gh",
				Args:    nil,
				Env:     nil,
			},
		},
	}

	tests := []struct {
		name       string
		searchName string
		wantKey    string
		wantFound  bool
	}{
		{
			name:       "exact match",
			searchName: "targetprocess",
			wantKey:    "targetprocess",
			wantFound:  true,
		},
		{
			name:       "case insensitive exact match",
			searchName: "TargetProcess",
			wantKey:    "targetprocess",
			wantFound:  true,
		},
		{
			name:       "partial match - target",
			searchName: "target",
			wantKey:    "targetprocess",
			wantFound:  true,
		},
		{
			name:       "partial match - target-process",
			searchName: "target-process",
			wantKey:    "targetprocess",
			wantFound:  true,
		},
		{
			name:       "partial match - jira",
			searchName: "jira",
			wantKey:    "jira-mcp",
			wantFound:  true,
		},
		{
			name:       "case insensitive partial match",
			searchName: "github",
			wantKey:    "GitHub",
			wantFound:  true,
		},
		{
			name:       "not found",
			searchName: "nonexistent",
			wantKey:    "",
			wantFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mcp.NewTestManager("", nil, nil)
			key, server, err := mcp.ManagerFindMCPByName(m, settings, tt.searchName)

			if tt.wantFound {
				assertServerFound(t, key, server, err, tt.wantKey)
			} else {
				assertServerNotFound(t, err)
			}
		})
	}
}

func TestEnable(t *testing.T) {
	tests := []struct {
		name         string
		mcpName      string
		settings     *mcp.Settings
		mockOutput   string
		shouldFail   bool
		wantErr      bool
		checkCommand func(t *testing.T, cmd string, args []string)
	}{
		{
			name:    "enables MCP successfully",
			mcpName: "targetprocess",
			settings: &mcp.Settings{
				MCPServers: map[string]mcp.Server{
					"targetprocess": {
						Type:    "local",
						Command: "node",
						Args:    []string{"server.js"},
						Env:     nil,
					},
				},
			},
			mockOutput: "",
			shouldFail: false,
			wantErr:    false,
			checkCommand: func(t *testing.T, _ string, args []string) {
				t.Helper()
				expectedArgs := []string{"mcp", "add", "targetprocess", "node", "server.js"}
				if !slicesEqual(args, expectedArgs) {
					t.Errorf("args = %v, want %v", args, expectedArgs)
				}
			},
		},
		{
			name:    "expands home directory in command",
			mcpName: "test",
			settings: &mcp.Settings{
				MCPServers: map[string]mcp.Server{
					"test": {
						Type:    "",
						Command: "~/bin/mcp",
						Args:    []string{"--port", "3000"},
						Env:     nil,
					},
				},
			},
			mockOutput: "",
			shouldFail: false,
			wantErr:    false,
			checkCommand: func(t *testing.T, _ string, args []string) {
				t.Helper()
				if len(args) < 4 {
					t.Fatalf("Not enough args: %v", args)
				}
				commandPath := args[3]
				if strings.Contains(commandPath, "~") {
					t.Error("Command path should have ~ expanded")
				}
				if !strings.Contains(commandPath, "/bin/mcp") {
					t.Error("Command path should contain /bin/mcp")
				}
			},
		},
		{
			name:    "handles already enabled",
			mcpName: "jira",
			settings: &mcp.Settings{
				MCPServers: map[string]mcp.Server{
					"jira": {
						Type:    "",
						Command: "jira-mcp",
						Args:    nil,
						Env:     nil,
					},
				},
			},
			mockOutput:   "MCP server 'jira' already exists",
			shouldFail:   true,
			wantErr:      false,
			checkCommand: nil,
		},
		{
			name:    "MCP not in settings",
			mcpName: "nonexistent",
			settings: &mcp.Settings{
				MCPServers: map[string]mcp.Server{},
			},
			mockOutput:   "",
			shouldFail:   false,
			wantErr:      true,
			checkCommand: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			settingsPath := filepath.Join(tmpDir, "settings.json")

			data, _ := json.MarshalIndent(tt.settings, "", "  ")
			os.WriteFile(settingsPath, data, 0o600)

			mockExec := &mockCommandExecutor{
				capturedCmd:    "",
				capturedArgs:   nil,
				mockOutput:     tt.mockOutput,
				shouldFail:     tt.shouldFail,
				commandHandler: nil,
			}

			out := output.NewTerminal(&bytes.Buffer{}, &bytes.Buffer{})
			m := mcp.NewTestManager(settingsPath, out, mockExec)

			err := m.Enable(context.Background(), tt.mcpName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Enable() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.checkCommand != nil && !tt.wantErr {
				tt.checkCommand(t, mockExec.capturedCmd, mockExec.capturedArgs)
			}
		})
	}
}

func TestDisable(t *testing.T) {
	tests := []struct {
		name         string
		mcpName      string
		settings     *mcp.Settings
		mockOutput   string
		shouldFail   bool
		wantErr      bool
		checkCommand func(t *testing.T, cmd string, args []string)
	}{
		{
			name:    "disables MCP successfully",
			mcpName: "targetprocess",
			settings: &mcp.Settings{
				MCPServers: map[string]mcp.Server{
					"targetprocess": {
						Type:    "local",
						Command: "",
						Args:    nil,
						Env:     nil,
					},
				},
			},
			mockOutput: "",
			shouldFail: false,
			wantErr:    false,
			checkCommand: func(t *testing.T, _ string, args []string) {
				t.Helper()
				expectedArgs := []string{"mcp", "remove", "targetprocess"}
				if !slicesEqual(args, expectedArgs) {
					t.Errorf("args = %v, want %v", args, expectedArgs)
				}
			},
		},
		{
			name:    "handles not found",
			mcpName: "nonexistent",
			settings: &mcp.Settings{
				MCPServers: map[string]mcp.Server{},
			},
			mockOutput:   "MCP server 'nonexistent' not found",
			shouldFail:   true,
			wantErr:      false,
			checkCommand: nil,
		},
		{
			name:    "uses provided name when not in settings",
			mcpName: "custom-mcp",
			settings: &mcp.Settings{
				MCPServers: map[string]mcp.Server{},
			},
			mockOutput: "",
			shouldFail: false,
			wantErr:    false,
			checkCommand: func(t *testing.T, _ string, args []string) {
				t.Helper()
				expectedArgs := []string{"mcp", "remove", "custom-mcp"}
				if !slicesEqual(args, expectedArgs) {
					t.Errorf("args = %v, want %v", args, expectedArgs)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			settingsPath := filepath.Join(tmpDir, "settings.json")

			data, _ := json.MarshalIndent(tt.settings, "", "  ")
			os.WriteFile(settingsPath, data, 0o600)

			mockExec := &mockCommandExecutor{
				capturedCmd:    "",
				capturedArgs:   nil,
				mockOutput:     tt.mockOutput,
				shouldFail:     tt.shouldFail,
				commandHandler: nil,
			}

			out := output.NewTerminal(&bytes.Buffer{}, &bytes.Buffer{})
			m := mcp.NewTestManager(settingsPath, out, mockExec)

			err := m.Disable(context.Background(), tt.mcpName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Disable() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.checkCommand != nil && !tt.wantErr {
				tt.checkCommand(t, mockExec.capturedCmd, mockExec.capturedArgs)
			}
		})
	}
}

// assertServersEnabled checks that all expected servers were attempted for enable.
func assertServersEnabled(t *testing.T, enabledServers map[string]bool, expected []string) {
	t.Helper()
	for _, name := range expected {
		if !enabledServers[name] {
			t.Errorf("Server %s was not attempted to enable", name)
		}
	}
}

func TestEnableAll(t *testing.T) {
	tests := []struct {
		name           string
		settings       *mcp.Settings
		enabledServers []string
		failServers    []string
		wantErr        bool
	}{
		{
			name: "enables all servers",
			settings: &mcp.Settings{
				MCPServers: map[string]mcp.Server{
					"server1": {Type: "", Command: "cmd1", Args: nil, Env: nil},
					"server2": {Type: "", Command: "cmd2", Args: nil, Env: nil},
					"server3": {Type: "", Command: "cmd3", Args: nil, Env: nil},
				},
			},
			enabledServers: []string{"server1", "server2", "server3"},
			failServers:    nil,
			wantErr:        false,
		},
		{
			name: "handles partial failures",
			settings: &mcp.Settings{
				MCPServers: map[string]mcp.Server{
					"server1": {Type: "", Command: "cmd1", Args: nil, Env: nil},
					"server2": {Type: "", Command: "cmd2", Args: nil, Env: nil},
				},
			},
			enabledServers: []string{"server1"},
			failServers:    []string{"server2"},
			wantErr:        true,
		},
		{
			name: "handles empty servers",
			settings: &mcp.Settings{
				MCPServers: map[string]mcp.Server{},
			},
			enabledServers: []string{},
			failServers:    nil,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			settingsPath := filepath.Join(tmpDir, "settings.json")

			data, _ := json.MarshalIndent(tt.settings, "", "  ")
			os.WriteFile(settingsPath, data, 0o600)

			enabledServers := make(map[string]bool)
			failServers := tt.failServers

			mockExec := &mockCommandExecutor{
				capturedCmd:  "",
				capturedArgs: nil,
				mockOutput:   "",
				shouldFail:   false,
				commandHandler: func(_ string, args []string) *exec.Cmd {
					if len(args) >= 3 && args[0] == "mcp" && args[1] == "add" {
						serverName := args[2]
						enabledServers[serverName] = true
						if slices.Contains(failServers, serverName) {
							return exec.Command("false")
						}
					}
					return exec.Command("echo", "success")
				},
			}

			out := output.NewTerminal(&bytes.Buffer{}, &bytes.Buffer{})
			m := mcp.NewTestManager(settingsPath, out, mockExec)

			err := m.EnableAll(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("EnableAll() error = %v, wantErr %v", err, tt.wantErr)
			}

			assertServersEnabled(t, enabledServers, tt.enabledServers)
		})
	}
}

// assertServersRemoved checks that expected servers were removed and no unexpected ones.
func assertServersRemoved(t *testing.T, removedServers map[string]bool, expected []string) {
	t.Helper()
	for _, name := range expected {
		if !removedServers[name] {
			t.Errorf("Server %s was not removed", name)
		}
	}
	for removed := range removedServers {
		if !slices.Contains(expected, removed) {
			t.Errorf("Unexpected server %s was removed", removed)
		}
	}
}

func TestDisableAll(t *testing.T) {
	tests := []struct {
		name           string
		listOutput     string
		expectedRemove []string
		wantErr        bool
	}{
		{
			name: "disables all listed servers",
			listOutput: "Checking MCP servers...\n" +
				"targetprocess: Running\n" +
				"jira-mcp: Running\n" +
				"github: Stopped",
			expectedRemove: []string{"targetprocess", "jira-mcp", "github"},
			wantErr:        false,
		},
		{
			name:           "handles no servers",
			listOutput:     "Checking MCP servers...\n",
			expectedRemove: []string{},
			wantErr:        false,
		},
		{
			name: "parses various output formats",
			listOutput: "Some header text\n" +
				"server1: Status information here\n" +
				"server2: Different status\n" +
				"Other text that should be ignored\n" +
				"server3: More status",
			expectedRemove: []string{"server1", "server2", "server3"},
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			removedServers := make(map[string]bool)
			listOutput := tt.listOutput

			mockExec := &mockCommandExecutor{
				capturedCmd:  "",
				capturedArgs: nil,
				mockOutput:   "",
				shouldFail:   false,
				commandHandler: func(_ string, args []string) *exec.Cmd {
					if len(args) >= 2 && args[0] == "mcp" {
						if args[1] == "list" {
							return exec.Command("echo", listOutput)
						}
						if args[1] == "remove" && len(args) >= 3 {
							removedServers[args[2]] = true
						}
					}
					return exec.Command("echo", "success")
				},
			}

			out := output.NewTerminal(&bytes.Buffer{}, &bytes.Buffer{})
			m := mcp.NewTestManager("", out, mockExec)

			err := m.DisableAll(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("DisableAll() error = %v, wantErr %v", err, tt.wantErr)
			}

			assertServersRemoved(t, removedServers, tt.expectedRemove)
		})
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		name       string
		shouldFail bool
		wantErr    bool
	}{
		{
			name:       "lists successfully",
			shouldFail: false,
			wantErr:    false,
		},
		{
			name:       "handles command error",
			shouldFail: true,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := &mockCommandExecutor{
				capturedCmd:    "",
				capturedArgs:   nil,
				mockOutput:     "MCP list output",
				shouldFail:     tt.shouldFail,
				commandHandler: nil,
			}

			out := output.NewTerminal(&bytes.Buffer{}, &bytes.Buffer{})
			m := mcp.NewTestManager("", out, mockExec)

			err := m.List(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				assertCapturedCommand(t, mockExec, "claude", []string{"mcp", "list"})
			}
		})
	}
}

// assertCapturedCommand verifies the captured command and args on the mock executor.
func assertCapturedCommand(t *testing.T, mockExec *mockCommandExecutor, wantCmd string, wantArgs []string) {
	t.Helper()
	if mockExec.capturedCmd != wantCmd {
		t.Errorf("command = %s, want %s", mockExec.capturedCmd, wantCmd)
	}
	if !slicesEqual(mockExec.capturedArgs, wantArgs) {
		t.Errorf("args = %v, want %v", mockExec.capturedArgs, wantArgs)
	}
}

func TestRemoveMCP(t *testing.T) {
	tests := []struct {
		name       string
		mcpName    string
		mockOutput string
		shouldFail bool
		wantErr    bool
	}{
		{
			name:       "removes successfully",
			mcpName:    "test-mcp",
			mockOutput: "",
			shouldFail: false,
			wantErr:    false,
		},
		{
			name:       "handles not found",
			mcpName:    "nonexistent",
			mockOutput: "MCP server 'nonexistent' not found",
			shouldFail: true,
			wantErr:    false,
		},
		{
			name:       "handles other errors",
			mcpName:    "test-mcp",
			mockOutput: "Some other error",
			shouldFail: true,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := &mockCommandExecutor{
				capturedCmd:    "",
				capturedArgs:   nil,
				mockOutput:     tt.mockOutput,
				shouldFail:     tt.shouldFail,
				commandHandler: nil,
			}

			out := output.NewTerminal(&bytes.Buffer{}, &bytes.Buffer{})
			m := mcp.NewTestManager("", out, mockExec)

			err := mcp.ManagerRemoveMCP(context.Background(), m, tt.mcpName)
			if (err != nil) != tt.wantErr {
				t.Errorf("removeMCP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHomeDirectoryExpansion(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "expands tilde",
			command:  "~/bin/mcp",
			expected: filepath.Join(homeDir, "bin/mcp"),
		},
		{
			name:     "leaves absolute path unchanged",
			command:  "/usr/local/bin/mcp",
			expected: "/usr/local/bin/mcp",
		},
		{
			name:     "leaves relative path unchanged",
			command:  "./bin/mcp",
			expected: "./bin/mcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := tt.command
			if strings.HasPrefix(command, "~/") {
				command = filepath.Join(homeDir, command[2:])
			}

			if command != tt.expected {
				t.Errorf("expanded command = %s, want %s", command, tt.expected)
			}
		})
	}
}

// slicesEqual compares two string slices for equality.
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
