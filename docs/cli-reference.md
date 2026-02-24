# CLI Reference

Complete reference for every `cc-tools` command, subcommand, flag, and environment variable.

`cc-tools` is a CLI companion for Claude Code that handles hook event dispatching, parallel lint/test validation, MCP server management, session tracking, debug logging, per-directory skip configuration, and learned instinct management.

## Global Options

| Flag | Description |
| --- | --- |
| `--version` | Print the version and exit |
| `--help`, `-h` | Show help for any command |

## hook

Handle Claude Code hook events. This is a hidden command -- Claude Code invokes it automatically through hook configuration. You do not call it directly.

### Synopsis

```
cc-tools hook
```

### Description

Reads hook event JSON from stdin, dispatches the event to the registered handler registry, and writes structured JSON output to stdout. If a handler blocks the event, the command writes feedback to stderr and exits with code 2.

### Input

Pipe a JSON hook event on stdin. The JSON structure depends on the hook type (`PreToolUse`, `PostToolUse`, `UserPromptSubmit`, `Stop`).

### Exit Codes

| Code | Meaning |
| --- | --- |
| `0` | Success -- event processed without blocking |
| `2` | Block -- handler rejected the event; feedback written to stderr |

### Example

```bash
echo '{"hook_type":"PreToolUse","tool_name":"Bash","tool_input":{"command":"npm install"}}' | cc-tools hook
```

---

## validate

Discover and run lint and test commands in parallel. Designed as a `PostToolUse` hook for Claude Code to validate changes after edits.

### Synopsis

```
cc-tools validate [flags]
```

### Flags

| Flag | Short | Default | Description |
| --- | --- | --- | --- |
| `--timeout` | `-t` | `60` | Timeout in seconds for the validation run |
| `--cooldown` | `-c` | `5` | Cooldown in seconds between consecutive runs |

### Environment Variables

| Variable | Description |
| --- | --- |
| `CC_TOOLS_HOOKS_VALIDATE_TIMEOUT_SECONDS` | Override the timeout value |
| `CC_TOOLS_HOOKS_VALIDATE_COOLDOWN_SECONDS` | Override the cooldown value |

### Configuration Precedence

Values resolve in this order (highest wins):

1. Environment variables (`CC_TOOLS_HOOKS_VALIDATE_*`)
2. Config file (`~/.config/cc-tools/config.json`)
3. Flag defaults

### Exit Codes

| Code | Meaning |
| --- | --- |
| `0` | All lint and test checks passed |
| non-zero | One or more checks failed; blocks the tool call |

### Examples

```bash
# Run with default settings, piping hook event JSON
echo '{"tool_input":{"file_path":"main.go"}}' | cc-tools validate

# Run with a longer timeout
echo '{"tool_input":{"file_path":"main.go"}}' | cc-tools validate --timeout 120

# Override timeout via environment variable
CC_TOOLS_HOOKS_VALIDATE_TIMEOUT_SECONDS=180 cc-tools validate
```

---

## session

Manage Claude Code sessions. Browse recent sessions, look up details, search by keyword, and create aliases for quick access.

### Synopsis

```
cc-tools session <subcommand>
```

### Subcommands

#### session list

List recent sessions in a tabular format.

```
cc-tools session list [--limit N]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--limit` | `10` | Maximum number of sessions to display |

```bash
cc-tools session list
cc-tools session list --limit 20
```

#### session info

Show detailed information about a session. Accepts a session ID or a previously defined alias.

```
cc-tools session info <id-or-alias>
```

Output is formatted as indented JSON.

```bash
cc-tools session info abc123
cc-tools session info mywork
```

#### session search

Search sessions by keyword. Matches against session titles and content.

```
cc-tools session search <query>
```

```bash
cc-tools session search refactor
cc-tools session search "config validation"
```

#### session alias set

Create or overwrite a named alias that maps to a session ID.

```
cc-tools session alias set <name> <session-id>
```

```bash
cc-tools session alias set mywork abc123
```

#### session alias remove

Delete a named session alias.

```
cc-tools session alias remove <name>
```

```bash
cc-tools session alias remove mywork
```

#### session alias list

List all defined session aliases.

```
cc-tools session alias list
```

---

## config

Read and write cc-tools configuration. Settings persist in `~/.config/cc-tools/config.json`.

### Synopsis

```
cc-tools config <subcommand>
```

### Subcommands

#### config get

Retrieve the current value of a configuration key. If the key does not exist, the command prints available keys and exits with an error.

