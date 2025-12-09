# Demo Assets

This directory contains VHS tape files and generated GIFs for BrewSync demos.

## 📼 Tape Files (Source)

- **demo-quick.tape** - Quick workflow demo (recommended for README)
- **demo-tui.tape** - Full TUI navigation showcase
- **demo-interactive.tape** - Mixed TUI and CLI demo
- **demo.tape** - Comprehensive command showcase

## 🎬 Generate GIFs

Run from the repository root:

```bash
# Generate all demos
vhs .github/demo/demo-quick.tape
vhs .github/demo/demo-tui.tape
vhs .github/demo/demo-interactive.tape

# Or use make commands (if added to Makefile)
make demo
```

## 📏 GIF Specifications

- **Size**: 1400-1600px width, 800-1000px height
- **Theme**: Catppuccin Mocha (matches app colors)
- **Font Size**: 22-28px
- **Format**: GIF (for GitHub README compatibility)

## 🎨 Customization

Edit the `.tape` files to:
- Adjust timing (`Sleep` commands)
- Change appearance (`Set FontSize`, `Set Theme`)
- Add/remove steps
- Show different commands

## 📝 Usage in README

```markdown
![BrewSync Demo](.github/demo/demo-quick.gif)
```

## 🔄 Updating Demos

1. Edit the `.tape` file
2. Regenerate: `vhs .github/demo/[tape-name].tape`
3. Commit both `.tape` and `.gif` files
4. Keep GIF file sizes reasonable (< 2MB if possible)
