package hooks_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/riddopic/cc-tools/internal/hooks"
	"github.com/riddopic/cc-tools/internal/skipregistry"
)

func TestSkippedTypesForFile(t *testing.T) {
	const (
		root    = "/wt/ams-core"
		pkgDir  = root + "/src/services/scraper"
		fileDir = pkgDir + "/tests"
	)

	tests := []struct {
		name       string
		data       skipregistry.RegistryData
		storageErr error
		wantLint   bool
		wantTest   bool
	}{
		{
			name:       "skip at repo root applies to a nested package",
			data:       skipregistry.RegistryData{root: {"lint", "test"}},
			storageErr: nil,
			wantLint:   true,
			wantTest:   true,
		},
		{
			name:       "skip at a nested package still applies",
			data:       skipregistry.RegistryData{pkgDir: {"test"}},
			storageErr: nil,
			wantLint:   false,
			wantTest:   true,
		},
		{
			name:       "skips from different ancestors combine",
			data:       skipregistry.RegistryData{root: {"lint"}, pkgDir: {"test"}},
			storageErr: nil,
			wantLint:   true,
			wantTest:   true,
		},
		{
			name:       "skip on the file's own directory applies",
			data:       skipregistry.RegistryData{fileDir: {"lint"}},
			storageErr: nil,
			wantLint:   true,
			wantTest:   false,
		},
		{
			name:       "skip above the repo root does not apply",
			data:       skipregistry.RegistryData{"/wt": {"lint", "test"}},
			storageErr: nil,
			wantLint:   false,
			wantTest:   false,
		},
		{
			name:       "skip on a sibling directory does not apply",
			data:       skipregistry.RegistryData{root + "/src/services/other": {"lint", "test"}},
			storageErr: nil,
			wantLint:   false,
			wantTest:   false,
		},
		{
			name:       "no skips",
			data:       skipregistry.RegistryData{},
			storageErr: nil,
			wantLint:   false,
			wantTest:   false,
		},
		{
			name:       "unreadable registry skips nothing",
			data:       nil,
			storageErr: errors.New("registry unreadable"),
			wantLint:   false,
			wantTest:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := skipregistry.NewRegistry(&mockSkipStorage{data: tt.data, err: tt.storageErr})

			skipLint, skipTest := hooks.SkippedTypesForTest(context.Background(), registry, fileDir, root)

			assert.Equal(t, tt.wantLint, skipLint, "skip lint")
			assert.Equal(t, tt.wantTest, skipTest, "skip test")
		})
	}
}
