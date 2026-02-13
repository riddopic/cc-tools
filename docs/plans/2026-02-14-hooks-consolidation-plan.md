# Hooks Consolidation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use executing-plans to implement this plan task-by-task.

**Goal:** Consolidate 12 hooks (JS/Python/Bash) into the cc-tools Go binary as `cc-tools hook <event-type>` subcommands.

**Architecture:** Flat subcommand dispatcher under `cc-tools hook`, with each event type running multiple handlers sequentially. New internal packages per domain (session, notify, observe, compact, superpowers, pkgmanager). Existing `internal/hooks` stays focused on validation.

**Tech Stack:** Go 1.26, gopxl/beep (audio), testify/mockery (testing). No Cobra — extends existing switch dispatcher.

**Design doc:** `docs/plans/2026-02-14-hooks-consolidation-design.md`

---

## Task 1: Extend config with hook settings

Add new hook configuration fields to the config system. This is the foundation — all subsequent handlers read these values.

**Files:**
- Modify: `internal/config/manager.go:21-35` (Values struct)
- Modify: `internal/config/manager.go:345-355` (getDefaultConfig)
- Modify: `internal/config/manager.go:358-367` (ensureDefaults)
- Modify: `internal/config/config.go:10-29` (Config struct, HooksConfig)
- Test: `internal/config/manager_test.go`

**Step 1: Write the failing test**

Add a test in `internal/config/manager_test.go` that loads a config file with the new hook fields and verifies they're parsed correctly.

```go
func TestManager_LoadsHookConfig(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantComp int
		wantQH   bool
		wantQHS  string
		wantQHE  string
	}{
		{
			name:     "defaults when empty",
			json:     `{}`,
			wantComp: 50,
			wantQH:   true,
			wantQHS:  "21:00",
			wantQHE:  "07:30",
		},
		{
			name:     "custom values",
			json:     `{"compact":{"threshold":100,"reminder_interval":50},"notify":{"quiet_hours":{"enabled":false,"start":"22:00","end":"08:00"}}}`,
			wantComp: 100,
			wantQH:   false,
			wantQHS:  "22:00",
			wantQHE:  "08:00",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfgPath := filepath.Join(tmpDir, "config.json")
			require.NoError(t, os.WriteFile(cfgPath, []byte(tt.json), 0o600))

			m := NewManagerWithPath(cfgPath)
			cfg, err := m.GetConfig(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.wantComp, cfg.Compact.Threshold)
			assert.Equal(t, tt.wantQH, cfg.Notify.QuietHours.Enabled)
			assert.Equal(t, tt.wantQHS, cfg.Notify.QuietHours.Start)
			assert.Equal(t, tt.wantQHE, cfg.Notify.QuietHours.End)
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestManager_LoadsHookConfig ./internal/config/...`
Expected: FAIL — `NewManagerWithPath` undefined, `cfg.Compact` undefined

**Step 3: Implement config struct extensions**

Add to `internal/config/manager.go` Values struct:

```go
type Values struct {
	Validate      ValidateValues      `json:"validate"`
	Notifications NotificationsValues `json:"notifications"`
	Compact       CompactValues       `json:"compact"`
	Notify        NotifyValues        `json:"notify"`
	Observe       ObserveValues       `json:"observe"`
	Learning      LearningValues      `json:"learning"`
	PreCommit     PreCommitValues     `json:"pre_commit_reminder"`
}

type CompactValues struct {
	Threshold        int `json:"threshold"`
	ReminderInterval int `json:"reminder_interval"`
}

type NotifyValues struct {
	QuietHours QuietHoursValues `json:"quiet_hours"`
	Audio      AudioValues      `json:"audio"`
	Desktop    DesktopValues    `json:"desktop"`
}

type QuietHoursValues struct {
	Enabled bool   `json:"enabled"`
	Start   string `json:"start"`
	End     string `json:"end"`
}

type AudioValues struct {
	Enabled   bool   `json:"enabled"`
	Directory string `json:"directory"`
}

type DesktopValues struct {
	Enabled bool `json:"enabled"`
}

type ObserveValues struct {
	Enabled       bool `json:"enabled"`
	MaxFileSizeMB int  `json:"max_file_size_mb"`
}

type LearningValues struct {
	MinSessionLength int    `json:"min_session_length"`
	LearnedSkillsPath string `json:"learned_skills_path"`
}

type PreCommitValues struct {
	Enabled bool   `json:"enabled"`
	Command string `json:"command"`
}
```

Add `NewManagerWithPath(path string) *Manager` constructor. Update `getDefaultConfig()` and `ensureDefaults()` with new field defaults:
- `Compact`: threshold=50, reminderInterval=25
- `Notify.QuietHours`: enabled=true, start="21:00", end="07:30"
- `Notify.Audio`: enabled=true, directory="~/.claude/audio"
- `Notify.Desktop`: enabled=true
- `Observe`: enabled=true, maxFileSizeMB=10
- `Learning`: minSessionLength=10, learnedSkillsPath=".claude/skills/learned"
- `PreCommit`: enabled=true, command="task pre-commit"

