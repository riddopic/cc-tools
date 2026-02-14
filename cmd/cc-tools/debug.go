package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/riddopic/cc-tools/internal/debug"
	"github.com/riddopic/cc-tools/internal/output"
	"github.com/riddopic/cc-tools/internal/shared"
)

func newDebugCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Configure debug logging for directories",
	}
	cmd.AddCommand(
		newDebugEnableCmd(),
		newDebugDisableCmd(),
		newDebugStatusCmd(),
		newDebugListCmd(),
		newDebugFilenameCmd(),
	)
	return cmd
}

func newDebugEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "enable",
		Short:   "Enable debug logging for the current directory",
		Example: "  cc-tools debug enable",
		RunE: func(_ *cobra.Command, _ []string) error {
			out := output.NewTerminal(os.Stdout, os.Stderr)
			manager := debug.NewManager()
			return enableDebug(context.Background(), out, manager)
		},
	}
}

func newDebugDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "disable",
		Short:   "Disable debug logging for the current directory",
		Example: "  cc-tools debug disable",
		RunE: func(_ *cobra.Command, _ []string) error {
			out := output.NewTerminal(os.Stdout, os.Stderr)
			manager := debug.NewManager()
			return disableDebug(context.Background(), out, manager)
		},
	}
}

func newDebugStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Short:   "Show debug status for the current directory",
		Example: "  cc-tools debug status",
		RunE: func(_ *cobra.Command, _ []string) error {
			out := output.NewTerminal(os.Stdout, os.Stderr)
			manager := debug.NewManager()
			return showDebugStatus(context.Background(), out, manager)
		},
	}
}

func newDebugListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Show all directories with debug logging enabled",
		RunE: func(_ *cobra.Command, _ []string) error {
			out := output.NewTerminal(os.Stdout, os.Stderr)
			manager := debug.NewManager()
			return listDebugDirs(context.Background(), out, manager)
		},
	}
}

func newDebugFilenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "filename",
		Short:   "Print the debug log filename for the current directory",
		Example: "  cc-tools debug filename",
		RunE: func(_ *cobra.Command, _ []string) error {
			out := output.NewTerminal(os.Stdout, os.Stderr)
			return showDebugFilename(out)
		},
	}
}

func enableDebug(ctx context.Context, out *output.Terminal, manager *debug.Manager) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	logFile, err := manager.Enable(ctx, dir)
	if err != nil {
		return fmt.Errorf("enable debug: %w", err)
	}

	_ = out.Success("✓ Debug logging enabled for %s", dir)
	_ = out.Info("  Log file: %s", logFile)
	_ = out.Write("")
	_ = out.Info("cc-tools-validate will write debug logs to this file.")

	return nil
}

func disableDebug(ctx context.Context, out *output.Terminal, manager *debug.Manager) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	if disableErr := manager.Disable(ctx, dir); disableErr != nil {
		return fmt.Errorf("disable debug: %w", disableErr)
	}

	_ = out.Success("✓ Debug logging disabled for %s", dir)

	return nil
}

func showDebugStatus(ctx context.Context, out *output.Terminal, manager *debug.Manager) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	enabled, err := manager.IsEnabled(ctx, dir)
	if err != nil {
		return fmt.Errorf("check debug status: %w", err)
	}

	table := output.NewTable(
		[]string{"Property", "Value"},
		[]int{15, 60},
	)

	if enabled {
		logFile := debug.GetLogFilePath(dir)
		table.AddRow([]string{"Status", "ENABLED"})
		table.AddRow([]string{"Log file", logFile})
	} else {
		table.AddRow([]string{"Status", "DISABLED"})
	}

	_ = out.Info("Debug status for %s:", dir)
	_ = out.Write(table.Render())

	return nil
}

func listDebugDirs(ctx context.Context, out *output.Terminal, manager *debug.Manager) error {
	dirs, err := manager.GetEnabledDirs(ctx)
	if err != nil {
		return fmt.Errorf("list debug directories: %w", err)
	}

	if len(dirs) == 0 {
		_ = out.Info("No directories have debug logging enabled")
		return nil
	}

	sort.Strings(dirs)

	table := output.NewTable(
		[]string{"Directory", "Log File", "Debug File"},
		[]int{30, 35, 35},
	)

	for _, dir := range dirs {
		logFile := debug.GetLogFilePath(dir)
		debugLogFile := shared.GetDebugLogPathForDir(dir)
		table.AddRow([]string{
			dir,
			logFile,
			debugLogFile,
		})
	}

	_ = out.Info("Directories with debug logging enabled:")
	_ = out.Write(table.Render())

	return nil
}

func showDebugFilename(out *output.Terminal) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}
	_ = out.Raw(shared.GetDebugLogPathForDir(wd))
	_ = out.Raw("\n")
	return nil
}
