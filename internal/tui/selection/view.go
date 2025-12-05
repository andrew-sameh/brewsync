package selection

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/andrew-sameh/brewsync/internal/tui/styles"
)

// render renders the full UI
func (m Model) render() string {
	var b strings.Builder

	// Title
	b.WriteString(styles.TitleStyle.Render(m.title))
	b.WriteString("\n\n")

	// Tabs
	b.WriteString(m.renderTabs())
	b.WriteString("\n\n")

	// Search bar (if searching)
	if m.searching {
		b.WriteString("Search: ")
		b.WriteString(m.searchText.View())
		b.WriteString("\n\n")
	}

	// Package list
	b.WriteString(m.renderList())
	b.WriteString("\n")

	// Status line
	b.WriteString(m.renderStatus())
	b.WriteString("\n\n")

	// Help
	if m.showHelp {
		b.WriteString(m.help.View(m.keys))
	} else {
		b.WriteString(m.renderShortHelp())
	}

	return b.String()
}

// renderTabs renders the category tabs
func (m Model) renderTabs() string {
	counts := m.countByCategory()
	var tabs []string

	for _, cat := range AllCategories() {
		c := counts[cat]
		label := fmt.Sprintf("%s (%d/%d)", cat, c.selected, c.total)

		if cat == m.category {
			tabs = append(tabs, styles.ActiveTabStyle.Render(label))
		} else {
			tabs = append(tabs, styles.InactiveTabStyle.Render(label))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

// renderList renders the package list
func (m Model) renderList() string {
	if len(m.filtered) == 0 {
		return styles.DimmedStyle.Render("  No packages in this category")
	}

	var lines []string

	// Calculate visible range
	visibleHeight := m.height - 12 // Account for header, tabs, status, help
	if visibleHeight < 5 {
		visibleHeight = 5
	}

	start := 0
	end := len(m.filtered)

	// Scroll window around cursor
	if len(m.filtered) > visibleHeight {
		half := visibleHeight / 2
		start = m.cursor - half
		if start < 0 {
			start = 0
		}
		end = start + visibleHeight
		if end > len(m.filtered) {
			end = len(m.filtered)
			start = end - visibleHeight
			if start < 0 {
				start = 0
			}
		}
	}

	// Show scroll indicator at top
	if start > 0 {
		lines = append(lines, styles.DimmedStyle.Render(fmt.Sprintf("  ↑ %d more above", start)))
	}

	// Render visible items
	for i := start; i < end; i++ {
		idx := m.filtered[i]
		item := m.items[idx]
		lines = append(lines, m.renderItem(item, i == m.cursor))
	}

	// Show scroll indicator at bottom
	if end < len(m.filtered) {
		lines = append(lines, styles.DimmedStyle.Render(fmt.Sprintf("  ↓ %d more below", len(m.filtered)-end)))
	}

	return strings.Join(lines, "\n")
}

// renderItem renders a single package item
func (m Model) renderItem(item Item, isCursor bool) string {
	var b strings.Builder

	// Cursor
	if isCursor {
		b.WriteString(styles.CursorStyle.Render("> "))
	} else {
		b.WriteString("  ")
	}

	// Checkbox
	if item.Ignored {
		b.WriteString(styles.DimmedStyle.Render("[-]"))
	} else if item.Selected {
		b.WriteString(styles.SelectedStyle.Render("[x]"))
	} else {
		b.WriteString("[ ]")
	}

	b.WriteString(" ")

	// Package type badge
	typeStyle := styles.GetCategoryStyle(string(item.Package.Type))
	b.WriteString(typeStyle.Render(fmt.Sprintf("%-6s", item.Package.Type)))
	b.WriteString(" ")

	// Package name
	name := item.Package.Name
	if len(name) > 50 {
		name = name[:47] + "..."
	}

	if item.Ignored {
		b.WriteString(styles.IgnoredStyle.Render(name))
	} else if isCursor {
		b.WriteString(styles.CursorStyle.Render(name))
	} else if item.Selected {
		b.WriteString(styles.SelectedStyle.Render(name))
	} else {
		b.WriteString(name)
	}

	return b.String()
}

// renderStatus renders the status line
func (m Model) renderStatus() string {
	selected := 0
	ignored := 0
	for _, item := range m.items {
		if item.Selected && !item.Ignored {
			selected++
		}
		if item.Ignored {
			ignored++
		}
	}

	status := fmt.Sprintf("Selected: %d | Ignored: %d | Total: %d",
		selected, ignored, len(m.items))

	return styles.SubtitleStyle.Render(status)
}

// renderShortHelp renders the short help line
func (m Model) renderShortHelp() string {
	keys := []string{
		"space:toggle",
		"a:all",
		"n:none",
		"i:ignore",
		"/:search",
		"1-7:tabs",
		"enter:confirm",
		"q:quit",
		"?:help",
	}

	return styles.HelpStyle.Render(strings.Join(keys, " | "))
}
