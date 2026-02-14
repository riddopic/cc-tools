# cc-tools Architecture Redesign Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use executing-plans to implement this plan task-by-task.

**Goal:** Migrate cc-tools from manual `os.Args` parsing to Cobra CLI with a handler registry pattern, keeping all existing functionality intact.

**Architecture:** Bottom-up migration — build new internal/handler/ package alongside existing code, create Cobra commands, then swap the entry point. Existing internal packages (hooks, config, session, etc.) stay unchanged.

**Tech Stack:** Go 1.26, Cobra (CLI framework), lipgloss (styling), testify (testing)

**Design doc:** `docs/plans/2026-02-14-cc-tools-redesign-design.md`

**Note on config format:** The design doc specifies YAML config. This plan keeps JSON config unchanged — YAML migration is deferred to a follow-up since it's orthogonal to the CLI restructure and the existing JSON config has comprehensive tests.

---

### Task 1: Add Cobra Dependency and Event Constants

**Files:**
- Modify: `go.mod`
- Create: `internal/hookcmd/events.go`
- Create: `internal/hookcmd/events_test.go`

**Context:** Cobra replaces ~200 lines of manual `os.Args` parsing. Event constants prevent typos in hook event name strings scattered across main.go and hookcmd.go. Currently event names are bare strings like `"SessionStart"` in the `hookEventMap` in `cmd/cc-tools/main.go:123-132`.

**Step 1: Add Cobra dependency**

Run:
```bash
cd /Users/riddopic/git/cc-stuff/q2 && go get github.com/spf13/cobra@latest
```

Expected: go.mod updated with cobra direct dependency, go.sum updated.

**Step 2: Write failing test for event constants**

Create `internal/hookcmd/events_test.go`:

```go
package hookcmd_test

import (
	"testing"

	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/stretchr/testify/assert"
)

func TestEventConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"SessionStart", hookcmd.EventSessionStart, "SessionStart"},
		{"SessionEnd", hookcmd.EventSessionEnd, "SessionEnd"},
		{"PreToolUse", hookcmd.EventPreToolUse, "PreToolUse"},
		{"PostToolUse", hookcmd.EventPostToolUse, "PostToolUse"},
		{"PostToolUseFailure", hookcmd.EventPostToolUseFailure, "PostToolUseFailure"},
		{"PreCompact", hookcmd.EventPreCompact, "PreCompact"},
		{"Notification", hookcmd.EventNotification, "Notification"},
		{"UserPromptSubmit", hookcmd.EventUserPromptSubmit, "UserPromptSubmit"},
		{"PermissionRequest", hookcmd.EventPermissionRequest, "PermissionRequest"},
		{"Stop", hookcmd.EventStop, "Stop"},
		{"SubagentStart", hookcmd.EventSubagentStart, "SubagentStart"},
		{"SubagentStop", hookcmd.EventSubagentStop, "SubagentStop"},
		{"TeammateIdle", hookcmd.EventTeammateIdle, "TeammateIdle"},
		{"TaskCompleted", hookcmd.EventTaskCompleted, "TaskCompleted"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

func TestAllEvents_ReturnsAllConstants(t *testing.T) {
	t.Parallel()
	events := hookcmd.AllEvents()
	assert.Len(t, events, 14)
	assert.Contains(t, events, hookcmd.EventSessionStart)
	assert.Contains(t, events, hookcmd.EventStop)
}
```

**Step 3: Run test to verify it fails**

Run: `go test ./internal/hookcmd/ -run TestEventConstants -v`
Expected: FAIL — `hookcmd.EventSessionStart` undefined.

**Step 4: Implement event constants**

Create `internal/hookcmd/events.go`:

```go
package hookcmd

// Event name constants matching Claude Code hook event names.
// See: https://docs.anthropic.com/en/docs/claude-code/hooks
const (
	EventSessionStart       = "SessionStart"
	EventSessionEnd         = "SessionEnd"
	EventPreToolUse         = "PreToolUse"
	EventPostToolUse        = "PostToolUse"
	EventPostToolUseFailure = "PostToolUseFailure"
	EventPreCompact         = "PreCompact"
	EventNotification       = "Notification"
	EventUserPromptSubmit   = "UserPromptSubmit"
	EventPermissionRequest  = "PermissionRequest"
	EventStop               = "Stop"
	EventSubagentStart      = "SubagentStart"
	EventSubagentStop       = "SubagentStop"
	EventTeammateIdle       = "TeammateIdle"
	EventTaskCompleted      = "TaskCompleted"
)

// AllEvents returns all known hook event names.
func AllEvents() []string {
	return []string{
		EventSessionStart,
		EventSessionEnd,
		EventPreToolUse,
		EventPostToolUse,
		EventPostToolUseFailure,
		EventPreCompact,
		EventNotification,
		EventUserPromptSubmit,
		EventPermissionRequest,
		EventStop,
		EventSubagentStart,
		EventSubagentStop,
		EventTeammateIdle,
		EventTaskCompleted,
	}
}
```

**Step 5: Run tests to verify they pass**

Run: `go test ./internal/hookcmd/ -v`
Expected: ALL PASS (existing tests + new tests).

**Step 6: Run lint**

Run: `golangci-lint run ./internal/hookcmd/`
Expected: No issues.

**Step 7: Commit**

```bash
git add go.mod go.sum internal/hookcmd/events.go internal/hookcmd/events_test.go
git commit -m "feat: add Cobra dependency and hook event constants"
```

---

### Task 2: Create Handler Package Foundation

**Files:**
- Create: `internal/handler/handler.go`
- Create: `internal/handler/registry.go`
- Create: `internal/handler/handler_test.go`
- Create: `internal/handler/registry_test.go`

**Context:** The current `hookcmd.Handler` interface uses `Run(ctx, input, out, errOut io.Writer) error`. The new `handler.Handler` returns a structured `*Response` instead. This enables testing without io.Writer mocking and supports the full Claude Code hooks JSON output protocol (exit codes, `hookSpecificOutput`, `additionalContext`, etc.). Both interfaces coexist during migration — the old one is removed in a later task.

**Step 1: Write failing tests for Handler types**

Create `internal/handler/handler_test.go`:

