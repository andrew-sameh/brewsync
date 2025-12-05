package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultCategories(t *testing.T) {
	expected := []string{"tap", "brew", "cask", "vscode", "cursor", "go", "mas"}
	assert.Equal(t, expected, DefaultCategories)
}

func TestDefaultCategories_ContainsAllTypes(t *testing.T) {
	// Ensure all expected package types are in defaults
	expectedTypes := []string{"tap", "brew", "cask", "vscode", "cursor", "go", "mas"}

	for _, expectedType := range expectedTypes {
		assert.Contains(t, DefaultCategories, expectedType,
			"DefaultCategories should contain %s", expectedType)
	}
}

func TestDefaultCommitMessage(t *testing.T) {
	assert.Equal(t, "brewsync: update {machine} Brewfile", DefaultCommitMessage)
	assert.Contains(t, DefaultCommitMessage, "{machine}")
}
