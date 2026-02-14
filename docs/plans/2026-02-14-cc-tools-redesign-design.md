# cc-tools Architecture Redesign

> **For Claude:** REQUIRED SUB-SKILL: Use executing-plans to implement this plan task-by-task.

**Goal:** Restructure cc-tools from manual arg parsing with business logic in cmd/ to idiomatic Go with Cobra CLI, a handler registry pattern, unified YAML config, and notification system.

**Architecture:** Thin Cobra command files in cmd/ delegate to internal/ packages. A handler package provides a registry that dispatches hook events to registered handlers returning structured responses. Notifications use a Notifier interface with ntfy, audio, and desktop backends.

**Tech Stack:** Go 1.24+, Cobra (CLI), gopkg.in/yaml.v3 (config), lipgloss (styling), testify (testing)

---

## 1. Problem Statement

The current architecture has several issues:

1. **Business logic in cmd/**: `main.go` is 757 lines containing session commands, hook handler adapter closures, registry building, and stdin reading
2. **Manual argument parsing**: Hand-rolled `os.Args` switching instead of a CLI framework
3. **No structured hook output**: Handlers write directly to io.Writer instead of returning structured data matching the Claude Code hooks JSON protocol
4. **Dual config systems**: `internal/config` manages YAML files while some config is passed via flags/environment
5. **Notification stubs**: Handlers exist but are no-ops; implementations exist but are not wired
6. **No CLI niceties**: No shell completion, no consistent --help, no flag validation

## 2. Package Structure

```
cmd/cc-tools/
    main.go          # Cobra root command + main()
    hook.go          # cc-tools hook (stdin dispatch)
    validate.go      # cc-tools validate (standalone)
    config.go        # cc-tools config get/set/list/reset
    session.go       # cc-tools session list/info/alias/unalias
    skip.go          # cc-tools skip/unskip
    debug.go         # cc-tools debug
    mcp.go           # cc-tools mcp install/uninstall/status

internal/
    handler/         # NEW - hook event handlers
        handler.go   # Handler interface, Response, HookOutput, Registry
        session.go   # SessionStart/SessionEnd handlers
        tooluse.go   # PreToolUse/PostToolUse handlers
        notify.go    # Notification event handler
        stop.go      # Stop handler
        compact.go   # PreCompact handler (suggest-compact)
        prompt.go    # UserPromptSubmit handler
    notify/          # NEW - notification backends
        notifier.go  # Notifier interface, Message, MultiNotifier
        ntfy.go      # NtfyNotifier (HTTP POST)
        audio.go     # AudioNotifier (afplay)
        desktop.go   # DesktopNotifier (osascript)
    config/          # EXISTING - unified YAML config
    hookcmd/         # EXISTING - input parsing, event constants
    hooks/           # EXISTING - validation, discovery, execution
    session/         # EXISTING - session store
    mcp/             # EXISTING - MCP server management
    output/          # EXISTING - formatting
    shared/          # EXISTING - shared utilities
    debug/           # EXISTING - debug logging
    compact/         # EXISTING - compact suggestion
    observe/         # EXISTING - file observation
    superpowers/     # EXISTING - superpowers injection
    pkgmanager/      # EXISTING - package manager detection
```

**Dependency direction:** `cmd/` → `internal/handler/` → other `internal/` packages. The handler package is the composition root that imports domain packages.

## 3. Cobra CLI Structure

### Root Command (`main.go`)

```go
func newRootCmd() *cobra.Command {
    root := &cobra.Command{
        Use:   "cc-tools",
        Short: "Claude Code integration tools",
        PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
            if debug {
                debugLog(cmd.Name(), args)
            }
            return nil
        },
    }

    root.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")

    root.AddCommand(
        newHookCmd(),
        newValidateCmd(),
        newConfigCmd(),
        newSessionCmd(),
        newSkipCmd(),
        newDebugCmd(),
        newMCPCmd(),
    )

    return root
}

func main() {
    if err := newRootCmd().Execute(); err != nil {
        os.Exit(1)
    }
}
```

### Hook Command (`hook.go`)

The primary entry point for Claude Code hooks. Reads JSON from stdin, dispatches to the handler registry, and writes structured JSON to stdout.

```go
func newHookCmd() *cobra.Command {
    return &cobra.Command{
        Use:    "hook",
        Short:  "Handle Claude Code hook events",
        Hidden: true, // Called by Claude Code, not users
        RunE: func(cmd *cobra.Command, args []string) error {
            input, err := hookcmd.ParseInput(os.Stdin)
            if err != nil {
                return err
            }

            cfg := config.MustLoad()
            registry := handler.NewRegistry(cfg)
            resp := registry.Dispatch(cmd.Context(), input)

            return writeResponse(os.Stdout, os.Stderr, resp)
        },
    }
}
```

### Validate Command (`validate.go`)

Standalone command for manual validation. Calls the same `hooks.RunValidation()` that the PostToolUse handler uses.

```go
func newValidateCmd() *cobra.Command {
    var timeout int
    cmd := &cobra.Command{
        Use:   "validate",
        Short: "Run lint and test validation",
        RunE: func(cmd *cobra.Command, args []string) error {
            cfg := config.MustLoad()
            return hooks.RunValidation(cmd.Context(), cfg.Validate)
        },
    }
    cmd.Flags().IntVarP(&timeout, "timeout", "t", 30, "timeout in seconds")
    return cmd
}
```

### Other Commands

Each command file follows the same pattern: create `*cobra.Command`, parse flags, delegate to internal packages. No business logic in cmd/.

## 4. Handler Package

### Handler Interface

```go
// Handler processes a hook event and returns a structured response.
type Handler interface {
    Name() string
    Handle(ctx context.Context, input *hookcmd.HookInput) (*Response, error)
}
```

Handlers are structs with state (config references, store instances). The `Name()` method provides context for error logging.

### Response Types

```go
// Response captures a handler's output for the Claude Code hooks protocol.
type Response struct {
    ExitCode int
    Stdout   *HookOutput
    Stderr   string
}

// HookOutput is the JSON written to stdout per the Claude Code hooks protocol.
type HookOutput struct {
    Continue          bool              `json:"continue,omitempty"`
    StopReason        string            `json:"stopReason,omitempty"`
    SuppressOutput    bool              `json:"suppressOutput,omitempty"`
    SystemMessage     string            `json:"systemMessage,omitempty"`
    HookSpecificOutput map[string]any   `json:"hookSpecificOutput,omitempty"`
    AdditionalContext []string          `json:"additionalContext,omitempty"`
    PermissionDecision string           `json:"permissionDecision,omitempty"`
    UpdatedInput      map[string]any    `json:"updatedInput,omitempty"`
}
```

`Continue` and `SuppressOutput` are plain `bool` (not `*bool`). The zero value `false` is the correct default; omitempty skips them when not set. No pointer indirection needed.

### Registry

```go
// Registry maps event names to handler slices.
type Registry struct {
    handlers map[string][]Handler
}

// NewRegistry creates a registry with all default handlers wired.
func NewRegistry(cfg *config.Values) *Registry {
    r := &Registry{handlers: make(map[string][]Handler)}

    // SessionStart handlers
    r.Register(hookcmd.EventSessionStart,
        NewSuperpowersHandler(cfg),
        NewPkgManagerHandler(cfg),
        NewSessionContextHandler(cfg),
    )

    // PreToolUse handlers
    r.Register(hookcmd.EventPreToolUse,
        NewSuggestCompactHandler(cfg),
        NewObserveHandler(cfg),
        NewPreCommitReminderHandler(cfg),
    )

    // PostToolUse handlers
    r.Register(hookcmd.EventPostToolUse,
        NewValidateHandler(cfg),
    )

    // Notification handlers
    r.Register(hookcmd.EventNotification,
        NewNotifyHandler(cfg),
    )

    // SessionEnd handlers
    r.Register(hookcmd.EventSessionEnd,
        NewSessionEndHandler(cfg),
    )

    return r
}
```

### Response Merging

When multiple handlers run for one event, responses merge simply:

```go
// Dispatch runs all handlers for the event and merges responses.
func (r *Registry) Dispatch(ctx context.Context, input *hookcmd.HookInput) *Response {
    handlers := r.handlers[input.HookEventName]
    if len(handlers) == 0 {
        return &Response{ExitCode: 0}
    }

    merged := &Response{ExitCode: 0}
    for _, h := range handlers {
        resp, err := h.Handle(ctx, input)
        if err != nil {
            merged.Stderr += fmt.Sprintf("[%s] error: %v\n", h.Name(), err)
            continue
        }
        if resp == nil {
            continue
        }
        if resp.ExitCode > merged.ExitCode {
            merged.ExitCode = resp.ExitCode
        }
        if resp.Stdout != nil && merged.Stdout == nil {
            merged.Stdout = resp.Stdout
        }
        if resp.Stderr != "" {
            merged.Stderr += resp.Stderr
        }
    }
    return merged
}
```

In practice, at most one handler per event produces stdout JSON. The merge takes the first non-nil Stdout, concatenates stderr, and uses max exit code. This is ~15 lines, not a complex merge engine.

## 5. Event Name Constants

Defined in `internal/hookcmd/` as typed constants:

```go
// Event name constants matching Claude Code hook event names.
const (
    EventSessionStart       = "SessionStart"
    EventSessionEnd         = "SessionEnd"
    EventPreToolUse         = "PreToolUse"
    EventPostToolUse        = "PostToolUse"
    EventPostToolUseFailure = "PostToolUseFailure"
    EventPreCompact         = "PreCompact"
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

Prevents typos in string literals and makes event references self-documenting.

## 6. Unified YAML Configuration

Single config file at `~/.config/cc-tools/config.yaml` using `gopkg.in/yaml.v3`.

### Config Structure

```go
// Values is the top-level configuration.
type Values struct {
    Validate   ValidateConfig   `yaml:"validate"`
    Notify     NotifyConfig     `yaml:"notify"`
    Compact    CompactConfig    `yaml:"compact"`
    Observe    ObserveConfig    `yaml:"observe"`
    Learning   LearningConfig   `yaml:"learning"`
    PreCommit  PreCommitConfig  `yaml:"precommit"`
}

type ValidateConfig struct {
    Enabled  bool `yaml:"enabled"`
    Timeout  int  `yaml:"timeout"`   // seconds, default 30
    Cooldown int  `yaml:"cooldown"`  // seconds, default 30
    SkipLint bool `yaml:"skip_lint"`
    SkipTest bool `yaml:"skip_test"`
}

type NotifyConfig struct {
    Ntfy       NtfyConfig       `yaml:"ntfy"`
    Audio      AudioConfig      `yaml:"audio"`
    Desktop    DesktopConfig    `yaml:"desktop"`
    QuietHours QuietHoursConfig `yaml:"quiet_hours"`
}

type NtfyConfig struct {
    Enabled  bool   `yaml:"enabled"`            // default false
    Topic    string `yaml:"topic"`              // required when enabled
    Server   string `yaml:"server"`             // default "https://ntfy.sh"
    Token    string `yaml:"token,omitempty"`     // optional bearer token
    Priority int    `yaml:"priority"`           // 1-5, default 3
}

type AudioConfig struct {
    Enabled   bool   `yaml:"enabled"`           // default true
    SoundFile string `yaml:"sound_file"`        // path to sound
}

type DesktopConfig struct {
    Enabled bool `yaml:"enabled"`               // default true
}

type QuietHoursConfig struct {
    Enabled bool   `yaml:"enabled"`
    Start   string `yaml:"start"`               // "22:00"
    End     string `yaml:"end"`                 // "08:00"
}
```

### Example config.yaml

```yaml
validate:
  enabled: true
  timeout: 30
  cooldown: 30

notify:
  ntfy:
    enabled: true
    topic: "my-cc-tools"
    server: "https://ntfy.sh"
    priority: 3
  audio:
    enabled: true
    sound_file: "/System/Library/Sounds/Ping.aiff"
  desktop:
    enabled: true
  quiet_hours:
    enabled: true
    start: "22:00"
    end: "08:00"

compact:
  threshold: 60
  warn_percentage: 80

observe:
  enabled: true
```

### Config Get/Set

The `config.Manager` uses explicit switch statements for dot-notation key access:

```go
func (m *Manager) Get(key string) (any, error) {
    switch key {
    case "validate.timeout":
        return m.values.Validate.Timeout, nil
    case "validate.cooldown":
        return m.values.Validate.Cooldown, nil
    case "notify.ntfy.enabled":
        return m.values.Notify.Ntfy.Enabled, nil
    case "notify.ntfy.topic":
        return m.values.Notify.Ntfy.Topic, nil
    // ... one case per key
    default:
        return nil, fmt.Errorf("unknown config key: %s", key)
    }
}
```

This follows Go's preference for explicit over clever. 20 keys = 20 case statements with full type safety. No reflection, no map-walking.

## 7. Notification System

### Notifier Interface

```go
// Notifier sends notifications through a specific backend.
type Notifier interface {
    Notify(ctx context.Context, msg Message) error
}

// Message contains notification content.
type Message struct {
    Title    string
    Body     string
    Priority int      // 1-5 (ntfy scale)
    Tags     []string
}
```

### NtfyNotifier

Sends HTTP POST to ntfy.sh with JSON body:

```go
type NtfyNotifier struct {
    topic    string
    server   string
    token    string
    priority int
    client   *http.Client
}

func (n *NtfyNotifier) Notify(ctx context.Context, msg Message) error {
    body := map[string]any{
        "topic":    n.topic,
        "title":    msg.Title,
        "message":  msg.Body,
        "priority": msg.Priority,
        "tags":     msg.Tags,
    }
    data, err := json.Marshal(body)
    if err != nil {
        return fmt.Errorf("marshal ntfy payload: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, "POST", n.server, bytes.NewReader(data))
    if err != nil {
        return fmt.Errorf("create ntfy request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")
    if n.token != "" {
        req.Header.Set("Authorization", "Bearer "+n.token)
    }

    resp, err := n.client.Do(req)
    if err != nil {
        return fmt.Errorf("send ntfy notification: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("ntfy returned status %d", resp.StatusCode)
    }
    return nil
}
```

### AudioNotifier

```go
type AudioNotifier struct {
    soundFile string
}

func (a *AudioNotifier) Notify(ctx context.Context, _ Message) error {
    cmd := exec.CommandContext(ctx, "afplay", a.soundFile)
    return cmd.Run()
}
```

### DesktopNotifier

```go
type DesktopNotifier struct{}

func (d *DesktopNotifier) Notify(ctx context.Context, msg Message) error {
    script := fmt.Sprintf(
        `display notification %q with title %q`,
        msg.Body, msg.Title,
    )
    cmd := exec.CommandContext(ctx, "osascript", "-e", script)
    return cmd.Run()
}
```

### MultiNotifier

Composites all enabled backends and respects quiet hours:

```go
type MultiNotifier struct {
    notifiers  []Notifier
    quietHours *QuietHoursConfig
}

func (m *MultiNotifier) Notify(ctx context.Context, msg Message) error {
    if m.isQuietTime() {
        return nil
    }
    var errs []error
    for _, n := range m.notifiers {
        if err := n.Notify(ctx, msg); err != nil {
            errs = append(errs, err)
        }
    }
    return errors.Join(errs...)
}
```

### Constructor

```go
func NewMultiNotifier(cfg *NotifyConfig) *MultiNotifier {
    var notifiers []Notifier

    if cfg.Ntfy.Enabled && cfg.Ntfy.Topic != "" {
        notifiers = append(notifiers, NewNtfyNotifier(cfg.Ntfy))
    }
    if cfg.Audio.Enabled {
        notifiers = append(notifiers, NewAudioNotifier(cfg.Audio))
    }
    if cfg.Desktop.Enabled {
        notifiers = append(notifiers, NewDesktopNotifier())
    }

    return &MultiNotifier{
        notifiers:  notifiers,
        quietHours: &cfg.QuietHours,
    }
}
```

## 8. Hook Wiring

The `.claude/settings.json` hooks configuration points all events at `cc-tools hook`:

```json
{
  "hooks": {
    "SessionStart": [
      { "matcher": "", "command": "cc-tools hook" }
    ],
    "PreToolUse": [
      { "matcher": "Edit|Write|MultiEdit|NotebookEdit", "command": "cc-tools hook" }
    ],
    "PostToolUse": [
      { "matcher": "Edit|Write|MultiEdit|NotebookEdit", "command": "cc-tools hook" }
    ],
    "Notification": [
      { "matcher": "", "command": "cc-tools hook" }
    ],
    "SessionEnd": [
      { "matcher": "", "command": "cc-tools hook" }
    ],
    "Stop": [
      { "matcher": "", "command": "cc-tools hook" }
    ]
  }
}
```

Each hook invocation runs `cc-tools hook`, which reads stdin JSON, dispatches through the registry, and writes structured output to stdout/stderr with the appropriate exit code.

## 9. Validate Dual-Path

The validate command works as both a standalone CLI command and a PostToolUse hook handler, calling the same implementation:

```
Standalone:  cc-tools validate  →  hooks.RunValidation()
PostToolUse: cc-tools hook      →  registry → ValidateHandler → hooks.RunValidation()
```

The `ValidateHandler` in `internal/handler/tooluse.go` calls the same `hooks.RunValidation()` that the standalone `cc-tools validate` command uses. No code duplication.

## 10. Debug Logging

Debug logging uses Cobra's `PersistentPreRunE` on the root command to log command name and args. For stdin-dependent commands (like `hook`), stdin content is logged in the individual command's `RunE` function after parsing.

```go
PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
    if debug {
        debug.Log("cmd=%s args=%v", cmd.Name(), args)
    }
    return nil
},
```

## 11. Dependencies

**Add:**
- `github.com/spf13/cobra` - CLI framework (replaces ~200 lines of manual arg parsing, provides shell completion, --help, flag validation)
- `gopkg.in/yaml.v3` - YAML config parsing

**Keep:**
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/stretchr/testify` - Test assertions

**Remove (eventually):**
- Manual `os.Args` parsing in `main.go`
- `handlerFunc` adapter and `buildHookRegistry()` in `main.go`

## 12. Migration Strategy

The redesign replaces the existing cmd/ layer and adds new internal/ packages. Existing internal/ packages (hooks, config, session, mcp, etc.) remain largely unchanged. The migration:

1. Add Cobra and yaml.v3 dependencies
2. Create `internal/handler/` with Handler interface, Registry, Response types
3. Create `internal/notify/` with Notifier backends
4. Migrate each command from main.go to its own Cobra command file
5. Wire handlers into registry
6. Update `.claude/settings.json` hook configuration
7. Remove old manual arg parsing

Tests for existing internal/ packages remain valid. New tests cover handler dispatch, response merging, notification backends, and Cobra command wiring.

## 13. Design Principles Applied

- **Thin cmd/, fat internal/**: All business logic in internal packages
- **Accept interfaces, return structs**: Handlers return `*Response`, notifiers accept `Notifier` interface
- **Composition root**: handler package wires dependencies; individual handlers import only what they need
- **Explicit over clever**: Switch statements for config keys, typed constants for events
- **YAGNI**: No *bool, no reflection, no complex merge engine
- **Testability**: Handlers are pure functions (input → response), no stdio mocking needed
- **Single responsibility**: One file per command, one file per handler group, one file per notification backend
