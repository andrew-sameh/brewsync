package brewfile

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Writer writes packages to Brewfile format
type Writer struct {
	packages Packages
}

// NewWriter creates a new Brewfile writer
func NewWriter(packages Packages) *Writer {
	return &Writer{packages: packages}
}

// Write writes the Brewfile to the given path
func (w *Writer) Write(path string) error {
	content := w.Format()
	return os.WriteFile(path, []byte(content), 0644)
}

// Format returns the Brewfile content as a string
func (w *Writer) Format() string {
	var sb strings.Builder

	// Group packages by type
	byType := w.packages.ByType()

	// Write in specific order
	typeOrder := []PackageType{TypeTap, TypeBrew, TypeCask, TypeMas, TypeVSCode, TypeCursor, TypeAntigravity, TypeGo}

	for _, t := range typeOrder {
		pkgs, ok := byType[t]
		if !ok || len(pkgs) == 0 {
			continue
		}

		// Sort packages by name
		sort.Slice(pkgs, func(i, j int) bool {
			return pkgs[i].Name < pkgs[j].Name
		})

		// Add section comment for non-standard types
		if t == TypeCursor || t == TypeAntigravity || t == TypeGo {
			sb.WriteString(fmt.Sprintf("\n# %s (brewsync extension)\n", t))
		} else if sb.Len() > 0 {
			sb.WriteString("\n")
		}

		for _, p := range pkgs {
			// Add description as a comment if available
			if p.Description != "" {
				sb.WriteString(fmt.Sprintf("# %s\n", p.Description))
			}
			sb.WriteString(formatPackage(p))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// formatPackage formats a single package entry
func formatPackage(p Package) string {
	switch p.Type {
	case TypeTap:
		return fmt.Sprintf(`tap "%s"`, p.Name)

	case TypeBrew:
		if len(p.Options) > 0 {
			return fmt.Sprintf(`brew "%s", %s`, p.Name, formatOptions(p.Options))
		}
		return fmt.Sprintf(`brew "%s"`, p.Name)

	case TypeCask:
		if len(p.Options) > 0 {
			return fmt.Sprintf(`cask "%s", %s`, p.Name, formatOptions(p.Options))
		}
		return fmt.Sprintf(`cask "%s"`, p.Name)

	case TypeMas:
		// mas entries need the id option
		if id, ok := p.Options["id"]; ok {
			name := p.FullName
			if name == "" {
				name = p.Name
			}
			return fmt.Sprintf(`mas "%s", id: %s`, name, id)
		}
		return fmt.Sprintf(`mas "%s"`, p.Name)

	case TypeVSCode:
		return fmt.Sprintf(`vscode "%s"`, p.Name)

	case TypeCursor:
		return fmt.Sprintf(`cursor "%s"`, p.Name)

	case TypeAntigravity:
		return fmt.Sprintf(`antigravity "%s"`, p.Name)

	case TypeGo:
		return fmt.Sprintf(`go "%s"`, p.Name)

	default:
		return fmt.Sprintf(`# unknown type: %s "%s"`, p.Type, p.Name)
	}
}

// formatOptions formats package options as Ruby hash syntax
func formatOptions(opts map[string]string) string {
	var parts []string
	for k, v := range opts {
		// Handle special cases
		if v == "true" || v == "false" {
			parts = append(parts, fmt.Sprintf("%s: %s", k, v))
		} else if strings.HasPrefix(v, ":") {
			// Symbol value
			parts = append(parts, fmt.Sprintf("%s: %s", k, v))
		} else if _, err := fmt.Sscanf(v, "%d", new(int)); err == nil {
			// Numeric value
			parts = append(parts, fmt.Sprintf("%s: %s", k, v))
		} else {
			// String value
			parts = append(parts, fmt.Sprintf(`%s: "%s"`, k, v))
		}
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

// Append adds packages to an existing Brewfile
func Append(path string, packages Packages) error {
	content := ""
	if data, err := os.ReadFile(path); err == nil {
		content = string(data)
	}

	// Add newline if file doesn't end with one
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	// Add new packages
	w := NewWriter(packages)
	content += w.Format()

	return os.WriteFile(path, []byte(content), 0644)
}
