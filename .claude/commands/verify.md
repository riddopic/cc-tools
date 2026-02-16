---
description: Verification Command
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash
  - Task
  - TaskCreate
  - TaskUpdate
  - TaskList
skills:
  - go-coding-standards
  - code-review
  - testing-patterns
---

# Verify Codebase

Run comprehensive verification on current codebase state following Go and project standards.

## Arguments

`$ARGUMENTS` can be:

- `quick` - Only build + format check (fastest)
- `full` - All checks including race detection (default)
- `pre-commit` - Checks relevant for commits (fmt, lint, test)
- `pre-pr` - Full checks plus security scan and coverage

## Execution Steps

Execute verification in this exact order. Stop on critical failures.

### 1. Build Check

```bash
task build
```

- If build fails, report errors and **STOP**
- Build must pass before other checks

### 2. Format Check

```bash
# Check formatting without modifying files
gofmt -l .
goimports -l .
```

- Report any files needing formatting
- For `quick` mode, stop here if passing

### 3. Lint Check

```bash
task lint
```

- Run golangci-lint with project configuration
- Report all warnings and errors with file:line
- Critical: Any errors block PR readiness

### 4. Unit Tests

```bash
task test
```

- Run fast unit tests (-short flag, 30s timeout)
- Report pass/fail count
- Must pass for `pre-commit` and above

### 5. Race Detection (full/pre-pr only)

```bash
task test-race
```

- Run tests with Go race detector enabled
- Critical for concurrent code correctness
- Report any data races detected

### 6. Test Coverage (pre-pr only)

```bash
task coverage
```

- Generate coverage report
- Target: ≥80% for new code
- Report coverage percentage per package

### 7. Security Scan (pre-pr only)

```bash
# Note: vulncheck is no longer a separate task target.
# Use `task lint` which includes security-related linting checks.
task lint
```

- Check for known vulnerabilities in dependencies
- Report any CVEs found

### 8. Debug Statement Audit

```bash
# Search for debug prints in source files (excluding tests and vendor)
rg "fmt\.Print|log\.Print|zap\.(Debug|Info)" --type go -g '!*_test.go' -g '!vendor/*' -l
```

- Report locations of potential debug statements
- Exclude legitimate logging in production code

### 9. Git Status

```bash
git status --short
git diff --stat HEAD
```

- Show uncommitted changes
- Show files modified since last commit

## Output Format

Produce a concise verification report:

```
VERIFICATION: [PASS/FAIL]

Mode:     [quick/full/pre-commit/pre-pr]
Build:    [OK/FAIL]
Format:   [OK/X files need formatting]
Lint:     [OK/X warnings, Y errors]
Tests:    [X/Y passed]
Race:     [OK/X races detected] (if run)
Coverage: [X%] (if run)
Security: [OK/X vulnerabilities] (if run)
Debug:    [OK/X statements found]

Ready for PR: [YES/NO]
```

## Critical Issues

If any critical issues found, list them with fix suggestions:

```
CRITICAL ISSUES:

1. Build failure in cmd/analyze.go:42
   → Fix: Missing import "context"

2. Race condition in internal/agent/orchestrator.go
   → Fix: Add mutex protection for shared state

SUGGESTED COMMANDS:
  task fmt      # Fix formatting
  task lint     # Re-check after fixes
  task test     # Verify tests pass
```

## Quick Reference

| Mode | Checks Run |
|------|------------|
| `quick` | build, format |
| `pre-commit` | build, format, lint, test |
| `full` | build, format, lint, test, race |
| `pre-pr` | build, format, lint, test, race, coverage, security |

## Integration

This command integrates with the project's standard tooling:

- Uses `task` targets defined in Taskfile.yml
- Follows `docs/CODING_GUIDELINES.md` standards
- Compatible with CI pipeline checks
- Respects `.golangci.yml` lint configuration
- Implements the `verification-before-completion` skill -- run it before any completion claims
