package views

import (
	"fmt"
	"strings"

	"distui/handlers"
	"distui/internal/gitcleanup"
	"github.com/charmbracelet/lipgloss"
)

func RenderFileSelection(m *handlers.FileSelectionModel) string {
	if m == nil || len(m.Files) == 0 {
		return "\nNo files to select.\n\nPress Esc to cancel\n"
	}

	var b strings.Builder

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true)

	selectedStyle := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color("237"))

	autoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46"))

	optionalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214"))

	grayStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))

	b.WriteString("\n")
	b.WriteString(headerStyle.Render("SELECT FILES TO COMMIT") + "\n\n")

	if m.CustomRules {
		b.WriteString(grayStyle.Render("Custom rules mode: All categorized files pre-selected") + "\n")
	} else {
		b.WriteString(grayStyle.Render("Default mode: Go files pre-selected, toggle others with Space") + "\n")
	}
	b.WriteString("\n")

	// Calculate visible range
	maxVisible := 15
	if m.Height > 0 {
		maxVisible = m.Height - 10
		if maxVisible < 5 {
			maxVisible = 5
		}
	}

	scrollStart := 0
	scrollEnd := len(m.Files)

	if len(m.Files) > maxVisible {
		// Center selected item
		scrollStart = m.SelectedIndex - maxVisible/2
		if scrollStart < 0 {
			scrollStart = 0
		}
		scrollEnd = scrollStart + maxVisible
		if scrollEnd > len(m.Files) {
			scrollEnd = len(m.Files)
			scrollStart = scrollEnd - maxVisible
			if scrollStart < 0 {
				scrollStart = 0
			}
		}
	}

	for i := scrollStart; i < scrollEnd; i++ {
		file := m.Files[i]

		// Checkbox
		checkbox := "[X]"
		if !file.Selected {
			checkbox = "[ ]"
		}

		// Category indicator
		categoryLabel := ""
		style := optionalStyle
		if file.IsAuto {
			categoryLabel = " (auto)"
			style = autoStyle
		} else {
			switch file.Category {
			case gitcleanup.CategoryDocs:
				categoryLabel = " (docs)"
			case gitcleanup.CategoryOther:
				categoryLabel = " (other)"
			}
		}

		// File path
		filePath := file.Path
		if len(filePath) > 60 {
			filePath = "..." + filePath[len(filePath)-57:]
		}

		line := fmt.Sprintf("%s %s%s", checkbox, filePath, categoryLabel)

		// Highlight current selection
		if i == m.SelectedIndex {
			line = "> " + line
			b.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			line = "  " + line
			b.WriteString(style.Render(line) + "\n")
		}
	}

	if scrollEnd < len(m.Files) {
		remaining := len(m.Files) - scrollEnd
		b.WriteString(grayStyle.Render(fmt.Sprintf("\n  ...%d more files\n", remaining)))
	}

	b.WriteString("\n")

	// Summary
	selectedCount := 0
	for _, f := range m.Files {
		if f.Selected {
			selectedCount++
		}
	}

	summaryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	b.WriteString(summaryStyle.Render(fmt.Sprintf("%d of %d files selected", selectedCount, len(m.Files))) + "\n\n")

	// Help
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	if m.CustomRules {
		b.WriteString(helpStyle.Render("↑/↓: navigate • space: toggle • enter: commit • esc: cancel"))
	} else {
		b.WriteString(helpStyle.Render("↑/↓: navigate • space: toggle non-auto files • enter: commit • esc: cancel"))
	}

	return b.String()
}
