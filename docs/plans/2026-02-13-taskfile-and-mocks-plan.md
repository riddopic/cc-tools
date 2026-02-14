# Taskfile and Mock Generation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use executing-plans to implement this plan task-by-task.

**Goal:** Adapt the Taskfile.yml for cc-tools and configure mockery to generate mocks for all project interfaces.

**Architecture:** Two independent changes — (1) rewrite Taskfile.yml to remove irrelevant tasks and fix project-specific references, (2) add packages section to `.mockery.yml` and generate mocks.

**Tech Stack:** [Task](https://taskfile.dev) v3, [mockery](https://vektra.github.io/mockery/) v3.2.5, testify/mock

---

### Task 1: Strip Taskfile variables

**Files:**
- Modify: `Taskfile.yml:1-31`

**Step 1: Replace the vars section**

Remove `VERSION`, `COMMIT`, `BUILD_DATE`, `BUILD_USER`, `GOOS`, `GOARCH`,
`PLATFORM`, `VERSION_FLAGS`, and `LDFLAGS`. Keep only `BINARY_NAME`, `BIN_DIR`,
`BINARY_PATH`, `MAIN_PATH`, `COVERAGE_DIR`, `COVERAGE_FILE`, `COVERAGE_HTML`.

Replace lines 1-31 with:

```yaml
# yaml-language-server: $schema=https://taskfile.dev/schema.json
version: "3"

vars:
  BINARY_NAME: cc-tools
  BIN_DIR: bin
  BINARY_PATH: "{{.BIN_DIR}}/{{.BINARY_NAME}}"
  MAIN_PATH: "./cmd/cc-tools"
  COVERAGE_DIR: coverage
  COVERAGE_FILE: "{{.COVERAGE_DIR}}/coverage.out"
  COVERAGE_HTML: "{{.COVERAGE_DIR}}/coverage.html"
```

**Step 2: Verify the file is valid YAML**

Run: `task --list 2>&1 | head -5`
Expected: task lists available tasks without errors

**Step 3: Commit**

```bash
git add Taskfile.yml
git commit -m "chore: strip version flags from Taskfile"
```

---

### Task 2: Remove release, cross-platform, security, and integration tasks

**Files:**
- Modify: `Taskfile.yml:70-322` (everything after `fmt`)

**Step 1: Update the check task**

Change `check` to remove the `vulncheck` dependency. Replace:

```yaml
  check:
    desc: Run all pre-commit checks (fmt + lint + test-race + vulncheck)
    aliases: [pre-commit]
    cmds:
      - task: fmt
      - task: lint
      - task: test-race
      - task: vulncheck
```

With:

```yaml
  check:
    desc: Run all pre-commit checks (fmt + lint + test-race)
    aliases: [pre-commit]
    cmds:
      - task: fmt
      - task: lint
      - task: test-race
```

**Step 2: Remove test-race-full task**

Delete the entire `test-race-full` task block (lines 91-97).

**Step 3: Remove integration task**

Delete the entire `integration` task block (lines 99-113).

**Step 4: Simplify the build task**

Replace the build command to remove ldflags:

```yaml
  build:
    desc: Build the binary
    aliases: [q]
    sources:
      - "**/*.go"
      - go.mod
      - go.sum
    generates:
      - "{{.BINARY_PATH}}"
    cmds:
      - mkdir -p {{.BIN_DIR}}
      - go build -o {{.BINARY_PATH}} {{.MAIN_PATH}}
```

**Step 5: Simplify the clean task**

Remove `dist/` reference:

```yaml
  clean:
    desc: Remove build artifacts
    cmds:
      - rm -rf {{.BIN_DIR}} {{.COVERAGE_DIR}}
      - go clean -cache -testcache
```

**Step 6: Delete release tasks**

Delete `release-check`, `release-dry`, and `release` task blocks (lines 146-165).

**Step 7: Simplify coverage exclusions**

Replace the coverage task's package filtering. Remove all the `rg -v` lines
for packages that don't exist (`foundry`, `blockchain/tvl`, `benchmark`,
`logger`, `gas`, `security`, `forge/parser`, `blockchain/chain`, `proxy`,
`dex/registry`, `cli/commands`, `interfaces`). Keep only the mocks exclusion:

```yaml
  coverage:
    desc: Generate test coverage report
    preconditions:
      - sh: command -v gotestsum
        msg: "gotestsum not installed. Run: task tools-install"
    cmds:
      - mkdir -p {{.COVERAGE_DIR}}
      - |
        PKGS_COVER=$(go list ./... | grep -v "/mocks")
        gotestsum --format pkgname -- -tags=testmode -coverprofile={{.COVERAGE_FILE}} -covermode=atomic $PKGS_COVER
      - go tool cover -html={{.COVERAGE_FILE}} -o {{.COVERAGE_HTML}}
      - go tool cover -func={{.COVERAGE_FILE}} | tail -n 1
```

Note: use `grep -v` instead of `rg -v` to avoid requiring ripgrep as a
dependency for the build system.

**Step 8: Delete security and vulncheck tasks**

Delete the `security` and `vulncheck` task blocks (lines 215-232).

