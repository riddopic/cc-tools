package shared_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/riddopic/cc-tools/internal/shared"
)

// Mock implementations for testing.
type mockFileSystem struct {
	statFunc  func(name string) (os.FileInfo, error)
	getwdFunc func() (string, error)
	absFunc   func(name string) (string, error)
}

func (m *mockFileSystem) Stat(name string) (os.FileInfo, error) {
	if m.statFunc != nil {
		return m.statFunc(name)
	}
	return nil, os.ErrNotExist
}

func (m *mockFileSystem) Getwd() (string, error) {
	if m.getwdFunc != nil {
		return m.getwdFunc()
	}
	return "/home/user/project", nil
}

func (m *mockFileSystem) Abs(path string) (string, error) {
	if m.absFunc != nil {
		return m.absFunc(path)
	}
	// Simple mock implementation.
	if path == "" || path[0] != '/' {
		return filepath.Join("/home/user/project", path), nil
	}
	return path, nil
}

// Mock FileInfo implementation.
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return m.size }
func (m mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m mockFileInfo) ModTime() time.Time { return m.modTime }
func (m mockFileInfo) IsDir() bool        { return m.isDir }

func (m mockFileInfo) Sys() any { return nil }

// newMockFileInfo creates a mockFileInfo with all fields explicitly set.
func newMockFileInfo(name string, isDir bool) mockFileInfo {
	return mockFileInfo{
		name:    name,
		size:    0,
		mode:    0,
		modTime: time.Time{},
		isDir:   isDir,
	}
}

// newMockFS creates a mockFileSystem with all fields explicitly set.
func newMockFS(
	statFunc func(string) (os.FileInfo, error),
	getwdFunc func() (string, error),
	absFunc func(string) (string, error),
) *mockFileSystem {
	return &mockFileSystem{
		statFunc:  statFunc,
		getwdFunc: getwdFunc,
		absFunc:   absFunc,
	}
}

// identityAbs returns an Abs function that returns paths unchanged.
func identityAbs() func(string) (string, error) {
	return func(path string) (string, error) {
		return path, nil
	}
}

// assertFindProjectRoot validates a single FindProjectRoot test case.
func assertFindProjectRoot(t *testing.T, startDir string, mockFS shared.FS, expected string, expectErr bool) {
	t.Helper()

	deps := &shared.Dependencies{FS: mockFS}
	result, err := shared.FindProjectRoot(startDir, deps)

	if expectErr {
		if err == nil {
			t.Errorf("expected error but got none")
		}
		return
	}

	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFindProjectRoot(t *testing.T) {
	tests := []struct {
		name      string
		startDir  string
		mockFS    shared.FS
		expected  string
		expectErr bool
	}{
		{
			name:      "finds git project root",
			startDir:  "/home/user/project/src/nested",
			mockFS:    newMockFS(statForDir("/home/user/project/.git", ".git"), nil, identityAbs()),
			expected:  "/home/user/project",
			expectErr: false,
		},
		{
			name:      "finds go.mod project root",
			startDir:  "/home/user/goproject/internal",
			mockFS:    newMockFS(statForFile("/home/user/goproject/go.mod", "go.mod"), nil, identityAbs()),
			expected:  "/home/user/goproject",
			expectErr: false,
		},
		{
			name:      "finds package.json project root",
			startDir:  "/home/user/jsproject/src",
			mockFS:    newMockFS(statForFile("/home/user/jsproject/package.json", "package.json"), nil, identityAbs()),
			expected:  "/home/user/jsproject",
			expectErr: false,
		},
		{
			name:      "returns original dir when no root found",
			startDir:  "/tmp/no-project",
			mockFS:    newMockFS(nothingExists(), nil, identityAbs()),
			expected:  "/tmp/no-project",
			expectErr: false,
		},
		{
			name:     "uses working dir when startDir is empty",
			startDir: "",
			mockFS: newMockFS(
				statForDir("/home/user/current/.git", ".git"),
				func() (string, error) { return "/home/user/current", nil },
				func(path string) (string, error) {
					if path == "" {
						return "/home/user/current", nil
					}
					return path, nil
				},
			),
			expected:  "/home/user/current",
			expectErr: false,
		},
		{
			name:      "error getting working directory",
			startDir:  "",
			mockFS:    newMockFS(nil, failingGetwd("permission denied"), nil),
			expected:  "",
			expectErr: true,
		},
		{
			name:      "error getting absolute path",
			startDir:  "relative/path",
			mockFS:    newMockFS(nil, nil, failingAbs("invalid path")),
			expected:  "",
			expectErr: true,
		},
		{
			name:      "finds Makefile project root",
			startDir:  "/home/user/makeproject/src",
			mockFS:    newMockFS(statForFile("/home/user/makeproject/Makefile", "Makefile"), nil, identityAbs()),
			expected:  "/home/user/makeproject",
			expectErr: false,
		},
		{
			name:      "finds justfile project root (lowercase)",
			startDir:  "/home/user/justproject/src",
			mockFS:    newMockFS(statForFile("/home/user/justproject/justfile", "justfile"), nil, identityAbs()),
			expected:  "/home/user/justproject",
			expectErr: false,
		},
		{
			name:     "finds Taskfile.yml project root",
			startDir: "/home/user/taskproject/src",
			mockFS: newMockFS(
				statForFile("/home/user/taskproject/Taskfile.yml", "Taskfile.yml"),
				nil,
				identityAbs(),
			),
			expected:  "/home/user/taskproject",
			expectErr: false,
		},
		{
			name:      "finds Cargo.toml project root",
			startDir:  "/home/user/rustproject/src",
			mockFS:    newMockFS(statForFile("/home/user/rustproject/Cargo.toml", "Cargo.toml"), nil, identityAbs()),
			expected:  "/home/user/rustproject",
			expectErr: false,
		},
		{
			name:     "finds pyproject.toml project root",
			startDir: "/home/user/pythonproject/lib",
			mockFS: newMockFS(
				statForFile("/home/user/pythonproject/pyproject.toml", "pyproject.toml"),
				nil,
				identityAbs(),
			),
			expected:  "/home/user/pythonproject",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertFindProjectRoot(t, tt.startDir, tt.mockFS, tt.expected, tt.expectErr)
		})
	}
}

