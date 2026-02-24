package instinct

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ImportAction describes what should happen to an instinct during import.
type ImportAction int

const (
	// ImportSkip means the instinct will not be imported.
	ImportSkip ImportAction = iota
	// ImportNew means the instinct is new and will be added.
	ImportNew
	// ImportOverwrite means the instinct exists and will be replaced.
	ImportOverwrite
)

// String returns a human-readable label for the action.
func (a ImportAction) String() string {
	switch a {
	case ImportSkip:
		return "skip"
	case ImportNew:
		return "import"
	case ImportOverwrite:
		return "overwrite"
	default:
		return "import"
	}
}

// Label returns the action label, optionally prefixed with [dry-run].
func (a ImportAction) Label(dryRun bool) string {
	prefix := ""
	if dryRun {
		prefix = "[dry-run] "
	}

	return prefix + a.String() + ":"
}

// ImportOptions controls import behavior.
type ImportOptions struct {
	DryRun        bool
	Force         bool
	MinConfidence float64
}

// ImportItem records the action taken for a single instinct.
type ImportItem struct {
	Instinct Instinct
	Action   ImportAction
}

// ImportResult holds the outcome of an import operation.
type ImportResult struct {
	Items []ImportItem
}

// Imported returns the count of non-skip items.
func (r *ImportResult) Imported() int {
	count := 0
	for _, item := range r.Items {
		if item.Action != ImportSkip {
			count++
		}
	}

	return count
}

// Verb returns "imported" or "would be imported" depending on dry-run.
func (r *ImportResult) Verb(dryRun bool) string {
	if dryRun {
		return "would be imported"
	}

	return "imported"
}

// ClassifyImport determines whether an instinct should be imported, skipped,
// or overwritten based on existing store state and options. It returns an
// error when the store lookup fails for reasons other than "not found".
func ClassifyImport(store *FileStore, inst Instinct, force bool, minConf float64) (ImportAction, error) {
	if minConf > 0 && inst.Confidence < minConf {
		return ImportSkip, nil
	}

	_, err := store.Get(inst.ID)
	if err == nil && !force {
		return ImportSkip, nil
	}

	if err == nil {
		return ImportOverwrite, nil
	}

	if errors.Is(err, ErrNotFound) {
		return ImportNew, nil
	}

	return ImportSkip, fmt.Errorf("classify import %s: %w", inst.ID, err)
}

// Import processes instincts and saves eligible ones to writeStore.
// It classifies each instinct against readStore and writes to writeStore.
func Import(
	readStore, writeStore *FileStore,
	instincts []Instinct,
	opts ImportOptions,
) (*ImportResult, error) {
	result := &ImportResult{Items: nil}

	for _, inst := range instincts {
		action, classErr := ClassifyImport(readStore, inst, opts.Force, opts.MinConfidence)
		if classErr != nil {
			return nil, fmt.Errorf("classify instinct %s: %w", inst.ID, classErr)
		}

		result.Items = append(result.Items, ImportItem{Instinct: inst, Action: action})

		if action == ImportSkip || opts.DryRun {
			continue
		}

		if err := writeStore.Save(inst); err != nil {
			return nil, fmt.Errorf("save instinct %s: %w", inst.ID, err)
		}
	}

	return result, nil
}

// ReadAndParseSource reads a file at source and parses its frontmatter.
// It rejects paths containing directory traversal sequences.
func ReadAndParseSource(source string) ([]Instinct, error) {
	cleanPath := filepath.Clean(source)
	if strings.Contains(cleanPath, "..") {
		return nil, errors.New("invalid path: directory traversal detected")
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("read source file: %w", err)
	}

	parsed, err := ParseFrontmatter(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse source file: %w", err)
	}

	return parsed, nil
}
