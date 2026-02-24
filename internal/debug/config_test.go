package debug_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/riddopic/cc-tools/internal/debug"
	"github.com/riddopic/cc-tools/internal/shared"
)

func TestNewManager(t *testing.T) {
	m := debug.NewManager()

	if m == nil {
		t.Fatal("NewManager() should not return nil")
	}

	cfg := m.ManagerConfig()
	if cfg == nil {
		t.Fatal("config should be initialized")
	}

	if cfg.EnabledDirs == nil {
		t.Error("EnabledDirs map should be initialized")
	}
}

// assertConfigDirCount is a test helper that checks the number of enabled dirs.
func assertConfigDirCount(t *testing.T, cfg *debug.Config, want int) {
	t.Helper()

	if len(cfg.EnabledDirs) != want {
		t.Errorf("expected %d enabled dirs, got %d", want, len(cfg.EnabledDirs))
	}
}

// assertDirEnabled is a test helper that checks if a directory is enabled.
func assertDirEnabled(t *testing.T, cfg *debug.Config, dir string) {
	t.Helper()

	if !cfg.EnabledDirs[dir] {
		t.Errorf("%s should be enabled", dir)
	}
}

func TestManagerLoad(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		setupFunc func(_ *testing.T, filepath string)
		checkFunc func(t *testing.T, m *debug.Manager)
		wantErr   bool
	}{
		{
			name: "loads valid config",
			setupFunc: func(_ *testing.T, configPath string) {
				config := &debug.Config{
					EnabledDirs: map[string]bool{
						"/path/one": true,
						"/path/two": true,
					},
				}
				data, _ := json.MarshalIndent(config, "", "  ")
				os.MkdirAll(filepath.Dir(configPath), 0o755)
				os.WriteFile(configPath, data, 0o600)
			},
			checkFunc: func(t *testing.T, m *debug.Manager) {
				t.Helper()
				cfg := m.ManagerConfig()
				assertConfigDirCount(t, cfg, 2)
				assertDirEnabled(t, cfg, "/path/one")
				assertDirEnabled(t, cfg, "/path/two")
			},
			wantErr: false,
		},
		{
			name:      "handles missing file",
			setupFunc: nil,
			checkFunc: func(t *testing.T, m *debug.Manager) {
				t.Helper()
				cfg := m.ManagerConfig()
				if cfg == nil {
					t.Error("config should be initialized even when file missing")
				}
				assertConfigDirCount(t, cfg, 0)
			},
			wantErr: false,
		},
		{
			name: "handles empty file",
			setupFunc: func(_ *testing.T, configPath string) {
				os.MkdirAll(filepath.Dir(configPath), 0o755)
				os.WriteFile(configPath, []byte(""), 0o600)
			},
			checkFunc: func(t *testing.T, m *debug.Manager) {
				t.Helper()
				assertConfigDirCount(t, m.ManagerConfig(), 0)
			},
			wantErr: false,
		},
		{
			name: "handles nil EnabledDirs in JSON",
			setupFunc: func(_ *testing.T, configPath string) {
				data := []byte(`{"enabled_dirs": null}`)
				os.MkdirAll(filepath.Dir(configPath), 0o755)
				os.WriteFile(configPath, data, 0o600)
			},
			checkFunc: func(t *testing.T, m *debug.Manager) {
				t.Helper()
				if m.ManagerConfig().EnabledDirs == nil {
					t.Error("EnabledDirs should be initialized even when null in JSON")
				}
			},
			wantErr: false,
		},
		{
			name: "handles corrupt JSON",
			setupFunc: func(_ *testing.T, configPath string) {
				os.MkdirAll(filepath.Dir(configPath), 0o755)
				os.WriteFile(configPath, []byte("{invalid json}"), 0o600)
			},
			checkFunc: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "debug-config.json")

			m := debug.NewTestManager(configPath)

			if tt.setupFunc != nil {
				tt.setupFunc(t, configPath)
			}

			err := m.Load(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && tt.checkFunc != nil {
				tt.checkFunc(t, m)
			}
		})
	}
}