Also update config key constants, `GetValue`, `Set`, `Reset`, `GetAllKeys`, and `GetAll` to handle the new keys.

**Step 4: Run test to verify it passes**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestManager_LoadsHookConfig ./internal/config/...`
Expected: PASS

**Step 5: Run full check**

Run: `task check`
Expected: All pass

**Step 6: Commit**

```bash
git add internal/config/
git commit -m "feat: extend config with hook settings for consolidation"
```

---

## Task 2: Create hookcmd package — input parsing and handler interface

The core dispatcher infrastructure. All subsequent handler tasks depend on this.

**Files:**
- Create: `internal/hookcmd/input.go`
- Create: `internal/hookcmd/handler.go`
- Create: `internal/hookcmd/hookcmd.go`
- Test: `internal/hookcmd/input_test.go`
- Test: `internal/hookcmd/handler_test.go`
- Test: `internal/hookcmd/hookcmd_test.go`

**Step 1: Write the failing test for input parsing**

```go
// internal/hookcmd/input_test.go
func TestParseInput(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantEvt string
		wantSID string
		wantTN  string
	}{
		{
			name:    "PreToolUse event",
			json:    `{"hook_event_name":"PreToolUse","session_id":"abc123","tool_name":"Bash","tool_input":{"command":"ls"},"cwd":"/tmp","transcript_path":"/tmp/t.jsonl"}`,
			wantEvt: "PreToolUse",
			wantSID: "abc123",
			wantTN:  "Bash",
		},
		{
			name:    "SessionStart event",
			json:    `{"hook_event_name":"SessionStart","session_id":"def456","source":"startup","cwd":"/tmp"}`,
			wantEvt: "SessionStart",
			wantSID: "def456",
		},
		{
			name:    "Stop event with stop_hook_active",
			json:    `{"hook_event_name":"Stop","session_id":"ghi789","stop_hook_active":true}`,
			wantEvt: "Stop",
			wantSID: "ghi789",
		},
		{
			name:    "lenient parsing ignores unknown fields",
			json:    `{"hook_event_name":"PreToolUse","session_id":"x","unknown_field":"ignored"}`,
			wantEvt: "PreToolUse",
			wantSID: "x",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := ParseInput(strings.NewReader(tt.json))
			require.NoError(t, err)
			assert.Equal(t, tt.wantEvt, input.HookEventName)
			assert.Equal(t, tt.wantSID, input.SessionID)
			if tt.wantTN != "" {
				assert.Equal(t, tt.wantTN, input.ToolName)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestParseInput ./internal/hookcmd/...`
Expected: FAIL — package does not exist

**Step 3: Implement input.go**

```go
// Package hookcmd dispatches Claude Code hook events to registered handlers.
package hookcmd

import (
	"encoding/json"
	"fmt"
	"io"
)

// HookInput represents the JSON input from Claude Code hooks.
type HookInput struct {
	// Common fields (present on ALL events)
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
	PermissionMode string `json:"permission_mode"`
	HookEventName  string `json:"hook_event_name"`

	// Tool events (PreToolUse, PostToolUse, PostToolUseFailure)
	ToolName   string          `json:"tool_name,omitempty"`
	ToolInput  json.RawMessage `json:"tool_input,omitempty"`
	ToolOutput json.RawMessage `json:"tool_response,omitempty"`
	ToolUseID  string          `json:"tool_use_id,omitempty"`

	// PostToolUseFailure specific
	Error       string `json:"error,omitempty"`
	IsInterrupt bool   `json:"is_interrupt,omitempty"`

	// SessionStart specific
	Source string `json:"source,omitempty"`
	Model  string `json:"model,omitempty"`

	// SessionEnd specific
	Reason string `json:"reason,omitempty"`

	// Stop / SubagentStop specific
	StopHookActive bool `json:"stop_hook_active,omitempty"`

	// UserPromptSubmit specific
	Prompt string `json:"prompt,omitempty"`

	// Notification specific
	Message          string `json:"message,omitempty"`
	Title            string `json:"title,omitempty"`
	NotificationType string `json:"notification_type,omitempty"`

	// PreCompact specific
	Trigger            string `json:"trigger,omitempty"`
	CustomInstructions string `json:"custom_instructions,omitempty"`
}

// ParseInput reads JSON from the given reader and parses it into HookInput.
func ParseInput(r io.Reader) (*HookInput, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}
	if len(data) == 0 {
		return &HookInput{}, nil
	}
	var input HookInput
	if unmarshalErr := json.Unmarshal(data, &input); unmarshalErr != nil {
		return nil, fmt.Errorf("parsing hook input JSON: %w", unmarshalErr)
	}
	return &input, nil
}

// GetToolInputString extracts a string field from tool_input JSON.
func (h *HookInput) GetToolInputString(key string) string {
	if len(h.ToolInput) == 0 {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(h.ToolInput, &m); err != nil {
		return ""
	}
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
```

**Step 4: Run input test to verify it passes**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestParseInput ./internal/hookcmd/...`
Expected: PASS

**Step 5: Write the failing test for handler runner**

```go
// internal/hookcmd/handler_test.go
func TestRunHandlers_SequentialExecution(t *testing.T) {
	var order []string
	h1 := &testHandler{name: "first", runFn: func() error { order = append(order, "first"); return nil }}
	h2 := &testHandler{name: "second", runFn: func() error { order = append(order, "second"); return nil }}

	var stdout, stderr bytes.Buffer
	RunHandlers(context.Background(), &HookInput{}, []Handler{h1, h2}, &stdout, &stderr)
	assert.Equal(t, []string{"first", "second"}, order)
}

func TestRunHandlers_ErrorContinues(t *testing.T) {
	var order []string
	h1 := &testHandler{name: "fails", runFn: func() error { order = append(order, "fails"); return errors.New("boom") }}
	h2 := &testHandler{name: "runs", runFn: func() error { order = append(order, "runs"); return nil }}

	var stdout, stderr bytes.Buffer
	RunHandlers(context.Background(), &HookInput{}, []Handler{h1, h2}, &stdout, &stderr)
	assert.Equal(t, []string{"fails", "runs"}, order)
	assert.Contains(t, stderr.String(), "boom")
}

func TestRunHandlers_PanicRecovery(t *testing.T) {
	h1 := &testHandler{name: "panics", runFn: func() error { panic("oops") }}
	h2 := &testHandler{name: "safe", runFn: func() error { return nil }}

	var stdout, stderr bytes.Buffer
	assert.NotPanics(t, func() {
		RunHandlers(context.Background(), &HookInput{}, []Handler{h1, h2}, &stdout, &stderr)
	})
	assert.Contains(t, stderr.String(), "panic")
}
```

**Step 6: Implement handler.go**

```go
package hookcmd

import (
	"context"
	"fmt"
	"io"
)

// Handler processes a hook event.
type Handler interface {
	Name() string
	Run(ctx context.Context, input *HookInput, out io.Writer, errOut io.Writer) error
}

// RunHandlers executes handlers sequentially. Errors are logged to errOut
// but do not stop subsequent handlers. Panics are recovered.
func RunHandlers(ctx context.Context, input *HookInput, handlers []Handler, out, errOut io.Writer) {
	for _, h := range handlers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Fprintf(errOut, "[%s] panic recovered: %v\n", h.Name(), r)
				}
			}()
			if err := h.Run(ctx, input, out, errOut); err != nil {
				fmt.Fprintf(errOut, "[%s] error: %v\n", h.Name(), err)
			}
		}()
	}
}
```

**Step 7: Run handler test to verify it passes**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestRunHandlers ./internal/hookcmd/...`
Expected: PASS

**Step 8: Write the failing test for the dispatcher**

```go
// internal/hookcmd/hookcmd_test.go
func TestDispatch_UnknownEventExitsCleanly(t *testing.T) {
	var stdout, stderr bytes.Buffer
	input := &HookInput{HookEventName: "SubagentStart"}
	exitCode := Dispatch(context.Background(), input, &stdout, &stderr, nil)
	assert.Equal(t, 0, exitCode)
}

func TestDispatch_RoutesToCorrectHandlers(t *testing.T) {
	called := false
	h := &testHandler{name: "test", runFn: func() error { called = true; return nil }}
	registry := map[string][]Handler{
		"PreToolUse": {h},
	}

	var stdout, stderr bytes.Buffer
	input := &HookInput{HookEventName: "PreToolUse"}
	exitCode := Dispatch(context.Background(), input, &stdout, &stderr, registry)
	assert.Equal(t, 0, exitCode)
	assert.True(t, called)
}
```

**Step 9: Implement hookcmd.go**

```go
package hookcmd

import (
	"context"
	"io"
)

// Dispatch routes a hook event to registered handlers.
// Returns the exit code (always 0 — errors are logged, not fatal).
func Dispatch(ctx context.Context, input *HookInput, out, errOut io.Writer, registry map[string][]Handler) int {
	handlers, ok := registry[input.HookEventName]
	if !ok {
		// Unknown event type — accept gracefully
		return 0
	}
	RunHandlers(ctx, input, handlers, out, errOut)
	return 0
}
```

**Step 10: Run all hookcmd tests**

Run: `gotestsum --format pkgname -- -tags=testmode ./internal/hookcmd/...`
Expected: PASS

**Step 11: Run full check**

Run: `task check`
Expected: All pass

**Step 12: Commit**

```bash
git add internal/hookcmd/
git commit -m "feat: add hookcmd package with input parsing, handler interface, and dispatcher"
```

---

## Task 3: Wire `cc-tools hook` command into main dispatcher

Connect the new hookcmd package to the CLI entry point.

**Files:**
- Modify: `cmd/cc-tools/main.go:40-63` (switch statement)
- Modify: `cmd/cc-tools/main.go:65-86` (printUsage)

**Step 1: Add `hook` and `session` cases to the switch**

Add these cases to the switch in `main()`:

```go
case "hook":
	runHookCommand()
case "session":
	runSessionCommand()
```

**Step 2: Implement `runHookCommand()`**

```go
func runHookCommand() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: cc-tools hook <event-type>")
		os.Exit(1)
	}

	// Map CLI subcommand names to hook event names
	eventMap := map[string]string{
		"session-start":        "SessionStart",
		"session-end":          "SessionEnd",
		"pre-tool-use":         "PreToolUse",
		"post-tool-use":        "PostToolUse",
		"post-tool-use-failure": "PostToolUseFailure",
		"pre-compact":          "PreCompact",
		"stop":                 "Stop",
		"notification":         "Notification",
	}

	subCmd := os.Args[2]
	eventName, ok := eventMap[subCmd]
	if !ok {
		// Accept unknown events gracefully (future-proofing)
		eventName = subCmd
	}

	input, err := hookcmd.ParseInput(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing hook input: %v\n", err)
		os.Exit(0) // still exit 0 — hooks must not block
	}
	input.HookEventName = eventName

	// Build handler registry (will be populated in later tasks)
	registry := buildHookRegistry()
	exitCode := hookcmd.Dispatch(context.Background(), input, os.Stdout, os.Stderr, registry)
	os.Exit(exitCode)
}

func buildHookRegistry() map[string][]hookcmd.Handler {
	return map[string][]hookcmd.Handler{
		// Handlers will be added in subsequent tasks
	}
}

func runSessionCommand() {
	// Stub — will be implemented in Task 11
	fmt.Fprintln(os.Stderr, "session command not yet implemented")
	os.Exit(1)
}
```

**Step 3: Update printUsage**

Add `hook` and `session` commands to the usage help text.

**Step 4: Update debugLog to also read stdin for hook command**

Change the `needsStdin` check in `debugLog()`:

```go
needsStdin := len(os.Args) > 1 && (os.Args[1] == "validate" || os.Args[1] == "hook")
```

**Step 5: Run full check**

Run: `task check`
Expected: All pass

**Step 6: Commit**

```bash
git add cmd/cc-tools/main.go
git commit -m "feat: wire hook and session commands into CLI dispatcher"
```

---

## Task 4: Quiet hours + notification handlers

Implement `internal/notify/` — quiet hours, desktop, and audio notifications.

**Files:**
- Create: `internal/notify/quiethours.go`
- Create: `internal/notify/desktop.go`
- Create: `internal/notify/audio.go`
- Test: `internal/notify/quiethours_test.go`
- Test: `internal/notify/desktop_test.go`
- Test: `internal/notify/audio_test.go`

**Step 1: Write quiet hours test**

```go
// internal/notify/quiethours_test.go
func TestIsQuietHours(t *testing.T) {
	tests := []struct {
		name  string
		now   time.Time
		start string
		end   string
		want  bool
	}{
		{name: "10pm is quiet", now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.Local), start: "21:00", end: "07:30", want: true},
		{name: "7am is quiet", now: time.Date(2026, 1, 1, 7, 0, 0, 0, time.Local), start: "21:00", end: "07:30", want: true},
		{name: "8am is not quiet", now: time.Date(2026, 1, 1, 8, 0, 0, 0, time.Local), start: "21:00", end: "07:30", want: false},
		{name: "3pm is not quiet", now: time.Date(2026, 1, 1, 15, 0, 0, 0, time.Local), start: "21:00", end: "07:30", want: false},
		{name: "exactly on start", now: time.Date(2026, 1, 1, 21, 0, 0, 0, time.Local), start: "21:00", end: "07:30", want: true},
		{name: "exactly on end", now: time.Date(2026, 1, 1, 7, 30, 0, 0, time.Local), start: "21:00", end: "07:30", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qh := QuietHours{Enabled: true, Start: tt.start, End: tt.end}
			assert.Equal(t, tt.want, qh.IsActive(tt.now))
		})
	}
}

