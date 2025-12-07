package config

import "github.com/spf13/viper"

// DefaultCategories is the default list of package types to sync
var DefaultCategories = []string{
	"tap",
	"brew",
	"cask",
	"vscode",
	"cursor",
	"antigravity",
	"go",
	"mas",
}

// DefaultCommitMessage is the default git commit message template
const DefaultCommitMessage = "brewsync: update {machine} Brewfile"

// setDefaults sets all default values in viper
func setDefaults() {
	// Machine detection
	viper.SetDefault("current_machine", "auto")

	// Default categories
	viper.SetDefault("default_categories", DefaultCategories)

	// Auto-dump settings
	viper.SetDefault("auto_dump.enabled", false)
	viper.SetDefault("auto_dump.after_install", false)
	viper.SetDefault("auto_dump.commit", false)
	viper.SetDefault("auto_dump.push", false)
	viper.SetDefault("auto_dump.commit_message", DefaultCommitMessage)

	// Dump settings
	viper.SetDefault("dump.use_brew_bundle", true) // Use 'brew bundle dump --describe' by default

	// Conflict resolution
	viper.SetDefault("conflict_resolution", string(ConflictAsk))

	// Output settings
	viper.SetDefault("output.color", true)
	viper.SetDefault("output.verbose", false)
	viper.SetDefault("output.show_descriptions", true)
}
