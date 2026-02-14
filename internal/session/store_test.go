package session_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/session"
)

func TestStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	store := session.NewStore(dir)

	started := time.Date(2026, 2, 14, 10, 0, 0, 0, time.UTC)
	ended := time.Date(2026, 2, 14, 11, 0, 0, 0, time.UTC)

	sess := &session.Session{
		Version:       "1",
		ID:            "abc123",
		Date:          "2026-02-14",
		Started:       started,
		Ended:         ended,
		Title:         "Test session",
		Summary:       "A test summary",
		ToolsUsed:     []string{"Bash", "Edit"},
		FilesModified: []string{"main.go"},
		MessageCount:  5,
	}

	saveErr := store.Save(sess)
	require.NoError(t, saveErr)

	loaded, loadErr := store.Load("abc123")
	require.NoError(t, loadErr)

	assert.Equal(t, sess.Version, loaded.Version)
	assert.Equal(t, sess.ID, loaded.ID)
	assert.Equal(t, sess.Date, loaded.Date)
	assert.Equal(t, sess.Title, loaded.Title)
	assert.Equal(t, sess.Summary, loaded.Summary)
	assert.Equal(t, sess.ToolsUsed, loaded.ToolsUsed)
	assert.Equal(t, sess.FilesModified, loaded.FilesModified)
	assert.Equal(t, sess.MessageCount, loaded.MessageCount)
	assert.True(t, sess.Started.Equal(loaded.Started))
	assert.True(t, sess.Ended.Equal(loaded.Ended))
}

func TestStore_SaveSetsDefaultVersion(t *testing.T) {
	dir := t.TempDir()
	store := session.NewStore(dir)

	sess := &session.Session{
		Version:       "",
		ID:            "ver001",
		Date:          "2026-02-14",
		Started:       time.Date(2026, 2, 14, 10, 0, 0, 0, time.UTC),
		Ended:         time.Time{},
		Title:         "Version test",
		Summary:       "",
		ToolsUsed:     nil,
		FilesModified: nil,
		MessageCount:  0,
	}

	saveErr := store.Save(sess)
	require.NoError(t, saveErr)

	loaded, loadErr := store.Load("ver001")
	require.NoError(t, loadErr)
	assert.Equal(t, "1", loaded.Version)
}

func TestStore_LoadReturnsErrorForNonexistentID(t *testing.T) {
	dir := t.TempDir()
	store := session.NewStore(dir)

	_, err := store.Load("nonexistent")
	require.Error(t, err)
	assert.ErrorIs(t, err, session.ErrNotFound)
}

func TestStore_LoadReturnsErrorForEmptyID(t *testing.T) {
	dir := t.TempDir()
	store := session.NewStore(dir)

	_, err := store.Load("")
	require.Error(t, err)
	assert.ErrorIs(t, err, session.ErrEmptyID)
}

func TestStore_SaveReturnsErrorForEmptyID(t *testing.T) {
	dir := t.TempDir()
	store := session.NewStore(dir)

	sess := &session.Session{
		Version:       "1",
		ID:            "",
		Date:          "2026-02-14",
		Started:       time.Date(2026, 2, 14, 10, 0, 0, 0, time.UTC),
		Ended:         time.Time{},
		Title:         "No ID",
		Summary:       "",
		ToolsUsed:     nil,
		FilesModified: nil,
		MessageCount:  0,
	}

	err := store.Save(sess)
	require.Error(t, err)
	assert.ErrorIs(t, err, session.ErrEmptyID)
}

