package views

import (
	"fmt"
	"strings"

	"distui/handlers"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	headerStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
	subtleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	focusedButton  = focusedStyle.Render("[ Save ]")
	blurredButton  = fmt.Sprintf("[ %s ]", blurredStyle.Render("Save"))
)

func RenderSettingsContent(model *handlers.SettingsModel) string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("SETTINGS"))
	content.WriteString("\n\n")

	if model == nil {
		content.WriteString("Loading settings...")
		return content.String()
	}

	if model.Editing {
		content.WriteString("Configure distui settings:\n\n")

		// Render input fields
		for i := range model.Inputs {
			var style lipgloss.Style
			if i == model.FocusIndex {
				style = focusedStyle
			} else {
				style = blurredStyle
			}

			label := []string{
				"GitHub Username:",
				"Homebrew Tap:",
				"NPM Scope:",
				"Version Bump:",
			}[i]

			content.WriteString(style.Render(fmt.Sprintf("%-20s", label)))
			content.WriteString(model.Inputs[i].View())
			content.WriteString("\n")
		}

		content.WriteString("\n")

		// Render save button
		button := blurredButton
		if model.FocusIndex == len(model.Inputs) {
			button = focusedButton
		}
		content.WriteString(button)

		if model.Saved {
			content.WriteString("\n\n")
			content.WriteString(focusedStyle.Render("✓ Settings saved!"))
		}

		content.WriteString("\n\n")
		content.WriteString(subtleStyle.Render("Tab/Enter: next • Shift+Tab: prev • Esc: cancel"))
	} else {
		// Display current settings
		if model.Config != nil {
			content.WriteString("Current Configuration:\n\n")
			content.WriteString(fmt.Sprintf("  GitHub Username: %s\n", model.Config.User.GitHubUsername))
			content.WriteString(fmt.Sprintf("  Homebrew Tap:   %s\n", model.Config.User.DefaultHomebrewTap))
			content.WriteString(fmt.Sprintf("  NPM Scope:      %s\n", model.Config.User.NPMScope))
			content.WriteString(fmt.Sprintf("  Version Bump:   %s\n", model.Config.Preferences.DefaultVersionBump))
		} else {
			content.WriteString("No configuration found.\n")
		}

		content.WriteString("\n")
		content.WriteString(focusedStyle.Render("[e] Edit Settings"))
		content.WriteString("\n\n")
		content.WriteString(subtleStyle.Render("p: project • g: global • tab: cycle • q: quit"))
	}

	return content.String()
}