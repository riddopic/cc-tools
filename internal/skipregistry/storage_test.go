package skipregistry_test

import (
	"context"
	"testing"

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
