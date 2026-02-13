# Development Philosophy

This directory contains the fundamental development philosophy and principles that guide the quanta project. These documents establish the foundational practices that enable high-quality, maintainable Go code.

## Philosophy Documents

### 📖 [TDD Principles](tdd-principles.md)

Core Test-Driven Development principles adapted for Go development.

**Key Topics:**

- The sacred Red-Green-Refactor cycle
- Why TDD is about design, not testing
- Common TDD violations and how to avoid them
- The TDD mindset and best practices
- Go-specific TDD examples using testify and table-driven tests

**When to Reference:**

- Starting any new feature or component
- When unsure about testing approach
- During code reviews to ensure TDD compliance
- When onboarding new team members

### 🎯 [LEVER Framework](lever-framework.md)

The LEVER Framework for building maintainable, scalable Go software.

**Framework Components:**

- **L**everage existing patterns and libraries
- **E**xtend before creating from scratch
- **V**erify through reactive patterns
- **E**liminate knowledge duplication
- **R**educe complexity at every opportunity

**When to Reference:**

- Architecture and design decisions
- Choosing between implementation approaches
- Refactoring existing code
- Code review discussions about complexity

### 🔄 [TDD Workflow](tdd-workflow.md)

Concrete examples of TDD workflow with detailed Go implementations.

**Covered Examples:**

- StatusLine renderer development
- Configuration loading with error handling
- Async service implementation
- Common anti-patterns to avoid
- StatusLine-specific development patterns

**When to Reference:**

- Learning TDD workflow step-by-step
- Understanding how to apply TDD to different scenarios
- Examples of proper test structure and organization
- Guidance on refactoring decisions

## Philosophy Integration

These principles work together to create a cohesive development approach:

1. **Start with TDD** - Every feature begins with a failing test
2. **Apply LEVER** - Use the framework to guide implementation decisions
3. **Follow the Workflow** - Use concrete examples to maintain consistency

## Relationship to Other Documentation

### 📋 [docs/CODING_GUIDELINES.md](../../docs/CODING_GUIDELINES.md)

The philosophy documents provide the **why** behind the coding guidelines. The coding guidelines provide the **how** and **what**.

### 🛠️ [docs/examples/patterns/](../patterns/)

Philosophy guides the selection and application of the implementation patterns.

### 📐 [docs/examples/standards/](../standards/)

Philosophy informs the creation and evolution of coding standards.

## Quick Reference

### Before Writing Code

1. **Write a failing test first** (TDD Principles)
2. **Check if existing patterns apply** (LEVER - Leverage)
3. **Consider extending rather than creating** (LEVER - Extend)

### During Implementation

1. **Write minimum code to pass** (TDD Workflow)
2. **Make it work, then make it simple** (LEVER - Reduce)
3. **Avoid knowledge duplication** (LEVER - Eliminate)

### After Implementation

1. **Assess refactoring opportunities** (TDD Workflow)
2. **Verify through reactive patterns** (LEVER - Verify)
3. **Commit and move to next increment** (TDD Workflow)

## Compliance Checking

Use these questions to ensure philosophy compliance:

### TDD Compliance

- [ ] Was the first line of code written to make a failing test pass?
- [ ] Are we writing the minimum code necessary?
- [ ] Have we assessed refactoring opportunities?
- [ ] Are tests focused on behavior, not implementation?

### LEVER Compliance

- [ ] Did we check for existing patterns first?
- [ ] Are we extending rather than recreating?
- [ ] Is the system self-verifying?
- [ ] Have we eliminated knowledge duplication?
- [ ] Is this the simplest solution that works?

## Examples in Context

### StatusLine Development

```go
// Philosophy-Driven Development Example
func TestStatusLineRenderer(t *testing.T) {
    // TDD: Start with failing test
    t.Run("should render session status", func(t *testing.T) {
        renderer := NewRenderer()
        status := Status{SessionID: "abc123", Active: true}

        result := renderer.Render(context.Background(), status)

        // LEVER: Verify through assertions
        assert.Contains(t, result, "abc123")
        assert.Contains(t, result, "●")
    })
}

// LEVER: Leverage existing patterns (context.Context)
// LEVER: Extend interfaces rather than creating from scratch
type Renderer interface {
    fmt.Stringer // Extend standard interface
    Render(ctx context.Context, status Status) string
}

// TDD: Minimum implementation to pass test
func (r *renderer) Render(ctx context.Context, status Status) string {
    // LEVER: Reduce complexity - simple implementation
    indicator := "○"
    if status.Active {
        indicator = "●"
    }
    return fmt.Sprintf("%s %s", status.SessionID, indicator)
}
```

## Best Practices Summary

1. **Never write production code without a failing test**
2. **Always look for existing patterns before creating new ones**
3. **Extend existing solutions when they're close but not perfect**
4. **Build systems that verify themselves through reactive patterns**
5. **Eliminate duplication of knowledge, not just code**
6. **Choose the simplest solution that solves the problem**
7. **Commit working code before refactoring**
8. **Refactor only when it adds clear value**

## Contributing to Philosophy

When updating these documents:

1. **Provide concrete Go examples** - Abstract principles need practical application
2. **Show both right and wrong approaches** - Contrast helps understanding
3. **Reference StatusLine context** - Keep examples relevant to our domain
4. **Update all related documentation** - Maintain consistency across the project
5. **Test examples compile and run** - Philosophy should be practical, not theoretical

## Related Resources

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Proverbs](https://go-proverbs.github.io/)
- [The Go Programming Language Specification](https://golang.org/ref/spec)
