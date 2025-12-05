package version

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
	// Default version should be "dev"
	assert.Equal(t, "dev", Info())
}

func TestFull(t *testing.T) {
	full := Full()

	// Should contain version, commit, and build date
	assert.Contains(t, full, Version)
	assert.Contains(t, full, Commit)
	assert.Contains(t, full, BuildDate)
	assert.Contains(t, full, "built")
}

func TestFullFormat(t *testing.T) {
	// Test with custom values
	origVersion, origCommit, origBuildDate := Version, Commit, BuildDate
	defer func() {
		Version, Commit, BuildDate = origVersion, origCommit, origBuildDate
	}()

	Version = "1.0.0"
	Commit = "abc123"
	BuildDate = "2024-01-01"

	full := Full()
	assert.Equal(t, "1.0.0 (abc123) built 2024-01-01", full)
}

func TestVersionVariables(t *testing.T) {
	// Verify default values
	assert.Equal(t, "dev", Version)
	assert.Equal(t, "unknown", Commit)
	assert.Equal(t, "unknown", BuildDate)
}

func TestInfoReturnsVersion(t *testing.T) {
	origVersion := Version
	defer func() { Version = origVersion }()

	testCases := []string{"1.0.0", "2.5.3", "0.0.1-beta"}
	for _, tc := range testCases {
		Version = tc
		assert.Equal(t, tc, Info(), "Info() should return the current Version")
	}
}

func TestFullContainsAllParts(t *testing.T) {
	full := Full()
	parts := strings.Split(full, " ")

	// Should have format: "version (commit) built date"
	assert.GreaterOrEqual(t, len(parts), 4, "Full() should have at least 4 parts")
}
