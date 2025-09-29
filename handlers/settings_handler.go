package handlers

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"distui/internal/config"
	"distui/internal/detection"
	"distui/internal/models"
)

type SettingsModel struct {
	FocusIndex int
	Inputs     []textinput.Model
	Config     *models.GlobalConfig
	Editing    bool
	Saved      bool
}

func NewSettingsModel(globalConfig *models.GlobalConfig) *SettingsModel {
	m := &SettingsModel{
		Inputs: make([]textinput.Model, 4),
		Config: globalConfig,
	}

	focusedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	cursorStyle := focusedStyle

	// Auto-detect user environment for better defaults
	userEnv, _ := detection.DetectUserEnvironment()

	// Create input fields
	for i := range m.Inputs {
		t := textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 64

		switch i {
		case 0:
			t.Placeholder = "GitHub Username"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
			// Prefer detected value over config value
			if userEnv != nil && userEnv.GitHubUser != "" {
				t.SetValue(userEnv.GitHubUser)
			} else if globalConfig != nil && globalConfig.User.GitHubUsername != "" {
				t.SetValue(globalConfig.User.GitHubUsername)
			}
		case 1:
			t.Placeholder = "Homebrew Tap (optional)"
			// Use existing config value or generate default from GitHub username
			if globalConfig != nil && globalConfig.User.DefaultHomebrewTap != "" {
				t.SetValue(globalConfig.User.DefaultHomebrewTap)
			} else if userEnv != nil && userEnv.GitHubUser != "" {
				t.SetValue(userEnv.GitHubUser + "/homebrew-tap")
			}
		case 2:
			t.Placeholder = "NPM Scope (optional)"
			if globalConfig != nil && globalConfig.User.NPMScope != "" {
				t.SetValue(globalConfig.User.NPMScope)
			} else if userEnv != nil && userEnv.GitHubUser != "" {
				t.SetValue("@" + userEnv.GitHubUser)
			}
		case 3:
			t.Placeholder = "Default Version Bump"
			if globalConfig != nil && globalConfig.Preferences.DefaultVersionBump != "" {
				t.SetValue(globalConfig.Preferences.DefaultVersionBump)
			} else {
				t.SetValue("patch")
			}
		}

		m.Inputs[i] = t
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
	m.Inputs[m.FocusIndex].Blur()
	m.FocusIndex++
	if m.FocusIndex > len(m.Inputs) {
		m.FocusIndex = 0
	}
	if m.FocusIndex < len(m.Inputs) {
		m.Inputs[m.FocusIndex].Focus()
	}
}

func (m *SettingsModel) prevInput() {
	m.Inputs[m.FocusIndex].Blur()
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

	m.Config.User.GitHubUsername = m.Inputs[0].Value()
	m.Config.User.DefaultHomebrewTap = m.Inputs[1].Value()
	m.Config.User.NPMScope = m.Inputs[2].Value()
	m.Config.Preferences.DefaultVersionBump = m.Inputs[3].Value()

	config.SaveGlobalConfig(m.Config)
}