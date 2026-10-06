package compact_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/compact"
)

func TestSuggestor_Check(t *testing.T) {
	const threshold = 100_000

	tests := []struct {
		name   string
		tokens []int
		want   []bool
	}{
		{
			name:   "below threshold never nudges",
			tokens: []int{10_000, 50_000, 99_999},
			want:   []bool{false, false, false},
		},
		{
			name:   "above threshold nudges once",
			tokens: []int{120_000, 130_000, 140_000},
			want:   []bool{true, false, false},
		},
		{
			name:   "exactly at threshold nudges",
			tokens: []int{threshold},
			want:   []bool{true},
		},
		{
			name:   "dropping below re-arms the nudge",
			tokens: []int{120_000, 125_000, 30_000, 40_000, 110_000, 115_000},
			want:   []bool{true, false, false, false, true, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := compact.NewSuggestor(t.TempDir(), threshold)

			for i, tokens := range tt.tokens {
				msg := s.Check("test-session", tokens)
				if tt.want[i] {
					assert.Contains(t, msg, "/compact", "call %d", i)
				} else {
					assert.Empty(t, msg, "call %d", i)
				}
			}
		})
	}
}

func TestSuggestor_MessageReportsTokens(t *testing.T) {
	s := compact.NewSuggestor(t.TempDir(), 100_000)

	assert.Equal(t,
		"[cc-tools] Context is ~152k tokens. Consider running /compact to reduce context usage.",
		s.Check("session", 152_345))
}

func TestSuggestor_IndependentSessions(t *testing.T) {
	s := compact.NewSuggestor(t.TempDir(), 100)

	assert.NotEmpty(t, s.Check("session-a", 200))
	assert.NotEmpty(t, s.Check("session-b", 200), "session B has its own state")
	assert.Empty(t, s.Check("session-a", 200))
}

func TestSuggestor_StateFileName(t *testing.T) {
	stateDir := t.TempDir()
	s := compact.NewSuggestor(stateDir, 1)

	s.Check("../../../etc/passwd", 10)

	entries, err := os.ReadDir(stateDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)

	fileName := entries[0].Name()
	assert.NotContains(t, fileName, "..")
	assert.NotContains(t, fileName, "/")
	assert.Regexp(t, `^cc-tools-compact-[a-f0-9]{16}\.json$`, fileName)
}

func TestSuggestor_NoStateWrittenBelowThreshold(t *testing.T) {
	stateDir := t.TempDir()
	s := compact.NewSuggestor(stateDir, 100)

	assert.Empty(t, s.Check("session", 10))

	entries, err := os.ReadDir(stateDir)
	require.NoError(t, err)
	assert.Empty(t, entries, "no state file until a nudge is recorded")
}

func TestSuggestor_MissingStateDir(t *testing.T) {
	stateDir := filepath.Join(t.TempDir(), "nonexistent", "subdir")
	s := compact.NewSuggestor(stateDir, 1)

	assert.Contains(t, s.Check("session-create", 10), "/compact")

	info, err := os.Stat(stateDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	assert.Empty(t, s.Check("session-create", 10), "state persists across calls")
}

func TestSuggestor_CorruptStateTreatedAsUnsuggested(t *testing.T) {
	stateDir := t.TempDir()
	s := compact.NewSuggestor(stateDir, 1)

	path := filepath.Join(stateDir, "cc-tools-compact-session.json")
	require.NoError(t, os.WriteFile(path, []byte("not json"), 0o600))

	assert.Contains(t, s.Check("session", 10), "/compact")
}
