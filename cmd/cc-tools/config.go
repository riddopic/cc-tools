package main

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/output"
)

const configSetArgs = 2

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration settings",
	}
	cmd.AddCommand(
		newConfigGetCmd(),
		newConfigSetCmd(),
		newConfigListCmd(),
		newConfigResetCmd(),
	)
	return cmd
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "get <key>",
		Short:   "Get a configuration value",
		Args:    cobra.ExactArgs(1),
		Example: "  cc-tools config get validate.timeout",
		RunE: func(_ *cobra.Command, args []string) error {
			return handleConfigGet(context.Background(), newTerminal(), newConfigManager(), args[0])
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "set <key> <value>",
		Short:   "Set a configuration value",
		Args:    cobra.ExactArgs(configSetArgs),
		Example: "  cc-tools config set validate.timeout 90",
		RunE: func(_ *cobra.Command, args []string) error {
			return handleConfigSet(context.Background(), newTerminal(), newConfigManager(), args[0], args[1])
		},
	}
}

func newConfigListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "Show all configuration with defaults and overrides",
		Aliases: []string{"show"},
		RunE: func(_ *cobra.Command, _ []string) error {
			return handleConfigList(context.Background(), newTerminal(), newConfigManager())
		},
	}
}

func newConfigResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "reset [key]",
		Short:   "Reset configuration to defaults (all or specific key)",
		Args:    cobra.MaximumNArgs(1),
		Example: "  cc-tools config reset validate.timeout\n  cc-tools config reset",
		RunE: func(_ *cobra.Command, args []string) error {
			var key string
			if len(args) > 0 {
				key = args[0]
			}
			return handleConfigReset(context.Background(), newTerminal(), newConfigManager(), key)
		},
	}
}

func handleConfigGet(ctx context.Context, out *output.Terminal, manager *config.Manager, key string) error {
	if err := manager.EnsureConfig(ctx); err != nil {
		return fmt.Errorf("ensure config: %w", err)
	}

	value, exists, err := manager.GetValue(ctx, key)
	if err != nil {
		return fmt.Errorf("get config value: %w", err)
	}

	if !exists {
		_ = out.Error("Key '%s' not found", key)
		_ = out.Info("Available keys:")
		keys, _ := manager.GetAllKeys(ctx)
		for _, k := range keys {
			_ = out.Info("  %s", k)
		}
		return errors.New("key not found")
	}

	_ = out.Raw(fmt.Sprintf("%v\n", value))
	return nil
}

func handleConfigSet(ctx context.Context, out *output.Terminal, manager *config.Manager, key, value string) error {
	if err := manager.EnsureConfig(ctx); err != nil {
		return fmt.Errorf("ensure config: %w", err)
	}

	if err := manager.Set(ctx, key, value); err != nil {
		return fmt.Errorf("set config value: %w", err)
	}

	_ = out.Success("✓ Set %s = %s", key, value)
	return nil
}

func handleConfigList(ctx context.Context, out *output.Terminal, manager *config.Manager) error {
	if err := manager.EnsureConfig(ctx); err != nil {
		return fmt.Errorf("ensure config: %w", err)
	}

	settings, err := manager.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("get all config: %w", err)
	}

	keys := make([]string, 0, len(settings))
	for key := range settings {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	table := output.NewTable(
		[]string{"Setting", "Value", "Status"},
		[]int{30, 25, 10},
	)

	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	customStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))

	for _, key := range keys {
		info := settings[key]
		var status string
		if info.IsDefault {
			status = defaultStyle.Render("default")
		} else {
			status = customStyle.Render("custom")
		}

		value := info.Value
		if value == "" {
			value = "(empty)"
		}

		table.AddRow([]string{key, value, status})
	}

	_ = out.Info("Configuration Settings")
	_ = out.Write(table.Render())

	configPath := manager.GetConfigPath()
	_ = out.Info("\nConfig file: %s", configPath)

	return nil
}

func handleConfigReset(ctx context.Context, out *output.Terminal, manager *config.Manager, key string) error {
	if key == "" {
		if err := manager.ResetAll(ctx); err != nil {
			return fmt.Errorf("reset all config: %w", err)
		}
		_ = out.Success("✓ Reset all configuration to defaults")
	} else {
		if err := manager.Reset(ctx, key); err != nil {
			return fmt.Errorf("reset config key: %w", err)
		}
		_ = out.Success("✓ Reset %s to default value", key)
	}

	return nil
}
