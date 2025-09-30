package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func RenderGenerateConfigConsent(files []string, width, height int) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true).
		Padding(1, 2)

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true).
		Padding(0, 2)

	infoStyle := lipgloss.NewStyle().
		Padding(0, 2)

	controlStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Padding(1, 2)

	var content strings.Builder

	// Check if we're regenerating (all common release files) or creating new
	isRegeneration := len(files) > 0 && files[0] == ".goreleaser.yaml"
	allFiles := true
	for _, f := range files {
		if f != ".goreleaser.yaml" && f != "package.json" {
			allFiles = false
			break
		}
	}
	isRegeneration = isRegeneration && allFiles && len(files) <= 2

	if isRegeneration {
		content.WriteString(headerStyle.Render("REGENERATE RELEASE FILES") + "\n\n")
		content.WriteString(warningStyle.Render("distui will regenerate these files with your current config:") + "\n\n")
	} else {
		content.WriteString(headerStyle.Render("GENERATE CONFIGURATION FILES") + "\n\n")
		content.WriteString(warningStyle.Render("distui needs to create these files in your repository:") + "\n\n")
	}

	for _, file := range files {
		content.WriteString(infoStyle.Render(fmt.Sprintf("  • %s", file)) + "\n")
	}

	content.WriteString("\n")
	content.WriteString(infoStyle.Render("These files are required for releases but will be added to your repo.") + "\n")
	content.WriteString(infoStyle.Render("You should commit them to version control.") + "\n\n")

	content.WriteString(infoStyle.Render("Generate these files?") + "\n\n")

	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("82")).
		Bold(true).
		Padding(0, 2)

	content.WriteString(successStyle.Render("After generating: Press ESC then [r] to release!") + "\n\n")

	content.WriteString(controlStyle.Render("[y] Yes, generate files  [n] No, cancel"))

	// Center the content
	contentStr := content.String()
	if width > 0 && height > 0 {
		return lipgloss.Place(
			width,
			height,
			lipgloss.Center,
			lipgloss.Center,
			contentStr,
		)
	}

	return contentStr
}