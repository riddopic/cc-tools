# Performance Guidelines

Performance optimization and profiling patterns for Go code.

## Optimization Principles

1. **Measure First**: Profile before optimizing
2. **Optimize Hot Paths**: Focus on frequently executed code
3. **Reduce Allocations**: Minimize heap allocations
4. **Use Buffering**: Buffer I/O operations
5. **Concurrent When Beneficial**: Use goroutines for parallel work

## Profiling with pprof

```go
// Enable profiling in development
import _ "net/http/pprof"

func init() {
    if os.Getenv("ENABLE_PROFILING") == "true" {
        go func() {
            log.Println(http.ListenAndServe("localhost:6060", nil))
        }()
    }
}
```

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./...
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=. ./...
go tool pprof mem.prof

# HTTP profiling (when server running)
go tool pprof http://localhost:6060/debug/pprof/profile
go tool pprof http://localhost:6060/debug/pprof/heap
```

## Benchmarking

```go
// ✅ DO: Write benchmarks for performance-critical code
func BenchmarkProcessData(b *testing.B) {
    data := generateTestData(1000)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ProcessData(data)
    }
}

// ✅ DO: Benchmark with memory allocation stats
func BenchmarkWithAllocs(b *testing.B) {
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        result := CreateResult()
        _ = result
    }
}
```

```bash
# Run benchmarks
task bench

# Compare benchmark results
go test -bench=. -count=5 > old.txt
# Make changes
go test -bench=. -count=5 > new.txt
benchstat old.txt new.txt
```

## Memory Optimization

```go
// ✅ DO: Pre-allocate slices when size is known
data := make([]string, 0, expectedSize)

// ✅ DO: Use sync.Pool for frequently allocated objects
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func processWithPool() {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()

    // Use buf...
}

// ✅ DO: Use strings.Builder for string concatenation
var b strings.Builder
b.Grow(expectedLen)  // Pre-allocate if size known
b.WriteString("Session: ")
b.WriteString(sessionID)
result := b.String()

// ❌ DON'T: Concatenate strings with +
result := "Session: " + sessionID + " Status: " + status
```

## Channel Optimization

```go
// ✅ DO: Buffer channels appropriately
ch := make(chan Event, 100)  // Buffered channel

// ✅ DO: Close channels when done sending
func producer(ch chan<- int) {
    defer close(ch)
    for i := 0; i < 100; i++ {
        ch <- i
    }
}

// ✅ DO: Use select with default for non-blocking operations
select {
case msg := <-ch:
    process(msg)
default:
    // Channel empty, do something else
}
```

## Model Selection for Agents

When using AI agents in workflows:

| Model | Use Case | Cost/Performance |
| ------- | ---------- | ------------------ |
| **Haiku 4.5** | Lightweight agents, frequent invocation, pair programming | 3x cost savings vs Sonnet |
| **Sonnet 4.5** | Main development, orchestration, complex coding | Best coding model |
| **Opus 4.5** | Complex architecture, deep reasoning, research | Maximum reasoning |

## Context Window Management

Avoid last 20% of context window for:

- Large-scale refactoring
- Multi-file feature implementation
- Complex debugging sessions

Lower context sensitivity tasks:

- Single-file edits
- Independent utility creation
- Documentation updates
- Simple bug fixes

## Quick Commands

```bash
# Run benchmarks with memory stats
task bench

# Profile a specific test
go test -cpuprofile=cpu.prof -memprofile=mem.prof \
    -run=XXX -bench=BenchmarkName ./pkg/...

# Analyze profile
go tool pprof -http=:8080 cpu.prof

# Check for race conditions
task test-race
```

## Performance Checklist

Before optimizing:

- [ ] Profiled to identify bottleneck
- [ ] Benchmark exists for the hot path
- [ ] Optimization addresses measured problem

After optimizing:

- [ ] Benchmark shows improvement
- [ ] No regressions in other benchmarks
- [ ] Code is still readable and maintainable
- [ ] Tests still pass
