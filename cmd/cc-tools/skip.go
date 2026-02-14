package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/riddopic/cc-tools/internal/output"
	"github.com/riddopic/cc-tools/internal/skipregistry"
)

func newSkipCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skip",
		Short: "Configure skip settings for directories",
	}
	cmd.AddCommand(
		newSkipLintCmd(),
		newSkipTestCmd(),
		newSkipAllCmd(),
		newSkipListCmd(),
		newSkipStatusCmd(),
	)
	return cmd
}

func newUnskipCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unskip",
		Short: "Remove skip settings from directories",
	}
	cmd.AddCommand(
		newUnskipLintCmd(),
		newUnskipTestCmd(),
		newUnskipAllCmd(),
	)
	// Default behavior when called without subcommand: clear all skips.
	cmd.RunE = func(_ *cobra.Command, _ []string) error {
		out := output.NewTerminal(os.Stdout, os.Stderr)
		storage := skipregistry.DefaultStorage()
		registry := skipregistry.NewRegistry(storage)
		return clearSkips(context.Background(), out, registry)
	}
	return cmd
}

func newSkipLintCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "lint",
		Short:   "Skip linting in the current directory",
		Example: "  cc-tools skip lint",
		RunE: func(_ *cobra.Command, _ []string) error {
			out := output.NewTerminal(os.Stdout, os.Stderr)
			storage := skipregistry.DefaultStorage()
			registry := skipregistry.NewRegistry(storage)
			return addSkip(context.Background(), out, registry, skipregistry.SkipTypeLint)
		},
	}
}

func newSkipTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "test",
		Short:   "Skip testing in the current directory",
		Example: "  cc-tools skip test",
		RunE: func(_ *cobra.Command, _ []string) error {
			out := output.NewTerminal(os.Stdout, os.Stderr)
			storage := skipregistry.DefaultStorage()
			registry := skipregistry.NewRegistry(storage)
			return addSkip(context.Background(), out, registry, skipregistry.SkipTypeTest)
		},
	}
}

func newSkipAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "all",
		Short:   "Skip both linting and testing in the current directory",
		Example: "  cc-tools skip all",
		RunE: func(_ *cobra.Command, _ []string) error {
			out := output.NewTerminal(os.Stdout, os.Stderr)
			storage := skipregistry.DefaultStorage()
			registry := skipregistry.NewRegistry(storage)
			return addSkip(context.Background(), out, registry, skipregistry.SkipTypeAll)
		},
	}
}

func newSkipListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Show all directories with skip configurations",
		RunE: func(_ *cobra.Command, _ []string) error {
			out := output.NewTerminal(os.Stdout, os.Stderr)
			storage := skipregistry.DefaultStorage()
			registry := skipregistry.NewRegistry(storage)
			return listSkips(context.Background(), out, registry)
		},
	}
}

func newSkipStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Short:   "Show skip status for the current directory",
		Example: "  cc-tools skip status",
		RunE: func(_ *cobra.Command, _ []string) error {
			out := output.NewTerminal(os.Stdout, os.Stderr)
			storage := skipregistry.DefaultStorage()
			registry := skipregistry.NewRegistry(storage)
			return showStatus(context.Background(), out, registry)
		},
	}
}

func newUnskipLintCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "lint",
		Short:   "Remove skip for linting in the current directory",
		Example: "  cc-tools unskip lint",
		RunE: func(_ *cobra.Command, _ []string) error {
			out := output.NewTerminal(os.Stdout, os.Stderr)
			storage := skipregistry.DefaultStorage()
			registry := skipregistry.NewRegistry(storage)
			return removeSkip(context.Background(), out, registry, skipregistry.SkipTypeLint)
		},
	}
}

func newUnskipTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "test",
		Short:   "Remove skip for testing in the current directory",
		Example: "  cc-tools unskip test",
		RunE: func(_ *cobra.Command, _ []string) error {
			out := output.NewTerminal(os.Stdout, os.Stderr)
			storage := skipregistry.DefaultStorage()
			registry := skipregistry.NewRegistry(storage)
			return removeSkip(context.Background(), out, registry, skipregistry.SkipTypeTest)
		},
	}
}

func newUnskipAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "all",
		Short:   "Remove all skips for the current directory",
		Example: "  cc-tools unskip all",
		RunE: func(_ *cobra.Command, _ []string) error {
			out := output.NewTerminal(os.Stdout, os.Stderr)
			storage := skipregistry.DefaultStorage()
			registry := skipregistry.NewRegistry(storage)
			return clearSkips(context.Background(), out, registry)
		},
	}
}

