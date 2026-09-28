package session

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	// maxSummaryRequest caps the quoted opening request, in runes.
	maxSummaryRequest = 150
	// maxSummaryFiles caps how many modified files are named.
	maxSummaryFiles = 5
	// maxSummary caps the whole summary, in runes. It is injected into
	// Claude's context at the start of later sessions, so it must stay small.
	maxSummary = 400
)

// BuildSummary renders a short, factual recap of a session: the user's
// opening request and the files it modified, with paths relative to cwd when
// they fall inside it. It returns "" when the session had no typed prompt.
func BuildSummary(ts *TranscriptSummary, cwd string) string {
	if ts == nil {
		return ""
	}

	request := truncateRunes(strings.Join(strings.Fields(ts.FirstPrompt), " "), maxSummaryRequest)
	if request == "" {
		return ""
	}

	summary := fmt.Sprintf("Request: %q.", request)
	if files := fileList(ts.FilesModified, cwd); files != "" {
		summary += " Files modified: " + files + "."
	}

	return truncateRunes(summary, maxSummary)
}

// fileList names up to maxSummaryFiles files and counts the rest.
func fileList(paths []string, cwd string) string {
	if len(paths) == 0 {
		return ""
	}

	shown := paths[:min(len(paths), maxSummaryFiles)]
	names := make([]string, 0, len(shown)+1)
	for _, p := range shown {
		names = append(names, displayPath(p, cwd))
	}

	if extra := len(paths) - len(shown); extra > 0 {
		names = append(names, fmt.Sprintf("+%d more", extra))
	}

	return strings.Join(names, ", ")
}

// displayPath shortens p to a cwd-relative path, or its base name when it
// lies outside cwd.
func displayPath(p, cwd string) string {
	if cwd != "" {
		if rel, err := filepath.Rel(cwd, p); err == nil && !strings.HasPrefix(rel, "..") {
			return rel
		}
	}

	return filepath.Base(p)
}

// truncateRunes shortens s to at most limit runes, ending with an ellipsis
// when it cuts.
func truncateRunes(s string, limit int) string {
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}

	return string(runes[:limit-1]) + "…"
}
