package hooks_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/riddopic/cc-tools/internal/hooks"
	"github.com/riddopic/cc-tools/internal/skipregistry"
)

// mockSkipStorage implements skipregistry.Storage for testing.
type mockSkipStorage struct {
	data skipregistry.RegistryData
	err  error
}

func (m *mockSkipStorage) Load(_ context.Context) (skipregistry.RegistryData, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.data == nil {
		return make(skipregistry.RegistryData), nil
	}
	return m.data, nil
}

func (m *mockSkipStorage) Save(_ context.Context, data skipregistry.RegistryData) error {
	if m.err != nil {
		return m.err
	}
	m.data = data
	return nil
}

// --- Skip validation helpers ---

// setupSkipTestProjectFS configures the mock filesystem for .git, go.mod, and Makefile detection.
func setupSkipTestProjectFS(deps *hooks.TestDependencies, paths ...string) {
	pathSet := make(map[string]bool, len(paths))
	for _, p := range paths {
		pathSet[p] = true
	}
	deps.MockFS.StatFunc = func(name string) (os.FileInfo, error) {
		for p := range pathSet {
			if name == p {
				return hooks.NewMockFileInfo(
					filepath.Base(name),
					0,
					0,
					time.Time{},
					filepath.Base(name) == ".git",
				), nil
			}
		}
		return nil, errors.New("not found")
	}
}

// setupSkipTestRunner configures the mock runner with discovery and execution for skip tests.
func setupSkipTestRunner(deps *hooks.TestDependencies, lintResult, testResult func() (*hooks.CommandOutput, error)) {
	deps.MockRunner.RunContextFunc = makeDiscoveryAndExecRunner(lintResult, testResult)
	deps.MockRunner.LookPathFunc = func(file string) (string, error) {
		if file == "make" {
			return "/usr/bin/make", nil
		}
		return "", errors.New("not found")
	}
}

// assertSkipResults checks BothPassed and that skipped commands were not executed.
func assertSkipResults(t *testing.T, result *hooks.ValidateResult, wantBoth bool, skipConfig *hooks.SkipConfig) {
	t.Helper()
	if result.BothPassed != wantBoth {
		t.Errorf("ExecuteValidations() BothPassed = %v, want %v", result.BothPassed, wantBoth)
	}
	if skipConfig == nil {
		return
	}
	if skipConfig.SkipLint && result.LintResult != nil {
		t.Errorf("Lint was supposed to be skipped but got result: %+v", result.LintResult)
	}
	if skipConfig.SkipTest && result.TestResult != nil {
		t.Errorf("Test was supposed to be skipped but got result: %+v", result.TestResult)
	}
}

func TestExecuteValidationsWithSkip(t *testing.T) {
	projectPaths := []string{"/project/.git", "/project/go.mod", "/project/Makefile"}
	projectPathsNoMakefile := []string{"/project/.git", "/project/go.mod"}

	tests := []struct {
		name       string
		skipConfig *hooks.SkipConfig
		setupMocks func(*hooks.TestDependencies)
		wantBoth   bool
	}{
		{
			name:       "no skip config runs both",
			skipConfig: nil,
			setupMocks: func(td *hooks.TestDependencies) {
				setupSkipTestProjectFS(td, projectPaths...)
				setupSkipTestRunner(td, successOutput("Success"), successOutput("Success"))
			},
			wantBoth: true,
		},
		{
			name: "skip lint only runs test",
			skipConfig: &hooks.SkipConfig{
				SkipLint: true,
				SkipTest: false,
			},
			setupMocks: func(td *hooks.TestDependencies) {
				setupSkipTestProjectFS(td, projectPaths...)
				setupSkipTestRunner(td, nil, successOutput("Test Success"))
			},
			wantBoth: true,
		},
		{
			name: "skip test only runs lint",
			skipConfig: &hooks.SkipConfig{
				SkipLint: false,
				SkipTest: true,
			},
			setupMocks: func(td *hooks.TestDependencies) {
				setupSkipTestProjectFS(td, projectPaths...)
				setupSkipTestRunner(td, successOutput("Lint Success"), nil)
			},
			wantBoth: true,
		},
		{
			name: "skip both returns success without running",
			skipConfig: &hooks.SkipConfig{
				SkipLint: true,
				SkipTest: true,
			},
			setupMocks: func(td *hooks.TestDependencies) {
				setupSkipTestProjectFS(td, projectPathsNoMakefile...)
				td.MockRunner.RunContextFunc = func(_ context.Context, _, _ string, _ ...string) (*hooks.CommandOutput, error) {
					return nil, errors.New("no commands should be run when both are skipped")
				}
			},
			wantBoth: true,
		},
		{
			name: "lint fails when not skipped",
			skipConfig: &hooks.SkipConfig{
				SkipLint: false,
				SkipTest: true,
			},
			setupMocks: func(td *hooks.TestDependencies) {
				setupSkipTestProjectFS(td, projectPaths...)
				setupSkipTestRunner(td, failOutput("lint errors"), nil)
			},
			wantBoth: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDeps := hooks.CreateTestDependencies()
			tt.setupMocks(testDeps)

			executor := hooks.NewParallelValidateExecutor(
				"/project", 10, false, tt.skipConfig, testDeps.Dependencies,
			)

			result, err := executor.ExecuteValidations(
				context.Background(), "/project", "/project/src",
			)
			if err != nil {
				t.Fatalf("ExecuteValidations() error = %v", err)
			}

			assertSkipResults(t, result, tt.wantBoth, tt.skipConfig)
		})
	}
}

