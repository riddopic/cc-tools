# Dual CLI Support Design

## Problem

cc-tools only works with Claude Code. Gemini CLI supports a nearly identical hook protocol — JSON via stdin, JSON via stdout, exit code 0/2 semantics — but uses different event names, tool names, and configuration formats. Users who work with both CLIs cannot reuse their hook infrastructure.

## Decisions

- **Architecture:** Adapter pattern. A thin adapter layer normalizes CLI-specific input/output into a canonical internal format before reaching the handler registry.
- **Detection:** Auto-detect from environment variables. Gemini sets `GEMINI_SESSION_ID` and `GEMINI_PROJECT_DIR`; their presence indicates Gemini CLI. Absence means Claude Code.
- **Canonical names:** CLI-neutral. Neither Claude nor Gemini names — a neutral set that both adapters map to.
- **Event scope:** Shared events first (Phase 1). Gemini-specific events (BeforeAgent, AfterAgent, BeforeModel, AfterModel, BeforeToolSelection) deferred to Phase 2.
- **Binary rename:** `cc-tools` → `hookd`. Module path stays `github.com/riddopic/cc-tools`.
- **Config migration:** `~/.config/hookd/` primary, auto-migrate from `~/.config/cc-tools/` on first run.

## Adapter Interface

```go
type CLIType string

const (
    CLIClaude CLIType = "claude"
    CLIGemini CLIType = "gemini"
)

type Adapter interface {
    ParseInput(data []byte) (*hookcmd.HookInput, error)
    FormatOutput(resp *handler.Response) ([]byte, error)
    CLIType() CLIType
}

func Detect() CLIType {
    if os.Getenv("GEMINI_SESSION_ID") != "" || os.Getenv("GEMINI_PROJECT_DIR") != "" {
        return CLIGemini
    }
    return CLIClaude
}
```

The hook command selects an adapter, parses input, dispatches to handlers, and formats output:

```
stdin JSON → Adapter.ParseInput → HookInput → Handler Registry → Response → Adapter.FormatOutput → stdout JSON
```

## Event Normalization

Canonical event names replace the current Claude-specific constants:

| Canonical | Claude Code | Gemini CLI |
|-----------|-------------|------------|
| `SessionStart` | `SessionStart` | `SessionStart` |
| `SessionEnd` | `SessionEnd` | `SessionEnd` |
| `BeforeTool` | `PreToolUse` | `BeforeTool` |
| `AfterTool` | `PostToolUse` | `AfterTool` |
| `AfterToolFailure` | `PostToolUseFailure` | *(none)* |
| `BeforeCompress` | `PreCompact` | `PreCompress` |
| `Notification` | `Notification` | `Notification` |
| `UserPromptSubmit` | `UserPromptSubmit` | *(none)* |
| `Stop` | `Stop` | *(none)* |
| `SubagentStop` | `SubagentStop` | *(none)* |

Each adapter maps its CLI-specific event names to canonical ones during `ParseInput`.

## Tool Name Normalization

Tool names differ between CLIs. Adapters normalize to lowercase canonical names:

| Canonical | Claude Code | Gemini CLI |
|-----------|-------------|------------|
| `bash` | `Bash` | `shell` |
| `write_file` | `Write` | `write_file` |
| `edit_file` | `Edit` | `replace` |
| `read_file` | `Read` | `read_file` |
| `glob` | `Glob` | `glob` |
| `grep` | `Grep` | `grep` |

Helper methods like `IsEditTool()` check canonical names.

## Input Normalization

`HookInput` gains a `CLIType` field so handlers can branch on CLI when necessary:

```go
type HookInput struct {
    CLIType        CLIType
    SessionID      string
    TranscriptPath string  // empty for Gemini
    Cwd            string
    EventName      string  // canonical
    ToolName       string  // canonical
    // ... remaining fields unchanged
}
```

Gemini passes `SessionID` and `Cwd` via environment variables rather than JSON. The Gemini adapter reads these from `os.Getenv` and merges them into `HookInput`.

## Output Formatting

Both CLIs use the same exit code semantics (0 = proceed, 2 = block with stderr reason). The `HookOutput` struct stays unchanged. If Gemini requires different JSON field names, the Gemini adapter handles the mapping in `FormatOutput`.

## Setup Command

New `hookd setup` command generates per-CLI configuration:

```
hookd setup              # auto-detect or prompt interactively
hookd setup --cli claude # generate .claude/settings.json
hookd setup --cli gemini # generate .gemini/settings.json + commands
```

For Gemini, the setup command:

1. Generates `.gemini/settings.json` with correct event names and regex matchers (Gemini uses `.*` where Claude uses `*`)
2. Converts `.claude/commands/*.md` to `.gemini/commands/*.toml`

### Command Conversion

Claude Code and Gemini CLI use different command formats:

| Feature | Claude Code | Gemini CLI |
|---------|-------------|------------|
| Format | Markdown (`.md`) | TOML (`.toml`) |
| Metadata | YAML frontmatter | TOML fields |
| Arguments | `$ARGUMENTS` | `{{args}}` |
| Prompt | Markdown body | `prompt` field |

The setup command reads each `.claude/commands/*.md` file, extracts the prompt body and frontmatter metadata, and writes a `.gemini/commands/*.toml` file with the `prompt` field and `description` mapped from frontmatter.

## Binary Rename

| What | Before | After |
|------|--------|-------|
| Entry point | `cmd/cc-tools/` | `cmd/hookd/` |
| Binary | `bin/cc-tools` | `bin/hookd` |
| Config dir | `~/.config/cc-tools/` | `~/.config/hookd/` |
| Cache dir | `~/.cache/cc-tools/` | `~/.cache/hookd/` |
| Module path | `github.com/riddopic/cc-tools` | *(unchanged)* |

On first run, if `~/.config/cc-tools/` exists and `~/.config/hookd/` does not, auto-migrate the directory. Same for cache.

## Files Modified

| File | Change |
|------|--------|
| `cmd/cc-tools/` → `cmd/hookd/` | Rename entry point directory |
| `internal/adapter/` | New package: `Adapter` interface, `ClaudeAdapter`, `GeminiAdapter`, `Detect()` |
| `internal/hookcmd/events.go` | Rename constants to canonical names |
| `internal/hookcmd/input.go` | Add `CLIType` field to `HookInput` |
| `internal/handler/defaults.go` | Update event registrations to canonical names |
| `internal/handler/*.go` | Update event/tool name references |
| `cmd/hookd/hook.go` | Use adapter for parse/format |
| `cmd/hookd/setup.go` | New command: generate settings + commands |
| `internal/config/` | Update default paths, add migration |
| `Taskfile.yml` | Update binary name and build paths |
| `CLAUDE.md` | Update all references |
| `tests` | Add adapter tests, update event constant references |

## What Does Not Change

- Go module path
- Internal package structure (beyond new `internal/adapter`)
- Handler interface
- Core handler logic (notification, session, observation, etc.)
- Validate command architecture

## Phases

**Phase 1 (this design):** Adapter layer, shared events, binary rename, setup command, command conversion.

**Phase 2 (future):** Gemini-specific events — BeforeAgent, AfterAgent, BeforeModel, AfterModel, BeforeToolSelection — with new handlers.
