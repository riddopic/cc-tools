// Package config manages application configuration.
package config

import (
	"context"
	"fmt"
)

// Config represents the application configuration.
type Config struct {
	Hooks         HooksConfig         `json:"hooks"`
	Notifications NotificationsConfig `json:"notifications"`
}

// HooksConfig represents hook-related settings.
type HooksConfig struct {
	Validate ValidateConfig `json:"validate"`
}

// ValidateConfig represents validate hook settings.
type ValidateConfig struct {
	CooldownSeconds int `json:"cooldown_seconds"`
	TimeoutSeconds  int `json:"timeout_seconds"`
}

// NotificationsConfig represents notification settings.
type NotificationsConfig struct {
	NtfyTopic string `json:"ntfy_topic"`
}

// Load loads configuration from the config file.
// It delegates to Manager for consistent config access.
func Load() (*Config, error) {
	manager := NewManager()

	values, err := manager.GetConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return &Config{
		Hooks: HooksConfig{
			Validate: ValidateConfig{
				CooldownSeconds: values.Validate.Cooldown,
				TimeoutSeconds:  values.Validate.Timeout,
			},
		},
		Notifications: NotificationsConfig{
			NtfyTopic: values.Notifications.NtfyTopic,
		},
	}, nil
}
