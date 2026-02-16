# Hook Parity Testing Design

## Problem

cc-tools reimplements 10 Claude Code hooks (originally JS/shell scripts) as Go handlers. Before switching `settings.json` to use the new handlers, we need tests that verify each handler produces the same behavior as the original script.

## Scope

Test all 10 migrated handlers for full I/O contract parity:

| Old Script | New Handler | Event |
|---|---|---|
| `suggest-compact.js` | SuggestCompactHandler | PreToolUse |
| `pre-commit-reminder.sh` | PreCommitReminderHandler | PreToolUse |
| `observe.sh` | ObserveHandler (x3) | Pre/Post/PostFailure ToolUse |
| `pre-compact.js` | LogCompactionHandler | PreCompact |
| `start-superpowers.sh` | SuperpowersHandler | SessionStart |
| `session-start.js` (pkg) | PkgManagerHandler | SessionStart |
| `session-start.js` (ctx) | SessionContextHandler | SessionStart |
| `session-end.js` + `evaluate-session.js` | SessionEndHandler | SessionEnd |
| `play_audio.py` | NotifyAudioHandler | Notification |
| `macos_notification.py` | NotifyDesktopHandler | Notification |

## Approach

Per-handler parity tests in `internal/handler/`. Each test:

1. Constructs a `HookInput` matching what Claude Code sends for that event
2. Calls `Handle()` directly on the handler
3. Asserts the full I/O contract:
   - **Exit code** (0 = success, 2 = blocking)
   - **Stdout JSON** (HookOutput structure, hookSpecificOutput, additionalContext)
   - **Stderr** (user-visible messages)
   - **File side effects** (counter files, JSONL observations, .env files, session files, compaction logs)

Tests use existing patterns: `newTestConfig()`, `t.TempDir()`, functional options for dependency injection, table-driven with testify.

## What "Parity" Means

The new handlers differ intentionally from the old scripts: configurable thresholds (not env vars), Go-native file operations (not python3/jq), different message prefixes (`[cc-tools]` vs `[StrategicCompact]`). Parity means **functional equivalence** — given the same scenario, the handler produces correct behavior — not byte-identical output.

## Test Cases Per Handler

### 1. SuggestCompactHandler (10 cases)

4 existing + 6 new:
- Counter file creation, increment, and content format
- No suggestion below threshold
- Suggestion at exact threshold
- Periodic suggestion at reminder intervals
- No suggestion between intervals
- Separate counters per session
- Threshold=0 never suggests
- Nil config graceful no-op

### 2. PreCommitReminderHandler (12 cases)

7 existing + 5 new:
- `git commit -am` variant detected
- `git push`, `ls -la` produce no reminder
- Custom command from config used in message
- `git commit` embedded in chained command detected
- Exit code 0 across all cases (non-blocking)

### 3. ObserveHandler (13 cases)

4 existing + 9 new:
- Post and failure modes record correct phase
- JSONL format validation (each line parses as JSON)
- Event fields: tool_name, session_id, timestamp, tool_input
- Multiple events appended, not overwritten
- Disabled via `.disabled` marker file
- File rotation when over max size
- Empty tool input handled gracefully
- Missing dir created automatically

### 4. LogCompactionHandler (5 cases)

0 existing + 5 new:
- Creates compaction-log.txt with timestamped entry
- Multiple calls append entries
- Log dir created if missing
- Entry format matches `[YYYY-MM-DD HH:MM:SS] compaction triggered`
- Exit code 0

### 5. SuperpowersHandler (5 cases)

3 existing + 2 new:
- Multiple skill files all included
- hookSpecificOutput structure matches protocol

### 6. PkgManagerHandler (7 cases)

4 existing + 3 new:
- package-lock.json detects npm
- pnpm-lock.yaml detects pnpm
- Re-run overwrites existing .env file

### 7. SessionContextHandler (6 cases)

5 existing + 1 new:
- Multiple sessions uses most recent

### 8. SessionEndHandler (9 cases)

0 existing + 9 new:
- Session saved with correct metadata (date, title)
- No transcript handled gracefully
- Transcript extracts tools/files/message count
- Long session emits learning signal in stderr
- Short session produces no learning signal
- Custom minLength from config
- Default minLength=10 with nil config
- Exit code 0

### 9. NotifyAudioHandler (6 cases)

0 existing + 6 new:
- Disabled config produces no play
- No player injected produces no error
- Enabled with mock player calls Play
- Quiet hours active skips play
- Quiet hours inactive plays
- Nil config no-op

### 10. NotifyDesktopHandler (8 cases)

0 existing + 8 new:
- Disabled config produces no notification
- No runner produces no error
- Enabled with mock runner calls Run
- Custom title/message from HookInput used
- Default title/message when input empty
- Quiet hours active skips notification
- Nil config no-op

## File Changes

| File | Action |
|---|---|
| `internal/handler/tooluse_test.go` | Add parity cases for SuggestCompact, Observe, PreCommitReminder |
| `internal/handler/compact_test.go` | Add parity cases for LogCompaction |
| `internal/handler/session_start_test.go` | Add parity cases for Superpowers, PkgManager, SessionContext |
| `internal/handler/session_end_test.go` | Add parity cases for SessionEnd |
| `internal/handler/notification_test.go` | Add parity cases for NotifyAudio, NotifyDesktop |

## Totals

- **27 existing tests** provide baseline coverage
- **54 new parity tests** verify behavioral equivalence
- **81 total test cases** across 10 handlers
