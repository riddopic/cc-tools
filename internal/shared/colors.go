// Package shared provides shared utilities for all cc-tools commands.
package shared

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Raw ANSI escape codes matching the bash hooks.
const (
	ANSIRed    = "\033[0;31m"
	ANSIGreen  = "\033[0;32m"
	ANSIYellow = "\033[0;33m"
	ANSIBlue   = "\033[0;34m"
	ANSICyan   = "\033[0;36m"
	ANSIReset  = "\033[0m"
)

// RawANSIStyle provides raw ANSI formatting matching bash hooks exactly.
type RawANSIStyle struct {
	color string
}

// NewRawANSIStyle creates a raw ANSI style.
func NewRawANSIStyle(color string) RawANSIStyle {
	return RawANSIStyle{color: color}
}

// Render applies the ANSI color codes to text.
func (s RawANSIStyle) Render(text string) string {
	return fmt.Sprintf("%s%s%s", s.color, text, ANSIReset)
}

// RawErrorStyle returns a raw ANSI style for error output.
func RawErrorStyle() RawANSIStyle { return NewRawANSIStyle(ANSIRed) }

// RawSuccessStyle returns a raw ANSI style for success output.
func RawSuccessStyle() RawANSIStyle { return NewRawANSIStyle(ANSIGreen) }

// RawWarningStyle returns a raw ANSI style for warning output.
func RawWarningStyle() RawANSIStyle { return NewRawANSIStyle(ANSIYellow) }

// RawInfoStyle returns a raw ANSI style for info output.
func RawInfoStyle() RawANSIStyle { return NewRawANSIStyle(ANSIBlue) }

// RawDebugStyle returns a raw ANSI style for debug output.
func RawDebugStyle() RawANSIStyle { return NewRawANSIStyle(ANSICyan) }

// Red returns the standard red color.
func Red() lipgloss.Color { return lipgloss.Color("#f38ba8") }

// Green returns the standard green color.
func Green() lipgloss.Color { return lipgloss.Color("#a6e3a1") }

// Yellow returns the standard yellow color.
func Yellow() lipgloss.Color { return lipgloss.Color("#f9e2af") }

// Blue returns the standard blue color.
func Blue() lipgloss.Color { return lipgloss.Color("#89dceb") }

// Cyan returns the standard cyan color.
func Cyan() lipgloss.Color { return lipgloss.Color("#94e2d5") }

// ErrorStyle returns the style for error output.
func ErrorStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(Red()) }

// SuccessStyle returns the style for success output.
func SuccessStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(Green()) }

// WarningStyle returns the style for warning output.
func WarningStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(Yellow()) }

// InfoStyle returns the style for info output.
func InfoStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(Blue()) }

// DebugStyle returns the style for debug output.
func DebugStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(Cyan()) }
