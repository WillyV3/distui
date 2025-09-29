package views

import (
	"fmt"
	"strings"

	"distui/handlers"
	"github.com/charmbracelet/lipgloss"
)

func RenderRepoBrowser(model *handlers.RepoBrowserModel) string {
	if model == nil {
		return "Loading repository..."
	}

	var content strings.Builder

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true)

	pathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214"))

	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("237"))

	// Header with current path
	content.WriteString(headerStyle.Render("REPOSITORY BROWSER") + "\n")
	content.WriteString(pathStyle.Render(fmt.Sprintf("Path: %s", model.CurrentDirectory)) + "\n")
	content.WriteString(strings.Repeat("─", model.Width-4) + "\n")

	// Column headers
	columnStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Italic(true)
	content.WriteString(columnStyle.Render("T  Name                                     Modified") + "\n")
	content.WriteString(strings.Repeat("─", model.Width-4) + "\n")

	// Handle error
	if model.Error != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
		content.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", model.Error)) + "\n")
		return content.String()
	}

	// Calculate visible entries
	headerLines := 6 // header + path + divider + column headers + divider
	footerLines := 3 // divider + legend + controls
	availableLines := model.Height - headerLines - footerLines
	if availableLines < 1 {
		availableLines = 1
	}

	// Calculate scroll position
	scrollStart := 0
	if model.Selected >= availableLines {
		scrollStart = model.Selected - availableLines + 1
	}
	scrollEnd := scrollStart + availableLines
	if scrollEnd > len(model.Entries) {
		scrollEnd = len(model.Entries)
	}

	// Show entries
	if len(model.Entries) == 0 {
		content.WriteString("(empty directory)\n")
	} else {
		for i := scrollStart; i < scrollEnd; i++ {
			entry := model.Entries[i]
			line := entry.String()

			// Truncate if too long
			maxLen := model.Width - 4
			if len(line) > maxLen && maxLen > 10 {
				line = line[:maxLen-3] + "..."
			}

			if i == model.Selected {
				content.WriteString(selectedStyle.Render(line) + "\n")
			} else {
				content.WriteString(line + "\n")
			}
		}

		// Show scroll indicator if needed
		remaining := len(model.Entries) - scrollEnd
		if len(model.Entries) > availableLines && remaining > 0 {
			content.WriteString(fmt.Sprintf("  ...%d more files below\n", remaining))
		}
	}

	// Fill remaining space
	linesShown := scrollEnd - scrollStart
	remaining := len(model.Entries) - scrollEnd
	if len(model.Entries) > availableLines && remaining > 0 {
		linesShown++
	}
	for i := linesShown; i < availableLines; i++ {
		content.WriteString("\n")
	}

	// Controls and legend
	controlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	legendStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	content.WriteString(strings.Repeat("─", model.Width-4) + "\n")
	content.WriteString(legendStyle.Render("Types: / = dir, g = go, m = md, j = json, y = yaml, t = txt, - = other") + "\n")
	content.WriteString(controlStyle.Render("[j/↓] Down  [k/↑] Up  [l/→/Enter] Open  [h/←/BS] Back  [q/Esc] Exit"))

	return content.String()
}