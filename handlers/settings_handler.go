package handlers

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"distui/internal/config"
	"distui/internal/detection"
	"distui/internal/models"
)

type SettingsModel struct {
	FocusIndex      int
	Inputs          []textinput.Model
	Config          *models.GlobalConfig
	Editing         bool
	Saved           bool
	SelectedAccount int // For managing accounts list
}

func NewSettingsModel(globalConfig *models.GlobalConfig) *SettingsModel {
	m := &SettingsModel{
		Inputs: make([]textinput.Model, 5), // Added one more for accounts
		Config: globalConfig,
	}

	// Auto-detect user environment for better defaults
	userEnv, _ := detection.DetectUserEnvironment()

	// Create input fields
	for i := range m.Inputs {
		t := textinput.New()
		t.CharLimit = 64

		switch i {
		case 0:
			t.Placeholder = ""
			t.Focus()
			// Prefer detected value over config value
			if userEnv != nil && userEnv.GitHubUser != "" {
				t.SetValue(userEnv.GitHubUser)
			} else if globalConfig != nil && globalConfig.User.GitHubUsername != "" {
				t.SetValue(globalConfig.User.GitHubUsername)
			}
		case 1:
			t.Placeholder = ""
			// Show all accounts including primary as comma-separated with @ prefix for orgs
			var accounts []string

			// Always include primary account first if it exists
			primaryUsername := ""
			if userEnv != nil && userEnv.GitHubUser != "" {
				primaryUsername = userEnv.GitHubUser
			} else if globalConfig != nil && globalConfig.User.GitHubUsername != "" {
				primaryUsername = globalConfig.User.GitHubUsername
			}

			if primaryUsername != "" {
				accounts = append(accounts, primaryUsername)
			}

			// Add any additional accounts from GitHubAccounts list
			if globalConfig != nil && len(globalConfig.User.GitHubAccounts) > 0 {
				for _, acc := range globalConfig.User.GitHubAccounts {
					// Skip if it's the same as primary (avoid duplicates)
					if acc.Username == primaryUsername && !acc.IsOrg {
						continue
					}
					if acc.IsOrg {
						accounts = append(accounts, "@"+acc.Username)
					} else {
						accounts = append(accounts, acc.Username)
					}
				}
			}

			if len(accounts) > 0 {
				t.SetValue(strings.Join(accounts, ", "))
			}
			t.CharLimit = 256 // Longer for multiple accounts
		case 2:
			t.Placeholder = ""
			// Use existing config value or generate default from GitHub username
			if globalConfig != nil && globalConfig.User.DefaultHomebrewTap != "" {
				t.SetValue(globalConfig.User.DefaultHomebrewTap)
			} else if userEnv != nil && userEnv.GitHubUser != "" {
				t.SetValue(userEnv.GitHubUser + "/homebrew-tap")
			}
		case 3:
			t.Placeholder = ""
			if globalConfig != nil && globalConfig.User.NPMScope != "" {
				t.SetValue(globalConfig.User.NPMScope)
			} else if userEnv != nil && userEnv.GitHubUser != "" {
				t.SetValue("@" + userEnv.GitHubUser)
			}
		case 4:
			t.Placeholder = ""
			if globalConfig != nil && globalConfig.Preferences.DefaultVersionBump != "" {
				t.SetValue(globalConfig.Preferences.DefaultVersionBump)
			} else {
				t.SetValue("patch")
			}
		}

		m.Inputs[i] = t
	}

	// Migrate old single username to accounts list if needed
	if globalConfig != nil && globalConfig.User.GitHubUsername != "" {
		if len(globalConfig.User.GitHubAccounts) == 0 {
			globalConfig.User.GitHubAccounts = []models.GitHubAccount{
				{
					Username: globalConfig.User.GitHubUsername,
					Default:  true,
				},
			}
		}
	}

	return m
}

