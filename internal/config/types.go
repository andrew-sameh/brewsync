package config

// Machine represents a macOS machine configuration
type Machine struct {
	Hostname    string `yaml:"hostname" mapstructure:"hostname"`
	Brewfile    string `yaml:"brewfile" mapstructure:"brewfile"`
	Description string `yaml:"description,omitempty" mapstructure:"description"`
}

// AutoDumpConfig configures automatic Brewfile updates
type AutoDumpConfig struct {
	Enabled       bool   `yaml:"enabled" mapstructure:"enabled"`
	AfterInstall  bool   `yaml:"after_install" mapstructure:"after_install"`
	Commit        bool   `yaml:"commit" mapstructure:"commit"`
	Push          bool   `yaml:"push" mapstructure:"push"`
	CommitMessage string `yaml:"commit_message" mapstructure:"commit_message"`
}

// DumpConfig configures how dump command works
type DumpConfig struct {
	UseBrewBundle bool `yaml:"use_brew_bundle" mapstructure:"use_brew_bundle"` // Use 'brew bundle dump --describe' for Homebrew packages
}

// PackageIgnoreList holds ignored packages by type
type PackageIgnoreList struct {
	Tap         []string `yaml:"tap,omitempty" mapstructure:"tap"`
	Brew        []string `yaml:"brew,omitempty" mapstructure:"brew"`
	Cask        []string `yaml:"cask,omitempty" mapstructure:"cask"`
	VSCode      []string `yaml:"vscode,omitempty" mapstructure:"vscode"`
	Cursor      []string `yaml:"cursor,omitempty" mapstructure:"cursor"`
	Antigravity []string `yaml:"antigravity,omitempty" mapstructure:"antigravity"`
	Go          []string `yaml:"go,omitempty" mapstructure:"go"`
	Mas         []string `yaml:"mas,omitempty" mapstructure:"mas"`
}

// IgnoreConfig holds category and package-level ignores
type IgnoreConfig struct {
	Categories []string          `yaml:"categories"` // Ignore entire categories (e.g., "mas", "go")
	Packages   PackageIgnoreList `yaml:"packages"`   // Ignore specific packages within non-ignored categories
}

// IgnoreFile represents the separate ignore.yaml file
type IgnoreFile struct {
	Global   IgnoreConfig            `yaml:"global"`   // Global ignores for all machines
	Machines map[string]IgnoreConfig `yaml:"machines"` // Per-machine ignores
}

// MachineSpecificConfig holds packages specific to each machine
type MachineSpecificConfig map[string]PackageIgnoreList

// OutputConfig configures CLI output behavior
type OutputConfig struct {
	Color            bool `yaml:"color" mapstructure:"color"`
	Verbose          bool `yaml:"verbose" mapstructure:"verbose"`
	ShowDescriptions bool `yaml:"show_descriptions" mapstructure:"show_descriptions"`
}

// HooksConfig holds shell commands to run at various points
type HooksConfig struct {
	PreInstall  string `yaml:"pre_install,omitempty" mapstructure:"pre_install"`
	PostInstall string `yaml:"post_install,omitempty" mapstructure:"post_install"`
	PreDump     string `yaml:"pre_dump,omitempty" mapstructure:"pre_dump"`
	PostDump    string `yaml:"post_dump,omitempty" mapstructure:"post_dump"`
}

// ConflictResolution defines how to handle conflicting ignore lists
type ConflictResolution string

const (
	ConflictAsk         ConflictResolution = "ask"
	ConflictSkip        ConflictResolution = "skip"
	ConflictSourceWins  ConflictResolution = "source-wins"
	ConflictCurrentWins ConflictResolution = "current-wins"
)

// Config is the main configuration structure
type Config struct {
	Machines           map[string]Machine    `yaml:"machines" mapstructure:"machines"`
	CurrentMachine     string                `yaml:"current_machine" mapstructure:"current_machine"`
	DefaultSource      string                `yaml:"default_source" mapstructure:"default_source"`
	DefaultCategories  []string              `yaml:"default_categories" mapstructure:"default_categories"`
	AutoDump           AutoDumpConfig        `yaml:"auto_dump" mapstructure:"auto_dump"`
	Dump               DumpConfig            `yaml:"dump" mapstructure:"dump"`
	MachineSpecific    MachineSpecificConfig `yaml:"machine_specific" mapstructure:"machine_specific"`
	ConflictResolution ConflictResolution    `yaml:"conflict_resolution" mapstructure:"conflict_resolution"`
	Output             OutputConfig          `yaml:"output" mapstructure:"output"`
	Hooks              HooksConfig           `yaml:"hooks" mapstructure:"hooks"`

	// Loaded separately from ignore.yaml (not in YAML)
	ignoreFile *IgnoreFile
}

// GetMachine returns the machine config for the given name
func (c *Config) GetMachine(name string) (Machine, bool) {
	m, ok := c.Machines[name]
	return m, ok
}

