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

// Manager handles configuration read/write operations.
type Manager struct {
	configPath string
	config     *Values
}

// Info contains information about a configuration value.
type Info struct {
	Value     string
	IsDefault bool
}

// NewManager creates a new configuration manager.
func NewManager() *Manager {
	return &Manager{
		configPath: getConfigFilePath(),
		config:     nil,
	}
}

// NewManagerWithPath creates a new configuration manager with a specific config file path.
func NewManagerWithPath(path string) *Manager {
	return &Manager{
		configPath: path,
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
	case keyCompactThreshold:
		return m.config.Compact.Threshold, true, nil
	case keyCompactReminderInterval:
		return m.config.Compact.ReminderInterval, true, nil
	case keyObserveMaxFileSizeMB:
		return m.config.Observe.MaxFileSizeMB, true, nil
	case keyLearningMinSessionLength:
		return m.config.Learning.MinSessionLength, true, nil
	case keyDriftMinEdits:
		return m.config.Drift.MinEdits, true, nil
	case keyStopReminderInterval:
		return m.config.StopReminder.Interval, true, nil
	case keyStopReminderWarnAt:
		return m.config.StopReminder.WarnAt, true, nil
	case keyInstinctMaxInstincts:
		return m.config.Instinct.MaxInstincts, true, nil
	case keyInstinctClusterThreshold:
		return m.config.Instinct.ClusterThreshold, true, nil
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
	case keyNotifyQuietHoursStart:
		return m.config.Notify.QuietHours.Start, true, nil
	case keyNotifyQuietHoursEnd:
		return m.config.Notify.QuietHours.End, true, nil
	case keyNotifyAudioDirectory:
		return m.config.Notify.Audio.Directory, true, nil
	case keyLearningLearnedSkillsPath:
		return m.config.Learning.LearnedSkillsPath, true, nil
	case keyPreCommitCommand:
		return m.config.PreCommit.Command, true, nil
	case keyPackageManagerPreferred:
		return m.config.PackageManager.Preferred, true, nil
	case keyInstinctPersonalPath:
		return m.config.Instinct.PersonalPath, true, nil
	case keyInstinctInheritedPath:
		return m.config.Instinct.InheritedPath, true, nil
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
	case keyCompactThreshold:
		return strconv.Itoa(m.config.Compact.Threshold), true, nil
	case keyCompactReminderInterval:
		return strconv.Itoa(m.config.Compact.ReminderInterval), true, nil
	case keyNotifyQuietHoursEnabled:
		return strconv.FormatBool(m.config.Notify.QuietHours.Enabled), true, nil
	case keyNotifyQuietHoursStart:
		return m.config.Notify.QuietHours.Start, true, nil
	case keyNotifyQuietHoursEnd:
		return m.config.Notify.QuietHours.End, true, nil
	case keyNotifyAudioEnabled:
		return strconv.FormatBool(m.config.Notify.Audio.Enabled), true, nil
	case keyNotifyAudioDirectory:
		return m.config.Notify.Audio.Directory, true, nil
	case keyNotifyDesktopEnabled:
		return strconv.FormatBool(m.config.Notify.Desktop.Enabled), true, nil
	case keyObserveEnabled:
		return strconv.FormatBool(m.config.Observe.Enabled), true, nil
	case keyObserveMaxFileSizeMB:
		return strconv.Itoa(m.config.Observe.MaxFileSizeMB), true, nil
	case keyLearningMinSessionLength:
		return strconv.Itoa(m.config.Learning.MinSessionLength), true, nil
	case keyLearningLearnedSkillsPath:
		return m.config.Learning.LearnedSkillsPath, true, nil
	case keyPreCommitEnabled:
		return strconv.FormatBool(m.config.PreCommit.Enabled), true, nil
	case keyPreCommitCommand:
		return m.config.PreCommit.Command, true, nil
	case keyPackageManagerPreferred:
		return m.config.PackageManager.Preferred, true, nil
	default:
		return m.config.getExtendedValue(key)
	}
}

