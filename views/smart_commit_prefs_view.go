package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"distui/handlers"
)

var (
	titleStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
	selectedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	normalStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	enabledStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	disabledStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	confirmStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
)

func RenderSmartCommitPrefs(model *handlers.SmartCommitPrefsModel) string {
	if model == nil {
		return "Loading preferences..."
	}

	var content strings.Builder

	content.WriteString(titleStyle.Render("SMART COMMIT PREFERENCES"))
	content.WriteString("\n\n")

	if model.ShowConfirm {
		content.WriteString(confirmStyle.Render("Reset custom rules to defaults?"))
		content.WriteString("\n\n")
		content.WriteString(dimStyle.Render("[y] Yes  [n] No  [ESC] Cancel"))
		return content.String()
	}

	useCustom := model.ProjectConfig.Config.SmartCommit != nil &&
		model.ProjectConfig.Config.SmartCommit.UseCustomRules

	if useCustom {
		content.WriteString(enabledStyle.Render("[✓] Use Custom Rules"))
	} else {
		content.WriteString(disabledStyle.Render("[ ] Use Custom Rules"))
	}
	content.WriteString("  ")
	content.WriteString(dimStyle.Render("(press [space] to toggle)"))
	content.WriteString("\n\n")

	if !useCustom {
		content.WriteString(normalStyle.Render("Using default categorization rules"))
		content.WriteString("\n\n")
		content.WriteString(dimStyle.Render("[space] Enable custom rules • [ESC] Back • [s] Save"))
		return content.String()
	}

	content.WriteString(normalStyle.Render("Categories:"))
	content.WriteString("\n")

	for i, category := range model.Categories {
		if i == model.SelectedCategory {
			content.WriteString(selectedStyle.Render("→ " + category))
		} else {
			content.WriteString(normalStyle.Render("  " + category))
		}
		content.WriteString("\n")
	}

	content.WriteString("\n")

	category := model.Categories[model.SelectedCategory]
	rules := model.ProjectConfig.Config.SmartCommit.Categories[category]

	content.WriteString(selectedStyle.Render(category + " Category:"))
	content.WriteString("\n\n")

	content.WriteString(normalStyle.Render("Extensions: "))
	if len(rules.Extensions) > 0 {
		content.WriteString(strings.Join(rules.Extensions, ", "))
	} else {
		content.WriteString(dimStyle.Render("(none)"))
	}
	content.WriteString("\n")

	content.WriteString(normalStyle.Render("Patterns:   "))
	if len(rules.Patterns) > 0 {
		content.WriteString(strings.Join(rules.Patterns, ", "))
	} else {
		content.WriteString(dimStyle.Render("(none)"))
	}
	content.WriteString("\n\n")

	switch model.EditMode {
	case handlers.ModeAddExtension:
		content.WriteString(focusedStyle.Render("Add Extension: "))
		content.WriteString(model.ExtensionInput.View())
		content.WriteString("\n")
		content.WriteString(dimStyle.Render("Enter: save • ESC: cancel"))
	case handlers.ModeAddPattern:
		content.WriteString(focusedStyle.Render("Add Pattern: "))
		content.WriteString(model.PatternInput.View())
		content.WriteString("\n")
		content.WriteString(dimStyle.Render("Enter: save • ESC: cancel"))
	default:
		if model.Saved {
			content.WriteString(enabledStyle.Render("✓ Settings saved!"))
			content.WriteString("\n\n")
		}
		content.WriteString(dimStyle.Render(fmt.Sprintf("[↑↓] Navigate • [e] Add extension • [p] Add pattern\n")))
		content.WriteString(dimStyle.Render("[r] Reset • [s] Save • [ESC] Back"))
	}

	return content.String()
}
