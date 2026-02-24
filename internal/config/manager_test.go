package config_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		Compact: config.CompactValues{
			Threshold:        config.ExportDefaultCompactThreshold(),
			ReminderInterval: config.ExportDefaultCompactReminderInterval(),
		},
		Notify: config.NotifyValues{
			QuietHours: config.QuietHoursValues{
				Enabled: config.ExportDefaultNotifyQuietHoursEnabled(),
				Start:   config.ExportDefaultNotifyQuietHoursStart(),
				End:     config.ExportDefaultNotifyQuietHoursEnd(),
			},
			Audio: config.AudioValues{
				Enabled:   config.ExportDefaultNotifyAudioEnabled(),
				Directory: config.ExportDefaultNotifyAudioDirectory(),
			},
			Desktop: config.DesktopValues{
				Enabled: config.ExportDefaultNotifyDesktopEnabled(),
			},
		},
		Observe: config.ObserveValues{
			Enabled:       config.ExportDefaultObserveEnabled(),
			MaxFileSizeMB: config.ExportDefaultObserveMaxFileSizeMB(),
		},
		Learning: config.LearningValues{
			MinSessionLength:  config.ExportDefaultLearningMinSessionLength(),
			LearnedSkillsPath: config.ExportDefaultLearningLearnedSkillsPath(),
		},
		PreCommit: config.PreCommitValues{
			Enabled: config.ExportDefaultPreCommitEnabled(),
			Command: config.ExportDefaultPreCommitCommand(),
		},
		PackageManager: config.PackageManagerValues{
			Preferred: config.ExportDefaultPackageManagerPreferred(),
		},
		Drift: config.DriftValues{
			Enabled:   config.ExportDefaultDriftEnabled(),
			MinEdits:  config.ExportDefaultDriftMinEdits(),
			Threshold: config.ExportDefaultDriftThreshold(),
		},
		StopReminder: config.StopReminderValues{
			Enabled:  config.ExportDefaultStopReminderEnabled(),
			Interval: config.ExportDefaultStopReminderInterval(),
			WarnAt:   config.ExportDefaultStopReminderWarnAt(),
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

	expectedKeys := config.ExportAllKeys()
	sort.Strings(expectedKeys)

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

func TestManager_LoadsHookConfig(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantComp int
		wantQH   bool
		wantQHS  string
		wantQHE  string
	}{
		{
			name:     "defaults when empty",
			json:     `{}`,
			wantComp: 50,
			wantQH:   true,
			wantQHS:  "21:00",
			wantQHE:  "07:30",
		},
		{
			name:     "custom values",
			json:     `{"compact":{"threshold":100,"reminder_interval":50},"notify":{"quiet_hours":{"enabled":false,"start":"22:00","end":"08:00"}}}`,
			wantComp: 100,
			wantQH:   false,
			wantQHS:  "22:00",
			wantQHE:  "08:00",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfgPath := filepath.Join(tmpDir, "config.json")
			require.NoError(t, os.WriteFile(cfgPath, []byte(tt.json), 0o600))

			m := config.NewManagerWithPath(cfgPath)
			cfg, err := m.GetConfig(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.wantComp, cfg.Compact.Threshold)
			assert.Equal(t, tt.wantQH, cfg.Notify.QuietHours.Enabled)
			assert.Equal(t, tt.wantQHS, cfg.Notify.QuietHours.Start)
			assert.Equal(t, tt.wantQHE, cfg.Notify.QuietHours.End)
		})
	}
}

func TestGetString_AllKeys(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		config    *config.Values
		key       string
		wantValue string
		wantFound bool
	}{
		{
			name:      "get notifications ntfy topic (empty default)",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyNotificationsNtfyTopic(),
			wantValue: "",
			wantFound: true,
		},
		{
			name: "get notifications ntfy topic (custom)",
			config: func() *config.Values {
				v := newTestValues(0, 0)
				v.Notifications.NtfyTopic = "my-topic"
				return v
			}(),
			key:       config.ExportKeyNotificationsNtfyTopic(),
			wantValue: "my-topic",
			wantFound: true,
		},
		{
			name:      "get notify quiet hours start",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyNotifyQuietHoursStart(),
			wantValue: config.ExportDefaultNotifyQuietHoursStart(),
			wantFound: true,
		},
		{
			name:      "get notify quiet hours end",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyNotifyQuietHoursEnd(),
			wantValue: config.ExportDefaultNotifyQuietHoursEnd(),
			wantFound: true,
		},
		{
			name:      "get notify audio directory",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyNotifyAudioDirectory(),
			wantValue: config.ExportDefaultNotifyAudioDirectory(),
			wantFound: true,
		},
		{
			name:      "get learning learned skills path",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyLearningLearnedSkillsPath(),
			wantValue: config.ExportDefaultLearningLearnedSkillsPath(),
			wantFound: true,
		},
		{
			name:      "get pre-commit command",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyPreCommitCommand(),
			wantValue: config.ExportDefaultPreCommitCommand(),
			wantFound: true,
		},
		{
			name:      "get package manager preferred (empty default)",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyPackageManagerPreferred(),
			wantValue: "",
			wantFound: true,
		},
		{
			name: "get package manager preferred (custom)",
			config: func() *config.Values {
				v := newTestValues(0, 0)
				v.PackageManager.Preferred = "bun"
				return v
			}(),
			key:       config.ExportKeyPackageManagerPreferred(),
			wantValue: "bun",
			wantFound: true,
		},
		{
			name: "get custom quiet hours start",
			config: func() *config.Values {
				v := newTestValues(0, 0)
				v.Notify.QuietHours.Start = "23:00"
				return v
			}(),
			key:       config.ExportKeyNotifyQuietHoursStart(),
			wantValue: "23:00",
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			m := config.NewTestManager(filepath.Join(tmpDir, "config.json"), tt.config)

			value, found, err := m.GetString(ctx, tt.key)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValue, value)
			assert.Equal(t, tt.wantFound, found)
		})
	}
}

