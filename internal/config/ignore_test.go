package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadIgnoreFile_NotExists(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Override config directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	ignoreFile, err := LoadIgnoreFile()
	require.NoError(t, err)
	assert.NotNil(t, ignoreFile)
	assert.Empty(t, ignoreFile.Global.Categories)
	assert.NotNil(t, ignoreFile.Machines)
}

func TestLoadIgnoreFile_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	ignorePath := filepath.Join(tmpDir, "ignore.yaml")

	// Create a test ignore file
	content := `global:
  categories:
    - mas
    - go
  packages:
    cask:
      - bluestacks
machines:
  mini:
    categories:
      - antigravity
    packages:
      brew:
        - postgresql
`
	err := os.WriteFile(ignorePath, []byte(content), 0644)
	require.NoError(t, err)

	// Override IgnorePath for testing
	SetIgnorePath(ignorePath)
	defer SetIgnorePath("")

	ignoreFile, err := LoadIgnoreFile()
	require.NoError(t, err)
	assert.Equal(t, []string{"mas", "go"}, ignoreFile.Global.Categories)
	assert.Equal(t, []string{"bluestacks"}, ignoreFile.Global.Packages.Cask)
	assert.Equal(t, []string{"antigravity"}, ignoreFile.Machines["mini"].Categories)
	assert.Equal(t, []string{"postgresql"}, ignoreFile.Machines["mini"].Packages.Brew)
}

func TestSaveIgnoreFile(t *testing.T) {
	tmpDir := t.TempDir()
	ignorePath := filepath.Join(tmpDir, "ignore.yaml")

	ignoreFile := &IgnoreFile{
		Global: IgnoreConfig{
			Categories: []string{"mas"},
			Packages: PackageIgnoreList{
				Cask: []string{"app1", "app2"},
			},
		},
		Machines: map[string]IgnoreConfig{
			"mini": {
				Categories: []string{"go"},
				Packages: PackageIgnoreList{
					Brew: []string{"tool1"},
				},
			},
		},
	}

	// Override IgnorePath for testing
	SetIgnorePath(ignorePath)
	defer SetIgnorePath("")

	err := SaveIgnoreFile(ignoreFile)
	require.NoError(t, err)

	// Verify file was created
	assert.FileExists(t, ignorePath)

	// Load it back and verify
	loaded, err := LoadIgnoreFile()
	require.NoError(t, err)
	assert.Equal(t, ignoreFile.Global.Categories, loaded.Global.Categories)
	assert.Equal(t, ignoreFile.Global.Packages.Cask, loaded.Global.Packages.Cask)
}

func TestCreateDefaultIgnoreFile(t *testing.T) {
	tmpDir := t.TempDir()
	ignorePath := filepath.Join(tmpDir, "ignore.yaml")

	// Override IgnorePath for testing
	SetIgnorePath(ignorePath)
	defer SetIgnorePath("")

	err := CreateDefaultIgnoreFile()
	require.NoError(t, err)
	assert.FileExists(t, ignorePath)

	// Load and verify structure
	loaded, err := LoadIgnoreFile()
	require.NoError(t, err)
	assert.Empty(t, loaded.Global.Categories)
	assert.NotNil(t, loaded.Global.Packages)
	assert.NotNil(t, loaded.Machines)
}

func TestAddCategoryIgnore_Global(t *testing.T) {
	tmpDir := t.TempDir()
	ignorePath := filepath.Join(tmpDir, "ignore.yaml")

	SetIgnorePath(ignorePath)
	defer func() { SetIgnorePath("") }()

	// Add global category ignore
	err := AddCategoryIgnore("", "mas", true)
	require.NoError(t, err)

	// Verify
	loaded, err := LoadIgnoreFile()
	require.NoError(t, err)
	assert.Contains(t, loaded.Global.Categories, "mas")
}

func TestAddCategoryIgnore_Machine(t *testing.T) {
	tmpDir := t.TempDir()
	ignorePath := filepath.Join(tmpDir, "ignore.yaml")

	SetIgnorePath(ignorePath)
	defer func() { SetIgnorePath("") }()

	// Add machine-specific category ignore
	err := AddCategoryIgnore("mini", "go", false)
	require.NoError(t, err)

	// Verify
	loaded, err := LoadIgnoreFile()
	require.NoError(t, err)
	assert.Contains(t, loaded.Machines["mini"].Categories, "go")
}

