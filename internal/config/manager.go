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

	keyCompactThreshold        = "compact.threshold"
	keyCompactReminderInterval = "compact.reminder_interval"

	keyNotifyQuietHoursEnabled = "notify.quiet_hours.enabled"
	keyNotifyQuietHoursStart   = "notify.quiet_hours.start"
	keyNotifyQuietHoursEnd     = "notify.quiet_hours.end"
	keyNotifyAudioEnabled      = "notify.audio.enabled"
	keyNotifyAudioDirectory    = "notify.audio.directory"
	keyNotifyDesktopEnabled    = "notify.desktop.enabled"

	keyObserveEnabled       = "observe.enabled"
	keyObserveMaxFileSizeMB = "observe.max_file_size_mb"

	keyLearningMinSessionLength  = "learning.min_session_length"
	keyLearningLearnedSkillsPath = "learning.learned_skills_path"

	keyPreCommitEnabled = "pre_commit_reminder.enabled"
	keyPreCommitCommand = "pre_commit_reminder.command"
)

// Values represents the concrete configuration structure.
type Values struct {
	Validate      ValidateValues      `json:"validate"`
	Notifications NotificationsValues `json:"notifications"`
	Compact       CompactValues       `json:"compact"`
	Notify        NotifyValues        `json:"notify"`
	Observe       ObserveValues       `json:"observe"`
	Learning      LearningValues      `json:"learning"`
	PreCommit     PreCommitValues     `json:"pre_commit_reminder"`
}

// NotificationsValues represents notification-related settings.
type NotificationsValues struct {
	NtfyTopic string `json:"ntfy_topic"`
}

// ValidateValues represents validate-related settings.
type ValidateValues struct {
	Timeout  int `json:"timeout"`
	Cooldown int `json:"cooldown"`
}

// CompactValues represents compact context reminder settings.
type CompactValues struct {
	Threshold        int `json:"threshold"`
	ReminderInterval int `json:"reminder_interval"`
}

// NotifyValues represents notification dispatch settings.
type NotifyValues struct {
	QuietHours QuietHoursValues `json:"quiet_hours"`
	Audio      AudioValues      `json:"audio"`
	Desktop    DesktopValues    `json:"desktop"`
}

// QuietHoursValues represents quiet hours configuration.
type QuietHoursValues struct {
	Enabled bool   `json:"enabled"`
	Start   string `json:"start"`
	End     string `json:"end"`
}

// AudioValues represents audio notification settings.
type AudioValues struct {
	Enabled   bool   `json:"enabled"`
	Directory string `json:"directory"`
}

// DesktopValues represents desktop notification settings.
type DesktopValues struct {
	Enabled bool `json:"enabled"`
}

// ObserveValues represents file observation settings.
type ObserveValues struct {
	Enabled       bool `json:"enabled"`
	MaxFileSizeMB int  `json:"max_file_size_mb"`
}

// LearningValues represents learning extraction settings.
type LearningValues struct {
	MinSessionLength  int    `json:"min_session_length"`
	LearnedSkillsPath string `json:"learned_skills_path"`
}

// PreCommitValues represents pre-commit reminder settings.
type PreCommitValues struct {
	Enabled bool   `json:"enabled"`
	Command string `json:"command"`
}

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

const (
	defaultValidateTimeout  = 60
	defaultValidateCooldown = 5

	defaultCompactThreshold        = 50
	defaultCompactReminderInterval = 25

	defaultNotifyQuietHoursEnabled = true
	defaultNotifyQuietHoursStart   = "21:00"
	defaultNotifyQuietHoursEnd     = "07:30"
	defaultNotifyAudioEnabled      = true
	defaultNotifyAudioDirectory    = "~/.claude/audio"
	defaultNotifyDesktopEnabled    = true

	defaultObserveEnabled       = true
	defaultObserveMaxFileSizeMB = 10

	defaultLearningMinSessionLength  = 10
	defaultLearningLearnedSkillsPath = ".claude/skills/learned"

	defaultPreCommitEnabled = true
	defaultPreCommitCommand = "task pre-commit"
)

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
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}
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

	defaults := getDefaultConfig()
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

