package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/riddopic/cc-tools/internal/observe"
)

// loadDriftState reads the persisted drift state for a session.
func loadDriftState(t *testing.T, stateDir string, sessionID hookcmd.SessionID) driftTestState {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(stateDir, "drift-"+string(sessionID)+".json"))
	require.NoError(t, err)

	var state driftTestState
	require.NoError(t, json.Unmarshal(data, &state))

	return state
}

// TestDriftHandler_IntentKeywordsSpanWholePrompt guards against starving the
// intent keyword set. An opening prompt whose first sentence carries a single
// meaningful word must still contribute the keywords that follow it; otherwise
// every later prompt scores zero overlap and warns.
func TestDriftHandler_IntentKeywordsSpanWholePrompt(t *testing.T) {
	t.Parallel()

	stateDir := t.TempDir()
	sessionID := hookcmd.SessionID("whole-prompt")
	h := handler.NewDriftHandler(driftConfig(true, 6, 0.2), handler.WithDriftStateDir(stateDir))

	_, err := h.Handle(context.Background(), &hookcmd.HookInput{
		SessionID: sessionID,
		Prompt:    "I want to refactor this. The auth middleware and session token validation both need work.",
	})
	require.NoError(t, err)

	state := loadDriftState(t, stateDir, sessionID)

	assert.Equal(t, "I want to refactor this.", state.Intent,
		"displayed intent stays the first sentence")
	assert.Subset(t, state.Keywords,
		[]string{"refactor", "auth", "middleware", "session", "token", "validation"},
		"keywords must come from the whole prompt, not just the first sentence")
}

// TestDriftHandler_RelatedFollowUpDoesNotWarn exercises the end-to-end effect
// of the wider keyword set: a follow-up working on a detail named later in the
// opening prompt is on-topic and must stay silent.
func TestDriftHandler_RelatedFollowUpDoesNotWarn(t *testing.T) {
	t.Parallel()

	stateDir := t.TempDir()
	sessionID := hookcmd.SessionID("related-followup")
	h := handler.NewDriftHandler(driftConfig(true, 1, 0.2), handler.WithDriftStateDir(stateDir))

	_, err := h.Handle(context.Background(), &hookcmd.HookInput{
		SessionID: sessionID,
		Prompt:    "I want to refactor this. The auth middleware and session token validation both need work.",
	})
	require.NoError(t, err)

	resp, err := h.Handle(context.Background(), &hookcmd.HookInput{
		SessionID: sessionID,
		Prompt:    "update the session token validation",
	})
	require.NoError(t, err)
	assert.Empty(t, resp.Stderr)
}

// TestDriftHandler_RepeatedKeywordDoesNotMaskDrift ensures a prompt cannot
// suppress a warning by repeating one on-topic word. Overlap is computed over
// distinct prompt keywords, so repetition carries no extra weight.
func TestDriftHandler_RepeatedKeywordDoesNotMaskDrift(t *testing.T) {
	t.Parallel()

	stateDir := t.TempDir()
	sessionID := hookcmd.SessionID("repeated-keyword")
	seedDriftState(t, stateDir, sessionID, &driftTestState{
		Intent:   "refactor the alpha beta gamma delta pipeline",
		Keywords: []string{"refactor", "alpha", "beta", "gamma", "delta", "pipeline"},
		Edits:    6,
	})

	h := handler.NewDriftHandler(driftConfig(true, 6, 0.2), handler.WithDriftStateDir(stateDir))
	resp, err := h.Handle(context.Background(), &hookcmd.HookInput{
		SessionID: sessionID,
		Prompt:    "alpha alpha alpha alpha zulu yankee xray whiskey victor",
	})
	require.NoError(t, err)
	assert.Contains(t, resp.Stderr, "Possible drift detected")
}

// TestDriftHandler_SparseIntentSuppressesWarning covers what a wider keyword
// set cannot fix: an opening prompt so terse it yields almost no keywords. With
// too little signal to judge against, silence beats warning on every turn.
func TestDriftHandler_SparseIntentSuppressesWarning(t *testing.T) {
	t.Parallel()

	stateDir := t.TempDir()
	sessionID := hookcmd.SessionID("sparse-intent")
	seedDriftState(t, stateDir, sessionID, &driftTestState{
		Intent:   "fix this",
		Keywords: []string{"fix"},
		Edits:    6,
	})

	h := handler.NewDriftHandler(driftConfig(true, 6, 0.2), handler.WithDriftStateDir(stateDir))
	resp, err := h.Handle(context.Background(), &hookcmd.HookInput{
		SessionID: sessionID,
		Prompt:    "run the linter and check for gosec findings",
	})
	require.NoError(t, err)
	assert.Empty(t, resp.Stderr)
}

// TestDriftHandler_LogsEvaluationWhenEnabled verifies that a scored prompt is
// appended to the eval log with everything needed to hand-label it later.
func TestDriftHandler_LogsEvaluationWhenEnabled(t *testing.T) {
	t.Parallel()

	stateDir := t.TempDir()
	evalDir := t.TempDir()
	sessionID := hookcmd.SessionID("eval-log")

	intentKeywords := []string{"refactor", "auth", "middleware", "session", "token", "validation"}
	seedDriftState(t, stateDir, sessionID, &driftTestState{
		Intent:   "refactor the auth middleware and session token validation",
		Keywords: intentKeywords,
		Edits:    6,
	})

	cfg := driftConfig(true, 6, 0.2)
	cfg.Drift.LogEvals = true

	h := handler.NewDriftHandler(cfg,
		handler.WithDriftStateDir(stateDir),
		handler.WithDriftEvalDir(evalDir),
	)
	_, err := h.Handle(context.Background(), &hookcmd.HookInput{
		SessionID: sessionID,
		Prompt:    "deploy the marketing site to production",
	})
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(evalDir, "drift-evals.jsonl"))
	require.NoError(t, err)

	var eval observe.DriftEval
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(data), &eval))

	assert.Equal(t, string(sessionID), eval.SessionID)
	assert.Equal(t, "deploy the marketing site to production", eval.Prompt)
	assert.Equal(t, intentKeywords, eval.IntentKeywords)
	assert.InDelta(t, 0.2, eval.Threshold, 0.001)
	assert.Equal(t, 7, eval.Edits)
	assert.True(t, eval.Warned)
	assert.InDelta(t, 0.0, eval.Overlap, 0.001)
}

// TestDriftHandler_DoesNotLogEvaluationByDefault keeps prompt text off disk
// unless the user has explicitly opted in to a measurement run.
func TestDriftHandler_DoesNotLogEvaluationByDefault(t *testing.T) {
	t.Parallel()

	stateDir := t.TempDir()
	evalDir := t.TempDir()
	sessionID := hookcmd.SessionID("eval-off")

	seedDriftState(t, stateDir, sessionID, &driftTestState{
		Intent:   "refactor the auth middleware",
		Keywords: []string{"refactor", "auth", "middleware"},
		Edits:    6,
	})

	h := handler.NewDriftHandler(driftConfig(true, 6, 0.2),
		handler.WithDriftStateDir(stateDir),
		handler.WithDriftEvalDir(evalDir),
	)
	_, err := h.Handle(context.Background(), &hookcmd.HookInput{
		SessionID: sessionID,
		Prompt:    "deploy the marketing site",
	})
	require.NoError(t, err)

	_, statErr := os.Stat(filepath.Join(evalDir, "drift-evals.jsonl"))
	assert.ErrorIs(t, statErr, os.ErrNotExist)
}
