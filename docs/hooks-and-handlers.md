# Hooks and Handlers

This document explains how cc-tools integrates with Claude Code's hook system, how events flow through the handler registry, and how the two execution paths --- `cc-tools hook` and `cc-tools validate` --- differ.

## Overview

Claude Code fires hook events at specific points during a session: when a session starts, before and after tool execution, when you submit a prompt, and so on. cc-tools processes these events through two commands:

- **`cc-tools hook`** --- A general-purpose event dispatcher. It reads event JSON from stdin, routes the event through a handler registry, and returns structured output.
- **`cc-tools validate`** --- A specialized PostToolUse handler. It discovers lint and test commands for the current project and runs them in parallel.

Both commands follow the Claude Code hooks protocol: they read JSON from stdin, perform work, and communicate results through exit codes, stdout JSON, and stderr text.

## Hook Events

Claude Code defines the following hook events (from `internal/hookcmd/events.go`). Each event fires at a specific point in the session lifecycle.

| Event | When It Fires |
|-------|---------------|
| `SessionStart` | A new Claude Code session begins |
| `SessionEnd` | A Claude Code session ends |
| `PreToolUse` | Before a tool is executed |
| `PostToolUse` | After a tool executes successfully |
| `PostToolUseFailure` | After a tool execution fails |
| `PreCompact` | Before context compaction |
| `Notification` | When Claude Code sends a notification |
| `UserPromptSubmit` | When you submit a prompt |
| `PermissionRequest` | When a permission request is made |
| `Stop` | When Claude Code stops generating |
| `SubagentStart` | When a subagent starts |
| `SubagentStop` | When a subagent stops |
| `TeammateIdle` | When a teammate goes idle |
| `TaskCompleted` | When a task is completed |

Not every event has a registered handler. Events without handlers return a zero-value response (exit code 0, no output) and have no effect.

## How `cc-tools hook` Works

The `cc-tools hook` command is the primary event dispatcher. Here is the sequence of operations when Claude Code invokes it:

1. Claude Code writes event JSON to stdin and invokes `cc-tools hook`.
2. cc-tools reads all stdin data and parses the JSON into a `HookInput` struct (`internal/hookcmd/input.go`). The struct includes common fields present on every event (`session_id`, `cwd`, `hook_event_name`) and event-specific fields (`tool_name`, `prompt`, `message`, etc.).
3. cc-tools loads configuration from disk via `config.NewManager()` and builds the default handler registry.
4. The registry looks up all handlers registered for that event name.
5. Each handler runs in sequence and returns a `Response` containing an exit code, optional stdout JSON, and optional stderr text.
6. Responses are merged: the highest exit code wins, the first non-nil stdout wins, and all stderr strings are concatenated.
7. cc-tools writes the merged response to stdout (JSON) and stderr (text), then exits with the merged exit code.

If stdin is empty or the JSON is malformed, `cc-tools hook` exits silently with code 0. Hooks must never block Claude Code due to input errors.

### The Response Protocol

The exit code determines how Claude Code reacts to the hook response:

| Exit Code | Meaning |
|-----------|---------|
| `0` | Success. Claude Code continues normally. |
| `2` | Block the action. Stderr text is shown to Claude as feedback. |

The stdout JSON (`HookOutput` in `internal/handler/handler.go`) can include these fields:

| Field | Purpose |
|-------|---------|
| `systemMessage` | Inject a system-level message into the conversation |
| `additionalContext` | Append context strings to the current turn |
| `suppressOutput` | Suppress the tool's output from the conversation |
| `hookSpecificOutput` | Arbitrary key-value data for hook-specific behavior |
| `permissionDecision` | Grant or deny a permission request |
| `updatedInput` | Modify the tool's input before execution |
| `continue` | Signal whether to continue processing |
| `stopReason` | Reason string when stopping generation |

## Handler Registry Architecture

The registry (`internal/handler/registry.go`) is a map from event names to ordered slices of handlers. Each handler implements the `Handler` interface: a `Name()` method for identification and a `Handle()` method that receives context and a `HookInput`, then returns a `Response`.

