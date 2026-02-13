# Interface Standards

This document defines standards for designing, implementing, and organizing interfaces in Go.

## Interface Design Principles

### 1. Small and Focused
Keep interfaces small with a single responsibility:

```go
// ✅ Good - Small, focused interface
type Reader interface {
    Read([]byte) (int, error)
}

// ❌ Bad - Too many responsibilities
type FileManager interface {
    Read([]byte) (int, error)
    Write([]byte) (int, error)
    Delete() error
    Rename(string) error
    GetInfo() (FileInfo, error)
}
```

### 2. Interface Segregation
Split large interfaces into smaller, composable ones:

```go
// ✅ Good - Segregated interfaces
type Reader interface {
    Read([]byte) (int, error)
}

type Writer interface {
    Write([]byte) (int, error)
}

type Closer interface {
    Close() error
}

// Compose interfaces when needed
type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}
```

### 3. Accept Interfaces, Return Concrete Types
Functions should accept interfaces and return concrete implementations:

```go
// ✅ Good
func ProcessData(r io.Reader) (*Result, error) {
    // Process and return concrete type
    return &Result{...}, nil
}

// ❌ Avoid returning interfaces unless necessary
func ProcessData(r io.Reader) (ResultInterface, error) {
    // Returning interface reduces flexibility
    return &Result{...}, nil
}
```

## Naming Conventions

### Interface Names
Use descriptive names that indicate behavior:

```go
// ✅ Good naming
type Reader interface {}      // -er suffix for single method
type Validator interface {}    // Describes behavior
type Repository interface {}   // Domain-specific role

// ❌ Bad naming
type IReader interface {}      // No Hungarian notation
type ReaderInterface interface {} // Redundant suffix
type ReadManager interface {}  // Vague purpose
```

### Method Names
Keep method names clear and consistent:

```go
type Validator interface {
    // Clear, action-oriented method name
    Validate(ctx context.Context, data interface{}) error
    
    // Use consistent parameter names across implementations
    IsValid(value string) bool
}
```

## Interface Organization

### Package Structure
Define interfaces in the package that uses them:

```go
// ✅ Good - Interface defined where it's used
// internal/service/interfaces.go
package service

type Repository interface {
    Get(ctx context.Context, id string) (*Entity, error)
    Save(ctx context.Context, entity *Entity) error
}

// internal/repository/implementation.go
package repository

import "internal/service"

type PostgresRepo struct {}

// Implements service.Repository
func (r *PostgresRepo) Get(ctx context.Context, id string) (*service.Entity, error) {
    // Implementation
}
```

### File Organization
Place all package interfaces in `interfaces.go`:

```go
// internal/statusline/interfaces.go
package statusline

import "context"

// Renderer defines the rendering behavior for statuslines.
type Renderer interface {
    Render(data *StatusData) (string, error)
    SetTheme(theme Theme)
}

// MetricsCollector defines how metrics are gathered.
type MetricsCollector interface {
    Collect(ctx context.Context) (*Metrics, error)
    Subscribe(ch chan<- MetricUpdate)
}

// ThemeProvider defines theme management behavior.
type ThemeProvider interface {
    GetTheme(name string) (Theme, error)
    ListThemes() []string
}
```

## Standard Interfaces

### Error Handling
Implement standard error interfaces when appropriate:

```go
// Implement error interface
type ValidationError struct {
    Field string
    Value interface{}
    Msg   string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Msg)
}

// Support error unwrapping
func (e ValidationError) Unwrap() error {
    return e.Cause
}

// Support error comparison
func (e ValidationError) Is(target error) bool {
    t, ok := target.(ValidationError)
    return ok && t.Field == e.Field
}
```

### Stringer Interface
Implement Stringer for custom string representation:

```go
type Status int

const (
    StatusPending Status = iota
    StatusRunning
    StatusComplete
    StatusError
)

func (s Status) String() string {
    switch s {
    case StatusPending:
        return "pending"
    case StatusRunning:
        return "running"
    case StatusComplete:
        return "complete"
    case StatusError:
        return "error"
    default:
        return "unknown"
    }
}
```

## Interface Composition

### Embedding Interfaces
Compose complex interfaces from simpler ones:

```go
// Base interfaces
type Reader interface {
    Read(ctx context.Context, key string) ([]byte, error)
}

type Writer interface {
    Write(ctx context.Context, key string, data []byte) error
}

type Deleter interface {
    Delete(ctx context.Context, key string) error
}

// Composed interface
type Storage interface {
    Reader
    Writer
    Deleter
}

// Extended interface
type CachedStorage interface {
    Storage
    Invalidate(ctx context.Context, key string) error
    Stats() CacheStats
}
```

