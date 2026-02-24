package instinct_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/instinct"
)

func newTestInstinct(id, domain string, confidence float64) instinct.Instinct {
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)

	return instinct.Instinct{
		ID:         id,
		Trigger:    "test trigger for " + id,
		Confidence: confidence,
		Domain:     domain,
		Source:     "observation",
		SourceRepo: "",
		Content:    "Test content for " + id,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func TestFileStore(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "save and get round-trip",
			fn:   testSaveAndGetRoundTrip,
		},
		{
			name: "list with no filters returns all",
			fn:   testListNoFilters,
		},
		{
			name: "list filtered by domain",
			fn:   testListFilteredByDomain,
		},
		{
			name: "list filtered by min confidence",
			fn:   testListFilteredByMinConfidence,
		},
		{
			name: "delete existing instinct",
			fn:   testDeleteExisting,
		},
		{
			name: "get nonexistent returns ErrNotFound",
			fn:   testGetNonexistent,
		},
		{
			name: "save creates directories",
			fn:   testSaveCreatesDirectories,
		},
		{
			name: "inherited instincts appear in list but not deletable from personal",
			fn:   testInheritedInstincts,
		},
		{
			name: "get rejects path traversal ID",
			fn:   testGetRejectsPathTraversal,
		},
		{
			name: "get rejects empty ID",
			fn:   testGetRejectsEmptyID,
		},
		{
			name: "delete rejects path traversal ID",
			fn:   testDeleteRejectsPathTraversal,
		},
		{
			name: "delete rejects empty ID",
			fn:   testDeleteRejectsEmptyID,
		},
		{
			name: "get with valid ID works after save",
			fn:   testGetValidIDAfterSave,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func testSaveAndGetRoundTrip(t *testing.T) {
	t.Helper()

	personalDir := t.TempDir()
	inheritedDir := t.TempDir()
	store := instinct.NewFileStore(personalDir, inheritedDir)

	inst := newTestInstinct("round-trip", "go", 0.7)

	require.NoError(t, store.Save(inst))

	got, err := store.Get("round-trip")
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, inst.ID, got.ID)
	assert.Equal(t, inst.Trigger, got.Trigger)
	assert.InDelta(t, inst.Confidence, got.Confidence, 0.001)
	assert.Equal(t, inst.Domain, got.Domain)
	assert.Equal(t, inst.Source, got.Source)
	assert.Equal(t, inst.SourceRepo, got.SourceRepo)
	assert.Equal(t, inst.Content, got.Content)
	assert.True(t, inst.CreatedAt.Equal(got.CreatedAt))
	assert.True(t, inst.UpdatedAt.Equal(got.UpdatedAt))
}

func testListNoFilters(t *testing.T) {
	t.Helper()

	personalDir := t.TempDir()
	inheritedDir := t.TempDir()
	store := instinct.NewFileStore(personalDir, inheritedDir)

	require.NoError(t, store.Save(newTestInstinct("alpha", "go", 0.5)))
	require.NoError(t, store.Save(newTestInstinct("beta", "python", 0.7)))

	got, err := store.List(instinct.ListOptions{
		Domain:        "",
		MinConfidence: 0,
		Source:        "",
	})
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, "alpha", got[0].ID)
	assert.Equal(t, "beta", got[1].ID)
}

func testListFilteredByDomain(t *testing.T) {
	t.Helper()

	personalDir := t.TempDir()
	inheritedDir := t.TempDir()
	store := instinct.NewFileStore(personalDir, inheritedDir)

	require.NoError(t, store.Save(newTestInstinct("go-inst", "go", 0.5)))
	require.NoError(t, store.Save(newTestInstinct("py-inst", "python", 0.7)))

	got, err := store.List(instinct.ListOptions{
		Domain:        "go",
		MinConfidence: 0,
		Source:        "",
	})
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "go-inst", got[0].ID)
}

