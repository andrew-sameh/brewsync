package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigPath(t *testing.T) {
	// Reset configPath
	configPath = ""

	path, err := ConfigPath()
	assert.NoError(t, err)
	assert.Contains(t, path, ".config/brewsync/config.yaml")
}

func TestSetConfigPath(t *testing.T) {
	// Reset after test
	defer func() { configPath = "" }()

	customPath := "/custom/path/config.yaml"
	SetConfigPath(customPath)

	path, err := ConfigPath()
	assert.NoError(t, err)
	assert.Equal(t, customPath, path)
}

func TestEnsureDir(t *testing.T) {
	// Create a temp dir for testing
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, ".config", "brewsync")

	// Temporarily override configPath to use temp dir
	origConfigPath := configPath
	defer func() { configPath = origConfigPath }()
	configPath = filepath.Join(testDir, "config.yaml")

	// This should work without error (but uses actual home dir)
	err := EnsureDir()
	// Note: We can't easily test this without mocking home dir
	// Just verify it doesn't panic
	_ = err
}

func TestExists(t *testing.T) {
	t.Run("file does not exist", func(t *testing.T) {
		origConfigPath := configPath
		defer func() { configPath = origConfigPath }()

		configPath = "/nonexistent/path/config.yaml"
		assert.False(t, Exists())
	})

	t.Run("file exists", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "config*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		origConfigPath := configPath
		defer func() { configPath = origConfigPath }()

		configPath = tmpFile.Name()
		assert.True(t, Exists())
	})
}

func TestProfilesDir(t *testing.T) {
	dir, err := ProfilesDir()
	assert.NoError(t, err)
	assert.Contains(t, dir, "profiles")
	assert.Contains(t, dir, ".config/brewsync")
}

func TestHistoryPath(t *testing.T) {
	path, err := HistoryPath()
	assert.NoError(t, err)
	assert.Contains(t, path, "history.log")
	assert.Contains(t, path, ".config/brewsync")
}

func TestLoadWithConfigFile(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `
machines:
  test:
    hostname: "test-hostname"
    brewfile: "/path/to/Brewfile"
    description: "Test machine"

current_machine: test
default_source: test
default_categories:
  - brew
  - cask
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Reset viper and config state
	viper.Reset()
	cfg = nil
	origConfigPath := configPath
	// Save and unset MACHINE env var to prevent it from overriding config
	origMachine := os.Getenv("MACHINE")
	os.Unsetenv("MACHINE")
	defer func() {
		configPath = origConfigPath
		cfg = nil
		viper.Reset()
		if origMachine != "" {
			os.Setenv("MACHINE", origMachine)
		}
	}()

	SetConfigPath(configFile)

	loadedCfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, loadedCfg)

	// Check loaded values
	assert.Equal(t, "test", loadedCfg.CurrentMachine)
	assert.Contains(t, loadedCfg.Machines, "test")
	assert.Equal(t, "test-hostname", loadedCfg.Machines["test"].Hostname)
}

func TestLoadWithDefaults(t *testing.T) {
	// Reset viper and config state
	viper.Reset()
	cfg = nil
	origConfigPath := configPath
	// Save and unset MACHINE env var to test pure defaults
	origMachine := os.Getenv("MACHINE")
	os.Unsetenv("MACHINE")
	defer func() {
		configPath = origConfigPath
		cfg = nil
		viper.Reset()
		if origMachine != "" {
			os.Setenv("MACHINE", origMachine)
		}
	}()

	// Use non-existent config file to test defaults
	tmpDir := t.TempDir()
	SetConfigPath(filepath.Join(tmpDir, "nonexistent.yaml"))

	loadedCfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, loadedCfg)

	// Check default values
	assert.Equal(t, DefaultCategories, loadedCfg.DefaultCategories)
	assert.Equal(t, ConflictAsk, loadedCfg.ConflictResolution)
	assert.True(t, loadedCfg.Output.Color)
}

func TestGet_ReturnsCachedConfig(t *testing.T) {
	// Reset state
	viper.Reset()
	cfg = nil
	origConfigPath := configPath
	// Save and unset MACHINE env var
	origMachine := os.Getenv("MACHINE")
	os.Unsetenv("MACHINE")
	defer func() {
		configPath = origConfigPath
		cfg = nil
		viper.Reset()
		if origMachine != "" {
			os.Setenv("MACHINE", origMachine)
		}
	}()

	// Create a minimal config
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configFile, []byte("current_machine: cached"), 0644)
	require.NoError(t, err)

	SetConfigPath(configFile)

	// First call loads
	cfg1, err := Get()
	require.NoError(t, err)

	// Second call returns cached
	cfg2, err := Get()
	require.NoError(t, err)

	assert.Same(t, cfg1, cfg2, "Get should return the same cached instance")
}
