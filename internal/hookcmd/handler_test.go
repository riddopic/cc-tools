package hookcmd_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/hookcmd"
)

// testHandler implements hookcmd.Handler for testing.
type testHandler struct {
	name  string
	runFn func() error
}

func (h *testHandler) Name() string { return h.name }

func (h *testHandler) Run(_ context.Context, _ *hookcmd.HookInput, _ io.Writer, _ io.Writer) error {
	return h.runFn()
}

func TestRunHandlers(t *testing.T) {
	tests := []struct {
		name           string
		handlers       []hookcmd.Handler
		wantErrOut     string
		expectNoErrOut bool
	}{
		{
			name: "sequential execution in order",
			handlers: func() []hookcmd.Handler {
				return []hookcmd.Handler{
					&testHandler{
						name: "first",
						runFn: func() error {
							return nil
						},
					},
					&testHandler{
						name: "second",
						runFn: func() error {
							return nil
						},
					},
					&testHandler{
						name: "third",
						runFn: func() error {
							return nil
						},
					},
				}
			}(),
			wantErrOut:     "",
			expectNoErrOut: true,
		},
		{
			name: "continues after error",
			handlers: func() []hookcmd.Handler {
				return []hookcmd.Handler{
					&testHandler{
						name: "failing",
						runFn: func() error {
							return errors.New("something broke")
						},
					},
					&testHandler{
						name: "succeeding",
						runFn: func() error {
							return nil
						},
					},
				}
			}(),
			wantErrOut:     "[failing] error: something broke",
			expectNoErrOut: false,
		},
		{
			name: "recovers from panic and continues",
			handlers: func() []hookcmd.Handler {
				return []hookcmd.Handler{
					&testHandler{
						name: "panicking",
						runFn: func() error {
							panic("unexpected panic")
						},
					},
					&testHandler{
						name: "after-panic",
						runFn: func() error {
							return nil
						},
					},
				}
			}(),
			wantErrOut:     "[panicking] panic recovered: unexpected panic",
			expectNoErrOut: false,
		},
		{
			name:           "empty handler list does nothing",
			handlers:       []hookcmd.Handler{},
			wantErrOut:     "",
			expectNoErrOut: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			ctx := context.Background()
			input := &hookcmd.HookInput{}

			hookcmd.RunHandlers(ctx, input, tt.handlers, &out, &errOut)

			if tt.expectNoErrOut {
				assert.Empty(t, errOut.String())
			}
			if tt.wantErrOut != "" {
				assert.Contains(t, errOut.String(), tt.wantErrOut)
			}
		})
	}
}

func TestRunHandlersExecutionOrder(t *testing.T) {
	var order []string
	handlers := []hookcmd.Handler{
		&testHandler{
			name: "first",
			runFn: func() error {
				order = append(order, "first")
				return nil
			},
		},
		&testHandler{
			name: "second",
			runFn: func() error {
				order = append(order, "second")
				return nil
			},
		},
		&testHandler{
			name: "third",
			runFn: func() error {
				order = append(order, "third")
				return nil
			},
		},
	}

	var out, errOut bytes.Buffer
	ctx := context.Background()
	input := &hookcmd.HookInput{}

	hookcmd.RunHandlers(ctx, input, handlers, &out, &errOut)

	require.Len(t, order, 3)
	assert.Equal(t, []string{"first", "second", "third"}, order)
}
