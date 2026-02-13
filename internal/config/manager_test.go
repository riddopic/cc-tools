package config_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/riddopic/cc-tools/internal/config"
)

// assertConfigSavedToFile reads the config file and unmarshals it, failing the
// test if the file cannot be read or parsed.
func assertConfigSavedToFile(t *testing.T, configPath string) config.Values {
	t.Helper()

	data, readErr := os.ReadFile(configPath)
	if readErr != nil {
		t.Fatalf("Failed to read config file: %v", readErr)
	}

	var saved config.Values
	if unmarshalErr := json.Unmarshal(data, &saved); unmarshalErr != nil {
		t.Fatalf("Failed to unmarshal saved config: %v", unmarshalErr)
	}

	return saved
}

// assertSavedConfigValid reads the config file and verifies it is valid JSON,
// pretty-printed, and has correct permissions.
func assertSavedConfigValid(t *testing.T, configPath string) {
	t.Helper()

	data, readErr := os.ReadFile(configPath)
	if readErr != nil {
		t.Fatalf("Failed to read saved config: %v", readErr)
	}

	var saved config.Values
	if unmarshalErr := json.Unmarshal(data, &saved); unmarshalErr != nil {
		t.Fatalf("Saved config is not valid JSON: %v", unmarshalErr)
	}

	if !strings.Contains(string(data), "  ") {
		t.Error("Config should be pretty-printed with indentation")
	}

	info, statErr := os.Stat(configPath)
	if statErr != nil {
		t.Fatalf("Failed to stat config file: %v", statErr)
	}

	if info.Mode().Perm() != 0o600 {
		t.Errorf("Config file permissions = %v, want 0600", info.Mode().Perm())
	}
}

// writeTestConfig marshals the given value to JSON and writes it to configPath,
// creating parent directories as needed.
func writeTestConfig(t *testing.T, configPath string, v any) {
	t.Helper()

	data, marshalErr := json.MarshalIndent(v, "", "  ")
	if marshalErr != nil {
		t.Fatalf("Failed to marshal test config: %v", marshalErr)
	}

	if mkdirErr := os.MkdirAll(filepath.Dir(configPath), 0o755); mkdirErr != nil {
		t.Fatalf("Failed to create config directory: %v", mkdirErr)
	}

	if writeErr := os.WriteFile(configPath, data, 0o600); writeErr != nil {
		t.Fatalf("Failed to write test config: %v", writeErr)
	}
}

// newTestValues creates a fully populated Values for testing.
func newTestValues(timeout, cooldown int) *config.Values {
	return &config.Values{
		Validate: config.ValidateValues{
			Timeout:  timeout,
			Cooldown: cooldown,
		},
		Notifications: config.NotificationsValues{
			NtfyTopic: "",
		},
	}
}

func TestNewManager(t *testing.T) {
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
			t.Setenv("XDG_CONFIG_HOME", tt.xdgHome)
			m := config.NewManager()
			path := m.GetConfigPath()

			if !filepath.IsAbs(path) {
				t.Errorf("config path should be absolute, got %s", path)
			}

			if tt.xdgHome != "" && path != tt.wantPathEnd {
				t.Errorf("unexpected config path = %s, want %s", path, tt.wantPathEnd)
			}
		})
	}
}

func TestEnsureConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("creates config file when missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")
		os.MkdirAll(filepath.Dir(configPath), 0o755)
		m := config.NewTestManager(configPath, nil)

		if err := m.EnsureConfig(ctx); err != nil {
			t.Fatalf("EnsureConfig() error = %v", err)
		}

		if _, err := os.Stat(m.GetConfigPath()); os.IsNotExist(err) {
			t.Error("config file should have been created")
		}

		cfg, err := m.GetConfig(ctx)
		if err != nil {
			t.Fatalf("failed to get config: %v", err)
		}

		if cfg.Validate.Timeout != config.ExportDefaultValidateTimeout() {
			t.Errorf("timeout = %d, want %d", cfg.Validate.Timeout, config.ExportDefaultValidateTimeout())
		}
	})

	t.Run("loads existing config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")
		writeTestConfig(t, configPath, newTestValues(120, 10))
		m := config.NewTestManager(configPath, nil)

		if err := m.EnsureConfig(ctx); err != nil {
			t.Fatalf("EnsureConfig() error = %v", err)
		}

		cfg, err := m.GetConfig(ctx)
		if err != nil {
			t.Fatalf("failed to get config: %v", err)
		}

		if cfg.Validate.Timeout != 120 {
			t.Errorf("timeout = %d, want 120", cfg.Validate.Timeout)
		}
	})

	t.Run("handles corrupt config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")
		os.MkdirAll(filepath.Dir(configPath), 0o755)
		os.WriteFile(configPath, []byte("invalid json"), 0o600)
		m := config.NewTestManager(configPath, nil)

		if err := m.EnsureConfig(ctx); err == nil {
			t.Error("EnsureConfig() expected error for corrupt config")
		}
	})
}

