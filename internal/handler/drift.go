package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/hookcmd"
)

// Compile-time interface check.
var _ Handler = (*DriftHandler)(nil)

const (
	// maxIntentLen is the maximum character length for the stored intent sentence.
	maxIntentLen = 200
	// minKeywordLen is the minimum word length to qualify as a keyword.
	minKeywordLen = 3
)

// driftState persists the original session intent across prompts.
type driftState struct {
	Intent   string   `json:"intent"`
	Keywords []string `json:"keywords"`
	Edits    int      `json:"edits"`
}

// DriftOption configures a DriftHandler.
type DriftOption func(*DriftHandler)

// WithDriftStateDir overrides the state directory for testing.
func WithDriftStateDir(dir string) DriftOption {
	return func(h *DriftHandler) {
		h.stateDir = dir
	}
}

// DriftHandler detects when a session drifts away from its original intent.
// It fires on UserPromptSubmit events, tracking keywords from the first prompt
// and warning when subsequent prompts diverge significantly.
type DriftHandler struct {
	cfg      *config.Values
	stateDir string
}

// NewDriftHandler creates a new DriftHandler.
func NewDriftHandler(cfg *config.Values, opts ...DriftOption) *DriftHandler {
	h := &DriftHandler{
		cfg:      cfg,
		stateDir: "",
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// Name returns the handler identifier.
func (h *DriftHandler) Name() string { return "drift-detection" }

// Handle processes a UserPromptSubmit event, tracking intent and detecting drift.
func (h *DriftHandler) Handle(_ context.Context, input *hookcmd.HookInput) (*Response, error) {
	if h.cfg == nil || !h.cfg.Drift.Enabled {
		return &Response{ExitCode: 0}, nil
	}

	prompt := strings.TrimSpace(input.Prompt)
	if prompt == "" {
		return &Response{ExitCode: 0}, nil
	}

	stateDir := h.stateDir
	if stateDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home directory: %w", err)
		}
		stateDir = filepath.Join(homeDir, ".cache", "cc-tools", "drift")
	}

	state := h.loadState(stateDir, input.SessionID)

	// Detect explicit intent changes.
	if isPivotPhrase(prompt) {
		state = h.initIntent(prompt)
		h.saveState(stateDir, input.SessionID, state)
		return &Response{ExitCode: 0}, nil
	}

	// First prompt: establish intent baseline.
	if state.Intent == "" {
		state = h.initIntent(prompt)
		h.saveState(stateDir, input.SessionID, state)
		return &Response{ExitCode: 0}, nil
	}

	// Subsequent prompts: increment edits and check drift.
	state.Edits++
	h.saveState(stateDir, input.SessionID, state)

	minEdits := h.cfg.Drift.MinEdits
	threshold := h.cfg.Drift.Threshold

	if state.Edits < minEdits {
		return &Response{ExitCode: 0}, nil
	}

	promptKeywords := extractKeywords(prompt)
	overlap := keywordOverlap(state.Keywords, promptKeywords)

	if overlap < threshold {
		msg := fmt.Sprintf(
			"[cc-tools] Possible drift detected — current work may be unrelated to original intent: %q\n",
			state.Intent,
		)
		return &Response{ExitCode: 0, Stderr: msg}, nil
	}

	return &Response{ExitCode: 0}, nil
}

// initIntent creates a new drift state from the given prompt.
func (h *DriftHandler) initIntent(prompt string) *driftState {
	intent := firstSentence(prompt, maxIntentLen)
	return &driftState{
		Intent:   intent,
		Keywords: extractKeywords(intent),
		Edits:    0,
	}
}

func (h *DriftHandler) statePath(dir, sessionID string) string {
	return filepath.Join(dir, "drift-"+hookcmd.FileSafeSessionKey(sessionID)+".json")
}

func (h *DriftHandler) loadState(dir, sessionID string) *driftState {
	data, err := os.ReadFile(h.statePath(dir, sessionID)) // #nosec G304 -- path built from stateDir
	if err != nil {
		return &driftState{}
	}
	var state driftState
	if unmarshalErr := json.Unmarshal(data, &state); unmarshalErr != nil {
		return &driftState{}
	}
	return &state
}

func (h *DriftHandler) saveState(dir, sessionID string, state *driftState) {
	_ = os.MkdirAll(dir, 0o750)
	data, err := json.Marshal(state)
	if err != nil {
		return
	}
	_ = os.WriteFile(h.statePath(dir, sessionID), data, 0o600)
}

// firstSentence extracts the first sentence from text, up to maxLen characters.
func firstSentence(text string, maxLen int) string {
	// Find end of first sentence.
	for i, ch := range text {
		if i >= maxLen {
			return text[:maxLen]
		}
		if ch == '.' || ch == '!' || ch == '?' {
			return text[:i+1]
		}
		if ch == '\n' {
			return text[:i]
		}
	}
	if len(text) > maxLen {
		return text[:maxLen]
	}
	return text
}

// extractKeywords splits text into lowercase words, removes stop words, and
// filters words shorter than 3 characters.
func extractKeywords(text string) []string {
	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	var keywords []string
	for _, w := range words {
		if len(w) < minKeywordLen {
			continue
		}
		if isStopWord(w) {
			continue
		}
		keywords = append(keywords, w)
	}
	return keywords
}

// keywordOverlap returns the ratio of prompt keywords found in the intent keywords.
func keywordOverlap(intentKW, promptKW []string) float64 {
	if len(promptKW) == 0 {
		return 1.0 // Empty prompt doesn't indicate drift.
	}
	intentSet := make(map[string]struct{}, len(intentKW))
	for _, kw := range intentKW {
		intentSet[kw] = struct{}{}
	}
	matches := 0
	for _, kw := range promptKW {
		if _, ok := intentSet[kw]; ok {
			matches++
		}
	}
	return float64(matches) / float64(len(promptKW))
}

// isPivotPhrase returns true if the prompt starts with an explicit intent change.
func isPivotPhrase(prompt string) bool {
	lower := strings.ToLower(prompt)
	for _, prefix := range pivotPrefixes() {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}

func pivotPrefixes() []string {
	return []string{
		"now let's", "now lets", "switch to", "actually,", "actually ",
		"new task", "forget that", "let's switch", "lets switch",
		"change of plan", "different topic",
	}
}

// isStopWord reports whether w is a common English stop word.
func isStopWord(w string) bool {
	switch w {
	case "the", "and", "for", "are", "but", "not",
		"you", "all", "can", "her", "was", "one",
		"our", "out", "has", "had", "its", "let",
		"say", "she", "too", "use", "how", "man",
		"did", "get", "may", "him", "old", "see",
		"now", "way", "who", "any", "new", "got",
		"been", "that", "this", "with", "have", "from",
		"they", "will", "when", "make", "like", "just",
		"over", "also", "into", "some", "than", "them",
		"want", "give", "most", "only", "what", "were",
		"then", "here", "does", "each", "more", "much",
		"these", "those", "about", "would", "there", "their",
		"which", "could", "other", "after", "being", "where",
		"should", "because", "before", "between", "through",
		"please", "using":
		return true
	default:
		return false
	}
}
