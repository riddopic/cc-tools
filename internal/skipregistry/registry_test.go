package skipregistry_test

import (
	"context"
	"errors"
	"testing"

	"github.com/riddopic/cc-tools/internal/skipregistry"
)

// mockStorage is a mock implementation of Storage for testing.
type mockStorage struct {
	data      skipregistry.RegistryData
	loadErr   error
	saveErr   error
	loadCalls int
	saveCalls int
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		data:      make(skipregistry.RegistryData),
		loadErr:   nil,
		saveErr:   nil,
		loadCalls: 0,
		saveCalls: 0,
	}
}

func (m *mockStorage) Load(_ context.Context) (skipregistry.RegistryData, error) {
	m.loadCalls++
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	// Return a copy to prevent mutations.
	dataCopy := make(skipregistry.RegistryData)
	for k, v := range m.data {
		types := make([]string, 0, len(v))
		dataCopy[k] = append(types, v...)
	}
	return dataCopy, nil
}

func (m *mockStorage) Save(_ context.Context, data skipregistry.RegistryData) error {
	m.saveCalls++
	if m.saveErr != nil {
		return m.saveErr
	}
	// Save a copy to prevent mutations.
	m.data = make(skipregistry.RegistryData)
	for k, v := range data {
		types := make([]string, 0, len(v))
		m.data[k] = append(types, v...)
	}
	return nil
}

func TestRegistry_IsSkipped(t *testing.T) {
	tests := []struct {
		name      string
		setupData skipregistry.RegistryData
		dir       skipregistry.DirectoryPath
		skipType  skipregistry.SkipType
		want      bool
		wantErr   bool
	}{
		{
			name:      "empty registry returns false",
			setupData: skipregistry.RegistryData{},
			dir:       "/project",
			skipType:  skipregistry.SkipTypeLint,
			want:      false,
			wantErr:   false,
		},
		{
			name: "finds lint skip",
			setupData: skipregistry.RegistryData{
				"/project": {"lint"},
			},
			dir:      "/project",
			skipType: skipregistry.SkipTypeLint,
			want:     true,
			wantErr:  false,
		},
		{
			name: "finds test skip",
			setupData: skipregistry.RegistryData{
				"/project": {"test"},
			},
			dir:      "/project",
			skipType: skipregistry.SkipTypeTest,
			want:     true,
			wantErr:  false,
		},
		{
			name: "multiple skip types",
			setupData: skipregistry.RegistryData{
				"/project": {"lint", "test"},
			},
			dir:      "/project",
			skipType: skipregistry.SkipTypeLint,
			want:     true,
			wantErr:  false,
		},
		{
			name: "different directory not skipped",
			setupData: skipregistry.RegistryData{
				"/project": {"lint"},
			},
			dir:      "/other",
			skipType: skipregistry.SkipTypeLint,
			want:     false,
			wantErr:  false,
		},
		{
			name:      "invalid path returns error",
			setupData: skipregistry.RegistryData{},
			dir:       "relative/path",
			skipType:  skipregistry.SkipTypeLint,
			want:      false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newMockStorage()
			storage.data = tt.setupData
			r := skipregistry.NewRegistry(storage)

			got, err := r.IsSkipped(context.Background(), tt.dir, tt.skipType)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsSkipped() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsSkipped() = %v, want %v", got, tt.want)
			}
		})
	}
}

// assertRegistryData validates that the storage data matches the expected data.
func assertRegistryData(t *testing.T, method string, storage *mockStorage, wantData skipregistry.RegistryData) {
	t.Helper()

	if len(storage.data) != len(wantData) {
		t.Errorf("%s() data length = %v, want %v", method, len(storage.data), len(wantData))
		return
	}

	for path, types := range wantData {
		gotTypes := storage.data[path]
		if len(gotTypes) != len(types) {
			t.Errorf("%s() types for %s = %v, want %v", method, path, gotTypes, types)
			continue
		}

		for i, wantType := range types {
			if i >= len(gotTypes) || gotTypes[i] != wantType {
				t.Errorf("%s() type[%d] for %s = %v, want %v", method, i, path, gotTypes[i], wantType)
			}
		}
	}
}