```go
package handler_test

import (
	"encoding/json"
	"testing"

	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponse_ZeroValue(t *testing.T) {
	t.Parallel()
	resp := &handler.Response{}
	assert.Equal(t, 0, resp.ExitCode)
	assert.Nil(t, resp.Stdout)
	assert.Empty(t, resp.Stderr)
}

func TestHookOutput_JSON_OmitsEmptyFields(t *testing.T) {
	t.Parallel()
	out := &handler.HookOutput{
		Continue: true,
	}
	data, err := json.Marshal(out)
	require.NoError(t, err)

	// Should have "continue" but not empty fields
	assert.Contains(t, string(data), `"continue":true`)
	assert.NotContains(t, string(data), `"stopReason"`)
	assert.NotContains(t, string(data), `"hookSpecificOutput"`)
}

func TestHookOutput_JSON_FullOutput(t *testing.T) {
	t.Parallel()
	out := &handler.HookOutput{
		Continue:      true,
		SystemMessage: "context loaded",
		AdditionalContext: []string{
			"Previous session summary",
		},
	}
	data, err := json.Marshal(out)
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.True(t, parsed["continue"].(bool))
	assert.Equal(t, "context loaded", parsed["systemMessage"])
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/handler/ -run TestResponse -v`
Expected: FAIL — package `handler` not found.

**Step 3: Implement handler types**

Create `internal/handler/handler.go`:

```go
// Package handler provides hook event handlers for the cc-tools CLI.
//
// Each handler processes a Claude Code hook event and returns a structured
// [Response] that maps to the hooks JSON output protocol.
package handler

import (
	"context"

	"github.com/riddopic/cc-tools/internal/hookcmd"
)

// Handler processes a hook event and returns a structured response.
type Handler interface {
	// Name returns a short identifier for logging and debugging.
	Name() string
	// Handle processes the hook event. It returns nil Response to indicate
	// no output (different from &Response{} which outputs exit code 0).
	Handle(ctx context.Context, input *hookcmd.HookInput) (*Response, error)
}

// Response captures a handler's output for the Claude Code hooks protocol.
// Exit code 0 = success, 2 = block with stderr feedback.
type Response struct {
	ExitCode int
	Stdout   *HookOutput
	Stderr   string
}

// HookOutput is the JSON written to stdout per the Claude Code hooks protocol.
type HookOutput struct {
	Continue           bool              `json:"continue,omitempty"`
	StopReason         string            `json:"stopReason,omitempty"`
	SuppressOutput     bool              `json:"suppressOutput,omitempty"`
	SystemMessage      string            `json:"systemMessage,omitempty"`
	HookSpecificOutput map[string]any    `json:"hookSpecificOutput,omitempty"`
	AdditionalContext  []string          `json:"additionalContext,omitempty"`
	PermissionDecision string            `json:"permissionDecision,omitempty"`
	UpdatedInput       map[string]any    `json:"updatedInput,omitempty"`
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/handler/ -v`
Expected: ALL PASS.

**Step 5: Write failing tests for Registry**

Create `internal/handler/registry_test.go`:

```go
package handler_test

import (
	"context"
	"testing"

	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		},
		&stubHandler{
			name: "second",
			resp: &handler.Response{
				ExitCode: 0,
				Stderr:   "log from second\n",
			},
		},
	)

	input := &hookcmd.HookInput{HookEventName: hookcmd.EventSessionStart}
	resp := r.Dispatch(context.Background(), input)

	require.NotNil(t, resp)
	// First handler's stdout wins
	require.NotNil(t, resp.Stdout)
	assert.Equal(t, "from first", resp.Stdout.SystemMessage)
	// Stderr concatenated
	assert.Contains(t, resp.Stderr, "log from second")
}

func TestRegistry_Dispatch_MaxExitCode(t *testing.T) {
	t.Parallel()
	r := handler.NewRegistry()
	r.Register(hookcmd.EventPreToolUse,
		&stubHandler{name: "pass", resp: &handler.Response{ExitCode: 0}},
		&stubHandler{name: "block", resp: &handler.Response{ExitCode: 2, Stderr: "blocked"}},
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
		err:  assert.AnError,
	})

	input := &hookcmd.HookInput{HookEventName: hookcmd.EventStop}
	resp := r.Dispatch(context.Background(), input)

	// Errors are logged to stderr, not fatal
	assert.Equal(t, 0, resp.ExitCode)
	assert.Contains(t, resp.Stderr, "[broken] error:")
}

func TestRegistry_Dispatch_NilResponse(t *testing.T) {
	t.Parallel()
	r := handler.NewRegistry()
	r.Register(hookcmd.EventNotification, &stubHandler{
		name: "silent",
		resp: nil,
	})

	input := &hookcmd.HookInput{HookEventName: hookcmd.EventNotification}
	resp := r.Dispatch(context.Background(), input)

	assert.Equal(t, 0, resp.ExitCode)
	assert.Nil(t, resp.Stdout)
}
```

**Step 6: Run test to verify it fails**

Run: `go test ./internal/handler/ -run TestRegistry -v`
Expected: FAIL — `handler.NewRegistry` undefined.

**Step 7: Implement Registry**

Create `internal/handler/registry.go`:

```go
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
		return &Response{ExitCode: 0}
	}

	merged := &Response{ExitCode: 0}
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
```

**Step 8: Run all tests**

Run: `go test ./internal/handler/ -v`
Expected: ALL PASS.

**Step 9: Run lint**

Run: `golangci-lint run ./internal/handler/`
Expected: No issues.

**Step 10: Commit**

```bash
git add internal/handler/handler.go internal/handler/registry.go internal/handler/handler_test.go internal/handler/registry_test.go
git commit -m "feat: add handler package with Handler interface and Registry"
```

---

### Task 3: Add NtfyNotifier

**Files:**
- Create: `internal/notify/ntfy.go`
- Create: `internal/notify/ntfy_test.go`
- Create: `internal/notify/multi.go`
- Create: `internal/notify/multi_test.go`

**Context:** The notify package already has `audio.go`, `desktop.go`, and `quiethours.go`. This task adds the NtfyNotifier (HTTP POST to ntfy.sh) and a MultiNotifier that composites all backends. The existing `notify.Audio` and `notify.Desktop` types have different interfaces (method names differ), so MultiNotifier wraps them with a common `Notifier` interface.

