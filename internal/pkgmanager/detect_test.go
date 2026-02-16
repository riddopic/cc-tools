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
			name:            "preserves existing PREFERRED_PACKAGE_MANAGER",
			existingContent: "PREFERRED_PACKAGE_MANAGER=bun\n",
			manager:         "npm",
			wantContent:     "PREFERRED_PACKAGE_MANAGER=bun\n",
		},
		{
			name:            "appends when other vars exist but no PREFERRED_PACKAGE_MANAGER",
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

func TestDetectWithPreferred(t *testing.T) {
	tests := []struct {
		name      string
		lockFiles []string
		envVar    string
		preferred string
		want      string
	}{
		{
			name:      "preferred overrides lock file",
			lockFiles: []string{"yarn.lock"},
			envVar:    "",
			preferred: "bun",
			want:      "bun",
		},
		{
			name:      "preferred overrides env var",
			lockFiles: nil,
			envVar:    "pnpm",
			preferred: "bun",
			want:      "bun",
		},
		{
			name:      "empty preferred falls through to env var",
			lockFiles: nil,
			envVar:    "yarn",
			preferred: "",
			want:      "yarn",
		},
		{
			name:      "empty preferred falls through to lock file",
			lockFiles: []string{"pnpm-lock.yaml"},
			envVar:    "",
			preferred: "",
			want:      "pnpm",
		},
		{
			name:      "empty preferred and no lock file defaults to npm",
			lockFiles: nil,
			envVar:    "",
			preferred: "",
			want:      "npm",
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

			got := pkgmanager.DetectWithPreferred(projectDir, tt.preferred)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWriteToEnvFile_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "claude.env")

	// Call WriteToEnvFile three times with the same manager.
	for range 3 {
		err := pkgmanager.WriteToEnvFile(envFile, "bun")
		require.NoError(t, err)
	}

	got, err := os.ReadFile(envFile)
	require.NoError(t, err)
	assert.Equal(t, "PREFERRED_PACKAGE_MANAGER=bun\n", string(got),
		"file should contain exactly one entry after multiple writes")
}

func TestWriteToEnvFileError(t *testing.T) {
	err := pkgmanager.WriteToEnvFile("/nonexistent/path/to/file.env", "npm")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "env file")
}
