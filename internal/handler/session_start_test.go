package handler_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/riddopic/cc-tools/internal/session"
)

// ---------------------------------------------------------------------
// SuperpowersHandler
// ---------------------------------------------------------------------

func TestSuperpowersHandler_Name(t *testing.T) {
	t.Parallel()
	h := handler.NewSuperpowersHandler()
	assert.Equal(t, "superpowers", h.Name())
}

func TestSuperpowersHandler_Handle_NoSkillFile(t *testing.T) {
	t.Parallel()
	h := handler.NewSuperpowersHandler()
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           t.TempDir(),
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
	assert.Nil(t, resp.Stdout, "no output when skill file is absent")
}

func TestSuperpowersHandler_Handle_WithSkillFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, ".claude", "skills", "using-superpowers")
	require.NoError(t, os.MkdirAll(skillDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(skillDir, "SKILL.md"),
		[]byte("Use /superpowers to discover skills."),
		0o600,
	))

	h := handler.NewSuperpowersHandler()
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           tmpDir,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
	require.NotNil(t, resp.Stdout, "should produce output when skill file exists")
	assert.NotNil(t, resp.Stdout.HookSpecificOutput, "should populate hookSpecificOutput")
}

func TestSuperpowersHandler_ImplementsHandler(t *testing.T) {
	t.Parallel()
	var _ handler.Handler = handler.NewSuperpowersHandler()
}

// ---------------------------------------------------------------------
// PkgManagerHandler
// ---------------------------------------------------------------------

func TestPkgManagerHandler_Name(t *testing.T) {
	t.Parallel()
	h := handler.NewPkgManagerHandler()
	assert.Equal(t, "pkg-manager", h.Name())
}

func TestPkgManagerHandler_Handle_CreatesEnvFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	h := handler.NewPkgManagerHandler()
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           tmpDir,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	// Verify .claude/.env was created.
	envFile := filepath.Join(tmpDir, ".claude", ".env")
	data, readErr := os.ReadFile(envFile)
	require.NoError(t, readErr, "env file should exist")
	assert.Contains(t, string(data), "PREFERRED_PACKAGE_MANAGER=")
}

func TestPkgManagerHandler_Handle_DetectsYarn(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create a yarn.lock file so detection picks yarn.
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "yarn.lock"), []byte(""), 0o600))

	h := handler.NewPkgManagerHandler()
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           tmpDir,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	envFile := filepath.Join(tmpDir, ".claude", ".env")
	data, readErr := os.ReadFile(envFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "PREFERRED_PACKAGE_MANAGER=yarn")
}

func TestPkgManagerHandler_Handle_NoStdout(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	h := handler.NewPkgManagerHandler()
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           tmpDir,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Nil(t, resp.Stdout, "pkg-manager handler should not produce stdout output")
}

func TestPkgManagerHandler_ImplementsHandler(t *testing.T) {
	t.Parallel()
	var _ handler.Handler = handler.NewPkgManagerHandler()
}

// ---------------------------------------------------------------------
// SessionContextHandler
// ---------------------------------------------------------------------

func TestSessionContextHandler_Name(t *testing.T) {
	t.Parallel()
	h := handler.NewSessionContextHandler()
	assert.Equal(t, "session-context", h.Name())
}

func TestSessionContextHandler_Handle_NoSessions(t *testing.T) {
	t.Parallel()
	tmpHome := t.TempDir()

	h := handler.NewSessionContextHandler(handler.WithHomeDir(tmpHome))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
	assert.Nil(t, resp.Stdout, "no output when no sessions exist")
}

func TestSessionContextHandler_Handle_WithPreviousSession(t *testing.T) {
	t.Parallel()
	tmpHome := t.TempDir()

	// Create a session file in the expected location.
	storeDir := filepath.Join(tmpHome, ".claude", "sessions")
	store := session.NewStore(storeDir)
	require.NoError(t, store.Save(&session.Session{
		Version:       "1",
		ID:            "test-session-123",
		Date:          "2025-01-15",
		Started:       time.Now(),
		Ended:         time.Time{},
		Title:         "Test session",
		Summary:       "Worked on refactoring",
		ToolsUsed:     nil,
		FilesModified: nil,
		MessageCount:  0,
	}))

	h := handler.NewSessionContextHandler(handler.WithHomeDir(tmpHome))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
	require.NotNil(t, resp.Stdout, "should produce output when previous session exists")
	assert.True(t, resp.Stdout.Continue)
	require.NotEmpty(t, resp.Stdout.AdditionalContext)
	assert.Contains(t, resp.Stdout.AdditionalContext[0], "Worked on refactoring")
	assert.Contains(t, resp.Stdout.AdditionalContext[0], "2025-01-15")
}

func TestSessionContextHandler_Handle_SessionWithEmptySummary(t *testing.T) {
	t.Parallel()
	tmpHome := t.TempDir()

	storeDir := filepath.Join(tmpHome, ".claude", "sessions")
	store := session.NewStore(storeDir)
	require.NoError(t, store.Save(&session.Session{
		Version:       "1",
		ID:            "empty-summary-session",
		Date:          "2025-01-15",
		Started:       time.Now(),
		Ended:         time.Time{},
		Title:         "No summary session",
		Summary:       "",
		ToolsUsed:     nil,
		FilesModified: nil,
		MessageCount:  0,
	}))

	h := handler.NewSessionContextHandler(handler.WithHomeDir(tmpHome))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
	// No stdout when summary is empty.
	assert.Nil(t, resp.Stdout)
}

func TestSessionContextHandler_Handle_WithAliases(t *testing.T) {
	t.Parallel()
	tmpHome := t.TempDir()

	// Create a session.
	storeDir := filepath.Join(tmpHome, ".claude", "sessions")
	store := session.NewStore(storeDir)
	require.NoError(t, store.Save(&session.Session{
		Version:       "1",
		ID:            "aliased-session",
		Date:          "2025-01-15",
		Started:       time.Now(),
		Ended:         time.Time{},
		Title:         "Aliased session",
		Summary:       "Has aliases",
		ToolsUsed:     nil,
		FilesModified: nil,
		MessageCount:  0,
	}))

	// Create aliases file.
	aliasFile := filepath.Join(tmpHome, ".claude", "session-aliases.json")
	aliasData, err := json.Marshal(map[string]string{
		"latest": "aliased-session",
		"prod":   "other-session",
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(aliasFile, aliasData, 0o600))

	h := handler.NewSessionContextHandler(handler.WithHomeDir(tmpHome))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Stderr, "[session-context]")
	assert.Contains(t, resp.Stderr, "2 alias(es)")
}

func TestSessionContextHandler_ImplementsHandler(t *testing.T) {
	t.Parallel()
	var _ handler.Handler = handler.NewSessionContextHandler()
}
