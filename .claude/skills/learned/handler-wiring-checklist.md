# Handler Wiring Checklist

**Extracted:** 2026-02-16
**Context:** Adding new notification backends or event handlers to cc-tools

## Problem

A backend implementation (e.g., `NtfyNotifier` in `internal/notify/`) can exist and be fully functional, yet produce no effect because it was never wrapped in a handler and registered in the dispatch system. This causes silent failures — no errors, no output, just missing behavior.

## Solution

When adding any new notification backend or event handler, follow this checklist:

1. **Backend exists** — verify the implementation in `internal/notify/` (or relevant package)
2. **Handler wrapper exists** — create a `Notify*Handler` struct in `internal/handler/notification.go` that:
   - Implements the `Handler` interface (Name + Handle methods)
   - Has a compile-time check: `var _ Handler = (*NotifyFooHandler)(nil)`
   - Accepts config and checks enabled/configured guards
   - Respects quiet hours via `notify.QuietHours`
   - Uses dependency injection interface for testability (e.g., `NtfySender`)
   - Creates the real backend when no mock is injected
3. **Handler registered** — add `NewNotifyFooHandler(cfg)` to the appropriate event in `internal/handler/defaults.go`
4. **Tests written** — cover nil config, disabled, quiet hours, mock injection, default/custom title+message
5. **Binary rebuilt** — `task build` after all changes

## Example

```go
// Step 1: Backend exists (internal/notify/ntfy.go)
// Step 2: Handler wrapper (internal/handler/notification.go)
var _ Handler = (*NotifyNtfyHandler)(nil)

// Step 3: Registration (internal/handler/defaults.go)
r.Register(hookcmd.EventNotification,
    NewNotifyNtfyHandler(cfg),
)
```

## When to Use

- Adding a new notification channel (Slack, Telegram, webhook, etc.)
- Adding a new event handler for any hook event
- Debugging "why isn't X working?" when the backend code looks correct
