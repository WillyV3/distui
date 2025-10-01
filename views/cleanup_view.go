package views

import (
	"fmt"
	"strings"

	"distui/handlers"
	"distui/internal/models"
	"github.com/charmbracelet/lipgloss"
)

func RenderCleanupStatus(model *handlers.CleanupModel) string {
	return RenderCleanupStatusWithMessage(model, "", nil)
}

func RenderCleanupStatusWithMessage(model *handlers.CleanupModel, statusMessage string, projectConfig *models.ProjectConfig) string {
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

	// Always reserve space for status message to prevent layout shifts
	if statusMessage != "" {
		successStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)
		lines = append(lines, "  "+successStyle.Render(statusMessage))
	} else {
		lines = append(lines, "")
	}

	// Add blank line after header
	lines = append(lines, "")

	// Color styles for status
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	grayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	blueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))

	// Git status with padding
	lines = append(lines, "  "+grayStyle.Render("Git Repository: ")+greenStyle.Render("Initialized (local)"))

	// GitHub status with padding - distinguish between remote config and actual repo
	if model.RepoInfo.RemoteExists {
		// Use cached value
		if model.GitHubRepoExists {
			lines = append(lines, fmt.Sprintf("  %s%s",
				grayStyle.Render("GitHub Repository: "),
				greenStyle.Render(fmt.Sprintf("%s/%s", model.RepoInfo.Owner, model.RepoInfo.RepoName))))
		} else {
			lines = append(lines, fmt.Sprintf("  %s%s",
				grayStyle.Render("GitHub Remote: "),
				yellowStyle.Render(fmt.Sprintf("%s/%s (not found on GitHub)", model.RepoInfo.Owner, model.RepoInfo.RepoName))))
		}
	} else {
		lines = append(lines, "  "+grayStyle.Render("GitHub Remote: ")+yellowStyle.Render("Not configured"))
	}

	// Changes summary with padding
	modified, added, deleted, untracked := model.GetFileSummary()
	total := modified + added + deleted + untracked

	if total > 0 {
		lines = append(lines, fmt.Sprintf("  %s%s",
			grayStyle.Render("Local Changes: "),
			yellowStyle.Render(fmt.Sprintf("%d uncommitted files", total))))
	} else {
		lines = append(lines, "  "+grayStyle.Render("Local Changes: ")+greenStyle.Render("Clean working directory"))
	}

	// Branch info with padding
	if model.RepoInfo.Branch != "" {
		lines = append(lines, fmt.Sprintf("  %s%s",
			grayStyle.Render("Branch: "),
			blueStyle.Render(model.RepoInfo.Branch)))
	}

	// Smart commit mode indicator
	customRulesEnabled := projectConfig != nil &&
		projectConfig.Config != nil &&
		projectConfig.Config.SmartCommit != nil &&
		projectConfig.Config.SmartCommit.UseCustomRules

	if customRulesEnabled {
		modeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
		lines = append(lines, fmt.Sprintf("  %s%s",
			grayStyle.Render("Smart Commit: "),
			modeStyle.Render("Custom Rules")))
	} else {
		modeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
		lines = append(lines, fmt.Sprintf("  %s%s",
			grayStyle.Render("Smart Commit: "),
			modeStyle.Render("Default (Go only)")))
	}

	// Show "All synced!" message if clean and pushed
	if total == 0 && model.RepoInfo.UnpushedCommits == 0 {
		lines = append(lines, "")
		syncedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
		lines = append(lines, "  "+syncedStyle.Render("Repository is clean and up to date."))
		lines = append(lines, "  "+syncedStyle.Render("Press escape to release a new version."))
	}

	// Divider with padding
	lines = append(lines, "")
	dividerWidth := model.Width - 8  // Account for padding on both sides
	if dividerWidth < 20 {
		dividerWidth = 20
	}
	lines = append(lines, "  "+strings.Repeat("─", dividerWidth))

	// Track actual header lines (everything before file/browser section)
	headerLineCount := len(lines)

	// Actions at bottom: 4 lines (blank + "Actions:" + action line + blank after)
	actionLines := 4

	// Calculate available lines for files
	// model.Height is the content height, subtract actual header and action space
	availableForFiles := model.Height - headerLineCount - actionLines
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
	// Calculate how many lines we've used for content (excluding header)
	contentLinesUsed := len(lines) - headerLineCount
	for contentLinesUsed < availableForFiles {
		lines = append(lines, "")
		contentLinesUsed++
	}

	// Actions section (always at bottom)
	lines = append(lines, "")
	lines = append(lines, "  Actions:")

	// Check if GitHub repo needs to be created (remote configured but doesn't exist)
	needsGitHubRepo := model.RepoInfo != nil && model.RepoInfo.RemoteExists && !model.GitHubRepoExists

	if model.HasChanges() {
		if needsGitHubRepo {
			// Show both commit and GitHub repo creation
			actionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
			lines = append(lines, "  [C] Commit  [s] Smart commit  "+actionStyle.Render("[G] Create GitHub Repo")+"  [p] Preferences  [r] Refresh")
		} else {
			lines = append(lines, "  [C] Commit  [s] Smart commit  [p] Preferences  [r] Refresh")
		}
	} else if model.RepoInfo != nil && model.RepoInfo.UnpushedCommits > 0 {
		lines = append(lines, "  [P] Push to remote  [p] Preferences  [r] Refresh")
	} else if needsGitHubRepo {
		// Remote is configured but repo doesn't exist on GitHub
		actionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
		lines = append(lines, "  "+actionStyle.Render("[G] Create Repository on GitHub")+"  [p] Preferences  [r] Refresh")
	} else if model.RepoInfo != nil && model.RepoInfo.RemoteExists {
		lines = append(lines, "  [p] Preferences  [r] Refresh")
	} else if model.NeedsGitHub() {
		lines = append(lines, "  [G] Set up GitHub  [p] Preferences  [r] Refresh")
	} else {
		lines = append(lines, "  [p] Preferences  [r] Refresh")
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

	// Calculate visible range - center the selected item
	visibleCount := availableLines - 1 // -1 for directory line
	if visibleCount < 1 {
		visibleCount = 1
	}

	scrollStart := 0
	scrollEnd := visibleCount

	if len(browser.Entries) <= visibleCount {
		// All items fit
		scrollStart = 0
		scrollEnd = len(browser.Entries)
	} else {
		// Center the selected item
		scrollStart = browser.Selected - visibleCount/2
		if scrollStart < 0 {
			scrollStart = 0
		}

		scrollEnd = scrollStart + visibleCount
		if scrollEnd > len(browser.Entries) {
			scrollEnd = len(browser.Entries)
			scrollStart = scrollEnd - visibleCount
			if scrollStart < 0 {
				scrollStart = 0
			}
		}
	}

	// Show entries with simple formatting
	for i := scrollStart; i < scrollEnd && i < len(browser.Entries); i++ {
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
	if scrollEnd < len(browser.Entries) {
		remaining := len(browser.Entries) - scrollEnd
		lines = append(lines, dirStyle.Render(fmt.Sprintf("  ...%d more items", remaining)))
	}

	// Fill remaining space with empty lines
	for len(lines) < availableLines {
		lines = append(lines, "")
	}

	return lines
}