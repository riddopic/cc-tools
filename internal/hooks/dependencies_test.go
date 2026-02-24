//go:build testmode

package hooks_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/hooks"
)

func TestRealCommandRunner_RunContext_ConcurrentOutput(t *testing.T) {
	t.Parallel()

	runner := hooks.NewRealCommandRunner()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Run a command that writes large interleaved stdout and stderr.
	// This would deadlock with sequential reads if the pipe buffer fills up,
	// because one pipe blocks while the other is not being drained.
	lineCount := 1000
	script := fmt.Sprintf(
		`for i in $(seq 1 %d); do echo "stdout line $i"; echo "stderr line $i" >&2; done`,
		lineCount,
	)

	output, err := runner.RunContext(ctx, "", "sh", "-c", script)
	require.NoError(t, err)

	// Verify stdout contains all expected lines
	stdoutStr := string(output.Stdout)
	for i := 1; i <= lineCount; i++ {
		expected := fmt.Sprintf("stdout line %d", i)
		assert.True(t, strings.Contains(stdoutStr, expected),
			"stdout missing line: %s", expected)
	}

	// Verify stderr contains all expected lines
	stderrStr := string(output.Stderr)
	for i := 1; i <= lineCount; i++ {
		expected := fmt.Sprintf("stderr line %d", i)
		assert.True(t, strings.Contains(stderrStr, expected),
			"stderr missing line: %s", expected)
	}

	// Count lines to ensure completeness
	stdoutLines := strings.Count(stdoutStr, "\n")
	stderrLines := strings.Count(stderrStr, "\n")
	assert.Equal(t, lineCount, stdoutLines, "stdout should have %d lines", lineCount)
	assert.Equal(t, lineCount, stderrLines, "stderr should have %d lines", lineCount)
}
