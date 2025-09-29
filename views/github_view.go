package views

import (
	"strings"

	"distui/handlers"
	"github.com/charmbracelet/lipgloss"
)

func RenderGitHubManagement(model *handlers.GitHubModel) string {
	if model == nil {
		return "Loading GitHub management..."
	}

	// Show repo browser if in overview state
	if model.State == 0 && model.RepoBrowser != nil { // githubOverview = 0
		return RenderRepoBrowser(model.RepoBrowser)
	}

	// Otherwise show the creation form
	var content strings.Builder

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true)

	// Calculate box width based on available space
	boxWidth := model.Width - 4
	if boxWidth < 40 {
		boxWidth = 40
	}
	if boxWidth > 80 {
		boxWidth = 80 // Max width for readability
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("69")).
		Padding(1).
		Width(boxWidth)

	content.WriteString(headerStyle.Render("GITHUB REPOSITORY MANAGEMENT") + "\n\n")

	content.WriteString("No GitHub repository configured\n\n")

	// Create repository form
	var form strings.Builder
	form.WriteString("Create GitHub Repository\n\n")

	form.WriteString("Name:\n")
	nameView := model.RepoName.View()
	if model.FocusIndex == 0 {
		nameView = "> " + nameView
	} else {
		nameView = "  " + nameView
	}
	form.WriteString(nameView + "\n\n")

	form.WriteString("Description:\n")
	descView := model.RepoDesc.View()
	if model.FocusIndex == 1 {
		descView = "> " + descView
	} else {
		descView = "  " + descView
	}
	form.WriteString(descView + "\n\n")

	form.WriteString("Visibility:\n")
	var visText string
	if model.IsPrivate {
		visText = "[●] Private  [ ] Public"
	} else {
		visText = "[ ] Private  [●] Public"
	}
	if model.FocusIndex == 2 {
		visText = "> " + visText
	} else {
		visText = "  " + visText
	}
	form.WriteString(visText + "\n")

	content.WriteString(boxStyle.Render(form.String()))
	content.WriteString("\n\n")

	content.WriteString("[Tab] Switch fields  [Enter] Create  [Esc] Cancel\n")

	return content.String()
}