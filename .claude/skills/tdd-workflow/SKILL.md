---
name: tdd-workflow
description: Apply Test-Driven Development workflow. Use when implementing new features, writing production code, fixing bugs, or when the user asks about TDD practices. TDD is mandatory in this project.
---

# Test-Driven Development Workflow

**TDD IS MANDATORY** - Every line of production code must be written in response to a failing test.

## The Sacred Cycle: Red-Green-Refactor

```
RED     -> Write a failing test that describes desired behavior
GREEN   -> Write MINIMUM code to make the test pass
REFACTOR -> Assess if code can be improved (only if it adds value)
COMMIT  -> Save working code before moving on
```

## TDD Checklist Before Writing Code

- [ ] Do I have a failing test that demands this code?
- [ ] Have I run the test and seen it FAIL?
- [ ] Am I writing the minimum code to make the test pass?
- [ ] Have I committed my working code before refactoring?

## Critical Rules

### 1. No Production Code Without a Failing Test

```go
// WRONG: Writing implementation first
func CalculateDiscount(price float64, tier string) float64 {
    // implementation
}

// RIGHT: Start with test
func TestCalculateDiscount(t *testing.T) {
    t.Run("should apply 20% discount for premium tier", func(t *testing.T) {
        result := CalculateDiscount(100.0, "premium")
        assert.Equal(t, 80.0, result)
    })
}
// NOW write the minimum code to pass
```

### 2. Write the Minimum Code to Pass

```go
// Test demands: return user by id
func TestFindUserByID(t *testing.T) {
    user, err := FindUserByID("123")
    assert.NoError(t, err)
    assert.Equal(t, "123", user.ID)
}

// MINIMUM implementation - don't add extras!
func FindUserByID(id string) (*User, error) {
    return &User{ID: id}, nil
}
```

### 3. Small Steps Win

```go
// WRONG: Too much at once
t.Run("should process order with discounts, tax, shipping", ...)

// RIGHT: One step at a time
t.Run("should calculate subtotal", ...)
// Make it pass, commit

t.Run("should apply discount", ...)
// Make it pass, commit
```

## Common Violations to Avoid

1. **Writing code "while you're there"** - Don't add untested features
2. **Writing multiple tests before going green** - One test at a time
3. **Skipping refactor assessment** - Always evaluate after green

## Detailed TDD Principles

For core TDD principles, see [tdd-principles.md](../../../docs/examples/philosophy/tdd-principles.md)
For TDD workflow examples, see [tdd-workflow.md](../../../docs/examples/philosophy/tdd-workflow.md)
