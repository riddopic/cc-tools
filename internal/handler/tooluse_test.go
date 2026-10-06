package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/riddopic/cc-tools/internal/observe"
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
			ContextTokens: 0,
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
		PackageManager: config.PackageManagerValues{
			Preferred: "",
		},
		Drift: config.DriftValues{
			Enabled:   false,
			MinEdits:  0,
			Threshold: 0,
		},
		StopReminder: config.StopReminderValues{
			Enabled:  false,
			Interval: 0,
			WarnAt:   0,
		},
		State: config.StateValues{
			MaxAgeDays: 0,
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

// writeCompactTranscript writes a transcript whose last main-chain assistant
// turn reports the given context size, and returns its path.
func writeCompactTranscript(t *testing.T, dir string, tokens int) string {
	t.Helper()

	path := filepath.Join(dir, "transcript.jsonl")
	line := fmt.Sprintf(
		`{"type":"assistant","isSidechain":false,"message":{"usage":`+
			`{"input_tokens":%d,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`,
		tokens,
	)
	require.NoError(t, os.WriteFile(path, []byte(line+"\n"), 0o600))

	return path
}

func newCompactHandler(t *testing.T, threshold int) (*handler.SuggestCompactHandler, string, string) {
	t.Helper()

	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, "compact")

	cfg := newTestConfig()
	cfg.Compact.ContextTokens = threshold

	return handler.NewSuggestCompactHandler(cfg, handler.WithCompactStateDir(stateDir)), stateDir, tmpDir
}

func compactInput(session, transcript string) *hookcmd.HookInput {
	return &hookcmd.HookInput{
		HookEventName:  hookcmd.EventPreToolUse,
		SessionID:      hookcmd.SessionID(session),
		TranscriptPath: transcript,
	}
}

func TestSuggestCompactHandler_BelowThreshold(t *testing.T) {
	t.Parallel()
	h, _, tmpDir := newCompactHandler(t, 100_000)
	input := compactInput("below-threshold", writeCompactTranscript(t, tmpDir, 50_000))

	for range 3 {
		resp, err := h.Handle(context.Background(), input)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.SystemMessage(), "no suggestion below threshold")
	}
}

func TestSuggestCompactHandler_NudgesOnceAboveThreshold(t *testing.T) {
	t.Parallel()
	h, stateDir, tmpDir := newCompactHandler(t, 100_000)
	input := compactInput("above-threshold", writeCompactTranscript(t, tmpDir, 152_000))

	first, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t,
		"[cc-tools] Context is ~152k tokens. Consider running /compact to reduce context usage.",
		first.SystemMessage())

	second, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Empty(t, second.SystemMessage(), "second call above threshold is silent")

	_, statErr := os.Stat(filepath.Join(stateDir, "cc-tools-compact-above-threshold.json"))
	assert.NoError(t, statErr, "state file should be created")
}

func TestSuggestCompactHandler_RearmsAfterDrop(t *testing.T) {
	t.Parallel()
	h, _, tmpDir := newCompactHandler(t, 100_000)

	steps := []struct {
		tokens int
		nudge  bool
	}{
		{tokens: 120_000, nudge: true},
		{tokens: 130_000, nudge: false},
		{tokens: 20_000, nudge: false},
		{tokens: 110_000, nudge: true},
		{tokens: 115_000, nudge: false},
	}

	for i, step := range steps {
		input := compactInput("rearm", writeCompactTranscript(t, tmpDir, step.tokens))
		resp, err := h.Handle(context.Background(), input)
		require.NoError(t, err)

		if step.nudge {
			assert.Contains(t, resp.SystemMessage(), "/compact", "step %d", i)
		} else {
			assert.Empty(t, resp.SystemMessage(), "step %d", i)
		}
	}
}

func TestSuggestCompactHandler_SkipsSubagents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		agentID   string
		agentType string
	}{
		{name: "agent_id set", agentID: "agent-123", agentType: ""},
		{name: "agent_type set", agentID: "", agentType: "Explore"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h, stateDir, tmpDir := newCompactHandler(t, 100_000)
			input := compactInput("subagent", writeCompactTranscript(t, tmpDir, 500_000))
			input.AgentID = tt.agentID
			input.AgentType = tt.agentType

			resp, err := h.Handle(context.Background(), input)
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Empty(t, resp.SystemMessage(), "subagent calls never nudge")

			_, statErr := os.Stat(stateDir)
			assert.True(t, os.IsNotExist(statErr), "no state written for subagent calls")
		})
	}
}

func TestSuggestCompactHandler_NoTranscript(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		transcript func(dir string) string
	}{
		{name: "empty path", transcript: func(string) string { return "" }},
		{name: "missing file", transcript: func(dir string) string { return filepath.Join(dir, "missing.jsonl") }},
		{name: "empty file", transcript: func(dir string) string {
			path := filepath.Join(dir, "empty.jsonl")
			_ = os.WriteFile(path, nil, 0o600)

			return path
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h, stateDir, tmpDir := newCompactHandler(t, 1)

			resp, err := h.Handle(context.Background(), compactInput("no-transcript", tt.transcript(tmpDir)))
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Empty(t, resp.SystemMessage())

			_, statErr := os.Stat(stateDir)
			assert.True(t, os.IsNotExist(statErr), "no state written without a transcript")
		})
	}
}

