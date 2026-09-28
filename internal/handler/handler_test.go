package handler_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
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
	data, err := json.Marshal(&handler.HookOutput{SystemMessage: "", HookSpecificOutput: nil})
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(data))
}

func TestHookOutput_JSON_MatchesHooksProtocol(t *testing.T) {
	t.Parallel()
	resp := handler.ContextResponse(hookcmd.EventSessionStart, "Previous session summary")
	data, err := json.Marshal(resp.Stdout)
	require.NoError(t, err)

	// additionalContext must be a string nested under hookSpecificOutput with
	// the event name; Claude Code ignores a top-level additionalContext.
	assert.JSONEq(t, `{"hookSpecificOutput":{"hookEventName":"SessionStart",`+
		`"additionalContext":"Previous session summary"}}`, string(data))
}

func TestUserMessageResponse_UsesSystemMessage(t *testing.T) {
	t.Parallel()
	resp := handler.UserMessageResponse("consider /compact")
	data, err := json.Marshal(resp.Stdout)
	require.NoError(t, err)
	assert.JSONEq(t, `{"systemMessage":"consider /compact"}`, string(data))
	assert.Empty(t, resp.Stderr, "stderr on exit 0 never reaches Claude or the user")
}

func TestResponse_Accessors_NilSafe(t *testing.T) {
	t.Parallel()
	var nilResp *handler.Response
	assert.Empty(t, nilResp.SystemMessage())
	assert.Empty(t, nilResp.AdditionalContext())
	assert.Empty(t, (&handler.Response{}).AdditionalContext())

	assert.Equal(t, "ctx", handler.ContextResponse(hookcmd.EventStop, "ctx").AdditionalContext())
	assert.Equal(t, "msg", handler.UserMessageResponse("msg").SystemMessage())
}
