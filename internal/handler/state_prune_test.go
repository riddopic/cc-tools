package handler_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
)

const pruneDay = 24 * time.Hour

// fixedPruneNow returns the reference clock for prune tests.
func fixedPruneNow() time.Time { return time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC) }

// writeAgedFile creates path (and its parent) with an mtime of now minus age.
func writeAgedFile(t *testing.T, path string, age time.Duration) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, []byte("x"), 0o600))
	mtime := fixedPruneNow().Add(-age)
	require.NoError(t, os.Chtimes(path, mtime, mtime))
}

func newPruneHandler(cacheDir string, cfg *config.Values) *handler.StatePruneHandler {
	return handler.NewStatePruneHandler(cfg,
		handler.WithPruneCacheDir(cacheDir),
		handler.WithPruneNow(fixedPruneNow),
	)
}

func runPrune(t *testing.T, h *handler.StatePruneHandler, sessionID string) {
	t.Helper()
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		SessionID:     hookcmd.SessionID(sessionID),
	}
	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
	assert.Nil(t, resp.Stdout)
	assert.Empty(t, resp.Stderr)
}

func TestStatePruneHandler_Name(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "state-prune", handler.NewStatePruneHandler(nil).Name())
}

func TestStatePruneHandler_RemovesOldKeepsFresh(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	oldFiles := []string{
		filepath.Join(dir, "drift", "drift-old.json"),
		filepath.Join(dir, "stop", "stop-old.count"),
		filepath.Join(dir, "compact", "cc-tools-compact-old.count"),
		filepath.Join(dir, "compact", "cc-tools-compact-older.json"),
	}
	freshFiles := []string{
		filepath.Join(dir, "drift", "drift-fresh.json"),
		filepath.Join(dir, "stop", "stop-fresh.count"),
		filepath.Join(dir, "compact", "cc-tools-compact-fresh.count"),
	}
	for _, f := range oldFiles {
		writeAgedFile(t, f, 8*pruneDay)
	}
	for _, f := range freshFiles {
		writeAgedFile(t, f, 6*pruneDay)
	}

	runPrune(t, newPruneHandler(dir, nil), "current")

	for _, f := range oldFiles {
		assert.NoFileExists(t, f)
	}
	for _, f := range freshFiles {
		assert.FileExists(t, f)
	}

	marker, err := os.Stat(filepath.Join(dir, ".last-prune"))
	require.NoError(t, err, "marker should be written after a prune")
	assert.True(t, marker.ModTime().Equal(fixedPruneNow()))
}

func TestStatePruneHandler_KeepsCurrentSession(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	const sessionID = "abc-123"

	current := []string{
		filepath.Join(dir, "drift", "drift-"+sessionID+".json"),
		filepath.Join(dir, "stop", "stop-"+sessionID+".count"),
		filepath.Join(dir, "compact", "cc-tools-compact-"+sessionID+".count"),
	}
	for _, f := range current {
		writeAgedFile(t, f, 30*pruneDay)
	}
	other := filepath.Join(dir, "stop", "stop-other.count")
	writeAgedFile(t, other, 30*pruneDay)

	runPrune(t, newPruneHandler(dir, nil), sessionID)

	for _, f := range current {
		assert.FileExists(t, f)
	}
	assert.NoFileExists(t, other)
}

func TestStatePruneHandler_KeepsCurrentHashedSession(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sessionID := hookcmd.SessionID("not/safe id")

	current := filepath.Join(dir, "drift", "drift-"+sessionID.FileKey()+".json")
	writeAgedFile(t, current, 30*pruneDay)

	runPrune(t, newPruneHandler(dir, nil), string(sessionID))

	assert.FileExists(t, current)
}