```
cc-tools config get <key>
```

```bash
cc-tools config get validate.timeout
```

#### config set

Set a configuration key to a new value.

```
cc-tools config set <key> <value>
```

```bash
cc-tools config set validate.timeout 90
cc-tools config set drift.enabled false
```

#### config list

Display all configuration settings in a table showing key, current value, and whether the value is a default or custom override. Also aliased as `show`.

```
cc-tools config list
cc-tools config show
```

#### config reset

Reset configuration to defaults. Pass a key to reset a single setting, or omit the key to reset everything.

```
cc-tools config reset [key]
```

```bash
# Reset a single key
cc-tools config reset validate.timeout

# Reset all configuration to defaults
cc-tools config reset
```

### Configuration Keys

| Key | Default | Description |
| --- | --- | --- |
| `validate.timeout` | `60` | Validation timeout in seconds |
| `validate.cooldown` | `5` | Cooldown between validation runs in seconds |
| `compact.threshold` | `50` | Context compaction threshold |
| `compact.reminder_interval` | `25` | Compaction reminder interval |
| `notify.quiet_hours.enabled` | `true` | Enable quiet hours for notifications |
| `notify.quiet_hours.start` | `21:00` | Quiet hours start time |
| `notify.quiet_hours.end` | `07:30` | Quiet hours end time |
| `notify.audio.enabled` | `true` | Enable audio notifications |
| `notify.audio.directory` | `~/.claude/audio` | Audio files directory |
| `notify.desktop.enabled` | `true` | Enable desktop notifications |
| `observe.enabled` | `true` | Enable tool usage observation |
| `observe.max_file_size_mb` | `10` | Maximum observation log file size in MB |
| `learning.min_session_length` | `10` | Minimum session length for learning |
| `learning.learned_skills_path` | `.claude/skills/learned` | Path for learned skills |
| `pre_commit_reminder.enabled` | `true` | Enable pre-commit reminder |
| `pre_commit_reminder.command` | `task pre-commit` | Pre-commit command to run |
| `drift.enabled` | `true` | Enable session drift detection |
| `drift.min_edits` | `6` | Minimum edits before drift check |
| `drift.threshold` | `0.2` | Drift detection threshold |
| `stop_reminder.enabled` | `true` | Enable stop reminders |
| `stop_reminder.interval` | `20` | Responses between reminders |
| `stop_reminder.warn_at` | `50` | Response count to trigger warning |
| `instinct.personal_path` | `~/.config/cc-tools/instincts/personal` | Personal instincts directory |
| `instinct.inherited_path` | `~/.config/cc-tools/instincts/inherited` | Inherited instincts directory |
| `instinct.min_confidence` | `0.3` | Minimum confidence for instincts |
| `instinct.auto_approve` | `0.7` | Auto-approve confidence threshold |
| `instinct.decay_rate` | `0.02` | Instinct confidence decay rate |
| `instinct.max_instincts` | `100` | Maximum number of instincts |
| `instinct.cluster_threshold` | `3` | Minimum instincts for cluster analysis |

---

## skip

Configure per-directory skip rules for linting and testing. Skips apply to the current working directory and are respected by `cc-tools validate`.

### Synopsis

```
cc-tools skip <subcommand>
```

### Subcommands

#### skip lint

Skip linting in the current directory.

```
cc-tools skip lint
```

#### skip test

Skip testing in the current directory.

```
cc-tools skip test
```

#### skip all

Skip both linting and testing in the current directory.

```
cc-tools skip all
```

#### skip list

Show all directories that have skip configurations.

```
cc-tools skip list
```

#### skip status

Show the skip status (active or skipped) for linting and testing in the current directory.

```
cc-tools skip status
```

### Examples

```bash
# Skip linting in a generated code directory
cd ~/projects/generated-api && cc-tools skip lint

# Check what is skipped in the current directory
cc-tools skip status

# See all directories with skip rules
cc-tools skip list
```

---

## unskip

Remove per-directory skip rules. Running `unskip` with no subcommand clears all skips for the current directory.

### Synopsis

```
cc-tools unskip [subcommand]
```

### Subcommands

#### unskip lint

Remove the lint skip for the current directory.

```
cc-tools unskip lint
```

#### unskip test

Remove the test skip for the current directory.

```
cc-tools unskip test
```

#### unskip all

Remove all skips for the current directory. This is the same behavior as running `unskip` with no subcommand.

```
cc-tools unskip all
```

### Examples