// Set updates a configuration value.
func (m *Manager) Set(_ context.Context, key string, value string) error {
	if m.config == nil {
		if err := m.loadConfig(); err != nil {
			return fmt.Errorf("load config: %w", err)
		}
	}

	if err := m.setField(key, value); err != nil {
		return err
	}

	// Save to file
	if err := m.saveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	return nil
}

// setField dispatches the value assignment to the correct config field.
func (m *Manager) setField(key string, value string) error {
	switch key {
	case keyValidateTimeout:
		return setIntField(&m.config.Validate.Timeout, value)
	case keyValidateCooldown:
		return setIntField(&m.config.Validate.Cooldown, value)
	case keyNotificationsNtfyTopic:
		m.config.Notifications.NtfyTopic = value
	case keyCompactThreshold:
		return setIntField(&m.config.Compact.Threshold, value)
	case keyCompactReminderInterval:
		return setIntField(&m.config.Compact.ReminderInterval, value)
	case keyNotifyQuietHoursEnabled:
		return setBoolField(&m.config.Notify.QuietHours.Enabled, value)
	case keyNotifyQuietHoursStart:
		m.config.Notify.QuietHours.Start = value
	case keyNotifyQuietHoursEnd:
		m.config.Notify.QuietHours.End = value
	case keyNotifyAudioEnabled:
		return setBoolField(&m.config.Notify.Audio.Enabled, value)
	case keyNotifyAudioDirectory:
		m.config.Notify.Audio.Directory = value
	case keyNotifyDesktopEnabled:
		return setBoolField(&m.config.Notify.Desktop.Enabled, value)
	case keyObserveEnabled:
		return setBoolField(&m.config.Observe.Enabled, value)
	case keyObserveMaxFileSizeMB:
		return setIntField(&m.config.Observe.MaxFileSizeMB, value)
	case keyLearningMinSessionLength:
		return setIntField(&m.config.Learning.MinSessionLength, value)
	case keyLearningLearnedSkillsPath:
		m.config.Learning.LearnedSkillsPath = value
	case keyPreCommitEnabled:
		return setBoolField(&m.config.PreCommit.Enabled, value)
	case keyPreCommitCommand:
		m.config.PreCommit.Command = value
	case keyPackageManagerPreferred:
		m.config.PackageManager.Preferred = value
	default:
		if handled, err := m.config.setExtendedField(key, value); handled {
			return err
		}
		return fmt.Errorf("unknown configuration key: %s", key)
	}
	return nil
}

// setFloatField parses and assigns a float64 value to the given field.
func setFloatField(field *float64, value string) error {
	floatVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("value must be a number: %w", err)
	}
	*field = floatVal
	return nil
}

// setIntField parses and assigns an integer value to the given field.
func setIntField(field *int, value string) error {
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("value must be an integer: %w", err)
	}
	*field = intVal
	return nil
}

// setBoolField parses and assigns a boolean value to the given field.
func setBoolField(field *bool, value string) error {
	boolVal, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("value must be a boolean: %w", err)
	}
	*field = boolVal
	return nil
}

// GetAll retrieves all configuration values with their metadata.
func (m *Manager) GetAll(ctx context.Context) (map[string]Info, error) {
	if m.config == nil {
		if err := m.loadConfig(); err != nil {
			return nil, fmt.Errorf("load config: %w", err)
		}
	}

	defaults := GetDefaultConfig()
	result := make(map[string]Info)

	// Process all configuration keys
	keys := allKeys()

	for _, key := range keys {
		value, _, _ := m.GetValue(ctx, key)
		defaultValue := getDefaultValue(defaults, key)

		result[key] = Info{
			Value:     value,
			IsDefault: value == defaultValue,
		}
	}

	return result, nil
}

