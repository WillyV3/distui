package handlers

import (
	tea "github.com/charmbracelet/bubbletea"

	"distui/internal/config"
	"distui/internal/models"
	"distui/internal/workflow"
)

type WorkflowGenModel struct {
	ProjectConfig    *models.ProjectConfig
	ProjectPath      string
	PreviewMode      bool
	PreviewContent   string
	GeneratedYAML    string
	RequiredSecrets  []string
	ShowConfirm      bool
	ConfirmOverwrite bool
	Width            int
	Height           int
	Error            string
	Success          bool
}

func NewWorkflowGenModel(projectConfig *models.ProjectConfig, projectPath string) *WorkflowGenModel {
	if projectConfig.Config.CICD == nil {
		projectConfig.Config.CICD = &models.CICDSettings{
			GitHubActions: &models.GitHubActionsConfig{
				Enabled:        false,
				WorkflowPath:   ".github/workflows/release.yml",
				IncludeTests:   true,
				AutoRegenerate: false,
			},
		}
	}

	if projectConfig.Config.CICD.GitHubActions == nil {
		projectConfig.Config.CICD.GitHubActions = &models.GitHubActionsConfig{
			Enabled:        false,
			WorkflowPath:   ".github/workflows/release.yml",
			IncludeTests:   true,
			AutoRegenerate: false,
		}
	}

	secrets := workflow.GetRequiredSecrets(projectConfig)

	return &WorkflowGenModel{
		ProjectConfig:   projectConfig,
		ProjectPath:     projectPath,
		RequiredSecrets: secrets,
	}
}

func (m *WorkflowGenModel) Update(msg tea.Msg) (*WorkflowGenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.ShowConfirm {
			return m.handleConfirm(msg), nil
		}

		if m.PreviewMode {
			return m.handlePreview(msg), nil
		}

		return m.handleNormalMode(msg), nil
	}
	return m, nil
}

func (m *WorkflowGenModel) handleNormalMode(msg tea.KeyMsg) *WorkflowGenModel {
	switch msg.String() {
	case "space":
		m.toggleEnabled()
	case "t":
		m.toggleIncludeTests()
	case "a":
		m.toggleAutoRegenerate()
	case "p":
		m.showPreview()
	case "g":
		m.promptGenerate()
	case "s":
		m.saveConfig()
	}
	return m
}

func (m *WorkflowGenModel) handlePreview(msg tea.KeyMsg) *WorkflowGenModel {
	switch msg.String() {
	case "esc", "q":
		m.PreviewMode = false
		m.PreviewContent = ""
	}
	return m
}

func (m *WorkflowGenModel) handleConfirm(msg tea.KeyMsg) *WorkflowGenModel {
	switch msg.String() {
	case "y":
		m.generateWorkflow()
		m.ShowConfirm = false
	case "n", "esc":
		m.ShowConfirm = false
	}
	return m
}

func (m *WorkflowGenModel) toggleEnabled() {
	if m.ProjectConfig.Config.CICD.GitHubActions == nil {
		return
	}
	m.ProjectConfig.Config.CICD.GitHubActions.Enabled =
		!m.ProjectConfig.Config.CICD.GitHubActions.Enabled
}

func (m *WorkflowGenModel) toggleIncludeTests() {
	if m.ProjectConfig.Config.CICD.GitHubActions == nil {
		return
	}
	m.ProjectConfig.Config.CICD.GitHubActions.IncludeTests =
		!m.ProjectConfig.Config.CICD.GitHubActions.IncludeTests
}

func (m *WorkflowGenModel) toggleAutoRegenerate() {
	if m.ProjectConfig.Config.CICD.GitHubActions == nil {
		return
	}
	m.ProjectConfig.Config.CICD.GitHubActions.AutoRegenerate =
		!m.ProjectConfig.Config.CICD.GitHubActions.AutoRegenerate
}

func (m *WorkflowGenModel) showPreview() {
	yamlContent, err := workflow.GenerateWorkflow(m.ProjectConfig)
	if err != nil {
		m.Error = err.Error()
		return
	}

	m.PreviewContent = yamlContent
	m.PreviewMode = true
	m.Error = ""
}

func (m *WorkflowGenModel) promptGenerate() {
	if !m.ProjectConfig.Config.CICD.GitHubActions.Enabled {
		m.Error = "Enable workflow generation first"
		return
	}

	yamlContent, err := workflow.GenerateWorkflow(m.ProjectConfig)
	if err != nil {
		m.Error = err.Error()
		return
	}

	m.GeneratedYAML = yamlContent

	exists := workflow.WorkflowExists(m.ProjectPath)
	m.ConfirmOverwrite = exists
	m.ShowConfirm = true
	m.Error = ""
}

func (m *WorkflowGenModel) generateWorkflow() {
	if m.GeneratedYAML == "" {
		m.Error = "No workflow to generate"
		return
	}

	err := workflow.WriteWorkflowFile(m.ProjectPath, m.GeneratedYAML)
	if err != nil {
		m.Error = err.Error()
		return
	}

	m.Success = true
	m.Error = ""
	m.GeneratedYAML = ""
}

func (m *WorkflowGenModel) saveConfig() {
	if m.ProjectConfig == nil {
		return
	}
	config.SaveProject(m.ProjectConfig)
	m.Success = true
}