func TestIsQuietHours_Disabled(t *testing.T) {
	qh := QuietHours{Enabled: false, Start: "21:00", End: "07:30"}
	midnight := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	assert.False(t, qh.IsActive(midnight))
}
```

**Step 2: Run test to verify it fails**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestIsQuietHours ./internal/notify/...`
Expected: FAIL — package does not exist

**Step 3: Implement quiethours.go**

```go
// Package notify provides desktop and audio notifications for hook events.
package notify

import (
	"fmt"
	"time"
)

// QuietHours configuration for suppressing notifications.
type QuietHours struct {
	Enabled bool
	Start   string // "HH:MM" format
	End     string // "HH:MM" format
}

// IsActive returns true if the given time falls within quiet hours.
// Returns false if quiet hours are disabled.
func (qh QuietHours) IsActive(now time.Time) bool {
	if !qh.Enabled {
		return false
	}
	startH, startM, err := parseTime(qh.Start)
	if err != nil {
		return false
	}
	endH, endM, err := parseTime(qh.End)
	if err != nil {
		return false
	}

	nowMinutes := now.Hour()*60 + now.Minute()
	startMinutes := startH*60 + startM
	endMinutes := endH*60 + endM

	if startMinutes <= endMinutes {
		// Same day range (e.g., 08:00 to 17:00)
		return nowMinutes >= startMinutes && nowMinutes < endMinutes
	}
	// Overnight range (e.g., 21:00 to 07:30)
	return nowMinutes >= startMinutes || nowMinutes < endMinutes
}

func parseTime(s string) (int, int, error) {
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil {
		return 0, 0, fmt.Errorf("parsing time %q: %w", s, err)
	}
	return h, m, nil
}
```

