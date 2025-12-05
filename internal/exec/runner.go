package exec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// DefaultTimeout is the default command timeout
const DefaultTimeout = 5 * time.Minute

// Runner executes shell commands
type Runner struct {
	Timeout time.Duration
	Verbose bool
}

// NewRunner creates a new command runner
func NewRunner() *Runner {
	return &Runner{
		Timeout: DefaultTimeout,
	}
}

// Run executes a command and returns its output
func (r *Runner) Run(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()

	return r.RunContext(ctx, name, args...)
}

// RunContext executes a command with the given context
func (r *Runner) RunContext(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Include stderr in error message for debugging
		errMsg := stderr.String()
		if errMsg != "" {
			return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(errMsg))
		}
		return "", err
	}

	return stdout.String(), nil
}

// RunLines executes a command and returns output as lines
func (r *Runner) RunLines(name string, args ...string) ([]string, error) {
	output, err := r.Run(name, args...)
	if err != nil {
		return nil, err
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return nil, nil
	}

	return strings.Split(output, "\n"), nil
}

// Exists checks if a command exists in PATH
func (r *Runner) Exists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// Which returns the path to a command
func (r *Runner) Which(name string) (string, error) {
	return exec.LookPath(name)
}

// Default is the default runner instance
var Default = NewRunner()

// Run executes a command using the default runner
func Run(name string, args ...string) (string, error) {
	return Default.Run(name, args...)
}

// RunLines executes a command and returns output as lines using the default runner
func RunLines(name string, args ...string) ([]string, error) {
	return Default.RunLines(name, args...)
}

// Exists checks if a command exists using the default runner
func Exists(name string) bool {
	return Default.Exists(name)
}
