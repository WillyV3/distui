package views

import (
	"fmt"
	"strings"

	"distui/handlers"
	"distui/internal/gitcleanup"
	"github.com/charmbracelet/lipgloss"
)

var (
	tealColor      = lipgloss.Color("#006666")
	tabStyle       = lipgloss.NewStyle().Foreground(tealColor).Padding(0, 1)
	activeTabStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(tealColor).Padding(0, 1)
	controlStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	spinnerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
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

	// Render tabs with border
	tabs := []string{"Distributions", "Build Settings", "Advanced", "Cleanup"}
	var renderedTabs []string

	for i, t := range tabs {
		style := tabStyle
		if i == configModel.ActiveTab {
			style = activeTabStyle
		}
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	// Apply border with width constraints
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(tealColor)

	if configModel.Width > 0 {
		borderStyle = borderStyle.Width(configModel.Width - 2)
	}

	content.WriteString(borderStyle.Render(row) + "\n")

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
		content.WriteString(configModel.Lists[configModel.ActiveTab].View())
	} else {
		content.WriteString("Initializing...")
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

