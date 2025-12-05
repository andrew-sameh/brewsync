package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/andrew-sameh/brewsync/internal/config"
	"github.com/andrew-sameh/brewsync/internal/exec"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Validate setup and diagnose issues",
	Long: `Check your BrewSync configuration and environment for potential issues.

Validates:
  - Config file exists and is valid
  - Current machine is detected
  - Brewfile paths exist
  - Required CLI tools are available (brew, code, cursor, mas, go)`,
	RunE: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

type checkResult struct {
	name    string
	ok      bool
	message string
}

func runDoctor(cmd *cobra.Command, args []string) error {
	var results []checkResult

	// Check config file
	results = append(results, checkConfigFile())

	// Load config for further checks
	cfg, err := config.Get()
	if err != nil {
		results = append(results, checkResult{
			name:    "Config loading",
			ok:      false,
			message: fmt.Sprintf("Failed to load config: %v", err),
		})
		printResults(results)
		return nil
	}

	// Check current machine
	results = append(results, checkCurrentMachine(cfg))

	// Check Brewfile paths
	results = append(results, checkBrewfilePaths(cfg)...)

	// Check default source
	if cfg.DefaultSource != "" {
		results = append(results, checkDefaultSource(cfg))
	}

	// Check CLI tools
	results = append(results, checkCLITools()...)

	printResults(results)
	return nil
}

func checkConfigFile() checkResult {
	if config.Exists() {
		path, _ := config.ConfigPath()
		return checkResult{
			name:    "Config file",
			ok:      true,
			message: fmt.Sprintf("Found at %s", path),
		}
	}
	return checkResult{
		name:    "Config file",
		ok:      false,
		message: "Not found. Run 'brewsync config init' to create one.",
	}
}

func checkCurrentMachine(cfg *config.Config) checkResult {
	if cfg.CurrentMachine == "" {
		return checkResult{
			name:    "Current machine",
			ok:      false,
			message: "Not detected. Check hostname configuration.",
		}
	}

	if _, ok := cfg.Machines[cfg.CurrentMachine]; !ok {
		return checkResult{
			name:    "Current machine",
			ok:      false,
			message: fmt.Sprintf("'%s' not found in machines config", cfg.CurrentMachine),
		}
	}

	return checkResult{
		name:    "Current machine",
		ok:      true,
		message: cfg.CurrentMachine,
	}
}

func checkBrewfilePaths(cfg *config.Config) []checkResult {
	var results []checkResult

	for name, machine := range cfg.Machines {
		if machine.Brewfile == "" {
			results = append(results, checkResult{
				name:    fmt.Sprintf("Brewfile (%s)", name),
				ok:      false,
				message: "Path not configured",
			})
			continue
		}

		if _, err := os.Stat(machine.Brewfile); os.IsNotExist(err) {
			// Only warn if it's the current machine
			if name == cfg.CurrentMachine {
				results = append(results, checkResult{
					name:    fmt.Sprintf("Brewfile (%s)", name),
					ok:      false,
					message: fmt.Sprintf("Not found at %s", machine.Brewfile),
				})
			}
		} else if err != nil {
			results = append(results, checkResult{
				name:    fmt.Sprintf("Brewfile (%s)", name),
				ok:      false,
				message: fmt.Sprintf("Error: %v", err),
			})
		} else {
			results = append(results, checkResult{
				name:    fmt.Sprintf("Brewfile (%s)", name),
				ok:      true,
				message: machine.Brewfile,
			})
		}
	}

	return results
}

func checkDefaultSource(cfg *config.Config) checkResult {
	if _, ok := cfg.Machines[cfg.DefaultSource]; !ok {
		return checkResult{
			name:    "Default source",
			ok:      false,
			message: fmt.Sprintf("'%s' not found in machines config", cfg.DefaultSource),
		}
	}

	return checkResult{
		name:    "Default source",
		ok:      true,
		message: cfg.DefaultSource,
	}
}

func checkCLITools() []checkResult {
	var results []checkResult

	tools := []struct {
		name     string
		command  string
		required bool
	}{
		{"Homebrew", "brew", true},
		{"brew bundle", "brew", true}, // Will check bundle separately
		{"VSCode CLI", "code", false},
		{"Cursor CLI", "cursor", false},
		{"Mac App Store CLI", "mas", false},
		{"Go", "go", false},
	}

	for _, tool := range tools {
		if tool.name == "brew bundle" {
			// Special check for brew bundle
			_, err := exec.Run("brew", "bundle", "--help")
			if err != nil {
				results = append(results, checkResult{
					name:    tool.name,
					ok:      false,
					message: "Not available. Run 'brew tap homebrew/bundle'",
				})
			} else {
				results = append(results, checkResult{
					name:    tool.name,
					ok:      true,
					message: "Available",
				})
			}
			continue
		}

		if exec.Exists(tool.command) {
			results = append(results, checkResult{
				name:    tool.name,
				ok:      true,
				message: "Installed",
			})
		} else {
			msg := "Not found"
			if !tool.required {
				msg += fmt.Sprintf(" (%s packages won't sync)", tool.name)
			}
			results = append(results, checkResult{
				name:    tool.name,
				ok:      !tool.required,
				message: msg,
			})
		}
	}

	return results
}

func printResults(results []checkResult) {
	for _, r := range results {
		var status string
		if r.ok {
			status = "✓"
		} else {
			status = "✗"
		}
		fmt.Printf("%s %s: %s\n", status, r.name, r.message)
	}

	// Summary
	var failures int
	var warnings int
	for _, r := range results {
		if !r.ok {
			failures++
		}
	}

	fmt.Println()
	if failures == 0 {
		fmt.Println("All checks passed!")
	} else {
		fmt.Printf("%d issue(s) found.\n", failures)
	}
	_ = warnings
}