// assertFileExists is a test helper that checks if a file exists.
func assertFileExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("file should exist: %s", path)
	}
}

// assertFilePermissions is a test helper that checks file permissions.
func assertFilePermissions(t *testing.T, path string, want os.FileMode) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Failed to stat %s: %v", path, err)
	}

	if info.Mode().Perm() != want {
		t.Errorf("permissions = %v, want %v", info.Mode().Perm(), want)
	}
}

// assertSavedConfigValid reads and validates a saved config file.
func assertSavedConfigValid(t *testing.T, configFile string) {
	t.Helper()

	data, readErr := os.ReadFile(configFile)
	if readErr != nil {
		t.Fatalf("Failed to read saved file: %v", readErr)
	}

	var saved debug.Config
	if unmarshalErr := json.Unmarshal(data, &saved); unmarshalErr != nil {
		t.Fatalf("Failed to unmarshal saved config: %v", unmarshalErr)
	}

	assertDirEnabled(t, &saved, "/test/path")
	assertFilePermissions(t, configFile, 0o600)

	if !strings.Contains(string(data), "  ") {
		t.Error("Config should be pretty-printed with indentation")
	}

	if data[len(data)-1] != '\n' {
		t.Error("Config file should end with newline")
	}
}

func TestManagerSave(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		config    *debug.Config
		setupFunc func(_ *testing.T, filepath string)
		wantErr   bool
		checkFunc func(t *testing.T, configFile string)
	}{
		{
			name: "saves config successfully",
			config: &debug.Config{
				EnabledDirs: map[string]bool{
					"/test/path": true,
					"/another":   true,
				},
			},
			setupFunc: nil,
			wantErr:   false,
			checkFunc: assertSavedConfigValid,
		},
		{
			name: "creates directory if missing",
			config: &debug.Config{
				EnabledDirs: map[string]bool{"/path": true},
			},
			setupFunc: func(_ *testing.T, configPath string) {
				os.RemoveAll(filepath.Dir(configPath))
			},
			wantErr: false,
			checkFunc: func(t *testing.T, configFile string) {
				t.Helper()
				assertFileExists(t, configFile)
				assertFilePermissions(t, filepath.Dir(configFile), 0o750)
			},
		},
		{
			name: "uses atomic write with temp file",
			config: &debug.Config{
				EnabledDirs: map[string]bool{"/atomic": true},
			},
			setupFunc: nil,
			wantErr:   false,
			checkFunc: func(t *testing.T, configFile string) {
				t.Helper()
				tempFile := configFile + ".tmp"
				if _, statErr := os.Stat(tempFile); !os.IsNotExist(statErr) {
					t.Error("Temp file should be cleaned up after successful write")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "subdir", "debug-config.json")

			if tt.setupFunc != nil {
				tt.setupFunc(t, configPath)
			}

			m := debug.NewTestManagerWithConfig(configPath, tt.config)

			err := m.Save(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && tt.checkFunc != nil {
				tt.checkFunc(t, configPath)
			}
		})
	}
}