**Docs to check:**
- `internal/notify/audio.go` — current Audio type and AudioPlayer interface
- `internal/notify/desktop.go` — current Desktop type and CmdRunner interface
- `internal/notify/quiethours.go` — QuietHours struct

**Step 1: Read existing notify types to understand interfaces**

Read: `internal/notify/audio.go`, `internal/notify/desktop.go`, `internal/notify/quiethours.go`

**Step 2: Write failing test for NtfyNotifier**

Create `internal/notify/ntfy_test.go`:

```go
package notify_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/riddopic/cc-tools/internal/notify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNtfyNotifier_Send(t *testing.T) {
	t.Parallel()

	var received map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &received))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	notifier := notify.NewNtfyNotifier(notify.NtfyConfig{
		Topic:    "test-topic",
		Server:   srv.URL,
		Priority: 3,
	})

	err := notifier.Send(context.Background(), "Test Title", "Test message")
	require.NoError(t, err)

	assert.Equal(t, "test-topic", received["topic"])
	assert.Equal(t, "Test Title", received["title"])
	assert.Equal(t, "Test message", received["message"])
	assert.EqualValues(t, 3, received["priority"])
}

func TestNtfyNotifier_Send_WithToken(t *testing.T) {
	t.Parallel()

	var authHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	notifier := notify.NewNtfyNotifier(notify.NtfyConfig{
		Topic:  "test-topic",
		Server: srv.URL,
		Token:  "tk_mytoken",
	})

	err := notifier.Send(context.Background(), "Title", "Body")
	require.NoError(t, err)
	assert.Equal(t, "Bearer tk_mytoken", authHeader)
}

func TestNtfyNotifier_Send_ServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	notifier := notify.NewNtfyNotifier(notify.NtfyConfig{
		Topic:  "test-topic",
		Server: srv.URL,
	})

	err := notifier.Send(context.Background(), "Title", "Body")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}
```

**Step 3: Run test to verify it fails**

Run: `go test ./internal/notify/ -run TestNtfyNotifier -v`
Expected: FAIL — `notify.NewNtfyNotifier` undefined.

**Step 4: Implement NtfyNotifier**

Create `internal/notify/ntfy.go`:

```go
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// NtfyConfig configures the ntfy notification backend.
type NtfyConfig struct {
	Topic    string // required
	Server   string // default "https://ntfy.sh"
	Token    string // optional bearer token
	Priority int    // 1-5, default 3
}

// NtfyNotifier sends notifications via ntfy.sh HTTP API.
type NtfyNotifier struct {
	config NtfyConfig
	client *http.Client
}

// NewNtfyNotifier creates a new ntfy notifier.
func NewNtfyNotifier(cfg NtfyConfig) *NtfyNotifier {
	if cfg.Server == "" {
		cfg.Server = "https://ntfy.sh"
	}
	if cfg.Priority == 0 {
		cfg.Priority = 3
	}
	return &NtfyNotifier{
		config: cfg,
		client: &http.Client{},
	}
}

// Send posts a notification to the configured ntfy topic.
func (n *NtfyNotifier) Send(ctx context.Context, title, message string) error {
	body := map[string]any{
		"topic":    n.config.Topic,
		"title":    title,
		"message":  message,
		"priority": n.config.Priority,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal ntfy payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.config.Server, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create ntfy request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if n.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+n.config.Token)
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("send ntfy notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ntfy returned status %d", resp.StatusCode)
	}
	return nil
}
```

**Step 5: Run tests to verify they pass**

Run: `go test ./internal/notify/ -run TestNtfyNotifier -v`
Expected: ALL PASS.

**Step 6: Write failing test for MultiNotifier**

Create `internal/notify/multi_test.go`:

```go
package notify_test

import (
	"context"
	"errors"
	"testing"

	"github.com/riddopic/cc-tools/internal/notify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSender struct {
	called bool
	err    error
}

func (m *mockSender) Send(_ context.Context, _, _ string) error {
	m.called = true
	return m.err
}

func TestMultiNotifier_Send_AllBackends(t *testing.T) {
	t.Parallel()

	s1 := &mockSender{}
	s2 := &mockSender{}

	multi := notify.NewMultiNotifier([]notify.Sender{s1, s2}, nil)
	err := multi.Send(context.Background(), "Title", "Body")
	require.NoError(t, err)

	assert.True(t, s1.called)
	assert.True(t, s2.called)
}

func TestMultiNotifier_Send_CollectsErrors(t *testing.T) {
	t.Parallel()

	s1 := &mockSender{err: errors.New("fail1")}
	s2 := &mockSender{err: errors.New("fail2")}

	multi := notify.NewMultiNotifier([]notify.Sender{s1, s2}, nil)
	err := multi.Send(context.Background(), "Title", "Body")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fail1")
	assert.Contains(t, err.Error(), "fail2")
}

func TestMultiNotifier_Send_QuietHours(t *testing.T) {
	t.Parallel()

	s1 := &mockSender{}
	qh := &notify.QuietHours{Enabled: true, Start: "00:00", End: "23:59"}

	multi := notify.NewMultiNotifier([]notify.Sender{s1}, qh)
	err := multi.Send(context.Background(), "Title", "Body")

	require.NoError(t, err)
	// During quiet hours, senders are NOT called
	assert.False(t, s1.called)
}

func TestMultiNotifier_Send_NoSenders(t *testing.T) {
	t.Parallel()

	multi := notify.NewMultiNotifier(nil, nil)
	err := multi.Send(context.Background(), "Title", "Body")
	assert.NoError(t, err)
}
```

**Step 7: Run test to verify it fails**

Run: `go test ./internal/notify/ -run TestMultiNotifier -v`
Expected: FAIL — `notify.Sender` and `notify.NewMultiNotifier` undefined.

**Step 8: Implement Sender interface and MultiNotifier**

Create `internal/notify/multi.go`:

