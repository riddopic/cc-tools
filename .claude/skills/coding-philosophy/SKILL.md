---
name: coding-philosophy
description: LEVER decision framework for evaluating approaches and making build-vs-reuse decisions. Use when designing features or choosing between implementation strategies.
---

# Coding Philosophy — LEVER Framework

**Core Mantra**: "The best code is no code. The second best code is code that already exists and works."

## The LEVER Principles

```
L - Leverage existing patterns (use what works)
E - Extend before creating (build on existing)
V - Verify through reactivity (self-validating systems)
E - Eliminate duplication (of knowledge, not just code)
R - Reduce complexity (simplest solution wins)
```

## Quick Decision Guide

Before writing new code:

1. **Leverage**: Does the standard library solve this? Does an existing internal package?
   > **Tip:** The `search-first` skill provides a systematic workflow for the Leverage step.
2. **Extend**: Can we extend existing code rather than create new?
3. **Verify**: Will this be self-validating through reactive patterns?
4. **Eliminate**: Am I duplicating business knowledge?
5. **Reduce**: Is this the simplest solution?

## Applying LEVER

```go
// L - Leverage: Use context.Context, not custom context
func ProcessRequest(ctx context.Context, userID string) error {
    return doWork(ctx)
}

// E - Extend: Embed existing interfaces
type EnhancedLogger interface {
    slog.Logger
    WithContext(ctx context.Context) EnhancedLogger
}

// E - Eliminate: Single source of truth
const FreeShippingThreshold = 50.0
type ShippingCalculator struct{}

// R - Reduce: Early returns over deep nesting
if user == nil { return "login-required" }
if !user.IsActive { return "activate-account" }
```

**Note**: Duplicate _code_ is acceptable if it represents different _knowledge_.

For complete examples, see [lever-framework.md](../../../docs/examples/philosophy/lever-framework.md)
