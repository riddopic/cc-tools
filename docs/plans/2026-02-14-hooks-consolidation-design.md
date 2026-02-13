# Hooks Consolidation into cc-tools

**Date:** 2026-02-14
**Status:** Approved
**Approach:** Flat subcommands under `cc-tools hook <event-type>`

## Summary

Consolidate all Claude Code hooks from `.claude/hooks/` (JS/Python/Bash scripts) and the CL-v2 observe hook into the cc-tools Go binary. One entry point per hook event type replaces 12 separate scripts across 3 languages. Session management is redesigned from regex-parsed markdown to structured JSON. All configuration moves to `~/.config/cc-tools/config.json`.

## Scope

### Hooks being absorbed

| Hook | Current Language | Hook Event | Handler Name |
|------|-----------------|------------|--------------|
| `session-start.js` | Node.js | SessionStart | load-session |
| `start-superpowers.sh` | Bash | SessionStart | inject-superpowers |
| `setup-package-manager.js` | Node.js | SessionStart | detect-package-manager |
| `session-end.js` | Node.js | SessionEnd | save-session |
| `evaluate-session.js` | Node.js | SessionEnd | evaluate-learning |
| `suggest-compact.js` | Node.js | PreToolUse | suggest-compact |
| `pre-commit-reminder.sh` | Bash | PreToolUse | pre-commit-reminder |
| `observe.sh` (CL-v2) | Bash+Python | PreToolUse/PostToolUse | observe |
| `pre-compact.js` | Node.js | PreCompact | log-compaction |
| `play_audio.py` | Python | Stop/Notification | play-audio |
| `macos_notification.py` | Python | Stop/Notification | notify-desktop |
| `evaluate-session.sh` (CL-v1) | Bash | Stop | evaluate-learning |

### Shared libraries being absorbed

- `lib/utils.js` — file ops, git helpers, date/time formatting
- `lib/session-manager.js` — session CRUD operations
- `lib/session-aliases.js` — alias management
- `lib/package-manager.js` — package manager detection

### Out of scope

- `~/.claude/hooks/sentinel` — external hook, stays as separate settings.json entry
- `utils/generate_audio_clips.py` — one-time generation tool using OpenAI TTS API, stays as Python script

## Command Structure

### New `hook` command

```
cc-tools hook <event-type> [flags]
```

Event types map 1:1 to Claude Code hook events:

| CLI Subcommand | Claude Code Hook Event | Handlers |
|---|---|---|
| `session-start` | SessionStart | load-session, inject-superpowers, detect-package-manager |
| `session-end` | SessionEnd | save-session, evaluate-learning |
| `pre-tool-use` | PreToolUse | suggest-compact, pre-commit-reminder, observe |
| `post-tool-use` | PostToolUse | observe |
| `post-tool-use-failure` | PostToolUseFailure | observe |
| `pre-compact` | PreCompact | log-compaction |
| `stop` | Stop | play-audio, notify-desktop, evaluate-learning |
| `notification` | Notification | play-audio, notify-desktop |

Unknown event types (e.g., `user-prompt-submit`, `permission-request`, `subagent-start`) are accepted gracefully — log to debug and exit 0.

### New `session` command

```
cc-tools session list [--limit N] [--date YYYY-MM-DD] [--search PATTERN]
cc-tools session load <id|alias>
cc-tools session info <id|alias>
cc-tools session alias <id> <name>
cc-tools session alias --remove <name>
cc-tools session aliases
```

### settings.json configuration

Replaces all individual hook entries with one entry per event type:

```json
{
  "hooks": {
    "SessionStart": [{ "matcher": "*", "hooks": [{ "type": "command", "command": "cc-tools hook session-start" }] }],
    "PreToolUse": [{ "matcher": "*", "hooks": [{ "type": "command", "command": "cc-tools hook pre-tool-use" }] }],
    "PostToolUse": [{ "matcher": "*", "hooks": [{ "type": "command", "command": "cc-tools hook post-tool-use" }] }],
    "PostToolUseFailure": [{ "matcher": "*", "hooks": [{ "type": "command", "command": "cc-tools hook post-tool-use-failure" }] }],
    "PreCompact": [{ "matcher": "*", "hooks": [{ "type": "command", "command": "cc-tools hook pre-compact" }] }],
    "Stop": [{ "matcher": "*", "hooks": [{ "type": "command", "command": "cc-tools hook stop" }] }],
    "Notification": [{ "matcher": "*", "hooks": [{ "type": "command", "command": "cc-tools hook notification" }] }],
    "SessionEnd": [{ "matcher": "*", "hooks": [{ "type": "command", "command": "cc-tools hook session-end" }] }]
  }
}
```

The existing `validate` command stays as a separate PostToolUse entry with its own matcher (`Write|Edit|MultiEdit`).