// allKeys returns all configuration keys in a consistent order.
func allKeys() []string {
	return []string{
		keyValidateTimeout,
		keyValidateCooldown,
		keyNotificationsNtfyTopic,
		keyCompactThreshold,
		keyCompactReminderInterval,
		keyNotifyQuietHoursEnabled,
		keyNotifyQuietHoursStart,
		keyNotifyQuietHoursEnd,
		keyNotifyAudioEnabled,
		keyNotifyAudioDirectory,
		keyNotifyDesktopEnabled,
		keyObserveEnabled,
		keyObserveMaxFileSizeMB,
		keyLearningMinSessionLength,
		keyLearningLearnedSkillsPath,
		keyPreCommitEnabled,
		keyPreCommitCommand,
	}
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
	m.config = getDefaultConfig()
	return m.saveConfig()
}

// getDefaultConfig returns a new config with default values.
func getDefaultConfig() *Values {
	return &Values{
		Validate: ValidateValues{
			Timeout:  defaultValidateTimeout,
			Cooldown: defaultValidateCooldown,
		},
		Notifications: NotificationsValues{
			NtfyTopic: "",
		},
		Compact: CompactValues{
			Threshold:        defaultCompactThreshold,
			ReminderInterval: defaultCompactReminderInterval,
		},
		Notify: NotifyValues{
			QuietHours: QuietHoursValues{
				Enabled: defaultNotifyQuietHoursEnabled,
				Start:   defaultNotifyQuietHoursStart,
				End:     defaultNotifyQuietHoursEnd,
			},
			Audio: AudioValues{
				Enabled:   defaultNotifyAudioEnabled,
				Directory: defaultNotifyAudioDirectory,
			},
			Desktop: DesktopValues{
				Enabled: defaultNotifyDesktopEnabled,
			},
		},
		Observe: ObserveValues{
			Enabled:       defaultObserveEnabled,
			MaxFileSizeMB: defaultObserveMaxFileSizeMB,
		},
		Learning: LearningValues{
			MinSessionLength:  defaultLearningMinSessionLength,
			LearnedSkillsPath: defaultLearningLearnedSkillsPath,
		},
		PreCommit: PreCommitValues{
			Enabled: defaultPreCommitEnabled,
			Command: defaultPreCommitCommand,
		},
	}
}

// ensureDefaults ensures all fields have values, using defaults for missing fields.
// Boolean fields are not checked here because we unmarshal into a defaults struct,
// which preserves default true values when the field is absent from JSON.
func (m *Manager) ensureDefaults() {
	defaults := getDefaultConfig()

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
}

// convertFromMap converts the old map-based config to the new structured format.
func (m *Manager) convertFromMap(mapConfig map[string]any) {
	// Initialize with defaults
	m.config = getDefaultConfig()

	convertValidateFromMap(&m.config.Validate, mapConfig)
	convertNotificationsFromMap(&m.config.Notifications, mapConfig)
	convertCompactFromMap(&m.config.Compact, mapConfig)
	convertObserveFromMap(&m.config.Observe, mapConfig)
	convertLearningFromMap(&m.config.Learning, mapConfig)
	convertPreCommitFromMap(&m.config.PreCommit, mapConfig)

	if notifyMap, notifyOk := mapConfig["notify"].(map[string]any); notifyOk {
		convertNotifyFromMap(&m.config.Notify, notifyMap)
	}
}

// convertValidateFromMap extracts validate settings from a map config.
func convertValidateFromMap(v *ValidateValues, mapConfig map[string]any) {
	section, sectionOk := mapConfig["validate"].(map[string]any)
	if !sectionOk {
		return
	}
	if timeout, timeoutOk := section["timeout"].(float64); timeoutOk {
		v.Timeout = int(timeout)
	}
	if cooldown, cooldownOk := section["cooldown"].(float64); cooldownOk {
		v.Cooldown = int(cooldown)
	}
}

// convertNotificationsFromMap extracts notification settings from a map config.
func convertNotificationsFromMap(n *NotificationsValues, mapConfig map[string]any) {
	section, sectionOk := mapConfig["notifications"].(map[string]any)
	if !sectionOk {
		return
	}
	if topic, topicOk := section["ntfy_topic"].(string); topicOk {
		n.NtfyTopic = topic
	}
}

// convertCompactFromMap extracts compact settings from a map config.
func convertCompactFromMap(c *CompactValues, mapConfig map[string]any) {
	section, sectionOk := mapConfig["compact"].(map[string]any)
	if !sectionOk {
		return
	}
	if threshold, thresholdOk := section["threshold"].(float64); thresholdOk {
		c.Threshold = int(threshold)
	}
	if interval, intervalOk := section["reminder_interval"].(float64); intervalOk {
		c.ReminderInterval = int(interval)
	}
}

