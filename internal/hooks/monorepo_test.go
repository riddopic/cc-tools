package hooks_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/riddopic/cc-tools/internal/hooks"
	"github.com/riddopic/cc-tools/internal/shared"
)

const (
	monoRoot      = "/repo"
	monoPkgDir    = "/repo/src/shared/lib"
	barePyproject = "[project]\nname = \"lib\"\n"
)

// newMonorepoDeps returns test dependencies whose filesystem holds exactly files,
// whose PATH provides ruff and pytest, and whose Makefiles declare every target.
func newMonorepoDeps(t *testing.T, files map[string]string) *hooks.TestDependencies {
	t.Helper()

	testDeps := hooks.CreateTestDependencies()
	testDeps.MockFS.StatFunc = func(path string) (os.FileInfo, error) {
		if _, ok := files[path]; ok {
			return hooks.NewMockFileInfo(filepath.Base(path), 0, 0, hooks.DefaultTime(), false), nil
		}
		return nil, os.ErrNotExist
	}
	testDeps.MockFS.ReadFileFunc = func(path string) ([]byte, error) {
		if content, ok := files[path]; ok {
			return []byte(content), nil
		}
		return nil, os.ErrNotExist
	}
	testDeps.MockRunner.LookPathFunc = func(file string) (string, error) {
		if file == "ruff" || file == "pytest" {
			return "/usr/bin/" + file, nil
		}
		return "", errors.New("not found")
	}
	testDeps.MockRunner.RunContextFunc = func(
		_ context.Context, _, name string, _ ...string,
	) (*hooks.CommandOutput, error) {
		if name == "make" {
			return &hooks.CommandOutput{Stdout: nil, Stderr: nil}, nil
		}
		return nil, errors.New("command failed")
	}
	return testDeps
}

func TestDiscoverCommand_MonorepoPackages(t *testing.T) {
	tests := []struct {
		name        string
		files       map[string]string
		cmdType     hooks.CommandType
		wantCommand string
		wantArgs    []string
		wantDir     string
	}{
		{
			name: "bare package manifest defers lint to repo root Makefile",
			files: map[string]string{
				monoRoot + "/Makefile":         "",
				monoPkgDir + "/pyproject.toml": barePyproject,
			},
			cmdType:     hooks.CommandTypeLint,
			wantCommand: "make",
			wantArgs:    []string{"lint"},
			wantDir:     monoRoot,
		},
		{
			name: "bare package manifest defers test to repo root Makefile",
			files: map[string]string{
				monoRoot + "/Makefile":         "",
				monoPkgDir + "/pyproject.toml": barePyproject,
			},
			cmdType:     hooks.CommandTypeTest,
			wantCommand: "make",
			wantArgs:    []string{"test"},
			wantDir:     monoRoot,
		},
		{
			name: "bare package manifest defers ruff to repo root",
			files: map[string]string{
				monoRoot + "/pyproject.toml":   "[tool.ruff]\nline-length = 120\n",
				monoPkgDir + "/pyproject.toml": barePyproject,
			},
			cmdType:     hooks.CommandTypeLint,
			wantCommand: "ruff",
			wantArgs:    []string{"check", "."},
			wantDir:     monoRoot,
		},
		{
			name: "bare package manifest defers pytest to repo root",
			files: map[string]string{
				monoRoot + "/pyproject.toml":   "[tool.pytest.ini_options]\n",
				monoPkgDir + "/pyproject.toml": barePyproject,
			},
			cmdType:     hooks.CommandTypeTest,
			wantCommand: "pytest",
			wantArgs:    []string{},
			wantDir:     monoRoot,
		},
		{
			name: "package declaring ruff config in pyproject lints in the package",
			files: map[string]string{
				monoRoot + "/Makefile":         "",
				monoPkgDir + "/pyproject.toml": barePyproject + "[tool.ruff.lint]\nselect = [\"E\"]\n",
			},
			cmdType:     hooks.CommandTypeLint,
			wantCommand: "ruff",
			wantArgs:    []string{"check", "."},
			wantDir:     monoPkgDir,
		},
		{
			name: "package with ruff.toml lints in the package",
			files: map[string]string{
				monoRoot + "/Makefile":         "",
				monoPkgDir + "/pyproject.toml": barePyproject,
				monoPkgDir + "/ruff.toml":      "",
			},
			cmdType:     hooks.CommandTypeLint,
			wantCommand: "ruff",
			wantArgs:    []string{"check", "."},
			wantDir:     monoPkgDir,
		},
		{
			name: "package with pytest.ini tests in the package",
			files: map[string]string{
				monoRoot + "/Makefile":         "",
				monoPkgDir + "/pyproject.toml": barePyproject,
				monoPkgDir + "/pytest.ini":     "",
			},
			cmdType:     hooks.CommandTypeTest,
			wantCommand: "pytest",
			wantArgs:    []string{},
			wantDir:     monoPkgDir,
		},
		{
			name: "nested go module still runs in the module",
			files: map[string]string{
				monoRoot + "/Makefile": "",
				monoPkgDir + "/go.mod": "",
			},
			cmdType:     hooks.CommandTypeTest,
			wantCommand: "go",
			wantArgs:    []string{"test", "./..."},
			wantDir:     monoPkgDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDeps := newMonorepoDeps(t, tt.files)
			discovery := hooks.NewCommandDiscovery(monoRoot, 20, testDeps.Dependencies)

			cmd, err := discovery.DiscoverCommand(context.Background(), tt.cmdType, monoPkgDir+"/tests")

			require.NoError(t, err)
			require.NotNil(t, cmd)
			assert.Equal(t, tt.wantCommand, cmd.Command)
			assert.Equal(t, tt.wantArgs, cmd.Args)
			assert.Equal(t, tt.wantDir, cmd.WorkingDir)
		})
	}
}

