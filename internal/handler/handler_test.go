package handler_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/handler"
)

func TestResponse_ZeroValue(t *testing.T) {
	t.Parallel()
	resp := &handler.Response{}
	assert.Equal(t, 0, resp.ExitCode)
	assert.Nil(t, resp.Stdout)
	assert.Empty(t, resp.Stderr)
}

func TestHookOutput_JSON_OmitsEmptyFields(t *testing.T) {
	t.Parallel()
	out := &handler.HookOutput{
		Continue: true,
	}
	data, err := json.Marshal(out)
	require.NoError(t, err)

	// Should have "continue" but not empty fields
	assert.Contains(t, string(data), `"continue":true`)
	assert.NotContains(t, string(data), `"stopReason"`)
	assert.NotContains(t, string(data), `"hookSpecificOutput"`)
}

func TestHookOutput_JSON_FullOutput(t *testing.T) {
	t.Parallel()
	out := &handler.HookOutput{
		Continue:      true,
		SystemMessage: "context loaded",
		AdditionalContext: []string{
			"Previous session summary",
		},
	}
	data, err := json.Marshal(out)
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.True(t, parsed["continue"].(bool))
	assert.Equal(t, "context loaded", parsed["systemMessage"])
}
