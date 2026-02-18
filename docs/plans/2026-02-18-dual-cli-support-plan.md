# Dual CLI Support Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use executing-plans to implement this plan task-by-task.

**Goal:** Rename the binary from cc-tools to hookd and add an adapter layer so both Claude Code and Gemini CLI can use the same hook handler.

**Architecture:** Thin adapter pattern — each CLI gets an adapter that normalizes its input/output into canonical internal types. The handler registry stays unchanged except for event constant renames.

**Tech Stack:** Go 1.26, Cobra, Task, gotestsum, golangci-lint

---

### Task 1: Rename binary from cc-tools to hookd

Mechanical rename of the entry point directory, Taskfile vars, and CLI metadata. No logic changes.

**Files:**
- Rename: `cmd/cc-tools/` → `cmd/hookd/` (all files move)
- Modify: `Taskfile.yml:5-8` (BINARY_NAME, MAIN_PATH vars)
- Modify: `cmd/hookd/main.go:31-32` (Use and Short fields)
- Modify: `cmd/hookd/main.go:66` (debug log message)
- Modify: `cmd/hookd/main.go:87` (fallback debug path)
- Modify: `.claude/settings.json:48,59,69,79,89,101,124` (all `cc-tools hook` → `hookd hook`, `cc-tools validate` → `hookd validate`)

**Step 1: Rename the directory**

```bash
git mv cmd/cc-tools cmd/hookd
```

**Step 2: Update Taskfile.yml vars**

Change lines 5-8 from:

```yaml
vars:
  BINARY_NAME: cc-tools
  BIN_DIR: bin
  BINARY_PATH: "{{.BIN_DIR}}/{{.BINARY_NAME}}"
  MAIN_PATH: "./cmd/cc-tools"
```

To:

```yaml
vars:
  BINARY_NAME: hookd
  BIN_DIR: bin
  BINARY_PATH: "{{.BIN_DIR}}/{{.BINARY_NAME}}"
  MAIN_PATH: "./cmd/hookd"
```

**Step 3: Update main.go CLI metadata**

In `cmd/hookd/main.go`, change:

```go
Use:     "cc-tools",
Short:   "Claude Code Tools",
```

To:

```go
Use:     "hookd",
Short:   "Hook daemon for AI coding CLIs",
```

**Step 4: Update debug log references in main.go**

Change the `cc-tools invoked` string to `hookd invoked` and the fallback path from `/tmp/cc-tools.debug` to `/tmp/hookd.debug`.

**Step 5: Update .claude/settings.json**

Replace all `cc-tools hook` with `hookd hook` and `cc-tools validate` with `hookd validate`.

**Step 6: Run tests and verify**

```bash
task check
```

Expected: All 1001 tests pass. Lint clean.

**Step 7: Commit**

```bash
git add cmd/hookd/ Taskfile.yml .claude/settings.json
git commit -m "refactor: rename binary from cc-tools to hookd"
```

---

### Task 2: Rename config and debug paths from cc-tools to hookd

Update all filesystem paths that reference `cc-tools` in directory names. Add auto-migration from old paths.

**Files:**
- Modify: `internal/shared/configdir.go:8-21`
- Modify: `internal/shared/configdir_test.go`
- Modify: `internal/shared/debug_paths.go:48`
- Modify: `internal/shared/debug_paths_test.go`
- Modify: `internal/config/manager.go:500-514`
- Modify: `internal/config/manager_test.go`
- Modify: `internal/debug/config.go` (package doc)
- Modify: `internal/debug/config_test.go`

**Step 1: Write failing test for config directory migration**

In `internal/shared/configdir_test.go`, add a test that verifies:
- `ConfigDir()` returns paths with `hookd` instead of `cc-tools`
- When `XDG_CONFIG_HOME` is set, returns `$XDG_CONFIG_HOME/hookd`
- Default returns `~/.config/hookd`

Run: `gotestsum --format pkgname -- -tags=testmode -run TestConfigDir ./internal/shared/...`
Expected: FAIL — paths still say `cc-tools`.

**Step 2: Update configdir.go**

Change `internal/shared/configdir.go`:

```go
// ConfigDir returns the hookd configuration directory.
// Respects $XDG_CONFIG_HOME; defaults to ~/.config/hookd.
func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "hookd")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("/tmp", ".config", "hookd")
	}

	return filepath.Join(home, ".config", "hookd")
}
```

