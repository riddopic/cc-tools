package shared_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/riddopic/cc-tools/internal/shared"
)

// statForPaths returns a Stat function that succeeds only for the given paths.
func statForPaths(paths ...string) func(string) (os.FileInfo, error) {
	existing := make(map[string]bool, len(paths))
	for _, p := range paths {
		existing[p] = true
	}
	return func(name string) (os.FileInfo, error) {
		if existing[name] {
			return newMockFileInfo(filepath.Base(name), false), nil
		}
		return nil, os.ErrNotExist
	}
}

func TestFindRepoRoot(t *testing.T) {
	tests := []struct {
		name      string
		startDir  string
		mockFS    shared.FS
		expected  string
		expectErr bool
	}{
		{
			name:     "skips nested package manifests to reach the git root",
			startDir: "/home/user/mono/src/shared/lib/tests",
			mockFS: newMockFS(statForPaths(
				"/home/user/mono/.git",
				"/home/user/mono/pyproject.toml",
				"/home/user/mono/src/shared/lib/pyproject.toml",
			), nil, identityAbs()),
			expected:  "/home/user/mono",
			expectErr: false,
		},
		{
			name:      "falls back to the nearest project marker outside a git repo",
			startDir:  "/home/user/proj/lib",
			mockFS:    newMockFS(statForPaths("/home/user/proj/pyproject.toml"), nil, identityAbs()),
			expected:  "/home/user/proj",
			expectErr: false,
		},
		{
			name:      "returns original dir when no markers exist",
			startDir:  "/tmp/no-project",
			mockFS:    newMockFS(nothingExists(), nil, identityAbs()),
			expected:  "/tmp/no-project",
			expectErr: false,
		},
		{
			name:      "error getting absolute path",
			startDir:  "relative/path",
			mockFS:    newMockFS(nil, nil, failingAbs("invalid path")),
			expected:  "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := &shared.Dependencies{FS: tt.mockFS}
			result, err := shared.FindRepoRoot(tt.startDir, deps)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
