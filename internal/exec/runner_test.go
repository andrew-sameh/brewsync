package exec

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRunner(t *testing.T) {
	runner := NewRunner()

	assert.NotNil(t, runner)
	assert.Equal(t, DefaultTimeout, runner.Timeout)
	assert.False(t, runner.Verbose)
}

func TestRunner_Run(t *testing.T) {
	runner := NewRunner()

	t.Run("successful command", func(t *testing.T) {
		output, err := runner.Run("echo", "hello")
		assert.NoError(t, err)
		assert.Equal(t, "hello\n", output)
	})

	t.Run("command with multiple args", func(t *testing.T) {
		output, err := runner.Run("echo", "hello", "world")
		assert.NoError(t, err)
		assert.Equal(t, "hello world\n", output)
	})

	t.Run("command not found", func(t *testing.T) {
		_, err := runner.Run("nonexistent-command-12345")
		assert.Error(t, err)
	})

	t.Run("command with error exit", func(t *testing.T) {
		_, err := runner.Run("false")
		assert.Error(t, err)
	})

	t.Run("command with stderr", func(t *testing.T) {
		// Use sh to write to stderr
		_, err := runner.Run("sh", "-c", "echo error >&2; exit 1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error")
	})
}

func TestRunner_RunContext(t *testing.T) {
	runner := NewRunner()

	t.Run("with valid context", func(t *testing.T) {
		ctx := context.Background()
		output, err := runner.RunContext(ctx, "echo", "test")
		assert.NoError(t, err)
		assert.Equal(t, "test\n", output)
	})

	t.Run("with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := runner.RunContext(ctx, "sleep", "10")
		assert.Error(t, err)
	})

	t.Run("with timeout context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := runner.RunContext(ctx, "sleep", "10")
		assert.Error(t, err)
	})
}

func TestRunner_RunLines(t *testing.T) {
	runner := NewRunner()

	t.Run("multiple lines", func(t *testing.T) {
		lines, err := runner.RunLines("printf", "line1\nline2\nline3")
		assert.NoError(t, err)
		assert.Equal(t, []string{"line1", "line2", "line3"}, lines)
	})

	t.Run("single line", func(t *testing.T) {
		lines, err := runner.RunLines("echo", "single")
		assert.NoError(t, err)
		assert.Equal(t, []string{"single"}, lines)
	})

	t.Run("empty output", func(t *testing.T) {
		lines, err := runner.RunLines("true")
		assert.NoError(t, err)
		assert.Nil(t, lines)
	})

	t.Run("command error", func(t *testing.T) {
		_, err := runner.RunLines("false")
		assert.Error(t, err)
	})
}

func TestRunner_Exists(t *testing.T) {
	runner := NewRunner()

	t.Run("existing command", func(t *testing.T) {
		assert.True(t, runner.Exists("echo"))
		assert.True(t, runner.Exists("ls"))
		assert.True(t, runner.Exists("sh"))
	})

	t.Run("non-existing command", func(t *testing.T) {
		assert.False(t, runner.Exists("nonexistent-command-12345"))
	})
}

func TestRunner_Which(t *testing.T) {
	runner := NewRunner()

	t.Run("existing command", func(t *testing.T) {
		path, err := runner.Which("echo")
		assert.NoError(t, err)
		assert.NotEmpty(t, path)
		assert.Contains(t, path, "echo")
	})

	t.Run("non-existing command", func(t *testing.T) {
		_, err := runner.Which("nonexistent-command-12345")
		assert.Error(t, err)
	})
}

func TestDefaultRunner(t *testing.T) {
	assert.NotNil(t, Default)
	assert.Equal(t, DefaultTimeout, Default.Timeout)
}

func TestPackageLevelFunctions(t *testing.T) {
	t.Run("Run", func(t *testing.T) {
		output, err := Run("echo", "test")
		assert.NoError(t, err)
		assert.Equal(t, "test\n", output)
	})

	t.Run("RunLines", func(t *testing.T) {
		lines, err := RunLines("printf", "a\nb")
		assert.NoError(t, err)
		assert.Equal(t, []string{"a", "b"}, lines)
	})

	t.Run("Exists", func(t *testing.T) {
		assert.True(t, Exists("echo"))
		assert.False(t, Exists("nonexistent-12345"))
	})
}

func TestRunner_Timeout(t *testing.T) {
	runner := &Runner{
		Timeout: 50 * time.Millisecond,
	}

	// This should timeout
	_, err := runner.Run("sleep", "10")
	assert.Error(t, err)
}

func TestRunner_RunWithWhitespace(t *testing.T) {
	runner := NewRunner()

	t.Run("leading/trailing whitespace", func(t *testing.T) {
		output, err := runner.Run("echo", "  spaces  ")
		require.NoError(t, err)
		assert.Equal(t, "  spaces  \n", output)
	})

	t.Run("tabs and newlines in output", func(t *testing.T) {
		output, err := runner.Run("printf", "a\tb\nc")
		require.NoError(t, err)
		assert.Equal(t, "a\tb\nc", output)
	})
}