```go
package notify

import (
	"context"
	"errors"
	"time"
)

// Sender sends a notification with a title and message body.
type Sender interface {
	Send(ctx context.Context, title, message string) error
}

// MultiNotifier composites multiple notification backends, respecting quiet hours.
type MultiNotifier struct {
	senders    []Sender
	quietHours *QuietHours
}

// NewMultiNotifier creates a notifier that fans out to all senders.
func NewMultiNotifier(senders []Sender, qh *QuietHours) *MultiNotifier {
	return &MultiNotifier{
		senders:    senders,
		quietHours: qh,
	}
}

// Send dispatches the notification to all backends. Errors are collected
// and returned as a joined error. Quiet hours suppress all notifications.
func (m *MultiNotifier) Send(ctx context.Context, title, message string) error {
	if m.quietHours != nil && m.quietHours.IsActive(time.Now()) {
		return nil
	}

	var errs []error
	for _, s := range m.senders {
		if err := s.Send(ctx, title, message); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
```

**Step 9: Run all notify tests**

Run: `go test ./internal/notify/ -v`
Expected: ALL PASS (existing + new tests).

**Step 10: Run lint**

Run: `golangci-lint run ./internal/notify/`
Expected: No issues.

**Step 11: Commit**

```bash
git add internal/notify/ntfy.go internal/notify/ntfy_test.go internal/notify/multi.go internal/notify/multi_test.go
git commit -m "feat: add NtfyNotifier and MultiNotifier to notify package"
```

---

### Task 4: Implement Handler Functions for SessionStart Events

**Files:**
- Create: `internal/handler/session_start.go`
- Create: `internal/handler/session_start_test.go`

**Context:** Currently `cmd/cc-tools/main.go:198-316` defines three SessionStart handler closures: `superpowersHandler()`, `pkgManagerHandler()`, and `sessionContextHandler()`. These need to be converted from `hookcmd.Handler` (writes to io.Writer) to `handler.Handler` (returns `*Response`). Each handler becomes a struct with a `Handle` method that returns structured output.

**Key reference files:**
- `internal/superpowers/injector.go` — `NewInjector(cwd).Run(ctx, out)` returns error, writes to `out`
- `internal/pkgmanager/detect.go` — `Detect(cwd)` returns string
- `internal/pkgmanager/envfile.go` — `WriteToEnvFile(path, manager)` returns error
- `internal/session/store.go` — `NewStore(dir)`, `store.List(limit)` returns `[]*Session`
- `internal/session/aliases.go` — `NewAliasManager(file)`, `aliases.List()` returns `map[string]string`

**Step 1: Write failing tests**

Create `internal/handler/session_start_test.go`:

```go
package handler_test

import (
	"context"
	"testing"

	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuperpowersHandler_Name(t *testing.T) {
	t.Parallel()
	h := handler.NewSuperpowersHandler()
	assert.Equal(t, "superpowers", h.Name())
}

func TestSuperpowersHandler_Handle(t *testing.T) {
	t.Parallel()
	h := handler.NewSuperpowersHandler()
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           t.TempDir(),
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	// Superpowers writes to stdout for Claude to ingest
	require.NotNil(t, resp)
}

func TestPkgManagerHandler_Name(t *testing.T) {
	t.Parallel()
	h := handler.NewPkgManagerHandler()
	assert.Equal(t, "pkg-manager", h.Name())
}

func TestPkgManagerHandler_Handle(t *testing.T) {
	t.Parallel()
	h := handler.NewPkgManagerHandler()
	tmpDir := t.TempDir()
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           tmpDir,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
}

func TestSessionContextHandler_Name(t *testing.T) {
	t.Parallel()
	h := handler.NewSessionContextHandler()
	assert.Equal(t, "session-context", h.Name())
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/handler/ -run TestSuperpowersHandler -v`
Expected: FAIL — `handler.NewSuperpowersHandler` undefined.

**Step 3: Implement SessionStart handlers**

Create `internal/handler/session_start.go`:

```go
package handler

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/riddopic/cc-tools/internal/pkgmanager"
	"github.com/riddopic/cc-tools/internal/session"
	"github.com/riddopic/cc-tools/internal/superpowers"
)

// SuperpowersHandler injects superpowers system message on session start.
type SuperpowersHandler struct{}

// NewSuperpowersHandler creates a new superpowers handler.
func NewSuperpowersHandler() *SuperpowersHandler {
	return &SuperpowersHandler{}
}

func (h *SuperpowersHandler) Name() string { return "superpowers" }

func (h *SuperpowersHandler) Handle(ctx context.Context, input *hookcmd.HookInput) (*Response, error) {
	var buf bytes.Buffer
	if err := superpowers.NewInjector(input.Cwd).Run(ctx, &buf); err != nil {
		return nil, fmt.Errorf("inject superpowers: %w", err)
	}

	msg := buf.String()
	if msg == "" {
		return &Response{ExitCode: 0}, nil
	}

	return &Response{
		ExitCode: 0,
		Stdout: &HookOutput{
			Continue:      true,
			SystemMessage: msg,
		},
	}, nil
}

// PkgManagerHandler detects the package manager and writes to .claude/.env.
type PkgManagerHandler struct{}

// NewPkgManagerHandler creates a new package manager handler.
func NewPkgManagerHandler() *PkgManagerHandler {
	return &PkgManagerHandler{}
}

func (h *PkgManagerHandler) Name() string { return "pkg-manager" }

func (h *PkgManagerHandler) Handle(_ context.Context, input *hookcmd.HookInput) (*Response, error) {
	manager := pkgmanager.Detect(input.Cwd)
	envDir := filepath.Join(input.Cwd, ".claude")
	if err := os.MkdirAll(envDir, 0o750); err != nil {
		return nil, fmt.Errorf("create .claude directory: %w", err)
	}
	envFile := filepath.Join(envDir, ".env")
	if err := pkgmanager.WriteToEnvFile(envFile, manager); err != nil {
		return nil, fmt.Errorf("write env file: %w", err)
	}
	return &Response{ExitCode: 0}, nil
}

// SessionContextHandler provides previous session context on start.
type SessionContextHandler struct{}

// NewSessionContextHandler creates a new session context handler.
func NewSessionContextHandler() *SessionContextHandler {
	return &SessionContextHandler{}
}

func (h *SessionContextHandler) Name() string { return "session-context" }

func (h *SessionContextHandler) Handle(_ context.Context, _ *hookcmd.HookInput) (*Response, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home directory: %w", err)
	}

	storeDir := filepath.Join(homeDir, ".claude", "sessions")
	store := session.NewStore(storeDir)

	sessions, _ := store.List(1)
	if len(sessions) == 0 {
		return &Response{ExitCode: 0}, nil
	}

	var context []string
	latest := sessions[0]
	if latest.Summary != "" {
		context = append(context, fmt.Sprintf("Previous session (%s): %s", latest.Date, latest.Summary))
	}

	var stderr string
	aliasFile := filepath.Join(homeDir, ".claude", "session-aliases.json")
	aliases := session.NewAliasManager(aliasFile)
	aliasList, aliasErr := aliases.List()
	if aliasErr == nil && len(aliasList) > 0 {
		names := make([]string, 0, len(aliasList))
		for name := range aliasList {
			names = append(names, name)
		}
		stderr = fmt.Sprintf("[session-context] %d alias(es): %s\n",
			len(aliasList), strings.Join(names, ", "))
	}

	resp := &Response{ExitCode: 0, Stderr: stderr}
	if len(context) > 0 {
		resp.Stdout = &HookOutput{
			Continue:          true,
			AdditionalContext: context,
		}
	}
	return resp, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/handler/ -v`
