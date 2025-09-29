package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderReleaseContent returns the content for the release creation view
func RenderReleaseContent() string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true).
		Padding(0, 1)

	fieldStyle := lipgloss.NewStyle().
		Padding(0, 1).
		MarginLeft(2)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("207")).
		Bold(true)

	actionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true).
		Padding(0, 1).
		MarginLeft(2)

	subtleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	var content strings.Builder

	content.WriteString(headerStyle.Render("CREATE NEW RELEASE") + "\n\n")

	content.WriteString(fieldStyle.Render("Project:         ") + valueStyle.Render("example-go-app") + "\n")
	content.WriteString(fieldStyle.Render("Current Version: ") + valueStyle.Render("v1.2.3") + "\n")
	content.WriteString(fieldStyle.Render("New Version:     ") + valueStyle.Render("v1.3.0") + " " + subtleStyle.Render("(patch/minor/major)") + "\n")

	content.WriteString("\n" + headerStyle.Render("RELEASE CONFIGURATION") + "\n\n")
	content.WriteString(fieldStyle.Render("Release Type:    ") + valueStyle.Render("Standard") + " " + subtleStyle.Render("(standard/hotfix/rc)") + "\n")
	content.WriteString(fieldStyle.Render("Build Targets:   ") + valueStyle.Render("linux/amd64, darwin/amd64, windows/amd64") + "\n")
	content.WriteString(fieldStyle.Render("Include Tests:   ") + valueStyle.Render("Yes") + "\n")
	content.WriteString(fieldStyle.Render("Create GitHub Release: ") + valueStyle.Render("Yes") + "\n")
	content.WriteString(fieldStyle.Render("Publish to Homebrew:  ") + valueStyle.Render("No") + "\n")

	content.WriteString("\n" + headerStyle.Render("RELEASE NOTES") + "\n\n")
	content.WriteString(fieldStyle.Render("Auto-generated from commits:") + "\n")
	content.WriteString(fieldStyle.Render("• Added new feature X") + "\n")
	content.WriteString(fieldStyle.Render("• Fixed bug in Y module") + "\n")
	content.WriteString(fieldStyle.Render("• Updated dependencies") + "\n")

	content.WriteString("\n" + headerStyle.Render("ACTIONS") + "\n\n")
	actions := []string{
		"[v] Change Version",
		"[t] Configure Targets",
		"[n] Edit Release Notes",
		"[p] Preview Build",
		"[enter] Start Release",
	}

	for _, action := range actions {
		content.WriteString(actionStyle.Render(fmt.Sprintf("  %s", action)) + "\n")
	}

	content.WriteString("\n" + subtleStyle.Render("Build and release your Go application"))
	content.WriteString("\n" + subtleStyle.Render("↑/↓: navigate • enter: execute • esc: back • q: quit"))

	return content.String()
}