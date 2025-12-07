# Makefile Quick Reference

## Most Common Commands

### Daily Development

```bash
make quick              # Quick build + test (fastest iteration)
make dev                # Build locally + show version
make build              # Build to ./bin/brewsync
make test               # Run all tests
make clean              # Remove build artifacts
```

### Before Committing

```bash
make pre-commit         # Format + vet + test with coverage
make ci                 # Run CI checks (format + vet + test)
make check              # Format + vet only
```

### Testing

```bash
make test                      # Run all tests
make test-coverage             # Run tests with coverage summary
make test-coverage-detail      # Generate coverage.html report
make test-verbose              # Verbose test output
make test-race                 # Run with race detector
make test-specific PKG=./internal/brewfile  # Test specific package
```

### Building

```bash
make build              # Build to ./bin/brewsync
make build-local        # Build to ./brewsync (current dir)
make install            # Install to $GOPATH/bin
make release            # Optimized production build
```

### Code Quality

```bash
make fmt                # Format all code
make vet                # Run go vet
make lint               # Run golangci-lint (if installed)
make check              # fmt + vet
```

### Dependencies

```bash
make deps               # Download dependencies
make deps-tidy          # Tidy go.mod
make deps-verify        # Verify checksums
make deps-update        # Update all dependencies
```

### Running

```bash
make run ARGS="status"          # Build and run with args
make run ARGS="dump --dry-run"  # Example: dump dry-run
make doctor                     # Build and run doctor
make debug                      # Build with debug symbols
```

### Manual Testing

```bash
make test-setup         # Setup test environment (see MANUAL_TEST_GUIDE.md)
make test-cleanup       # Cleanup test environment
```

### Git & Release

```bash
make tag VERSION=v1.0.0  # Create and tag release
make status              # Show git status + build info
```

### Information

```bash
make help               # Show all commands with descriptions
make info               # Show project and Go environment info
make version            # Show current version
make list               # List all available targets
```

### Composite Commands

```bash
make all                # clean + build + test (full rebuild)
make quick              # build-local + test (fastest)
make ci                 # fmt + vet + test (CI pipeline)
make pre-commit         # fmt + vet + test-coverage (before commit)
```

---

## Typical Workflows

### Starting Work

```bash
make deps               # Ensure dependencies are current
make quick              # Verify everything works
```

### During Development

```bash
# Make code changes...
make quick              # Quick feedback loop
# or
make run ARGS="dump --dry-run"  # Test specific command
```

### Before Commit

```bash
make pre-commit         # Run all checks
# If passes, commit your changes
```

### Creating a Release

```bash
make ci                 # Verify CI passes
make tag VERSION=v1.0.0 # Tag the release
make release            # Build optimized binary
git push origin v1.0.0  # Push tag to remote
```

### Debugging Issues

```bash
make test-verbose       # See detailed test output
make debug              # Build with debug symbols
make run ARGS="doctor --verbose"  # Diagnose setup issues
```

### Clean Slate

```bash
make clean-all          # Remove all artifacts and caches
make deps               # Re-download dependencies
make all                # Full rebuild and test
```

---

## Examples

### Test Specific Package

```bash
# Test only brewfile package
make test-specific PKG=./internal/brewfile

# Test with verbose output
make test-specific PKG=./internal/config
```

### Run Commands with Arguments

```bash
# Dump packages
make run ARGS="dump --dry-run --verbose"

# Import from another machine
make run ARGS="import --from test2 --dry-run"

# Show status
make run ARGS="status"

# Run doctor
make doctor
```

### Generate Coverage Report

```bash
make test-coverage-detail
# Opens coverage.html in browser
open coverage.html
```

### Update Dependencies

```bash
# Update all dependencies to latest
make deps-update

# Test after update
make test
```

---

## Environment Variables

The Makefile uses these variables (you can override them):

```bash
BINARY_NAME=brewsync              # Binary name
BUILD_DIR=./bin                    # Build output directory
INSTALL_PATH=$GOPATH/bin           # Install location

# Override examples:
make build BUILD_DIR=./output
make install INSTALL_PATH=/usr/local/bin
```

---

## Tips

1. **Use `make quick` for fastest feedback** during development
2. **Run `make pre-commit`** before committing changes
3. **Use `make help`** to discover all available commands
4. **Use `make info`** to see current build configuration
5. **Combine with watch tools** for auto-rebuild:
   ```bash
   # Using fswatch (macOS)
   fswatch -o . | xargs -n1 -I{} make quick
   ```

---

## Color Output

The Makefile produces colorized output:
- 🟢 Green: Success messages
- 🟡 Yellow: Warnings and cleanup
- 🔵 Blue: Info and process messages
- ⚪ White: Regular output

To disable colors, redirect to a file or pipe:
```bash
make build > build.log 2>&1
```

---

## Integration with IDE

### VS Code Tasks

Add to `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Build",
      "type": "shell",
      "command": "make quick",
      "group": {
        "kind": "build",
        "isDefault": true
      }
    },
    {
      "label": "Test",
      "type": "shell",
      "command": "make test-verbose"
    }
  ]
}
```

### GoLand

1. Run → Edit Configurations
2. Add "Shell Script" configuration
3. Script: `make quick`

---

## Troubleshooting

### "command not found: make"

Install make:
```bash
xcode-select --install  # macOS
```

### "golangci-lint: command not found"

```bash
brew install golangci-lint
# or
make lint  # Shows install hint if missing
```

### Build fails with import errors

```bash
make deps-tidy    # Fix go.mod
make clean-all    # Clean caches
make build        # Try again
```

### Tests fail after dependency update

```bash
make deps-tidy    # Ensure go.mod is correct
make clean-all    # Clear test cache
make test-verbose # See detailed error
```

---

## Maintenance

### Keep Makefile Updated

As you add new features, update the Makefile:

1. Add new test targets for new packages
2. Add new run configurations for common operations
3. Update version/build info as needed
4. Keep help text synchronized

### Performance Tips

- `make quick` is fastest (local build + test)
- `make build` outputs to ./bin (keeps dir clean)
- `make ci` runs all checks (for CI/CD)
- `make release` creates optimized binary (for distribution)
