package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/riddopic/cc-tools/internal/pkgmanager"
	"github.com/riddopic/cc-tools/internal/session"
	"github.com/riddopic/cc-tools/internal/superpowers"
)

// Compile-time interface checks.
var (
	_ Handler = (*SuperpowersHandler)(nil)
	_ Handler = (*PkgManagerHandler)(nil)
	_ Handler = (*SessionContextHandler)(nil)
)

// ---------------------------------------------------------------------
// SuperpowersHandler
// ---------------------------------------------------------------------

// SuperpowersHandler injects superpowers system message on session start.
type SuperpowersHandler struct{}

// NewSuperpowersHandler creates a new SuperpowersHandler.
func NewSuperpowersHandler() *SuperpowersHandler { return &SuperpowersHandler{} }

// Name returns the handler identifier.
func (h *SuperpowersHandler) Name() string { return "superpowers" }

// Handle runs the superpowers injector and returns hookSpecificOutput if a
// skill file is present.
func (h *SuperpowersHandler) Handle(ctx context.Context, input *hookcmd.HookInput) (*Response, error) {
	var buf bytes.Buffer

	if err := superpowers.NewInjector(input.Cwd).Run(ctx, &buf); err != nil {
		return nil, fmt.Errorf("inject superpowers: %w", err)
	}

	if buf.Len() == 0 {
		return &Response{ExitCode: 0}, nil
	}

	// The injector writes JSON with a hookSpecificOutput field that maps
	// directly to the HookOutput struct. Unmarshal to preserve fidelity.
	var out HookOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		return nil, fmt.Errorf("parse superpowers output: %w", err)
	}

	return &Response{
		ExitCode: 0,
		Stdout:   &out,
	}, nil
}

// ---------------------------------------------------------------------
// PkgManagerHandler
// ---------------------------------------------------------------------

// PkgManagerHandler detects the package manager and writes to .claude/.env.
type PkgManagerHandler struct {
	cfg *config.Values
}

// NewPkgManagerHandler creates a new PkgManagerHandler.
func NewPkgManagerHandler(cfg *config.Values) *PkgManagerHandler {
	return &PkgManagerHandler{cfg: cfg}
}

// Name returns the handler identifier.
func (h *PkgManagerHandler) Name() string { return "pkg-manager" }

// Handle detects the project's package manager and persists it in the
// .claude/.env file so it is available to Bash commands during the session.
func (h *PkgManagerHandler) Handle(_ context.Context, input *hookcmd.HookInput) (*Response, error) {
	var preferred string
	if h.cfg != nil {
		preferred = h.cfg.PackageManager.Preferred
	}
	manager := pkgmanager.DetectWithPreferred(input.Cwd, preferred)

	envDir := filepath.Join(input.Cwd, ".claude")
	if err := os.MkdirAll(envDir, 0o750); err != nil {
		return nil, fmt.Errorf("create .claude directory: %w", err)
	}

	envFile := filepath.Join(envDir, ".env")
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
// context as additional context and alias info on stderr.
func (h *SessionContextHandler) Handle(_ context.Context, _ *hookcmd.HookInput) (*Response, error) {
	homeDir := h.homeDir
	if homeDir == "" {
		var err error

		homeDir, err = os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home directory: %w", err)
		}
	}

	storeDir := filepath.Join(homeDir, ".claude", "sessions")
	store := session.NewStore(storeDir)

	sessions, _ := store.List(1)
	if len(sessions) == 0 {
		return &Response{ExitCode: 0}, nil
	}

	var additionalCtx []string

	latest := sessions[0]
	if latest.Summary != "" {
		additionalCtx = append(additionalCtx,
			fmt.Sprintf("Previous session (%s): %s", latest.Date, latest.Summary))
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
	if len(additionalCtx) > 0 {
		resp.Stdout = &HookOutput{
			Continue:          true,
			AdditionalContext: additionalCtx,
		}
	}

	return resp, nil
}
