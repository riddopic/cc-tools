//go:build testmode

package main

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/session"
)

func newTestSessionStore(t *testing.T) *session.Store {
	t.Helper()
	return session.NewStore(filepath.Join(t.TempDir(), "sessions"))
}

func newTestAliasManager(t *testing.T) *session.AliasManager {
	t.Helper()
	return session.NewAliasManager(filepath.Join(t.TempDir(), "aliases.json"))
}

func seedSession(t *testing.T, store *session.Store, id, date, title string) {
	t.Helper()
	sess := &session.Session{
		Version: "1",
		ID:      id,
		Date:    date,
		Started: time.Now(),
		Title:   title,
	}
	require.NoError(t, store.Save(sess))
}

func TestListSessions(t *testing.T) {
	t.Run("empty store", func(t *testing.T) {
		store := newTestSessionStore(t)
		var buf bytes.Buffer

		err := listSessions(&buf, store, defaultSessionLimit)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "No sessions found.")
	})

	t.Run("populated store", func(t *testing.T) {
		store := newTestSessionStore(t)
		seedSession(t, store, "abc123", "2026-02-20", "Refactor auth module")
		seedSession(t, store, "def456", "2026-02-21", "Add session tracking")

		var buf bytes.Buffer
		err := listSessions(&buf, store, defaultSessionLimit)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "DATE")
		assert.Contains(t, output, "ID")
		assert.Contains(t, output, "TITLE")
		assert.Contains(t, output, "abc123")
		assert.Contains(t, output, "def456")
		assert.Contains(t, output, "Refactor auth module")
	})

	t.Run("respects limit", func(t *testing.T) {
		store := newTestSessionStore(t)
		seedSession(t, store, "s1", "2026-02-01", "First")
		seedSession(t, store, "s2", "2026-02-02", "Second")
		seedSession(t, store, "s3", "2026-02-03", "Third")

		var buf bytes.Buffer
		err := listSessions(&buf, store, 2)
		require.NoError(t, err)

		output := buf.String()
		// Most recent two should appear; oldest should not.
		assert.Contains(t, output, "s3")
		assert.Contains(t, output, "s2")
		assert.NotContains(t, output, "First")
	})
}

func TestShowSessionInfo(t *testing.T) {
	t.Run("session found", func(t *testing.T) {
		store := newTestSessionStore(t)
		aliases := newTestAliasManager(t)
		seedSession(t, store, "abc123", "2026-02-20", "Test session")

		var buf bytes.Buffer
		err := showSessionInfo(&buf, store, aliases, "abc123")
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "abc123")
		assert.Contains(t, buf.String(), "Test session")
	})

	t.Run("session not found", func(t *testing.T) {
		store := newTestSessionStore(t)
		aliases := newTestAliasManager(t)

		var buf bytes.Buffer
		err := showSessionInfo(&buf, store, aliases, "nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})

	t.Run("alias resolution", func(t *testing.T) {
		store := newTestSessionStore(t)
		aliases := newTestAliasManager(t)
		seedSession(t, store, "abc123", "2026-02-20", "Aliased session")
		require.NoError(t, aliases.Set("mywork", "abc123"))

		var buf bytes.Buffer
		err := showSessionInfo(&buf, store, aliases, "mywork")
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "abc123")
		assert.Contains(t, buf.String(), "Aliased session")
	})
}

func TestSetSessionAlias(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		aliases := newTestAliasManager(t)
		var buf bytes.Buffer

		err := setSessionAlias(&buf, aliases, "mywork", "abc123")
		require.NoError(t, err)
		assert.Contains(t, buf.String(), `"mywork"`)
		assert.Contains(t, buf.String(), "abc123")

		// Verify persistence.
		resolved, resolveErr := aliases.Resolve("mywork")
		require.NoError(t, resolveErr)
		assert.Equal(t, "abc123", resolved)
	})
}

func TestRemoveSessionAlias(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		aliases := newTestAliasManager(t)
		require.NoError(t, aliases.Set("mywork", "abc123"))

		var buf bytes.Buffer
		err := removeSessionAlias(&buf, aliases, "mywork")
		require.NoError(t, err)
		assert.Contains(t, buf.String(), `"mywork"`)
		assert.Contains(t, buf.String(), "removed")
	})

	t.Run("not found", func(t *testing.T) {
		aliases := newTestAliasManager(t)
		var buf bytes.Buffer

		err := removeSessionAlias(&buf, aliases, "nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "remove alias")
	})
}

