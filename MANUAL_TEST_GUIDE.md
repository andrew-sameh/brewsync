# BrewSync Manual Testing Guide

This guide provides step-by-step instructions to manually test all BrewSync functionality from the command line.

## Prerequisites

- macOS machine
- Homebrew installed
- Go 1.21+ installed
- Terminal access

## Setup Test Environment

### 1. Build the Binary

```bash
cd /Users/andrewsam/Code/brewsync
go build -o brewsync ./cmd/brewsync

# Optional: Install to PATH
go install ./cmd/brewsync
```

### 2. Create Test Directories

```bash
# Create a test dotfiles directory
mkdir -p ~/brewsync-test-dotfiles/_brew_test1
mkdir -p ~/brewsync-test-dotfiles/_brew_test2

# Backup existing config if it exists
[ -d ~/.config/brewsync ] && mv ~/.config/brewsync ~/.config/brewsync.backup

# Create fresh config directory
mkdir -p ~/.config/brewsync/profiles
```

---

## Test Suite

### TEST 1: Version & Help

**Purpose**: Verify basic CLI functionality

```bash
# Test: Show version
./brewsync --version
# Expected: Shows version number

# Test: Show help
./brewsync --help
# Expected: Lists all available commands

# Test: Command-specific help
./brewsync dump --help
./brewsync import --help
./brewsync config --help
# Expected: Shows detailed help for each command
```

**✅ Pass Criteria**: All help text displays correctly, version shows

---

### TEST 2: Config Initialization

**Purpose**: Test configuration setup

```bash
# Test: Initialize config interactively
./brewsync config init
# Expected: Interactive prompts for machine setup
# - Enter machine ID (e.g., "test1")
# - Confirm hostname detection
# - Enter Brewfile path: ~/brewsync-test-dotfiles/_brew_test1/Brewfile
# - Enter description: "Test Machine 1"

# Test: Verify config was created
./brewsync config show
# Expected: Displays YAML config with your machine

# Test: Show config path
./brewsync config path
# Expected: Shows ~/.config/brewsync/config.yaml

# Test: Verify config file exists
cat ~/.config/brewsync/config.yaml
# Expected: Valid YAML with your machine configuration
```

**✅ Pass Criteria**: Config file created, shows correct machine info

---

### TEST 3: Doctor Command

**Purpose**: Validate setup and dependencies

```bash
# Test: Run doctor
./brewsync doctor
# Expected: Checks and reports:
# - ✓ Config file exists
# - ✓ Current machine detected
# - ✓ Brewfile path configured
# - ✓/✗ Homebrew available
# - ✓/✗ VSCode CLI available
# - ✓/✗ Cursor CLI available
# - ✓/✗ mas CLI available
# - ✓/✗ Go available

# Test: Verbose mode
./brewsync doctor --verbose
# Expected: More detailed output

# Test: Quiet mode
./brewsync doctor --quiet
# Expected: Minimal output, only errors
```

**✅ Pass Criteria**: Doctor runs, detects tools correctly

---

### TEST 4: Dump Command

**Purpose**: Export current packages to Brewfile

```bash
# Test: Dry-run dump
./brewsync dump --dry-run
# Expected: Shows what would be dumped without writing

# Test: Verbose dry-run
./brewsync dump --dry-run --verbose
# Expected: Detailed output of package collection

# Test: Actual dump
./brewsync dump
# Expected: Creates Brewfile at configured path

# Test: Verify Brewfile was created
ls -la ~/brewsync-test-dotfiles/_brew_test1/Brewfile
cat ~/brewsync-test-dotfiles/_brew_test1/Brewfile | head -20
# Expected: File exists with packages in Brewfile format

# Test: Dump with commit flag (not implemented yet)
./brewsync dump --commit
# Expected: Warning that flag is not implemented

# Test: Dump with push flag (not implemented yet)
./brewsync dump --push
# Expected: Warning that flag is not implemented
```

**✅ Pass Criteria**: Brewfile created with current packages

---

### TEST 5: List Command

**Purpose**: Display packages in Brewfiles

```bash
# Test: List current machine packages
./brewsync list
# Expected: Shows all packages from current Brewfile

# Test: List specific types only
./brewsync list --only tap,brew
# Expected: Shows only taps and formulae

# Test: List with type filter
./brewsync list --only cask
# Expected: Shows only casks

# Test: Verbose list
./brewsync list --verbose
# Expected: More detailed package information

# Test: JSON output
./brewsync list --format json
# Expected: JSON array of packages

# Test: JSON with pretty print
./brewsync list --format json | jq '.'
# Expected: Formatted JSON (if jq is installed)
```

