package handlers

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"distui/internal/config"
	"distui/internal/detection"
	"distui/internal/fileops"
	"distui/internal/models"
)

// Update the configure model
func (m *ConfigureModel) Update(msg tea.Msg) (*ConfigureModel, tea.Cmd) {
	// Skip expensive list size updates unless it's a window resize
	// List sizes are already set correctly in NewConfigureModel and on WindowSizeMsg
	// Updating on every keystroke causes significant lag

	switch msg := msg.(type) {
	case struct{}:
		// Clear status message after timeout
		m.CreateStatus = ""
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.CreateSpinner, cmd = m.CreateSpinner.Update(msg)
		// Only continue ticking if we're showing the spinner
		if m.IsCreating || m.Loading || m.GeneratingFiles || m.NPMNameStatus == "checking" {
			return m, cmd
		}
		return m, nil
	case loadCompleteMsg:
		m.Loading = false
		m.Initialized = true
		m.CleanupModel = msg.cleanupModel
		m.Lists[0].SetItems(m.loadGitStatus())

		// Create project config file if it doesn't exist
		if m.ProjectConfig != nil && m.ProjectConfig.Project != nil {
			m.saveConfigWithRegenFlag(false) // Don't trigger regen warning on initial load
		}

		// Start background git watcher (polls every 2 seconds)
		if !m.GitWatcherActive {
			m.GitWatcherActive = true
			return m, StartGitWatcherCmd()
		}

		return m, nil

	case gitWatchTickMsg:
		// Background git status refresh
		return m.HandleGitWatchTick()

	case githubStatusMsg:
		// Update GitHub status from async load (non-blocking)
		m.HasGitRemote = msg.hasRemote
		m.GitHubOwner = msg.owner
		m.GitHubRepo = msg.repo
		m.GitHubRepoExists = msg.repoExists

		// Refresh cleanup model to reflect updated GitHub status
		if m.CleanupModel != nil {
			m.CleanupModel.Refresh()
		}

		return m, nil

	case repoCreatedMsg:
		m.IsCreating = false
		if msg.err == nil {
			// Success - refresh and clear inputs
			m.CreatingRepo = false
			m.RepoNameInput.SetValue("")
			m.RepoDescInput.SetValue("")
			m.RepoIsPrivate = false
			m.RepoInputFocus = 0
			m.refreshGitHubStatus()
			if m.CleanupModel != nil {
				m.CleanupModel.Refresh()
			}
			m.Lists[0].SetItems(m.loadGitStatus())
			m.CreateStatus = "✓ Repository created successfully!"
			// Clear status after 1 second
			return m, tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
				return struct{}{}
			})
		} else {
			m.CreateStatus = fmt.Sprintf("✗ Failed: %v", msg.err)
			// Clear status after 1 second
			return m, tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
				return struct{}{}
			})
		}
	case pushCompleteMsg:
		m.IsCreating = false
		if msg.err == nil {
			m.CreateStatus = "✓ Pushed to remote successfully!"
			if m.CleanupModel != nil {
				m.CleanupModel.Refresh()
			}
		} else {
			m.CreateStatus = fmt.Sprintf("✗ Push failed: %v", msg.err)
		}
		// Clear status after 1 second
		return m, tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
			return struct{}{}
		})
	case filesGeneratedMsg:
		m.GeneratingFiles = false
		if msg.err == nil {
			m.GenerateStatus = "✓ Release files updated successfully!"
			m.CurrentView = TabView
			m.PendingGenerateFiles = nil
			m.PendingDeleteFiles = nil
			m.NeedsRegeneration = false
			// Reload git status to show the newly generated files
			m.Lists[0].SetItems(m.loadGitStatus())
		} else {
			m.GenerateStatus = fmt.Sprintf("✗ Generation failed: %v", msg.err)
		}
		// Clear status after 1 second
		return m, tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
			return struct{}{}
		})
	case distributionDetectedMsg:
		m.DetectingDistributions = false

		// Apply fallbacks for missing values
		detectedHomebrewTap := msg.homebrewTap
		detectedHomebrewFormula := msg.homebrewFormula
		detectedNPMPackage := msg.npmPackage

		// Fallback: Use binary name for formula if empty
		if detectedHomebrewFormula == "" && m.DetectedProject != nil && m.DetectedProject.Binary != nil {
			detectedHomebrewFormula = m.DetectedProject.Binary.Name
		}

		// Initialize text inputs
		homebrewTap := textinput.New()
		homebrewTap.Placeholder = "owner/repo"
		homebrewTap.CharLimit = 100
		homebrewTap.Width = 40

		homebrewFormula := textinput.New()
		homebrewFormula.Placeholder = "formula-name"
		homebrewFormula.CharLimit = 100
		homebrewFormula.Width = 40

		npmPackage := textinput.New()
		npmPackage.Placeholder = "package-name or @scope/package-name"
		npmPackage.CharLimit = 214
		npmPackage.Width = 40

		// Check if config files exist and are custom (not distui-generated)
		customFiles := []string{}

		// Check .goreleaser.yaml - detect as custom if file exists and not distui-generated
		// Don't require homebrew config to be present in the file
		if m.DetectedProject != nil {
			goreleaserPath := filepath.Join(m.DetectedProject.Path, ".goreleaser.yaml")
			goreleaserYmlPath := filepath.Join(m.DetectedProject.Path, ".goreleaser.yml")
			if detection.IsCustomConfig(goreleaserPath) {
				customFiles = append(customFiles, ".goreleaser.yaml")
			} else if detection.IsCustomConfig(goreleaserYmlPath) {
				customFiles = append(customFiles, ".goreleaser.yml")
			}

			// Check package.json - detect as custom if file exists and not distui-generated
			packagePath := filepath.Join(m.DetectedProject.Path, "package.json")
			if detection.IsCustomConfig(packagePath) {
				customFiles = append(customFiles, "package.json")
			}
		}

		if len(customFiles) > 0 {
			// Custom files detected - show choice dialog with huh.Select
			m.FirstTimeSetupCustomChoice = true
			m.CustomFilesDetected = customFiles
			m.HomebrewTapInput = homebrewTap
			m.HomebrewFormulaInput = homebrewFormula
			m.NPMPackageInput = npmPackage
			homebrewTap.SetValue(detectedHomebrewTap)
			homebrewFormula.SetValue(detectedHomebrewFormula)
			npmPackage.SetValue(detectedNPMPackage)
			if msg.homebrewFromFile {
				m.HomebrewCheckEnabled = true
			}
			if msg.npmFromFile {
				m.NPMCheckEnabled = true
			}

			// Create huh.Select for custom file choice
			var choice string
			m.CustomFilesForm = huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Key("choice").
						Title("What would you like to do?").
						Options(
							huh.NewOption("Use distui-managed files", "distui"),
							huh.NewOption("Keep my custom files", "custom"),
						).
						Value(&choice),
				),
			)

			return m, nil
		} else if msg.homebrewFromFile || msg.npmFromFile {
			// Files are distui-generated or don't exist - safe to skip setup
			if msg.homebrewFromFile && detectedHomebrewTap != "" && detectedHomebrewFormula != "" {
				homebrewTap.SetValue(detectedHomebrewTap)
				homebrewFormula.SetValue(detectedHomebrewFormula)
				m.HomebrewCheckEnabled = true

				// Save Homebrew config
				if m.ProjectConfig.Config.Distributions.Homebrew == nil {
					m.ProjectConfig.Config.Distributions.Homebrew = &models.HomebrewConfig{}
				}
				m.ProjectConfig.Config.Distributions.Homebrew.Enabled = true
				m.ProjectConfig.Config.Distributions.Homebrew.TapRepo = detectedHomebrewTap
				m.ProjectConfig.Config.Distributions.Homebrew.FormulaName = detectedHomebrewFormula
			}

			if msg.npmFromFile && detectedNPMPackage != "" {
				npmPackage.SetValue(detectedNPMPackage)
				m.NPMCheckEnabled = true

				// Save NPM package name to config
				if m.ProjectConfig.Config.Distributions.NPM == nil {
					m.ProjectConfig.Config.Distributions.NPM = &models.NPMConfig{}
				}
				m.ProjectConfig.Config.Distributions.NPM.Enabled = true
				m.ProjectConfig.Config.Distributions.NPM.PackageName = detectedNPMPackage
			}

			m.HomebrewTapInput = homebrewTap
			m.HomebrewFormulaInput = homebrewFormula
			m.NPMPackageInput = npmPackage

			// Skip first-time setup - go straight to normal view
			m.FirstTimeSetup = false
			m.CurrentView = TabView
			m.ProjectConfig.FirstTimeSetupCompleted = true

			// Save config with detected distribution info
			config.SaveProject(m.ProjectConfig)
		} else if msg.homebrewExists || msg.npmExists {
			// Found in registry but no config files - show confirmation
			m.AutoDetected = true
			m.HomebrewDetectedFromFile = msg.homebrewFromFile
			m.NPMDetectedFromFile = msg.npmFromFile

			// Enable Homebrew if found in registry
			if msg.homebrewExists && detectedHomebrewTap != "" && detectedHomebrewFormula != "" {
				m.HomebrewCheckEnabled = true
				homebrewTap.SetValue(detectedHomebrewTap)
				homebrewFormula.SetValue(detectedHomebrewFormula)
			}

			// Enable NPM if found in registry
			if msg.npmExists && detectedNPMPackage != "" {
				m.NPMCheckEnabled = true
				npmPackage.SetValue(detectedNPMPackage)
			}

			m.HomebrewTapInput = homebrewTap
			m.HomebrewFormulaInput = homebrewFormula
			m.NPMPackageInput = npmPackage

			// Go to confirmation screen
			m.FirstTimeSetupConfirmation = true
		} else {
			// Nothing found, show manual input screen
			m.AutoDetected = false

			// Prefill with defaults from global config and detected project
			if detectedHomebrewTap != "" {
				homebrewTap.SetValue(detectedHomebrewTap)
			} else if m.GlobalConfig != nil && m.GlobalConfig.User.DefaultHomebrewTap != "" {
				homebrewTap.SetValue(m.GlobalConfig.User.DefaultHomebrewTap)
			}

			if detectedHomebrewFormula != "" {
				homebrewFormula.SetValue(detectedHomebrewFormula)
			}

			if detectedNPMPackage != "" {
				npmPackage.SetValue(detectedNPMPackage)
			} else if m.DetectedProject != nil && m.DetectedProject.Module != nil {
				npmPackage.SetValue(m.DetectedProject.Module.Name)
			}

			m.HomebrewTapInput = homebrewTap
			m.HomebrewFormulaInput = homebrewFormula
			m.NPMPackageInput = npmPackage
		}

		return m, nil

	case distributionVerifiedMsg:
		m.VerifyingDistributions = false
		if msg.err != nil {
			m.DistributionVerifyError = msg.err.Error()
			return m, nil
		}

		// Update project config with detected versions
		if msg.homebrewExists && msg.homebrewVersion != "" {
			if m.ProjectConfig.Config.Distributions.Homebrew == nil {
				m.ProjectConfig.Config.Distributions.Homebrew = &models.HomebrewConfig{}
			}
			m.ProjectConfig.Config.Distributions.Homebrew.Enabled = true
			m.ProjectConfig.Config.Distributions.Homebrew.TapRepo = m.HomebrewTapInput.Value()
			m.ProjectConfig.Config.Distributions.Homebrew.FormulaName = m.HomebrewFormulaInput.Value()
			if m.ProjectConfig.Project != nil && m.ProjectConfig.Project.Module != nil {
				m.ProjectConfig.Project.Module.Version = msg.homebrewVersion
			}
		}

		if msg.npmExists && msg.npmVersion != "" {
			if m.ProjectConfig.Config.Distributions.NPM == nil {
				m.ProjectConfig.Config.Distributions.NPM = &models.NPMConfig{}
			}
			m.ProjectConfig.Config.Distributions.NPM.Enabled = true
			m.ProjectConfig.Config.Distributions.NPM.PackageName = m.NPMPackageInput.Value()
			if m.ProjectConfig.Project != nil && m.ProjectConfig.Project.Module != nil {
				if msg.homebrewVersion == "" {
					m.ProjectConfig.Project.Module.Version = msg.npmVersion
				}
			}
		}

		// Save config and return to normal view
		m.FirstTimeSetup = false
		m.CurrentView = TabView
		// Mark first-time setup as completed
		m.ProjectConfig.FirstTimeSetupCompleted = true
		config.SaveProject(m.ProjectConfig)

		// Rebuild distributions list with updated config
		distItems := BuildDistributionsList(m.ProjectConfig, m.DetectedProject, m.NPMNameStatus)
		distributions := make([]list.Item, len(distItems))
		for i, item := range distItems {
			distributions[i] = item
		}
		m.Lists[1].SetItems(distributions)

		return m, nil
	case npmNameCheckMsg:
		m.NPMNameStatus = string(msg.result.Status)
		m.NPMNameError = msg.result.Error
		m.NPMNameSuggestions = msg.result.Suggestions

		// CRITICAL: Rebuild distributions list to show status indicator in NPM item
		distItems := BuildDistributionsList(m.ProjectConfig, m.DetectedProject, m.NPMNameStatus)
		distributions := make([]list.Item, len(distItems))
		for i, item := range distItems {
			distributions[i] = item
		}
		m.Lists[1].SetItems(distributions)

		// CRITICAL: Recalculate list height when NPM status changes
		// The NPM UI is appended to list content, so list must shrink to make room
		chromeLines := 15
		if m.NeedsRegeneration {
			chromeLines = 16
		}
		if m.ActiveTab == 1 && m.NPMNameStatus == "unavailable" && len(m.NPMNameSuggestions) > 0 {
			chromeLines += 9
		} else if m.ActiveTab == 1 && m.NPMNameStatus != "" {
			chromeLines += 3
		}
		listHeight := m.Height - chromeLines
		if listHeight < 5 {
			listHeight = 5
		}
		m.Lists[1].SetHeight(listHeight)

		return m, nil
	case commitCompleteMsg:
		m.IsCreating = false
		if msg.err == nil {
			m.CreateStatus = fmt.Sprintf("✓ Committed: %s", msg.message)
			if m.CleanupModel != nil {
				m.CleanupModel.Refresh()
			}
			m.Lists[0].SetItems(m.loadGitStatus())
			// Return to main view
			m.CurrentView = TabView
			m.CommitModel = nil // Clean up
		} else {
			m.CreateStatus = fmt.Sprintf("✗ Commit failed: %v", msg.err)
		}
		// Clear status after 1 second
		return m, tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
			return struct{}{}
		})
	case tea.WindowSizeMsg:
		// Note: app.go will update m.Width and m.Height after this handler returns
		// So we should use the current m.Width/m.Height which are already adjusted
		// by app.go (minus border and padding), not msg.Width/msg.Height

		// If model doesn't have dimensions yet, use msg dimensions minus app.go's chrome
		width := m.Width
		height := m.Height
		if width == 0 || height == 0 {
			width = msg.Width - 4   // border (2) + padding (2) from app.go View()
			height = msg.Height - 4
		}

		// Update list sizes with same calculation as NewConfigureModel
		// Total UI chrome: 4 (app wrapper) + 11 (view) = 15 lines, +1 if warning
		chromeLines := 15
		if m.NeedsRegeneration {
			chromeLines = 16
		}
		// NPM UI is appended to list content with "\n\n" prefix (2 blank lines)
		// NPM unavailable: 2 (blanks) + 7 (content) = 9, Other status: 2 (blanks) + 1 (status) = 3
		if m.ActiveTab == 1 && m.NPMNameStatus == "unavailable" && len(m.NPMNameSuggestions) > 0 {
			chromeLines += 9
		} else if m.ActiveTab == 1 && m.NPMNameStatus != "" {
			chromeLines += 3
		}
		listHeight := height - chromeLines
		if listHeight < 5 {
			listHeight = 5
		}
		// Content box has just border, no horizontal padding
		listWidth := width - 2
		if listWidth < 40 {
			listWidth = 40
		}

		// Update sub-models with CONTENT dimensions, not window dimensions
		if m.CleanupModel != nil {
			m.CleanupModel.Update(listWidth, listHeight)
		}
		if m.GitHubModel != nil {
			m.GitHubModel.SetSize(listWidth, listHeight)
		}
		if m.CommitModel != nil {
			m.CommitModel.SetSize(listWidth, listHeight)
		}
		if m.SmartCommitPrefsModel != nil {
			m.SmartCommitPrefsModel.SetSize(listWidth, listHeight)
		}
		for i := range m.Lists {
			m.Lists[i].SetWidth(listWidth)
			m.Lists[i].SetHeight(listHeight)
		}
		// Update text input widths for repo creation
		m.RepoNameInput.Width = msg.Width - 4
		m.RepoDescInput.Width = msg.Width - 4
		m.Initialized = true

	case tea.KeyMsg:
		// Handle overwrite warning modal first (highest priority)
		if m.ShowOverwriteWarning {
			switch msg.String() {
			case "y", "Y":
				// User confirms overwrite
				m.ShowOverwriteWarning = false
				// Proceed with save and regeneration
				m.saveConfigWithRegenFlag(true)
				m.FilesToOverwrite = nil
				return m, nil

			case "n", "N", "esc":
				// User cancels
				m.ShowOverwriteWarning = false
				m.FilesToOverwrite = nil
				// Revert the list item toggle
				if m.ActiveTab == 1 {
					// Re-toggle the distribution item back
					currentList := &m.Lists[m.ActiveTab]
					if selectedItem := currentList.SelectedItem(); selectedItem != nil {
						if i, ok := selectedItem.(DistributionItem); ok {
							i.Enabled = !i.Enabled // Toggle back
							items := currentList.Items()
							items[currentList.Index()] = i
							currentList.SetItems(items)
						}
					}
				}
				return m, nil
			}
			return m, nil // Consume all other inputs during warning
		}

		// Handle mode switch warning
		if m.CurrentView == ModeSwitchWarning {
			switch msg.String() {
			case "y", "Y":
				// User confirms switch
				filesToReplace := m.detectFilesToOverwrite()

				// Archive custom files
				backupPath, err := fileops.ArchiveCustomFiles(m.DetectedProject.Path, filesToReplace)
				if err != nil {
					m.CreateStatus = fmt.Sprintf("✗ Backup failed: %v", err)
					m.CurrentView = TabView
					return m, nil
				}

				// Set managed mode
				m.ProjectConfig.CustomFilesMode = false
				m.saveConfig()

				// Generate distui-managed files
				err = GenerateConfigFiles(m.DetectedProject, m.ProjectConfig, filesToReplace)
				if err != nil {
					m.CreateStatus = fmt.Sprintf("✗ Generation failed: %v", err)
					m.CurrentView = TabView
					return m, nil
				}

				m.CreateStatus = fmt.Sprintf("✓ Switched to managed mode. Backup: %s", backupPath)
				m.CurrentView = TabView
				return m, nil

			case "n", "N", "esc":
				// User cancels
				m.CurrentView = TabView
				return m, nil
			}
			return m, nil
		}

		// Handle first-time setup view
		if m.CurrentView == FirstTimeSetupView {
			return m.handleFirstTimeSetupKeys(msg)
		}

		// Handle NPM name editing mode first
		if m.NPMEditMode {
			switch msg.String() {
			case "enter":
				// Save the new package name
				newName := m.NPMNameInput.Value()
				if newName != "" {
					if m.ProjectConfig.Config.Distributions.NPM == nil {
						m.ProjectConfig.Config.Distributions.NPM = &models.NPMConfig{}
					}
					m.ProjectConfig.Config.Distributions.NPM.PackageName = newName
					m.saveConfig()

					// Trigger name check first (before rebuilding list)
					username := ""
					if m.DetectedProject != nil && m.DetectedProject.Repository != nil {
						username = m.DetectedProject.Repository.Owner
					}
					m.NPMNameStatus = "checking"

					// Rebuild distributions list with new package name and checking status
					distItems := BuildDistributionsList(m.ProjectConfig, m.DetectedProject, m.NPMNameStatus)
					listItems := make([]list.Item, len(distItems))
					for i, item := range distItems {
						listItems[i] = item
					}
					m.Lists[1].SetItems(listItems)
					m.NPMEditMode = false
					m.NPMNameInput.Blur()
					return m, tea.Batch(m.CreateSpinner.Tick, checkNPMNameCmd(newName, username))
				}
				m.NPMEditMode = false
				m.NPMNameInput.Blur()
				return m, nil
			case "esc":
				// Cancel editing
				m.NPMEditMode = false
				m.NPMNameInput.Blur()
				return m, nil
			default:
				// Pass to text input
				var cmd tea.Cmd
				m.NPMNameInput, cmd = m.NPMNameInput.Update(msg)
				return m, cmd
			}
		}

		// If we're on the cleanup tab and there are no changes, delegate navigation to the repo browser
		if m.ActiveTab == 0 && m.CleanupModel != nil && !m.CleanupModel.HasChanges() {
			// Check if this is a navigation key that should go to the repo browser
			switch msg.String() {
			case "j", "down", "k", "up", "g", "G", "h", "left", "backspace", "l", "right", "enter":
				var cmd tea.Cmd
				m.CleanupModel, cmd = m.CleanupModel.HandleKey(msg)
				return m, cmd
			}
		}

		switch msg.String() {
		case "tab":
			oldTab := m.ActiveTab
			m.ActiveTab = (m.ActiveTab + 1) % 4

			// Refresh cleanup tab when entering it
			if m.ActiveTab == 0 && oldTab != 0 {
				if m.CleanupModel == nil {
					// First time - load everything
					m.Loading = true
					listWidth := m.Width - 2
					listHeight := m.Height - 13
					return m, tea.Batch(m.CreateSpinner.Tick, LoadCleanupCmd(listWidth, listHeight), RefreshGitHubStatusCmd())
				} else {
					// Already loaded - just refresh GitHub status asynchronously
					return m, RefreshGitHubStatusCmd()
				}
			}

			// NPM name checking removed from tab switch for performance
			// NPM will only be checked when user explicitly enables it or changes package name

			return m, nil

		case "C":
			// Only allow in custom mode
			if m.ProjectConfig != nil && m.ProjectConfig.CustomFilesMode {
				// Store files to replace and show mode switch warning
				m.FilesToOverwrite = m.detectFilesToOverwrite()
				m.CurrentView = ModeSwitchWarning
				return m, nil
			}
			return m, nil

		case "shift+tab":
			oldTab := m.ActiveTab
			m.ActiveTab = (m.ActiveTab + 3) % 4

			// Refresh cleanup tab when entering it
			if m.ActiveTab == 0 && oldTab != 0 {
				if m.CleanupModel == nil {
					// First time - load everything
					m.Loading = true
					listWidth := m.Width - 2
					listHeight := m.Height - 13
					return m, tea.Batch(m.CreateSpinner.Tick, LoadCleanupCmd(listWidth, listHeight), RefreshGitHubStatusCmd())
				} else {
					// Already loaded - just refresh GitHub status asynchronously
					return m, RefreshGitHubStatusCmd()
				}
			}

			return m, nil
		case " ", "space":
			// Toggle the selected item
			if m.ActiveTab < 0 || m.ActiveTab >= len(m.Lists) {
				return m, nil
			}
			currentList := &m.Lists[m.ActiveTab]
			selectedItem := currentList.SelectedItem()
			if selectedItem == nil {
				return m, nil
			}

			if i, ok := selectedItem.(DistributionItem); ok {
				// If in custom mode, prompt to switch instead of saving
				if m.ProjectConfig != nil && m.ProjectConfig.CustomFilesMode {
					// Don't toggle - user needs to switch modes first
					// Revert the toggle
					items := currentList.Items()
					items[currentList.Index()] = i // Keep old value
					currentList.SetItems(items)
					// Store files to replace and show mode switch warning
					m.FilesToOverwrite = m.detectFilesToOverwrite()
					m.CurrentView = ModeSwitchWarning
					return m, nil
				}

				i.Enabled = !i.Enabled
				items := currentList.Items()
				items[currentList.Index()] = i
				currentList.SetItems(items)

				// Save config after toggle
				// In managed mode, files are always distui-generated, no warning needed
				m.saveConfig()

				// If NPM was just enabled, trigger name check immediately
				if i.Key == "npm" && i.Enabled {
					packageName := ""
					if m.ProjectConfig != nil && m.ProjectConfig.Config != nil &&
						m.ProjectConfig.Config.Distributions.NPM != nil {
						packageName = m.ProjectConfig.Config.Distributions.NPM.PackageName
					}
					// If no package name yet, use project name
					if packageName == "" && m.DetectedProject != nil {
						if m.DetectedProject.Binary != nil && m.DetectedProject.Binary.Name != "" {
							packageName = m.DetectedProject.Binary.Name
						} else if m.DetectedProject.Module != nil && m.DetectedProject.Module.Name != "" {
							packageName = m.DetectedProject.Module.Name
						}
					}

					username := ""
					if m.DetectedProject != nil && m.DetectedProject.Repository != nil {
						username = m.DetectedProject.Repository.Owner
					}

					m.NPMNameStatus = "checking"
					return m, tea.Batch(m.CreateSpinner.Tick, checkNPMNameCmd(packageName, username))
				} else if i.Key == "npm" && !i.Enabled {
					// NPM was disabled, clear status
					m.NPMNameStatus = ""
					m.NPMNameError = ""
					m.NPMNameSuggestions = nil
				}
			} else if i, ok := currentList.SelectedItem().(BuildItem); ok {
				i.Enabled = !i.Enabled
				items := currentList.Items()
				items[currentList.Index()] = i
				currentList.SetItems(items)
				// Save config after toggle
				m.saveConfig()
			} else if i, ok := currentList.SelectedItem().(CleanupItem); ok {
				// Special handling for GitHub repo creation
				if i.Category == "github-new" || i.Category == "github-push" {
					if i.Action == "create" {
						i.Action = "skip"
					} else {
						i.Action = "create"
					}
				} else {
					// Cycle through actions: commit -> skip -> ignore -> commit
					switch i.Action {
					case "commit":
						i.Action = "skip"
					case "skip":
						i.Action = "ignore"
					case "ignore":
						i.Action = "commit"
					default:
						i.Action = "commit"
					}
				}
				items := currentList.Items()
				items[currentList.Index()] = i
				currentList.SetItems(items)
			}
			return m, nil
		case "e":
			// Edit package name when on NPM item in Distributions tab
			if m.ActiveTab == 1 {
				selectedItem := m.Lists[1].SelectedItem()
				if dist, ok := selectedItem.(DistributionItem); ok && dist.Key == "npm" {
					// Enter edit mode
					m.NPMEditMode = true
					currentName := ""
					if m.ProjectConfig != nil && m.ProjectConfig.Config != nil &&
						m.ProjectConfig.Config.Distributions.NPM != nil {
						currentName = m.ProjectConfig.Config.Distributions.NPM.PackageName
					}
					m.NPMNameInput.SetValue(currentName)
					m.NPMNameInput.Focus()
					return m, nil
				}
			}
			return m, nil
		case "a":
			// Check/uncheck all in current tab
			if m.ActiveTab == 1 {
				items := m.Lists[1].Items()
				allChecked := true
				for _, item := range items {
					if dist, ok := item.(DistributionItem); ok && !dist.Enabled {
						allChecked = false
						break
					}
				}
				// Toggle all
				for i, item := range items {
					if dist, ok := item.(DistributionItem); ok {
						dist.Enabled = !allChecked
						items[i] = dist
					}
				}
				m.Lists[1].SetItems(items)
			}
			return m, nil
		default:
			// Don't pass navigation to the list if we're showing the repo browser
			if m.ActiveTab == 0 && m.CleanupModel != nil && !m.CleanupModel.HasChanges() {
				// Already handled above
				return m, nil
			}
			// Pass through to the active list
			var cmd tea.Cmd
			m.Lists[m.ActiveTab], cmd = m.Lists[m.ActiveTab].Update(msg)
			return m, cmd
		}
	}

	// Don't update the list if we're on cleanup tab with repo browser
	if m.ActiveTab == 0 && m.CleanupModel != nil && !m.CleanupModel.HasChanges() {
		return m, nil
	}

	// Update the active list
	var cmd tea.Cmd
	m.Lists[m.ActiveTab], cmd = m.Lists[m.ActiveTab].Update(msg)
	return m, cmd
}
