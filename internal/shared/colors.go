// Package shared provides shared utilities for all cc-tools commands.
package shared

// Raw ANSI escape codes matching the bash hooks.
const (
	ANSIRed    = "\033[0;31m"
	ANSIGreen  = "\033[0;32m"
	ANSIYellow = "\033[0;33m"
	ANSIBlue   = "\033[0;34m"
	ANSICyan   = "\033[0;36m"
	ANSIReset  = "\033[0m"
)
