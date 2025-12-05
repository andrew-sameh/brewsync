package brewfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllTypes(t *testing.T) {
	types := AllTypes()

	assert.Len(t, types, 7)
	assert.Contains(t, types, TypeTap)
	assert.Contains(t, types, TypeBrew)
	assert.Contains(t, types, TypeCask)
	assert.Contains(t, types, TypeVSCode)
	assert.Contains(t, types, TypeCursor)
	assert.Contains(t, types, TypeGo)
	assert.Contains(t, types, TypeMas)
}

func TestParsePackageType(t *testing.T) {
	testCases := []struct {
		input    string
		expected PackageType
		hasError bool
	}{
		{"tap", TypeTap, false},
		{"TAP", TypeTap, false},
		{"Tap", TypeTap, false},
		{"brew", TypeBrew, false},
		{"BREW", TypeBrew, false},
		{"cask", TypeCask, false},
		{"vscode", TypeVSCode, false},
		{"cursor", TypeCursor, false},
		{"go", TypeGo, false},
		{"mas", TypeMas, false},
		{"invalid", "", true},
		{"", "", true},
		{"homebrew", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := ParsePackageType(tc.input)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestNewPackage(t *testing.T) {
	pkg := NewPackage(TypeBrew, "git")

	assert.Equal(t, TypeBrew, pkg.Type)
	assert.Equal(t, "git", pkg.Name)
	assert.Empty(t, pkg.FullName)
	assert.Nil(t, pkg.Options)
	assert.Empty(t, pkg.Description)
}

func TestPackage_WithOption(t *testing.T) {
	pkg := NewPackage(TypeBrew, "libpq")

	t.Run("add first option", func(t *testing.T) {
		newPkg := pkg.WithOption("link", "true")
		assert.Equal(t, "true", newPkg.Options["link"])
		assert.Nil(t, pkg.Options, "original should be unchanged")
	})

	t.Run("add multiple options", func(t *testing.T) {
		newPkg := pkg.WithOption("link", "true").WithOption("force", "true")
		assert.Equal(t, "true", newPkg.Options["link"])
		assert.Equal(t, "true", newPkg.Options["force"])
	})
}

func TestPackage_ID(t *testing.T) {
	testCases := []struct {
		pkg      Package
		expected string
	}{
		{NewPackage(TypeBrew, "git"), "brew:git"},
		{NewPackage(TypeCask, "raycast"), "cask:raycast"},
		{NewPackage(TypeTap, "homebrew/core"), "tap:homebrew/core"},
		{NewPackage(TypeVSCode, "golang.go"), "vscode:golang.go"},
		{NewPackage(TypeMas, "497799835"), "mas:497799835"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.pkg.ID())
		})
	}
}

func TestPackage_String(t *testing.T) {
	t.Run("without fullname", func(t *testing.T) {
		pkg := NewPackage(TypeBrew, "git")
		assert.Equal(t, "git", pkg.String())
	})

	t.Run("with fullname", func(t *testing.T) {
		pkg := NewPackage(TypeMas, "497799835")
		pkg.FullName = "Xcode"
		assert.Equal(t, "497799835 (Xcode)", pkg.String())
	})
}

func TestPackages_ByType(t *testing.T) {
	packages := Packages{
		NewPackage(TypeBrew, "git"),
		NewPackage(TypeBrew, "fzf"),
		NewPackage(TypeCask, "raycast"),
		NewPackage(TypeTap, "homebrew/core"),
	}

	byType := packages.ByType()

	assert.Len(t, byType[TypeBrew], 2)
	assert.Len(t, byType[TypeCask], 1)
	assert.Len(t, byType[TypeTap], 1)
	assert.Len(t, byType[TypeVSCode], 0)
}

func TestPackages_Filter(t *testing.T) {
	packages := Packages{
		NewPackage(TypeBrew, "git"),
		NewPackage(TypeBrew, "fzf"),
		NewPackage(TypeCask, "raycast"),
		NewPackage(TypeTap, "homebrew/core"),
	}

	t.Run("filter single type", func(t *testing.T) {
		filtered := packages.Filter(TypeBrew)
		assert.Len(t, filtered, 2)
	})

	t.Run("filter multiple types", func(t *testing.T) {
		filtered := packages.Filter(TypeBrew, TypeCask)
		assert.Len(t, filtered, 3)
	})

	t.Run("filter no types returns all", func(t *testing.T) {
		filtered := packages.Filter()
		assert.Len(t, filtered, 4)
	})

	t.Run("filter non-matching type", func(t *testing.T) {
		filtered := packages.Filter(TypeVSCode)
		assert.Len(t, filtered, 0)
	})
}

func TestPackages_Names(t *testing.T) {
	packages := Packages{
		NewPackage(TypeBrew, "git"),
		NewPackage(TypeBrew, "fzf"),
		NewPackage(TypeCask, "raycast"),
	}

	names := packages.Names()
	assert.Equal(t, []string{"git", "fzf", "raycast"}, names)
}

func TestPackages_Contains(t *testing.T) {
	packages := Packages{
		NewPackage(TypeBrew, "git"),
		NewPackage(TypeCask, "raycast"),
	}

	t.Run("contains existing", func(t *testing.T) {
		assert.True(t, packages.Contains("brew:git"))
		assert.True(t, packages.Contains("cask:raycast"))
	})

	t.Run("does not contain", func(t *testing.T) {
		assert.False(t, packages.Contains("brew:fzf"))
		assert.False(t, packages.Contains("cask:git"))
	})
}

func TestPackages_Empty(t *testing.T) {
	var packages Packages

	assert.Len(t, packages.ByType(), 0)
	assert.Len(t, packages.Filter(TypeBrew), 0)
	assert.Len(t, packages.Names(), 0)
	assert.False(t, packages.Contains("brew:git"))
}
