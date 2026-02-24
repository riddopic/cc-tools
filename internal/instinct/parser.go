package instinct

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ErrNoID indicates a frontmatter block has no id field.
var ErrNoID = errors.New("frontmatter block has no id field")

// ParseFrontmatter parses YAML frontmatter blocks from input and returns
// instincts. Each block is delimited by "---" lines. Instincts without an
// id field are skipped. Returns nil for empty input or input without
// frontmatter delimiters.
func ParseFrontmatter(input string) ([]Instinct, error) {
	if strings.TrimSpace(input) == "" {
		return nil, nil
	}

	blocks := splitFrontmatterBlocks(input)
	if len(blocks) == 0 {
		return nil, nil
	}

	var result []Instinct

	for _, block := range blocks {
		inst, err := parseBlock(block)
		if err != nil {
			if errors.Is(err, ErrNoID) {
				continue
			}

			return nil, fmt.Errorf("parse block: %w", err)
		}

		result = append(result, *inst)
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result, nil
}

// frontmatterBlock holds the raw frontmatter and content sections.
type frontmatterBlock struct {
	frontmatter string
	content     string
}

// splitFrontmatterBlocks splits input into frontmatter blocks delimited by
// "---" lines. Returns nil if no valid frontmatter is found.
func splitFrontmatterBlocks(input string) []frontmatterBlock {
	const delimiter = "---"

	lines := strings.Split(input, "\n")

	var blocks []frontmatterBlock
	var fmLines []string

	inFrontmatter := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == delimiter {
			if !inFrontmatter {
				inFrontmatter = true
				fmLines = nil

				continue
			}

			content := extractContent(lines, i+1)

			blocks = append(blocks, frontmatterBlock{
				frontmatter: strings.Join(fmLines, "\n"),
				content:     content,
			})

			inFrontmatter = false

			continue
		}

		if inFrontmatter {
			fmLines = append(fmLines, line)
		}
	}

	return blocks
}

// extractContent collects lines after the closing delimiter until the next
// opening delimiter or end of input.
func extractContent(lines []string, startIdx int) string {
	const delimiter = "---"

	var contentLines []string

	for i := startIdx; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == delimiter {
			break
		}

		contentLines = append(contentLines, lines[i])
	}

	if len(contentLines) == 0 {
		return ""
	}

	return strings.Join(contentLines, "\n")
}

// parseBlock converts a frontmatter block into an Instinct. Returns ErrNoID
// if the block has no id field.
func parseBlock(block frontmatterBlock) (*Instinct, error) {
	fields := parseFrontmatterFields(block.frontmatter)

	id := fields["id"]
	if id == "" {
		return nil, ErrNoID
	}

	confidence, err := parseFloat(fields["confidence"])
	if err != nil {
		return nil, fmt.Errorf("parse confidence: %w", err)
	}

	createdAt, err := parseTimestamp(fields["created_at"])
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}

	updatedAt, err := parseTimestamp(fields["updated_at"])
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}

	inst := &Instinct{
		ID:         id,
		Trigger:    fields["trigger"],
		Confidence: confidence,
		Domain:     fields["domain"],
		Source:     fields["source"],
		SourceRepo: fields["source_repo"],
		Content:    block.content,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}

	return inst, nil
}

// parseFrontmatterFields splits key: value lines into a map. Handles both
// quoted and unquoted values.
func parseFrontmatterFields(frontmatter string) map[string]string {
	fields := make(map[string]string)

	for line := range strings.SplitSeq(frontmatter, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		key, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = unquote(value)
		fields[key] = value
	}

	return fields
}

// unquote removes surrounding double quotes from a string if present.
func unquote(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}

	return s
}

// parseFloat converts a string to float64, returning 0 for empty strings.
func parseFloat(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid float %q: %w", s, err)
	}

	return v, nil
}

// parseTimestamp parses an RFC3339 timestamp string.
func parseTimestamp(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp %q: %w", s, err)
	}

	return t, nil
}