**Step 3: Run test to verify it passes**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestConfigDir ./internal/shared/...`
Expected: PASS

**Step 4: Update debug_paths.go**

Change the format string in `GetDebugLogPathForDir` from `cc-tools` to `hookd`:

```go
return fmt.Sprintf("/tmp/hookd-%s-%s.debug", namePart, hashStr)
```

**Step 5: Update config/manager.go getConfigFilePath()**

Change `internal/config/manager.go:500-514` to use `hookd` instead of `cc-tools`:

```go
func getConfigFilePath() string {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "hookd", "config.json")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "config.json"
	}

	return filepath.Join(homeDir, ".config", "hookd", "config.json")
}
```

**Step 6: Update all test expectations**

Update test files that assert path strings containing `cc-tools`:
- `internal/shared/configdir_test.go`: `"cc-tools"` → `"hookd"`
- `internal/shared/debug_paths_test.go`: `"/tmp/cc-tools-"` → `"/tmp/hookd-"`
- `internal/config/manager_test.go`: `"cc-tools"` → `"hookd"`
- `internal/debug/config_test.go`: `"cc-tools"` → `"hookd"`

**Step 7: Update package doc comments**

- `internal/shared/colors.go:1`: `cc-tools` → `hookd`
- `internal/debug/config.go:1`: `cc-tools` → `hookd`

**Step 8: Run full test suite**

```bash
task check
```

Expected: All tests pass.

**Step 9: Commit**

```bash
git add internal/shared/ internal/config/ internal/debug/
git commit -m "refactor: rename config and debug paths from cc-tools to hookd"
```

---

### Task 3: Normalize event constants to CLI-neutral names

Rename event constants from Claude Code-specific names to CLI-neutral canonical names. This is a mechanical rename with no logic changes.

**Files:**
- Modify: `internal/hookcmd/events.go:3-20`
- Modify: `internal/handler/defaults.go:13-45`
- Modify: `internal/handler/registry.go:28` (uses `HookEventName` field)
- Modify: `internal/hooks/executor.go:121` (checks `HookEventName`)
- Modify: All test files that reference `hookcmd.Event*` constants

**Step 1: Update event constants in events.go**

Change `internal/hookcmd/events.go`:

```go
// Canonical event name constants for hook dispatch.
// CLI-neutral: adapters map CLI-specific names to these.
const (
	EventSessionStart       = "SessionStart"
	EventSessionEnd         = "SessionEnd"
	EventBeforeTool         = "BeforeTool"
	EventAfterTool          = "AfterTool"
	EventAfterToolFailure   = "AfterToolFailure"
	EventBeforeCompress     = "BeforeCompress"
	EventNotification       = "Notification"
	EventUserPromptSubmit   = "UserPromptSubmit"
	EventPermissionRequest  = "PermissionRequest"
	EventStop               = "Stop"
	EventSubagentStart      = "SubagentStart"
	EventSubagentStop       = "SubagentStop"
	EventTeammateIdle       = "TeammateIdle"
	EventTaskCompleted      = "TaskCompleted"
)
```

Note: `SessionStart`, `SessionEnd`, `Notification`, `Stop` stay the same. The renames are:
- `EventPreToolUse` → `EventBeforeTool`
- `EventPostToolUse` → `EventAfterTool`
- `EventPostToolUseFailure` → `EventAfterToolFailure`
- `EventPreCompact` → `EventBeforeCompress`

**Step 2: Update AllEvents()**

Update the `AllEvents()` function to use the new constant names.

**Step 3: Update handler/defaults.go registrations**

Replace all old constant references with new ones:
- `hookcmd.EventPreToolUse` → `hookcmd.EventBeforeTool`
- `hookcmd.EventPostToolUse` → `hookcmd.EventAfterTool`
- `hookcmd.EventPostToolUseFailure` → `hookcmd.EventAfterToolFailure`
- `hookcmd.EventPreCompact` → `hookcmd.EventBeforeCompress`

**Step 4: Update hooks/executor.go**

Line 121: Change `"PostToolUse"` string literal to `hookcmd.EventAfterTool`. Also note that the `HookEventName` field still holds the raw value from JSON — the adapter (built in Task 5) will set this to the canonical name. For now, update the string comparison.

**Step 5: Update all test files**

Search and replace across all test files:
- `hookcmd.EventPreToolUse` → `hookcmd.EventBeforeTool`
- `hookcmd.EventPostToolUse` → `hookcmd.EventAfterTool`
- `hookcmd.EventPostToolUseFailure` → `hookcmd.EventAfterToolFailure`
- `hookcmd.EventPreCompact` → `hookcmd.EventBeforeCompress`

Also update string literals in tests:
- `"PreToolUse"` → `"BeforeTool"` (where used as event name values)
- `"PostToolUse"` → `"AfterTool"`
- `"PreCompact"` → `"BeforeCompress"`
- `"PostToolUseFailure"` → `"AfterToolFailure"`

Key test files:
- `internal/handler/registry_test.go`
- `internal/handler/tooluse_test.go`
- `internal/handler/compact_test.go`
- `internal/handler/session_start_test.go`
- `internal/handler/session_end_test.go`
- `internal/handler/notification_test.go`
- `internal/hookcmd/hookcmd_test.go`
- `internal/hookcmd/input_test.go`
- `internal/hooks/executor_test.go`
- `internal/hooks/hooks_test.go`
- `internal/hooks/validate_test.go`
- `internal/hooks/input_test.go`
- `cmd/hookd/hook_test.go`

**Step 6: Run full test suite**

```bash
task check
```

Expected: All tests pass. This step is critical — every reference must be updated or tests fail.

**Step 7: Commit**

```bash
git add internal/hookcmd/ internal/handler/ internal/hooks/ cmd/hookd/
git commit -m "refactor: rename event constants to CLI-neutral canonical names"
```

---

### Task 4: Add CLIType to HookInput and create adapter package

Create the adapter interface and Claude adapter. The Claude adapter maps Claude Code's event/tool names to canonical names.

**Files:**
- Modify: `internal/hookcmd/input.go:11-17` (add CLIType field)
- Create: `internal/adapter/adapter.go`
- Create: `internal/adapter/adapter_test.go`
- Create: `internal/adapter/claude.go`
- Create: `internal/adapter/claude_test.go`

**Step 1: Write failing test for Detect()**

Create `internal/adapter/adapter_test.go`:

```go
package adapter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/riddopic/cc-tools/internal/adapter"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    adapter.CLIType
	}{
		{
			name:    "defaults to claude when no env vars",
			envVars: map[string]string{},
			want:    adapter.CLIClaude,
		},
		{
			name:    "detects gemini from GEMINI_SESSION_ID",
			envVars: map[string]string{"GEMINI_SESSION_ID": "sess-123"},
			want:    adapter.CLIGemini,
		},
		{
			name:    "detects gemini from GEMINI_PROJECT_DIR",
			envVars: map[string]string{"GEMINI_PROJECT_DIR": "/project"},
			want:    adapter.CLIGemini,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env
			t.Setenv("GEMINI_SESSION_ID", "")
			t.Setenv("GEMINI_PROJECT_DIR", "")
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}
			got := adapter.Detect()
			assert.Equal(t, tt.want, got)
		})
	}
}
```

Run: `gotestsum --format pkgname -- -tags=testmode -run TestDetect ./internal/adapter/...`
Expected: FAIL — package doesn't exist.

**Step 2: Create adapter.go with interface and Detect**

Create `internal/adapter/adapter.go`:

```go
// Package adapter normalizes CLI-specific hook input/output into canonical
// internal types. Each supported CLI (Claude Code, Gemini CLI) gets an
// adapter that maps its event names, tool names, and JSON fields.
package adapter

