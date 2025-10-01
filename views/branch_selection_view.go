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

	b.WriteString("\n┌─ SELECT BRANCH TO PUSH ──────────────────────\n")
	b.WriteString("│\n")

	for i, branch := range m.Branches {
		prefix := "  "
		if i == m.SelectedIndex {
			prefix = "> "
		}

		branchDisplay := branch.Name
		if branch.IsCurrent {
			branchDisplay += " (current)"
		}

		tracking := ""
		if branch.TrackingBranch != "" {
			tracking = fmt.Sprintf(" → %s", branch.TrackingBranch)
			if branch.AheadCount > 0 {
				tracking += fmt.Sprintf(" (ahead %d)", branch.AheadCount)
			}
			if branch.BehindCount > 0 {
				tracking += fmt.Sprintf(" (behind %d)", branch.BehindCount)
			}
		} else {
			tracking = " (no tracking)"
		}

		line := fmt.Sprintf("│  %s%s%s", prefix, branchDisplay, tracking)

		if i == m.SelectedIndex {
			b.WriteString(branchSelectedStyle.Render(line) + "\n")
		} else if branch.IsCurrent {
			b.WriteString(branchCurrentStyle.Render(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}

	b.WriteString("│\n")
	b.WriteString("│  ↑/↓: navigate • enter: push • esc: cancel\n")
	b.WriteString("└──────────────────────────────────────────────\n")

	return b.String()
}
