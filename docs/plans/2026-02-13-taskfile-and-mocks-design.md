# Taskfile and Mock Generation Design

## Problem

The project has a Taskfile.yml copied from another project. It references
packages, version flags, and tasks that do not exist in cc-tools. The project
also has hand-written mocks scattered across test files, with a `.mockery.yml`
that has no packages configured.

## Decisions

### Taskfile

**Strip to daily-dev tasks only.** Remove version/ldflags embedding, release
tasks, cross-platform builds, and security checks. Fix coverage exclusions to
match this project's actual packages.

**Tasks to keep (14):**

| Task | Aliases | Notes |
|------|---------|-------|
| `default` | | Lists available tasks |
| `test` | `qt` | Fast unit tests with gotestsum |
| `watch` | | Auto-run tests on file changes |
| `lint` | `ql` | golangci-lint |
| `fmt` | | gofmt + goimports |
| `check` | `pre-commit` | fmt, lint, test-race |
| `test-race` | | Race detector, short mode |
| `build` | `q` | Plain go build, no ldflags |
| `install` | | Copy binary to GOPATH/bin |
| `clean` | | Remove build artifacts |
| `coverage` | | HTML coverage report |
| `bench` | | Benchmarks with gotestsum |
| `mocks` | | Generate mocks via mockery |
| `polish` | | Format, auto-fix, clean backups |
| `doctor` | | Check dev environment |
| `tools-install` | | Install required tools |

**Tasks to remove:** `release-check`, `release-dry`, `release`, `security`,
`vulncheck`, `build-all`, `build-linux`, `build-darwin`, `build-windows`,
`build-cross`, `version`, `integration`, `test-race-full`.

**Coverage exclusions** simplify to filtering out `/mocks` directories only.

### Mock Generation

**Configure mockery for all 8 interfaces across 2 packages.** Generate into
per-package `mocks/` subdirectories.

**Interfaces to mock:**

`internal/hooks`:
- `CommandRunner` (2 methods)
- `ProcessManager` (3 methods)
- `Clock` (1 method)
- `InputReader` (2 methods)
- `OutputWriter` (embeds io.Writer)

`internal/shared`:
- `HooksFS` (6 methods)
- `RegistryFS` (4 methods)
- `SharedFS` (3 methods)

**Output structure:**
```
internal/hooks/mocks/MockCommandRunner.go
internal/hooks/mocks/MockProcessManager.go
internal/hooks/mocks/MockClock.go
internal/hooks/mocks/MockInputReader.go
internal/hooks/mocks/MockOutputWriter.go
internal/shared/mocks/MockHooksFS.go
internal/shared/mocks/MockRegistryFS.go
internal/shared/mocks/MockSharedFS.go
```

**Migration strategy:** Hand-written mocks in `mocks_test.go` remain
untouched. New tests use generated mocks. Old tests migrate when modified.
