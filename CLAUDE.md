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
