package hooks

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	// summaryHeadLines is how many leading lines of command output to keep.
	// Linters usually report the first issues at the top.
	summaryHeadLines = 20
	// summaryTailLines is how many trailing lines to keep. Test runners put
	// failure summaries and exit reasons at the bottom.
	summaryTailLines = 40
	// summaryMaxChars caps one command's output so a lint and a test failure
	// together stay well under Claude Code's 10,000-character hook output cap.
	summaryMaxChars = 3000
	// capHeadShare is the fraction (1/capHeadShare) of the character budget
	// spent on the start of oversized output; the rest goes to the end.
	capHeadShare = 3
)

// ansiEscape matches CSI escape sequences such as color codes.
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

// summarizeOutput prepares command output for Claude: it strips ANSI escapes,
// normalizes line endings, drops surrounding blank lines, and keeps the head
// and tail of long output within a fixed character budget.
func summarizeOutput(raw string) string {
	text := ansiEscape.ReplaceAllString(raw, "")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.Trim(text, " \t\r\n")
	if text == "" {
		return ""
	}

	return capChars(keepHeadAndTail(text), summaryMaxChars)
}

// keepHeadAndTail elides the middle lines of text when it is too long.
func keepHeadAndTail(text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) <= summaryHeadLines+summaryTailLines {
		return text
	}

	omitted := len(lines) - summaryHeadLines - summaryTailLines
	head := strings.Join(lines[:summaryHeadLines], "\n")
	tail := strings.Join(lines[len(lines)-summaryTailLines:], "\n")

	return fmt.Sprintf("%s\n... %d lines omitted ...\n%s", head, omitted, tail)
}

// capChars splits the budget between the start and end of text when it
// exceeds limit bytes, cutting only on UTF-8 rune boundaries.
func capChars(text string, limit int) string {
	if len(text) <= limit {
		return text
	}

	headBudget := limit / capHeadShare
	headEnd := runeBoundary(text, headBudget)
	tailStart := runeBoundary(text, len(text)-(limit-headBudget))
	omitted := tailStart - headEnd

	return fmt.Sprintf("%s\n... %d chars omitted ...\n%s", text[:headEnd], omitted, text[tailStart:])
}

// runeBoundary moves i backward to the start of the rune containing it.
func runeBoundary(text string, i int) int {
	for i > 0 && !utf8.RuneStart(text[i]) {
		i--
	}

	return i
}