**Step 4: Run quiet hours test**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestIsQuietHours ./internal/notify/...`
Expected: PASS

**Step 5: Write desktop notification test**

Test using a mock exec runner interface to avoid calling osascript in tests. The test verifies the correct command is constructed.

```go
// internal/notify/desktop_test.go
func TestDesktopNotify_BuildsOsascriptCommand(t *testing.T) {
	var capturedCmd string
	runner := &mockRunner{runFn: func(name string, args ...string) error {
		capturedCmd = name + " " + strings.Join(args, " ")
		return nil
	}}
	d := NewDesktop(runner)
	err := d.Send("Test Title", "Test Message")
	require.NoError(t, err)
	assert.Contains(t, capturedCmd, "osascript")
	assert.Contains(t, capturedCmd, "Test Title")
	assert.Contains(t, capturedCmd, "Test Message")
}
```

**Step 6: Implement desktop.go**

Define a `CmdRunner` interface for testability, implement `Desktop` struct that calls `osascript -e 'display notification...'`.

**Step 7: Write audio playback test**

Test that audio handler picks a random MP3, calls the player interface, and respects quiet hours. Use an `AudioPlayer` interface for mocking.

**Step 8: Implement audio.go**

Define `AudioPlayer` interface. The real implementation uses `gopxl/beep` for MP3 playback. Picks a random file from the configured audio directory. Add `gopxl/beep` to go.mod.

Run: `go get github.com/gopxl/beep/v2 github.com/gopxl/beep/v2/mp3 github.com/gopxl/beep/v2/speaker`

**Step 9: Run all notify tests**

Run: `gotestsum --format pkgname -- -tags=testmode ./internal/notify/...`
Expected: PASS

**Step 10: Run full check**

Run: `task check`
Expected: All pass

**Step 11: Commit**

```bash
git add internal/notify/ go.mod go.sum
git commit -m "feat: add notify package with quiet hours, desktop, and audio playback"
```

---

## Task 5: Compact suggestion handler

Implement `internal/compact/` — tool call counting and `/compact` suggestion.

**Files:**
- Create: `internal/compact/suggest.go`
- Create: `internal/compact/log.go`
- Test: `internal/compact/suggest_test.go`
- Test: `internal/compact/log_test.go`

**Step 1: Write suggest-compact test**

```go
func TestSuggestCompact_FirstThreshold(t *testing.T) {
	tmpDir := t.TempDir()
	var stderr bytes.Buffer
	s := NewSuggestor(tmpDir, 50, 25)

	// Simulate 49 calls — no suggestion
	for i := 0; i < 49; i++ {
		s.RecordCall("session1", &stderr)
		stderr.Reset()
	}
	// 50th call — suggestion
	s.RecordCall("session1", &stderr)
	assert.Contains(t, stderr.String(), "/compact")
}

