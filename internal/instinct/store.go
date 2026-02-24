package instinct

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ErrNotFound indicates the requested instinct was not found.
var ErrNotFound = errors.New("instinct not found")

// FileStore stores instincts as YAML frontmatter files on disk.
type FileStore struct {
	personalDir  string
	inheritedDir string
}

// NewFileStore creates a FileStore that reads from personalDir and
// inheritedDir, and writes to personalDir.
func NewFileStore(personalDir, inheritedDir string) *FileStore {
	return &FileStore{
		personalDir:  personalDir,
		inheritedDir: inheritedDir,
	}
}

// Save writes an instinct to personalDir/{id}.yaml in YAML frontmatter
// format. It creates directories as needed.
func (s *FileStore) Save(inst Instinct) error {
	if err := validateID(inst.ID); err != nil {
		return fmt.Errorf("save instinct: %w", err)
	}

	if err := os.MkdirAll(s.personalDir, 0o750); err != nil {
		return fmt.Errorf("create instinct directory: %w", err)
	}

	data := marshalFrontmatter(inst)
	path := filepath.Join(s.personalDir, inst.ID+".yaml")

	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		return fmt.Errorf("write instinct file: %w", err)
	}

	return nil
}

// Get retrieves an instinct by ID, searching personalDir first then
// inheritedDir. Returns ErrNotFound if the instinct does not exist in
// either directory.
func (s *FileStore) Get(id string) (*Instinct, error) {
	if err := validateID(id); err != nil {
		return nil, fmt.Errorf("get instinct: %w", err)
	}

	inst, err := s.loadFromDir(s.personalDir, id)
	if err == nil {
		return inst, nil
	}

	if s.inheritedDir != "" {
		inst, err = s.loadFromDir(s.inheritedDir, id)
		if err == nil {
			return inst, nil
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrNotFound, id)
}

// List returns all instincts from both directories, filtered by opts,
// sorted by ID.
func (s *FileStore) List(opts ListOptions) ([]Instinct, error) {
	seen := make(map[string]bool)

	var result []Instinct

	personal, err := s.globDir(s.personalDir)
	if err != nil {
		return nil, fmt.Errorf("list personal instincts: %w", err)
	}

	for _, inst := range personal {
		if matchesFilter(inst, opts) {
			result = append(result, inst)
			seen[inst.ID] = true
		}
	}

	if s.inheritedDir != "" {
		inherited, inheritErr := s.globDir(s.inheritedDir)
		if inheritErr != nil {
			return nil, fmt.Errorf("list inherited instincts: %w", inheritErr)
		}

		for _, inst := range inherited {
			if !seen[inst.ID] && matchesFilter(inst, opts) {
				result = append(result, inst)
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}

// Delete removes an instinct from personalDir. It does not remove
// inherited instincts.
func (s *FileStore) Delete(id string) error {
	if err := validateID(id); err != nil {
		return fmt.Errorf("delete instinct: %w", err)
	}

	path := filepath.Join(s.personalDir, id+".yaml")

	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("delete instinct: %w: %s", ErrNotFound, id)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete instinct file: %w", err)
	}

	return nil
}

// loadFromDir reads and parses a single instinct file from a directory.
func (s *FileStore) loadFromDir(dir, id string) (*Instinct, error) {
	path := filepath.Join(dir, id+".yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read instinct %s: %w", id, err)
	}

	instincts, err := ParseFrontmatter(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse instinct %s: %w", id, err)
	}

	if len(instincts) == 0 {
		return nil, fmt.Errorf("no instinct found in %s", path)
	}

	return &instincts[0], nil
}

// globDir reads all .yaml files in a directory and parses them.
func (s *FileStore) globDir(dir string) ([]Instinct, error) {
	pattern := filepath.Join(dir, "*.yaml")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob %s: %w", pattern, err)
	}

	var result []Instinct

	for _, path := range matches {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, fmt.Errorf("read %s: %w", path, readErr)
		}

		instincts, parseErr := ParseFrontmatter(string(data))
		if parseErr != nil {
			return nil, fmt.Errorf("parse %s: %w", path, parseErr)
		}

		result = append(result, instincts...)
	}

	return result, nil
}

// matchesFilter checks whether an instinct matches the given list options.
func matchesFilter(inst Instinct, opts ListOptions) bool {
	if opts.Domain != "" && inst.Domain != opts.Domain {
		return false
	}

	if opts.MinConfidence > 0 && inst.Confidence < opts.MinConfidence {
		return false
	}

	if opts.Source != "" && inst.Source != opts.Source {
		return false
	}

	return true
}

// validateID checks that an instinct ID is safe for use as a filename.
func validateID(id string) error {
	if id == "" {
		return errors.New("instinct ID must not be empty")
	}

	cleaned := filepath.Clean(id)
	if strings.Contains(cleaned, "..") {
		return errors.New("instinct ID must not contain path traversal")
	}

	if cleaned != id {
		return fmt.Errorf("instinct ID %q is not clean", id)
	}

	return nil
}

// marshalFrontmatter serializes an instinct to YAML frontmatter format.
func marshalFrontmatter(inst Instinct) string {
	var sb strings.Builder

	sb.WriteString("---\n")
	writeFrontmatterField(&sb, "id", inst.ID)
	writeFrontmatterField(&sb, "trigger", inst.Trigger)
	writeFrontmatterFloat(&sb, "confidence", inst.Confidence)
	writeFrontmatterField(&sb, "domain", inst.Domain)
	writeFrontmatterField(&sb, "source", inst.Source)

	if inst.SourceRepo != "" {
		writeFrontmatterField(&sb, "source_repo", inst.SourceRepo)
	}

	writeFrontmatterField(&sb, "created_at", inst.CreatedAt.Format(time.RFC3339))
	writeFrontmatterField(&sb, "updated_at", inst.UpdatedAt.Format(time.RFC3339))
	sb.WriteString("---\n")

	if inst.Content != "" {
		sb.WriteString(inst.Content)
	}

	return sb.String()
}

// writeFrontmatterField writes a key: value line to the builder.
func writeFrontmatterField(sb *strings.Builder, key, value string) {
	sb.WriteString(key)
	sb.WriteString(": ")
	sb.WriteString(value)
	sb.WriteString("\n")
}

// writeFrontmatterFloat writes a key: float line to the builder.
func writeFrontmatterFloat(sb *strings.Builder, key string, value float64) {
	sb.WriteString(key)
	sb.WriteString(": ")
	fmt.Fprintf(sb, "%g", value)
	sb.WriteString("\n")
}