func TestGetInt_AllKeys(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		config    *config.Values
		key       string
		wantValue int
		wantFound bool
	}{
		{
			name:      "get compact threshold default",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyCompactThreshold(),
			wantValue: config.ExportDefaultCompactThreshold(),
			wantFound: true,
		},
		{
			name: "get compact threshold custom",
			config: func() *config.Values {
				v := newTestValues(0, 0)
				v.Compact.Threshold = 100
				return v
			}(),
			key:       config.ExportKeyCompactThreshold(),
			wantValue: 100,
			wantFound: true,
		},
		{
			name:      "get compact reminder interval default",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyCompactReminderInterval(),
			wantValue: config.ExportDefaultCompactReminderInterval(),
			wantFound: true,
		},
		{
			name:      "get observe max file size mb default",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyObserveMaxFileSizeMB(),
			wantValue: config.ExportDefaultObserveMaxFileSizeMB(),
			wantFound: true,
		},
		{
			name: "get observe max file size mb custom",
			config: func() *config.Values {
				v := newTestValues(0, 0)
				v.Observe.MaxFileSizeMB = 25
				return v
			}(),
			key:       config.ExportKeyObserveMaxFileSizeMB(),
			wantValue: 25,
			wantFound: true,
		},
		{
			name:      "get learning min session length default",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyLearningMinSessionLength(),
			wantValue: config.ExportDefaultLearningMinSessionLength(),
			wantFound: true,
		},
		{
			name: "get learning min session length custom",
			config: func() *config.Values {
				v := newTestValues(0, 0)
				v.Learning.MinSessionLength = 30
				return v
			}(),
			key:       config.ExportKeyLearningMinSessionLength(),
			wantValue: 30,
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			m := config.NewTestManager(filepath.Join(tmpDir, "config.json"), tt.config)

			value, found, err := m.GetInt(ctx, tt.key)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValue, value)
			assert.Equal(t, tt.wantFound, found)
		})
	}
}

func TestSetBoolField(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
		check   func(t *testing.T, cfg *config.Values)
	}{
		{
			name:    "set quiet hours enabled to true",
			key:     config.ExportKeyNotifyQuietHoursEnabled(),
			value:   "true",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.True(t, cfg.Notify.QuietHours.Enabled)
			},
		},
		{
			name:    "set quiet hours enabled to false",
			key:     config.ExportKeyNotifyQuietHoursEnabled(),
			value:   "false",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.False(t, cfg.Notify.QuietHours.Enabled)
			},
		},
		{
			name:    "set audio enabled to 1",
			key:     config.ExportKeyNotifyAudioEnabled(),
			value:   "1",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.True(t, cfg.Notify.Audio.Enabled)
			},
		},
		{
			name:    "set audio enabled to 0",
			key:     config.ExportKeyNotifyAudioEnabled(),
			value:   "0",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.False(t, cfg.Notify.Audio.Enabled)
			},
		},
		{
			name:    "set observe enabled to false",
			key:     config.ExportKeyObserveEnabled(),
			value:   "false",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.False(t, cfg.Observe.Enabled)
			},
		},
		{
			name:    "set observe enabled to true",
			key:     config.ExportKeyObserveEnabled(),
			value:   "true",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.True(t, cfg.Observe.Enabled)
			},
		},
		{
			name:    "set pre-commit enabled to false",
			key:     config.ExportKeyPreCommitEnabled(),
			value:   "false",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.False(t, cfg.PreCommit.Enabled)
			},
		},
		{
			name:    "set pre-commit enabled to true",
			key:     config.ExportKeyPreCommitEnabled(),
			value:   "true",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.True(t, cfg.PreCommit.Enabled)
			},
		},
		{
			name:    "invalid bool value returns error",
			key:     config.ExportKeyNotifyQuietHoursEnabled(),
			value:   "invalid",
			wantErr: true,
			check:   nil,
		},
		{
			name:    "set desktop enabled to false",
			key:     config.ExportKeyNotifyDesktopEnabled(),
			value:   "false",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.False(t, cfg.Notify.Desktop.Enabled)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			m := config.NewTestManager(
				filepath.Join(tmpDir, "config.json"),
				config.ExportGetDefaultConfig(),
			)

			err := m.Set(ctx, tt.key, tt.value)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			cfg, getErr := m.GetConfig(ctx)
			require.NoError(t, getErr)
			tt.check(t, cfg)

			assertConfigSavedToFile(t, config.ManagerConfigPath(m))
		})
	}
}

