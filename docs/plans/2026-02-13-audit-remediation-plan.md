# Audit Remediation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use executing-plans to implement this plan task-by-task.

**Goal:** Fix all six findings from `docs/audits/context-audit-20260213-213500.md` — temp file leak, silent errors, interface escape, config location, dead code, and docs/naming.

**Architecture:** Six independent, incremental changes. Each fix is self-contained and can be committed separately. Order matters only for Fix 3 (must precede Fix 4 since Fix 4 changes paths that use the interface).

**Tech Stack:** Go 1.26, mockery v3.5 (testify template), gotestsum, golangci-lint.

---

### Task 1: Fix temp file leak in `debugLog()`

`cmd/cc-tools/main.go:154` creates a temp file via `os.CreateTemp` every time `validate` runs. The file is never removed.

**Files:**
- Modify: `cmd/cc-tools/main.go:132-188`

**Step 1: Add `stdinTempFile` package variable and cleanup**

In `cmd/cc-tools/main.go`, add a package-level variable and clean up after `runValidate()`:

```go
// Add after line 25 (var version = "dev")
var stdinTempFile string
```

In `debugLog()`, capture the temp file path (around line 154):

```go
if tmpFile, tmpErr := os.CreateTemp("", "cc-tools-stdin-"); tmpErr == nil {
    _, _ = tmpFile.Write(stdinDebugData)
    _, _ = tmpFile.Seek(0, 0)
    os.Stdin = tmpFile //nolint:reassign // Resetting stdin for subsequent reads
    stdinTempFile = tmpFile.Name()
}
```

In the `switch` block, clean up after validate (around line 39-40):

```go
case "validate":
    runValidate()
    if stdinTempFile != "" {
        _ = os.Remove(stdinTempFile)
    }
```

**Step 2: Verify the fix compiles and tests pass**

Run: `task check`
Expected: All checks pass (fmt, lint, test).

**Step 3: Commit**

```bash
git add cmd/cc-tools/main.go
git commit -m "$(cat <<'EOF'
fix: clean up temp file created by debugLog()

debugLog() creates a temp file per validate invocation to buffer stdin.
The file was never deleted, leaking one file per run. Add package-level
tracking and cleanup after runValidate() returns.
EOF
)"
```

---

### Task 2: Surface silent discovery errors in debug mode

`internal/hooks/validate.go:148-151` discards errors from `DiscoverCommand()`. When debug mode is on, these should be logged to stderr.

**Files:**
- Modify: `internal/hooks/validate.go:138-155`
- Test: `internal/hooks/validate_test.go` (existing file)

**Step 1: Write the failing test**

Add to `internal/hooks/validate_test.go` a test that verifies discovery errors are logged in debug mode. The test should use the existing `ParallelValidateExecutor` test infrastructure.

Create a test that:
1. Sets up a `ParallelValidateExecutor` with `debug: true`
2. Mocks `DiscoverCommand` to return an error
3. Asserts the error appears on stderr

Since `discoverCommands` is a private method and the executor writes to `os.Stderr`, the most practical approach is a test via the public `RunValidateHook` entry point or by checking the new behavior through the existing test helpers. However, the `ParallelValidateExecutor` uses `pve.debug` but currently has no stderr writer. We need to add an `io.Writer` for debug output.

Actually, looking at the design doc more closely, the fix is simpler — just add `fmt.Fprintf(os.Stderr, ...)` conditionally. This doesn't need a new test since it's debug-only logging that goes to `os.Stderr` directly. But to test it properly, we should route debug output through the dependency-injected stderr.

The cleanest approach: add a `stderr` field to `ParallelValidateExecutor` and use it for debug logging.

