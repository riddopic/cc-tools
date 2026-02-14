package handler_test

import (
	"context"
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

// ---------------------------------------------------------------------
// SessionEndHandler
// ---------------------------------------------------------------------

func TestSessionEndHandler_Name(t *testing.T) {
	t.Parallel()
	h := handler.NewSessionEndHandler(nil)
	assert.Equal(t, "session-end", h.Name())
}

func TestSessionEndHandler_NilConfig(t *testing.T) {
	t.Parallel()
	tmpHome := t.TempDir()

	h := handler.NewSessionEndHandler(nil, handler.WithSessionEndHomeDir(tmpHome))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionEnd,
		SessionID:     "nil-config-session",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
}

func TestSessionEndHandler_SavesSession(t *testing.T) {
	t.Parallel()
	tmpHome := t.TempDir()

	cfg := &config.Values{}

	h := handler.NewSessionEndHandler(cfg, handler.WithSessionEndHomeDir(tmpHome))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionEnd,
		SessionID:     "save-test-session",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	// Session file should exist.
	sessDir := filepath.Join(tmpHome, ".claude", "sessions")
	matches, _ := filepath.Glob(filepath.Join(sessDir, "*save-test-session.json"))
	assert.NotEmpty(t, matches, "session file should be created")
}

func TestSessionEndHandler_WithTranscript(t *testing.T) {
	t.Parallel()
	tmpHome := t.TempDir()

	// Create a transcript file.
	transcriptDir := t.TempDir()
	transcriptPath := filepath.Join(transcriptDir, "transcript.jsonl")
	lines := []string{
		`{"type":"human","content":"hello"}`,
		`{"type":"human","content":"fix it"}`,
		`{"type":"tool_use","name":"Edit","input":{"file_path":"/tmp/test.go"}}`,
	}
	require.NoError(t, os.WriteFile(transcriptPath, []byte(
		lines[0]+"\n"+lines[1]+"\n"+lines[2]+"\n",
	), 0o600))

	cfg := &config.Values{}

	h := handler.NewSessionEndHandler(cfg, handler.WithSessionEndHomeDir(tmpHome))
	input := &hookcmd.HookInput{
		HookEventName:  hookcmd.EventSessionEnd,
		SessionID:      "transcript-session",
		TranscriptPath: transcriptPath,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
}

func TestSessionEndHandler_LearningSignal(t *testing.T) {
	t.Parallel()
	tmpHome := t.TempDir()

	// Create a transcript with enough messages to trigger learning signal.
	transcriptDir := t.TempDir()
	transcriptPath := filepath.Join(transcriptDir, "transcript.jsonl")

	var b strings.Builder
	for range 15 {
		b.WriteString("{\"type\":\"human\",\"content\":\"message\"}\n")
	}
	content := b.String()
	require.NoError(t, os.WriteFile(transcriptPath, []byte(content), 0o600))

	cfg := &config.Values{
		Learning: config.LearningValues{
			MinSessionLength: 10,
		},
	}

	h := handler.NewSessionEndHandler(cfg, handler.WithSessionEndHomeDir(tmpHome))
	input := &hookcmd.HookInput{
		HookEventName:  hookcmd.EventSessionEnd,
		SessionID:      "learning-session",
		TranscriptPath: transcriptPath,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Stderr, "evaluate for extractable patterns")
}

func TestSessionEndHandler_DefaultMinSessionLength(t *testing.T) {
	t.Parallel()
	tmpHome := t.TempDir()

	// Create a transcript with exactly 10 messages (default threshold).
	transcriptDir := t.TempDir()
	transcriptPath := filepath.Join(transcriptDir, "transcript.jsonl")

	var b strings.Builder
	for range 10 {
		b.WriteString("{\"type\":\"human\",\"content\":\"msg\"}\n")
	}
	content := b.String()
	require.NoError(t, os.WriteFile(transcriptPath, []byte(content), 0o600))

	// No learning config (use default minLength of 10).
	cfg := &config.Values{}

	h := handler.NewSessionEndHandler(cfg, handler.WithSessionEndHomeDir(tmpHome))
	input := &hookcmd.HookInput{
		HookEventName:  hookcmd.EventSessionEnd,
		SessionID:      "default-min-session",
		TranscriptPath: transcriptPath,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Stderr, "evaluate for extractable patterns")
}

func TestSessionEndHandler_ImplementsHandler(t *testing.T) {
	t.Parallel()
	var _ handler.Handler = handler.NewSessionEndHandler(nil)
}
