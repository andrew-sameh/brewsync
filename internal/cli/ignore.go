package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/andrew-sameh/brewsync/internal/config"
)

var ignoreCmd = &cobra.Command{
	Use:   "ignore",
	Short: "Manage ignore lists",
	Long: `Manage packages that should be ignored during sync/import.

Packages can be ignored globally (on all machines) or per-machine.
Use the format "type:name" to specify packages, e.g., "cask:bluestacks".

Subcommands:
  list    Show all ignored packages
  add     Add a package to ignore list
  remove  Remove a package from ignore list
  clear   Clear all ignored packages`,
}

var (
	ignoreMachine string
	ignoreGlobal  bool
)

var ignoreListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show all ignored packages",
	RunE:  runIgnoreList,
}

var ignoreAddCmd = &cobra.Command{
	Use:   "add [type:name]",
	Short: "Add a package to ignore list",
	Long: `Add a package to the ignore list.

Format: type:name
Examples:
  brewsync ignore add cask:bluestacks
  brewsync ignore add brew:postgresql --machine mini
  brewsync ignore add vscode:some.extension --global`,
	Args: cobra.ExactArgs(1),
	RunE: runIgnoreAdd,
}

var ignoreRemoveCmd = &cobra.Command{
	Use:   "remove [type:name]",
	Short: "Remove a package from ignore list",
	Args:  cobra.ExactArgs(1),
	RunE:  runIgnoreRemove,
}

var ignoreClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all ignored packages",
	RunE:  runIgnoreClear,
}

func init() {
	ignoreListCmd.Flags().StringVar(&ignoreMachine, "machine", "", "show only for specific machine")

	ignoreAddCmd.Flags().StringVar(&ignoreMachine, "machine", "", "add to specific machine's ignore list")
	ignoreAddCmd.Flags().BoolVar(&ignoreGlobal, "global", false, "add to global ignore list")

	ignoreRemoveCmd.Flags().StringVar(&ignoreMachine, "machine", "", "remove from specific machine's ignore list")
	ignoreRemoveCmd.Flags().BoolVar(&ignoreGlobal, "global", false, "remove from global ignore list")

	ignoreClearCmd.Flags().StringVar(&ignoreMachine, "machine", "", "clear only specific machine's ignore list")

	ignoreCmd.AddCommand(ignoreListCmd)
	ignoreCmd.AddCommand(ignoreAddCmd)
	ignoreCmd.AddCommand(ignoreRemoveCmd)
	ignoreCmd.AddCommand(ignoreClearCmd)
	rootCmd.AddCommand(ignoreCmd)
}

func runIgnoreList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	hasEntries := false

	// Show global ignore list
	if ignoreMachine == "" {
		globalEntries := listIgnoreEntries(cfg.Ignore.Global)
		if len(globalEntries) > 0 {
			fmt.Println("Global ignore list:")
			for _, entry := range globalEntries {
				fmt.Printf("  %s\n", entry)
			}
			hasEntries = true
		}
	}

	// Show machine-specific ignore lists
	for machine, ignoreList := range cfg.Ignore.ByMachine {
		if ignoreMachine != "" && machine != ignoreMachine {
			continue
		}

		entries := listIgnoreEntries(ignoreList)
		if len(entries) > 0 {
			if hasEntries {
				fmt.Println()
			}
			fmt.Printf("Machine '%s' ignore list:\n", machine)
			for _, entry := range entries {
				fmt.Printf("  %s\n", entry)
			}
			hasEntries = true
		}
	}

	if !hasEntries {
		fmt.Println("No packages are being ignored.")
	}

	return nil
}

func listIgnoreEntries(list config.PackageIgnoreList) []string {
	var entries []string

	for _, name := range list.Tap {
		entries = append(entries, "tap:"+name)
	}
	for _, name := range list.Brew {
		entries = append(entries, "brew:"+name)
	}
	for _, name := range list.Cask {
		entries = append(entries, "cask:"+name)
	}
	for _, name := range list.VSCode {
		entries = append(entries, "vscode:"+name)
	}
	for _, name := range list.Cursor {
		entries = append(entries, "cursor:"+name)
	}
	for _, name := range list.Go {
		entries = append(entries, "go:"+name)
	}
	for _, name := range list.Mas {
		entries = append(entries, "mas:"+name)
	}

	return entries
}