func TestGetInt(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		config    *config.Values
		key       string
		wantValue int
		wantFound bool
		wantErr   bool
	}{
		{
			name:      "get validate timeout",
			config:    newTestValues(90, 0),
			key:       config.ExportKeyValidateTimeout(),
			wantValue: 90,
			wantFound: true,
			wantErr:   false,
		},
		{
			name:      "get validate cooldown",
			config:    newTestValues(0, 15),
			key:       config.ExportKeyValidateCooldown(),
			wantValue: 15,
			wantFound: true,
			wantErr:   false,
		},
		{
			name:      "unknown key returns not found",
			config:    newTestValues(0, 0),
			key:       "unknown.key",
			wantValue: 0,
			wantFound: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			m := config.NewTestManager(filepath.Join(tmpDir, "config.json"), tt.config)

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
		config    *config.Values
		key       string
		wantValue string
		wantFound bool
	}{
		{
			name:      "unknown key returns not found",
			config:    newTestValues(0, 0),
			key:       "unknown.key",
			wantValue: "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			m := config.NewTestManager(filepath.Join(tmpDir, "config.json"), tt.config)

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
		config    *config.Values
		key       string
		wantValue string
		wantFound bool
	}{
		{
			name:      "get int value as string",
			config:    newTestValues(45, 0),
			key:       config.ExportKeyValidateTimeout(),
			wantValue: "45",
			wantFound: true,
		},
		{
			name:      "unknown key",
			config:    newTestValues(0, 0),
			key:       "unknown",
			wantValue: "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			m := config.NewTestManager(filepath.Join(tmpDir, "config.json"), tt.config)

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

	t.Run("set validate timeout", func(t *testing.T) {
		tmpDir := t.TempDir()
		m := config.NewTestManager(
			filepath.Join(tmpDir, "config.json"),
			config.ExportGetDefaultConfig(),
		)

		if err := m.Set(ctx, config.ExportKeyValidateTimeout(), "180"); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		assertConfigSavedToFile(t, config.ManagerConfigPath(m))

		cfg, err := m.GetConfig(ctx)
		if err != nil {
			t.Fatalf("GetConfig() error = %v", err)
		}

		if cfg.Validate.Timeout != 180 {
			t.Errorf("timeout = %d, want 180", cfg.Validate.Timeout)
		}
	})

	t.Run("invalid int value", func(t *testing.T) {
		tmpDir := t.TempDir()
		m := config.NewTestManager(
			filepath.Join(tmpDir, "config.json"),
			config.ExportGetDefaultConfig(),
		)

		if err := m.Set(ctx, config.ExportKeyValidateTimeout(), "not-a-number"); err == nil {
			t.Error("Set() expected error for non-integer value")
		}
	})

	t.Run("unknown key", func(t *testing.T) {
		tmpDir := t.TempDir()
		m := config.NewTestManager(
			filepath.Join(tmpDir, "config.json"),
			config.ExportGetDefaultConfig(),
		)

		if err := m.Set(ctx, "unknown.key", "value"); err == nil {
			t.Error("Set() expected error for unknown key")
		}
	})
}

func TestGetAll(t *testing.T) {
	ctx := context.Background()

	cfg := newTestValues(config.ExportDefaultValidateTimeout(), 10)
	tmpDir := t.TempDir()
	m := config.NewTestManager(filepath.Join(tmpDir, "config.json"), cfg)

	all, err := m.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	// Check that all keys are present
	expectedKeys := []string{
		config.ExportKeyValidateTimeout(),
		config.ExportKeyValidateCooldown(),
		config.ExportKeyNotificationsNtfyTopic(),
	}

	for _, key := range expectedKeys {
		if _, ok := all[key]; !ok {
			t.Errorf("GetAll() missing key %s", key)
		}
	}

	// Check IsDefault flags
	if !all[config.ExportKeyValidateTimeout()].IsDefault {
		t.Error("validate.timeout should be marked as default")
	}
	if all[config.ExportKeyValidateCooldown()].IsDefault {
		t.Error("validate.cooldown should not be marked as default")
	}
}

func TestGetAllKeys(t *testing.T) {
	ctx := context.Background()

	m := config.NewManager()
	keys, err := m.GetAllKeys(ctx)
	if err != nil {
		t.Fatalf("GetAllKeys() error = %v", err)
	}

	expectedKeys := []string{
		"notifications.ntfy_topic",
		"validate.cooldown",
		"validate.timeout",
	}

	if len(keys) != len(expectedKeys) {
		t.Errorf("GetAllKeys() returned %d keys, want %d", len(keys), len(expectedKeys))
	}

	for i, key := range keys {
		if i < len(expectedKeys) && key != expectedKeys[i] {
			t.Errorf("GetAllKeys()[%d] = %s, want %s", i, key, expectedKeys[i])
		}
	}
}

func TestReset(t *testing.T) {
	ctx := context.Background()

	t.Run("reset validate timeout", func(t *testing.T) {
		tmpDir := t.TempDir()
		m := config.NewTestManager(
			filepath.Join(tmpDir, "config.json"),
			config.ExportGetDefaultConfig(),
		)

		if err := m.Set(ctx, config.ExportKeyValidateTimeout(), "999"); err != nil {
			t.Fatalf("Set() setup error = %v", err)
		}

		if err := m.Reset(ctx, config.ExportKeyValidateTimeout()); err != nil {
			t.Fatalf("Reset() error = %v", err)
		}

		cfg, err := m.GetConfig(ctx)
		if err != nil {
			t.Fatalf("GetConfig() error = %v", err)
		}

		if cfg.Validate.Timeout != config.ExportDefaultValidateTimeout() {
			t.Errorf("timeout = %d, want %d", cfg.Validate.Timeout, config.ExportDefaultValidateTimeout())
		}

		assertConfigSavedToFile(t, config.ManagerConfigPath(m))
	})

	t.Run("unknown key", func(t *testing.T) {
		tmpDir := t.TempDir()
		m := config.NewTestManager(
			filepath.Join(tmpDir, "config.json"),
			config.ExportGetDefaultConfig(),
		)

		if err := m.Reset(ctx, "unknown.key"); err == nil {
			t.Error("Reset() expected error for unknown key")
		}
	})
}

func TestResetAll(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	m := config.NewTestManager(filepath.Join(tmpDir, "config.json"), newTestValues(999, 999))

	err := m.ResetAll(ctx)
	if err != nil {
		t.Fatalf("ResetAll() error = %v", err)
	}

	defaults := config.ExportGetDefaultConfig()

	cfg, getErr := m.GetConfig(ctx)
	if getErr != nil {
		t.Fatalf("GetConfig() error = %v", getErr)
	}

	if cfg.Validate.Timeout != defaults.Validate.Timeout {
		t.Errorf("timeout not reset to default")
	}
	if cfg.Validate.Cooldown != defaults.Validate.Cooldown {
		t.Errorf("cooldown not reset to default")
	}

	saved := assertConfigSavedToFile(t, config.ManagerConfigPath(m))
	if saved.Validate.Timeout != defaults.Validate.Timeout {
		t.Error("saved config not reset to defaults")
	}
}

func TestLoadConfig(t *testing.T) {
	t.Run("loads structured config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")
		writeTestConfig(t, configPath, newTestValues(120, 10))
		m := config.NewTestManager(configPath, nil)

		if err := config.ManagerLoadConfig(m); err != nil {
			t.Fatalf("loadConfig() error = %v", err)
		}

		cfg := config.ManagerConfig(m)
		if cfg.Validate.Timeout != 120 {
			t.Errorf("timeout = %d, want 120", cfg.Validate.Timeout)
		}
	})

	t.Run("loads map-based config for backward compatibility", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")
		writeTestConfig(t, configPath, map[string]any{
			"validate": map[string]any{
				"timeout":  90.0,
				"cooldown": 5.0,
			},
		})
		m := config.NewTestManager(configPath, nil)

		if err := config.ManagerLoadConfig(m); err != nil {
			t.Fatalf("loadConfig() error = %v", err)
		}

		cfg := config.ManagerConfig(m)
		if cfg.Validate.Timeout != 90 {
			t.Errorf("timeout = %d, want 90", cfg.Validate.Timeout)
		}
	})

	t.Run("uses defaults when file doesn't exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")
		m := config.NewTestManager(configPath, nil)

		if err := config.ManagerLoadConfig(m); err != nil {
			t.Fatalf("loadConfig() error = %v", err)
		}

		defaults := config.ExportGetDefaultConfig()
		cfg := config.ManagerConfig(m)
		if cfg.Validate.Timeout != defaults.Validate.Timeout {
			t.Errorf("timeout = %d, want %d", cfg.Validate.Timeout, defaults.Validate.Timeout)
		}
	})

	t.Run("fills in missing fields with defaults", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")
		writeTestConfig(t, configPath, newTestValues(100, 0))
		m := config.NewTestManager(configPath, nil)

		if err := config.ManagerLoadConfig(m); err != nil {
			t.Fatalf("loadConfig() error = %v", err)
		}

		cfg := config.ManagerConfig(m)
		if cfg.Validate.Timeout != 100 {
			t.Errorf("timeout = %d, want 100", cfg.Validate.Timeout)
		}
		if cfg.Validate.Cooldown != config.ExportDefaultValidateCooldown() {
			t.Errorf("cooldown = %d, want default %d", cfg.Validate.Cooldown, config.ExportDefaultValidateCooldown())
		}
	})

	t.Run("handles corrupt JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")
		os.MkdirAll(filepath.Dir(configPath), 0o755)
		os.WriteFile(configPath, []byte("{invalid json}"), 0o600)
		m := config.NewTestManager(configPath, nil)

		if err := config.ManagerLoadConfig(m); err == nil {
			t.Error("loadConfig() expected error for corrupt JSON")
		}
	})
}

