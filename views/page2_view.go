package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderPage2Content returns the content for page2
func RenderPage2Content() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true).
		Padding(0, 1).
		MarginLeft(2)

	itemStyle := lipgloss.NewStyle().
		Padding(0, 1).
		MarginLeft(4)

	subtleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	categories := []struct {
		name  string
		items []string
	}{
		{
			name: "Category A",
			items: []string{
				"Item A1 - Description for first item",
				"Item A2 - Description for second item",
				"Item A3 - Description for third item",
			},
		},
		{
			name: "Category B",
			items: []string{
				"Item B1 - Another sample item",
				"Item B2 - Yet another item",
			},
		},
	}

	var content strings.Builder
	content.WriteString("This is the second sample page with categories:\n\n")

	for _, category := range categories {
		content.WriteString(titleStyle.Render(fmt.Sprintf("📁 %s", category.name)) + "\n")
		for _, item := range category.items {
			content.WriteString(itemStyle.Render(fmt.Sprintf("• %s", item)) + "\n")
		}
		content.WriteString("\n")
	}

	content.WriteString(subtleStyle.Render("This page demonstrates categorized content display"))
	content.WriteString("\n\n" + subtleStyle.Render("esc: back to home • q: quit"))

	return content.String()
}