func testListFilteredByMinConfidence(t *testing.T) {
	t.Helper()

	personalDir := t.TempDir()
	inheritedDir := t.TempDir()
	store := instinct.NewFileStore(personalDir, inheritedDir)

	require.NoError(t, store.Save(newTestInstinct("low", "go", 0.3)))
	require.NoError(t, store.Save(newTestInstinct("high", "go", 0.8)))

	got, err := store.List(instinct.ListOptions{
		Domain:        "",
		MinConfidence: 0.5,
		Source:        "",
	})
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "high", got[0].ID)
}

func testDeleteExisting(t *testing.T) {
	t.Helper()

	personalDir := t.TempDir()
	inheritedDir := t.TempDir()
	store := instinct.NewFileStore(personalDir, inheritedDir)

	inst := newTestInstinct("to-delete", "go", 0.5)
	require.NoError(t, store.Save(inst))

	require.NoError(t, store.Delete("to-delete"))

	_, err := store.Get("to-delete")
	assert.ErrorIs(t, err, instinct.ErrNotFound)
}

func testGetNonexistent(t *testing.T) {
	t.Helper()

	personalDir := t.TempDir()
	inheritedDir := t.TempDir()
	store := instinct.NewFileStore(personalDir, inheritedDir)

	_, err := store.Get("does-not-exist")
	assert.ErrorIs(t, err, instinct.ErrNotFound)
}

func testSaveCreatesDirectories(t *testing.T) {
	t.Helper()

	base := t.TempDir()
	personalDir := filepath.Join(base, "nested", "personal")
	inheritedDir := t.TempDir()
	store := instinct.NewFileStore(personalDir, inheritedDir)

	inst := newTestInstinct("create-dirs", "go", 0.5)
	require.NoError(t, store.Save(inst))

	info, err := os.Stat(personalDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	got, err := store.Get("create-dirs")
	require.NoError(t, err)
	assert.Equal(t, "create-dirs", got.ID)
}

func testInheritedInstincts(t *testing.T) {
	t.Helper()

	personalDir := t.TempDir()
	inheritedDir := t.TempDir()

	// Write an inherited instinct directly to the inherited directory.
	inheritedStore := instinct.NewFileStore(inheritedDir, "")
	require.NoError(t, inheritedStore.Save(newTestInstinct("inherited-inst", "go", 0.6)))

	store := instinct.NewFileStore(personalDir, inheritedDir)

	// Get should find inherited instinct.
	got, err := store.Get("inherited-inst")
	require.NoError(t, err)
	assert.Equal(t, "inherited-inst", got.ID)

	// List should include inherited instinct.
	all, err := store.List(instinct.ListOptions{
		Domain:        "",
		MinConfidence: 0,
		Source:        "",
	})
	require.NoError(t, err)
	require.Len(t, all, 1)
	assert.Equal(t, "inherited-inst", all[0].ID)

	// Delete from personal should fail since it only exists in inherited.
	err = store.Delete("inherited-inst")
	assert.Error(t, err)
}

func testGetRejectsPathTraversal(t *testing.T) {
	t.Helper()

	store := instinct.NewFileStore(t.TempDir(), "")

	_, err := store.Get("../traversal")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get instinct")
}

func testGetRejectsEmptyID(t *testing.T) {
	t.Helper()

	store := instinct.NewFileStore(t.TempDir(), "")

	_, err := store.Get("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get instinct")
}

func testDeleteRejectsPathTraversal(t *testing.T) {
	t.Helper()

	store := instinct.NewFileStore(t.TempDir(), "")

	err := store.Delete("../traversal")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path traversal")
}

func testDeleteRejectsEmptyID(t *testing.T) {
	t.Helper()

	store := instinct.NewFileStore(t.TempDir(), "")

	err := store.Delete("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must not be empty")
}

func testGetValidIDAfterSave(t *testing.T) {
	t.Helper()

	store := instinct.NewFileStore(t.TempDir(), "")
	inst := newTestInstinct("valid-id", "go", 0.8)

	require.NoError(t, store.Save(inst))

	got, err := store.Get("valid-id")
	require.NoError(t, err)
	assert.Equal(t, "valid-id", got.ID)
}