import (
	"os"

	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
)

// CLIType identifies which AI coding CLI is invoking the hook.
type CLIType string

const (
	// CLIClaude indicates Claude Code.
	CLIClaude CLIType = "claude"
	// CLIGemini indicates Gemini CLI.
	CLIGemini CLIType = "gemini"
)

// Adapter normalizes CLI-specific hook I/O into canonical internal types.
type Adapter interface {
	// ParseInput normalizes raw JSON bytes into a canonical HookInput.
	ParseInput(data []byte) (*hookcmd.HookInput, error)

	// FormatOutput converts an internal Response to CLI-specific JSON bytes.
	FormatOutput(resp *handler.Response) ([]byte, error)

	// Type returns which CLI this adapter handles.
	Type() CLIType
}

// Detect returns which CLI is calling the hook based on environment variables.
// Gemini CLI sets GEMINI_SESSION_ID and GEMINI_PROJECT_DIR; their presence
// indicates Gemini. Otherwise, Claude Code is assumed.
func Detect() CLIType {
	if os.Getenv("GEMINI_SESSION_ID") != "" || os.Getenv("GEMINI_PROJECT_DIR") != "" {
		return CLIGemini
	}
	return CLIClaude
}

// ForCLI returns the adapter for the given CLI type.
func ForCLI(cli CLIType) Adapter {
	switch cli {
	case CLIGemini:
		return &GeminiAdapter{}
	default:
		return &ClaudeAdapter{}
	}
}
```

**Step 3: Run Detect test**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestDetect ./internal/adapter/...`
Expected: PASS (after creating stub GeminiAdapter to satisfy compilation)