func validateSkipPath(dir string) (string, error) {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}

	cleanPath := filepath.Clean(absPath)
	if strings.Contains(cleanPath, "..") {
		return "", errors.New("invalid path: directory traversal not allowed")
	}

	return cleanPath, nil
}

func addSkip(
	ctx context.Context,
	out *output.Terminal,
	registry skipregistry.Registry,
	skipType skipregistry.SkipType,
) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	dir, err = validateSkipPath(dir)
	if err != nil {
		return err
	}

	if addErr := registry.AddSkip(ctx, skipregistry.DirectoryPath(dir), skipType); addErr != nil {
		return fmt.Errorf("add skip: %w", addErr)
	}

	switch skipType {
	case skipregistry.SkipTypeLint:
		_ = out.Success("✓ Linting will be skipped in %s", dir)
	case skipregistry.SkipTypeTest:
		_ = out.Success("✓ Testing will be skipped in %s", dir)
	case skipregistry.SkipTypeAll:
		_ = out.Success("✓ Linting and testing will be skipped in %s", dir)
	}

	return nil
}

func removeSkip(
	ctx context.Context,
	out *output.Terminal,
	registry skipregistry.Registry,
	skipType skipregistry.SkipType,
) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	if removeErr := registry.RemoveSkip(ctx, skipregistry.DirectoryPath(dir), skipType); removeErr != nil {
		return fmt.Errorf("remove skip: %w", removeErr)
	}

	switch skipType {
	case skipregistry.SkipTypeLint:
		_ = out.Success("✓ Linting will no longer be skipped in %s", dir)
	case skipregistry.SkipTypeTest:
		_ = out.Success("✓ Testing will no longer be skipped in %s", dir)
	case skipregistry.SkipTypeAll:
		// This case won't occur as we expand SkipTypeAll earlier
	}

	return nil
}

func clearSkips(
	ctx context.Context,
	out *output.Terminal,
	registry skipregistry.Registry,
) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	if clearErr := registry.Clear(ctx, skipregistry.DirectoryPath(dir)); clearErr != nil {
		return fmt.Errorf("clear skips: %w", clearErr)
	}

	_ = out.Success("✓ All skips removed from %s", dir)
	return nil
}

func listSkips(
	ctx context.Context,
	out *output.Terminal,
	registry skipregistry.Registry,
) error {
	entries, err := registry.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("list all: %w", err)
	}

	if len(entries) == 0 {
		_ = out.Info("No directories have skip configurations")
		return nil
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path.String() < entries[j].Path.String()
	})

	table := output.NewTable(
		[]string{"Directory", "Skip Types"},
		[]int{50, 30},
	)

	for _, entry := range entries {
		var typeStrs []string
		for _, t := range entry.Types {
			typeStrs = append(typeStrs, string(t))
		}
		table.AddRow([]string{
			entry.Path.String(),
			strings.Join(typeStrs, ", "),
		})
	}

	_ = out.Info("Skip configurations:")
	_ = out.Write(table.Render())

	return nil
}

func showStatus(
	ctx context.Context,
	out *output.Terminal,
	registry skipregistry.Registry,
) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	types, err := registry.GetSkipTypes(ctx, skipregistry.DirectoryPath(dir))
	if err != nil {
		return fmt.Errorf("get skip types: %w", err)
	}

	if len(types) == 0 {
		_ = out.Info("No skips configured for %s", dir)
		return nil
	}

	table := output.NewTable(
		[]string{"Type", "Status"},
		[]int{20, 30},
	)

	hasLint := false
	hasTest := false
	for _, t := range types {
		switch t {
		case skipregistry.SkipTypeLint:
			hasLint = true
		case skipregistry.SkipTypeTest:
			hasTest = true
		case skipregistry.SkipTypeAll:
			hasLint = true
			hasTest = true
		}
	}

	if hasLint {
		table.AddRow([]string{"Linting", "SKIPPED"})
	} else {
		table.AddRow([]string{"Linting", "Active"})
	}

	if hasTest {
		table.AddRow([]string{"Testing", "SKIPPED"})
	} else {
		table.AddRow([]string{"Testing", "Active"})
	}

	_ = out.Info("Skip status for %s:", dir)
	_ = out.Write(table.Render())

	return nil
}