**✅ Pass Criteria**: Package lists display correctly in all formats

---

### TEST 6: Status Command

**Purpose**: Show current machine state

```bash
# Test: Basic status
./brewsync status
# Expected: Shows:
# - Current machine name
# - Hostname
# - Brewfile path
# - Package counts by type
# - Last operations (if any)

# Test: Verbose status
./brewsync status --verbose
# Expected: More detailed information

# Test: Status with no history
./brewsync status
# Expected: Shows "No recent operations" or similar
```

**✅ Pass Criteria**: Status displays machine info and package counts

---

### TEST 7: Add Second Machine

**Purpose**: Test multi-machine configuration

```bash
# Test: Add another machine
./brewsync config add-machine test2 \
  --hostname "Test-Machine-2" \
  --brewfile ~/brewsync-test-dotfiles/_brew_test2/Brewfile \
  --description "Test Machine 2"
# Expected: Machine added to config

# Test: Verify both machines exist
./brewsync config show
# Expected: Shows both test1 and test2

# Test: Manually create a simple Brewfile for test2
cat > ~/brewsync-test-dotfiles/_brew_test2/Brewfile << 'EOF'
tap "homebrew/bundle"
brew "git"
brew "wget"
brew "htop"
cask "firefox"
vscode "ms-python.python"
EOF

# Test: Verify test2 Brewfile
cat ~/brewsync-test-dotfiles/_brew_test2/Brewfile
# Expected: Shows the packages we just created
```

**✅ Pass Criteria**: Second machine added, config shows both

---

### TEST 8: Diff Command

**Purpose**: Compare packages between machines

```bash
# Test: Diff against test2
./brewsync diff --from test2
# Expected: Shows:
# - Packages only in test2 (missing locally)
# - Packages only in current machine
# - Common packages

# Test: Diff specific types
./brewsync diff --from test2 --only brew,cask
# Expected: Shows diff for only brew and cask types

# Test: Diff with verbose
./brewsync diff --from test2 --verbose
# Expected: More detailed comparison

# Test: Diff JSON output
./brewsync diff --from test2 --format json
# Expected: JSON representation of differences

# Test: Diff with invalid source
./brewsync diff --from nonexistent
# Expected: Error message about unknown machine
```

**✅ Pass Criteria**: Diff shows package differences correctly

---

### TEST 9: Import Command (Non-Interactive)

**Purpose**: Import packages from another machine

```bash
# Test: Dry-run import
./brewsync import --from test2 --dry-run
# Expected: Shows what would be installed without installing

# Test: Import specific types
./brewsync import --from test2 --only brew --dry-run
# Expected: Shows only brew packages that would be installed

# Test: Import with --yes flag (skip interactive)
./brewsync import --from test2 --only brew --yes --dry-run
# Expected: Would install all without prompts (dry-run prevents actual install)

# Test: Import excluding types
./brewsync import --from test2 --skip vscode,cask --dry-run
# Expected: Shows packages except vscode and cask types

# Test: Multiple source machines
echo 'brew "curl"' >> ~/brewsync-test-dotfiles/_brew_test2/Brewfile
./brewsync import --from test2 --dry-run
# Expected: Shows union of packages from both sources
```

**✅ Pass Criteria**: Import shows correct packages to install

---

### TEST 10: Import Command (Interactive TUI)

**Purpose**: Test interactive package selection

```bash
# Test: Interactive import (opens TUI)
./brewsync import --from test2

# In TUI:
# - Press 'space' to toggle package selection
# - Press 'a' to select all
# - Press 'n' to deselect all
# - Press '1', '2', '3', etc. to filter by category
# - Press '/' to search (if implemented)
# - Press 'i' to mark packages as ignored
# - Press 'q' to quit
# - Press 'enter' to confirm and install

# Expected: Interactive selection interface appears
# Note: This requires actual interaction, not CLI-only

# Test: Import with assume-yes (bypass TUI)
./brewsync import --from test2 --yes --dry-run
# Expected: Skips TUI, shows what would be installed
```

**✅ Pass Criteria**: TUI launches and responds to inputs

---

### TEST 11: Sync Command

**Purpose**: Bidirectional sync (add + remove)

