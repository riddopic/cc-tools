package hooks

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// CommandType represents the type of command to discover.
type CommandType string

const (
	// CommandTypeLint represents lint commands (used internally by validate).
	CommandTypeLint CommandType = "lint"
	// CommandTypeTest represents test commands (used internally by validate).
	CommandTypeTest CommandType = "test"
)

// Project markers and sources used by language-specific discovery.
const (
	goModFile           = "go.mod"
	pythonProjectSource = "Python project"
)

// errNotExecutable marks a scripts/ candidate that exists but lacks the
// executable bit, so probing continues to the next candidate name.
var errNotExecutable = errors.New("script is not executable")

// targetNames returns the target, task, recipe, and script names probed for
// cmdType, most specific first. The generic names are aliases projects commonly
// use for a combined check. The two lists are kept disjoint so a single generic
// target is never discovered as both the lint and the test command and then run
// twice in parallel.
func targetNames(cmdType CommandType) []string {
	if cmdType == CommandTypeTest {
		return []string{string(CommandTypeTest), "tests"}
	}

	return []string{string(CommandTypeLint), "check", "verify", "validate", "ci"}
}

// firstMatchingTarget returns the first candidate for which probe reports no
// error, or an empty string when none match.
func firstMatchingTarget(candidates []string, probe func(string) error) string {
	for _, candidate := range candidates {
		if probe(candidate) == nil {
			return candidate
		}
	}

	return ""
}

// DiscoveredCommand represents a discovered command.
type DiscoveredCommand struct {
	Type       CommandType
	Command    string
	Args       []string
	WorkingDir string
	Source     string // Where it was found (e.g., "Makefile", "package.json")
}

// CommandDiscovery handles discovering project commands with injected dependencies.
type CommandDiscovery struct {
	projectRoot string
	timeout     int
	debug       bool
	deps        *Dependencies
}

// NewCommandDiscovery creates a new command discovery instance with dependencies.
func NewCommandDiscovery(projectRoot string, timeoutSecs int, deps *Dependencies) *CommandDiscovery {
	if deps == nil {
		deps = NewDefaultDependencies()
	}
	return &CommandDiscovery{
		projectRoot: projectRoot,
		timeout:     timeoutSecs,
		debug:       false,
		deps:        deps,
	}
}

// SetDebug enables debug logging for discovery operations.
func (cd *CommandDiscovery) SetDebug(debug bool) {
	cd.debug = debug
}

// debugf writes a debug message to stderr when debug mode is enabled.
func (cd *CommandDiscovery) debugf(format string, args ...any) {
	if cd.debug {
		_, _ = fmt.Fprintf(cd.deps.Stderr, "[discovery] "+format+"\n", args...)
	}
}

// DiscoverCommand searches for and returns a command of the specified type.
func (cd *CommandDiscovery) DiscoverCommand(
	ctx context.Context,
	cmdType CommandType,
	startDir string,
) (*DiscoveredCommand, error) {
	currentDir := startDir
	if currentDir == "" {
		currentDir = cd.projectRoot
	}

	// Walk up from current directory to project root
	for {
		// Check for Makefile
		if cmd := cd.checkMakefile(ctx, currentDir, cmdType); cmd != nil {
			return cmd, nil
		}

		// Check for Taskfile
		if cmd := cd.checkTaskfile(ctx, currentDir, cmdType); cmd != nil {
			return cmd, nil
		}

		// Check for justfile
		if cmd := cd.checkJustfile(ctx, currentDir, cmdType); cmd != nil {
			return cmd, nil
		}

		// Check for package.json (Node.js)
		if cmd := cd.checkPackageJSON(ctx, currentDir, cmdType); cmd != nil {
			return cmd, nil
		}

		// Check for scripts directory
		if cmd := cd.checkScriptsDir(ctx, currentDir, cmdType); cmd != nil {
			return cmd, nil
		}

		// Check for language-specific tools
		if cmd := cd.checkLanguageSpecific(ctx, currentDir, cmdType); cmd != nil {
			return cmd, nil
		}

		// Stop at project root or filesystem root
		if currentDir == cd.projectRoot || currentDir == "/" {
			break
		}

		// Move up one directory
		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			break
		}
		currentDir = parent
	}

	return nil, fmt.Errorf("no command found for type %s", cmdType)
}

// buildfileProbe describes one build-file family: the filenames to look for,
// the command that runs a target, and the arguments preceding the target name.
type buildfileProbe struct {
	filenames []string
	command   string
	args      func(path string) []string
}

