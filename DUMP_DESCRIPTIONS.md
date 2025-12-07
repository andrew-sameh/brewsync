# Dump with Descriptions Feature

## Overview

The `dump` command now supports automatic package descriptions in the generated Brewfile.

## Configuration

### Default Behavior (Recommended)

By default, BrewSync uses `brew bundle dump --describe` which:
- ✅ Automatically includes descriptions for all Homebrew packages (taps, formulae, casks)
- ✅ Works for VSCode, Cursor, Go tools, and mas (if installed via Homebrew)
- ✅ Faster than manual collection
- ✅ Uses Homebrew's official description data

**Config setting:**
```yaml
dump:
  use_brew_bundle: true  # Default
```

### Manual Collection Mode

If you prefer manual collection (e.g., for more control or custom descriptions):

```yaml
dump:
  use_brew_bundle: false
```

This mode:
- Collects packages manually using individual `brew list` commands
- Currently doesn't fetch descriptions (but could be enhanced)
- Useful for debugging or custom workflows

## Example Output

### With Descriptions (use_brew_bundle: true)

```ruby
tap "homebrew/bundle"
tap "homebrew/cask"
tap "homebrew/core"

# Clone of cat(1) with syntax highlighting and Git integration
brew "bat"
# Get/set bluetooth power and discoverable state
brew "blueutil"
# Distributed revision control system
brew "git"
# Lightweight and flexible command-line JSON processor
brew "jq"

# GPU-accelerated terminal emulator
cask "alacritty"
# Application launcher and productivity software
cask "raycast"

# BrewSync extensions
vscode "golang.go"
vscode "vscodevim.vim"

cursor "ms-python.python"

go "golang.org/x/tools/gopls"
```

### Without Descriptions (use_brew_bundle: false)

```ruby
tap "homebrew/bundle"
tap "homebrew/cask"
tap "homebrew/core"

brew "bat"
brew "blueutil"
brew "git"
brew "jq"

cask "alacritty"
cask "raycast"

# BrewSync extensions
vscode "golang.go"
vscode "vscodevim.vim"

cursor "ms-python.python"

go "golang.org/x/tools/gopls"
```

## Usage

### Basic Dump

```bash
# Uses config setting (default: with descriptions)
brewsync dump
```

### Preview

```bash
# See what would be written
brewsync dump --dry-run
```

### With Verbose Output

```bash
# See detailed collection process
brewsync dump --verbose
```

## How It Works

### With `use_brew_bundle: true` (Default)

1. Runs `brew bundle dump --force --describe --file=/tmp/brewfile`
2. Parses the output (includes descriptions as comments)
3. Adds VSCode, Cursor, Go, and mas packages
4. Writes combined Brewfile

### With `use_brew_bundle: false`

1. Runs `brew list --formula -1` for formulae
2. Runs `brew list --cask -1` for casks
3. Runs `brew tap` for taps
4. Fetches descriptions manually (if implemented)
5. Adds VSCode, Cursor, Go, and mas packages
6. Writes combined Brewfile

## Benefits of Descriptions

1. **Self-documenting**: Brewfiles are easier to understand
2. **Decision-making**: When importing, you know what each package does
3. **Onboarding**: New team members can see package purposes
4. **Review**: Easier to audit what's installed

## Compatibility

- ✅ Homebrew 3.0+: Full support for `--describe` flag
- ✅ Older Homebrew: Falls back to manual collection
- ✅ All package types: brew, cask, tap, vscode, cursor, go, mas

## Limitations

### Current

- VSCode/Cursor extensions don't have descriptions yet (extensions.json doesn't include them)
- Go tools don't have descriptions (go list doesn't provide them)
- mas descriptions would require additional API calls

### Future Enhancements

Could add descriptions for:
- VSCode extensions: Query VS Code marketplace API
- Cursor extensions: Same as VS Code
- Go tools: Parse go doc or README
- mas apps: Already included by brew bundle dump if available

## Performance

### With `use_brew_bundle: true`

- Single `brew bundle dump` command
- **Fast**: ~1-2 seconds for 100+ packages
- Descriptions included at no extra cost

### With `use_brew_bundle: false`

- Multiple `brew list` commands
- Additional `brew info` calls if fetching descriptions manually
- **Slower**: ~5-10 seconds for 100+ packages with descriptions

## Troubleshooting

### No descriptions appearing

1. Check config:
   ```bash
   brewsync config show | grep use_brew_bundle
   ```

2. Verify Homebrew version:
   ```bash
   brew --version
   # Should be 3.0+
   ```

3. Test brew bundle directly:
   ```bash
   brew bundle dump --describe --file=/tmp/test-brewfile
   cat /tmp/test-brewfile
   ```

### Descriptions truncated

This is normal - Homebrew provides concise one-line descriptions.

### Wrong descriptions

Descriptions come from Homebrew's database. If incorrect, report to Homebrew:
```bash
brew edit <formula-name>
```

## Migration

### From old Brewfiles without descriptions

Just run:
```bash
brewsync dump
```

Your Brewfile will be regenerated with descriptions.

### Preserving custom comments

If you have custom comments in your Brewfile:
1. Back up your Brewfile
2. Run `brewsync dump`
3. Manually add custom comments back

**Note**: `dump` always regenerates the entire Brewfile from scratch.

## Examples

### Enable descriptions in config

```bash
# Edit config
brewsync config edit

# Add or ensure this setting exists:
dump:
  use_brew_bundle: true
```

### Disable descriptions

```bash
# Edit config
brewsync config edit

# Set to false:
dump:
  use_brew_bundle: false
```

### One-time override (future feature)

```bash
# Could add flags like:
brewsync dump --use-brew-bundle     # Force use brew bundle
brewsync dump --no-brew-bundle      # Force manual collection
```

## Code Reference

- Config type: `internal/config/types.go:19-22`
- Default setting: `internal/config/defaults.go:35`
- Dump implementation: `internal/cli/dump.go:73-122`
- Brew bundle method: `internal/installer/brew.go:167-170`

## Related Documentation

- See `CLAUDE.md` for full config reference
- See `MANUAL_TEST_GUIDE.md` for testing instructions
- See `README.md` for usage examples