```bash
# Re-enable linting after skipping it
cc-tools unskip lint

# Clear all skips at once
cc-tools unskip
```

---

## mcp

Manage Claude MCP (Model Context Protocol) servers. Enable, disable, or list servers defined in your Claude settings.

### Synopsis

```
cc-tools mcp <subcommand>
```

### Subcommands

#### mcp list

Show all MCP servers and their current status (enabled or disabled).

```
cc-tools mcp list
```

#### mcp enable

Enable a single MCP server by name.

```
cc-tools mcp enable <name>
```

```bash
cc-tools mcp enable jira
```

#### mcp disable

Disable a single MCP server by name.

```
cc-tools mcp disable <name>
```

```bash
cc-tools mcp disable playwright
```

#### mcp enable-all

Enable all MCP servers defined in your settings.

```
cc-tools mcp enable-all
```

#### mcp disable-all

Disable all MCP servers.

```
cc-tools mcp disable-all
```

### Examples

```bash
# Check which servers are running
cc-tools mcp list

# Enable a server for the current session
cc-tools mcp enable context7

# Disable everything before a focused session
cc-tools mcp disable-all
```

---

## debug

Configure debug logging on a per-directory basis. Debug logs are written to `~/.cache/cc-tools/debug/`.

### Synopsis

```
cc-tools debug <subcommand>
```

### Subcommands

#### debug enable

Enable debug logging for the current directory. Prints the log file path on success.

```
cc-tools debug enable
```

#### debug disable

Disable debug logging for the current directory.

```
cc-tools debug disable
```

#### debug status

Show whether debug logging is enabled for the current directory, and the log file path if active.

```
cc-tools debug status
```

#### debug list

Show all directories that have debug logging enabled, along with their log file paths.

```
cc-tools debug list
```

#### debug filename

Print the debug log filename for the current directory. Useful for piping into other commands.

```
cc-tools debug filename
```

### Examples

```bash
# Enable debug logging and view the log
cc-tools debug enable
tail -f "$(cc-tools debug filename)"

# Check debug status across all projects
cc-tools debug list

# Disable when done investigating
cc-tools debug disable
```

---

## instinct

Manage learned instincts. Instincts are behavioral patterns that cc-tools learns from tool usage observations and stores for future reference.

### Synopsis

```
cc-tools instinct <subcommand>
```

### Subcommands

#### instinct status

List all instincts grouped by domain, with visual confidence bars.

```
cc-tools instinct status [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--domain` | (none) | Filter instincts to a specific domain |
| `--min-confidence` | `0` | Only show instincts at or above this confidence |

```bash
cc-tools instinct status
cc-tools instinct status --domain testing
cc-tools instinct status --domain testing --min-confidence 0.5
```

#### instinct export

Export instincts to YAML or JSON. Writes to stdout by default, or to a file with `--output`.

```
cc-tools instinct export [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--output` | (stdout) | Output file path |
| `--domain` | (none) | Filter by domain |
| `--min-confidence` | `0` | Minimum confidence threshold |
| `--format` | `yaml` | Output format: `yaml` or `json` |

```bash
cc-tools instinct export
cc-tools instinct export --format json --output instincts.json
cc-tools instinct export --domain testing --min-confidence 0.6
```

#### instinct import

Import instincts from a YAML or JSON file into the inherited instincts directory.

```
cc-tools instinct import <source> [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--dry-run` | `false` | Preview what would be imported without saving |
| `--force` | `false` | Overwrite existing instincts |
| `--min-confidence` | `0` | Only import instincts at or above this confidence |

```bash
cc-tools instinct import instincts.yaml
cc-tools instinct import instincts.yaml --dry-run
cc-tools instinct import shared-instincts.json --force --min-confidence 0.5
```

#### instinct evolve

Analyze instinct clusters and suggest candidates for promotion to skills, commands, or agents. Requires a minimum number of instincts (configured by `instinct.cluster_threshold`, default 3).

```
cc-tools instinct evolve
```

The output includes three sections:

- **Skill candidates** -- Groups of 3+ related instincts that could become a skill file.
- **Command candidates** -- High-confidence workflow instincts (confidence >= 0.7) that could become CLI commands.
- **Agent candidates** -- Large clusters (3+ instincts, average confidence >= 0.75) that could justify a dedicated agent.

```bash
cc-tools instinct evolve
```

---

## version

Print the cc-tools version string.

### Synopsis

```
cc-tools version
cc-tools --version
```

### Example

```bash
$ cc-tools version
cc-tools version v0.12.0
```
