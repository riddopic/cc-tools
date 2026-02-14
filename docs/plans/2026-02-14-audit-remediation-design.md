# Audit Remediation & Hook Migration Design

**Date:** 2026-02-14
**Source:** `docs/audits/context-audit-20260214-051757.md`
**Scope:** Fix 6 fragility clusters, migrate 2,750 lines of JS/Python/Bash hooks to Go, address open questions

---

## Context

cc-tools has:
- 6 fragility clusters identified by context audit
- 12 hook scripts in `.claude/hooks/` (JS/Python/Bash, 2,750 lines) that duplicate or extend Go implementations in `internal/`
- 8 open architectural questions

This design addresses all three categories in dependency order.

## Decisions Made

1. **Interleave** architectural fixes and hook migration (fix foundation → migrate → quality)
2. **Unify** `hooks.HookInput` and `hookcmd.HookInput` into single struct in `hookcmd`
3. **Wire** Go notification implementations, delete Python scripts
4. **Read stdin once** in `main()`, pass as `[]byte` — no temp files, no global mutation
5. **Merge** `evaluate-session.js` into `sessionEndHandler()`
6. **Split** `config/manager.go` (844 lines) into 3 files
7. **Defer** command discovery refactoring and lock manager PID reuse fix

---

## Phase 1: Architectural Foundation

### 1.1 Unify HookInput

**Problem**: Two `HookInput` structs (`hooks.HookInput` and `hookcmd.HookInput`) parse the same stdin JSON with different field names (`EventName` vs `HookEventName`).

**Solution**: `hookcmd.HookInput` becomes the canonical struct.

**Changes**:
- Add `IsEditTool()`, `GetFilePath()` methods to `hookcmd.HookInput` (currently on `hooks.HookInput`)
- Update `hooks/validate.go` and `hooks/validate_skip.go` to accept `*hookcmd.HookInput` instead of `io.Reader`
- Delete `internal/hooks/input.go`
- Update all callers in `cmd/cc-tools/main.go`

**Files**: `internal/hookcmd/input.go`, `internal/hooks/input.go` (delete), `internal/hooks/validate.go`, `internal/hooks/validate_skip.go`, `internal/hooks/executor.go`, `cmd/cc-tools/main.go`

### 1.2 Fix Stdin Consumption

**Problem**: `debugLog()` reads all of stdin, writes to temp file, replaces `os.Stdin` globally. Fragile: temp file failure loses stdin silently.

**Solution**: Read stdin once at top of `main()`, pass `[]byte` to both `debugLog()` and command handlers.

```go
func main() {
    var stdinData []byte
    if needsStdin(os.Args) {
        stdinData, _ = io.ReadAll(os.Stdin)
    }
    debugLog(os.Args, stdinData)  // log only, no mutation
    // pass stdinData to handlers...
}
```

**Files**: `cmd/cc-tools/main.go`

### 1.3 Remove stopGuardHandler

**Problem**: No-op handler — checks `StopHookActive` but does nothing in either branch.

**Solution**: Delete handler and its entry in `buildHookRegistry()`.

**Files**: `cmd/cc-tools/main.go`

---

## Phase 2: Hook Migration

### Migration Map

| Script | Event | Go Equivalent | Action |
|---|---|---|---|
| `start-superpowers.sh` | SessionStart | `superpowersHandler()` | Delete script (Go already wired) |
| `suggest-compact.js` | PreToolUse | `suggestCompactHandler()` | Delete script (Go already wired) |
| `pre-compact.js` | PreCompact | `logCompactionHandler()` | Delete script (Go already wired) |
| `setup-package-manager.js` | standalone | `pkgManagerHandler()` | Delete script (Go already wired) |
| `play_audio.py` | Stop/Notification | `notifyAudioHandler()` | Wire Go implementation |
| `macos_notification.py` | Stop/Notification | `notifyDesktopHandler()` | Wire Go implementation |
| `session-start.js` | SessionStart | New: `sessionContextHandler()` | Implement in Go |
| `session-end.js` | SessionEnd | New: `sessionEndHandler()` | Implement in Go |
| `evaluate-session.js` | SessionEnd | Merged into `sessionEndHandler()` | Implement in Go |
| `pre-commit-reminder.sh` | PreToolUse(Bash) | New: `preCommitReminderHandler()` | Implement in Go |

### 2.1 Wire Notification Handlers

Replace stub handlers with real implementations:

**`notifyAudioHandler()`**: Instantiate `notify.NewAudio()` with config values (audio directory from `cfg.Notify.Audio.Directory`, quiet hours from config). Call `PlayRandom()`.

**`notifyDesktopHandler()`**: Instantiate `notify.NewDesktop()` with a real `CmdRunner` that wraps `exec.Command("osascript", ...)`. Call `Send()` with title/message derived from hook input.

**Files**: `cmd/cc-tools/main.go`

### 2.2 Session Context Handler (SessionStart)

New handler `sessionContextHandler()` replaces `session-start.js`:

1. Read latest session file from `~/.claude/sessions/` via `session.Store.List(1)`
2. If found, output session summary to stdout (becomes Claude context)
3. List active aliases via `session.AliasManager.List()`
4. Log alias count and names to stderr

**Files**: `cmd/cc-tools/main.go`

### 2.3 Session End Handler (SessionEnd)

