# Configuration Reference

cc-tools reads configuration from a single JSON file and exposes all keys through a unified CLI interface.

**Config file location:** `~/.config/cc-tools/config.json`

**CLI management:**

```bash
cc-tools config get <key>       # Read a single key
cc-tools config set <key> <val> # Write a single key
cc-tools config list            # Show all keys and current values
cc-tools config reset [key]     # Reset one key or all keys to defaults
```

## Precedence

When the same setting is available through multiple sources, cc-tools applies this resolution order (highest wins):

1. Environment variables
2. Config file (`~/.config/cc-tools/config.json`)
3. Built-in defaults

## Validation

Controls timeout and cooldown for the `cc-tools validate` command, which runs lint and test commands in parallel.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `validate.timeout` | int | `60` | Validation timeout in seconds |
| `validate.cooldown` | int | `5` | Cooldown between validation runs in seconds |

**Environment variable overrides:**

| Variable | Overrides |
|----------|-----------|
| `CC_TOOLS_HOOKS_VALIDATE_TIMEOUT_SECONDS` | `validate.timeout` |
| `CC_TOOLS_HOOKS_VALIDATE_COOLDOWN_SECONDS` | `validate.cooldown` |

## Notifications

Configures push notifications through the [ntfy.sh](https://ntfy.sh) service.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `notifications.ntfy_topic` | string | `""` | ntfy.sh topic for push notifications |

Set this to your ntfy topic name to receive push notifications on your phone or desktop when long-running operations complete. Leave empty to disable.

## Compact Context

Tracks tool-call volume and suggests running `/compact` when the context window grows large.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `compact.threshold` | int | `50` | Tool-call count that triggers a compact suggestion |
| `compact.reminder_interval` | int | `25` | Tool calls between subsequent compact reminders |

## Notification Dispatch

Fine-grained control over how and when cc-tools delivers local notifications. Covers quiet hours, audio alerts, and macOS desktop banners.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `notify.quiet_hours.enabled` | bool | `true` | Suppress notifications during quiet hours |
| `notify.quiet_hours.start` | string | `"21:00"` | Quiet hours start time (HH:MM, 24-hour format) |
| `notify.quiet_hours.end` | string | `"07:30"` | Quiet hours end time (HH:MM, 24-hour format) |
| `notify.audio.enabled` | bool | `true` | Enable audio notification sounds |
| `notify.audio.directory` | string | `"~/.claude/audio"` | Path to directory containing MP3 files |
| `notify.desktop.enabled` | bool | `true` | Enable macOS desktop notifications |

Audio notifications play a random MP3 from the configured directory. Place your preferred sound files there to customize the alert.

## Observation

Controls the tool-use observation logger that feeds the instinct learning system.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `observe.enabled` | bool | `true` | Enable tool-use observation logging |
| `observe.max_file_size_mb` | int | `10` | Max observation file size in MB before rotation |

Observations are written to `~/.cache/cc-tools/observations/observations.jsonl` in newline-delimited JSON format.

## Learning

Configures automatic skill extraction from session history.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `learning.min_session_length` | int | `10` | Minimum session length (in tool calls) for learning extraction |
| `learning.learned_skills_path` | string | `".claude/skills/learned"` | Path for learned skill files |

## Pre-Commit Reminder

Reminds you to run quality checks before committing code through a `PreToolUse` hook on `git commit`.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `pre_commit_reminder.enabled` | bool | `true` | Remind to run checks before git commit |
| `pre_commit_reminder.command` | string | `"task pre-commit"` | Command to suggest before commits |

## Package Manager

Overrides automatic package manager detection for projects that use multiple managers.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `package_manager.preferred` | string | `""` | Preferred package manager (overrides auto-detection) |

When empty, cc-tools auto-detects the package manager from lock files in the current directory. Set this to `npm`, `pnpm`, `yarn`, or another manager name to force a specific choice.

## Drift Detection

Monitors session prompts for topic drift and warns when you stray from the original intent.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `drift.enabled` | bool | `true` | Enable drift detection on prompts |
| `drift.min_edits` | int | `6` | Minimum prompt count before checking for drift |
| `drift.threshold` | float | `0.2` | Keyword overlap ratio below which drift is flagged |

The detector extracts keywords from your first prompt and compares subsequent prompts against them. A lower threshold makes detection more sensitive. Pivot phrases like "now let's" or "switch to" reset the baseline automatically.

## Stop Reminder

Emits periodic reminders during long sessions to encourage natural stopping points.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `stop_reminder.enabled` | bool | `true` | Enable periodic session reminders |
| `stop_reminder.interval` | int | `20` | Responses between reminders |
| `stop_reminder.warn_at` | int | `50` | Response count that triggers a strong wrap-up warning |

## Instinct Management

Controls the instinct learning system that captures, evolves, and applies behavioral patterns from your sessions.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `instinct.personal_path` | string | `"~/.config/cc-tools/instincts/personal"` | Directory for personal instincts |
| `instinct.inherited_path` | string | `"~/.config/cc-tools/instincts/inherited"` | Directory for imported instincts |
| `instinct.min_confidence` | float | `0.3` | Minimum confidence for instinct activation |
| `instinct.auto_approve` | float | `0.7` | Confidence threshold for automatic approval |
| `instinct.decay_rate` | float | `0.02` | Confidence decay per week without reinforcement |
| `instinct.max_instincts` | int | `100` | Maximum number of instincts to retain |
| `instinct.cluster_threshold` | int | `3` | Minimum instincts in a cluster for evolve analysis |

Instincts below `min_confidence` are not activated. Those above `auto_approve` are applied without prompting. The `decay_rate` reduces confidence by the configured amount for each full week since the instinct's `updated_at` timestamp. Decay is evaluated at read time during `status`, `export`, and `evolve` without mutating stored files. During `import`, decay is applied and the decayed values are persisted to the inherited store. Instincts that fall below `min_confidence` through decay become candidates for pruning.

## File Paths

cc-tools reads from and writes to several well-known locations on disk.

| Path | Purpose |
|------|---------|
| `~/.config/cc-tools/config.json` | Configuration file |
| `~/.cache/cc-tools/debug/` | Debug logs |
| `~/.cache/cc-tools/observations/observations.jsonl` | Tool-use observation log |
| `~/.config/cc-tools/instincts/personal/` | Personal instincts |
| `~/.config/cc-tools/instincts/inherited/` | Imported instincts |
| `~/.claude/sessions/` | Session data |
| `~/.claude/session-aliases.json` | Session alias mappings |
| `~/.claude/audio/` | Audio notification files (MP3) |

## Examples

Common configuration scenarios:

```bash
# Increase validation timeout for large projects
cc-tools config set validate.timeout 120

# Disable audio notifications
cc-tools config set notify.audio.enabled false

# Set quiet hours
cc-tools config set notify.quiet_hours.start 22:00
cc-tools config set notify.quiet_hours.end 08:00

# Configure ntfy push notifications
cc-tools config set notifications.ntfy_topic my-cc-tools

# Disable drift detection
cc-tools config set drift.enabled false

# View all settings
cc-tools config list

# Reset a single key
cc-tools config reset validate.timeout

# Reset all settings
cc-tools config reset
```