**Step 4: Add CLIType field to HookInput**

In `internal/hookcmd/input.go`, add to the struct:

```go
type HookInput struct {
	// CLIType identifies which CLI sent this event. Set by the adapter.
	CLIType string `json:"-"` // not from JSON, set programmatically

	// Common fields (present on ALL events).
	SessionID      string `json:"session_id"`
	// ... rest unchanged
}
```

**Step 5: Write failing test for ClaudeAdapter.ParseInput**

Create `internal/adapter/claude_test.go` with table-driven tests covering:
- Normal PreToolUse event → maps to BeforeTool
- PostToolUse event → maps to AfterTool
- PreCompact event → maps to BeforeCompress
- SessionStart event → passes through
- Tool name normalization (Write → write_file, Edit → edit_file, etc.)
- Empty input returns empty HookInput
- Invalid JSON returns error

**Step 6: Implement ClaudeAdapter**

Create `internal/adapter/claude.go`:

```go
package adapter

import (
	"encoding/json"
	"fmt"

	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
)

// claudeEventMap maps Claude Code event names to canonical names.
var claudeEventMap = map[string]string{
	"PreToolUse":         hookcmd.EventBeforeTool,
	"PostToolUse":        hookcmd.EventAfterTool,
	"PostToolUseFailure": hookcmd.EventAfterToolFailure,
	"PreCompact":         hookcmd.EventBeforeCompress,
	"SessionStart":       hookcmd.EventSessionStart,
	"SessionEnd":         hookcmd.EventSessionEnd,
	"Notification":       hookcmd.EventNotification,
	"UserPromptSubmit":   hookcmd.EventUserPromptSubmit,
	"PermissionRequest":  hookcmd.EventPermissionRequest,
	"Stop":               hookcmd.EventStop,
	"SubagentStart":      hookcmd.EventSubagentStart,
	"SubagentStop":       hookcmd.EventSubagentStop,
	"TeammateIdle":       hookcmd.EventTeammateIdle,
	"TaskCompleted":      hookcmd.EventTaskCompleted,
}

// claudeToolMap maps Claude Code tool names to canonical lowercase names.
var claudeToolMap = map[string]string{
	"Bash":         "bash",
	"Write":        "write_file",
	"Edit":         "edit_file",
	"MultiEdit":    "multi_edit",
	"Read":         "read_file",
	"Glob":         "glob",
	"Grep":         "grep",
	"NotebookEdit": "notebook_edit",
}

// ClaudeAdapter normalizes Claude Code hook JSON into canonical types.
type ClaudeAdapter struct{}

var _ Adapter = (*ClaudeAdapter)(nil)

func (a *ClaudeAdapter) ParseInput(data []byte) (*hookcmd.HookInput, error) {
	if len(data) == 0 {
		return &hookcmd.HookInput{CLIType: string(CLIClaude)}, nil
	}

	var input hookcmd.HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("parsing claude hook input: %w", err)
	}

	input.CLIType = string(CLIClaude)

	// Normalize event name
	if canonical, ok := claudeEventMap[input.HookEventName]; ok {
		input.HookEventName = canonical
	}

	// Normalize tool name
	if canonical, ok := claudeToolMap[input.ToolName]; ok {
		input.ToolName = canonical
	}

	return &input, nil
}

func (a *ClaudeAdapter) FormatOutput(resp *handler.Response) ([]byte, error) {
	if resp.Stdout == nil {
		return nil, nil
	}
	data, err := json.Marshal(resp.Stdout)
	if err != nil {
		return nil, fmt.Errorf("marshal claude hook output: %w", err)
	}
	return data, nil
}

func (a *ClaudeAdapter) Type() CLIType {
	return CLIClaude
}
```

**Step 7: Run adapter tests**

