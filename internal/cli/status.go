package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
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

	const tableWidth = 80

	// Current machine info
	currentMachine := cfg.CurrentMachine
	if currentMachine == "" {
		errorBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(catRed).
			Padding(0, 2).
			Width(tableWidth).
			Align(lipgloss.Center).
			Foreground(catRed)

		fmt.Println()
		fmt.Println(errorBox.Render("⚠ Current machine not detected"))
		fmt.Println()
		printInfo("Run 'brewsync config init' to set up your machine.")
		return nil
	}

	machine, ok := cfg.Machines[currentMachine]
	if !ok {
		errorBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(catYellow).
			Padding(0, 2).
			Width(tableWidth).
			Align(lipgloss.Center).
			Foreground(catYellow)

		fmt.Println()
		fmt.Println(errorBox.Render(fmt.Sprintf("⚠ Machine '%s' not configured", currentMachine)))
		fmt.Println()
		return nil
	}

	// Build all content in a single box
	var allLines []string

	// Header
	headerText := fmt.Sprintf("📊 Status: %s", currentMachine)
	if machine.Description != "" {
		headerText += fmt.Sprintf(" - %s", machine.Description)
	}
	header := lipgloss.NewStyle().
		Foreground(catLavender).
		Bold(true).
		Render(headerText)
	allLines = append(allLines, header)

	// Separator
	separator := lipgloss.NewStyle().
		Foreground(catOverlay0).
		Render(strings.Repeat("─", tableWidth-4))
	allLines = append(allLines, separator, "")

	// Machine info
	if machine.Hostname != "" {
		allLines = append(allLines, formatStatusLine("🖥", "Hostname", machine.Hostname, catText))
	}
	allLines = append(allLines, formatStatusLine("📄", "Brewfile", machine.Brewfile, catSubtext0))
	if cfg.DefaultSource != "" {
		allLines = append(allLines, formatStatusLine("🔗", "Source", cfg.DefaultSource, catText))
	}

	// Package counts
	packages, err := brewfile.Parse(machine.Brewfile)
	if err == nil {
		counts := formatPackageCountsCompact(packages)
		allLines = append(allLines, formatStatusLine("📦", "Packages", counts, catText))
	}

	// Metadata (if available)
	metaPath := filepath.Join(filepath.Dir(machine.Brewfile), ".brewsync-meta")
	meta, err := loadMetadata(metaPath)
	if err == nil && meta != nil {
		allLines = append(allLines, "")
		if !meta.LastDump.IsZero() {
			allLines = append(allLines, formatStatusLine("💾", "Last Dump", formatTimeAgo(meta.LastDump), catGreen))
		}
		if meta.LastSync != nil && !meta.LastSync.At.IsZero() {
			syncDetails := fmt.Sprintf("%s from %s", formatTimeAgo(meta.LastSync.At), meta.LastSync.From)
			if meta.LastSync.Added > 0 || meta.LastSync.Removed > 0 {
				syncDetails += fmt.Sprintf(" (+%d/-%d)", meta.LastSync.Added, meta.LastSync.Removed)
			}
			allLines = append(allLines, formatStatusLine("🔄", "Last Sync", syncDetails, catBlue))
		}
	}

	// Pending changes (if any)
	if cfg.DefaultSource != "" && cfg.DefaultSource != currentMachine && packages != nil {
		sourceMachine, ok := cfg.Machines[cfg.DefaultSource]
		if ok {
			sourcePackages, err := brewfile.Parse(sourceMachine.Brewfile)
			if err == nil {
				diff := brewfile.Diff(sourcePackages, packages)
				if !diff.IsEmpty() {
					allLines = append(allLines, "")
					pendingHeader := lipgloss.NewStyle().
						Foreground(catYellow).
						Bold(true).
						Render(fmt.Sprintf("⚡ Pending from %s", cfg.DefaultSource))
					allLines = append(allLines, pendingHeader)
					allLines = append(allLines, formatPendingCompact(diff))
				}
			}
		}
	}

	// Single status box
	statusBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(catOverlay0).
		Padding(1, 2).
		Width(tableWidth)

	fmt.Println()
	fmt.Println(statusBox.Render(strings.Join(allLines, "\n")))
	fmt.Println()

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

