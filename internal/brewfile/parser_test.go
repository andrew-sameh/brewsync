package brewfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_ParseString_Taps(t *testing.T) {
	content := `
tap "homebrew/bundle"
tap "charmbracelet/tap"
tap "homebrew/cask-fonts"
`
	parser := NewParser()
	packages, err := parser.ParseString(content)
	require.NoError(t, err)
	assert.Len(t, packages, 3)

	for _, pkg := range packages {
		assert.Equal(t, TypeTap, pkg.Type)
	}
	assert.Equal(t, "homebrew/bundle", packages[0].Name)
	assert.Equal(t, "charmbracelet/tap", packages[1].Name)
}

func TestParser_ParseString_Brews(t *testing.T) {
	content := `
brew "git"
brew "fzf"
brew "bat"
brew "libpq", link: true
brew "python@3.11"
`
	parser := NewParser()
	packages, err := parser.ParseString(content)
	require.NoError(t, err)
	assert.Len(t, packages, 5)

	for _, pkg := range packages {
		assert.Equal(t, TypeBrew, pkg.Type)
	}
	assert.Equal(t, "git", packages[0].Name)
	assert.Equal(t, "libpq", packages[3].Name)
	assert.Equal(t, "true", packages[3].Options["link"])
}

func TestParser_ParseString_Casks(t *testing.T) {
	content := `
cask "raycast"
cask "slack"
cask "visual-studio-code"
`
	parser := NewParser()
	packages, err := parser.ParseString(content)
	require.NoError(t, err)
	assert.Len(t, packages, 3)

	for _, pkg := range packages {
		assert.Equal(t, TypeCask, pkg.Type)
	}
	assert.Equal(t, "raycast", packages[0].Name)
}

func TestParser_ParseString_MAS(t *testing.T) {
	content := `
mas "Xcode", id: 497799835
mas "TestFlight", id: 899247664
`
	parser := NewParser()
	packages, err := parser.ParseString(content)
	require.NoError(t, err)
	assert.Len(t, packages, 2)

	for _, pkg := range packages {
		assert.Equal(t, TypeMas, pkg.Type)
	}
	assert.Equal(t, "Xcode", packages[0].Name)
	assert.Equal(t, "497799835", packages[0].Options["id"])
}

func TestParser_ParseString_VSCode(t *testing.T) {
	content := `
vscode "golang.go"
vscode "ms-python.python"
vscode "vscodevim.vim"
`
	parser := NewParser()
	packages, err := parser.ParseString(content)
	require.NoError(t, err)
	assert.Len(t, packages, 3)

	for _, pkg := range packages {
		assert.Equal(t, TypeVSCode, pkg.Type)
	}
	assert.Equal(t, "golang.go", packages[0].Name)
}

func TestParser_ParseString_Cursor(t *testing.T) {
	content := `
# BrewSync extension: Cursor extensions
cursor "golang.go"
cursor "ms-python.python"
`
	parser := NewParser()
	packages, err := parser.ParseString(content)
	require.NoError(t, err)
	assert.Len(t, packages, 2)

	for _, pkg := range packages {
		assert.Equal(t, TypeCursor, pkg.Type)
	}
}

func TestParser_ParseString_Go(t *testing.T) {
	content := `
# BrewSync extension: Go tools
go "golang.org/x/tools/gopls"
go "github.com/go-delve/delve/cmd/dlv"
`
	parser := NewParser()
	packages, err := parser.ParseString(content)
	require.NoError(t, err)
	assert.Len(t, packages, 2)

	for _, pkg := range packages {
		assert.Equal(t, TypeGo, pkg.Type)
	}
	assert.Equal(t, "golang.org/x/tools/gopls", packages[0].Name)
}

func TestParser_ParseString_Comments(t *testing.T) {
	content := `
# This is a comment
tap "homebrew/bundle"
# Another comment
brew "git"
`
	parser := NewParser()
	packages, err := parser.ParseString(content)
	require.NoError(t, err)
	assert.Len(t, packages, 2)
}

func TestParser_ParseString_EmptyLines(t *testing.T) {
	content := `

tap "homebrew/bundle"

brew "git"

`
	parser := NewParser()
	packages, err := parser.ParseString(content)
	require.NoError(t, err)
	assert.Len(t, packages, 2)
}

func TestParser_ParseString_MixedContent(t *testing.T) {
	content := `
# Taps
tap "homebrew/bundle"

# Formulae
brew "git"
brew "fzf"

# Casks
cask "raycast"

# VSCode
vscode "golang.go"

# Cursor (BrewSync extension)
cursor "ms-python.python"

# Go tools (BrewSync extension)
go "golang.org/x/tools/gopls"

# Mac App Store
mas "Xcode", id: 497799835
`
	parser := NewParser()
	packages, err := parser.ParseString(content)
	require.NoError(t, err)

	byType := packages.ByType()
	assert.Len(t, byType[TypeTap], 1)
	assert.Len(t, byType[TypeBrew], 2)
	assert.Len(t, byType[TypeCask], 1)
	assert.Len(t, byType[TypeVSCode], 1)
	assert.Len(t, byType[TypeCursor], 1)
	assert.Len(t, byType[TypeGo], 1)
	assert.Len(t, byType[TypeMas], 1)
}

func TestParser_ParseFile(t *testing.T) {
	// Create a temp file
	tmpDir := t.TempDir()
	brewfilePath := filepath.Join(tmpDir, "Brewfile")

	content := `tap "homebrew/bundle"
brew "git"
cask "raycast"
`
	err := os.WriteFile(brewfilePath, []byte(content), 0644)
	require.NoError(t, err)

	parser := NewParser()
	packages, err := parser.ParseFile(brewfilePath)
	require.NoError(t, err)
	assert.Len(t, packages, 3)
}

func TestParser_ParseFile_NotFound(t *testing.T) {
	parser := NewParser()
	_, err := parser.ParseFile("/nonexistent/Brewfile")
	assert.Error(t, err)
}

func TestParse_Convenience(t *testing.T) {
	tmpDir := t.TempDir()
	brewfilePath := filepath.Join(tmpDir, "Brewfile")

	content := `brew "git"`
	err := os.WriteFile(brewfilePath, []byte(content), 0644)
	require.NoError(t, err)

	packages, err := Parse(brewfilePath)
	require.NoError(t, err)
	assert.Len(t, packages, 1)
}

func TestParseContent_Convenience(t *testing.T) {
	content := `brew "git"`
	packages, err := ParseContent(content)
	require.NoError(t, err)
	assert.Len(t, packages, 1)
}

func TestParser_ParseOptions_Link(t *testing.T) {
	content := `brew "libpq", link: true`
	packages, err := ParseContent(content)
	require.NoError(t, err)
	assert.Len(t, packages, 1)
	assert.Equal(t, "true", packages[0].Options["link"])
}

func TestParser_ParseOptions_Args(t *testing.T) {
	content := `brew "neovim", args: ["HEAD"]`
	packages, err := ParseContent(content)
	require.NoError(t, err)
	assert.Len(t, packages, 1)
	assert.Contains(t, packages[0].Options["args"], "HEAD")
}

func TestParser_UnrecognizedLines(t *testing.T) {
	content := `
brew "git"
some random text
cask "raycast"
`
	packages, err := ParseContent(content)
	require.NoError(t, err)
	assert.Len(t, packages, 2) // Only valid lines
}
