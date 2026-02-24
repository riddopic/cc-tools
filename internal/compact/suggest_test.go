package compact_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/compact"
)

func TestSuggestor_RecordCall(t *testing.T) {
	tests := []struct {
		name       string
		threshold  int
		interval   int
		calls      int
		wantOutput bool
		wantSubstr string
	}{
		{
			name:       "first threshold hit triggers suggestion",
			threshold:  5,
			interval:   3,
			calls:      5,
			wantOutput: true,
			wantSubstr: "/compact",
		},
		{
			name:       "calls before threshold produce no suggestion",
			threshold:  50,
			interval:   10,
			calls:      49,
			wantOutput: false,
			wantSubstr: "",
		},
		{
			name:       "first reminder interval after threshold triggers suggestion",
			threshold:  5,
			interval:   3,
			calls:      8,
			wantOutput: true,
			wantSubstr: "/compact",
		},
		{
			name:       "second reminder interval triggers suggestion",
			threshold:  5,
			interval:   3,
			calls:      11,
			wantOutput: true,
			wantSubstr: "/compact",
		},
		{
			name:       "call between reminder intervals produces no suggestion",
			threshold:  5,
			interval:   3,
			calls:      7,
			wantOutput: false,
			wantSubstr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			s := compact.NewSuggestor(stateDir, tt.threshold, tt.interval)

			var buf bytes.Buffer

			for range tt.calls {
				buf.Reset()
				s.RecordCall("test-session", &buf)
			}

			if tt.wantOutput {
				assert.Contains(t, buf.String(), tt.wantSubstr)
			} else {
				assert.Empty(t, buf.String())
			}
		})
	}
}

func TestSuggestor_IndependentSessions(t *testing.T) {
	stateDir := t.TempDir()

	const threshold = 3

	s := compact.NewSuggestor(stateDir, threshold, 5)

	var buf bytes.Buffer

	// Advance session A to threshold - 1.
	for range threshold - 1 {
		buf.Reset()
		s.RecordCall("session-a", &buf)
	}

	assert.Empty(t, buf.String(), "session A should not suggest before threshold")

	// Advance session B to threshold.
	for range threshold {
		buf.Reset()
		s.RecordCall("session-b", &buf)
	}

	assert.Contains(t, buf.String(), "/compact",
		"session B should suggest at threshold")

	// Session A's next call should still not trigger (it is at threshold - 1 + 1 = threshold).
	buf.Reset()
	s.RecordCall("session-a", &buf)

	assert.Contains(t, buf.String(), "/compact",
		"session A should suggest when it reaches threshold independently")
}

func TestSuggestor_SafeSessionID(t *testing.T) {
	stateDir := t.TempDir()
	s := compact.NewSuggestor(stateDir, 1, 1)

	maliciousID := "../../../etc/passwd"
	var buf bytes.Buffer
	s.RecordCall(maliciousID, &buf)

	// Verify the counter file was created with a safe name inside stateDir.
	entries, err := os.ReadDir(stateDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)

	fileName := entries[0].Name()
	assert.NotContains(t, fileName, "..",
		"counter file name must not contain path traversal characters")
	assert.NotContains(t, fileName, "/",
		"counter file name must not contain path separators")
	assert.True(t, filepath.IsAbs(filepath.Join(stateDir, fileName)),
		"counter file must resolve to an absolute path within stateDir")
}

func TestSuggestor_MissingStateDir(t *testing.T) {
	stateDir := filepath.Join(t.TempDir(), "nonexistent", "subdir")

	// Verify the directory does not exist yet.
	_, err := os.Stat(stateDir)
	require.True(t, os.IsNotExist(err))

	s := compact.NewSuggestor(stateDir, 1, 1)

	var buf bytes.Buffer
	s.RecordCall("session-create", &buf)

	assert.Contains(t, buf.String(), "/compact",
		"should suggest even when state dir did not exist")

	// Verify the directory was created.
	info, statErr := os.Stat(stateDir)
	require.NoError(t, statErr)
	assert.True(t, info.IsDir())
}
