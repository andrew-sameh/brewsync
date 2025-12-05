package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/andrew-sameh/brewsync/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long: `Manage BrewSync configuration.

Subcommands:
  show         Display current configuration
  edit         Open config file in editor
  path         Show config file path
  init         Initialize configuration (interactive)
  add-machine  Add a new machine configuration`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	RunE:  runConfigShow,
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open config file in editor",
	RunE:  runConfigEdit,
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show config file path",
	RunE:  runConfigPath,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	Long: `Initialize BrewSync configuration interactively.

Creates a new config file with machine settings based on
the current hostname.`,
	RunE: runConfigInit,
}

var (
	addMachineHostname    string
	addMachineBrewfile    string
	addMachineDescription string
)

var configAddMachineCmd = &cobra.Command{
	Use:   "add-machine [name]",
	Short: "Add a new machine configuration",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigAddMachine,
}

func init() {
	configAddMachineCmd.Flags().StringVar(&addMachineHostname, "hostname", "", "hostname for auto-detection")
	configAddMachineCmd.Flags().StringVar(&addMachineBrewfile, "brewfile", "", "path to Brewfile")
	configAddMachineCmd.Flags().StringVar(&addMachineDescription, "description", "", "machine description")

	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configAddMachineCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Marshal to YAML for display
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to format config: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func runConfigEdit(cmd *cobra.Command, args []string) error {
	path, err := config.ConfigPath()
	if err != nil {
		return err
	}

	if !config.Exists() {
		return fmt.Errorf("config file does not exist; run 'brewsync config init' first")
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}

	editorCmd := exec.Command(editor, path)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	return editorCmd.Run()
}

func runConfigPath(cmd *cobra.Command, args []string) error {
	path, err := config.ConfigPath()
	if err != nil {
		return err
	}

	fmt.Println(path)
	return nil
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	path, err := config.ConfigPath()
	if err != nil {
		return err
	}

	if config.Exists() {
		printWarning("Config file already exists at %s", path)
		if !assumeYes {
			var overwrite bool
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("Overwrite existing config?").
						Value(&overwrite),
				),
			)
			if err := form.Run(); err != nil {
				return err
			}
			if !overwrite {
				printInfo("Init cancelled")
				return nil
			}
		}
	}

	// Detect hostname
	hostname, err := config.GetLocalHostname()
	if err != nil {
		hostname = "unknown"
	}

	// Create a simple machine name from hostname
	suggestedName := suggestMachineName(hostname)

	// Default Brewfile path
	home, _ := os.UserHomeDir()
	suggestedBrewfile := filepath.Join(home, "dotfiles", fmt.Sprintf("_brew_%s", suggestedName), "Brewfile")

	// Form values
	var (
		machineName     = suggestedName
		machineHostname = hostname
		brewfilePath    = suggestedBrewfile
		description     = fmt.Sprintf("Machine %s", suggestedName)
	)

	fmt.Printf("Detected hostname: %s\n\n", hostname)

	// Interactive form using Huh
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Machine name").
				Description("Short identifier for this machine (e.g., 'mini', 'air')").
				Value(&machineName).
				Validate(func(s string) error {
					s = strings.TrimSpace(s)
					if s == "" {
						return fmt.Errorf("machine name is required")
					}
					if strings.ContainsAny(s, " \t\n/\\") {
						return fmt.Errorf("machine name cannot contain spaces or slashes")
					}
					return nil
				}),

			huh.NewInput().
				Title("Hostname").
				Description("System hostname for auto-detection").
				Value(&machineHostname),

			huh.NewInput().
				Title("Brewfile path").
				Description("Path where this machine's Brewfile will be stored").
				Value(&brewfilePath).
				Validate(func(s string) error {
					s = strings.TrimSpace(s)
					if s == "" {
						return fmt.Errorf("Brewfile path is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Description").
				Description("Optional description for this machine").
				Value(&description),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	// Trim values
	machineName = strings.TrimSpace(machineName)
	machineHostname = strings.TrimSpace(machineHostname)
	brewfilePath = strings.TrimSpace(brewfilePath)
	description = strings.TrimSpace(description)

	// Update brewfile path if machine name changed
	if machineName != suggestedName && brewfilePath == suggestedBrewfile {
		brewfilePath = filepath.Join(home, "dotfiles", fmt.Sprintf("_brew_%s", machineName), "Brewfile")
	}

	// Expand ~ in path
	if strings.HasPrefix(brewfilePath, "~/") {
		brewfilePath = filepath.Join(home, brewfilePath[2:])
	}

	// Create initial config
	initialConfig := map[string]interface{}{
		"machines": map[string]interface{}{
			machineName: map[string]interface{}{
				"hostname":    machineHostname,
				"brewfile":    brewfilePath,
				"description": description,
			},
		},
		"current_machine":     "auto",
		"default_source":      machineName,
		"default_categories":  config.DefaultCategories,
		"conflict_resolution": config.ConflictAsk,
		"output": map[string]interface{}{
			"color":   true,
			"verbose": false,
		},
	}

	// Ensure directory exists
	if err := config.EnsureDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config
	data, err := yaml.Marshal(initialConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Println()
	printInfo("Created config file at %s", path)
	printInfo("Machine '%s' configured with Brewfile at %s", machineName, brewfilePath)
	printInfo("Run 'brewsync dump' to create your Brewfile")
	printInfo("Edit with 'brewsync config edit' to add more machines")

	return nil
}

func runConfigAddMachine(cmd *cobra.Command, args []string) error {
	machineName := args[0]

	path, err := config.ConfigPath()
	if err != nil {
		return err
	}

	// Load existing config
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

	// Ensure machines map exists
	machines, ok := cfg["machines"].(map[string]interface{})
	if !ok {
		machines = make(map[string]interface{})
		cfg["machines"] = machines
	}

	// Check if machine already exists
	if _, exists := machines[machineName]; exists {
		return fmt.Errorf("machine '%s' already exists", machineName)
	}

	// Build machine config
	machineConfig := make(map[string]interface{})

	if addMachineHostname != "" {
		machineConfig["hostname"] = addMachineHostname
	}

	if addMachineBrewfile != "" {
		machineConfig["brewfile"] = addMachineBrewfile
	} else {
		home, _ := os.UserHomeDir()
		machineConfig["brewfile"] = fmt.Sprintf("%s/dotfiles/_brew_%s/Brewfile", home, machineName)
	}

	if addMachineDescription != "" {
		machineConfig["description"] = addMachineDescription
	}

	machines[machineName] = machineConfig

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

	printInfo("Added machine '%s' to config", machineName)
	return nil
}

func suggestMachineName(hostname string) string {
	// Simple heuristics to suggest a short name
	name := hostname

	// Common patterns
	replacements := map[string]string{
		"Andrews-Mac-mini":    "mini",
		"Andrews-MacBook-Air": "air",
		"Andrews-MacBook-Pro": "pro",
		"Andrews-iMac":        "imac",
	}

	if short, ok := replacements[hostname]; ok {
		return short
	}

	// Default to lowercase first word
	if len(name) > 10 {
		name = name[:10]
	}

	return name
}