// checkBuildfile looks for the first candidate target defined in any of the
// probe's build files, preferring the exact name over the generic aliases.
func (cd *CommandDiscovery) checkBuildfile(
	ctx context.Context,
	dir string,
	cmdType CommandType,
	probe buildfileProbe,
) *DiscoveredCommand {
	for _, filename := range probe.filenames {
		path := filepath.Join(dir, filename)
		if _, err := cd.deps.FS.Stat(path); err != nil {
			continue
		}

		target := firstMatchingTarget(targetNames(cmdType), func(candidate string) error {
			timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(cd.timeout)*time.Second)
			defer cancel()
			_, err := cd.deps.Runner.RunContext(
				timeoutCtx, dir, probe.command, append(probe.args(path), candidate)...,
			)

			return err
		})
		if target == "" {
			cd.debugf("%s: no %s target found in %s", probe.command, cmdType, path)

			continue
		}

		return &DiscoveredCommand{
			Type:       cmdType,
			Command:    probe.command,
			Args:       []string{target},
			WorkingDir: dir,
			Source:     filename,
		}
	}

	return nil
}

// checkMakefile checks for Makefile targets.
func (cd *CommandDiscovery) checkMakefile(
	ctx context.Context,
	dir string,
	cmdType CommandType,
) *DiscoveredCommand {
	return cd.checkBuildfile(ctx, dir, cmdType, buildfileProbe{
		filenames: []string{"Makefile", "makefile"},
		command:   "make",
		args:      func(path string) []string { return []string{"-f", path, "-n"} },
	})
}

// checkTaskfile checks for Taskfile tasks.
func (cd *CommandDiscovery) checkTaskfile(
	ctx context.Context,
	dir string,
	cmdType CommandType,
) *DiscoveredCommand {
	return cd.checkBuildfile(ctx, dir, cmdType, buildfileProbe{
		filenames: []string{"Taskfile.yml", "Taskfile.yaml"},
		command:   "task",
		args:      func(path string) []string { return []string{"--taskfile", path, "--dry"} },
	})
}

// checkJustfile checks for justfile recipes.
func (cd *CommandDiscovery) checkJustfile(
	ctx context.Context,
	dir string,
	cmdType CommandType,
) *DiscoveredCommand {
	return cd.checkBuildfile(ctx, dir, cmdType, buildfileProbe{
		filenames: []string{"justfile", "Justfile", ".justfile"},
		command:   "just",
		args:      func(path string) []string { return []string{"--justfile", path, "--show"} },
	})
}

// checkPackageJSON checks for npm/yarn/pnpm scripts.
func (cd *CommandDiscovery) checkPackageJSON(
	ctx context.Context,
	dir string,
	cmdType CommandType,
) *DiscoveredCommand {
	packagePath := filepath.Join(dir, "package.json")
	if _, err := cd.deps.FS.Stat(packagePath); err != nil {
		return nil
	}

	// Use jq to check which candidate script exists
	script := firstMatchingTarget(targetNames(cmdType), func(candidate string) error {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(cd.timeout)*time.Second)
		defer cancel()
		_, err := cd.deps.Runner.RunContext(timeoutCtx, dir, "jq", "-e",
			fmt.Sprintf(".scripts.%q", candidate), packagePath)

		return err
	})
	if script == "" {
		cd.debugf("package.json: no %s script found in %s", cmdType, packagePath)

		return nil
	}

	// Detect package manager
	pm := cd.detectPackageManager(dir)

	return &DiscoveredCommand{
		Type:       cmdType,
		Command:    pm,
		Args:       []string{"run", script},
		WorkingDir: dir,
		Source:     "package.json",
	}
}

// checkScriptsDir checks for executable scripts in scripts/ directory.
func (cd *CommandDiscovery) checkScriptsDir(
	_ context.Context,
	dir string,
	cmdType CommandType,
) *DiscoveredCommand {
	script := firstMatchingTarget(targetNames(cmdType), func(candidate string) error {
		info, err := cd.deps.FS.Stat(filepath.Join(dir, "scripts", candidate))
		if err != nil {
			return err
		}
		// Check if it's executable
		if info.Mode()&0o111 == 0 {
			cd.debugf("scripts/: %s exists but is not executable", candidate)

			return errNotExecutable
		}

		return nil
	})
	if script == "" {
		return nil
	}

	return &DiscoveredCommand{
		Type:       cmdType,
		Command:    "./scripts/" + script,
		Args:       []string{},
		WorkingDir: dir,
		Source:     "scripts/",
	}
}

// checkLanguageSpecific checks for language-specific tools.
func (cd *CommandDiscovery) checkLanguageSpecific(
	ctx context.Context,
	dir string,
	cmdType CommandType,
) *DiscoveredCommand {
	// Check for various project markers
	projectTypes := cd.detectProjectTypes(dir)

	for _, projectType := range projectTypes {
		switch projectType {
		case "go":
			if cmd := cd.checkGoCommands(ctx, dir, cmdType); cmd != nil {
				return cmd
			}
		case "rust":
			if cmd := cd.checkRustCommands(ctx, dir, cmdType); cmd != nil {
				return cmd
			}
		case "python":
			if cmd := cd.checkPythonCommands(ctx, dir, cmdType); cmd != nil {
				return cmd
			}
		}
	}

	return nil
}