## Package Structure

```
internal/
├── hookcmd/              # Hook command dispatcher + handler registry
│   ├── hookcmd.go        # Top-level dispatch: parses subcommand, runs handlers
│   ├── input.go          # Stdin JSON parsing (shared across all handlers)
│   └── handler.go        # Handler interface + sequential runner
├── session/              # Session persistence (redesigned)
│   ├── store.go          # Session CRUD — JSON-based storage
│   ├── alias.go          # Alias management
│   └── transcript.go     # Transcript parsing (extract summary from JSONL)
├── observe/              # CL-v2 observation recording
│   ├── observe.go        # Write tool events to JSONL observations file
│   └── archive.go        # File rotation when size exceeds threshold
├── notify/               # Desktop + audio notifications
│   ├── desktop.go        # macOS notification via osascript (os/exec)
│   ├── audio.go          # Audio playback via Go library (gopxl/beep)
│   └── quiethours.go     # Configurable quiet hours check
├── compact/              # Compact suggestion logic
│   ├── suggest.go        # Tool call counter + threshold suggestions
│   └── log.go            # Pre-compact state logging
├── superpowers/          # Skill injection for SessionStart
│   └── inject.go         # Reads skill file, outputs hookSpecificOutput JSON
└── pkgmanager/           # Package manager detection
    └── detect.go         # Lock file / package.json / config detection
```

## Stdin JSON Contract