func TestListSessionAliases(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		aliases := newTestAliasManager(t)
		var buf bytes.Buffer

		err := listSessionAliases(&buf, aliases)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "No aliases defined.")
	})

	t.Run("populated", func(t *testing.T) {
		aliases := newTestAliasManager(t)
		require.NoError(t, aliases.Set("work", "abc123"))
		require.NoError(t, aliases.Set("hobby", "def456"))

		var buf bytes.Buffer
		err := listSessionAliases(&buf, aliases)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "ALIAS")
		assert.Contains(t, output, "SESSION ID")
		assert.Contains(t, output, "work")
		assert.Contains(t, output, "abc123")
		assert.Contains(t, output, "hobby")
		assert.Contains(t, output, "def456")
	})
}

func TestSearchSessions(t *testing.T) {
	t.Run("no matches", func(t *testing.T) {
		store := newTestSessionStore(t)
		seedSession(t, store, "abc123", "2026-02-20", "Build auth module")

		var buf bytes.Buffer
		err := searchSessions(&buf, store, "nonexistent")
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "No matching sessions found.")
	})

	t.Run("matches found", func(t *testing.T) {
		store := newTestSessionStore(t)
		seedSession(t, store, "abc123", "2026-02-20", "Build auth module")
		seedSession(t, store, "def456", "2026-02-21", "Fix auth bug")
		seedSession(t, store, "ghi789", "2026-02-22", "Add logging")

		var buf bytes.Buffer
		err := searchSessions(&buf, store, "auth")
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "abc123")
		assert.Contains(t, output, "def456")
		assert.NotContains(t, output, "ghi789")
	})
}

// Command-execution tests exercise the Cobra RunE wrappers to cover
// the os.UserHomeDir → Store/AliasManager → handler delegation path.

func setupSessionHome(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	return tmpDir
}

func TestSessionListCmd(t *testing.T) {
	homeDir := setupSessionHome(t)

	// Seed a session file so there's something to list.
	store := session.NewStore(filepath.Join(homeDir, ".claude", "sessions"))
	seedSession(t, store, "cmd-test-1", "2026-02-23", "Command test")

	cmd := newSessionListCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.RunE(cmd, nil)
	require.NoError(t, err)
}

func TestSessionInfoCmd(t *testing.T) {
	homeDir := setupSessionHome(t)

	store := session.NewStore(filepath.Join(homeDir, ".claude", "sessions"))
	seedSession(t, store, "info-test-1", "2026-02-23", "Info test")

	cmd := newSessionInfoCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.RunE(cmd, []string{"info-test-1"})
	require.NoError(t, err)
}

func TestSessionAliasSetCmd(t *testing.T) {
	setupSessionHome(t)

	cmd := newSessionAliasSetCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.RunE(cmd, []string{"myalias", "some-session-id"})
	require.NoError(t, err)
}

func TestSessionAliasRemoveCmd(t *testing.T) {
	homeDir := setupSessionHome(t)

	// Create an alias first so removal succeeds.
	aliases := session.NewAliasManager(filepath.Join(homeDir, ".claude", "session-aliases.json"))
	require.NoError(t, aliases.Set("removeme", "some-id"))

	cmd := newSessionAliasRemoveCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.RunE(cmd, []string{"removeme"})
	require.NoError(t, err)
}

func TestSessionAliasListCmd(t *testing.T) {
	setupSessionHome(t)

	cmd := newSessionAliasListCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.RunE(cmd, nil)
	require.NoError(t, err)
}

func TestSessionSearchCmd(t *testing.T) {
	homeDir := setupSessionHome(t)

	store := session.NewStore(filepath.Join(homeDir, ".claude", "sessions"))
	seedSession(t, store, "search-test", "2026-02-23", "Searchable title")

	cmd := newSessionSearchCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.RunE(cmd, []string{"Searchable"})
	require.NoError(t, err)
}
