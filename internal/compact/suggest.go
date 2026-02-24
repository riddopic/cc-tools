// Package compact provides tool call counting and /compact suggestion
// for Claude Code sessions.
package compact

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/riddopic/cc-tools/internal/hookcmd"
)

// Suggestor tracks tool call counts per session and suggests running /compact
// when a threshold is reached.
type Suggestor struct {
	stateDir         string
	threshold        int
	reminderInterval int
}

// NewSuggestor creates a new Suggestor that stores per-session counters in
// stateDir and triggers suggestions at the given threshold and reminder interval.
func NewSuggestor(stateDir string, threshold, reminderInterval int) *Suggestor {
	return &Suggestor{
		stateDir:         stateDir,
		threshold:        threshold,
		reminderInterval: reminderInterval,
	}
}

// RecordCall increments the tool call counter for the given session and writes
// a /compact suggestion to errOut when the threshold or reminder interval is hit.
func (s *Suggestor) RecordCall(id hookcmd.SessionID, errOut io.Writer) {
	count := s.readCount(id)
	count++
	s.writeCount(id, count)

	if s.shouldSuggest(count) {
		fmt.Fprintf(errOut,
			"[cc-tools] You have made %d tool calls in this session. "+
				"Consider running /compact to reduce context usage.\n",
			count,
		)
	}
}

func (s *Suggestor) shouldSuggest(count int) bool {
	if count == s.threshold {
		return true
	}

	if count > s.threshold && s.reminderInterval > 0 {
		return (count-s.threshold)%s.reminderInterval == 0
	}

	return false
}

func (s *Suggestor) counterPath(id hookcmd.SessionID) string {
	return filepath.Join(s.stateDir, "cc-tools-compact-"+id.FileKey()+".count")
}

func (s *Suggestor) readCount(id hookcmd.SessionID) int {
	data, err := os.ReadFile(s.counterPath(id)) // #nosec G304 -- path built from stateDir
	if err != nil {
		return 0
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0
	}

	return count
}

func (s *Suggestor) writeCount(id hookcmd.SessionID, count int) {
	// Ensure the state directory exists.
	_ = os.MkdirAll(s.stateDir, 0o750)

	_ = os.WriteFile(
		s.counterPath(id),
		[]byte(strconv.Itoa(count)),
		0o600,
	)
}