func TestSuggestCompact_ReminderInterval(t *testing.T) {
	tmpDir := t.TempDir()
	var stderr bytes.Buffer
	s := NewSuggestor(tmpDir, 5, 3)

	// Calls 1-4: no suggestion
	for i := 0; i < 4; i++ {
		s.RecordCall("s1", &stderr)
		stderr.Reset()
	}
	// Call 5: threshold hit
	s.RecordCall("s1", &stderr)
	assert.Contains(t, stderr.String(), "/compact")
	stderr.Reset()

	// Calls 6-7: no suggestion
	for i := 0; i < 2; i++ {
		s.RecordCall("s1", &stderr)
		stderr.Reset()
	}
	// Call 8: reminder interval
	s.RecordCall("s1", &stderr)
	assert.Contains(t, stderr.String(), "/compact")
}
```

**Step 2: Run test to verify it fails**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestSuggestCompact ./internal/compact/...`
Expected: FAIL

**Step 3: Implement suggest.go**

Counter stored in `<tmpDir>/cc-tools-compact-<sessionID>.count`. Read count, increment, check threshold, write back.

**Step 4: Write log-compaction test**

Test that `LogCompaction` writes a timestamped entry to the compaction log file.

**Step 5: Implement log.go**

Append `[timestamp] compaction triggered` to `.claude/sessions/compaction-log.txt`.

**Step 6: Run all compact tests**

Run: `gotestsum --format pkgname -- -tags=testmode ./internal/compact/...`
Expected: PASS

**Step 7: Run full check and commit**

```bash
task check
git add internal/compact/
git commit -m "feat: add compact package with tool call counting and compaction logging"
```