## Implementation Patterns

### Compile-Time Interface Checks
Ensure types implement interfaces at compile time:

```go
// Ensure StatusLine implements the Runner interface
var _ Runner = (*StatusLine)(nil)

// Or with explicit type
var _ Runner = new(StatusLine)

// Multiple interface checks
var (
    _ Runner    = (*StatusLine)(nil)
    _ io.Closer = (*StatusLine)(nil)
)
```

### Optional Interface Methods
Use type assertions for optional behavior:

```go
type Renderer interface {
    Render(data *StatusData) (string, error)
}

// Optional interface
type ThemeableRenderer interface {
    Renderer
    SetTheme(theme Theme)
}

func UpdateTheme(r Renderer, theme Theme) {
    // Check if renderer supports themes
    if tr, ok := r.(ThemeableRenderer); ok {
        tr.SetTheme(theme)
    }
}
```

### Adapter Pattern
Create adapters to satisfy interfaces:

```go
// Adapt function to interface
type HandlerFunc func(ctx context.Context, data []byte) error

type Handler interface {
    Handle(ctx context.Context, data []byte) error
}

// HandlerFunc implements Handler
func (f HandlerFunc) Handle(ctx context.Context, data []byte) error {
    return f(ctx, data)
}

// Usage
var h Handler = HandlerFunc(func(ctx context.Context, data []byte) error {
    // Handle data
    return nil
})
```

## Testing with Interfaces

### Mock Generation
Generate mocks for interfaces:

```go
//go:generate mockery --name=Repository --output=mocks --outpkg=mocks

type Repository interface {
    Get(ctx context.Context, id string) (*Entity, error)
    Save(ctx context.Context, entity *Entity) error
}
```

### Test Doubles
Create simple test implementations:

```go
type stubRenderer struct {
    renderFunc func(*StatusData) (string, error)
}

func (s stubRenderer) Render(data *StatusData) (string, error) {
    if s.renderFunc != nil {
        return s.renderFunc(data)
    }
    return "", nil
}

// Usage in tests
func TestStatusLine(t *testing.T) {
    renderer := stubRenderer{
        renderFunc: func(data *StatusData) (string, error) {
            return "test output", nil
        },
    }
    
    sl := NewStatusLine(renderer)
    // Test with stub
}
```

## Anti-Patterns to Avoid

### 1. Interface Pollution
Don't create interfaces "just in case":

```go
// ❌ Bad - Unnecessary interface
type UserService interface {
    GetUser(id string) (*User, error)
}

type userService struct {}

func (s *userService) GetUser(id string) (*User, error) {
    // Only one implementation exists
}

// ✅ Good - Use concrete type until interface is needed
type UserService struct {}

func (s *UserService) GetUser(id string) (*User, error) {
    // Start with concrete type
}
```

### 2. Leaky Abstractions
Don't expose implementation details:

```go
// ❌ Bad - Exposes SQL details
type Repository interface {
    ExecuteSQL(query string, args ...interface{}) (sql.Result, error)
}

// ✅ Good - Abstract interface
type Repository interface {
    Get(ctx context.Context, id string) (*Entity, error)
    Save(ctx context.Context, entity *Entity) error
}
```

### 3. God Interfaces
Avoid interfaces with too many methods:

```go
// ❌ Bad - Too many responsibilities
type Service interface {
    // User operations
    GetUser(id string) (*User, error)
    CreateUser(user *User) error
    UpdateUser(user *User) error
    DeleteUser(id string) error
    
    // Order operations
    GetOrder(id string) (*Order, error)
    CreateOrder(order *Order) error
    
    // Payment operations
    ProcessPayment(payment *Payment) error
    RefundPayment(id string) error
}

// ✅ Good - Separate interfaces
type UserService interface {
    GetUser(id string) (*User, error)
    CreateUser(user *User) error
}

type OrderService interface {
    GetOrder(id string) (*Order, error)
    CreateOrder(order *Order) error
}
```

## Best Practices Summary

1. **Keep interfaces small** - Prefer many small interfaces over few large ones
2. **Define at point of use** - Define interfaces where they're used, not with implementations
3. **Name by behavior** - Use names that describe what the interface does
4. **Document clearly** - Explain expected behavior in interface documentation
5. **Compose when needed** - Build complex interfaces from simple ones
6. **Check at compile time** - Use compile-time checks to ensure implementation
7. **Return concrete types** - Only return interfaces when necessary
8. **Avoid premature abstraction** - Create interfaces when you need them
9. **Use standard interfaces** - Implement error, Stringer, etc. when appropriate
10. **Test with interfaces** - Use interfaces to enable easy testing