func runIgnoreAdd(cmd *cobra.Command, args []string) error {
	pkgSpec := args[0]

	// Parse package spec
	parts := strings.SplitN(pkgSpec, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format; use 'type:name' (e.g., 'cask:bluestacks')")
	}

	pkgType := strings.ToLower(parts[0])
	pkgName := parts[1]

	// Validate type
	validTypes := map[string]bool{
		"tap": true, "brew": true, "cask": true,
		"vscode": true, "cursor": true, "go": true, "mas": true,
	}
	if !validTypes[pkgType] {
		return fmt.Errorf("invalid package type '%s'; valid types: tap, brew, cask, vscode, cursor, go, mas", pkgType)
	}

	// Determine scope
	scope := "global"
	if ignoreMachine != "" {
		scope = ignoreMachine
	} else if !ignoreGlobal {
		// Default to current machine if not specified
		cfg, err := config.Get()
		if err == nil && cfg.CurrentMachine != "" {
			scope = cfg.CurrentMachine
		}
	}

	// Update config
	if err := updateIgnoreList(pkgType, pkgName, scope, true); err != nil {
		return err
	}

	if scope == "global" {
		printInfo("Added %s to global ignore list", pkgSpec)
	} else {
		printInfo("Added %s to %s ignore list", pkgSpec, scope)
	}

	return nil
}

func runIgnoreRemove(cmd *cobra.Command, args []string) error {
	pkgSpec := args[0]

	// Parse package spec
	parts := strings.SplitN(pkgSpec, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format; use 'type:name' (e.g., 'cask:bluestacks')")
	}

	pkgType := strings.ToLower(parts[0])
	pkgName := parts[1]

	// Determine scope
	scope := "global"
	if ignoreMachine != "" {
		scope = ignoreMachine
	} else if !ignoreGlobal {
		cfg, err := config.Get()
		if err == nil && cfg.CurrentMachine != "" {
			scope = cfg.CurrentMachine
		}
	}

	// Update config
	if err := updateIgnoreList(pkgType, pkgName, scope, false); err != nil {
		return err
	}

	if scope == "global" {
		printInfo("Removed %s from global ignore list", pkgSpec)
	} else {
		printInfo("Removed %s from %s ignore list", pkgSpec, scope)
	}

	return nil
}

func runIgnoreClear(cmd *cobra.Command, args []string) error {
	if !assumeYes {
		scope := "all"
		if ignoreMachine != "" {
			scope = ignoreMachine
		}
		fmt.Printf("Clear %s ignore list(s)? [y/N] ", scope)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			return nil
		}
	}

	path, err := config.ConfigPath()
	if err != nil {
		return err
	}

	// Load config as raw map
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var cfg map[string]interface{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	ignore, ok := cfg["ignore"].(map[string]interface{})
	if !ok {
		printInfo("No ignore lists to clear")
		return nil
	}

	if ignoreMachine != "" {
		// Clear specific machine
		delete(ignore, ignoreMachine)
		printInfo("Cleared ignore list for %s", ignoreMachine)
	} else {
		// Clear all
		cfg["ignore"] = map[string]interface{}{}
		printInfo("Cleared all ignore lists")
	}

	// Write updated config
	outData, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, outData, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func updateIgnoreList(pkgType, pkgName, scope string, add bool) error {
	path, err := config.ConfigPath()
	if err != nil {
		return err
	}

	// Load config as raw map for manipulation
	var cfg map[string]interface{}
	if config.Exists() {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read config: %w", err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
	} else {
		cfg = make(map[string]interface{})
	}

	// Ensure ignore structure exists
	ignore, ok := cfg["ignore"].(map[string]interface{})
	if !ok {
		ignore = make(map[string]interface{})
		cfg["ignore"] = ignore
	}

	var targetList map[string]interface{}
	if scope == "global" {
		global, ok := ignore["global"].(map[string]interface{})
		if !ok {
			global = make(map[string]interface{})
			ignore["global"] = global
		}
		targetList = global
	} else {
		machines, ok := ignore["machines"].(map[string]interface{})
		if !ok {
			machines = make(map[string]interface{})
			ignore["machines"] = machines
		}
		machine, ok := machines[scope].(map[string]interface{})
		if !ok {
			machine = make(map[string]interface{})
			machines[scope] = machine
		}
		targetList = machine
	}

	// Get current list for this type
	var currentList []string
	if existing, ok := targetList[pkgType].([]interface{}); ok {
		for _, v := range existing {
			if s, ok := v.(string); ok {
				currentList = append(currentList, s)
			}
		}
	}

	if add {
		// Check if already in list
		for _, name := range currentList {
			if name == pkgName {
				return fmt.Errorf("%s:%s is already in the ignore list", pkgType, pkgName)
			}
		}
		currentList = append(currentList, pkgName)
	} else {
		// Remove from list
		var newList []string
		found := false
		for _, name := range currentList {
			if name == pkgName {
				found = true
			} else {
				newList = append(newList, name)
			}
		}
		if !found {
			return fmt.Errorf("%s:%s is not in the ignore list", pkgType, pkgName)
		}
		currentList = newList
	}

	targetList[pkgType] = currentList

	// Ensure directory exists
	if err := config.EnsureDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write updated config
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