// failedValidation builds a failed ValidationResult for a command run in dir.
func failedValidation(cmdType hooks.CommandType, dir, command string, args ...string) *hooks.ValidationResult {
	return &hooks.ValidationResult{
		Type:     cmdType,
		Success:  false,
		ExitCode: 1,
		Message:  "",
		Command: &hooks.DiscoveredCommand{
			Type:       cmdType,
			Command:    command,
			Args:       args,
			WorkingDir: dir,
			Source:     "",
		},
		Error: nil,
	}
}

func TestValidateResult_FormatMessage_EachCommandCarriesItsDir(t *testing.T) {
	result := &hooks.ValidateResult{
		LintResult: failedValidation(hooks.CommandTypeLint, "/repo/pkg", "ruff", "check", "."),
		TestResult: failedValidation(hooks.CommandTypeTest, "/repo", "make", "test"),
		BothPassed: false,
	}

	msg := result.FormatMessage()

	assert.Contains(t, msg, "cd /repo/pkg && ruff check .")
	assert.Contains(t, msg, "cd /repo && make test")
}

func TestRunValidateHook_NestedPackageRunsFromRepoRoot(t *testing.T) {
	repo := t.TempDir()
	pkg := filepath.Join(repo, "src", "shared", "lib")
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".git"), 0o750))
	require.NoError(t, os.MkdirAll(pkg, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "Makefile"), []byte("lint:\n\t@true\ntest:\n\t@true\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(pkg, "pyproject.toml"), []byte(barePyproject), 0o600))

	testDeps := newMonorepoDeps(t, nil)
	testDeps.Dependencies.FS = &shared.RealFS{}

	var mu sync.Mutex
	var ran []string
	testDeps.MockRunner.RunContextFunc = func(
		_ context.Context, dir, name string, args ...string,
	) (*hooks.CommandOutput, error) {
		if name == "make" && slices.Contains(args, "-n") {
			return &hooks.CommandOutput{Stdout: nil, Stderr: nil}, nil
		}
		mu.Lock()
		ran = append(ran, dir+": "+name+" "+strings.Join(args, " "))
		mu.Unlock()
		return &hooks.CommandOutput{Stdout: nil, Stderr: []byte("failed")}, errors.New("exit status 1")
	}

	input := &hookcmd.HookInput{
		HookEventName: "PostToolUse",
		ToolName:      "Edit",
		ToolInput:     hooks.MustMarshalJSON(map[string]any{"file_path": filepath.Join(pkg, "mod.py")}),
	}

	exitCode := hooks.RunValidateHook(context.Background(), input, false, 20, 0, testDeps.Dependencies)

	assert.Equal(t, hooks.ExitCodeShowMessage, exitCode)
	assert.ElementsMatch(t, []string{repo + ": make lint", repo + ": make test"}, ran)
	stderr := testDeps.MockStderr.String()
	assert.Contains(t, stderr, "cd "+repo+" && make lint")
	assert.Contains(t, stderr, "cd "+repo+" && make test")
}
