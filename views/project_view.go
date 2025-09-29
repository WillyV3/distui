package views

import (
	"fmt"
	"strings"

	"distui/handlers"
	"distui/internal/models"
	"github.com/charmbracelet/lipgloss"
)

func RenderProjectContent(project *models.ProjectInfo, config *models.ProjectConfig, globalConfig *models.GlobalConfig, releaseModel *handlers.ReleaseModel) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true).
		Padding(0, 1)

	infoStyle := lipgloss.NewStyle().
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

	// GitHub status
	if globalConfig != nil && globalConfig.User.GitHubUsername != "" {
		content.WriteString(successStyle.Render(fmt.Sprintf("✓ GitHub: %s", globalConfig.User.GitHubUsername)) + "\n\n")
	} else {
		content.WriteString(warningStyle.Render("⚠ GitHub not configured") + "\n\n")
	}

	// UNCONFIGURED project - minimal view
	if project != nil && config == nil {
		content.WriteString(headerStyle.Render("UNCONFIGURED PROJECT") + "\n\n")
		content.WriteString(infoStyle.Render(fmt.Sprintf("%s", project.Module.Name)) + "\n")
		content.WriteString(subtleStyle.Render(fmt.Sprintf("%s", project.Path)) + "\n\n")
		content.WriteString(warningStyle.Render("Press [c] to configure this project for releases") + "\n\n")
		content.WriteString(subtleStyle.Render("c: configure • g: global • s: settings • q: quit"))
		return content.String()
	}

	// NO project detected
	if project == nil {
		content.WriteString(headerStyle.Render("NO PROJECT") + "\n\n")
		content.WriteString(infoStyle.Render("No Go project detected in current directory") + "\n\n")
		content.WriteString(subtleStyle.Render("g: global • s: settings • q: quit"))
		return content.String()
	}

	// CONFIGURED project - full view
	content.WriteString(headerStyle.Render(project.Module.Name) + "\n\n")
	content.WriteString(infoStyle.Render(fmt.Sprintf("Version: %s", project.Module.Version)) + "\n")

	if project.Repository != nil {
		content.WriteString(infoStyle.Render(fmt.Sprintf("Repo: %s/%s",
			project.Repository.Owner, project.Repository.Name)) + "\n")
	}

	// Inline release section (appears when [r] pressed)
	if releaseModel != nil {
		content.WriteString("\n")
		content.WriteString(renderInlineReleaseSection(releaseModel))
		content.WriteString("\n")
	}

	// Recent releases (only if history exists)
	if config.History != nil && len(config.History.Releases) > 0 {
		content.WriteString("\n" + headerStyle.Render("RECENT RELEASES") + "\n\n")
		for i, release := range config.History.Releases[:min(3, len(config.History.Releases))] {
			if i > 2 {
				break
			}
			status := "✓"
			if release.Status == "failed" {
				status = "✗"
			}
			content.WriteString(infoStyle.Render(fmt.Sprintf("%s %s (%s)",
				status, release.Version, release.Duration)) + "\n")
		}
		content.WriteString("\n")
	}

	content.WriteString(subtleStyle.Render("r: release • c: configure • g: global • s: settings • q: quit"))

	return content.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func renderInlineReleaseSection(m *handlers.ReleaseModel) string {
	if m == nil {
		return ""
	}

	switch m.Phase {
	case models.PhaseVersionSelect:
		return renderCompactVersionSelect(m)
	case models.PhaseComplete:
		return RenderSuccess(m)
	case models.PhaseFailed:
		return RenderFailure(m)
	default:
		return RenderProgress(m)
	}
}

func renderCompactVersionSelect(m *handlers.ReleaseModel) string {
	var content strings.Builder

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true).
		Padding(0, 1)

	fieldStyle := lipgloss.NewStyle().
		Padding(0, 1).
		MarginLeft(2)

	actionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true).
		Padding(0, 1).
		MarginLeft(2)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true).
		Padding(0, 1).
		MarginLeft(2)

	subtleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	content.WriteString(headerStyle.Render("SELECT RELEASE VERSION") + "\n\n")
	content.WriteString(fieldStyle.Render(fmt.Sprintf("Current: %s", m.CurrentVersion)) + "\n\n")

	versions := []string{
		"Patch (bug fixes)",
		"Minor (new features)",
		"Major (breaking changes)",
		"Custom version",
	}

	for i, ver := range versions {
		prefix := "  "
		style := actionStyle
		if i == m.SelectedVersion {
			prefix = "> "
			style = selectedStyle
		}
		content.WriteString(style.Render(prefix+ver) + "\n")
	}

	if m.SelectedVersion == 3 {
		content.WriteString("\n" + fieldStyle.Render("Enter version: ") + m.VersionInput.View() + "\n")
	}

	content.WriteString("\n" + subtleStyle.Render("↑/↓: navigate • enter: start • esc: cancel"))

	return content.String()
}