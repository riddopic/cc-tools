package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/hookcmd"
)

// Compile-time interface check.
var _ Handler = (*StopReminderHandler)(nil)

// StopReminderOption configures a StopReminderHandler.
type StopReminderOption func(*StopReminderHandler)

// WithStopStateDir overrides the state directory for testing.
func WithStopStateDir(dir string) StopReminderOption {
	return func(h *StopReminderHandler) {
		h.stateDir = dir
	}
}

// StopReminderHandler tracks response count per session and emits rotating
// reminders at configurable intervals. It fires on Stop events.
type StopReminderHandler struct {
	cfg      *config.Values
	stateDir string
}

// NewStopReminderHandler creates a new StopReminderHandler.
func NewStopReminderHandler(cfg *config.Values, opts ...StopReminderOption) *StopReminderHandler {
	h := &StopReminderHandler{
		cfg:      cfg,
		stateDir: "",
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// Name returns the handler identifier.
func (h *StopReminderHandler) Name() string { return "stop-reminder" }

// Handle processes a Stop event, incrementing the response counter and emitting
// a reminder when the configured interval or warning threshold is reached.
func (h *StopReminderHandler) Handle(_ context.Context, input *hookcmd.HookInput) (*Response, error) {
	if h.cfg == nil || !h.cfg.StopReminder.Enabled {
		return &Response{ExitCode: 0}, nil
	}

	stateDir := h.stateDir
	if stateDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home directory: %w", err)
		}
		stateDir = filepath.Join(homeDir, ".cache", "cc-tools", "stop")
	}

	count := h.readCount(stateDir, input.SessionID)
	count++
	h.writeCount(stateDir, input.SessionID, count)

	msg := h.reminderMessage(count)
	if msg != "" {
		return &Response{ExitCode: 0, Stderr: msg}, nil
	}

	return &Response{ExitCode: 0}, nil
}

func (h *StopReminderHandler) reminderMessage(count int) string {
	interval := h.cfg.StopReminder.Interval
	warnAt := h.cfg.StopReminder.WarnAt

	if warnAt > 0 && count >= warnAt {
		return fmt.Sprintf(
			"[cc-tools] Session has %d+ responses — strongly consider wrapping up and committing progress.\n",
			warnAt,
		)
	}

	if interval > 0 && count > 0 && count%interval == 0 {
		return stopReminders()[reminderIndex(count, interval)]
	}

	return ""
}

func reminderIndex(count, interval int) int {
	idx := (count / interval) - 1
	msgs := stopReminders()
	return idx % len(msgs)
}

func stopReminders() []string {
	return []string{
		"[cc-tools] Consider running /compact — context is getting heavy.\n",
		"[cc-tools] Long session — consider committing progress and capturing learnings.\n",
		"[cc-tools] Extended session — review your work and consider a checkpoint.\n",
	}
}

func (h *StopReminderHandler) counterPath(dir, sessionID string) string {
	return filepath.Join(dir, "stop-"+sessionID+".count")
}

func (h *StopReminderHandler) readCount(dir, sessionID string) int {
	data, err := os.ReadFile(h.counterPath(dir, sessionID)) // #nosec G304 -- path built from stateDir
	if err != nil {
		return 0
	}

	count, parseErr := strconv.Atoi(strings.TrimSpace(string(data)))
	if parseErr != nil {
		return 0
	}

	return count
}

func (h *StopReminderHandler) writeCount(dir, sessionID string, count int) {
	_ = os.MkdirAll(dir, 0o750)
	_ = os.WriteFile(
		h.counterPath(dir, sessionID),
		[]byte(strconv.Itoa(count)),
		0o600,
	)
}
