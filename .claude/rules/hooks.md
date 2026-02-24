# Claude Code Hooks

Claude Code hooks automate actions before/after tool execution.

## Hook Types

| Type | Trigger | Use Case |
| ------ | --------- | ---------- |
| **PreToolUse** | Before tool execution | Validation, parameter modification |
| **PostToolUse** | After tool execution | Auto-formatting, checks |
| **UserPromptSubmit** | User sends a prompt | Drift detection, prompt analysis |
| **Stop** | Session ends | Response counting, verification |

## Current Hooks

Configured in `~/.claude/settings.json`:

### PreToolUse Hooks

- **tmux reminder**: Suggests tmux for long-running commands (npm, pnpm, yarn, cargo, etc.)
- **git push review**: Opens editor for review before push
- **doc blocker**: Blocks creation of unnecessary .md/.txt files

### PostToolUse Hooks

- **PR creation**: Logs PR URL and GitHub Actions status
- **Prettier**: Auto-formats JS/TS files after edit
- **TypeScript check**: Runs tsc after editing .ts/.tsx files
- **console.log warning**: Warns about console.log in edited files

### UserPromptSubmit Handlers

- **drift detection** (`DriftHandler`): Tracks session intent from the first prompt, extracts keywords, and warns when subsequent prompts diverge significantly. Configurable via `drift.enabled`, `drift.min_edits` (default 6), `drift.threshold` (default 0.2). Recognizes pivot phrases ("now let's", "switch to", etc.) to reset intent.

### Stop Handlers

- **stop reminder** (`StopReminderHandler`): Tracks response count per session and emits rotating reminders at configurable intervals. Configurable via `stop_reminder.enabled`, `stop_reminder.interval` (default 20), `stop_reminder.warn_at` (default 50).
- **console.log audit**: Checks all modified files for console.log before session ends

### Built-in Handlers (cc-tools hook)

- **observe** (`ObserveHandler`): Captures tool usage events (tool name, input, output, errors) to `~/.cache/cc-tools/observations/observations.jsonl`. Feeds the instinct learning system. No manual hook configuration needed — dispatched automatically by `cc-tools hook`.

## Hook Configuration

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "command": "your-validation-script.sh"
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Edit",
        "command": "your-post-edit-script.sh"
      }
    ]
  }
}
```

## Best Practices

### Auto-Accept Permissions

Use with caution:

- Enable for trusted, well-defined plans
- Disable for exploratory work
- Never use `dangerously-skip-permissions` flag
- Configure `allowedTools` in `~/.claude.json` instead

### TodoWrite Usage

Use TodoWrite tool to:

- Track progress on multi-step tasks
- Verify understanding of instructions
- Enable real-time steering
- Show granular implementation steps

Todo list reveals:

- Out of order steps
- Missing items
- Extra unnecessary items
- Wrong granularity
- Misinterpreted requirements

## Hook Troubleshooting

If a hook blocks your action:

1. Check the blocked message for guidance
2. Determine if you can adjust your action
3. If not adjustable, ask user to check hooks configuration

Common issues:

- Doc blocker preventing legitimate documentation
- Long-running command warnings for quick operations
- Pre-commit checks failing on partial work

## Project-Specific Hooks

Ensure these run before commits:

```bash
# These should pass before any commit
task fmt      # Format code
task lint     # Check for issues
task test     # Run tests
```

Consider adding as PreToolUse hooks for `git commit` operations.
