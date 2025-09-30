package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"distui/internal/executor"
	"distui/internal/models"
)

type ReleaseModel struct {
	Phase       models.ReleasePhase
	Packages    []Package
	Installing  int
	Installed   []int
	Progress    progress.Model
	Spinner     spinner.Model
	Output      []string
	Version     string
	StartTime   time.Time
	Error       error
	Width       int
	Height      int

	VersionInput textinput.Model
	SelectedVersion int
	CurrentVersion string

	ProjectPath string
	ProjectName string
	RepoOwner   string
	RepoName    string

	EnableHomebrew bool
	EnableNPM      bool
	HomebrewTap    string
}

type Package struct {
	Name     string
	Status   string
	Output   []string
	Duration time.Duration
}

func NewReleaseModel(width, height int, projectPath, projectName, currentVersion, repoOwner, repoName string, projectConfig *models.ProjectConfig) *ReleaseModel {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	ti := textinput.New()
	ti.Placeholder = "v0.1.0"
	ti.CharLimit = 20

	packages := []Package{
		{Name: "Pre-flight Checks", Status: "pending"},
		{Name: "Running Tests", Status: "pending"},
		{Name: "Creating Tag", Status: "pending"},
		{Name: "GoReleaser", Status: "pending"},
	}

	// Load config settings
	enableHomebrew := false
	enableNPM := false
	homebrewTap := ""

	if projectConfig != nil && projectConfig.Config != nil {
		if projectConfig.Config.Distributions.Homebrew != nil {
			enableHomebrew = projectConfig.Config.Distributions.Homebrew.Enabled
			homebrewTap = projectConfig.Config.Distributions.Homebrew.TapRepo
		}
		if projectConfig.Config.Distributions.NPM != nil {
			enableNPM = projectConfig.Config.Distributions.NPM.Enabled
		}
	}

	return &ReleaseModel{
		Phase:           models.PhaseVersionSelect,
		Packages:        packages,
		Installing:      -1,
		Installed:       []int{},
		Progress:        p,
		Spinner:         s,
		Output:          []string{},
		Width:           width,
		Height:          height,
		VersionInput:    ti,
		SelectedVersion: 0,
		CurrentVersion:  currentVersion,
		ProjectPath:     projectPath,
		ProjectName:     projectName,
		RepoOwner:       repoOwner,
		RepoName:        repoName,
		EnableHomebrew:  enableHomebrew,
		EnableNPM:       enableNPM,
		HomebrewTap:     homebrewTap,
	}
}

func (m *ReleaseModel) Update(msg tea.Msg) (*ReleaseModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		progressModel, cmd := m.Progress.Update(msg)
		m.Progress = progressModel.(progress.Model)
		return m, cmd

	case models.ReleasePhaseMsg:
		m.Phase = msg.Phase
		m.Installing = int(msg.Phase) - 1
		if m.Installing >= 0 && m.Installing < len(m.Packages) {
			m.Packages[m.Installing].Status = "installing"
		}
		return m, m.Spinner.Tick

	case models.ReleasePhaseCompleteMsg:
		idx := int(msg.Phase) - 1
		if idx >= 0 && idx < len(m.Packages) {
			if msg.Success {
				m.Packages[idx].Status = "done"
				m.Packages[idx].Duration = msg.Duration
				m.Installed = append(m.Installed, idx)
			} else {
				m.Packages[idx].Status = "failed"
			}
		}
		return m, nil

	case models.CommandOutputMsg:
		m.Output = append(m.Output, msg.Line)
		if len(m.Output) > 100 {
			m.Output = m.Output[1:]
		}
		return m, nil

	case models.ReleaseCompleteMsg:
		if msg.Success {
			m.Phase = models.PhaseComplete
		} else {
			m.Phase = models.PhaseFailed
			m.Error = fmt.Errorf("failed at step: %s", msg.FailedStep)
		}
		return m, nil
	}

	return m, nil
}

func (m *ReleaseModel) handleKeyPress(msg tea.KeyMsg) (*ReleaseModel, tea.Cmd) {
	if m.Phase == models.PhaseVersionSelect {
		switch msg.String() {
		case "up", "k":
			if m.SelectedVersion > 0 {
				m.SelectedVersion--
			}
		case "down", "j":
			if m.SelectedVersion < 3 {
				m.SelectedVersion++
			}
		case "enter":
			return m.startRelease()
		}

		if m.SelectedVersion == 3 {
			var cmd tea.Cmd
			m.VersionInput, cmd = m.VersionInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m *ReleaseModel) startRelease() (*ReleaseModel, tea.Cmd) {
	version := m.getSelectedVersion()
	if version == "" {
		return m, nil
	}

	m.Version = version
	m.Phase = models.PhasePreFlight
	m.StartTime = time.Now()

	releaseConfig := executor.ReleaseConfig{
		Version:        version,
		SkipTests:      false,
		EnableHomebrew: m.EnableHomebrew,
		EnableNPM:      m.EnableNPM,
		HomebrewTap:    m.HomebrewTap,
		RepoOwner:      m.RepoOwner,
		RepoName:       m.RepoName,
		ProjectName:    m.ProjectName,
	}

	releaseExecutor := executor.NewReleaseExecutor(m.ProjectPath, releaseConfig)

	return m, tea.Batch(
		m.Spinner.Tick,
		releaseExecutor.ExecuteReleasePhases(context.Background()),
	)
}

func (m *ReleaseModel) getSelectedVersion() string {
	baseVersion := m.CurrentVersion
	if baseVersion == "" {
		baseVersion = "v0.1.0"
	}

	switch m.SelectedVersion {
	case 0:
		return bumpPatch(baseVersion)
	case 1:
		return bumpMinor(baseVersion)
	case 2:
		return bumpMajor(baseVersion)
	case 3:
		return m.VersionInput.Value()
	}

	return ""
}

func bumpPatch(version string) string {
	return version
}

func bumpMinor(version string) string {
	return version
}

func bumpMajor(version string) string {
	return version
}

func UpdateReleaseView(currentPage, previousPage int, msg tea.Msg, releaseModel *ReleaseModel) (int, bool, tea.Cmd, *ReleaseModel) {
	if releaseModel == nil {
		return currentPage, false, nil, releaseModel
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return currentPage, true, tea.Quit, releaseModel
		case "esc":
			if releaseModel.Phase == models.PhaseVersionSelect {
				return 0, false, nil, releaseModel
			}
		}
	}

	updatedModel, cmd := releaseModel.Update(msg)
	return currentPage, false, cmd, updatedModel
}