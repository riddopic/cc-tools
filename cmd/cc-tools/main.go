// Package main implements the cc-tools CLI application.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/riddopic/cc-tools/internal/shared"
)

// Build-time variables.
var version = "dev"

func main() {
	root := newRootCmd()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:     "cc-tools",
		Short:   "Claude Code Tools",
		Version: version,
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			writeDebugLog(os.Args, nil)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		newHookCmd(),
		newSessionCmd(),
		newConfigCmd(),
		newSkipCmd(),
		newUnskipCmd(),
		newDebugCmd(),
		newMCPCmd(),
		newValidateCmd(),
	)

	return root
}

func writeDebugLog(args []string, stdinData []byte) {
	debugFile := getDebugLogPath()

	f, err := os.OpenFile(debugFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	_, _ = fmt.Fprintf(f, "\n========================================\n")
	_, _ = fmt.Fprintf(f, "[%s] cc-tools invoked\n", timestamp)
	_, _ = fmt.Fprintf(f, "Args: %v\n", args)
	_, _ = fmt.Fprintf(f, "  CLAUDE_HOOKS_DEBUG: %s\n", os.Getenv("CLAUDE_HOOKS_DEBUG"))

	if wd, wdErr := os.Getwd(); wdErr == nil {
		_, _ = fmt.Fprintf(f, "  Working Dir: %s\n", wd)
	}

	if len(stdinData) > 0 {
		_, _ = fmt.Fprintf(f, "Stdin: %s\n", string(stdinData))
	} else {
		_, _ = fmt.Fprintf(f, "Stdin: (no data)\n")
	}
}

// getDebugLogPath returns the debug log path for the current directory.
func getDebugLogPath() string {
	wd, err := os.Getwd()
	if err != nil {
		return "/tmp/cc-tools.debug"
	}
	return shared.GetDebugLogPathForDir(wd)
}
