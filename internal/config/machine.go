package config

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetLocalHostname returns the local hostname using scutil
func GetLocalHostname() (string, error) {
	cmd := exec.Command("scutil", "--get", "LocalHostName")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// DetectMachine attempts to detect the current machine based on hostname
func DetectMachine(machines map[string]Machine) (string, error) {
	hostname, err := GetLocalHostname()
	if err != nil {
		return "", err
	}

	for name, machine := range machines {
		if machine.Hostname == hostname {
			return name, nil
		}
	}

	return "", fmt.Errorf("no machine found matching hostname %q", hostname)
}
