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

	var content strings.Builder

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true)

	statusStyle := lipgloss.NewStyle().
		Padding(1, 0)

	content.WriteString(headerStyle.Render("REPOSITORY STATUS") + "\n")

	// Show repo status
	if model.RepoInfo == nil {
		content.WriteString("❌ Not a git repository\n")
		return content.String()
	}

	// Git status
	content.WriteString("Git Repository: ✅ Initialized\n")

	// GitHub status
	if model.RepoInfo.RemoteExists {
		content.WriteString(fmt.Sprintf("GitHub Remote: ✅ %s/%s\n",
			model.RepoInfo.Owner, model.RepoInfo.RepoName))
	} else {
		content.WriteString("GitHub Remote: ⚠️  Not configured\n")
	}

	// Changes summary
	modified, added, deleted, untracked := model.GetFileSummary()
	total := modified + added + deleted + untracked

	if total > 0 {
		content.WriteString(fmt.Sprintf("Local Changes: 🔴 %d uncommitted files\n", total))
	} else {
		content.WriteString("Local Changes: ✅ Clean working directory\n")
	}

	// Branch info
	if model.RepoInfo.Branch != "" {
		content.WriteString(fmt.Sprintf("Branch: %s\n", model.RepoInfo.Branch))
	}

	// Use actual width for divider
	dividerWidth := model.Width - 4
	if dividerWidth < 20 {
		dividerWidth = 20
	}
	content.WriteString(strings.Repeat("─", dividerWidth) + "\n")

	// Show file changes
	if len(model.FileChanges) > 0 {
		// Calculate available lines for files
		// This view is rendered inside a content box with:
		// - Tabs and header: 6 lines
		// - Content box border: 2 lines
		// - Content padding: 2 lines
		// - Controls at bottom: 3 lines
		// Total chrome outside this view: ~13 lines
		// Within this view we have:
		// - Status lines: 5 (header + git + github + changes + branch)
		// - Divider: 2
		// - Files label: 1
		// - Actions section: 3-4
		// - Padding: 2
		// Total: ~13 internal + 13 external = 26 lines of overhead
		// With our reduced newlines, we can use fewer lines
		availableLines := model.Height - 22
		if availableLines < 2 {
			availableLines = 2
		}

		// Determine how many files to show
		filesToShow := len(model.FileChanges)
		if filesToShow > availableLines {
			filesToShow = availableLines - 1 // Save one line for "and X more..."
		}

		content.WriteString("Files:\n")
		for i := 0; i < filesToShow; i++ {
			change := model.FileChanges[i]
			path := change.Path
			// Truncate path if too long
			maxPathLen := model.Width - 20 // Account for icon, status, and padding
			if maxPathLen > 0 && len(path) > maxPathLen {
				path = "..." + path[len(path)-maxPathLen+3:]
			}
			content.WriteString(fmt.Sprintf("%s %s (%s)\n",
				change.Icon, path, change.StatusText))
		}

		// Show count of remaining files if any
		remaining := len(model.FileChanges) - filesToShow
		if remaining > 0 {
			content.WriteString(fmt.Sprintf("  ...and %d more files\n", remaining))
		}

	}

	// Action hints
	content.WriteString(statusStyle.Render("Actions:\n"))
	if model.NeedsGitHub() {
		content.WriteString("[G] Set up GitHub repository\n")
	}
	if model.HasChanges() {
		content.WriteString("[C] Commit changes  [s] Smart commit\n")
	} else {
		content.WriteString("[G] GitHub settings\n")
	}
	if model.IsClean() {
		content.WriteString("✅ Ready for release!\n")
	}
	content.WriteString("[r] Refresh  [Tab] Next tab  [Esc] Back")

	return content.String()
}