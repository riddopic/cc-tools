package main

import (
	"os"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/debug"
	"github.com/riddopic/cc-tools/internal/mcp"
	"github.com/riddopic/cc-tools/internal/output"
	"github.com/riddopic/cc-tools/internal/skipregistry"
)

// Subcommand names and table labels shared across commands.
const (
	subcmdList   = "list"
	subcmdStatus = "status"
	statusLabel  = "Status"
)

func newTerminal() *output.Terminal {
	return output.NewTerminal(os.Stdout, os.Stderr)
}

func newSkipRegistry() *skipregistry.JSONRegistry {
	return skipregistry.NewRegistry(skipregistry.DefaultStorage())
}

func newConfigManager() *config.Manager {
	return config.NewManager()
}

func newDebugManager() *debug.Manager {
	return debug.NewManager()
}

func newMCPManager(out *output.Terminal) *mcp.Manager {
	return mcp.NewManager(out)
}