Expected: ALL PASS.

**Step 5: Run lint**

Run: `golangci-lint run ./internal/handler/`
Expected: No issues.

**Step 6: Commit**

```bash
git add internal/handler/session_start.go internal/handler/session_start_test.go
git commit -m "feat: add SessionStart handler implementations"
```

---

### Task 5: Implement Handlers for PreToolUse, PostToolUse, and Remaining Events

**Files:**
- Create: `internal/handler/tooluse.go`
- Create: `internal/handler/tooluse_test.go`
- Create: `internal/handler/notification.go`
- Create: `internal/handler/notification_test.go`
- Create: `internal/handler/session_end.go`
- Create: `internal/handler/session_end_test.go`
- Create: `internal/handler/compact.go`
- Create: `internal/handler/compact_test.go`

**Context:** Migrate the remaining handler closures from `cmd/cc-tools/main.go`:
- `suggestCompactHandler` (lines 222-239) → PreToolUse
- `observeHandler` (lines 241-263) → PreToolUse, PostToolUse, PostToolUseFailure
- `preCommitReminderHandler` (lines 380-405) → PreToolUse
- `notifyAudioHandler` (lines 407-425) + `notifyDesktopHandler` (lines 434-466) → Notification
- `sessionEndHandler` (lines 318-378) → SessionEnd
- `logCompactionHandler` (lines 265-277) → PreCompact

Each handler receives `*config.Values` via constructor (dependency injection).

**Key reference files:**
- `internal/compact/suggestor.go` — `NewSuggestor(stateDir, threshold, interval)`, `.RecordCall(sessionID, errOut)`
- `internal/observe/observer.go` — `NewObserver(dir, maxSize)`, `.Record(event)`
- `internal/config/values.go` — Config struct hierarchy
- `internal/notify/audio.go` — `NewAudio(player, dir, qh, logger)`, `.PlayRandom()`
- `internal/notify/desktop.go` — `NewDesktop(runner)`, `.Send(title, msg)`

**Step 1: Write failing tests for tooluse handlers**

Create `internal/handler/tooluse_test.go` with tests for SuggestCompactHandler, ObserveHandler, PreCommitReminderHandler. Key test: `PreCommitReminderHandler` should write stderr when tool_input contains "git commit".

```go
package handler_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreCommitReminderHandler_Name(t *testing.T) {
	t.Parallel()
	h := handler.NewPreCommitReminderHandler(nil)
	assert.Equal(t, "pre-commit-reminder", h.Name())
}

func TestPreCommitReminderHandler_NoConfig(t *testing.T) {
	t.Parallel()
	h := handler.NewPreCommitReminderHandler(nil)
	input := &hookcmd.HookInput{ToolName: "Bash"}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Stderr)
}

func TestPreCommitReminderHandler_GitCommit(t *testing.T) {
	t.Parallel()
	cfg := &config.Values{}
	cfg.PreCommit.Enabled = true
	cfg.PreCommit.Command = "task pre-commit"

	toolInput, _ := json.Marshal(map[string]string{"command": "git commit -m 'test'"})
	h := handler.NewPreCommitReminderHandler(cfg)
	input := &hookcmd.HookInput{
		ToolName:  "Bash",
		ToolInput: toolInput,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Stderr, "task pre-commit")
}

func TestObserveHandler_Name(t *testing.T) {
	t.Parallel()
	h := handler.NewObserveHandler(nil, "pre")
	assert.Equal(t, "observe-pre", h.Name())
}

func TestSuggestCompactHandler_Name(t *testing.T) {
	t.Parallel()
	h := handler.NewSuggestCompactHandler(nil)
	assert.Equal(t, "suggest-compact", h.Name())
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/handler/ -run TestPreCommitReminderHandler -v`
Expected: FAIL.

**Step 3: Implement all remaining handlers**

Create the four handler files following the same pattern as session_start.go — each handler is a struct with config, implementing `Handle(ctx, input) (*Response, error)`. The logic is a direct port from the closures in main.go, but returning `*Response` with stderr strings instead of writing to `io.Writer`.

For each handler file, follow the pattern:
- Constructor accepts `*config.Values` (nil-safe — check `cfg == nil` early)
- Return `&Response{ExitCode: 0}` for no-op cases
- Return stderr via `Response.Stderr` instead of `fmt.Fprintf(errOut, ...)`

**Step 4: Run all handler tests**

Run: `go test ./internal/handler/ -v`
Expected: ALL PASS.

**Step 5: Run lint**

Run: `golangci-lint run ./internal/handler/`
Expected: No issues.

**Step 6: Commit**

```bash
git add internal/handler/tooluse.go internal/handler/tooluse_test.go \
  internal/handler/notification.go internal/handler/notification_test.go \
  internal/handler/session_end.go internal/handler/session_end_test.go \
  internal/handler/compact.go internal/handler/compact_test.go
git commit -m "feat: add PreToolUse, Notification, SessionEnd, and PreCompact handlers"
```

---

### Task 6: Create Default Registry Wiring

**Files:**
- Create: `internal/handler/defaults.go`
- Create: `internal/handler/defaults_test.go`