```bash
gotestsum --format pkgname -- -tags=testmode ./internal/adapter/...
```

Expected: PASS

**Step 8: Update IsEditTool to use canonical names**

In `internal/hookcmd/input.go`, update `IsEditTool()`:

```go
func (h *HookInput) IsEditTool() bool {
	switch h.ToolName {
	case "edit_file", "multi_edit", "write_file", "notebook_edit":
		return true
	// Also accept raw Claude Code names for backward compatibility
	// in code paths that haven't gone through the adapter yet.
	case "Edit", "MultiEdit", "Write", "NotebookEdit":
		return true
	default:
		return false
	}
}
```

**Step 9: Run full test suite**

```bash
task check
```

Expected: All tests pass.

**Step 10: Commit**

```bash
git add internal/adapter/ internal/hookcmd/input.go
git commit -m "feat: add adapter package with CLI detection and Claude adapter"
```

---

### Task 5: Create Gemini adapter

Implement the Gemini CLI adapter that reads environment variables and maps Gemini-specific event/tool names.

**Files:**
- Create: `internal/adapter/gemini.go`
- Create: `internal/adapter/gemini_test.go`

**Step 1: Write failing tests for GeminiAdapter.ParseInput**

Create `internal/adapter/gemini_test.go` with table-driven tests:
- BeforeTool event passes through (already canonical)
- AfterTool event passes through
- PreCompress → BeforeCompress
- SessionStart → SessionStart
- Tool name mapping (shell → bash, write_file → write_file, replace → edit_file)
- SessionID read from GEMINI_SESSION_ID env var
- Cwd read from GEMINI_CWD env var
- Empty input returns HookInput with env vars populated

Run: `gotestsum --format pkgname -- -tags=testmode -run TestGeminiAdapter ./internal/adapter/...`
Expected: FAIL — GeminiAdapter not implemented.

**Step 2: Implement GeminiAdapter**

Create `internal/adapter/gemini.go`:

```go
package adapter

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
)

// geminiEventMap maps Gemini CLI event names to canonical names.
var geminiEventMap = map[string]string{
	"BeforeTool":  hookcmd.EventBeforeTool,
	"AfterTool":   hookcmd.EventAfterTool,
	"PreCompress": hookcmd.EventBeforeCompress,
	// These are already canonical:
	"SessionStart":  hookcmd.EventSessionStart,
	"SessionEnd":    hookcmd.EventSessionEnd,
	"Notification":  hookcmd.EventNotification,
}

// geminiToolMap maps Gemini CLI tool names to canonical lowercase names.
var geminiToolMap = map[string]string{
	"shell":      "bash",
	"write_file": "write_file",
	"replace":    "edit_file",
	"read_file":  "read_file",
	"glob":       "glob",
	"grep":       "grep",
}

// GeminiAdapter normalizes Gemini CLI hook JSON into canonical types.
type GeminiAdapter struct{}

var _ Adapter = (*GeminiAdapter)(nil)

func (a *GeminiAdapter) ParseInput(data []byte) (*hookcmd.HookInput, error) {
	input := &hookcmd.HookInput{
		CLIType:   string(CLIGemini),
		SessionID: os.Getenv("GEMINI_SESSION_ID"),
		Cwd:       os.Getenv("GEMINI_CWD"),
	}

	if len(data) == 0 {
		return input, nil
	}

	if err := json.Unmarshal(data, input); err != nil {
		return nil, fmt.Errorf("parsing gemini hook input: %w", err)
	}

	input.CLIType = string(CLIGemini)

	// Populate from env if not in JSON
	if input.SessionID == "" {
		input.SessionID = os.Getenv("GEMINI_SESSION_ID")
	}
	if input.Cwd == "" {
		input.Cwd = os.Getenv("GEMINI_CWD")
	}

	// Normalize event name
	if canonical, ok := geminiEventMap[input.HookEventName]; ok {
		input.HookEventName = canonical
	}

	// Normalize tool name
	if canonical, ok := geminiToolMap[input.ToolName]; ok {
		input.ToolName = canonical
	}

	return input, nil
}

func (a *GeminiAdapter) FormatOutput(resp *handler.Response) ([]byte, error) {
	if resp.Stdout == nil {
		return nil, nil
	}
	data, err := json.Marshal(resp.Stdout)
	if err != nil {
		return nil, fmt.Errorf("marshal gemini hook output: %w", err)
	}
	return data, nil
}

func (a *GeminiAdapter) Type() CLIType {
	return CLIGemini
}
```