func TestSaveConfig(t *testing.T) {
	t.Run("saves config successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "subdir", "config.json")
		m := config.NewTestManager(configPath, newTestValues(100, 10))

		if err := config.ManagerSaveConfig(m); err != nil {
			t.Fatalf("saveConfig() error = %v", err)
		}

		assertSavedConfigValid(t, configPath)
	})

	t.Run("creates directory if missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "subdir", "config.json")
		m := config.NewTestManager(configPath, newTestValues(60, 5))

		if err := config.ManagerSaveConfig(m); err != nil {
			t.Fatalf("saveConfig() error = %v", err)
		}

		assertSavedConfigValid(t, configPath)
	})

	t.Run("handles permission error", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "subdir", "config.json")
		dir := filepath.Dir(configPath)
		os.MkdirAll(dir, 0o755)
		os.Chmod(dir, 0o444)
		t.Cleanup(func() {
			os.Chmod(dir, 0o755)
		})

		m := config.NewTestManager(configPath, newTestValues(0, 0))

		if err := config.ManagerSaveConfig(m); err == nil {
			t.Error("saveConfig() expected error for permission denied")
		}
	})
}

func TestGetConfig(t *testing.T) {
	ctx := context.Background()

	expectedConfig := newTestValues(90, 10)

	tmpDir := t.TempDir()
	m := config.NewTestManager(filepath.Join(tmpDir, "config.json"), expectedConfig)

	cfg, err := m.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if cfg != expectedConfig {
		t.Error("GetConfig() should return the same config instance")
	}

	// Test lazy loading
	m2 := config.NewTestManager(filepath.Join(tmpDir, "config2.json"), nil)
	writeTestConfig(t, config.ManagerConfigPath(m2), expectedConfig)

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
	m := config.NewTestManager(expectedPath, nil)

	path := m.GetConfigPath()
	if path != expectedPath {
		t.Errorf("GetConfigPath() = %s, want %s", path, expectedPath)
	}
}

