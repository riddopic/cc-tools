package handler

import (
	"context"
	"fmt"

	"github.com/riddopic/cc-tools/internal/hookcmd"
)

// Registry maps hook event names to handler slices.
type Registry struct {
	handlers map[string][]Handler
}

// NewRegistry creates an empty handler registry.
func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string][]Handler)}
}

// Register adds one or more handlers for the given event name.
func (r *Registry) Register(event string, handlers ...Handler) {
	r.handlers[event] = append(r.handlers[event], handlers...)
}

// Dispatch runs all handlers for the event and merges their responses.
// Unknown events return a zero-value Response (exit code 0, no output).
func (r *Registry) Dispatch(ctx context.Context, input *hookcmd.HookInput) *Response {
	handlers := r.handlers[input.HookEventName]
	if len(handlers) == 0 {
		return &Response{}
	}

	merged := &Response{}
	for _, h := range handlers {
		resp, err := h.Handle(ctx, input)
		if err != nil {
			merged.Stderr += fmt.Sprintf("[%s] error: %v\n", h.Name(), err)

			continue
		}

		if resp == nil {
			continue
		}

		if resp.ExitCode > merged.ExitCode {
			merged.ExitCode = resp.ExitCode
		}

		if resp.Stdout != nil && merged.Stdout == nil {
			merged.Stdout = resp.Stdout
		}

		if resp.Stderr != "" {
			merged.Stderr += resp.Stderr
		}
	}

	return merged
}
