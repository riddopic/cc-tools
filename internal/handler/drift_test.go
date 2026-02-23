package handler_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
)

func TestDriftHandler_Name(t *testing.T) {
	t.Parallel()
	h := handler.NewDriftHandler(nil)
	assert.Equal(t, "drift-detection", h.Name())
}

func TestDriftHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cfg        *config.Values
		seedState  *driftTestState
		prompt     string
		wantStderr string
		wantErr    bool
	}{
		{
			name:       "nil config returns exit 0",
			cfg:        nil,
			seedState:  nil,
			prompt:     "fix the bug",
			wantStderr: "",
			wantErr:    false,
		},
		{
			name:       "disabled drift returns exit 0",
			cfg:        driftConfig(false, 6, 0.2),
			seedState:  nil,
			prompt:     "fix the bug",
			wantStderr: "",
			wantErr:    false,
		},
		{
			name:       "empty prompt returns exit 0",
			cfg:        driftConfig(true, 6, 0.2),
			seedState:  nil,
			prompt:     "",
			wantStderr: "",
			wantErr:    false,
		},
		{
			name:       "first prompt establishes intent",
			cfg:        driftConfig(true, 6, 0.2),
			seedState:  nil,
			prompt:     "refactor the authentication module",
			wantStderr: "",
			wantErr:    false,
		},
		{
			name: "below min edits skips drift check",
			cfg:  driftConfig(true, 6, 0.2),
			seedState: &driftTestState{
				Intent:   "refactor the authentication module",
				Keywords: []string{"refactor", "authentication", "module"},
				Edits:    3,
			},
			prompt:     "completely unrelated topic about cooking",
			wantStderr: "",
			wantErr:    false,
		},
		{
			name: "above threshold overlap no warning",
			cfg:  driftConfig(true, 2, 0.5),
			seedState: &driftTestState{
				Intent:   "refactor the authentication module",
				Keywords: []string{"refactor", "authentication", "module"},
				Edits:    5,
			},
			prompt:     "refactor authentication module again",
			wantStderr: "",
			wantErr:    false,
		},
		{
			name: "below threshold overlap triggers warning",
			cfg:  driftConfig(true, 2, 0.2),
			seedState: &driftTestState{
				Intent:   "refactor the authentication module",
				Keywords: []string{"refactor", "authentication", "module"},
				Edits:    5,
			},
			prompt:     "update the database migration scripts for postgres",
			wantStderr: "Possible drift detected",
			wantErr:    false,
		},
		{
			name: "pivot phrase resets intent",
			cfg:  driftConfig(true, 2, 0.2),
			seedState: &driftTestState{
				Intent:   "refactor the authentication module",
				Keywords: []string{"refactor", "authentication", "module"},
				Edits:    10,
			},
			prompt:     "now let's work on the database layer",
			wantStderr: "",
			wantErr:    false,
		},
		{
			name: "empty prompt keywords return no drift",
			cfg:  driftConfig(true, 2, 0.2),
			seedState: &driftTestState{
				Intent:   "refactor the authentication module",
				Keywords: []string{"refactor", "authentication", "module"},
				Edits:    10,
			},
			prompt:     "ok",
			wantStderr: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stateDir := t.TempDir()
			sessionID := "test-session"

			if tt.seedState != nil {
				seedDriftState(t, stateDir, sessionID, tt.seedState)
			}

			h := handler.NewDriftHandler(tt.cfg, handler.WithDriftStateDir(stateDir))
			resp, err := h.Handle(context.Background(), &hookcmd.HookInput{
				SessionID: sessionID,
				Prompt:    tt.prompt,
			})

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, 0, resp.ExitCode)

			if tt.wantStderr != "" {
				assert.Contains(t, resp.Stderr, tt.wantStderr)
			} else {
				assert.Empty(t, resp.Stderr)
			}
		})
	}
}

func TestDriftHandler_IntentPersistence(t *testing.T) {
	t.Parallel()

	stateDir := t.TempDir()
	sessionID := "persist-test"
	cfg := driftConfig(true, 6, 0.2)
	h := handler.NewDriftHandler(cfg, handler.WithDriftStateDir(stateDir))

	// First prompt establishes intent.
	resp, err := h.Handle(context.Background(), &hookcmd.HookInput{
		SessionID: sessionID,
		Prompt:    "implement user authentication with JWT tokens",
	})
	require.NoError(t, err)
	assert.Empty(t, resp.Stderr)

	// Verify state file was created.
	statePath := filepath.Join(stateDir, "drift-"+sessionID+".json")
	data, err := os.ReadFile(statePath)
	require.NoError(t, err)

	var state driftTestState
	require.NoError(t, json.Unmarshal(data, &state))
	assert.NotEmpty(t, state.Intent)
	assert.NotEmpty(t, state.Keywords)
	assert.Equal(t, 0, state.Edits)
}