func TestEnsureDefaults(t *testing.T) {
	t.Run("fills in zero values with defaults", func(t *testing.T) {
		m := config.NewTestManager("", newTestValues(0, 0))
		config.ManagerEnsureDefaults(m)
		cfg := config.ManagerConfig(m)

		if cfg.Validate.Timeout != config.ExportDefaultValidateTimeout() {
			t.Errorf("timeout = %d, want %d", cfg.Validate.Timeout, config.ExportDefaultValidateTimeout())
		}
		if cfg.Validate.Cooldown != config.ExportDefaultValidateCooldown() {
			t.Errorf("cooldown = %d, want %d", cfg.Validate.Cooldown, config.ExportDefaultValidateCooldown())
		}
	})

	t.Run("preserves non-zero values", func(t *testing.T) {
		m := config.NewTestManager("", newTestValues(100, 10))
		config.ManagerEnsureDefaults(m)
		cfg := config.ManagerConfig(m)

		if cfg.Validate.Timeout != 100 {
			t.Errorf("timeout = %d, want 100", cfg.Validate.Timeout)
		}
	})
}

func TestConvertFromMap(t *testing.T) {
	t.Run("converts all fields", func(t *testing.T) {
		m := config.NewTestManager("", nil)
		config.ManagerConvertFromMap(m, map[string]any{
			"validate": map[string]any{
				"timeout":  120.0,
				"cooldown": 10.0,
			},
		})
		cfg := config.ManagerConfig(m)

		if cfg.Validate.Timeout != 120 {
			t.Errorf("timeout = %d, want 120", cfg.Validate.Timeout)
		}
		if cfg.Validate.Cooldown != 10 {
			t.Errorf("cooldown = %d, want 10", cfg.Validate.Cooldown)
		}
	})

	t.Run("handles missing sections", func(t *testing.T) {
		m := config.NewTestManager("", nil)
		config.ManagerConvertFromMap(m, map[string]any{
			"validate": map[string]any{
				"timeout": 90.0,
			},
		})
		cfg := config.ManagerConfig(m)

		if cfg.Validate.Timeout != 90 {
			t.Errorf("timeout = %d, want 90", cfg.Validate.Timeout)
		}
		if cfg.Validate.Cooldown != config.ExportDefaultValidateCooldown() {
			t.Errorf("cooldown = %d, want default %d", cfg.Validate.Cooldown, config.ExportDefaultValidateCooldown())
		}
	})

	t.Run("handles empty map", func(t *testing.T) {
		m := config.NewTestManager("", nil)
		config.ManagerConvertFromMap(m, map[string]any{})
		cfg := config.ManagerConfig(m)

		defaults := config.ExportGetDefaultConfig()
		if cfg.Validate.Timeout != defaults.Validate.Timeout {
			t.Errorf("should have default timeout")
		}
	})

	t.Run("handles wrong types gracefully", func(t *testing.T) {
		m := config.NewTestManager("", nil)
		config.ManagerConvertFromMap(m, map[string]any{
			"validate": "not-a-map",
		})
		cfg := config.ManagerConfig(m)

		defaults := config.ExportGetDefaultConfig()
		if cfg.Validate.Timeout != defaults.Validate.Timeout {
			t.Errorf("should have default timeout when validate is wrong type")
		}
	})
}

func TestGetDefaultValue(t *testing.T) {
	defaults := config.ExportGetDefaultConfig()

	tests := []struct {
		key  string
		want string
	}{
		{config.ExportKeyValidateTimeout(), "60"},
		{config.ExportKeyValidateCooldown(), "5"},
		{"unknown.key", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := config.ExportGetDefaultValue(defaults, tt.key)
			if got != tt.want {
				t.Errorf("getDefaultValue(%s) = %s, want %s", tt.key, got, tt.want)
			}
		})
	}
}

func TestConfigFilePath(t *testing.T) {
	tests := []struct {
		name         string
		xdgHome      string
		homeDir      string
		wantContains string
	}{
		{
			name:         "uses XDG_CONFIG_HOME",
			xdgHome:      "/custom/xdg",
			homeDir:      "",
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
			t.Setenv("XDG_CONFIG_HOME", tt.xdgHome)
			if tt.homeDir != "" {
				t.Setenv("HOME", tt.homeDir)
			}

			path := config.ExportGetConfigFilePath()
			if !strings.Contains(path, tt.wantContains) {
				t.Errorf("getConfigFilePath() = %s, want to contain %s", path, tt.wantContains)
			}
		})
	}
}
