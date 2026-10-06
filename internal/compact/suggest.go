// Package compact suggests /compact to the user when a Claude Code session's
// context grows large, and logs compaction events.
package compact

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/riddopic/cc-tools/internal/hookcmd"
)

// tokensPerK converts a token count to thousands for display.
const tokensPerK = 1000

// Suggestor nudges the user to run /compact once per session when the context
// size reaches a token threshold. The nudge re-arms when the context drops
// back below the threshold, which happens after a compaction.
type Suggestor struct {
	stateDir  string
	threshold int
}

// suggestState is the per-session state persisted between hook calls.
type suggestState struct {
	Suggested bool `json:"suggested"`
}

// NewSuggestor creates a Suggestor that stores per-session state in stateDir
// and nudges once the context reaches threshold tokens.
func NewSuggestor(stateDir string, threshold int) *Suggestor {
	return &Suggestor{
		stateDir:  stateDir,
		threshold: threshold,
	}
}

// Check compares the session's current context size with the threshold and
// returns a /compact suggestion, or an empty string when none is due.
func (s *Suggestor) Check(id hookcmd.SessionID, tokens int) string {
	suggested := s.readState(id).Suggested

	switch {
	case tokens >= s.threshold && !suggested:
		s.writeState(id, suggestState{Suggested: true})

		return fmt.Sprintf(
			"[cc-tools] Context is ~%dk tokens. Consider running /compact to reduce context usage.",
			tokens/tokensPerK,
		)
	case tokens < s.threshold && suggested:
		s.writeState(id, suggestState{Suggested: false})
	}

	return ""
}

// statePath keeps the cc-tools-compact- prefix so stale state files can be
// pruned by prefix.
func (s *Suggestor) statePath(id hookcmd.SessionID) string {
	return filepath.Join(s.stateDir, "cc-tools-compact-"+id.FileKey()+".json")
}

func (s *Suggestor) readState(id hookcmd.SessionID) suggestState {
	var state suggestState

	data, err := os.ReadFile(s.statePath(id)) // #nosec G304 -- path built from stateDir
	if err != nil {
		return state
	}

	if jsonErr := json.Unmarshal(data, &state); jsonErr != nil {
		return suggestState{Suggested: false}
	}

	return state
}

func (s *Suggestor) writeState(id hookcmd.SessionID, state suggestState) {
	data, err := json.Marshal(state)
	if err != nil {
		return
	}

	_ = os.MkdirAll(s.stateDir, 0o750)
	_ = os.WriteFile(s.statePath(id), data, 0o600)
}
