package handler_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
)

// newTestConfig returns a config.Values with all fields populated to satisfy
// exhaustruct. Callers should override fields as needed after construction.
func newTestConfig() *config.Values {
	return &config.Values{
		Validate: config.ValidateValues{
			Timeout:  0,
			Cooldown: 0,
		},
		Notifications: config.NotificationsValues{
			NtfyTopic: "",
		},
		Compact: config.CompactValues{
			Threshold:        0,
			ReminderInterval: 0,
		},
		Notify: config.NotifyValues{
			QuietHours: config.QuietHoursValues{
				Enabled: false,
				Start:   "",
				End:     "",
			},
			Audio: config.AudioValues{
				Enabled:   false,
				Directory: "",
			},
			Desktop: config.DesktopValues{
				Enabled: false,
			},
		},
		Observe: config.ObserveValues{
			Enabled:       false,
			MaxFileSizeMB: 0,
		},
		Learning: config.LearningValues{
			MinSessionLength:  0,
			LearnedSkillsPath: "",
		},
		PreCommit: config.PreCommitValues{
			Enabled: false,
			Command: "",
		},
	}
}

// ---------------------------------------------------------------------
// SuggestCompactHandler
// ---------------------------------------------------------------------

func TestSuggestCompactHandler_Name(t *testing.T) {
	t.Parallel()
	h := handler.NewSuggestCompactHandler(nil)
	assert.Equal(t, "suggest-compact", h.Name())
}

func TestSuggestCompactHandler_NilConfig(t *testing.T) {
	t.Parallel()
	h := handler.NewSuggestCompactHandler(nil)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		SessionID:     "test-session",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
}

func TestSuggestCompactHandler_RecordsCall(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, "compact")

	cfg := newTestConfig()
	cfg.Compact.Threshold = 5
	cfg.Compact.ReminderInterval = 10

	h := handler.NewSuggestCompactHandler(cfg, handler.WithCompactStateDir(stateDir))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		SessionID:     "test-session-record",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	// Counter file should now exist.
	counterFile := filepath.Join(stateDir, "cc-tools-compact-test-session-record.count")
	_, statErr := os.Stat(counterFile)
	assert.NoError(t, statErr, "counter file should be created")
}

func TestSuggestCompactHandler_SuggestsAtThreshold(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, "compact")

	cfg := newTestConfig()
	cfg.Compact.Threshold = 3
	cfg.Compact.ReminderInterval = 5

	h := handler.NewSuggestCompactHandler(cfg, handler.WithCompactStateDir(stateDir))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		SessionID:     "threshold-session",
	}

	// Make 3 calls (threshold).
	var lastResp *handler.Response
	for range 3 {
		resp, err := h.Handle(context.Background(), input)
		require.NoError(t, err)
		lastResp = resp
	}

	require.NotNil(t, lastResp)
	assert.Contains(t, lastResp.Stderr, "/compact",
		"should suggest /compact at threshold")
}

func TestSuggestCompactHandler_BelowThreshold(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, "compact")

	cfg := newTestConfig()
	cfg.Compact.Threshold = 5
	cfg.Compact.ReminderInterval = 10

	h := handler.NewSuggestCompactHandler(cfg, handler.WithCompactStateDir(stateDir))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		SessionID:     "below-threshold",
	}

	// Make 4 calls (threshold-1) — none should suggest.
	for range 4 {
		resp, err := h.Handle(context.Background(), input)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Stderr, "no suggestion below threshold")
	}
}

func TestSuggestCompactHandler_ReminderInterval(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, "compact")

	cfg := newTestConfig()
	cfg.Compact.Threshold = 3
	cfg.Compact.ReminderInterval = 2

	h := handler.NewSuggestCompactHandler(cfg, handler.WithCompactStateDir(stateDir))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		SessionID:     "reminder-interval",
	}

	var responses [5]*handler.Response
	for i := range 5 {
		resp, err := h.Handle(context.Background(), input)
		require.NoError(t, err)
		responses[i] = resp
	}

	// Calls 1 and 2: below threshold, no suggestion.
	assert.Empty(t, responses[0].Stderr, "call 1: no suggestion")
	assert.Empty(t, responses[1].Stderr, "call 2: no suggestion")

	// Call 3: hits threshold, should suggest.
	assert.NotEmpty(t, responses[2].Stderr, "call 3: suggestion at threshold")

	// Call 4: 1 past threshold, interval=2, no suggestion.
	assert.Empty(t, responses[3].Stderr, "call 4: no suggestion between intervals")

	// Call 5: 2 past threshold, interval=2, should suggest.
	assert.NotEmpty(t, responses[4].Stderr, "call 5: suggestion at reminder interval")
}