func TestValidateCommandWithSkipRegistry(t *testing.T) {
	tests := []struct {
		name         string
		registryData skipregistry.RegistryData
		filePath     string
		wantSkipLint bool
		wantSkipTest bool
	}{
		{
			name: "directory with lint skip",
			registryData: skipregistry.RegistryData{
				"/project/src": []string{"lint"},
			},
			filePath:     "/project/src/main.go",
			wantSkipLint: true,
			wantSkipTest: false,
		},
		{
			name: "directory with test skip",
			registryData: skipregistry.RegistryData{
				"/project/src": []string{"test"},
			},
			filePath:     "/project/src/main.go",
			wantSkipLint: false,
			wantSkipTest: true,
		},
		{
			name: "directory with both skips",
			registryData: skipregistry.RegistryData{
				"/project/src": []string{"lint", "test"},
			},
			filePath:     "/project/src/main.go",
			wantSkipLint: true,
			wantSkipTest: true,
		},
		{
			name:         "directory with no skips",
			registryData: skipregistry.RegistryData{},
			filePath:     "/project/src/main.go",
			wantSkipLint: false,
			wantSkipTest: false,
		},
		{
			name: "different directory not skipped",
			registryData: skipregistry.RegistryData{
				"/other/path": []string{"lint", "test"},
			},
			filePath:     "/project/src/main.go",
			wantSkipLint: false,
			wantSkipTest: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &mockSkipStorage{
				data: tt.registryData,
				err:  nil,
			}
			registry := skipregistry.NewRegistry(mockStorage)

			ctx := context.Background()
			dir := filepath.Dir(tt.filePath)

			assertSkipRegistryResult(ctx, t, registry, dir, tt.wantSkipLint, tt.wantSkipTest)
		})
	}
}

// assertSkipRegistryResult checks IsSkipped results for both lint and test.
func assertSkipRegistryResult(
	ctx context.Context,
	t *testing.T,
	registry *skipregistry.JSONRegistry,
	dir string,
	wantSkipLint, wantSkipTest bool,
) {
	t.Helper()

	skipLint, err := registry.IsSkipped(ctx, skipregistry.DirectoryPath(dir), skipregistry.SkipTypeLint)
	if err != nil {
		t.Fatalf("IsSkipped(lint) error = %v", err)
	}
	if skipLint != wantSkipLint {
		t.Errorf("IsSkipped(lint) = %v, want %v", skipLint, wantSkipLint)
	}

	skipTest, err := registry.IsSkipped(ctx, skipregistry.DirectoryPath(dir), skipregistry.SkipTypeTest)
	if err != nil {
		t.Fatalf("IsSkipped(test) error = %v", err)
	}
	if skipTest != wantSkipTest {
		t.Errorf("IsSkipped(test) = %v, want %v", skipTest, wantSkipTest)
	}
}
