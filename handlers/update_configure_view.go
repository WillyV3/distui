package handlers

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"distui/internal/config"
	"distui/internal/gitcleanup"
)

// UpdateConfigureView handles configure view updates and navigation
func UpdateConfigureView(currentPage, previousPage int, msg tea.Msg, configModel *ConfigureModel) (int, bool, tea.Cmd, *ConfigureModel) {
	// Model will be created in app.go with proper dimensions

	switch msg := msg.(type) {
	case repoCreatedMsg, pushCompleteMsg, commitCompleteMsg, filesGeneratedMsg:
		// Pass these messages directly to the model's Update
		if configModel != nil {
			newModel, cmd := configModel.Update(msg)
			return currentPage, false, cmd, newModel
		}
	case branchesLoadedMsg, pushResultMsg:
		// Route branch modal messages
		if configModel != nil && configModel.ShowingBranchModal && configModel.BranchModal != nil {
			newModal, cmd := configModel.BranchModal.Update(msg)
			*configModel.BranchModal = newModal

			// Check if push completed successfully - close modal and refresh
			if _, ok := msg.(pushResultMsg); ok {
				if !configModel.BranchModal.Pushing && configModel.BranchModal.Error == "" {
					// Success - close modal and refresh
					configModel.ShowingBranchModal = false
					configModel.CleanupModel.Refresh()
					return currentPage, false, nil, configModel
				}
			}
			return currentPage, false, cmd, configModel
		}
	case scanCompleteMsg:
		// Route scan result to RepoCleanupModel
		if configModel != nil && configModel.CurrentView == RepoCleanupView && configModel.RepoCleanupModel != nil {
			newModel, cmd := configModel.RepoCleanupModel.Update(msg)
			*configModel.RepoCleanupModel = newModel
			configModel.ScanningRepo = false
			return currentPage, false, cmd, configModel
		}
	case spinner.TickMsg:
		// Route spinner to repo cleanup model if scanning
		if configModel != nil && configModel.CurrentView == RepoCleanupView && configModel.RepoCleanupModel != nil {
			newModel, cmd := configModel.RepoCleanupModel.Update(msg)
			*configModel.RepoCleanupModel = newModel
			return currentPage, false, cmd, configModel
		}
		// Route spinner to branch modal if showing
		if configModel != nil && configModel.ShowingBranchModal && configModel.BranchModal != nil {
			newModal, cmd := configModel.BranchModal.Update(msg)
			*configModel.BranchModal = newModal
			return currentPage, false, cmd, configModel
		}
		// Otherwise pass to main model
		if configModel != nil {
			newModel, cmd := configModel.Update(msg)
			return currentPage, false, cmd, newModel
		}
	case tea.KeyMsg:
		// Handle branch modal first (highest priority when showing)
		if configModel.ShowingBranchModal && configModel.BranchModal != nil {
			newModal, cmd := configModel.BranchModal.Update(msg)
			*configModel.BranchModal = newModal

			// ESC closes modal immediately (handled in branch handler)
			if msg.String() == "esc" {
				configModel.ShowingBranchModal = false
				return currentPage, false, nil, configModel
			}

			return currentPage, false, cmd, configModel
		}

		// Handle view switching
		if configModel.CurrentView == CommitView {
			switch msg.String() {
			case "esc":
				configModel.CurrentView = TabView
				configModel.CommitModel = nil // Reset
				configModel.CleanupModel.Refresh()
				return currentPage, false, nil, configModel
			case "enter":
				// Only process if we're at the commit message stage and have files staged
				if configModel.CommitModel != nil && configModel.CommitModel.IsComplete() && configModel.CommitModel.HasStagedFiles() {
					// Execute the commit
					message := configModel.CommitModel.CommitMessage.Value()
					if message != "" {
						configModel.IsCreating = true
						configModel.CreateStatus = "Committing files..."
						// Stage the selected files and commit
						stagedFiles := configModel.CommitModel.GetStagedFiles()
						return currentPage, false, tea.Batch(
							configModel.CreateSpinner.Tick,
							regularCommitCmd(stagedFiles, message),
						), configModel
					}
				}
			default:
				// Pass key to commit model
				if configModel.CommitModel != nil {
					var cmd tea.Cmd
					configModel.CommitModel, cmd = configModel.CommitModel.Update(msg)
					return currentPage, false, cmd, configModel
				}
			}
		} else if configModel.CurrentView == SmartCommitConfirm {
			switch msg.String() {
			case "y", "Y":
				// Check if custom rules are enabled
				customRulesEnabled := configModel.ProjectConfig != nil &&
					configModel.ProjectConfig.Config != nil &&
					configModel.ProjectConfig.Config.SmartCommit != nil &&
					configModel.ProjectConfig.Config.SmartCommit.UseCustomRules

				// Create file selection model and switch to selection view
				changes, _ := gitcleanup.GetFileChanges()
				configModel.FileSelectionModel = NewFileSelectionModel(changes, customRulesEnabled, configModel.ProjectConfig)
				configModel.FileSelectionModel.Width = configModel.Width
				configModel.FileSelectionModel.Height = configModel.Height
				configModel.CurrentView = SmartCommitFileSelection
				return currentPage, false, nil, configModel

			case "n", "N", "esc":
				// User cancelled
				configModel.CurrentView = TabView
				return currentPage, false, nil, configModel
			}
		} else if configModel.CurrentView == SmartCommitFileSelection {
			switch msg.String() {
			case "esc":
				// Cancel file selection
				configModel.CurrentView = TabView
				configModel.FileSelectionModel = nil
				return currentPage, false, nil, configModel

			case "enter":
				// Commit selected files
				if configModel.FileSelectionModel != nil && configModel.FileSelectionModel.HasSelectedFiles() {
					items := configModel.FileSelectionModel.GetSelectedFiles()

					customRulesEnabled := configModel.ProjectConfig != nil &&
						configModel.ProjectConfig.Config != nil &&
						configModel.ProjectConfig.Config.SmartCommit != nil &&
						configModel.ProjectConfig.Config.SmartCommit.UseCustomRules

					configModel.CurrentView = TabView
					configModel.IsCreating = true
					if customRulesEnabled {
						configModel.CreateStatus = "Committing with custom rules..."
					} else {
						configModel.CreateStatus = "Committing selected files..."
					}
					configModel.FileSelectionModel = nil
					return currentPage, false, tea.Batch(
						configModel.CreateSpinner.Tick,
						smartCommitCmd(items),
					), configModel
				}
				return currentPage, false, nil, configModel

			default:
				// Pass keys to file selection model
				if configModel.FileSelectionModel != nil {
					var cmd tea.Cmd
					configModel.FileSelectionModel, cmd = configModel.FileSelectionModel.Update(msg)
					return currentPage, false, cmd, configModel
				}
			}
		} else if configModel.CurrentView == GenerateConfigConsent {
			switch msg.String() {
			case "y", "Y":
				// Start async file generation with spinner
				configModel.GeneratingFiles = true
				configModel.GenerateStatus = "Generating release files..."
				return currentPage, false, tea.Batch(
					configModel.CreateSpinner.Tick,
					generateFilesCmd(configModel.DetectedProject, configModel.ProjectConfig, configModel.PendingGenerateFiles, configModel.PendingDeleteFiles),
				), configModel
			case "n", "N", "esc":
				configModel.CurrentView = TabView
				configModel.PendingGenerateFiles = nil
				configModel.PendingDeleteFiles = nil
				return currentPage, false, nil, configModel
			}
		} else if configModel.CurrentView == GitHubView {
			switch msg.String() {
			case "esc":
				configModel.CurrentView = TabView
				configModel.CleanupModel.Refresh()
				return currentPage, false, nil, configModel
			default:
				newModel, cmd := configModel.GitHubModel.Update(msg)
				configModel.GitHubModel = newModel
				return currentPage, false, cmd, configModel
			}
		} else if configModel.CurrentView == SmartCommitPrefsView {
			// Always delegate to model first to handle edit mode transitions
			if configModel.SmartCommitPrefsModel != nil {
				// Check if we're in normal mode BEFORE processing ESC
				wasInNormalMode := configModel.SmartCommitPrefsModel.EditMode == ModeNormal

				newModel, cmd := configModel.SmartCommitPrefsModel.Update(msg)
				configModel.SmartCommitPrefsModel = newModel

				// Only exit to TabView if ESC was pressed while already in normal mode
				if msg.String() == "esc" && wasInNormalMode {
					configModel.CurrentView = TabView
					// Save any changes before returning
					if configModel.ProjectConfig != nil {
						config.SaveProject(configModel.ProjectConfig)
					}
					return currentPage, false, nil, configModel
				}
				return currentPage, false, cmd, configModel
			}
		} else if configModel.CurrentView == RepoCleanupView {
			if configModel.RepoCleanupModel != nil {
				if msg.String() == "esc" {
					configModel.CurrentView = TabView
					configModel.ScanningRepo = false
					configModel.CleanupModel.Refresh()
					return currentPage, false, nil, configModel
				}

				newModel, cmd := configModel.RepoCleanupModel.Update(msg)
				*configModel.RepoCleanupModel = newModel
				return currentPage, false, cmd, configModel
			}
		}

		// G key handler removed - repo browser is now embedded in cleanup view

		// Handle 'C' key to switch to Commit view (only in TabView)
		if msg.String() == "C" && configModel.CurrentView == TabView && configModel.ActiveTab == 0 {
			if configModel.CleanupModel != nil && configModel.CleanupModel.HasChanges() {
				// Initialize commit model if needed
				if configModel.CommitModel == nil {
					configModel.CommitModel = NewCommitModel(configModel.Width-2, configModel.Height-13)
				}
				configModel.CurrentView = CommitView
			}
			return currentPage, false, nil, configModel
		}

		// Handle 'p' key to switch to Smart Commit Preferences view (only in TabView, Cleanup tab)
		if msg.String() == "p" && configModel.CurrentView == TabView && configModel.ActiveTab == 0 {
			// Initialize preferences model if needed
			if configModel.SmartCommitPrefsModel == nil {
				configModel.SmartCommitPrefsModel = NewSmartCommitPrefsModel(configModel.ProjectConfig, configModel.Width-2, configModel.Height-13)
			}
			configModel.CurrentView = SmartCommitPrefsView
			return currentPage, false, nil, configModel
		}

		// Handle 'f' key to trigger repository file scan (only in TabView, Cleanup tab)
		if msg.String() == "f" && configModel.CurrentView == TabView && configModel.ActiveTab == 0 {
			if configModel.RepoCleanupModel == nil {
				model := NewRepoCleanupModel(configModel.Width-2, configModel.Height-13)
				configModel.RepoCleanupModel = &model
			}
			configModel.ScanningRepo = true
			configModel.CurrentView = RepoCleanupView
			return currentPage, false, configModel.RepoCleanupModel.Init(), configModel
		}

		// Handle 'R' key to confirm and generate/regenerate release files (only in TabView, not on Cleanup tab)
		if (msg.String() == "r" || msg.String() == "R") && configModel.CurrentView == TabView && configModel.ActiveTab != 0 {
			missing := CheckMissingConfigFiles(configModel.DetectedProject, configModel.ProjectConfig)
			if len(missing) > 0 {
				// Files don't exist - show consent for creation
				configModel.PendingGenerateFiles = missing
				configModel.PendingDeleteFiles = nil
			} else {
				// Files exist - check what needs to be regenerated/deleted
				changes := GetConfigFileChanges(configModel.DetectedProject, configModel.ProjectConfig)
				configModel.PendingGenerateFiles = changes.FilesToGenerate
				configModel.PendingDeleteFiles = changes.FilesToDelete
			}
			configModel.CurrentView = GenerateConfigConsent
			return currentPage, false, nil, configModel
		}

		// Handle 'P' key to open branch selection modal (only in TabView, Cleanup tab)
		if msg.String() == "P" && configModel.CurrentView == TabView && configModel.ActiveTab == 0 {
			if configModel.CleanupModel != nil && configModel.CleanupModel.RepoInfo != nil &&
				configModel.CleanupModel.RepoInfo.RemoteExists {
				// Open branch selection modal
				if configModel.BranchModal == nil {
					model := NewBranchSelectionModel(configModel.Width, configModel.Height)
					configModel.BranchModal = &model
				}
				configModel.ShowingBranchModal = true
				return currentPage, false, configModel.BranchModal.Init(), configModel
			}
			return currentPage, false, nil, configModel
		}


		// Handle 'G' key to create GitHub repository (only in TabView, Cleanup tab)
		if msg.String() == "G" && configModel != nil && !configModel.CreatingRepo && configModel.CurrentView == TabView && configModel.ActiveTab == 0 {
			// Check if we need to create a GitHub repo
			if gitcleanup.HasGitRepo() {
				if !gitcleanup.HasGitHubRemote() || !gitcleanup.CheckGitHubRepoExists() {
					// Enter repo creation mode
					configModel.CreatingRepo = true
					configModel.RepoInputFocus = 0

					// Create fresh inputs to avoid cursor issues
					nameInput := textinput.New()
					defaultName := gitcleanup.GetDefaultRepoName()
					nameInput.Placeholder = fmt.Sprintf("Repository name (default: %s)", defaultName)
					nameInput.CharLimit = 100
					if configModel.Width > 0 {
						nameInput.Width = configModel.Width - 4
					}
					// Explicitly clear value and reset before focusing
					nameInput.SetValue("")
					nameInput.Focus()
					// Clear again after focusing in case something got buffered
					nameInput.SetValue("")
					nameInput.CursorStart()
					configModel.RepoNameInput = nameInput

					descInput := textinput.New()
					descInput.Placeholder = "Repository description (optional)"
					descInput.CharLimit = 200
					if configModel.Width > 0 {
						descInput.Width = configModel.Width - 4
					}
					configModel.RepoDescInput = descInput
				}
			}
			// Always return here to consume the 'G' key
			return currentPage, false, nil, configModel
		}

		// Handle repo creation mode inputs
		if configModel != nil && configModel.CreatingRepo {
			switch msg.String() {
			case "esc":
				// Cancel repo creation
				configModel.CreatingRepo = false
				return currentPage, false, nil, configModel
			case "tab":
				// Cycle through name, description, private toggle, and account
				maxFields := 3
				if len(configModel.GitHubAccounts) > 0 {
					maxFields = 4
				}

				if configModel.RepoInputFocus == 0 {
					configModel.RepoInputFocus = 1
					configModel.RepoNameInput.Blur()
					configModel.RepoDescInput.Focus()
				} else if configModel.RepoInputFocus == 1 {
					configModel.RepoInputFocus = 2
					configModel.RepoDescInput.Blur()
				} else if configModel.RepoInputFocus == 2 && maxFields == 4 {
					configModel.RepoInputFocus = 3
				} else {
					configModel.RepoInputFocus = 0
					configModel.RepoNameInput.Focus()
				}
				return currentPage, false, nil, configModel
			case "enter":
				// Don't allow creation if already creating
				if configModel.IsCreating {
					return currentPage, false, nil, configModel
				}

				// Get repo details (use default if empty)
				repoName := configModel.RepoNameInput.Value()
				if repoName == "" {
					repoName = gitcleanup.GetDefaultRepoName()
				}
				repoDesc := configModel.RepoDescInput.Value()

				// Start creating with spinner
				configModel.IsCreating = true
				configModel.CreateStatus = "Creating repository..."

				// Get the owner (account) to create under
				owner := ""
				if len(configModel.GitHubAccounts) > 0 && configModel.SelectedAccountIdx < len(configModel.GitHubAccounts) {
					owner = configModel.GitHubAccounts[configModel.SelectedAccountIdx].Username
				}

				// Return commands for both spinner and repo creation
				return currentPage, false, tea.Batch(
					configModel.CreateSpinner.Tick,
					createRepoCmd(configModel.RepoIsPrivate, repoName, repoDesc, owner),
				), configModel
			case " ":
				// Toggle based on focus
				if configModel.RepoInputFocus == 2 {
					// Toggle private/public
					configModel.RepoIsPrivate = !configModel.RepoIsPrivate
					return currentPage, false, nil, configModel
				}
				if configModel.RepoInputFocus == 3 && len(configModel.GitHubAccounts) > 0 {
					// Cycle through accounts
					configModel.SelectedAccountIdx++
					if configModel.SelectedAccountIdx >= len(configModel.GitHubAccounts) {
						configModel.SelectedAccountIdx = 0
					}
					return currentPage, false, nil, configModel
				}
				// For text inputs, fall through to default to handle space as text
				fallthrough
			default:
				// Update the focused input
				if configModel.RepoInputFocus == 0 {
					configModel.RepoNameInput, _ = configModel.RepoNameInput.Update(msg)
				} else if configModel.RepoInputFocus == 1 {
					configModel.RepoDescInput, _ = configModel.RepoDescInput.Update(msg)
				}
				return currentPage, false, nil, configModel
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return currentPage, true, tea.Quit, configModel
		case "esc":
			// If in NPM edit mode, delegate to model's Update to handle cancellation
			if configModel != nil && configModel.NPMEditMode {
				newModel, cmd := configModel.Update(msg)
				return currentPage, false, cmd, newModel
			}
			// If we're in a nested view, return to TabView (shouldn't normally reach here)
			if configModel != nil && configModel.CurrentView != TabView {
				configModel.CurrentView = TabView
				if configModel.ProjectConfig != nil {
					config.SaveProject(configModel.ProjectConfig)
				}
				return currentPage, false, nil, configModel
			}
			// From TabView, go back to project view
			return 0, false, nil, configModel
		case "r":
			// Refresh git status in cleanup tab
			if configModel != nil && configModel.ActiveTab == 0 {
				configModel.refreshGitHubStatus()
				configModel.Lists[0].SetItems(configModel.loadGitStatus())
				if configModel.CleanupModel != nil {
					configModel.CleanupModel.Refresh()
				}
			}
			return currentPage, false, nil, configModel
		case "s":
			// Show smart commit confirmation
			if configModel != nil && configModel.ActiveTab == 0 && !configModel.IsCreating {
				if configModel.CleanupModel != nil && configModel.CleanupModel.HasChanges() {
					// Switch to confirmation view
					configModel.CurrentView = SmartCommitConfirm
					return currentPage, false, nil, configModel
				}
			}
			return currentPage, false, nil, configModel // stay in configure view
		default:
			// Let unhandled keys fall through to the model's Update
		}
	}

	// Update the model if it exists
	if configModel != nil {
		newModel, cmd := configModel.Update(msg)
		return currentPage, false, cmd, newModel
	}
	return currentPage, false, nil, configModel
}