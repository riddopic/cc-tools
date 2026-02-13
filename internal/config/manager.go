package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

// Configuration keys.
const (
	keyValidateTimeout        = "validate.timeout"
	keyValidateCooldown       = "validate.cooldown"
	keyNotificationsNtfyTopic = "notifications.ntfy_topic"
)

// ConfigValues represents the concrete configuration structure.
type ConfigValues struct {
	Validate      ValidateConfigValues      `json:"validate"`
	Notifications NotificationsConfigValues `json:"notifications"`
}

// NotificationsConfigValues represents notification-related settings.
type NotificationsConfigValues struct {
	NtfyTopic string `json:"ntfy_topic"`
}

// ValidateConfigValues represents validate-related settings.
type ValidateConfigValues struct {
	Timeout  int `json:"timeout"`
	Cooldown int `json:"cooldown"`
}

// Manager handles configuration read/write operations.
type Manager struct {
	configPath string
	config     *ConfigValues
}

// ConfigInfo contains information about a configuration value.
type ConfigInfo struct {
	Value     string
	IsDefault bool
}

const (
	defaultValidateTimeout  = 60
	defaultValidateCooldown = 5
)

// NewManager creates a new configuration manager.
func NewManager() *Manager {
	return &Manager{
		configPath: getConfigFilePath(),
		config:     nil,
	}
}

// EnsureConfig ensures the configuration file exists with defaults.
func (m *Manager) EnsureConfig(_ context.Context) error {
	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// Create config directory if it doesn't exist
		configDir := filepath.Dir(m.configPath)
		if mkErr := os.MkdirAll(configDir, 0o750); mkErr != nil {
			return fmt.Errorf("create config directory: %w", mkErr)
		}

		// Create default config
		if createErr := m.createDefaultConfig(); createErr != nil {
			return fmt.Errorf("create default config: %w", createErr)
		}
	}

	// Load the config
	if err := m.loadConfig(); err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	return nil
}

// GetInt retrieves an integer configuration value.
func (m *Manager) GetInt(_ context.Context, key string) (int, bool, error) {
	if m.config == nil {
		if err := m.loadConfig(); err != nil {
			return 0, false, fmt.Errorf("load config: %w", err)
		}
	}

	switch key {
	case keyValidateTimeout:
		return m.config.Validate.Timeout, true, nil
	case keyValidateCooldown:
		return m.config.Validate.Cooldown, true, nil
	default:
		return 0, false, nil
	}
}

// GetString retrieves a string configuration value.
func (m *Manager) GetString(_ context.Context, key string) (string, bool, error) {
	if m.config == nil {
		if err := m.loadConfig(); err != nil {
			return "", false, fmt.Errorf("load config: %w", err)
		}
	}

	switch key {
	case keyNotificationsNtfyTopic:
		return m.config.Notifications.NtfyTopic, true, nil
	default:
		return "", false, nil
	}
}

// GetValue retrieves a configuration value as a string.
// This is used for display purposes in the config command.
func (m *Manager) GetValue(_ context.Context, key string) (string, bool, error) {
	if m.config == nil {
		if err := m.loadConfig(); err != nil {
			return "", false, fmt.Errorf("load config: %w", err)
		}
	}

	switch key {
	case keyValidateTimeout:
		return strconv.Itoa(m.config.Validate.Timeout), true, nil
	case keyValidateCooldown:
		return strconv.Itoa(m.config.Validate.Cooldown), true, nil
	case keyNotificationsNtfyTopic:
		return m.config.Notifications.NtfyTopic, true, nil
	default:
		return "", false, nil
	}
}

// Set updates a configuration value.
func (m *Manager) Set(_ context.Context, key string, value string) error {
	if m.config == nil {
		if err := m.loadConfig(); err != nil {
			return fmt.Errorf("load config: %w", err)
		}
	}

	// Parse and set the value
	switch key {
	case keyValidateTimeout:
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("value must be an integer: %w", err)
		}
		m.config.Validate.Timeout = intVal
	case keyValidateCooldown:
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("value must be an integer: %w", err)
		}
		m.config.Validate.Cooldown = intVal
	case keyNotificationsNtfyTopic:
		m.config.Notifications.NtfyTopic = value
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	// Save to file
	if err := m.saveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	return nil
}

// GetAll retrieves all configuration values with their metadata.
func (m *Manager) GetAll(ctx context.Context) (map[string]ConfigInfo, error) {
	if m.config == nil {
		if err := m.loadConfig(); err != nil {
			return nil, fmt.Errorf("load config: %w", err)
		}
	}

	defaults := getDefaultConfig()
	result := make(map[string]ConfigInfo)

	// Process all configuration keys
	keys := []string{
		keyValidateTimeout,
		keyValidateCooldown,
		keyNotificationsNtfyTopic,
	}

	for _, key := range keys {
		value, _, _ := m.GetValue(ctx, key)
		defaultValue := getDefaultValue(defaults, key)

		result[key] = ConfigInfo{
			Value:     value,
			IsDefault: value == defaultValue,
		}
	}

	return result, nil
}

// GetAllKeys returns all available configuration keys.
func (m *Manager) GetAllKeys(_ context.Context) ([]string, error) {
	keys := []string{
		keyValidateTimeout,
		keyValidateCooldown,
		keyNotificationsNtfyTopic,
	}
	sort.Strings(keys)
	return keys, nil
}

