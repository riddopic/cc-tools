package skipregistry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/riddopic/cc-tools/internal/shared"
)

// JSONStorage implements Storage using JSON files.
type JSONStorage struct {
	fs       shared.RegistryFS
	filePath string
}

// NewJSONStorage creates a new JSON storage backend.
func NewJSONStorage(fs shared.RegistryFS, filePath string) *JSONStorage {
	return &JSONStorage{
		fs:       fs,
		filePath: filePath,
	}
}

// Load reads the registry from the JSON file.
func (s *JSONStorage) Load(_ context.Context) (RegistryData, error) {
	// Read the file
	data, err := s.fs.ReadFile(s.filePath)
	if err != nil {
		// Check if file doesn't exist using errors.Is to handle wrapped errors
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("read registry file: %w", err)
	}

	// Parse JSON
	var registry RegistryData
	if len(data) == 0 {
		// Empty file, return empty registry
		return make(RegistryData), nil
	}

	if unmarshalErr := json.Unmarshal(data, &registry); unmarshalErr != nil {
		return nil, fmt.Errorf("parse registry JSON: %w", unmarshalErr)
	}

	if registry == nil {
		registry = make(RegistryData)
	}

	return registry, nil
}

// Save atomically writes the registry to the JSON file.
func (s *JSONStorage) Save(_ context.Context, data RegistryData) error {
	// Ensure directory exists
	dir := filepath.Dir(s.filePath)
	const dirPerm = 0o755
	if err := s.fs.MkdirAll(dir, os.FileMode(dirPerm)); err != nil {
		return fmt.Errorf("create registry directory: %w", err)
	}

	// Marshal to JSON with indentation for readability
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal registry JSON: %w", err)
	}

	// Add newline at end for better text file formatting
	jsonData = append(jsonData, '\n')

	// Write atomically by writing to temp file and moving
	tempFile := s.filePath + ".tmp"
	const filePerm = 0o644
	if writeErr := s.fs.WriteFile(tempFile, jsonData, os.FileMode(filePerm)); writeErr != nil {
		return fmt.Errorf("write temp registry file: %w", writeErr)
	}

	// Rename temp file to actual file (atomic on most filesystems)
	if renameErr := s.fs.Rename(tempFile, s.filePath); renameErr != nil {
		// Try to clean up temp file
		_ = s.fs.Remove(tempFile)
		return fmt.Errorf("rename registry file: %w", renameErr)
	}

	return nil
}

// DefaultStorage creates a storage instance with the default file path.
func DefaultStorage() *JSONStorage {
	return NewJSONStorage(&shared.RealFS{}, getRegistryPath())
}
