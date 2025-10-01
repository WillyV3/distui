package views

import (
	"fmt"
	"strings"

	"distui/internal/models"
	"github.com/charmbracelet/lipgloss"
)

func RenderGlobalContent(projects []models.ProjectConfig, selectedIndex int, deleteMode bool) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true).
		Padding(0, 1)

	projectStyle := lipgloss.NewStyle().
		Padding(0, 1).
		MarginLeft(2)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true).
		Padding(0, 1).
		MarginLeft(2)

	actionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("207")).
		Bold(true).
		Padding(0, 1).
		MarginLeft(2)

	subtleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)

	var content strings.Builder

	content.WriteString(headerStyle.Render("ALL PROJECTS") + "\n\n")

	if len(projects) > 0 {
		for i, project := range projects {
			style := projectStyle
			if i == selectedIndex {
				style = selectedStyle
			}

			if project.Project == nil {
				continue
			}

			name := project.Project.Identifier
			version := "v0.0.0"
			status := "Active"

			if project.Project.Module != nil && project.Project.Module.Version != "" {
				version = project.Project.Module.Version
			}

			if project.History != nil && len(project.History.Releases) > 0 {
				lastRelease := project.History.Releases[0]
				if lastRelease.Status == "failed" {
					status = "Failed"
				} else if lastRelease.Status == "success" {
					status = "Active"
				}
			}

			projectLine := fmt.Sprintf("%-30s %-12s %s",
				truncate(name, 30),
				version,
				status)

			marker := "  "
			if i == selectedIndex {
				marker = "→ "
			}
			content.WriteString(style.Render(fmt.Sprintf("%s%s", marker, projectLine)) + "\n")

			if i == selectedIndex && project.Project.Path != "" {
				content.WriteString(subtleStyle.Render(fmt.Sprintf("    %s", project.Project.Path)) + "\n")
			}
		}
	} else {
		content.WriteString(projectStyle.Render("  No projects registered yet") + "\n")
		content.WriteString(projectStyle.Render("  Press [a] to add your first project") + "\n")
	}

	content.WriteString("\n" + headerStyle.Render("ACTIONS") + "\n\n")

	if deleteMode {
		content.WriteString(errorStyle.Render("  DELETE MODE ACTIVE") + "\n")
		content.WriteString(errorStyle.Render("  Press [enter] to delete or [esc] to cancel") + "\n")
	} else {
		actions := []string{
			"[a] Add New Project",
			"[d] Delete Selected",
			"[enter] Open Selected Project",
		}

		for _, action := range actions {
			content.WriteString(actionStyle.Render(fmt.Sprintf("  %s", action)) + "\n")
		}
	}

	content.WriteString("\n" + subtleStyle.Render("Manage all your Go projects from one place"))
	content.WriteString("\n" + subtleStyle.Render("↑/↓: navigate • enter: select • p: back to project • q: quit"))

	return content.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}