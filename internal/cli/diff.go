package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/andrew-sameh/brewsync/internal/brewfile"
	"github.com/andrew-sameh/brewsync/internal/config"
)

var (
	diffFrom   string
	diffOnly   []string
	diffFormat string
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show differences between machines",
	Long: `Show differences between the current machine and a source machine.

Without arguments, compares with the default source machine.
Use --from to specify a different source machine.

Examples:
  brewsync diff                  # Compare with default source
  brewsync diff --from air       # Compare with specific machine
  brewsync diff --only brew,cask # Filter to specific types
  brewsync diff --format json    # Output as JSON`,
	RunE: runDiff,
}

func init() {
	diffCmd.Flags().StringVar(&diffFrom, "from", "", "source machine to compare with")
	diffCmd.Flags().StringSliceVar(&diffOnly, "only", nil, "only include these package types")
	diffCmd.Flags().StringVar(&diffFormat, "format", "table", "output format: table, json")
	rootCmd.AddCommand(diffCmd)
}

func runDiff(cmd *cobra.Command, args []string) error {
	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Determine source machine
	source := diffFrom
	if source == "" {
		source = cfg.DefaultSource
	}
	if source == "" {
		return fmt.Errorf("no source machine specified and no default_source in config")
	}

	// Get current machine
	currentMachine := cfg.CurrentMachine
	if currentMachine == "" {
		return fmt.Errorf("current machine not detected; run 'brewsync config init'")
	}

	// Can't diff with self
	if source == currentMachine {
		return fmt.Errorf("cannot diff machine with itself")
	}

	// Get source machine config
	sourceMachine, ok := cfg.Machines[source]
	if !ok {
		return fmt.Errorf("source machine '%s' not found in config", source)
	}

	// Get current machine config
	current, ok := cfg.Machines[currentMachine]
	if !ok {
		return fmt.Errorf("current machine '%s' not found in config", currentMachine)
	}

	printInfo("Comparing %s -> %s", source, currentMachine)

	// Parse source Brewfile
	printVerbose("Parsing source Brewfile: %s", sourceMachine.Brewfile)
	sourcePackages, err := brewfile.Parse(sourceMachine.Brewfile)
	if err != nil {
		return fmt.Errorf("failed to parse source Brewfile: %w", err)
	}

	// Parse current Brewfile
	printVerbose("Parsing current Brewfile: %s", current.Brewfile)
	currentPackages, err := brewfile.Parse(current.Brewfile)
	if err != nil {
		if os.IsNotExist(err) {
			currentPackages = brewfile.Packages{}
			printWarning("Current Brewfile not found, assuming empty")
		} else {
			return fmt.Errorf("failed to parse current Brewfile: %w", err)
		}
	}

	// Compute diff
	var diff *brewfile.DiffResult
	if len(diffOnly) > 0 {
		types := parsePackageTypes(diffOnly)
		diff = brewfile.DiffByType(sourcePackages, currentPackages, types)
	} else {
		diff = brewfile.Diff(sourcePackages, currentPackages)
	}

	// Output results
	switch diffFormat {
	case "json":
		return outputDiffJSON(diff)
	default:
		return outputDiffTable(diff, source, currentMachine)
	}
}

func parsePackageTypes(types []string) []brewfile.PackageType {
	var result []brewfile.PackageType
	for _, t := range types {
		switch strings.ToLower(strings.TrimSpace(t)) {
		case "tap":
			result = append(result, brewfile.TypeTap)
		case "brew":
			result = append(result, brewfile.TypeBrew)
		case "cask":
			result = append(result, brewfile.TypeCask)
		case "vscode":
			result = append(result, brewfile.TypeVSCode)
		case "cursor":
			result = append(result, brewfile.TypeCursor)
		case "go":
			result = append(result, brewfile.TypeGo)
		case "mas":
			result = append(result, brewfile.TypeMas)
		}
	}
	return result
}

func outputDiffJSON(diff *brewfile.DiffResult) error {
	output := map[string]interface{}{
		"additions": packageNames(diff.Additions),
		"removals":  packageNames(diff.Removals),
		"common":    len(diff.Common),
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

func packageNames(pkgs brewfile.Packages) map[string][]string {
	result := make(map[string][]string)
	for _, pkg := range pkgs {
		typeName := string(pkg.Type)
		result[typeName] = append(result[typeName], pkg.Name)
	}
	return result
}

func outputDiffTable(diff *brewfile.DiffResult, source, current string) error {
	if diff.IsEmpty() {
		printInfo("No differences between %s and %s", source, current)
		return nil
	}

	// Print additions
	if len(diff.Additions) > 0 {
		fmt.Printf("\nTo be installed (from %s): %d packages\n", source, len(diff.Additions))
		printPackagesByType(diff.Additions, "+")
	}

	// Print removals
	if len(diff.Removals) > 0 {
		fmt.Printf("\nTo be removed (not in %s): %d packages\n", source, len(diff.Removals))
		printPackagesByType(diff.Removals, "-")
	}

	// Summary
	fmt.Printf("\nSummary: %s\n", diff.Summary())

	return nil
}

func printPackagesByType(pkgs brewfile.Packages, prefix string) {
	byType := pkgs.ByType()
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

		fmt.Printf("  %s (%d):\n", t, len(typePkgs))
		for _, pkg := range typePkgs {
			fmt.Printf("    %s %s\n", prefix, pkg.Name)
		}
	}
}