// convertNotifyFromMap extracts notify settings (quiet hours, audio, desktop) from a map.
func convertNotifyFromMap(n *NotifyValues, notifyMap map[string]any) {
	if qhMap, qhOk := notifyMap["quiet_hours"].(map[string]any); qhOk {
		if enabled, enabledOk := qhMap["enabled"].(bool); enabledOk {
			n.QuietHours.Enabled = enabled
		}
		if start, startOk := qhMap["start"].(string); startOk {
			n.QuietHours.Start = start
		}
		if end, endOk := qhMap["end"].(string); endOk {
			n.QuietHours.End = end
		}
	}
	if audioMap, audioOk := notifyMap["audio"].(map[string]any); audioOk {
		if enabled, enabledOk := audioMap["enabled"].(bool); enabledOk {
			n.Audio.Enabled = enabled
		}
		if dir, dirOk := audioMap["directory"].(string); dirOk {
			n.Audio.Directory = dir
		}
	}
	if desktopMap, desktopOk := notifyMap["desktop"].(map[string]any); desktopOk {
		if enabled, enabledOk := desktopMap["enabled"].(bool); enabledOk {
			n.Desktop.Enabled = enabled
		}
	}
}

// convertObserveFromMap extracts observe settings from a map config.
func convertObserveFromMap(o *ObserveValues, mapConfig map[string]any) {
	section, sectionOk := mapConfig["observe"].(map[string]any)
	if !sectionOk {
		return
	}
	if enabled, enabledOk := section["enabled"].(bool); enabledOk {
		o.Enabled = enabled
	}
	if maxSize, maxSizeOk := section["max_file_size_mb"].(float64); maxSizeOk {
		o.MaxFileSizeMB = int(maxSize)
	}
}

// convertLearningFromMap extracts learning settings from a map config.
func convertLearningFromMap(l *LearningValues, mapConfig map[string]any) {
	section, sectionOk := mapConfig["learning"].(map[string]any)
	if !sectionOk {
		return
	}
	if minLen, minLenOk := section["min_session_length"].(float64); minLenOk {
		l.MinSessionLength = int(minLen)
	}
	if path, pathOk := section["learned_skills_path"].(string); pathOk {
		l.LearnedSkillsPath = path
	}
}

// convertPreCommitFromMap extracts pre-commit reminder settings from a map config.
func convertPreCommitFromMap(p *PreCommitValues, mapConfig map[string]any) {
	section, sectionOk := mapConfig["pre_commit_reminder"].(map[string]any)
	if !sectionOk {
		return
	}
	if enabled, enabledOk := section["enabled"].(bool); enabledOk {
		p.Enabled = enabled
	}
	if cmd, cmdOk := section["command"].(string); cmdOk {
		p.Command = cmd
	}
}

// getDefaultValue returns the default value for a key as a string.
func getDefaultValue(defaults *Values, key string) string {
	switch key {
	case keyValidateTimeout:
		return strconv.Itoa(defaults.Validate.Timeout)
	case keyValidateCooldown:
		return strconv.Itoa(defaults.Validate.Cooldown)
	case keyNotificationsNtfyTopic:
		return defaults.Notifications.NtfyTopic
	case keyCompactThreshold:
		return strconv.Itoa(defaults.Compact.Threshold)
	case keyCompactReminderInterval:
		return strconv.Itoa(defaults.Compact.ReminderInterval)
	case keyNotifyQuietHoursEnabled:
		return strconv.FormatBool(defaults.Notify.QuietHours.Enabled)
	case keyNotifyQuietHoursStart:
		return defaults.Notify.QuietHours.Start
	case keyNotifyQuietHoursEnd:
		return defaults.Notify.QuietHours.End
	case keyNotifyAudioEnabled:
		return strconv.FormatBool(defaults.Notify.Audio.Enabled)
	case keyNotifyAudioDirectory:
		return defaults.Notify.Audio.Directory
	case keyNotifyDesktopEnabled:
		return strconv.FormatBool(defaults.Notify.Desktop.Enabled)
	case keyObserveEnabled:
		return strconv.FormatBool(defaults.Observe.Enabled)
	case keyObserveMaxFileSizeMB:
		return strconv.Itoa(defaults.Observe.MaxFileSizeMB)
	case keyLearningMinSessionLength:
		return strconv.Itoa(defaults.Learning.MinSessionLength)
	case keyLearningLearnedSkillsPath:
		return defaults.Learning.LearnedSkillsPath
	case keyPreCommitEnabled:
		return strconv.FormatBool(defaults.PreCommit.Enabled)
	case keyPreCommitCommand:
		return defaults.PreCommit.Command
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
