package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"distui/handlers"
)

var (
	branchCurrentStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	branchSelectedStyle = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("12"))
	branchNormalStyle   = lipgloss.NewStyle()
)

func RenderBranchSelection(m handlers.BranchSelectionModel) string {
	if m.Loading {
		return fmt.Sprintf("\n%s Loading branches...\n", m.LoadSpinner.View())
	}

	if m.Error != "" {
		return fmt.Sprintf("\nError: %s\n\nPress Esc to cancel\n", m.Error)
	}

	if len(m.Branches) == 0 {
		return "\nNo branches found.\n\nPress Esc to cancel\n"
	}

	var b strings.Builder

	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	actionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))

	// Find current branch
	currentBranch := ""
	for _, branch := range m.Branches {
		if branch.IsCurrent {
			currentBranch = branch.Name
			break
		}
	}

	b.WriteString("\n")
	b.WriteString(headerStyle.Render("PUSH YOUR CODE") + "\n\n")

	if currentBranch != "" {
		b.WriteString(infoStyle.Render(fmt.Sprintf("Current branch: %s", currentBranch)) + "\n")
		b.WriteString(infoStyle.Render("Select where to push your changes:") + "\n\n")
	}

	for i, branch := range m.Branches {
		prefix := "  "
		if i == m.SelectedIndex {
			prefix = "> "
		}

		branchDisplay := branch.Name
		action := ""

		if branch.IsCurrent {
			// Pushing to current branch
			action = actionStyle.Render(" → Push to origin/" + branch.Name)
		} else if branch.Name == "main" || branch.Name == "master" {
			// Pushing to main
			action = actionStyle.Render(" → Merge into " + branch.Name + " (creates PR)")
		} else {
			// Pushing to another branch
			action = " → Push to " + branch.Name
		}

		tracking := ""
		if branch.TrackingBranch != "" && branch.IsCurrent {
			if branch.AheadCount > 0 {
				tracking = fmt.Sprintf(" (%d commits ahead)", branch.AheadCount)
			}
		}

		line := fmt.Sprintf("%s%s%s%s", prefix, branchDisplay, action, tracking)

		if i == m.SelectedIndex {
			b.WriteString(branchSelectedStyle.Render(line) + "\n")
		} else if branch.IsCurrent {
			b.WriteString(branchCurrentStyle.Render(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}

	b.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	b.WriteString(helpStyle.Render("↑/↓: navigate • enter: push • esc: cancel") + "\n")

	return b.String()
}