```bash
# Test: Sync dry-run (default)
./brewsync sync --from test2
# Expected: Shows packages that would be:
# - Added (in test2, not in current)
# - Removed (in current, not in test2)

# Test: Sync specific types
./brewsync sync --from test2 --only brew
# Expected: Shows sync actions only for brew packages

# Test: Sync with verbose
./brewsync sync --from test2 --verbose
# Expected: Detailed sync plan

# Test: Sync preview mode
./brewsync sync --from test2
# Expected: Shows changes but doesn't apply

# Test: Sync apply (CAUTION: Actually makes changes)
./brewsync sync --from test2 --apply --dry-run
# Note: --dry-run prevents actual changes even with --apply

# Test: Sync with confirmation
./brewsync sync --from test2 --apply
# Expected: Asks for confirmation before proceeding
# (Cancel with 'n')

# Test: Sync with --yes flag
./brewsync sync --from test2 --apply --yes --dry-run
# Expected: Would skip confirmation (dry-run prevents actual changes)
```

**✅ Pass Criteria**: Sync shows correct add/remove actions

---

### TEST 12: Ignore Management

**Purpose**: Test ignore list functionality

```bash
# Test: List ignores (empty initially)
./brewsync ignore list
# Expected: Shows empty ignore lists or "No ignored packages"

# Test: Add package to ignore list
./brewsync ignore add brew:wget
# Expected: Confirmation that package was added to ignore list

# Test: Add cask to ignore
./brewsync ignore add cask:firefox
# Expected: Confirmation

# Test: Add globally
./brewsync ignore add brew:htop --global
# Expected: Added to global ignore list

# Test: Add for specific machine
./brewsync ignore add brew:git --machine test2
# Expected: Added to test2's ignore list

# Test: List ignores again
./brewsync ignore list
# Expected: Shows all ignored packages organized by scope

# Test: Verify in config file
cat ~/.config/brewsync/config.yaml | grep -A 10 ignore
# Expected: Shows ignore section with added packages

# Test: Remove from ignore list
./brewsync ignore remove brew:wget
# Expected: Confirmation that package was removed

# Test: Clear all ignores
./brewsync ignore clear
# Expected: Asks for confirmation, then clears all

# Test: Clear with --yes
./brewsync ignore clear --yes
# Expected: Clears without confirmation
```

**✅ Pass Criteria**: Ignore operations work, config updates correctly

---

### TEST 13: Profile Management

**Purpose**: Test profile creation and usage

```bash
# Test: List profiles (empty initially)
./brewsync profile list
# Expected: Shows "No profiles found" or empty list

# Test: Create a profile manually
cat > ~/.config/brewsync/profiles/test-core.yaml << 'EOF'
name: test-core
description: "Core testing tools"

packages:
  tap:
    - homebrew/bundle
  brew:
    - git
    - curl
    - wget
  cask:
    - firefox
  vscode:
    - vscodevim.vim
EOF

# Test: List profiles again
./brewsync profile list
# Expected: Shows test-core profile

# Test: Show profile contents
./brewsync profile show test-core
# Expected: Displays profile packages

# Test: Show non-existent profile
./brewsync profile show nonexistent
# Expected: Error message

# Test: Create profile via CLI
./brewsync profile create test-dev --description "Development tools"
# Expected: Opens $EDITOR to edit profile
# (May need to set EDITOR="nano" or similar)

# Test: Edit existing profile
EDITOR=cat ./brewsync profile edit test-core
# Expected: Opens profile in editor (cat just displays it)

# Test: Install from profile (dry-run)
./brewsync profile install test-core --dry-run
# Expected: Shows packages that would be installed

# Test: Install multiple profiles
./brewsync profile install test-core,test-dev --dry-run
# Expected: Shows union of packages from both profiles

# Test: Install profile with --yes
./brewsync profile install test-core --yes --dry-run
# Expected: Would install without prompts

# Test: Delete profile
./brewsync profile delete test-dev
# Expected: Asks for confirmation, then deletes

# Test: Delete with --yes
./brewsync profile delete test-core --yes
# Expected: Deletes without confirmation

# Test: Verify deletion
./brewsync profile list
# Expected: Profiles are gone
```

**✅ Pass Criteria**: Profile CRUD operations work correctly

---

### TEST 14: History Command

**Purpose**: View operation history

```bash
# Test: View history (may be empty)
./brewsync history
# Expected: Shows recent operations or "No history"

# Test: Perform some operations to create history
./brewsync dump
./brewsync list > /dev/null
./brewsync status > /dev/null

# Test: View history again
./brewsync history
# Expected: Shows logged operations

# Test: Check history file directly
cat ~/.config/brewsync/history.log
# Expected: Shows append-only log entries

# Test: History with verbose
./brewsync history --verbose
# Expected: More detailed history

# Test: History with limit (if supported)
./brewsync history | head -10
# Expected: Shows first 10 entries
```

