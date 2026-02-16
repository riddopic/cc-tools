package handler

import (
	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/riddopic/cc-tools/internal/notify"
)

// NewDefaultRegistry creates a registry with all default handlers wired.
func NewDefaultRegistry(cfg *config.Values) *Registry {
	r := NewRegistry()

	r.Register(hookcmd.EventSessionStart,
		NewSuperpowersHandler(),
		NewPkgManagerHandler(),
		NewSessionContextHandler(),
	)

	r.Register(hookcmd.EventSessionEnd,
		NewSessionEndHandler(cfg),
	)

	r.Register(hookcmd.EventPreToolUse,
		NewSuggestCompactHandler(cfg),
		NewObserveHandler(cfg, "pre"),
		NewPreCommitReminderHandler(cfg),
	)

	r.Register(hookcmd.EventPostToolUse,
		NewObserveHandler(cfg, "post"),
	)

	r.Register(hookcmd.EventPostToolUseFailure,
		NewObserveHandler(cfg, "failure"),
	)

	r.Register(hookcmd.EventPreCompact,
		NewLogCompactionHandler(),
	)

	r.Register(hookcmd.EventNotification,
		NewNotifyAudioHandler(cfg, WithAudioPlayer(&notify.AFPlayer{})),
		NewNotifyDesktopHandler(cfg, WithCmdRunner(&notify.OSRunner{})),
		NewNotifyNtfyHandler(cfg),
	)

	return r
}

// HasHandlers reports whether the registry has handlers for the given event.
func (r *Registry) HasHandlers(event string) bool {
	return len(r.handlers[event]) > 0
}
