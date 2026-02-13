# Go File Decomposition Pattern

**Extracted:** 2026-02-10
**Context:** Splitting large Go files into multiple files within the same package

## Problem

Large Go files (1000+ lines) need decomposition into focused files while maintaining compilation, tests, and lint compliance.

## Solution

### Step 1: Plan the split by function clusters

Group functions by responsibility (command setup, config, deps/factories, execution, output).

### Step 2: Create new files with `package cmd` only

Don't manually write imports — let `goimports` handle them.

### Step 3: Run goimports with local prefix

```bash
goimports -local github.com/riddopic/quanta -w cmd/file1.go cmd/file2.go ...
```

The `-local` flag is **required** to produce 3 import groups (stdlib, third-party, internal) that satisfy the `golangci-lint` `goimports` linter with `local-prefixes` config.

### Step 4: Update `.golangci.yml` exclusion rules

Lint exclusions reference specific file paths. When functions move to new files, update path patterns:

```yaml
# Before (single file)
- path: 'cmd/analyze\.go'
  linters: [ireturn]

# After (multiple files)
- path: 'cmd/analyze(_\w+)?\.go'
  linters: [ireturn]
```

### Step 5: Fix test assertions

If the decomposition changed error messages (e.g., new config constructor), update test `assert.Contains` strings to match.

### Step 6: Remove orphaned `//nolint` directives

When functions move to a new file, their `//nolint` comments may remain in the original file above unrelated functions. The `nolintlint` linter catches these.

## When to Use

- Go file exceeds 1000 lines
- Functions cluster into clear responsibility groups
- File has multiple `//nolint:funlen` or `//nolint:gocognit` pragmas