// assertDetectProjectType validates a single DetectProjectType test case.
func assertDetectProjectType(t *testing.T, projectDir string, mockFS shared.FS, expected []string) {
	t.Helper()

	deps := &shared.Dependencies{FS: mockFS}
	result := shared.DetectProjectType(projectDir, deps)

	if len(result) != len(expected) {
		t.Errorf("expected %v, got %v", expected, result)
		return
	}

	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("expected %v, got %v", expected, result)
			return
		}
	}
}

// statForFile returns a stat function that recognizes only the given path as a file.
func statForFile(matchPath, fileName string) func(string) (os.FileInfo, error) {
	return func(name string) (os.FileInfo, error) {
		if name == matchPath {
			return newMockFileInfo(fileName, false), nil
		}
		return nil, os.ErrNotExist
	}
}

// statForDir returns a stat function that recognizes only the given path as a directory.
func statForDir(matchPath, dirName string) func(string) (os.FileInfo, error) {
	return func(name string) (os.FileInfo, error) {
		if name == matchPath {
			return newMockFileInfo(dirName, true), nil
		}
		return nil, os.ErrNotExist
	}
}

// nothingExists returns a stat function where no files exist.
func nothingExists() func(string) (os.FileInfo, error) {
	return func(_ string) (os.FileInfo, error) {
		return nil, os.ErrNotExist
	}
}

// failingGetwd returns a Getwd function that always returns an error.
func failingGetwd(msg string) func() (string, error) {
	return func() (string, error) {
		return "", errors.New(msg)
	}
}

// failingAbs returns an Abs function that always returns an error.
func failingAbs(msg string) func(string) (string, error) {
	return func(_ string) (string, error) {
		return "", errors.New(msg)
	}
}

