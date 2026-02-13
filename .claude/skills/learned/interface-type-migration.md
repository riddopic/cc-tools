# Interface Type Migration Across Package Boundaries

**Extracted:** 2026-02-12
**Context:** Migrating a concrete type (e.g., *zap.Logger) to an interface across a large Go codebase

## Problem
When you change a struct field type in a core package (e.g., `Config.Logger` from `*zap.Logger` to `interfaces.SugaredLogger`), every downstream consumer that constructs that struct breaks. The blast radius is invisible until you compile — grep only finds direct references, not transitive consumers through function calls.

## Solution
Use iterative `go build ./...` to discover each layer of breakage, fixing from leaf packages inward:

1. **Change the core package first** — update all field types, constructors, and call sites within the package
2. **Build and fix layer by layer** — run `go build ./...`, fix the reported errors, repeat
3. **At package boundaries, create bridge helpers** — when a consumer needs BOTH the old type (for other deps) and the new type (for your migrated package), create a small adapter:
   ```go
   // Consumer still needs *zap.Logger for sourcecache, coordinator, etc.
   // But LLM configs now need interfaces.SugaredLogger
   slog := logger.NewSugaredLogger("info", true)  // for LLM configs
   // zapLogger still used for non-LLM downstream
   ```
4. **Don't change the return type of shared helpers** — if `newProductionLogger()` returns `*zap.Logger` and is used by both migrated and non-migrated consumers, add a NEW helper (`newSugaredLogger()`) rather than changing the existing one
5. **Config-only files may not need the implementation import** — if a file only declares `Logger interfaces.SugaredLogger` but never constructs one, it only needs the `interfaces` import, NOT the `applogger` import

## Key Gotchas
- **IDE diagnostics cascade badly** — when package A doesn't compile, the IDE reports phantom errors in packages B, C, D that import A. Only trust `go build` output.
- **Import shadowing** — if the package name matches a local variable (e.g., `logger` package vs `logger` struct field), use an alias like `applogger`
- **Unused imports from batch edits** — sed/batch replacements often add imports to files that don't need them (config files that only declare the type but never construct it)
- **`replace_all` for test files** — test files often have identical patterns (e.g., `zap.NewNop()`) repeated many times; use replace_all=true

## When to Use
- Migrating a concrete type to an interface across 10+ files
- Changing a field type on a widely-used Config struct
- Any refactoring where the core change is in an `internal/` package consumed by `cmd/` and other `internal/` packages
