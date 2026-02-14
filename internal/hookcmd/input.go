// Package hookcmd dispatches Claude Code hook events to registered handlers.
package hookcmd

import (
	"encoding/json"
	"fmt"
	"io"
)

// HookInput represents the JSON input from Claude Code hooks.
type HookInput struct {
	// Common fields (present on ALL events).
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
	PermissionMode string `json:"permission_mode"`
	HookEventName  string `json:"hook_event_name"`

	// Tool events (PreToolUse, PostToolUse, PostToolUseFailure).
	ToolName   string          `json:"tool_name,omitempty"`
	ToolInput  json.RawMessage `json:"tool_input,omitempty"`
	ToolOutput json.RawMessage `json:"tool_response,omitempty"`
	ToolUseID  string          `json:"tool_use_id,omitempty"`

	// PostToolUseFailure specific.
	Error       string `json:"error,omitempty"`
	IsInterrupt bool   `json:"is_interrupt,omitempty"`

	// SessionStart specific.
	Source string `json:"source,omitempty"`
	Model  string `json:"model,omitempty"`

	// SessionEnd specific.
	Reason string `json:"reason,omitempty"`

	// Stop / SubagentStop specific.
	StopHookActive bool `json:"stop_hook_active,omitempty"`

	// UserPromptSubmit specific.
	Prompt string `json:"prompt,omitempty"`

	// Notification specific.
	Message          string `json:"message,omitempty"`
	Title            string `json:"title,omitempty"`
	NotificationType string `json:"notification_type,omitempty"`

	// PreCompact specific.
	Trigger            string `json:"trigger,omitempty"`
	CustomInstructions string `json:"custom_instructions,omitempty"`
}

// ParseInput reads JSON from the given reader and parses it into [HookInput].
func ParseInput(r io.Reader) (*HookInput, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	if len(data) == 0 {
		return &HookInput{}, nil
	}

	var input HookInput
	if unmarshalErr := json.Unmarshal(data, &input); unmarshalErr != nil {
		return nil, fmt.Errorf("parsing hook input JSON: %w", unmarshalErr)
	}

	return &input, nil
}

// GetToolInputString extracts a string field from tool_input JSON.
func (h *HookInput) GetToolInputString(key string) string {
	if len(h.ToolInput) == 0 {
		return ""
	}

	var m map[string]any
	if err := json.Unmarshal(h.ToolInput, &m); err != nil {
		return ""
	}

	if v, ok := m[key].(string); ok {
		return v
	}

	return ""
}
