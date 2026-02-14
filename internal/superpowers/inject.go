// Package superpowers injects skill content into SessionStart hook events.
package superpowers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// skillRelPath is the relative path from the project directory to the
// using-superpowers skill file.
const skillRelPath = ".claude/skills/using-superpowers/SKILL.md"

// hookOutput represents the JSON envelope expected by Claude Code hooks.
type hookOutput struct {
	HookSpecificOutput hookSpecificOutput `json:"hookSpecificOutput"`
}

// hookSpecificOutput carries the event name and injected context.
type hookSpecificOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext"`
}

// Injector reads skill file and outputs hookSpecificOutput JSON.
type Injector struct {
	projectDir string
}

// NewInjector creates a new Injector for the given project directory.
func NewInjector(projectDir string) *Injector {
	return &Injector{
		projectDir: projectDir,
	}
}

// Run reads the using-superpowers SKILL.md and writes hookSpecificOutput JSON
// to the provided writer. Returns nil if the skill file does not exist
// (silent skip).
func (inj *Injector) Run(_ context.Context, out io.Writer) error {
	skillPath := filepath.Join(inj.projectDir, skillRelPath)

	data, err := os.ReadFile(skillPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("reading skill file: %w", err)
	}

	content := string(data)
	wrapped := "<EXTREMELY_IMPORTANT>\n" + content + "\n</EXTREMELY_IMPORTANT>"

	payload := hookOutput{
		HookSpecificOutput: hookSpecificOutput{
			HookEventName:     "SessionStart",
			AdditionalContext: wrapped,
		},
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encoding hook output: %w", err)
	}

	if _, err = out.Write(encoded); err != nil {
		return fmt.Errorf("writing hook output: %w", err)
	}

	return nil
}
