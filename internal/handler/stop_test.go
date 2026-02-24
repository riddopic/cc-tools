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
			sessionID := hookcmd.SessionID("test-session")

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
	sessionID := hookcmd.SessionID("counter-test")
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
	counterPath := filepath.Join(stateDir, "stop-"+string(sessionID)+".count")
	data, err := os.ReadFile(counterPath)
	require.NoError(t, err)
	assert.Equal(t, "3", string(data))
}

func TestStopReminderHandler_CorruptCounterFile(t *testing.T) {
	t.Parallel()

	stateDir := t.TempDir()
	sessionID := hookcmd.SessionID("corrupt-test")
	cfg := stopConfig(true, 100, 200)

	// Write corrupt counter value.
	err := os.WriteFile(
		filepath.Join(stateDir, "stop-"+string(sessionID)+".count"),
		[]byte("not-a-number"),
		0o600,
	)
	require.NoError(t, err)

	h := handler.NewStopReminderHandler(cfg, handler.WithStopStateDir(stateDir))
	resp, handleErr := h.Handle(context.Background(), &hookcmd.HookInput{
		SessionID: sessionID,
	})
	require.NoError(t, handleErr)
	assert.Empty(t, resp.Stderr)

	// Verify counter was reset to 1 (corrupt treated as 0, then incremented).
	data, readErr := os.ReadFile(filepath.Join(stateDir, "stop-"+string(sessionID)+".count"))
	require.NoError(t, readErr)
	assert.Equal(t, "1", string(data))
}

func TestStopReminderHandler_IntervalZeroNoReminder(t *testing.T) {
	t.Parallel()

	stateDir := t.TempDir()
	sessionID := hookcmd.SessionID("zero-interval")
	cfg := stopConfig(true, 0, 0)
	h := handler.NewStopReminderHandler(cfg, handler.WithStopStateDir(stateDir))

	// Even at high counts, interval=0 and warnAt=0 should never emit.
	seedStopCount(t, stateDir, sessionID, 99)
	resp, err := h.Handle(context.Background(), &hookcmd.HookInput{
		SessionID: sessionID,
	})
	require.NoError(t, err)
	assert.Empty(t, resp.Stderr)
}

func TestStopReminderHandler_CounterPath_SafeSessionID(t *testing.T) {
	t.Parallel()

	stateDir := t.TempDir()
	cfg := stopConfig(true, 100, 200)
	h := handler.NewStopReminderHandler(cfg, handler.WithStopStateDir(stateDir))

	resp, err := h.Handle(context.Background(), &hookcmd.HookInput{
		SessionID: "../traversal",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	// Verify no file was created outside the state directory.
	entries, err := os.ReadDir(filepath.Dir(stateDir))
	require.NoError(t, err)
	for _, entry := range entries {
		if entry.Name() == filepath.Base(stateDir) {
			continue
		}
		assert.NotContains(t, entry.Name(), "stop-",
			"stop counter file must not escape stateDir")
	}

	// Verify the file was created inside stateDir with a safe name.
	stateEntries, err := os.ReadDir(stateDir)
	require.NoError(t, err)
	require.Len(t, stateEntries, 1)
	assert.NotContains(t, stateEntries[0].Name(), "..",
		"counter file name must not contain path traversal characters")
}

func stopConfig(enabled bool, interval, warnAt int) *config.Values {
	cfg := newTestConfig()
	cfg.StopReminder.Enabled = enabled
	cfg.StopReminder.Interval = interval
	cfg.StopReminder.WarnAt = warnAt
	return cfg
}

func seedStopCount(t *testing.T, stateDir string, sessionID hookcmd.SessionID, count int) {
	t.Helper()
	err := os.WriteFile(
		filepath.Join(stateDir, "stop-"+string(sessionID)+".count"),
		fmt.Appendf(nil, "%d", count),
		0o600,
	)
	require.NoError(t, err)
}
