package hooks

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

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

	fileDir, err := filepath.Abs(filepath.Dir(filePath))
	if err != nil {
		if debug {
			_, _ = fmt.Fprintf(stderr, "Failed to get absolute path: %v\n", err)
		}
		return false, false
	}

	// Resolve the same root validate runs commands from
	projectRoot, err := shared.FindRepoRoot(fileDir, nil)
	if err != nil {
		if debug {
			_, _ = fmt.Fprintf(stderr, "Failed to find project root: %v\n", err)
		}
		// If we can't find project root, check the file's directory as fallback
		projectRoot = fileDir
	}

	registry := skipregistry.NewRegistry(skipregistry.DefaultStorage())
	skipLint, skipTest := skippedTypes(ctx, registry, fileDir, projectRoot)

	if debug {
		_, _ = fmt.Fprintf(stderr, "File: %s\n", filePath)
		_, _ = fmt.Fprintf(stderr, "Checking skips for project root: %s (walking up from %s)\n", projectRoot, fileDir)
		if skipLint {
			_, _ = fmt.Fprintf(stderr, "Skipping lint for: %s\n", filePath)
		}
		if skipTest {
			_, _ = fmt.Fprintf(stderr, "Skipping test for: %s\n", filePath)
		}
	}

	return skipLint, skipTest
}

// skippedTypes reports which validations are skipped for files in fileDir. A
// skip registered on fileDir or on any ancestor up to and including root
// applies, so `cc-tools skip` at a repository root also covers packages with
// their own manifests nested below it. Registry read errors skip nothing.
func skippedTypes(ctx context.Context, reader skipregistry.Reader, fileDir, root string) (bool, bool) {
	if !isWithinDir(fileDir, root) {
		root = fileDir
	}

	var skipLint, skipTest bool
	for dir := fileDir; ; dir = filepath.Dir(dir) {
		lint, _ := reader.IsSkipped(ctx, skipregistry.DirectoryPath(dir), skipregistry.SkipTypeLint)
		test, _ := reader.IsSkipped(ctx, skipregistry.DirectoryPath(dir), skipregistry.SkipTypeTest)
		skipLint = skipLint || lint
		skipTest = skipTest || test

		if dir == root || dir == filepath.Dir(dir) {
			return skipLint, skipTest
		}
	}
}

// isWithinDir reports whether path is dir itself or lies below it.
func isWithinDir(path, dir string) bool {
	rel, err := filepath.Rel(dir, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