Based on the official Claude Code hooks reference (https://code.claude.com/docs/en/hooks).

### Common fields (all events)

```go
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

    // Notification specific
    Message          string `json:"message,omitempty"`
    Title            string `json:"title,omitempty"`
    NotificationType string `json:"notification_type,omitempty"`

    // PreCompact specific
    Trigger            string `json:"trigger,omitempty"`
    CustomInstructions string `json:"custom_instructions,omitempty"`
}
```

The parser is lenient — unknown fields are ignored, missing fields default to zero values.

## Handler Interface

```go
type Handler interface {
    Name() string
    Run(ctx context.Context, input *HookInput, out io.Writer, errOut io.Writer) error
}
```

Handlers are run sequentially. Errors are logged to stderr but do not cause non-zero exit. The dispatcher catches panics from individual handlers.

### Error handling contract

- Handler errors: logged to stderr, continue to next handler, exit 0
- Panic recovery: caught by dispatcher, logged, continue
- Only exception: if a handler needs to output JSON to stdout (e.g., inject-superpowers), failure means degraded behavior but not a crash

### Stop hook infinite loop prevention

The Stop handler MUST check `stop_hook_active` before doing anything:

```go
func (h *StopHandler) Run(ctx context.Context, input *HookInput, out, errOut io.Writer) error {
    if input.StopHookActive {
        return nil // prevent infinite loop
    }
    // ... normal handler logic
}
```

## stdout/stderr Contract

| Stream | Purpose | When processed |
|--------|---------|----------------|
| stdout | JSON output returned to Claude Code | Only on exit 0 |
| stderr | Log messages visible in Claude Code UI (verbose mode) or fed back to Claude on exit 2 | Always |

### JSON output format

SessionStart context injection:

```json
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "Context string here"
  }
}
```

Plain text stdout also works for SessionStart and UserPromptSubmit — it's added as context for Claude.

## Session Redesign

### Storage format

`.claude/sessions/{date}-{shortid}.json`

```json
{
  "version": "2.0",
  "id": "a1b2c3d4",
  "date": "2026-02-14",
  "started": "2026-02-14T09:15:00Z",
  "lastUpdated": "2026-02-14T11:30:00Z",
  "title": "Session: 2026-02-14",
  "summary": {
    "tasks": ["Implement hooks migration", "Fix lint errors"],
    "filesModified": ["cmd/cc-tools/main.go", "internal/hookcmd/hookcmd.go"],
    "toolsUsed": ["Edit", "Bash", "Read"],
    "totalMessages": 42
  },
  "compactions": [
    {"timestamp": "2026-02-14T10:00:00Z"}
  ],
  "notes": ""
}
```

### Aliases

Stay as `.claude/session-aliases.json` (same format). Code moves to `internal/session/alias.go`.

### Migration

On `session-start`, if cc-tools finds `.tmp` markdown files without a corresponding `.json` file, it reads and converts them. Old files are left in place (no destructive migration).

### /sessions command update

`.claude/commands/sessions.md` simplifies from inline Node.js scripts to `cc-tools session` calls.

## Configuration

All configurable values consolidated in `~/.config/cc-tools/config.json`:

```json
{
  "hooks": {
    "validate": {
      "timeoutSeconds": 60,
      "cooldownSeconds": 5
    },
    "notify": {
      "quietHours": {
        "enabled": true,
        "start": "21:00",
        "end": "07:30"
      },
      "audio": {
        "enabled": true,
        "directory": "~/.claude/audio"
      },
      "desktop": {
        "enabled": true
      }
    },
    "compact": {
      "threshold": 50,
      "reminderInterval": 25
    },
    "observe": {
      "enabled": true,
      "maxFileSizeMB": 10
    },
    "learning": {
      "minSessionLength": 10,
      "learnedSkillsPath": ".claude/skills/learned"
    },
    "preCommitReminder": {
      "enabled": true,
      "command": "task pre-commit"
    }
  }
}
```

Existing `validate` config stays where it is. Env var overrides (`CC_TOOLS_HOOKS_VALIDATE_*`) continue to work for validate only.

## Notification System

### Desktop notifications

`internal/notify/desktop.go` uses `os/exec` to call `osascript` on macOS. Notification message is contextual based on the hook event.

### Audio playback

`internal/notify/audio.go` uses `gopxl/beep` for MP3 decoding and playback. Audio files loaded from configurable directory (default `~/.claude/audio/`).

### Quiet hours

Configurable via `hooks.notify.quietHours` in config. Both audio and desktop notifications respect quiet hours. Default: 9 PM to 7:30 AM.

## Environment Variables

### Used by cc-tools

| Variable | Used By | Purpose |
|----------|---------|---------|
| `CLAUDE_PROJECT_DIR` | inject-superpowers | Resolve skill file path |
| `CLAUDE_ENV_FILE` | detect-package-manager | Persist `PREFERRED_PACKAGE_MANAGER` for session |
| `CLAUDE_HOOKS_DEBUG` | all handlers | Enable debug logging |

### No longer needed (replaced by config)

| Variable | Replaced By |
|----------|-------------|
| `COMPACT_THRESHOLD` | `hooks.compact.threshold` |

## Handler Details

### session-start handlers

1. **load-session**: Find most recent session JSON, output summary to stdout as context
2. **inject-superpowers**: Read `using-superpowers` skill file from `$CLAUDE_PROJECT_DIR/.claude/skills/using-superpowers`, output as `hookSpecificOutput.additionalContext` JSON
3. **detect-package-manager**: Detection priority: env var > project config > package.json > lock file > global config > default npm. Write to `$CLAUDE_ENV_FILE` if available. Log detection to stderr.

### session-end handlers

1. **save-session**: Read transcript from `transcript_path`, extract summary (user messages, tools used, files modified), write/update session JSON file
2. **evaluate-learning**: Count user messages in transcript. If above `learning.minSessionLength` threshold, log evaluation signal to stderr

### pre-tool-use handlers

1. **suggest-compact**: Track tool call count in temp file keyed by `session_id`. At `compact.threshold` and every `compact.reminderInterval` after, log `/compact` suggestion to stderr
2. **pre-commit-reminder**: Check if `tool_input.command` contains `git commit`. If so, log reminder to run `preCommitReminder.command` to stderr
3. **observe**: Parse tool event, write JSONL to `.claude/homunculus/observations.jsonl`. Check for disabled file, handle file rotation

### post-tool-use / post-tool-use-failure handlers

1. **observe**: Same as pre-tool-use observe but records completion/failure events

### pre-compact handlers

1. **log-compaction**: Append timestamped event to compaction log. Update active session JSON `compactions` array

### stop handlers

1. **Check `stop_hook_active`** — if true, exit immediately (infinite loop prevention)
2. **play-audio**: Check quiet hours, play MP3 via beep library
3. **notify-desktop**: Check quiet hours, send macOS notification via osascript
4. **evaluate-learning**: Same as session-end evaluate-learning handler

### notification handlers

1. **play-audio**: Same as stop, check quiet hours first
2. **notify-desktop**: Same as stop, check quiet hours first

## Testing Strategy

- Each handler gets its own `_test.go` file with table-driven tests
- Use dependency injection for filesystem, clock, and audio interfaces
- Mock `os/exec` calls for osascript and external commands
- Test quiet hours with injected clock
- Test observe file rotation with injected filesystem
- Test session migration from markdown to JSON format
- Integration test: pipe real JSON stdin through `cc-tools hook session-start` and verify stdout/stderr

## Dependencies

New Go dependencies:

- `gopxl/beep` — MP3 decoding and audio playback (replaces Python `afplay` call)
- No other new external dependencies expected

## Migration Checklist

After implementation:

1. Update `~/.claude/settings.json` to point all hooks at `cc-tools hook <event>`
2. Update `.claude/commands/sessions.md` to use `cc-tools session` commands
3. Verify all hooks work via `claude --debug`
4. Archive `.claude/hooks/` scripts (keep for reference, no longer invoked)
5. Remove `lib/` Node.js libraries