func TestSuggestCompactHandler_SeparateSessions(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, "compact")

	cfg := newTestConfig()
	cfg.Compact.Threshold = 3
	cfg.Compact.ReminderInterval = 10

	h := handler.NewSuggestCompactHandler(cfg, handler.WithCompactStateDir(stateDir))

	inputA := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		SessionID:     "session-a",
	}
	inputB := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		SessionID:     "session-b",
	}

	// 2 calls on session-a (below threshold).
	for range 2 {
		resp, err := h.Handle(context.Background(), inputA)
		require.NoError(t, err)
		assert.Empty(t, resp.Stderr, "session-a below threshold")
	}

	// 3 calls on session-b — independent counter hits threshold at call 3.
	var lastB *handler.Response
	for range 3 {
		resp, err := h.Handle(context.Background(), inputB)
		require.NoError(t, err)
		lastB = resp
	}
	assert.NotEmpty(t, lastB.Stderr, "session-b should hit threshold independently")

	// Verify session-a counter file contains "2".
	counterA := filepath.Join(stateDir, "cc-tools-compact-session-a.count")
	data, err := os.ReadFile(counterA)
	require.NoError(t, err)
	assert.Equal(t, "2", strings.TrimSpace(string(data)),
		"session-a counter should be 2")
}

func TestSuggestCompactHandler_CounterFileIncrement(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, "compact")

	cfg := newTestConfig()
	cfg.Compact.Threshold = 100
	cfg.Compact.ReminderInterval = 100

	h := handler.NewSuggestCompactHandler(cfg, handler.WithCompactStateDir(stateDir))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		SessionID:     "counter-test",
	}

	for range 7 {
		_, err := h.Handle(context.Background(), input)
		require.NoError(t, err)
	}

	counterFile := filepath.Join(stateDir, "cc-tools-compact-counter-test.count")
	data, err := os.ReadFile(counterFile)
	require.NoError(t, err)
	assert.Equal(t, "7", strings.TrimSpace(string(data)),
		"counter file should contain 7 after 7 calls")
}

func TestSuggestCompactHandler_ZeroThreshold(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, "compact")

	cfg := newTestConfig()
	cfg.Compact.Threshold = 0
	cfg.Compact.ReminderInterval = 0

	h := handler.NewSuggestCompactHandler(cfg, handler.WithCompactStateDir(stateDir))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		SessionID:     "zero-threshold",
	}

	// Make 10 calls — none should ever suggest.
	for range 10 {
		resp, err := h.Handle(context.Background(), input)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Stderr, "zero threshold should never suggest")
	}
}

func TestSuggestCompactHandler_SuggestionMessage(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, "compact")

	cfg := newTestConfig()
	cfg.Compact.Threshold = 2
	cfg.Compact.ReminderInterval = 5

	h := handler.NewSuggestCompactHandler(cfg, handler.WithCompactStateDir(stateDir))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		SessionID:     "msg-test",
	}

	// Make 2 calls to hit threshold.
	var lastResp *handler.Response
	for range 2 {
		resp, err := h.Handle(context.Background(), input)
		require.NoError(t, err)
		lastResp = resp
	}

	require.NotNil(t, lastResp)
	assert.Contains(t, lastResp.Stderr, "2 tool calls",
		"message should mention tool call count")
	assert.Contains(t, lastResp.Stderr, "/compact",
		"message should mention /compact")
}

func TestSuggestCompactHandler_ImplementsHandler(t *testing.T) {
	t.Parallel()
	var _ handler.Handler = handler.NewSuggestCompactHandler(nil)
}

// ---------------------------------------------------------------------
// ObserveHandler
// ---------------------------------------------------------------------

func TestObserveHandler_Name(t *testing.T) {
	t.Parallel()

	tests := []struct {
		phase string
		want  string
	}{
		{"pre", "observe-pre"},
		{"post", "observe-post"},
		{"failure", "observe-failure"},
	}

	for _, tt := range tests {
		t.Run(tt.phase, func(t *testing.T) {
			t.Parallel()
			h := handler.NewObserveHandler(nil, tt.phase)
			assert.Equal(t, tt.want, h.Name())
		})
	}
}

func TestObserveHandler_NilConfig(t *testing.T) {
	t.Parallel()
	h := handler.NewObserveHandler(nil, "pre")
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		ToolName:      "Bash",
		SessionID:     "test-session",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
}

func TestObserveHandler_Disabled(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()
	cfg.Observe.Enabled = false

	h := handler.NewObserveHandler(cfg, "pre")
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		ToolName:      "Read",
		SessionID:     "disabled-session",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
}

