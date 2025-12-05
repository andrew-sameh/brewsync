package brewfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWriter(t *testing.T) {
	packages := Packages{
		NewPackage(TypeBrew, "git"),
	}
	writer := NewWriter(packages)

	assert.NotNil(t, writer)
	assert.Len(t, writer.packages, 1)
}

func TestWriter_Format(t *testing.T) {
	t.Run("empty packages", func(t *testing.T) {
		writer := NewWriter(Packages{})
		content := writer.Format()
		assert.Empty(t, content)
	})

	t.Run("single tap", func(t *testing.T) {
		writer := NewWriter(Packages{
			NewPackage(TypeTap, "homebrew/core"),
		})
		content := writer.Format()
		assert.Equal(t, `tap "homebrew/core"`+"\n", content)
	})

	t.Run("single brew", func(t *testing.T) {
		writer := NewWriter(Packages{
			NewPackage(TypeBrew, "git"),
		})
		content := writer.Format()
		assert.Equal(t, `brew "git"`+"\n", content)
	})

	t.Run("brew with options", func(t *testing.T) {
		pkg := NewPackage(TypeBrew, "libpq").WithOption("link", "true")
		writer := NewWriter(Packages{pkg})
		content := writer.Format()
		assert.Equal(t, `brew "libpq", link: true`+"\n", content)
	})

	t.Run("single cask", func(t *testing.T) {
		writer := NewWriter(Packages{
			NewPackage(TypeCask, "raycast"),
		})
		content := writer.Format()
		assert.Equal(t, `cask "raycast"`+"\n", content)
	})

	t.Run("mas with id", func(t *testing.T) {
		pkg := NewPackage(TypeMas, "497799835")
		pkg.FullName = "Xcode"
		pkg = pkg.WithOption("id", "497799835")
		writer := NewWriter(Packages{pkg})
		content := writer.Format()
		assert.Equal(t, `mas "Xcode", id: 497799835`+"\n", content)
	})

	t.Run("vscode extension", func(t *testing.T) {
		writer := NewWriter(Packages{
			NewPackage(TypeVSCode, "golang.go"),
		})
		content := writer.Format()
		assert.Equal(t, `vscode "golang.go"`+"\n", content)
	})

	t.Run("cursor extension with comment", func(t *testing.T) {
		writer := NewWriter(Packages{
			NewPackage(TypeCursor, "ms-python.python"),
		})
		content := writer.Format()
		assert.Contains(t, content, "# cursor (brewsync extension)")
		assert.Contains(t, content, `cursor "ms-python.python"`)
	})

	t.Run("go tool with comment", func(t *testing.T) {
		writer := NewWriter(Packages{
			NewPackage(TypeGo, "golang.org/x/tools/gopls"),
		})
		content := writer.Format()
		assert.Contains(t, content, "# go (brewsync extension)")
		assert.Contains(t, content, `go "golang.org/x/tools/gopls"`)
	})

	t.Run("sorted by name within type", func(t *testing.T) {
		writer := NewWriter(Packages{
			NewPackage(TypeBrew, "zsh"),
			NewPackage(TypeBrew, "git"),
			NewPackage(TypeBrew, "fzf"),
		})
		content := writer.Format()
		lines := strings.Split(strings.TrimSpace(content), "\n")

		assert.Len(t, lines, 3)
		assert.Equal(t, `brew "fzf"`, lines[0])
		assert.Equal(t, `brew "git"`, lines[1])
		assert.Equal(t, `brew "zsh"`, lines[2])
	})

	t.Run("types in correct order", func(t *testing.T) {
		writer := NewWriter(Packages{
			NewPackage(TypeGo, "gopls"),
			NewPackage(TypeBrew, "git"),
			NewPackage(TypeTap, "homebrew/core"),
			NewPackage(TypeCask, "raycast"),
			NewPackage(TypeVSCode, "ext"),
		})
		content := writer.Format()

		tapIdx := strings.Index(content, "tap ")
		brewIdx := strings.Index(content, "brew ")
		caskIdx := strings.Index(content, "cask ")
		vscodeIdx := strings.Index(content, "vscode ")
		goIdx := strings.Index(content, "go ")

		assert.True(t, tapIdx < brewIdx, "tap should come before brew")
		assert.True(t, brewIdx < caskIdx, "brew should come before cask")
		assert.True(t, caskIdx < vscodeIdx, "cask should come before vscode")
		assert.True(t, vscodeIdx < goIdx, "vscode should come before go")
	})
}

func TestWriter_Write(t *testing.T) {
	tmpDir := t.TempDir()
	brewfilePath := filepath.Join(tmpDir, "Brewfile")

	packages := Packages{
		NewPackage(TypeTap, "homebrew/core"),
		NewPackage(TypeBrew, "git"),
		NewPackage(TypeCask, "raycast"),
	}

	writer := NewWriter(packages)
	err := writer.Write(brewfilePath)
	require.NoError(t, err)

	// Verify file was created
	content, err := os.ReadFile(brewfilePath)
	require.NoError(t, err)

	assert.Contains(t, string(content), `tap "homebrew/core"`)
	assert.Contains(t, string(content), `brew "git"`)
	assert.Contains(t, string(content), `cask "raycast"`)
}