func TestDetectProjectType(t *testing.T) {
	tests := []struct {
		name       string
		projectDir string
		mockFS     shared.FS
		expected   []string
	}{
		{
			name:       "go project with go.mod",
			projectDir: "/project",
			mockFS:     newMockFS(statForFile("/project/go.mod", "go.mod"), nil, nil),
			expected:   []string{"go"},
		},
		{
			name:       "go project with go.sum",
			projectDir: "/project",
			mockFS:     newMockFS(statForFile("/project/go.sum", "go.sum"), nil, nil),
			expected:   []string{"go"},
		},
		{
			name:       "python project with pyproject.toml",
			projectDir: "/project",
			mockFS:     newMockFS(statForFile("/project/pyproject.toml", "pyproject.toml"), nil, nil),
			expected:   []string{"python"},
		},
		{
			name:       "python project with setup.py",
			projectDir: "/project",
			mockFS:     newMockFS(statForFile("/project/setup.py", "setup.py"), nil, nil),
			expected:   []string{"python"},
		},
		{
			name:       "python project with requirements.txt",
			projectDir: "/project",
			mockFS:     newMockFS(statForFile("/project/requirements.txt", "requirements.txt"), nil, nil),
			expected:   []string{"python"},
		},
		{
			name:       "javascript project with package.json",
			projectDir: "/project",
			mockFS:     newMockFS(statForFile("/project/package.json", "package.json"), nil, nil),
			expected:   []string{"javascript"},
		},
		{
			name:       "typescript project with tsconfig.json",
			projectDir: "/project",
			mockFS:     newMockFS(statForFile("/project/tsconfig.json", "tsconfig.json"), nil, nil),
			expected:   []string{"javascript"},
		},
		{
			name:       "rust project",
			projectDir: "/project",
			mockFS:     newMockFS(statForFile("/project/Cargo.toml", "Cargo.toml"), nil, nil),
			expected:   []string{"rust"},
		},
		{
			name:       "nix project with flake.nix",
			projectDir: "/project",
			mockFS:     newMockFS(statForFile("/project/flake.nix", "flake.nix"), nil, nil),
			expected:   []string{"nix"},
		},
		{
			name:       "nix project with default.nix",
			projectDir: "/project",
			mockFS:     newMockFS(statForFile("/project/default.nix", "default.nix"), nil, nil),
			expected:   []string{"nix"},
		},
		{
			name:       "nix project with shell.nix",
			projectDir: "/project",
			mockFS:     newMockFS(statForFile("/project/shell.nix", "shell.nix"), nil, nil),
			expected:   []string{"nix"},
		},
		{
			name:       "multi-language project",
			projectDir: "/project",
			mockFS: newMockFS(
				func(name string) (os.FileInfo, error) {
					switch name {
					case "/project/go.mod":
						return newMockFileInfo("go.mod", false), nil
					case "/project/package.json":
						return newMockFileInfo("package.json", false), nil
					case "/project/requirements.txt":
						return newMockFileInfo("requirements.txt", false), nil
					}
					return nil, os.ErrNotExist
				},
				nil,
				nil,
			),
			expected: []string{"go", "python", "javascript"},
		},
		{
			name:       "unknown project type",
			projectDir: "/project",
			mockFS: newMockFS(
				func(_ string) (os.FileInfo, error) {
					return nil, os.ErrNotExist
				},
				nil,
				nil,
			),
			expected: []string{"unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertDetectProjectType(t, tt.projectDir, tt.mockFS, tt.expected)
		})
	}
}

