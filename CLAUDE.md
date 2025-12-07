# BrewSync

A CLI tool to sync Homebrew packages, casks, taps, VSCode/Cursor extensions, Go tools, and Mac App Store apps across multiple macOS machines.

## Design Philosophy

- **Machine-centric**: Machines are the source of truth (actual installed state)
- **Git-based**: Designed for dotfiles workflows, not cloud storage
- **Profiles as overlays**: Optional curated package groups for convenience
- **Interactive by default**: User picks what to install, with batch options
- **Non-destructive by default**: Import mode only adds, never removes

---

## Core Concepts

### Machines

Physical macOS computers identified by hostname. Each machine has:
- A unique identifier (e.g., `mini`, `air`)
- A hostname for auto-detection (e.g., `Andrews-Mac-mini`)
- A Brewfile path (e.g., `/Users/andrew/dotfiles/_brew_mini/Brewfile`)

### Profiles

Optional curated groups of packages (e.g., `core`, `dev-go`, `k8s`). Profiles are:
- Manually curated (not auto-generated)
- Additive only (install, never remove)
- Cross-machine (not tied to specific hardware)
- Composable (`--profile core,dev-go,k8s`)

### Package Types

| Type | Source | Example |
|------|--------|---------|
| `tap` | Homebrew taps | `charmbracelet/tap` |
| `brew` | Homebrew formulae | `git`, `fzf`, `bat` |
| `cask` | Homebrew casks | `raycast`, `slack` |
| `vscode` | VSCode extensions | `golang.go` |
| `cursor` | Cursor extensions | `ms-python.python` |
| `antigravity` | Antigravity editor extensions | `python.lsp` |
| `go` | Go tools | `golang.org/x/tools/gopls` |
| `mas` | Mac App Store | `497799835` (Xcode) |

### Ignore System

Ignores are stored in a separate `~/.config/brewsync/ignore.yaml` file with two explicit layers:

**1. Categories** - Ignore entire package types:
- Stored in `categories: [...]` list
- Examples: `mas`, `go`, `antigravity`
- Effect: ALL packages of that type are ignored

**2. Packages** - Ignore specific packages within non-ignored categories:
- Stored in `packages:` with type-specific lists
- Format: `cask: ["app1", "app2"]`
- Effect: Only specified packages are ignored

**How to add ignores:**
- Interactive selection: Press `i` in TUI
- Category command: `brewsync ignore category add mas`
- Package command: `brewsync ignore add cask:bluestacks`
- Direct file edit: Edit `~/.config/brewsync/ignore.yaml`

**Scope**: Global (all machines) or per-machine

**Important**: Ignore lists apply to `import`, `sync`, and `diff` operations but **not** to `dump`. The dump command captures everything installed (source of truth), regardless of ignore lists.

### Machine-Specific Packages

Packages designated for specific machines only:
- Won't be suggested for other machines during import
- Won't be removed during sync on designated machine
- Explicit opt-in with `--include-machine-specific`

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI Interface                            │
│  (commands, flags, interactive prompts)                          │
└─────────────────────────────┬───────────────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────────────┐
│                      Command Router                              │
│  import | sync | diff | dump | status | config | ignore | ...   │
└─────────────────────────────┬───────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌───────────────┐   ┌─────────────────┐   ┌─────────────────┐
│ Config Manager │   │ Brewfile Parser │   │ Package Manager │
│ - Load/save   │   │ - Parse files   │   │ - brew install  │
│ - Validate    │   │ - Compare/diff  │   │ - brew uninstall│
│ - Machine     │   │ - Merge/filter  │   │ - mas/go install│
│   detection   │   │                 │   │                 │
└───────┬───────┘   └────────┬────────┘   └────────┬────────┘
        │                    │                     │
        ▼                    ▼                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Data Layer                                │
