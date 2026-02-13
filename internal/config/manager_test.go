package config

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	// Save original env
	origHome := os.Getenv("XDG_CONFIG_HOME")
	defer func() { os.Setenv("XDG_CONFIG_HOME", origHome) }()

	tests := []struct {
		name        string
		xdgHome     string
		wantPathEnd string
	}{
		{
			name:        "uses XDG_CONFIG_HOME when set",
			xdgHome:     "/custom/config",
			wantPathEnd: "/custom/config/cc-tools/config.json",
		},
		{
			name:        "falls back to home directory",
			xdgHome:     "",
			wantPathEnd: "/.config/cc-tools/config.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("XDG_CONFIG_HOME", tt.xdgHome)
			m := NewManager()

			if !filepath.IsAbs(m.configPath) {
				t.Errorf("config path should be absolute, got %s", m.configPath)
			}

			if tt.xdgHome != "" && m.configPath != tt.wantPathEnd {
				t.Errorf("unexpected config path = %s, want %s", m.configPath, tt.wantPathEnd)
			}
		})
	}
}

func TestEnsureConfig(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, configPath string)
		wantErr   bool
		checkFunc func(t *testing.T, m *Manager)
	}{
		{
			name: "creates config file when missing",
			setupFunc: func(t *testing.T, configPath string) {
				// Ensure parent dir exists but config file doesn't
				os.MkdirAll(filepath.Dir(configPath), 0755)
			},
			wantErr: false,
			checkFunc: func(t *testing.T, m *Manager) {
				if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
					t.Error("config file should have been created")
				}

				// Check defaults are loaded
				if m.config == nil {
					t.Fatal("config should be loaded")
				}
				if m.config.Validate.Timeout != defaultValidateTimeout {
					t.Errorf("timeout = %d, want %d", m.config.Validate.Timeout, defaultValidateTimeout)
				}
			},
		},
		{
			name: "loads existing config file",
			setupFunc: func(t *testing.T, configPath string) {
				os.MkdirAll(filepath.Dir(configPath), 0755)
				config := &ConfigValues{
					Validate: ValidateConfigValues{
						Timeout:  120,
						Cooldown: 10,
					},
				}
				data, _ := json.MarshalIndent(config, "", "  ")
				os.WriteFile(configPath, data, 0600)
			},
			wantErr: false,
			checkFunc: func(t *testing.T, m *Manager) {
				if m.config == nil {
					t.Fatal("config should be loaded")
				}
				if m.config.Validate.Timeout != 120 {
					t.Errorf("timeout = %d, want 120", m.config.Validate.Timeout)
				}
			},
		},
		{
			name: "handles corrupt config file",
			setupFunc: func(t *testing.T, configPath string) {
				os.MkdirAll(filepath.Dir(configPath), 0755)
				os.WriteFile(configPath, []byte("invalid json"), 0600)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")

			m := &Manager{
				configPath: configPath,
			}

			if tt.setupFunc != nil {
				tt.setupFunc(t, configPath)
			}

			err := m.EnsureConfig(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && tt.checkFunc != nil {
				tt.checkFunc(t, m)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		config    *ConfigValues
		key       string
		wantValue int
		wantFound bool
		wantErr   bool
	}{
		{
			name: "get validate timeout",
			config: &ConfigValues{
				Validate: ValidateConfigValues{Timeout: 90},
			},
			key:       keyValidateTimeout,
			wantValue: 90,
			wantFound: true,
		},
		{
			name: "get validate cooldown",
			config: &ConfigValues{
				Validate: ValidateConfigValues{Cooldown: 15},
			},
			key:       keyValidateCooldown,
			wantValue: 15,
			wantFound: true,
		},
		{
			name:      "unknown key returns not found",
			config:    &ConfigValues{},
			key:       "unknown.key",
			wantValue: 0,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			m := &Manager{
				configPath: filepath.Join(tmpDir, "config.json"),
				config:     tt.config,
			}

			value, found, err := m.GetInt(ctx, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInt() error = %v, wantErr %v", err, tt.wantErr)
			}
			if value != tt.wantValue {
				t.Errorf("GetInt() value = %d, want %d", value, tt.wantValue)
			}
			if found != tt.wantFound {
				t.Errorf("GetInt() found = %v, want %v", found, tt.wantFound)
			}
		})
	}
}

func TestGetString(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		config    *ConfigValues
		key       string
		wantValue string
		wantFound bool
	}{
		{
			name:      "unknown key returns not found",
			config:    &ConfigValues{},
			key:       "unknown.key",
			wantValue: "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			m := &Manager{
				configPath: filepath.Join(tmpDir, "config.json"),
				config:     tt.config,
			}

			value, found, err := m.GetString(ctx, tt.key)
			if err != nil {
				t.Errorf("GetString() unexpected error = %v", err)
			}
			if value != tt.wantValue {
				t.Errorf("GetString() value = %s, want %s", value, tt.wantValue)
			}
			if found != tt.wantFound {
				t.Errorf("GetString() found = %v, want %v", found, tt.wantFound)
			}
		})
	}
}