func TestGetPackageManager(t *testing.T) {
	tests := []struct {
		name       string
		projectDir string
		mockFS     shared.FS
		expected   string
	}{
		{
			name:       "yarn project",
			projectDir: "/project",
			mockFS:     newMockFS(statForFile("/project/yarn.lock", "yarn.lock"), nil, nil),
			expected:   "yarn",
		},
		{
			name:       "pnpm project",
			projectDir: "/project",
			mockFS:     newMockFS(statForFile("/project/pnpm-lock.yaml", "pnpm-lock.yaml"), nil, nil),
			expected:   "pnpm",
		},
		{
			name:       "bun project",
			projectDir: "/project",
			mockFS:     newMockFS(statForFile("/project/bun.lockb", "bun.lockb"), nil, nil),
			expected:   "bun",
		},
		{
			name:       "npm project (default)",
			projectDir: "/project",
			mockFS: newMockFS(
				func(_ string) (os.FileInfo, error) {
					return nil, os.ErrNotExist
				},
				nil,
				nil,
			),
			expected: "npm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := &shared.Dependencies{FS: tt.mockFS}
			result := shared.GetPackageManager(tt.projectDir, deps)

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestShouldSkipFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		// Skip patterns.
		{"vendor directory", "/project/vendor/github.com/lib/file.go", true},
		{"node_modules", "/project/node_modules/package/index.js", true},
		{"build directory", "/project/build/output.js", true},
		{"git directory", "/project/.git/config", true},
		{"dist directory", "/project/dist/bundle.js", true},
		{"python cache", "/project/__pycache__/module.pyc", true},
		{"cache directory", "/project/.cache/data", true},
		{"rust target", "/project/target/release/binary", true},
		{"next.js cache", "/project/.next/static/chunks/main.js", true},

		// Test files.
		{"go test file", "/project/main_test.go", true},
		{"python test file", "/project/test_module_test.py", true},
		{"js test file", "/project/component.test.js", true},
		{"ts test file", "/project/component.test.ts", true},
		{"js spec file", "/project/component.spec.js", true},
		{"ts spec file", "/project/component.spec.ts", true},
		{"jsx test file", "/project/component.test.jsx", true},
		{"tsx test file", "/project/component.test.tsx", true},
		{"jsx spec file", "/project/component.spec.jsx", true},
		{"tsx spec file", "/project/component.spec.tsx", true},

		// Generated files.
		{"generated go file", "/project/api.generated.go", true},
		{"protobuf go file", "/project/service.pb.go", true},
		{"gen go file", "/project/models.gen.go", true},
		{"underscore gen go file", "/project/types_gen.go", true},

		// Should not skip.
		{"regular go file", "/project/main.go", false},
		{"regular js file", "/project/index.js", false},
		{"regular py file", "/project/module.py", false},
		{"vendored but not in vendor dir", "/project/vendored_lib.go", false},
		{"test in name but not suffix", "/project/test_utils.go", false},
		{"generated in name but not suffix", "/project/generated_utils.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shared.ShouldSkipFile(tt.filePath)
			if result != tt.expected {
				t.Errorf("ShouldSkipFile(%q) = %v, expected %v", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestProjectHelper(t *testing.T) {
	t.Run("NewProjectHelper with nil deps", func(t *testing.T) {
		helper := shared.NewProjectHelper(nil)
		if helper == nil {
			t.Fatal("expected non-nil helper")
		}
		if helper.ProjectHelperDeps() == nil {
			t.Error("expected non-nil deps")
		}
	})

	t.Run("NewProjectHelper with deps", func(t *testing.T) {
		deps := &shared.Dependencies{
			FS: newMockFS(nil, nil, nil),
		}
		helper := shared.NewProjectHelper(deps)
		if helper == nil {
			t.Fatal("expected non-nil helper")
		}
		if helper.ProjectHelperDeps() != deps {
			t.Error("expected same deps")
		}
	})
}

func TestFileExistsWithDeps(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		mockFS   shared.FS
		expected bool
	}{
		{
			name:     "file exists",
			path:     "/project/file.txt",
			mockFS:   newMockFS(statForFile("/project/file.txt", "file.txt"), nil, nil),
			expected: true,
		},
		{
			name: "file does not exist",
			path: "/project/missing.txt",
			mockFS: newMockFS(
				func(_ string) (os.FileInfo, error) {
					return nil, os.ErrNotExist
				},
				nil,
				nil,
			),
			expected: false,
		},
		{
			name: "permission error treated as not exists",
			path: "/project/forbidden.txt",
			mockFS: newMockFS(
				func(_ string) (os.FileInfo, error) {
					return nil, os.ErrPermission
				},
				nil,
				nil,
			),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := &shared.Dependencies{FS: tt.mockFS}
			result := shared.ExportFileExists(tt.path, deps)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	// Test convenience wrapper - just ensure it doesn't panic.
	result := shared.ExportFileExists("/probably/nonexistent/file.txt", nil)
	if result {
		t.Log("Unexpectedly found file /probably/nonexistent/file.txt")
	}
}

func TestNewDefaultDependencies(t *testing.T) {
	deps := shared.NewDefaultDependencies()
	if deps == nil {
		t.Fatal("expected non-nil dependencies")
	}
	if deps.FS == nil {
		t.Fatal("expected non-nil filesystem")
	}

	// Test that the real filesystem implementation works.
	_, err := deps.FS.Getwd()
	if err != nil {
		t.Errorf("Getwd failed: %v", err)
	}
}