**Step 3: Run tests**

```bash
gotestsum --format pkgname -- -tags=testmode ./internal/adapter/...
```

Expected: PASS

**Step 4: Commit**

```bash
git add internal/adapter/gemini.go internal/adapter/gemini_test.go
git commit -m "feat: add Gemini CLI adapter with event and tool normalization"
```

---

### Task 6: Wire adapter into hook command

Replace direct `hookcmd.ParseInput` with adapter-based parsing in the hook command.

**Files:**
- Modify: `cmd/hookd/hook.go:28-47`
- Modify: `cmd/hookd/hook.go:58-77` (writeHookResponse)
- Modify: `cmd/hookd/hook_test.go`

**Step 1: Update runHook to use adapter**

Change `cmd/hookd/hook.go`:

```go
import (
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/riddopic/cc-tools/internal/adapter"
	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/handler"
)

func newHookCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "hook",
		Short:  "Handle hook events from AI coding CLIs",
		Long:   "Reads hook event JSON from stdin, dispatches to registered handlers, and writes structured output.",
		Hidden: true,
		RunE:   runHook,
	}
}

func runHook(cmd *cobra.Command, _ []string) error {
	data, readErr := io.ReadAll(os.Stdin)
	if readErr != nil {
		return nil //nolint:nilerr // hooks must not block on stdin errors
	}
	if len(data) == 0 {
		return nil
	}

	a := adapter.ForCLI(adapter.Detect())
	input, parseErr := a.ParseInput(data)
	if parseErr != nil {
		return nil //nolint:nilerr // hooks must not block on parse errors
	}

	cfg := loadConfig()
	registry := handler.NewDefaultRegistry(cfg)
	resp := registry.Dispatch(cmd.Context(), input)

	return writeHookResponse(os.Stdout, os.Stderr, resp, a)
}
```

Update `writeHookResponse` to accept an adapter:

```go
func writeHookResponse(stdout, stderr io.Writer, resp *handler.Response, a adapter.Adapter) error {
	if resp.Stderr != "" {
		_, _ = stderr.Write([]byte(resp.Stderr))
	}

	if resp.Stdout != nil {
		data, err := a.FormatOutput(resp)
		if err != nil {
			return fmt.Errorf("format hook output: %w", err)
		}
		if data != nil {
			_, _ = stdout.Write(data)
			_, _ = io.WriteString(stdout, "\n")
		}
	}

	if resp.ExitCode != 0 {
		return &exitError{code: resp.ExitCode}
	}

	return nil
}
```

**Step 2: Update hook_test.go**

Update test cases to work with the adapter-based flow. Tests that construct `HookInput` directly still work because the handler registry accepts any `HookInput` regardless of source.

**Step 3: Run full test suite**

```bash
task check
```

Expected: All tests pass.

**Step 4: Commit**

```bash
git add cmd/hookd/hook.go cmd/hookd/hook_test.go
git commit -m "feat: wire adapter into hook command for CLI-agnostic parsing"
```

---

### Task 7: Create setup command

Add `hookd setup` command that generates hook configuration for Claude Code or Gemini CLI.

**Files:**
- Create: `cmd/hookd/setup.go`
- Create: `cmd/hookd/setup_test.go`
- Modify: `cmd/hookd/main.go:41` (add newSetupCmd to root)

**Step 1: Write failing test for setup command**

Create `cmd/hookd/setup_test.go` with tests:
- `hookd setup --cli claude` generates `.claude/settings.json` with `hookd hook` commands
- `hookd setup --cli gemini` generates `.gemini/settings.json` with `hookd hook` commands and regex matchers
- `hookd setup` without flag auto-detects or prompts

**Step 2: Implement setup command**

Create `cmd/hookd/setup.go` with:
- `--cli` flag accepting `claude` or `gemini`
- Claude mode: generates `.claude/settings.json` hook entries
- Gemini mode: generates `.gemini/settings.json` hook entries with regex matchers
- Gemini mode: converts `.claude/commands/*.md` → `.gemini/commands/*.toml`

The command conversion logic:
1. Read each `.md` file
2. Extract YAML frontmatter (if any) for description
3. Extract markdown body as the prompt
4. Replace `$ARGUMENTS` with `{{args}}`
5. Write `.toml` file with `prompt` and optional `description` fields