// Reset resets a specific configuration key to its default value.
func (m *Manager) Reset(_ context.Context, key string) error {
	if m.config == nil {
		if err := m.loadConfig(); err != nil {
			return fmt.Errorf("load config: %w", err)
		}
	}

	defaults := getDefaultConfig()

	// Reset to default value
	switch key {
	case keyValidateTimeout:
		m.config.Validate.Timeout = defaults.Validate.Timeout
	case keyValidateCooldown:
		m.config.Validate.Cooldown = defaults.Validate.Cooldown
	case keyNotificationsNtfyTopic:
		m.config.Notifications.NtfyTopic = defaults.Notifications.NtfyTopic
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	// Save to file
	if err := m.saveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	return nil
}

// ResetAll resets all configuration to defaults.
func (m *Manager) ResetAll(_ context.Context) error {
	// Create new config with defaults
	m.config = getDefaultConfig()

	// Save to file
	if err := m.saveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	return nil
}

// GetConfig returns the current configuration structure.
// This is used by the Load function to get typed configuration.
func (m *Manager) GetConfig(_ context.Context) (*ConfigValues, error) {
	if m.config == nil {
		if err := m.loadConfig(); err != nil {
			return nil, fmt.Errorf("load config: %w", err)
		}
	}
	return m.config, nil
}

// GetConfigPath returns the path to the configuration file.
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// loadConfig loads the configuration from file.
func (m *Manager) loadConfig() error {
	// Initialize with defaults
	m.config = getDefaultConfig()

	// Read file if it exists
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, use defaults
			return nil
		}
		return fmt.Errorf("read config file: %w", err)
	}

	// Try to parse as structured config first
	var structuredConfig ConfigValues
	if unmarshalErr := json.Unmarshal(data, &structuredConfig); unmarshalErr == nil {
		// Successfully parsed as structured config
		m.config = &structuredConfig
		// Ensure all fields have values (use defaults for missing fields)
		m.ensureDefaults()
		return nil
	}

	// Try parsing as nested map for backward compatibility
	var mapConfig map[string]any
	if unmarshalErr := json.Unmarshal(data, &mapConfig); unmarshalErr != nil {
		return fmt.Errorf("parse config file: %w", unmarshalErr)
	}

	// Convert from map to structured config
	m.convertFromMap(mapConfig)
	m.ensureDefaults()

	return nil
}

// saveConfig saves the current configuration to file.
func (m *Manager) saveConfig() error {
	// Ensure directory exists
	configDir := filepath.Dir(m.configPath)
	if mkErr := os.MkdirAll(configDir, 0o750); mkErr != nil {
		return fmt.Errorf("create config directory: %w", mkErr)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	// Write to file
	if writeErr := os.WriteFile(m.configPath, data, 0o600); writeErr != nil {
		return fmt.Errorf("write config file: %w", writeErr)
	}

	return nil
}

// createDefaultConfig creates a configuration file with default values.
func (m *Manager) createDefaultConfig() error {
	m.config = getDefaultConfig()
	return m.saveConfig()
}

// getDefaultConfig returns a new config with default values.
func getDefaultConfig() *ConfigValues {
	return &ConfigValues{
		Validate: ValidateConfigValues{
			Timeout:  defaultValidateTimeout,
			Cooldown: defaultValidateCooldown,
		},
	}
}

// ensureDefaults ensures all fields have values, using defaults for missing fields.
func (m *Manager) ensureDefaults() {
	defaults := getDefaultConfig()

	if m.config.Validate.Timeout == 0 {
		m.config.Validate.Timeout = defaults.Validate.Timeout
	}
	if m.config.Validate.Cooldown == 0 {
		m.config.Validate.Cooldown = defaults.Validate.Cooldown
	}
}

// convertFromMap converts the old map-based config to the new structured format.
func (m *Manager) convertFromMap(mapConfig map[string]any) {
	// Initialize with defaults
	m.config = getDefaultConfig()

	// Convert validate settings
	if validateMap, validateOk := mapConfig["validate"].(map[string]any); validateOk {
		if timeout, timeoutOk := validateMap["timeout"].(float64); timeoutOk {
			m.config.Validate.Timeout = int(timeout)
		}
		if cooldown, cooldownOk := validateMap["cooldown"].(float64); cooldownOk {
			m.config.Validate.Cooldown = int(cooldown)
		}
	}

	// Convert notification settings
	if notifMap, notifOk := mapConfig["notifications"].(map[string]any); notifOk {
		if topic, topicOk := notifMap["ntfy_topic"].(string); topicOk {
			m.config.Notifications.NtfyTopic = topic
		}
	}
}

// getDefaultValue returns the default value for a key as a string.
func getDefaultValue(defaults *ConfigValues, key string) string {
	switch key {
	case keyValidateTimeout:
		return strconv.Itoa(defaults.Validate.Timeout)
	case keyValidateCooldown:
		return strconv.Itoa(defaults.Validate.Cooldown)
	case keyNotificationsNtfyTopic:
		return defaults.Notifications.NtfyTopic
	default:
		return ""
	}
}

// getConfigFilePath returns the path to the configuration file.
func getConfigFilePath() string {
	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "cc-tools", "config.json")
	}

	// Default to ~/.config/cc-tools/config.json
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if we can't get home
		return "config.json"
	}

	return filepath.Join(homeDir, ".config", "cc-tools", "config.json")
}