// GetCurrentMachine returns the current machine's config
func (c *Config) GetCurrentMachine() (Machine, bool) {
	return c.GetMachine(c.CurrentMachine)
}

// GetIgnoredCategories returns all ignored categories for a machine (global + machine-specific)
func (c *Config) GetIgnoredCategories(machine string) []string {
	if c.ignoreFile == nil {
		return []string{}
	}

	seen := make(map[string]bool)
	var result []string

	// Add global ignored categories
	for _, cat := range c.ignoreFile.Global.Categories {
		if !seen[cat] {
			seen[cat] = true
			result = append(result, cat)
		}
	}

	// Add machine-specific ignored categories
	if machineIgnore, ok := c.ignoreFile.Machines[machine]; ok {
		for _, cat := range machineIgnore.Categories {
			if !seen[cat] {
				seen[cat] = true
				result = append(result, cat)
			}
		}
	}

	return result
}

// GetIgnoredPackages returns all ignored package IDs for a machine (global + machine-specific)
// Package IDs are in format "type:name" (e.g., "cask:bluestacks")
// Does NOT include packages from ignored categories
func (c *Config) GetIgnoredPackages(machine string) []string {
	if c.ignoreFile == nil {
		return []string{}
	}

	var result []string

	// Add global ignored packages
	result = append(result, addPrefix("tap", c.ignoreFile.Global.Packages.Tap)...)
	result = append(result, addPrefix("brew", c.ignoreFile.Global.Packages.Brew)...)
	result = append(result, addPrefix("cask", c.ignoreFile.Global.Packages.Cask)...)
	result = append(result, addPrefix("vscode", c.ignoreFile.Global.Packages.VSCode)...)
	result = append(result, addPrefix("cursor", c.ignoreFile.Global.Packages.Cursor)...)
	result = append(result, addPrefix("antigravity", c.ignoreFile.Global.Packages.Antigravity)...)
	result = append(result, addPrefix("go", c.ignoreFile.Global.Packages.Go)...)
	result = append(result, addPrefix("mas", c.ignoreFile.Global.Packages.Mas)...)

	// Add machine-specific ignored packages
	if machineIgnore, ok := c.ignoreFile.Machines[machine]; ok {
		result = append(result, addPrefix("tap", machineIgnore.Packages.Tap)...)
		result = append(result, addPrefix("brew", machineIgnore.Packages.Brew)...)
		result = append(result, addPrefix("cask", machineIgnore.Packages.Cask)...)
		result = append(result, addPrefix("vscode", machineIgnore.Packages.VSCode)...)
		result = append(result, addPrefix("cursor", machineIgnore.Packages.Cursor)...)
		result = append(result, addPrefix("antigravity", machineIgnore.Packages.Antigravity)...)
		result = append(result, addPrefix("go", machineIgnore.Packages.Go)...)
		result = append(result, addPrefix("mas", machineIgnore.Packages.Mas)...)
	}

	return result
}

// GetMachineSpecificPackages returns machine-specific packages grouped by machine name
// Each value is a list of package IDs in format "type:name"
func (c *Config) GetMachineSpecificPackages() map[string][]string {
	result := make(map[string][]string)

	for machine, pkgs := range c.MachineSpecific {
		var ids []string
		ids = append(ids, addPrefix("tap", pkgs.Tap)...)
		ids = append(ids, addPrefix("brew", pkgs.Brew)...)
		ids = append(ids, addPrefix("cask", pkgs.Cask)...)
		ids = append(ids, addPrefix("vscode", pkgs.VSCode)...)
		ids = append(ids, addPrefix("cursor", pkgs.Cursor)...)
		ids = append(ids, addPrefix("antigravity", pkgs.Antigravity)...)
		ids = append(ids, addPrefix("go", pkgs.Go)...)
		ids = append(ids, addPrefix("mas", pkgs.Mas)...)
		result[machine] = ids
	}

	return result
}

// IsCategoryIgnored checks if an entire package category is ignored
func (c *Config) IsCategoryIgnored(machine, pkgType string) bool {
	if c.ignoreFile == nil {
		return false
	}

	// Check global ignored categories
	for _, cat := range c.ignoreFile.Global.Categories {
		if cat == pkgType {
			return true
		}
	}

	// Check machine-specific ignored categories
	if machineIgnore, ok := c.ignoreFile.Machines[machine]; ok {
		for _, cat := range machineIgnore.Categories {
			if cat == pkgType {
				return true
			}
		}
	}

	return false
}

// IsPackageIgnored checks if a specific package is ignored (not category)
// Package ID format: "type:name" (e.g., "cask:bluestacks")
func (c *Config) IsPackageIgnored(machine, pkgID string) bool {
	ignoredPackages := c.GetIgnoredPackages(machine)
	for _, ignored := range ignoredPackages {
		if ignored == pkgID {
			return true
		}
	}
	return false
}

// addPrefix adds a type prefix to each package name
func addPrefix(pkgType string, names []string) []string {
	result := make([]string, len(names))
	for i, name := range names {
		result[i] = pkgType + ":" + name
	}
	return result
}
