package skipregistry_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/shared/mocks"
	"github.com/riddopic/cc-tools/internal/skipregistry"
)

func TestJSONStorageSaveUsesInterfaceMethods(t *testing.T) {
	mockFS := mocks.NewMockRegistryFS(t)
	storage := skipregistry.NewJSONStorage(mockFS, "/tmp/test-registry.json")

	mockFS.EXPECT().MkdirAll("/tmp", mock.Anything).Return(nil).Once()
	mockFS.EXPECT().WriteFile("/tmp/test-registry.json.tmp", mock.Anything, mock.Anything).Return(nil).Once()
	mockFS.EXPECT().Rename("/tmp/test-registry.json.tmp", "/tmp/test-registry.json").Return(nil).Once()

	err := storage.Save(context.Background(), skipregistry.RegistryData{})
	require.NoError(t, err)
}

func TestJSONStorage_Load(t *testing.T) {
	tests := []struct {
		name       string
		fileData   []byte
		fileErr    error
		wantData   skipregistry.RegistryData
		wantErr    bool
		wantErrIs  error
		wantLength int
	}{
		{
			name:       "valid JSON with data",
			fileData:   []byte(`{"/project":["lint","test"]}`),
			fileErr:    nil,
			wantData:   nil,
			wantErr:    false,
			wantErrIs:  nil,
			wantLength: 1,
		},
		{
			name:       "file not found returns ErrNotFound",
			fileData:   nil,
			fileErr:    os.ErrNotExist,
			wantData:   nil,
			wantErr:    true,
			wantErrIs:  skipregistry.ErrNotFound,
			wantLength: 0,
		},
		{
			name:       "empty file returns empty RegistryData",
			fileData:   []byte{},
			fileErr:    nil,
			wantData:   nil,
			wantErr:    false,
			wantErrIs:  nil,
			wantLength: 0,
		},
		{
			name:       "corrupt JSON returns error",
			fileData:   []byte(`{invalid json`),
			fileErr:    nil,
			wantData:   nil,
			wantErr:    true,
			wantErrIs:  nil,
			wantLength: 0,
		},
		{
			name:       "empty JSON object returns empty RegistryData",
			fileData:   []byte(`{}`),
			fileErr:    nil,
			wantData:   nil,
			wantErr:    false,
			wantErrIs:  nil,
			wantLength: 0,
		},
		{
			name:       "JSON null returns empty RegistryData",
			fileData:   []byte(`null`),
			fileErr:    nil,
			wantData:   nil,
			wantErr:    false,
			wantErrIs:  nil,
			wantLength: 0,
		},
		{
			name:       "other read error returns wrapped error",
			fileData:   nil,
			fileErr:    errors.New("permission denied"),
			wantData:   nil,
			wantErr:    true,
			wantErrIs:  nil,
			wantLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := mocks.NewMockRegistryFS(t)
			storage := skipregistry.NewJSONStorage(mockFS, "/tmp/test-registry.json")

			mockFS.EXPECT().ReadFile("/tmp/test-registry.json").Return(tt.fileData, tt.fileErr).Once()

			data, err := storage.Load(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrIs != nil {
					require.ErrorIs(t, err, tt.wantErrIs)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, data)
			assert.Len(t, data, tt.wantLength)
		})
	}
}

func TestJSONStorage_Load_ValidDataContent(t *testing.T) {
	// Verify that loaded data has the correct content, not just the correct length.
	mockFS := mocks.NewMockRegistryFS(t)
	storage := skipregistry.NewJSONStorage(mockFS, "/tmp/test-registry.json")

	mockFS.EXPECT().ReadFile("/tmp/test-registry.json").
		Return([]byte(`{"/project":["lint"],"/other":["test","lint"]}`), nil).Once()

	data, err := storage.Load(context.Background())
	require.NoError(t, err)
	require.Len(t, data, 2)

	projectTypes := data["/project"]
	require.Len(t, projectTypes, 1)
	assert.Equal(t, "lint", projectTypes[0])

	otherTypes := data["/other"]
	require.Len(t, otherTypes, 2)
	assert.Equal(t, "test", otherTypes[0])
	assert.Equal(t, "lint", otherTypes[1])
}

func TestDefaultStorage_ReturnsNonNil(t *testing.T) {
	storage := skipregistry.DefaultStorage()
	assert.NotNil(t, storage)
}
