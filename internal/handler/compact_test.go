package handler_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
)

// ---------------------------------------------------------------------
// LogCompactionHandler
// ---------------------------------------------------------------------

func TestLogCompactionHandler_Name(t *testing.T) {
	t.Parallel()
	h := handler.NewLogCompactionHandler()
	assert.Equal(t, "log-compaction", h.Name())
}

func TestLogCompactionHandler_Handle(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	h := handler.NewLogCompactionHandler(handler.WithCompactLogDir(tmpDir))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreCompact,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	// Compaction log file should exist.
	logFile := filepath.Join(tmpDir, "compaction-log.txt")
	data, readErr := os.ReadFile(logFile)
	require.NoError(t, readErr, "compaction log file should be created")
	assert.Contains(t, string(data), "compaction triggered")
}

func TestLogCompactionHandler_ImplementsHandler(t *testing.T) {
	t.Parallel()
	var _ handler.Handler = handler.NewLogCompactionHandler()
}

func TestLogCompactionHandler_AppendsOnMultipleCalls(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	h := handler.NewLogCompactionHandler(handler.WithCompactLogDir(tmpDir))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreCompact,
	}

	for range 3 {
		resp, err := h.Handle(context.Background(), input)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 0, resp.ExitCode)
	}

	logFile := filepath.Join(tmpDir, "compaction-log.txt")
	data, err := os.ReadFile(logFile)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	assert.Len(t, lines, 3, "expected 3 log lines after 3 Handle calls")

	for i, line := range lines {
		assert.Contains(t, line, "compaction triggered",
			"line %d should contain 'compaction triggered'", i+1)
	}
}

func TestLogCompactionHandler_EntryFormat(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	h := handler.NewLogCompactionHandler(handler.WithCompactLogDir(tmpDir))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreCompact,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	logFile := filepath.Join(tmpDir, "compaction-log.txt")
	data, err := os.ReadFile(logFile)
	require.NoError(t, err)

	line := strings.TrimSpace(string(data))
	assert.Regexp(t, `^\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] compaction triggered$`,
		line, "log entry should match timestamp format")
}
