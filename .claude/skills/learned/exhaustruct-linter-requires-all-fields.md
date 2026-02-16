# exhaustruct Linter Requires All Struct Fields

**Extracted:** 2026-02-16
**Context:** Creating struct literals in cc-tools where golangci-lint has exhaustruct enabled

## Problem

The `exhaustruct` linter in this project's golangci-lint configuration requires every field in a struct literal to be explicitly set, even when the zero value is intended. Omitting fields causes a lint error like:

```
notify.NtfyConfig is missing fields Server, Token, Priority (exhaustruct)
```

This catches you when creating structs from external packages where you only care about a subset of fields.

## Solution

Always specify every field in struct literals, using explicit zero values for fields you don't need:

```go
// BAD — exhaustruct rejects this
sender = notify.NewNtfyNotifier(notify.NtfyConfig{
    Topic: h.cfg.Notifications.NtfyTopic,
})

// GOOD — all fields specified
sender = notify.NewNtfyNotifier(notify.NtfyConfig{
    Topic:    h.cfg.Notifications.NtfyTopic,
    Server:   "",
    Token:    "",
    Priority: 0,
})
```

The `NewNtfyNotifier` constructor applies defaults (server `https://ntfy.sh`, priority 3) when it sees zero values, so this is safe.

## When to Use

- Creating any struct literal in this project
- Especially when using structs from `internal/notify`, `internal/config`, or other packages with multiple fields
- After seeing an `exhaustruct` lint error
