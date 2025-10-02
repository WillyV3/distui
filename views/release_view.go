package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"distui/handlers"
	"distui/internal/models"
)

var (
	releaseHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("117")).
			Bold(true).
			Padding(0, 1)

	releaseFieldStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginLeft(2)

	releaseValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("207")).
			Bold(true)

	releaseActionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true).
			Padding(0, 1).
			MarginLeft(2)

	releaseSelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true).
			Padding(0, 1).
			MarginLeft(2)

	releaseSubtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	releaseCheckMark = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	releaseCrossMark = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).SetString("✗")

	releaseCurrentPkgNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
)

func RenderReleaseContent(releaseModel *handlers.ReleaseModel) string {
	if releaseModel == nil {
		return "No release model initialized"
	}

	switch releaseModel.Phase {
	case models.PhaseVersionSelect:
		return RenderVersionSelection(releaseModel)
	case models.PhaseComplete:
		return RenderSuccess(releaseModel)
	case models.PhaseFailed:
		return RenderFailure(releaseModel)
	default:
		return RenderProgress(releaseModel)
	}
}

func RenderVersionSelection(m *handlers.ReleaseModel) string {
	var content strings.Builder

	configureStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117")) // Teal/cyan color for configure
	configureSelectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)

	content.WriteString(releaseHeaderStyle.Render("SELECT RELEASE VERSION") + "\n\n")
	content.WriteString(releaseFieldStyle.Render(fmt.Sprintf("Current Version: %s", m.CurrentVersion)) + "\n\n")

	versions := []string{
		"Configure Project",
		"Patch (bug fixes)",
		"Minor (new features)",
		"Major (breaking changes)",
		"Custom version",
	}

	for i, ver := range versions {
		prefix := "  "
		style := releaseActionStyle

		// Special styling for Configure Project (item 0)
		if i == 0 {
			style = configureStyle
			if i == m.SelectedVersion {
				prefix = "> "
				style = configureSelectedStyle
			}
		} else {
			if i == m.SelectedVersion {
				prefix = "> "
				style = releaseSelectedStyle
			}
		}

		content.WriteString(style.Render(prefix+ver) + "\n")
	}

	if m.SelectedVersion == 4 {
		content.WriteString("\n" + releaseFieldStyle.Render("Enter version: ") + m.VersionInput.View() + "\n")
	}

	// Show changelog input if enabled and a release version is selected (not Configure Project, not Custom)
	needsChangelog := false
	debugInfo := ""
	if m.ProjectConfig == nil {
		debugInfo = "[ProjectConfig is nil]"
	} else if m.ProjectConfig.Config == nil {
		debugInfo = "[ProjectConfig.Config is nil]"
	} else if m.ProjectConfig.Config.Release == nil {
		debugInfo = "[ProjectConfig.Config.Release is nil]"
	} else {
		needsChangelog = m.ProjectConfig.Config.Release.GenerateChangelog
		debugInfo = fmt.Sprintf("[GenerateChangelog=%v]", needsChangelog)
	}

	// Temporary debug output
	if m.SelectedVersion > 0 && m.SelectedVersion < 4 {
		content.WriteString("\n" + releaseSubtleStyle.Render(debugInfo) + "\n")
	}

	if needsChangelog && m.SelectedVersion > 0 && m.SelectedVersion < 4 {
		content.WriteString("\n" + releaseFieldStyle.Render("Changelog: ") + m.ChangelogInput.View() + "\n")
	}

	content.WriteString("\n" + releaseSubtleStyle.Render("↑/↓: navigate • enter: start release • esc: back"))

	return content.String()
}

