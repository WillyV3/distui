package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"distui/handlers"
	"distui/internal/config"
	"distui/internal/detection"
	"distui/internal/models"
	"distui/views"
)

type pageState uint

const (
	projectView pageState = iota
	globalView
	settingsView
	configureView
	newProjectView
)

// Styles
var (
	titleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
	subtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	mainStyle   = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("#006666")).
			Padding(1, 2)
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	dotStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("236")).SetString(" • ")
)

type model struct {
	currentPage    pageState
	width          int
	height         int
	spinner        spinner.Model
	quitting       bool
	asciiArt       string

	// Real data
	globalConfig   *models.GlobalConfig
	currentProject *models.ProjectConfig
	allProjects    []models.ProjectConfig
	detectedProject *models.ProjectInfo

	// UI state
	selectedProjectIndex int
	configureModel      *handlers.ConfigureModel
	settingsModel       *handlers.SettingsModel
	globalModel         *handlers.GlobalModel
	releaseModel        *handlers.ReleaseModel
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("117"))

	// Load ASCII art
	asciiArt := ""
	if data, err := os.ReadFile("ascii-art-txt"); err == nil {
		asciiArt = string(data)
	}

	// Load global config
	globalConfig, err := config.LoadGlobalConfig()
	if err != nil {
		globalConfig = nil
	}

	// Try to detect current project
	detectedProject, err := detection.DetectProject(".")
	if err != nil {
		// Log the error for debugging but don't fail
		fmt.Fprintf(os.Stderr, "Detection error: %v\n", err)
	}

	// Load current project if it exists
	var currentProject *models.ProjectConfig
	if detectedProject != nil {
		currentProject, _ = config.LoadProject(detectedProject.Identifier)
	}

	// Load all projects (simplified for now)
	var allProjects []models.ProjectConfig
	if distui, err := config.LoadProject("distui"); err == nil {
		allProjects = append(allProjects, *distui)
	}

	// Always start at project view
	initialPage := projectView

	return model{
		currentPage:     initialPage,
		spinner:         s,
		asciiArt:        asciiArt,
		globalConfig:    globalConfig,
		currentProject:  currentProject,
		allProjects:     allProjects,
		detectedProject: detectedProject,
		configureModel:  nil,
		settingsModel:   nil,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tea.SetWindowTitle("distui - Go Release Manager"),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Don't return early - let the message pass through to handlers
	}

	// Update spinner
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)

	// Route to page handlers
	switch m.currentPage {
	case projectView:
		// Initialize releaseModel if needed
		if m.releaseModel == nil && m.width > 0 && m.height > 0 && m.detectedProject != nil {
			width := m.width - 4
			height := m.height - 4
			projectPath := m.detectedProject.Path
			projectName := m.detectedProject.Module.Name
			currentVersion := m.detectedProject.Module.Version
			repoOwner := ""
			repoName := ""
			if m.detectedProject.Repository != nil {
				repoOwner = m.detectedProject.Repository.Owner
				repoName = m.detectedProject.Repository.Name
			}
			m.releaseModel = handlers.NewReleaseModel(width, height, projectPath, projectName, currentVersion, repoOwner, repoName, m.currentProject)
		}

		newPage, quitting, pageCmd, newReleaseModel := handlers.UpdateProjectView(int(m.currentPage), int(projectView), msg, m.releaseModel, m.configureModel)
		m.releaseModel = newReleaseModel

		// Pre-create configure model if navigating to it
		if newPage == int(configureView) && m.configureModel == nil && m.width > 0 && m.height > 0 {
			width := m.width - 4   // border (2) + padding (2)
			height := m.height - 4 // border (2) + padding (2)
			accounts := extractGitHubAccounts(m.globalConfig)
			m.configureModel = handlers.NewConfigureModel(width, height, accounts, m.currentProject, m.detectedProject, m.globalConfig)
			// Change page NOW, start spinner and trigger async load
			m.currentPage = pageState(newPage)
			m.quitting = quitting
			listWidth := width - 2
			listHeight := height - 13

			// If first-time setup, also trigger auto-detection
			if m.configureModel.FirstTimeSetup {
				detectionCmd := handlers.StartDistributionDetectionCmd(m.detectedProject, m.globalConfig)
				return m, tea.Batch(cmd, pageCmd, m.configureModel.CreateSpinner.Tick, handlers.LoadCleanupCmd(listWidth, listHeight), detectionCmd, tea.ClearScreen)
			}

			return m, tea.Batch(cmd, pageCmd, m.configureModel.CreateSpinner.Tick, handlers.LoadCleanupCmd(listWidth, listHeight), tea.ClearScreen)
		}
		m.currentPage = pageState(newPage)
		m.quitting = quitting
		return m, tea.Batch(cmd, pageCmd)
	case globalView:
		if m.globalModel == nil {
			m.globalModel = handlers.NewGlobalModel(m.allProjects)
		}
		newPage, quitting, pageCmd, newGlobalModel := handlers.UpdateGlobalView(
			int(m.currentPage), int(projectView), msg, m.globalModel)
		m.currentPage = pageState(newPage)
		m.quitting = quitting
		m.globalModel = newGlobalModel
		// Sync selectedIndex back to model
		if m.globalModel != nil {
			m.selectedProjectIndex = m.globalModel.SelectedIndex
		}
		return m, tea.Batch(cmd, pageCmd)
	case settingsView:
		if m.settingsModel == nil {
			m.settingsModel = handlers.NewSettingsModel(m.globalConfig)
		}
		newPage, quitting, pageCmd, newSettingsModel := handlers.UpdateSettingsView(
			int(m.currentPage), int(projectView), msg, m.settingsModel)
		m.currentPage = pageState(newPage)
		m.quitting = quitting
		m.settingsModel = newSettingsModel
		return m, tea.Batch(cmd, pageCmd)
	case configureView:
		// Update dimensions on every frame if model exists
		if m.configureModel != nil && m.width > 0 && m.height > 0 {
			m.configureModel.Width = m.width - 4   // border (2) + padding (2)
			m.configureModel.Height = m.height - 4 // border (2) + padding (2)
		}
		newPage, quitting, pageCmd, newConfigModel := handlers.UpdateConfigureView(
			int(m.currentPage), int(projectView), msg, m.configureModel)
		m.currentPage = pageState(newPage)
		m.quitting = quitting
		m.configureModel = newConfigModel

		// Sync currentProject with configureModel's updated ProjectConfig
		if m.configureModel != nil && m.configureModel.ProjectConfig != nil {
			m.currentProject = m.configureModel.ProjectConfig
		}

		return m, tea.Batch(cmd, pageCmd)
	case newProjectView:
		newPage, quitting, pageCmd := handlers.UpdateNewProjectView(int(m.currentPage), int(globalView), msg)
		m.currentPage = pageState(newPage)
		m.quitting = quitting
		return m, tea.Batch(cmd, pageCmd)
	}

	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "\n  Goodbye!\n\n"
	}

	var s string

	// Route to appropriate view
	switch m.currentPage {
	case projectView:
		s = m.renderProjectView()
	case globalView:
		s = m.renderGlobalView()
	case settingsView:
		s = views.RenderSettingsContent(m.settingsModel)
	case configureView:
		s = m.renderConfigureView()
	case newProjectView:
		s = m.renderNewProjectView()
	default:
		s = "Unknown page"
	}

	// Use the window dimensions to create full-screen border
	// The content will fill the available space
	if m.width > 0 && m.height > 0 {
		// Calculate available space inside border and padding
		// ThickBorder = 2 chars on each side, padding = 2 on each side
		contentWidth := m.width - 2  // Account for border
		contentHeight := m.height - 2 // Account for border

		// Create a style that fills the window
		fullScreenStyle := lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("#006666")).
			Padding(1, 1).
			Width(contentWidth).
			Height(contentHeight)

		return fullScreenStyle.Render(s)
	}

	return mainStyle.Render(s)
}

