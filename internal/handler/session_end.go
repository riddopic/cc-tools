package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/riddopic/cc-tools/internal/session"
)

// Compile-time interface check.
var _ Handler = (*SessionEndHandler)(nil)

// defaultMinSessionLength is the default minimum message count to trigger
// a continuous learning signal.
const defaultMinSessionLength = 10

// SessionEndOption configures a SessionEndHandler.
type SessionEndOption func(*SessionEndHandler)

// WithSessionEndHomeDir overrides the home directory for testing.
func WithSessionEndHomeDir(dir string) SessionEndOption {
	return func(h *SessionEndHandler) {
		h.homeDir = dir
	}
}

// SessionEndHandler saves session metadata and emits a learning signal
// on session end.
type SessionEndHandler struct {
	cfg     *config.Values
	homeDir string
}

// NewSessionEndHandler creates a new SessionEndHandler.
func NewSessionEndHandler(cfg *config.Values, opts ...SessionEndOption) *SessionEndHandler {
	h := &SessionEndHandler{
		cfg:     cfg,
		homeDir: "",
	}
	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Name returns the handler identifier.
func (h *SessionEndHandler) Name() string { return "session-end" }

// Handle saves the session and emits a continuous learning signal when
// the session had enough messages.
func (h *SessionEndHandler) Handle(_ context.Context, input *hookcmd.HookInput) (*Response, error) {
	homeDir := h.homeDir
	if homeDir == "" {
		var err error

		homeDir, err = os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home directory: %w", err)
		}
	}

	storeDir := filepath.Join(homeDir, ".claude", "sessions")
	store := session.NewStore(storeDir)

	// Parse transcript if available.
	var summary *session.TranscriptSummary
	if input.TranscriptPath != "" {
		summary, _ = session.ParseTranscript(input.TranscriptPath)
	}

	// Build session metadata.
	now := time.Now()

	var toolsUsed []string
	var filesModified []string
	var messageCount int

	if summary != nil {
		toolsUsed = summary.ToolsUsed
		filesModified = summary.FilesModified
		messageCount = summary.TotalMessages
	}

	sess := &session.Session{
		Version:       "1",
		ID:            string(input.SessionID),
		Date:          now.Format("2006-01-02"),
		Started:       now,
		Ended:         now,
		Title:         fmt.Sprintf("Session %s", now.Format("15:04")),
		Summary:       "",
		ToolsUsed:     toolsUsed,
		FilesModified: filesModified,
		MessageCount:  messageCount,
	}

	var stderr string

	if saveErr := store.Save(sess); saveErr != nil {
		stderr += fmt.Sprintf("[session-end] save error: %v\n", saveErr)
	}

	// Continuous learning signal.
	minLength := defaultMinSessionLength
	if h.cfg != nil && h.cfg.Learning.MinSessionLength > 0 {
		minLength = h.cfg.Learning.MinSessionLength
	}

	if summary != nil && summary.TotalMessages >= minLength {
		stderr += fmt.Sprintf(
			"[session-end] %d messages — evaluate for extractable patterns\n",
			summary.TotalMessages)
	}

	return &Response{
		ExitCode: 0,
		Stderr:   stderr,
	}, nil
}
