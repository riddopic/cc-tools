---
description: Safely identify and remove dead code with test verification
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash
  - Edit
  - Task
  - TaskCreate
  - TaskUpdate
  - TaskList
  - AskUserQuestion
skills:
  - go-coding-standards
  - code-review
  - testing-patterns
---

# Refactor Clean

Safely identify and remove dead code from the Go codebase with comprehensive test verification. This command follows the project's coding guidelines and ensures all changes are validated before and after each modification.

**ULTRA THINK**

## Pre-Refactor Validation

**Run these commands FIRST to establish a baseline:**

```bash
# Ensure tests pass before starting
task test

# Run tests with race detector
task test-race

# Verify linting passes
task lint

# Check current coverage
task coverage
```

## Dead Code Detection

### Go-Specific Analysis Tools

```bash
# Find unused functions/variables (U1000)
staticcheck -checks="U1000" ./...

# Find unreachable code (SA9003)
staticcheck -checks="SA9003" ./...

# Find unused dependencies
go mod why -m all | grep "# "

# Check for empty Go files
find . -type f -name "*.go" -size 0

# Find ignored errors (potential dead code indicators)
rg "_ =" --type go -g '!*_test.go'

# Find commented-out code (excluding legitimate comments)
rg "^[[:space:]]*//.*" --type go | grep -v "TODO\|FIXME\|NOTE\|Copyright\|Package\|nolint"

# Find unused struct fields (manual inspection needed)
staticcheck -checks="U1000" ./... 2>&1 | grep "field"

# Find orphaned test files (tests for deleted code)
for f in $(find . -name "*_test.go"); do
    base=$(basename "$f" _test.go)
    dir=$(dirname "$f")
    if [ ! -f "$dir/$base.go" ] && [ ! -f "$dir/${base}s.go" ]; then
        echo "Orphaned test: $f"
    fi
done
```

### Additional Checks

```bash
# Run comprehensive linting to find more dead code
task lint

# Check for unused imports (auto-fixed by goimports)
goimports -l .

# Find TODO/FIXME markers that might indicate incomplete removal
grep -rn "TODO\|FIXME" --include="*.go" . | grep -i "remove\|delete\|dead\|unused"
```

## Generate Analysis Report

Create comprehensive report in `.reports/dead-code-analysis.md`:

```markdown
# Dead Code Analysis Report

Generated: [DATE]
Baseline Tests: PASS/FAIL
Baseline Coverage: X%

## Summary

| Category | Count | Status |
|----------|-------|--------|
| Unused Functions | X | Pending |
| Unused Variables | X | Pending |
| Unreachable Code | X | Pending |
| Commented Code | X | Pending |
| Empty Files | X | Pending |
| Orphaned Tests | X | Pending |

## Findings by Severity

### 🟢 SAFE (Auto-Delete Candidates)
- Test helpers only used in deleted tests
- Unused internal utilities
- Backup/temporary files

### 🟡 CAUTION (Manual Review Required)
- Functions referenced via interface (reflection)
- Code guarded by build tags
- Generated code (mocks, protobuf)

### 🔴 DANGER (Do Not Delete Without Investigation)
- Public API functions (exported)
- Configuration handlers
- Main entry points
- Cobra command implementations

## Detailed Findings

[List each finding with file:line reference]
```

## Categorization Guidelines

### 🟢 SAFE to Delete

- Unused private functions (`lowercase`)
- Unused local variables
- Unreachable code blocks
- Empty files
- Commented-out code (not TODO/FIXME)
- Orphaned test files

### 🟡 CAUTION - Verify First

- Unused exported functions (check external usage)
- Code in `internal/` packages (may have interface-based usage)
- Functions matching interface signatures
- Code with `//go:generate` directives

### 🔴 DANGER - Do Not Auto-Delete

- Cobra commands (`cmd/`)
- Interface implementations
- Functions registered via reflection
- Build-tag conditional code (`//go:build`)
- Plugin implementations

## Safe Deletion Workflow

**🚨 CRITICAL: Follow this workflow for EVERY deletion!**

For each identified dead code item:

1. **Document the Finding**
   ```bash
   echo "Removing: [file:line] - [reason]" >> .reports/deletion-log.md
   ```

2. **Run Full Test Suite BEFORE Deletion**
   ```bash
   task test-race
   ```

3. **Verify Tests Pass**
   - If tests fail, investigate FIRST
   - Do NOT proceed with deletion if tests fail

4. **Apply the Change**
   - Delete the dead code
   - Run `task fmt` to fix imports

5. **Re-Run Tests AFTER Deletion**
   ```bash
   task test-race
   ```

6. **Verify and Commit or Rollback**
   - If tests pass: Continue to next item
   - If tests fail: `git checkout -- [file]` and investigate

7. **Run Linting**
   ```bash
   task lint
   ```

## Final Verification

Apply `verification-before-completion` before claiming refactoring is complete.

After all deletions are complete:

```bash
# Full validation suite
task check

# Verify coverage hasn't dropped significantly
task coverage

# Run race detector
task test-race

# Final lint check
task lint
```

## Summary Report

Generate final summary:

```markdown
# Refactor Clean Summary

## Changes Made

| Action | File | Reason |
|--------|------|--------|
| Deleted | path/file.go:123 | Unused function |
| Deleted | path/file.go:456 | Unreachable code |
| Kept | path/file.go:789 | Used via interface |

## Metrics

- Files Modified: X
- Lines Removed: X
- Test Status: PASS
- Coverage Before: X%
- Coverage After: X%

## Verification

- [ ] All tests pass (`task test`)
- [ ] Race detector passes (`task test-race`)
- [ ] Linting passes (`task lint`)
- [ ] Pre-commit passes (`task pre-commit`)
```

## Important Guidelines

**🚨 CRITICAL — TEST BEHAVIOR, NOT IMPLEMENTATION!** Tests should remain valid even if the implementation changes completely. Only delete dead code, not code that tests depend on.

**IMPORTANT**: Do NOT delete code without running `task test-race` before AND after each deletion.

**CRITICAL**: Always follow the patterns in `docs/CODING_GUIDELINES.md` and `docs/examples/` when evaluating what constitutes "dead code".

**Reference Documentation:**
- `docs/CODING_GUIDELINES.md` - Coding standards
- `docs/examples/patterns/testing.md` - Test patterns
- `docs/examples/standards/go-specific.md` - Go idioms
