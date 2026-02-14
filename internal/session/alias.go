package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrAliasNotFound indicates the requested alias does not exist.
var ErrAliasNotFound = errors.New("alias not found")

// AliasManager manages named shortcuts for session IDs.
type AliasManager struct {
	path string
}

// NewAliasManager creates a new AliasManager that persists aliases at the given file path.
func NewAliasManager(path string) *AliasManager {
	return &AliasManager{path: path}
}

// Set creates or overwrites a named alias pointing to a session ID.
func (am *AliasManager) Set(alias, sessionID string) error {
	aliases, err := am.loadAliases()
	if err != nil {
		return err
	}

	aliases[alias] = sessionID

	return am.saveAliases(aliases)
}

// Resolve returns the session ID associated with the given alias.
func (am *AliasManager) Resolve(alias string) (string, error) {
	aliases, err := am.loadAliases()
	if err != nil {
		return "", err
	}

	sessionID, ok := aliases[alias]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrAliasNotFound, alias)
	}

	return sessionID, nil
}

// Remove deletes a named alias.
func (am *AliasManager) Remove(alias string) error {
	aliases, err := am.loadAliases()
	if err != nil {
		return err
	}

	if _, ok := aliases[alias]; !ok {
		return fmt.Errorf("%w: %s", ErrAliasNotFound, alias)
	}

	delete(aliases, alias)

	return am.saveAliases(aliases)
}

// List returns all aliases as a map from alias name to session ID.
func (am *AliasManager) List() (map[string]string, error) {
	return am.loadAliases()
}

func (am *AliasManager) loadAliases() (map[string]string, error) {
	data, err := os.ReadFile(am.path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}

		return nil, fmt.Errorf("read alias file: %w", err)
	}

	var aliases map[string]string
	if unmarshalErr := json.Unmarshal(data, &aliases); unmarshalErr != nil {
		return nil, fmt.Errorf("unmarshal alias file: %w", unmarshalErr)
	}

	return aliases, nil
}

func (am *AliasManager) saveAliases(aliases map[string]string) error {
	dir := filepath.Dir(am.path)
	if mkdirErr := os.MkdirAll(dir, 0o750); mkdirErr != nil {
		return fmt.Errorf("create alias directory: %w", mkdirErr)
	}

	data, err := json.MarshalIndent(aliases, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal aliases: %w", err)
	}

	if writeErr := os.WriteFile(am.path, data, 0o600); writeErr != nil {
		return fmt.Errorf("write alias file: %w", writeErr)
	}

	return nil
}
