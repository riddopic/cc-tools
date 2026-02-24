---
name: code-review
description: Apply code review standards. Use when reviewing code, evaluating pull requests, or checking code quality before commits. Covers Go idioms, testing patterns, security, and project standards.
---

# Code Review Process

How to review code effectively. Coding conventions are in `.claude/rules/` (coding-style.md, testing.md, security.md) — this skill focuses on the review workflow and judgment calls.

## Review Workflow

1. Run `task pre-commit` (or `task check`) to catch mechanical issues first
2. Apply the checklists below for issues automation cannot catch
3. Follow the verification protocol below before reporting any findings

## Pre-Report Verification Checklist

Before flagging ANY issue, verify:

- [ ] **I read the actual code** — Not just the diff context, but the full function
- [ ] **I searched for usages** — Before claiming "unused", searched all references
- [ ] **I checked surrounding code** — The issue may be handled elsewhere
- [ ] **I verified syntax against current docs** — Library syntax evolves
- [ ] **I distinguished "wrong" from "different style"** — Both approaches may be valid
- [ ] **I considered intentional design** — Checked comments, CLAUDE.md, architectural context

## Judgment-Based Review Checklist

Issues that require judgment beyond what linters catch:

- [ ] No goroutine leaks (channels closed, contexts canceled)
- [ ] Interfaces defined by consumers, small (1-2 methods)
- [ ] Mutexes protect shared state correctly
- [ ] Functions have single responsibility
- [ ] Resources closed with `defer` immediately after creation
- [ ] Test names describe behavior, not implementation
- [ ] Error messages include input, got, and want
- [ ] Parallel tests don't share mutable state

## Verification by Issue Type

### "Unused Variable/Function"

Before flagging, you MUST:
1. Search for ALL references in the codebase (use `rg` for exact matches)
2. Check if it's exported and used by external consumers
3. Check if it's used via reflection or interface satisfaction
4. Verify it's not a callback passed to a framework

### "Missing Validation/Error Handling"

Before flagging, you MUST:
1. Check if validation exists at a higher level (caller, middleware)
2. Check if the type system already enforces constraints
3. Verify the "missing" check isn't present in a different form

### "Type Assertion/Unsafe Cast"

Before flagging, you MUST:
1. Confirm the assertion is actually unsafe, not guarded by a type switch
2. Check if the interface value is narrowed by runtime checks
3. Verify if the caller guarantees the concrete type

Valid patterns often flagged incorrectly:
```go
// Type switch makes each branch safe
switch v := val.(type) {
case string:
    fmt.Println(v)
}

// Comma-ok pattern handles the failure case
if s, ok := val.(string); ok {
    fmt.Println(s)
}
```

### "Potential Memory Leak/Race Condition"

Before flagging, you MUST:
1. Verify cleanup is actually missing (not in a defer or different location)
2. Check if context cancellation is propagated correctly
3. Confirm the goroutine can actually outlive its parent scope

## Severity Calibration

| Severity | Use For |
|----------|---------|
| **Critical** (block merge) | Security vulnerabilities, data corruption, crash-causing bugs, breaking API changes |
| **Major** (should fix) | Logic bugs, missing error handling causing poor UX, measurable performance issues |
| **Minor** (consider) | Code clarity, documentation gaps, non-critical test coverage |
| **Do NOT flag** | Style preferences, unmeasurable optimizations, test code simplicity, generated code |

## Valid Patterns (Do NOT Flag)

- `_ = err` with reason comment — intentionally ignored errors
- `//nolint` directives with explanation
- Channel without close when consumer stops via context cancellation
- Naked returns in functions < 5 lines with named returns
- `+?` lazy quantifier in regex — prevents over-matching
- Multiple returns in function — can improve readability

## Context-Sensitive Rules

| Issue | Flag ONLY IF |
|-------|--------------|
| Missing error check | Error return is actionable |
| Goroutine leak | No context cancellation path exists |
| Missing defer | Resource isn't explicitly closed before next acquisition or return |
| Interface pollution | Interface has > 1 method AND only one consumer |
| Missing try/catch | No error boundary at higher level AND crash would result |

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

## Before Submitting Review

1. Re-read each finding and ask: "Did I verify this is actually an issue?"
2. For each finding, can you point to the specific line that proves it?
3. Would a domain expert agree this is a problem, or is it a style preference?
4. Does fixing this provide real value, or is it busywork?

If uncertain: remove the finding, mark as a question, or read more context.