**Context:** The design calls for `handler.NewDefaultRegistry(cfg)` that pre-registers all handlers. This is the composition root — it's the one place that knows which handlers serve which events. Currently this logic lives in `buildHookRegistry()` in `cmd/cc-tools/main.go:175-187`.

**Step 1: Write failing test**

Create `internal/handler/defaults_test.go`:

```go
package handler_test

import (
	"testing"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/stretchr/testify/assert"
)

func TestNewDefaultRegistry_RegistersAllEvents(t *testing.T) {
	t.Parallel()

	cfg := config.GetDefaultConfig()
	r := handler.NewDefaultRegistry(cfg)

	// Verify handlers registered for expected events
	assert.True(t, r.HasHandlers(hookcmd.EventSessionStart))
	assert.True(t, r.HasHandlers(hookcmd.EventSessionEnd))
	assert.True(t, r.HasHandlers(hookcmd.EventPreToolUse))
	assert.True(t, r.HasHandlers(hookcmd.EventPostToolUse))
	assert.True(t, r.HasHandlers(hookcmd.EventPreCompact))
	assert.True(t, r.HasHandlers(hookcmd.EventNotification))
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/handler/ -run TestNewDefaultRegistry -v`
Expected: FAIL.

**Step 3: Implement defaults.go**

Create `internal/handler/defaults.go`:

```go
package handler

import (
	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/hookcmd"
)

// NewDefaultRegistry creates a registry with all default handlers wired.
func NewDefaultRegistry(cfg *config.Values) *Registry {
	r := NewRegistry()

	r.Register(hookcmd.EventSessionStart,
		NewSuperpowersHandler(),
		NewPkgManagerHandler(),
		NewSessionContextHandler(),
	)

	r.Register(hookcmd.EventSessionEnd,
		NewSessionEndHandler(cfg),
	)

	r.Register(hookcmd.EventPreToolUse,
		NewSuggestCompactHandler(cfg),
		NewObserveHandler(cfg, "pre"),
		NewPreCommitReminderHandler(cfg),
	)

	r.Register(hookcmd.EventPostToolUse,
		NewObserveHandler(cfg, "post"),
	)

	r.Register(hookcmd.EventPostToolUseFailure,
		NewObserveHandler(cfg, "failure"),
	)

	r.Register(hookcmd.EventPreCompact,
		NewLogCompactionHandler(),
	)

	r.Register(hookcmd.EventNotification,
		NewNotificationHandler(cfg),
	)

	return r
}

// HasHandlers reports whether the registry has handlers for the given event.
func (r *Registry) HasHandlers(event string) bool {
	return len(r.handlers[event]) > 0
}
```

**Note:** This also requires exporting `config.GetDefaultConfig()` — currently it's unexported (`getDefaultConfig()`). Add a one-line wrapper: `func GetDefaultConfig() *Values { return getDefaultConfig() }` to `internal/config/keys.go`.

**Step 4: Run tests**

Run: `go test ./internal/handler/ -v`
Expected: ALL PASS.

**Step 5: Commit**

```bash
git add internal/handler/defaults.go internal/handler/defaults_test.go internal/config/keys.go
git commit -m "feat: add default handler registry wiring"
```

---

### Task 7: Create Cobra Root Command and Hook Command

**Files:**
- Rewrite: `cmd/cc-tools/main.go`
- Create: `cmd/cc-tools/hook.go`

**Context:** This is the core migration. The current `main()` function (lines 47-93) does manual `os.Args` switching. Replace with Cobra root command. The `hook` command reads stdin, dispatches through the handler registry, and writes structured JSON output — this replaces `runHookCommand()` (lines 136-161).

**Critical:** The hook command reads JSON from stdin and writes JSON to stdout. Exit code 0 = success, exit code 2 = blocking with stderr. This must match the Claude Code hooks protocol exactly.

**Step 1: Create hook.go**

Create `cmd/cc-tools/hook.go`:

```go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
)

func newHookCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "hook",
		Short:  "Handle Claude Code hook events",
		Long:   "Reads hook event JSON from stdin, dispatches to registered handlers, and writes structured output.",
		Hidden: true,
		RunE:   runHook,
	}
}

func runHook(cmd *cobra.Command, _ []string) error {
	input, err := hookcmd.ParseInput(os.Stdin)
	if err != nil {
		// Hook errors must not block Claude — exit 0
		return nil
	}

	cfg := loadConfig()
	registry := handler.NewDefaultRegistry(cfg)
	resp := registry.Dispatch(cmd.Context(), input)

	return writeHookResponse(os.Stdout, os.Stderr, resp)
}

func loadConfig() *config.Values {
	mgr := config.NewManager()
	cfg, err := mgr.GetConfig(nil)
	if err != nil {
		return nil
	}
	return cfg
}

func writeHookResponse(stdout, stderr io.Writer, resp *handler.Response) error {
	if resp.Stderr != "" {
		_, _ = fmt.Fprint(stderr, resp.Stderr)
	}

	if resp.Stdout != nil {
		data, err := json.Marshal(resp.Stdout)
		if err != nil {
			return fmt.Errorf("marshal hook output: %w", err)
		}
		_, _ = fmt.Fprintln(stdout, string(data))
	}

	if resp.ExitCode != 0 {
		os.Exit(resp.ExitCode)
	}
	return nil
}
```

**Step 2: Rewrite main.go with Cobra root**

Rewrite `cmd/cc-tools/main.go` to use Cobra as the root command. Keep the existing `writeDebugLog` function. Move `version` into Cobra's `Version` field. Add all subcommands.

During this step, the existing command functions (`runConfigCommand`, `runSkipCommand`, etc.) are temporarily wrapped as Cobra commands that call the old functions.

**Step 3: Build and verify**

Run:
```bash
go build -o ~/.claude/bin/cc-tools ./cmd/cc-tools/
echo '{"hook_event_name":"SessionStart","cwd":"/tmp"}' | ~/.claude/bin/cc-tools hook
```

Expected: Exits 0 with JSON output to stdout.

**Step 4: Run all tests**

Run: `go test ./...`
Expected: ALL PASS.

**Step 5: Commit**

```bash
git add cmd/cc-tools/main.go cmd/cc-tools/hook.go
git commit -m "feat: replace manual arg parsing with Cobra root and hook commands"
```

