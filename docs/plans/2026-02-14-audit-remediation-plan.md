# Audit Remediation & Hook Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use executing-plans to implement this plan task-by-task.

**Goal:** Fix 6 architectural fragility clusters and migrate all JS/Python/Bash hooks to Go, enabling deletion of `.claude/hooks/`.

**Architecture:** Three-phase bottom-up approach: fix foundation issues (unified HookInput, stdin fix), migrate hooks to Go handlers, then quality improvements (config split, traversal check, discovery logging).

**Tech Stack:** Go 1.26, lipgloss, testify, Mockery v3.5. No new dependencies.

---

### Task 1: Add IsEditTool and GetFilePath to hookcmd.HookInput

**Files:**
- Modify: `internal/hookcmd/input.go`
- Modify: `internal/hookcmd/input_test.go`

**Step 1: Write failing tests for IsEditTool**

Add to `internal/hookcmd/input_test.go`:

```go
func TestHookInput_IsEditTool(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		want     bool
	}{
		{name: "Edit is edit tool", toolName: "Edit", want: true},
		{name: "MultiEdit is edit tool", toolName: "MultiEdit", want: true},
		{name: "Write is edit tool", toolName: "Write", want: true},
		{name: "NotebookEdit is edit tool", toolName: "NotebookEdit", want: true},
		{name: "Read is not edit tool", toolName: "Read", want: false},
		{name: "Bash is not edit tool", toolName: "Bash", want: false},
		{name: "empty is not edit tool", toolName: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &hookcmd.HookInput{ToolName: tt.toolName}
			assert.Equal(t, tt.want, input.IsEditTool())
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestHookInput_IsEditTool ./internal/hookcmd/`
Expected: FAIL — `IsEditTool` method not found.

**Step 3: Write failing tests for GetFilePath**

Add to `internal/hookcmd/input_test.go`:

```go
func TestHookInput_GetFilePath(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		input    string
		want     string
	}{
		{
			name:     "extracts file_path from Edit tool",
			toolName: "Edit",
			input:    `{"file_path": "/tmp/test.go"}`,
			want:     "/tmp/test.go",
		},
		{
			name:     "extracts notebook_path from NotebookEdit",
			toolName: "NotebookEdit",
			input:    `{"notebook_path": "/tmp/nb.ipynb"}`,
			want:     "/tmp/nb.ipynb",
		},
		{
			name:     "returns empty for missing file_path",
			toolName: "Edit",
			input:    `{"other": "value"}`,
			want:     "",
		},
		{
			name:     "returns empty for empty tool_input",
			toolName: "Edit",
			input:    "",
			want:     "",
		},
		{
			name:     "returns empty for invalid JSON",
			toolName: "Edit",
			input:    `{invalid`,
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &hookcmd.HookInput{
				ToolName:  tt.toolName,
				ToolInput: json.RawMessage(tt.input),
			}
			assert.Equal(t, tt.want, input.GetFilePath())
		})
	}
}
```

**Step 4: Run test to verify it fails**

Run: `go test -v -run TestHookInput_GetFilePath ./internal/hookcmd/`
Expected: FAIL — `GetFilePath` method not found.

**Step 5: Implement IsEditTool and GetFilePath**

Add to `internal/hookcmd/input.go`:

```go
// IsEditTool returns true if the tool modifies files.
func (h *HookInput) IsEditTool() bool {
	switch h.ToolName {
	case "Edit", "MultiEdit", "Write", "NotebookEdit":
		return true
	default:
		return false
	}
}

// GetFilePath extracts the file path from tool_input based on tool type.
func (h *HookInput) GetFilePath() string {
	if len(h.ToolInput) == 0 {
		return ""
	}

	var fields map[string]any
	if err := json.Unmarshal(h.ToolInput, &fields); err != nil {
		return ""
	}

	// NotebookEdit uses notebook_path.
	if h.ToolName == "NotebookEdit" {
		if p, ok := fields["notebook_path"].(string); ok {
			return p
		}
	}

	if p, ok := fields["file_path"].(string); ok {
		return p
	}

	return ""
}
```

**Step 6: Run tests to verify they pass**

Run: `go test -v -run 'TestHookInput_(IsEditTool|GetFilePath)' ./internal/hookcmd/`
Expected: PASS

**Step 7: Commit**

```bash
git add internal/hookcmd/input.go internal/hookcmd/input_test.go
git commit -m "feat: add IsEditTool and GetFilePath methods to hookcmd.HookInput"
```

---

### Task 2: Migrate validate to use hookcmd.HookInput

**Files:**
- Modify: `internal/hooks/validate.go:254-321` (runValidateHookInternal)
- Modify: `internal/hooks/validate_skip.go` (ValidateWithSkipCheck)
- Modify: `internal/hooks/executor.go:119-138` (validateHookEvent, handleInputError)
- Modify: `internal/hooks/validate_test.go`
- Modify: `internal/hooks/validate_skip_test.go`
- Modify: `internal/hooks/hooks_test.go`
- Delete: `internal/hooks/input.go`
- Delete: `internal/hooks/input_test.go`

This is a refactor task. The key changes:

1. `runValidateHookInternal` currently calls `ReadHookInput(deps.Input)` which returns `*hooks.HookInput`. Change it to accept `*hookcmd.HookInput` as a parameter instead.
2. `validateHookEvent` currently takes `*hooks.HookInput`. Change to `*hookcmd.HookInput`.
3. `ValidateWithSkipCheck` currently reads stdin itself. Change to accept `[]byte` and parse via `hookcmd.ParseInput`.
4. Remove `hooks.HookInput`, `ReadHookInput`, `ErrNoInput`, `InputReader`, `bytesInputReader`.
5. Move `ErrNoInput` to `hookcmd` package if still needed, or remove entirely (callers check for empty input differently).

**Step 1: Update ValidateWithSkipCheck signature**

Change `internal/hooks/validate_skip.go`:

```go
// ValidateWithSkipCheck parses stdin bytes, checks skip registry, and runs validation.
func ValidateWithSkipCheck(
	ctx context.Context,
	stdinData []byte,
	stdout io.Writer,
	stderr io.Writer,
	debug bool,
	timeoutSecs int,
	cooldownSecs int,
) int {
```

Replace `io.ReadAll(stdin)` call with using `stdinData` directly. Remove the `bytesInputReader` struct — it's no longer needed. Pass a parsed `*hookcmd.HookInput` to `runValidateHookInternal`.

**Step 2: Update runValidateHookInternal to accept parsed input**

Change signature to accept `*hookcmd.HookInput`:

```go
func runValidateHookInternal(
	ctx context.Context,
	input *hookcmd.HookInput,
	debug bool,
	timeoutSecs int,
	cooldownSecs int,
	skipConfig *SkipConfig,
	deps *Dependencies,
) int {
```

Remove the `ReadHookInput(deps.Input)` call. Use `input.IsEditTool()`, `input.GetFilePath()`, `input.HookEventName` from the hookcmd struct.

**Step 3: Update validateHookEvent**

Change to accept `*hookcmd.HookInput`:

```go
func validateHookEvent(input *hookcmd.HookInput, debug bool, stderr io.Writer) (string, bool) {
	if input == nil || input.HookEventName != "PostToolUse" || !input.IsEditTool() {
```

**Step 4: Remove InputReader from Dependencies**

In `internal/hooks/dependencies.go`, remove the `Input InputReader` field and the `InputReader` interface. Remove `OutputWriter` if it's only used for `stderr` typing (check if `io.Writer` suffices).

**Step 5: Update RunValidateHook and RunValidateHookWithSkip**

Both now accept `*hookcmd.HookInput`:

```go
func RunValidateHookWithSkip(ctx context.Context, input *hookcmd.HookInput, debug bool, timeoutSecs int, cooldownSecs int, skipConfig *SkipConfig, deps *Dependencies) int {
	return runValidateHookInternal(ctx, input, debug, timeoutSecs, cooldownSecs, skipConfig, deps)
}
```

**Step 6: Update cmd/cc-tools/main.go runValidate()**

In `runValidate()`, parse stdinData with `hookcmd.ParseInput` and pass the result:

```go
func runValidate(stdinData []byte) {
	timeoutSecs, cooldownSecs := loadValidateConfig()
	debug := os.Getenv("CLAUDE_HOOKS_DEBUG") == "1"

	input, err := hookcmd.ParseInput(bytes.NewReader(stdinData))
	if err != nil {
		// Can't parse input — exit silently (hooks must not block).
		os.Exit(0)
	}

	exitCode := hooks.ValidateWithSkipCheck(
		context.Background(),
		input,
		stdinData,
		os.Stdout,
		os.Stderr,
		debug,
		timeoutSecs,
		cooldownSecs,
	)
	os.Exit(exitCode)
}
```

**Step 7: Delete hooks/input.go and hooks/input_test.go**

Remove `internal/hooks/input.go` and `internal/hooks/input_test.go`. All functionality now lives in `hookcmd`.

**Step 8: Update all tests**

Update test files that create `hooks.HookInput` to use `hookcmd.HookInput`. Update mock setup for `InputReader` removal.

**Step 9: Run full test suite**

Run: `task test`
Expected: All tests pass.

Run: `task lint`
Expected: Zero issues.

**Step 10: Commit**

```bash
git add -A internal/hooks/ internal/hookcmd/ cmd/cc-tools/
git commit -m "refactor: unify HookInput into hookcmd package

Merge hooks.HookInput and hookcmd.HookInput into a single struct in
hookcmd. Remove InputReader interface from hooks.Dependencies. Validate
functions now accept parsed *hookcmd.HookInput instead of reading stdin."
```

---

### Task 3: Fix stdin consumption pattern

**Files:**
- Modify: `cmd/cc-tools/main.go`

**Step 1: Rewrite main() to read stdin once**

Replace `debugLog()` (lines 561-618) and `main()` (lines 42-80):

```go
func main() {
	out := output.NewTerminal(os.Stdout, os.Stderr)

	// Read stdin once for commands that need it.
	var stdinData []byte
	if len(os.Args) > 1 && needsStdin(os.Args[1]) {
		if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
			stdinData, _ = io.ReadAll(os.Stdin)
		}
	}

	// Debug log (never mutates os.Stdin).
	writeDebugLog(os.Args, stdinData)

	if len(os.Args) < minArgs {
		printUsage(out)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "validate":
		runValidate(stdinData)
	case "hook":
		runHookCommand(stdinData)
	// ... remaining cases unchanged ...
	}
}

func needsStdin(cmd string) bool {
	return cmd == "validate" || cmd == "hook"
}
```

**Step 2: Rewrite debugLog as writeDebugLog**

```go
func writeDebugLog(args []string, stdinData []byte) {
	debugFile := getDebugLogPath()

	f, err := os.OpenFile(debugFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	_, _ = fmt.Fprintf(f, "\n========================================\n")
	_, _ = fmt.Fprintf(f, "[%s] cc-tools invoked\n", timestamp)
	_, _ = fmt.Fprintf(f, "Args: %v\n", args)
	_, _ = fmt.Fprintf(f, "  CLAUDE_HOOKS_DEBUG: %s\n", os.Getenv("CLAUDE_HOOKS_DEBUG"))

	if wd, wdErr := os.Getwd(); wdErr == nil {
		_, _ = fmt.Fprintf(f, "  Working Dir: %s\n", wd)
	}

	if len(stdinData) > 0 {
		_, _ = fmt.Fprintf(f, "Stdin: %s\n", string(stdinData))
	} else {
		_, _ = fmt.Fprintf(f, "Stdin: (no data)\n")
	}
}
```

**Step 3: Remove stdinTempFile global and all its references**

Delete the `var stdinTempFile string` global. Remove `stdinTempFile` cleanup in `runValidate()` and `runHookCommand()`.

**Step 4: Update runHookCommand to accept stdinData**

```go
func runHookCommand(stdinData []byte) {
	out := output.NewTerminal(os.Stdout, os.Stderr)
	// ... arg validation unchanged ...

	input, err := hookcmd.ParseInput(bytes.NewReader(stdinData))
	if err != nil {
		_ = out.Error("error parsing hook input: %v", err)
		os.Exit(0)
	}
	input.HookEventName = eventName

	registry := buildHookRegistry()
	exitCode := hookcmd.Dispatch(context.Background(), input, os.Stdout, os.Stderr, registry)
	os.Exit(exitCode)
}
```

**Step 5: Run tests and lint**

Run: `task test && task lint`
Expected: All pass, zero lint issues.

**Step 6: Commit**

```bash
git add cmd/cc-tools/main.go
git commit -m "refactor: read stdin once in main(), eliminate temp file pattern

Replace debugLog() stdin consumption + temp file + os.Stdin mutation with
a single io.ReadAll at top of main(). Pass stdinData as []byte to both
debug logging and command handlers. Remove stdinTempFile global."
```

---

### Task 4: Remove stopGuardHandler and wire notification handlers

**Files:**
- Modify: `cmd/cc-tools/main.go`

**Step 1: Delete stopGuardHandler**

Remove `stopGuardHandler()` function (lines 270-280) and its entry in `buildHookRegistry()`.

**Step 2: Wire notifyAudioHandler**

Replace the stub with a real implementation:

```go
func notifyAudioHandler(cfg *config.Values) *handlerFunc {
	return &handlerFunc{
		name: "notify-audio",
		fn: func(_ context.Context, _ *hookcmd.HookInput, _, _ io.Writer) error {
			if cfg == nil || !cfg.Notify.Audio.Enabled {
				return nil
			}

			player := &afPlayer{}
			qh := notify.QuietHours{
				Enabled: cfg.Notify.QuietHours.Enabled,
				Start:   cfg.Notify.QuietHours.Start,
				End:     cfg.Notify.QuietHours.End,
			}
			audio := notify.NewAudio(player, cfg.Notify.Audio.Directory, qh, nil)
			return audio.PlayRandom()
		},
	}
}

// afPlayer implements notify.AudioPlayer using macOS afplay.
type afPlayer struct{}

func (a *afPlayer) Play(filepath string) error {
	return exec.Command("afplay", filepath).Run()
}
```

**Step 3: Wire notifyDesktopHandler**

Replace the stub:

```go
func notifyDesktopHandler(cfg *config.Values) *handlerFunc {
	return &handlerFunc{
		name: "notify-desktop",
		fn: func(_ context.Context, input *hookcmd.HookInput, _, _ io.Writer) error {
			if cfg == nil || !cfg.Notify.Desktop.Enabled {
				return nil
			}

			qh := notify.QuietHours{
				Enabled: cfg.Notify.QuietHours.Enabled,
				Start:   cfg.Notify.QuietHours.Start,
				End:     cfg.Notify.QuietHours.End,
			}
			if qh.IsActive(time.Now()) {
				return nil
			}

			runner := &osascriptRunner{}
			desktop := notify.NewDesktop(runner)

			title := "Claude Code"
			message := "Task completed"
			if input.Title != "" {
				title = input.Title
			}
			if input.Message != "" {
				message = input.Message
			}

			return desktop.Send(title, message)
		},
	}
}

// osascriptRunner implements notify.CmdRunner.
type osascriptRunner struct{}

func (o *osascriptRunner) Run(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}
```

**Step 4: Update buildHookRegistry**

Remove `Stop` → `stopGuardHandler` entry. Add audio + desktop to `Stop` event:

```go
"Stop":         {notifyAudioHandler(cfg), notifyDesktopHandler(cfg)},
```

**Step 5: Add `os/exec` import if not already present**

**Step 6: Run tests and lint**

Run: `task test && task lint`
Expected: All pass, zero lint issues.

**Step 7: Commit**

```bash
git add cmd/cc-tools/main.go
git commit -m "feat: wire notification handlers, remove stop guard no-op

Connect internal/notify Audio and Desktop implementations to the handler
functions. Add afPlayer and osascriptRunner adapters for real command
execution. Remove the no-op stopGuardHandler."
```

---

### Task 5: Add session context handler (SessionStart)

**Files:**
- Modify: `cmd/cc-tools/main.go`

**Step 1: Implement sessionContextHandler**

```go
func sessionContextHandler() *handlerFunc {
	return &handlerFunc{
		name: "session-context",
		fn: func(_ context.Context, _ *hookcmd.HookInput, out, errOut io.Writer) error {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("get home directory: %w", err)
			}

			storeDir := filepath.Join(homeDir, ".claude", "sessions")
			store := session.NewStore(storeDir)

			// Load most recent session.
			sessions, listErr := store.List(1)
			if listErr != nil || len(sessions) == 0 {
				return nil // No sessions — nothing to inject.
			}

			latest := sessions[0]
			if latest.Summary != "" {
				_, _ = fmt.Fprintf(out, "Previous session (%s): %s\n", latest.Date, latest.Summary)
			}

			// Report aliases.
			aliasFile := filepath.Join(homeDir, ".claude", "session-aliases.json")
			aliases := session.NewAliasManager(aliasFile)
			aliasList, aliasErr := aliases.List()
			if aliasErr == nil && len(aliasList) > 0 {
				names := make([]string, 0, len(aliasList))
				for name := range aliasList {
					names = append(names, name)
				}
				_, _ = fmt.Fprintf(errOut, "[session-context] %d alias(es): %s\n",
					len(aliasList), strings.Join(names, ", "))
			}

			return nil
		},
	}
}
```

**Step 2: Add to buildHookRegistry**

```go
"SessionStart": {superpowersHandler(), pkgManagerHandler(), sessionContextHandler()},
```

**Step 3: Run tests and lint**

Run: `task test && task lint`

**Step 4: Commit**

```bash
git add cmd/cc-tools/main.go
git commit -m "feat: add session context handler for SessionStart

Load most recent session summary and report active aliases on session
start. Replaces session-start.js."
```

---

### Task 6: Add session end handler (SessionEnd)

**Files:**
- Modify: `cmd/cc-tools/main.go`

**Step 1: Implement sessionEndHandler**

```go
func sessionEndHandler(cfg *config.Values) *handlerFunc {
	return &handlerFunc{
		name: "session-end",
		fn: func(_ context.Context, input *hookcmd.HookInput, _, errOut io.Writer) error {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("get home directory: %w", err)
			}

			storeDir := filepath.Join(homeDir, ".claude", "sessions")
			store := session.NewStore(storeDir)

			// Parse transcript if available.
			var summary *session.TranscriptSummary
			if input.TranscriptPath != "" {
				summary, _ = session.ParseTranscript(input.TranscriptPath)
			}

			// Build session metadata.
			now := time.Now()
			sess := &session.Session{
				ID:      input.SessionID,
				Date:    now.Format("2006-01-02"),
				Started: now,
				Ended:   now,
				Title:   fmt.Sprintf("Session %s", now.Format("15:04")),
			}

			if summary != nil {
				sess.ToolsUsed = summary.ToolsUsed
				sess.FilesModified = summary.FilesModified
				sess.MessageCount = summary.TotalMessages
			}

			if saveErr := store.Save(sess); saveErr != nil {
				_, _ = fmt.Fprintf(errOut, "[session-end] save error: %v\n", saveErr)
			}

			// Continuous learning signal.
			minLength := 10
			if cfg != nil && cfg.Learning.MinSessionLength > 0 {
				minLength = cfg.Learning.MinSessionLength
			}
			if summary != nil && summary.TotalMessages >= minLength {
				_, _ = fmt.Fprintf(errOut,
					"[session-end] %d messages — evaluate for extractable patterns\n",
					summary.TotalMessages)
			}

			return nil
		},
	}
}
```

**Step 2: Add SessionEnd to buildHookRegistry**

```go
"SessionEnd": {sessionEndHandler(cfg)},
```

**Step 3: Run tests and lint**

Run: `task test && task lint`

**Step 4: Commit**

```bash
git add cmd/cc-tools/main.go
git commit -m "feat: add session end handler for SessionEnd

Extract transcript summary on session end, save session metadata, and
signal continuous learning for sessions above message threshold. Replaces
session-end.js and evaluate-session.js."
```

---

### Task 7: Add pre-commit reminder handler (PreToolUse)

**Files:**
- Modify: `cmd/cc-tools/main.go`

**Step 1: Implement preCommitReminderHandler**

```go
func preCommitReminderHandler(cfg *config.Values) *handlerFunc {
	return &handlerFunc{
		name: "pre-commit-reminder",
		fn: func(_ context.Context, input *hookcmd.HookInput, _, errOut io.Writer) error {
			if cfg == nil || !cfg.PreCommit.Enabled {
				return nil
			}

			if input.ToolName != "Bash" {
				return nil
			}

			command := input.GetToolInputString("command")
			if strings.Contains(command, "git commit") {
				reminder := cfg.PreCommit.Command
				if reminder == "" {
					reminder = "task pre-commit"
				}
				_, _ = fmt.Fprintf(errOut, "Reminder: Run '%s' (fmt + lint + test) before committing.\n", reminder)
			}

			return nil
		},
	}
}
```

**Step 2: Add to PreToolUse handler chain**

```go
"PreToolUse": {suggestCompactHandler(cfg), observeHandler(cfg, "pre"), preCommitReminderHandler(cfg)},
```

**Step 3: Run tests and lint**

Run: `task test && task lint`

**Step 4: Commit**

```bash
git add cmd/cc-tools/main.go
git commit -m "feat: add pre-commit reminder handler for PreToolUse

Check if Bash tool input contains 'git commit' and print a reminder to
run pre-commit checks. Replaces pre-commit-reminder.sh."
```

---

### Task 8: Update .claude/settings.json and delete migrated scripts

**Files:**
- Modify: `.claude/settings.json`
- Delete: `.claude/hooks/start-superpowers.sh`
- Delete: `.claude/hooks/suggest-compact.js`
- Delete: `.claude/hooks/pre-compact.js`
- Delete: `.claude/hooks/session-start.js`
- Delete: `.claude/hooks/session-end.js`
- Delete: `.claude/hooks/evaluate-session.js`
- Delete: `.claude/hooks/setup-package-manager.js`
- Delete: `.claude/hooks/play_audio.py`
- Delete: `.claude/hooks/macos_notification.py`
- Delete: `.claude/hooks/pre-commit-reminder.sh`
- Delete: `.claude/hooks/lib/` (entire directory)
- Delete: `.claude/hooks/utils/` (entire directory)

**Step 1: Update .claude/settings.json**

Replace JS/Python/Bash hook commands with cc-tools. Keep external hooks (`sentinel`, `continuous-learning-v2/observe.sh`, `claude-docs-helper.sh`).

**Step 2: Delete all migrated scripts**

```bash
rm .claude/hooks/start-superpowers.sh
rm .claude/hooks/suggest-compact.js
rm .claude/hooks/pre-compact.js
rm .claude/hooks/session-start.js
rm .claude/hooks/session-end.js
rm .claude/hooks/evaluate-session.js
rm .claude/hooks/setup-package-manager.js
rm .claude/hooks/play_audio.py
rm .claude/hooks/macos_notification.py
rm .claude/hooks/pre-commit-reminder.sh
rm -rf .claude/hooks/lib/
rm -rf .claude/hooks/utils/
```

**Step 3: Verify no scripts remain that should have been migrated**

```bash
ls .claude/hooks/
```

Expected: Empty directory (or only files not managed by cc-tools).

**Step 4: Verify hooks work**

Test each event type manually:

```bash
echo '{"hook_event_name":"SessionStart","session_id":"test"}' | cc-tools hook session-start
echo '{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"git commit -m test"}}' | cc-tools hook pre-tool-use
echo '{"hook_event_name":"PreCompact"}' | cc-tools hook pre-compact
echo '{"hook_event_name":"Notification","title":"Test","message":"Hello"}' | cc-tools hook notification
echo '{"hook_event_name":"SessionEnd","session_id":"test-123","transcript_path":"/tmp/fake.jsonl"}' | cc-tools hook session-end
```

**Step 5: Run full test suite**

Run: `task test && task lint`

**Step 6: Commit**

```bash
git add .claude/settings.json
git add .claude/hooks/  # stages deletions
git commit -m "chore: migrate hooks to cc-tools and delete JS/Python/Bash scripts

Update .claude/settings.json to use cc-tools for all hook events.
Delete 10 scripts (2,750 lines of JS/Python/Bash) now replaced by Go
handlers in cmd/cc-tools/main.go."
```

---

### Task 9: Split config/manager.go into three files

**Files:**
- Modify: `internal/config/manager.go`
- Create: `internal/config/keys.go`
- Create: `internal/config/values.go`

**Step 1: Extract Values and section structs to values.go**

Move `Values`, `ValidateValues`, `CompactValues`, `NotifyValues`, `ObserveValues`, `LearningValues`, `PreCommitValues`, `NotificationsValues`, and any backward-compat parsing to `internal/config/values.go`.

**Step 2: Extract key constants and definitions to keys.go**

Move all `key*` constants, `keyDefinition` type (if it exists, or create), validation logic, and default values to `internal/config/keys.go`.

**Step 3: Keep Manager in manager.go**

Keep `Manager` struct, `NewManager()`, `Get/Set/Reset/GetAll/GetConfig`, load/save in `manager.go`.

**Step 4: Verify no behavior change**

Run: `task test && task lint`
Expected: All 646+ tests pass, zero lint issues.

**Step 5: Commit**

```bash
git add internal/config/
git commit -m "refactor: split config/manager.go into manager, keys, and values

Extract Values structs to values.go and key constants with validation
to keys.go. No behavior changes — pure file reorganization to reduce
cognitive load (844 lines to ~300 each)."
```

---

### Task 10: Add skip directory traversal check

**Files:**
- Modify: `cmd/cc-tools/skip.go`

**Step 1: Add path validation function**

```go
func validateSkipPath(dir string) (string, error) {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}

	cleanPath := filepath.Clean(absPath)
	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("invalid path: directory traversal not allowed")
	}

	return cleanPath, nil
}
```

**Step 2: Call validateSkipPath in addSkip()**

Add validation before calling `registry.Skip()`.

**Step 3: Run tests and lint**

Run: `task test && task lint`

**Step 4: Commit**

```bash
git add cmd/cc-tools/skip.go
git commit -m "fix: reject directory traversal in skip command paths

Validate that skip directory paths do not contain '..' after cleanup.
Prevents traversal attacks via the skip command."
```

---

### Task 11: Add discovery error visibility in debug mode

**Files:**
- Modify: `internal/hooks/discovery.go`

**Step 1: Add debug logging to discovery methods**

In each discovery method that returns `nil` on failure, add debug logging. The `CommandDiscovery` struct has access to `deps.Stderr` for debug output. Add a `debug` field or pass debug state.

Key locations to add logging:
- When Makefile/Taskfile/justfile target is not found
- When package.json script is not found
- When scripts directory doesn't exist
- When language-specific tool is not found via LookPath

Pattern:

```go
if ce.debug {
    _, _ = fmt.Fprintf(ce.stderr, "[discovery] %s: target %q not found in %s\n", runner, target, dir)
}
```

**Step 2: Run tests and lint**

Run: `task test && task lint`

**Step 3: Commit**

```bash
git add internal/hooks/discovery.go
git commit -m "feat: surface command discovery failures in debug mode

Add debug logging for each failed discovery attempt (task runner
targets, package.json scripts, language-specific tools). Failures were
previously silent."
```

---

### Task 12: Final verification and rebuild binary

**Step 1: Run all checks**

```bash
task test
task lint
task test-race
```

**Step 2: Rebuild cc-tools binary**

```bash
go build -o ~/.claude/bin/cc-tools-validate ./cmd/cc-tools/
```

**Step 3: Verify grep checks**

```bash
# No hooks scripts remaining
ls .claude/hooks/

# All tests pass
task test
```

**Step 4: Final commit if needed**

Any cleanup or fixes discovered during verification.
