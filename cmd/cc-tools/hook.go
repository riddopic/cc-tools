package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/handler"
	"github.com/riddopic/cc-tools/internal/hookcmd"
)

func newHookCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "hook",
		Short:  "Handle Claude Code hook events",
		Long:   "Reads hook event JSON from stdin, dispatches to registered handlers, and writes structured output.",
		Hidden: true,
		RunE:   runHook,
	}
}

func runHook(cmd *cobra.Command, _ []string) error {
	data, readErr := io.ReadAll(os.Stdin)
	if readErr != nil {
		return nil //nolint:nilerr // hooks must not block on stdin errors
	}
	if len(data) == 0 {
		return nil
	}

	input, parseErr := hookcmd.ParseInput(bytes.NewReader(data))
	if parseErr != nil {
		return nil //nolint:nilerr // hooks must not block on parse errors
	}

	cfg := loadConfig()
	registry := handler.NewDefaultRegistry(cfg)
	resp := registry.Dispatch(cmd.Context(), input)

	return writeHookResponse(os.Stdout, os.Stderr, resp)
}

func loadConfig() *config.Values {
	mgr := config.NewManager()
	cfg, err := mgr.GetConfig(context.TODO())
	if err != nil {
		return nil
	}
	return cfg
}

func writeHookResponse(stdout, stderr io.Writer, resp *handler.Response) error {
	if resp.Stderr != "" {
		_, _ = stderr.Write([]byte(resp.Stderr))
	}

	if resp.Stdout != nil {
		data, err := json.Marshal(resp.Stdout)
		if err != nil {
			return fmt.Errorf("marshal hook output: %w", err)
		}
		_, _ = stdout.Write(data)
		_, _ = io.WriteString(stdout, "\n")
	}

	if resp.ExitCode != 0 {
		return &exitError{code: resp.ExitCode}
	}

	return nil
}
