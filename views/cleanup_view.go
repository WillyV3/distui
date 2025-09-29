package views

import (
	"fmt"
	"strings"

	"distui/handlers"
	"github.com/charmbracelet/lipgloss"
)

func RenderCleanupStatus(model *handlers.CleanupModel) string {
	if model == nil {
		return "Loading repository status..."
	}

	var lines []string

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true)

	// Add top padding
	lines = append(lines, "")

	// Header with left padding
	lines = append(lines, "  "+headerStyle.Render("REPOSITORY STATUS"))

	// Show repo status
	if model.RepoInfo == nil {
		lines = append(lines, "  Not a git repository")
		return strings.Join(lines, "\n")
	}

	// Add blank line after header
	lines = append(lines, "")

	// Git status with padding
	lines = append(lines, "  Git Repository: Initialized")

	// GitHub status with padding
	if model.RepoInfo.RemoteExists {
		lines = append(lines, fmt.Sprintf("  GitHub Remote: %s/%s",
			model.RepoInfo.Owner, model.RepoInfo.RepoName))
	} else {
		lines = append(lines, "  GitHub Remote: Not configured")
	}

	// Changes summary with padding
	modified, added, deleted, untracked := model.GetFileSummary()
	total := modified + added + deleted + untracked

	if total > 0 {
		lines = append(lines, fmt.Sprintf("  Local Changes: %d uncommitted files", total))
	} else {
		lines = append(lines, "  Local Changes: Clean working directory")
	}

	// Branch info with padding
	if model.RepoInfo.Branch != "" {
		lines = append(lines, fmt.Sprintf("  Branch: %s", model.RepoInfo.Branch))
	}

	// Divider with padding
	lines = append(lines, "")
	dividerWidth := model.Width - 8  // Account for padding on both sides
	if dividerWidth < 20 {
		dividerWidth = 20
	}
	lines = append(lines, "  "+strings.Repeat("─", dividerWidth))

	// Calculate how many lines we've used so far
	headerLines := len(lines)

	// Reserve lines for actions at the bottom (4 lines: blank + "Actions:" + action line + potential bottom padding)
	actionLines := 4

	// Calculate available lines for files
	// model.Height is already the available content height (after UI chrome)
	// We just need to subtract our own header and action lines
	availableForFiles := model.Height - headerLines - actionLines
	if availableForFiles < 1 {
		availableForFiles = 1
	}

	// Show file changes
	if len(model.FileChanges) > 0 {
		lines = append(lines, "")
		lines = append(lines, "  Files:")

		filesToShow := len(model.FileChanges)
		if filesToShow > availableForFiles-2 { // -2 for blank line and "Files:" label
			filesToShow = availableForFiles - 3 // Reserve space for "...and X more"
			if filesToShow < 1 {
				filesToShow = 1
			}
		}

		for i := 0; i < filesToShow && i < len(model.FileChanges); i++ {
			change := model.FileChanges[i]
			path := change.Path
			// Truncate path if too long (account for padding)
			maxPathLen := model.Width - 24
			if maxPathLen > 0 && len(path) > maxPathLen {
				path = "..." + path[len(path)-maxPathLen+3:]
			}

			// Use simple status indicators without emojis
			statusChar := "M"
			if change.Status == "??" {
				statusChar = "?"
			} else if strings.HasPrefix(change.Status, "A") {
				statusChar = "+"
			} else if strings.HasPrefix(change.Status, "D") {
				statusChar = "-"
			}

			lines = append(lines, fmt.Sprintf("    [%s] %s", statusChar, path))
		}

		remaining := len(model.FileChanges) - filesToShow
		if remaining > 0 {
			lines = append(lines, fmt.Sprintf("    ...and %d more files", remaining))
		}
	} else {
		// Add empty lines to fill space when no files
		for i := 0; i < availableForFiles; i++ {
			lines = append(lines, "")
		}
	}

	// Fill remaining space to push actions to bottom
	currentLines := len(lines) - headerLines
	for currentLines < availableForFiles {
		lines = append(lines, "")
		currentLines++
	}

	// Actions section (always at bottom)
	lines = append(lines, "")
	lines = append(lines, "  Actions:")
	if model.HasChanges() {
		lines = append(lines, "  [C] Commit  [s] Smart commit  [G] GitHub browser  [r] Refresh")
	} else if model.RepoInfo != nil && model.RepoInfo.RemoteExists {
		lines = append(lines, "  [G] GitHub browser  [P] Push  [r] Refresh")
	} else if model.NeedsGitHub() {
		lines = append(lines, "  [G] Set up GitHub  [r] Refresh")
	} else {
		lines = append(lines, "  [G] GitHub browser  [r] Refresh")
	}

	return strings.Join(lines, "\n")
}