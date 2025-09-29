package views

import (
	"fmt"
	"strings"

	"distui/handlers"
	"github.com/charmbracelet/lipgloss"
)

func RenderGitHubManagement(model *handlers.GitHubModel) string {
	if model == nil {
		return "Loading GitHub management..."
	}

	var content strings.Builder

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("69")).
		Padding(1).
		Width(model.Width - 4)

	content.WriteString(headerStyle.Render("GITHUB REPOSITORY MANAGEMENT") + "\n\n")

	// Show current status
	if model.RepoInfo != nil && model.RepoInfo.RemoteExists {
		content.WriteString(fmt.Sprintf("✅ Connected: github.com/%s/%s\n\n",
			model.RepoInfo.Owner, model.RepoInfo.RepoName))
		content.WriteString("[Esc] Return to Configure\n")
		return content.String()
	}

	content.WriteString("⚠️  No GitHub repository configured\n\n")

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