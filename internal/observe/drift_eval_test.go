package observe_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/observe"
)

func TestObserver_RecordDriftEval(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	obs := observe.NewObserver(dir, 10)

	eval := observe.DriftEval{
		Timestamp:      time.Date(2026, time.September, 22, 12, 0, 0, 0, time.UTC),
		SessionID:      "abc123",
		Intent:         "refactor the auth middleware",
		Prompt:         "deploy the marketing site",
		IntentKeywords: []string{"refactor", "auth", "middleware"},
		PromptKeywords: []string{"deploy", "marketing", "site"},
		Overlap:        0.0,
		Threshold:      0.2,
		Edits:          7,
		Warned:         true,
	}
	require.NoError(t, obs.RecordDriftEval(eval))

	data, err := os.ReadFile(filepath.Join(dir, "drift-evals.jsonl"))
	require.NoError(t, err)
	require.True(t, bytes.HasSuffix(data, []byte("\n")), "each eval must be its own JSONL line")

	var got observe.DriftEval
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(data), &got))
	assert.Equal(t, eval, got)
}

// evalWithPrompt returns a fully populated DriftEval carrying prompt, so tests
// that only care about file mechanics do not restate every field.
func evalWithPrompt(prompt string) observe.DriftEval {
	return observe.DriftEval{
		Timestamp:      time.Date(2026, time.September, 22, 12, 0, 0, 0, time.UTC),
		SessionID:      "session",
		Intent:         "refactor the auth middleware",
		Prompt:         prompt,
		IntentKeywords: []string{"refactor", "auth", "middleware"},
		PromptKeywords: []string{"deploy", "marketing", "site"},
		Overlap:        0.0,
		Threshold:      0.2,
		Edits:          7,
		Warned:         true,
	}
}

func TestObserver_RecordDriftEval_AppendsLines(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	obs := observe.NewObserver(dir, 10)

	for _, prompt := range []string{"first", "second", "third"} {
		require.NoError(t, obs.RecordDriftEval(evalWithPrompt(prompt)))
	}

	data, err := os.ReadFile(filepath.Join(dir, "drift-evals.jsonl"))
	require.NoError(t, err)
	assert.Len(t, bytes.Split(bytes.TrimSpace(data), []byte("\n")), 3)
}

// TestObserver_RecordDriftEval_RespectsDisabledMarker confirms the existing
// observation kill switch also silences drift eval logging, so a user who has
// turned observation off never gets prompt text written to disk.
func TestObserver_RecordDriftEval_RespectsDisabledMarker(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".disabled"), nil, 0o600))

	obs := observe.NewObserver(dir, 10)
	require.NoError(t, obs.RecordDriftEval(evalWithPrompt("deploy the marketing site")))

	_, err := os.Stat(filepath.Join(dir, "drift-evals.jsonl"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

// TestObserver_RecordDriftEval_SeparateFromObservations keeps the eval log out
// of observations.jsonl, which the learning prompts consume.
func TestObserver_RecordDriftEval_SeparateFromObservations(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	obs := observe.NewObserver(dir, 10)

	require.NoError(t, obs.RecordDriftEval(evalWithPrompt("deploy the marketing site")))

	_, err := os.Stat(filepath.Join(dir, "observations.jsonl"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}
