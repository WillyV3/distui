package views

import (
	"fmt"
	"strings"

	"distui/handlers"
	"distui/internal/gitcleanup"
	"github.com/charmbracelet/lipgloss"
)

func RenderSmartCommitConfirm(model *handlers.CleanupModel) string {
	var content strings.Builder

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true)

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("214")).
		Padding(1, 2).
		Width(60)

	content.WriteString(warningStyle.Render("SMART COMMIT CONFIRMATION") + "\n\n")

	var boxContent strings.Builder
	boxContent.WriteString(headerStyle.Render("What Smart Commit Does:") + "\n\n")

	boxContent.WriteString(infoStyle.Render("1. Automatically stages ALL changed files\n"))
	boxContent.WriteString(infoStyle.Render("2. Generates a commit message based on:\n"))
	boxContent.WriteString(infoStyle.Render("   • File types modified (.go, .md, etc.)\n"))
	boxContent.WriteString(infoStyle.Render("   • Number of changes\n"))
	boxContent.WriteString(infoStyle.Render("   • Type of changes (add/modify/delete)\n\n"))

	boxContent.WriteString(warningStyle.Render("WARNING: This will commit ALL changes at once!") + "\n\n")

	// Show what will be committed
	if model != nil && len(model.FileChanges) > 0 {
		boxContent.WriteString(headerStyle.Render("Files to be committed:") + "\n")
		maxFiles := 10
		for i, change := range model.FileChanges {
			if i >= maxFiles {
				remaining := len(model.FileChanges) - maxFiles
				boxContent.WriteString(fmt.Sprintf("   ...and %d more files\n", remaining))
				break
			}
			statusPrefix := "["
			switch change.Status {
			case "M", " M", "MM":
				statusPrefix += "M"
			case "A", " A":
				statusPrefix += "+"
			case "D", " D":
				statusPrefix += "-"
			case "??":
				statusPrefix += "?"
			default:
				statusPrefix += " "
			}
			statusPrefix += "]"
			boxContent.WriteString(fmt.Sprintf("   %s %s\n", statusPrefix, change.Path))
		}
		boxContent.WriteString("\n")

		// Try to predict the commit message
		changes, _ := gitcleanup.GetFileChanges()
		var files []string
		for _, c := range changes {
			files = append(files, c.Path)
		}
		suggestedMsg := gitcleanup.SuggestCommitMessage(files)
		boxContent.WriteString(headerStyle.Render("Predicted commit message:") + "\n")
		boxContent.WriteString(fmt.Sprintf("   \"%s\"\n", suggestedMsg))
	}

	content.WriteString(boxStyle.Render(boxContent.String()))
	content.WriteString("\n\n")

	content.WriteString("Do you want to proceed with Smart Commit?\n\n")
	content.WriteString("[Y] Yes, commit all  [N] No, cancel  [Esc] Cancel")

	return content.String()
}