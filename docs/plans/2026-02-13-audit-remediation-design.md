# Audit Remediation Design

**Date:** 2026-02-13
**Source:** `docs/audits/context-audit-20260213-213500.md`
**Approach:** Incremental fixes — six independent, testable changes.

## Fix 1: Temp file leak in `debugLog()`

`cmd/cc-tools/main.go:154` creates a temp file per `validate` invocation, never deleted.

Add a package-level `stdinTempFile` variable set by `debugLog()`. Clean up after `runValidate()` returns:

```go
var stdinTempFile string

case "validate":
    runValidate()
    if stdinTempFile != "" {
        _ = os.Remove(stdinTempFile)
    }
```

## Fix 2: Silent discovery errors

`validate.go:148-152` discards `DiscoverCommand` errors. Surface them when debug mode is enabled. No behavioral change for non-debug mode.

```go
if !skipLint {
    lintCmd, err = pve.discovery.DiscoverCommand(ctx, CommandTypeLint, fileDir)
    if err != nil && pve.debug {
        _, _ = fmt.Fprintf(os.Stderr, "Lint discovery error: %v\n", err)
    }
}
```

Same for test discovery.

## Fix 3: `os.Rename` escapes `RegistryFS` interface

`skipregistry/storage.go:84` calls `os.Rename` directly. Add `Rename` and `Remove` methods to `shared.RegistryFS`:

```go
type RegistryFS interface {
    ReadFile(name string) ([]byte, error)
    WriteFile(name string, data []byte, perm os.FileMode) error
    MkdirAll(path string, perm os.FileMode) error
    UserHomeDir() (string, error)
    Rename(oldpath, newpath string) error  // new
    Remove(name string) error              // new
}
```

Implement on `RealFS`. Replace direct `os.Rename` and `os.Remove` in `JSONStorage.Save()`. Regenerate mocks via `task mocks`.

## Fix 4: Config location consolidation

Move skip registry and debug config to `~/.config/cc-tools/` (respecting `$XDG_CONFIG_HOME`):

| State | Before | After |
|-------|--------|-------|
| App config | `~/.config/cc-tools/config.json` | unchanged |
| Skip registry | `~/.claude/skip-registry.json` | `~/.config/cc-tools/skip-registry.json` |
| Debug config | `~/.claude/debug-config.json` | `~/.config/cc-tools/debug-config.json` |

Changes:
1. Extract shared `ConfigDir() string` helper into `internal/shared/`.
2. Update `skipregistry/registry.go` `getRegistryPath()` to use it.
3. Update `debug/config.go` `getConfigDir()` to use it.
4. Add one-time migration: if old file exists at `~/.claude/` and no file at new location, copy it. Leave old file in place.

## Fix 5: Remove dead code

Delete `RunSmartHook`, `ExecuteForHook`, and all their helpers and tests. No production callers exist — only the parallel `ValidateWithSkipCheck` path is used.

**Delete from `executor.go`:**
- `ExecuteForHook()`
- `initLogger()`, `logHookStart()`, `processHookInput()`
- `RunSmartHook()`
- `discoverCommand()` (package-level function, not the method)
- `executeCommand()` (package-level function)
- `discoverAndExecute()`

**Keep:**
- `CommandExecutor` struct and `Execute()` method
- `ExitCodeShowMessage` constant
- `handleInputError()`, `validateHookEvent()`, `acquireLock()`

**Delete tests** for removed functions in `executor_test.go` and `additional_test.go`.

## Fix 6: Documentation and naming cleanup

**6a:** CLAUDE.md — change "YAML/JSON config persistence" to "JSON config persistence."

**6b:** Unify debug log naming. Change `debug.GetLogFilePath()` to delegate to `shared.GetDebugLogPathForDir()`. Single naming pattern: `cc-tools-{name}-{hash}.debug`. The `.log` variant disappears.
