---
name: go-coding-standards
description: Apply Go idioms and project coding standards. Use when writing Go code, reviewing implementations, making architectural decisions, or when the user asks about Go best practices.
---

# Go Coding Standards

Apply these standards when writing or reviewing Go code in this project.

## Quick Reference

| Principle | Rule |
|-----------|------|
| Interfaces | Accept interfaces, return concrete types |
| Errors | Errors are values - handle explicitly with context wrapping |
| Functions | Keep under 50 lines, use early returns |
| Receivers | Be consistent - all pointer or all value |
| Zero values | Make them useful - types should work without initialization |

## Core Go Idioms

1. **Errors are values** - No exceptions, explicit error handling
2. **Make zero values useful** - Design types to work without initialization
3. **Accept interfaces, return concrete types**
4. **Composition over inheritance** - Use embedding and interfaces
5. **Small interfaces** - One or two methods per interface
6. **Early returns** - Reduce nesting with guard clauses

## Error Handling Pattern

```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("loading config %s: %w", path, err)
}

// Define sentinel errors
var ErrNotFound = errors.New("not found")

// Check specific errors
if errors.Is(err, ErrNotFound) {
    // Handle not found
}
```

## Struct Patterns

For optional parameters, use functional options:

```go
type Option func(*StatusLine)

func WithTheme(theme string) Option {
    return func(s *StatusLine) {
        s.theme = theme
    }
}

func NewStatusLine(options ...Option) *StatusLine {
    s := &StatusLine{theme: "default"} // defaults
    for _, opt := range options {
        opt(s)
    }
    return s
}
```

## Memory Optimization

- Preallocate slices when size is known: `make([]T, 0, expectedSize)`
- Use `strings.Builder` for concatenation
- Use `sync.Pool` for frequently allocated objects

## Detailed Standards

For complete Go idioms, see [go-specific.md](../../../docs/examples/standards/go-specific.md)
For interface design, see [interfaces.md](../../../docs/examples/standards/interfaces.md)
For documentation standards, see [documentation.md](../../../docs/examples/standards/documentation.md)
For project coding guidelines, see [CODING_GUIDELINES.md](../../../docs/CODING_GUIDELINES.md)
