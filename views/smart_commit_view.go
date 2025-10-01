package views

import (
	"fmt"
	"strings"

	"distui/handlers"
	"distui/internal/gitcleanup"
	"distui/internal/models"
	"github.com/charmbracelet/lipgloss"
)

func RenderSmartCommitConfirm(model *handlers.CleanupModel, projectConfig *models.ProjectConfig) string {
	customRulesEnabled := projectConfig != nil &&
		projectConfig.Config != nil &&
		projectConfig.Config.SmartCommit != nil &&
		projectConfig.Config.SmartCommit.UseCustomRules
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
	if customRulesEnabled {
		boxContent.WriteString(headerStyle.Render("Custom Smart Commit Mode:") + "\n\n")
		boxContent.WriteString("1. Uses your custom categorization rules\n")
		boxContent.WriteString("2. Commits ALL files matching your rules\n")
		boxContent.WriteString("3. Skips only files marked as 'ignore'\n")
		boxContent.WriteString("4. Generates commit message based on file types\n\n")
	} else {
		boxContent.WriteString(headerStyle.Render("What Smart Commit Does:") + "\n\n")
		boxContent.WriteString("1. Groups files by type (code, config, docs, etc.)\n")
		boxContent.WriteString("2. Auto-commits Go files in Go code directories\n")
		boxContent.WriteString("3. Auto-commits project files (go.mod, .goreleaser.yaml)\n")
		boxContent.WriteString("4. Skips documentation and non-code files\n")
		boxContent.WriteString("5. Generates descriptive commit messages\n\n")

		boxContent.WriteString(headerStyle.Render("Go-Aware Categorization:") + "\n")
		boxContent.WriteString("   • .go files in Go directories → Auto-commit\n")
		boxContent.WriteString("   • Non-Go files in Go directories → Skip\n")
		boxContent.WriteString("   • Files in non-code directories → Skip\n\n")
	}

	// Show what will be committed
	if model != nil && len(model.FileChanges) > 0 {
		// Categorize files
		commitFiles := []gitcleanup.FileChange{}
		skipFiles := []gitcleanup.FileChange{}

		for _, change := range model.FileChanges {
			category := gitcleanup.CategorizeFileWithConfig(change.Path, projectConfig)

			if customRulesEnabled {
				// Custom rules: commit everything except ignore
				if category != gitcleanup.CategoryIgnore {
					commitFiles = append(commitFiles, change)
				} else {
					skipFiles = append(skipFiles, change)
				}
			} else {
				// Default mode: only commit auto files
				if category == gitcleanup.CategoryAuto {
					commitFiles = append(commitFiles, change)
				} else if category != gitcleanup.CategoryIgnore {
					skipFiles = append(skipFiles, change)
				}
			}
		}

		maxFiles := maxLines - 18
		if maxFiles < 3 {
			maxFiles = 3
		}
		if maxFiles > 8 {
			maxFiles = 8
		}

		// Show commit files
		if len(commitFiles) > 0 {
			commitStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
			if customRulesEnabled {
				boxContent.WriteString(commitStyle.Render("Will commit (using custom rules):") + "\n")
			} else {
				boxContent.WriteString(commitStyle.Render("Will commit (Go files):") + "\n")
			}
			shown := 0
			for i, change := range commitFiles {
				if shown >= maxFiles/2 {
					remaining := len(commitFiles) - shown
					boxContent.WriteString(fmt.Sprintf("   ...and %d more\n", remaining))
					break
				}
				boxContent.WriteString(fmt.Sprintf("   ✓ %s\n", change.Path))
				shown++
				if i >= len(commitFiles)-1 {
					break
				}
			}
			boxContent.WriteString("\n")
		}

		// Show skip files
		if len(skipFiles) > 0 {
			skipStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
			boxContent.WriteString(skipStyle.Render("Will skip:") + "\n")
			shown := 0
			for i, change := range skipFiles {
				if shown >= maxFiles/2 {
					remaining := len(skipFiles) - shown
					boxContent.WriteString(fmt.Sprintf("   ...and %d more\n", remaining))
					break
				}
				boxContent.WriteString(fmt.Sprintf("   - %s\n", change.Path))
				shown++
				if i >= len(skipFiles)-1 {
					break
				}
			}
			boxContent.WriteString("\n")
		}

		// Summary
		if len(commitFiles) > 0 {
			boxContent.WriteString(headerStyle.Render("Summary:") + "\n")
			if customRulesEnabled {
				boxContent.WriteString(fmt.Sprintf("   %d files will be committed with custom rules\n", len(commitFiles)))
			} else {
				boxContent.WriteString(fmt.Sprintf("   %d files will be auto-committed\n", len(commitFiles)))
			}
			if len(skipFiles) > 0 {
				boxContent.WriteString(fmt.Sprintf("   %d files will be skipped\n", len(skipFiles)))
			}
		}
	}

	content.WriteString(boxStyle.Render(boxContent.String()))
	content.WriteString("\n\n")

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	content.WriteString("Proceed with Smart Commit?\n")
	if customRulesEnabled {
		content.WriteString(dimStyle.Render("(All categorized files will be committed)") + "\n\n")
	} else {
		content.WriteString(dimStyle.Render("(Only Go files will be committed, docs/other files will be skipped)") + "\n\n")
	}
	content.WriteString("[Y] Yes, start smart commit  [N] No, cancel  [Esc] Cancel")

	return content.String()
}