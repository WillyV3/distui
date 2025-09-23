package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderPage3Content returns the content for page3
func RenderPage3Content() string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true).
		Align(lipgloss.Center).
		Padding(1, 0)

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("117")).
		Padding(1, 2).
		Margin(1, 0)

	subtleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	cards := []struct {
		title       string
		description string
		details     []string
	}{
		{
			title:       "📊 Data Card",
			description: "Sample data visualization",
			details: []string{
				"• Metric 1: 85%",
				"• Metric 2: 142 items",
				"• Metric 3: Active",
			},
		},
		{
			title:       "⚙️ Config Card",
			description: "Configuration settings",
			details: []string{
				"• Setting A: Enabled",
				"• Setting B: Auto",
				"• Setting C: Custom",
			},
		},
	}

	var content strings.Builder
	content.WriteString(headerStyle.Render("Third Sample Page") + "\n")
	content.WriteString("This page demonstrates card-based layouts:\n\n")

	for _, card := range cards {
		cardContent := fmt.Sprintf("%s\n%s\n\n%s",
			card.title,
			card.description,
			strings.Join(card.details, "\n"))

		content.WriteString(cardStyle.Render(cardContent) + "\n")
	}

	content.WriteString(subtleStyle.Render("This page demonstrates card-based content layout"))
	content.WriteString("\n\n" + subtleStyle.Render("esc: back to home • q: quit"))

	return content.String()
}