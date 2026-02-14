package main

import (
	"context"
	"time"

	"github.com/spf13/cobra"
)

const mcpTimeout = 30 * time.Second

func newMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Manage Claude MCP servers",
	}
	cmd.AddCommand(
		newMCPListCmd(),
		newMCPEnableCmd(),
		newMCPDisableCmd(),
		newMCPEnableAllCmd(),
		newMCPDisableAllCmd(),
	)
	return cmd
}

func newMCPListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "Show all MCP servers and their status",
		Example: "  cc-tools mcp list",
		RunE: func(_ *cobra.Command, _ []string) error {
			out := newTerminal()
			ctx, cancel := context.WithTimeout(context.Background(), mcpTimeout)
			defer cancel()
			return newMCPManager(out).List(ctx)
		},
	}
}

func newMCPEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "enable <name>",
		Short:   "Enable an MCP server",
		Args:    cobra.ExactArgs(1),
		Example: "  cc-tools mcp enable jira",
		RunE: func(_ *cobra.Command, args []string) error {
			out := newTerminal()
			ctx, cancel := context.WithTimeout(context.Background(), mcpTimeout)
			defer cancel()
			return newMCPManager(out).Enable(ctx, args[0])
		},
	}
}

func newMCPDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "disable <name>",
		Short:   "Disable an MCP server",
		Args:    cobra.ExactArgs(1),
		Example: "  cc-tools mcp disable playwright",
		RunE: func(_ *cobra.Command, args []string) error {
			out := newTerminal()
			ctx, cancel := context.WithTimeout(context.Background(), mcpTimeout)
			defer cancel()
			return newMCPManager(out).Disable(ctx, args[0])
		},
	}
}

func newMCPEnableAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "enable-all",
		Short:   "Enable all MCP servers from settings",
		Example: "  cc-tools mcp enable-all",
		RunE: func(_ *cobra.Command, _ []string) error {
			out := newTerminal()
			ctx, cancel := context.WithTimeout(context.Background(), mcpTimeout)
			defer cancel()
			return newMCPManager(out).EnableAll(ctx)
		},
	}
}

func newMCPDisableAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "disable-all",
		Short:   "Disable all MCP servers",
		Example: "  cc-tools mcp disable-all",
		RunE: func(_ *cobra.Command, _ []string) error {
			out := newTerminal()
			ctx, cancel := context.WithTimeout(context.Background(), mcpTimeout)
			defer cancel()
			return newMCPManager(out).DisableAll(ctx)
		},
	}
}