func (m model) renderProjectView() string {
	return views.RenderProjectContent(m.detectedProject, m.currentProject, m.globalConfig, m.releaseModel, m.configureModel)
}

func (m model) renderGlobalView() string {
	deleteMode := false
	if m.globalModel != nil {
		deleteMode = m.globalModel.DeletingMode
	}
	return views.RenderGlobalContent(m.allProjects, m.selectedProjectIndex, deleteMode)
}

func (m model) renderConfigureView() string {
	projectName := "distui"
	if m.detectedProject != nil {
		projectName = m.detectedProject.Module.Name
	}
	return views.RenderConfigureContent(projectName, m.configureModel)
}

func (m model) renderNewProjectView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("New Project Setup"))
	b.WriteString("\n\n")

	if m.detectedProject != nil {
		b.WriteString("Detected Project Information:\n\n")
		b.WriteString(fmt.Sprintf("  Module: %s\n", m.detectedProject.Module.Name))
		b.WriteString(fmt.Sprintf("  Path:   %s\n", m.detectedProject.Path))
		if m.detectedProject.Repository != nil {
			b.WriteString(fmt.Sprintf("  Repo:   %s/%s\n",
				m.detectedProject.Repository.Owner,
				m.detectedProject.Repository.Name))
		}
		b.WriteString("\n")
		b.WriteString("[s] Save this project\n")
		b.WriteString("[e] Edit details\n")
		b.WriteString("[c] Cancel\n")
	} else {
		b.WriteString("No project detected in current directory.\n")
		b.WriteString("Please navigate to a Go project directory.\n")
	}

	return b.String()
}

func extractGitHubAccounts(cfg *models.GlobalConfig) []models.GitHubAccount {
	if cfg == nil {
		return []models.GitHubAccount{}
	}

	accounts := []models.GitHubAccount{}

	// Add legacy personal account first if configured
	if cfg.User.GitHubUsername != "" {
		accounts = append(accounts, models.GitHubAccount{
			Username: cfg.User.GitHubUsername,
			IsOrg:    false,
			Default:  true,
		})
	}

	// Add accounts from GitHubAccounts list
	for _, acc := range cfg.User.GitHubAccounts {
		accounts = append(accounts, acc)
	}

	return accounts
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}