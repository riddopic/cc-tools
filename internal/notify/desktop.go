package notify

import (
	"fmt"
	"strings"
)

// CmdRunner abstracts command execution for testing.
type CmdRunner interface {
	Run(name string, args ...string) error
}

// Desktop sends desktop notifications via osascript.
type Desktop struct {
	runner CmdRunner
}

// NewDesktop creates a new Desktop notifier with the given command runner.
func NewDesktop(runner CmdRunner) *Desktop {
	return &Desktop{
		runner: runner,
	}
}

// Send displays a desktop notification with the given title and message.
func (d *Desktop) Send(title, message string) error {
	escapedTitle := escapeAppleScript(title)
	escapedMessage := escapeAppleScript(message)

	script := fmt.Sprintf(
		`display notification "%s" with title "%s"`,
		escapedMessage,
		escapedTitle,
	)

	if err := d.runner.Run("osascript", "-e", script); err != nil {
		return fmt.Errorf("send desktop notification: %w", err)
	}

	return nil
}

// escapeAppleScript escapes special characters for AppleScript string literals.
func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)

	return s
}