func TestManagerEnable(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		inputDir     string
		initialDirs  map[string]bool
		wantLogPath  bool
		checkEnabled []string
	}{
		{
			name:         "enables relative directory",
			inputDir:     ".",
			initialDirs:  nil,
			wantLogPath:  true,
			checkEnabled: []string{"."},
		},
		{
			name:         "enables absolute directory",
			inputDir:     "/test/absolute/path",
			initialDirs:  nil,
			wantLogPath:  true,
			checkEnabled: []string{"/test/absolute/path"},
		},
		{
			name:     "adds to existing enabled dirs",
			inputDir: "/new/path",
			initialDirs: map[string]bool{
				"/existing/path": true,
			},
			wantLogPath:  true,
			checkEnabled: []string{"/new/path", "/existing/path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "debug-config.json")

			config := &debug.Config{
				EnabledDirs: make(map[string]bool),
			}
			if tt.initialDirs != nil {
				config.EnabledDirs = tt.initialDirs
			}

			m := debug.NewTestManagerWithConfig(configPath, config)

			logPath, err := m.Enable(ctx, tt.inputDir)
			if err != nil {
				t.Fatalf("Enable() error = %v", err)
			}

			if tt.wantLogPath && logPath == "" {
				t.Error("Expected log path but got empty string")
			}

			absDir, _ := filepath.Abs(tt.inputDir)
			assertDirEnabled(t, m.ManagerConfig(), absDir)

			assertDirPersistedToDisk(t, configPath, absDir)

			for _, dir := range tt.checkEnabled {
				absCheckDir, _ := filepath.Abs(dir)
				assertDirEnabled(t, m.ManagerConfig(), absCheckDir)
			}
		})
	}
}

// assertDirPersistedToDisk verifies a directory is saved in the config file on disk.
func assertDirPersistedToDisk(t *testing.T, configPath, absDir string) {
	t.Helper()

	var saved debug.Config

	data, _ := os.ReadFile(configPath)
	json.Unmarshal(data, &saved)

	if !saved.EnabledDirs[absDir] {
		t.Error("Enabled directory should be persisted to disk")
	}
}

func TestManagerDisable(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		inputDir    string
		initialDirs map[string]bool
		checkFunc   func(t *testing.T, m *debug.Manager)
	}{
		{
			name:     "disables existing directory",
			inputDir: "/test/path",
			initialDirs: map[string]bool{
				"/test/path":  true,
				"/other/path": true,
			},
			checkFunc: func(t *testing.T, m *debug.Manager) {
				t.Helper()
				absDir, _ := filepath.Abs("/test/path")
				if m.ManagerConfig().EnabledDirs[absDir] {
					t.Error("/test/path should be disabled")
				}

				otherDir, _ := filepath.Abs("/other/path")
				assertDirEnabled(t, m.ManagerConfig(), otherDir)
			},
		},
		{
			name:        "handles disabling non-existent directory",
			inputDir:    "/not/enabled",
			initialDirs: map[string]bool{"/other": true},
			checkFunc: func(t *testing.T, m *debug.Manager) {
				t.Helper()
				assertConfigDirCount(t, m.ManagerConfig(), 1)
			},
		},
		{
			name:        "handles relative paths",
			inputDir:    "./relative",
			initialDirs: map[string]bool{},
			checkFunc: func(t *testing.T, m *debug.Manager) {
				t.Helper()
				assertConfigDirCount(t, m.ManagerConfig(), 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "debug-config.json")

			absDirs := make(map[string]bool)
			for dir, enabled := range tt.initialDirs {
				absDir, _ := filepath.Abs(dir)
				absDirs[absDir] = enabled
			}

			m := debug.NewTestManagerWithConfig(configPath, &debug.Config{
				EnabledDirs: absDirs,
			})

			err := m.Disable(ctx, tt.inputDir)
			if err != nil {
				t.Fatalf("Disable() error = %v", err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, m)
			}

			assertDisabledDirNotPersisted(t, configPath, tt.inputDir)
		})
	}
}

// assertDisabledDirNotPersisted verifies a disabled directory is not in the saved config.
func assertDisabledDirNotPersisted(t *testing.T, configPath, inputDir string) {
	t.Helper()

	var saved debug.Config

	data, _ := os.ReadFile(configPath)
	if len(data) == 0 {
		return
	}

	json.Unmarshal(data, &saved)

	absInput, _ := filepath.Abs(inputDir)
	if saved.EnabledDirs[absInput] {
		t.Error("Disabled directory should not be in saved config")
	}
}

