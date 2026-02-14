package handler_test

import (
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
	assert.True(t, r.HasHandlers(hookcmd.EventNotification))
}

func TestNewDefaultRegistry_NilConfig(t *testing.T) {
	t.Parallel()

	r := handler.NewDefaultRegistry(nil)
	assert.True(t, r.HasHandlers(hookcmd.EventSessionStart))
}
