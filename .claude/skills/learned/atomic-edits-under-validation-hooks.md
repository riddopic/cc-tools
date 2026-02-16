# Atomic Edits Under Validation Hooks

**Extracted:** 2026-02-16
**Context:** When cc-tools validate runs as a PostToolUse hook on Edit/Write

## Problem

When a PostToolUse validation hook (like `cc-tools validate`) runs after every Edit or Write, incremental code changes that leave the file in a non-compilable state get blocked. For example, adding a compile-time interface check `var _ Handler = (*Foo)(nil)` without the `Foo` struct existing in the same file causes a compilation error, and the validation hook rejects the edit.

## Solution

When adding code that has cross-references within the same file (compile-time checks referencing new types, functions calling new helpers, etc.), use the Write tool to replace the entire file in a single operation rather than making multiple incremental Edit calls. This ensures every intermediate state compiles.

**Decision flow:**

1. If the change is self-contained (no new cross-references) → use Edit
2. If the change adds references to not-yet-existing code in the same file → use Write to include everything atomically
3. If Edit fails validation → revert with another Edit, then use Write for the full replacement

## Example

```text
# BAD: Two separate edits — first one fails validation
Edit 1: Add `var _ Handler = (*NotifyNtfyHandler)(nil)` to interface checks
  → BLOCKED: NotifyNtfyHandler undefined
Edit 2: Add NotifyNtfyHandler struct and methods
  → Never reached

# GOOD: Single Write with complete file
Write: Full file containing both the interface check AND the struct/methods
  → Compiles and passes validation
```

## When to Use

- Adding new types referenced by compile-time interface checks
- Adding helper functions called by code in the same file
- Any multi-part change within a single file when PostToolUse validation is active
