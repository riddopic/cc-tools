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

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/riddopic/cc-tools/internal/session"
)

// ---------------------------------------------------------------------
// PkgManagerHandler
// ---------------------------------------------------------------------

// noEnv stands in for [os.Getenv] so a PREFERRED_PACKAGE_MANAGER exported in the
// developer's shell cannot override lock file detection under test.
func noEnv(string) string { return "" }

func TestPkgManagerHandler_Name(t *testing.T) {
	t.Parallel()
	h := handler.NewPkgManagerHandler(nil)
	assert.Equal(t, "pkg-manager", h.Name())
}

func TestPkgManagerHandler_Handle_CreatesEnvFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	h := handler.NewPkgManagerHandler(nil,
		handler.WithEnvFilePath(filepath.Join(tmpDir, "claude.env")),
		handler.WithPkgManagerGetenv(noEnv),
	)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           tmpDir,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	// Verify the session env file was written.
	envFile := filepath.Join(tmpDir, "claude.env")
	data, readErr := os.ReadFile(envFile)
	require.NoError(t, readErr, "env file should exist")
	assert.Contains(t, string(data), "PREFERRED_PACKAGE_MANAGER=")
}

func TestPkgManagerHandler_Handle_DetectsYarn(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create a yarn.lock file so detection picks yarn.
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "yarn.lock"), []byte(""), 0o600))

	h := handler.NewPkgManagerHandler(nil,
		handler.WithEnvFilePath(filepath.Join(tmpDir, "claude.env")),
		handler.WithPkgManagerGetenv(noEnv),
	)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           tmpDir,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	envFile := filepath.Join(tmpDir, "claude.env")
	data, readErr := os.ReadFile(envFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "PREFERRED_PACKAGE_MANAGER=yarn")
}

func TestPkgManagerHandler_Handle_DetectsNpm(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "package-lock.json"), []byte("{}"), 0o600,
	))

	h := handler.NewPkgManagerHandler(nil,
		handler.WithEnvFilePath(filepath.Join(tmpDir, "claude.env")),
		handler.WithPkgManagerGetenv(noEnv),
	)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           tmpDir,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)

	envFile := filepath.Join(tmpDir, "claude.env")
	data, readErr := os.ReadFile(envFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "PREFERRED_PACKAGE_MANAGER=npm")
}

func TestPkgManagerHandler_Handle_DetectsPnpm(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "pnpm-lock.yaml"), []byte(""), 0o600,
	))

	h := handler.NewPkgManagerHandler(nil,
		handler.WithEnvFilePath(filepath.Join(tmpDir, "claude.env")),
		handler.WithPkgManagerGetenv(noEnv),
	)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           tmpDir,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)

	envFile := filepath.Join(tmpDir, "claude.env")
	data, readErr := os.ReadFile(envFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "PREFERRED_PACKAGE_MANAGER=pnpm")
}

func TestPkgManagerHandler_Handle_DetectsBun(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "bun.lock"), []byte(""), 0o600,
	))

	h := handler.NewPkgManagerHandler(nil,
		handler.WithEnvFilePath(filepath.Join(tmpDir, "claude.env")),
		handler.WithPkgManagerGetenv(noEnv),
	)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           tmpDir,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)

	envFile := filepath.Join(tmpDir, "claude.env")
	data, readErr := os.ReadFile(envFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "PREFERRED_PACKAGE_MANAGER=bun")
}

func TestPkgManagerHandler_Handle_EnvVarOverridesLockFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "yarn.lock"), []byte(""), 0o600))

	getenv := func(key string) string {
		if key == "PREFERRED_PACKAGE_MANAGER" {
			return "bun"
		}
		return ""
	}
	h := handler.NewPkgManagerHandler(nil,
		handler.WithEnvFilePath(filepath.Join(tmpDir, "claude.env")),
		handler.WithPkgManagerGetenv(getenv),
	)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           tmpDir,
	}

	_, err := h.Handle(context.Background(), input)
	require.NoError(t, err)

	data, readErr := os.ReadFile(filepath.Join(tmpDir, "claude.env"))
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "PREFERRED_PACKAGE_MANAGER=bun")
}

