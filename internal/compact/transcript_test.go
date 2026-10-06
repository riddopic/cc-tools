package compact_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/compact"
)

const (
	userLine  = `{"type":"user","isSidechain":false,"message":{"role":"user","content":"hi"}}`
	firstAsst = `{"type":"assistant","isSidechain":false,"message":{"usage":` +
		`{"input_tokens":1,"cache_read_input_tokens":1000,"cache_creation_input_tokens":200,"output_tokens":5}}}`
	lastAsst = `{"type":"assistant","isSidechain":false,"message":{"usage":` +
		`{"input_tokens":2,"cache_read_input_tokens":112416,"cache_creation_input_tokens":9608,"output_tokens":7}}}`
	sidechainAsst = `{"type":"assistant","isSidechain":true,"message":{"usage":` +
		`{"input_tokens":9,"cache_read_input_tokens":900000,"cache_creation_input_tokens":0,"output_tokens":1}}}`
	noUsageAsst = `{"type":"assistant","isSidechain":false,"message":{"content":"x"}}`

	lastAsstTokens  = 2 + 112416 + 9608
	firstAsstTokens = 1 + 1000 + 200
)

func writeTranscript(t *testing.T, lines ...string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "transcript.jsonl")
	require.NoError(t, os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600))

	return path
}

func TestContextTokens(t *testing.T) {
	tests := []struct {
		name    string
		lines   []string
		want    int
		wantErr bool
	}{
		{
			name:    "last usage entry wins",
			lines:   []string{userLine, firstAsst, userLine, lastAsst, userLine},
			want:    lastAsstTokens,
			wantErr: false,
		},
		{
			name:    "sidechain entries are ignored",
			lines:   []string{lastAsst, sidechainAsst},
			want:    lastAsstTokens,
			wantErr: false,
		},
		{
			name:    "entries without usage are skipped",
			lines:   []string{firstAsst, noUsageAsst},
			want:    firstAsstTokens,
			wantErr: false,
		},
		{
			name:    "malformed lines are skipped",
			lines:   []string{firstAsst, `{not json`},
			want:    firstAsstTokens,
			wantErr: false,
		},
		{
			name:    "no usage at all is an error",
			lines:   []string{userLine, noUsageAsst, sidechainAsst},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compact.ContextTokens(writeTranscript(t, tt.lines...))
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContextTokens_MissingFile(t *testing.T) {
	got, err := compact.ContextTokens(filepath.Join(t.TempDir(), "missing.jsonl"))
	require.Error(t, err)
	assert.Zero(t, got)
}

func TestContextTokens_EmptyPath(t *testing.T) {
	got, err := compact.ContextTokens("")
	require.Error(t, err)
	assert.Zero(t, got)
}

func TestContextTokens_ReadsOnlyTail(t *testing.T) {
	head := `{"type":"assistant","isSidechain":false,"message":{"usage":` +
		`{"input_tokens":7,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`
	padding := `{"type":"user","message":{"content":"` + strings.Repeat("x", 1<<20) + `"}}`
	tail := `{"type":"assistant","isSidechain":false,"message":{"usage":` +
		`{"input_tokens":3,"cache_read_input_tokens":4,"cache_creation_input_tokens":5}}}`

	got, err := compact.ContextTokens(writeTranscript(t, head, padding, tail))
	require.NoError(t, err)
	assert.Equal(t, 12, got)

	// The head entry sits outside the tail window, so it must not be found.
	got, err = compact.ContextTokens(writeTranscript(t, head, padding))
	require.Error(t, err)
	assert.Zero(t, got)
}

func TestContextTokens_LargeLinesInTail(t *testing.T) {
	big := `{"type":"user","message":{"content":"` + strings.Repeat("y", 200*1024) + `"}}`

	got, err := compact.ContextTokens(writeTranscript(t, big, lastAsst, big))
	require.NoError(t, err)
	assert.Equal(t, lastAsstTokens, got)
}

func TestContextTokens_NoTrailingNewline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "transcript.jsonl")
	require.NoError(t, os.WriteFile(path, []byte(userLine+"\n"+lastAsst), 0o600))

	got, err := compact.ContextTokens(path)
	require.NoError(t, err)
	assert.Equal(t, lastAsstTokens, got)
}
