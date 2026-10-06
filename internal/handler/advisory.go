package handler

import (
	"context"
	"os"
	"strings"

	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/riddopic/cc-tools/internal/shared"
	"github.com/riddopic/cc-tools/internal/skipregistry"
)

// Environment variables that mark a session as unattended. AMS_UNATTENDED is
// honored so fleet runners need only one flag.
const (
	envUnattended    = "CC_TOOLS_UNATTENDED"
	envAMSUnattended = "AMS_UNATTENDED"
)

// AdvisoryOption configures an [AdvisoryHandler].
type AdvisoryOption func(*AdvisoryHandler)

// WithAdvisoryGetenv overrides how environment variables are read, which
// otherwise uses [os.Getenv].
func WithAdvisoryGetenv(getenv func(string) string) AdvisoryOption {
	return func(h *AdvisoryHandler) {
		h.getenv = getenv
	}
}

// WithAdvisorySkipReader overrides the skip registry consulted for the
// nudges skip type, which otherwise is the user's registry file.
func WithAdvisorySkipReader(reader skipregistry.Reader) AdvisoryOption {
	return func(h *AdvisoryHandler) {
		h.skipReader = reader
	}
}

// AdvisoryHandler wraps a handler whose output is advice for a person at the
// keyboard, and drops that output when the session is unattended. The inner
// handler always runs, so its counters and logs keep advancing.
type AdvisoryHandler struct {
	inner      Handler
	getenv     func(string) string
	skipReader skipregistry.Reader
}

// NewAdvisoryHandler wraps inner so its nudges are silenced in unattended
// sessions: loops, fleet runners and subagents.
func NewAdvisoryHandler(inner Handler, opts ...AdvisoryOption) *AdvisoryHandler {
	h := &AdvisoryHandler{inner: inner, getenv: os.Getenv, skipReader: nil}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// Name returns the wrapped handler's name.
func (h *AdvisoryHandler) Name() string { return h.inner.Name() }

// Handle runs the inner handler and suppresses its user-facing output when
// the session is unattended. Errors and non-zero exit codes pass through.
func (h *AdvisoryHandler) Handle(ctx context.Context, input *hookcmd.HookInput) (*Response, error) {
	resp, err := h.inner.Handle(ctx, input)
	if err != nil || !hasAdvice(resp) {
		return resp, err
	}

	if !h.unattended(ctx, input) {
		return resp, nil
	}

	return &Response{ExitCode: 0, Stdout: nil, Stderr: ""}, nil
}

// hasAdvice reports whether resp carries output a person or Claude would read.
func hasAdvice(resp *Response) bool {
	if resp == nil || resp.ExitCode != 0 {
		return false
	}
	return resp.SystemMessage() != "" || resp.AdditionalContext() != ""
}

// unattended checks the triggers cheapest first: environment, then subagent
// fields in the payload, then the skip registry.
func (h *AdvisoryHandler) unattended(ctx context.Context, input *hookcmd.HookInput) bool {
	if isTruthy(h.getenv(envUnattended)) || isTruthy(h.getenv(envAMSUnattended)) {
		return true
	}

	if input == nil {
		return false
	}
	if input.AgentID != "" || input.AgentType != "" {
		return true
	}

	return h.nudgesSkipped(ctx, input.Cwd)
}

// nudgesSkipped reports whether the skip registry silences nudges for cwd or
// an ancestor up to the repository root. Lookup failures count as attended.
func (h *AdvisoryHandler) nudgesSkipped(ctx context.Context, cwd string) bool {
	if cwd == "" {
		return false
	}

	root, err := shared.FindRepoRoot(cwd, nil)
	if err != nil {
		root = cwd
	}

	reader := h.skipReader
	if reader == nil {
		reader = skipregistry.NewRegistry(skipregistry.DefaultStorage())
	}

	skips := skipregistry.Effective(
		ctx, reader, skipregistry.DirectoryPath(cwd), skipregistry.DirectoryPath(root),
	)
	return skips.Nudges.Skipped
}

// isTruthy reports whether an environment value enables a flag.
func isTruthy(value string) bool {
	v := strings.TrimSpace(value)
	return v == "1" || strings.EqualFold(v, "true")
}