func TestSetStringAndIntFields(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
		check   func(t *testing.T, cfg *config.Values)
	}{
		{
			name:    "set ntfy topic",
			key:     config.ExportKeyNotificationsNtfyTopic(),
			value:   "my-notifications",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, "my-notifications", cfg.Notifications.NtfyTopic)
			},
		},
		{
			name:    "set quiet hours start",
			key:     config.ExportKeyNotifyQuietHoursStart(),
			value:   "22:30",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, "22:30", cfg.Notify.QuietHours.Start)
			},
		},
		{
			name:    "set quiet hours end",
			key:     config.ExportKeyNotifyQuietHoursEnd(),
			value:   "08:00",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, "08:00", cfg.Notify.QuietHours.End)
			},
		},
		{
			name:    "set audio directory",
			key:     config.ExportKeyNotifyAudioDirectory(),
			value:   "/custom/audio/path",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, "/custom/audio/path", cfg.Notify.Audio.Directory)
			},
		},
		{
			name:    "set learned skills path",
			key:     config.ExportKeyLearningLearnedSkillsPath(),
			value:   "custom/skills/dir",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, "custom/skills/dir", cfg.Learning.LearnedSkillsPath)
			},
		},
		{
			name:    "set pre-commit command",
			key:     config.ExportKeyPreCommitCommand(),
			value:   "make lint && make test",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, "make lint && make test", cfg.PreCommit.Command)
			},
		},
		{
			name:    "set compact threshold",
			key:     config.ExportKeyCompactThreshold(),
			value:   "75",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, 75, cfg.Compact.Threshold)
			},
		},
		{
			name:    "set compact reminder interval",
			key:     config.ExportKeyCompactReminderInterval(),
			value:   "10",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, 10, cfg.Compact.ReminderInterval)
			},
		},
		{
			name:    "set observe max file size mb",
			key:     config.ExportKeyObserveMaxFileSizeMB(),
			value:   "50",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, 50, cfg.Observe.MaxFileSizeMB)
			},
		},
		{
			name:    "set learning min session length",
			key:     config.ExportKeyLearningMinSessionLength(),
			value:   "20",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, 20, cfg.Learning.MinSessionLength)
			},
		},
		{
			name:    "set compact threshold invalid int",
			key:     config.ExportKeyCompactThreshold(),
			value:   "abc",
			wantErr: true,
			check:   nil,
		},
		{
			name:    "set package manager preferred",
			key:     config.ExportKeyPackageManagerPreferred(),
			value:   "bun",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, "bun", cfg.PackageManager.Preferred)
			},
		},
		{
			name:    "set package manager preferred to empty (reset to auto-detect)",
			key:     config.ExportKeyPackageManagerPreferred(),
			value:   "",
			wantErr: false,
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Empty(t, cfg.PackageManager.Preferred)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			m := config.NewTestManager(
				filepath.Join(tmpDir, "config.json"),
				config.ExportGetDefaultConfig(),
			)

			err := m.Set(ctx, tt.key, tt.value)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			cfg, getErr := m.GetConfig(ctx)
			require.NoError(t, getErr)
			tt.check(t, cfg)

			assertConfigSavedToFile(t, config.ManagerConfigPath(m))
		})
	}
}