`NewDefaultRegistry()` in `internal/handler/defaults.go` wires all built-in handlers. The following sections describe each handler grouped by event.

### SessionStart Handlers

These run once at the beginning of every Claude Code session.

| Handler | What It Does |
|---------|--------------|
| **SuperpowersHandler** | Injects system context (skill discovery information) at session start |
| **PkgManagerHandler** | Detects the project's package manager (npm, yarn, pnpm, cargo, etc.) and injects context about available commands |
| **SessionContextHandler** | Stores session metadata (session ID, start time, working directory) for later retrieval |

### SessionEnd Handlers

These run when a Claude Code session terminates.

| Handler | What It Does |
|---------|--------------|
| **SessionEndHandler** | Persists session metadata to disk for post-session analysis |

### PreToolUse Handlers

These run before every tool execution. They can inject context, log events, or block tool calls.

| Handler | What It Does |
|---------|--------------|
| **SuggestCompactHandler** | Monitors tool call count and suggests context compaction when a threshold is reached. Configurable via `compact.threshold` and `compact.reminder_interval`. |
| **ObserveHandler** (pre phase) | Logs tool usage events to `~/.cache/cc-tools/observations/observations.jsonl` for the instinct learning system |
| **PreCommitReminderHandler** | Reminds you to run `task pre-commit` before git commit operations. Configurable via `pre_commit_reminder.enabled` and `pre_commit_reminder.command`. |

### PostToolUse Handlers

These run after a tool executes successfully.

| Handler | What It Does |
|---------|--------------|
| **ObserveHandler** (post phase) | Logs tool completion events to the observations file |

### PostToolUseFailure Handlers

These run after a tool execution fails.

| Handler | What It Does |
|---------|--------------|
| **ObserveHandler** (failure phase) | Logs tool failure events to the observations file |

### PreCompact Handlers

These run before Claude Code compacts the context window.

| Handler | What It Does |
|---------|--------------|
| **LogCompactionHandler** | Records compaction events for debugging |

### UserPromptSubmit Handlers

These run when you submit a prompt.

| Handler | What It Does |
|---------|--------------|
| **DriftHandler** | Tracks session intent from the first prompt, extracts keywords, and warns when subsequent prompts diverge significantly. Recognizes pivot phrases ("now let's", "switch to") to reset intent. Configurable via `drift.enabled`, `drift.min_edits`, `drift.threshold`. |

### Stop Handlers

These run when Claude Code stops generating.

| Handler | What It Does |
|---------|--------------|
| **StopReminderHandler** | Tracks response count per session and emits rotating reminders at configurable intervals. Configurable via `stop_reminder.enabled`, `stop_reminder.interval`, `stop_reminder.warn_at`. |

### Notification Handlers

These run when Claude Code sends a notification.

| Handler | What It Does |
|---------|--------------|
| **NotifyAudioHandler** | Plays a random MP3 from the audio directory using `afplay` (macOS). Respects quiet hours. |
| **NotifyDesktopHandler** | Sends macOS desktop notifications via `osascript`. Respects quiet hours. |
| **NotifyNtfyHandler** | Sends push notifications to an ntfy.sh topic. Respects quiet hours. |

## How `cc-tools validate` Differs

`cc-tools validate` is a standalone validation pipeline that does **not** use the handler registry. It exists as a separate command because its job --- discovering and running lint and test commands in parallel --- is fundamentally different from the dispatch-and-merge pattern of `cc-tools hook`.

Here is the validation sequence:

1. Reads PostToolUse event JSON from stdin.
2. Extracts the file path from the tool input and checks whether the file should be skipped (via the skip registry or file-type filters).
3. Finds the project root by walking up the directory tree.
4. Acquires a per-project lock with a configurable cooldown to avoid redundant back-to-back runs.
5. Discovers lint and test commands for the project by inspecting Taskfile, Makefile, package.json, and other build system files.
6. Runs lint and test commands in parallel with a configurable timeout.
7. Returns exit code 0 if both pass, or exit code 2 (block) with a descriptive error message if either fails.