func TestStatePruneHandler_LeavesForeignFilesAlone(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	foreign := []string{
		filepath.Join(dir, "drift", "notes.json"),
		filepath.Join(dir, "drift", "drift-evals.jsonl.bak", "drift-nested.json"),
		filepath.Join(dir, "stop", "drift-misplaced.json"),
		filepath.Join(dir, "compact", "compact-old.count"),
		filepath.Join(dir, "observations", "observations.jsonl"),
		filepath.Join(dir, "other", "drift-old.json"),
		filepath.Join(dir, "drift-old.json"),
		filepath.Join(dir, "stop-old.count"),
	}
	for _, f := range foreign {
		writeAgedFile(t, f, 365*pruneDay)
	}
	staleDir := filepath.Join(dir, "stop", "stop-dir")
	require.NoError(t, os.MkdirAll(staleDir, 0o700))
	require.NoError(t, os.Chtimes(staleDir, fixedPruneNow().Add(-365*pruneDay), fixedPruneNow().Add(-365*pruneDay)))

	runPrune(t, newPruneHandler(dir, nil), "current")

	for _, f := range foreign {
		assert.FileExists(t, f)
	}
	assert.DirExists(t, staleDir)
}

func TestStatePruneHandler_DoesNotFollowSymlinks(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	outside := t.TempDir()

	target := filepath.Join(outside, "drift-target.json")
	writeAgedFile(t, target, 365*pruneDay)
	link := filepath.Join(dir, "drift", "drift-link.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(link), 0o700))
	require.NoError(t, os.Symlink(target, link))

	outsideStop := filepath.Join(outside, "stop")
	writeAgedFile(t, filepath.Join(outsideStop, "stop-old.count"), 365*pruneDay)
	require.NoError(t, os.Symlink(outsideStop, filepath.Join(dir, "stop")))

	runPrune(t, newPruneHandler(dir, nil), "current")

	assert.FileExists(t, target)
	_, err := os.Lstat(link)
	require.NoError(t, err, "symlink itself should be left in place")
	assert.FileExists(t, filepath.Join(outsideStop, "stop-old.count"))
}

func TestStatePruneHandler_RecentMarkerSkipsPrune(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	old := filepath.Join(dir, "drift", "drift-old.json")
	writeAgedFile(t, old, 30*pruneDay)
	marker := filepath.Join(dir, ".last-prune")
	writeAgedFile(t, marker, 23*time.Hour)

	runPrune(t, newPruneHandler(dir, nil), "current")

	assert.FileExists(t, old)
	info, err := os.Stat(marker)
	require.NoError(t, err)
	assert.True(t, info.ModTime().Equal(fixedPruneNow().Add(-23*time.Hour)), "marker should be untouched")
}

func TestStatePruneHandler_StaleMarkerAllowsPrune(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	old := filepath.Join(dir, "drift", "drift-old.json")
	writeAgedFile(t, old, 30*pruneDay)
	marker := filepath.Join(dir, ".last-prune")
	writeAgedFile(t, marker, 25*time.Hour)

	runPrune(t, newPruneHandler(dir, nil), "current")

	assert.NoFileExists(t, old)
	info, err := os.Stat(marker)
	require.NoError(t, err)
	assert.True(t, info.ModTime().Equal(fixedPruneNow()))
}

func TestStatePruneHandler_HonorsMaxAgeDays(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	cfg := newTestConfig()
	cfg.State.MaxAgeDays = 2

	threeDays := filepath.Join(dir, "stop", "stop-three.count")
	oneDay := filepath.Join(dir, "stop", "stop-one.count")
	writeAgedFile(t, threeDays, 3*pruneDay)
	writeAgedFile(t, oneDay, 1*pruneDay)

	runPrune(t, newPruneHandler(dir, cfg), "current")

	assert.NoFileExists(t, threeDays)
	assert.FileExists(t, oneDay)
}

func TestStatePruneHandler_MissingCacheDir(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "does-not-exist")

	runPrune(t, newPruneHandler(dir, nil), "current")

	assert.NoDirExists(t, dir, "prune must not create the cache directory")
}
