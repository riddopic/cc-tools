---
name: interface-design
description: Apply Go interface design patterns. Use when defining interfaces, composing behaviors, implementing mocking boundaries, or reviewing interface usage. Helps create small, focused, testable interfaces.
---

# Go Interface Design for Quanta

## Core Principles

1. **Small interfaces** - 1-2 methods per interface
2. **Define at point of use** - Interfaces live where they're used
3. **Accept interfaces, return concrete types**
4. **Compose through embedding**

## Interface Size

```go
// GOOD: Small, focused interface
type Reader interface {
    Read([]byte) (int, error)
}

// BAD: Too many responsibilities
type FileManager interface {
    Read([]byte) (int, error)
    Write([]byte) (int, error)
    Delete() error
    Rename(string) error
}
```

## Interface Segregation

```go
// Split large interfaces into composable pieces
type Reader interface { Read([]byte) (int, error) }
type Writer interface { Write([]byte) (int, error) }
type Closer interface { Close() error }

// Compose when needed
type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}
```

## Naming Conventions

```go
// GOOD naming
type Reader interface {}     // -er suffix for single method
type Validator interface {}  // Describes behavior
type Repository interface {} // Domain role

// BAD naming
type IReader interface {}         // No Hungarian notation
type ReaderInterface interface {} // Redundant suffix
```

## Package Organization

Define interfaces in the package that uses them:

```go
// internal/service/interfaces.go
package service

type Repository interface {
    Get(ctx context.Context, id string) (*Entity, error)
}

// internal/repository/postgres.go
package repository

// Implements service.Repository
type PostgresRepo struct {}
```

## Compile-Time Checks

```go
// Ensure type implements interface at compile time
var _ Runner = (*StatusLine)(nil)

// Multiple checks
var (
    _ Runner    = (*StatusLine)(nil)
    _ io.Closer = (*StatusLine)(nil)
)
```

## Optional Behavior

```go
type Renderer interface {
    Render(data *StatusData) (string, error)
}

type ThemeableRenderer interface {
    Renderer
    SetTheme(theme Theme)
}

func UpdateTheme(r Renderer, theme Theme) {
    if tr, ok := r.(ThemeableRenderer); ok {
        tr.SetTheme(theme)
    }
}
```

## Anti-Patterns

1. **Interface pollution** - Don't create interfaces "just in case"
2. **Leaky abstractions** - Don't expose implementation details
3. **God interfaces** - Avoid interfaces with many methods

```go
// BAD: Premature abstraction
type UserService interface { GetUser(id string) (*User, error) }
// Only one implementation - use concrete type until needed

// GOOD: Start with concrete type
type UserService struct {}
func (s *UserService) GetUser(id string) (*User, error) { ... }
```

## Detailed Patterns

For complete interface patterns, see [interfaces.md](../../../docs/examples/standards/interfaces.md)
