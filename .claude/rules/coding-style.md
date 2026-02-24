---
paths:
  - "**/*.go"
---

# Go Coding Standards

Project-specific constraints for cc-tools. Standard Go idioms (error handling, naming, early returns) are assumed knowledge.

## Project Constraints

- **No ad-hoc adapters in `cmd/`**: Adapter structs satisfying `internal/` interfaces belong in `internal/`, not `cmd/`
- **One interface definition**: Never duplicate an interface — search existing definitions first (`rg "type.*interface" internal/ --glob "*.go"`)
- **Functions under 50 lines**, nesting under 3 levels
- **Compile-time interface checks**: `var _ Interface = (*Impl)(nil)`

## Import Organization

```go
import (
    // Standard library
    "context"
    "fmt"

    // Third-party packages
    "github.com/spf13/cobra"

    // Internal packages
    "github.com/riddopic/cc-tools/internal/config"
)
```

Enforced by `goimports -local github.com/riddopic/cc-tools`.

## Code Quality Checklist

Before committing Go code:

- [ ] Code passes `task fmt` (gofmt + goimports)
- [ ] No linter warnings (`task lint`)
- [ ] All tests pass (`task test`)
- [ ] Race detector passes (`task test-race`)
- [ ] No commented-out code
- [ ] No TODO without issue reference