// checkGoCommands checks for Go-specific commands.
func (cd *CommandDiscovery) checkGoCommands(
	_ context.Context,
	dir string,
	cmdType CommandType,
) *DiscoveredCommand {
	// Only check if go.mod exists in this directory
	if _, err := cd.deps.FS.Stat(filepath.Join(dir, goModFile)); err != nil {
		return nil
	}

	switch cmdType {
	case CommandTypeLint:
		// Try golangci-lint first
		if _, err := cd.deps.Runner.LookPath("golangci-lint"); err == nil {
			return &DiscoveredCommand{
				Type:       cmdType,
				Command:    "golangci-lint",
				Args:       []string{"run"},
				WorkingDir: dir,
				Source:     goModFile,
			}
		}
		// Fall back to go vet
		return &DiscoveredCommand{
			Type:       cmdType,
			Command:    "go",
			Args:       []string{"vet", "./..."},
			WorkingDir: dir,
			Source:     goModFile,
		}
	case CommandTypeTest:
		return &DiscoveredCommand{
			Type:       cmdType,
			Command:    "go",
			Args:       []string{"test", "./..."},
			WorkingDir: dir,
			Source:     goModFile,
		}
	}

	return nil
}

// checkRustCommands checks for Rust-specific commands.
func (cd *CommandDiscovery) checkRustCommands(
	_ context.Context,
	dir string,
	cmdType CommandType,
) *DiscoveredCommand {
	// Only check if Cargo.toml exists in this directory
	if _, err := cd.deps.FS.Stat(filepath.Join(dir, "Cargo.toml")); err != nil {
		return nil
	}

	switch cmdType {
	case CommandTypeLint:
		return &DiscoveredCommand{
			Type:       cmdType,
			Command:    "cargo",
			Args:       []string{"clippy", "--", "-D", "warnings"},
			WorkingDir: dir,
			Source:     "Cargo.toml",
		}
	case CommandTypeTest:
		return &DiscoveredCommand{
			Type:       cmdType,
			Command:    "cargo",
			Args:       []string{"test"},
			WorkingDir: dir,
			Source:     "Cargo.toml",
		}
	}

	return nil
}

// checkPythonCommands checks for Python-specific commands.
func (cd *CommandDiscovery) checkPythonCommands(
	_ context.Context,
	dir string,
	cmdType CommandType,
) *DiscoveredCommand {
	// Check if this is a Python project directory
	pythonMarkers := []string{"pyproject.toml", "setup.py", "requirements.txt"}
	hasPython := false
	for _, marker := range pythonMarkers {
		if _, err := cd.deps.FS.Stat(filepath.Join(dir, marker)); err == nil {
			hasPython = true
			break
		}
	}

	if !hasPython {
		return nil
	}

	switch cmdType {
	case CommandTypeLint:
		return cd.pythonLintCommand(dir)
	case CommandTypeTest:
		return cd.pythonTestCommand(dir)
	}

	return nil
}

// pythonLintCommand returns the preferred Python linter that can run from dir.
func (cd *CommandDiscovery) pythonLintCommand(dir string) *DiscoveredCommand {
	linters := []struct {
		name string
		args []string
	}{
		{"ruff", []string{"check", "."}},
		{"flake8", []string{"."}},
		{"pylint", []string{"."}},
	}

	for _, linter := range linters {
		if _, err := cd.deps.Runner.LookPath(linter.name); err != nil {
			cd.debugf("python: linter %q not found in PATH", linter.name)
			continue
		}
		if !cd.isPythonToolRoot(dir, linter.name) {
			continue
		}
		return &DiscoveredCommand{
			Type:       CommandTypeLint,
			Command:    linter.name,
			Args:       linter.args,
			WorkingDir: dir,
			Source:     pythonProjectSource,
		}
	}

	return nil
}

// pythonTestCommand returns the Python test runner that can run from dir.
func (cd *CommandDiscovery) pythonTestCommand(dir string) *DiscoveredCommand {
	if !cd.isPythonToolRoot(dir, "pytest") {
		return nil
	}

	if _, err := cd.deps.Runner.LookPath("pytest"); err == nil {
		return &DiscoveredCommand{
			Type:       CommandTypeTest,
			Command:    "pytest",
			Args:       []string{},
			WorkingDir: dir,
			Source:     pythonProjectSource,
		}
	}

	return &DiscoveredCommand{
		Type:       CommandTypeTest,
		Command:    "python",
		Args:       []string{"-m", "unittest"},
		WorkingDir: dir,
		Source:     pythonProjectSource,
	}
}