func TestRegistry_AddSkip(t *testing.T) {
	tests := []struct {
		name      string
		setupData skipregistry.RegistryData
		dir       skipregistry.DirectoryPath
		skipType  skipregistry.SkipType
		wantData  skipregistry.RegistryData
		wantErr   bool
	}{
		{
			name:      "add lint to empty registry",
			setupData: skipregistry.RegistryData{},
			dir:       "/project",
			skipType:  skipregistry.SkipTypeLint,
			wantData: skipregistry.RegistryData{
				"/project": {"lint"},
			},
			wantErr: false,
		},
		{
			name: "add test to existing lint",
			setupData: skipregistry.RegistryData{
				"/project": {"lint"},
			},
			dir:      "/project",
			skipType: skipregistry.SkipTypeTest,
			wantData: skipregistry.RegistryData{
				"/project": {"lint", "test"},
			},
			wantErr: false,
		},
		{
			name:      "add all expands to both",
			setupData: skipregistry.RegistryData{},
			dir:       "/project",
			skipType:  skipregistry.SkipTypeAll,
			wantData: skipregistry.RegistryData{
				"/project": {"lint", "test"},
			},
			wantErr: false,
		},
		{
			name: "add duplicate is idempotent",
			setupData: skipregistry.RegistryData{
				"/project": {"lint"},
			},
			dir:      "/project",
			skipType: skipregistry.SkipTypeLint,
			wantData: skipregistry.RegistryData{
				"/project": {"lint"},
			},
			wantErr: false,
		},
		{
			name:      "invalid path returns error",
			setupData: skipregistry.RegistryData{},
			dir:       "relative/path",
			skipType:  skipregistry.SkipTypeLint,
			wantData:  skipregistry.RegistryData{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newMockStorage()
			storage.data = tt.setupData
			r := skipregistry.NewRegistry(storage)

			err := r.AddSkip(context.Background(), tt.dir, tt.skipType)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddSkip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assertRegistryData(t, "AddSkip", storage, tt.wantData)
			}
		})
	}
}

func TestRegistry_RemoveSkip(t *testing.T) {
	tests := []struct {
		name      string
		setupData skipregistry.RegistryData
		dir       skipregistry.DirectoryPath
		skipType  skipregistry.SkipType
		wantData  skipregistry.RegistryData
		wantErr   bool
	}{
		{
			name: "remove lint keeps test",
			setupData: skipregistry.RegistryData{
				"/project": {"lint", "test"},
			},
			dir:      "/project",
			skipType: skipregistry.SkipTypeLint,
			wantData: skipregistry.RegistryData{
				"/project": {"test"},
			},
			wantErr: false,
		},
		{
			name: "remove last type removes entry",
			setupData: skipregistry.RegistryData{
				"/project": {"lint"},
			},
			dir:      "/project",
			skipType: skipregistry.SkipTypeLint,
			wantData: skipregistry.RegistryData{},
			wantErr:  false,
		},
		{
			name: "remove all removes both",
			setupData: skipregistry.RegistryData{
				"/project": {"lint", "test"},
			},
			dir:      "/project",
			skipType: skipregistry.SkipTypeAll,
			wantData: skipregistry.RegistryData{},
			wantErr:  false,
		},
		{
			name: "remove from non-existent is idempotent",
			setupData: skipregistry.RegistryData{
				"/other": {"lint"},
			},
			dir:      "/project",
			skipType: skipregistry.SkipTypeLint,
			wantData: skipregistry.RegistryData{
				"/other": {"lint"},
			},
			wantErr: false,
		},
		{
			name:      "invalid path returns error",
			setupData: skipregistry.RegistryData{},
			dir:       "relative/path",
			skipType:  skipregistry.SkipTypeLint,
			wantData:  skipregistry.RegistryData{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newMockStorage()
			storage.data = tt.setupData
			r := skipregistry.NewRegistry(storage)

			err := r.RemoveSkip(context.Background(), tt.dir, tt.skipType)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveSkip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assertRegistryData(t, "RemoveSkip", storage, tt.wantData)
			}
		})
	}
}

// assertClearData validates the storage data after a Clear operation.
func assertClearData(t *testing.T, storage *mockStorage, wantData skipregistry.RegistryData) {
	t.Helper()

	if len(storage.data) != len(wantData) {
		t.Errorf("Clear() data length = %v, want %v", len(storage.data), len(wantData))
		return
	}

	for path, types := range wantData {
		gotTypes := storage.data[path]
		if len(gotTypes) != len(types) {
			t.Errorf("Clear() types for %s = %v, want %v", path, gotTypes, types)
		}
	}
}

