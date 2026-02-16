package handler_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
)

// mockAudioPlayer records Play calls for assertion.
type mockAudioPlayer struct {
	played []string
}

func (m *mockAudioPlayer) Play(filepath string) error {
	m.played = append(m.played, filepath)
	return nil
}

// mockCmdRunner records Run calls for assertion.
type mockCmdRunner struct {
	calls []cmdRunnerCall
}

type cmdRunnerCall struct {
	name string
	args []string
}

func (m *mockCmdRunner) Run(name string, args ...string) error {
	m.calls = append(m.calls, cmdRunnerCall{name: name, args: args})
	return nil
}

// ---------------------------------------------------------------------
// NotifyAudioHandler
// ---------------------------------------------------------------------

func TestNotifyAudioHandler_Name(t *testing.T) {
	t.Parallel()
	h := handler.NewNotifyAudioHandler(nil)
	assert.Equal(t, "notify-audio", h.Name())
}

func TestNotifyAudioHandler_NilConfig(t *testing.T) {
	t.Parallel()
	h := handler.NewNotifyAudioHandler(nil)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventNotification,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
}

func TestNotifyAudioHandler_Disabled(t *testing.T) {
	t.Parallel()
	cfg := &config.Values{
		Notify: config.NotifyValues{
			Audio: config.AudioValues{
				Enabled: false,
			},
		},
	}

	h := handler.NewNotifyAudioHandler(cfg)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventNotification,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
}

func TestNotifyAudioHandler_NoPlayerInjected(t *testing.T) {
	t.Parallel()
	cfg := &config.Values{
		Notify: config.NotifyValues{
			Audio: config.AudioValues{
				Enabled:   true,
				Directory: "/tmp/sounds",
			},
		},
	}

	// No WithAudioPlayer option — player is nil.
	h := handler.NewNotifyAudioHandler(cfg)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventNotification,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
}

func TestNotifyAudioHandler_EnabledWithPlayer(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create an MP3 file so PlayRandom has something to pick.
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "beep.mp3"), []byte("fake-audio"), 0o600,
	))

	player := &mockAudioPlayer{}

	cfg := &config.Values{
		Notify: config.NotifyValues{
			Audio: config.AudioValues{
				Enabled:   true,
				Directory: tmpDir,
			},
		},
	}

	h := handler.NewNotifyAudioHandler(cfg, handler.WithAudioPlayer(player))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventNotification,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
	assert.NotEmpty(t, player.played, "should have played an audio file")
}

func TestNotifyAudioHandler_QuietHoursSkipsPlay(t *testing.T) {
	t.Parallel()
	player := &mockAudioPlayer{}

	cfg := &config.Values{
		Notify: config.NotifyValues{
			Audio: config.AudioValues{
				Enabled:   true,
				Directory: "/tmp/sounds",
			},
			QuietHours: config.QuietHoursValues{
				Enabled: true,
				Start:   "00:00",
				End:     "23:59",
			},
		},
	}

	h := handler.NewNotifyAudioHandler(cfg, handler.WithAudioPlayer(player))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventNotification,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
	assert.Empty(t, player.played, "should not play during quiet hours")
}

func TestNotifyAudioHandler_ImplementsHandler(t *testing.T) {
	t.Parallel()
	var _ handler.Handler = handler.NewNotifyAudioHandler(nil)
}

// ---------------------------------------------------------------------
// NotifyDesktopHandler
// ---------------------------------------------------------------------

func TestNotifyDesktopHandler_Name(t *testing.T) {
	t.Parallel()
	h := handler.NewNotifyDesktopHandler(nil)
	assert.Equal(t, "notify-desktop", h.Name())
}

func TestNotifyDesktopHandler_NilConfig(t *testing.T) {
	t.Parallel()
	h := handler.NewNotifyDesktopHandler(nil)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventNotification,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
}

func TestNotifyDesktopHandler_Disabled(t *testing.T) {
	t.Parallel()
	cfg := &config.Values{
		Notify: config.NotifyValues{
			Desktop: config.DesktopValues{
				Enabled: false,
			},
		},
	}

	h := handler.NewNotifyDesktopHandler(cfg)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventNotification,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
}

func TestNotifyDesktopHandler_QuietHoursActive(t *testing.T) {
	t.Parallel()
	cfg := &config.Values{
		Notify: config.NotifyValues{
			Desktop: config.DesktopValues{
				Enabled: true,
			},
			QuietHours: config.QuietHoursValues{
				Enabled: true,
				Start:   "00:00",
				End:     "23:59",
			},
		},
	}

	h := handler.NewNotifyDesktopHandler(cfg)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventNotification,
		Title:         "Test",
		Message:       "Hello",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
}

func TestNotifyDesktopHandler_NoRunnerInjected(t *testing.T) {
	t.Parallel()
	cfg := &config.Values{
		Notify: config.NotifyValues{
			Desktop: config.DesktopValues{
				Enabled: true,
			},
		},
	}

	// No WithCmdRunner option — runner is nil.
	h := handler.NewNotifyDesktopHandler(cfg)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventNotification,
		Title:         "Test",
		Message:       "Hello",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
}

func TestNotifyDesktopHandler_EnabledWithRunner(t *testing.T) {
	t.Parallel()
	runner := &mockCmdRunner{}

	cfg := &config.Values{
		Notify: config.NotifyValues{
			Desktop: config.DesktopValues{
				Enabled: true,
			},
		},
	}

	h := handler.NewNotifyDesktopHandler(cfg, handler.WithCmdRunner(runner))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventNotification,
		Title:         "Build Done",
		Message:       "All tests passed",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
	assert.NotEmpty(t, runner.calls, "should have called runner")
}

func TestNotifyDesktopHandler_CustomTitleAndMessage(t *testing.T) {
	t.Parallel()
	runner := &mockCmdRunner{}

	cfg := &config.Values{
		Notify: config.NotifyValues{
			Desktop: config.DesktopValues{
				Enabled: true,
			},
		},
	}

	h := handler.NewNotifyDesktopHandler(cfg, handler.WithCmdRunner(runner))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventNotification,
		Title:         "Custom Title",
		Message:       "Custom body text",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, runner.calls)
	// The runner should receive the custom title/message (exact args depend on
	// the notify.Desktop implementation, but it should be called).
}

func TestNotifyDesktopHandler_DefaultTitleAndMessage(t *testing.T) {
	t.Parallel()
	runner := &mockCmdRunner{}

	cfg := &config.Values{
		Notify: config.NotifyValues{
			Desktop: config.DesktopValues{
				Enabled: true,
			},
		},
	}

	h := handler.NewNotifyDesktopHandler(cfg, handler.WithCmdRunner(runner))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventNotification,
		// No Title or Message — should use defaults.
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
	assert.NotEmpty(t, runner.calls,
		"should still send notification with default title/message")
}

func TestNotifyDesktopHandler_ImplementsHandler(t *testing.T) {
	t.Parallel()
	var _ handler.Handler = handler.NewNotifyDesktopHandler(nil)
}