```go
// In validate_test.go, add this test:
func TestDiscoverCommandsLogsErrorsInDebugMode(t *testing.T) {
    testDeps := hooks.CreateTestDependencies()

    // Make discovery fail
    testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, _ string, _ ...string) (*hooks.CommandOutput, error) {
        return nil, errors.New("command not found")
    }
    testDeps.MockRunner.LookPathFunc = func(_ string) (string, error) {
        return "", errors.New("not found")
    }
    testDeps.MockFS.StatFunc = func(_ string) (os.FileInfo, error) {
        return nil, os.ErrNotExist
    }

    skipConfig := &hooks.SkipConfig{SkipLint: false, SkipTest: false}
    pve := hooks.NewParallelValidateExecutor("/project", 20, true, skipConfig, testDeps.Dependencies)

    result, err := pve.ExecuteValidations(context.Background(), "/project", "/project")
    require.NoError(t, err)

    // With no commands found, both passed (no commands = nothing to fail)
    assert.True(t, result.BothPassed)

    // Debug output should mention discovery errors
    stderrOutput := testDeps.MockStderr.String()
    assert.Contains(t, stderrOutput, "discovery error")
}
```

Run: `gotestsum --format pkgname -- -tags=testmode -run TestDiscoverCommandsLogsErrorsInDebugMode ./internal/hooks/...`
Expected: FAIL — no debug output appears yet.

**Step 2: Implement the fix**

Modify `internal/hooks/validate.go`. Add a `stderr` field to `ParallelValidateExecutor` and update `discoverCommands`:

```go
// ParallelValidateExecutor — add stderr field
type ParallelValidateExecutor struct {
    discovery  *CommandDiscovery
    executor   *CommandExecutor
    timeout    int
    debug      bool
    skipConfig *SkipConfig
    stderr     io.Writer  // add this
}
```

Update `NewParallelValidateExecutor` to accept stderr from deps:

```go
return &ParallelValidateExecutor{
    discovery:  NewCommandDiscovery(projectRoot, timeout, deps),
    executor:   NewCommandExecutor(timeout, debug, deps),
    timeout:    timeout,
    debug:      debug,
    skipConfig: skipConfig,
    stderr:     deps.Stderr,  // add this
}
```

Update `discoverCommands` at lines 147-151:

```go
var lintCmd, testCmd *DiscoveredCommand
if !skipLint {
    var err error
    lintCmd, err = pve.discovery.DiscoverCommand(ctx, CommandTypeLint, fileDir)
    if err != nil && pve.debug {
        _, _ = fmt.Fprintf(pve.stderr, "Lint discovery error: %v\n", err)
    }
}
if !skipTest {
    var err error
    testCmd, err = pve.discovery.DiscoverCommand(ctx, CommandTypeTest, fileDir)
    if err != nil && pve.debug {
        _, _ = fmt.Fprintf(pve.stderr, "Test discovery error: %v\n", err)
    }
}
```

Add `"io"` to the import block if not already present.

**Step 3: Run test to verify it passes**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestDiscoverCommandsLogsErrorsInDebugMode ./internal/hooks/...`
Expected: PASS

**Step 4: Run full suite**

Run: `task check`
Expected: All checks pass.

**Step 5: Commit**

```bash
git add internal/hooks/validate.go internal/hooks/validate_test.go
git commit -m "$(cat <<'EOF'
fix: surface command discovery errors in debug mode

