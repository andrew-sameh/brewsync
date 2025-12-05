# BrewSync

A CLI tool to sync Homebrew packages, casks, taps, VSCode/Cursor extensions, Go tools, and Mac App Store apps across multiple macOS machines.

## Overview

BrewSync solves the problem of keeping multiple macOS machines in sync. When you install something on one machine, you can easily replicate that on your other machines without manual tracking.

**Key Design Principles:**
- **Machine-centric**: Each machine maintains its own Brewfile as the source of truth
- **Git-based**: Designed for dotfiles workflows, not cloud storage
- **Non-destructive by default**: Import only adds packages, never removes
- **Interactive**: Choose what to install with batch options

## Features

- Sync packages across multiple macOS machines
- Support for multiple package types:
  - Homebrew taps, formulae, and casks
  - VSCode extensions
  - Cursor extensions
  - Go tools
  - Mac App Store apps
- Profile system for curated package groups
- Ignore lists (global and per-machine)
- Diff view between machines
- Operation history logging
- Doctor command for setup validation

## Installation

```bash
# Build from source
go build -o brewsync ./cmd/brewsync

# Install to GOPATH/bin
go install ./cmd/brewsync
```

## Quick Start

### 1. Initialize Configuration

```bash
brewsync config init
```

This creates `~/.config/brewsync/config.yaml` with your machine settings based on hostname.

### 2. Dump Current Packages

```bash
brewsync dump
```

Creates/updates your machine's Brewfile with all installed packages.

### 3. Check Status

```bash
brewsync status
```

Shows current machine info, package counts, and pending changes.

## Usage Flow

### Setting Up Multiple Machines

```
Machine A (main workstation)          Machine B (laptop)
─────────────────────────────         ─────────────────────────
1. brewsync config init               1. brewsync config init
2. brewsync dump                      2. brewsync dump
3. git commit & push                  3. git pull
                                      4. brewsync diff --from machineA
                                      5. brewsync import --from machineA
```

### Daily Workflow

```bash
# On your main machine - install new tools
brew install newtool
brew install --cask newapp

# Save state at end of day
brewsync dump
cd ~/dotfiles && git add -A && git commit -m "Update brewfile" && git push

# On another machine - sync up
cd ~/dotfiles && git pull
brewsync status           # See what's new
brewsync import           # Install missing packages
```

### Using Profiles

```bash
# Create a profile for core tools
brewsync profile create core --description "Essential tools"
brewsync profile edit core

# Install from profile on any machine
brewsync profile install core

# Install multiple profiles
brewsync profile install core,dev-go,k8s
```

## Commands Reference

### Core Commands

| Command | Description |
|---------|-------------|
| `dump` | Update Brewfile from installed packages |
| `list` | List packages in a Brewfile |
| `diff` | Show differences between machines |
| `import` | Install missing packages from another machine (interactive TUI) |
| `sync` | Make current machine match source exactly (preview + apply) |

### Status & Diagnostics

| Command | Description |
|---------|-------------|
| `status` | Show current machine state overview |
| `doctor` | Validate setup and diagnose issues |
| `history` | View operation history |

### Configuration

| Command | Description |
|---------|-------------|
| `config show` | Display current configuration |
| `config edit` | Open config in $EDITOR |
| `config path` | Show config file path |
| `config init` | Initialize configuration |
| `config add-machine` | Add a new machine |

### Ignore Management

| Command | Description |
|---------|-------------|
| `ignore list` | Show all ignored packages |
| `ignore add` | Add package to ignore list |
| `ignore remove` | Remove from ignore list |
| `ignore clear` | Clear all ignored packages |

### Profile Management

| Command | Description |
|---------|-------------|
| `profile list` | List available profiles |
| `profile show` | Display profile contents |
| `profile install` | Install packages from profile(s) |
| `profile create` | Create a new profile |
| `profile edit` | Edit profile in $EDITOR |
| `profile delete` | Delete a profile |

### Global Flags

```
--config string   Config file (default ~/.config/brewsync/config.yaml)
--dry-run         Preview without executing
--verbose, -v     Detailed output
--quiet, -q       Minimal output
--no-color        Disable colored output
--yes, -y         Skip confirmations
```

## Command Examples

### dump

```bash
brewsync dump                    # Update Brewfile
brewsync dump --commit           # Commit changes to git
brewsync dump --push             # Commit and push
brewsync dump --dry-run          # Preview changes
```

### import

```bash
brewsync import                    # Interactive TUI selection
brewsync import --from air         # From specific machine
brewsync import --from mini,air    # Union of multiple machines
brewsync import --only brew,cask   # Filter categories
brewsync import --skip vscode      # Exclude categories
brewsync import --yes              # Install all without prompts
brewsync import --dry-run          # Preview only
brewsync import --include-machine-specific  # Include machine-specific packages
```

