---
name: code-review
description: Apply Quanta code review standards. Use when reviewing code, evaluating pull requests, or checking code quality before commits. Ensures Go idioms, security, and project standards are followed.
---

# Code Review Standards for Quanta

## Pre-Commit Checklist

Before committing, ensure:

- [ ] Code passes formatting: `make fmt`
- [ ] No linter warnings: `make lint`
- [ ] All tests pass: `make test`
- [ ] Race detector passes: `make test-race`
- [ ] New code has test coverage: `make coverage`
- [ ] Error messages are clear and actionable
- [ ] No commented-out code
- [ ] No TODO without issue references

**Quick check**: `make pre-commit`

## Code Quality Focus Areas

### 1. Go Idioms

- [ ] Errors are handled explicitly (no ignored errors)
- [ ] Context is passed as first parameter
- [ ] Interfaces are small (1-2 methods)
- [ ] Zero values are useful
- [ ] Early returns reduce nesting

### 2. Error Handling

```go
// GOOD: Wrapped with context
return fmt.Errorf("loading config %s: %w", path, err)

// BAD: Lost context
return err
```

### 3. Function Design

- [ ] Functions under 50 lines
- [ ] Single responsibility
- [ ] Clear naming (verbs for functions, nouns for types)

### 4. Testing

- [ ] Tests focus on behavior, not implementation
- [ ] Table-driven tests for multiple scenarios
- [ ] No hardcoded secrets in tests
- [ ] Coverage ≥80% for new code

### 5. Security

- [ ] No hardcoded secrets
- [ ] Input validation at boundaries
- [ ] Proper error messages (no stack traces to users)
- [ ] Secure file permissions (0600/0750)

### 6. Concurrency

- [ ] Context used for cancellation
- [ ] No goroutine leaks
- [ ] Channels closed by sender
- [ ] Race detector passes

## Red Flags

**Stop and address these immediately:**

- `_ = someFunc()` - Ignored error
- `panic()` for normal error handling
- Magic numbers without constants
- Deeply nested code (>3 levels)
- Functions over 50 lines
- Missing godoc on exported items

## Review Questions

1. **Does this code need to exist?** (YAGNI)
2. **Is there a simpler solution?** (KISS)
3. **Can we leverage existing code?** (LEVER)
4. **Is it testable?** (TDD requirement)
5. **Would I understand this in 6 months?**

## Validation Commands

```bash
make fmt           # Format code
make lint          # Lint check
make test          # Run tests
make test-race     # Race detection
make coverage      # Coverage report
make pre-commit    # All checks
```

## Project References

- [CODING_GUIDELINES.md](../../../docs/CODING_GUIDELINES.md)
- [go-specific.md](../../../docs/examples/standards/go-specific.md)
- [testing.md](../../../docs/examples/patterns/testing.md)
