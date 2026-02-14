package pkgmanager_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/pkgmanager"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name      string
		lockFiles []string
		envVar    string
		want      string
	}{
		{
			name:      "bun.lock detected",
			lockFiles: []string{"bun.lock"},
			envVar:    "",
			want:      "bun",
		},
		{
			name:      "bun.lockb detected",
			lockFiles: []string{"bun.lockb"},
			envVar:    "",
			want:      "bun",
		},
		{
			name:      "pnpm-lock.yaml detected",
			lockFiles: []string{"pnpm-lock.yaml"},
			envVar:    "",
			want:      "pnpm",
		},
		{
			name:      "yarn.lock detected",
			lockFiles: []string{"yarn.lock"},
			envVar:    "",
			want:      "yarn",
		},
		{
			name:      "package-lock.json detected",
			lockFiles: []string{"package-lock.json"},
			envVar:    "",
			want:      "npm",
		},
		{
			name:      "no lock file defaults to npm",
			lockFiles: nil,
			envVar:    "",
			want:      "npm",
		},
		{
			name:      "multiple lock files uses first in priority order",
			lockFiles: []string{"bun.lock", "yarn.lock", "package-lock.json"},
			envVar:    "",
			want:      "bun",
		},
		{
			name:      "env var overrides lock file detection",
			lockFiles: []string{"yarn.lock"},
			envVar:    "pnpm",
			want:      "pnpm",
		},
		{
			name:      "env var used when no lock files present",
			lockFiles: nil,
			envVar:    "bun",
			want:      "bun",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir := t.TempDir()

			for _, lf := range tt.lockFiles {
				lockPath := filepath.Join(projectDir, lf)
				require.NoError(t, os.WriteFile(lockPath, []byte(""), 0o600))
			}

			if tt.envVar != "" {
				t.Setenv("PREFERRED_PACKAGE_MANAGER", tt.envVar)
			} else {
				t.Setenv("PREFERRED_PACKAGE_MANAGER", "")
			}

			got := pkgmanager.Detect(projectDir)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWriteToEnvFile(t *testing.T) {
	tests := []struct {
		name            string
		existingContent string
		manager         string
		wantContent     string
	}{
		{
			name:            "creates file and writes correct content",
			existingContent: "",
			manager:         "pnpm",
			wantContent:     "PREFERRED_PACKAGE_MANAGER=pnpm\n",
		},
		{
			name:            "appends to existing file",
			existingContent: "SOME_VAR=value\n",
			manager:         "yarn",
			wantContent:     "SOME_VAR=value\nPREFERRED_PACKAGE_MANAGER=yarn\n",
		},
		{
			name:            "written content has correct format for bun",
			existingContent: "",
			manager:         "bun",
			wantContent:     "PREFERRED_PACKAGE_MANAGER=bun\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			envFile := filepath.Join(tmpDir, "claude.env")

			if tt.existingContent != "" {
				require.NoError(t, os.WriteFile(envFile, []byte(tt.existingContent), 0o644))
			}

			err := pkgmanager.WriteToEnvFile(envFile, tt.manager)
			require.NoError(t, err)

			got, err := os.ReadFile(envFile)
			require.NoError(t, err)
			assert.Equal(t, tt.wantContent, string(got))
		})
	}
}

func TestWriteToEnvFileError(t *testing.T) {
	err := pkgmanager.WriteToEnvFile("/nonexistent/path/to/file.env", "npm")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "open env file")
}
