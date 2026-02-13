package shared

import (
	"fmt"
	"os"
	"path/filepath"
)

// HooksFS provides filesystem operations needed by the hooks package.
type HooksFS interface {
	Stat(name string) (os.FileInfo, error)
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
	TempDir() string
	CreateExclusive(name string, data []byte, perm os.FileMode) error
	Remove(name string) error
}

// RegistryFS provides filesystem operations needed by the skipregistry package.
type RegistryFS interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	UserHomeDir() (string, error)
}

// SharedFS provides filesystem operations needed by the shared package.
type SharedFS interface {
	Stat(name string) (os.FileInfo, error)
	Getwd() (string, error)
	Abs(path string) (string, error)
}

// RealFS implements HooksFS, RegistryFS, and SharedFS using the real filesystem.
type RealFS struct{}

// Compile-time interface checks.
var (
	_ HooksFS    = (*RealFS)(nil)
	_ RegistryFS = (*RealFS)(nil)
	_ SharedFS   = (*RealFS)(nil)
)

func (r *RealFS) Stat(name string) (os.FileInfo, error) {
	info, err := os.Stat(name)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", name, err)
	}
	return info, nil
}

func (r *RealFS) ReadFile(name string) ([]byte, error) {
	data, err := os.ReadFile(name) // #nosec G304 - file path is from trusted source
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", name, err)
	}
	return data, nil
}

func (r *RealFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	if err := os.WriteFile(name, data, perm); err != nil {
		return fmt.Errorf("write file %s: %w", name, err)
	}
	return nil
}

func (r *RealFS) TempDir() string {
	return os.TempDir()
}

func (r *RealFS) CreateExclusive(name string, data []byte, perm os.FileMode) error {
	// Use O_EXCL to atomically create the file only if it doesn't exist
	// #nosec G304 - file path is from trusted source
	file, err := os.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
	if err != nil {
		return fmt.Errorf("create exclusive %s: %w", name, err)
	}
	defer func() { _ = file.Close() }()

	if _, writeErr := file.Write(data); writeErr != nil {
		// Try to clean up on write failure
		_ = os.Remove(name)
		return fmt.Errorf("write exclusive %s: %w", name, writeErr)
	}
	return nil
}

func (r *RealFS) Remove(name string) error {
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("remove %s: %w", name, err)
	}
	return nil
}

func (r *RealFS) Getwd() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	return dir, nil
}

func (r *RealFS) Abs(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("get absolute path %s: %w", path, err)
	}
	return abs, nil
}

func (r *RealFS) MkdirAll(path string, perm os.FileMode) error {
	if err := os.MkdirAll(path, perm); err != nil {
		return fmt.Errorf("mkdir all %s: %w", path, err)
	}
	return nil
}

func (r *RealFS) UserHomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get user home dir: %w", err)
	}
	return homeDir, nil
}
