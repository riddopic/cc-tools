# Go-Specific Development Standards

This document consolidates Go-specific coding standards, best practices, and language-specific guidelines.

## Language Fundamentals

### Go Version and Environment
- **Development**: Target Go 1.25 (latest development version)
- **Production**: Use Go 1.23 (latest stable)
- **Modules**: Always use Go modules (`GO111MODULE=on`)
- **Vendoring**: Only vendor for reproducible builds

### Zero Values
Make zero values useful:

```go
// ✅ Good - Zero value is usable
type Buffer struct {
    buf []byte
}

func (b *Buffer) Write(p []byte) (int, error) {
    b.buf = append(b.buf, p...) // Works even if buf is nil
    return len(p), nil
}

// Usage
var b Buffer // Zero value
b.Write([]byte("hello")) // Works immediately
```

### Value vs Pointer Receivers
Be consistent with receiver types:

```go
// If one method has a pointer receiver, all should
type StatusLine struct {
    config Config
    data   []byte
}

// ✅ Consistent - All pointer receivers
func (s *StatusLine) Start() error { }
func (s *StatusLine) Stop() error { }
func (s *StatusLine) Update(data []byte) { }

// ❌ Inconsistent - Mixed receivers
func (s StatusLine) Start() error { }
func (s *StatusLine) Stop() error { }
```

## Memory Management

### Slice Preallocation
Preallocate slices when size is known:

```go
// ❌ Bad - Multiple allocations
var results []Result
for _, item := range items {
    results = append(results, process(item))
}

// ✅ Good - Single allocation
results := make([]Result, 0, len(items))
for _, item := range items {
    results = append(results, process(item))
}

// ✅ Even better if size is exact
results := make([]Result, len(items))
for i, item := range items {
    results[i] = process(item)
}
```

### String Building
Use strings.Builder for concatenation:

```go
// ❌ Bad - Creates many temporary strings
s := ""
for _, word := range words {
    s += word + " "
}

// ✅ Good - Efficient string building
var b strings.Builder
b.Grow(estimatedSize) // Optional: preallocate
for _, word := range words {
    b.WriteString(word)
    b.WriteByte(' ')
}
s := b.String()
```

### sync.Pool for Temporary Objects
Reuse frequently allocated objects:

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func ProcessData(data []byte) string {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()
    
    buf.Write(data)
    // Process buffer
    return buf.String()
}
```

## Error Handling Patterns

### Custom Error Types
Create domain-specific error types:

```go
type ErrorCode int

const (
    ErrCodeInvalid ErrorCode = iota + 1
    ErrCodeNotFound
    ErrCodePermission
    ErrCodeTimeout
)

type AppError struct {
    Code    ErrorCode
    Message string
    Cause   error
}

func (e AppError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}

func (e AppError) Unwrap() error {
    return e.Cause
}

// Usage
return AppError{
    Code:    ErrCodeNotFound,
    Message: "configuration not found",
    Cause:   err,
}
```

### Error Wrapping
Always add context when wrapping errors:

```go
// ✅ Good - Clear error context
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("reading config file %s: %w", path, err)
    }
    
    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parsing config JSON: %w", err)
    }
    
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("validating config: %w", err)
    }
    
    return &cfg, nil
}
```

## Struct Patterns

### Functional Options
For structs with many optional fields:

```go
type Option func(*StatusLine)

func WithTheme(theme string) Option {
    return func(s *StatusLine) {
        s.theme = theme
    }
}

func WithRefreshRate(rate time.Duration) Option {
    return func(s *StatusLine) {
        s.refreshRate = rate
    }
}

func NewStatusLine(options ...Option) *StatusLine {
    // Set defaults
    s := &StatusLine{
        theme:       "default",
        refreshRate: time.Second,
    }
    
    // Apply options
    for _, opt := range options {
        opt(s)
    }
    
    return s
}

// Usage
sl := NewStatusLine(
    WithTheme("powerline"),
    WithRefreshRate(2 * time.Second),
)
```

### Builder Pattern
For complex object construction:

```go
type ConfigBuilder struct {
    cfg *Config
    err error
}

func NewConfigBuilder() *ConfigBuilder {
    return &ConfigBuilder{
        cfg: &Config{},
    }
}

func (b *ConfigBuilder) Theme(theme string) *ConfigBuilder {
    if b.err != nil {
        return b
    }
    if !isValidTheme(theme) {
        b.err = fmt.Errorf("invalid theme: %s", theme)
        return b
    }
    b.cfg.Theme = theme
    return b
}

func (b *ConfigBuilder) RefreshInterval(d time.Duration) *ConfigBuilder {
    if b.err != nil {
        return b
    }
    if d < time.Second {
        b.err = errors.New("refresh interval must be at least 1 second")
        return b
    }
    b.cfg.RefreshInterval = d
    return b
}

func (b *ConfigBuilder) Build() (*Config, error) {
    if b.err != nil {
        return nil, b.err
    }
    return b.cfg, nil
}

// Usage
cfg, err := NewConfigBuilder().
    Theme("powerline").
    RefreshInterval(2 * time.Second).
    Build()
```

## Initialization Patterns

### Lazy Initialization
Initialize resources only when needed:

```go
type Client struct {
    mu     sync.Mutex
    conn   *Connection
    config *Config
}