---

## Task 6: Observe handler (CL-v2)

Implement `internal/observe/` — tool event recording with JSONL and file rotation.

**Files:**
- Create: `internal/observe/observe.go`
- Create: `internal/observe/archive.go`
- Test: `internal/observe/observe_test.go`
- Test: `internal/observe/archive_test.go`

**Step 1: Write observe test**

```go
func TestObserve_WritesJSONL(t *testing.T) {
	tmpDir := t.TempDir()
	obs := NewObserver(tmpDir, 10)

	event := Event{
		Timestamp: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		Phase:     "pre",
		ToolName:  "Bash",
		ToolInput: json.RawMessage(`{"command":"ls"}`),
		SessionID: "abc",
	}
	require.NoError(t, obs.Record(event))

	data, err := os.ReadFile(filepath.Join(tmpDir, "observations.jsonl"))
	require.NoError(t, err)

	var parsed Event
	require.NoError(t, json.Unmarshal(data[:len(data)-1], &parsed)) // strip newline
	assert.Equal(t, "Bash", parsed.ToolName)
	assert.Equal(t, "pre", parsed.Phase)
}
```

**Step 2: Implement observe.go**

`Event` struct, `Observer` struct with `Record(event)` method that appends JSONL. Checks for `.claude/homunculus/disabled` file.

**Step 3: Write archive rotation test**

Test that when file exceeds max size, it's rotated to `observations-{timestamp}.jsonl`.

**Step 4: Implement archive.go**

Check file size before writing, rotate if over threshold.

**Step 5: Run all observe tests, check, commit**

```bash
gotestsum --format pkgname -- -tags=testmode ./internal/observe/...
task check
git add internal/observe/
git commit -m "feat: add observe package with JSONL recording and file rotation"
```

---

## Task 7: Superpowers injection handler

Implement `internal/superpowers/` — reads skill file, outputs hookSpecificOutput JSON.

**Files:**
- Create: `internal/superpowers/inject.go`
- Test: `internal/superpowers/inject_test.go`

**Step 1: Write inject test**

```go
func TestInject_OutputsHookSpecificJSON(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, ".claude", "skills", "using-superpowers")
	require.NoError(t, os.MkdirAll(skillDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Superpowers\nBe awesome"), 0o644))

	var stdout bytes.Buffer
	inj := NewInjector(tmpDir)
	require.NoError(t, inj.Run(context.Background(), &stdout))

	var result map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &result))
	hso, ok := result["hookSpecificOutput"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "SessionStart", hso["hookEventName"])
	assert.Contains(t, hso["additionalContext"], "Superpowers")
}
```

**Step 2: Implement inject.go**

Read the SKILL.md file from `$CLAUDE_PROJECT_DIR/.claude/skills/using-superpowers/SKILL.md`. Output JSON:

```json
{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"<file contents>"}}
```

**Step 3: Run tests, check, commit**

```bash
gotestsum --format pkgname -- -tags=testmode ./internal/superpowers/...
task check
git add internal/superpowers/
git commit -m "feat: add superpowers injection handler for SessionStart"
```

---

## Task 8: Package manager detection handler

Implement `internal/pkgmanager/` — detects and persists preferred package manager.

**Files:**
- Create: `internal/pkgmanager/detect.go`
- Test: `internal/pkgmanager/detect_test.go`

**Step 1: Write detection test**

```go
func TestDetect_LockFileOrder(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  string
	}{
		{name: "bun.lock", files: []string{"bun.lock"}, want: "bun"},
		{name: "pnpm-lock.yaml", files: []string{"pnpm-lock.yaml"}, want: "pnpm"},
		{name: "yarn.lock", files: []string{"yarn.lock"}, want: "yarn"},
		{name: "package-lock.json", files: []string{"package-lock.json"}, want: "npm"},
		{name: "no lock file", files: []string{}, want: "npm"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			for _, f := range tt.files {
				require.NoError(t, os.WriteFile(filepath.Join(tmpDir, f), []byte{}, 0o644))
			}
			assert.Equal(t, tt.want, Detect(tmpDir))
		})
	}
}
```

**Step 2: Implement detect.go**

Detection priority: `PREFERRED_PACKAGE_MANAGER` env var → lock file in project dir → default npm. Write to `CLAUDE_ENV_FILE` if the env var is set.

**Step 3: Run tests, check, commit**

```bash
gotestsum --format pkgname -- -tags=testmode ./internal/pkgmanager/...
task check
git add internal/pkgmanager/
git commit -m "feat: add pkgmanager detection handler"
```

---

## Task 9: Session store (redesigned JSON format)

Implement `internal/session/` — session CRUD with structured JSON.

**Files:**
- Create: `internal/session/store.go`
- Create: `internal/session/alias.go`
- Create: `internal/session/transcript.go`
- Test: `internal/session/store_test.go`
- Test: `internal/session/alias_test.go`
- Test: `internal/session/transcript_test.go`

**Step 1: Write store test**

