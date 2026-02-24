// Package hookcmd dispatches Claude Code hook events to registered handlers.
package hookcmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
)

// SessionID is a typed wrapper for Claude Code session identifiers.
type SessionID string

// String returns the raw session ID value.
func (id SessionID) String() string { return string(id) }

// IsEmpty reports whether the session ID is empty.
func (id SessionID) IsEmpty() bool { return id == "" }

// FileKey returns a filesystem-safe key derived from the session ID.
// IDs matching ^[a-zA-Z0-9-]+$ pass through unchanged.
// All other IDs are hashed to a 16-character hex string.
func (id SessionID) FileKey() string {
	if id == "" {
		return ""
	}

	if safeIDPattern.MatchString(string(id)) {
		return string(id)
	}

	h := sha256.Sum256([]byte(id))

	return hex.EncodeToString(h[:])[:16]
}

// HookInput represents the JSON input from Claude Code hooks.
type HookInput struct {
	// Common fields (present on ALL events).
	SessionID      SessionID `json:"session_id"`
	TranscriptPath string    `json:"transcript_path"`
	Cwd            string    `json:"cwd"`
	PermissionMode string    `json:"permission_mode"`
	HookEventName  string    `json:"hook_event_name"`

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

// IsEditTool reports whether the hook event involves a file-editing tool.
func (h *HookInput) IsEditTool() bool {
	switch h.ToolName {
	case "Edit", "MultiEdit", "Write", "NotebookEdit":
		return true
	default:
		return false
	}
}

// GetFilePath extracts the target file path from tool_input JSON.
// For NotebookEdit it reads "notebook_path"; for all other tools it reads "file_path".
// It returns an empty string when ToolInput is empty, invalid, or the field is absent.
func (h *HookInput) GetFilePath() string {
	key := "file_path"
	if h.ToolName == "NotebookEdit" {
		key = "notebook_path"
	}

	return h.GetToolInputString(key)
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

// safeIDPattern matches session IDs that contain only alphanumeric characters
// and hyphens. Claude Code produces UUID-format session IDs, so this covers
// all legitimate values while rejecting dots, underscores, glob metacharacters,
// and path separators.
var safeIDPattern = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
