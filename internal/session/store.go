// Package session manages session metadata storage as structured JSON files.
package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// sessionFileVersion is the current schema version for session files.
const sessionFileVersion = "1"

// Session represents a recorded Claude Code session.
type Session struct {
	Version       string    `json:"version"`
	ID            string    `json:"id"`
	Date          string    `json:"date"`
	Started       time.Time `json:"started"`
	Ended         time.Time `json:"ended,omitzero"`
	Title         string    `json:"title"`
	Summary       string    `json:"summary,omitempty"`
	ToolsUsed     []string  `json:"tools_used,omitempty"`
	FilesModified []string  `json:"files_modified,omitempty"`
	MessageCount  int       `json:"message_count,omitempty"`
}

// Store manages session files in a directory.
type Store struct {
	dir string
}

// Sentinel errors for session operations.
var (
	// ErrNotFound indicates the requested session was not found.
	ErrNotFound = errors.New("session not found")
	// ErrEmptyID indicates an empty session ID was provided.
	ErrEmptyID = errors.New("session ID must not be empty")
)

// NewStore creates a new Store rooted at the given directory.
func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

// Save persists a session as {date}-{id}.json in the store directory.
func (s *Store) Save(session *Session) error {
	if session.ID == "" {
		return ErrEmptyID
	}

	if session.Version == "" {
		session.Version = sessionFileVersion
	}

	if err := os.MkdirAll(s.dir, 0o750); err != nil {
		return fmt.Errorf("create session directory: %w", err)
	}

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	filename := s.filename(session.Date, session.ID)

	if writeErr := os.WriteFile(filepath.Join(s.dir, filename), data, 0o600); writeErr != nil {
		return fmt.Errorf("write session file: %w", writeErr)
	}

	return nil
}

// Load retrieves a session by its ID using glob matching.
func (s *Store) Load(id string) (*Session, error) {
	if id == "" {
		return nil, ErrEmptyID
	}

	pattern := filepath.Join(s.dir, "*-"+id+".json")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob session files: %w", err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, id)
	}

	return s.readSessionFile(matches[0])
}

// List returns the most recent sessions, limited by count.
// Sessions are sorted by filename in descending order (most recent first).
func (s *Store) List(limit int) ([]*Session, error) {
	entries, err := s.readAllSessions()
	if err != nil {
		return nil, err
	}

	// Sort descending by filename (date prefix gives chronological order).
	slices.Reverse(entries)

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return entries, nil
}

// FindByDate returns sessions whose date field starts with the given prefix.
func (s *Store) FindByDate(date string) ([]*Session, error) {
	entries, err := s.readAllSessions()
	if err != nil {
		return nil, err
	}

	result := make([]*Session, 0, len(entries))
	for _, entry := range entries {
		if strings.HasPrefix(entry.Date, date) {
			result = append(result, entry)
		}
	}

	return result, nil
}

// Search returns sessions whose title or summary contains the query string (case-insensitive).
func (s *Store) Search(query string) ([]*Session, error) {
	entries, err := s.readAllSessions()
	if err != nil {
		return nil, err
	}

	lowerQuery := strings.ToLower(query)
	result := make([]*Session, 0, len(entries))

	for _, entry := range entries {
		titleMatch := strings.Contains(strings.ToLower(entry.Title), lowerQuery)
		summaryMatch := strings.Contains(strings.ToLower(entry.Summary), lowerQuery)

		if titleMatch || summaryMatch {
			result = append(result, entry)
		}
	}

	return result, nil
}

func (s *Store) filename(date, id string) string {
	return date + "-" + id + ".json"
}

func (s *Store) readSessionFile(path string) (*Session, error) {
	// #nosec G304 -- path is built from a controlled directory.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read session file: %w", err)
	}

	var sess Session
	if unmarshalErr := json.Unmarshal(data, &sess); unmarshalErr != nil {
		return nil, fmt.Errorf("unmarshal session file %s: %w", filepath.Base(path), unmarshalErr)
	}

	return &sess, nil
}

func (s *Store) readAllSessions() ([]*Session, error) {
	pattern := filepath.Join(s.dir, "*.json")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob session files: %w", err)
	}

	// Glob returns sorted results, so filenames with date prefixes are chronologically ordered.
	sessions := make([]*Session, 0, len(matches))

	for _, match := range matches {
		sess, readErr := s.readSessionFile(match)
		if readErr != nil {
			continue
		}

		sessions = append(sessions, sess)
	}

	return sessions, nil
}
