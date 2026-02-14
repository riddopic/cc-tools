package handler_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
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

func TestPreCommitReminderHandler_ImplementsHandler(t *testing.T) {
	t.Parallel()
	var _ handler.Handler = handler.NewPreCommitReminderHandler(nil)
}
