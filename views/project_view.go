package views

import (
	"fmt"
	"strings"

	"distui/internal/models"
	"github.com/charmbracelet/lipgloss"
)

func RenderProjectContent(project *models.ProjectInfo, config *models.ProjectConfig, globalConfig *models.GlobalConfig) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true).
		Padding(0, 1)

	infoStyle := lipgloss.NewStyle().
		Padding(0, 1).
		MarginLeft(2)

	actionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("207")).
		Bold(true).
		Padding(0, 1).
		MarginLeft(2)

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true).
		Padding(0, 1).
		MarginLeft(2)

	subtleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))

	var content strings.Builder

	// GitHub status indicator
	if globalConfig != nil && globalConfig.User.GitHubUsername != "" {
		content.WriteString(successStyle.Render(fmt.Sprintf("✓ GitHub: %s", globalConfig.User.GitHubUsername)) + "\n\n")
	} else {
		content.WriteString(warningStyle.Render("⚠ GitHub not configured - press [s] then [e] to set up") + "\n\n")
	}

	if project != nil && config == nil {
		content.WriteString(headerStyle.Render("UNCONFIGURED PROJECT DETECTED") + "\n\n")
		content.WriteString(warningStyle.Render("Unreleased Go project detected!") + "\n")
		content.WriteString(infoStyle.Render(fmt.Sprintf("Module: %s", project.Module.Name)) + "\n")
		content.WriteString(infoStyle.Render(fmt.Sprintf("Path: %s", project.Path)) + "\n\n")
		content.WriteString(warningStyle.Render("This project is not configured for releases.") + "\n")
		content.WriteString(warningStyle.Render("Press [c] to configure this project") + "\n")
	} else {
		content.WriteString(headerStyle.Render("PROJECT OVERVIEW") + "\n\n")
	}

	if project != nil && config != nil {
		content.WriteString(infoStyle.Render(fmt.Sprintf("Name: %s", project.Module.Name)) + "\n")
		content.WriteString(infoStyle.Render(fmt.Sprintf("Path: %s", project.Path)) + "\n")
		content.WriteString(infoStyle.Render(fmt.Sprintf("Version: %s", project.Module.Version)) + "\n")

		if project.Repository != nil {
			content.WriteString(infoStyle.Render(fmt.Sprintf("Repository: %s/%s",
				project.Repository.Owner, project.Repository.Name)) + "\n")
			content.WriteString(infoStyle.Render(fmt.Sprintf("Branch: %s",
				project.Repository.DefaultBranch)) + "\n")
		}

		if project.Binary != nil {
			content.WriteString(infoStyle.Render(fmt.Sprintf("Binary: %s", project.Binary.Name)) + "\n")
		}
	} else if project == nil {
		content.WriteString(infoStyle.Render("No project detected") + "\n")
		content.WriteString(infoStyle.Render("Navigate to a Go project directory") + "\n")
	}

	content.WriteString("\n" + headerStyle.Render("RELEASE HISTORY") + "\n\n")

	if config != nil && config.History != nil && len(config.History.Releases) > 0 {
		for i, release := range config.History.Releases[:min(3, len(config.History.Releases))] {
			if i > 2 {
				break
			}
			status := "✓"
			if release.Status == "failed" {
				status = "✗"
			}
			content.WriteString(infoStyle.Render(fmt.Sprintf("%s %s - %s (%s)",
				status, release.Version, release.Status, release.Duration)) + "\n")
		}
	} else {
		content.WriteString(infoStyle.Render("No releases yet") + "\n")
		content.WriteString(infoStyle.Render("Press [r] to create your first release") + "\n")
	}

	content.WriteString("\n" + headerStyle.Render("QUICK ACTIONS") + "\n\n")

	if project != nil && config == nil {
		content.WriteString(warningStyle.Render(fmt.Sprintf("  %s", "[c] Configure Project (REQUIRED)")) + "\n")
		content.WriteString(actionStyle.Render(fmt.Sprintf("  %s", "[t] Run Tests")) + "\n")
		content.WriteString(actionStyle.Render(fmt.Sprintf("  %s", "[b] Build Project")) + "\n")
	} else {
		actions := []string{
			"[r] Create New Release",
			"[c] Configure Project",
			"[h] View History",
			"[t] Run Tests",
			"[b] Build Project",
		}

		for _, action := range actions {
			content.WriteString(actionStyle.Render(fmt.Sprintf("  %s", action)) + "\n")
		}
	}

	content.WriteString("\n" + subtleStyle.Render("Project management and release tools"))
	content.WriteString("\n" + subtleStyle.Render("g: all projects • s: settings • r: release • c: configure • q: quit"))

	return content.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}