func RenderProgress(m *handlers.ReleaseModel) string {
	var content strings.Builder

	content.WriteString(releaseHeaderStyle.Render("RELEASING "+m.Version) + "\n\n")

	// Show progress bar
	content.WriteString(m.Progress.View() + "\n\n")

	// Show output stream (last 10 lines)
	if len(m.Output) > 0 {
		outputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

		start := 0
		if len(m.Output) > 10 {
			start = len(m.Output) - 10
		}

		for _, line := range m.Output[start:] {
			// Highlight successful completions
			if strings.Contains(line, "✓") {
				content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(line) + "\n")
			} else if strings.Contains(line, "...") {
				// Current action
				content.WriteString(m.Spinner.View() + " " + outputStyle.Render(line) + "\n")
			} else {
				content.WriteString(outputStyle.Render(line) + "\n")
			}
		}
	}

	// Add some spacing
	for i := len(m.Output); i < 10; i++ {
		content.WriteString("\n")
	}

	elapsed := time.Since(m.StartTime).Round(time.Second)
	content.WriteString("\n" + releaseSubtleStyle.Render(fmt.Sprintf("Elapsed: %s", elapsed)))

	return content.String()
}

func RenderSuccess(m *handlers.ReleaseModel) string {
	var content strings.Builder

	successHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("42")).
		Bold(true).
		Padding(0, 1)

	content.WriteString(releaseCheckMark.String() + " " + successHeaderStyle.Render("RELEASE COMPLETE") + "\n\n")

	content.WriteString(releaseFieldStyle.Render("Version:  ") + releaseValueStyle.Render(m.Version) + "\n")

	// Use the captured CompletedDuration instead of recalculating
	duration := m.CompletedDuration
	if duration == 0 {
		// Fallback if CompletedDuration wasn't set
		duration = time.Since(m.StartTime)
	}
	content.WriteString(releaseFieldStyle.Render("Duration: ") + releaseValueStyle.Render(duration.Round(time.Second).String()) + "\n\n")

	content.WriteString(releaseHeaderStyle.Render("PUBLISHED TO") + "\n")
	successCount := 0
	for _, pkg := range m.Packages {
		if pkg.Status == "done" {
			content.WriteString("  " + releaseCheckMark.String() + " " + pkg.Name)
			if pkg.Duration > 0 {
				content.WriteString(releaseSubtleStyle.Render(fmt.Sprintf(" (%s)", pkg.Duration.Round(time.Second))))
			}
			content.WriteString("\n")
			successCount++
		}
	}

	// Show summary if we have completed steps
	if successCount > 0 {
		content.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(
			fmt.Sprintf("Successfully completed %d/%d steps", successCount, len(m.Packages))))
	}

	// GitHub release reminder
	if m.RepoOwner != "" && m.RepoName != "" && m.Version != "" {
		githubURL := fmt.Sprintf("https://github.com/%s/%s/releases/tag/%s", m.RepoOwner, m.RepoName, m.Version)
		reminderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
		content.WriteString("\n\n" + reminderStyle.Render("→ Head over to " + githubURL))
		content.WriteString("\n" + reminderStyle.Render("  to edit the release and tell your users what changed!"))
	}

	content.WriteString("\n\n" + releaseSubtleStyle.Render("Press ESC to return"))

	return content.String()
}

func RenderFailure(m *handlers.ReleaseModel) string {
	var content strings.Builder

	content.WriteString(releaseCrossMark.String() + " " + releaseHeaderStyle.Render("RELEASE FAILED") + "\n\n")

	if m.Error != nil {
		content.WriteString(releaseFieldStyle.Render("Error: ") + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.Error.Error()) + "\n\n")
	}

	// Show elapsed time if available
	if m.CompletedDuration > 0 {
		duration := m.CompletedDuration.Round(time.Second)
		content.WriteString(releaseFieldStyle.Render("Elapsed: ") + releaseSubtleStyle.Render(duration.String()) + "\n\n")
	}

	content.WriteString(releaseHeaderStyle.Render("COMPLETED STEPS") + "\n")
	for _, pkg := range m.Packages {
		status := ""
		switch pkg.Status {
		case "done":
			status = releaseCheckMark.String()
		case "failed":
			status = releaseCrossMark.String()
		default:
			status = releaseSubtleStyle.Render("○")
		}
		content.WriteString(status + " " + pkg.Name + "\n")
	}

	content.WriteString("\n" + releaseSubtleStyle.Render("Press ESC to return • R to retry"))

	return content.String()
}