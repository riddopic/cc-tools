# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

**cc-tools** is a CLI companion for Claude Code that automates hook event handling, parallel lint/test validation, MCP server management, notifications, session tracking, and per-directory skip configuration. It runs as a Claude Code hook to dispatch registered handlers for every hook event.

Module: `github.com/riddopic/cc-tools` — Go 1.26

## Build & Development

Uses [Task](https://taskfile.dev) (not Make). Binary: `bin/cc-tools`, entry point: `./cmd/cc-tools`.

```bash
task build              # Build binary
task test               # Fast tests (-short, 30s timeout)
task lint               # golangci-lint (5m timeout)
task fmt                # gofmt + goimports
task check              # fmt + lint + test-race (alias: pre-commit)
task test-race          # Tests with race detector
task mocks              # Generate mocks via mockery
task coverage           # HTML coverage report
task doctor             # Check dev environment
task tools-install      # Install required tools
```

Run a single test:

```bash
gotestsum --format pkgname -- -tags=testmode -run TestFunctionName ./internal/hooks/...
```

All test commands require the `-tags=testmode` build tag. Tests use `gotestsum` (not raw `go test`).

## Architecture

CLI built with [Cobra](https://github.com/spf13/cobra) — `cmd/cc-tools/main.go` registers subcommands via `root.AddCommand()`. Commands: `hook`, `validate`, `session`, `config`, `skip`, `unskip`, `mcp`, `debug`, `version`.

Two separate execution paths:

- **`cc-tools hook`** — reads hook event JSON from stdin, dispatches to a handler registry (`internal/handler`), returns structured output
- **`cc-tools validate`** — reads tool call JSON from stdin, discovers lint/test commands, runs them in parallel

### Internal packages

| Package | Purpose |
|---------|---------|
| `internal/handler` | Handler registry and event dispatch; all hook handlers (notification, session, observation, compaction, superpowers, pre-commit, drift detection, stop reminder) |
| `internal/hookcmd` | Hook event constants, JSON input parsing, `HookInput` struct |
| `internal/hooks` | Core validation: discovers lint/test commands, runs them in parallel with `sync.WaitGroup`, manages cooldown locks |
| `internal/notify` | Notification backends: `NtfyNotifier` (HTTP push), `Audio` (afplay), `Desktop` (osascript), `QuietHours`, `MultiNotifier` |
| `internal/config` | JSON config persistence at `~/.config/cc-tools/config.json`; `Values` struct, `Manager` with get/set/reset |
| `internal/session` | Session metadata storage, alias management, search |
| `internal/observe` | Tool use observation logging to filesystem |
| `internal/compact` | Context compaction suggestions and log tracking |
| `internal/superpowers` | Skill injection for SessionStart events |
| `internal/pkgmanager` | Package manager detection (npm, pnpm, bun, yarn) |
| `internal/shared` | Filesystem interfaces (`HooksFS`, `RegistryFS`, `FS`), project detection, dependency injection container, debug path utilities |
| `internal/skipregistry` | Persistent skip settings (lint/test/all) per directory with JSON storage |
| `internal/output` | Thread-safe terminal writer using `charmbracelet/lipgloss` for styled output |
| `internal/mcp` | MCP server enable/disable management |
| `internal/debug` | Debug log infrastructure at `~/.cache/cc-tools/debug/` |

### Handler registry

`internal/handler/defaults.go` wires all handlers to their events:

| Event | Handlers |
|-------|----------|
| SessionStart | `SuperpowersHandler`, `PkgManagerHandler`, `SessionContextHandler` |
| SessionEnd | `SessionEndHandler` |
| PreToolUse | `SuggestCompactHandler`, `ObserveHandler(pre)`, `PreCommitReminderHandler` |
| PostToolUse | `ObserveHandler(post)` |
| PostToolUseFailure | `ObserveHandler(failure)` |
| UserPromptSubmit | `DriftHandler` |
| Stop | `StopReminderHandler` |
| PreCompact | `LogCompactionHandler` |
| Notification | `NotifyAudioHandler`, `NotifyDesktopHandler`, `NotifyNtfyHandler` |

### Key design patterns

- **Handler interface**: `Handler.Handle(ctx, *HookInput) (*Response, error)` with `Name()` for identification
- **Dependency injection**: Functional options pattern (`WithAudioPlayer`, `WithCmdRunner`, `WithNtfySender`) for testability
- **Registry dispatch**: Iterates handlers for an event; errors logged to stderr, dispatch continues
- **Three filesystem interfaces**: `HooksFS` (hook operations), `RegistryFS` (registry persistence), `FS` (general stat/getwd) — each scoped to its concern
- **Compile-time interface checks**: `var _ Interface = (*Impl)(nil)`
- **Parallel validation**: lint and test run concurrently via `sync.WaitGroup`; `LockManager` prevents concurrent validation races

## Testing

TDD is mandatory. Mock generation uses mockery v3.5 with testify template:

```bash
task mocks              # Regenerate all mocks
```

Mocks live in `internal/{package}/mocks/`. Mocked interfaces:

- `hooks`: `CommandRunner`, `ProcessManager`, `Clock`, `InputReader`, `OutputWriter`
- `notify`: `AudioPlayer`, `CmdRunner`
- `shared`: `HooksFS`, `RegistryFS`, `FS`

Config: `.mockery.yml` at project root.

## Linting

golangci-lint v2 with strict settings: max cyclomatic complexity 30, max cognitive complexity 20, max function length 100 lines/50 statements, max line length 120 chars. Over 60 linters enabled including `gosec`, `errcheck`, `errorlint`, `exhaustruct`.

## Imports

Three groups separated by blank lines: stdlib, third-party, then `github.com/riddopic/cc-tools/internal/...`. Enforced by `goimports -local github.com/riddopic/cc-tools`.

## Conventions

- Conventional commits: `feat:`, `fix:`, `refactor:`, `test:`, `chore:`, `docs:`, `perf:`, `ci:`
- Run `task check` before every commit
- No project management references (sprint numbers, ticket IDs) in code comments
- Functions under 50 lines, nesting under 3 levels
- Wrap errors with context using `fmt.Errorf("...: %w", err)`
- Early returns with guard clauses over nested conditionals
