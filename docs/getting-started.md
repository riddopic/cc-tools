# Getting Started with cc-tools

In this tutorial, you will install cc-tools, connect it to Claude Code as a hook, and verify the integration end to end. By the end, you will have a working setup where cc-tools automatically validates file edits, dispatches hook events, and sends notifications during your Claude Code sessions.

## Prerequisites

Before you begin, make sure you have the following installed:

- **Go 1.26+** -- download from [go.dev/dl](https://go.dev/dl/)
- **Task** -- a task runner that replaces Make. Install from [taskfile.dev](https://taskfile.dev)
- **Claude Code** -- the Anthropic CLI. Install from [docs.anthropic.com](https://docs.anthropic.com/en/docs/claude-code)

Verify your Go version:

```bash
go version
```

You should see output like:

```
go version go1.26.0 darwin/arm64
```

Verify Task is available:

```bash
task --version
```

You should see output like:

```
Task version: v3.40.1
```

## Build and Install

Clone the repository and build the binary:

```bash
git clone https://github.com/riddopic/cc-tools.git
cd cc-tools
task build
```

You should see no errors. The compiled binary is now at `./bin/cc-tools`.

Install the binary to your `$GOPATH/bin` so it is available on your `PATH`:

```bash
task install
```

Confirm the binary is accessible:

```bash
cc-tools version
```

You should see output like:

```
cc-tools version 0.1.0
```

If you see `command not found`, make sure `$GOPATH/bin` is in your `PATH`. Add it to your shell profile if needed:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Configure Claude Code Hooks

Claude Code supports hooks that run commands before and after tool execution. You configure cc-tools as a hook so it receives every event from your Claude Code sessions.

Open (or create) the file `~/.claude/settings.json` and add the following `hooks` block:

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

If `~/.claude/settings.json` already exists with other settings, merge the `hooks` key into the existing JSON object. Do not replace the entire file.

This configuration does the following:

- **PreToolUse** -- runs `cc-tools hook` before every tool call for observation logging and pre-commit reminders.
- **PostToolUse** -- runs `cc-tools validate` after file edits (`Write`, `Edit`, `MultiEdit`) to lint and test in parallel. Also runs `cc-tools hook` for observation logging on all tool calls.
- **SessionStart / SessionEnd** -- tracks session lifecycle for metadata and context.
- **UserPromptSubmit** -- enables drift detection to warn when a session diverges from its original intent.
- **Stop** -- tracks response counts and emits session reminders.
- **Notification** -- plays audio, shows desktop alerts, and sends push notifications.
- **PreCompact** -- logs context compaction events.

## Verify the Installation

Run a quick smoke test by piping a hook event directly to cc-tools:

```bash
echo '{"hook_event_name":"Notification","title":"Test","message":"Hello from cc-tools"}' | cc-tools hook
```

You should hear an audio notification (if your system supports it) and see a macOS desktop alert. If audio is not available, cc-tools processes the event silently without error.

Next, start a Claude Code session and ask it to make a small file edit. Watch the terminal output. After the edit, cc-tools runs your project's linter and test suite automatically. If either fails, the edit is blocked with a formatted error message.

## Configure Basic Settings

cc-tools stores its configuration at `~/.config/cc-tools/config.json`. You manage settings with the `config` subcommand.

List all current settings and their values:

```bash
cc-tools config list
```

You should see a table showing each setting, its current value, and whether it is a default or custom value.

Adjust the validation timeout (in seconds) if your test suite needs more time:

```bash
cc-tools config set validate.timeout 120
```

You should see:

```
Set validate.timeout = 120
```

Disable audio notifications if you prefer silent operation:

```bash
cc-tools config set notify.audio.enabled false
```

Read back a setting to confirm it was saved:

```bash
cc-tools config get notify.audio.enabled
```

You should see:

```
false
```

To restore a setting to its default value:

```bash
cc-tools config reset notify.audio.enabled
```

## Enable Debug Logging

If something is not working as expected, enable debug logging for your project directory. Debug logs capture every cc-tools invocation with timestamps, arguments, and stdin data.

Navigate to your project directory, then enable debug logging:

```bash
cc-tools debug enable
```

You should see output like:

```
Debug logging enabled for /Users/you/projects/my-project
  Log file: /Users/you/.cache/cc-tools/debug/my-project.log
```

Find the log file path at any time:

```bash
cc-tools debug filename
```

Logs are written to `~/.cache/cc-tools/debug/`. Open the log file to inspect hook invocations, stdin payloads, and handler outputs. Each entry includes a timestamp and the full argument list.

When you are done debugging, disable logging:

```bash
cc-tools debug disable
```

## Next Steps

You now have a working cc-tools installation connected to Claude Code. From here, explore the rest of the documentation:

- [CLI Reference](cli-reference.md) -- full command reference for all subcommands and flags
- [Configuration](configuration.md) -- complete list of configuration keys and their defaults
- [Hooks and Handlers](hooks-and-handlers.md) -- how the hook dispatch system works and what each handler does
- [Troubleshooting](troubleshooting.md) -- common issues and solutions
