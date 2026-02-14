// Package handler provides hook event handlers for the cc-tools CLI.
//
// Each handler processes a Claude Code hook event and returns a structured
// [Response] that maps to the hooks JSON output protocol.
package handler

import (
	"context"

	"github.com/riddopic/cc-tools/internal/hookcmd"
)

// Handler processes a hook event and returns a structured response.
type Handler interface {
	// Name returns a short identifier for logging and debugging.
	Name() string
	// Handle processes the hook event. It returns nil Response to indicate
	// no output (different from &Response{} which outputs exit code 0).
	Handle(ctx context.Context, input *hookcmd.HookInput) (*Response, error)
}

// Response captures a handler's output for the Claude Code hooks protocol.
// Exit code 0 = success, 2 = block with stderr feedback.
type Response struct {
	ExitCode int
	Stdout   *HookOutput
	Stderr   string
}

// HookOutput is the JSON written to stdout per the Claude Code hooks protocol.
type HookOutput struct {
	Continue           bool           `json:"continue,omitempty"`
	StopReason         string         `json:"stopReason,omitempty"`
	SuppressOutput     bool           `json:"suppressOutput,omitempty"`
	SystemMessage      string         `json:"systemMessage,omitempty"`
	HookSpecificOutput map[string]any `json:"hookSpecificOutput,omitempty"`
	AdditionalContext  []string       `json:"additionalContext,omitempty"`
	PermissionDecision string         `json:"permissionDecision,omitempty"`
	UpdatedInput       map[string]any `json:"updatedInput,omitempty"`
}
