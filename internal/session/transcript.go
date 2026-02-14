package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"slices"
)

// TranscriptSummary holds aggregated info from a transcript file.
type TranscriptSummary struct {
	TotalMessages int
	ToolsUsed     []string
	FilesModified []string
}

// ParseTranscript reads a JSONL transcript file and produces an aggregated summary.
// It counts lines with "type":"human" as messages, extracts tool names from
// "type":"tool_use" lines, and extracts file paths from tool_use inputs.
// Both ToolsUsed and FilesModified are deduplicated.
func ParseTranscript(path string) (*TranscriptSummary, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open transcript: %w", err)
	}
	defer f.Close()

	summary := &TranscriptSummary{
		TotalMessages: 0,
		ToolsUsed:     []string{},
		FilesModified: []string{},
	}

	seenTools := make(map[string]bool)
	seenFiles := make(map[string]bool)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry transcriptEntry
		if unmarshalErr := json.Unmarshal(line, &entry); unmarshalErr != nil {
			continue
		}

		if entry.Type == "human" {
			summary.TotalMessages++
		}

		if entry.Type == "tool_use" && entry.Name != "" {
			if !seenTools[entry.Name] {
				seenTools[entry.Name] = true
				summary.ToolsUsed = append(summary.ToolsUsed, entry.Name)
			}

			filePath := extractFilePath(entry.Input)
			if filePath != "" && !seenFiles[filePath] {
				seenFiles[filePath] = true
				summary.FilesModified = append(summary.FilesModified, filePath)
			}
		}
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return nil, fmt.Errorf("scan transcript: %w", scanErr)
	}

	slices.Sort(summary.ToolsUsed)
	slices.Sort(summary.FilesModified)

	return summary, nil
}

// transcriptEntry represents a single line in a JSONL transcript.
type transcriptEntry struct {
	Type  string          `json:"type"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

// extractFilePath attempts to read "file_path" from a JSON input object.
func extractFilePath(input json.RawMessage) string {
	if len(input) == 0 {
		return ""
	}

	var fields struct {
		FilePath string `json:"file_path"`
	}

	if err := json.Unmarshal(input, &fields); err != nil {
		return ""
	}

	return fields.FilePath
}