Configuration is resolved with this precedence: environment variables > config file > command-line flags.

| Setting | Flag | Environment Variable |
|---------|------|---------------------|
| Timeout (seconds) | `--timeout`, `-t` | `CC_TOOLS_HOOKS_VALIDATE_TIMEOUT_SECONDS` |
| Cooldown (seconds) | `--cooldown`, `-c` | `CC_TOOLS_HOOKS_VALIDATE_COOLDOWN_SECONDS` |

## Configuring Hooks in Claude Code

Hooks are configured in `~/.claude/settings.json` under the `hooks` key. Each entry maps an event name to an array of hook definitions. A hook definition includes a `matcher` (glob pattern for tool names) and a `hooks` array with the commands to run.

A typical setup routes all events to `cc-tools hook` and uses `cc-tools validate` specifically for file-editing PostToolUse events:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "",
        "hooks": [
          { "type": "command", "command": "cc-tools hook" }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "",
        "hooks": [
          { "type": "command", "command": "cc-tools hook" }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit|NotebookEdit",
        "hooks": [
          { "type": "command", "command": "cc-tools validate" }
        ]
      },
      {
        "matcher": "",
        "hooks": [
          { "type": "command", "command": "cc-tools hook" }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "matcher": "",
        "hooks": [
          { "type": "command", "command": "cc-tools hook" }
        ]
      }
    ],
    "Stop": [
      {
        "matcher": "",
        "hooks": [
          { "type": "command", "command": "cc-tools hook" }
        ]
      }
    ],
    "Notification": [
      {
        "matcher": "",
        "hooks": [
          { "type": "command", "command": "cc-tools hook" }
        ]
      }
    ]
  }
}
```

The `matcher` field filters which tool names trigger the hook. An empty string matches all tools. Pipe-separated values (e.g., `"Write|Edit|MultiEdit|NotebookEdit"`) match any of the listed tool names.

## Data Flow

The following diagram shows how events flow through the two execution paths during a Claude Code session.

```
Claude Code Session
    |
    +-- SessionStart ----------> cc-tools hook --> Superpowers, PkgManager, SessionContext
    +-- PreToolUse ------------> cc-tools hook --> CompactSuggest, Observe, PreCommitReminder
    +-- PostToolUse (edit) ----> cc-tools validate --> Lint + Test (parallel)
    +-- PostToolUse (*) -------> cc-tools hook --> Observe
    +-- PostToolUseFailure ----> cc-tools hook --> Observe
    +-- UserPromptSubmit ------> cc-tools hook --> DriftDetection
    +-- Stop ------------------> cc-tools hook --> StopReminder
    +-- Notification ----------> cc-tools hook --> Audio, Desktop, Ntfy
    +-- PreCompact ------------> cc-tools hook --> LogCompaction
    +-- SessionEnd ------------> cc-tools hook --> SessionPersistence
```

Edit events (`Write`, `Edit`, `MultiEdit`, `NotebookEdit`) take the `cc-tools validate` path. All other PostToolUse events and every other event type route through `cc-tools hook` and the handler registry.

## Key Source Files

| File | Purpose |
|------|---------|
| `cmd/cc-tools/hook.go` | Entry point for the `hook` command |
| `cmd/cc-tools/validate.go` | Entry point for the `validate` command |
| `internal/hookcmd/events.go` | Hook event name constants |
| `internal/hookcmd/input.go` | `HookInput` struct and JSON parsing |
| `internal/handler/handler.go` | `Handler` interface, `Response`, and `HookOutput` types |
| `internal/handler/registry.go` | `Registry` type with `Register` and `Dispatch` methods |
| `internal/handler/defaults.go` | `NewDefaultRegistry()` wiring all built-in handlers |
| `internal/hooks/validate.go` | Parallel validation executor and orchestration |
| `internal/hooks/discovery.go` | Lint and test command discovery logic |
| `internal/hooks/executor.go` | Command execution with timeout support |