func TestReset_AllKeyTypes(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		setupKey string
		setupVal string
		resetKey string
		check    func(t *testing.T, cfg *config.Values)
	}{
		{
			name:     "reset string key ntfy topic",
			setupKey: config.ExportKeyNotificationsNtfyTopic(),
			setupVal: "custom-topic",
			resetKey: config.ExportKeyNotificationsNtfyTopic(),
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Empty(t, cfg.Notifications.NtfyTopic)
			},
		},
		{
			name:     "reset bool key quiet hours enabled",
			setupKey: config.ExportKeyNotifyQuietHoursEnabled(),
			setupVal: "false",
			resetKey: config.ExportKeyNotifyQuietHoursEnabled(),
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultNotifyQuietHoursEnabled(), cfg.Notify.QuietHours.Enabled)
			},
		},
		{
			name:     "reset int key compact threshold",
			setupKey: config.ExportKeyCompactThreshold(),
			setupVal: "999",
			resetKey: config.ExportKeyCompactThreshold(),
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultCompactThreshold(), cfg.Compact.Threshold)
			},
		},
		{
			name:     "reset compact reminder interval",
			setupKey: config.ExportKeyCompactReminderInterval(),
			setupVal: "999",
			resetKey: config.ExportKeyCompactReminderInterval(),
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultCompactReminderInterval(), cfg.Compact.ReminderInterval)
			},
		},
		{
			name:     "reset quiet hours start",
			setupKey: config.ExportKeyNotifyQuietHoursStart(),
			setupVal: "23:30",
			resetKey: config.ExportKeyNotifyQuietHoursStart(),
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultNotifyQuietHoursStart(), cfg.Notify.QuietHours.Start)
			},
		},
		{
			name:     "reset quiet hours end",
			setupKey: config.ExportKeyNotifyQuietHoursEnd(),
			setupVal: "09:00",
			resetKey: config.ExportKeyNotifyQuietHoursEnd(),
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultNotifyQuietHoursEnd(), cfg.Notify.QuietHours.End)
			},
		},
		{
			name:     "reset audio enabled",
			setupKey: config.ExportKeyNotifyAudioEnabled(),
			setupVal: "false",
			resetKey: config.ExportKeyNotifyAudioEnabled(),
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultNotifyAudioEnabled(), cfg.Notify.Audio.Enabled)
			},
		},
		{
			name:     "reset audio directory",
			setupKey: config.ExportKeyNotifyAudioDirectory(),
			setupVal: "/tmp/audio",
			resetKey: config.ExportKeyNotifyAudioDirectory(),
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultNotifyAudioDirectory(), cfg.Notify.Audio.Directory)
			},
		},
		{
			name:     "reset desktop enabled",
			setupKey: config.ExportKeyNotifyDesktopEnabled(),
			setupVal: "false",
			resetKey: config.ExportKeyNotifyDesktopEnabled(),
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultNotifyDesktopEnabled(), cfg.Notify.Desktop.Enabled)
			},
		},
		{
			name:     "reset observe enabled",
			setupKey: config.ExportKeyObserveEnabled(),
			setupVal: "false",
			resetKey: config.ExportKeyObserveEnabled(),
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultObserveEnabled(), cfg.Observe.Enabled)
			},
		},
		{
			name:     "reset observe max file size mb",
			setupKey: config.ExportKeyObserveMaxFileSizeMB(),
			setupVal: "99",
			resetKey: config.ExportKeyObserveMaxFileSizeMB(),
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultObserveMaxFileSizeMB(), cfg.Observe.MaxFileSizeMB)
			},
		},
		{
			name:     "reset learning min session length",
			setupKey: config.ExportKeyLearningMinSessionLength(),
			setupVal: "99",
			resetKey: config.ExportKeyLearningMinSessionLength(),
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultLearningMinSessionLength(), cfg.Learning.MinSessionLength)
			},
		},
		{
			name:     "reset learning learned skills path",
			setupKey: config.ExportKeyLearningLearnedSkillsPath(),
			setupVal: "/custom/path",
			resetKey: config.ExportKeyLearningLearnedSkillsPath(),
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultLearningLearnedSkillsPath(), cfg.Learning.LearnedSkillsPath)
			},
		},
		{
			name:     "reset pre-commit enabled",
			setupKey: config.ExportKeyPreCommitEnabled(),
			setupVal: "false",
			resetKey: config.ExportKeyPreCommitEnabled(),
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultPreCommitEnabled(), cfg.PreCommit.Enabled)
			},
		},
		{
			name:     "reset pre-commit command",
			setupKey: config.ExportKeyPreCommitCommand(),
			setupVal: "make check",
			resetKey: config.ExportKeyPreCommitCommand(),
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultPreCommitCommand(), cfg.PreCommit.Command)
			},
		},
		{
			name:     "reset package manager preferred",
			setupKey: config.ExportKeyPackageManagerPreferred(),
			setupVal: "bun",
			resetKey: config.ExportKeyPackageManagerPreferred(),
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultPackageManagerPreferred(), cfg.PackageManager.Preferred)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			m := config.NewTestManager(
				filepath.Join(tmpDir, "config.json"),
				config.ExportGetDefaultConfig(),
			)

			err := m.Set(ctx, tt.setupKey, tt.setupVal)
			require.NoError(t, err, "Set() setup failed")

			err = m.Reset(ctx, tt.resetKey)
			require.NoError(t, err, "Reset() failed")

			cfg, getErr := m.GetConfig(ctx)
			require.NoError(t, getErr)
			tt.check(t, cfg)

			assertConfigSavedToFile(t, config.ManagerConfigPath(m))
		})
	}
}

func TestConvertNotifyFromMap(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]any
		check func(t *testing.T, cfg *config.Values)
	}{
		{
			name: "full map with all fields",
			input: map[string]any{
				"notify": map[string]any{
					"quiet_hours": map[string]any{
						"enabled": false,
						"start":   "23:00",
						"end":     "06:00",
					},
					"audio": map[string]any{
						"enabled":   false,
						"directory": "/custom/audio",
					},
					"desktop": map[string]any{
						"enabled": false,
					},
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.False(t, cfg.Notify.QuietHours.Enabled)
				assert.Equal(t, "23:00", cfg.Notify.QuietHours.Start)
				assert.Equal(t, "06:00", cfg.Notify.QuietHours.End)
				assert.False(t, cfg.Notify.Audio.Enabled)
				assert.Equal(t, "/custom/audio", cfg.Notify.Audio.Directory)
				assert.False(t, cfg.Notify.Desktop.Enabled)
			},
		},
		{
			name: "partial map with only quiet hours",
			input: map[string]any{
				"notify": map[string]any{
					"quiet_hours": map[string]any{
						"start": "22:00",
					},
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, "22:00", cfg.Notify.QuietHours.Start)
				// Other notify defaults should be preserved
				assert.Equal(t, config.ExportDefaultNotifyAudioEnabled(), cfg.Notify.Audio.Enabled)
				assert.Equal(t, config.ExportDefaultNotifyAudioDirectory(), cfg.Notify.Audio.Directory)
			},
		},
		{
			name:  "empty notify map preserves defaults",
			input: map[string]any{},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultNotifyQuietHoursEnabled(), cfg.Notify.QuietHours.Enabled)
				assert.Equal(t, config.ExportDefaultNotifyQuietHoursStart(), cfg.Notify.QuietHours.Start)
				assert.Equal(t, config.ExportDefaultNotifyQuietHoursEnd(), cfg.Notify.QuietHours.End)
				assert.Equal(t, config.ExportDefaultNotifyAudioEnabled(), cfg.Notify.Audio.Enabled)
				assert.Equal(t, config.ExportDefaultNotifyAudioDirectory(), cfg.Notify.Audio.Directory)
				assert.Equal(t, config.ExportDefaultNotifyDesktopEnabled(), cfg.Notify.Desktop.Enabled)
			},
		},
		{
			name: "wrong types in notify map preserves defaults",
			input: map[string]any{
				"notify": map[string]any{
					"quiet_hours": map[string]any{
						"enabled": "not-a-bool",
						"start":   123,
						"end":     true,
					},
					"audio": map[string]any{
						"enabled":   "not-a-bool",
						"directory": 456,
					},
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				// Wrong types should be ignored; defaults should be preserved
				assert.Equal(t, config.ExportDefaultNotifyQuietHoursEnabled(), cfg.Notify.QuietHours.Enabled)
				assert.Equal(t, config.ExportDefaultNotifyQuietHoursStart(), cfg.Notify.QuietHours.Start)
				assert.Equal(t, config.ExportDefaultNotifyQuietHoursEnd(), cfg.Notify.QuietHours.End)
				assert.Equal(t, config.ExportDefaultNotifyAudioEnabled(), cfg.Notify.Audio.Enabled)
				assert.Equal(t, config.ExportDefaultNotifyAudioDirectory(), cfg.Notify.Audio.Directory)
			},
		},
		{
			name: "notify is not a map",
			input: map[string]any{
				"notify": "not-a-map",
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultNotifyQuietHoursEnabled(), cfg.Notify.QuietHours.Enabled)
				assert.Equal(t, config.ExportDefaultNotifyQuietHoursStart(), cfg.Notify.QuietHours.Start)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := config.NewTestManager("", nil)
			config.ManagerConvertFromMap(m, tt.input)
			cfg := config.ManagerConfig(m)
			tt.check(t, cfg)
		})
	}
}

