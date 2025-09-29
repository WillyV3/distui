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

	// Show unpushed commits warning if applicable
	if model.RepoInfo.UnpushedCommits > 0 {
		warningStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)
		warningText := fmt.Sprintf("⚠ %d unpushed commit", model.RepoInfo.UnpushedCommits)
		if model.RepoInfo.UnpushedCommits > 1 {
			warningText = fmt.Sprintf("⚠ %d unpushed commits", model.RepoInfo.UnpushedCommits)
		}
		lines = append(lines, "  "+warningStyle.Render(warningText+" - [P] to push!"))
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

	// Show file changes or repo browser
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
	} else if model.RepoBrowser != nil {
		// No changes, show repo browser inline
		lines = append(lines, "")
		lines = append(lines, "  Repository Contents:")

		// Create a mini repo browser view
		browserLines := createMiniRepoBrowser(model.RepoBrowser, availableForFiles-2)
		for _, line := range browserLines {
			lines = append(lines, "  "+line)
		}
	} else {
		// Add empty lines to fill space when no browser available
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
		lines = append(lines, "  [C] Commit  [s] Smart commit  [r] Refresh")
	} else if model.RepoInfo != nil && model.RepoInfo.RemoteExists {
		lines = append(lines, "  [P] Push to remote  [r] Refresh")
	} else if model.NeedsGitHub() {
		lines = append(lines, "  [G] Set up GitHub  [r] Refresh")
	} else {
		lines = append(lines, "  [r] Refresh")
	}

	return strings.Join(lines, "\n")
}

func createMiniRepoBrowser(browser *handlers.RepoBrowserModel, availableLines int) []string {
	var lines []string

	if browser == nil || browser.Error != nil {
		lines = append(lines, "(unable to browse repository)")
		return lines
	}

	// Show current directory
	dirStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	lines = append(lines, dirStyle.Render(browser.CurrentDirectory+"/"))

	// Calculate how many entries we can show
	entriesToShow := len(browser.Entries)
	if entriesToShow > availableLines-1 { // -1 for directory line
		entriesToShow = availableLines - 2 // Reserve space for "...more"
		if entriesToShow < 1 {
			entriesToShow = 1
		}
	}

	// Show entries with simple formatting
	for i := 0; i < entriesToShow && i < len(browser.Entries); i++ {
		entry := browser.Entries[i]

		// Simple type indicator
		typeChar := "-"
		if entry.IsDir {
			typeChar = "/"
		} else if strings.HasSuffix(entry.Name, ".go") {
			typeChar = "g"
		} else if strings.HasSuffix(entry.Name, ".md") {
			typeChar = "m"
		}

		// Format name
		name := entry.Name
		if entry.IsDir {
			name = name + "/"
		}
		if len(name) > 40 {
			name = name[:37] + "..."
		}

		// Add selection indicator
		line := ""
		if i == browser.Selected {
			line = "> " + fmt.Sprintf("%s %s", typeChar, name)
		} else {
			line = "  " + fmt.Sprintf("%s %s", typeChar, name)
		}

		// Color based on type (but highlight if selected)
		if i == browser.Selected {
			line = lipgloss.NewStyle().
				Background(lipgloss.Color("237")).
				Foreground(lipgloss.Color("255")).
				Render(line)
		} else if entry.IsDir {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(line)
		} else if strings.HasSuffix(entry.Name, ".go") {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Render(line)
		} else {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(line)
		}

		lines = append(lines, line)
	}

	// Show if there are more files
	if len(browser.Entries) > entriesToShow {
		remaining := len(browser.Entries) - entriesToShow
		lines = append(lines, dirStyle.Render(fmt.Sprintf("  ...%d more items", remaining)))
	}

	// Fill remaining space with empty lines
	for len(lines) < availableLines {
		lines = append(lines, "")
	}

	return lines
}