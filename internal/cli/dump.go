package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/andrew-sameh/brewsync/internal/brewfile"
	"github.com/andrew-sameh/brewsync/internal/config"
	"github.com/andrew-sameh/brewsync/internal/installer"
)

var (
	dumpCommit  bool
	dumpPush    bool
	dumpMessage string
)

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Update current machine's Brewfile from installed packages",
	Long: `Dump captures the current state of installed packages and writes them
to the machine's Brewfile. This includes:
- Homebrew taps, formulae, and casks
- VSCode extensions
- Cursor extensions
- Antigravity extensions
- Go tools
- Mac App Store apps

The Brewfile location is determined from the config for the current machine.`,
	RunE: runDump,
}

func init() {
	dumpCmd.Flags().BoolVar(&dumpCommit, "commit", false, "commit changes after dump")
	dumpCmd.Flags().BoolVar(&dumpPush, "push", false, "commit and push changes")
	dumpCmd.Flags().StringVarP(&dumpMessage, "message", "m", "", "custom commit message")
}

func runDump(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get current machine
	machine, ok := cfg.GetCurrentMachine()
	if !ok {
		return fmt.Errorf("current machine not configured (detected: %s)", cfg.CurrentMachine)
	}

	brewfilePath := machine.Brewfile
	if brewfilePath == "" {
		return fmt.Errorf("no Brewfile path configured for machine %s", cfg.CurrentMachine)
	}

	printInfo("Dumping packages for machine: %s", cfg.CurrentMachine)
	printVerbose("Brewfile path: %s", brewfilePath)

	// Ensure directory exists
	dir := filepath.Dir(brewfilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Collect packages
	var allPackages brewfile.Packages
	brewInst := installer.NewBrewInstaller()

	// Use brew bundle dump if configured (default), otherwise collect manually
	if cfg.Dump.UseBrewBundle && brewInst.IsAvailable() {
		printVerbose("Using 'brew bundle dump --describe' for Homebrew packages...")

		// Create temp file for brew bundle dump
		tmpFile := brewfilePath + ".brewbundle.tmp"
		if err := brewInst.DumpToFile(tmpFile); err != nil {
			printWarning("Failed to run 'brew bundle dump': %v", err)
		} else {
			// Parse the brew bundle output (includes taps, formulae, casks with descriptions)
			brewPkgs, err := brewfile.Parse(tmpFile)
			if err != nil {
				printWarning("Failed to parse brew bundle output: %v", err)
			} else {
				allPackages = append(allPackages, brewPkgs...)
				printVerbose("  Found %d Homebrew packages (with descriptions)", len(brewPkgs))
			}
			// Clean up temp file
			os.Remove(tmpFile)
		}
	} else if brewInst.IsAvailable() {
		// Manual collection (without descriptions for now, but could add them)
		printVerbose("Collecting Homebrew packages manually...")

		taps, err := brewInst.ListTaps()
		if err != nil {
			printWarning("Failed to list taps: %v", err)
		} else {
			allPackages = append(allPackages, taps...)
			printVerbose("  Found %d taps", len(taps))
		}

		formulae, err := brewInst.ListFormulae()
		if err != nil {
			printWarning("Failed to list formulae: %v", err)
		} else {
			allPackages = append(allPackages, formulae...)
			printVerbose("  Found %d formulae", len(formulae))
		}

		casks, err := brewInst.ListCasks()
		if err != nil {
			printWarning("Failed to list casks: %v", err)
		} else {
			allPackages = append(allPackages, casks...)
			printVerbose("  Found %d casks", len(casks))
		}
	} else {
		printWarning("Homebrew not available, skipping brew packages")
	}

	// VSCode extensions (only add if not already in brew bundle dump output)
	vscodeInst := installer.NewVSCodeInstaller()
	if vscodeInst.IsAvailable() {
		printVerbose("Collecting VSCode extensions...")
		extensions, err := vscodeInst.List()
		if err != nil {
			printWarning("Failed to list VSCode extensions: %v", err)
		} else {
			beforeCount := len(allPackages)
			allPackages = allPackages.AddUnique(extensions...)
			addedCount := len(allPackages) - beforeCount
			printVerbose("  Found %d extensions (%d new, %d already in Brewfile)", len(extensions), addedCount, len(extensions)-addedCount)
		}
	} else {
		printVerbose("VSCode CLI not available, skipping extensions")
	}

	// Cursor extensions (only add if not already in brew bundle dump output)
	cursorInst := installer.NewCursorInstaller()
	if cursorInst.IsAvailable() {
		printVerbose("Collecting Cursor extensions...")
		extensions, err := cursorInst.List()
		if err != nil {
			printWarning("Failed to list Cursor extensions: %v", err)
		} else {
			beforeCount := len(allPackages)
			allPackages = allPackages.AddUnique(extensions...)
			addedCount := len(allPackages) - beforeCount
			printVerbose("  Found %d extensions (%d new, %d already in Brewfile)", len(extensions), addedCount, len(extensions)-addedCount)
		}
	} else {
		printVerbose("Cursor CLI not available, skipping extensions")
	}

	// Antigravity extensions (only add if not already in brew bundle dump output)
	antigravityInst := installer.NewAntigravityInstaller()
	if antigravityInst.IsAvailable() {
		printVerbose("Collecting Antigravity extensions...")
		extensions, err := antigravityInst.List()
		if err != nil {
			printWarning("Failed to list Antigravity extensions: %v", err)
		} else {
			beforeCount := len(allPackages)
			allPackages = allPackages.AddUnique(extensions...)
			addedCount := len(allPackages) - beforeCount
			printVerbose("  Found %d extensions (%d new, %d already in Brewfile)", len(extensions), addedCount, len(extensions)-addedCount)
		}
	} else {
		printVerbose("Antigravity CLI not available, skipping extensions")
	}

	// Go tools (only add if not already in brew bundle dump output)
	goInst := installer.NewGoToolsInstaller()
	if goInst.IsAvailable() {
		printVerbose("Collecting Go tools...")
		tools, err := goInst.List()
		if err != nil {
			printWarning("Failed to list Go tools: %v", err)
		} else {
			beforeCount := len(allPackages)
			allPackages = allPackages.AddUnique(tools...)
			addedCount := len(allPackages) - beforeCount
			printVerbose("  Found %d tools (%d new, %d already in Brewfile)", len(tools), addedCount, len(tools)-addedCount)
		}
	} else {
		printVerbose("Go not available, skipping tools")
	}

	// Mac App Store apps (only add if not already in brew bundle dump output)
	masInst := installer.NewMasInstaller()
	if masInst.IsAvailable() {
		printVerbose("Collecting Mac App Store apps...")
		apps, err := masInst.List()
		if err != nil {
			printWarning("Failed to list Mac App Store apps: %v", err)
		} else {
			beforeCount := len(allPackages)
			allPackages = allPackages.AddUnique(apps...)
			addedCount := len(allPackages) - beforeCount
			printVerbose("  Found %d apps (%d new, %d already in Brewfile)", len(apps), addedCount, len(apps)-addedCount)
		}
	} else {
		printVerbose("mas CLI not available, skipping App Store apps")
	}

	// Dry run - just show what would be written
	if dryRun {
		printInfo("\nDry run - would write %d packages to %s:", len(allPackages), brewfilePath)
		byType := allPackages.ByType()
		for _, t := range brewfile.AllTypes() {
			if pkgs, ok := byType[t]; ok && len(pkgs) > 0 {
				printInfo("  %s: %d", t, len(pkgs))
			}
		}
		return nil
	}

	// Write Brewfile
	writer := brewfile.NewWriter(allPackages)
	if err := writer.Write(brewfilePath); err != nil {
		return fmt.Errorf("failed to write Brewfile: %w", err)
	}

	// Print summary
	byType := allPackages.ByType()
	printInfo("\nWrote %d packages to %s:", len(allPackages), brewfilePath)
	for _, t := range brewfile.AllTypes() {
		if pkgs, ok := byType[t]; ok && len(pkgs) > 0 {
			printInfo("  %s: %d", t, len(pkgs))
		}
	}

	// TODO: Handle --commit and --push flags
	if dumpCommit || dumpPush {
		printWarning("--commit and --push flags not yet implemented")
	}

	return nil
}