// GetAllKeys returns all available configuration keys.
func (m *Manager) GetAllKeys(_ context.Context) ([]string, error) {
	keys := allKeys()
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

	defaults := GetDefaultConfig()

	// Reset to default value
	switch key {
	case keyValidateTimeout:
		m.config.Validate.Timeout = defaults.Validate.Timeout
	case keyValidateCooldown:
		m.config.Validate.Cooldown = defaults.Validate.Cooldown
	case keyNotificationsNtfyTopic:
		m.config.Notifications.NtfyTopic = defaults.Notifications.NtfyTopic
	case keyCompactThreshold:
		m.config.Compact.Threshold = defaults.Compact.Threshold
	case keyCompactReminderInterval:
		m.config.Compact.ReminderInterval = defaults.Compact.ReminderInterval
	case keyNotifyQuietHoursEnabled:
		m.config.Notify.QuietHours.Enabled = defaults.Notify.QuietHours.Enabled
	case keyNotifyQuietHoursStart:
		m.config.Notify.QuietHours.Start = defaults.Notify.QuietHours.Start
	case keyNotifyQuietHoursEnd:
		m.config.Notify.QuietHours.End = defaults.Notify.QuietHours.End
	case keyNotifyAudioEnabled:
		m.config.Notify.Audio.Enabled = defaults.Notify.Audio.Enabled
	case keyNotifyAudioDirectory:
		m.config.Notify.Audio.Directory = defaults.Notify.Audio.Directory
	case keyNotifyDesktopEnabled:
		m.config.Notify.Desktop.Enabled = defaults.Notify.Desktop.Enabled
	case keyObserveEnabled:
		m.config.Observe.Enabled = defaults.Observe.Enabled
	case keyObserveMaxFileSizeMB:
		m.config.Observe.MaxFileSizeMB = defaults.Observe.MaxFileSizeMB
	case keyLearningMinSessionLength:
		m.config.Learning.MinSessionLength = defaults.Learning.MinSessionLength
	case keyLearningLearnedSkillsPath:
		m.config.Learning.LearnedSkillsPath = defaults.Learning.LearnedSkillsPath
	case keyPreCommitEnabled:
		m.config.PreCommit.Enabled = defaults.PreCommit.Enabled
	case keyPreCommitCommand:
		m.config.PreCommit.Command = defaults.PreCommit.Command
	case keyPackageManagerPreferred:
		m.config.PackageManager.Preferred = defaults.PackageManager.Preferred
	default:
		if !m.config.resetExtended(key, defaults) {
			return fmt.Errorf("unknown configuration key: %s", key)
		}
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
	m.config = GetDefaultConfig()

	// Save to file
	if err := m.saveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	return nil
}

// GetConfig returns the current configuration structure.
func (m *Manager) GetConfig(_ context.Context) (*Values, error) {
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
	m.config = GetDefaultConfig()

	// Read file if it exists
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, use defaults
			return nil
		}
		return fmt.Errorf("read config file: %w", err)
	}

	// Try to parse as structured config first, unmarshaling into defaults
	// so that missing fields retain their default values (especially booleans).
	if unmarshalErr := json.Unmarshal(data, m.config); unmarshalErr == nil {
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
	m.config = GetDefaultConfig()
	return m.saveConfig()
}