func TestPkgManagerHandler_Handle_NoStdout(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	h := handler.NewPkgManagerHandler(nil,
		handler.WithEnvFilePath(filepath.Join(tmpDir, "claude.env")),
		handler.WithPkgManagerGetenv(noEnv),
	)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           tmpDir,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Nil(t, resp.Stdout, "pkg-manager handler should not produce stdout output")
}

func TestPkgManagerHandler_Handle_ConfigPreferredOverridesLockFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create a yarn.lock so detection would normally pick yarn.
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "yarn.lock"), []byte(""), 0o600))

	cfg := config.GetDefaultConfig()
	cfg.PackageManager.Preferred = "bun"

	h := handler.NewPkgManagerHandler(cfg,
		handler.WithEnvFilePath(filepath.Join(tmpDir, "claude.env")),
		handler.WithPkgManagerGetenv(noEnv),
	)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           tmpDir,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	envFile := filepath.Join(tmpDir, "claude.env")
	data, readErr := os.ReadFile(envFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "PREFERRED_PACKAGE_MANAGER=bun",
		"config preferred should override lock file detection")
}

func TestPkgManagerHandler_Handle_EmptyConfigFallsBackToDetection(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create a pnpm-lock.yaml so detection picks pnpm.
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "pnpm-lock.yaml"), []byte(""), 0o600))

	cfg := config.GetDefaultConfig()
	// Preferred is empty — should fall through to lock file detection.

	h := handler.NewPkgManagerHandler(cfg,
		handler.WithEnvFilePath(filepath.Join(tmpDir, "claude.env")),
		handler.WithPkgManagerGetenv(noEnv),
	)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           tmpDir,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)

	envFile := filepath.Join(tmpDir, "claude.env")
	data, readErr := os.ReadFile(envFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "PREFERRED_PACKAGE_MANAGER=pnpm",
		"empty config preferred should fall back to lock file detection")
}

func TestPkgManagerHandler_Handle_NoEnvFileIsNoop(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Without CLAUDE_ENV_FILE there is nowhere Claude Code will read the value
	// from, so the handler must not litter the project with a .claude/.env.
	h := handler.NewPkgManagerHandler(nil,
		handler.WithEnvFilePath(""),
		handler.WithPkgManagerGetenv(noEnv),
	)
	resp, err := h.Handle(context.Background(), &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           tmpDir,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NoDirExists(t, filepath.Join(tmpDir, ".claude"))
}

func TestPkgManagerHandler_ImplementsHandler(t *testing.T) {
	t.Parallel()
	var _ handler.Handler = handler.NewPkgManagerHandler(nil)
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
		Cwd:           "/proj",
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
		Cwd:           "/proj",
		Title:         "Test session",
		Summary:       "Worked on refactoring",
		ToolsUsed:     nil,
		FilesModified: nil,
		MessageCount:  0,
	}))

	h := handler.NewSessionContextHandler(handler.WithHomeDir(tmpHome))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           "/proj",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
	require.NotNil(t, resp.Stdout, "should produce output when previous session exists")
	require.NotNil(t, resp.Stdout.HookSpecificOutput)
	assert.Equal(t, hookcmd.EventSessionStart, resp.Stdout.HookSpecificOutput.HookEventName)
	assert.Contains(t, resp.AdditionalContext(), "Worked on refactoring")
	assert.Contains(t, resp.AdditionalContext(), "2025-01-15")
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
		Cwd:           "/proj",
		Title:         "No summary session",
		Summary:       "",
		ToolsUsed:     nil,
		FilesModified: nil,
		MessageCount:  0,
	}))

	h := handler.NewSessionContextHandler(handler.WithHomeDir(tmpHome))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           "/proj",
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
		Cwd:           "/proj",
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
		Cwd:           "/proj",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Stderr, "[session-context]")
	assert.Contains(t, resp.Stderr, "2 alias(es)")
}