func TestConvertCompactFromMap(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]any
		check func(t *testing.T, cfg *config.Values)
	}{
		{
			name: "full compact settings",
			input: map[string]any{
				"compact": map[string]any{
					"threshold":         80.0,
					"reminder_interval": 40.0,
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, 80, cfg.Compact.Threshold)
				assert.Equal(t, 40, cfg.Compact.ReminderInterval)
			},
		},
		{
			name: "partial compact with threshold only",
			input: map[string]any{
				"compact": map[string]any{
					"threshold": 100.0,
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, 100, cfg.Compact.Threshold)
				assert.Equal(t, config.ExportDefaultCompactReminderInterval(), cfg.Compact.ReminderInterval)
			},
		},
		{
			name: "compact wrong types",
			input: map[string]any{
				"compact": map[string]any{
					"threshold":         "not-a-number",
					"reminder_interval": true,
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultCompactThreshold(), cfg.Compact.Threshold)
				assert.Equal(t, config.ExportDefaultCompactReminderInterval(), cfg.Compact.ReminderInterval)
			},
		},
		{
			name: "compact section is not a map",
			input: map[string]any{
				"compact": "not-a-map",
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultCompactThreshold(), cfg.Compact.Threshold)
				assert.Equal(t, config.ExportDefaultCompactReminderInterval(), cfg.Compact.ReminderInterval)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := config.NewTestManager("", nil)
			config.ManagerConvertFromMap(m, tt.input)
			cfg := config.ManagerConfig(m)
			tt.check(t, cfg)
		})
	}
}

func TestConvertObserveFromMap(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]any
		check func(t *testing.T, cfg *config.Values)
	}{
		{
			name: "full observe settings",
			input: map[string]any{
				"observe": map[string]any{
					"enabled":          false,
					"max_file_size_mb": 25.0,
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.False(t, cfg.Observe.Enabled)
				assert.Equal(t, 25, cfg.Observe.MaxFileSizeMB)
			},
		},
		{
			name: "partial observe with enabled only",
			input: map[string]any{
				"observe": map[string]any{
					"enabled": false,
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.False(t, cfg.Observe.Enabled)
				assert.Equal(t, config.ExportDefaultObserveMaxFileSizeMB(), cfg.Observe.MaxFileSizeMB)
			},
		},
		{
			name: "observe wrong types",
			input: map[string]any{
				"observe": map[string]any{
					"enabled":          "not-a-bool",
					"max_file_size_mb": "not-a-number",
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultObserveEnabled(), cfg.Observe.Enabled)
				assert.Equal(t, config.ExportDefaultObserveMaxFileSizeMB(), cfg.Observe.MaxFileSizeMB)
			},
		},
		{
			name: "observe section is not a map",
			input: map[string]any{
				"observe": 42,
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultObserveEnabled(), cfg.Observe.Enabled)
				assert.Equal(t, config.ExportDefaultObserveMaxFileSizeMB(), cfg.Observe.MaxFileSizeMB)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := config.NewTestManager("", nil)
			config.ManagerConvertFromMap(m, tt.input)
			cfg := config.ManagerConfig(m)
			tt.check(t, cfg)
		})
	}
}

