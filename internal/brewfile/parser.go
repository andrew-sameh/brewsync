package brewfile

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Parser parses Brewfile format files
type Parser struct{}

// NewParser creates a new Parser
func NewParser() *Parser {
	return &Parser{}
}

// Patterns for parsing Brewfile lines
var (
	// Match: tap "name" or tap "name", args
	tapPattern = regexp.MustCompile(`^tap\s+"([^"]+)"(?:\s*,\s*(.+))?`)
	// Match: brew "name" or brew "name", options
	brewPattern = regexp.MustCompile(`^brew\s+"([^"]+)"(?:\s*,\s*(.+))?`)
	// Match: cask "name" or cask "name", options
	caskPattern = regexp.MustCompile(`^cask\s+"([^"]+)"(?:\s*,\s*(.+))?`)
	// Match: mas "name", id: 123
	masPattern = regexp.MustCompile(`^mas\s+"([^"]+)"(?:\s*,\s*(.+))?`)
	// Match: vscode "name"
	vscodePattern = regexp.MustCompile(`^vscode\s+"([^"]+)"`)
	// Match: cursor "name" (BrewSync extension)
	cursorPattern = regexp.MustCompile(`^cursor\s+"([^"]+)"`)
	// Match: antigravity "name" (BrewSync extension)
	antigravityPattern = regexp.MustCompile(`^antigravity\s+"([^"]+)"`)
	// Match: go "name" (BrewSync extension)
	goPattern = regexp.MustCompile(`^go\s+"([^"]+)"`)
	// Match options like: link: true, args: ["--foo"]
	optionPattern = regexp.MustCompile(`(\w+):\s*(.+?)(?:,\s*|$)`)
)

// ParseFile parses a Brewfile from the given path
func (p *Parser) ParseFile(path string) (Packages, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var packages Packages
	scanner := bufio.NewScanner(file)
	lineNum := 0
	var lastComment string // Track comment from previous line

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Capture comments as potential descriptions
		if strings.HasPrefix(line, "#") {
			// Extract comment text (remove leading # and whitespace)
			lastComment = strings.TrimSpace(strings.TrimPrefix(line, "#"))
			continue
		}

		pkg, ok := p.parseLine(line)
		if !ok {
			// Skip lines that don't match any pattern
			lastComment = "" // Reset if we skip a line
			continue
		}

		// Attach the last comment as description if available
		if lastComment != "" {
			pkg.Description = lastComment
			lastComment = "" // Reset after using
		}

		packages = append(packages, pkg)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return packages, nil
}

// ParseString parses Brewfile content from a string
func (p *Parser) ParseString(content string) (Packages, error) {
	var packages Packages
	scanner := bufio.NewScanner(strings.NewReader(content))
	var lastComment string // Track comment from previous line

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Capture comments as potential descriptions
		if strings.HasPrefix(line, "#") {
			// Extract comment text (remove leading # and whitespace)
			lastComment = strings.TrimSpace(strings.TrimPrefix(line, "#"))
			continue
		}

		pkg, ok := p.parseLine(line)
		if !ok {
			lastComment = "" // Reset if we skip a line
			continue
		}

		// Attach the last comment as description if available
		if lastComment != "" {
			pkg.Description = lastComment
			lastComment = "" // Reset after using
		}

		packages = append(packages, pkg)
	}

	return packages, nil
}

// parseLine parses a single Brewfile line
func (p *Parser) parseLine(line string) (Package, bool) {
	// Try each pattern in order
	if matches := tapPattern.FindStringSubmatch(line); matches != nil {
		pkg := NewPackage(TypeTap, matches[1])
		if len(matches) > 2 && matches[2] != "" {
			pkg = p.parseOptions(matches[2], pkg)
		}
		return pkg, true
	}

	if matches := brewPattern.FindStringSubmatch(line); matches != nil {
		pkg := NewPackage(TypeBrew, matches[1])
		if len(matches) > 2 && matches[2] != "" {
			pkg = p.parseOptions(matches[2], pkg)
		}
		return pkg, true
	}

	if matches := caskPattern.FindStringSubmatch(line); matches != nil {
		pkg := NewPackage(TypeCask, matches[1])
		if len(matches) > 2 && matches[2] != "" {
			pkg = p.parseOptions(matches[2], pkg)
		}
		return pkg, true
	}

	if matches := masPattern.FindStringSubmatch(line); matches != nil {
		pkg := NewPackage(TypeMas, matches[1])
		if len(matches) > 2 && matches[2] != "" {
			pkg = p.parseOptions(matches[2], pkg)
		}
		return pkg, true
	}

	if matches := vscodePattern.FindStringSubmatch(line); matches != nil {
		return NewPackage(TypeVSCode, matches[1]), true
	}

	if matches := cursorPattern.FindStringSubmatch(line); matches != nil {
		return NewPackage(TypeCursor, matches[1]), true
	}

	if matches := antigravityPattern.FindStringSubmatch(line); matches != nil {
		return NewPackage(TypeAntigravity, matches[1]), true
	}

	if matches := goPattern.FindStringSubmatch(line); matches != nil {
		return NewPackage(TypeGo, matches[1]), true
	}

	return Package{}, false
}

// parseOptions parses option string like "link: true, args: [\"--foo\"]"
func (p *Parser) parseOptions(optStr string, pkg Package) Package {
	if pkg.Options == nil {
		pkg.Options = make(map[string]string)
	}
	// Handle simple key: value patterns
	matches := optionPattern.FindAllStringSubmatch(optStr, -1)
	for _, match := range matches {
		if len(match) >= 3 {
			key := match[1]
			value := strings.TrimSpace(match[2])
			// Remove trailing comma if present
			value = strings.TrimSuffix(value, ",")
			// Store the option as string
			pkg.Options[key] = p.parseValueAsString(value)
		}
	}
	return pkg
}

// parseValueAsString converts a value to string representation
func (p *Parser) parseValueAsString(value string) string {
	// Remove quotes if present
	value = strings.Trim(value, `"'`)
	return value
}

// Parse is a convenience function to parse a file
func Parse(path string) (Packages, error) {
	return NewParser().ParseFile(path)
}

// ParseContent is a convenience function to parse content string
func ParseContent(content string) (Packages, error) {
	return NewParser().ParseString(content)
}
