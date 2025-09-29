package views

import (
	"fmt"
	"strings"

	"distui/handlers"
	"distui/internal/gitcleanup"
	"github.com/charmbracelet/lipgloss"
)

var (
	tealColor    = lipgloss.Color("#006666")
	controlStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
)

// RenderConfigureContent returns the content for the project configuration view
func RenderConfigureContent(project string, configModel *handlers.ConfigureModel) string {
	if configModel == nil {
		return "Loading configuration..."
	}

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true)

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Padding(0, 1)

	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("82"))

	var content strings.Builder

	content.WriteString(headerStyle.Render(fmt.Sprintf("CONFIGURE PROJECT: %s", project)) + "\n")

	// Show GitHub remote status (using cached values for performance)
	statusText := ""
	if gitcleanup.HasGitRepo() {
		if configModel.HasGitRemote && configModel.GitHubOwner != "" && configModel.GitHubRepo != "" {
			remoteURL := fmt.Sprintf("github.com/%s/%s", configModel.GitHubOwner, configModel.GitHubRepo)
			// Truncate if too long for terminal width
			if configModel.Width > 0 && len(remoteURL) > configModel.Width-20 {
				remoteURL = remoteURL[:configModel.Width-23] + "..."
			}
			if configModel.GitHubRepoExists {
				statusText = fmt.Sprintf("✓ Remote: %s", remoteURL)
				content.WriteString(successStyle.Render(statusText) + "\n")
			} else {
				statusText = fmt.Sprintf("⚠ Remote not found: %s", remoteURL)
				content.WriteString(statusStyle.Render(statusText) + "\n")
			}
		} else {
			statusText = "📦 No GitHub remote configured"
			content.WriteString(statusStyle.Render(statusText) + "\n")
		}
	} else {
		statusText = "Not a git repository"
		content.WriteString(controlStyle.Render(statusText) + "\n")
	}
	content.WriteString("\n")

	// Render tabs as flexbox-style boxes
	tabs := []string{"Distributions", "Build", "Advanced", "Cleanup"}

	// Calculate dynamic tab width based on window width
	// Distribute width evenly, accounting for rounding
	baseTabWidth := 18 // Default width
	extraWidth := 0
	if configModel.Width > 8 {
		baseTabWidth = configModel.Width / 4
		extraWidth = configModel.Width % 4 // Handle remainder
		if baseTabWidth < 12 {
			baseTabWidth = 12 // Minimum readable width
		}
		if baseTabWidth > 25 {
			baseTabWidth = 25 // Maximum reasonable width
			extraWidth = 0
		}
	}

	var renderedTabs []string
	for i, t := range tabs {
		// Give extra width to last tab to fill entire width
		tabWidth := baseTabWidth
		if i == 3 && extraWidth > 0 {
			tabWidth += extraWidth
		}

		// Create styles with appropriate width
		style := lipgloss.NewStyle().
			Width(tabWidth).
			Height(1).
			Align(lipgloss.Center)

		if i == configModel.ActiveTab {
			style = style.
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(tealColor)
		} else {
			style = style.
				Foreground(lipgloss.Color("240"))
		}

		renderedTabs = append(renderedTabs, style.Render(t))
	}

	// Join tabs horizontally
	content.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...) + "\n\n")

	// Create content area box that matches tab width
	contentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 0) // Vertical padding only, no horizontal

	// Set content box width to match total width
	if configModel.Width > 8 {
		// Use full width minus border
		contentBox = contentBox.Width(configModel.Width - 2)
	}

	// Render the active list or repo creation form
	if configModel.CreatingRepo {
		// Show repo creation form (available from any tab)
		formStyle := lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(tealColor).
			Padding(2).
			Width(configModel.Width - 4)

		var form strings.Builder

		// Show spinner if creating
		if configModel.IsCreating {
			form.WriteString(spinnerStyle.Render(configModel.CreateSpinner.View()) + " ")
			form.WriteString(configModel.CreateStatus + "\n")
		} else if configModel.CreateStatus != "" {
			form.WriteString(configModel.CreateStatus + "\n\n")
		} else {
			form.WriteString(headerStyle.Render("Create GitHub Repository") + "\n\n")
		}

		// Only show form fields if not currently creating
		if !configModel.IsCreating {
			form.WriteString("Repository Name:\n")
			nameView := configModel.RepoNameInput.View()
			if configModel.RepoInputFocus == 0 {
				nameView = "> " + nameView
			} else {
				nameView = "  " + nameView
			}
			form.WriteString(nameView + "\n\n")

			form.WriteString("Description:\n")
			descView := configModel.RepoDescInput.View()
			if configModel.RepoInputFocus == 1 {
				descView = "> " + descView
			} else {
				descView = "  " + descView
			}
			form.WriteString(descView + "\n\n")

			form.WriteString("Visibility:\n")
			visibilityText := ""
			if configModel.RepoIsPrivate {
				visibilityText = "[●] Private  [ ] Public"
			} else {
				visibilityText = "[ ] Private  [●] Public"
			}
			if configModel.RepoInputFocus == 2 {
				visibilityText = "> " + visibilityText + " [Space to toggle]"
			} else {
				visibilityText = "  " + visibilityText
			}
			form.WriteString(visibilityText + "\n\n")

			form.WriteString(controlStyle.Render("[Tab] Switch fields  [Enter] Create  [Esc] Cancel"))
		}

		content.WriteString(formStyle.Render(form.String()))
	} else if configModel.Initialized {
		// Wrap list content in the content box
		listContent := configModel.Lists[configModel.ActiveTab].View()
		content.WriteString(contentBox.Render(listContent))
	} else {
		content.WriteString(contentBox.Render("Initializing..."))
	}

	// Controls
	if configModel.Width > 0 {
		divider := strings.Repeat("─", configModel.Width)
		content.WriteString("\n" + controlStyle.Render(divider))
	} else {
		content.WriteString("\n" + controlStyle.Render("──────────────────────────────────────────"))
	}

	// Check if GitHub repo needs creation (using cached values)
	needsGitHub := false
	if gitcleanup.HasGitRepo() {
		if !configModel.HasGitRemote || !configModel.GitHubRepoExists {
			needsGitHub = true
		}
	}

	// Show appropriate controls based on active tab and GitHub status
	controlLine1 := ""
	controlLine2 := ""

	if configModel.ActiveTab == 3 {
		// Cleanup tab specific controls
		if needsGitHub {
			controlLine1 = "[Space] Cycle  [s] Execute  [G] GitHub  [r] Refresh"
		} else {
			controlLine1 = "[Space] Cycle Action  [s] Execute  [r] Refresh"
		}
		controlLine2 = "[Tab] Next Tab  [ESC] Cancel  [↑/↓] Navigate"
	} else {
		// Other tabs controls
		if needsGitHub {
			controlLine1 = "[Space] Toggle  [a] Check All  [G] GitHub  [Tab] Next"
		} else {
			controlLine1 = "[Space] Toggle  [a] Check All  [Tab] Next Tab"
		}
		controlLine2 = "[s] Save  [ESC] Cancel  [↑/↓] Navigate"
	}

	// Truncate control lines if needed
	if configModel.Width > 0 {
		if len(controlLine1) > configModel.Width {
			controlLine1 = controlLine1[:configModel.Width-3] + "..."
		}
		if len(controlLine2) > configModel.Width {
			controlLine2 = controlLine2[:configModel.Width-3] + "..."
		}
	}

	content.WriteString("\n" + controlStyle.Render(controlLine1))
	content.WriteString("\n" + controlStyle.Render(controlLine2))

	return content.String()
}

