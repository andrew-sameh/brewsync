package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ignorePath can be overridden for testing
var ignorePath string

// IgnorePath returns the path to the ignore.yaml file
func IgnorePath() string {
	if ignorePath != "" {
		return ignorePath
	}
	return filepath.Join(ConfigDir(), "ignore.yaml")
}

// SetIgnorePath overrides the ignore file path (for testing)
func SetIgnorePath(path string) {
	ignorePath = path
}

// LoadIgnoreFile loads the ignore.yaml file
// Returns an empty IgnoreFile if the file doesn't exist (not an error)
func LoadIgnoreFile() (*IgnoreFile, error) {
	path := IgnorePath()

	// If file doesn't exist, return empty ignore file
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &IgnoreFile{
			Global: IgnoreConfig{
				Categories: []string{},
				Packages:   PackageIgnoreList{},
			},
			Machines: make(map[string]IgnoreConfig),
		}, nil
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read ignore file: %w", err)
	}

	// Parse YAML
	var ignoreFile IgnoreFile
	if err := yaml.Unmarshal(data, &ignoreFile); err != nil {
		return nil, fmt.Errorf("failed to parse ignore file: %w", err)
	}

	// Initialize empty maps/slices if nil
	if ignoreFile.Machines == nil {
		ignoreFile.Machines = make(map[string]IgnoreConfig)
	}
	if ignoreFile.Global.Categories == nil {
		ignoreFile.Global.Categories = []string{}
	}

	return &ignoreFile, nil
}

// SaveIgnoreFile writes the ignore file to disk
func SaveIgnoreFile(ignoreFile *IgnoreFile) error {
	path := IgnorePath()

	// Ensure config directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(ignoreFile)
	if err != nil {
		return fmt.Errorf("failed to marshal ignore file: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write ignore file: %w", err)
	}

	return nil
}

// CreateDefaultIgnoreFile creates an ignore.yaml file with default empty structure
func CreateDefaultIgnoreFile() error {
	ignoreFile := &IgnoreFile{
		Global: IgnoreConfig{
			Categories: []string{},
			Packages: PackageIgnoreList{
				Tap:         []string{},
				Brew:        []string{},
				Cask:        []string{},
				VSCode:      []string{},
				Cursor:      []string{},
				Antigravity: []string{},
				Go:          []string{},
				Mas:         []string{},
			},
		},
		Machines: make(map[string]IgnoreConfig),
	}

	return SaveIgnoreFile(ignoreFile)
}

// AddCategoryIgnore adds a category to the ignore list
func AddCategoryIgnore(machine, category string, global bool) error {
	ignoreFile, err := LoadIgnoreFile()
	if err != nil {
		return err
	}

	if global || machine == "" {
		// Add to global
		if !contains(ignoreFile.Global.Categories, category) {
			ignoreFile.Global.Categories = append(ignoreFile.Global.Categories, category)
		}
	} else {
		// Add to machine-specific
		machineIgnore, ok := ignoreFile.Machines[machine]
		if !ok {
			machineIgnore = IgnoreConfig{
				Categories: []string{},
				Packages:   PackageIgnoreList{},
			}
		}

		if !contains(machineIgnore.Categories, category) {
			machineIgnore.Categories = append(machineIgnore.Categories, category)
		}

		ignoreFile.Machines[machine] = machineIgnore
	}

	return SaveIgnoreFile(ignoreFile)
}

// RemoveCategoryIgnore removes a category from the ignore list
func RemoveCategoryIgnore(machine, category string, global bool) error {
	ignoreFile, err := LoadIgnoreFile()
	if err != nil {
		return err
	}

	if global || machine == "" {
		// Remove from global
		ignoreFile.Global.Categories = removeString(ignoreFile.Global.Categories, category)
	} else {
		// Remove from machine-specific
		if machineIgnore, ok := ignoreFile.Machines[machine]; ok {
			machineIgnore.Categories = removeString(machineIgnore.Categories, category)
			ignoreFile.Machines[machine] = machineIgnore
		}
	}

	return SaveIgnoreFile(ignoreFile)
}

