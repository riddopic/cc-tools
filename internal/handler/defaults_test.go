package handler_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
)

func TestNewDefaultRegistry_RegistersAllEvents(t *testing.T) {
	t.Parallel()

	cfg := config.GetDefaultConfig()
	r := handler.NewDefaultRegistry(cfg)

	// Verify handlers registered for expected events.
	assert.True(t, r.HasHandlers(hookcmd.EventSessionStart))
	assert.True(t, r.HasHandlers(hookcmd.EventSessionEnd))
	assert.True(t, r.HasHandlers(hookcmd.EventPreToolUse))
	assert.True(t, r.HasHandlers(hookcmd.EventPostToolUse))
	assert.True(t, r.HasHandlers(hookcmd.EventPreCompact))
	assert.True(t, r.HasHandlers(hookcmd.EventUserPromptSubmit))
	assert.True(t, r.HasHandlers(hookcmd.EventStop))
	assert.True(t, r.HasHandlers(hookcmd.EventNotification))
}

// The stop reminder is wrapped as advisory: a subagent's Stop event still
// counts but shows no reminder, while the main session sees it.
func TestNewDefaultRegistry_StopReminderSilencedForSubagents(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("CC_TOOLS_UNATTENDED", "")
	t.Setenv("AMS_UNATTENDED", "")

	cfg := config.GetDefaultConfig()
	cfg.StopReminder.Enabled = true
	cfg.StopReminder.Interval = 1
	r := handler.NewDefaultRegistry(cfg)

	subagent := r.Dispatch(context.Background(), &hookcmd.HookInput{
		HookEventName: hookcmd.EventStop,
		SessionID:     "sub-session",
		AgentID:       "a-1",
	})
	assert.Empty(t, subagent.SystemMessage(), "subagent stop must stay silent")

	interactive := r.Dispatch(context.Background(), &hookcmd.HookInput{
		HookEventName: hookcmd.EventStop,
		SessionID:     "main-session",
	})
	assert.NotEmpty(t, interactive.SystemMessage(), "interactive stop still gets the reminder")
}

func TestNewDefaultRegistry_NilConfig(t *testing.T) {
	t.Parallel()

	r := handler.NewDefaultRegistry(nil)
	assert.True(t, r.HasHandlers(hookcmd.EventSessionStart))
}
