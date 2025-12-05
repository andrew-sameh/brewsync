package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/andrew-sameh/brewsync/internal/config"
	"github.com/andrew-sameh/brewsync/pkg/version"
)

var (
	// Global flags
	cfgFile   string
	dryRun    bool
	verbose   bool
	quiet     bool
	noColor   bool
	assumeYes bool
)

// rootCmd is the base command
var rootCmd = &cobra.Command{
	Use:   "brewsync",
	Short: "Sync Homebrew packages across macOS machines",
	Long: `BrewSync is a CLI tool to sync Homebrew packages, casks, taps,
VSCode/Cursor extensions, Go tools, and Mac App Store apps
across multiple macOS machines.

It uses a git-based dotfiles workflow where each machine has its own
Brewfile, and provides commands to import, sync, and diff packages
between machines.`,
	Version: version.Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config loading for version command
		if cmd.Name() == "version" || cmd.Name() == "help" {
			return nil
		}

		// Set config path if provided
		if cfgFile != "" {
			config.SetConfigPath(cfgFile)
		}

		// Initialize config
		return config.Init()
	},
	SilenceUsage: true,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default ~/.config/brewsync/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "preview without executing")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "detailed output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "minimal output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVarP(&assumeYes, "yes", "y", false, "skip confirmations")

	// Add subcommands
	rootCmd.AddCommand(dumpCmd)
}

// printInfo prints an info message (respects quiet flag)
func printInfo(format string, args ...interface{}) {
	if !quiet {
		fmt.Printf(format+"\n", args...)
	}
}

// printVerbose prints a verbose message (respects verbose flag)
func printVerbose(format string, args ...interface{}) {
	if verbose && !quiet {
		fmt.Printf(format+"\n", args...)
	}
}

// printError prints an error message
func printError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}

// printWarning prints a warning message
func printWarning(format string, args ...interface{}) {
	if !quiet {
		fmt.Fprintf(os.Stderr, "Warning: "+format+"\n", args...)
	}
}