// ensureDefaults ensures all fields have values, using defaults for missing fields.
// Boolean fields are not checked here because we unmarshal into a defaults struct,
// which preserves default true values when the field is absent from JSON.
func (m *Manager) ensureDefaults() {
	defaults := GetDefaultConfig()

	if m.config.Validate.Timeout == 0 {
		m.config.Validate.Timeout = defaults.Validate.Timeout
	}
	if m.config.Validate.Cooldown == 0 {
		m.config.Validate.Cooldown = defaults.Validate.Cooldown
	}
	if m.config.Compact.Threshold == 0 {
		m.config.Compact.Threshold = defaults.Compact.Threshold
	}
	if m.config.Compact.ReminderInterval == 0 {
		m.config.Compact.ReminderInterval = defaults.Compact.ReminderInterval
	}
	if m.config.Notify.QuietHours.Start == "" {
		m.config.Notify.QuietHours.Start = defaults.Notify.QuietHours.Start
	}
	if m.config.Notify.QuietHours.End == "" {
		m.config.Notify.QuietHours.End = defaults.Notify.QuietHours.End
	}
	if m.config.Notify.Audio.Directory == "" {
		m.config.Notify.Audio.Directory = defaults.Notify.Audio.Directory
	}
	if m.config.Observe.MaxFileSizeMB == 0 {
		m.config.Observe.MaxFileSizeMB = defaults.Observe.MaxFileSizeMB
	}
	if m.config.Learning.MinSessionLength == 0 {
		m.config.Learning.MinSessionLength = defaults.Learning.MinSessionLength
	}
	if m.config.Learning.LearnedSkillsPath == "" {
		m.config.Learning.LearnedSkillsPath = defaults.Learning.LearnedSkillsPath
	}
	if m.config.PreCommit.Command == "" {
		m.config.PreCommit.Command = defaults.PreCommit.Command
	}
	if m.config.Drift.MinEdits == 0 {
		m.config.Drift.MinEdits = defaults.Drift.MinEdits
	}
	if m.config.Drift.Threshold == 0 {
		m.config.Drift.Threshold = defaults.Drift.Threshold
	}
	if m.config.StopReminder.Interval == 0 {
		m.config.StopReminder.Interval = defaults.StopReminder.Interval
	}
	if m.config.StopReminder.WarnAt == 0 {
		m.config.StopReminder.WarnAt = defaults.StopReminder.WarnAt
	}
	if m.config.Instinct.PersonalPath == "" {
		m.config.Instinct.PersonalPath = defaults.Instinct.PersonalPath
	}
	if m.config.Instinct.InheritedPath == "" {
		m.config.Instinct.InheritedPath = defaults.Instinct.InheritedPath
	}
	if m.config.Instinct.MinConfidence == 0 {
		m.config.Instinct.MinConfidence = defaults.Instinct.MinConfidence
	}
	if m.config.Instinct.AutoApprove == 0 {
		m.config.Instinct.AutoApprove = defaults.Instinct.AutoApprove
	}
	if m.config.Instinct.DecayRate == 0 {
		m.config.Instinct.DecayRate = defaults.Instinct.DecayRate
	}
	if m.config.Instinct.MaxInstincts == 0 {
		m.config.Instinct.MaxInstincts = defaults.Instinct.MaxInstincts
	}
	if m.config.Instinct.ClusterThreshold == 0 {
		m.config.Instinct.ClusterThreshold = defaults.Instinct.ClusterThreshold
	}
}

// convertFromMap converts the old map-based config to the new structured format.
func (m *Manager) convertFromMap(mapConfig map[string]any) {
	// Initialize with defaults
	m.config = GetDefaultConfig()

	convertValidateFromMap(&m.config.Validate, mapConfig)
	convertNotificationsFromMap(&m.config.Notifications, mapConfig)
	convertCompactFromMap(&m.config.Compact, mapConfig)
	convertObserveFromMap(&m.config.Observe, mapConfig)
	convertLearningFromMap(&m.config.Learning, mapConfig)
	convertPreCommitFromMap(&m.config.PreCommit, mapConfig)
	convertPackageManagerFromMap(&m.config.PackageManager, mapConfig)
	convertDriftFromMap(&m.config.Drift, mapConfig)
	convertStopReminderFromMap(&m.config.StopReminder, mapConfig)
	convertInstinctFromMap(&m.config.Instinct, mapConfig)

	if notifyMap, notifyOk := mapConfig["notify"].(map[string]any); notifyOk {
		convertNotifyFromMap(&m.config.Notify, notifyMap)
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
