# cc-tools

A CLI companion for [Claude Code](https://docs.anthropic.com/en/docs/claude-code) that automates hook event handling, parallel lint/test validation, MCP server management, notifications, session tracking, and per-directory skip configuration.

## What it does

cc-tools plugs into Claude Code's hook system to run handlers on every hook event — session lifecycle, tool use, notifications, and context compaction. When configured as a `PostToolUse` hook, the `validate` command intercepts file edits and runs your project's linter and test suite in parallel before accepting the change. If either fails, the tool call is blocked with a formatted error message.

Beyond validation, cc-tools provides:

- **Hook dispatch** — a handler registry that routes all Claude Code hook events to purpose-built handlers
- **Notifications** — audio playback, macOS desktop alerts, and ntfy push notifications with quiet hours
- **Session tracking** — stores session metadata, supports aliases and search
- **Observation logging** — records tool use events for analysis
- **MCP management** — enable/disable MCP server integrations
- **Skip registry** — per-directory skip rules for lint, test, or both

## Install

Requires Go 1.26+ and [Task](https://taskfile.dev).

```bash
task build      # Build to ./bin/cc-tools
task install    # Copy to $GOPATH/bin
```

## Commands

```bash
cc-tools <command> [arguments]
```

| Command | Description |
|---------|-------------|
| `hook` | Dispatch Claude Code hook events to registered handlers (reads JSON from stdin) |
| `validate` | Run lint and test in parallel (reads JSON from stdin) |
| `session` | List, search, and manage session metadata and aliases |
| `config` | Get, set, list, and reset application settings |
| `skip` | Configure directories to skip validation (lint, test, or all) |
| `unskip` | Remove skip settings from directories |
| `mcp` | Manage Claude MCP servers (list, enable, disable) |
| `debug` | Configure debug logging (enable, disable, status, list, filename) |

### Hook integration

Add to your Claude Code settings (`~/.claude/settings.json`):

```json
{
  "hooks": {
    "PreToolUse": [
      { "matcher": "*", "command": "cc-tools hook" }
    ],
    "PostToolUse": [
      { "matcher": "Write|Edit|MultiEdit", "command": "cc-tools validate" },
      { "matcher": "*", "command": "cc-tools hook" }
    ],
    "SessionStart": [
      { "matcher": "*", "command": "cc-tools hook" }
    ],
    "SessionEnd": [
      { "matcher": "*", "command": "cc-tools hook" }
    ],
    "Notification": [
      { "matcher": "*", "command": "cc-tools hook" }
    ],
    "PreCompact": [
      { "matcher": "*", "command": "cc-tools hook" }
    ]
  }
}
```

The `hook` command reads event JSON from stdin, dispatches to registered handlers, and returns structured output. The `validate` command discovers lint/test commands for the project and runs them concurrently.

### Registered handlers

| Event | Handlers |
|-------|----------|
| SessionStart | Superpowers injection, package manager detection, session context |
| SessionEnd | Session metadata persistence |
| PreToolUse | Compact suggestion, observation logging, pre-commit reminder |
| PostToolUse | Observation logging |
| PreCompact | Log compaction |
| Notification | Audio playback, desktop alert, ntfy push |

### Examples

```bash
# Dispatch a hook event (stdin receives event JSON)
echo '{"hook_event_name":"Notification","title":"Done","message":"Tests passed"}' | cc-tools hook

# Validate a file edit
echo '{"tool_name":"Edit","tool_input":{"file_path":"main.go"}}' | cc-tools validate

# Manage sessions
cc-tools session list
cc-tools session info <session-id>
cc-tools session search "auth refactor"
cc-tools session alias set latest <session-id>

# Manage MCP servers
cc-tools mcp list
cc-tools mcp enable jira
cc-tools mcp disable jira

# Skip validation for a directory
cc-tools skip lint /path/to/generated
cc-tools skip list

# Configure settings
cc-tools config set validate.timeout 120
cc-tools config get validate.timeout
cc-tools config list
```

## Configuration

Settings are stored at `~/.config/cc-tools/config.json`.

| Key | Default | Description |
|-----|---------|-------------|
| `validate.timeout` | 60 | Validation timeout in seconds |
| `validate.cooldown` | 5 | Cooldown between validation runs |
| `notifications.ntfy_topic` | `""` | ntfy.sh topic for push notifications |
| `notify.audio.enabled` | false | Enable audio notification sounds |
| `notify.audio.directory` | `""` | Path to directory of MP3 files |
| `notify.desktop.enabled` | false | Enable macOS desktop notifications |
| `notify.quiet_hours.enabled` | false | Suppress notifications during quiet hours |
| `notify.quiet_hours.start` | `""` | Quiet hours start time (HH:MM) |
| `notify.quiet_hours.end` | `""` | Quiet hours end time (HH:MM) |
| `compact.threshold` | 0 | Token threshold for compact suggestions |
| `observe.enabled` | false | Enable tool use observation logging |
| `pre_commit_reminder.enabled` | false | Remind to run checks before git commit |

Debug logs are written to `~/.cache/cc-tools/debug/`.

## Development

```bash
task doctor         # Check required tools
task tools-install  # Install gotestsum, golangci-lint, goimports, mockery
task build          # Build binary
task test           # Fast tests (-short, 30s)
task lint           # golangci-lint
task check          # fmt + lint + test-race (run before committing)
task test-race      # Tests with race detector
task mocks          # Regenerate mocks
task coverage       # HTML coverage report
```

## License

See [LICENSE](LICENSE) for details.