DiscoverCommand errors were silently discarded. Now log them to stderr
when debug mode is enabled. No behavioral change for non-debug mode.
EOF
)"
```

---

### Task 3: Add `Rename` and `Remove` to `RegistryFS` interface

`internal/skipregistry/storage.go:84-86` calls `os.Rename` and `os.Remove` directly, bypassing the `RegistryFS` abstraction. This breaks testability.

**Files:**
- Modify: `internal/shared/fs.go:19-25` (RegistryFS interface)
- Modify: `internal/skipregistry/storage.go:84-86`
- Regenerate: `internal/shared/mocks/RegistryFS.go`
- Test: `internal/skipregistry/storage_test.go`

**Step 1: Write the failing test**

Add a test in `internal/skipregistry/storage_test.go` that verifies `Save()` calls `Rename` and `Remove` through the interface rather than directly via `os`:

```go
func TestJSONStorageSaveUsesInterfaceMethods(t *testing.T) {
    mockFS := mocks.NewMockRegistryFS(t)
    storage := skipregistry.NewJSONStorage(mockFS, "/tmp/test-registry.json")

    // MkdirAll for the directory
    mockFS.EXPECT().MkdirAll("/tmp", mock.Anything).Return(nil).Once()
    // WriteFile for the temp file
    mockFS.EXPECT().WriteFile("/tmp/test-registry.json.tmp", mock.Anything, mock.Anything).Return(nil).Once()
    // Rename from temp to actual
    mockFS.EXPECT().Rename("/tmp/test-registry.json.tmp", "/tmp/test-registry.json").Return(nil).Once()

    err := storage.Save(context.Background(), skipregistry.RegistryData{})
    require.NoError(t, err)
}
```

Run: `gotestsum --format pkgname -- -tags=testmode -run TestJSONStorageSaveUsesInterfaceMethods ./internal/skipregistry/...`
Expected: FAIL — `Rename` method doesn't exist on `RegistryFS` yet.

**Step 2: Extend the interface**

In `internal/shared/fs.go`, add `Rename` and `Remove` to `RegistryFS`:

```go
type RegistryFS interface {
    ReadFile(name string) ([]byte, error)
    WriteFile(name string, data []byte, perm os.FileMode) error
    MkdirAll(path string, perm os.FileMode) error
    UserHomeDir() (string, error)
    Rename(oldpath, newpath string) error
    Remove(name string) error
}
```

Add implementations to `RealFS` (it already has `Remove` for `HooksFS`, but we need `Rename`):

```go
func (r *RealFS) Rename(oldpath, newpath string) error {
    if err := os.Rename(oldpath, newpath); err != nil {
        return fmt.Errorf("rename %s to %s: %w", oldpath, newpath, err)
    }
    return nil
}
```

**Step 3: Regenerate mocks**

Run: `task mocks`
Expected: `internal/shared/mocks/RegistryFS.go` is regenerated with `Rename` and `Remove` methods.

**Step 4: Update `JSONStorage.Save()` to use interface**

In `internal/skipregistry/storage.go`, replace lines 84-86:

```go
// Before:
if renameErr := os.Rename(tempFile, s.filePath); renameErr != nil {
    _ = os.Remove(tempFile)

// After:
if renameErr := s.fs.Rename(tempFile, s.filePath); renameErr != nil {
    _ = s.fs.Remove(tempFile)
```

**Step 5: Run the test**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestJSONStorageSaveUsesInterfaceMethods ./internal/skipregistry/...`
Expected: PASS

**Step 6: Run full suite**

Run: `task check`
Expected: All checks pass.

**Step 7: Commit**

```bash
git add internal/shared/fs.go internal/shared/mocks/RegistryFS.go internal/skipregistry/storage.go internal/skipregistry/storage_test.go
git commit -m "$(cat <<'EOF'
refactor: route Rename/Remove through RegistryFS interface

JSONStorage.Save() called os.Rename and os.Remove directly, bypassing
the RegistryFS abstraction. Add Rename and Remove methods to the
interface, implement on RealFS, and update storage to use them.
Regenerate mocks.
EOF
)"
```

---

### Task 4: Consolidate config locations to `~/.config/cc-tools/`

Skip registry lives in `~/.claude/skip-registry.json` and debug config lives in `~/.claude/debug-config.json`. Both should move to `~/.config/cc-tools/` (respecting `$XDG_CONFIG_HOME`).

**Files:**
- Create: `internal/shared/configdir.go`
- Modify: `internal/skipregistry/registry.go:291-305`
- Modify: `internal/debug/config.go:194-200`
- Test: `internal/shared/configdir_test.go`

**Step 1: Write the failing test for ConfigDir**

Create `internal/shared/configdir_test.go`:

```go
package shared_test

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/riddopic/cc-tools/internal/shared"
)

func TestConfigDir(t *testing.T) {
    t.Run("returns XDG_CONFIG_HOME when set", func(t *testing.T) {
        t.Setenv("XDG_CONFIG_HOME", "/custom/config")
        got := shared.ConfigDir()
        assert.Equal(t, filepath.Join("/custom/config", "cc-tools"), got)
    })

    t.Run("defaults to ~/.config/cc-tools", func(t *testing.T) {
        t.Setenv("XDG_CONFIG_HOME", "")
        home, err := os.UserHomeDir()
        require.NoError(t, err)
        got := shared.ConfigDir()
        assert.Equal(t, filepath.Join(home, ".config", "cc-tools"), got)
    })
}
```

Run: `gotestsum --format pkgname -- -tags=testmode -run TestConfigDir ./internal/shared/...`
Expected: FAIL — `ConfigDir` doesn't exist.

**Step 2: Implement ConfigDir**

Create `internal/shared/configdir.go`:

```go
package shared

import (
    "os"
    "path/filepath"
)

// ConfigDir returns the cc-tools configuration directory.
// Respects $XDG_CONFIG_HOME; defaults to ~/.config/cc-tools.
func ConfigDir() string {
    if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
        return filepath.Join(xdg, "cc-tools")
    }
    home, err := os.UserHomeDir()
    if err != nil {
        return filepath.Join("/tmp", ".config", "cc-tools")
    }
    return filepath.Join(home, ".config", "cc-tools")
}
```

**Step 3: Run test to verify it passes**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestConfigDir ./internal/shared/...`
Expected: PASS

**Step 4: Update `skipregistry/registry.go` to use ConfigDir**

Replace `getClaudeDir()` and `getRegistryPath()` at lines 291-305:

```go
func getRegistryPath() string {
    return filepath.Join(shared.ConfigDir(), "skip-registry.json")
}
```

Delete the `getClaudeDir()` function (lines 296-305). Add a one-time migration function:

```go
// migrateRegistryIfNeeded copies skip-registry.json from ~/.claude/ to the
// new config dir if the old file exists and the new one does not.
func migrateRegistryIfNeeded() {
    newPath := getRegistryPath()
    if _, err := os.Stat(newPath); err == nil {
        return // new file already exists
    }

    home, err := os.UserHomeDir()
    if err != nil {
        return
    }
    oldPath := filepath.Join(home, ".claude", "skip-registry.json")
    data, err := os.ReadFile(oldPath)
    if err != nil {
        return // old file doesn't exist or unreadable
    }

    dir := filepath.Dir(newPath)
    _ = os.MkdirAll(dir, 0o750)
    _ = os.WriteFile(newPath, data, 0o644)
}
```

Call `migrateRegistryIfNeeded()` from `DefaultStorage()`:

```go
func DefaultStorage() *JSONStorage {
    migrateRegistryIfNeeded()
    return NewJSONStorage(&shared.RealFS{}, getRegistryPath())
}
```

Remove the `"os"` import if `getClaudeDir()` was the only user of it (it's not — `os.PathSeparator` is used elsewhere, but check).

**Step 5: Update `debug/config.go` to use ConfigDir**

Replace `getConfigDir()` at lines 194-200:

```go
func getConfigDir() string {
    return shared.ConfigDir()
}
```

Add the same migration pattern:

```go
func init() {
    migrateDebugConfigIfNeeded()
}

func migrateDebugConfigIfNeeded() {
    newPath := filepath.Join(shared.ConfigDir(), "debug-config.json")
    if _, err := os.Stat(newPath); err == nil {
        return
    }

    home, err := os.UserHomeDir()
    if err != nil {
        return
    }
    oldPath := filepath.Join(home, ".claude", "debug-config.json")
    data, err := os.ReadFile(oldPath)
    if err != nil {
        return
    }

    dir := filepath.Dir(newPath)
    _ = os.MkdirAll(dir, 0o750)
    _ = os.WriteFile(newPath, data, 0o644)
}
```

**Step 6: Run full suite**

Run: `task check`
Expected: All checks pass.

**Step 7: Commit**

```bash
git add internal/shared/configdir.go internal/shared/configdir_test.go \
        internal/skipregistry/registry.go internal/debug/config.go
git commit -m "$(cat <<'EOF'
refactor: consolidate config locations to ~/.config/cc-tools/

Move skip registry and debug config from ~/.claude/ to
~/.config/cc-tools/ (respecting $XDG_CONFIG_HOME). Add one-time
migration that copies old files to new location on first access.
EOF
)"
```

---

### Task 5: Remove dead code from `executor.go`

`RunSmartHook`, `ExecuteForHook`, and their helpers have no production callers — only the parallel `ValidateWithSkipCheck` path is used.

**Files:**
- Modify: `internal/hooks/executor.go` — delete functions
- Modify: `internal/hooks/executor_test.go` — delete dead tests
- Modify: `internal/hooks/additional_test.go` — delete dead tests

**Step 1: Identify what to delete**

**Delete from `executor.go`:**
- `ExecuteForHook()` (lines 116-162) — method on CommandExecutor
- `initLogger()` (lines 164-169)
- `logHookStart()` (lines 171-181)
- `processHookInput()` (lines 183-212)
- `RunSmartHook()` (lines 214-278)
- `discoverCommand()` (lines 336-376) — package-level function
- `executeCommand()` (lines 378-404) — package-level function
- `discoverAndExecute()` (lines 406-430)

**Keep in `executor.go`:**
- `ExitCodeShowMessage` constant (line 19)
- `ExecutorResult` struct (lines 22-30)
- `CommandExecutor` struct (lines 32-37)
- `NewCommandExecutor()` (lines 39-49)
- `Execute()` method (lines 51-113)
- `handleInputError()` (lines 280-286)
- `validateHookEvent()` (lines 288-307)
- `acquireLock()` (lines 309-334)

**Delete from `executor_test.go`:**
- `TestRunSmartHook` (lines 142-289) — tests the deleted `RunSmartHook`
- `TestCommandExecutor` subtests for `ExecuteForHook` (lines 328-355) — tests the deleted method

**Delete from `additional_test.go`:**
- `TestExecutorEdgeCases` subtests for `ExecuteForHook` (lines 31-109)
- `TestRunSmartHookEdgeCases` (lines 257-413)

**Keep in `executor_test.go`:**
- Helper functions (`assertExitCode`, `assertStderrContains`, etc.)
- `setupEditInput`, `setupGitProjectFS`, `setupGitMakefileFS`, `setupLockAvailable`
- `TestCommandExecutor` subtests for `Execute` (lines 293-326)
- `TestValidateHookEvent` (lines 382-444)

**Keep in `additional_test.go`:**
- `TestExecutorEdgeCases` subtest "Execute with nil command" (lines 18-29)
- `TestDiscoveryEdgeCases` (lines 112-255)
- `TestHandleInputError` (lines 415-449)
- `TestLockManagerCleanupOnExit` (lines 452-494)

**Step 2: Also check if `ExecuteForHook` is called from outside executor.go**

The method `ExecuteForHook` is called only from `executeCommand()` (line 394) which is itself dead code. The package-level `discoverAndExecute()` calls both. All three are only called from `RunSmartHook`. Confirmed safe to delete.

**Step 3: Delete the dead code**

Remove the functions listed above from `executor.go`. Remove corresponding tests from `executor_test.go` and `additional_test.go`.

After deletion, check if any imports in `executor.go` become unused:
- `debuglog "github.com/riddopic/cc-tools/internal/debug"` — only used by `initLogger()` → remove
- `"github.com/riddopic/cc-tools/internal/output"` — only used by `ExecuteForHook()` → remove
- `"os"` — only used by `initLogger()` → remove
- `"path/filepath"` — used by `RunSmartHook()` and `acquireLock`/`validateHookEvent` → check if still needed after deletion (it is used in `validateHookEvent` indirectly... actually no, `validateHookEvent` doesn't use filepath. But `acquireLock` is kept and doesn't use filepath either.) Check more carefully after deletion.
- `"time"` — used by `CommandExecutor.timeout` field → keep

Also in `executor_test.go`, check if `makeLintDryRunRunner` and `newTestDiscoveredCommand` helpers are still used after removing RunSmartHook tests. If not, remove them too.

And check `additional_test.go` — `TestExecutorEdgeCases` will only have the "Execute with nil command" subtest left. The `ExecuteForHook` subtests need to be removed.

**Step 4: Run tests**

Run: `task check`
Expected: All checks pass. Dead code is gone, remaining tests still pass.

**Step 5: Commit**

```bash
git add internal/hooks/executor.go internal/hooks/executor_test.go internal/hooks/additional_test.go
git commit -m "$(cat <<'EOF'
refactor: remove dead code from executor.go

Delete RunSmartHook, ExecuteForHook, and all their helpers (initLogger,
logHookStart, processHookInput, discoverCommand, executeCommand,
discoverAndExecute). No production callers exist — only the parallel
ValidateWithSkipCheck path is used. Remove corresponding tests.
EOF
)"
```

---

### Task 6: Documentation and naming cleanup

Two sub-fixes: CLAUDE.md accuracy and debug log naming unification.

**Files:**
- Modify: `CLAUDE.md:46`
- Modify: `internal/debug/config.go:179-192`

**Step 6a: Fix CLAUDE.md**

In `CLAUDE.md` line 46, change:

```
| `internal/config` | YAML/JSON config persistence at `~/.config/cc-tools/config.yaml` |
```

to:

```
| `internal/config` | JSON config persistence at `~/.config/cc-tools/config.json` |
```

**Step 6b: Write the failing test for unified naming**

The two functions that generate debug log paths are:
1. `shared.GetDebugLogPathForDir()` — returns `/tmp/cc-tools-{name}-{hash}.debug`
2. `debug.GetLogFilePath()` — returns `/tmp/cc-tools-validate-{name}-{hash}.log`

Unify by making `GetLogFilePath` delegate to `GetDebugLogPathForDir`:

```go
// In debug/config.go, replace GetLogFilePath:
func GetLogFilePath(dir string) string {
    return shared.GetDebugLogPathForDir(dir)
}
```

Add a test in an existing test file (or a new `internal/debug/config_test.go` if one doesn't exist) that verifies the naming convention matches:

```go
func TestGetLogFilePathUsesSharedNaming(t *testing.T) {
    dir := "/some/project"
    logPath := debug.GetLogFilePath(dir)
    sharedPath := shared.GetDebugLogPathForDir(dir)
    assert.Equal(t, sharedPath, logPath, "GetLogFilePath should delegate to shared.GetDebugLogPathForDir")
    assert.True(t, strings.HasSuffix(logPath, ".debug"), "should use .debug extension")
    assert.False(t, strings.Contains(logPath, ".log"), "should not use .log extension")
}
```

Run: `gotestsum --format pkgname -- -tags=testmode -run TestGetLogFilePathUsesSharedNaming ./internal/debug/...`
Expected: FAIL — currently returns `.log` suffix.

**Step 6c: Implement the fix**

Replace `GetLogFilePath` in `internal/debug/config.go` (lines 179-192):

```go
// GetLogFilePath generates a log file path for a directory.
// Delegates to shared.GetDebugLogPathForDir for consistent naming.
func GetLogFilePath(dir string) string {
    return shared.GetDebugLogPathForDir(dir)
}
```

Remove the now-unused imports from the function body: `"crypto/sha256"`, `"encoding/hex"`, `"strings"` — but only if no other functions in the file use them (they are — `IsEnabled` uses `strings`, `Save` uses `sha256`... actually checking: `strings` is used in `IsEnabled` at line 154. `crypto/sha256` and `encoding/hex` are only used in `GetLogFilePath`. So remove `"crypto/sha256"` and `"encoding/hex"` from imports.

**Step 6d: Run tests**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestGetLogFilePathUsesSharedNaming ./internal/debug/...`
Expected: PASS

**Step 6e: Run full suite**

Run: `task check`
Expected: All checks pass.

**Step 6f: Commit**

```bash
git add CLAUDE.md internal/debug/config.go internal/debug/config_test.go
git commit -m "$(cat <<'EOF'
fix: correct CLAUDE.md config description and unify debug log naming

Change "YAML/JSON config persistence" to "JSON config persistence" in
CLAUDE.md. Unify debug log path generation by delegating GetLogFilePath
to shared.GetDebugLogPathForDir, eliminating the duplicate .log naming
pattern.
EOF
)"
```

---

## Final Verification

After all six tasks are complete:

Run: `task check && task test-race`
Expected: All formatting, linting, tests, and race detection pass.

Verify no temp files leak:
```bash
ls /tmp/cc-tools-stdin-* 2>/dev/null && echo "LEAK FOUND" || echo "Clean"
```

Verify config paths are consolidated:
```bash
grep -r '\.claude' internal/ --include='*.go' | grep -v test | grep -v mock | grep -v _test
```
Expected: No references to `~/.claude/` for skip-registry or debug-config.
