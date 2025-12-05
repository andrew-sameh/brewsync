package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_GetMachine(t *testing.T) {
	cfg := &Config{
		Machines: map[string]Machine{
			"mini": {
				Hostname:    "Andrews-Mac-mini",
				Brewfile:    "/path/to/mini/Brewfile",
				Description: "Mac Mini",
			},
			"air": {
				Hostname:    "Andrews-MacBook-Air",
				Brewfile:    "/path/to/air/Brewfile",
				Description: "MacBook Air",
			},
		},
	}

	t.Run("existing machine", func(t *testing.T) {
		machine, ok := cfg.GetMachine("mini")
		assert.True(t, ok)
		assert.Equal(t, "Andrews-Mac-mini", machine.Hostname)
		assert.Equal(t, "/path/to/mini/Brewfile", machine.Brewfile)
	})

	t.Run("non-existing machine", func(t *testing.T) {
		_, ok := cfg.GetMachine("nonexistent")
		assert.False(t, ok)
	})

	t.Run("empty machines map", func(t *testing.T) {
		emptyCfg := &Config{}
		_, ok := emptyCfg.GetMachine("any")
		assert.False(t, ok)
	})
}

func TestConfig_GetCurrentMachine(t *testing.T) {
	cfg := &Config{
		Machines: map[string]Machine{
			"mini": {Hostname: "Andrews-Mac-mini", Brewfile: "/path/to/Brewfile"},
		},
		CurrentMachine: "mini",
	}

	t.Run("current machine exists", func(t *testing.T) {
		machine, ok := cfg.GetCurrentMachine()
		assert.True(t, ok)
		assert.Equal(t, "Andrews-Mac-mini", machine.Hostname)
	})

	t.Run("current machine not set", func(t *testing.T) {
		cfg.CurrentMachine = ""
		_, ok := cfg.GetCurrentMachine()
		assert.False(t, ok)
	})

	t.Run("current machine not in map", func(t *testing.T) {
		cfg.CurrentMachine = "nonexistent"
		_, ok := cfg.GetCurrentMachine()
		assert.False(t, ok)
	})
}

func TestConflictResolution_Constants(t *testing.T) {
	assert.Equal(t, ConflictResolution("ask"), ConflictAsk)
	assert.Equal(t, ConflictResolution("skip"), ConflictSkip)
	assert.Equal(t, ConflictResolution("source-wins"), ConflictSourceWins)
	assert.Equal(t, ConflictResolution("current-wins"), ConflictCurrentWins)
}

func TestMachine_Fields(t *testing.T) {
	m := Machine{
		Hostname:    "test-host",
		Brewfile:    "/path/to/Brewfile",
		Description: "Test machine",
	}

	assert.Equal(t, "test-host", m.Hostname)
	assert.Equal(t, "/path/to/Brewfile", m.Brewfile)
	assert.Equal(t, "Test machine", m.Description)
}

func TestAutoDumpConfig_Defaults(t *testing.T) {
	cfg := AutoDumpConfig{}

	// Default values should be false/empty
	assert.False(t, cfg.Enabled)
	assert.False(t, cfg.AfterInstall)
	assert.False(t, cfg.Commit)
	assert.False(t, cfg.Push)
	assert.Empty(t, cfg.CommitMessage)
}

func TestOutputConfig_Defaults(t *testing.T) {
	cfg := OutputConfig{}

	// Default values should be false
	assert.False(t, cfg.Color)
	assert.False(t, cfg.Verbose)
	assert.False(t, cfg.ShowDescriptions)
}

func TestPackageIgnoreList_Empty(t *testing.T) {
	list := PackageIgnoreList{}

	assert.Nil(t, list.Tap)
	assert.Nil(t, list.Brew)
	assert.Nil(t, list.Cask)
	assert.Nil(t, list.VSCode)
	assert.Nil(t, list.Cursor)
	assert.Nil(t, list.Go)
	assert.Nil(t, list.Mas)
}

func TestPackageIgnoreList_WithValues(t *testing.T) {
	list := PackageIgnoreList{
		Brew: []string{"pkg1", "pkg2"},
		Cask: []string{"app1"},
	}

	assert.Equal(t, []string{"pkg1", "pkg2"}, list.Brew)
	assert.Equal(t, []string{"app1"}, list.Cask)
	assert.Nil(t, list.Tap)
}

func TestHooksConfig(t *testing.T) {
	hooks := HooksConfig{
		PreInstall:  "echo pre-install",
		PostInstall: "echo post-install",
		PreDump:     "echo pre-dump",
		PostDump:    "echo post-dump",
	}

	assert.Equal(t, "echo pre-install", hooks.PreInstall)
	assert.Equal(t, "echo post-install", hooks.PostInstall)
	assert.Equal(t, "echo pre-dump", hooks.PreDump)
	assert.Equal(t, "echo post-dump", hooks.PostDump)
}
