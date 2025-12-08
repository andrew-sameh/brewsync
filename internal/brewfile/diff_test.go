package brewfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiff_NoChanges(t *testing.T) {
	source := Packages{
		NewPackage(TypeBrew, "git"),
		NewPackage(TypeBrew, "fzf"),
	}
	current := Packages{
		NewPackage(TypeBrew, "git"),
		NewPackage(TypeBrew, "fzf"),
	}

	diff := Diff(source, current)

	assert.True(t, diff.IsEmpty())
	assert.Len(t, diff.Additions, 0)
	assert.Len(t, diff.Removals, 0)
	assert.Len(t, diff.Common, 2)
}

func TestDiff_Additions(t *testing.T) {
	source := Packages{
		NewPackage(TypeBrew, "git"),
		NewPackage(TypeBrew, "fzf"),
		NewPackage(TypeBrew, "bat"),
	}
	current := Packages{
		NewPackage(TypeBrew, "git"),
	}

	diff := Diff(source, current)

	assert.False(t, diff.IsEmpty())
	assert.Len(t, diff.Additions, 2)
	assert.Len(t, diff.Removals, 0)
	assert.Len(t, diff.Common, 1)

	additionNames := diff.Additions.Names()
	assert.Contains(t, additionNames, "fzf")
	assert.Contains(t, additionNames, "bat")
}

func TestDiff_Removals(t *testing.T) {
	source := Packages{
		NewPackage(TypeBrew, "git"),
	}
	current := Packages{
		NewPackage(TypeBrew, "git"),
		NewPackage(TypeBrew, "fzf"),
		NewPackage(TypeBrew, "bat"),
	}

	diff := Diff(source, current)

	assert.False(t, diff.IsEmpty())
	assert.Len(t, diff.Additions, 0)
	assert.Len(t, diff.Removals, 2)
	assert.Len(t, diff.Common, 1)

	removalNames := diff.Removals.Names()
	assert.Contains(t, removalNames, "fzf")
	assert.Contains(t, removalNames, "bat")
}

func TestDiff_MixedChanges(t *testing.T) {
	source := Packages{
		NewPackage(TypeBrew, "git"),
		NewPackage(TypeBrew, "fzf"),
		NewPackage(TypeCask, "raycast"),
	}
	current := Packages{
		NewPackage(TypeBrew, "git"),
		NewPackage(TypeBrew, "bat"),
		NewPackage(TypeCask, "slack"),
	}

	diff := Diff(source, current)

	assert.False(t, diff.IsEmpty())
	assert.Len(t, diff.Additions, 2) // fzf, raycast
	assert.Len(t, diff.Removals, 2)  // bat, slack
	assert.Len(t, diff.Common, 1)    // git
}

func TestDiff_DifferentTypes(t *testing.T) {
	// Same name but different types should be treated as different packages
	source := Packages{
		NewPackage(TypeBrew, "python"),
	}
	current := Packages{
		NewPackage(TypeCask, "python"), // Different type
	}

	diff := Diff(source, current)

	assert.Len(t, diff.Additions, 1)
	assert.Len(t, diff.Removals, 1)
	assert.Len(t, diff.Common, 0)
}

func TestDiff_EmptyLists(t *testing.T) {
	t.Run("both empty", func(t *testing.T) {
		diff := Diff(Packages{}, Packages{})
		assert.True(t, diff.IsEmpty())
	})

	t.Run("source empty", func(t *testing.T) {
		current := Packages{NewPackage(TypeBrew, "git")}
		diff := Diff(Packages{}, current)
		assert.Len(t, diff.Additions, 0)
		assert.Len(t, diff.Removals, 1)
	})

	t.Run("current empty", func(t *testing.T) {
		source := Packages{NewPackage(TypeBrew, "git")}
		diff := Diff(source, Packages{})
		assert.Len(t, diff.Additions, 1)
		assert.Len(t, diff.Removals, 0)
	})
}