```go
func TestStore_CreateAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	s := &Session{
		Version: "2.0",
		ID:      "abc123",
		Date:    "2026-02-14",
		Started: time.Date(2026, 2, 14, 9, 0, 0, 0, time.UTC),
		Title:   "Test Session",
	}
	require.NoError(t, store.Save(s))

	loaded, err := store.Load("abc123")
	require.NoError(t, err)
	assert.Equal(t, "abc123", loaded.ID)
	assert.Equal(t, "Test Session", loaded.Title)
}

func TestStore_ListRecent(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	for i := 0; i < 5; i++ {
		s := &Session{
			Version: "2.0",
			ID:      fmt.Sprintf("id%d", i),
			Date:    "2026-02-14",
			Started: time.Date(2026, 2, 14, i, 0, 0, 0, time.UTC),
		}
		require.NoError(t, store.Save(s))
	}

	sessions, err := store.List(3)
	require.NoError(t, err)
	assert.Len(t, sessions, 3)
	// Most recent first
	assert.Equal(t, "id4", sessions[0].ID)
}
```

**Step 2: Implement store.go**

`Session` struct matching design doc format. `Store` with `Save`, `Load`, `List`, `FindByDate`, `Search` methods. Files stored as `.claude/sessions/{date}-{id}.json`.

**Step 3: Write alias test**

```go
func TestAlias_SetAndResolve(t *testing.T) {
	tmpDir := t.TempDir()
	am := NewAliasManager(filepath.Join(tmpDir, "aliases.json"))

	require.NoError(t, am.Set("mywork", "abc123"))
	id, err := am.Resolve("mywork")
	require.NoError(t, err)
	assert.Equal(t, "abc123", id)
}
```

**Step 4: Implement alias.go**

JSON file at `.claude/session-aliases.json`. `AliasManager` with `Set`, `Resolve`, `Remove`, `List`.

**Step 5: Write transcript parsing test**

```go
func TestParseTranscript_ExtractsSummary(t *testing.T) {
	tmpDir := t.TempDir()
	lines := []string{
		`{"type":"human","content":"Fix the bug"}`,
		`{"type":"tool_use","name":"Edit","input":{"file_path":"main.go"}}`,
		`{"type":"tool_use","name":"Bash","input":{"command":"go test"}}`,
		`{"type":"human","content":"Looks good, commit it"}`,
	}
	tPath := filepath.Join(tmpDir, "transcript.jsonl")
	require.NoError(t, os.WriteFile(tPath, []byte(strings.Join(lines, "\n")), 0o644))

	summary, err := ParseTranscript(tPath)
	require.NoError(t, err)
	assert.Equal(t, 2, summary.TotalMessages)
	assert.Contains(t, summary.ToolsUsed, "Edit")
	assert.Contains(t, summary.ToolsUsed, "Bash")
	assert.Contains(t, summary.FilesModified, "main.go")
}
```

**Step 6: Implement transcript.go**

Read JSONL line by line. Count `"type":"human"` entries. Collect unique tool names and file paths.

**Step 7: Run all session tests, check, commit**

```bash
gotestsum --format pkgname -- -tags=testmode ./internal/session/...
task check
git add internal/session/
git commit -m "feat: add session package with JSON store, aliases, and transcript parsing"
```

---

## Task 10: Wire all handlers into the registry

Connect all implemented handlers to `buildHookRegistry()` in main.go.

**Files:**
- Modify: `cmd/cc-tools/main.go` (buildHookRegistry function)

**Step 1: Create handler wrapper types**

Each domain package needs a struct implementing `hookcmd.Handler`. Create thin wrapper types that bridge each package to the Handler interface. These can live in the domain packages themselves (e.g., `compact.SuggestHandler`, `notify.AudioHandler`, etc.) or as anonymous handlers in `buildHookRegistry`.

**Step 2: Implement buildHookRegistry**

Load config. Build and return the registry map:

```go
func buildHookRegistry() map[string][]hookcmd.Handler {
	cfg := loadHookConfig()
	return map[string][]hookcmd.Handler{
		"SessionStart":        {sessionLoadHandler(cfg), superpowersHandler(), pkgManagerHandler()},
		"SessionEnd":          {sessionSaveHandler(cfg), evaluateLearningHandler(cfg)},
		"PreToolUse":          {suggestCompactHandler(cfg), preCommitHandler(cfg), observeHandler(cfg, "pre")},
		"PostToolUse":         {observeHandler(cfg, "post")},
		"PostToolUseFailure":  {observeHandler(cfg, "failure")},
		"PreCompact":          {logCompactionHandler()},
		"Stop":                {stopGuardHandler(), audioHandler(cfg), desktopHandler(cfg), evaluateLearningHandler(cfg)},
		"Notification":        {audioHandler(cfg), desktopHandler(cfg)},
	}
}
```

The `stopGuardHandler` checks `stop_hook_active` and returns early if true.

**Step 3: Run full check**

Run: `task check`
Expected: All pass

**Step 4: Manual integration test**

