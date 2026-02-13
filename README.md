# cc-tools

A CLI companion for [Claude Code](https://claude.ai/code) that runs lint and test validation in parallel as a hook, manages MCP servers, and handles per-directory skip configuration.

## What it does

When configured as a Claude Code `PreToolUse` hook, cc-tools intercepts tool calls and runs your project's linter and test suite in parallel before accepting the change. If either fails, the tool call is blocked with a formatted error message.

Beyond validation, it provides commands for managing MCP server integrations, configuring per-directory skip rules, and tuning timeouts.

## Install

Requires Go 1.26+ and [Task](https://taskfile.dev).

```bash
task build      # Build to ./bin/cc-tools
task install    # Copy to $GOPATH/bin
```

## Usage

```bash
cc-tools <command> [arguments]
```

| Command | Description |
|---------|-------------|
| `validate` | Run lint and test in parallel (reads JSON from stdin) |
| `skip` | Configure directories to skip validation |
| `unskip` | Remove skip settings from directories |
| `debug` | Configure debug logging |
| `mcp` | Manage Claude MCP servers (list, enable, disable) |
| `config` | Manage application settings |
| `version` | Print version |

### Hook integration

Add to your Claude Code settings as a `PreToolUse` hook:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit",
        "command": "cc-tools validate"
      }
    ]
  }
}
```

The `validate` command reads the tool call JSON from stdin, determines the project type, discovers available lint/test commands, and runs them concurrently.

### Examples

```bash
# Validate via hook (stdin receives tool call JSON)
echo '{"file_path": "main.go"}' | cc-tools validate

# Manage MCP servers
cc-tools mcp list
cc-tools mcp enable jira
cc-tools mcp disable jira

# Skip validation for a directory
cc-tools skip --type lint /path/to/generated

# Configure validation timeout
cc-tools config set hooks.validate.timeout_seconds 120
```

## Configuration

Settings are stored at `~/.config/cc-tools/config.yaml`. Environment variables with the `CC_TOOLS_` prefix override config file values.

| Setting | Env Override | Default | Description |
|---------|-------------|---------|-------------|
| `hooks.validate.timeout_seconds` | `CC_TOOLS_HOOKS_VALIDATE_TIMEOUT_SECONDS` | 60 | Validation timeout |
| `hooks.validate.cooldown_seconds` | `CC_TOOLS_HOOKS_VALIDATE_COOLDOWN_SECONDS` | 5 | Cooldown between runs |

Debug logs are written to `~/.cache/cc-tools/debug/`.

## Development

```bash
task doctor         # Check required tools
task tools-install  # Install gotestsum, golangci-lint, goimports, mockery
task test           # Fast tests (-short, 30s)
task lint           # golangci-lint
task check          # fmt + lint + test-race (run before committing)
task mocks          # Regenerate mocks
task coverage       # HTML coverage report
```

## License

See [LICENSE](LICENSE) for details.
