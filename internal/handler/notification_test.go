package handler_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
)

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

func TestNotifyDesktopHandler_ImplementsHandler(t *testing.T) {
	t.Parallel()
	var _ handler.Handler = handler.NewNotifyDesktopHandler(nil)
}