func TestGetValue(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		config    *ConfigValues
		key       string
		wantValue string
		wantFound bool
	}{
		{
			name: "get int value as string",
			config: &ConfigValues{
				Validate: ValidateConfigValues{Timeout: 45},
			},
			key:       keyValidateTimeout,
			wantValue: "45",
			wantFound: true,
		},
		{
			name:      "unknown key",
			config:    &ConfigValues{},
			key:       "unknown",
			wantValue: "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			m := &Manager{
				configPath: filepath.Join(tmpDir, "config.json"),
				config:     tt.config,
			}

			value, found, err := m.GetValue(ctx, tt.key)
			if err != nil {
				t.Errorf("GetValue() unexpected error = %v", err)
			}
			if value != tt.wantValue {
				t.Errorf("GetValue() value = %s, want %s", value, tt.wantValue)
			}
			if found != tt.wantFound {
				t.Errorf("GetValue() found = %v, want %v", found, tt.wantFound)
			}
		})
	}
}

func TestSet(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		key       string
		value     string
		wantErr   bool
		checkFunc func(t *testing.T, m *Manager)
	}{
		{
			name:  "set validate timeout",
			key:   keyValidateTimeout,
			value: "180",
			checkFunc: func(t *testing.T, m *Manager) {
				if m.config.Validate.Timeout != 180 {
					t.Errorf("timeout = %d, want 180", m.config.Validate.Timeout)
				}
			},
		},
		{
			name:    "invalid int value",
			key:     keyValidateTimeout,
			value:   "not-a-number",
			wantErr: true,
		},
		{
			name:    "unknown key",
			key:     "unknown.key",
			value:   "value",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			m := &Manager{
				configPath: filepath.Join(tmpDir, "config.json"),
				config:     getDefaultConfig(),
			}

			err := m.Set(ctx, tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				// Verify config was saved to file
				data, readErr := os.ReadFile(m.configPath)
				if readErr != nil {
					t.Fatalf("Failed to read config file: %v", readErr)
				}

				var savedConfig ConfigValues
				if unmarshalErr := json.Unmarshal(data, &savedConfig); unmarshalErr != nil {
					t.Fatalf("Failed to unmarshal saved config: %v", unmarshalErr)
				}

				if tt.checkFunc != nil {
					tt.checkFunc(t, m)
				}
			}
		})
	}
}

func TestGetAll(t *testing.T) {
	ctx := context.Background()

	config := &ConfigValues{
		Validate: ValidateConfigValues{
			Timeout:  defaultValidateTimeout,
			Cooldown: 10, // non-default
		},
	}

	tmpDir := t.TempDir()
	m := &Manager{
		configPath: filepath.Join(tmpDir, "config.json"),
		config:     config,
	}

	all, err := m.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	// Check that all keys are present
	expectedKeys := []string{
		keyValidateTimeout,
		keyValidateCooldown,
	}

	for _, key := range expectedKeys {
		if _, ok := all[key]; !ok {
			t.Errorf("GetAll() missing key %s", key)
		}
	}

	// Check IsDefault flags
	if !all[keyValidateTimeout].IsDefault {
		t.Error("validate.timeout should be marked as default")
	}
	if all[keyValidateCooldown].IsDefault {
		t.Error("validate.cooldown should not be marked as default")
	}
}

func TestGetAllKeys(t *testing.T) {
	ctx := context.Background()

	m := NewManager()
	keys, err := m.GetAllKeys(ctx)
	if err != nil {
		t.Fatalf("GetAllKeys() error = %v", err)
	}

	expectedKeys := []string{
		"validate.cooldown",
		"validate.timeout",
	}

	if len(keys) != len(expectedKeys) {
		t.Errorf("GetAllKeys() returned %d keys, want %d", len(keys), len(expectedKeys))
	}

	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("GetAllKeys()[%d] = %s, want %s", i, key, expectedKeys[i])
		}
	}
}

