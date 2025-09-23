package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderPage1Content returns the content for page1
func RenderPage1Content() string {
	listItemStyle := lipgloss.NewStyle().Padding(0, 1).MarginLeft(2)
	selectedItemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true).
		Padding(0, 1).
		MarginLeft(2)
	subtleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	items := []string{
		"Sample Item One",
		"Sample Item Two",
		"Sample Item Three",
		"Sample Item Four",
		"Sample Item Five",
	}

	var content strings.Builder
	content.WriteString("This is the first sample page:\n\n")

	for i, item := range items {
		style := listItemStyle
		if i == 0 {
			style = selectedItemStyle
		}
		content.WriteString(style.Render(fmt.Sprintf("• %s", item)) + "\n")
	}

	content.WriteString("\n" + subtleStyle.Render("This page demonstrates a simple list display"))
	content.WriteString("\n\n" + subtleStyle.Render("esc: back to home • q: quit"))

	return content.String()
}