# BrewSync Interactive Demo
# Generate with: vhs demo-interactive.tape

Output .github/demo/demo-interactive.gif

# Terminal settings
Set FontSize 24
Set Width 1600
Set Height 1000
Set TypingSpeed 40ms
Set Theme "Catppuccin Mocha"
Set Padding 20

# Show the main TUI
Type "# BrewSync - Interactive TUI Demo"
Sleep 1s
Enter
Enter
Sleep 500ms

Type "# Launch the interactive interface"
Sleep 500ms
Enter
Type "brewsync"
Sleep 800ms
Enter
Sleep 2s

# Navigate through screens with number keys
# 1=Dashboard, 2=Import, 3=Sync, 4=Diff, 5=Dump, 6=List
Sleep 1s

# Go to Import screen
Type "2"
Sleep 2s

# Go to Sync screen
Type "3"
Sleep 2s

# Go to Diff screen
Type "4"
Sleep 2s

# Go to List screen
Type "6"
Sleep 2s

# Back to Dashboard
Type "1"
Sleep 1.5s

# Press 'q' to quit
Type "q"
Sleep 500ms

# Show import command
Ctrl+L
Sleep 500ms
