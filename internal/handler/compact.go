package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/riddopic/cc-tools/internal/compact"
	"github.com/riddopic/cc-tools/internal/hookcmd"
)

// Compile-time interface check.
var _ Handler = (*LogCompactionHandler)(nil)

// LogCompactionOption configures a LogCompactionHandler.
type LogCompactionOption func(*LogCompactionHandler)

// WithCompactLogDir overrides the log directory for testing.
func WithCompactLogDir(dir string) LogCompactionOption {
	return func(h *LogCompactionHandler) {
		h.logDir = dir
	}
}

// LogCompactionHandler logs PreCompact events for debugging.
type LogCompactionHandler struct {
	logDir string
}

// NewLogCompactionHandler creates a new LogCompactionHandler.
func NewLogCompactionHandler(opts ...LogCompactionOption) *LogCompactionHandler {
	h := &LogCompactionHandler{
		logDir: "",
	}
	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Name returns the handler identifier.
func (h *LogCompactionHandler) Name() string { return "log-compaction" }

// Handle logs the compaction event to a timestamped log file.
func (h *LogCompactionHandler) Handle(_ context.Context, _ *hookcmd.HookInput) (*Response, error) {
	logDir := h.logDir
	if logDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home directory: %w", err)
		}

		logDir = filepath.Join(homeDir, ".cache", "cc-tools")
	}

	if err := compact.LogCompaction(logDir); err != nil {
		return nil, fmt.Errorf("log compaction: %w", err)
	}

	return &Response{ExitCode: 0}, nil
}
