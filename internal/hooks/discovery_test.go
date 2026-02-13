package hooks_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/riddopic/cc-tools/internal/hooks"
)

func testDiscoversMakefileLintTarget(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "Makefile") {
			return hooks.NewMockFileInfo("Makefile", 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, name string, args ...string) (*hooks.CommandOutput, error) {
		if name == "make" && len(args) >= 3 &&
			args[len(args)-1] == "lint" && args[len(args)-2] == "-n" {
			return &hooks.CommandOutput{
				Stdout: []byte("golangci-lint run"),
				Stderr: nil,
			}, nil
		}
		return nil, errors.New("command failed")
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(
		context.Background(),
		hooks.CommandTypeLint,
		"/project",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("Expected to find command")
	}
	if cmd.Command != "make" || len(cmd.Args) != 1 || cmd.Args[0] != "lint" {
		t.Errorf("Unexpected command: %s %v", cmd.Command, cmd.Args)
	}
	if cmd.Source != "Makefile" {
		t.Errorf("Expected source 'Makefile', got %s", cmd.Source)
	}
}

func testDiscoversJustfileRecipe(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.Contains(path, "justfile") {
			return hooks.NewMockFileInfo("justfile", 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, name string, args ...string) (*hooks.CommandOutput, error) {
		if name == "just" && len(args) >= 2 &&
			args[len(args)-1] == "test" && args[len(args)-2] == "--show" {
			return &hooks.CommandOutput{
				Stdout: []byte("test:\n\tgo test ./..."),
				Stderr: nil,
			}, nil
		}
		return nil, errors.New("command failed")
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(
		context.Background(),
		hooks.CommandTypeTest,
		"/project",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("Expected to find command")
	}
	if cmd.Command != "just" || len(cmd.Args) != 1 || cmd.Args[0] != "test" {
		t.Errorf("Unexpected command: %s %v", cmd.Command, cmd.Args)
	}
}

func testDiscoversTaskfileTask(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "Taskfile.yml") {
			return hooks.NewMockFileInfo("Taskfile.yml", 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, name string, args ...string) (*hooks.CommandOutput, error) {
		if name == "task" && len(args) >= 3 &&
			args[len(args)-1] == "lint" && args[len(args)-2] == "--dry" {
			return &hooks.CommandOutput{
				Stdout: []byte("task: golangci-lint run"),
				Stderr: nil,
			}, nil
		}
		return nil, errors.New("command failed")
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(
		context.Background(),
		hooks.CommandTypeLint,
		"/project",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("Expected to find command")
	}
	if cmd.Command != "task" || len(cmd.Args) != 1 || cmd.Args[0] != "lint" {
		t.Errorf("Unexpected command: %s %v", cmd.Command, cmd.Args)
	}
	if cmd.Source != "Taskfile.yml" {
		t.Errorf("Expected source 'Taskfile.yml', got %s", cmd.Source)
	}
}

func testDiscoversNpmScripts(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "package.json") {
			return hooks.NewMockFileInfo("package.json", 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, name string, args ...string) (*hooks.CommandOutput, error) {
		if name == "jq" && len(args) >= 2 {
			if strings.Contains(args[1], "lint") {
				return &hooks.CommandOutput{
					Stdout: []byte(`"eslint ."`),
					Stderr: nil,
				}, nil
			}
		}
		return nil, errors.New("command failed")
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(
		context.Background(),
		hooks.CommandTypeLint,
		"/project",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("Expected to find command")
	}
	if cmd.Command != "npm" {
		t.Errorf("Expected npm, got %s", cmd.Command)
	}
	if len(cmd.Args) != 2 || cmd.Args[0] != "run" || cmd.Args[1] != "lint" {
		t.Errorf("Unexpected args: %v", cmd.Args)
	}
}

func testDetectsYarnFromLockFile(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "package.json") ||
			strings.HasSuffix(path, "yarn.lock") {
			return hooks.NewMockFileInfo(path, 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, name string, args ...string) (*hooks.CommandOutput, error) {
		if name == "jq" && strings.Contains(args[1], "test") {
			return &hooks.CommandOutput{
				Stdout: []byte(`"jest"`),
				Stderr: nil,
			}, nil
		}
		return nil, errors.New("command failed")
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(
		context.Background(),
		hooks.CommandTypeTest,
		"/project",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("Expected to find command")
	}
	if cmd.Command != "yarn" {
		t.Errorf("Expected yarn, got %s", cmd.Command)
	}
}

func testDetectsPnpmFromLockFile(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "package.json") ||
			strings.HasSuffix(path, "pnpm-lock.yaml") {
			return hooks.NewMockFileInfo(path, 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, name string, _ ...string) (*hooks.CommandOutput, error) {
		if name == "jq" {
			return &hooks.CommandOutput{
				Stdout: []byte(`"test script"`),
				Stderr: nil,
			}, nil
		}
		return nil, errors.New("command failed")
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(
		context.Background(),
		hooks.CommandTypeTest,
		"/project",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd.Command != "pnpm" {
		t.Errorf("Expected pnpm, got %s", cmd.Command)
	}
}

func testDiscoversExecutableScript(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "scripts/lint") {
			return hooks.NewMockFileInfo("lint", 0, 0o755, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(
		context.Background(),
		hooks.CommandTypeLint,
		"/project",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("Expected to find command")
	}
	if cmd.Command != "./scripts/lint" {
		t.Errorf("Expected ./scripts/lint, got %s", cmd.Command)
	}
	if cmd.Source != "scripts/" {
		t.Errorf("Expected source 'scripts/', got %s", cmd.Source)
	}
}

func testSkipsNonExecutableScript(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "scripts/test") {
			return hooks.NewMockFileInfo("test", 0, 0o644, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(
		context.Background(),
		hooks.CommandTypeTest,
		"/project",
	)

	if err == nil {
		t.Fatal("Expected error for no command found")
	}
	if cmd != nil {
		t.Fatal("Expected no command for non-executable script")
	}
}

func testDiscoversGoLintWithGolangciLint(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "go.mod") {
			return hooks.NewMockFileInfo("go.mod", 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	testDeps.MockRunner.LookPathFunc = func(file string) (string, error) {
		if file == "golangci-lint" {
			return "/usr/local/bin/golangci-lint", nil
		}
		return "", errors.New("not found")
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(
		context.Background(),
		hooks.CommandTypeLint,
		"/project",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("Expected to find command")
	}
	if cmd.Command != "golangci-lint" {
		t.Errorf("Expected golangci-lint, got %s", cmd.Command)
	}
	if len(cmd.Args) != 1 || cmd.Args[0] != "run" {
		t.Errorf("Expected args [run], got %v", cmd.Args)
	}
}

func testFallsBackToGoVet(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "go.mod") {
			return hooks.NewMockFileInfo("go.mod", 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	testDeps.MockRunner.LookPathFunc = func(_ string) (string, error) {
		return "", errors.New("not found")
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(
		context.Background(),
		hooks.CommandTypeLint,
		"/project",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("Expected to find command")
	}
	if cmd.Command != "go" {
		t.Errorf("Expected go, got %s", cmd.Command)
	}
	if len(cmd.Args) != 2 || cmd.Args[0] != "vet" || cmd.Args[1] != "./..." {
		t.Errorf("Expected args [vet ./...], got %v", cmd.Args)
	}
}

func testDiscoversGoTestCommand(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "go.mod") {
			return hooks.NewMockFileInfo("go.mod", 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(
		context.Background(),
		hooks.CommandTypeTest,
		"/project",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("Expected to find command")
	}
	if cmd.Command != "go" || cmd.Args[0] != "test" || cmd.Args[1] != "./..." {
		t.Errorf("Unexpected command: %s %v", cmd.Command, cmd.Args)
	}
}

func testDiscoversRustClippy(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "Cargo.toml") {
			return hooks.NewMockFileInfo("Cargo.toml", 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(
		context.Background(),
		hooks.CommandTypeLint,
		"/project",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("Expected to find command")
	}
	if cmd.Command != "cargo" {
		t.Errorf("Expected cargo, got %s", cmd.Command)
	}
	expectedArgs := []string{"clippy", "--", "-D", "warnings"}
	if len(cmd.Args) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d", len(expectedArgs), len(cmd.Args))
	}
	for i, arg := range expectedArgs {
		if i < len(cmd.Args) && cmd.Args[i] != arg {
			t.Errorf("Arg %d: expected %s, got %s", i, arg, cmd.Args[i])
		}
	}
}

func testDiscoversRustTest(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "Cargo.toml") {
			return hooks.NewMockFileInfo("Cargo.toml", 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(
		context.Background(),
		hooks.CommandTypeTest,
		"/project",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("Expected to find command")
	}
	if cmd.Command != "cargo" || len(cmd.Args) != 1 || cmd.Args[0] != "test" {
		t.Errorf("Unexpected command: %s %v", cmd.Command, cmd.Args)
	}
}

func testDiscoversPythonRuff(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "pyproject.toml") {
			return hooks.NewMockFileInfo(
				"pyproject.toml", 0, 0, time.Time{}, false,
			), nil
		}
		return nil, os.ErrNotExist
	}

	testDeps.MockRunner.LookPathFunc = func(file string) (string, error) {
		if file == "ruff" {
			return "/usr/local/bin/ruff", nil
		}
		return "", errors.New("not found")
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(
		context.Background(),
		hooks.CommandTypeLint,
		"/project",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("Expected to find command")
	}
	if cmd.Command != "ruff" {
		t.Errorf("Expected ruff, got %s", cmd.Command)
	}
	if len(cmd.Args) != 2 || cmd.Args[0] != "check" || cmd.Args[1] != "." {
		t.Errorf("Expected args [check .], got %v", cmd.Args)
	}
}

func testFallsBackToFlake8(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "setup.py") {
			return hooks.NewMockFileInfo("setup.py", 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	testDeps.MockRunner.LookPathFunc = func(file string) (string, error) {
		if file == "flake8" {
			return "/usr/local/bin/flake8", nil
		}
		return "", errors.New("not found")
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(
		context.Background(),
		hooks.CommandTypeLint,
		"/project",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("Expected to find command")
	}
	if cmd.Command != "flake8" {
		t.Errorf("Expected flake8, got %s", cmd.Command)
	}
}

func testDiscoversPytest(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "requirements.txt") {
			return hooks.NewMockFileInfo(
				"requirements.txt", 0, 0, time.Time{}, false,
			), nil
		}
		return nil, os.ErrNotExist
	}

	testDeps.MockRunner.LookPathFunc = func(file string) (string, error) {
		if file == "pytest" {
			return "/usr/local/bin/pytest", nil
		}
		return "", errors.New("not found")
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(
		context.Background(),
		hooks.CommandTypeTest,
		"/project",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("Expected to find command")
	}
	if cmd.Command != "pytest" {
		t.Errorf("Expected pytest, got %s", cmd.Command)
	}
}

func testFallsBackToUnittest(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "requirements.txt") {
			return hooks.NewMockFileInfo(
				"requirements.txt", 0, 0, time.Time{}, false,
			), nil
		}
		return nil, os.ErrNotExist
	}

	testDeps.MockRunner.LookPathFunc = func(_ string) (string, error) {
		return "", errors.New("not found")
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(
		context.Background(),
		hooks.CommandTypeTest,
		"/project",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("Expected to find command")
	}
	if cmd.Command != "python" {
		t.Errorf("Expected python, got %s", cmd.Command)
	}
	if len(cmd.Args) != 2 || cmd.Args[0] != "-m" || cmd.Args[1] != "unittest" {
		t.Errorf("Expected args [-m unittest], got %v", cmd.Args)
	}
}

func testWalksUpDirectoryTree(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if path == "/project/Makefile" {
			return hooks.NewMockFileInfo("Makefile", 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, name string, args ...string) (*hooks.CommandOutput, error) {
		if name == "make" && args[len(args)-1] == "lint" {
			return &hooks.CommandOutput{
				Stdout: []byte("lint"),
				Stderr: nil,
			}, nil
		}
		return nil, errors.New("command failed")
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(
		context.Background(),
		hooks.CommandTypeLint,
		"/project/src/subdir",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("Expected to find command")
	}
	if cmd.WorkingDir != "/project" {
		t.Errorf("Expected working dir /project, got %s", cmd.WorkingDir)
	}
}

func testStopsAtProjectRoot(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(_ string) (os.FileInfo, error) {
		return nil, os.ErrNotExist
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(
		context.Background(),
		hooks.CommandTypeLint,
		"/project/src",
	)

	if err == nil {
		t.Fatal("Expected error for no command found")
	}
	if cmd != nil {
		t.Fatal("Expected no command")
	}
	if !strings.Contains(err.Error(), "no command found") {
		t.Errorf("Expected 'no command found' error, got: %v", err)
	}
}

func testHandlesTimeout(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "Makefile") {
			return hooks.NewMockFileInfo("Makefile", 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	testDeps.MockRunner.RunContextFunc = func(ctx context.Context, _, _ string, _ ...string) (*hooks.CommandOutput, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	discovery := hooks.NewCommandDiscovery("/project", 0, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(
		context.Background(),
		hooks.CommandTypeLint,
		"/project",
	)

	if err == nil {
		t.Fatal("Expected error due to timeout")
	}
	if cmd != nil {
		t.Fatal("Expected no command due to timeout")
	}
}

func testDetectsMultipleProjectTypes(t *testing.T) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "go.mod") ||
			strings.HasSuffix(path, "package.json") {
			return hooks.NewMockFileInfo(path, 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	types := discovery.DetectProjectTypesForTest("/project")

	hasGo := false
	hasJS := false
	for _, pt := range types {
		if pt == "go" {
			hasGo = true
		}
		if pt == "javascript" {
			hasJS = true
		}
	}

	if !hasGo {
		t.Error("Expected to detect Go project")
	}
	if !hasJS {
		t.Error("Expected to detect JavaScript project")
	}
}

func TestCommandDiscovery(t *testing.T) {
	t.Run("discovers Makefile lint target", testDiscoversMakefileLintTarget)
	t.Run("discovers justfile recipe", testDiscoversJustfileRecipe)
	t.Run("discovers Taskfile task", testDiscoversTaskfileTask)
	t.Run("discovers package.json scripts with npm", testDiscoversNpmScripts)
	t.Run("detects yarn from lock file", testDetectsYarnFromLockFile)
	t.Run("detects pnpm from lock file", testDetectsPnpmFromLockFile)
	t.Run("discovers executable script", testDiscoversExecutableScript)
	t.Run("skips non-executable script", testSkipsNonExecutableScript)
	t.Run("discovers Go lint with golangci-lint", testDiscoversGoLintWithGolangciLint)
	t.Run("falls back to go vet", testFallsBackToGoVet)
	t.Run("discovers Go test command", testDiscoversGoTestCommand)
	t.Run("discovers Rust clippy for lint", testDiscoversRustClippy)
	t.Run("discovers Rust test command", testDiscoversRustTest)
	t.Run("discovers Python lint with ruff", testDiscoversPythonRuff)
	t.Run("falls back to flake8", testFallsBackToFlake8)
	t.Run("discovers Python test with pytest", testDiscoversPytest)
	t.Run("falls back to unittest", testFallsBackToUnittest)
	t.Run("walks up directory tree", testWalksUpDirectoryTree)
	t.Run("stops at project root", testStopsAtProjectRoot)
	t.Run("handles timeout", testHandlesTimeout)
	t.Run("detects multiple project types", testDetectsMultipleProjectTypes)
}

func TestDiscoveredCommandString(t *testing.T) {
	tests := []struct {
		name     string
		cmd      *hooks.DiscoveredCommand
		expected string
	}{
		{
			name:     "nil command",
			cmd:      nil,
			expected: "",
		},
		{
			name: "command without args",
			cmd: &hooks.DiscoveredCommand{
				Type:       "",
				Command:    "make",
				Args:       []string{},
				WorkingDir: "",
				Source:     "",
			},
			expected: "make",
		},
		{
			name: "command with single arg",
			cmd: &hooks.DiscoveredCommand{
				Type:       "",
				Command:    "make",
				Args:       []string{"lint"},
				WorkingDir: "",
				Source:     "",
			},
			expected: "make lint",
		},
		{
			name: "command with multiple args",
			cmd: &hooks.DiscoveredCommand{
				Type:       "",
				Command:    "cargo",
				Args:       []string{"clippy", "--", "-D", "warnings"},
				WorkingDir: "",
				Source:     "",
			},
			expected: "cargo clippy -- -D warnings",
		},
		{
			name: "script path",
			cmd: &hooks.DiscoveredCommand{
				Type:       "",
				Command:    "./scripts/test",
				Args:       []string{},
				WorkingDir: "",
				Source:     "",
			},
			expected: "./scripts/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cmd.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func BenchmarkCommandDiscovery(b *testing.B) {
	testDeps := hooks.CreateTestDependencies()

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, "Makefile") {
			return hooks.NewMockFileInfo("Makefile", 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	testDeps.MockRunner.RunContextFunc = func(_ context.Context, _, name string, args ...string) (*hooks.CommandOutput, error) {
		if name == "make" && args[len(args)-1] == "lint" {
			return &hooks.CommandOutput{
				Stdout: []byte("lint"),
				Stderr: nil,
			}, nil
		}
		return nil, errors.New("command failed")
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)

	b.ResetTimer()
	for range b.N {
		discovery.DiscoverCommand(
			context.Background(),
			hooks.CommandTypeLint,
			"/project",
		)
	}
}

func BenchmarkDetectPackageManager(b *testing.B) {
	testDeps := hooks.CreateTestDependencies()

	lockFiles := []string{
		"yarn.lock",
		"pnpm-lock.yaml",
		"bun.lockb",
		"package-lock.json",
	}
	currentLock := 0

	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, lockFiles[currentLock%len(lockFiles)]) {
			return hooks.NewMockFileInfo(path, 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)

	b.ResetTimer()
	for i := range b.N {
		currentLock = i
		discovery.DetectPackageManagerForTest("/project")
	}
}
