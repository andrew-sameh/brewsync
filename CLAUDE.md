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
| `go` | Go tools | `golang.org/x/tools/gopls` |
| `mas` | Mac App Store | `497799835` (Xcode) |

### Ignore System

Packages can be ignored globally or per-machine via:
- Interactive selection (press `i` in TUI)
- Config file (`ignore` section)
- CLI command (`brewsync ignore add`)
- Inline flag (`--ignore "cask:app"`)

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

**Dump Flow**: Run `brew bundle dump` → Append VSCode/Cursor extensions → Append Go tools → Update metadata → Optionally commit/push

---

## File Structure

### Config Directory (`~/.config/brewsync/`)

```
~/.config/brewsync/
├── config.yaml           # Main configuration
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
default_categories: [tap, brew, cask, vscode, cursor, go, mas]

auto_dump:
  enabled: false
  after_install: false
  commit: false
  push: false
  commit_message: "brewsync: update {machine} Brewfile"

ignore:
  global:
    tap: []
    brew: []
    cask: ["company-vpn"]
    vscode: []
    cursor: []
    go: []
    mas: []
  mini:
    cask: ["bluestacks"]
  air:
    brew: ["scrcpy"]

machine_specific:
  mini:
    brew: ["postgresql@16", "redis"]
    cask: ["orbstack"]
  air:
    brew: ["ollama"]
    cask: ["ollama-app"]

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
  go: 9
  mas: 3
macos_version: "14.2"
brewsync_version: "1.0.0"
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

brewsync ignore list                     # Show ignored packages
brewsync ignore add "cask:app"           # Add to ignore list
brewsync ignore add "cask:app" --machine mini  # Machine-specific ignore
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
    Type    PackageType // tap, brew, cask, vscode, cursor, go, mas
    Name    string      // Package identifier
    Options string      // Optional: "link: true", "id: 123"
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

---

## Build Commands

```bash
go build ./...                           # Build all
go build -o brewsync ./cmd/brewsync      # Build binary
go install ./cmd/brewsync                # Install to GOPATH/bin
go run ./cmd/brewsync --help             # Run directly
```

---

## Brewfile Format

Standard Homebrew Bundle format plus BrewSync extensions:

```ruby
# Standard entries
tap "homebrew/bundle"
brew "git"
brew "libpq", link: true
cask "raycast"
mas "Xcode", id: 497799835
vscode "golang.go"

# BrewSync extensions
cursor "golang.go"
go "golang.org/x/tools/gopls"
```
