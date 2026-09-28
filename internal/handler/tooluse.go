package handler

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/riddopic/cc-tools/internal/compact"
	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/riddopic/cc-tools/internal/observe"
)

// Compile-time interface checks.
var (
	_ Handler = (*SuggestCompactHandler)(nil)
	_ Handler = (*ObserveHandler)(nil)
	_ Handler = (*PreCommitReminderHandler)(nil)
)

// defaultPreCommitCommand is the fallback pre-commit command.
const defaultPreCommitCommand = "task pre-commit"

// ---------------------------------------------------------------------
// SuggestCompactHandler
// ---------------------------------------------------------------------

// SuggestCompactOption configures a SuggestCompactHandler.
type SuggestCompactOption func(*SuggestCompactHandler)

// WithCompactStateDir overrides the state directory for testing.
func WithCompactStateDir(dir string) SuggestCompactOption {
	return func(h *SuggestCompactHandler) {
		h.stateDir = dir
	}
}

// SuggestCompactHandler records tool calls and suggests compaction when a
// threshold is exceeded.
type SuggestCompactHandler struct {
	cfg      *config.Values
	stateDir string
}

// NewSuggestCompactHandler creates a new SuggestCompactHandler.
func NewSuggestCompactHandler(cfg *config.Values, opts ...SuggestCompactOption) *SuggestCompactHandler {
	h := &SuggestCompactHandler{
		cfg:      cfg,
		stateDir: "",
	}
	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Name returns the handler identifier.
func (h *SuggestCompactHandler) Name() string { return "suggest-compact" }

// Handle records a tool call and shows the user a /compact suggestion when the
// session threshold is reached. Only the user can run /compact, so the
// suggestion is a system message rather than context for Claude.
func (h *SuggestCompactHandler) Handle(_ context.Context, input *hookcmd.HookInput) (*Response, error) {
	if h.cfg == nil {
		return &Response{ExitCode: 0}, nil
	}

	stateDir := h.stateDir
	if stateDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home directory: %w", err)
		}

		stateDir = filepath.Join(homeDir, ".cache", "cc-tools", "compact")
	}

	s := compact.NewSuggestor(stateDir, h.cfg.Compact.Threshold, h.cfg.Compact.ReminderInterval)

	var buf bytes.Buffer
	s.RecordCall(input.SessionID, &buf)

	msg := strings.TrimSpace(buf.String())
	if msg == "" {
		return &Response{ExitCode: 0}, nil
	}

	return UserMessageResponse(msg), nil
}

// ---------------------------------------------------------------------
// ObserveHandler
// ---------------------------------------------------------------------

// ObserveOption configures an ObserveHandler.
type ObserveOption func(*ObserveHandler)

// WithObserveDir overrides the observation directory for testing.
func WithObserveDir(dir string) ObserveOption {
	return func(h *ObserveHandler) {
		h.dir = dir
	}
}

// ObserveHandler records tool usage events for analytics.
type ObserveHandler struct {
	cfg   *config.Values
	phase string
	dir   string
}

// NewObserveHandler creates a new ObserveHandler for the given phase.
// Phase should be "pre", "post", or "failure".
func NewObserveHandler(cfg *config.Values, phase string, opts ...ObserveOption) *ObserveHandler {
	h := &ObserveHandler{
		cfg:   cfg,
		phase: phase,
		dir:   "",
	}
	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Name returns the handler identifier.
func (h *ObserveHandler) Name() string { return "observe-" + h.phase }

// Handle records a tool usage event to the observations JSONL file.
func (h *ObserveHandler) Handle(_ context.Context, input *hookcmd.HookInput) (*Response, error) {
	if h.cfg == nil || !h.cfg.Observe.Enabled {
		return &Response{ExitCode: 0}, nil
	}

	dir := h.dir
	if dir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home directory: %w", err)
		}

		dir = filepath.Join(homeDir, ".cache", "cc-tools", "observations")
	}

	obs := observe.NewObserver(dir, h.cfg.Observe.MaxFileSizeMB)

	if err := obs.Record(observe.Event{
		Timestamp:  time.Now(),
		Phase:      h.phase,
		ToolName:   input.ToolName,
		ToolInput:  input.ToolInput,
		ToolOutput: input.ToolOutput,
		Error:      input.Error,
		SessionID:  string(input.SessionID),
	}); err != nil {
		return nil, fmt.Errorf("record observation: %w", err)
	}

	return &Response{ExitCode: 0}, nil
}

// ---------------------------------------------------------------------
// PreCommitReminderHandler
// ---------------------------------------------------------------------

// PreCommitReminderHandler adds a reminder to Claude's context when a git
// commit command is detected.
type PreCommitReminderHandler struct {
	cfg *config.Values
}

// NewPreCommitReminderHandler creates a new PreCommitReminderHandler.
func NewPreCommitReminderHandler(cfg *config.Values) *PreCommitReminderHandler {
	return &PreCommitReminderHandler{cfg: cfg}
}

// Name returns the handler identifier.
func (h *PreCommitReminderHandler) Name() string { return "pre-commit-reminder" }

// Handle checks if the tool input contains a git commit command and tells
// Claude which pre-commit command the project expects. Claude is the one
// running the commit, so the reminder goes to its context, not the user.
func (h *PreCommitReminderHandler) Handle(_ context.Context, input *hookcmd.HookInput) (*Response, error) {
	if h.cfg == nil || !h.cfg.PreCommit.Enabled {
		return &Response{ExitCode: 0}, nil
	}

	if input.ToolName != "Bash" {
		return &Response{ExitCode: 0}, nil
	}

	command := input.GetToolInputString("command")
	if !strings.Contains(command, "git commit") {
		return &Response{ExitCode: 0}, nil
	}

	reminder := h.cfg.PreCommit.Command
	if reminder == "" {
		reminder = defaultPreCommitCommand
	}

	return ContextResponse(hookcmd.EventPreToolUse,
		fmt.Sprintf("This project runs '%s' (fmt + lint + test) before every commit.", reminder)), nil
}