func TestManagerIsEnabled(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		checkDir    string
		enabledDirs map[string]bool
		want        bool
	}{
		{
			name:        "exact match enabled",
			checkDir:    "/test/path",
			enabledDirs: map[string]bool{"/test/path": true},
			want:        true,
		},
		{
			name:        "child directory enabled",
			checkDir:    "/test/path/subdir/deep",
			enabledDirs: map[string]bool{"/test/path": true},
			want:        true,
		},
		{
			name:        "parent directory not enabled",
			checkDir:    "/test",
			enabledDirs: map[string]bool{"/test/path": true},
			want:        false,
		},
		{
			name:        "no directories enabled",
			checkDir:    "/any/path",
			enabledDirs: map[string]bool{},
			want:        false,
		},
		{
			name:     "multiple enabled dirs, one matches",
			checkDir: "/project/src/file",
			enabledDirs: map[string]bool{
				"/other":   true,
				"/project": true,
				"/another": true,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "debug-config.json")

			absDirs := make(map[string]bool)
			for dir, enabled := range tt.enabledDirs {
				absDir, _ := filepath.Abs(dir)
				absDirs[absDir] = enabled
			}

			m := debug.NewTestManagerWithConfig(configPath, &debug.Config{
				EnabledDirs: absDirs,
			})

			m.Save(ctx)

			m = debug.NewTestManager(configPath)

			enabled, err := m.IsEnabled(ctx, tt.checkDir)
			if err != nil {
				t.Fatalf("IsEnabled() error = %v", err)
			}

			if enabled != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", enabled, tt.want)
			}
		})
	}
}

func TestManagerGetEnabledDirs(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		enabledDirs map[string]bool
		wantCount   int
	}{
		{
			name:        "returns empty list when none enabled",
			enabledDirs: map[string]bool{},
			wantCount:   0,
		},
		{
			name: "returns all enabled directories",
			enabledDirs: map[string]bool{
				"/path/one":   true,
				"/path/two":   true,
				"/path/three": true,
			},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "debug-config.json")

			absDirs := make(map[string]bool)
			expectedDirs := make(map[string]bool)
			for dir, enabled := range tt.enabledDirs {
				absDir, _ := filepath.Abs(dir)
				absDirs[absDir] = enabled
				expectedDirs[absDir] = true
			}

			m := debug.NewTestManagerWithConfig(configPath, &debug.Config{
				EnabledDirs: absDirs,
			})

			m.Save(ctx)
			m = debug.NewTestManager(configPath)

			dirs, err := m.GetEnabledDirs(ctx)
			if err != nil {
				t.Fatalf("GetEnabledDirs() error = %v", err)
			}

			if len(dirs) != tt.wantCount {
				t.Errorf("GetEnabledDirs() returned %d dirs, want %d", len(dirs), tt.wantCount)
			}

			for _, dir := range dirs {
				if !expectedDirs[dir] {
					t.Errorf("Unexpected directory in result: %s", dir)
				}
			}
		})
	}
}

func TestGetLogFilePath(t *testing.T) {
	tempPrefix := filepath.Join(os.TempDir(), "cc-tools-") //nolint:usetesting // verifying production os.TempDir usage

	tests := []struct {
		name       string
		inputDir   string
		wantPrefix string
		wantSuffix string
	}{
		{
			name:       "generates path for normal directory",
			inputDir:   "/home/user/project",
			wantPrefix: tempPrefix + "user-project-",
			wantSuffix: ".debug",
		},
		{
			name:       "handles root directory",
			inputDir:   "/",
			wantPrefix: tempPrefix + "root-",
			wantSuffix: ".debug",
		},
		{
			name:       "handles relative path",
			inputDir:   ".",
			wantPrefix: tempPrefix,
			wantSuffix: ".debug",
		},
		{
			name:       "sanitizes directory name",
			inputDir:   "/path/with/many/levels",
			wantPrefix: tempPrefix + "many-levels-",
			wantSuffix: ".debug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logPath := debug.GetLogFilePath(tt.inputDir)

			assertHasPrefix(t, logPath, tt.wantPrefix)
			assertHasSuffix(t, logPath, tt.wantSuffix)
			assertHashLength(t, logPath)
		})
	}
}

