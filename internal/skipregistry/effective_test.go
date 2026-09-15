package skipregistry_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/riddopic/cc-tools/internal/skipregistry"
)

// staticStorage serves fixed registry data, or a fixed load error.
type staticStorage struct {
	data skipregistry.RegistryData
	err  error
}

func (s *staticStorage) Load(_ context.Context) (skipregistry.RegistryData, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.data, nil
}

func (s *staticStorage) Save(_ context.Context, _ skipregistry.RegistryData) error {
	return nil
}

func TestEffective(t *testing.T) {
	const (
		root = "/wt/ams-core"
		pkg  = root + "/src/services/scraper"
		dir  = pkg + "/tests"
	)

	notSkipped := skipregistry.Decision{Skipped: false, Source: ""}
	skippedBy := func(source string) skipregistry.Decision {
		return skipregistry.Decision{Skipped: true, Source: skipregistry.DirectoryPath(source)}
	}

	tests := []struct {
		name       string
		data       skipregistry.RegistryData
		storageErr error
		dir        string
		root       string
		wantLint   skipregistry.Decision
		wantTest   skipregistry.Decision
	}{
		{
			name:       "root entry applies to a nested directory and is named as the source",
			data:       skipregistry.RegistryData{root: {"lint", "test"}},
			storageErr: nil,
			dir:        dir,
			root:       root,
			wantLint:   skippedBy(root),
			wantTest:   skippedBy(root),
		},
		{
			name:       "nearest entry is reported as the source",
			data:       skipregistry.RegistryData{root: {"lint", "test"}, pkg: {"lint"}},
			storageErr: nil,
			dir:        dir,
			root:       root,
			wantLint:   skippedBy(pkg),
			wantTest:   skippedBy(root),
		},
		{
			name:       "entry on the directory itself",
			data:       skipregistry.RegistryData{dir: {"test"}},
			storageErr: nil,
			dir:        dir,
			root:       root,
			wantLint:   notSkipped,
			wantTest:   skippedBy(dir),
		},
		{
			name:       "stored all type skips both validations",
			data:       skipregistry.RegistryData{pkg: {"all"}},
			storageErr: nil,
			dir:        dir,
			root:       root,
			wantLint:   skippedBy(pkg),
			wantTest:   skippedBy(pkg),
		},
		{
			name:       "entry above the root does not apply",
			data:       skipregistry.RegistryData{"/wt": {"lint", "test"}},
			storageErr: nil,
			dir:        dir,
			root:       root,
			wantLint:   notSkipped,
			wantTest:   notSkipped,
		},
		{
			name:       "entry on a sibling directory does not apply",
			data:       skipregistry.RegistryData{root + "/src/services/other": {"lint", "test"}},
			storageErr: nil,
			dir:        dir,
			root:       root,
			wantLint:   notSkipped,
			wantTest:   notSkipped,
		},
		{
			name:       "directory outside the root consults only itself",
			data:       skipregistry.RegistryData{"/other": {"lint"}},
			storageErr: nil,
			dir:        "/other/pkg",
			root:       root,
			wantLint:   notSkipped,
			wantTest:   notSkipped,
		},
		{
			name:       "unreadable registry skips nothing",
			data:       nil,
			storageErr: errors.New("registry unreadable"),
			dir:        dir,
			root:       root,
			wantLint:   notSkipped,
			wantTest:   notSkipped,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := skipregistry.NewRegistry(&staticStorage{data: tt.data, err: tt.storageErr})

			skips := skipregistry.Effective(
				context.Background(),
				registry,
				skipregistry.DirectoryPath(tt.dir),
				skipregistry.DirectoryPath(tt.root),
			)

			assert.Equal(t, tt.wantLint, skips.Lint, "lint decision")
			assert.Equal(t, tt.wantTest, skips.Test, "test decision")
		})
	}
}
