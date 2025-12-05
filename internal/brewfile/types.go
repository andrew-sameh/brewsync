package brewfile

import (
	"fmt"
	"strings"
)

// PackageType represents the type of package
type PackageType string

const (
	TypeTap    PackageType = "tap"
	TypeBrew   PackageType = "brew"
	TypeCask   PackageType = "cask"
	TypeVSCode PackageType = "vscode"
	TypeCursor PackageType = "cursor"
	TypeGo     PackageType = "go"
	TypeMas    PackageType = "mas"
)

// AllTypes returns all package types
func AllTypes() []PackageType {
	return []PackageType{
		TypeTap,
		TypeBrew,
		TypeCask,
		TypeVSCode,
		TypeCursor,
		TypeGo,
		TypeMas,
	}
}

// ParsePackageType parses a string into a PackageType
func ParsePackageType(s string) (PackageType, error) {
	switch strings.ToLower(s) {
	case "tap":
		return TypeTap, nil
	case "brew":
		return TypeBrew, nil
	case "cask":
		return TypeCask, nil
	case "vscode":
		return TypeVSCode, nil
	case "cursor":
		return TypeCursor, nil
	case "go":
		return TypeGo, nil
	case "mas":
		return TypeMas, nil
	default:
		return "", fmt.Errorf("unknown package type: %s", s)
	}
}

// Package represents a single package entry
type Package struct {
	Type        PackageType       `json:"type" yaml:"type"`
	Name        string            `json:"name" yaml:"name"`
	FullName    string            `json:"full_name,omitempty" yaml:"full_name,omitempty"` // For mas: app name
	Options     map[string]string `json:"options,omitempty" yaml:"options,omitempty"`     // link: true, id: 123, etc.
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
}

// NewPackage creates a new package
func NewPackage(t PackageType, name string) Package {
	return Package{
		Type: t,
		Name: name,
	}
}

// WithOption adds an option to the package
func (p Package) WithOption(key, value string) Package {
	if p.Options == nil {
		p.Options = make(map[string]string)
	}
	p.Options[key] = value
	return p
}

// ID returns a unique identifier for the package
func (p Package) ID() string {
	return fmt.Sprintf("%s:%s", p.Type, p.Name)
}

// String returns a human-readable representation
func (p Package) String() string {
	if p.FullName != "" {
		return fmt.Sprintf("%s (%s)", p.Name, p.FullName)
	}
	return p.Name
}

// Packages is a collection of packages
type Packages []Package

// ByType groups packages by their type
func (ps Packages) ByType() map[PackageType][]Package {
	result := make(map[PackageType][]Package)
	for _, p := range ps {
		result[p.Type] = append(result[p.Type], p)
	}
	return result
}

// Filter returns packages matching the given types
func (ps Packages) Filter(types ...PackageType) Packages {
	if len(types) == 0 {
		return ps
	}

	typeSet := make(map[PackageType]bool)
	for _, t := range types {
		typeSet[t] = true
	}

	var result Packages
	for _, p := range ps {
		if typeSet[p.Type] {
			result = append(result, p)
		}
	}
	return result
}

// Names returns just the names of packages
func (ps Packages) Names() []string {
	names := make([]string, len(ps))
	for i, p := range ps {
		names[i] = p.Name
	}
	return names
}

// Contains checks if a package with the given ID exists
func (ps Packages) Contains(id string) bool {
	for _, p := range ps {
		if p.ID() == id {
			return true
		}
	}
	return false
}
