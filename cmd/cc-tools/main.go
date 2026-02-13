// Package main implements the cc-tools CLI application.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/hooks"
	"github.com/riddopic/cc-tools/internal/output"
	"github.com/riddopic/cc-tools/internal/shared"
)

const (
	minArgs     = 2
	helpFlag    = "--help"
	helpCommand = "help"
)

// Build-time variables.
var version = "dev"

func main() {
	out := output.NewTerminal(os.Stdout, os.Stderr)

	// Debug logging - log all invocations to a file
	debugLog()

	if len(os.Args) < minArgs {
		printUsage(out)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "validate":
		runValidate()
	case "skip":
		runSkipCommand()
	case "unskip":
		runUnskipCommand()
	case "debug":
		runDebugCommand()
	case "mcp":
		runMCPCommand()
	case "config":
		runConfigCommand()
	case "version":
		// Print version to stdout as intended output
		out.Raw(fmt.Sprintf("cc-tools %s\n", version))
	case helpCommand, "-h", helpFlag:
		printUsage(out)
	default:
		out.Error("Unknown command: %s", os.Args[1])
		printUsage(out)
		os.Exit(1)
	}
}

func printUsage(out *output.Terminal) {
	out.RawError(`cc-tools - Claude Code Tools

Usage:
  cc-tools <command> [arguments]

Commands:
  validate      Run smart validation (lint and test in parallel)
  skip          Configure skip settings for directories
  unskip        Remove skip settings from directories
  debug         Configure debug logging for directories
  mcp           Manage Claude MCP servers
  config        Manage configuration settings
  version       Print version information
  help          Show this help message

Examples:
  echo '{"file_path": "main.go"}' | cc-tools validate
  cc-tools mcp list
  cc-tools mcp enable jira
`)
}

func loadValidateConfig() (int, int) {
	timeoutSecs := 60
	cooldownSecs := 5

	// Load configuration
	cfg, _ := config.Load()
	if cfg != nil {
		if cfg.Hooks.Validate.TimeoutSeconds > 0 {
			timeoutSecs = cfg.Hooks.Validate.TimeoutSeconds
		}
		if cfg.Hooks.Validate.CooldownSeconds > 0 {
			cooldownSecs = cfg.Hooks.Validate.CooldownSeconds
		}
	}

	// Environment variables override config
	if envTimeout := os.Getenv("CC_TOOLS_HOOKS_VALIDATE_TIMEOUT_SECONDS"); envTimeout != "" {
		if val, err := strconv.Atoi(envTimeout); err == nil && val > 0 {
			timeoutSecs = val
		}
	}
	if envCooldown := os.Getenv("CC_TOOLS_HOOKS_VALIDATE_COOLDOWN_SECONDS"); envCooldown != "" {
		if val, err := strconv.Atoi(envCooldown); err == nil && val >= 0 {
			cooldownSecs = val
		}
	}

	return timeoutSecs, cooldownSecs
}

func runValidate() {
	timeoutSecs, cooldownSecs := loadValidateConfig()
	debug := os.Getenv("CLAUDE_HOOKS_DEBUG") == "1"

	exitCode := hooks.ValidateWithSkipCheck(
		context.Background(),
		os.Stdin,
		os.Stdout,
		os.Stderr,
		debug,
		timeoutSecs,
		cooldownSecs,
	)
	os.Exit(exitCode)
}

func debugLog() {
	// Create or append to debug log file for current directory
	debugFile := getDebugLogPath()
	//nolint:gosec // Debug log file path is controlled
	f, err := os.OpenFile(debugFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return // Silently fail if we can't write debug log
	}
	defer func() { _ = f.Close() }()

	// Read stdin and save it for both debug and actual use
	// Only read stdin for commands that actually need it
	var stdinDebugData []byte
	needsStdin := len(os.Args) > 1 && os.Args[1] == "validate"

	if needsStdin {
		if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
			// There's data in stdin
			stdinDebugData, _ = io.ReadAll(os.Stdin)
			// Create a new reader from the data we just read
			// This will be used by the actual commands
			// Actually, we need to pipe it back - create a temp file
			if tmpFile, tmpErr := os.CreateTemp("", "cc-tools-stdin-"); tmpErr == nil { //nolint:forbidigo // Debug temp file
				_, _ = tmpFile.Write(stdinDebugData)
				_, _ = tmpFile.Seek(0, 0)
				os.Stdin = tmpFile //nolint:reassign // Resetting stdin for subsequent reads
			}
		}
	}

	// Log the invocation details
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	_, _ = fmt.Fprintf(f, "\n========================================\n")
	_, _ = fmt.Fprintf(f, "[%s] cc-tools invoked\n", timestamp)
	_, _ = fmt.Fprintf(f, "Args: %v\n", os.Args)
	_, _ = fmt.Fprintf(f, "Environment:\n")
	_, _ = fmt.Fprintf(f, "  CLAUDE_HOOKS_DEBUG: %s\n", os.Getenv("CLAUDE_HOOKS_DEBUG"))
	_, _ = fmt.Fprintf(f, "  Working Dir: %s\n", func() string {
		if wd, wdErr := os.Getwd(); wdErr == nil {
			return wd
		}
		return "unknown"
	}())

	if len(stdinDebugData) > 0 {
		_, _ = fmt.Fprintf(f, "Stdin: %s\n", string(stdinDebugData))
	} else {
		_, _ = fmt.Fprintf(f, "Stdin: (no data available)\n")
	}

	_, _ = fmt.Fprintf(f, "Command: %s\n", func() string {
		if len(os.Args) > 1 {
			return os.Args[1]
		}
		return "(none)"
	}())
}

// getDebugLogPath returns the debug log path for the current directory.
func getDebugLogPath() string {
	wd, err := os.Getwd()
	if err != nil {
		// Fallback to generic log if we can't get working directory
		return "/tmp/cc-tools.debug"
	}
	return shared.GetDebugLogPathForDir(wd)
}
