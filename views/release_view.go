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

	content.WriteString(releaseHeaderStyle.Render("SELECT RELEASE VERSION") + "\n\n")
	content.WriteString(releaseFieldStyle.Render(fmt.Sprintf("Current Version: %s", m.CurrentVersion)) + "\n\n")

	versions := []string{
		"Patch (bug fixes)",
		"Minor (new features)",
		"Major (breaking changes)",
		"Custom version",
	}

	for i, ver := range versions {
		prefix := "  "
		style := releaseActionStyle
		if i == m.SelectedVersion {
			prefix = "> "
			style = releaseSelectedStyle
		}
		content.WriteString(style.Render(prefix+ver) + "\n")
	}

	if m.SelectedVersion == 3 {
		content.WriteString("\n" + releaseFieldStyle.Render("Enter version: ") + m.VersionInput.View() + "\n")
	}

	content.WriteString("\n" + releaseSubtleStyle.Render("↑/↓: navigate • enter: start release • esc: back"))

	return content.String()
}

func RenderProgress(m *handlers.ReleaseModel) string {
	var content strings.Builder

	content.WriteString(releaseHeaderStyle.Render("RELEASING "+m.Version) + "\n\n")

	// Calculate progress
	n := len(m.Packages)
	installed := len(m.Installed)

	// Format counter with consistent width
	w := lipgloss.Width(fmt.Sprintf("%d", n))
	pkgCount := fmt.Sprintf(" %*d/%*d", w, installed, w, n)

	// Current step being processed
	currentStepName := ""
	if m.Installing >= 0 && m.Installing < len(m.Packages) {
		currentStepName = m.Packages[m.Installing].Name
		// Show spinner and current step on one line
		spin := m.Spinner.View() + " "
		info := releaseCurrentPkgNameStyle.Render(currentStepName)
		prog := m.Progress.View()

		// Calculate available space for the step name
		cellsAvail := 60 // Default width
		pkgNameStyled := releaseCurrentPkgNameStyle.Render(currentStepName)
		info = lipgloss.NewStyle().MaxWidth(cellsAvail).Render(pkgNameStyled)

		// Single line progress display
		content.WriteString(spin + info + "\n")
		content.WriteString(prog + pkgCount + "\n\n")
	} else {
		// Just show the progress bar and count
		content.WriteString(m.Progress.View() + pkgCount + "\n\n")
	}

	// Show all steps with their status
	for _, pkg := range m.Packages {
		status := ""
		style := releaseSubtleStyle

		switch pkg.Status {
		case "done":
			status = releaseCheckMark.String()
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		case "failed":
			status = releaseCrossMark.String()
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		case "installing":
			status = m.Spinner.View()
			style = releaseCurrentPkgNameStyle
		default:
			status = "○"
			style = releaseSubtleStyle
		}

		line := fmt.Sprintf("%s %s", status, style.Render(pkg.Name))

		// Add duration for completed steps
		if pkg.Status == "done" && pkg.Duration > 0 {
			line += releaseSubtleStyle.Render(fmt.Sprintf(" (%s)", pkg.Duration.Round(time.Millisecond)))
		}

		content.WriteString(line + "\n")
	}

	// Show condensed output (last few lines only)
	if len(m.Output) > 0 && m.Installing >= 0 {
		content.WriteString("\n")
		// Show only last 3 lines of output
		start := 0
		if len(m.Output) > 3 {
			start = len(m.Output) - 3
		}
		for _, line := range m.Output[start:] {
			// Truncate long lines
			if len(line) > 70 {
				line = line[:67] + "..."
			}
			content.WriteString(releaseSubtleStyle.Render("  " + line) + "\n")
		}
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

	elapsed := time.Since(m.StartTime).Round(time.Second)
	content.WriteString(releaseFieldStyle.Render("Duration: ") + releaseValueStyle.Render(elapsed.String()) + "\n\n")

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

	// Show summary
	content.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(
		fmt.Sprintf("Successfully completed %d/%d steps", successCount, len(m.Packages))))

	content.WriteString("\n\n" + releaseSubtleStyle.Render("Press ESC to return"))

	return content.String()
}

func RenderFailure(m *handlers.ReleaseModel) string {
	var content strings.Builder

	content.WriteString(releaseCrossMark.String() + " " + releaseHeaderStyle.Render("RELEASE FAILED") + "\n\n")

	if m.Error != nil {
		content.WriteString(releaseFieldStyle.Render("Error: ") + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.Error.Error()) + "\n\n")
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