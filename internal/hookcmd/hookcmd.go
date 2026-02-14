package hookcmd

import (
	"context"
	"io"
)

// Dispatch routes a hook event to registered handlers.
// Returns the exit code (always 0 -- errors are logged, not fatal).
func Dispatch(ctx context.Context, input *HookInput, out, errOut io.Writer, registry map[string][]Handler) int {
	handlers, ok := registry[input.HookEventName]
	if !ok {
		// Unknown event type -- accept gracefully.
		return 0
	}

	RunHandlers(ctx, input, handlers, out, errOut)

	return 0
}
