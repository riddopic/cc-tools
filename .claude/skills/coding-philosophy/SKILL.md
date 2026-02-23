---
name: coding-philosophy
description: Decision framework and execution guidelines for writing code. Use when designing features (LEVER), evaluating approaches, writing/reviewing code (Karpathy), or making build-vs-reuse decisions.
---

# Coding Philosophy

## Part 1: Decision Framework (LEVER)

**Core Mantra**: "The best code is no code. The second best code is code that already exists and works."

### The LEVER Principles

```
L - Leverage existing patterns (use what works)
E - Extend before creating (build on existing)
V - Verify through reactivity (self-validating systems)
E - Eliminate duplication (of knowledge, not just code)
R - Reduce complexity (simplest solution wins)
```

### Quick Decision Guide

Before writing new code:

1. **Leverage**: Does the standard library solve this? Does an existing internal package?
   > **Tip:** The `search-first` skill provides a systematic workflow for the Leverage step.
2. **Extend**: Can we extend existing code rather than create new?
3. **Verify**: Will this be self-validating through reactive patterns?
4. **Eliminate**: Am I duplicating business knowledge?
5. **Reduce**: Is this the simplest solution?

### L - Leverage Existing Patterns

```go
// WRONG: Custom context management
type CustomContext struct {
    data map[string]interface{}
}

// RIGHT: Leverage context.Context
import "context"

func ProcessRequest(ctx context.Context, userID string) error {
    ctx = context.WithValue(ctx, "userID", userID)
    return doWork(ctx)
}
```

### E - Extend Before Creating

```go
// RIGHT: Extend existing interfaces through embedding
type EnhancedLogger interface {
    slog.Logger  // Embed standard logger
    WithContext(ctx context.Context) EnhancedLogger
}

// RIGHT: Extend base types through embedding
type User struct {
    BaseEntity        // Embedded struct
    Email string
}
```

### V - Verify Through Reactivity

```go
// RIGHT: Self-verifying with context cancellation
func (sm *StatusMonitor) Start() {
    for {
        select {
        case <-sm.ctx.Done():
            return
        case <-ticker.C:
            sm.notifyWatchers(sm.getCurrentStatus())
        }
    }
}
```

### E - Eliminate Duplication

```go
// WRONG: Same business rule in multiple places
func calculateShipping1() { if total > 50 { ... } }
func calculateShipping2() { if total > 50 { ... } }

// RIGHT: Single source of truth
const FreeShippingThreshold = 50.0
type ShippingCalculator struct{}
func (sc *ShippingCalculator) Calculate(total float64) float64 { ... }
```

**Note**: Duplicate _code_ is acceptable if it represents different _knowledge_.

### R - Reduce Complexity

```go
// WRONG: Deeply nested
if user != nil {
    if user.IsActive {
        if user.Subscription != nil {
            // more nesting...
        }
    }
}

// RIGHT: Early returns
if user == nil { return "login-required" }
if !user.IsActive { return "activate-account" }
if user.Subscription == nil { return "create-subscription" }
```

For complete examples, see [lever-framework.md](../../../docs/examples/philosophy/lever-framework.md)

---

## Part 2: Execution Guidelines (Karpathy)

Behavioral guidelines to reduce common LLM coding mistakes, derived from [Andrej Karpathy's observations](https://x.com/karpathy/status/2015883857489522876).

**Tradeoff:** These guidelines bias toward caution over speed. For trivial tasks, use judgment.

### 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:

- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

### 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

### 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:

- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:

- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

### 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:

- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:

```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.