**✅ Pass Criteria**: History logs and displays operations

---

### TEST 15: Global Flags

**Purpose**: Test flags that work across all commands

```bash
# Test: Dry-run flag on various commands
./brewsync dump --dry-run
./brewsync import --from test2 --dry-run --yes
./brewsync sync --from test2 --dry-run
# Expected: All show previews without executing

# Test: Verbose flag
./brewsync status --verbose
./brewsync list --verbose
./brewsync doctor --verbose
# Expected: All show more detailed output

# Test: Quiet flag
./brewsync status --quiet
./brewsync dump --quiet
# Expected: Minimal output

# Test: No-color flag
./brewsync status --no-color
./brewsync list --no-color
# Expected: Output without ANSI color codes

# Test: Assume-yes flag
./brewsync ignore clear --yes
./brewsync profile delete test-core --yes 2>&1 || true
# Expected: Skips confirmations

# Test: Custom config file
echo "machines: {}" > /tmp/test-config.yaml
./brewsync --config /tmp/test-config.yaml config show
# Expected: Uses specified config file

# Test: Combine multiple flags
./brewsync list --verbose --no-color --quiet
# Expected: Quiet takes precedence over verbose
```

**✅ Pass Criteria**: Flags modify behavior as expected

---

### TEST 16: Error Handling

**Purpose**: Test error cases and edge conditions

```bash
# Test: Invalid command
./brewsync invalid-command
# Expected: Error message, suggests 'help'

# Test: Missing required argument
./brewsync config add-machine
# Expected: Error about missing arguments

# Test: Invalid machine name
./brewsync diff --from nonexistent
# Expected: Error "unknown machine"

# Test: Import from current machine
./brewsync import --from test1
# Expected: Error "cannot import from current machine"

# Test: Non-existent config file
./brewsync --config /nonexistent/config.yaml status
# Expected: Error about config file

# Test: Invalid package type in --only
./brewsync list --only invalid-type
# Expected: Error or warning about invalid type

# Test: Malformed Brewfile
echo "invalid brewfile content" > /tmp/bad-brewfile
./brewsync list --from test2
# Note: May need to temporarily break test2's Brewfile
# Expected: Handles parse errors gracefully

# Test: Missing Brewfile
./brewsync config add-machine test3 --hostname "test3" --brewfile /nonexistent/Brewfile
./brewsync list --from test3 2>&1 || true
# Expected: Error about missing Brewfile
```

**✅ Pass Criteria**: Errors are clear and helpful

---

### TEST 17: Edge Cases

**Purpose**: Test boundary conditions

```bash
# Test: Empty Brewfile
echo "" > ~/brewsync-test-dotfiles/_brew_test2/Brewfile
./brewsync list --from test2
# Expected: Shows "No packages" or empty list

# Test: Brewfile with only comments
cat > ~/brewsync-test-dotfiles/_brew_test2/Brewfile << 'EOF'
# This is a comment
# Another comment
EOF
./brewsync list --from test2
# Expected: Shows no packages

# Test: Import with no differences
./brewsync dump
./brewsync import --from test1 --dry-run
# Expected: Shows "nothing to import" or similar

# Test: Sync with identical machines
./brewsync sync --from test1 --dry-run
# Expected: Shows "already in sync"

# Test: Very long package list
# (Create a Brewfile with 100+ packages)
seq 1 100 | xargs -I {} echo 'brew "package{}"' > ~/brewsync-test-dotfiles/_brew_test2/Brewfile
./brewsync list --from test2
# Expected: Handles large lists without issues

# Test: Special characters in package names
cat > ~/brewsync-test-dotfiles/_brew_test2/Brewfile << 'EOF'
brew "git"
brew "python@3.11"
brew "node@18"
cask "android-studio"
EOF
./brewsync list --from test2
# Expected: Parses special characters correctly
```

**✅ Pass Criteria**: Handles edge cases gracefully

---

### TEST 18: Config Editing

**Purpose**: Test config modification

```bash
# Test: Show current config
./brewsync config show

# Test: Edit config in editor
EDITOR=cat ./brewsync config edit
# Expected: Opens config (cat just displays it)

# Test: Manual config edit
cp ~/.config/brewsync/config.yaml ~/.config/brewsync/config.yaml.bak
cat >> ~/.config/brewsync/config.yaml << 'EOF'

# Custom settings
output:
  color: true
  verbose: false
  show_descriptions: true
EOF

# Test: Validate config after edit
./brewsync config show
# Expected: Shows updated config

# Test: Doctor should validate config
./brewsync doctor
# Expected: Validates updated config

# Test: Restore backup
mv ~/.config/brewsync/config.yaml.bak ~/.config/brewsync/config.yaml
```