func TestConvertLearningFromMap(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]any
		check func(t *testing.T, cfg *config.Values)
	}{
		{
			name: "full learning settings",
			input: map[string]any{
				"learning": map[string]any{
					"min_session_length":  20.0,
					"learned_skills_path": "custom/path",
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, 20, cfg.Learning.MinSessionLength)
				assert.Equal(t, "custom/path", cfg.Learning.LearnedSkillsPath)
			},
		},
		{
			name: "partial learning with min session length only",
			input: map[string]any{
				"learning": map[string]any{
					"min_session_length": 30.0,
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, 30, cfg.Learning.MinSessionLength)
				assert.Equal(t, config.ExportDefaultLearningLearnedSkillsPath(), cfg.Learning.LearnedSkillsPath)
			},
		},
		{
			name: "partial learning with path only",
			input: map[string]any{
				"learning": map[string]any{
					"learned_skills_path": "another/path",
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultLearningMinSessionLength(), cfg.Learning.MinSessionLength)
				assert.Equal(t, "another/path", cfg.Learning.LearnedSkillsPath)
			},
		},
		{
			name: "learning wrong types",
			input: map[string]any{
				"learning": map[string]any{
					"min_session_length":  "not-a-number",
					"learned_skills_path": 123,
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultLearningMinSessionLength(), cfg.Learning.MinSessionLength)
				assert.Equal(t, config.ExportDefaultLearningLearnedSkillsPath(), cfg.Learning.LearnedSkillsPath)
			},
		},
		{
			name: "learning section is not a map",
			input: map[string]any{
				"learning": []string{"not", "a", "map"},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultLearningMinSessionLength(), cfg.Learning.MinSessionLength)
				assert.Equal(t, config.ExportDefaultLearningLearnedSkillsPath(), cfg.Learning.LearnedSkillsPath)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := config.NewTestManager("", nil)
			config.ManagerConvertFromMap(m, tt.input)
			cfg := config.ManagerConfig(m)
			tt.check(t, cfg)
		})
	}
}

func TestConvertPreCommitFromMap(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]any
		check func(t *testing.T, cfg *config.Values)
	}{
		{
			name: "full pre-commit settings",
			input: map[string]any{
				"pre_commit_reminder": map[string]any{
					"enabled": false,
					"command": "make check",
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.False(t, cfg.PreCommit.Enabled)
				assert.Equal(t, "make check", cfg.PreCommit.Command)
			},
		},
		{
			name: "partial pre-commit with enabled only",
			input: map[string]any{
				"pre_commit_reminder": map[string]any{
					"enabled": false,
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.False(t, cfg.PreCommit.Enabled)
				assert.Equal(t, config.ExportDefaultPreCommitCommand(), cfg.PreCommit.Command)
			},
		},
		{
			name: "partial pre-commit with command only",
			input: map[string]any{
				"pre_commit_reminder": map[string]any{
					"command": "npm run lint",
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultPreCommitEnabled(), cfg.PreCommit.Enabled)
				assert.Equal(t, "npm run lint", cfg.PreCommit.Command)
			},
		},
		{
			name: "pre-commit wrong types",
			input: map[string]any{
				"pre_commit_reminder": map[string]any{
					"enabled": "not-a-bool",
					"command": 12345,
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultPreCommitEnabled(), cfg.PreCommit.Enabled)
				assert.Equal(t, config.ExportDefaultPreCommitCommand(), cfg.PreCommit.Command)
			},
		},
		{
			name: "pre-commit section is not a map",
			input: map[string]any{
				"pre_commit_reminder": true,
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, config.ExportDefaultPreCommitEnabled(), cfg.PreCommit.Enabled)
				assert.Equal(t, config.ExportDefaultPreCommitCommand(), cfg.PreCommit.Command)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := config.NewTestManager("", nil)
			config.ManagerConvertFromMap(m, tt.input)
			cfg := config.ManagerConfig(m)
			tt.check(t, cfg)
		})
	}
}

func TestGetValue_AllKeys(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		config    *config.Values
		key       string
		wantValue string
		wantFound bool
	}{
		{
			name:      "get cooldown as string",
			config:    newTestValues(0, 15),
			key:       config.ExportKeyValidateCooldown(),
			wantValue: "15",
			wantFound: true,
		},
		{
			name:      "get ntfy topic as string",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyNotificationsNtfyTopic(),
			wantValue: "",
			wantFound: true,
		},
		{
			name:      "get compact threshold as string",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyCompactThreshold(),
			wantValue: "50",
			wantFound: true,
		},
		{
			name:      "get compact reminder interval as string",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyCompactReminderInterval(),
			wantValue: "25",
			wantFound: true,
		},
		{
			name:      "get quiet hours enabled as string",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyNotifyQuietHoursEnabled(),
			wantValue: "true",
			wantFound: true,
		},
		{
			name:      "get quiet hours start as string",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyNotifyQuietHoursStart(),
			wantValue: "21:00",
			wantFound: true,
		},
		{
			name:      "get quiet hours end as string",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyNotifyQuietHoursEnd(),
			wantValue: "07:30",
			wantFound: true,
		},
		{
			name:      "get audio enabled as string",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyNotifyAudioEnabled(),
			wantValue: "true",
			wantFound: true,
		},
		{
			name:      "get audio directory as string",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyNotifyAudioDirectory(),
			wantValue: "~/.claude/audio",
			wantFound: true,
		},
		{
			name:      "get desktop enabled as string",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyNotifyDesktopEnabled(),
			wantValue: "true",
			wantFound: true,
		},
		{
			name:      "get observe enabled as string",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyObserveEnabled(),
			wantValue: "true",
			wantFound: true,
		},
		{
			name:      "get observe max file size as string",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyObserveMaxFileSizeMB(),
			wantValue: "10",
			wantFound: true,
		},
		{
			name:      "get learning min session length as string",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyLearningMinSessionLength(),
			wantValue: "10",
			wantFound: true,
		},
		{
			name:      "get learning learned skills path as string",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyLearningLearnedSkillsPath(),
			wantValue: ".claude/skills/learned",
			wantFound: true,
		},
		{
			name:      "get pre-commit enabled as string",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyPreCommitEnabled(),
			wantValue: "true",
			wantFound: true,
		},
		{
			name:      "get pre-commit command as string",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyPreCommitCommand(),
			wantValue: "task pre-commit",
			wantFound: true,
		},
		{
			name:      "get package manager preferred as string (empty default)",
			config:    newTestValues(0, 0),
			key:       config.ExportKeyPackageManagerPreferred(),
			wantValue: "",
			wantFound: true,
		},
		{
			name: "get package manager preferred as string (custom)",
			config: func() *config.Values {
				v := newTestValues(0, 0)
				v.PackageManager.Preferred = "pnpm"
				return v
			}(),
			key:       config.ExportKeyPackageManagerPreferred(),
			wantValue: "pnpm",
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			m := config.NewTestManager(filepath.Join(tmpDir, "config.json"), tt.config)

			value, found, err := m.GetValue(ctx, tt.key)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValue, value)
			assert.Equal(t, tt.wantFound, found)
		})
	}
}

