package compact_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/compact"
)

func TestLogCompaction(t *testing.T) {
	tests := []struct {
		name          string
		setupExisting string
		wantLines     int
		wantPattern   string
	}{
		{
			name:          "creates file and writes entry when file does not exist",
			setupExisting: "",
			wantLines:     1,
			wantPattern:   `^\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] compaction triggered$`,
		},
		{
			name:          "appends to existing file",
			setupExisting: "[2025-01-01 00:00:00] compaction triggered\n",
			wantLines:     2,
			wantPattern:   `^\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] compaction triggered$`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logDir := t.TempDir()
			logFile := filepath.Join(logDir, "compaction-log.txt")

			if tt.setupExisting != "" {
				err := os.WriteFile(logFile, []byte(tt.setupExisting), 0o600)
				require.NoError(t, err)
			}

			err := compact.LogCompaction(logDir)
			require.NoError(t, err)

			data, readErr := os.ReadFile(logFile)
			require.NoError(t, readErr)

			lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
			assert.Len(t, lines, tt.wantLines)

			// Verify the last line matches the expected timestamp pattern.
			lastLine := lines[len(lines)-1]
			matched, matchErr := regexp.MatchString(tt.wantPattern, lastLine)
			require.NoError(t, matchErr)
			assert.True(t, matched,
				"last line %q should match pattern %q", lastLine, tt.wantPattern)
		})
	}
}

func TestLogCompaction_CreatesDirectory(t *testing.T) {
	logDir := filepath.Join(t.TempDir(), "nested", "log", "dir")

	err := compact.LogCompaction(logDir)
	require.NoError(t, err)

	logFile := filepath.Join(logDir, "compaction-log.txt")
	data, readErr := os.ReadFile(logFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "compaction triggered")
}