func TestSuggestCompactHandler_SeparateSessions(t *testing.T) {
	t.Parallel()
	h, _, tmpDir := newCompactHandler(t, 100_000)
	transcript := writeCompactTranscript(t, tmpDir, 200_000)

	respA, err := h.Handle(context.Background(), compactInput("session-a", transcript))
	require.NoError(t, err)
	assert.NotEmpty(t, respA.SystemMessage())

	respB, err := h.Handle(context.Background(), compactInput("session-b", transcript))
	require.NoError(t, err)
	assert.NotEmpty(t, respB.SystemMessage(), "session-b has independent state")
}

func TestSuggestCompactHandler_ZeroThresholdDisables(t *testing.T) {
	t.Parallel()
	h, stateDir, tmpDir := newCompactHandler(t, 0)
	input := compactInput("zero-threshold", writeCompactTranscript(t, tmpDir, 500_000))

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.SystemMessage(), "zero threshold should never suggest")

	_, statErr := os.Stat(stateDir)
	assert.True(t, os.IsNotExist(statErr))
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

func TestObserveHandler_PostPhaseRecordsToolOutput(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	obsDir := filepath.Join(tmpDir, "observations")

	cfg := newTestConfig()
	cfg.Observe.Enabled = true
	cfg.Observe.MaxFileSizeMB = 10

	h := handler.NewObserveHandler(cfg, "post", handler.WithObserveDir(obsDir))

	toolInput, _ := json.Marshal(map[string]string{"command": "echo hello"})
	toolOutput := json.RawMessage(`{"stdout":"hello\n","stderr":""}`)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPostToolUse,
		ToolName:      "Bash",
		ToolInput:     toolInput,
		ToolOutput:    toolOutput,
		SessionID:     "post-output-session",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	obsFile := filepath.Join(obsDir, "observations.jsonl")
	data, readErr := os.ReadFile(obsFile)
	require.NoError(t, readErr, "observations file should exist")

	var event observe.Event
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(data), &event))
	assert.Equal(t, "post", event.Phase)
	assert.JSONEq(t, string(toolOutput), string(event.ToolOutput),
		"recorded event should contain the tool output")
	assert.Empty(t, event.Error, "post-phase event should have no error")
}

func TestObserveHandler_FailurePhaseRecordsError(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	obsDir := filepath.Join(tmpDir, "observations")

	cfg := newTestConfig()
	cfg.Observe.Enabled = true
	cfg.Observe.MaxFileSizeMB = 10

	h := handler.NewObserveHandler(cfg, "failure", handler.WithObserveDir(obsDir))

	toolInput, _ := json.Marshal(map[string]string{"command": "rm /protected"})
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPostToolUseFailure,
		ToolName:      "Bash",
		ToolInput:     toolInput,
		Error:         "permission denied: /protected",
		SessionID:     "failure-error-session",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	obsFile := filepath.Join(obsDir, "observations.jsonl")
	data, readErr := os.ReadFile(obsFile)
	require.NoError(t, readErr, "observations file should exist")

	var event observe.Event
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(data), &event))
	assert.Equal(t, "failure", event.Phase)
	assert.Equal(t, "permission denied: /protected", event.Error,
		"recorded event should contain the error string")
	assert.Nil(t, event.ToolOutput, "failure-phase event should have no tool output")
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
	assert.Empty(t, resp.AdditionalContext(), "no reminder when disabled")
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
	assert.Empty(t, resp.AdditionalContext(), "no reminder for non-Bash tools")
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
	assert.Contains(t, resp.AdditionalContext(), "task pre-commit",
		"should remind about pre-commit command")
	require.NotNil(t, resp.Stdout)
	require.NotNil(t, resp.Stdout.HookSpecificOutput)
	assert.Equal(t, hookcmd.EventPreToolUse, resp.Stdout.HookSpecificOutput.HookEventName)
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
	assert.Contains(t, resp.AdditionalContext(), "task pre-commit",
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
	assert.Empty(t, resp.AdditionalContext(), "no reminder for non-commit git commands")
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
	assert.Contains(t, resp.AdditionalContext(), "task pre-commit",
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
	assert.Contains(t, resp.AdditionalContext(), "task pre-commit",
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
	assert.Contains(t, resp.AdditionalContext(), "make check",
		"should use custom pre-commit command")
	assert.NotContains(t, resp.AdditionalContext(), "task pre-commit",
		"should not contain default command when custom is configured")
}

func TestPreCommitReminderHandler_ImplementsHandler(t *testing.T) {
	t.Parallel()
	var _ handler.Handler = handler.NewPreCommitReminderHandler(nil)
}
