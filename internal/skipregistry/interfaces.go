package skipregistry

import (
	"context"
	"io"
)

// Reader provides read operations for the skip registry.
type Reader interface {
	// IsSkipped checks if a directory has any skip types configured.
	IsSkipped(ctx context.Context, dir DirectoryPath, skipType SkipType) (bool, error)
	// GetSkipTypes returns all skip types configured for a directory.
	GetSkipTypes(ctx context.Context, dir DirectoryPath) ([]SkipType, error)
	// ListAll returns all directories and their skip configurations.
	ListAll(ctx context.Context) ([]RegistryEntry, error)
}

// Writer provides write operations for the skip registry.
type Writer interface {
	// AddSkip adds a skip type to a directory.
	AddSkip(ctx context.Context, dir DirectoryPath, skipType SkipType) error
	// RemoveSkip removes a skip type from a directory.
	RemoveSkip(ctx context.Context, dir DirectoryPath, skipType SkipType) error
	// Clear removes all skip configurations for a directory.
	Clear(ctx context.Context, dir DirectoryPath) error
}

// Storage provides persistence operations for the registry.
type Storage interface {
	// Load reads the registry from storage.
	Load(ctx context.Context) (RegistryData, error)
	// Save atomically writes the registry to storage.
	Save(ctx context.Context, data RegistryData) error
}

// Registry combines all registry operations.
type Registry interface {
	Reader
	Writer
}

// OutputWriter provides output operations.
type OutputWriter interface {
	io.Writer
}
