package session_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/session"
)

func TestParseTranscript_ExtractsSummary(t *testing.T) {
	lines := []string{
		`{"type":"human","content":"Fix the bug"}`,
		`{"type":"tool_use","name":"Edit","input":{"file_path":"main.go"}}`,
		`{"type":"tool_use","name":"Bash","input":{"command":"go test"}}`,
		`{"type":"human","content":"Looks good, commit it"}`,
		`{"type":"tool_use","name":"Edit","input":{"file_path":"handler.go"}}`,
	}

	tPath := filepath.Join(t.TempDir(), "transcript.jsonl")
	require.NoError(t, os.WriteFile(tPath, []byte(strings.Join(lines, "\n")), 0o600))

	summary, parseErr := session.ParseTranscript(tPath)
	require.NoError(t, parseErr)

	assert.Equal(t, 2, summary.TotalMessages)
	assert.Contains(t, summary.ToolsUsed, "Edit")
	assert.Contains(t, summary.ToolsUsed, "Bash")
	assert.Contains(t, summary.FilesModified, "main.go")
	assert.Contains(t, summary.FilesModified, "handler.go")
}

func TestParseTranscript_DeduplicatesToolsAndFiles(t *testing.T) {
	lines := []string{
		`{"type":"tool_use","name":"Edit","input":{"file_path":"main.go"}}`,
		`{"type":"tool_use","name":"Edit","input":{"file_path":"main.go"}}`,
		`{"type":"tool_use","name":"Bash","input":{"command":"ls"}}`,
		`{"type":"tool_use","name":"Bash","input":{"command":"pwd"}}`,
	}

	tPath := filepath.Join(t.TempDir(), "transcript.jsonl")
	require.NoError(t, os.WriteFile(tPath, []byte(strings.Join(lines, "\n")), 0o600))

	summary, parseErr := session.ParseTranscript(tPath)
	require.NoError(t, parseErr)

	assert.Equal(t, 0, summary.TotalMessages)
	assert.Len(t, summary.ToolsUsed, 2)
	assert.Len(t, summary.FilesModified, 1)
}

func TestParseTranscript_ReturnsEmptySummaryForEmptyFile(t *testing.T) {
	tPath := filepath.Join(t.TempDir(), "empty.jsonl")
	require.NoError(t, os.WriteFile(tPath, []byte(""), 0o600))

	summary, parseErr := session.ParseTranscript(tPath)
	require.NoError(t, parseErr)

	assert.Equal(t, 0, summary.TotalMessages)
	assert.Empty(t, summary.ToolsUsed)
	assert.Empty(t, summary.FilesModified)
}

func TestParseTranscript_ReturnsErrorForMissingFile(t *testing.T) {
	_, err := session.ParseTranscript("/tmp/nonexistent-transcript-xyz.jsonl")
	require.Error(t, err)
}

func TestParseTranscript_SkipsInvalidJSONLines(t *testing.T) {
	lines := []string{
		`{"type":"human","content":"Hello"}`,
		`not valid json`,
		`{"type":"tool_use","name":"Read","input":{"file_path":"config.go"}}`,
	}

	tPath := filepath.Join(t.TempDir(), "transcript.jsonl")
	require.NoError(t, os.WriteFile(tPath, []byte(strings.Join(lines, "\n")), 0o600))

	summary, parseErr := session.ParseTranscript(tPath)
	require.NoError(t, parseErr)

	assert.Equal(t, 1, summary.TotalMessages)
	assert.Len(t, summary.ToolsUsed, 1)
	assert.Contains(t, summary.ToolsUsed, "Read")
	assert.Contains(t, summary.FilesModified, "config.go")
}

func TestParseTranscript_HandlesMissingFilePath(t *testing.T) {
	lines := []string{
		`{"type":"tool_use","name":"Bash","input":{"command":"go build"}}`,
	}

	tPath := filepath.Join(t.TempDir(), "transcript.jsonl")
	require.NoError(t, os.WriteFile(tPath, []byte(strings.Join(lines, "\n")), 0o600))

	summary, parseErr := session.ParseTranscript(tPath)
	require.NoError(t, parseErr)

	assert.Len(t, summary.ToolsUsed, 1)
	assert.Contains(t, summary.ToolsUsed, "Bash")
	assert.Empty(t, summary.FilesModified)
}