func TestObserveHandler_RecordsEvent(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	obsDir := filepath.Join(tmpDir, "observations")

	cfg := newTestConfig()
	cfg.Observe.Enabled = true
	cfg.Observe.MaxFileSizeMB = 10

	h := handler.NewObserveHandler(cfg, "pre", handler.WithObserveDir(obsDir))

	toolInput, _ := json.Marshal(map[string]string{"command": "ls"})
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		ToolName:      "Bash",
		ToolInput:     toolInput,
		SessionID:     "observe-session",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	// Observations file should exist.
	obsFile := filepath.Join(obsDir, "observations.jsonl")
	data, readErr := os.ReadFile(obsFile)
	require.NoError(t, readErr, "observations file should exist")
	assert.Contains(t, string(data), "Bash")
	assert.Contains(t, string(data), "observe-session")
}

func TestObserveHandler_PostPhase(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	obsDir := filepath.Join(tmpDir, "observations")

	cfg := newTestConfig()
	cfg.Observe.Enabled = true
	cfg.Observe.MaxFileSizeMB = 10

	h := handler.NewObserveHandler(cfg, "post", handler.WithObserveDir(obsDir))

	toolInput, _ := json.Marshal(map[string]string{"command": "ls"})
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPostToolUse,
		ToolName:      "Bash",
		ToolInput:     toolInput,
		SessionID:     "post-phase-session",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	obsFile := filepath.Join(obsDir, "observations.jsonl")
	data, readErr := os.ReadFile(obsFile)
	require.NoError(t, readErr, "observations file should exist")
	assert.Contains(t, string(data), `"phase":"post"`)
}

func TestObserveHandler_FailurePhase(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	obsDir := filepath.Join(tmpDir, "observations")

	cfg := newTestConfig()
	cfg.Observe.Enabled = true
	cfg.Observe.MaxFileSizeMB = 10

	h := handler.NewObserveHandler(cfg, "failure", handler.WithObserveDir(obsDir))

	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPostToolUseFailure,
		ToolName:      "Bash",
		SessionID:     "failure-phase-session",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	obsFile := filepath.Join(obsDir, "observations.jsonl")
	data, readErr := os.ReadFile(obsFile)
	require.NoError(t, readErr, "observations file should exist")
	assert.Contains(t, string(data), `"phase":"failure"`)
}

func TestObserveHandler_MultipleEventsAppend(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	obsDir := filepath.Join(tmpDir, "observations")

	cfg := newTestConfig()
	cfg.Observe.Enabled = true
	cfg.Observe.MaxFileSizeMB = 10

	h := handler.NewObserveHandler(cfg, "pre", handler.WithObserveDir(obsDir))

	tools := []string{"Bash", "Read", "Edit"}
	for _, tool := range tools {
		input := &hookcmd.HookInput{
			HookEventName: hookcmd.EventPreToolUse,
			ToolName:      tool,
			SessionID:     "multi-event-session",
		}

		resp, err := h.Handle(context.Background(), input)
		require.NoError(t, err)
		require.NotNil(t, resp)
	}

	obsFile := filepath.Join(obsDir, "observations.jsonl")
	data, readErr := os.ReadFile(obsFile)
	require.NoError(t, readErr, "observations file should exist")

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	assert.Len(t, lines, 3, "should have 3 lines for 3 events")

	for i, line := range lines {
		assert.True(t, json.Valid([]byte(line)), "line %d should be valid JSON", i+1)
	}
}

func TestObserveHandler_DisabledMarkerFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	obsDir := filepath.Join(tmpDir, "observations")

	cfg := newTestConfig()
	cfg.Observe.Enabled = true
	cfg.Observe.MaxFileSizeMB = 10

	// Create the obsDir and place a .disabled marker file inside it.
	require.NoError(t, os.MkdirAll(obsDir, 0o750))
	disabledPath := filepath.Join(obsDir, ".disabled")
	require.NoError(t, os.WriteFile(disabledPath, []byte{}, 0o600))

	h := handler.NewObserveHandler(cfg, "pre", handler.WithObserveDir(obsDir))

	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		ToolName:      "Bash",
		SessionID:     "disabled-marker-session",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	obsFile := filepath.Join(obsDir, "observations.jsonl")
	_, statErr := os.Stat(obsFile)
	assert.True(t, os.IsNotExist(statErr), "observations.jsonl should not exist when .disabled marker is present")
}

func TestObserveHandler_EmptyToolInput(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	obsDir := filepath.Join(tmpDir, "observations")

	cfg := newTestConfig()
	cfg.Observe.Enabled = true
	cfg.Observe.MaxFileSizeMB = 10

	h := handler.NewObserveHandler(cfg, "pre", handler.WithObserveDir(obsDir))

	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		ToolName:      "Read",
		ToolInput:     nil,
		SessionID:     "empty-input",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	obsFile := filepath.Join(obsDir, "observations.jsonl")
	data, readErr := os.ReadFile(obsFile)
	require.NoError(t, readErr, "observations file should exist")
	assert.Contains(t, string(data), "Read")
}

