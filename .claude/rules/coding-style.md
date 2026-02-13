---
paths:
  - "**/*.go"
---

# Go Coding Standards

Go idioms and coding standards. Reference `docs/CODING_GUIDELINES.md` for comprehensive details.

## Core Principles

- **Simplicity First**: Favor simple, obvious solutions over clever ones
- **Explicit Over Implicit**: Make intentions clear in code
- **Composition Over Inheritance**: Use interfaces and embedding
- **Early Returns**: Reduce nesting with guard clauses
- **Small Functions**: Keep functions focused and under 50 lines
- **Testability**: Design code to be easily testable
- **No Ad-Hoc Adapters in cmd/**: Adapter structs that satisfy `internal/` interfaces belong in `internal/`, not `cmd/`
- **One Interface Definition**: Never duplicate an interface — search for existing definitions first

## Error Handling

```go
// ✅ DO: Wrap errors with context
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
    }
    return parseConfig(data)
}

// ❌ DON'T: Return bare errors without context
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err  // No context about what failed
    }
    return parseConfig(data)
}
```

## Early Returns

```go
// ✅ DO: Use guard clauses for early returns
func ProcessUser(user *User) error {
    if user == nil {
        return ErrNilUser
    }
    if user.ID == "" {
        return ErrMissingID
    }
    if !user.IsActive {
        return ErrInactiveUser
    }

    // Happy path at the end
    return processActiveUser(user)
}

// ❌ DON'T: Deeply nest conditions
func ProcessUser(user *User) error {
    if user != nil {
        if user.ID != "" {
            if user.IsActive {
                return processActiveUser(user)
            }
        }
    }
    return errors.New("invalid user")
}
```

## Function Size

```go
// ✅ DO: Keep functions focused (<50 lines)
func (s *Service) ProcessOrder(ctx context.Context, order *Order) error {
    if err := s.validateOrder(order); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    total := s.calculateTotal(order.Items)
    payment, err := s.processPayment(ctx, order.CustomerID, total)
    if err != nil {
        return fmt.Errorf("payment failed: %w", err)
    }

    return s.fulfillOrder(ctx, order, payment)
}

// ❌ DON'T: Write monolithic functions with multiple responsibilities
```

## Naming Conventions

```go
// Package names: lowercase, singular
package user    // ✅
package users   // ❌
package userPkg // ❌

// Interfaces: behavior-focused, -er suffix for single method
type Reader interface { Read([]byte) (int, error) }     // ✅
type IReader interface {}                                // ❌ No Hungarian notation
type ReaderInterface interface {}                        // ❌ Redundant suffix

// Exported vs unexported
type StatusLine struct {}      // ✅ Exported (public)
type renderer struct {}        // ✅ Unexported (private)
func NewStatusLine() {}        // ✅ Constructor
func parseConfig() {}          // ✅ Internal helper
```

## Import Organization

```go
import (
    // Standard library
    "context"
    "fmt"
    "io"

    // Third-party packages
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "go.uber.org/zap"

    // Internal packages
    "github.com/riddopic/cc-tools/internal/config"
    "github.com/riddopic/cc-tools/internal/statusline"
)
```

## Interface Design

```go
// ✅ DO: Accept interfaces, return concrete types
func ProcessData(r io.Reader) (*Result, error) {
    return &Result{}, nil
}

// ✅ DO: Keep interfaces small and focused
type Validator interface {
    Validate(ctx context.Context, data interface{}) error
}

// ❌ DON'T: Create god interfaces
type Service interface {
    GetUser() (*User, error)
    CreateUser() error
    GetOrder() (*Order, error)
    ProcessPayment() error
    // Too many responsibilities!
}
```

## Code Quality Checklist

Before committing Go code:

- [ ] Code passes `task fmt` (gofmt + goimports)
- [ ] No linter warnings (`task lint`)
- [ ] All tests pass (`task test`)
- [ ] Race detector passes (`task test-race`)
- [ ] Functions are under 50 lines
- [ ] No deep nesting (>3 levels)
- [ ] Errors are wrapped with context
- [ ] No commented-out code
- [ ] No TODO without issue reference
