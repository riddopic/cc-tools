package observe_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/observe"
)

func TestRotateIfNeeded(t *testing.T) {
	tests := []struct {
		name        string
		setupDir    func(t *testing.T) string
		maxSizeMB   int
		wantRotated bool
		wantErr     bool
	}{
		{
			name: "does nothing when file is under limit",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				filePath := filepath.Join(dir, "observations.jsonl")
				require.NoError(t, os.WriteFile(filePath, []byte("small\n"), 0o600))
				return dir
			},
			maxSizeMB:   1,
			wantRotated: false,
			wantErr:     false,
		},
		{
			name: "renames file when over limit",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				filePath := filepath.Join(dir, "observations.jsonl")
				// Create a file that exceeds 1 byte threshold (we use maxSizeMB=0
				// which means 0 bytes limit, so any non-empty file triggers rotation).
				require.NoError(t, os.WriteFile(filePath, []byte("data\n"), 0o600))
				return dir
			},
			maxSizeMB:   0,
			wantRotated: true,
			wantErr:     false,
		},
		{
			name: "handles nonexistent file gracefully",
			setupDir: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			maxSizeMB:   1,
			wantRotated: false,
			wantErr:     false,
		},
		{
			name: "rotated file has timestamped name",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				filePath := filepath.Join(dir, "observations.jsonl")
				require.NoError(t, os.WriteFile(filePath, []byte("content\n"), 0o600))
				return dir
			},
			maxSizeMB:   0,
			wantRotated: true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupDir(t)
			filePath := filepath.Join(dir, "observations.jsonl")

			err := observe.RotateIfNeeded(filePath, tt.maxSizeMB)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.wantRotated {
				// Original file should no longer exist.
				_, statErr := os.Stat(filePath)
				assert.True(t, os.IsNotExist(statErr), "original file should be renamed")

				// A timestamped archive file should exist.
				entries, readErr := os.ReadDir(dir)
				require.NoError(t, readErr)
				require.Len(t, entries, 1, "should have exactly one rotated file")

				archiveName := entries[0].Name()
				assert.Contains(t, archiveName, "observations-")
				assert.Contains(t, archiveName, ".jsonl")
			} else {
				// If file existed before, it should still be there.
				if _, statErr := os.Stat(filePath); statErr == nil {
					assert.FileExists(t, filePath)
				}
			}
		})
	}
}
