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

	// Show spinner while loading or generating
	if configModel.Loading {
		spinnerView := spinnerStyle.Render(configModel.CreateSpinner.View()) + " Loading repository status..."
		return lipgloss.Place(
			configModel.Width,
			configModel.Height,
			lipgloss.Center,
			lipgloss.Center,
			spinnerView,
		)
	}

	if configModel.GeneratingFiles {
		spinnerView := spinnerStyle.Render(configModel.CreateSpinner.View()) + " " + configModel.GenerateStatus
		return lipgloss.Place(
			configModel.Width,
			configModel.Height,
			lipgloss.Center,
			lipgloss.Center,
			spinnerView,
		)
	}

	// Check for modal overlay first (highest priority)
	if configModel.ShowingBranchModal && configModel.BranchModal != nil {
		return RenderBranchSelection(*configModel.BranchModal)
	}

	// Check if we're in a sub-view
	switch configModel.CurrentView {
	case handlers.FirstTimeSetupView:
		return RenderFirstTimeSetup(configModel)
	case handlers.GitHubView:
		return RenderGitHubManagement(configModel.GitHubModel)
	case handlers.SmartCommitConfirm:
		return RenderSmartCommitConfirm(configModel.CleanupModel, configModel.ProjectConfig)
	case handlers.SmartCommitFileSelection:
		return RenderFileSelection(configModel.FileSelectionModel)
	case handlers.CommitView:
		return RenderCommitView(configModel.CommitModel)
	case handlers.GenerateConfigConsent:
		return RenderGenerateConfigConsent(configModel.PendingGenerateFiles, configModel.PendingDeleteFiles, configModel.Width, configModel.Height)
	case handlers.SmartCommitPrefsView:
		return RenderSmartCommitPrefs(configModel.SmartCommitPrefsModel)
	case handlers.RepoCleanupView:
		if configModel.RepoCleanupModel != nil {
			return RenderRepoCleanup(*configModel.RepoCleanupModel)
		}
		return "Loading cleanup view..."
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
	tabs := []string{"Cleanup", "Distributions", "Build", "Advanced"}

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

	// Create content area box that matches tab width and height
	contentBox := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 0) // Vertical padding only, no horizontal

	// Set content box dimensions
	if configModel.Width > 8 {
		// Use full width minus border
		contentBox = contentBox.Width(configModel.Width - 2)
	}

	// Set height constraint - use the listHeight calculation
	boxWidth := configModel.Width - 2
	if boxWidth < 40 {
		boxWidth = 40
	}
	// Calculate box height: handler already subtracted chrome based on warning state
	// Use the same calculation as the handler
	chromeLines := 13
	if configModel.NeedsRegeneration {
		chromeLines = 14
	}
	// Add NPM status lines when on Distributions tab and status exists
	if configModel.ActiveTab == 1 && configModel.NPMNameStatus == "unavailable" && len(configModel.NPMNameSuggestions) > 0 {
		chromeLines += 10 // 2 blanks + status + 2 blanks + header + 3 suggestions + help text
	} else if configModel.ActiveTab == 1 && configModel.NPMNameStatus != "" {
		chromeLines += 3 // 2 blanks + status line
	}
	boxHeight := configModel.Height - chromeLines
	if boxHeight < 5 {
		boxHeight = 5
	}
	contentBox = contentBox.Height(boxHeight)

	// First, render the base content
	var baseContent string
	if configModel.CreatingRepo {
		// Show repo creation form (available from any tab)
		boxWidth := configModel.Width - 2
		if boxWidth < 40 {
			boxWidth = 40
		}

		formWidth := boxWidth - 4
		if formWidth < 40 {
			formWidth = 40
		}

		formStyle := lipgloss.NewStyle().
			PaddingLeft(2).
			Width(formWidth)

		var form strings.Builder

		// Add top padding
		form.WriteString("\n")

		// Show spinner if creating
		if configModel.IsCreating {
			form.WriteString("  " + spinnerStyle.Render(configModel.CreateSpinner.View()) + " ")
			form.WriteString(configModel.CreateStatus + "\n")
		} else if configModel.CreateStatus != "" {
			form.WriteString("  " + configModel.CreateStatus + "\n\n")
		} else {
			form.WriteString("  " + headerStyle.Render("CREATE GITHUB REPOSITORY") + "\n\n")
		}

		// Only show form fields if not currently creating
		if !configModel.IsCreating {
			form.WriteString("  Repository Name:\n")
			nameView := configModel.RepoNameInput.View()
			if configModel.RepoInputFocus == 0 {
				nameView = "  > " + nameView
			} else {
				nameView = "    " + nameView
			}
			form.WriteString(nameView + "\n\n")

			form.WriteString("  Description:\n")
			descView := configModel.RepoDescInput.View()
			if configModel.RepoInputFocus == 1 {
				descView = "  > " + descView
			} else {
				descView = "    " + descView
			}
			form.WriteString(descView + "\n\n")

			form.WriteString("  Visibility:\n")
			if configModel.RepoInputFocus == 2 {
				form.WriteString("  > ")
			} else {
				form.WriteString("    ")
			}
			if configModel.RepoIsPrivate {
				form.WriteString("[●] Private  [ ] Public")
			} else {
				form.WriteString("[ ] Private  [●] Public")
			}
			if configModel.RepoInputFocus == 2 {
				form.WriteString("  [Space]")
			}
			form.WriteString("\n\n")

			// Account selection
			if len(configModel.GitHubAccounts) > 0 {
				form.WriteString("  Account:\n")

				// Show inline if 2 or fewer accounts, otherwise stack vertically
				if len(configModel.GitHubAccounts) <= 2 {
					if configModel.RepoInputFocus == 3 {
						form.WriteString("  > ")
					} else {
						form.WriteString("    ")
					}

					var parts []string
					for i, acc := range configModel.GitHubAccounts {
						checkbox := "[ ]"
						if i == configModel.SelectedAccountIdx {
							checkbox = "[●]"
						}

						accType := "user"
						if acc.IsOrg {
							accType = "org"
						}

						parts = append(parts, fmt.Sprintf("%s %s (%s)", checkbox, acc.Username, accType))
					}

					form.WriteString(strings.Join(parts, "  "))

					if configModel.RepoInputFocus == 3 {
						form.WriteString("  [Space]")
					}
					form.WriteString("\n")
				} else {
					// Multiple accounts - show vertically
					for i, acc := range configModel.GitHubAccounts {
						if configModel.RepoInputFocus == 3 && i == configModel.SelectedAccountIdx {
							form.WriteString("  > ")
						} else {
							form.WriteString("    ")
						}

						checkbox := "[ ]"
						if i == configModel.SelectedAccountIdx {
							checkbox = "[●]"
						}

						accType := "user"
						if acc.IsOrg {
							accType = "org"
						}

						form.WriteString(fmt.Sprintf("%s %s (%s)", checkbox, acc.Username, accType))

						if configModel.RepoInputFocus == 3 && i == configModel.SelectedAccountIdx {
							form.WriteString("  [Space]")
						}

						form.WriteString("\n")
					}
				}
				form.WriteString("\n")
			}

			form.WriteString("  " + controlStyle.Render("[Tab] Switch fields  [Enter] Create  [Esc] Cancel"))
		}

		baseContent = formStyle.Render(form.String())
	} else if configModel.ActiveTab == 0 {
		// Special handling for Cleanup tab - show status instead of list
		// Add status message if present
		statusMessage := ""
		if configModel.CreateStatus != "" && !configModel.IsCreating {
			statusMessage = configModel.CreateStatus
		}
		statusContent := RenderCleanupStatusWithMessage(configModel.CleanupModel, statusMessage, configModel.ProjectConfig)
		baseContent = statusContent
	} else if configModel.Initialized {
		// Wrap list content in the content box
		listContent := configModel.Lists[configModel.ActiveTab].View()
		baseContent = listContent
	} else {
		baseContent = "Initializing..."
	}

	// Render the content box with base content
	var renderedContent string
	if configModel.CreatingRepo {
		// Center the form in the content box
		centeredForm := lipgloss.Place(
			boxWidth,
			boxHeight,
			lipgloss.Center,
			lipgloss.Center,
			baseContent,
		)
		renderedContent = contentBox.Render(centeredForm)
	} else if configModel.IsCreating && configModel.CreateStatus != "" {
		// Create spinner overlay
		spinnerBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(tealColor).
			Padding(1).
			Width(40).
			Align(lipgloss.Center)

		spinnerContent := fmt.Sprintf("%s %s",
			configModel.CreateSpinner.View(),
			configModel.CreateStatus)

		overlayContent := lipgloss.Place(
			boxWidth,
			boxHeight,
			lipgloss.Center,
			lipgloss.Center,
			spinnerBox.Render(spinnerContent),
		)
		renderedContent = contentBox.Render(overlayContent)
	} else {
		renderedContent = contentBox.Render(baseContent)
	}
	content.WriteString(renderedContent)

	// Show NPM name edit UI if in edit mode
	if configModel.NPMEditMode {
		content.WriteString("\n\n")
		editStyle := lipgloss.NewStyle().Padding(0, 2).Foreground(lipgloss.Color("117"))
		content.WriteString(editStyle.Render("Edit NPM Package Name:"))
		content.WriteString("\n" + editStyle.Render(configModel.NPMNameInput.View()))
		content.WriteString("\n" + statusStyle.Render(
			lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("[Enter] Save  [ESC] Cancel")))
	} else if configModel.ActiveTab == 1 && configModel.NPMNameStatus != "" {
		// Show NPM name validation status (only on Distributions tab when not editing)
		content.WriteString("\n\n")

		statusStyle := lipgloss.NewStyle().Padding(0, 2)

		switch configModel.NPMNameStatus {
		case "checking":
			content.WriteString(statusStyle.Render(
				lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Render("⏳ Checking package name availability...")))
		case "available":
			message := "✓ Package name is available"
			if configModel.NPMNameError != "" {
				// Show ownership info if present
				message = "✓ " + configModel.NPMNameError
			}
			content.WriteString(statusStyle.Render(
				lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render(message)))
		case "unavailable":
			// Show the reason (e.g., "similar package exists: dist-ui")
			reason := "Package name unavailable"
			if configModel.NPMNameError != "" {
				reason = configModel.NPMNameError
			}
			warningMsg := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("✗ " + reason)
			content.WriteString(statusStyle.Render(warningMsg))

			if len(configModel.NPMNameSuggestions) > 0 {
				content.WriteString("\n\n" + statusStyle.Render(
					lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("Try one of these alternatives:")))
				for i, suggestion := range configModel.NPMNameSuggestions {
					if i >= 3 {
						break // Show max 3 suggestions
					}
					suggestionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
					content.WriteString("\n" + statusStyle.Render("  → "+suggestionStyle.Render(suggestion)))
				}
				content.WriteString("\n" + statusStyle.Render(
					lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("To change: ESC > s > e")))
			}
		case "error":
			content.WriteString(statusStyle.Render(
				lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗ Error checking name: "+configModel.NPMNameError)))
		}
	}

	// Controls
	if configModel.Width > 0 {
		divider := strings.Repeat("─", configModel.Width)
		content.WriteString("\n" + controlStyle.Render(divider))
	} else {
		content.WriteString("\n" + controlStyle.Render("──────────────────────────────────────────"))
	}

	// Show regeneration needed indicator
	if configModel.NeedsRegeneration && configModel.ActiveTab != 0 {
		warningStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)
		content.WriteString("\n" + warningStyle.Render("⚠ Configuration changed - Press [R] to regenerate release files"))
	}

	// Show appropriate controls based on active tab
	controlLine1 := ""
	controlLine2 := ""
	controlLine3 := ""

	if configModel.ActiveTab == 0 {
		// Cleanup tab specific controls
		controlLine1 = "[Space] Cycle  [s] Execute  [r] Refresh"
		controlLine2 = "[Tab] Next Tab  [ESC] Cancel  [↑/↓] Navigate"
	} else if configModel.ActiveTab == 1 {
		// Distributions tab - show hint about editing package name
		controlLine1 = "[Space] Toggle  [a] Check All  [e] Edit Package  [Tab] Next Tab"
		controlLine2 = "[R] Confirm & Generate Release Files  [ESC] Back"
	} else {
		// Other tabs controls
		controlLine1 = "[Space] Toggle  [a] Check All  [Tab] Next Tab"
		controlLine2 = "[R] Confirm & Generate Release Files  [ESC] Back"
	}

	// Truncate control lines if needed
	if configModel.Width > 0 {
		if len(controlLine1) > configModel.Width {
			controlLine1 = controlLine1[:configModel.Width-3] + "..."
		}
		if len(controlLine2) > configModel.Width {
			controlLine2 = controlLine2[:configModel.Width-3] + "..."
		}
		if len(controlLine3) > configModel.Width {
			controlLine3 = controlLine3[:configModel.Width-3] + "..."
		}
	}

	content.WriteString("\n" + controlStyle.Render(controlLine1))
	content.WriteString("\n" + controlStyle.Render(controlLine2))
	if controlLine3 != "" {
		hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
		content.WriteString("\n" + hintStyle.Render(controlLine3))
	}

	return content.String()
}

