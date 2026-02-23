package handler_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
)

func TestStopReminderHandler_Name(t *testing.T) {
	t.Parallel()
	h := handler.NewStopReminderHandler(nil)
	assert.Equal(t, "stop-reminder", h.Name())
}

func TestStopReminderHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cfg        *config.Values
		seedCount  int
		wantStderr string
		wantErr    bool
	}{
		{
			name:       "nil config returns exit 0",
			cfg:        nil,
			seedCount:  0,
			wantStderr: "",
			wantErr:    false,
		},
		{
			name:       "disabled returns exit 0",
			cfg:        stopConfig(false, 20, 50),
			seedCount:  0,
			wantStderr: "",
			wantErr:    false,
		},
		{
			name:       "below interval no reminder",
			cfg:        stopConfig(true, 20, 50),
			seedCount:  5,
			wantStderr: "",
			wantErr:    false,
		},
		{
			name:       "at interval emits first reminder",
			cfg:        stopConfig(true, 20, 50),
			seedCount:  19,
			wantStderr: "running /compact",
			wantErr:    false,
		},
		{
			name:       "at second interval emits second reminder",
			cfg:        stopConfig(true, 20, 50),
			seedCount:  39,
			wantStderr: "committing progress",
			wantErr:    false,
		},
		{
			name:       "at third interval emits third reminder",
			cfg:        stopConfig(true, 20, 100),
			seedCount:  59,
			wantStderr: "checkpoint",
			wantErr:    false,
		},
		{
			name:       "at warn threshold emits strong warning",
			cfg:        stopConfig(true, 20, 50),
			seedCount:  49,
			wantStderr: "strongly consider wrapping up",
			wantErr:    false,
		},
		{
			name:       "above warn threshold still warns",
			cfg:        stopConfig(true, 20, 50),
			seedCount:  55,
			wantStderr: "strongly consider wrapping up",
			wantErr:    false,
		},
		{
			name:       "warn at zero disables strong warning",
			cfg:        stopConfig(true, 20, 0),
			seedCount:  59,
			wantStderr: "checkpoint",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stateDir := t.TempDir()
			sessionID := "test-session"

			if tt.seedCount > 0 {
				seedStopCount(t, stateDir, sessionID, tt.seedCount)
			}

			h := handler.NewStopReminderHandler(tt.cfg, handler.WithStopStateDir(stateDir))
			resp, err := h.Handle(context.Background(), &hookcmd.HookInput{
				SessionID: sessionID,
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

func TestStopReminderHandler_CounterPersistence(t *testing.T) {
	t.Parallel()

	stateDir := t.TempDir()
	sessionID := "counter-test"
	cfg := stopConfig(true, 100, 200)
	h := handler.NewStopReminderHandler(cfg, handler.WithStopStateDir(stateDir))

	// Call three times.
	for range 3 {
		_, err := h.Handle(context.Background(), &hookcmd.HookInput{
			SessionID: sessionID,
		})
		require.NoError(t, err)
	}

	// Verify counter file.
	counterPath := filepath.Join(stateDir, "stop-"+sessionID+".count")
	data, err := os.ReadFile(counterPath)
	require.NoError(t, err)
	assert.Equal(t, "3", string(data))
}

func stopConfig(enabled bool, interval, warnAt int) *config.Values {
	cfg := newTestConfig()
	cfg.StopReminder.Enabled = enabled
	cfg.StopReminder.Interval = interval
	cfg.StopReminder.WarnAt = warnAt
	return cfg
}

func seedStopCount(t *testing.T, stateDir, sessionID string, count int) {
	t.Helper()
	err := os.WriteFile(
		filepath.Join(stateDir, "stop-"+sessionID+".count"),
		fmt.Appendf(nil, "%d", count),
		0o600,
	)
	require.NoError(t, err)
}