func TestRemoveCategoryIgnore(t *testing.T) {
	tmpDir := t.TempDir()
	ignorePath := filepath.Join(tmpDir, "ignore.yaml")

	SetIgnorePath(ignorePath)
	defer func() { SetIgnorePath("") }()

	// Add then remove
	err := AddCategoryIgnore("", "mas", true)
	require.NoError(t, err)

	err = RemoveCategoryIgnore("", "mas", true)
	require.NoError(t, err)

	// Verify removed
	loaded, err := LoadIgnoreFile()
	require.NoError(t, err)
	assert.NotContains(t, loaded.Global.Categories, "mas")
}

func TestAddPackageIgnore_Global(t *testing.T) {
	tmpDir := t.TempDir()
	ignorePath := filepath.Join(tmpDir, "ignore.yaml")

	SetIgnorePath(ignorePath)
	defer func() { SetIgnorePath("") }()

	// Add global package ignore
	err := AddPackageIgnore("", "cask:bluestacks", true)
	require.NoError(t, err)

	// Verify
	loaded, err := LoadIgnoreFile()
	require.NoError(t, err)
	assert.Contains(t, loaded.Global.Packages.Cask, "bluestacks")
}

func TestAddPackageIgnore_Machine(t *testing.T) {
	tmpDir := t.TempDir()
	ignorePath := filepath.Join(tmpDir, "ignore.yaml")

	SetIgnorePath(ignorePath)
	defer func() { SetIgnorePath("") }()

	// Add machine-specific package ignore
	err := AddPackageIgnore("mini", "brew:postgresql", false)
	require.NoError(t, err)

	// Verify
	loaded, err := LoadIgnoreFile()
	require.NoError(t, err)
	assert.Contains(t, loaded.Machines["mini"].Packages.Brew, "postgresql")
}

func TestRemovePackageIgnore(t *testing.T) {
	tmpDir := t.TempDir()
	ignorePath := filepath.Join(tmpDir, "ignore.yaml")

	SetIgnorePath(ignorePath)
	defer func() { SetIgnorePath("") }()

	// Add then remove
	err := AddPackageIgnore("", "cask:bluestacks", true)
	require.NoError(t, err)

	err = RemovePackageIgnore("", "cask:bluestacks", true)
	require.NoError(t, err)

	// Verify removed
	loaded, err := LoadIgnoreFile()
	require.NoError(t, err)
	assert.NotContains(t, loaded.Global.Packages.Cask, "bluestacks")
}

func TestAddPackageIgnore_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	ignorePath := filepath.Join(tmpDir, "ignore.yaml")

	SetIgnorePath(ignorePath)
	defer func() { SetIgnorePath("") }()

	// Invalid format (no colon)
	err := AddPackageIgnore("", "bluestacks", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid package ID format")
}

func TestAddPackageIgnore_Duplicate(t *testing.T) {
	tmpDir := t.TempDir()
	ignorePath := filepath.Join(tmpDir, "ignore.yaml")

	SetIgnorePath(ignorePath)
	defer func() { SetIgnorePath("") }()

	// Add package twice
	err := AddPackageIgnore("", "cask:app", true)
	require.NoError(t, err)

	// Second add should not duplicate
	err = AddPackageIgnore("", "cask:app", true)
	require.NoError(t, err)

	// Verify only one entry
	loaded, err := LoadIgnoreFile()
	require.NoError(t, err)
	count := 0
	for _, pkg := range loaded.Global.Packages.Cask {
		if pkg == "app" {
			count++
		}
	}
	assert.Equal(t, 1, count, "Should only have one entry")
}

func TestAddCategoryIgnore_Duplicate(t *testing.T) {
	tmpDir := t.TempDir()
	ignorePath := filepath.Join(tmpDir, "ignore.yaml")

	SetIgnorePath(ignorePath)
	defer func() { SetIgnorePath("") }()

	// Add category twice
	err := AddCategoryIgnore("", "mas", true)
	require.NoError(t, err)

	err = AddCategoryIgnore("", "mas", true)
	require.NoError(t, err)

	// Verify only one entry
	loaded, err := LoadIgnoreFile()
	require.NoError(t, err)
	count := 0
	for _, cat := range loaded.Global.Categories {
		if cat == "mas" {
			count++
		}
	}
	assert.Equal(t, 1, count, "Should only have one entry")
}
