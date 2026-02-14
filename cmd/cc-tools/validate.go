package main

import (
	"context"
	"io"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/hooks"
)

func newValidateCmd() *cobra.Command {
	var timeout int
	var cooldown int

	defaults := config.GetDefaultConfig()

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Run lint and test validation in parallel",
		Long:  "Discovers and runs lint and test commands in parallel, reporting results. Used as a PostToolUse hook for Claude Code.",
		Example: `  echo '{"tool_input":{"file_path":"main.go"}}' | cc-tools validate
  cc-tools validate --timeout 120`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			timeout, cooldown = resolveValidateConfig(
				defaults, timeout, cooldown,
			)
			return runValidate(cmd, timeout, cooldown)
		},
	}

	cmd.Flags().IntVarP(&timeout, "timeout", "t", defaults.Validate.Timeout, "timeout in seconds")
	cmd.Flags().IntVarP(&cooldown, "cooldown", "c", defaults.Validate.Cooldown, "cooldown between runs in seconds")

	return cmd
}

// resolveValidateConfig applies config file and env var overrides to the
// flag defaults. Precedence: env vars > config file > flag defaults.
func resolveValidateConfig(defaults *config.Values, timeout, cooldown int) (int, int) {
	// Config file overrides flag defaults.
	mgr := config.NewManager()
	cfg, err := mgr.GetConfig(context.Background())
	if err == nil && cfg != nil {
		if timeout == defaults.Validate.Timeout && cfg.Validate.Timeout > 0 {
			timeout = cfg.Validate.Timeout
		}
		if cooldown == defaults.Validate.Cooldown && cfg.Validate.Cooldown > 0 {
			cooldown = cfg.Validate.Cooldown
		}
	}

	// Environment variables take highest precedence.
	if envTimeout := os.Getenv("CC_TOOLS_HOOKS_VALIDATE_TIMEOUT_SECONDS"); envTimeout != "" {
		if val, parseErr := strconv.Atoi(envTimeout); parseErr == nil && val > 0 {
			timeout = val
		}
	}
	if envCooldown := os.Getenv("CC_TOOLS_HOOKS_VALIDATE_COOLDOWN_SECONDS"); envCooldown != "" {
		if val, parseErr := strconv.Atoi(envCooldown); parseErr == nil && val >= 0 {
			cooldown = val
		}
	}

	return timeout, cooldown
}

func runValidate(cmd *cobra.Command, timeout, cooldown int) error {
	debug := os.Getenv("CLAUDE_HOOKS_DEBUG") == "1"

	var stdinData []byte
	if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
		stdinData, _ = io.ReadAll(os.Stdin)
	}

	exitCode := hooks.ValidateWithSkipCheck(
		cmd.Context(),
		stdinData,
		os.Stdout,
		os.Stderr,
		debug,
		timeout,
		cooldown,
	)

	if exitCode != 0 {
		return &exitError{code: exitCode}
	}
	return nil
}
