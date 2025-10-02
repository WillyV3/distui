package views

import (
	"fmt"
	"strings"

	"distui/handlers"
	"distui/internal/models"
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
		content.WriteString("Config file:  ~/.distui/config.yaml:\n\n")
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
				"Primary GitHub:",
				"All Accounts:",
				"Homebrew Tap:",
				"NPM Scope:",
				"Version Bump:",
			}[i]

			content.WriteString(style.Render(fmt.Sprintf("%-20s", label)))
			content.WriteString(" ") // Add space between label and input
			content.WriteString(model.Inputs[i].View())
			if i == 1 {
				content.WriteString("\n")
				content.WriteString(subtleStyle.Render("                       (comma-separated, prefix with '@' for orgs: user1, @org1, user2)"))
			}
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

			// Show GitHub accounts
			content.WriteString("  GitHub Accounts:\n")

			// Build the complete accounts list including primary
			var allAccounts []models.GitHubAccount
			primaryUsername := model.Config.User.GitHubUsername
			primaryAdded := false

			// First, add all existing accounts
			for _, acc := range model.Config.User.GitHubAccounts {
				allAccounts = append(allAccounts, acc)
				if acc.Username == primaryUsername && !acc.IsOrg {
					primaryAdded = true
				}
			}

			// If primary wasn't in the list, add it first
			if primaryUsername != "" && !primaryAdded {
				primaryAccount := models.GitHubAccount{
					Username: primaryUsername,
					IsOrg:    false,
					Default:  true,
				}
				allAccounts = append([]models.GitHubAccount{primaryAccount}, allAccounts...)
			}

			// Display all accounts
			if len(allAccounts) > 0 {
				for _, acc := range allAccounts {
					prefix := "    •"
					if acc.Default || acc.Username == primaryUsername {
						prefix = "    ✓"
					}
					username := acc.Username
					if acc.IsOrg {
						username = "@" + username
					}
					content.WriteString(fmt.Sprintf("%s %s\n", prefix, username))
				}
			
			} else {
				content.WriteString("    (none)\n")
			}

			content.WriteString(fmt.Sprintf("\n  Homebrew Tap:   %s\n", model.Config.User.DefaultHomebrewTap))
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