func TestReset(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		key       string
		initValue string
		wantErr   bool
		checkFunc func(t *testing.T, m *Manager)
	}{
		{
			name:      "reset validate timeout",
			key:       keyValidateTimeout,
			initValue: "999",
			checkFunc: func(t *testing.T, m *Manager) {
				if m.config.Validate.Timeout != defaultValidateTimeout {
					t.Errorf("timeout = %d, want %d", m.config.Validate.Timeout, defaultValidateTimeout)
				}
			},
		},
		{
			name:    "unknown key",
			key:     "unknown.key",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			m := &Manager{
				configPath: filepath.Join(tmpDir, "config.json"),
				config:     getDefaultConfig(),
			}

			// Set initial non-default value if provided
			if tt.initValue != "" && !tt.wantErr {
				m.Set(ctx, tt.key, tt.initValue)
			}

			err := m.Reset(ctx, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reset() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && tt.checkFunc != nil {
				tt.checkFunc(t, m)

				// Verify config was saved
				data, _ := os.ReadFile(m.configPath)
				var savedConfig ConfigValues
				json.Unmarshal(data, &savedConfig)
			}
		})
	}
}

func TestResetAll(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	m := &Manager{
		configPath: filepath.Join(tmpDir, "config.json"),
		config: &ConfigValues{
			Validate: ValidateConfigValues{
				Timeout:  999,
				Cooldown: 999,
			},
		},
	}

	err := m.ResetAll(ctx)
	if err != nil {
		t.Fatalf("ResetAll() error = %v", err)
	}

	// Check all values are reset to defaults
	defaults := getDefaultConfig()
	if m.config.Validate.Timeout != defaults.Validate.Timeout {
		t.Errorf("timeout not reset to default")
	}
	if m.config.Validate.Cooldown != defaults.Validate.Cooldown {
		t.Errorf("cooldown not reset to default")
	}

	// Verify saved to file
	data, _ := os.ReadFile(m.configPath)
	var savedConfig ConfigValues
	json.Unmarshal(data, &savedConfig)
	if savedConfig.Validate.Timeout != defaults.Validate.Timeout {
		t.Error("saved config not reset to defaults")
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T, configPath string)
		checkFunc func(t *testing.T, m *Manager)
		wantErr   bool
	}{
		{
			name: "loads structured config",
			setupFunc: func(t *testing.T, configPath string) {
				config := &ConfigValues{
					Validate: ValidateConfigValues{
						Timeout:  120,
						Cooldown: 10,
					},
				}
				data, _ := json.MarshalIndent(config, "", "  ")
				os.MkdirAll(filepath.Dir(configPath), 0755)
				os.WriteFile(configPath, data, 0600)
			},
			checkFunc: func(t *testing.T, m *Manager) {
				if m.config.Validate.Timeout != 120 {
					t.Errorf("timeout = %d, want 120", m.config.Validate.Timeout)
				}
			},
		},
		{
			name: "loads map-based config for backward compatibility",
			setupFunc: func(t *testing.T, configPath string) {
				mapConfig := map[string]any{
					"validate": map[string]any{
						"timeout":  90.0,
						"cooldown": 5.0,
					},
				}
				data, _ := json.MarshalIndent(mapConfig, "", "  ")
				os.MkdirAll(filepath.Dir(configPath), 0755)
				os.WriteFile(configPath, data, 0600)
			},
			checkFunc: func(t *testing.T, m *Manager) {
				if m.config.Validate.Timeout != 90 {
					t.Errorf("timeout = %d, want 90", m.config.Validate.Timeout)
				}
			},
		},
		{
			name: "uses defaults when file doesn't exist",
			checkFunc: func(t *testing.T, m *Manager) {
				defaults := getDefaultConfig()
				if m.config.Validate.Timeout != defaults.Validate.Timeout {
					t.Errorf("timeout = %d, want %d", m.config.Validate.Timeout, defaults.Validate.Timeout)
				}
			},
		},
		{
			name: "fills in missing fields with defaults",
			setupFunc: func(t *testing.T, configPath string) {
				// Partial config with some fields missing
				config := &ConfigValues{
					Validate: ValidateConfigValues{
						Timeout: 100,
						// Cooldown missing - should use default
					},
				}
				data, _ := json.MarshalIndent(config, "", "  ")
				os.MkdirAll(filepath.Dir(configPath), 0755)
				os.WriteFile(configPath, data, 0600)
			},
			checkFunc: func(t *testing.T, m *Manager) {
				if m.config.Validate.Timeout != 100 {
					t.Errorf("timeout = %d, want 100", m.config.Validate.Timeout)
				}
				if m.config.Validate.Cooldown != defaultValidateCooldown {
					t.Errorf("cooldown = %d, want default %d", m.config.Validate.Cooldown, defaultValidateCooldown)
				}
			},
		},
		{
			name: "handles corrupt JSON",
			setupFunc: func(t *testing.T, configPath string) {
				os.MkdirAll(filepath.Dir(configPath), 0755)
				os.WriteFile(configPath, []byte("{invalid json}"), 0600)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")

			if tt.setupFunc != nil {
				tt.setupFunc(t, configPath)
			}

			m := &Manager{
				configPath: configPath,
			}

			err := m.loadConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("loadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && tt.checkFunc != nil {
				tt.checkFunc(t, m)
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *ConfigValues
		setupFunc func(t *testing.T, configPath string)
		wantErr   bool
	}{
		{
			name: "saves config successfully",
			config: &ConfigValues{
				Validate: ValidateConfigValues{
					Timeout:  100,
					Cooldown: 10,
				},
			},
		},
		{
			name: "creates directory if missing",
			config: &ConfigValues{
				Validate: ValidateConfigValues{
					Timeout:  60,
					Cooldown: 5,
				},
			},
		},
		{
			name:   "handles permission error",
			config: &ConfigValues{},
			setupFunc: func(t *testing.T, configPath string) {
				// Create a read-only directory
				dir := filepath.Dir(configPath)
				os.MkdirAll(dir, 0755)
				os.Chmod(dir, 0444)
				// Cleanup function to restore permissions
				t.Cleanup(func() {
					os.Chmod(dir, 0755)
				})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "subdir", "config.json")

			if tt.setupFunc != nil {
				tt.setupFunc(t, configPath)
			}

			m := &Manager{
				configPath: configPath,
				config:     tt.config,
			}

			err := m.saveConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("saveConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				// Verify file was created and contains valid JSON
				data, readErr := os.ReadFile(configPath)
				if readErr != nil {
					t.Fatalf("Failed to read saved config: %v", readErr)
				}

				var savedConfig ConfigValues
				if unmarshalErr := json.Unmarshal(data, &savedConfig); unmarshalErr != nil {
					t.Fatalf("Saved config is not valid JSON: %v", unmarshalErr)
				}

				// Check indentation (should be pretty-printed)
				if !contains(string(data), "  ") {
					t.Error("Config should be pretty-printed with indentation")
				}

				// Verify file permissions
				info, _ := os.Stat(configPath)
				mode := info.Mode()
				if mode.Perm() != 0600 {
					t.Errorf("Config file permissions = %v, want 0600", mode.Perm())
				}
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	ctx := context.Background()

	expectedConfig := &ConfigValues{
		Validate: ValidateConfigValues{
			Timeout:  90,
			Cooldown: 10,
		},
	}

	tmpDir := t.TempDir()
	m := &Manager{
		configPath: filepath.Join(tmpDir, "config.json"),
		config:     expectedConfig,
	}

	config, err := m.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if config != expectedConfig {
		t.Error("GetConfig() should return the same config instance")
	}

	// Test lazy loading
	m2 := &Manager{
		configPath: filepath.Join(tmpDir, "config2.json"),
		config:     nil,
	}

	// Create config file
	data, _ := json.MarshalIndent(expectedConfig, "", "  ")
	os.WriteFile(m2.configPath, data, 0600)

	config2, err := m2.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() with lazy load error = %v", err)
	}

	if config2 == nil {
		t.Fatal("GetConfig() should have loaded config")
	}

	if config2.Validate.Timeout != expectedConfig.Validate.Timeout {
		t.Errorf("Lazy loaded timeout = %d, want %d", config2.Validate.Timeout, expectedConfig.Validate.Timeout)
	}
}

func TestManagerGetConfigPath(t *testing.T) {
	expectedPath := "/custom/path/config.json"
	m := &Manager{
		configPath: expectedPath,
	}

	path := m.GetConfigPath()
	if path != expectedPath {
		t.Errorf("GetConfigPath() = %s, want %s", path, expectedPath)
	}
}

func TestEnsureDefaults(t *testing.T) {
	tests := []struct {
		name  string
		input *ConfigValues
		check func(t *testing.T, config *ConfigValues)
	}{
		{
			name: "fills in zero values with defaults",
			input: &ConfigValues{
				Validate: ValidateConfigValues{},
			},
			check: func(t *testing.T, config *ConfigValues) {
				if config.Validate.Timeout != defaultValidateTimeout {
					t.Errorf("timeout = %d, want %d", config.Validate.Timeout, defaultValidateTimeout)
				}
				if config.Validate.Cooldown != defaultValidateCooldown {
					t.Errorf("cooldown = %d, want %d", config.Validate.Cooldown, defaultValidateCooldown)
				}
			},
		},
		{
			name: "preserves non-zero values",
			input: &ConfigValues{
				Validate: ValidateConfigValues{
					Timeout:  100,
					Cooldown: 10,
				},
			},
			check: func(t *testing.T, config *ConfigValues) {
				if config.Validate.Timeout != 100 {
					t.Errorf("timeout = %d, want 100", config.Validate.Timeout)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{
				config: tt.input,
			}
			m.ensureDefaults()
			tt.check(t, m.config)
		})
	}
}

func TestConvertFromMap(t *testing.T) {
	tests := []struct {
		name     string
		mapInput map[string]any
		check    func(t *testing.T, config *ConfigValues)
	}{
		{
			name: "converts all fields",
			mapInput: map[string]any{
				"validate": map[string]any{
					"timeout":  120.0,
					"cooldown": 10.0,
				},
			},
			check: func(t *testing.T, config *ConfigValues) {
				if config.Validate.Timeout != 120 {
					t.Errorf("timeout = %d, want 120", config.Validate.Timeout)
				}
				if config.Validate.Cooldown != 10 {
					t.Errorf("cooldown = %d, want 10", config.Validate.Cooldown)
				}
			},
		},
		{
			name: "handles missing sections",
			mapInput: map[string]any{
				"validate": map[string]any{
					"timeout": 90.0,
				},
			},
			check: func(t *testing.T, config *ConfigValues) {
				if config.Validate.Timeout != 90 {
					t.Errorf("timeout = %d, want 90", config.Validate.Timeout)
				}
				// Should have defaults for missing values
				if config.Validate.Cooldown != defaultValidateCooldown {
					t.Errorf("cooldown = %d, want default %d", config.Validate.Cooldown, defaultValidateCooldown)
				}
			},
		},
		{
			name:     "handles empty map",
			mapInput: map[string]any{},
			check: func(t *testing.T, config *ConfigValues) {
				defaults := getDefaultConfig()
				if config.Validate.Timeout != defaults.Validate.Timeout {
					t.Errorf("should have default timeout")
				}
			},
		},
		{
			name: "handles wrong types gracefully",
			mapInput: map[string]any{
				"validate": "not-a-map",
			},
			check: func(t *testing.T, config *ConfigValues) {
				// Should use defaults when types are wrong
				defaults := getDefaultConfig()
				if config.Validate.Timeout != defaults.Validate.Timeout {
					t.Errorf("should have default timeout when validate is wrong type")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{}
			m.convertFromMap(tt.mapInput)
			tt.check(t, m.config)
		})
	}
}

func TestGetDefaultValue(t *testing.T) {
	defaults := getDefaultConfig()

	tests := []struct {
		key  string
		want string
	}{
		{keyValidateTimeout, "60"},
		{keyValidateCooldown, "5"},
		{"unknown.key", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := getDefaultValue(defaults, tt.key)
			if got != tt.want {
				t.Errorf("getDefaultValue(%s) = %s, want %s", tt.key, got, tt.want)
			}
		})
	}
}

func TestConfigFilePath(t *testing.T) {
	// Save original env
	origHome := os.Getenv("XDG_CONFIG_HOME")
	origUserHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", origHome)
		os.Setenv("HOME", origUserHome)
	}()

	tests := []struct {
		name         string
		xdgHome      string
		homeDir      string
		wantContains string
	}{
		{
			name:         "uses XDG_CONFIG_HOME",
			xdgHome:      "/custom/xdg",
			wantContains: "/custom/xdg/cc-tools/config.json",
		},
		{
			name:         "falls back to HOME/.config",
			xdgHome:      "",
			homeDir:      "/home/user",
			wantContains: "/.config/cc-tools/config.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("XDG_CONFIG_HOME", tt.xdgHome)
			if tt.homeDir != "" {
				os.Setenv("HOME", tt.homeDir)
			}

			path := getConfigFilePath()
			if !contains(path, tt.wantContains) {
				t.Errorf("getConfigFilePath() = %s, want to contain %s", path, tt.wantContains)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr ||
		(len(substr) > 0 && len(s) > 0 && s == substr) ||
		(len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
