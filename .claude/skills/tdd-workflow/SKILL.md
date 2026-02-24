---
name: tdd-workflow
description: Apply Test-Driven Development workflow. Use when implementing new features, writing production code, fixing bugs, or when the user asks about TDD practices. TDD is mandatory in this project.
---

# Test-Driven Development Workflow

**TDD IS MANDATORY** — Every line of production code must be written in response to a failing test.

## The Iron Law

```
NO PRODUCTION CODE WITHOUT A FAILING TEST FIRST
```

Write code before the test? Delete it. Start over. No exceptions.

## Red-Green-Refactor Cycle

```
RED     → Write a failing test that describes desired behavior
GREEN   → Write MINIMUM code to make the test pass
REFACTOR → Clean up (only if it adds value), keep tests green
COMMIT  → Save working code before moving on
```

### RED — Write Failing Test

Write one minimal test showing what should happen. One behavior, clear name, real code.

```go
func TestRetryOperation_RetriesFailedOperations3Times(t *testing.T) {
    attempts := 0
    operation := func() (string, error) {
        attempts++
        if attempts < 3 {
            return "", errors.New("fail")
        }
        return "success", nil
    }

    result, err := RetryOperation(operation)

    require.NoError(t, err)
    assert.Equal(t, "success", result)
    assert.Equal(t, 3, attempts)
}
```

### Verify RED — Watch It Fail (MANDATORY)

```bash
gotestsum --format pkgname -- -tags=testmode -run TestName ./path/to/package/...
```

Confirm: test fails (not errors), failure is expected, fails because feature is missing.

### GREEN — Minimal Code

Write the simplest code to pass. Don't add features, options, or "improvements" beyond the test.

### Verify GREEN (MANDATORY)

Run test again. Confirm it passes and other tests still pass.

### REFACTOR — Clean Up

After green only: remove duplication, improve names, extract helpers. Keep tests green. Don't add behavior.

### Repeat

Next failing test for next behavior.

## Bug Fix Flow

Bug found → write failing test reproducing it → follow TDD cycle. Test proves fix and prevents regression. Never fix bugs without a test.

## Common Rationalizations (All Mean: Start Over with TDD)

| Excuse | Reality |
|--------|---------|
| "Too simple to test" | Simple code breaks. Test takes 30 seconds. |
| "I'll test after" | Tests passing immediately prove nothing. |
| "Need to explore first" | Fine. Throw away exploration, start with TDD. |
| "Test hard = design unclear" | Hard to test = hard to use. Simplify interface. |
| "Already manually tested" | Ad-hoc ≠ systematic. No record, can't re-run. |

## Verification Checklist

- [ ] Every new function/method has a test
- [ ] Watched each test fail before implementing
- [ ] Wrote minimal code to pass each test
- [ ] All tests pass
- [ ] Tests use real code (mocks only if unavoidable)
- [ ] Edge cases and errors covered

Can't check all boxes? You skipped TDD. Start over.