The interactive TUI lets you:
- Toggle packages with `space`
- Select all/none with `a`/`n`
- Filter by category with number keys `1-7`
- Search with `/`
- Mark as ignored with `i`
- Confirm with `enter`

### sync

```bash
brewsync sync                    # Preview mode (shows changes)
brewsync sync --apply            # Execute changes
brewsync sync --from air         # Sync from specific machine
brewsync sync --only brew        # Only sync specific types
brewsync sync --apply --yes      # Apply without confirmation
```

Sync differs from import:
- Import only **adds** missing packages
- Sync **adds AND removes** to match source exactly
- Protected packages (machine-specific, ignored) are never removed

### list

```bash
brewsync list                    # Current machine
brewsync list --from mini        # Another machine
brewsync list --only brew,cask   # Filter by type
brewsync list --format json      # JSON output
```

### diff

```bash
brewsync diff                    # Compare with default source
brewsync diff --from air         # Compare with specific machine
brewsync diff --only brew,cask   # Filter to specific types
brewsync diff --format json      # Output as JSON
```

### ignore

```bash
brewsync ignore add cask:bluestacks              # Ignore on current machine
brewsync ignore add brew:postgresql --global     # Ignore globally
brewsync ignore add cask:steam --machine mini    # Ignore on specific machine
brewsync ignore list                             # Show all ignored
brewsync ignore remove cask:bluestacks           # Remove from ignore
```

### profile

```bash
brewsync profile list                           # List profiles
brewsync profile show core                      # Show profile contents
brewsync profile install core                   # Install from profile
brewsync profile install core,dev-go            # Install multiple
brewsync profile create web-dev                 # Create new profile
brewsync profile edit core                      # Edit in $EDITOR
brewsync profile delete old-profile             # Delete profile
```

## Configuration

Configuration is stored at `~/.config/brewsync/config.yaml`.

### Example Configuration

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

current_machine: auto  # Auto-detect from hostname
default_source: mini   # Default machine for import/diff

default_categories:
  - tap
  - brew
  - cask
  - vscode
  - cursor
  - go
  - mas

ignore:
  global:
    cask:
      - "company-vpn"  # Manually installed
  mini:
    cask:
      - "bluestacks"   # Don't need on workstation

output:
  color: true
  verbose: false
```

## Profiles

Profiles are YAML files stored in `~/.config/brewsync/profiles/`.

### Example Profile (`~/.config/brewsync/profiles/core.yaml`)

```yaml
name: core
description: "Essential tools for any machine"

packages:
  tap:
    - homebrew/bundle
  brew:
    - git
    - fzf
    - bat
    - eza
    - fd
    - ripgrep
    - lazygit
    - starship
  cask:
    - raycast
    - iterm2
  vscode:
    - vscodevim.vim
    - eamodio.gitlens
```

## Directory Structure

```
~/.config/brewsync/
├── config.yaml           # Main configuration
├── history.log           # Operation history
└── profiles/             # Profile definitions
    ├── core.yaml
    ├── dev-go.yaml
    └── dev-python.yaml

~/dotfiles/               # Your dotfiles repo
├── _brew_mini/
│   └── Brewfile          # Mini's package list
├── _brew_air/
│   └── Brewfile          # Air's package list
└── ...
```

## Package Types

| Type | Source | Example |
|------|--------|---------|
| `tap` | Homebrew taps | `charmbracelet/tap` |
| `brew` | Homebrew formulae | `git`, `fzf`, `bat` |
| `cask` | Homebrew casks | `raycast`, `slack` |
| `vscode` | VSCode extensions | `golang.go` |
| `cursor` | Cursor extensions | `ms-python.python` |
| `go` | Go tools | `golang.org/x/tools/gopls` |
| `mas` | Mac App Store | `497799835` (Xcode) |

## Brewfile Format

BrewSync uses the standard Brewfile format with extensions:

```ruby
# Standard Homebrew entries
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

## Troubleshooting

### Run the doctor command

```bash
brewsync doctor
```

This checks:
- Config file exists and is valid
- Current machine is detected
- Brewfile paths exist
- Required CLI tools are available

### Common Issues

| Issue | Solution |
|-------|----------|
| "Machine not recognized" | Run `brewsync config init` or add machine manually |
| "Brewfile not found" | Run `brewsync dump` to create it |
| "brew command failed" | Check package name, verify network |
| CLI not available | Install missing tool (code, cursor, mas, go) |

## Requirements

- macOS
- Go 1.21+ (for building from source)
- Homebrew
- Optional: VSCode (`code` CLI), Cursor (`cursor` CLI), mas-cli, Go

## Development

```bash
# Build
go build -o brewsync ./cmd/brewsync

# Run tests
go test ./...

# Run with coverage
go test ./... -cover

# Install locally
go install ./cmd/brewsync
```

## License

MIT
