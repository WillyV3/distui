package views

import (
	"fmt"
	"strings"

	"distui/handlers"
	"github.com/charmbracelet/lipgloss"
)

func RenderCommitView(model *handlers.CommitModel) string {
	if model == nil {
		return "Loading commit interface..."
	}

	var lines []string
	maxLines := model.Height
	if maxLines < 10 {
		maxLines = 10
	}

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true)

	fileStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true)

	// Header
	lines = append(lines, "")
	lines = append(lines, "  "+headerStyle.Render("FILE-BY-FILE COMMIT"))
	lines = append(lines, "")

	// If we're done with files, show commit message input
	if model.IsComplete() {
		if !model.HasStagedFiles() {
			lines = append(lines, "  No files staged for commit!")
			lines = append(lines, "")
			lines = append(lines, "  [p] Go back  [Esc] Cancel")
			return strings.Join(lines, "\n")
		}

		lines = append(lines, "  Files to commit:")

		// Calculate how many file lines we can show
		availableForFiles := maxLines - 11
		if availableForFiles < 3 {
			availableForFiles = 3
		}

		stagedCount := 0
		displayedFiles := 0
		for _, change := range model.FileChanges {
			if model.Decisions[change.Path] == "stage" {
				stagedCount++
				if displayedFiles < availableForFiles {
					statusChar := "+"
					if strings.HasPrefix(change.Status, "M") {
						statusChar = "M"
					} else if strings.HasPrefix(change.Status, "D") {
						statusChar = "-"
					}
					lines = append(lines, fmt.Sprintf("    [%s] %s", statusChar, change.Path))
					displayedFiles++
				}
			}
		}

		if stagedCount > displayedFiles {
			lines = append(lines, fmt.Sprintf("    ...and %d more", stagedCount-displayedFiles))
		}

		lines = append(lines, "")
		lines = append(lines, "  Commit message:")
		lines = append(lines, "  "+model.CommitMessage.View())
		lines = append(lines, "")
		lines = append(lines, "  [Enter] Commit  [p] Go back  [Esc] Cancel")
		return strings.Join(lines, "\n")
	}

	// Show current file
	currentFile := model.FileChanges[model.CurrentIndex]
	lines = append(lines, fmt.Sprintf("  File %d of %d", model.CurrentIndex+1, len(model.FileChanges)))
	lines = append(lines, "")

	// Show file path with status
	statusText := "Modified"
	statusColor := "214"
	if currentFile.Status == "??" {
		statusText = "Untracked"
		statusColor = "226"
	} else if strings.HasPrefix(currentFile.Status, "A") {
		statusText = "Added"
		statusColor = "82"
	} else if strings.HasPrefix(currentFile.Status, "D") {
		statusText = "Deleted"
		statusColor = "196"
	}

	lines = append(lines, "  "+fileStyle.Render(currentFile.Path))
	lines = append(lines, "  "+lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Render(statusText))

	// Show what will happen
	lines = append(lines, "")
	decision := model.Decisions[currentFile.Path]
	if decision == "stage" {
		lines = append(lines, "  Status: Will be staged for commit")
	} else if decision == "skip" {
		lines = append(lines, "  Status: Will be skipped")
	} else if decision == "ignore" {
		lines = append(lines, "  Status: Will be added to .gitignore")
	}

	// Show staged count so far
	stagedSoFar := 0
	for i := 0; i < model.CurrentIndex; i++ {
		if model.Decisions[model.FileChanges[i].Path] == "stage" {
			stagedSoFar++
		}
	}
	if stagedSoFar > 0 {
		lines = append(lines, fmt.Sprintf("  (%d files staged so far)", stagedSoFar))
	}

	// Actions
	lines = append(lines, "")
	lines = append(lines, "  What to do with this file?")
	lines = append(lines, "")
	lines = append(lines, "  [a] Add to commit (stage)")
	lines = append(lines, "  [s] Skip for now")
	lines = append(lines, "  [i] Add to .gitignore")
	lines = append(lines, "")
	lines = append(lines, "  [p] Previous file  [Esc] Cancel")

	// Ensure content fits within available height
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	return strings.Join(lines, "\n")
}