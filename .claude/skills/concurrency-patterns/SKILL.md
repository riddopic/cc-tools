---
name: concurrency-patterns
description: Apply Go concurrency patterns safely. Use when working with goroutines, channels, worker pools, context cancellation, or any concurrent code. Helps avoid goroutine leaks and race conditions.
---

# Go Concurrency Patterns for Quanta

## Core Principles

1. **Always use context for cancellation**
2. **Never leak goroutines** - ensure all goroutines can exit
3. **Close channels from sender side only**
4. **Use sync primitives appropriately**

## Context Propagation

```go
// Always pass context as first parameter
func GetUser(ctx context.Context, id string) (*User, error)

// Check context in long operations
select {
case <-ctx.Done():
    return ctx.Err()
case result := <-resultCh:
    return result, nil
}
```

## Worker Pool Pattern

```go
func ProcessItems(ctx context.Context, items []Item, workers int) error {
    jobs := make(chan Item, len(items))
    results := make(chan error, len(items))

    // Start workers
    var wg sync.WaitGroup
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for item := range jobs {
                select {
                case <-ctx.Done():
                    return
                default:
                    results <- processItem(item)
                }
            }
        }()
    }

    // Send jobs
    for _, item := range items {
        jobs <- item
    }
    close(jobs)

    // Wait and collect
    go func() {
        wg.Wait()
        close(results)
    }()

    for err := range results {
        if err != nil {
            return err
        }
    }
    return nil
}
```

## sync.Once for Singletons

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

## Lazy Initialization with Mutex

```go
type Client struct {
    mu   sync.Mutex
    conn *Connection
}

func (c *Client) getConnection() (*Connection, error) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if c.conn == nil {
        conn, err := dial()
        if err != nil {
            return nil, err
        }
        c.conn = conn
    }
    return c.conn, nil
}
```

## Timeout Pattern

```go
func ExecuteWithTimeout(ctx context.Context, timeout time.Duration) error {
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()

    resultCh := make(chan error, 1)
    go func() {
        resultCh <- doWork(ctx)
    }()

    select {
    case err := <-resultCh:
        return err
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

## Common Pitfalls to Avoid

### Goroutine Leak

```go
// WRONG: Goroutine never exits if channel not read
go func() {
    ch <- result  // Blocks forever if no receiver
}()

// RIGHT: Use select with context
go func() {
    select {
    case ch <- result:
    case <-ctx.Done():
    }
}()
```

### Defer in Loops

```go
// WRONG: All files stay open
for _, path := range paths {
    file, _ := os.Open(path)
    defer file.Close()  // Accumulates!
}

// RIGHT: Use function scope
for _, path := range paths {
    func() {
        file, _ := os.Open(path)
        defer file.Close()
        // process
    }()
}
```

## Race Detection

Always test concurrent code with race detector:

```bash
make test-race
# or
go test -race ./...
```

## Detailed Concurrency Patterns

For complete patterns, see [concurrency.md](../../../docs/examples/patterns/concurrency.md)
