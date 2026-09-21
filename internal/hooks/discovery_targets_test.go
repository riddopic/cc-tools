package hooks_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/hooks"
)

// statOnly reports success for paths ending in name and ErrNotExist otherwise.
func statOnly(name string) func(string) (os.FileInfo, error) {
	return func(path string) (os.FileInfo, error) {
		if strings.HasSuffix(path, name) {
			return hooks.NewMockFileInfo(name, 0, 0, time.Time{}, false), nil
		}
		return nil, os.ErrNotExist
	}
}

// makeTargetRunner succeeds for `make -n <target>` when target is in accepted.
func makeTargetRunner(accepted ...string) func(
	context.Context, string, string, ...string,
) (*hooks.CommandOutput, error) {
	allowed := make(map[string]struct{}, len(accepted))
	for _, target := range accepted {
		allowed[target] = struct{}{}
	}

	return func(_ context.Context, _, name string, args ...string) (*hooks.CommandOutput, error) {
		if name != "make" || len(args) < 3 || args[len(args)-2] != "-n" {
			return nil, errors.New("command failed")
		}
		if _, ok := allowed[args[len(args)-1]]; !ok {
			return nil, errors.New("no such target")
		}
		return &hooks.CommandOutput{Stdout: []byte("ok"), Stderr: nil}, nil
	}
}

// TestDiscoverCommand_MakefileTargetAliases covers projects whose Makefile
// names its targets something other than the literal "lint" or "test". Before
// this, a Makefile with only a `check` target was invisible to discovery.
func TestDiscoverCommand_MakefileTargetAliases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cmdType    hooks.CommandType
		targets    []string
		wantTarget string
	}{
		{
			name:       "exact lint target",
			cmdType:    hooks.CommandTypeLint,
			targets:    []string{"lint"},
			wantTarget: "lint",
		},
		{
			name:       "check stands in for lint",
			cmdType:    hooks.CommandTypeLint,
			targets:    []string{"check"},
			wantTarget: "check",
		},
		{
			name:       "verify stands in for lint",
			cmdType:    hooks.CommandTypeLint,
			targets:    []string{"verify"},
			wantTarget: "verify",
		},
		{
			name:       "validate stands in for lint",
			cmdType:    hooks.CommandTypeLint,
			targets:    []string{"validate"},
			wantTarget: "validate",
		},
		{
			name:       "ci stands in for lint",
			cmdType:    hooks.CommandTypeLint,
			targets:    []string{"ci"},
			wantTarget: "ci",
		},
		{
			name:       "exact lint wins over aliases",
			cmdType:    hooks.CommandTypeLint,
			targets:    []string{"ci", "check", "lint", "verify"},
			wantTarget: "lint",
		},
		{
			name:       "check wins over later aliases",
			cmdType:    hooks.CommandTypeLint,
			targets:    []string{"ci", "verify", "check"},
			wantTarget: "check",
		},
		{
			name:       "exact test target",
			cmdType:    hooks.CommandTypeTest,
			targets:    []string{"test"},
			wantTarget: "test",
		},
		{
			name:       "tests stands in for test",
			cmdType:    hooks.CommandTypeTest,
			targets:    []string{"tests"},
			wantTarget: "tests",
		},
		{
			name:       "exact test wins over tests",
			cmdType:    hooks.CommandTypeTest,
			targets:    []string{"tests", "test"},
			wantTarget: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testDeps := hooks.CreateTestDependencies()
			testDeps.MockFS.StatFunc = statOnly("Makefile")
			testDeps.MockRunner.RunContextFunc = makeTargetRunner(tt.targets...)

			discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
			cmd, err := discovery.DiscoverCommand(context.Background(), tt.cmdType, "/project")

			require.NoError(t, err)
			require.NotNil(t, cmd)
			assert.Equal(t, "make", cmd.Command)
			assert.Equal(t, []string{tt.wantTarget}, cmd.Args)
			assert.Equal(t, "Makefile", cmd.Source)
		})
	}
}

// TestDiscoverCommand_LintAliasesDoNotLeakIntoTest keeps the two candidate
// lists disjoint, so a generic `check` target is never run twice in parallel
// as both the lint and the test command.
func TestDiscoverCommand_LintAliasesDoNotLeakIntoTest(t *testing.T) {
	t.Parallel()

	testDeps := hooks.CreateTestDependencies()
	testDeps.MockFS.StatFunc = statOnly("Makefile")
	testDeps.MockRunner.RunContextFunc = makeTargetRunner("check", "verify", "validate", "ci")

	discovery := hooks.NewCommandDiscovery("/project", 20, testDeps.Dependencies)
	cmd, err := discovery.DiscoverCommand(context.Background(), hooks.CommandTypeTest, "/project")

	require.Error(t, err)
	assert.Nil(t, cmd)
}