func UpdateSettingsView(currentPage, previousPage int, msg tea.Msg, model *SettingsModel) (int, bool, tea.Cmd, *SettingsModel) {
	if model == nil {
		model = NewSettingsModel(nil)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return currentPage, true, tea.Quit, model
		case "esc":
			if model.Editing {
				model.Editing = false
				return currentPage, false, nil, model
			}
			return 0, false, nil, model // projectView
		case "e":
			if !model.Editing {
				model.Editing = true
				return currentPage, false, nil, model
			}
		case "enter":
			if model.Editing {
				if model.FocusIndex == len(model.Inputs) {
					// Save button pressed
					model.saveConfig()
					model.Editing = false
					model.Saved = true
					return currentPage, false, nil, model
				}
				model.nextInput()
			}
		case "shift+tab", "up":
			if model.Editing {
				model.prevInput()
			}
		case "tab", "down":
			if model.Editing {
				model.nextInput()
			} else {
				return 3, false, nil, model // releaseView
			}
		case "p":
			return 0, false, nil, model // projectView
		case "g":
			return 1, false, nil, model // globalView
		case "r":
			return 3, false, nil, model // releaseView
		case "c":
			return 4, false, nil, model // configureView
		case "n":
			return 5, false, nil, model // newProjectView
		}
	}

	// Update inputs
	if model.Editing {
		cmd := model.updateInputs(msg)
		return currentPage, false, cmd, model
	}

	return currentPage, false, nil, model
}

func (m *SettingsModel) nextInput() {
	// Blur current input if it's a valid index
	if m.FocusIndex < len(m.Inputs) {
		m.Inputs[m.FocusIndex].Blur()
	}
	m.FocusIndex++
	if m.FocusIndex > len(m.Inputs) {
		m.FocusIndex = 0
	}
	if m.FocusIndex < len(m.Inputs) {
		m.Inputs[m.FocusIndex].Focus()
	}
}

func (m *SettingsModel) prevInput() {
	// Blur current input if it's a valid index
	if m.FocusIndex < len(m.Inputs) {
		m.Inputs[m.FocusIndex].Blur()
	}
	m.FocusIndex--
	if m.FocusIndex < 0 {
		m.FocusIndex = len(m.Inputs)
	}
	if m.FocusIndex < len(m.Inputs) {
		m.Inputs[m.FocusIndex].Focus()
	}
}

func (m *SettingsModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.Inputs))

	for i := range m.Inputs {
		m.Inputs[i], cmds[i] = m.Inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m *SettingsModel) saveConfig() {
	if m.Config == nil {
		m.Config = &models.GlobalConfig{
			Version: "1.0",
			Preferences: models.Preferences{
				ConfirmBeforeRelease: true,
				ShowCommandOutput:    true,
				AutoDetectProjects:   true,
			},
			UI: models.UIConfig{
				Theme:       "default",
				CompactMode: false,
				ShowHints:   true,
			},
		}
	}

	// Primary GitHub username from first field
	primaryUsername := m.Inputs[0].Value()
	m.Config.User.GitHubUsername = primaryUsername

	// Parse all accounts from comma-separated input with @ prefix for orgs
	accountsStr := m.Inputs[1].Value()
	m.Config.User.GitHubAccounts = []models.GitHubAccount{}

	if accountsStr != "" {
		accounts := strings.Split(accountsStr, ",")
		primaryFound := false

		for _, acc := range accounts {
			acc = strings.TrimSpace(acc)
			isOrg := false
			// Check for @ prefix to indicate organization
			if strings.HasPrefix(acc, "@") {
				isOrg = true
				acc = strings.TrimPrefix(acc, "@")
			}
			if acc != "" {
				isDefault := false
				// First occurrence of primary username (not as org) is the default
				if acc == primaryUsername && !isOrg && !primaryFound {
					isDefault = true
					primaryFound = true
				}
				m.Config.User.GitHubAccounts = append(m.Config.User.GitHubAccounts, models.GitHubAccount{
					Username: acc,
					IsOrg:    isOrg,
					Default:  isDefault,
				})
			}
		}

		// If primary wasn't in the list, add it as first/default
		if primaryUsername != "" && !primaryFound {
			// Prepend primary account
			primaryAccount := models.GitHubAccount{
				Username: primaryUsername,
				IsOrg:    false,
				Default:  true,
			}
			m.Config.User.GitHubAccounts = append([]models.GitHubAccount{primaryAccount}, m.Config.User.GitHubAccounts...)
		}
	} else if primaryUsername != "" {
		// If only primary exists, create single account entry
		m.Config.User.GitHubAccounts = []models.GitHubAccount{
			{
				Username: primaryUsername,
				IsOrg:    false,
				Default:  true,
			},
		}
	}

	m.Config.User.DefaultHomebrewTap = m.Inputs[2].Value()
	m.Config.User.NPMScope = m.Inputs[3].Value()
	m.Config.Preferences.DefaultVersionBump = m.Inputs[4].Value()

	config.SaveGlobalConfig(m.Config)
}