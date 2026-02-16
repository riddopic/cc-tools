//go:build testmode

package main

import (
	"bytes"
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/output"
	"github.com/riddopic/cc-tools/internal/skipregistry"
)

// testMockStorage is a simple in-memory Storage implementation for testing.
type testMockStorage struct {
	mu   sync.Mutex
	data skipregistry.RegistryData
}

func newTestMockStorage() *testMockStorage {
	return &testMockStorage{
		mu:   sync.Mutex{},
		data: make(skipregistry.RegistryData),
	}
}

func (m *testMockStorage) Load(_ context.Context) (skipregistry.RegistryData, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	dataCopy := make(skipregistry.RegistryData)
	for k, v := range m.data {
		types := make([]string, 0, len(v))
		dataCopy[k] = append(types, v...)
	}
	return dataCopy, nil
}

func (m *testMockStorage) Save(_ context.Context, data skipregistry.RegistryData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(skipregistry.RegistryData)
	for k, v := range data {
		types := make([]string, 0, len(v))
		m.data[k] = append(types, v...)
	}
	return nil
}

func newSkipTestTerminal(t *testing.T) (*output.Terminal, *bytes.Buffer) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	out := output.NewTerminal(&stdout, &stderr)
	return out, &stdout
}

func TestAddSkip(t *testing.T) {
	tests := []struct {
		name       string
		skipType   skipregistry.SkipType
		wantSubstr string
	}{
		{
			name:       "add lint skip",
			skipType:   skipregistry.SkipTypeLint,
			wantSubstr: "Linting will be skipped",
		},
		{
			name:       "add test skip",
			skipType:   skipregistry.SkipTypeTest,
			wantSubstr: "Testing will be skipped",
		},
		{
			name:       "add all skip",
			skipType:   skipregistry.SkipTypeAll,
			wantSubstr: "Linting and testing will be skipped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			t.Chdir(tmpDir)

			storage := newTestMockStorage()
			registry := skipregistry.NewRegistry(storage)
			out, stdout := newSkipTestTerminal(t)
			ctx := context.Background()

			err := addSkip(ctx, out, registry, tt.skipType)
			require.NoError(t, err)
			assert.Contains(t, stdout.String(), tt.wantSubstr)
		})
	}
}

func TestRemoveSkip(t *testing.T) {
	tests := []struct {
		name       string
		addType    skipregistry.SkipType
		removeType skipregistry.SkipType
		wantSubstr string
	}{
		{
			name:       "remove lint skip",
			addType:    skipregistry.SkipTypeLint,
			removeType: skipregistry.SkipTypeLint,
			wantSubstr: "Linting will no longer be skipped",
		},
		{
			name:       "remove test skip",
			addType:    skipregistry.SkipTypeTest,
			removeType: skipregistry.SkipTypeTest,
			wantSubstr: "Testing will no longer be skipped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			t.Chdir(tmpDir)

			storage := newTestMockStorage()
			registry := skipregistry.NewRegistry(storage)
			ctx := context.Background()

			// First add a skip.
			addOut, _ := newSkipTestTerminal(t)
			addErr := addSkip(ctx, addOut, registry, tt.addType)
			require.NoError(t, addErr)

			// Then remove it.
			out, stdout := newSkipTestTerminal(t)
			err := removeSkip(ctx, out, registry, tt.removeType)
			require.NoError(t, err)
			assert.Contains(t, stdout.String(), tt.wantSubstr)
		})
	}
}

func TestClearSkips(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	storage := newTestMockStorage()
	registry := skipregistry.NewRegistry(storage)
	ctx := context.Background()

	// Add some skips first.
	addOut, _ := newSkipTestTerminal(t)
	require.NoError(t, addSkip(ctx, addOut, registry, skipregistry.SkipTypeAll))

	// Clear all skips.
	out, stdout := newSkipTestTerminal(t)
	err := clearSkips(ctx, out, registry)
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "All skips removed")

	// Verify the skips are gone.
	types, getErr := registry.GetSkipTypes(ctx, skipregistry.DirectoryPath(tmpDir))
	require.NoError(t, getErr)
	assert.Empty(t, types)
}

func TestListSkips(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		storage := newTestMockStorage()
		registry := skipregistry.NewRegistry(storage)
		out, stdout := newSkipTestTerminal(t)
		ctx := context.Background()

		err := listSkips(ctx, out, registry)
		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "No directories have skip configurations")
	})

	t.Run("populated list", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		storage := newTestMockStorage()
		registry := skipregistry.NewRegistry(storage)
		ctx := context.Background()

		// Add a skip entry.
		addOut, _ := newSkipTestTerminal(t)
		require.NoError(t, addSkip(ctx, addOut, registry, skipregistry.SkipTypeLint))

		out, stdout := newSkipTestTerminal(t)
		err := listSkips(ctx, out, registry)
		require.NoError(t, err)

		outputStr := stdout.String()
		assert.Contains(t, outputStr, "Skip configurations")
		assert.Contains(t, outputStr, "Directory")
		assert.Contains(t, outputStr, "lint")
	})
}

func TestShowStatus(t *testing.T) {
	t.Run("no skips configured", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		storage := newTestMockStorage()
		registry := skipregistry.NewRegistry(storage)
		out, stdout := newSkipTestTerminal(t)
		ctx := context.Background()

		err := showStatus(ctx, out, registry)
		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "No skips configured")
	})

	t.Run("lint skipped", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		storage := newTestMockStorage()
		registry := skipregistry.NewRegistry(storage)
		ctx := context.Background()

		// Add lint skip.
		addOut, _ := newSkipTestTerminal(t)
		require.NoError(t, addSkip(ctx, addOut, registry, skipregistry.SkipTypeLint))

		out, stdout := newSkipTestTerminal(t)
		err := showStatus(ctx, out, registry)
		require.NoError(t, err)

		outputStr := stdout.String()
		assert.Contains(t, outputStr, "SKIPPED")
		assert.Contains(t, outputStr, "Active")
	})
}
