package views

import (
	"fmt"
	"strings"

	"distui/internal/models"
	"github.com/charmbracelet/lipgloss"
)

func RenderGlobalContent(projects []models.ProjectConfig, selectedIndex int, detecting bool, status string, spinnerView string, settingWorkingDir bool, workingDirInput string, workingDirResults []string, workingDirSelected int) string {
	if settingWorkingDir {
		return renderWorkingDirPicker(workingDirInput, workingDirResults, workingDirSelected)
	}
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
		content.WriteString(projectStyle.Render("  No projects detected yet") + "\n")
		content.WriteString(projectStyle.Render("  Press [D] to detect & import from NPM and Homebrew") + "\n")
	}

	if detecting {
		content.WriteString("\n" + spinnerView + " Searching Homebrew & NPM for your distributions...\n")
	} else if status != "" {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Padding(0, 1)
		content.WriteString("\n" + statusStyle.Render(status) + "\n")
	}

	content.WriteString("\n" + headerStyle.Render("ACTIONS") + "\n\n")

	actions := []string{
		"[D] Detect & Import All Distributions",
		"[enter] Open Selected Project",
	}

	for _, action := range actions {
		content.WriteString(actionStyle.Render(fmt.Sprintf("  %s", action)) + "\n")
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

func renderWorkingDirPicker(inputView string, results []string, selected int) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	subtleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	var content strings.Builder

	content.WriteString(headerStyle.Render("SET WORKING DIRECTORY"))
	content.WriteString("\n\n")
	content.WriteString(subtleStyle.Render("This project needs a working directory to track releases."))
	content.WriteString("\n")
	content.WriteString(subtleStyle.Render("Type to search (min 2 chars):"))
	content.WriteString("\n\n")

	content.WriteString(inputView)
	content.WriteString("\n\n")

	if len(results) > 0 {
		content.WriteString(headerStyle.Render("RESULTS"))
		content.WriteString("\n\n")

		for i, result := range results {
			if i >= 3 {
				break
			}

			marker := "  "
			style := normalStyle
			if i == selected {
				marker = "→ "
				style = selectedStyle
			}

			content.WriteString(style.Render(marker + result))
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")
	content.WriteString(subtleStyle.Render("↑/↓: navigate • enter: select • esc: cancel"))

	return content.String()
}