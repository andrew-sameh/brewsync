package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectMachine(t *testing.T) {
	// Get actual hostname for testing
	actualHostname, err := GetLocalHostname()
	if err != nil {
		t.Skipf("Cannot get local hostname: %v", err)
	}

	t.Run("matching hostname", func(t *testing.T) {
		machines := map[string]Machine{
			"test": {Hostname: actualHostname},
		}

		name, err := DetectMachine(machines)
		assert.NoError(t, err)
		assert.Equal(t, "test", name)
	})

	t.Run("no matching hostname", func(t *testing.T) {
		machines := map[string]Machine{
			"other": {Hostname: "some-other-hostname"},
		}

		_, err := DetectMachine(machines)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no machine found matching hostname")
	})

	t.Run("empty machines map", func(t *testing.T) {
		machines := map[string]Machine{}

		_, err := DetectMachine(machines)
		assert.Error(t, err)
	})

	t.Run("multiple machines one match", func(t *testing.T) {
		machines := map[string]Machine{
			"other1": {Hostname: "hostname1"},
			"match":  {Hostname: actualHostname},
			"other2": {Hostname: "hostname2"},
		}

		name, err := DetectMachine(machines)
		assert.NoError(t, err)
		assert.Equal(t, "match", name)
	})
}

func TestGetLocalHostname(t *testing.T) {
	hostname, err := GetLocalHostname()

	// This should work on macOS
	assert.NoError(t, err)
	assert.NotEmpty(t, hostname)

	// Hostname shouldn't contain newlines
	assert.NotContains(t, hostname, "\n")
}