New handler `sessionEndHandler()` replaces `session-end.js` + `evaluate-session.js`:

1. Get `transcript_path` from hook input
2. Parse transcript via `session.ParseTranscript()` (already exists)
3. Build session metadata (date, ID, tools used, files modified, message count)
4. Save via `session.Store.Save()`
5. If message count >= configured threshold, log continuous learning signal

Add `SessionEnd` to `buildHookRegistry()`.

**Files**: `cmd/cc-tools/main.go`

### 2.4 Pre-Commit Reminder Handler (PreToolUse)

New handler `preCommitReminderHandler()` replaces `pre-commit-reminder.sh`:

1. Check if tool is `Bash`
2. Parse `tool_input.command` from hook input
3. If command contains `git commit`, print reminder to stderr
4. Non-blocking (always returns nil)

Add to PreToolUse handler chain.

**Files**: `cmd/cc-tools/main.go`

### 2.5 Update .claude/settings.json

Replace JS/Python/Bash hook entries with cc-tools:

```json
{
  "SessionStart": [
    {"matcher": "*", "hooks": [{"type": "command", "command": "cc-tools hook session-start"}]}
  ],
  "PreToolUse": [
    {"matcher": "*", "hooks": [{"type": "command", "command": "cc-tools hook pre-tool-use"}]},
    {"matcher": "*", "hooks": [{"type": "command", "command": "${CLAUDE_PROJECT_DIR}/.claude/skills/continuous-learning-v2/hooks/observe.sh pre"}]}
  ],
  "PostToolUse": [
    {"matcher": "Write|Edit|MultiEdit", "hooks": [{"type": "command", "command": "~/.claude/bin/cc-tools-validate"}]},
    {"matcher": "*", "hooks": [{"type": "command", "command": "${CLAUDE_PROJECT_DIR}/.claude/skills/continuous-learning-v2/hooks/observe.sh post"}]}
  ],
  "PreCompact": [
    {"matcher": "*", "hooks": [{"type": "command", "command": "cc-tools hook pre-compact"}]}
  ],
  "Stop": [
    {"matcher": "*", "hooks": [
      {"type": "command", "command": "~/.claude/hooks/sentinel"},
      {"type": "command", "command": "cc-tools hook stop"}
    ]}
  ],
  "Notification": [
    {"matcher": "*", "hooks": [
      {"type": "command", "command": "~/.claude/hooks/sentinel"},
      {"type": "command", "command": "cc-tools hook notification"}
    ]}
  ],
  "SessionEnd": [
    {"matcher": "*", "hooks": [{"type": "command", "command": "cc-tools hook session-end"}]}
  ]
}
```

**Preserved external hooks**: `sentinel`, `continuous-learning-v2/observe.sh`, `claude-docs-helper.sh` (in user settings).

### 2.6 Delete Migrated Scripts

After migration is verified working:

```
.claude/hooks/start-superpowers.sh      → delete
.claude/hooks/suggest-compact.js        → delete
.claude/hooks/pre-compact.js            → delete
.claude/hooks/session-start.js          → delete
.claude/hooks/session-end.js            → delete
.claude/hooks/evaluate-session.js       → delete
.claude/hooks/setup-package-manager.js  → delete
.claude/hooks/play_audio.py             → delete
.claude/hooks/macos_notification.py     → delete
.claude/hooks/pre-commit-reminder.sh    → delete
.claude/hooks/lib/                      → delete (shared JS utilities)
.claude/hooks/utils/                    → delete (Python utilities)
```

**Keep**: Nothing in `.claude/hooks/` — directory can be removed entirely.

---

## Phase 3: Quality Improvements

### 3.1 Split Config Manager

Split `internal/config/manager.go` (844 lines) into:

| File | Contents | ~Lines |
|---|---|---|
| `manager.go` | `Manager` struct, `NewManager()`, Get/Set/Reset/GetAll/GetConfig, load/save | ~300 |
| `keys.go` | Key constants, `keyDefinition` type, key registry map, type validation | ~300 |
| `values.go` | `Values` struct, section structs, backward-compatible parsing | ~250 |

Pure refactoring — no behavior changes.

### 3.2 Skip Directory Traversal Check

Add path validation in `cmd/cc-tools/skip.go`:

```go
cleanPath := filepath.Clean(absPath)
if strings.Contains(cleanPath, "..") {
    return fmt.Errorf("invalid path: directory traversal not allowed")
}
```

### 3.3 Discovery Error Visibility

Add debug logging in `internal/hooks/discovery.go` for each failed discovery attempt. Currently failures are silent; with debug enabled, each check logs what was tried and why it failed.

---

## Deferred

- **Command Discovery refactoring** (Cluster 1): Works correctly, refactoring is project-sized
- **Lock Manager PID reuse** (Cluster 5): Theoretical, not practical
- **Config key compile-time validation** (Open Question 4): Nice-to-have

## Verification

After all phases:
1. `task test` — all tests pass
2. `task lint` — zero lint issues
3. `task test-race` — no race conditions
4. `grep -r "quanta\|Quanta" .claude/hooks/` — no results (directory deleted)
5. Verify each hook event fires correctly by running `cc-tools hook <event>` with sample JSON input
6. Verify `.claude/settings.json` references only cc-tools and preserved external hooks