Run: `echo '{"hook_event_name":"PreToolUse","session_id":"test","tool_name":"Bash","tool_input":{"command":"ls"}}' | go run ./cmd/cc-tools hook pre-tool-use`
Expected: exits 0, may log to stderr

**Step 5: Commit**

```bash
git add cmd/cc-tools/main.go
git commit -m "feat: wire all hook handlers into registry"
```

---

## Task 11: Session command implementation

Implement `cc-tools session` subcommands using the session package.

**Files:**
- Modify: `cmd/cc-tools/main.go` (runSessionCommand)

**Step 1: Implement runSessionCommand**

Switch on `os.Args[2]`:
- `list`: call `session.Store.List()`, format output table
- `load`: resolve alias if needed, load session, print summary to stdout
- `info`: load session, print full JSON details
- `alias`: set/remove alias
- `aliases`: list all aliases

**Step 2: Run full check and commit**

```bash
task check
git add cmd/cc-tools/main.go
git commit -m "feat: implement session command with list, load, info, alias subcommands"
```

---

## Task 12: Update mockery config and generate mocks

Add new interfaces to `.mockery.yml` and regenerate.

**Files:**
- Modify: `.mockery.yml`

**Step 1: Add new interfaces**

Add entries for any new interfaces (e.g., `notify.CmdRunner`, `notify.AudioPlayer`) to `.mockery.yml`.

**Step 2: Generate mocks**

Run: `task mocks`

**Step 3: Run full check and commit**

```bash
task check
git add .mockery.yml internal/*/mocks/
git commit -m "chore: add new interfaces to mockery and regenerate mocks"
```

---

## Task 13: Update .claude/commands/sessions.md

Update the sessions command to use cc-tools instead of inline JS.

**Files:**
- Modify: `.claude/commands/sessions.md`

**Step 1: Replace inline scripts with cc-tools calls**

Replace the Node.js session management logic with `cc-tools session list`, `cc-tools session load`, etc.

**Step 2: Commit**

```bash
git add .claude/commands/sessions.md
git commit -m "refactor: update sessions command to use cc-tools session"
```

---

## Task 14: Final integration test and cleanup

Verify end-to-end functionality and run all checks.

**Step 1: Run full test suite**

Run: `task check`
Expected: All pass, zero lint issues

**Step 2: Build binary**

Run: `task build`
Expected: Binary builds successfully

**Step 3: Manual smoke test — all hook events**

```bash
# SessionStart
echo '{"hook_event_name":"SessionStart","session_id":"test1","source":"startup","cwd":"/tmp"}' | ./bin/cc-tools hook session-start

# PreToolUse
echo '{"hook_event_name":"PreToolUse","session_id":"test1","tool_name":"Bash","tool_input":{"command":"ls"}}' | ./bin/cc-tools hook pre-tool-use

# PostToolUse
echo '{"hook_event_name":"PostToolUse","session_id":"test1","tool_name":"Edit","tool_input":{"file_path":"test.go"}}' | ./bin/cc-tools hook post-tool-use

# Stop (with stop_hook_active=false)
echo '{"hook_event_name":"Stop","session_id":"test1","stop_hook_active":false}' | ./bin/cc-tools hook stop

# Stop (with stop_hook_active=true — should exit immediately)
echo '{"hook_event_name":"Stop","session_id":"test1","stop_hook_active":true}' | ./bin/cc-tools hook stop

# Unknown event — should exit 0
echo '{"hook_event_name":"SubagentStart","session_id":"test1"}' | ./bin/cc-tools hook subagent-start

# SessionEnd
echo '{"hook_event_name":"SessionEnd","session_id":"test1","reason":"prompt_input_exit"}' | ./bin/cc-tools hook session-end
```

**Step 4: Commit any final fixes**

```bash
task check
git add -A
git commit -m "test: add integration smoke tests for all hook events"
```

---

## Task Summary

| # | Task | Depends On | Estimated Steps |
|---|------|-----------|----------------|
| 1 | Config extensions | — | 6 |
| 2 | hookcmd package (input, handler, dispatch) | — | 12 |
| 3 | Wire `cc-tools hook` into main | 2 | 6 |
| 4 | Notify package (quiet hours, desktop, audio) | 1 | 11 |
| 5 | Compact package (suggest, log) | 1 | 7 |
| 6 | Observe package (JSONL, rotation) | — | 5 |
| 7 | Superpowers injection | — | 3 |
| 8 | Package manager detection | — | 3 |
| 9 | Session store (JSON, aliases, transcript) | — | 7 |
| 10 | Wire handlers into registry | 1-9 | 5 |
| 11 | Session command implementation | 9 | 2 |
| 12 | Mockery config update | 4, 6 | 3 |
| 13 | Update sessions.md command | 11 | 2 |
| 14 | Final integration and smoke test | 10-13 | 4 |

**Parallelizable:** Tasks 1, 2, 6, 7, 8, 9 can all run in parallel (no dependencies). Tasks 4, 5 depend on Task 1 (config). Task 3 depends on Task 2. Task 10 depends on all packages. Tasks 11-14 are sequential.
