package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/andrew-sameh/brewsync/internal/brewfile"
	"github.com/andrew-sameh/brewsync/internal/config"
)

var (
	listFrom   string
	listOnly   []string
	listFormat string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List packages in a Brewfile",
	Long: `List packages in a machine's Brewfile.

Without arguments, lists packages from the current machine.
Use --from to list from a different machine.

Examples:
  brewsync list                  # Current machine
  brewsync list --from mini      # Another machine
  brewsync list --only brew      # Filter by type
  brewsync list --format json    # JSON output`,
	RunE: runList,
}

func init() {
	listCmd.Flags().StringVar(&listFrom, "from", "", "machine to list packages from")
	listCmd.Flags().StringSliceVar(&listOnly, "only", nil, "only include these package types")
	listCmd.Flags().StringVar(&listFormat, "format", "table", "output format: table, json")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Determine which machine to list
	machineName := listFrom
	if machineName == "" {
		machineName = cfg.CurrentMachine
	}
	if machineName == "" {
		return fmt.Errorf("no machine specified and current machine not detected")
	}

	// Get machine config
	machine, ok := cfg.Machines[machineName]
	if !ok {
		return fmt.Errorf("machine '%s' not found in config", machineName)
	}

	printVerbose("Reading Brewfile: %s", machine.Brewfile)

	// Parse Brewfile
	packages, err := brewfile.Parse(machine.Brewfile)
	if err != nil {
		if os.IsNotExist(err) {
			printInfo("No Brewfile found at %s", machine.Brewfile)
			return nil
		}
		return fmt.Errorf("failed to parse Brewfile: %w", err)
	}

	// Filter by type if specified
	if len(listOnly) > 0 {
		types := parsePackageTypes(listOnly)
		packages = packages.Filter(types...)
	}

	// Output results
	switch listFormat {
	case "json":
		return outputListJSON(packages, machineName)
	default:
		return outputListTable(packages, machineName)
	}
}

func outputListJSON(packages brewfile.Packages, machine string) error {
	output := map[string]interface{}{
		"machine":  machine,
		"packages": packageNames(packages),
		"counts":   packageCounts(packages),
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

func packageCounts(pkgs brewfile.Packages) map[string]int {
	counts := make(map[string]int)
	for _, pkg := range pkgs {
		counts[string(pkg.Type)]++
	}
	return counts
}

func outputListTable(packages brewfile.Packages, machine string) error {
	if len(packages) == 0 {
		printInfo("No packages found for %s", machine)
		return nil
	}

	fmt.Printf("Packages for %s: %d total\n\n", machine, len(packages))

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

	for _, t := range typeOrder {
		typePkgs := byType[t]
		if len(typePkgs) == 0 {
			continue
		}

		fmt.Printf("%s (%d):\n", t, len(typePkgs))
		for _, pkg := range typePkgs {
			fmt.Printf("  %s\n", pkg.Name)
		}
		fmt.Println()
	}

	return nil
}
