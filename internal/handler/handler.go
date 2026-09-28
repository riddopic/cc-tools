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
// Only fields Claude Code honors are modeled: stderr on exit code 0 reaches
// neither Claude nor the user, so anything meant to be read must go here.
type HookOutput struct {
	// SystemMessage is shown to the user and never sent to Claude.
	SystemMessage string `json:"systemMessage,omitempty"`
	// HookSpecificOutput carries context that Claude Code injects into
	// Claude's context window as a system reminder.
	HookSpecificOutput *HookSpecificOutput `json:"hookSpecificOutput,omitempty"`
}

// HookSpecificOutput is the event-scoped part of [HookOutput].
type HookSpecificOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

// ContextResponse returns a response that adds text to Claude's context for
// the given event. Every token here is billed on each later request in the
// session, so keep the text short and factual.
func ContextResponse(event, text string) *Response {
	return &Response{
		ExitCode: 0,
		Stdout: &HookOutput{
			SystemMessage: "",
			HookSpecificOutput: &HookSpecificOutput{
				HookEventName:     event,
				AdditionalContext: text,
			},
		},
		Stderr: "",
	}
}

// UserMessageResponse returns a response that shows text to the user without
// spending any of Claude's context.
func UserMessageResponse(text string) *Response {
	return &Response{
		ExitCode: 0,
		Stdout:   &HookOutput{SystemMessage: text, HookSpecificOutput: nil},
		Stderr:   "",
	}
}

// SystemMessage returns the user-facing message, or "" if there is none.
func (r *Response) SystemMessage() string {
	if r == nil || r.Stdout == nil {
		return ""
	}

	return r.Stdout.SystemMessage
}

// AdditionalContext returns the text added to Claude's context, or "".
func (r *Response) AdditionalContext() string {
	if r == nil || r.Stdout == nil || r.Stdout.HookSpecificOutput == nil {
		return ""
	}

	return r.Stdout.HookSpecificOutput.AdditionalContext
}
