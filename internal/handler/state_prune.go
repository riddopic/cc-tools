package handler

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/hookcmd"
)

// Compile-time interface check.
var _ Handler = (*StatePruneHandler)(nil)

const (
	// pruneMarkerName is touched after each prune; its mtime throttles runs.
	pruneMarkerName = ".last-prune"
	// pruneInterval is the minimum time between two prune runs.
	pruneInterval = 24 * time.Hour
	// defaultPruneMaxAgeDays applies when the config is missing or invalid.
	defaultPruneMaxAgeDays = 7
)

// prunePattern names a state subdirectory and the file prefix the handler
// that owns it writes there.
type prunePattern struct {
	dir    string
	prefix string
}

// prunePatterns lists the only per-session state files the prune may remove.
// Matching is on prefix alone so a change of file extension stays covered.
func prunePatterns() []prunePattern {
	return []prunePattern{
		{dir: "drift", prefix: "drift-"},
		{dir: "stop", prefix: "stop-"},
		{dir: "compact", prefix: "cc-tools-compact-"},
	}
}

// StatePruneOption configures a StatePruneHandler.
type StatePruneOption func(*StatePruneHandler)

// WithPruneCacheDir overrides the cache root, which defaults to
// ~/.cache/cc-tools.
func WithPruneCacheDir(dir string) StatePruneOption {
	return func(h *StatePruneHandler) {
		h.cacheDir = dir
	}
}

// WithPruneNow overrides the clock used for age and throttle checks.
func WithPruneNow(now func() time.Time) StatePruneOption {
	return func(h *StatePruneHandler) {
		h.now = now
	}
}

// StatePruneHandler removes stale per-session state files left behind by the
// drift, stop and compact handlers. It runs on SessionStart at most once a
// day and never reports anything to the user.
type StatePruneHandler struct {
	cfg      *config.Values
	cacheDir string
	now      func() time.Time
}

// NewStatePruneHandler creates a new StatePruneHandler.
func NewStatePruneHandler(cfg *config.Values, opts ...StatePruneOption) *StatePruneHandler {
	h := &StatePruneHandler{cfg: cfg, cacheDir: "", now: time.Now}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// Name returns the handler identifier.
func (h *StatePruneHandler) Name() string { return "state-prune" }

// Handle prunes stale state files when the last prune is more than a day old.
// Pruning is housekeeping, so every failure is ignored.
func (h *StatePruneHandler) Handle(_ context.Context, input *hookcmd.HookInput) (*Response, error) {
	cacheDir := h.resolveCacheDir()
	if cacheDir == "" {
		return &Response{ExitCode: 0}, nil
	}

	now := h.now()
	marker := filepath.Join(cacheDir, pruneMarkerName)
	if !isDir(cacheDir) || prunedRecently(marker, now) {
		return &Response{ExitCode: 0}, nil
	}

	cutoff := now.Add(-time.Duration(h.maxAgeDays()) * 24 * time.Hour)
	keep := input.SessionID.FileKey()
	for _, p := range prunePatterns() {
		pruneDir(filepath.Join(cacheDir, p.dir), p.prefix, keep, cutoff)
	}

	touchMarker(marker, now)
	return &Response{ExitCode: 0}, nil
}

func (h *StatePruneHandler) resolveCacheDir() string {
	if h.cacheDir != "" {
		return h.cacheDir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".cache", "cc-tools")
}

func (h *StatePruneHandler) maxAgeDays() int {
	if h.cfg == nil || h.cfg.State.MaxAgeDays <= 0 {
		return defaultPruneMaxAgeDays
	}
	return h.cfg.State.MaxAgeDays
}

// isDir reports whether path is a real directory, not a symlink to one.
func isDir(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.IsDir()
}

// prunedRecently reports whether the marker was touched within pruneInterval.
func prunedRecently(marker string, now time.Time) bool {
	info, err := os.Lstat(marker)
	return err == nil && now.Sub(info.ModTime()) < pruneInterval
}

// pruneDir removes regular files in dir whose name starts with prefix, does
// not contain keep, and was last modified before cutoff. A dir that is itself
// a symlink is skipped, and directory entries are never followed.
func pruneDir(dir, prefix, keep string, cutoff time.Time) {
	if !isDir(dir) {
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !isPrunable(entry, prefix, keep, cutoff) {
			continue
		}
		_ = os.Remove(filepath.Join(dir, entry.Name()))
	}
}

func isPrunable(entry os.DirEntry, prefix, keep string, cutoff time.Time) bool {
	name := entry.Name()
	if !entry.Type().IsRegular() || !strings.HasPrefix(name, prefix) {
		return false
	}
	if keep != "" && strings.Contains(name, keep) {
		return false
	}
	info, err := entry.Info()
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	return info.ModTime().Before(cutoff)
}

// touchMarker creates the marker if needed and sets its mtime to now.
func touchMarker(path string, now time.Time) {
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		return
	}
	_ = os.Chtimes(path, now, now)
}