func TestDriftHandler_PivotResetsState(t *testing.T) {
	t.Parallel()

	stateDir := t.TempDir()
	sessionID := "pivot-test"
	cfg := driftConfig(true, 2, 0.2)
	h := handler.NewDriftHandler(cfg, handler.WithDriftStateDir(stateDir))

	// Establish initial intent.
	_, err := h.Handle(context.Background(), &hookcmd.HookInput{
		SessionID: sessionID,
		Prompt:    "work on authentication module",
	})
	require.NoError(t, err)

	// Pivot to new topic.
	_, err = h.Handle(context.Background(), &hookcmd.HookInput{
		SessionID: sessionID,
		Prompt:    "switch to database optimization work",
	})
	require.NoError(t, err)

	// Verify intent was updated.
	statePath := filepath.Join(stateDir, "drift-"+sessionID+".json")
	data, err := os.ReadFile(statePath)
	require.NoError(t, err)

	var state driftTestState
	require.NoError(t, json.Unmarshal(data, &state))
	assert.Contains(t, state.Intent, "database optimization")
	assert.Equal(t, 0, state.Edits)
}

func TestDriftHandler_IntentTruncation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		prompt       string
		wantContains string
		wantMaxLen   int
	}{
		{
			name:         "sentence ending with period",
			prompt:       "Fix the login bug. Then refactor the auth module.",
			wantContains: "Fix the login bug.",
			wantMaxLen:   18,
		},
		{
			name:         "sentence ending with exclamation",
			prompt:       "Fix this now! It is broken badly.",
			wantContains: "Fix this now!",
			wantMaxLen:   13,
		},
		{
			name:         "sentence ending with question mark",
			prompt:       "Can you fix the bug? I think it is in auth.",
			wantContains: "Can you fix the bug?",
			wantMaxLen:   20,
		},
		{
			name:         "newline truncates intent",
			prompt:       "Fix the auth module\nAlso update the tests",
			wantContains: "Fix the auth module",
			wantMaxLen:   19,
		},
		{
			name:         "short text without sentence enders",
			prompt:       "fix auth",
			wantContains: "fix auth",
			wantMaxLen:   8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stateDir := t.TempDir()
			sessionID := "truncation-test"
			cfg := driftConfig(true, 6, 0.2)
			h := handler.NewDriftHandler(cfg, handler.WithDriftStateDir(stateDir))

			_, err := h.Handle(context.Background(), &hookcmd.HookInput{
				SessionID: sessionID,
				Prompt:    tt.prompt,
			})
			require.NoError(t, err)

			statePath := filepath.Join(stateDir, "drift-"+sessionID+".json")
			data, readErr := os.ReadFile(statePath)
			require.NoError(t, readErr)

			var state driftTestState
			require.NoError(t, json.Unmarshal(data, &state))
			assert.Equal(t, tt.wantContains, state.Intent)
			assert.LessOrEqual(t, len(state.Intent), tt.wantMaxLen)
		})
	}
}

func TestDriftHandler_CorruptStateFile(t *testing.T) {
	t.Parallel()

	stateDir := t.TempDir()
	sessionID := "corrupt-test"
	cfg := driftConfig(true, 6, 0.2)

	// Write corrupt JSON to state file.
	err := os.WriteFile(
		filepath.Join(stateDir, "drift-"+sessionID+".json"),
		[]byte("not valid json{{{"),
		0o600,
	)
	require.NoError(t, err)

	h := handler.NewDriftHandler(cfg, handler.WithDriftStateDir(stateDir))
	resp, handleErr := h.Handle(context.Background(), &hookcmd.HookInput{
		SessionID: sessionID,
		Prompt:    "start fresh after corrupt state",
	})
	require.NoError(t, handleErr)
	assert.Empty(t, resp.Stderr)

	// Verify state was re-initialized (corrupt state treated as empty).
	data, readErr := os.ReadFile(filepath.Join(stateDir, "drift-"+sessionID+".json"))
	require.NoError(t, readErr)

	var state driftTestState
	require.NoError(t, json.Unmarshal(data, &state))
	assert.NotEmpty(t, state.Intent)
	assert.Equal(t, 0, state.Edits)
}

// driftTestState mirrors the internal driftState struct for test seeding.
type driftTestState struct {
	Intent   string   `json:"intent"`
	Keywords []string `json:"keywords"`
	Edits    int      `json:"edits"`
}

func driftConfig(enabled bool, minEdits int, threshold float64) *config.Values {
	cfg := newTestConfig()
	cfg.Drift.Enabled = enabled
	cfg.Drift.MinEdits = minEdits
	cfg.Drift.Threshold = threshold
	return cfg
}

func seedDriftState(t *testing.T, stateDir, sessionID string, state *driftTestState) {
	t.Helper()
	data, err := json.Marshal(state)
	require.NoError(t, err)
	err = os.WriteFile(
		filepath.Join(stateDir, "drift-"+sessionID+".json"),
		data,
		0o600,
	)
	require.NoError(t, err)
}