**Step 9: Update doctor to remove gosec**

Replace doctor:

```yaml
  doctor:
    desc: Check development environment
    silent: true
    cmds:
      - 'echo -n "  Go:            " && go version | cut -d" " -f3'
      - 'echo -n "  golangci-lint: " && (golangci-lint version --short 2>/dev/null || echo "not installed")'
      - 'echo -n "  goimports:     " && (which goimports > /dev/null 2>&1 && echo "installed" || echo "not installed")'
      - 'echo -n "  gotestsum:     " && (gotestsum --version 2>/dev/null || echo "not installed (REQUIRED)")'
      - 'echo -n "  mockery:       " && (mockery 2>&1 | grep -o "v[0-9.]*" | head -1 || echo "not installed")'
```

**Step 10: Update tools-install to remove gosec and govulncheck**

```yaml
  tools-install:
    desc: Install required development tools
    cmds:
      - go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
      - go install golang.org/x/tools/cmd/goimports@latest
      - go install gotest.tools/gotestsum@latest
      - go install github.com/vektra/mockery/v3@v3.2.5
```

**Step 11: Delete cross-platform build tasks**

Delete `build-all`, `build-linux`, `build-darwin`, `build-windows`,
`build-cross`, and `version` task blocks (lines 267-322).

**Step 12: Verify the final Taskfile**

Run: `task --list`
Expected: 14 tasks listed (default, test, watch, lint, fmt, check, test-race,
build, install, clean, coverage, bench, mocks, polish, doctor, tools-install)

Run: `task build`
Expected: Binary builds at `bin/cc-tools`

**Step 13: Commit**

```bash
git add Taskfile.yml
git commit -m "chore: strip Taskfile to daily-dev tasks for cc-tools"
```

---

### Task 3: Configure mockery packages

**Files:**
- Modify: `.mockery.yml:74-76`

**Step 1: Add packages section to .mockery.yml**

Append the packages configuration after line 76. The interfaces to mock are:

From `internal/hooks/dependencies.go`:
- `CommandRunner` (methods: RunContext, LookPath)
- `ProcessManager` (methods: GetPID, FindProcess, ProcessExists)
- `Clock` (methods: Now)
- `InputReader` (methods: ReadAll, IsTerminal)
- `OutputWriter` (embeds io.Writer)

From `internal/shared/fs.go`:
- `HooksFS` (methods: Stat, ReadFile, WriteFile, TempDir, CreateExclusive, Remove)
- `RegistryFS` (methods: ReadFile, WriteFile, MkdirAll, UserHomeDir)
- `SharedFS` (methods: Stat, Getwd, Abs)

Append this YAML:

```yaml
packages:
  github.com/riddopic/cc-tools/internal/hooks:
    interfaces:
      CommandRunner:
      ProcessManager:
      Clock:
      InputReader:
      OutputWriter:

  github.com/riddopic/cc-tools/internal/shared:
    interfaces:
      HooksFS:
      RegistryFS:
      SharedFS:
```

**Step 2: Strip verbose comments from .mockery.yml**

The current file has multi-line explanatory comments for every field. Replace
the entire file with a concise version:

```yaml
all: false
filename: "{{.InterfaceName}}.go"
force-file-write: true
formatter: goimports
log-level: info
structname: "Mock{{.InterfaceName}}"
pkgname: "mocks"
recursive: false
require-template-schema-exists: true
template: testify
template-schema: "{{.Template}}.schema.json"

packages:
  github.com/riddopic/cc-tools/internal/hooks:
    interfaces:
      CommandRunner:
      ProcessManager:
      Clock:
      InputReader:
      OutputWriter:

  github.com/riddopic/cc-tools/internal/shared:
    interfaces:
      HooksFS:
      RegistryFS:
      SharedFS:
```

**Step 3: Commit**

```bash
git add .mockery.yml
git commit -m "chore: configure mockery for all project interfaces"
```

---

### Task 4: Generate mocks and verify

**Files:**
- Create: `internal/hooks/mocks/` (5 files, auto-generated)
- Create: `internal/shared/mocks/` (3 files, auto-generated)

**Step 1: Run mockery**

Run: `task mocks`
Expected: 8 mock files generated without errors

**Step 2: Verify generated files exist**

Run: `ls internal/hooks/mocks/ internal/shared/mocks/`
Expected:
```
internal/hooks/mocks/:
Clock.go  CommandRunner.go  InputReader.go  OutputWriter.go  ProcessManager.go

internal/shared/mocks/:
HooksFS.go  RegistryFS.go  SharedFS.go
```

**Step 3: Verify the project builds with new mocks**

Run: `go build ./...`
Expected: Clean build, no errors

**Step 4: Verify existing tests still pass**

Run: `go test -tags=testmode -short -timeout=30s ./internal/hooks/... ./internal/shared/... ./internal/skipregistry/...`
Expected: All tests pass (existing hand-written mocks are unaffected)

**Step 5: Commit generated mocks**

```bash
git add internal/hooks/mocks/ internal/shared/mocks/
git commit -m "chore: generate mockery mocks for all interfaces"
```
