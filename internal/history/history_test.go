package history

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatEntry(t *testing.T) {
	entry := Entry{
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Operation: OpDump,
		Machine:   "mini",
		Details:   "tap:5,brew:10",
		Summary:   "completed",
	}

	formatted := formatEntry(entry)

	assert.Contains(t, formatted, "2024-01-15T10:30:00Z")
	assert.Contains(t, formatted, "dump")
	assert.Contains(t, formatted, "mini")
	assert.Contains(t, formatted, "tap:5,brew:10")
	assert.Contains(t, formatted, "completed")
}

func TestParseEntry(t *testing.T) {
	line := "2024-01-15T10:30:00Z|dump|mini|tap:5,brew:10|completed"

	entry, err := parseEntry(line)

	assert.NoError(t, err)
	assert.Equal(t, OpDump, entry.Operation)
	assert.Equal(t, "mini", entry.Machine)
	assert.Equal(t, "tap:5,brew:10", entry.Details)
	assert.Equal(t, "completed", entry.Summary)
}

func TestParseEntry_Invalid(t *testing.T) {
	testCases := []string{
		"",
		"invalid",
		"a|b|c", // Too few parts
		"invalid-time|dump|mini|details|summary",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			_, err := parseEntry(tc)
			assert.Error(t, err)
		})
	}
}

func TestEntryFormat(t *testing.T) {
	entry := Entry{
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Operation: OpDump,
		Machine:   "mini",
		Details:   "tap:5,brew:10",
		Summary:   "completed",
	}

	t.Run("short format", func(t *testing.T) {
		short := entry.Format(false)
		assert.Contains(t, short, "2024-01-15")
		assert.Contains(t, short, "dump")
		assert.Contains(t, short, "mini")
		assert.Contains(t, short, "completed")
		assert.NotContains(t, short, "tap:5,brew:10") // Details not in short format
	})

	t.Run("detailed format", func(t *testing.T) {
		detailed := entry.Format(true)
		assert.Contains(t, detailed, "2024-01-15")
		assert.Contains(t, detailed, "dump")
		assert.Contains(t, detailed, "mini")
		assert.Contains(t, detailed, "tap:5,brew:10")
		assert.Contains(t, detailed, "completed")
	})
}

func TestParseCounts(t *testing.T) {
	t.Run("valid counts", func(t *testing.T) {
		counts := ParseCounts("tap:6,brew:85,cask:42")

		assert.Equal(t, 6, counts["tap"])
		assert.Equal(t, 85, counts["brew"])
		assert.Equal(t, 42, counts["cask"])
	})

	t.Run("empty string", func(t *testing.T) {
		counts := ParseCounts("")
		assert.Empty(t, counts)
	})

	t.Run("invalid values", func(t *testing.T) {
		counts := ParseCounts("tap:invalid,brew:10")

		assert.Equal(t, 0, counts["tap"]) // Invalid skipped
		assert.Equal(t, 10, counts["brew"])
	})

	t.Run("single count", func(t *testing.T) {
		counts := ParseCounts("brew:100")
		assert.Equal(t, 100, counts["brew"])
	})
}

func TestOperationConstants(t *testing.T) {
	assert.Equal(t, Operation("dump"), OpDump)
	assert.Equal(t, Operation("import"), OpImport)
	assert.Equal(t, Operation("sync"), OpSync)
	assert.Equal(t, Operation("ignore"), OpIgnore)
	assert.Equal(t, Operation("profile"), OpProfile)
}

func TestFormatAndParse_Roundtrip(t *testing.T) {
	original := Entry{
		Timestamp: time.Now().Truncate(time.Second), // Truncate to match RFC3339 precision
		Operation: OpSync,
		Machine:   "air",
		Details:   "←mini,+5,-2", // Use comma instead of pipe to avoid delimiter conflict
		Summary:   "applied",
	}

	formatted := formatEntry(original)
	parsed, err := parseEntry(formatted)

	assert.NoError(t, err)
	assert.Equal(t, original.Operation, parsed.Operation)
	assert.Equal(t, original.Machine, parsed.Machine)
	assert.Equal(t, original.Details, parsed.Details)
	assert.Equal(t, original.Summary, parsed.Summary)
	assert.WithinDuration(t, original.Timestamp, parsed.Timestamp, time.Second)
}