func formatStatusLine(icon, label, value string, color lipgloss.Color) string {
	iconStyled := lipgloss.NewStyle().
		Foreground(catMauve).
		Render(icon)

	labelStyled := lipgloss.NewStyle().
		Foreground(catMauve).
		Bold(true).
		Render(label)

	valueStyled := lipgloss.NewStyle().
		Foreground(color).
		Render(value)

	return fmt.Sprintf("%s  %s: %s", iconStyled, labelStyled, valueStyled)
}

func formatPackageCountsCompact(packages brewfile.Packages) string {
	byType := packages.ByType()

	total := len(packages)
	var counts []string

	if c := len(byType[brewfile.TypeBrew]); c > 0 {
		counts = append(counts, fmt.Sprintf("%d brew", c))
	}
	if c := len(byType[brewfile.TypeCask]); c > 0 {
		counts = append(counts, fmt.Sprintf("%d cask", c))
	}
	if c := len(byType[brewfile.TypeVSCode]); c > 0 {
		counts = append(counts, fmt.Sprintf("%d vscode", c))
	}
	if c := len(byType[brewfile.TypeCursor]); c > 0 {
		counts = append(counts, fmt.Sprintf("%d cursor", c))
	}
	if c := len(byType[brewfile.TypeAntigravity]); c > 0 {
		counts = append(counts, fmt.Sprintf("%d antigravity", c))
	}

	if len(counts) == 0 {
		return "none"
	}

	return fmt.Sprintf("%d total (%s)", total, strings.Join(counts, ", "))
}

func formatPendingCompact(diff *brewfile.DiffResult) string {
	var lines []string

	// Summary line
	addCount := len(diff.Additions)
	remCount := len(diff.Removals)

	var summaryParts []string
	if addCount > 0 {
		addText := lipgloss.NewStyle().
			Foreground(catGreen).
			Render(fmt.Sprintf("+%d to install", addCount))
		summaryParts = append(summaryParts, addText)
	}
	if remCount > 0 {
		remText := lipgloss.NewStyle().
			Foreground(catRed).
			Render(fmt.Sprintf("-%d to remove", remCount))
		summaryParts = append(summaryParts, remText)
	}

	lines = append(lines, strings.Join(summaryParts, ", "))

	// Type breakdown
	typeOrder := []brewfile.PackageType{
		brewfile.TypeTap,
		brewfile.TypeBrew,
		brewfile.TypeCask,
		brewfile.TypeVSCode,
		brewfile.TypeCursor,
		brewfile.TypeAntigravity,
		brewfile.TypeGo,
		brewfile.TypeMas,
	}

	typeIcons := map[brewfile.PackageType]string{
		brewfile.TypeTap:         "🚰",
		brewfile.TypeBrew:        "🍺",
		brewfile.TypeCask:        "📦",
		brewfile.TypeVSCode:      "💻",
		brewfile.TypeCursor:      "✏️",
		brewfile.TypeAntigravity: "🚀",
		brewfile.TypeGo:          "🔷",
		brewfile.TypeMas:         "🍎",
	}

	addByType := diff.AdditionsByType()
	remByType := diff.RemovalsByType()

	for _, t := range typeOrder {
		adds := len(addByType[t])
		rems := len(remByType[t])

		if adds == 0 && rems == 0 {
			continue
		}

		icon := typeIcons[t]
		var typeParts []string

		if adds > 0 {
			addText := lipgloss.NewStyle().
				Foreground(catGreen).
				Render(fmt.Sprintf("+%d", adds))
			typeParts = append(typeParts, addText)
		}
		if rems > 0 {
			remText := lipgloss.NewStyle().
				Foreground(catRed).
				Render(fmt.Sprintf("-%d", rems))
			typeParts = append(typeParts, remText)
		}

		line := fmt.Sprintf("  %s %s: %s", icon, t, strings.Join(typeParts, " "))
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}
