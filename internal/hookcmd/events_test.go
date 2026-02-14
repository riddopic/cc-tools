package hookcmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/riddopic/cc-tools/internal/hookcmd"
)

func TestEventConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"SessionStart", hookcmd.EventSessionStart, "SessionStart"},
		{"SessionEnd", hookcmd.EventSessionEnd, "SessionEnd"},
		{"PreToolUse", hookcmd.EventPreToolUse, "PreToolUse"},
		{"PostToolUse", hookcmd.EventPostToolUse, "PostToolUse"},
		{"PostToolUseFailure", hookcmd.EventPostToolUseFailure, "PostToolUseFailure"},
		{"PreCompact", hookcmd.EventPreCompact, "PreCompact"},
		{"Notification", hookcmd.EventNotification, "Notification"},
		{"UserPromptSubmit", hookcmd.EventUserPromptSubmit, "UserPromptSubmit"},
		{"PermissionRequest", hookcmd.EventPermissionRequest, "PermissionRequest"},
		{"Stop", hookcmd.EventStop, "Stop"},
		{"SubagentStart", hookcmd.EventSubagentStart, "SubagentStart"},
		{"SubagentStop", hookcmd.EventSubagentStop, "SubagentStop"},
		{"TeammateIdle", hookcmd.EventTeammateIdle, "TeammateIdle"},
		{"TaskCompleted", hookcmd.EventTaskCompleted, "TaskCompleted"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

func TestAllEvents_ReturnsAllConstants(t *testing.T) {
	t.Parallel()
	events := hookcmd.AllEvents()
	assert.Len(t, events, 14)
	assert.Contains(t, events, hookcmd.EventSessionStart)
	assert.Contains(t, events, hookcmd.EventStop)
}
