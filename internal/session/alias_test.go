package session_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/session"
)

func TestAliasManager_SetAndResolve(t *testing.T) {
	tests := []struct {
		name      string
		alias     string
		sessionID string
	}{
		{
			name:      "simple alias",
			alias:     "mywork",
			sessionID: "abc123",
		},
		{
			name:      "alias with hyphens",
			alias:     "bug-fix-session",
			sessionID: "def456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "aliases.json")
			am := session.NewAliasManager(path)

			setErr := am.Set(tt.alias, tt.sessionID)
			require.NoError(t, setErr)

			resolved, resolveErr := am.Resolve(tt.alias)
			require.NoError(t, resolveErr)
			assert.Equal(t, tt.sessionID, resolved)
		})
	}
}

func TestAliasManager_ResolveReturnsErrorForUnknownAlias(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aliases.json")
	am := session.NewAliasManager(path)

	_, err := am.Resolve("nonexistent")
	require.Error(t, err)
	assert.ErrorIs(t, err, session.ErrAliasNotFound)
}

func TestAliasManager_RemoveDeletesAlias(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aliases.json")
	am := session.NewAliasManager(path)

	require.NoError(t, am.Set("temp", "sess1"))

	removeErr := am.Remove("temp")
	require.NoError(t, removeErr)

	_, resolveErr := am.Resolve("temp")
	require.Error(t, resolveErr)
	assert.ErrorIs(t, resolveErr, session.ErrAliasNotFound)
}

func TestAliasManager_RemoveReturnsErrorForUnknownAlias(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aliases.json")
	am := session.NewAliasManager(path)

	err := am.Remove("nonexistent")
	require.Error(t, err)
	assert.ErrorIs(t, err, session.ErrAliasNotFound)
}

func TestAliasManager_ListReturnsAllAliases(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aliases.json")
	am := session.NewAliasManager(path)

	require.NoError(t, am.Set("work", "sess1"))
	require.NoError(t, am.Set("debug", "sess2"))

	aliases, listErr := am.List()
	require.NoError(t, listErr)
	assert.Len(t, aliases, 2)
	assert.Equal(t, "sess1", aliases["work"])
	assert.Equal(t, "sess2", aliases["debug"])
}

func TestAliasManager_ListReturnsEmptyMapWhenNoAliases(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aliases.json")
	am := session.NewAliasManager(path)

	aliases, listErr := am.List()
	require.NoError(t, listErr)
	assert.Empty(t, aliases)
}

func TestAliasManager_SetOverwritesExistingAlias(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aliases.json")
	am := session.NewAliasManager(path)

	require.NoError(t, am.Set("latest", "sess1"))
	require.NoError(t, am.Set("latest", "sess2"))

	resolved, resolveErr := am.Resolve("latest")
	require.NoError(t, resolveErr)
	assert.Equal(t, "sess2", resolved)
}

func TestAliasManager_CreatesDirectoryIfNotExists(t *testing.T) {
	nested := filepath.Join(t.TempDir(), "nested", "dir", "aliases.json")
	am := session.NewAliasManager(nested)

	setErr := am.Set("test", "sess1")
	require.NoError(t, setErr)

	resolved, resolveErr := am.Resolve("test")
	require.NoError(t, resolveErr)
	assert.Equal(t, "sess1", resolved)
}
