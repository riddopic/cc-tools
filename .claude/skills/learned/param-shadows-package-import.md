# Parameter Name Shadows Package Import in Go

**Extracted:** 2026-02-12
**Context:** When migrating function signatures where a parameter name matches an imported package name

## Problem
When a Go function has a parameter named `logger` and the file also imports a package named `logger`, the parameter shadows the package inside the function body. You cannot call `logger.FromZapLogger(logger)` because both `logger` references resolve to the parameter.

This commonly occurs during interface migrations where you're wrapping old types with adapter functions from the same-named package.

## Solution
Three approaches, in order of preference:

1. **Rename the parameter** (best for unexported functions):
   ```go
   // Before: logger shadows the package
   func createFoo(logger *zap.Logger) {
       // Can't call logger.FromZapLogger() here!
   }

   // After: zapLogger doesn't shadow
   func createFoo(zapLogger *zap.Logger) {
       slog := logger.FromZapLogger(zapLogger)
   }
   ```

2. **Pre-compute before the shadowed scope** (best for exported functions where renaming breaks API):
   ```go
   slog := logger.FromZapLogger(zapLogger)
   result := callThatNeedsSugared(slog)
   ```

3. **Use an import alias** (last resort — changes all call sites in the file):
   ```go
   import applogger "github.com/example/internal/logger"
   ```

## When to Use
- Migrating `*zap.Logger` parameters to `interfaces.SugaredLogger`
- Any refactoring where a parameter name collides with a package import
- When you see `undefined: X` or `X has no field or method Y` where X is both a param and a package