│  ~/.config/brewsync/          ~/dotfiles/ (user repo)           │
│  ├── config.yaml              ├── _brew_mini/Brewfile           │
│  ├── history.log              ├── _brew_air/Brewfile            │
│  └── profiles/*.yaml          └── ...                           │
└─────────────────────────────────────────────────────────────────┘
```

### Data Flows

**Import Flow**: Load config → Parse source & current Brewfiles → Compute diff → Filter ignored/machine-specific → Interactive selection → Execute installations → Log to history

**Sync Flow**: Load config → Parse Brewfiles → Compute additions & removals → Filter → Preview (dry-run default) → On `--apply`: execute changes → Log & dump

**Dump Flow**: Run `brew bundle dump --describe` (default) → Parse with descriptions → Deduplicate and append VSCode/Cursor/Antigravity/Go/mas extensions → Write Brewfile → Update metadata → Optionally commit/push

---

## File Structure

### Config Directory (`~/.config/brewsync/`)

```
~/.config/brewsync/
├── config.yaml           # Main configuration
├── ignore.yaml           # Ignore rules (categories + packages)
├── history.log           # Operation history (append-only)
└── profiles/             # User-defined profiles
    ├── core.yaml
    └── dev-go.yaml
```

### Dotfiles Structure (User's Repo)

```
~/dotfiles/
├── _brew_mini/
│   ├── Brewfile          # Machine's package list
│   └── .brewsync-meta    # Metadata (last dump, counts)
└── _brew_air/
    ├── Brewfile
    └── .brewsync-meta
```

---

## Configuration

### Main Config (`~/.config/brewsync/config.yaml`)

```yaml
machines:
  mini:
    hostname: "Andrews-Mac-mini"
    brewfile: "/Users/andrew/dotfiles/_brew_mini/Brewfile"
    description: "Mac Mini - primary workstation"
  air:
    hostname: "Andrews-MacBook-Air"
    brewfile: "/Users/andrew/dotfiles/_brew_air/Brewfile"
    description: "MacBook Air - portable"

current_machine: auto  # or explicit: "mini", "air"
default_source: mini
default_categories: [tap, brew, cask, vscode, cursor, antigravity, go, mas]

auto_dump:
  enabled: false
  after_install: false
  commit: false
  push: false
  commit_message: "brewsync: update {machine} Brewfile"

dump:
  use_brew_bundle: true  # Use 'brew bundle dump --describe' for Homebrew packages (includes descriptions)

machine_specific:
  mini:
    brew: ["postgresql@16", "redis"]
    cask: ["orbstack"]
    antigravity: []
  air:
    brew: ["ollama"]
    cask: ["ollama-app"]
    antigravity: []

conflict_resolution: ask  # ask | skip | source-wins | current-wins

output:
  color: true
  verbose: false
  show_descriptions: true

hooks:
  pre_install: ""
  post_install: ""
  pre_dump: ""
  post_dump: ""
```

### Ignore File (`~/.config/brewsync/ignore.yaml`)

```yaml
# Global ignores (apply to all machines)
global:
  categories:
    - mas           # Ignore ALL Mac App Store apps
    - go            # Ignore ALL Go tools

  packages:
    tap: []
    brew: []
    cask: ["company-vpn"]
    vscode: []
    cursor: []
    antigravity: []
    go: []         # Individual packages (if category not ignored)
    mas: []

# Machine-specific ignores
machines:
  mini:
    categories:
      - antigravity  # Don't use Antigravity on mini

    packages:
      tap: []
      brew: []
      cask: ["bluestacks"]
      vscode: []
      cursor: []
      antigravity: []
      go: []
      mas: []

  air:
    categories: []

    packages:
      tap: []
      brew: ["scrcpy"]
      cask: []
      vscode: []
      cursor: []
      antigravity: []
      go: []
      mas: []
```

### Profile File (`profiles/core.yaml`)

```yaml
name: core
description: "Essential tools for any machine"
packages:
  tap: ["homebrew/bundle"]
  brew: [git, fzf, bat, eza, fd, ripgrep, lazygit, starship, zoxide]
  cask: [raycast, hiddenbar, stats, iterm2]
  vscode: [vscodevim.vim, eamodio.gitlens]
```

### Metadata File (`.brewsync-meta`)

```yaml
machine: mini
last_dump: "2024-01-16T10:30:00Z"
last_sync:
  from: air
  at: "2024-01-15T14:20:00Z"
  added: 5
  removed: 0
package_counts:
  tap: 6
  brew: 85
  cask: 42
  vscode: 120
  cursor: 45
  antigravity: 12
  go: 9
  mas: 3
macos_version: "14.2"
brewsync_version: "1.0.0"
```

---

## Package Descriptions

The `dump.use_brew_bundle` configuration controls how package descriptions are captured and stored in Brewfiles.

### Default Behavior (use_brew_bundle: true)

- Uses `brew bundle dump --describe` to capture Homebrew descriptions
- Descriptions stored as comments in Brewfile (# comment above each package)
- Applied to: tap, brew, cask packages (from Homebrew database)
- VSCode, Cursor, Antigravity, Go, and mas packages added afterward with deduplication
- Performance: ~1-2 seconds for 100+ packages

**Example output**:
```ruby
# Clone of cat(1) with syntax highlighting and Git integration
brew "bat"
# Distributed revision control system
brew "git"
# Launcher and productivity tool
cask "raycast"
```

### Manual Collection (use_brew_bundle: false)

- Collects packages via individual `brew list` commands
- No descriptions included (basic list only)
- Useful for custom workflows or older Homebrew versions
- Performance: ~5-10 seconds for 100+ packages

### Deduplication Behavior

When collecting extensions (VSCode, Cursor, Antigravity, Go, mas):
- Uses `AddUnique()` to prevent duplicates with brew bundle output
- Verbose output shows: `"(X new, Y already in Brewfile)"`
- If `brew bundle dump` already includes a package (e.g., mas app, vscode extension installed via Homebrew), manual collection skips it
- Existing packages with descriptions are preserved

**Example verbose output**:
```
Found 45 extensions (3 new, 42 already in Brewfile)
```

---

## Commands

### Core Operations

```bash
# Import - install missing packages from another machine
brewsync import                          # From default source, interactive
brewsync import --from air               # From specific machine
brewsync import --only brew,cask         # Filter categories
brewsync import --yes                    # Install all without prompts
brewsync import --dry-run                # Preview only

# Sync - make current machine match source exactly (adds AND removes)
brewsync sync                            # Preview mode (dry-run)
brewsync sync --apply                    # Execute changes
brewsync sync --from air --only brew     # Specific source and categories

# Diff - show differences without changes
brewsync diff                            # Compare with default source
brewsync diff --format json              # JSON output

# Dump - update Brewfile from installed packages
brewsync dump                            # Update Brewfile
brewsync dump --commit --push            # Commit and push changes
```

### Profile Operations

```bash
brewsync profile list                    # List profiles
brewsync profile show core               # Show profile contents
brewsync profile install core,dev-go     # Install from profiles
brewsync profile create web-dev --pick   # Create interactively
```

### Status & Info

```bash
brewsync status                          # Current state overview
brewsync list                            # List packages in Brewfile
brewsync history                         # View operation history
brewsync doctor                          # Validate setup
```

### Configuration

```bash
brewsync config init                     # Interactive setup wizard
brewsync config show                     # Display current config
brewsync config edit                     # Open in $EDITOR
brewsync config set default_source air   # Set config value
brewsync config add-machine dev --hostname "Dev-Mac" --brewfile "/path/to/Brewfile"

# Ignore commands (two-layer system)
brewsync ignore category add mas                    # Ignore ALL mas packages globally
brewsync ignore category add go --machine mini      # Ignore ALL go tools on mini
brewsync ignore category remove mas                 # Remove category ignore
brewsync ignore category list                       # List ignored categories

brewsync ignore add cask:app                        # Add package ignore (current machine)
brewsync ignore add cask:app --global               # Add package ignore (global)
brewsync ignore add cask:app --machine mini         # Add package ignore (specific machine)
brewsync ignore remove cask:app                     # Remove package ignore
brewsync ignore list                                # Show all ignores (categories + packages)
brewsync ignore path                                # Show ignore file location
brewsync ignore init                                # Create default ignore.yaml
```

### Global Flags

```bash
--dry-run           # Preview without executing
--verbose, -v       # Detailed output
--quiet, -q         # Minimal output
--no-color          # Disable colored output
--yes, -y           # Skip confirmations
--config <path>     # Use alternate config file
```

---

## Interactive UI

```
┌─────────────────────────────────────────────────────────────────┐
│ BrewSync Import - Missing packages from [mini]                  │
├─────────────────────────────────────────────────────────────────┤
│ TAPS (2)                                                        │
│   [ ] charmbracelet/tap                                         │
│   [x] homebrew/bundle                                           │
│                                                                 │
│ BREWS (15)                                                      │
│   [x] bat                  Cat clone with syntax highlighting   │
│   [ ] croc                 Securely send things between computers│
│   ... (13 more)                                                 │
│                                                                 │
│ CASKS (8)                                                       │
│   [x] raycast              Launcher and productivity tool       │
│   ... (7 more)                                                  │
├─────────────────────────────────────────────────────────────────┤
│ [space] toggle  [a] all  [n] none  [enter] install  [q] quit   │
│ [t] taps  [b] brews  [c] casks  [v] vscode  [g] go  [m] mas    │
│ [i] ignore selected  [s] save as profile  [/] search           │
└─────────────────────────────────────────────────────────────────┘
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Nothing to do |
| 3 | User cancelled |
| 4 | Config error |
| 5 | Brewfile not found |
| 6 | Machine not recognized |

---

## History Log Format

Append-only log at `~/.config/brewsync/history.log`:

```
2024-01-16T10:30:00Z|dump|mini|tap:6,brew:85,cask:42|committed
2024-01-15T14:20:00Z|import|mini←air|+brew:bat,eza;+cask:raycast|5 packages
2024-01-14T09:00:00Z|ignore|mini|+cask:bluestacks|added to ignore
```

---

## Error Handling

| Error | Cause | Recovery |
|-------|-------|----------|
| "Machine not recognized" | Hostname doesn't match config | Run `brewsync config init` |
| "Brewfile not found" | Path in config is wrong | `brewsync config edit` |
| "brew command failed" | Package doesn't exist or network issue | Check package name, retry |
| "Permission denied" | Can't write to Brewfile path | Check file permissions |

All destructive operations support `--dry-run` for preview.

---

## Technical Stack

### Language: Go

Single binary distribution, excellent CLI ecosystem (Charm libraries), fast execution, cross-compilation.

### Libraries

| Library | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | Command structure, flag parsing, help generation |
| `github.com/spf13/viper` | Config loading, env vars, flag binding |
| `github.com/charmbracelet/bubbletea` | Interactive TUI framework |
| `github.com/charmbracelet/bubbles` | TUI components (list, spinner, progress) |
| `github.com/charmbracelet/lipgloss` | Terminal styling and layout |
| `github.com/charmbracelet/huh` | Forms and prompts |
| `github.com/charmbracelet/log` | Structured logging |
| `github.com/sahilm/fuzzy` | Fuzzy search in package selection |
| `gopkg.in/yaml.v3` | YAML parsing for config/profiles |

### Project Structure (Actual)

```
brewsync/
├── cmd/brewsync/main.go           # Entry point - calls cli.Execute()
├── internal/
│   ├── cli/                       # Cobra commands (all register in init())
│   │   ├── root.go                # Root command, global flags (dryRun, verbose, noColor, assumeYes)
│   │   ├── import.go              # Import with TUI selection
│   │   ├── sync.go                # Sync with preview/apply modes
│   │   ├── diff.go                # Diff between machines
│   │   ├── dump.go                # Dump installed packages to Brewfile
│   │   ├── list.go                # List packages from Brewfile
│   │   ├── status.go              # Show machine status
│   │   ├── doctor.go              # Validate setup
│   │   ├── history.go             # View operation history
│   │   ├── profile.go             # Profile subcommands
│   │   ├── config.go              # Config subcommands (uses huh forms)
│   │   └── ignore.go              # Ignore list management
│   ├── config/                    # Configuration (Viper-based)
│   │   ├── config.go              # Load(), Get(), Init(), ConfigPath()
│   │   ├── types.go               # Config, Machine, IgnoreConfig structs + helper methods
│   │   ├── machine.go             # DetectMachine(), GetLocalHostname()
│   │   └── defaults.go            # Default values, DefaultCategories
│   ├── brewfile/                  # Brewfile parsing
│   │   ├── types.go               # Package, Packages, PackageType, DiffResult
│   │   ├── parser.go              # Parse() - reads Brewfile
│   │   ├── writer.go              # Write() - writes Brewfile
│   │   └── diff.go                # Diff() - compares two Packages
│   ├── profile/
│   │   └── profile.go             # Profile struct, Load(), Save(), MergePackages()
│   ├── installer/                 # Package installation
│   │   ├── installer.go           # Manager orchestrator, InstallMany(), UninstallMany()
│   │   ├── brew.go                # BrewInstaller - taps, formulae, casks
│   │   ├── vscode.go              # VSCodeInstaller
│   │   ├── cursor.go              # CursorInstaller
│   │   ├── antigravity.go         # AntigravityInstaller
│   │   ├── mas.go                 # MasInstaller (Mac App Store)
│   │   └── gotools.go             # GoToolsInstaller
│   ├── tui/                       # Bubble Tea UI components
│   │   ├── styles/styles.go       # Shared lipgloss styles and colors
│   │   ├── selection/             # Package selection TUI
│   │   │   ├── keys.go            # KeyMap with all keybindings
│   │   │   ├── model.go           # Bubble Tea Model, Init/Update/View
│   │   │   └── view.go            # Rendering logic
│   │   └── progress/
│   │       └── model.go           # Installation progress UI
│   ├── history/
│   │   └── history.go             # Log(), LogDump(), LogImport(), LogSync(), Read()
│   └── exec/
│       └── exec.go                # Run() - executes shell commands safely
└── pkg/version/
    └── version.go                 # Version, Commit, Date variables
```

---

## Key Types

### brewfile.Package
```go
type Package struct {
    Type        PackageType       // tap, brew, cask, vscode, cursor, antigravity, go, mas
    Name        string            // Package identifier
    FullName    string            // For mas: app name
    Options     map[string]string // link: true, id: 123, etc.
    Description string            // From brew bundle dump --describe
}
func (p Package) ID() string  // Returns "type:name"
```

### brewfile.DiffResult
```go
type DiffResult struct {
    Additions Packages  // In source but not current
    Removals  Packages  // In current but not source
    Common    Packages  // In both
}
```

### config.Config
```go
type Config struct {
    Machines           map[string]Machine
    CurrentMachine     string              // Auto-detected or explicit
    DefaultSource      string              // For import/diff
    DefaultCategories  []string
    Ignore             IgnoreConfig
    MachineSpecific    MachineSpecificConfig
    // ...
}
func (c *Config) GetIgnoredPackages(machine string) []string
func (c *Config) GetMachineSpecificPackages() map[string][]string
```

### installer.Manager
```go
type Manager struct { /* contains type-specific installers */ }
func NewManager() *Manager
func (m *Manager) Install(pkg brewfile.Package) error
func (m *Manager) Uninstall(pkg brewfile.Package) error
func (m *Manager) InstallMany(pkgs Packages, onProgress func(pkg, i, total, err))
func (m *Manager) UninstallMany(pkgs Packages, onProgress func(pkg, i, total, err))
```

---

## TUI Architecture

The TUI uses Bubble Tea (Elm architecture):

**selection.Model** - Interactive package picker
- `New(title string, packages Packages) Model`
- `SetIgnored(map[string]bool)` - Mark packages as ignored
- `SetSelected(map[string]bool)` - Pre-select packages
- `Selected() Packages` - Get selected packages after confirm
- `Ignored() Packages` - Get newly ignored packages
- `Cancelled() bool` / `Confirmed() bool`

**progress.Model** - Installation progress display
- `New(title string, packages Packages, installFn func(Package) error) Model`
- Shows spinner, progress bar, recent results
- `Installed() int` / `Failed() int`

---

## Important Patterns

### Config Loading
```go
cfg, err := config.Load()  // Singleton, auto-detects current machine
currentMachine := cfg.CurrentMachine
```

### Brewfile Operations
```go
pkgs, err := brewfile.Parse(path)           // Read
diff := brewfile.Diff(sourcePkgs, currentPkgs)  // Compare
err = brewfile.Write(path, pkgs)            // Write
```

### History Logging
```go
history.LogImport(machine, source, []string{"brew:git"})
history.LogSync(machine, source, addedCount, removedCount)
history.LogDump(machine, map[string]int{"brew": 10}, committed)
```

### Deduplication Patterns

When dumping packages, BrewSync handles deduplication automatically:

**Adding Unique Packages**:
```go
// Add packages that don't already exist
allPackages = allPackages.AddUnique(extensions...)

