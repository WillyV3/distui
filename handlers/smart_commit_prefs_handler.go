package handlers

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"distui/internal/config"
	"distui/internal/gitcleanup"
	"distui/internal/models"
)

type SmartCommitPrefsModel struct {
	ProjectConfig    *models.ProjectConfig
	Categories       []string
	SelectedCategory int
	EditMode         EditModeType
	ExtensionInput   textinput.Model
	PatternInput     textinput.Model
	Width            int
	Height           int
	ShowConfirm      bool
	Saved            bool
}

type EditModeType int

const (
	ModeNormal EditModeType = iota
	ModeAddExtension
	ModeAddPattern
)

func NewSmartCommitPrefsModel(projectConfig *models.ProjectConfig) *SmartCommitPrefsModel {
	if projectConfig.Config.SmartCommit == nil {
		projectConfig.Config.SmartCommit = &models.SmartCommitPrefs{
			Enabled:        true,
			UseCustomRules: false,
			Categories:     make(map[string]models.CategoryRules),
		}
	}

	extInput := textinput.New()
	extInput.Placeholder = ".ext"
	extInput.CharLimit = 32

	patInput := textinput.New()
	patInput.Placeholder = "**/*.ext"
	patInput.CharLimit = 128

	categories := []string{"code", "config", "docs", "build", "test", "assets", "data"}

	return &SmartCommitPrefsModel{
		ProjectConfig:    projectConfig,
		Categories:       categories,
		SelectedCategory: 0,
		ExtensionInput:   extInput,
		PatternInput:     patInput,
		EditMode:         ModeNormal,
	}
}

func (m *SmartCommitPrefsModel) Update(msg tea.Msg) (*SmartCommitPrefsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.ShowConfirm {
			return m.handleConfirm(msg), nil
		}

		switch m.EditMode {
		case ModeAddExtension:
			return m.handleExtensionInput(msg)
		case ModeAddPattern:
			return m.handlePatternInput(msg)
		default:
			return m.handleNormalMode(msg)
		}
	}
	return m, nil
}

func (m *SmartCommitPrefsModel) handleNormalMode(msg tea.KeyMsg) (*SmartCommitPrefsModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.SelectedCategory > 0 {
			m.SelectedCategory--
		}
	case "down", "j":
		if m.SelectedCategory < len(m.Categories)-1 {
			m.SelectedCategory++
		}
	case "space":
		m.toggleCustomRules()
	case "e":
		m.EditMode = ModeAddExtension
		m.ExtensionInput.Focus()
	case "p":
		m.EditMode = ModeAddPattern
		m.PatternInput.Focus()
	case "d":
		m.removeSelected()
	case "r":
		m.ShowConfirm = true
	case "s":
		m.saveConfig()
		m.Saved = true
	}
	return m, nil
}

func (m *SmartCommitPrefsModel) handleExtensionInput(msg tea.KeyMsg) (*SmartCommitPrefsModel, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.addExtension()
		m.EditMode = ModeNormal
		m.ExtensionInput.Blur()
		m.ExtensionInput.SetValue("")
	case "esc":
		m.EditMode = ModeNormal
		m.ExtensionInput.Blur()
		m.ExtensionInput.SetValue("")
	default:
		var cmd tea.Cmd
		m.ExtensionInput, cmd = m.ExtensionInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *SmartCommitPrefsModel) handlePatternInput(msg tea.KeyMsg) (*SmartCommitPrefsModel, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.addPattern()
		m.EditMode = ModeNormal
		m.PatternInput.Blur()
		m.PatternInput.SetValue("")
	case "esc":
		m.EditMode = ModeNormal
		m.PatternInput.Blur()
		m.PatternInput.SetValue("")
	default:
		var cmd tea.Cmd
		m.PatternInput, cmd = m.PatternInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *SmartCommitPrefsModel) handleConfirm(msg tea.KeyMsg) *SmartCommitPrefsModel {
	switch msg.String() {
	case "y":
		m.resetToDefaults()
		m.ShowConfirm = false
	case "n", "esc":
		m.ShowConfirm = false
	}
	return m
}

func (m *SmartCommitPrefsModel) toggleCustomRules() {
	if m.ProjectConfig.Config.SmartCommit == nil {
		return
	}
	m.ProjectConfig.Config.SmartCommit.UseCustomRules = !m.ProjectConfig.Config.SmartCommit.UseCustomRules

	if m.ProjectConfig.Config.SmartCommit.UseCustomRules && len(m.ProjectConfig.Config.SmartCommit.Categories) == 0 {
		m.initializeDefaultCategories()
	}
}

func (m *SmartCommitPrefsModel) initializeDefaultCategories() {
	defaultRules := gitcleanup.GetDefaultRules()
	m.ProjectConfig.Config.SmartCommit.Categories = defaultRules
}

func (m *SmartCommitPrefsModel) addExtension() {
	if m.ProjectConfig.Config.SmartCommit == nil || !m.ProjectConfig.Config.SmartCommit.UseCustomRules {
		return
	}

	ext := m.ExtensionInput.Value()
	if ext == "" {
		return
	}

	if ext[0] != '.' {
		ext = "." + ext
	}

	category := m.Categories[m.SelectedCategory]
	rules := m.ProjectConfig.Config.SmartCommit.Categories[category]
	rules.Extensions = append(rules.Extensions, ext)
	m.ProjectConfig.Config.SmartCommit.Categories[category] = rules
}

func (m *SmartCommitPrefsModel) addPattern() {
	if m.ProjectConfig.Config.SmartCommit == nil || !m.ProjectConfig.Config.SmartCommit.UseCustomRules {
		return
	}

	pattern := m.PatternInput.Value()
	if pattern == "" {
		return
	}

	category := m.Categories[m.SelectedCategory]
	rules := m.ProjectConfig.Config.SmartCommit.Categories[category]
	rules.Patterns = append(rules.Patterns, pattern)
	m.ProjectConfig.Config.SmartCommit.Categories[category] = rules
}

func (m *SmartCommitPrefsModel) removeSelected() {
	// Placeholder for removing selected extension/pattern
	// Would need additional state to track what's selected
}

func (m *SmartCommitPrefsModel) resetToDefaults() {
	if m.ProjectConfig.Config.SmartCommit == nil {
		return
	}
	m.ProjectConfig.Config.SmartCommit.UseCustomRules = false
	m.ProjectConfig.Config.SmartCommit.Categories = make(map[string]models.CategoryRules)
	m.saveConfig()
}

func (m *SmartCommitPrefsModel) saveConfig() {
	if m.ProjectConfig == nil {
		return
	}
	config.SaveProject(m.ProjectConfig)
}
