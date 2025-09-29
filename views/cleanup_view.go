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

	content.WriteString(headerStyle.Render("REPOSITORY STATUS") + "\n\n")

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

	content.WriteString("\n" + strings.Repeat("─", 40) + "\n\n")

	// Show file changes
	if len(model.FileChanges) > 0 {
		content.WriteString("Files:\n")
		for _, change := range model.FileChanges {
			content.WriteString(fmt.Sprintf("%s %s (%s)\n",
				change.Icon, change.Path, change.StatusText))
		}
		content.WriteString("\n")
	}

	// Action hints
	content.WriteString(statusStyle.Render("\nActions:\n"))
	if model.NeedsGitHub() {
		content.WriteString("Press [G] to set up GitHub repository\n")
	}
	if model.HasChanges() {
		content.WriteString("Press [C] to commit changes\n")
	}
	if model.IsClean() {
		content.WriteString("✅ Ready for release!\n")
	}

	return content.String()
}