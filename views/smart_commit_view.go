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

	maxLines := 20
	availableWidth := 80
	if model != nil {
		if model.Height > 0 {
			maxLines = model.Height
			if maxLines < 15 {
				maxLines = 15
			}
		}
		if model.Width > 0 {
			availableWidth = model.Width - 4
			if availableWidth < 60 {
				availableWidth = 60
			}
		}
	}

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true)

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("214")).
		Padding(1, 2).
		Width(availableWidth)

	content.WriteString(warningStyle.Render("SMART COMMIT CONFIRMATION") + "\n\n")

	var boxContent strings.Builder
	boxContent.WriteString(headerStyle.Render("What Smart Commit Does:") + "\n\n")

	boxContent.WriteString("1. Groups files by type (code, config, docs, etc.)\n")
	boxContent.WriteString("2. Auto-commits Go files in Go code directories\n")
	boxContent.WriteString("3. Auto-commits project files (go.mod, .goreleaser.yaml)\n")
	boxContent.WriteString("4. Creates separate commits for each category\n")
	boxContent.WriteString("5. Generates descriptive commit messages\n\n")

	boxContent.WriteString(headerStyle.Render("Go-Aware Categorization:") + "\n")
	boxContent.WriteString("   • .go files in Go directories → Auto-commit\n")
	boxContent.WriteString("   • Non-Go files in Go directories → Ask user\n")
	boxContent.WriteString("   • Files in non-code directories → Ask user\n\n")

	// Show what will be committed
	if model != nil && len(model.FileChanges) > 0 {
		// Categorize files
		autoCommitFiles := []gitcleanup.FileChange{}
		askUserFiles := []gitcleanup.FileChange{}

		for _, change := range model.FileChanges {
			category := gitcleanup.CategorizeFile(change.Path)
			if category == gitcleanup.CategoryAuto {
				autoCommitFiles = append(autoCommitFiles, change)
			} else if category != gitcleanup.CategoryIgnore {
				askUserFiles = append(askUserFiles, change)
			}
		}

		maxFiles := maxLines - 18
		if maxFiles < 3 {
			maxFiles = 3
		}
		if maxFiles > 8 {
			maxFiles = 8
		}

		// Show auto-commit files
		if len(autoCommitFiles) > 0 {
			autoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
			boxContent.WriteString(autoStyle.Render("Auto-commit (Go files):") + "\n")
			shown := 0
			for i, change := range autoCommitFiles {
				if shown >= maxFiles/2 {
					remaining := len(autoCommitFiles) - shown
					boxContent.WriteString(fmt.Sprintf("   ...and %d more\n", remaining))
					break
				}
				boxContent.WriteString(fmt.Sprintf("   ✓ %s\n", change.Path))
				shown++
				if i >= len(autoCommitFiles)-1 {
					break
				}
			}
			boxContent.WriteString("\n")
		}

		// Show ask-user files
		if len(askUserFiles) > 0 {
			askStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
			boxContent.WriteString(askStyle.Render("Will prompt for (non-Go/mixed dirs):") + "\n")
			shown := 0
			for i, change := range askUserFiles {
				if shown >= maxFiles/2 {
					remaining := len(askUserFiles) - shown
					boxContent.WriteString(fmt.Sprintf("   ...and %d more\n", remaining))
					break
				}
				boxContent.WriteString(fmt.Sprintf("   ? %s\n", change.Path))
				shown++
				if i >= len(askUserFiles)-1 {
					break
				}
			}
			boxContent.WriteString("\n")
		}

		// Summary
		if len(autoCommitFiles) > 0 {
			boxContent.WriteString(headerStyle.Render("Summary:") + "\n")
			boxContent.WriteString(fmt.Sprintf("   %d files will be auto-committed\n", len(autoCommitFiles)))
			if len(askUserFiles) > 0 {
				boxContent.WriteString(fmt.Sprintf("   %d files will prompt for confirmation\n", len(askUserFiles)))
			}
		}
	}

	content.WriteString(boxStyle.Render(boxContent.String()))
	content.WriteString("\n\n")

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	content.WriteString("Proceed with Smart Commit?\n")
	content.WriteString(dimStyle.Render("(You'll be prompted for non-Go files during the process)") + "\n\n")
	content.WriteString("[Y] Yes, start smart commit  [N] No, cancel  [Esc] Cancel")

	return content.String()
}