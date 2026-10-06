package handler

import (
	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/riddopic/cc-tools/internal/notify"
)

// NewDefaultRegistry creates a registry with all default handlers wired.
//
// Handlers whose output is only advice for a person at the keyboard are
// wrapped with [NewAdvisoryHandler] so they stay silent in unattended
// sessions. Blocking and validating handlers are never wrapped.
func NewDefaultRegistry(cfg *config.Values) *Registry {
	r := NewRegistry()

	r.Register(hookcmd.EventSessionStart,
		NewPkgManagerHandler(cfg),
		NewSessionContextHandler(),
		NewStatePruneHandler(cfg),
	)

	r.Register(hookcmd.EventSessionEnd,
		NewSessionEndHandler(cfg),
	)

	r.Register(hookcmd.EventPreToolUse,
		NewAdvisoryHandler(NewSuggestCompactHandler(cfg)),
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

	r.Register(hookcmd.EventUserPromptSubmit,
		NewAdvisoryHandler(NewDriftHandler(cfg)),
	)

	r.Register(hookcmd.EventStop,
		NewAdvisoryHandler(NewStopReminderHandler(cfg)),
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
