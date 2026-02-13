# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

**cc-tools** is a CLI tool for Claude Code that automates validation hooks, MCP server management, skip registry, and configuration. It runs as a Claude Code hook to execute lint and test in parallel before tool use is accepted.

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

Hand-rolled CLI (no Cobra/Viper) — `cmd/cc-tools/main.go` dispatches via `switch` on `os.Args[1]`. Commands: `validate`, `skip`, `unskip`, `debug`, `mcp`, `config`, `version`.

### Internal packages

| Package | Purpose |
|---------|---------|
| `internal/hooks` | Core validation: discovers lint/test commands, runs them in parallel with `sync.WaitGroup`, manages cooldown locks |
| `internal/shared` | Filesystem interfaces (`HooksFS`, `RegistryFS`, `FS`), project detection, dependency injection container, debug path utilities |
| `internal/config` | JSON config persistence at `~/.config/cc-tools/config.json` |
| `internal/skipregistry` | Persistent skip settings (lint/test/all) per directory with JSON storage |
| `internal/output` | Thread-safe terminal writer using `charmbracelet/lipgloss` for styled output |
| `internal/mcp` | MCP server enable/disable management |
| `internal/debug` | Debug log infrastructure at `~/.cache/cc-tools/debug/` |

### Key design patterns

- **Dependency injection**: `shared.Dependencies` struct holds `FS`, `Runner`, `Stdout`, `Stderr` — injected into hooks for testability
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
- `shared`: `HooksFS`, `RegistryFS`, `FS`

Config: `.mockery.yml` at project root.

## Linting

golangci-lint v2 with strict settings: max cyclomatic complexity 30, max cognitive complexity 20, max function length 100 lines/50 statements, max line length 120 chars. Over 60 linters enabled including `gosec`, `errcheck`, `errorlint`.

## Imports

Three groups separated by blank lines: stdlib, third-party, then `github.com/riddopic/cc-tools/internal/...`. Enforced by `goimports -local github.com/riddopic/cc-tools`.

## Conventions

- Conventional commits: `feat:`, `fix:`, `refactor:`, `test:`, `chore:`, `docs:`, `perf:`, `ci:`
- Run `task check` before every commit
- No project management references (sprint numbers, ticket IDs) in code comments
- Functions under 50 lines, nesting under 3 levels
- Wrap errors with context using `fmt.Errorf("...: %w", err)`
- Early returns with guard clauses over nested conditionals
