package styles

import "github.com/charmbracelet/lipgloss"

// Colors
var (
	PrimaryColor   = lipgloss.Color("212") // Blue
	SuccessColor   = lipgloss.Color("42")  // Green
	WarningColor   = lipgloss.Color("214") // Yellow
	ErrorColor     = lipgloss.Color("196") // Red
	MutedColor     = lipgloss.Color("241") // Gray
	HighlightColor = lipgloss.Color("99")  // Purple
)

// Base styles
var (
	// Title style for headers
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			MarginBottom(1)

	// Subtitle style
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			MarginBottom(1)

	// Box style for panels
	BoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(MutedColor).
			Padding(1, 2)

	// Selected item style
	SelectedStyle = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true)

	// Cursor style for current item
	CursorStyle = lipgloss.NewStyle().
			Foreground(HighlightColor).
			Bold(true)

	// Dimmed style for inactive items
	DimmedStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	// Added package style (+)
	AddedStyle = lipgloss.NewStyle().
			Foreground(SuccessColor)

	// Removed package style (-)
	RemovedStyle = lipgloss.NewStyle().
			Foreground(ErrorColor)

	// Warning style
	WarningStyle = lipgloss.NewStyle().
			Foreground(WarningColor)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)

	// Ignored package style
	IgnoredStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			Strikethrough(true)

	// Help style for keybindings
	HelpStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	// Category tab styles
	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(PrimaryColor).
			Padding(0, 1)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(MutedColor).
				Padding(0, 1)

	// Progress bar styles
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor)

	ProgressCompleteStyle = lipgloss.NewStyle().
				Foreground(SuccessColor)

	// Status indicators
	CheckmarkStyle = lipgloss.NewStyle().
			Foreground(SuccessColor).
			SetString("✓")

	CrossStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			SetString("✗")

	WarningMarkStyle = lipgloss.NewStyle().
				Foreground(WarningColor).
				SetString("⚠")

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor)
)

// Symbols
const (
	CheckedBox   = "[x]"
	UncheckedBox = "[ ]"
	Cursor       = ">"
	NoCursor     = " "
	Separator    = "─"
)

// CategoryColors maps package types to colors
var CategoryColors = map[string]lipgloss.Color{
	"tap":    lipgloss.Color("212"), // Blue
	"brew":   lipgloss.Color("42"),  // Green
	"cask":   lipgloss.Color("214"), // Yellow
	"vscode": lipgloss.Color("99"),  // Purple
	"cursor": lipgloss.Color("135"), // Light purple
	"go":     lipgloss.Color("39"),  // Cyan
	"mas":    lipgloss.Color("196"), // Red
}

// GetCategoryStyle returns a style for the given package type
func GetCategoryStyle(pkgType string) lipgloss.Style {
	if color, ok := CategoryColors[pkgType]; ok {
		return lipgloss.NewStyle().Foreground(color)
	}
	return lipgloss.NewStyle().Foreground(MutedColor)
}

// RenderCheckbox renders a checkbox with the given state
func RenderCheckbox(checked bool) string {
	if checked {
		return SelectedStyle.Render(CheckedBox)
	}
	return UncheckedBox
}

// RenderCursor renders a cursor or empty space
func RenderCursor(active bool) string {
	if active {
		return CursorStyle.Render(Cursor)
	}
	return NoCursor
}

// RenderDiff renders a diff indicator (+/-)
func RenderDiff(isAdd bool) string {
	if isAdd {
		return AddedStyle.Render("+")
	}
	return RemovedStyle.Render("-")
}
