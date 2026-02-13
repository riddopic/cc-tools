# Go Package Rename Strategy

**Extracted:** 2026-02-11
**Context:** Bulk renaming a Go package across a large codebase (e.g., `pkg/dexutils` → `internal/dex/pricing`)

## Problem

When renaming a Go package, a naive find-and-replace of `oldpkg.` → `newpkg.` corrupts string literals that happen to contain the old package name but are NOT Go type qualifiers:

- Viper config keys: `"dexutils.price_source"` (should stay as-is)
- Solidity function patterns: `"dexutils.swaptokentobase"` (should stay as-is)
- Log messages, comments, or docs referencing the old name

Additionally, macOS BSD `sed` does NOT support `\b` word boundaries — `sed -i '' 's/\bdexutils\./pricing./g'` silently matches nothing, leaving files unchanged with no error.

## Solution

### Step 1: Move files with git mv

```bash
mkdir -p internal/dex/pricing
for f in pkg/dexutils/*.go; do git mv "$f" internal/dex/pricing/; done
```

Always use glob patterns — don't assume filenames (e.g., `chainlink_pricer_internal_test.go` might actually be `chainlink_internal_test.go`).

### Step 2: Update package declarations

```bash
# In the moved files only
sed -i '' 's/^package dexutils/package pricing/' internal/dex/pricing/*.go
sed -i '' 's/^package dexutils_test/package pricing_test/' internal/dex/pricing/*_test.go
```

### Step 3: Update import paths (safe — these are always in Go import blocks)

```bash
# All Go files that import the old path
rg -l 'pkg/dexutils' --type go | xargs sed -i '' 's|pkg/dexutils|internal/dex/pricing|g'
```

### Step 4: Update type qualifiers (targeted — only files that import the package)

Build a targeted file list of ACTUAL importers, then replace only in those files:

```bash
# Files that import the new path (these are the ones with Go type qualifiers)
rg -l 'internal/dex/pricing' --type go | xargs sed -i '' 's/dexutils\./pricing./g'
```

Do NOT run this replacement globally — it would corrupt string literals in files that don't import the package but contain the string for other reasons.

### Step 5: Verify and format

```bash
make fmt   # gofmt + goimports fixes any import ordering
make lint  # catch anything missed
make test  # confirm nothing broke
```

## Key Rules

1. **Import path changes are always safe** — they're inside `import (...)` blocks
2. **Type qualifier changes require targeting** — only replace in files that actually import the package
3. **Never use `\b` in sed on macOS** — BSD sed doesn't support it; use targeted file lists instead
4. **Don't forget config files** — `.golangci.yml`, `.mockery.yml`, `.vscode/settings.json` may reference old paths
5. **Always glob for filenames** — don't hardcode assumed filenames in `git mv` commands

## When to Use

- Renaming or relocating any Go package
- Moving code between `pkg/` and `internal/`
- Consolidating packages under a new path
- Any bulk rename where the old name appears in both code and string literals