// Usage: Adds VSCode/Cursor/Antigravity/Go packages only if not from brew bundle dump
beforeCount := len(allPackages)
allPackages = allPackages.AddUnique(vscodeExtensions...)
addedCount := len(allPackages) - beforeCount
printVerbose("Found %d extensions (%d new, %d already in Brewfile)",
    len(vscodeExtensions), addedCount, len(vscodeExtensions)-addedCount)
```

**Merging Lists with Preservation**:
```go
// Merge two lists, keeping packages with best descriptions
merged := list1.MergeUnique(list2)
// If a package exists in both, list2's version is used (better descriptions)
```

This pattern prevents duplicate entries when collecting from multiple sources (e.g., `brew bundle dump` may include VSCode extensions installed via Homebrew casks, so manual VSCode extension collection should skip duplicates).

---

## Testing

```bash
go test ./...                    # Run all tests
go test ./... -cover             # With coverage
go test ./internal/brewfile -v   # Specific package
```

Test files use `_test.go` suffix. Key test coverage:
- `brewfile`: 95% - parser, diff logic
- `config`: 60% - loading, defaults
- `exec`: 100% - command execution
- `history`: 26% - log parsing

### Using Makefile

The Makefile provides convenient shortcuts for testing:

```bash
make test              # Run all tests
make test-coverage     # With coverage report
make test-specific PKG=./internal/brewfile  # Test specific package
make test-race         # With race detector
make test-bench        # Run benchmarks
```

See [MAKEFILE_GUIDE.md](MAKEFILE_GUIDE.md) for complete testing documentation.

---

## Build Commands

```bash
go build ./...                           # Build all
go build -o brewsync ./cmd/brewsync      # Build binary
go install ./cmd/brewsync                # Install to GOPATH/bin
go run ./cmd/brewsync --help             # Run directly
```

---

## Build System & Makefile

The project includes a comprehensive Makefile with 50+ commands organized into categories.

### Quick Reference

**Development**:
- `make dev` - Build and show version
- `make quick` - Build and test (fastest)
- `make ci` - Run CI checks (format, vet, test)
- `make pre-commit` - Pre-commit checklist

**Testing**:
- `make test` - Run all tests
- `make test-coverage` - With coverage report
- `make test-coverage-detail` - Generate coverage.html
- `make test-race` - Run with race detector
- `make test-bench` - Run benchmarks

**Build & Release**:
- `make build` - Build to ./bin/brewsync
- `make release` - Optimized release build
- `make install` - Install to GOPATH/bin
- `make clean` - Clean build artifacts

**Code Quality**:
- `make fmt` - Format with gofmt
- `make vet` - Run go vet
- `make lint` - Run golangci-lint
- `make check` - Run fmt + vet

**Dependencies**:
- `make deps` - Download dependencies
- `make deps-tidy` - Tidy dependencies
- `make deps-verify` - Verify dependencies
- `make deps-update` - Update all dependencies

**Manual Testing**:
- `make test-setup` - Setup test environment
- `make test-cleanup` - Clean test environment

For complete documentation, see [MAKEFILE_GUIDE.md](MAKEFILE_GUIDE.md).

---

## Brewfile Format

Standard Homebrew Bundle format plus BrewSync extensions:

```ruby
# Standard entries with descriptions
tap "homebrew/bundle"
# Distributed revision control system
brew "git"
brew "libpq", link: true
# Launcher and productivity tool
cask "raycast"
mas "Xcode", id: 497799835
vscode "golang.go"

# BrewSync extensions
cursor "golang.go"
antigravity "python.lsp"
go "golang.org/x/tools/gopls"
```

**Package Descriptions**: Comments above packages (e.g., `# Distributed revision control system`) are automatically captured by `brew bundle dump --describe` when `dump.use_brew_bundle: true` (default). This makes your Brewfile self-documenting and helps when reviewing packages across machines.