func TestObserveHandler_ImplementsHandler(t *testing.T) {
	t.Parallel()
	var _ handler.Handler = handler.NewObserveHandler(nil, "pre")
}

// ---------------------------------------------------------------------
// PreCommitReminderHandler
// ---------------------------------------------------------------------

func TestPreCommitReminderHandler_Name(t *testing.T) {
	t.Parallel()
	h := handler.NewPreCommitReminderHandler(nil)
	assert.Equal(t, "pre-commit-reminder", h.Name())
}

func TestPreCommitReminderHandler_NilConfig(t *testing.T) {
	t.Parallel()
	h := handler.NewPreCommitReminderHandler(nil)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		ToolName:      "Bash",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
}

func TestPreCommitReminderHandler_Disabled(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()
	cfg.PreCommit.Enabled = false

	h := handler.NewPreCommitReminderHandler(cfg)

	toolInput, _ := json.Marshal(map[string]string{"command": "git commit -m 'test'"})
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		ToolName:      "Bash",
		ToolInput:     toolInput,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Stderr, "no reminder when disabled")
}

func TestPreCommitReminderHandler_NonBashTool(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()
	cfg.PreCommit.Enabled = true
	cfg.PreCommit.Command = "task pre-commit"

	h := handler.NewPreCommitReminderHandler(cfg)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		ToolName:      "Read",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Stderr, "no reminder for non-Bash tools")
}

func TestPreCommitReminderHandler_GitCommitDetected(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()
	cfg.PreCommit.Enabled = true
	cfg.PreCommit.Command = "task pre-commit"

	h := handler.NewPreCommitReminderHandler(cfg)

	toolInput, _ := json.Marshal(map[string]string{"command": "git commit -m 'fix: something'"})
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		ToolName:      "Bash",
		ToolInput:     toolInput,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Stderr, "task pre-commit",
		"should remind about pre-commit command")
}

func TestPreCommitReminderHandler_DefaultCommand(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()
	cfg.PreCommit.Enabled = true

	h := handler.NewPreCommitReminderHandler(cfg)

	toolInput, _ := json.Marshal(map[string]string{"command": "git commit -am 'test'"})
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		ToolName:      "Bash",
		ToolInput:     toolInput,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Stderr, "task pre-commit",
		"should use default command when not configured")
}

func TestPreCommitReminderHandler_NoGitCommit(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()
	cfg.PreCommit.Enabled = true
	cfg.PreCommit.Command = "task pre-commit"

	h := handler.NewPreCommitReminderHandler(cfg)

	toolInput, _ := json.Marshal(map[string]string{"command": "git status"})
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		ToolName:      "Bash",
		ToolInput:     toolInput,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Stderr, "no reminder for non-commit git commands")
}

func TestPreCommitReminderHandler_GitCommitAmFlag(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()
	cfg.PreCommit.Enabled = true
	cfg.PreCommit.Command = "task pre-commit"

	h := handler.NewPreCommitReminderHandler(cfg)

	toolInput, _ := json.Marshal(map[string]string{"command": "git commit -am 'quick fix'"})
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		ToolName:      "Bash",
		ToolInput:     toolInput,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Stderr, "task pre-commit",
		"should remind about pre-commit for git commit -am")
}

func TestPreCommitReminderHandler_ChainedGitCommit(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()
	cfg.PreCommit.Enabled = true
	cfg.PreCommit.Command = "task pre-commit"

	h := handler.NewPreCommitReminderHandler(cfg)

	toolInput, _ := json.Marshal(map[string]string{"command": "git add . && git commit -m 'fix: resolve race'"})
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		ToolName:      "Bash",
		ToolInput:     toolInput,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Stderr, "task pre-commit",
		"should remind about pre-commit for chained git commit")
}

func TestPreCommitReminderHandler_CustomCommand(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()
	cfg.PreCommit.Enabled = true
	cfg.PreCommit.Command = "make check"

	h := handler.NewPreCommitReminderHandler(cfg)

	toolInput, _ := json.Marshal(map[string]string{"command": "git commit -m 'test'"})
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		ToolName:      "Bash",
		ToolInput:     toolInput,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Stderr, "make check",
		"should use custom pre-commit command")
	assert.NotContains(t, resp.Stderr, "task pre-commit",
		"should not contain default command when custom is configured")
}

func TestPreCommitReminderHandler_ImplementsHandler(t *testing.T) {
	t.Parallel()
	var _ handler.Handler = handler.NewPreCommitReminderHandler(nil)
}
