package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func RenderGenerateConfigConsent(filesToGenerate []string, filesToDelete []string, width, height int) string {
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

	hasChanges := len(filesToGenerate) > 0 || len(filesToDelete) > 0

	content.WriteString(headerStyle.Render("UPDATE RELEASE CONFIGURATION") + "\n\n")

	// Show files to generate
	if len(filesToGenerate) > 0 {
		content.WriteString(warningStyle.Render("Files to generate/update:") + "\n\n")
		for _, file := range filesToGenerate {
			content.WriteString(infoStyle.Render(fmt.Sprintf("  ✓ %s", file)) + "\n")
		}
		content.WriteString("\n")
	}

	// Show files to delete
	if len(filesToDelete) > 0 {
		deleteStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Padding(0, 2)

		content.WriteString(deleteStyle.Render("Files to delete (no longer needed):") + "\n\n")
		for _, file := range filesToDelete {
			content.WriteString(infoStyle.Render(fmt.Sprintf("  ✗ %s", file)) + "\n")
		}
		content.WriteString("\n")
	}

	if !hasChanges {
		content.WriteString(infoStyle.Render("No changes needed - configuration is up to date.") + "\n\n")
		content.WriteString(controlStyle.Render("[ESC] Return"))
	} else {
		content.WriteString(infoStyle.Render("These changes reflect your current distribution settings.") + "\n")
		content.WriteString(infoStyle.Render("You should commit these changes to version control.") + "\n\n")
		content.WriteString(infoStyle.Render("Apply these changes?") + "\n\n")

		successStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true).
			Padding(0, 2)

		content.WriteString(successStyle.Render("After applying: Press ESC then [r] to release!") + "\n\n")

		content.WriteString(controlStyle.Render("[y] Yes, apply changes  [n] No, cancel"))
	}

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