package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
)

const (
	// initialTranscriptBuf is the scanner's starting buffer size.
	initialTranscriptBuf = 64 << 10
	// maxTranscriptLine bounds a single JSONL line. Tool results in real
	// transcripts regularly exceed bufio's 64 KiB default token size.
	maxTranscriptLine = 16 << 20
)

// TranscriptSummary holds aggregated info from a transcript file.
type TranscriptSummary struct {
	// TotalMessages counts prompts the user typed, excluding tool results,
	// injected reminders, and shell-mode input.
	TotalMessages int
	// FirstPrompt is the text of the first prompt the user typed. For a slash
	// command it is the command's arguments.
	FirstPrompt string
	// ToolsUsed lists distinct tool names, sorted.
	ToolsUsed []string
	// FilesModified lists distinct files touched by editing tools, in the
	// order they were first edited.
	FilesModified []string
}

// ParseTranscript reads a Claude Code JSONL transcript and aggregates the
// prompts the user typed and the tools Claude called. Malformed lines are
// skipped.
func ParseTranscript(path string) (*TranscriptSummary, error) {
	f, err := os.Open(path) // #nosec G304 -- path supplied by Claude Code
	if err != nil {
		return nil, fmt.Errorf("open transcript: %w", err)
	}
	defer f.Close()

	agg := newAggregator()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, initialTranscriptBuf), maxTranscriptLine)
	for scanner.Scan() {
		var entry transcriptEntry
		if json.Unmarshal(scanner.Bytes(), &entry) != nil {
			continue
		}
		agg.add(&entry)
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return nil, fmt.Errorf("scan transcript: %w", scanErr)
	}

	return agg.summary(), nil
}

// transcriptEntry is the subset of a transcript line this package reads.
type transcriptEntry struct {
	Type    string `json:"type"`
	IsMeta  bool   `json:"isMeta"`
	Message struct {
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

// contentBlock is one element of a message's content array.
type contentBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

type aggregator struct {
	result    TranscriptSummary
	seenTools map[string]bool
	seenFiles map[string]bool
}

func newAggregator() *aggregator {
	return &aggregator{
		result: TranscriptSummary{
			TotalMessages: 0,
			FirstPrompt:   "",
			ToolsUsed:     []string{},
			FilesModified: []string{},
		},
		seenTools: make(map[string]bool),
		seenFiles: make(map[string]bool),
	}
}

func (a *aggregator) add(entry *transcriptEntry) {
	switch entry.Type {
	case "user":
		a.addPrompt(entry)
	case "assistant":
		for _, block := range contentBlocks(entry.Message.Content) {
			if block.Type == "tool_use" && block.Name != "" {
				a.addToolUse(block)
			}
		}
	}
}

func (a *aggregator) addPrompt(entry *transcriptEntry) {
	if entry.IsMeta {
		return
	}

	text := promptText(entry.Message.Content)
	if text == "" {
		return
	}

	a.result.TotalMessages++
	if a.result.FirstPrompt == "" {
		a.result.FirstPrompt = text
	}
}

func (a *aggregator) addToolUse(block contentBlock) {
	if !a.seenTools[block.Name] {
		a.seenTools[block.Name] = true
		a.result.ToolsUsed = append(a.result.ToolsUsed, block.Name)
	}

	path := editedPath(block)
	if path != "" && !a.seenFiles[path] {
		a.seenFiles[path] = true
		a.result.FilesModified = append(a.result.FilesModified, path)
	}
}

func (a *aggregator) summary() *TranscriptSummary {
	slices.Sort(a.result.ToolsUsed)
	return &a.result
}

// contentBlocks decodes a content array. A plain-string content yields a
// single text block.
func contentBlocks(raw json.RawMessage) []contentBlock {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return []contentBlock{{Type: "text", Text: text, Name: "", Input: nil}}
	}

	var blocks []contentBlock
	if json.Unmarshal(raw, &blocks) != nil {
		return nil
	}

	return blocks
}

var (
	systemReminder = regexp.MustCompile(`(?s)<system-reminder>.*?</system-reminder>`)
	commandArgs    = regexp.MustCompile(`(?s)<command-args>(.*?)</command-args>`)
)

// noisePrefixes mark user-role lines Claude Code writes on the user's behalf:
// slash command wrappers without arguments, shell-mode I/O, and local command
// output.
func noisePrefixes() []string {
	return []string{
		"<command-", "<bash-input>", "<bash-stdout>", "<bash-stderr>",
		"<local-command-", "Caveat:",
	}
}

// StripSystemReminders removes the <system-reminder> blocks Claude Code adds to
// user-role text and trims the surrounding whitespace, leaving what the user
// actually typed.
func StripSystemReminders(text string) string {
	return strings.TrimSpace(systemReminder.ReplaceAllString(text, ""))
}

// promptText returns what the user typed in a user-role line, or "" when the
// line is a tool result or text Claude Code generated.
func promptText(raw json.RawMessage) string {
	var parts []string
	for _, block := range contentBlocks(raw) {
		if block.Type == "text" {
			parts = append(parts, block.Text)
		}
	}

	text := StripSystemReminders(strings.Join(parts, "\n"))
	if m := commandArgs.FindStringSubmatch(text); m != nil {
		return strings.TrimSpace(m[1])
	}

	for _, prefix := range noisePrefixes() {
		if strings.HasPrefix(text, prefix) {
			return ""
		}
	}

	return text
}

// editedPath returns the file an editing tool call changed, or "".
func editedPath(block contentBlock) string {
	var key string

	switch block.Name {
	case "Edit", "MultiEdit", "Write":
		key = "file_path"
	case "NotebookEdit":
		key = "notebook_path"
	default:
		return ""
	}

	var fields map[string]any
	if json.Unmarshal(block.Input, &fields) != nil {
		return ""
	}

	path, _ := fields[key].(string)

	return path
}