func TestSessionContextHandler_Handle_MultipleSessionsUsesRecent(t *testing.T) {
	t.Parallel()
	tmpHome := t.TempDir()

	storeDir := filepath.Join(tmpHome, ".claude", "sessions")
	store := session.NewStore(storeDir)

	// Save two sessions — older first, then newer.
	require.NoError(t, store.Save(&session.Session{
		Version:       "1",
		ID:            "older-session",
		Date:          "2025-01-10",
		Started:       time.Date(2025, 1, 10, 9, 0, 0, 0, time.UTC),
		Ended:         time.Time{},
		Cwd:           "/proj",
		Title:         "Older session",
		Summary:       "Old work done here",
		ToolsUsed:     nil,
		FilesModified: nil,
		MessageCount:  0,
	}))
	require.NoError(t, store.Save(&session.Session{
		Version:       "1",
		ID:            "newer-session",
		Date:          "2025-01-15",
		Started:       time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC),
		Ended:         time.Time{},
		Cwd:           "/proj",
		Title:         "Newer session",
		Summary:       "Recent work done here",
		ToolsUsed:     nil,
		FilesModified: nil,
		MessageCount:  0,
	}))

	h := handler.NewSessionContextHandler(handler.WithHomeDir(tmpHome))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           "/proj",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Stdout)
	assert.Contains(t, resp.AdditionalContext(), "Recent work done here",
		"should use most recent session's summary")
}

func TestSessionContextHandler_Handle_BySource(t *testing.T) {
	t.Parallel()
	tmpHome := t.TempDir()

	store := session.NewStore(filepath.Join(tmpHome, ".claude", "sessions"))
	require.NoError(t, store.Save(&session.Session{
		Version:       "1",
		ID:            "prior-session",
		Date:          "2025-01-15",
		Started:       time.Now(),
		Ended:         time.Time{},
		Cwd:           "/proj",
		Title:         "Prior",
		Summary:       "Prior work",
		ToolsUsed:     nil,
		FilesModified: nil,
		MessageCount:  0,
	}))

	tests := []struct {
		source     string
		wantInject bool
	}{
		{source: "", wantInject: true},
		{source: "startup", wantInject: true},
		{source: "clear", wantInject: true},
		// On resume/fork the "previous" session is this conversation, and after
		// compaction Claude already has its own summary of the session.
		{source: "resume", wantInject: false},
		{source: "fork", wantInject: false},
		{source: "compact", wantInject: false},
	}

	for _, tt := range tests {
		t.Run("source="+tt.source, func(t *testing.T) {
			t.Parallel()
			h := handler.NewSessionContextHandler(handler.WithHomeDir(tmpHome))
			input := &hookcmd.HookInput{HookEventName: hookcmd.EventSessionStart, Cwd: "/proj", Source: tt.source}

			resp, err := h.Handle(context.Background(), input)
			require.NoError(t, err)
			if tt.wantInject {
				assert.Contains(t, resp.AdditionalContext(), "Prior work")
			} else {
				assert.Empty(t, resp.AdditionalContext())
			}
		})
	}
}

func TestSessionContextHandler_Handle_ScopedToProject(t *testing.T) {
	t.Parallel()
	tmpHome := t.TempDir()
	store := session.NewStore(filepath.Join(tmpHome, ".claude", "sessions"))

	save := func(id, date, cwd, summary string) {
		require.NoError(t, store.Save(&session.Session{
			Version:       "1",
			ID:            id,
			Date:          date,
			Cwd:           cwd,
			Started:       time.Now(),
			Ended:         time.Now(),
			Title:         id,
			Summary:       summary,
			ToolsUsed:     nil,
			FilesModified: nil,
			MessageCount:  0,
		}))
	}
	save("mine", "2025-01-10", "/proj/mine", `Request: "Fix the parser".`)
	save("legacy", "2025-01-12", "", "legacy file without a project")
	save("mine-empty", "2025-01-13", "/proj/mine", "")
	save("other", "2025-01-15", "/proj/other", "other project work")

	h := handler.NewSessionContextHandler(handler.WithHomeDir(tmpHome))

	resp, err := h.Handle(context.Background(), &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           "/proj/mine",
	})
	require.NoError(t, err)
	assert.Equal(t, `Previous session in this project (2025-01-10): Request: "Fix the parser".`,
		resp.AdditionalContext(), "newest summarized session from the same project wins")

	resp, err = h.Handle(context.Background(), &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           "/proj/unknown",
	})
	require.NoError(t, err)
	assert.Empty(t, resp.AdditionalContext(), "no context from other projects or legacy files")
}

func TestSessionContextHandler_ImplementsHandler(t *testing.T) {
	t.Parallel()
	var _ handler.Handler = handler.NewSessionContextHandler()
}