func (c *Client) getConnection() (*Connection, error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if c.conn == nil {
        conn, err := dial(c.config)
        if err != nil {
            return nil, err
        }
        c.conn = conn
    }
    
    return c.conn, nil
}
```

### sync.Once for Singletons
Ensure initialization happens exactly once:

```go
type Manager struct {
    once     sync.Once
    instance *service
    err      error
}

func (m *Manager) GetService() (*service, error) {
    m.once.Do(func() {
        m.instance, m.err = initializeService()
    })
    return m.instance, m.err
}
```

## Package Design

### Internal Packages
Use internal for private code:

```
project/
├── cmd/
│   └── app/
│       └── main.go
├── internal/          # Cannot be imported by external packages
│   ├── config/
│   ├── database/
│   └── services/
└── pkg/              # Public API
    └── client/
```

### Package Naming
Follow Go conventions:

```go
// ✅ Good package names
package http      // Not httputil
package sql       // Not sqlutil
package user      // Not users
package account   // Not accounts

// ❌ Bad package names
package util      // Too generic
package common    // Too vague
package helpers   // Not descriptive
package models    // Prefer domain names
```

## Type Assertions and Conversions

### Safe Type Assertions
Always use two-value form:

```go
// ❌ Bad - Can panic
val := x.(string)

// ✅ Good - Safe assertion
val, ok := x.(string)
if !ok {
    // Handle type mismatch
}

// ✅ Good - Type switch for multiple types
switch v := x.(type) {
case string:
    // v is string
case int:
    // v is int
default:
    // Unknown type
}
```

## Defer Patterns

### Resource Cleanup
Always defer cleanup immediately after resource acquisition:

```go
func ProcessFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close() // Immediately after successful open
    
    // Process file
    return nil
}
```

### Defer in Loops
Be careful with defer in loops:

```go
// ❌ Bad - Defers accumulate until function returns
for _, path := range paths {
    file, err := os.Open(path)
    if err != nil {
        continue
    }
    defer file.Close() // All files stay open until function returns
    // Process
}

// ✅ Good - Use function to limit defer scope
for _, path := range paths {
    err := func() error {
        file, err := os.Open(path)
        if err != nil {
            return err
        }
        defer file.Close() // Closes after each iteration
        // Process
        return nil
    }()
    if err != nil {
        // Handle error
    }
}
```

## Context Usage

### Context Propagation
Always pass context as first parameter:

```go
// ✅ Good
func GetUser(ctx context.Context, id string) (*User, error)
func (s *Service) Process(ctx context.Context, data []byte) error

// ❌ Bad
func GetUser(id string, ctx context.Context) (*User, error)
func Process(data []byte) error // Missing context
```

### Context Values
Use custom types for context keys:

```go
// Define key type
type contextKey string

const (
    requestIDKey contextKey = "requestID"
    userIDKey    contextKey = "userID"
)

// Set value
ctx := context.WithValue(ctx, requestIDKey, "12345")

// Get value
if requestID, ok := ctx.Value(requestIDKey).(string); ok {
    // Use requestID
}
```

## Reflection Guidelines

### Minimize Reflection Use
Only use reflection when necessary:

```go
// ❌ Avoid reflection for simple cases
func SetField(obj interface{}, field string, value interface{}) {
    reflect.ValueOf(obj).Elem().FieldByName(field).Set(reflect.ValueOf(value))
}

// ✅ Prefer type-safe approaches
type Config struct {
    Theme string
}

func (c *Config) SetTheme(theme string) {
    c.Theme = theme
}
```

## Build Tags and Constraints

### Platform-Specific Code
Use build tags for platform-specific implementations:

```go
//go:build linux
// +build linux

package system

func GetMemoryStats() (*MemStats, error) {
    // Linux-specific implementation
}
```

### Test-Only Code
Separate test utilities:

```go
//go:build test
// +build test

package testutil

func MockDatabase() *sql.DB {
    // Test-only mock
}
```

## Performance Considerations

### Avoid Unnecessary Allocations
Reuse variables where possible:

```go
// ❌ Bad - Allocates in loop
for i := 0; i < n; i++ {
    s := fmt.Sprintf("item_%d", i) // Allocates each iteration
    process(s)
}

// ✅ Good - Reuse buffer
var buf strings.Builder
for i := 0; i < n; i++ {
    buf.Reset()
    fmt.Fprintf(&buf, "item_%d", i)
    process(buf.String())
}
```

### Benchmark-Driven Optimization
Always measure before optimizing:

```go
func BenchmarkProcess(b *testing.B) {
    data := prepareData()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        Process(data)
    }
}

// Run: task bench
// Or directly: go test -bench=. -benchmem
```

## Go-Specific Best Practices Summary

1. **Embrace simplicity** - Go favors simple, obvious code
2. **Make zero values useful** - Design types to work without initialization
3. **Use composition** - Embed types instead of inheritance
4. **Handle errors explicitly** - No exceptions, check all errors
5. **Keep interfaces small** - Single responsibility principle
6. **Prefer standard library** - Use standard packages when possible
7. **Minimize allocations** - Reuse objects, preallocate slices
8. **Document exported items** - All public APIs need documentation
9. **Test with the race detector** - Always run tests with `-race`
10. **Profile before optimizing** - Measure, don't guess about performance