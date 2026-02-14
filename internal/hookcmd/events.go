package hookcmd

// Event name constants matching Claude Code hook event names.
// See: https://docs.anthropic.com/en/docs/claude-code/hooks
const (
	EventSessionStart       = "SessionStart"
	EventSessionEnd         = "SessionEnd"
	EventPreToolUse         = "PreToolUse"
	EventPostToolUse        = "PostToolUse"
	EventPostToolUseFailure = "PostToolUseFailure"
	EventPreCompact         = "PreCompact"
	EventNotification       = "Notification"
	EventUserPromptSubmit   = "UserPromptSubmit"
	EventPermissionRequest  = "PermissionRequest"
	EventStop               = "Stop"
	EventSubagentStart      = "SubagentStart"
	EventSubagentStop       = "SubagentStop"
	EventTeammateIdle       = "TeammateIdle"
	EventTaskCompleted      = "TaskCompleted"
)

// AllEvents returns all known hook event names.
func AllEvents() []string {
	return []string{
		EventSessionStart,
		EventSessionEnd,
		EventPreToolUse,
		EventPostToolUse,
		EventPostToolUseFailure,
		EventPreCompact,
		EventNotification,
		EventUserPromptSubmit,
		EventPermissionRequest,
		EventStop,
		EventSubagentStart,
		EventSubagentStop,
		EventTeammateIdle,
		EventTaskCompleted,
	}
}