---

### Task 8: Migrate Session, Config, Skip, Debug, MCP Commands to Cobra

**Files:**
- Rewrite: `cmd/cc-tools/config.go`
- Rewrite: `cmd/cc-tools/skip.go`
- Rewrite: `cmd/cc-tools/debug.go`
- Rewrite: `cmd/cc-tools/mcp.go`
- Create: `cmd/cc-tools/session.go`
- Modify: `cmd/cc-tools/main.go` (remove session functions)

**Context:** Each existing command file already delegates to internal packages. The migration wraps each in `cobra.Command` with proper `Use`, `Short`, `Example`, `RunE` fields. Flag parsing replaces manual `os.Args` indexing.

Session commands currently live in `main.go` (lines 475-675) and need to be extracted to `session.go`.

For each command:
1. Create `newXxxCmd()` returning `*cobra.Command`
2. Add subcommands as children
3. Use `cmd.Flags()` for flag definitions
4. Keep the same internal package calls

**Step 1: Extract session commands to session.go**

Move `runSessionCommand`, `runSessionList`, `runSessionInfo`, `runSessionAlias`, `runSessionAliasList`, `runSessionAliasSet`, `runSessionAliasRemove`, `runSessionSearch` from `main.go` to `session.go`, wrapped in Cobra commands.

**Step 2: Rewrite config.go, skip.go, debug.go, mcp.go**

Each file follows the same pattern — wrap existing logic in Cobra commands. Example for config:

```go
func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration settings",
	}
	cmd.AddCommand(
		newConfigGetCmd(),
		newConfigSetCmd(),
		newConfigListCmd(),
		newConfigResetCmd(),
	)
	return cmd
}
```

**Step 3: Update main.go root command**

Add all commands to root:
```go
root.AddCommand(
	newHookCmd(),
	newValidateCmd(),
	newConfigCmd(),
	newSessionCmd(),
	newSkipCmd(),
	newDebugCmd(),
	newMCPCmd(),
)
```

**Step 4: Remove old command functions from main.go**

Delete `runConfigCommand`, `runSkipCommand`, `runUnskipCommand`, `runDebugCommand`, `runMCPCommand`, `runSessionCommand`, all session helper functions, `printUsage`, manual arg parsing constants.

**Step 5: Build and verify all commands**

Run:
```bash
go build -o ~/.claude/bin/cc-tools ./cmd/cc-tools/
~/.claude/bin/cc-tools --help
~/.claude/bin/cc-tools config list
~/.claude/bin/cc-tools session list
~/.claude/bin/cc-tools skip status
~/.claude/bin/cc-tools debug status
~/.claude/bin/cc-tools mcp list
```

Expected: All commands work identically to before.

**Step 6: Run all tests**

Run: `go test ./...`
Expected: ALL PASS.

**Step 7: Commit**

```bash
git add cmd/cc-tools/
git commit -m "refactor: migrate all commands to Cobra"
```

---

### Task 9: Migrate Validate Command and Wire PostToolUse Handler

**Files:**
- Rewrite: `cmd/cc-tools/validate.go` (currently part of main.go)
- Modify: `cmd/cc-tools/main.go` (remove `runValidate`, `loadValidateConfig`)

**Context:** The validate command has a dual-path design (from the design doc): it works as a standalone `cc-tools validate` CLI command AND as a PostToolUse handler calling the same `hooks.ValidateWithSkipCheck()`. Currently `runValidate()` is in `main.go:707-721` and `loadValidateConfig()` is in `main.go:677-705`.

**Step 1: Create validate.go**

Create `cmd/cc-tools/validate.go`:

```go
package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/riddopic/cc-tools/internal/hooks"
)

func newValidateCmd() *cobra.Command {
	var timeout int
	var cooldown int

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Run lint and test validation",
		Long:  "Discovers and runs lint and test commands in parallel, reporting results.",
		Example: `  echo '{"tool_input":{"file_path":"main.go"}}' | cc-tools validate
  CC_TOOLS_HOOKS_VALIDATE_TIMEOUT_SECONDS=120 cc-tools validate`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			debug := os.Getenv("CLAUDE_HOOKS_DEBUG") == "1"

			// Read stdin
			var stdinData []byte
			if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
				stdinData, _ = io.ReadAll(os.Stdin)
			}

			exitCode := hooks.ValidateWithSkipCheck(
				cmd.Context(),
				stdinData,
				os.Stdout,
				os.Stderr,
				debug,
				timeout,
				cooldown,
			)

			if exitCode != 0 {
				os.Exit(exitCode)
			}
			return nil
		},
	}

	cmd.Flags().IntVarP(&timeout, "timeout", "t", 60, "timeout in seconds")
	cmd.Flags().IntVarP(&cooldown, "cooldown", "c", 5, "cooldown between runs in seconds")

	return cmd
}
```

**Step 2: Remove old validate code from main.go**

Delete `runValidate()` and `loadValidateConfig()` from main.go.

**Step 3: Build and verify**

Run:
```bash
go build -o ~/.claude/bin/cc-tools ./cmd/cc-tools/
echo '{"tool_input":{"file_path":"internal/handler/handler.go"}}' | ~/.claude/bin/cc-tools validate
```

Expected: Runs validation as before.

**Step 4: Commit**

```bash
git add cmd/cc-tools/validate.go cmd/cc-tools/main.go
git commit -m "refactor: migrate validate command to Cobra"
```

---

### Task 10: Clean Up Old Code and Remove Legacy Patterns

**Files:**
- Modify: `cmd/cc-tools/main.go` — remove old handler closures, adapter types
- Modify: `internal/hookcmd/hookcmd.go` — keep for backward compatibility (legacy `cc-tools-validate` binary may still use it)

**Context:** After Tasks 7-9, all commands use Cobra. The old `handlerFunc` adapter struct, `buildHookRegistry()`, `hookEventMap`, all handler closure functions (`superpowersHandler()`, `pkgManagerHandler()`, etc.), and the `afPlayer`/`osascriptRunner` adapter types in `main.go` are dead code.

**Step 1: Remove dead code from main.go**

