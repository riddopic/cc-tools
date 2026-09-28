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
		resp, err := r.dispatchOne(ctx, h, input)
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

		merged.Stdout = mergeOutput(merged.Stdout, resp.Stdout)

		if resp.Stderr != "" {
			merged.Stderr += resp.Stderr
		}
	}

	return merged
}

// mergeOutput combines two handlers' stdout. Claude Code reads a single JSON
// object per hook invocation, so dropping all but the first handler's output
// would silently discard context and messages from the rest.
func mergeOutput(a, b *HookOutput) *HookOutput {
	if a == nil {
		return b
	}

	if b == nil {
		return a
	}

	return &HookOutput{
		SystemMessage:      joinNonEmpty("\n", a.SystemMessage, b.SystemMessage),
		HookSpecificOutput: mergeSpecific(a.HookSpecificOutput, b.HookSpecificOutput),
	}
}

func mergeSpecific(a, b *HookSpecificOutput) *HookSpecificOutput {
	if a == nil {
		return b
	}

	if b == nil {
		return a
	}

	return &HookSpecificOutput{
		HookEventName:     a.HookEventName,
		AdditionalContext: joinNonEmpty("\n\n", a.AdditionalContext, b.AdditionalContext),
	}
}

func joinNonEmpty(sep, a, b string) string {
	switch {
	case a == "":
		return b
	case b == "":
		return a
	default:
		return a + sep + b
	}
}

// dispatchOne calls a single handler with panic recovery. If the handler
// panics, the panic value is captured and returned as an error.
//
//nolint:nonamedreturns // named returns required for defer/recover to assign err
func (r *Registry) dispatchOne(
	ctx context.Context, h Handler, input *hookcmd.HookInput,
) (resp *Response, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("panic: %v", p)
		}
	}()

	return h.Handle(ctx, input)
}