**Step 3: Register in main.go**

Add `newSetupCmd()` to the `root.AddCommand()` call.

**Step 4: Run tests**

```bash
gotestsum --format pkgname -- -tags=testmode ./cmd/hookd/...
```

Expected: PASS

**Step 5: Commit**

```bash
git add cmd/hookd/setup.go cmd/hookd/setup_test.go cmd/hookd/main.go
git commit -m "feat: add setup command for Claude and Gemini CLI configuration"
```

---

### Task 8: Update package documentation and CLAUDE.md

Update all documentation references from cc-tools to hookd.

**Files:**
- Modify: `CLAUDE.md` (all cc-tools references)
- Modify: `internal/hookcmd/input.go:1` (package doc)
- Modify: `internal/handler/handler.go:1-4` (package doc)
- Modify: `internal/shared/colors.go:1` (package doc)

**Step 1: Update CLAUDE.md**

Replace all `cc-tools` references with `hookd`:
- Project description
- Binary name
- Build commands
- Config paths
- Hook command references

Update the project description to mention both Claude Code and Gemini CLI support.

**Step 2: Update package doc comments**

- `internal/hookcmd/input.go:1`: "dispatches Claude Code hook events" → "dispatches hook events from AI coding CLIs"
- `internal/handler/handler.go:1-4`: "Claude Code hooks protocol" → "hook protocol"
- `internal/handler/handler.go:24`: Response doc comment
- `internal/handler/handler.go:31`: HookOutput doc comment

**Step 3: Run lint to verify**

```bash
task lint
```

Expected: Clean

**Step 4: Commit**

```bash
git add CLAUDE.md internal/hookcmd/ internal/handler/ internal/shared/
git commit -m "docs: update all references from cc-tools to hookd"
```

---

### Task 9: Config directory auto-migration

Add one-time migration from `~/.config/cc-tools/` to `~/.config/hookd/` on first access.

**Files:**
- Modify: `internal/config/manager.go` (add migration in loadConfig)
- Create: `internal/config/migrate.go`
- Create: `internal/config/migrate_test.go`

**Step 1: Write failing test for migration**

Create `internal/config/migrate_test.go` with tests:
- When `~/.config/cc-tools/` exists and `~/.config/hookd/` does not, migrate
- When both exist, prefer `hookd` (no migration)
- When neither exists, create `hookd` fresh
- Migration copies config.json content

**Step 2: Implement migrate.go**

```go
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// migrateFromLegacyPath copies config from the old cc-tools directory
// to the new hookd directory if the old path exists and new does not.
func migrateFromLegacyPath(newPath string) error {
	oldPath := legacyConfigPath(newPath)
	if oldPath == "" {
		return nil
	}

	// Check if old config exists
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return nil
	}

	// Check if new config already exists
	if _, err := os.Stat(newPath); err == nil {
		return nil // new path exists, no migration needed
	}

	// Create new directory
	newDir := filepath.Dir(newPath)
	if err := os.MkdirAll(newDir, 0o750); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// Copy old config to new location
	data, err := os.ReadFile(oldPath)
	if err != nil {
		return fmt.Errorf("read legacy config: %w", err)
	}

	if err := os.WriteFile(newPath, data, 0o600); err != nil {
		return fmt.Errorf("write migrated config: %w", err)
	}

	return nil
}

// legacyConfigPath returns the old cc-tools config path corresponding
// to the given hookd config path, or empty string if not applicable.
func legacyConfigPath(hookdPath string) string {
	dir := filepath.Dir(hookdPath)
	base := filepath.Base(hookdPath)
	parent := filepath.Dir(dir)

	if filepath.Base(dir) != "hookd" {
		return ""
	}

	return filepath.Join(parent, "cc-tools", base)
}
```

**Step 3: Wire migration into loadConfig**

In `internal/config/manager.go:loadConfig()`, add migration call before reading:

```go
func (m *Manager) loadConfig() error {
	// Attempt migration from legacy cc-tools path
	_ = migrateFromLegacyPath(m.configPath)

	// Initialize with defaults
	m.config = GetDefaultConfig()
	// ... rest unchanged
}
```

**Step 4: Run tests**

```bash
gotestsum --format pkgname -- -tags=testmode ./internal/config/...
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/migrate.go internal/config/migrate_test.go internal/config/manager.go
git commit -m "feat: add auto-migration from cc-tools to hookd config paths"
```
