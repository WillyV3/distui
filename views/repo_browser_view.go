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

	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("237"))

	// Header
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true)
	content.WriteString(titleStyle.Render("REPOSITORY BROWSER") + "\n")

	// Current path
	pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	content.WriteString(pathStyle.Render(model.CurrentDirectory) + "\n")

	// Type legend (compact)
	legendStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	legendText := "/ dir  b bin  g go  m md  j json  y yaml  t txt  - other"
	content.WriteString(legendStyle.Render(legendText) + "\n")
	content.WriteString(strings.Repeat("─", model.Width-4) + "\n")

	// Column headers
	columnStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Bold(true)
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
	headerLines := 6 // title + path + legend + divider + headers + divider
	footerLines := 2 // divider + controls
	availableLines := model.Height - headerLines - footerLines
	if availableLines < 1 {
		availableLines = 1
	}

	// Use viewport offset from model
	scrollStart := model.ViewportOffset
	scrollEnd := scrollStart + availableLines

	// Clamp to valid range
	if scrollEnd > len(model.Entries) {
		scrollEnd = len(model.Entries)
	}

	// Check if we need to show the "more items" indicator
	showMoreIndicator := scrollEnd < len(model.Entries)

	// If showing indicator, reduce visible items by 1 to make room
	if showMoreIndicator && scrollEnd > scrollStart {
		scrollEnd--
	}

	// Define colors for file types
	dirStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))    // Blue for directories
	goStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("51"))     // Cyan for Go files
	mdStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))    // Orange for Markdown
	jsonStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))  // Yellow for JSON
	yamlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141"))  // Purple for YAML
	txtStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))   // Light gray for text
	binStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))    // Green for binaries
	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244")) // Gray for others

	// Show entries
	if len(model.Entries) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		content.WriteString(emptyStyle.Render("   (empty directory)") + "\n")
	} else {
		for i := scrollStart; i < scrollEnd; i++ {
			entry := model.Entries[i]
			line := entry.String()

			// Truncate if too long
			maxLen := model.Width - 4
			if len(line) > maxLen && maxLen > 10 {
				line = line[:maxLen-3] + "..."
			}

			// Apply color based on file type
			var styledLine string
			if entry.IsDir {
				styledLine = dirStyle.Render(line)
			} else if strings.HasSuffix(entry.Name, ".go") {
				styledLine = goStyle.Render(line)
			} else if strings.HasSuffix(entry.Name, ".md") {
				styledLine = mdStyle.Render(line)
			} else if strings.HasSuffix(entry.Name, ".json") {
				styledLine = jsonStyle.Render(line)
			} else if strings.HasSuffix(entry.Name, ".yaml") || strings.HasSuffix(entry.Name, ".yml") {
				styledLine = yamlStyle.Render(line)
			} else if strings.HasSuffix(entry.Name, ".txt") {
				styledLine = txtStyle.Render(line)
			} else if entry.Mode&0111 != 0 && !strings.Contains(entry.Name, ".") {
				styledLine = binStyle.Render(line)
			} else {
				styledLine = defaultStyle.Render(line)
			}

			if i == model.Selected {
				content.WriteString(selectedStyle.Render(line) + "\n")
			} else {
				content.WriteString(styledLine + "\n")
			}
		}

		// Show scroll indicator if needed
		if showMoreIndicator {
			remaining := len(model.Entries) - scrollEnd
			content.WriteString(fmt.Sprintf("  ...%d more files below\n", remaining))
		}
	}

	// Fill remaining space
	linesShown := scrollEnd - scrollStart
	if showMoreIndicator {
		linesShown++ // Account for the indicator line
	}
	for i := linesShown; i < availableLines; i++ {
		content.WriteString("\n")
	}

	// Footer section
	dividerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	content.WriteString(dividerStyle.Render(strings.Repeat("─", model.Width-4)) + "\n")

	// Controls
	controlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	navControls := "[j/↓] Down  [k/↑] Up  [l/→/Enter] Open  [h/←/BS] Back  [q/Esc] Exit"
	content.WriteString(controlStyle.Render(navControls))

	return content.String()
}