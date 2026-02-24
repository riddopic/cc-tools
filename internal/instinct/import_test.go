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

func TestClassifyImport(t *testing.T) {
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)

	makeInst := func(id string, conf float64) instinct.Instinct {
		return instinct.Instinct{
			ID:         id,
			Trigger:    "test trigger",
			Confidence: conf,
			Domain:     "go",
			Source:     "observation",
			SourceRepo: "",
			Content:    "",
			CreatedAt:  now,
			UpdatedAt:  now,
		}
	}

	tests := []struct {
		name    string
		setup   func(t *testing.T, store *instinct.FileStore)
		inst    instinct.Instinct
		force   bool
		minConf float64
		want    instinct.ImportAction
	}{
		{
			name:  "new instinct",
			setup: func(_ *testing.T, _ *instinct.FileStore) {},
			inst:  makeInst("brand-new", 0.8),
			force: false, minConf: 0,
			want: instinct.ImportNew,
		},
		{
			name: "existing without force skips",
			setup: func(t *testing.T, store *instinct.FileStore) {
				t.Helper()
				require.NoError(t, store.Save(makeInst("exists", 0.5)))
			},
			inst:  makeInst("exists", 0.9),
			force: false, minConf: 0,
			want: instinct.ImportSkip,
		},
		{
			name: "existing with force overwrites",
			setup: func(t *testing.T, store *instinct.FileStore) {
				t.Helper()
				require.NoError(t, store.Save(makeInst("exists", 0.5)))
			},
			inst:  makeInst("exists", 0.9),
			force: true, minConf: 0,
			want: instinct.ImportOverwrite,
		},
		{
			name:  "below min confidence skips",
			setup: func(_ *testing.T, _ *instinct.FileStore) {},
			inst:  makeInst("low-conf", 0.3),
			force: false, minConf: 0.5,
			want: instinct.ImportSkip,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			store := instinct.NewFileStore(dir, "")
			tt.setup(t, store)

			got := instinct.ClassifyImport(store, tt.inst, tt.force, tt.minConf)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestImport(t *testing.T) {
	t.Run("happy path imports new and skips existing", func(t *testing.T) {
		readDir := t.TempDir()
		writeDir := t.TempDir()
		readStore := instinct.NewFileStore(readDir, "")
		writeStore := instinct.NewFileStore(writeDir, "")

		require.NoError(t, readStore.Save(newTestInstinct("existing", "go", 0.5)))

		input := []instinct.Instinct{
			newTestInstinct("new-one", "go", 0.7),
			newTestInstinct("existing", "go", 0.9),
			newTestInstinct("new-two", "go", 0.8),
		}

		opts := instinct.ImportOptions{
			DryRun:        false,
			Force:         false,
			MinConfidence: 0,
		}
		result, err := instinct.Import(readStore, writeStore, input, opts)
		require.NoError(t, err)
		assert.Equal(t, 2, result.Imported())

		got, err := writeStore.Get("new-one")
		require.NoError(t, err)
		assert.Equal(t, "new-one", got.ID)
	})

	t.Run("dry run does not save", func(t *testing.T) {
		readDir := t.TempDir()
		writeDir := t.TempDir()
		readStore := instinct.NewFileStore(readDir, "")
		writeStore := instinct.NewFileStore(writeDir, "")

		input := []instinct.Instinct{newTestInstinct("dry-inst", "go", 0.8)}
		opts := instinct.ImportOptions{
			DryRun:        true,
			Force:         false,
			MinConfidence: 0,
		}
		result, err := instinct.Import(readStore, writeStore, input, opts)
		require.NoError(t, err)
		assert.Equal(t, 1, result.Imported())

		_, err = writeStore.Get("dry-inst")
		require.Error(t, err)
	})

	t.Run("force overwrites existing", func(t *testing.T) {
		readDir := t.TempDir()
		writeDir := t.TempDir()
		readStore := instinct.NewFileStore(readDir, "")
		writeStore := instinct.NewFileStore(writeDir, "")

		require.NoError(t, readStore.Save(newTestInstinct("overwrite-me", "go", 0.4)))

		input := []instinct.Instinct{newTestInstinct("overwrite-me", "go", 0.9)}
		opts := instinct.ImportOptions{
			DryRun:        false,
			Force:         true,
			MinConfidence: 0,
		}
		result, err := instinct.Import(readStore, writeStore, input, opts)
		require.NoError(t, err)
		assert.Equal(t, 1, result.Imported())

		got, err := writeStore.Get("overwrite-me")
		require.NoError(t, err)
		assert.InDelta(t, 0.9, got.Confidence, 0.001)
	})

	t.Run("save error propagation", func(t *testing.T) {
		readDir := t.TempDir()
		writeStore := instinct.NewFileStore("/dev/null/impossible", "")
		readStore := instinct.NewFileStore(readDir, "")

		input := []instinct.Instinct{newTestInstinct("fail-save", "go", 0.8)}
		opts := instinct.ImportOptions{
			DryRun:        false,
			Force:         false,
			MinConfidence: 0,
		}
		_, err := instinct.Import(readStore, writeStore, input, opts)
		require.Error(t, err)
	})
}

func TestImportResult_Imported(t *testing.T) {
	empty := instinct.Instinct{
		ID:         "",
		Trigger:    "",
		Confidence: 0,
		Domain:     "",
		Source:     "",
		SourceRepo: "",
		Content:    "",
		CreatedAt:  time.Time{},
		UpdatedAt:  time.Time{},
	}

	tests := []struct {
		name  string
		items []instinct.ImportItem
		want  int
	}{
		{
			name: "counts non-skip items",
			items: []instinct.ImportItem{
				{Instinct: empty, Action: instinct.ImportNew},
				{Instinct: empty, Action: instinct.ImportSkip},
				{Instinct: empty, Action: instinct.ImportOverwrite},
				{Instinct: empty, Action: instinct.ImportSkip},
			},
			want: 2,
		},
		{
			name:  "empty returns zero",
			items: nil,
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &instinct.ImportResult{Items: tt.items}
			assert.Equal(t, tt.want, r.Imported())
		})
	}
}

func TestImportAction_Label(t *testing.T) {
	tests := []struct {
		name   string
		action instinct.ImportAction
		dryRun bool
		want   string
	}{
		{"new", instinct.ImportNew, false, "import:"},
		{"new dry-run", instinct.ImportNew, true, "[dry-run] import:"},
		{"overwrite", instinct.ImportOverwrite, false, "overwrite:"},
		{"overwrite dry-run", instinct.ImportOverwrite, true, "[dry-run] overwrite:"},
		{"skip", instinct.ImportSkip, false, "skip:"},
		{"skip dry-run", instinct.ImportSkip, true, "[dry-run] skip:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.action.Label(tt.dryRun))
		})
	}
}

func TestReadAndParseSource(t *testing.T) {
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)

	t.Run("valid file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "instincts.yaml")
		content := "---\nid: test-inst\ntrigger: test trigger\n" +
			"confidence: 0.8\ndomain: go\nsource: observation\n" +
			"created_at: " + now.Format(time.RFC3339) +
			"\nupdated_at: " + now.Format(time.RFC3339) + "\n---\n"
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

		got, err := instinct.ReadAndParseSource(path)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "test-inst", got[0].ID)
	})

	t.Run("directory traversal rejected", func(t *testing.T) {
		_, err := instinct.ReadAndParseSource("../../../etc/passwd")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "traversal")
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := instinct.ReadAndParseSource("/nonexistent/path/file.yaml")
		require.Error(t, err)
	})

	t.Run("malformed content", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "bad.yaml")
		content := "---\nconfidence: not-a-number\n---\n"
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

		got, err := instinct.ReadAndParseSource(path)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
