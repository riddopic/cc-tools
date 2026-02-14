package main

import (
	"io"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/hooks"
)

const (
	defaultValidateTimeout  = 60
	defaultValidateCooldown = 5
)

func newValidateCmd() *cobra.Command {
	var timeout int
	var cooldown int

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Run lint and test validation in parallel",
		Long:  "Discovers and runs lint and test commands in parallel, reporting results. Used as a PostToolUse hook for Claude Code.",
		Example: `  echo '{"tool_input":{"file_path":"main.go"}}' | cc-tools validate
  cc-tools validate --timeout 120`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			timeout, cooldown = resolveValidateConfig(timeout, cooldown)
			return runValidate(cmd, timeout, cooldown)
		},
	}

	cmd.Flags().IntVarP(&timeout, "timeout", "t", defaultValidateTimeout, "timeout in seconds")
	cmd.Flags().IntVarP(&cooldown, "cooldown", "c", defaultValidateCooldown, "cooldown between runs in seconds")

	return cmd
}

// resolveValidateConfig applies config file and env var overrides to the
// flag defaults. Precedence: env vars > config file > flag defaults.
func resolveValidateConfig(timeout, cooldown int) (int, int) {
	// Config file overrides flag defaults.
	cfg, _ := config.Load()
	if cfg != nil {
		if timeout == defaultValidateTimeout && cfg.Hooks.Validate.TimeoutSeconds > 0 {
			timeout = cfg.Hooks.Validate.TimeoutSeconds
		}
		if cooldown == defaultValidateCooldown && cfg.Hooks.Validate.CooldownSeconds > 0 {
			cooldown = cfg.Hooks.Validate.CooldownSeconds
		}
	}

	// Environment variables take highest precedence.
	if envTimeout := os.Getenv("CC_TOOLS_HOOKS_VALIDATE_TIMEOUT_SECONDS"); envTimeout != "" {
		if val, err := strconv.Atoi(envTimeout); err == nil && val > 0 {
			timeout = val
		}
	}
	if envCooldown := os.Getenv("CC_TOOLS_HOOKS_VALIDATE_COOLDOWN_SECONDS"); envCooldown != "" {
		if val, err := strconv.Atoi(envCooldown); err == nil && val >= 0 {
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
