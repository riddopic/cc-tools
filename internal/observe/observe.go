// Package observe records tool usage events as JSONL for the continuous learning system.
package observe

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// observationsFile is the name of the JSONL file that stores observations.
const observationsFile = "observations.jsonl"

// disabledFile is the name of the marker file that disables observation recording.
const disabledFile = ".disabled"

// Event represents a single tool usage observation.
type Event struct {
	Timestamp  time.Time       `json:"timestamp"`
	Phase      string          `json:"phase"` // "pre", "post", or "failure".
	ToolName   string          `json:"tool_name"`
	ToolInput  json.RawMessage `json:"tool_input,omitempty"`
	ToolOutput json.RawMessage `json:"tool_output,omitempty"`
	Error      string          `json:"error,omitempty"`
	SessionID  string          `json:"session_id"`
}

// Observer records tool events to a JSONL file.
type Observer struct {
	dir           string
	maxFileSizeMB int
}

// NewObserver creates a new Observer.
func NewObserver(dir string, maxFileSizeMB int) *Observer {
	return &Observer{
		dir:           dir,
		maxFileSizeMB: maxFileSizeMB,
	}
}

// Record appends an event as a JSON line to observations.jsonl.
// It checks file size before writing and rotates if over maxFileSizeMB.
// Returns nil if observation recording is disabled.
func (o *Observer) Record(event Event) error {
	if o.isDisabled() {
		return nil
	}

	if err := os.MkdirAll(o.dir, 0o750); err != nil {
		return fmt.Errorf("create observe directory: %w", err)
	}

	filePath := filepath.Join(o.dir, observationsFile)

	if err := RotateIfNeeded(filePath, o.maxFileSizeMB); err != nil {
		return fmt.Errorf("rotate observations file: %w", err)
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	data = append(data, '\n')

	// #nosec G304 -- filePath is built from a controlled directory.
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open observations file: %w", err)
	}
	defer f.Close()

	if _, writeErr := f.Write(data); writeErr != nil {
		return fmt.Errorf("write event: %w", writeErr)
	}

	return nil
}

func (o *Observer) isDisabled() bool {
	disabledPath := filepath.Join(o.dir, disabledFile)
	_, err := os.Stat(disabledPath)

	return err == nil
}
