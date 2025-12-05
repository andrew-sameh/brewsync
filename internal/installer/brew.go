package installer

import (
	"strings"

	"github.com/andrew-sameh/brewsync/internal/brewfile"
	"github.com/andrew-sameh/brewsync/internal/exec"
)

// BrewInstaller handles Homebrew formulae and casks
type BrewInstaller struct {
	runner *exec.Runner
}

// NewBrewInstaller creates a new Homebrew installer
func NewBrewInstaller() *BrewInstaller {
	return &BrewInstaller{
		runner: exec.Default,
	}
}

// ListTaps returns all installed taps
func (b *BrewInstaller) ListTaps() (brewfile.Packages, error) {
	lines, err := b.runner.RunLines("brew", "tap")
	if err != nil {
		return nil, err
	}

	var packages brewfile.Packages
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			packages = append(packages, brewfile.NewPackage(brewfile.TypeTap, line))
		}
	}
	return packages, nil
}

// ListFormulae returns all installed formulae
func (b *BrewInstaller) ListFormulae() (brewfile.Packages, error) {
	lines, err := b.runner.RunLines("brew", "list", "--formula", "-1")
	if err != nil {
		return nil, err
	}

	var packages brewfile.Packages
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			packages = append(packages, brewfile.NewPackage(brewfile.TypeBrew, line))
		}
	}
	return packages, nil
}

// ListCasks returns all installed casks
func (b *BrewInstaller) ListCasks() (brewfile.Packages, error) {
	lines, err := b.runner.RunLines("brew", "list", "--cask", "-1")
	if err != nil {
		return nil, err
	}

	var packages brewfile.Packages
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			packages = append(packages, brewfile.NewPackage(brewfile.TypeCask, line))
		}
	}
	return packages, nil
}

// ListAll returns all taps, formulae, and casks
func (b *BrewInstaller) ListAll() (brewfile.Packages, error) {
	var all brewfile.Packages

	taps, err := b.ListTaps()
	if err != nil {
		return nil, err
	}
	all = append(all, taps...)

	formulae, err := b.ListFormulae()
	if err != nil {
		return nil, err
	}
	all = append(all, formulae...)

	casks, err := b.ListCasks()
	if err != nil {
		return nil, err
	}
	all = append(all, casks...)

	return all, nil
}

// List returns all installed packages (alias for ListAll)
func (b *BrewInstaller) List() (brewfile.Packages, error) {
	return b.ListAll()
}

// Install installs a package
func (b *BrewInstaller) Install(pkg brewfile.Package) error {
	switch pkg.Type {
	case brewfile.TypeTap:
		_, err := b.runner.Run("brew", "tap", pkg.Name)
		return err
	case brewfile.TypeBrew:
		_, err := b.runner.Run("brew", "install", pkg.Name)
		return err
	case brewfile.TypeCask:
		_, err := b.runner.Run("brew", "install", "--cask", pkg.Name)
		return err
	default:
		return nil
	}
}

// Uninstall removes a package
func (b *BrewInstaller) Uninstall(pkg brewfile.Package) error {
	switch pkg.Type {
	case brewfile.TypeTap:
		_, err := b.runner.Run("brew", "untap", pkg.Name)
		return err
	case brewfile.TypeBrew:
		_, err := b.runner.Run("brew", "uninstall", pkg.Name)
		return err
	case brewfile.TypeCask:
		_, err := b.runner.Run("brew", "uninstall", "--cask", pkg.Name)
		return err
	default:
		return nil
	}
}

// IsAvailable checks if brew is available
func (b *BrewInstaller) IsAvailable() bool {
	return b.runner.Exists("brew")
}

// DumpToFile runs brew bundle dump to a file
func (b *BrewInstaller) DumpToFile(path string) error {
	_, err := b.runner.Run("brew", "bundle", "dump", "--force", "--file="+path)
	return err
}
