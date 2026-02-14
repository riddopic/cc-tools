package hookcmd_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/riddopic/cc-tools/internal/hookcmd"
)

func TestDispatch(t *testing.T) {
	tests := []struct {
		name       string
		input      *hookcmd.HookInput
		registry   map[string][]hookcmd.Handler
		wantExit   int
		wantErrOut string
	}{
		{
			name: "unknown event exits cleanly",
			input: &hookcmd.HookInput{
				HookEventName: "UnknownEvent",
			},
			registry: map[string][]hookcmd.Handler{
				"PreToolUse": {
					&testHandler{
						name:  "pre-tool",
						runFn: func() error { return nil },
					},
				},
			},
			wantExit:   0,
			wantErrOut: "",
		},
		{
			name: "routes to correct handlers",
			input: &hookcmd.HookInput{
				HookEventName: "PreToolUse",
			},
			registry: func() map[string][]hookcmd.Handler {
				return map[string][]hookcmd.Handler{
					"PreToolUse": {
						&testHandler{
							name:  "pre-handler",
							runFn: func() error { return nil },
						},
					},
					"PostToolUse": {
						&testHandler{
							name: "post-handler",
							runFn: func() error {
								// This should not be called.
								panic("wrong handler called")
							},
						},
					},
				}
			}(),
			wantExit:   0,
			wantErrOut: "",
		},
		{
			name: "nil registry exits cleanly",
			input: &hookcmd.HookInput{
				HookEventName: "PreToolUse",
			},
			registry:   nil,
			wantExit:   0,
			wantErrOut: "",
		},
		{
			name: "empty handlers list for event",
			input: &hookcmd.HookInput{
				HookEventName: "PreToolUse",
			},
			registry: map[string][]hookcmd.Handler{
				"PreToolUse": {},
			},
			wantExit:   0,
			wantErrOut: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			ctx := context.Background()

			exitCode := hookcmd.Dispatch(ctx, tt.input, &out, &errOut, tt.registry)

			assert.Equal(t, tt.wantExit, exitCode)
			if tt.wantErrOut != "" {
				assert.Contains(t, errOut.String(), tt.wantErrOut)
			}
		})
	}
}

func TestDispatchRoutesToCorrectHandler(t *testing.T) {
	var called bool
	registry := map[string][]hookcmd.Handler{
		"PreToolUse": {
			&testHandler{
				name: "tracker",
				runFn: func() error {
					called = true
					return nil
				},
			},
		},
	}

	input := &hookcmd.HookInput{
		HookEventName: "PreToolUse",
	}

	var out, errOut bytes.Buffer
	ctx := context.Background()

	exitCode := hookcmd.Dispatch(ctx, input, &out, &errOut, registry)

	assert.Equal(t, 0, exitCode)
	assert.True(t, called, "expected handler to be called")
}