// AddPackageIgnore adds a package to the ignore list
// pkgID format: "type:name" (e.g., "cask:bluestacks")
func AddPackageIgnore(machine, pkgID string, global bool) error {
	ignoreFile, err := LoadIgnoreFile()
	if err != nil {
		return err
	}

	// Parse package ID
	pkgType, pkgName, err := parsePackageID(pkgID)
	if err != nil {
		return err
	}

	if global || machine == "" {
		// Add to global packages
		addPackageToList(&ignoreFile.Global.Packages, pkgType, pkgName)
	} else {
		// Add to machine-specific packages
		machineIgnore, ok := ignoreFile.Machines[machine]
		if !ok {
			machineIgnore = IgnoreConfig{
				Categories: []string{},
				Packages:   PackageIgnoreList{},
			}
		}

		addPackageToList(&machineIgnore.Packages, pkgType, pkgName)
		ignoreFile.Machines[machine] = machineIgnore
	}

	return SaveIgnoreFile(ignoreFile)
}

// RemovePackageIgnore removes a package from the ignore list
func RemovePackageIgnore(machine, pkgID string, global bool) error {
	ignoreFile, err := LoadIgnoreFile()
	if err != nil {
		return err
	}

	// Parse package ID
	pkgType, pkgName, err := parsePackageID(pkgID)
	if err != nil {
		return err
	}

	if global || machine == "" {
		// Remove from global packages
		removePackageFromList(&ignoreFile.Global.Packages, pkgType, pkgName)
	} else {
		// Remove from machine-specific packages
		if machineIgnore, ok := ignoreFile.Machines[machine]; ok {
			removePackageFromList(&machineIgnore.Packages, pkgType, pkgName)
			ignoreFile.Machines[machine] = machineIgnore
		}
	}

	return SaveIgnoreFile(ignoreFile)
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func removeString(slice []string, item string) []string {
	result := []string{}
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

func parsePackageID(pkgID string) (pkgType, pkgName string, err error) {
	parts := splitPackageID(pkgID)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid package ID format: %s (expected type:name)", pkgID)
	}
	return parts[0], parts[1], nil
}

func splitPackageID(pkgID string) []string {
	// Split on first colon only
	for i, c := range pkgID {
		if c == ':' {
			return []string{pkgID[:i], pkgID[i+1:]}
		}
	}
	return []string{pkgID}
}

func addPackageToList(list *PackageIgnoreList, pkgType, pkgName string) {
	switch pkgType {
	case "tap":
		if !contains(list.Tap, pkgName) {
			list.Tap = append(list.Tap, pkgName)
		}
	case "brew":
		if !contains(list.Brew, pkgName) {
			list.Brew = append(list.Brew, pkgName)
		}
	case "cask":
		if !contains(list.Cask, pkgName) {
			list.Cask = append(list.Cask, pkgName)
		}
	case "vscode":
		if !contains(list.VSCode, pkgName) {
			list.VSCode = append(list.VSCode, pkgName)
		}
	case "cursor":
		if !contains(list.Cursor, pkgName) {
			list.Cursor = append(list.Cursor, pkgName)
		}
	case "antigravity":
		if !contains(list.Antigravity, pkgName) {
			list.Antigravity = append(list.Antigravity, pkgName)
		}
	case "go":
		if !contains(list.Go, pkgName) {
			list.Go = append(list.Go, pkgName)
		}
	case "mas":
		if !contains(list.Mas, pkgName) {
			list.Mas = append(list.Mas, pkgName)
		}
	}
}

func removePackageFromList(list *PackageIgnoreList, pkgType, pkgName string) {
	switch pkgType {
	case "tap":
		list.Tap = removeString(list.Tap, pkgName)
	case "brew":
		list.Brew = removeString(list.Brew, pkgName)
	case "cask":
		list.Cask = removeString(list.Cask, pkgName)
	case "vscode":
		list.VSCode = removeString(list.VSCode, pkgName)
	case "cursor":
		list.Cursor = removeString(list.Cursor, pkgName)
	case "antigravity":
		list.Antigravity = removeString(list.Antigravity, pkgName)
	case "go":
		list.Go = removeString(list.Go, pkgName)
	case "mas":
		list.Mas = removeString(list.Mas, pkgName)
	}
}
