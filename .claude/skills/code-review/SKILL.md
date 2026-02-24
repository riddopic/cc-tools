---
name: code-review
description: Apply code review standards. Use when reviewing code, evaluating pull requests, or checking code quality before commits. Covers Go idioms, testing patterns, security, and project standards.
---

# Code Review Standards

## Pre-Commit Checklist

```bash
task pre-commit         # Or: task check (fmt + lint + test-race)
```

- [ ] Code passes formatting: `task fmt`
- [ ] No linter warnings: `task lint`
- [ ] All tests pass: `task test`
- [ ] Race detector passes: `task test-race`

## Go Code Review Checklist

- [ ] All errors checked and wrapped with context (`fmt.Errorf("...: %w", err)`)
- [ ] Resources closed with `defer` immediately after creation
- [ ] No goroutine leaks (channels closed, contexts canceled)
- [ ] Interfaces defined by consumers, small (1-2 methods)
- [ ] Context passed as first parameter
- [ ] Mutexes protect shared state
- [ ] Functions under 50 lines, single responsibility
- [ ] Early returns reduce nesting
- [ ] No commented-out code, no TODO without issue reference

## Go Test Review Checklist

- [ ] Tests are table-driven with clear case names
- [ ] Test names describe behavior, not implementation
- [ ] Error messages include input, got, and want
- [ ] Parallel tests don't share mutable state
- [ ] Cleanup registered with `t.Cleanup`
- [ ] Tests verify behavior, not implementation details
- [ ] Coverage includes edge cases and error paths
- [ ] Coverage ≥80% for new code

## Valid Patterns (Do NOT Flag)

- `_ = err` with reason comment — intentionally ignored errors
- `//nolint` directives with explanation
- Channel without close when consumer stops via context cancellation
- Naked returns in functions < 5 lines with named returns

## Context-Sensitive Rules

| Issue | Flag ONLY IF |
|-------|--------------|
| Missing error check | Error return is actionable |
| Goroutine leak | No context cancellation path exists |
| Missing defer | Resource isn't explicitly closed before next acquisition or return |
| Interface pollution | Interface has > 1 method AND only one consumer |

## Red Flags — Stop and Address

- `_ = someFunc()` — Ignored error without reason
- `panic()` for normal error handling
- Magic numbers without constants
- Deeply nested code (>3 levels)
- Functions over 50 lines

## Review Questions

1. **Does this code need to exist?** (YAGNI)
2. **Is there a simpler solution?** (KISS)
3. **Can we leverage existing code?** (LEVER)
4. **Is it testable?** (TDD requirement)
5. **Would I understand this in 6 months?**

## Before Submitting Findings

Load and follow `review-verification-protocol` before reporting any issue.
