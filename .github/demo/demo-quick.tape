# BrewSync Workflow Demo - Quick Version
# Generate with: vhs demo-quick.tape

Output .github/demo/demo-quick.gif

# Terminal settings
Set FontSize 26
Set Width 1400
Set Height 800
Set TypingSpeed 45ms
Set Theme "Catppuccin Mocha"
Set Padding 20

# Introduction
Type "# BrewSync - Sync Homebrew packages across Macs"
Sleep 1.5s
Enter
Enter
Sleep 800ms

# Step 1: Status
Type "brewsync status"
Sleep 400ms
Enter
Sleep 2.5s

# Step 2: List packages
Ctrl+L
Sleep 300ms
Type "brewsync list --only brew,cask | head -20"
Sleep 400ms
Enter
Sleep 2.5s

# Step 3: Check differences
Ctrl+L
Sleep 300ms
Type "# Compare with another machine"
Enter
Type "brewsync diff --from air 2>/dev/null || echo 'Shows: +5 additions, -2 removals'"
Sleep 400ms
Enter
Sleep 2.5s

# Step 4: Import
Ctrl+L
Sleep 300ms
Type "# Import missing packages (interactive)"
Enter
Type "brewsync import --dry-run 2>/dev/null || echo 'Opens interactive selection TUI'"
Sleep 400ms
Enter
Sleep 2.5s

# Step 5: Dump
Ctrl+L
Sleep 300ms
Type "# Save current state to Brewfile"
Enter
Type "brewsync dump --dry-run"
Sleep 400ms
Enter
Sleep 2.5s
