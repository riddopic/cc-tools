package config

// Exports for use by config_test package.

// ExportKeyValidateTimeout returns the unexported keyValidateTimeout constant.
func ExportKeyValidateTimeout() string { return keyValidateTimeout }

// ExportKeyValidateCooldown returns the unexported keyValidateCooldown constant.
func ExportKeyValidateCooldown() string { return keyValidateCooldown }

// ExportKeyNotificationsNtfyTopic returns the unexported keyNotificationsNtfyTopic constant.
func ExportKeyNotificationsNtfyTopic() string { return keyNotificationsNtfyTopic }

// ExportDefaultValidateTimeout returns the unexported defaultValidateTimeout constant.
func ExportDefaultValidateTimeout() int { return defaultValidateTimeout }

// ExportDefaultValidateCooldown returns the unexported defaultValidateCooldown constant.
func ExportDefaultValidateCooldown() int { return defaultValidateCooldown }

// ExportGetDefaultConfig exposes getDefaultConfig for testing.
func ExportGetDefaultConfig() *Values { return getDefaultConfig() }

// ExportGetDefaultValue exposes getDefaultValue for testing.
func ExportGetDefaultValue(defaults *Values, key string) string {
	return getDefaultValue(defaults, key)
}

// ExportGetConfigFilePath exposes getConfigFilePath for testing.
func ExportGetConfigFilePath() string { return getConfigFilePath() }

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
