# Bool Field Addition Breaks Existing Tests

**Extracted:** 2026-02-11
**Context:** Adding a new boolean field to a widely-used config struct

## Problem

Adding a new `bool` field to a Go struct silently changes behavior for all existing struct literals that don't set it. Go's zero value for `bool` is `false`, so if the new field controls a feature that should default to "enabled", every existing test config gets "disabled" behavior without any compiler error.

The `exhaustruct` linter only enforces field completeness for **new** struct literals written after the linter is configured. Existing test files that were passing before the field was added continue to compile — they just silently get `false` for the new field.

## Example

```go
// Adding this field to AgentConfig:
Gate25Enabled bool `json:"gate25_enabled"`

// Existing test configs DON'T set it:
cfg := &interfaces.AgentConfig{
    GatePipeline: true,
    // Gate25Enabled is implicitly false!
}

// Tests that expected Gate 2.5 to run now fail because it's skipped.
```

## Solution

After adding a `bool` field where `true` is the desired default:

1. **Grep ALL struct literals** constructing that type across the entire codebase:

   ```bash
   rg "AgentConfig\{" --type go -l
   ```

2. **Add the new field explicitly** to every literal, especially in test files.

3. **Consider `SetDefaults()`** — but note that `SetDefaults()` cannot distinguish "explicitly set to false" from "zero value false", so it can only safely default non-bool fields (ints, strings). For bools where `true` is the default, you must either:
   - Use a pointer (`*bool`) with nil = unset
   - Accept that every callsite must explicitly set it
   - Use the inverse naming (`Gate25Disabled bool`) so zero value = enabled

4. **Run the full test suite** immediately after adding the field — don't wait for lint.

## When to Use

- Adding any `bool` field to a struct used in 5+ places
- Adding a feature toggle where "enabled" is the default
- Any struct that has `exhaustruct` enforcement but has existing test files
