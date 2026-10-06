package handler_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/riddopic/cc-tools/internal/skipregistry"
)

// countingHandler is a fake advisory handler that records how often it runs.
type countingHandler struct {
	resp  *handler.Response
	err   error
	calls int
}

func (c *countingHandler) Name() string { return "fake-advisory" }

func (c *countingHandler) Handle(_ context.Context, _ *hookcmd.HookInput) (*handler.Response, error) {
	c.calls++
	return c.resp, c.err
}

// memSkipStorage serves fixed skip registry data.
type memSkipStorage struct {
	data skipregistry.RegistryData
	err  error
}

func (m *memSkipStorage) Load(_ context.Context) (skipregistry.RegistryData, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.data == nil {
		return skipregistry.RegistryData{}, nil
	}
	return m.data, nil
}

func (m *memSkipStorage) Save(_ context.Context, _ skipregistry.RegistryData) error { return nil }

func envOf(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

func emptySkipReader() skipregistry.Reader {
	return skipregistry.NewRegistry(&memSkipStorage{data: nil, err: nil})
}

// newRepoTree creates a repo root with a nested directory and returns both.
func newRepoTree(t *testing.T) (string, string) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	require.NoError(t, os.Mkdir(filepath.Join(root, ".git"), 0o750))
	nested := filepath.Join(root, "src", "pkg")
	require.NoError(t, os.MkdirAll(nested, 0o750))
	return root, nested
}

func TestAdvisoryHandler_SuppressesWhenUnattended(t *testing.T) {
	t.Parallel()

	root, nested := newRepoTree(t)

	tests := []struct {
		name   string
		env    map[string]string
		input  hookcmd.HookInput
		reader skipregistry.Reader
	}{
		{
			name:   "CC_TOOLS_UNATTENDED=1",
			env:    map[string]string{"CC_TOOLS_UNATTENDED": "1"},
			input:  hookcmd.HookInput{Cwd: nested},
			reader: emptySkipReader(),
		},
		{
			name:   "AMS_UNATTENDED=true",
			env:    map[string]string{"AMS_UNATTENDED": "true"},
			input:  hookcmd.HookInput{Cwd: nested},
			reader: emptySkipReader(),
		},
		{
			name:   "subagent agent_id",
			env:    nil,
			input:  hookcmd.HookInput{Cwd: nested, AgentID: "a-123"},
			reader: emptySkipReader(),
		},
		{
			name:   "subagent agent_type",
			env:    nil,
			input:  hookcmd.HookInput{Cwd: nested, AgentType: "Explore"},
			reader: emptySkipReader(),
		},
		{
			name:  "nudges skipped on an ancestor directory",
			env:   nil,
			input: hookcmd.HookInput{Cwd: nested},
			reader: skipregistry.NewRegistry(&memSkipStorage{
				data: skipregistry.RegistryData{root: {"nudges"}},
				err:  nil,
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			inner := &countingHandler{resp: handler.UserMessageResponse("drift warning"), err: nil, calls: 0}
			h := handler.NewAdvisoryHandler(inner,
				handler.WithAdvisoryGetenv(envOf(tt.env)),
				handler.WithAdvisorySkipReader(tt.reader),
			)

			input := tt.input
			resp, err := h.Handle(context.Background(), &input)

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, 0, resp.ExitCode)
			assert.Nil(t, resp.Stdout, "nudge must be suppressed")
			assert.Empty(t, resp.SystemMessage())
			assert.Equal(t, 1, inner.calls, "inner handler still runs so counters advance")
		})
	}
}

func TestAdvisoryHandler_InteractivePassesThrough(t *testing.T) {
	t.Parallel()

	_, nested := newRepoTree(t)

	tests := []struct {
		name string
		env  map[string]string
		resp *handler.Response
	}{
		{
			name: "user message",
			env:  nil,
			resp: handler.UserMessageResponse("drift warning"),
		},
		{
			name: "context",
			env:  nil,
			resp: handler.ContextResponse(hookcmd.EventPreToolUse, "consider compacting"),
		},
		{
			name: "falsy env value",
			env:  map[string]string{"CC_TOOLS_UNATTENDED": "0", "AMS_UNATTENDED": "no"},
			resp: handler.UserMessageResponse("stop reminder"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			inner := &countingHandler{resp: tt.resp, err: nil, calls: 0}
			h := handler.NewAdvisoryHandler(inner,
				handler.WithAdvisoryGetenv(envOf(tt.env)),
				handler.WithAdvisorySkipReader(emptySkipReader()),
			)

			resp, err := h.Handle(context.Background(), &hookcmd.HookInput{Cwd: nested})

			require.NoError(t, err)
			assert.Same(t, tt.resp, resp)
			assert.Equal(t, 1, inner.calls)
		})
	}
}

func TestAdvisoryHandler_NoOutputSkipsTriggerChecks(t *testing.T) {
	t.Parallel()

	failEnv := func(key string) string {
		t.Errorf("getenv(%q) called with nothing to suppress", key)
		return ""
	}

	for _, resp := range []*handler.Response{nil, {ExitCode: 0, Stdout: nil, Stderr: ""}} {
		inner := &countingHandler{resp: resp, err: nil, calls: 0}
		h := handler.NewAdvisoryHandler(inner, handler.WithAdvisoryGetenv(failEnv))

		got, err := h.Handle(context.Background(), &hookcmd.HookInput{AgentID: "a-1"})

		require.NoError(t, err)
		assert.Same(t, resp, got)
		assert.Equal(t, 1, inner.calls)
	}
}

func TestAdvisoryHandler_InnerErrorReturnedUnchanged(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")
	inner := &countingHandler{resp: handler.UserMessageResponse("x"), err: wantErr, calls: 0}
	h := handler.NewAdvisoryHandler(inner,
		handler.WithAdvisoryGetenv(envOf(map[string]string{"CC_TOOLS_UNATTENDED": "1"})),
	)

	_, err := h.Handle(context.Background(), &hookcmd.HookInput{})

	require.ErrorIs(t, err, wantErr)
}

func TestAdvisoryHandler_RegistryErrorMeansAttended(t *testing.T) {
	t.Parallel()

	_, nested := newRepoTree(t)
	want := handler.UserMessageResponse("drift warning")
	inner := &countingHandler{resp: want, err: nil, calls: 0}
	h := handler.NewAdvisoryHandler(inner,
		handler.WithAdvisoryGetenv(envOf(nil)),
		handler.WithAdvisorySkipReader(skipregistry.NewRegistry(&memSkipStorage{
			data: nil,
			err:  errors.New("unreadable"),
		})),
	)

	resp, err := h.Handle(context.Background(), &hookcmd.HookInput{Cwd: nested})

	require.NoError(t, err)
	assert.Same(t, want, resp)
}

func TestAdvisoryHandler_NameDelegates(t *testing.T) {
	t.Parallel()

	inner := &countingHandler{resp: nil, err: nil, calls: 0}
	assert.Equal(t, "fake-advisory", handler.NewAdvisoryHandler(inner).Name())
}
