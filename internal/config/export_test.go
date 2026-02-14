package config

// Exports for use by config_test package.

// ExportKeyValidateTimeout returns the unexported keyValidateTimeout constant.
func ExportKeyValidateTimeout() string { return keyValidateTimeout }

// ExportKeyValidateCooldown returns the unexported keyValidateCooldown constant.
func ExportKeyValidateCooldown() string { return keyValidateCooldown }

// ExportKeyNotificationsNtfyTopic returns the unexported keyNotificationsNtfyTopic constant.
func ExportKeyNotificationsNtfyTopic() string { return keyNotificationsNtfyTopic }

// ExportKeyCompactThreshold returns the unexported keyCompactThreshold constant.
func ExportKeyCompactThreshold() string { return keyCompactThreshold }

// ExportKeyCompactReminderInterval returns the unexported keyCompactReminderInterval constant.
func ExportKeyCompactReminderInterval() string { return keyCompactReminderInterval }

// ExportKeyNotifyQuietHoursEnabled returns the unexported key constant.
func ExportKeyNotifyQuietHoursEnabled() string { return keyNotifyQuietHoursEnabled }

// ExportKeyNotifyQuietHoursStart returns the unexported key constant.
func ExportKeyNotifyQuietHoursStart() string { return keyNotifyQuietHoursStart }

// ExportKeyNotifyQuietHoursEnd returns the unexported key constant.
func ExportKeyNotifyQuietHoursEnd() string { return keyNotifyQuietHoursEnd }

// ExportKeyNotifyAudioEnabled returns the unexported key constant.
func ExportKeyNotifyAudioEnabled() string { return keyNotifyAudioEnabled }

// ExportKeyNotifyAudioDirectory returns the unexported key constant.
func ExportKeyNotifyAudioDirectory() string { return keyNotifyAudioDirectory }

// ExportKeyNotifyDesktopEnabled returns the unexported key constant.
func ExportKeyNotifyDesktopEnabled() string { return keyNotifyDesktopEnabled }

// ExportKeyObserveEnabled returns the unexported key constant.
func ExportKeyObserveEnabled() string { return keyObserveEnabled }

// ExportKeyObserveMaxFileSizeMB returns the unexported key constant.
func ExportKeyObserveMaxFileSizeMB() string { return keyObserveMaxFileSizeMB }

// ExportKeyLearningMinSessionLength returns the unexported key constant.
func ExportKeyLearningMinSessionLength() string { return keyLearningMinSessionLength }

// ExportKeyLearningLearnedSkillsPath returns the unexported key constant.
func ExportKeyLearningLearnedSkillsPath() string { return keyLearningLearnedSkillsPath }

// ExportKeyPreCommitEnabled returns the unexported key constant.
func ExportKeyPreCommitEnabled() string { return keyPreCommitEnabled }

// ExportKeyPreCommitCommand returns the unexported key constant.
func ExportKeyPreCommitCommand() string { return keyPreCommitCommand }

// ExportDefaultValidateTimeout returns the unexported defaultValidateTimeout constant.
func ExportDefaultValidateTimeout() int { return defaultValidateTimeout }

// ExportDefaultValidateCooldown returns the unexported defaultValidateCooldown constant.
func ExportDefaultValidateCooldown() int { return defaultValidateCooldown }

// ExportDefaultCompactThreshold returns the unexported defaultCompactThreshold constant.
func ExportDefaultCompactThreshold() int { return defaultCompactThreshold }

// ExportDefaultCompactReminderInterval returns the unexported default constant.
func ExportDefaultCompactReminderInterval() int { return defaultCompactReminderInterval }

// ExportDefaultNotifyQuietHoursEnabled returns the unexported default constant.
func ExportDefaultNotifyQuietHoursEnabled() bool { return defaultNotifyQuietHoursEnabled }

// ExportDefaultNotifyQuietHoursStart returns the unexported default constant.
func ExportDefaultNotifyQuietHoursStart() string { return defaultNotifyQuietHoursStart }

// ExportDefaultNotifyQuietHoursEnd returns the unexported default constant.
func ExportDefaultNotifyQuietHoursEnd() string { return defaultNotifyQuietHoursEnd }

// ExportDefaultNotifyAudioEnabled returns the unexported default constant.
func ExportDefaultNotifyAudioEnabled() bool { return defaultNotifyAudioEnabled }

// ExportDefaultNotifyAudioDirectory returns the unexported default constant.
func ExportDefaultNotifyAudioDirectory() string { return defaultNotifyAudioDirectory }

// ExportDefaultNotifyDesktopEnabled returns the unexported default constant.
func ExportDefaultNotifyDesktopEnabled() bool { return defaultNotifyDesktopEnabled }

// ExportDefaultObserveEnabled returns the unexported default constant.
func ExportDefaultObserveEnabled() bool { return defaultObserveEnabled }

// ExportDefaultObserveMaxFileSizeMB returns the unexported default constant.
func ExportDefaultObserveMaxFileSizeMB() int { return defaultObserveMaxFileSizeMB }

// ExportDefaultLearningMinSessionLength returns the unexported default constant.
func ExportDefaultLearningMinSessionLength() int { return defaultLearningMinSessionLength }

// ExportDefaultLearningLearnedSkillsPath returns the unexported default constant.
func ExportDefaultLearningLearnedSkillsPath() string { return defaultLearningLearnedSkillsPath }

// ExportDefaultPreCommitEnabled returns the unexported default constant.
func ExportDefaultPreCommitEnabled() bool { return defaultPreCommitEnabled }

// ExportDefaultPreCommitCommand returns the unexported default constant.
func ExportDefaultPreCommitCommand() string { return defaultPreCommitCommand }

// ExportGetDefaultConfig exposes GetDefaultConfig for testing.
func ExportGetDefaultConfig() *Values { return GetDefaultConfig() }

// ExportGetDefaultValue exposes getDefaultValue for testing.
func ExportGetDefaultValue(defaults *Values, key string) string {
	return getDefaultValue(defaults, key)
}

// ExportGetConfigFilePath exposes getConfigFilePath for testing.
func ExportGetConfigFilePath() string { return getConfigFilePath() }

// ExportAllKeys exposes allKeys for testing.
func ExportAllKeys() []string { return allKeys() }

// NewTestManager creates a Manager with the given config path and values for testing.
func NewTestManager(configPath string, cfg *Values) *Manager {
	return &Manager{
		configPath: configPath,
		config:     cfg,
	}
}

// ManagerConfig returns the unexported config field from a Manager.
func ManagerConfig(m *Manager) *Values {
	return m.config
}

// ManagerConfigPath returns the unexported configPath field from a Manager.
func ManagerConfigPath(m *Manager) string {
	return m.configPath
}

// ManagerLoadConfig exposes the unexported loadConfig method for testing.
func ManagerLoadConfig(m *Manager) error {
	return m.loadConfig()
}

// ManagerSaveConfig exposes the unexported saveConfig method for testing.
func ManagerSaveConfig(m *Manager) error {
	return m.saveConfig()
}

// ManagerEnsureDefaults exposes the unexported ensureDefaults method for testing.
func ManagerEnsureDefaults(m *Manager) {
	m.ensureDefaults()
}

// ManagerConvertFromMap exposes the unexported convertFromMap method for testing.
func ManagerConvertFromMap(m *Manager, mapConfig map[string]any) {
	m.convertFromMap(mapConfig)
}