Delete from `cmd/cc-tools/main.go`:
- `handlerFunc` struct and methods (lines 164-173)
- `buildHookRegistry()` function (lines 175-187)
- `loadHookConfig()` function (lines 189-196)
- `hookEventMap` variable (lines 123-132)
- All handler closure functions: `superpowersHandler()`, `pkgManagerHandler()`, `suggestCompactHandler()`, `observeHandler()`, `logCompactionHandler()`, `sessionContextHandler()`, `sessionEndHandler()`, `preCommitReminderHandler()`, `notifyAudioHandler()`, `notifyDesktopHandler()`
- `afPlayer` and `osascriptRunner` adapter types
- `needsStdin()` function
- Old `printUsage()` function
- All `minArgs`, `minSessionArgs`, etc. constants that are no longer used
- Old imports no longer needed

The resulting `main.go` should be ~30 lines: just the root command, `main()`, and `writeDebugLog()`.

**Step 2: Verify build compiles**

Run: `go build ./cmd/cc-tools/`
Expected: Compiles cleanly.

**Step 3: Run all tests**

Run: `go test ./...`
Expected: ALL PASS.

**Step 4: Run lint**

Run: `golangci-lint run ./...`
Expected: No issues.

**Step 5: Commit**

```bash
git add cmd/cc-tools/main.go
git commit -m "refactor: remove legacy handler closures and manual arg parsing"
```

---

### Task 11: Update Hook Wiring in .claude/settings.json

**Files:**
- Modify: `.claude/settings.json`

**Context:** Currently hooks call `cc-tools hook session-start`, `cc-tools hook pre-tool-use`, etc. with kebab-case event subcommands. With the Cobra redesign, the hook command reads `hook_event_name` directly from stdin JSON — it no longer needs a subcommand argument. All hooks can use the same command: `~/.claude/bin/cc-tools hook`.

However, `cc-tools-validate` (the standalone validation binary) is still used in PostToolUse. Keep that wiring unchanged since it's a separate binary.

**Step 1: Update settings.json hooks section**

Change all `cc-tools hook <event>` entries to just `cc-tools hook`:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "~/.claude/bin/cc-tools hook"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "~/.claude/bin/cc-tools hook"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "~/.claude/bin/cc-tools-validate"
          }
        ]
      },
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "~/.claude/bin/cc-tools hook"
          }
        ]
      }
    ],
    "PreCompact": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "~/.claude/bin/cc-tools hook"
          }
        ]
      }
    ],
    "Notification": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "~/.claude/bin/cc-tools hook"
          }
        ]
      }
    ],
    "SessionEnd": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "~/.claude/bin/cc-tools hook"
          }
        ]
      }
    ]
  }
}
```

**Step 2: Rebuild and install**

Run:
```bash
go build -o ~/.claude/bin/cc-tools ./cmd/cc-tools/
```

**Step 3: Test hook invocation**

Run:
```bash
echo '{"hook_event_name":"SessionStart","cwd":"/tmp","session_id":"test"}' | ~/.claude/bin/cc-tools hook
echo $?
```

Expected: Exit code 0, outputs JSON to stdout.

**Step 4: Commit**

```bash
git add .claude/settings.json
git commit -m "chore: simplify hook wiring to use single cc-tools hook command"
```

---

### Task 12: Final Verification and Quality Check

**Files:** None modified.

**Context:** Verify the entire migration is complete, all tests pass, lint is clean, and the binary works correctly.

**Step 1: Run full test suite**

Run: `go test ./... -count=1`
Expected: ALL PASS across all packages.

**Step 2: Run lint**

Run: `golangci-lint run ./...`
Expected: No issues.

**Step 3: Run race detector**

Run: `go test -race ./...`
Expected: No races detected.

**Step 4: Verify binary**

Run:
```bash
go build -o ~/.claude/bin/cc-tools ./cmd/cc-tools/
~/.claude/bin/cc-tools --help
~/.claude/bin/cc-tools version
~/.claude/bin/cc-tools config list
~/.claude/bin/cc-tools session list
~/.claude/bin/cc-tools skip status
~/.claude/bin/cc-tools debug status
~/.claude/bin/cc-tools mcp list
~/.claude/bin/cc-tools hook --help
~/.claude/bin/cc-tools validate --help
```

Expected: All commands work, help text is generated by Cobra.

**Step 5: Verify hook integration**

Run:
```bash
echo '{"hook_event_name":"SessionStart","cwd":"'$(pwd)'","session_id":"verify-test"}' | ~/.claude/bin/cc-tools hook
echo '{"hook_event_name":"Notification","message":"test","title":"test"}' | ~/.claude/bin/cc-tools hook
echo '{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"echo hello"}}' | ~/.claude/bin/cc-tools hook
```

Expected: Each exits 0, hooks fire correctly.

**Step 6: Verify line count reduction**

Run:
```bash
wc -l cmd/cc-tools/*.go
```

Expected: `main.go` should be under 50 lines (down from 757). Total cmd/ code should be comparable but better structured.

**Step 7: Check for unused imports/code**

Run: `go vet ./...`
Expected: No issues.

---

## Verification Checklist

After all tasks complete:

- [ ] `go test ./... -count=1` — all pass
- [ ] `golangci-lint run ./...` — no issues
- [ ] `go test -race ./...` — no races
- [ ] `go vet ./...` — no issues
- [ ] `cc-tools --help` — Cobra-generated help
- [ ] `cc-tools hook` — reads stdin JSON, dispatches correctly
- [ ] `cc-tools validate` — runs lint+test in parallel
- [ ] `cc-tools config list` — shows all config keys
- [ ] `cc-tools session list` — shows sessions
- [ ] `cc-tools skip status` — shows skip status
- [ ] `cc-tools debug status` — shows debug status
- [ ] `cc-tools mcp list` — shows MCP servers
- [ ] `main.go` < 50 lines
- [ ] No handler closures in cmd/
- [ ] All business logic in internal/

## Future Work (Not in This Plan)

- **JSON → YAML config migration** — Change config file format from `config.json` to `config.yaml`
- **Shell completions** — `cc-tools completion bash/zsh/fish` (free from Cobra)
- **ntfy wiring** — Add `NtfyConfig` fields to config and wire NtfyNotifier into notification handler
- **Structured logging** — Replace debug file logging with structured JSON logs
