package hooks

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/riddopic/cc-tools/internal/shared"
	"github.com/riddopic/cc-tools/internal/skipregistry"
)

// ValidateWithSkipCheck parses stdinData into a hookcmd.HookInput, checks the
// skip registry, and runs validation. This is the main entry point for both
// cc-tools validate and cc-tools-validate binaries.
func ValidateWithSkipCheck(
	ctx context.Context,
	stdinData []byte,
	stdout io.Writer,
	stderr io.Writer,
	debug bool,
	timeoutSecs int,
	cooldownSecs int,
) int {
	// Parse stdin into HookInput
	input, err := hookcmd.ParseInput(bytes.NewReader(stdinData))
	if err != nil {
		handleInputError(err, debug, stderr)
		return 0
	}

	// Check if directory should be skipped
	skipLint, skipTest := checkSkipsFromInput(ctx, input, debug, stderr)

	// If both are skipped, exit silently
	if skipLint && skipTest {
		if debug {
			_, _ = fmt.Fprintf(stderr, "Both lint and test skipped, exiting silently\n")
		}
		return 0
	}

	// Pass skip information to the validate hook
	skipConfig := &SkipConfig{
		SkipLint: skipLint,
		SkipTest: skipTest,
	}

	// Create dependencies
	defaults := NewDefaultDependencies()
	deps := &Dependencies{
		Stdout:  stdout,
		Stderr:  stderr,
		FS:      defaults.FS,
		Runner:  defaults.Runner,
		Process: defaults.Process,
		Clock:   defaults.Clock,
	}

	return RunValidateHookWithSkip(ctx, input, debug, timeoutSecs, cooldownSecs, skipConfig, deps)
}

// checkSkipsFromInput checks the skip registry using the parsed HookInput.
func checkSkipsFromInput(ctx context.Context, input *hookcmd.HookInput, debug bool, stderr io.Writer) (bool, bool) {
	if input == nil {
		return false, false
	}

	// Get file path from input using the canonical method
	filePath := input.GetFilePath()

	if filePath == "" {
		// No file path, don't skip
		if debug {
			_, _ = fmt.Fprintf(stderr, "No file path found in input\n")
		}
		return false, false
	}

	// Get directory from file path
	fileDir := filepath.Dir(filePath)

	// Find the project root - same as we do for discovering lint/test commands
	projectRoot, err := shared.FindProjectRoot(fileDir, nil)
	if err != nil {
		if debug {
			_, _ = fmt.Fprintf(stderr, "Failed to find project root: %v\n", err)
		}
		// If we can't find project root, check the file's directory as fallback
		projectRoot = fileDir
	}

	// Convert to absolute path
	absProjectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		if debug {
			_, _ = fmt.Fprintf(stderr, "Failed to get absolute path: %v\n", err)
		}
		return false, false
	}

	// Check skip registry for the project root
	storage := skipregistry.DefaultStorage()
	registry := skipregistry.NewRegistry(storage)

	skipLint, _ := registry.IsSkipped(ctx, skipregistry.DirectoryPath(absProjectRoot), skipregistry.SkipTypeLint)
	skipTest, _ := registry.IsSkipped(ctx, skipregistry.DirectoryPath(absProjectRoot), skipregistry.SkipTypeTest)

	if debug {
		_, _ = fmt.Fprintf(stderr, "File: %s\n", filePath)
		_, _ = fmt.Fprintf(stderr, "Project root: %s\n", absProjectRoot)
		_, _ = fmt.Fprintf(stderr, "Checking skips for project root: %s\n", absProjectRoot)
		if skipLint {
			_, _ = fmt.Fprintf(stderr, "Skipping lint for project: %s\n", absProjectRoot)
		}
		if skipTest {
			_, _ = fmt.Fprintf(stderr, "Skipping test for project: %s\n", absProjectRoot)
		}
	}

	return skipLint, skipTest
}