func TestRegistry_Clear(t *testing.T) {
	tests := []struct {
		name      string
		setupData skipregistry.RegistryData
		dir       skipregistry.DirectoryPath
		wantData  skipregistry.RegistryData
		wantErr   bool
	}{
		{
			name: "clear removes all types",
			setupData: skipregistry.RegistryData{
				"/project": {"lint", "test"},
				"/other":   {"lint"},
			},
			dir: "/project",
			wantData: skipregistry.RegistryData{
				"/other": {"lint"},
			},
			wantErr: false,
		},
		{
			name: "clear non-existent is idempotent",
			setupData: skipregistry.RegistryData{
				"/other": {"lint"},
			},
			dir: "/project",
			wantData: skipregistry.RegistryData{
				"/other": {"lint"},
			},
			wantErr: false,
		},
		{
			name:      "invalid path returns error",
			setupData: skipregistry.RegistryData{},
			dir:       "relative/path",
			wantData:  skipregistry.RegistryData{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newMockStorage()
			storage.data = tt.setupData
			r := skipregistry.NewRegistry(storage)

			err := r.Clear(context.Background(), tt.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("Clear() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assertClearData(t, storage, tt.wantData)
			}
		})
	}
}

func TestRegistry_ListAll(t *testing.T) {
	tests := []struct {
		name      string
		setupData skipregistry.RegistryData
		wantCount int
		wantErr   bool
	}{
		{
			name:      "empty registry returns empty list",
			setupData: skipregistry.RegistryData{},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "returns all entries",
			setupData: skipregistry.RegistryData{
				"/project1": {"lint"},
				"/project2": {"test"},
				"/project3": {"lint", "test"},
			},
			wantCount: 3,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newMockStorage()
			storage.data = tt.setupData
			r := skipregistry.NewRegistry(storage)

			got, err := r.ListAll(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("ListAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("ListAll() returned %v entries, want %v", len(got), tt.wantCount)
			}
		})
	}
}

func TestRegistry_SaveError(t *testing.T) {
	// Test that save errors are handled properly.
	storage := newMockStorage()
	storage.saveErr = errors.New("save failed")
	r := skipregistry.NewRegistry(storage)

	// Try to add a skip.
	err := r.AddSkip(context.Background(), "/project", skipregistry.SkipTypeLint)
	if err == nil {
		t.Errorf("AddSkip() should have returned error when save fails")
	}

	// Verify the cache was reverted.
	isSkipped, _ := r.IsSkipped(context.Background(), "/project", skipregistry.SkipTypeLint)
	if isSkipped {
		t.Errorf("Cache should have been reverted after save failure")
	}
}

func TestRegistry_LoadOnce(t *testing.T) {
	// Test that Load is only called once.
	storage := newMockStorage()
	storage.data = skipregistry.RegistryData{
		"/project": {"lint"},
	}
	r := skipregistry.NewRegistry(storage)

	// First call should load.
	_, _ = r.IsSkipped(context.Background(), "/project", skipregistry.SkipTypeLint)
	if storage.loadCalls != 1 {
		t.Errorf("Expected 1 load call, got %d", storage.loadCalls)
	}

	// Second call should not load again.
	_, _ = r.IsSkipped(context.Background(), "/project", skipregistry.SkipTypeTest)
	if storage.loadCalls != 1 {
		t.Errorf("Expected 1 load call after second operation, got %d", storage.loadCalls)
	}
}

func TestParseSkipType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    skipregistry.SkipType
		wantErr bool
	}{
		{
			name:    "parse lint",
			input:   "lint",
			want:    skipregistry.SkipTypeLint,
			wantErr: false,
		},
		{
			name:    "parse test",
			input:   "test",
			want:    skipregistry.SkipTypeTest,
			wantErr: false,
		},
		{
			name:    "parse all",
			input:   "all",
			want:    skipregistry.SkipTypeAll,
			wantErr: false,
		},
		{
			name:    "invalid type",
			input:   "invalid",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := skipregistry.ParseSkipType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSkipType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseSkipType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDirectoryPath_Validate(t *testing.T) {
	tests := []struct {
		name    string
		path    skipregistry.DirectoryPath
		wantErr bool
	}{
		{
			name:    "absolute path is valid",
			path:    "/project",
			wantErr: false,
		},
		{
			name:    "relative path is invalid",
			path:    "project",
			wantErr: true,
		},
		{
			name:    "relative path with dots is invalid",
			path:    "./project",
			wantErr: true,
		},
		{
			name:    "empty path is invalid",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.path.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DirectoryPath.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
