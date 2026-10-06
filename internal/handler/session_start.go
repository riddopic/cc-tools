package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/riddopic/cc-tools/internal/pkgmanager"
	"github.com/riddopic/cc-tools/internal/session"
)

// Compile-time interface checks.
var (
	_ Handler = (*PkgManagerHandler)(nil)
	_ Handler = (*SessionContextHandler)(nil)
)

// ---------------------------------------------------------------------
// PkgManagerHandler
// ---------------------------------------------------------------------

// PkgManagerOption configures a PkgManagerHandler.
type PkgManagerOption func(*PkgManagerHandler)

// WithEnvFilePath overrides the env file path, which otherwise comes from the
// CLAUDE_ENV_FILE variable Claude Code sets for SessionStart hooks.
func WithEnvFilePath(path string) PkgManagerOption {
	return func(h *PkgManagerHandler) {
		h.envFile = path
		h.envFileSet = true
	}
}

// WithPkgManagerGetenv overrides how PREFERRED_PACKAGE_MANAGER is read during
// detection, which otherwise uses [os.Getenv].
func WithPkgManagerGetenv(getenv func(string) string) PkgManagerOption {
	return func(h *PkgManagerHandler) {
		h.getenv = getenv
	}
}

// PkgManagerHandler detects the package manager and exports it to the
// session's Bash environment.
type PkgManagerHandler struct {
	cfg        *config.Values
	envFile    string
	envFileSet bool
	getenv     func(string) string
}

// NewPkgManagerHandler creates a new PkgManagerHandler.
func NewPkgManagerHandler(cfg *config.Values, opts ...PkgManagerOption) *PkgManagerHandler {
	h := &PkgManagerHandler{cfg: cfg, envFile: "", envFileSet: false, getenv: os.Getenv}
	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Name returns the handler identifier.
func (h *PkgManagerHandler) Name() string { return "pkg-manager" }

// Handle detects the project's package manager and appends it to the file
// named by CLAUDE_ENV_FILE, which Claude Code sources before each Bash
// command. Without that file there is nowhere the value would be read from,
// so the handler does nothing.
func (h *PkgManagerHandler) Handle(_ context.Context, input *hookcmd.HookInput) (*Response, error) {
	envFile := h.envFile
	if !h.envFileSet {
		envFile = os.Getenv("CLAUDE_ENV_FILE")
	}

	if envFile == "" {
		return &Response{ExitCode: 0}, nil
	}

	manager := pkgmanager.DetectWithEnv(input.Cwd, h.getenv)
	if h.cfg != nil && h.cfg.PackageManager.Preferred != "" {
		manager = h.cfg.PackageManager.Preferred
	}

	if err := pkgmanager.WriteToEnvFile(envFile, manager); err != nil {
		return nil, fmt.Errorf("write env file: %w", err)
	}

	return &Response{ExitCode: 0}, nil
}

// ---------------------------------------------------------------------
// SessionContextHandler
// ---------------------------------------------------------------------

// SessionContextOption configures a SessionContextHandler.
type SessionContextOption func(*SessionContextHandler)

// WithHomeDir overrides the home directory for testing.
func WithHomeDir(dir string) SessionContextOption {
	return func(h *SessionContextHandler) {
		h.homeDir = dir
	}
}

// SessionContextHandler provides previous session context on start.
type SessionContextHandler struct {
	homeDir string
}

// NewSessionContextHandler creates a new SessionContextHandler.
func NewSessionContextHandler(opts ...SessionContextOption) *SessionContextHandler {
	h := &SessionContextHandler{
		homeDir: "",
	}
	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Name returns the handler identifier.
func (h *SessionContextHandler) Name() string { return "session-context" }

// Handle loads the most recent session and any aliases, returning session
// context as additional context and alias info on stderr. It only runs for
// fresh contexts: on resume or fork the "previous" session is the current
// conversation, and after compaction Claude already has its own summary.
func (h *SessionContextHandler) Handle(_ context.Context, input *hookcmd.HookInput) (*Response, error) {
	if continuesTranscript(input.Source) || input.Source == "compact" {
		return &Response{ExitCode: 0}, nil
	}

	homeDir := h.homeDir
	if homeDir == "" {
		var err error

		homeDir, err = os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home directory: %w", err)
		}
	}

	store := session.NewStore(filepath.Join(homeDir, ".claude", "sessions"))

	// Only a summarized session from this project is worth Claude's context;
	// sessions from other directories, and files that predate the project
	// field, are skipped.
	latest, _ := store.Latest(func(s *session.Session) bool {
		return input.Cwd != "" && s.Cwd == input.Cwd && s.Summary != ""
	})

	var additionalCtx string
	if latest != nil {
		additionalCtx = fmt.Sprintf("Previous session in this project (%s): %s", latest.Date, latest.Summary)
	}

	var stderr string

	aliasFile := filepath.Join(homeDir, ".claude", "session-aliases.json")
	aliases := session.NewAliasManager(aliasFile)

	aliasList, aliasErr := aliases.List()
	if aliasErr == nil && len(aliasList) > 0 {
		names := make([]string, 0, len(aliasList))
		for name := range aliasList {
			names = append(names, name)
		}

		stderr = fmt.Sprintf("[session-context] %d alias(es): %s\n",
			len(aliasList), strings.Join(names, ", "))
	}

	resp := &Response{ExitCode: 0, Stderr: stderr}
	if additionalCtx != "" {
		resp = ContextResponse(hookcmd.EventSessionStart, additionalCtx)
		resp.Stderr = stderr
	}

	return resp, nil
}

// continuesTranscript reports whether a SessionStart source reopens an
// existing transcript, which already holds any context injected earlier.
func continuesTranscript(source string) bool {
	return source == "resume" || source == "fork"
}
