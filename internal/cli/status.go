package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/andrew-sameh/brewsync/internal/brewfile"
	"github.com/andrew-sameh/brewsync/internal/config"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current state overview",
	Long: `Show current machine info, package counts, and pending changes.

Displays:
  - Current machine identification
  - Package counts by type
  - Pending changes from default source (if configured)
  - Last dump/sync times (from metadata)`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

// Metadata represents the .brewsync-meta file
type Metadata struct {
	Machine        string            `yaml:"machine"`
	LastDump       time.Time         `yaml:"last_dump"`
	LastSync       *LastSyncInfo     `yaml:"last_sync,omitempty"`
	PackageCounts  map[string]int    `yaml:"package_counts"`
	MacOSVersion   string            `yaml:"macos_version,omitempty"`
	BrewsyncVersion string           `yaml:"brewsync_version,omitempty"`
}

type LastSyncInfo struct {
	From    string    `yaml:"from"`
	At      time.Time `yaml:"at"`
	Added   int       `yaml:"added"`
	Removed int       `yaml:"removed"`
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Current machine info
	currentMachine := cfg.CurrentMachine
	if currentMachine == "" {
		fmt.Println("Current machine: not detected")
		fmt.Println("Run 'brewsync config init' to set up your machine.")
		return nil
	}

	machine, ok := cfg.Machines[currentMachine]
	if !ok {
		fmt.Printf("Current machine: %s (not configured)\n", currentMachine)
		return nil
	}

	fmt.Printf("Current machine: %s", currentMachine)
	if machine.Description != "" {
		fmt.Printf(" (%s)", machine.Description)
	}
	fmt.Println()

	if machine.Hostname != "" {
		printVerbose("Hostname: %s", machine.Hostname)
	}
	printVerbose("Brewfile: %s", machine.Brewfile)

	// Default source
	if cfg.DefaultSource != "" {
		fmt.Printf("Default source: %s\n", cfg.DefaultSource)
	}

	fmt.Println()

	// Package counts from Brewfile
	packages, err := brewfile.Parse(machine.Brewfile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Brewfile: not found")
			fmt.Printf("  Run 'brewsync dump' to create %s\n", machine.Brewfile)
		} else {
			fmt.Printf("Brewfile: error reading (%v)\n", err)
		}
	} else {
		fmt.Println("Package counts:")
		printPackageCounts(packages)
	}

	// Try to load metadata
	metaPath := filepath.Join(filepath.Dir(machine.Brewfile), ".brewsync-meta")
	meta, err := loadMetadata(metaPath)
	if err == nil && meta != nil {
		fmt.Println()
		if !meta.LastDump.IsZero() {
			fmt.Printf("Last dump: %s\n", formatTimeAgo(meta.LastDump))
		}
		if meta.LastSync != nil && !meta.LastSync.At.IsZero() {
			fmt.Printf("Last sync: %s from %s (+%d/-%d)\n",
				formatTimeAgo(meta.LastSync.At),
				meta.LastSync.From,
				meta.LastSync.Added,
				meta.LastSync.Removed)
		}
	}

	// Show pending changes if default source is configured
	if cfg.DefaultSource != "" && cfg.DefaultSource != currentMachine {
		sourceMachine, ok := cfg.Machines[cfg.DefaultSource]
		if ok {
			sourcePackages, err := brewfile.Parse(sourceMachine.Brewfile)
			if err == nil {
				diff := brewfile.Diff(sourcePackages, packages)
				if !diff.IsEmpty() {
					fmt.Println()
					fmt.Printf("Pending from %s:\n", cfg.DefaultSource)
					printPendingSummary(diff)
				}
			}
		}
	}

	return nil
}

func printPackageCounts(packages brewfile.Packages) {
	byType := packages.ByType()
	typeOrder := []brewfile.PackageType{
		brewfile.TypeTap,
		brewfile.TypeBrew,
		brewfile.TypeCask,
		brewfile.TypeVSCode,
		brewfile.TypeCursor,
		brewfile.TypeGo,
		brewfile.TypeMas,
	}

	var parts []string
	for _, t := range typeOrder {
		count := len(byType[t])
		if count > 0 {
			parts = append(parts, fmt.Sprintf("%s: %d", t, count))
		}
	}

	if len(parts) == 0 {
		fmt.Println("  (none)")
		return
	}

	// Print in a single line if short enough
	line := "  "
	for i, part := range parts {
		if i > 0 {
			line += " | "
		}
		line += part
	}
	fmt.Println(line)
}

func printPendingSummary(diff *brewfile.DiffResult) {
	if len(diff.Additions) > 0 {
		addByType := diff.AdditionsByType()
		var parts []string
		for t, pkgs := range addByType {
			if len(pkgs) > 0 {
				parts = append(parts, fmt.Sprintf("+%d %s", len(pkgs), t))
			}
		}
		fmt.Printf("  %s\n", strings.Join(parts, ", "))
	}
	if len(diff.Removals) > 0 {
		remByType := diff.RemovalsByType()
		var parts []string
		for t, pkgs := range remByType {
			if len(pkgs) > 0 {
				parts = append(parts, fmt.Sprintf("-%d %s", len(pkgs), t))
			}
		}
		fmt.Printf("  %s\n", strings.Join(parts, ", "))
	}
}

func loadMetadata(path string) (*Metadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var meta Metadata
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	}
	if duration < time.Hour {
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	if duration < 48*time.Hour {
		return "yesterday"
	}
	days := int(duration.Hours() / 24)
	return fmt.Sprintf("%d days ago", days)
}