func TestStore_List(t *testing.T) {
	dir := t.TempDir()
	store := session.NewStore(dir)

	sessions := []session.Session{
		{
			Version:       "1",
			ID:            "aaa",
			Date:          "2026-02-12",
			Started:       time.Date(2026, 2, 12, 10, 0, 0, 0, time.UTC),
			Ended:         time.Time{},
			Title:         "First",
			Summary:       "",
			ToolsUsed:     nil,
			FilesModified: nil,
			MessageCount:  0,
		},
		{
			Version:       "1",
			ID:            "bbb",
			Date:          "2026-02-13",
			Started:       time.Date(2026, 2, 13, 10, 0, 0, 0, time.UTC),
			Ended:         time.Time{},
			Title:         "Second",
			Summary:       "",
			ToolsUsed:     nil,
			FilesModified: nil,
			MessageCount:  0,
		},
		{
			Version:       "1",
			ID:            "ccc",
			Date:          "2026-02-14",
			Started:       time.Date(2026, 2, 14, 10, 0, 0, 0, time.UTC),
			Ended:         time.Time{},
			Title:         "Third",
			Summary:       "",
			ToolsUsed:     nil,
			FilesModified: nil,
			MessageCount:  0,
		},
	}

	for i := range sessions {
		require.NoError(t, store.Save(&sessions[i]))
	}

	listed, listErr := store.List(2)
	require.NoError(t, listErr)
	require.Len(t, listed, 2)

	// Most recent first.
	assert.Equal(t, "ccc", listed[0].ID)
	assert.Equal(t, "bbb", listed[1].ID)
}

func TestStore_ListReturnsAllWhenLimitIsZero(t *testing.T) {
	dir := t.TempDir()
	store := session.NewStore(dir)

	for _, id := range []string{"x1", "x2", "x3"} {
		sess := &session.Session{
			Version:       "1",
			ID:            id,
			Date:          "2026-02-14",
			Started:       time.Date(2026, 2, 14, 10, 0, 0, 0, time.UTC),
			Ended:         time.Time{},
			Title:         "Session " + id,
			Summary:       "",
			ToolsUsed:     nil,
			FilesModified: nil,
			MessageCount:  0,
		}
		require.NoError(t, store.Save(sess))
	}

	listed, listErr := store.List(0)
	require.NoError(t, listErr)
	assert.Len(t, listed, 3)
}

func TestStore_FindByDate(t *testing.T) {
	dir := t.TempDir()
	store := session.NewStore(dir)

	sessions := []session.Session{
		{
			Version:       "1",
			ID:            "d1",
			Date:          "2026-02-13",
			Started:       time.Date(2026, 2, 13, 10, 0, 0, 0, time.UTC),
			Ended:         time.Time{},
			Title:         "Yesterday",
			Summary:       "",
			ToolsUsed:     nil,
			FilesModified: nil,
			MessageCount:  0,
		},
		{
			Version:       "1",
			ID:            "d2",
			Date:          "2026-02-14",
			Started:       time.Date(2026, 2, 14, 9, 0, 0, 0, time.UTC),
			Ended:         time.Time{},
			Title:         "Today morning",
			Summary:       "",
			ToolsUsed:     nil,
			FilesModified: nil,
			MessageCount:  0,
		},
		{
			Version:       "1",
			ID:            "d3",
			Date:          "2026-02-14",
			Started:       time.Date(2026, 2, 14, 15, 0, 0, 0, time.UTC),
			Ended:         time.Time{},
			Title:         "Today afternoon",
			Summary:       "",
			ToolsUsed:     nil,
			FilesModified: nil,
			MessageCount:  0,
		},
	}

	for i := range sessions {
		require.NoError(t, store.Save(&sessions[i]))
	}

	found, findErr := store.FindByDate("2026-02-14")
	require.NoError(t, findErr)
	require.Len(t, found, 2)

	assert.Equal(t, "d2", found[0].ID)
	assert.Equal(t, "d3", found[1].ID)
}

