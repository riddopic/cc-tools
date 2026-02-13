# Claude Code Hooks

Claude Code hooks automate actions before/after tool execution.

## Hook Types

| Type | Trigger | Use Case |
| ------ | --------- | ---------- |
| **PreToolUse** | Before tool execution | Validation, parameter modification |
| **PostToolUse** | After tool execution | Auto-formatting, checks |
| **Stop** | Session ends | Final verification |

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

### Stop Hooks

- **console.log audit**: Checks all modified files for console.log before session ends

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

For Quanta, ensure these run before commits:

```bash
# These should pass before any commit
task fmt      # Format code
task lint     # Check for issues
task test     # Run tests
```

Consider adding as PreToolUse hooks for `git commit` operations.
