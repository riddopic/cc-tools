//go:build testmode

package handler_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
)

// stubHandler is a test handler that returns a fixed response.
type stubHandler struct {
	name string
	resp *handler.Response
	err  error
}

func (s *stubHandler) Name() string { return s.name }

func (s *stubHandler) Handle(_ context.Context, _ *hookcmd.HookInput) (*handler.Response, error) {
	return s.resp, s.err
}

// panicHandler is a test handler that panics when Handle is called.
type panicHandler struct {
	name string
	msg  string
}

func (p *panicHandler) Name() string { return p.name }

func (p *panicHandler) Handle(_ context.Context, _ *hookcmd.HookInput) (*handler.Response, error) {
	panic(p.msg)
}

func TestRegistry_Dispatch_NoHandlers(t *testing.T) {
	t.Parallel()
	r := handler.NewRegistry()
	input := &hookcmd.HookInput{HookEventName: hookcmd.EventSessionStart}

	resp := r.Dispatch(context.Background(), input)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
}

func TestRegistry_Dispatch_SingleHandler(t *testing.T) {
	t.Parallel()
	r := handler.NewRegistry()
	r.Register(hookcmd.EventSessionStart, &stubHandler{
		name: "test",
		resp: &handler.Response{
			ExitCode: 0,
			Stdout: &handler.HookOutput{
				Continue:      true,
				SystemMessage: "hello",
			},
		},
		err: nil,
	})

	input := &hookcmd.HookInput{HookEventName: hookcmd.EventSessionStart}
	resp := r.Dispatch(context.Background(), input)

	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
	require.NotNil(t, resp.Stdout)
	assert.Equal(t, "hello", resp.Stdout.SystemMessage)
}

func TestRegistry_Dispatch_MergesMultipleHandlers(t *testing.T) {
	t.Parallel()
	r := handler.NewRegistry()
	r.Register(hookcmd.EventSessionStart,
		&stubHandler{
			name: "first",
			resp: &handler.Response{
				ExitCode: 0,
				Stdout: &handler.HookOutput{
					SystemMessage: "from first",
				},
			},
			err: nil,
		},
		&stubHandler{
			name: "second",
			resp: &handler.Response{
				ExitCode: 0,
				Stderr:   "log from second\n",
			},
			err: nil,
		},
	)

	input := &hookcmd.HookInput{HookEventName: hookcmd.EventSessionStart}
	resp := r.Dispatch(context.Background(), input)

	require.NotNil(t, resp)
	// First handler's stdout wins.
	require.NotNil(t, resp.Stdout)
	assert.Equal(t, "from first", resp.Stdout.SystemMessage)
	// Stderr concatenated.
	assert.Contains(t, resp.Stderr, "log from second")
}

func TestRegistry_Dispatch_MaxExitCode(t *testing.T) {
	t.Parallel()
	r := handler.NewRegistry()
	r.Register(hookcmd.EventPreToolUse,
		&stubHandler{name: "pass", resp: &handler.Response{ExitCode: 0}, err: nil},
		&stubHandler{name: "block", resp: &handler.Response{ExitCode: 2, Stderr: "blocked"}, err: nil},
	)

	input := &hookcmd.HookInput{HookEventName: hookcmd.EventPreToolUse}
	resp := r.Dispatch(context.Background(), input)

	assert.Equal(t, 2, resp.ExitCode)
	assert.Contains(t, resp.Stderr, "blocked")
}

func TestRegistry_Dispatch_HandlerError(t *testing.T) {
	t.Parallel()
	r := handler.NewRegistry()
	r.Register(hookcmd.EventStop, &stubHandler{
		name: "broken",
		resp: nil,
		err:  assert.AnError,
	})

	input := &hookcmd.HookInput{HookEventName: hookcmd.EventStop}
	resp := r.Dispatch(context.Background(), input)

	// Errors are logged to stderr, not fatal.
	assert.Equal(t, 0, resp.ExitCode)
	assert.Contains(t, resp.Stderr, "[broken] error:")
}

func TestRegistry_Dispatch_NilResponse(t *testing.T) {
	t.Parallel()
	r := handler.NewRegistry()
	r.Register(hookcmd.EventNotification, &stubHandler{
		name: "silent",
		resp: nil,
		err:  nil,
	})

	input := &hookcmd.HookInput{HookEventName: hookcmd.EventNotification}
	resp := r.Dispatch(context.Background(), input)

	assert.Equal(t, 0, resp.ExitCode)
	assert.Nil(t, resp.Stdout)
}

func TestRegistry_Dispatch_PanicRecovery(t *testing.T) {
	t.Parallel()
	r := handler.NewRegistry()
	r.Register(hookcmd.EventPreToolUse,
		&panicHandler{name: "crasher", msg: "unexpected nil pointer"},
		&stubHandler{
			name: "normal",
			resp: &handler.Response{
				ExitCode: 0,
				Stdout: &handler.HookOutput{
					SystemMessage: "still here",
				},
			},
			err: nil,
		},
	)

	input := &hookcmd.HookInput{HookEventName: hookcmd.EventPreToolUse}
	resp := r.Dispatch(context.Background(), input)

	require.NotNil(t, resp)
	// The panicking handler's error is captured in stderr.
	assert.Contains(t, resp.Stderr, "[crasher] error:")
	assert.Contains(t, resp.Stderr, "panic:")
	assert.Contains(t, resp.Stderr, "unexpected nil pointer")
	// The normal handler's response is still included.
	require.NotNil(t, resp.Stdout)
	assert.Equal(t, "still here", resp.Stdout.SystemMessage)
}