func TestStore_FindByDateWithMonthPrefix(t *testing.T) {
	dir := t.TempDir()
	store := session.NewStore(dir)

	sess := &session.Session{
		Version:       "1",
		ID:            "m1",
		Date:          "2026-02-14",
		Started:       time.Date(2026, 2, 14, 10, 0, 0, 0, time.UTC),
		Ended:         time.Time{},
		Title:         "February session",
		Summary:       "",
		ToolsUsed:     nil,
		FilesModified: nil,
		MessageCount:  0,
	}
	require.NoError(t, store.Save(sess))

	found, findErr := store.FindByDate("2026-02")
	require.NoError(t, findErr)
	assert.Len(t, found, 1)
}

func TestStore_Search(t *testing.T) {
	dir := t.TempDir()
	store := session.NewStore(dir)

	sessions := []session.Session{
		{
			Version:       "1",
			ID:            "s1",
			Date:          "2026-02-14",
			Started:       time.Date(2026, 2, 14, 10, 0, 0, 0, time.UTC),
			Ended:         time.Time{},
			Title:         "Refactoring hooks",
			Summary:       "Consolidated hook logic",
			ToolsUsed:     nil,
			FilesModified: nil,
			MessageCount:  0,
		},
		{
			Version:       "1",
			ID:            "s2",
			Date:          "2026-02-14",
			Started:       time.Date(2026, 2, 14, 11, 0, 0, 0, time.UTC),
			Ended:         time.Time{},
			Title:         "Bug fix in config",
			Summary:       "Fixed JSON parsing",
			ToolsUsed:     nil,
			FilesModified: nil,
			MessageCount:  0,
		},
		{
			Version:       "1",
			ID:            "s3",
			Date:          "2026-02-14",
			Started:       time.Date(2026, 2, 14, 12, 0, 0, 0, time.UTC),
			Ended:         time.Time{},
			Title:         "Add tests",
			Summary:       "Added hooks tests",
			ToolsUsed:     nil,
			FilesModified: nil,
			MessageCount:  0,
		},
	}

	for i := range sessions {
		require.NoError(t, store.Save(&sessions[i]))
	}

	// Search by title.
	found, searchErr := store.Search("hooks")
	require.NoError(t, searchErr)
	require.Len(t, found, 2)
	assert.Equal(t, "s1", found[0].ID)
	assert.Equal(t, "s3", found[1].ID)
}

func TestStore_SearchIsCaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	store := session.NewStore(dir)

	sess := &session.Session{
		Version:       "1",
		ID:            "ci1",
		Date:          "2026-02-14",
		Started:       time.Date(2026, 2, 14, 10, 0, 0, 0, time.UTC),
		Ended:         time.Time{},
		Title:         "UPPERCASE Title",
		Summary:       "",
		ToolsUsed:     nil,
		FilesModified: nil,
		MessageCount:  0,
	}
	require.NoError(t, store.Save(sess))

	found, searchErr := store.Search("uppercase")
	require.NoError(t, searchErr)
	assert.Len(t, found, 1)
}

func TestStore_SaveCreatesDirectoryIfNotExists(t *testing.T) {
	nested := filepath.Join(t.TempDir(), "nested", "sessions")
	store := session.NewStore(nested)

	sess := &session.Session{
		Version:       "1",
		ID:            "dir1",
		Date:          "2026-02-14",
		Started:       time.Date(2026, 2, 14, 10, 0, 0, 0, time.UTC),
		Ended:         time.Time{},
		Title:         "Nested dir test",
		Summary:       "",
		ToolsUsed:     nil,
		FilesModified: nil,
		MessageCount:  0,
	}

	saveErr := store.Save(sess)
	require.NoError(t, saveErr)

	// Verify the directory was created and file exists.
	info, statErr := os.Stat(nested)
	require.NoError(t, statErr)
	assert.True(t, info.IsDir())

	loaded, loadErr := store.Load("dir1")
	require.NoError(t, loadErr)
	assert.Equal(t, "dir1", loaded.ID)
}

func TestStore_ListEmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	store := session.NewStore(dir)

	listed, listErr := store.List(10)
	require.NoError(t, listErr)
	assert.Empty(t, listed)
}