**✅ Pass Criteria**: Config can be edited and validated

---

### TEST 19: Machine-Specific Packages

**Purpose**: Test machine-specific package handling

```bash
# Test: Add machine-specific packages to config
cat >> ~/.config/brewsync/config.yaml << 'EOF'

machine_specific:
  test1:
    brew:
      - postgresql@16
      - redis
    cask:
      - docker
  test2:
    brew:
      - nginx
EOF

# Test: Verify config
./brewsync config show | grep -A 10 machine_specific

# Test: Import without machine-specific
./brewsync import --from test2 --dry-run
# Expected: Doesn't suggest test2's machine-specific packages

# Test: Import with machine-specific flag
./brewsync import --from test2 --include-machine-specific --dry-run
# Expected: Includes test2's machine-specific packages

# Test: Sync respects machine-specific
./brewsync sync --from test2 --dry-run
# Expected: Doesn't remove current machine's machine-specific packages
```

**✅ Pass Criteria**: Machine-specific packages handled correctly

---

### TEST 20: Concurrent Operations

**Purpose**: Test file locking and concurrent access

```bash
# Test: Run multiple commands simultaneously
./brewsync status &
./brewsync list &
./brewsync doctor &
wait
# Expected: All complete without file corruption

# Test: Multiple dumps (should be safe to read)
./brewsync dump --dry-run &
./brewsync dump --dry-run &
wait
# Expected: Both complete successfully

# Note: Actual concurrent writes are harder to test safely
```

**✅ Pass Criteria**: Concurrent reads don't cause errors

---

## Cleanup Test Environment

```bash
# Remove test directories
rm -rf ~/brewsync-test-dotfiles

# Remove test config
rm -rf ~/.config/brewsync

# Restore original config if backed up
[ -d ~/.config/brewsync.backup ] && mv ~/.config/brewsync.backup ~/.config/brewsync

# Remove test binary
rm -f ./brewsync
```

---

## Test Results Checklist

Use this checklist to track your testing progress:

- [ ] TEST 1: Version & Help
- [ ] TEST 2: Config Initialization
- [ ] TEST 3: Doctor Command
- [ ] TEST 4: Dump Command
- [ ] TEST 5: List Command
- [ ] TEST 6: Status Command
- [ ] TEST 7: Add Second Machine
- [ ] TEST 8: Diff Command
- [ ] TEST 9: Import Command (Non-Interactive)
- [ ] TEST 10: Import Command (Interactive TUI)
- [ ] TEST 11: Sync Command
- [ ] TEST 12: Ignore Management
- [ ] TEST 13: Profile Management
- [ ] TEST 14: History Command
- [ ] TEST 15: Global Flags
- [ ] TEST 16: Error Handling
- [ ] TEST 17: Edge Cases
- [ ] TEST 18: Config Editing
- [ ] TEST 19: Machine-Specific Packages
- [ ] TEST 20: Concurrent Operations

---

## Known Issues to Verify

Based on the status report, verify these known limitations:

1. **Git integration**: `--commit` and `--push` flags show "not implemented" warning
2. **Metadata files**: `.brewsync-meta` files not generated
3. **Hooks**: Config hooks section not executed
4. **Auto-dump**: Auto-dump after installs not functional

---

## Reporting Issues

When you find issues, note:
- Exact command used
- Expected behavior
- Actual behavior
- Error messages (full output)
- Terminal output/screenshots
- Environment details (macOS version, shell, etc.)

---

## Tips for Effective Testing

1. **Use --dry-run liberally**: Prevents unwanted changes
2. **Use --verbose**: Provides more context when debugging
3. **Check config file**: `cat ~/.config/brewsync/config.yaml` after operations
4. **Check history**: `cat ~/.config/brewsync/history.log` to see what was logged
5. **Test incrementally**: Don't skip to advanced tests without basics working
6. **Keep notes**: Document what works and what doesn't
7. **Use different machines**: If possible, test on actual different Macs

---

## Success Criteria

The tool is working correctly if:
- ✅ All commands execute without crashes
- ✅ Help text is clear and accurate
- ✅ Brewfiles are parsed and written correctly
- ✅ Config operations update config.yaml properly
- ✅ Diff shows accurate package differences
- ✅ Import/sync preview mode is accurate
- ✅ TUI is responsive and functional
- ✅ Error messages are helpful
- ✅ History is logged correctly
- ✅ Flags modify behavior as documented