// assertHasPrefix is a test helper that checks string prefix.
func assertHasPrefix(t *testing.T, s, prefix string) {
	t.Helper()

	if !strings.HasPrefix(s, prefix) {
		t.Errorf("got %s, want prefix %s", s, prefix)
	}
}

// assertHasSuffix is a test helper that checks string suffix.
func assertHasSuffix(t *testing.T, s, suffix string) {
	t.Helper()

	if !strings.HasSuffix(s, suffix) {
		t.Errorf("got %s, want suffix %s", s, suffix)
	}
}

// assertHashLength validates the hash portion of a log file path is 8 chars.
func assertHashLength(t *testing.T, logPath string) {
	t.Helper()

	parts := strings.Split(logPath, "-")
	lastPart := parts[len(parts)-1]
	hashPart := strings.TrimSuffix(lastPart, ".debug")

	if len(hashPart) != 8 {
		t.Errorf("Hash part should be 8 chars (4 bytes hex), got %d: %s", len(hashPart), hashPart)
	}
}

func TestGetLogFilePathConsistency(t *testing.T) {
	dir := "/test/directory"

	path1 := debug.GetLogFilePath(dir)
	path2 := debug.GetLogFilePath(dir)

	if path1 != path2 {
		t.Errorf("GetLogFilePath() should be consistent, got %s and %s", path1, path2)
	}

	path3 := debug.GetLogFilePath("/different/directory")
	if path1 == path3 {
		t.Error("Different directories should produce different log paths")
	}
}

func TestGetLogFilePathUsesSharedNaming(t *testing.T) {
	dir := "/some/project"
	logPath := debug.GetLogFilePath(dir)
	sharedPath := shared.GetDebugLogPathForDir(dir)
	assert.Equal(t, sharedPath, logPath, "GetLogFilePath should delegate to shared.GetDebugLogPathForDir")
	assert.True(t, strings.HasSuffix(logPath, ".debug"), "should use .debug extension")
	assert.NotContains(t, logPath, ".log", "should not use .log extension")
}

func TestManagerConcurrency(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "debug-config.json")

	m := debug.NewTestManager(configPath)

	done := make(chan bool, 3)

	go func() {
		for i := range 10 {
			dir := filepath.Join("/test", string(rune('a'+i)))
			m.Enable(ctx, dir)
		}
		done <- true
	}()

	go func() {
		for i := range 10 {
			dir := filepath.Join("/test", string(rune('a'+i)))
			m.Disable(ctx, dir)
		}
		done <- true
	}()

	go func() {
		for range 10 {
			m.IsEnabled(ctx, "/test/a")
			m.GetEnabledDirs(ctx)
		}
		done <- true
	}()

	for range 3 {
		<-done
	}
}

func TestGetConfigDir(t *testing.T) {
	tests := []struct {
		name    string
		homeDir string
		wantDir string
	}{
		{
			name:    "uses home directory",
			homeDir: "/home/testuser",
			wantDir: "/home/testuser/.config/cc-tools",
		},
		{
			name:    "falls back to /tmp when home not available",
			homeDir: "",
			wantDir: "/tmp/.config/cc-tools",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", "")
			if tt.homeDir != "" {
				t.Setenv("HOME", tt.homeDir)
			} else {
				t.Setenv("HOME", "")
				os.Unsetenv("HOME")
			}

			configDir := debug.ExportGetConfigDir()

			if configDir == "" {
				t.Error("getConfigDir() should not return empty string")
			}

			if !strings.HasSuffix(configDir, "cc-tools") {
				t.Errorf("getConfigDir() = %s, should end with cc-tools", configDir)
			}
		})
	}
}
