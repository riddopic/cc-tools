# cc-tools

A CLI companion for [Claude Code](https://docs.anthropic.com/en/docs/claude-code) that automates hook event handling, parallel lint/test validation, MCP server management, notifications, session tracking, per-directory skip configuration, and instinct-based learning.

## What it does

cc-tools plugs into Claude Code's hook system to run handlers on every hook event — session lifecycle, tool use, notifications, and context compaction. When configured as a `PostToolUse` hook, the `validate` command intercepts file edits and runs your project's linter and test suite in parallel before accepting the change. If either fails, the tool call is blocked with a formatted error message.

Beyond validation, cc-tools provides:

- **Hook dispatch** — a handler registry that routes all Claude Code hook events to purpose-built handlers
- **Notifications** — audio playback, macOS desktop alerts, and ntfy push notifications with quiet hours
- **Session tracking** — stores session metadata, supports aliases and search
- **Observation logging** — records tool use events for analysis
- **Instinct learning** — observes tool usage, builds instincts with confidence scoring, evolves them into skills
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
| `instinct` | Manage learned instincts (status, export, import, evolve) |

### Hook integration

Add to your Claude Code settings (`~/.claude/settings.json`):

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [{ "type": "command", "command": "cc-tools hook" }]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit",
        "hooks": [{ "type": "command", "command": "cc-tools validate" }]
      },
      {
        "matcher": "*",
        "hooks": [{ "type": "command", "command": "cc-tools hook" }]
      }
    ],
    "SessionStart": [
      {
        "matcher": "*",
        "hooks": [{ "type": "command", "command": "cc-tools hook" }]
      }
    ],
    "SessionEnd": [
      {
        "matcher": "*",
        "hooks": [{ "type": "command", "command": "cc-tools hook" }]
      }
    ],
    "UserPromptSubmit": [
      {
        "matcher": "*",
        "hooks": [{ "type": "command", "command": "cc-tools hook" }]
      }
    ],
    "Stop": [
      {
        "matcher": "*",
        "hooks": [{ "type": "command", "command": "cc-tools hook" }]
      }
    ],
    "Notification": [
      {
        "matcher": "*",
        "hooks": [{ "type": "command", "command": "cc-tools hook" }]
      }
    ],
    "PreCompact": [
      {
        "matcher": "*",
        "hooks": [{ "type": "command", "command": "cc-tools hook" }]
      }
    ]
  }
}
```

For a guided walkthrough, see [Getting Started](docs/getting-started.md). For details on every handler, see [Hooks and Handlers](docs/hooks-and-handlers.md).

### Examples

```bash
# Manage sessions
cc-tools session list
cc-tools session info <session-id>
cc-tools session search "auth refactor"
cc-tools session alias set latest <session-id>

# Configure settings
cc-tools config set validate.timeout 120
cc-tools config list

# Skip validation in the current directory
cc-tools skip lint
cc-tools skip list

# Manage MCP servers
cc-tools mcp list
cc-tools mcp enable jira

# View learned instincts
cc-tools instinct status
cc-tools instinct export --format yaml
```

See [CLI Reference](docs/cli-reference.md) for the complete command reference.

## Configuration

Settings are stored at `~/.config/cc-tools/config.json`. Manage them with `cc-tools config`:

```bash
cc-tools config list                          # View all settings with defaults
cc-tools config set validate.timeout 120      # Change a value
cc-tools config reset validate.timeout        # Reset to default
```

Configuration covers 11 groups with 31 keys total:

| Group | Keys | Controls |
|-------|------|----------|
| `validate.*` | 2 | Timeout and cooldown for lint/test runs |
| `notifications.*` | 1 | ntfy.sh push notification topic |
| `compact.*` | 2 | Context compaction suggestion thresholds |
| `notify.*` | 6 | Quiet hours, audio, and desktop notifications |
| `observe.*` | 2 | Tool use observation logging |
| `learning.*` | 2 | Skill extraction from sessions |
| `pre_commit_reminder.*` | 2 | Pre-commit check reminders |
| `package_manager.*` | 1 | Preferred package manager override |
| `drift.*` | 3 | Session topic drift detection |
| `stop_reminder.*` | 3 | Periodic session length reminders |
| `instinct.*` | 7 | Instinct storage, confidence, and evolution |

See [Configuration Reference](docs/configuration.md) for all keys, types, defaults, and examples.

## Documentation

| Document | Type | Description |
|----------|------|-------------|
| [Getting Started](docs/getting-started.md) | Tutorial | Install, configure, and verify cc-tools from scratch |
| [CLI Reference](docs/cli-reference.md) | Reference | Every command, flag, and environment variable |
| [Configuration](docs/configuration.md) | Reference | All 31 configuration keys with types and defaults |
| [Hooks and Handlers](docs/hooks-and-handlers.md) | Explanation | How the hook system dispatches events to handlers |
| [Instincts](docs/instincts.md) | Explanation | The learning system lifecycle — observation to evolution |
| [Skills and Commands](docs/skills-and-commands.md) | Reference | All skills and slash commands with trigger contexts |
| [Troubleshooting](docs/troubleshooting.md) | How-to | Common issues and solutions |

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
