# uint64 Underflow in Go Memory Benchmark Tests

**Extracted:** 2026-02-11
**Context:** Memory profiling tests that measure `runtime.MemStats.Alloc` deltas

## Problem

Go memory tests that compute growth as `endAlloc - startAlloc` use `uint64` subtraction. When the function under test short-circuits (returns an error immediately, does no real work), GC can reclaim memory between measurements, making `endAlloc < startAlloc`. The unsigned subtraction wraps to a huge value (~18 exabytes), causing the assertion to fail with a misleading message:

```
Memory growth per analysis too high: 18446744073709550568 bytes > 10485760 bytes
```

This happens when refactoring removes the real execution path but leaves the memory test intact. The test still calls the function, but now it returns immediately with an error instead of doing work.

## Solution

Three options, depending on context:

1. **Delete the test** if the execution path it measured no longer exists. A memory test that measures a no-op is meaningless.

2. **Use signed arithmetic** if the test must survive both paths:

   ```go
   growth := int64(endAlloc) - int64(startAlloc)
   if growth > 0 {
       assert.LessOrEqual(t, growth, int64(maxGrowth))
   }
   ```

3. **Guard the subtraction**:

   ```go
   if endAlloc > startAlloc {
       growth := endAlloc - startAlloc
       assert.LessOrEqual(t, growth, maxGrowth)
   }
   ```

## When to Use

- Refactoring removes or short-circuits a function that memory tests exercise
- Memory tests suddenly report multi-exabyte "growth"
- Any `uint64` subtraction of `runtime.MemStats` fields
- Tests that previously passed now fail after changing execution paths