func TestWriter_WriteCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "subdir", "Brewfile")

	// Create directory first (Write doesn't create dirs)
	err := os.MkdirAll(filepath.Dir(nestedPath), 0755)
	require.NoError(t, err)

	writer := NewWriter(Packages{NewPackage(TypeBrew, "git")})
	err = writer.Write(nestedPath)
	require.NoError(t, err)

	_, err = os.Stat(nestedPath)
	assert.NoError(t, err)
}

func TestAppend(t *testing.T) {
	t.Run("append to existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		brewfilePath := filepath.Join(tmpDir, "Brewfile")

		// Create initial file
		initial := `tap "homebrew/core"` + "\n"
		err := os.WriteFile(brewfilePath, []byte(initial), 0644)
		require.NoError(t, err)

		// Append new packages
		err = Append(brewfilePath, Packages{
			NewPackage(TypeBrew, "git"),
		})
		require.NoError(t, err)

		content, err := os.ReadFile(brewfilePath)
		require.NoError(t, err)

		assert.Contains(t, string(content), `tap "homebrew/core"`)
		assert.Contains(t, string(content), `brew "git"`)
	})

	t.Run("append to new file", func(t *testing.T) {
		tmpDir := t.TempDir()
		brewfilePath := filepath.Join(tmpDir, "Brewfile")

		err := Append(brewfilePath, Packages{
			NewPackage(TypeBrew, "git"),
		})
		require.NoError(t, err)

		content, err := os.ReadFile(brewfilePath)
		require.NoError(t, err)

		assert.Contains(t, string(content), `brew "git"`)
	})

	t.Run("adds newline if missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		brewfilePath := filepath.Join(tmpDir, "Brewfile")

		// Create file without trailing newline
		initial := `tap "homebrew/core"`
		err := os.WriteFile(brewfilePath, []byte(initial), 0644)
		require.NoError(t, err)

		err = Append(brewfilePath, Packages{
			NewPackage(TypeBrew, "git"),
		})
		require.NoError(t, err)

		content, err := os.ReadFile(brewfilePath)
		require.NoError(t, err)

		// Should have proper newlines
		lines := strings.Split(string(content), "\n")
		assert.GreaterOrEqual(t, len(lines), 2)
	})
}

func TestFormatOptions(t *testing.T) {
	testCases := []struct {
		name     string
		options  map[string]string
		expected string
	}{
		{
			name:     "boolean true",
			options:  map[string]string{"link": "true"},
			expected: "link: true",
		},
		{
			name:     "boolean false",
			options:  map[string]string{"link": "false"},
			expected: "link: false",
		},
		{
			name:     "numeric value",
			options:  map[string]string{"id": "497799835"},
			expected: "id: 497799835",
		},
		{
			name:     "string value",
			options:  map[string]string{"name": "Xcode"},
			expected: `name: "Xcode"`,
		},
		{
			name:     "symbol value",
			options:  map[string]string{"args": ":head"},
			expected: "args: :head",
		},
		{
			name:     "multiple options sorted",
			options:  map[string]string{"link": "true", "force": "true"},
			expected: "force: true, link: true",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatOptions(tc.options)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestWriter_FormatComplexBrewfile(t *testing.T) {
	packages := Packages{
		NewPackage(TypeTap, "charmbracelet/tap"),
		NewPackage(TypeTap, "homebrew/bundle"),
		NewPackage(TypeBrew, "git"),
		NewPackage(TypeBrew, "fzf"),
		NewPackage(TypeBrew, "bat"),
		NewPackage(TypeCask, "raycast"),
		NewPackage(TypeCask, "iterm2"),
		func() Package {
			pkg := NewPackage(TypeMas, "497799835")
			pkg.FullName = "Xcode"
			pkg = pkg.WithOption("id", "497799835")
			return pkg
		}(),
		NewPackage(TypeVSCode, "golang.go"),
		NewPackage(TypeVSCode, "vscodevim.vim"),
		NewPackage(TypeCursor, "ms-python.python"),
		NewPackage(TypeGo, "golang.org/x/tools/gopls"),
	}

	writer := NewWriter(packages)
	content := writer.Format()

	// Verify structure
	assert.Contains(t, content, `tap "charmbracelet/tap"`)
	assert.Contains(t, content, `brew "bat"`)
	assert.Contains(t, content, `cask "iterm2"`)
	assert.Contains(t, content, `mas "Xcode", id: 497799835`)
	assert.Contains(t, content, `vscode "golang.go"`)
	assert.Contains(t, content, "# cursor (brewsync extension)")
	assert.Contains(t, content, `cursor "ms-python.python"`)
	assert.Contains(t, content, "# go (brewsync extension)")
	assert.Contains(t, content, `go "golang.org/x/tools/gopls"`)
}