func TestGetDefaultValue_AllKeys(t *testing.T) {
	defaults := config.ExportGetDefaultConfig()

	tests := []struct {
		key  string
		want string
	}{
		{config.ExportKeyNotificationsNtfyTopic(), ""},
		{config.ExportKeyCompactThreshold(), "50"},
		{config.ExportKeyCompactReminderInterval(), "25"},
		{config.ExportKeyNotifyQuietHoursEnabled(), "true"},
		{config.ExportKeyNotifyQuietHoursStart(), "21:00"},
		{config.ExportKeyNotifyQuietHoursEnd(), "07:30"},
		{config.ExportKeyNotifyAudioEnabled(), "true"},
		{config.ExportKeyNotifyAudioDirectory(), "~/.claude/audio"},
		{config.ExportKeyNotifyDesktopEnabled(), "true"},
		{config.ExportKeyObserveEnabled(), "true"},
		{config.ExportKeyObserveMaxFileSizeMB(), "10"},
		{config.ExportKeyLearningMinSessionLength(), "10"},
		{config.ExportKeyLearningLearnedSkillsPath(), ".claude/skills/learned"},
		{config.ExportKeyPreCommitEnabled(), "true"},
		{config.ExportKeyPreCommitCommand(), "task pre-commit"},
		{config.ExportKeyPackageManagerPreferred(), ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := config.ExportGetDefaultValue(defaults, tt.key)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConvertNotificationsFromMap(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]any
		check func(t *testing.T, cfg *config.Values)
	}{
		{
			name: "full notifications settings",
			input: map[string]any{
				"notifications": map[string]any{
					"ntfy_topic": "test-topic",
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, "test-topic", cfg.Notifications.NtfyTopic)
			},
		},
		{
			name: "notifications section is not a map",
			input: map[string]any{
				"notifications": "not-a-map",
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Empty(t, cfg.Notifications.NtfyTopic)
			},
		},
		{
			name: "notifications with wrong type for ntfy_topic",
			input: map[string]any{
				"notifications": map[string]any{
					"ntfy_topic": 12345,
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Empty(t, cfg.Notifications.NtfyTopic)
			},
		},
		{
			name:  "missing notifications section",
			input: map[string]any{},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Empty(t, cfg.Notifications.NtfyTopic)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := config.NewTestManager("", nil)
			config.ManagerConvertFromMap(m, tt.input)
			cfg := config.ManagerConfig(m)
			tt.check(t, cfg)
		})
	}
}

func TestConvertPackageManagerFromMap(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]any
		check func(t *testing.T, cfg *config.Values)
	}{
		{
			name: "full package manager settings",
			input: map[string]any{
				"package_manager": map[string]any{
					"preferred": "bun",
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Equal(t, "bun", cfg.PackageManager.Preferred)
			},
		},
		{
			name: "package manager section is not a map",
			input: map[string]any{
				"package_manager": "not-a-map",
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Empty(t, cfg.PackageManager.Preferred)
			},
		},
		{
			name: "package manager with wrong type for preferred",
			input: map[string]any{
				"package_manager": map[string]any{
					"preferred": 12345,
				},
			},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Empty(t, cfg.PackageManager.Preferred)
			},
		},
		{
			name:  "missing package manager section",
			input: map[string]any{},
			check: func(t *testing.T, cfg *config.Values) {
				t.Helper()
				assert.Empty(t, cfg.PackageManager.Preferred)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := config.NewTestManager("", nil)
			config.ManagerConvertFromMap(m, tt.input)
			cfg := config.ManagerConfig(m)
			tt.check(t, cfg)
		})
	}
}