// isPythonToolRoot reports whether tool should run from dir. In a monorepo a
// package's pyproject.toml marks a build unit, not a tooling root. Accepting it
// ends discovery before the repository's own lint and test targets are found,
// so the bare tool on PATH runs instead of the project's pinned toolchain, and
// pytest takes its rootdir from the package and skips the root conftest.py.
// Below the project root, dir qualifies only if it declares tool's config.
func (cd *CommandDiscovery) isPythonToolRoot(dir, tool string) bool {
	if dir == cd.projectRoot {
		return true
	}

	cfg := pythonToolConfig(tool)
	for _, name := range cfg.files {
		if _, err := cd.deps.FS.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	for name, header := range cfg.sections {
		data, err := cd.deps.FS.ReadFile(filepath.Join(dir, name))
		if err == nil && hasSectionHeader(data, header) {
			return true
		}
	}

	cd.debugf("python: %s declares no %s config, deferring to a parent directory", dir, tool)
	return false
}

// Shared Python config files that can carry sections for several tools.
const (
	pyprojectFile = "pyproject.toml"
	setupCfgFile  = "setup.cfg"
	toxIniFile    = "tox.ini"
)

// toolConfig describes where a Python tool's configuration can be declared.
type toolConfig struct {
	files    []string          // files whose presence alone declares config
	sections map[string]string // file name to the section header prefix that declares config
}

// pythonToolConfig returns the config locations recognised for tool.
func pythonToolConfig(tool string) toolConfig {
	switch tool {
	case "ruff":
		return toolConfig{
			files:    []string{"ruff.toml", ".ruff.toml"},
			sections: map[string]string{pyprojectFile: "[tool.ruff"},
		}
	case "flake8":
		return toolConfig{
			files:    []string{".flake8"},
			sections: map[string]string{setupCfgFile: "[flake8]", toxIniFile: "[flake8]"},
		}
	case "pylint":
		return toolConfig{
			files:    []string{".pylintrc", "pylintrc"},
			sections: map[string]string{pyprojectFile: "[tool.pylint", setupCfgFile: "[pylint"},
		}
	case "pytest":
		return toolConfig{
			files: []string{"pytest.ini"},
			sections: map[string]string{
				pyprojectFile: "[tool.pytest.ini_options]",
				toxIniFile:    "[pytest]",
				setupCfgFile:  "[tool:pytest]",
			},
		}
	}
	return toolConfig{files: nil, sections: nil}
}

// hasSectionHeader reports whether any line of data starts with header.
func hasSectionHeader(data []byte, header string) bool {
	for line := range strings.SplitSeq(string(data), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), header) {
			return true
		}
	}
	return false
}

// detectPackageManager detects which package manager to use based on lock files.
func (cd *CommandDiscovery) detectPackageManager(dir string) string {
	if _, err := cd.deps.FS.Stat(filepath.Join(dir, "yarn.lock")); err == nil {
		return "yarn"
	}
	if _, err := cd.deps.FS.Stat(filepath.Join(dir, "pnpm-lock.yaml")); err == nil {
		return "pnpm"
	}
	if _, err := cd.deps.FS.Stat(filepath.Join(dir, "bun.lockb")); err == nil {
		return "bun"
	}
	return "npm"
}

// detectProjectTypes detects the types of project in the directory.
func (cd *CommandDiscovery) detectProjectTypes(dir string) []string {
	var types []string

	// Go project
	if _, err := cd.deps.FS.Stat(filepath.Join(dir, goModFile)); err == nil {
		types = append(types, "go")
	}

	// Rust project
	if _, err := cd.deps.FS.Stat(filepath.Join(dir, "Cargo.toml")); err == nil {
		types = append(types, "rust")
	}

	// Python project
	pythonMarkers := []string{"pyproject.toml", "setup.py", "requirements.txt"}
	for _, marker := range pythonMarkers {
		if _, err := cd.deps.FS.Stat(filepath.Join(dir, marker)); err == nil {
			types = append(types, "python")
			break
		}
	}

	// JavaScript/TypeScript project
	if _, err := cd.deps.FS.Stat(filepath.Join(dir, "package.json")); err == nil {
		types = append(types, "javascript")
	}

	return types
}

// String returns a string representation of the discovered command.
func (dc *DiscoveredCommand) String() string {
	if dc == nil {
		return ""
	}
	args := strings.Join(dc.Args, " ")
	if args != "" {
		return fmt.Sprintf("%s %s", dc.Command, args)
	}
	return dc.Command
}
