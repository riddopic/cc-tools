package handler

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/riddopic/cc-tools/internal/notify"
)

// Compile-time interface checks.
var (
	_ Handler = (*NotifyAudioHandler)(nil)
	_ Handler = (*NotifyDesktopHandler)(nil)
	_ Handler = (*NotifyNtfyHandler)(nil)
)

// AudioPlayer abstracts audio file playback for dependency injection.
type AudioPlayer interface {
	Play(filepath string) error
}

// CmdRunner abstracts command execution for dependency injection.
type CmdRunner interface {
	Run(name string, args ...string) error
}

// NtfySender abstracts ntfy notification sending for dependency injection.
type NtfySender interface {
	Send(ctx context.Context, title, message string) error
}

// ---------------------------------------------------------------------
// NotifyAudioHandler
// ---------------------------------------------------------------------

// NotifyAudioOption configures a NotifyAudioHandler.
type NotifyAudioOption func(*NotifyAudioHandler)

// WithAudioPlayer overrides the audio player for testing.
func WithAudioPlayer(player AudioPlayer) NotifyAudioOption {
	return func(h *NotifyAudioHandler) {
		h.player = player
	}
}

// NotifyAudioHandler plays an audio notification sound.
type NotifyAudioHandler struct {
	cfg    *config.Values
	player AudioPlayer
}

// NewNotifyAudioHandler creates a new NotifyAudioHandler.
func NewNotifyAudioHandler(
	cfg *config.Values,
	opts ...NotifyAudioOption,
) *NotifyAudioHandler {
	h := &NotifyAudioHandler{
		cfg:    cfg,
		player: nil,
	}
	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Name returns the handler identifier.
func (h *NotifyAudioHandler) Name() string { return "notify-audio" }

// Handle plays a random audio notification if audio is enabled and quiet
// hours are not active.
func (h *NotifyAudioHandler) Handle(
	_ context.Context,
	_ *hookcmd.HookInput,
) (*Response, error) {
	if h.cfg == nil || !h.cfg.Notify.Audio.Enabled {
		return &Response{ExitCode: 0}, nil
	}

	player := h.player
	if player == nil {
		return &Response{ExitCode: 0}, nil
	}

	dir := expandHome(h.cfg.Notify.Audio.Directory)

	// Skip silently when the audio directory does not exist.
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return &Response{ExitCode: 0}, nil
	}

	qh := notify.QuietHours{
		Enabled: h.cfg.Notify.QuietHours.Enabled,
		Start:   h.cfg.Notify.QuietHours.Start,
		End:     h.cfg.Notify.QuietHours.End,
	}

	audio := notify.NewAudio(player, dir, qh, nil)
	if err := audio.PlayRandom(); err != nil {
		return nil, err
	}

	return &Response{ExitCode: 0}, nil
}

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	return filepath.Join(home, strings.TrimPrefix(path, "~"))
}

// ---------------------------------------------------------------------
// NotifyDesktopHandler
// ---------------------------------------------------------------------

// NotifyDesktopOption configures a NotifyDesktopHandler.
type NotifyDesktopOption func(*NotifyDesktopHandler)

// WithCmdRunner overrides the command runner for testing.
func WithCmdRunner(runner CmdRunner) NotifyDesktopOption {
	return func(h *NotifyDesktopHandler) {
		h.runner = runner
	}
}

// NotifyDesktopHandler sends a desktop notification.
type NotifyDesktopHandler struct {
	cfg    *config.Values
	runner CmdRunner
}

// NewNotifyDesktopHandler creates a new NotifyDesktopHandler.
func NewNotifyDesktopHandler(
	cfg *config.Values,
	opts ...NotifyDesktopOption,
) *NotifyDesktopHandler {
	h := &NotifyDesktopHandler{
		cfg:    cfg,
		runner: nil,
	}
	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Name returns the handler identifier.
func (h *NotifyDesktopHandler) Name() string { return "notify-desktop" }

// Handle sends a desktop notification if desktop notifications are enabled
// and quiet hours are not active.
func (h *NotifyDesktopHandler) Handle(
	_ context.Context,
	input *hookcmd.HookInput,
) (*Response, error) {
	if h.cfg == nil || !h.cfg.Notify.Desktop.Enabled {
		return &Response{ExitCode: 0}, nil
	}

	qh := notify.QuietHours{
		Enabled: h.cfg.Notify.QuietHours.Enabled,
		Start:   h.cfg.Notify.QuietHours.Start,
		End:     h.cfg.Notify.QuietHours.End,
	}

	if qh.IsActive(time.Now()) {
		return &Response{ExitCode: 0}, nil
	}

	runner := h.runner
	if runner == nil {
		return &Response{ExitCode: 0}, nil
	}

	desktop := notify.NewDesktop(runner)

	title := "Claude Code"
	message := "Task completed"

	if input.Title != "" {
		title = input.Title
	}

	if input.Message != "" {
		message = input.Message
	}

	if err := desktop.Send(title, message); err != nil {
		return nil, err
	}

	return &Response{ExitCode: 0}, nil
}

// ---------------------------------------------------------------------
// NotifyNtfyHandler
// ---------------------------------------------------------------------

// NotifyNtfyOption configures a NotifyNtfyHandler.
type NotifyNtfyOption func(*NotifyNtfyHandler)

// WithNtfySender overrides the ntfy sender for testing.
func WithNtfySender(sender NtfySender) NotifyNtfyOption {
	return func(h *NotifyNtfyHandler) {
		h.sender = sender
	}
}

// NotifyNtfyHandler sends a push notification via ntfy.
type NotifyNtfyHandler struct {
	cfg    *config.Values
	sender NtfySender
}

// NewNotifyNtfyHandler creates a new NotifyNtfyHandler.
func NewNotifyNtfyHandler(
	cfg *config.Values,
	opts ...NotifyNtfyOption,
) *NotifyNtfyHandler {
	h := &NotifyNtfyHandler{
		cfg:    cfg,
		sender: nil,
	}
	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Name returns the handler identifier.
func (h *NotifyNtfyHandler) Name() string { return "notify-ntfy" }

// Handle sends a push notification via ntfy if a topic is configured
// and quiet hours are not active.
func (h *NotifyNtfyHandler) Handle(
	ctx context.Context,
	input *hookcmd.HookInput,
) (*Response, error) {
	if h.cfg == nil || h.cfg.Notifications.NtfyTopic == "" {
		return &Response{ExitCode: 0}, nil
	}

	qh := notify.QuietHours{
		Enabled: h.cfg.Notify.QuietHours.Enabled,
		Start:   h.cfg.Notify.QuietHours.Start,
		End:     h.cfg.Notify.QuietHours.End,
	}

	if qh.IsActive(time.Now()) {
		return &Response{ExitCode: 0}, nil
	}

	sender := h.sender
	if sender == nil {
		sender = notify.NewNtfyNotifier(notify.NtfyConfig{
			Topic:    h.cfg.Notifications.NtfyTopic,
			Server:   "",
			Token:    "",
			Priority: 0,
		})
	}

	title := "Claude Code"
	message := "Task completed"

	if input.Title != "" {
		title = input.Title
	}

	if input.Message != "" {
		message = input.Message
	}

	if err := sender.Send(ctx, title, message); err != nil {
		return nil, err
	}

	return &Response{ExitCode: 0}, nil
}
