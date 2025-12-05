package brewfile

import "strconv"

// DiffResult contains the results of comparing two package lists
type DiffResult struct {
	// Additions are packages in source but not in current
	Additions Packages
	// Removals are packages in current but not in source
	Removals Packages
	// Common are packages in both
	Common Packages
}

// IsEmpty returns true if there are no differences
func (d *DiffResult) IsEmpty() bool {
	return len(d.Additions) == 0 && len(d.Removals) == 0
}

// AdditionsByType returns additions grouped by package type
func (d *DiffResult) AdditionsByType() map[PackageType][]Package {
	return d.Additions.ByType()
}

// RemovalsByType returns removals grouped by package type
func (d *DiffResult) RemovalsByType() map[PackageType][]Package {
	return d.Removals.ByType()
}

// Diff computes the differences between source and current package lists
// source: the packages we want to have (e.g., from another machine)
// current: the packages we currently have
func Diff(source, current Packages) *DiffResult {
	result := &DiffResult{
		Additions: make(Packages, 0),
		Removals:  make(Packages, 0),
		Common:    make(Packages, 0),
	}

	// Build a map of current packages for quick lookup
	currentMap := make(map[string]Package)
	for _, pkg := range current {
		key := packageKey(pkg)
		currentMap[key] = pkg
	}

	// Build a map of source packages
	sourceMap := make(map[string]Package)
	for _, pkg := range source {
		key := packageKey(pkg)
		sourceMap[key] = pkg
	}

	// Find additions (in source but not in current)
	for _, pkg := range source {
		key := packageKey(pkg)
		if _, exists := currentMap[key]; !exists {
			result.Additions = append(result.Additions, pkg)
		} else {
			result.Common = append(result.Common, pkg)
		}
	}

	// Find removals (in current but not in source)
	for _, pkg := range current {
		key := packageKey(pkg)
		if _, exists := sourceMap[key]; !exists {
			result.Removals = append(result.Removals, pkg)
		}
	}

	return result
}

// DiffByType computes differences filtered to specific package types
func DiffByType(source, current Packages, types []PackageType) *DiffResult {
	// Filter both lists to only include specified types
	filteredSource := source.Filter(types...)
	filteredCurrent := current.Filter(types...)

	return Diff(filteredSource, filteredCurrent)
}

// packageKey returns a unique key for a package based on type and name
func packageKey(pkg Package) string {
	return string(pkg.Type) + ":" + pkg.Name
}

// FilterIgnored removes packages from a diff result that should be ignored
func (d *DiffResult) FilterIgnored(ignoredPackages map[string]bool) *DiffResult {
	return &DiffResult{
		Additions: filterByKey(d.Additions, ignoredPackages),
		Removals:  filterByKey(d.Removals, ignoredPackages),
		Common:    d.Common,
	}
}

// FilterMachineSpecific removes packages that are designated for a specific machine
func (d *DiffResult) FilterMachineSpecific(machinePackages map[string]bool) *DiffResult {
	return &DiffResult{
		Additions: filterByKey(d.Additions, machinePackages),
		Removals:  filterByKey(d.Removals, machinePackages),
		Common:    d.Common,
	}
}

// filterByKey filters out packages whose keys are in the excluded map
func filterByKey(pkgs Packages, excluded map[string]bool) Packages {
	var result Packages
	for _, pkg := range pkgs {
		key := packageKey(pkg)
		if !excluded[key] {
			result = append(result, pkg)
		}
	}
	return result
}

// Summary returns a human-readable summary of the diff
func (d *DiffResult) Summary() string {
	if d.IsEmpty() {
		return "No differences"
	}

	var parts []string
	if len(d.Additions) > 0 {
		parts = append(parts, formatCount(len(d.Additions), "addition"))
	}
	if len(d.Removals) > 0 {
		parts = append(parts, formatCount(len(d.Removals), "removal"))
	}

	return join(parts, ", ")
}

// formatCount formats a count with singular/plural noun
func formatCount(count int, noun string) string {
	if count == 1 {
		return "1 " + noun
	}
	return strconv.Itoa(count) + " " + noun + "s"
}

// join joins strings with a separator
func join(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}