func TestConvertFromMap_AllSections(t *testing.T) {
	t.Run("all sections populated", func(t *testing.T) {
		m := config.NewTestManager("", nil)
		config.ManagerConvertFromMap(m, map[string]any{
			"validate": map[string]any{
				"timeout":  300.0,
				"cooldown": 20.0,
			},
			"notifications": map[string]any{
				"ntfy_topic": "all-sections-topic",
			},
			"compact": map[string]any{
				"threshold":         75.0,
				"reminder_interval": 30.0,
			},
			"notify": map[string]any{
				"quiet_hours": map[string]any{
					"enabled": false,
					"start":   "20:00",
					"end":     "09:00",
				},
				"audio": map[string]any{
					"enabled":   false,
					"directory": "/all/audio",
				},
				"desktop": map[string]any{
					"enabled": false,
				},
			},
			"observe": map[string]any{
				"enabled":          false,
				"max_file_size_mb": 50.0,
			},
			"learning": map[string]any{
				"min_session_length":  5.0,
				"learned_skills_path": "all/skills",
			},
			"pre_commit_reminder": map[string]any{
				"enabled": false,
				"command": "make all",
			},
			"package_manager": map[string]any{
				"preferred": "bun",
			},
		})
		cfg := config.ManagerConfig(m)

		assert.Equal(t, 300, cfg.Validate.Timeout)
		assert.Equal(t, 20, cfg.Validate.Cooldown)
		assert.Equal(t, "all-sections-topic", cfg.Notifications.NtfyTopic)
		assert.Equal(t, 75, cfg.Compact.Threshold)
		assert.Equal(t, 30, cfg.Compact.ReminderInterval)
		assert.False(t, cfg.Notify.QuietHours.Enabled)
		assert.Equal(t, "20:00", cfg.Notify.QuietHours.Start)
		assert.Equal(t, "09:00", cfg.Notify.QuietHours.End)
		assert.False(t, cfg.Notify.Audio.Enabled)
		assert.Equal(t, "/all/audio", cfg.Notify.Audio.Directory)
		assert.False(t, cfg.Notify.Desktop.Enabled)
		assert.False(t, cfg.Observe.Enabled)
		assert.Equal(t, 50, cfg.Observe.MaxFileSizeMB)
		assert.Equal(t, 5, cfg.Learning.MinSessionLength)
		assert.Equal(t, "all/skills", cfg.Learning.LearnedSkillsPath)
		assert.False(t, cfg.PreCommit.Enabled)
		assert.Equal(t, "make all", cfg.PreCommit.Command)
		assert.Equal(t, "bun", cfg.PackageManager.Preferred)
	})
}

func TestInstinctConfigDefaults(t *testing.T) {
	ctx := context.Background()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	m := config.NewManagerWithPath(configPath)

	require.NoError(t, m.EnsureConfig(ctx))

	cfg, err := m.GetConfig(ctx)
	require.NoError(t, err)

	assert.Equal(t, config.ExportDefaultInstinctPersonalPath(), cfg.Instinct.PersonalPath)
	assert.Equal(t, config.ExportDefaultInstinctInheritedPath(), cfg.Instinct.InheritedPath)
	assert.InDelta(t, config.ExportDefaultInstinctMinConfidence(), cfg.Instinct.MinConfidence, 0.001)
	assert.InDelta(t, config.ExportDefaultInstinctAutoApprove(), cfg.Instinct.AutoApprove, 0.001)
	assert.InDelta(t, config.ExportDefaultInstinctDecayRate(), cfg.Instinct.DecayRate, 0.001)
	assert.Equal(t, config.ExportDefaultInstinctMaxInstincts(), cfg.Instinct.MaxInstincts)
	assert.Equal(t, config.ExportDefaultInstinctClusterThreshold(), cfg.Instinct.ClusterThreshold)
}

func TestInstinctConfigSetGet(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		key       string
		setValue  string
		wantValue string
	}{
		{
			name:      "set and get personal path",
			key:       config.ExportKeyInstinctPersonalPath(),
			setValue:   "/custom/personal",
			wantValue: "/custom/personal",
		},
		{
			name:      "set and get inherited path",
			key:       config.ExportKeyInstinctInheritedPath(),
			setValue:   "/custom/inherited",
			wantValue: "/custom/inherited",
		},
		{
			name:      "set and get min confidence",
			key:       config.ExportKeyInstinctMinConfidence(),
			setValue:   "0.5",
			wantValue: "0.5",
		},
		{
			name:      "set and get auto approve",
			key:       config.ExportKeyInstinctAutoApprove(),
			setValue:   "0.9",
			wantValue: "0.9",
		},
		{
			name:      "set and get decay rate",
			key:       config.ExportKeyInstinctDecayRate(),
			setValue:   "0.05",
			wantValue: "0.05",
		},
		{
			name:      "set and get max instincts",
			key:       config.ExportKeyInstinctMaxInstincts(),
			setValue:   "200",
			wantValue: "200",
		},
		{
			name:      "set and get cluster threshold",
			key:       config.ExportKeyInstinctClusterThreshold(),
			setValue:   "5",
			wantValue: "5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")
			m := config.NewManagerWithPath(configPath)
			require.NoError(t, m.EnsureConfig(ctx))

			require.NoError(t, m.Set(ctx, tt.key, tt.setValue))

			value, found, err := m.GetValue(ctx, tt.key)
			require.NoError(t, err)
			assert.True(t, found)
			assert.Equal(t, tt.wantValue, value)

			// Verify persistence by reloading
			m2 := config.NewManagerWithPath(configPath)
			require.NoError(t, m2.EnsureConfig(ctx))

			value2, found2, err2 := m2.GetValue(ctx, tt.key)
			require.NoError(t, err2)
			assert.True(t, found2)
			assert.Equal(t, tt.wantValue, value2)
		})
	}
}
