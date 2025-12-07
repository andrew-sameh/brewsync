package exec

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
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

// RunWithOutput executes a command and streams output to a callback
func (r *Runner) RunWithOutput(name string, args []string, onOutput func(line string)) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()

	return r.RunWithOutputContext(ctx, name, args, onOutput)
}

// RunWithOutputContext executes a command with context and streams output
func (r *Runner) RunWithOutputContext(ctx context.Context, name string, args []string, onOutput func(line string)) error {
	cmd := exec.CommandContext(ctx, name, args...)

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return err
	}

	// Stream both stdout and stderr to the callback
	done := make(chan error, 2)
	go streamLines(stdout, onOutput, done)
	go streamLines(stderr, onOutput, done)

	// Wait for streaming to complete
	<-done
	<-done

	// Wait for command to finish
	return cmd.Wait()
}

// streamLines reads lines from a reader and sends them to the callback
func streamLines(r io.Reader, onOutput func(line string), done chan error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if onOutput != nil {
			onOutput(line)
		}
	}
	done <- scanner.Err()
}