func TestDiffByType(t *testing.T) {
	source := Packages{
		NewPackage(TypeTap, "homebrew/bundle"),
		NewPackage(TypeBrew, "git"),
		NewPackage(TypeBrew, "fzf"),
		NewPackage(TypeCask, "raycast"),
		NewPackage(TypeVSCode, "golang.go"),
	}
	current := Packages{
		NewPackage(TypeTap, "homebrew/bundle"),
		NewPackage(TypeBrew, "git"),
		NewPackage(TypeCask, "slack"),
	}

	// Only compare brew and cask
	diff := DiffByType(source, current, []PackageType{TypeBrew, TypeCask})

	// Should not include tap or vscode
	assert.Len(t, diff.Additions, 2) // fzf (brew), raycast (cask)
	assert.Len(t, diff.Removals, 1)  // slack (cask)
	assert.Len(t, diff.Common, 1)    // git (brew)

	// Verify no tap or vscode in results
	for _, pkg := range diff.Additions {
		assert.NotEqual(t, TypeTap, pkg.Type)
		assert.NotEqual(t, TypeVSCode, pkg.Type)
	}
}

func TestDiffResult_AdditionsByType(t *testing.T) {
	diff := &DiffResult{
		Additions: Packages{
			NewPackage(TypeBrew, "git"),
			NewPackage(TypeBrew, "fzf"),
			NewPackage(TypeCask, "raycast"),
		},
	}

	byType := diff.AdditionsByType()
	assert.Len(t, byType[TypeBrew], 2)
	assert.Len(t, byType[TypeCask], 1)
}

func TestDiffResult_RemovalsByType(t *testing.T) {
	diff := &DiffResult{
		Removals: Packages{
			NewPackage(TypeBrew, "bat"),
			NewPackage(TypeCask, "slack"),
		},
	}

	byType := diff.RemovalsByType()
	assert.Len(t, byType[TypeBrew], 1)
	assert.Len(t, byType[TypeCask], 1)
}

func TestDiffResult_FilterIgnored(t *testing.T) {
	diff := &DiffResult{
		Additions: Packages{
			NewPackage(TypeBrew, "git"),
			NewPackage(TypeBrew, "fzf"),
			NewPackage(TypeCask, "slack"),
		},
		Removals: Packages{
			NewPackage(TypeBrew, "bat"),
			NewPackage(TypeCask, "raycast"),
		},
	}

	ignored := map[string]bool{
		"brew:fzf":     true,
		"cask:raycast": true,
	}

	filtered := diff.FilterIgnored(ignored)

	assert.Len(t, filtered.Additions, 2) // git, slack (fzf filtered)
	assert.Len(t, filtered.Removals, 1)  // bat (raycast filtered)
}

func TestDiffResult_FilterMachineSpecific(t *testing.T) {
	diff := &DiffResult{
		Additions: Packages{
			NewPackage(TypeBrew, "git"),
			NewPackage(TypeBrew, "postgresql"),
			NewPackage(TypeCask, "orbstack"),
		},
		Removals: Packages{
			NewPackage(TypeBrew, "bat"),
		},
	}

	machineSpecific := map[string]bool{
		"brew:postgresql": true,
		"cask:orbstack":   true,
	}

	filtered := diff.FilterMachineSpecific(machineSpecific)

	assert.Len(t, filtered.Additions, 1) // git (postgresql, orbstack filtered)
	assert.Len(t, filtered.Removals, 1)  // bat
}

func TestDiffResult_Summary(t *testing.T) {
	t.Run("no differences", func(t *testing.T) {
		diff := &DiffResult{}
		assert.Equal(t, "No differences", diff.Summary())
	})

	t.Run("only additions", func(t *testing.T) {
		diff := &DiffResult{
			Additions: Packages{
				NewPackage(TypeBrew, "git"),
			},
		}
		assert.Contains(t, diff.Summary(), "addition")
	})

	t.Run("only removals", func(t *testing.T) {
		diff := &DiffResult{
			Removals: Packages{
				NewPackage(TypeBrew, "git"),
			},
		}
		assert.Contains(t, diff.Summary(), "removal")
	})

	t.Run("both", func(t *testing.T) {
		diff := &DiffResult{
			Additions: Packages{NewPackage(TypeBrew, "git")},
			Removals:  Packages{NewPackage(TypeBrew, "bat")},
		}
		summary := diff.Summary()
		assert.Contains(t, summary, "addition")
		assert.Contains(t, summary, "removal")
	})
}
