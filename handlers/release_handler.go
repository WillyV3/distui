package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"distui/internal/executor"
	"distui/internal/gitcleanup"
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
	CompletedDuration time.Duration  // Store final duration when release completes
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

// Messages for progress updates
type ProgressTickMsg struct{}

func tickProgress() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return ProgressTickMsg{}
	})
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

	case ProgressTickMsg:
		// Gradually increment progress while release is running
		if m.Phase != models.PhaseComplete && m.Phase != models.PhaseFailed && m.Phase != models.PhaseVersionSelect {
			// Animate progress smoothly
			cmd := m.Progress.IncrPercent(0.002) // Slow increment
			return m, tea.Batch(cmd, tickProgress())
		}
		return m, nil

	case models.ReleasePhaseMsg:
		m.Phase = msg.Phase

		// Map phase to package index
		var pkgIdx int
		switch msg.Phase {
		case models.PhasePreFlight:
			pkgIdx = 0
		case models.PhaseTests:
			pkgIdx = 1
		case models.PhaseTag:
			pkgIdx = 2
		case models.PhaseGoReleaser:
			pkgIdx = 3
		case models.PhaseHomebrew:
			pkgIdx = 4
		default:
			pkgIdx = -1
		}

		if pkgIdx >= 0 && pkgIdx < len(m.Packages) {
			// Mark previous as done if exists
			if m.Installing >= 0 && m.Installing < len(m.Packages) && m.Packages[m.Installing].Status == "installing" {
				m.Packages[m.Installing].Status = "done"
			}
			// Mark new one as installing
			m.Installing = pkgIdx
			m.Packages[pkgIdx].Status = "installing"
		}
		return m, tea.Batch(m.Spinner.Tick, tickProgress())

	case models.ReleasePhaseCompleteMsg:
		// Map phase to package index
		var pkgIdx int
		switch msg.Phase {
		case models.PhasePreFlight:
			pkgIdx = 0
		case models.PhaseTests:
			pkgIdx = 1
		case models.PhaseTag:
			pkgIdx = 2
		case models.PhaseGoReleaser:
			pkgIdx = 3
		case models.PhaseHomebrew:
			pkgIdx = 4
		default:
			pkgIdx = -1
		}

		if pkgIdx >= 0 && pkgIdx < len(m.Packages) {
			if msg.Success {
				m.Packages[pkgIdx].Status = "done"
				m.Packages[pkgIdx].Duration = msg.Duration
				m.Installed = append(m.Installed, pkgIdx)
			} else {
				m.Packages[pkgIdx].Status = "failed"
			}
		}
		return m, tickProgress()

	case models.CommandOutputMsg:
		m.Output = append(m.Output, msg.Line)
		if len(m.Output) > 100 {
			m.Output = m.Output[1:]
		}
		return m, nil

	case models.ReleaseCompleteMsg:
		if msg.Success {
			m.Phase = models.PhaseComplete
			m.CompletedDuration = msg.Duration  // Capture the final duration

			// Mark all steps as complete
			for i := range m.Packages {
				if m.Packages[i].Status != "failed" {
					m.Packages[i].Status = "done"
					// Set a duration if not already set
					if m.Packages[i].Duration == 0 {
						m.Packages[i].Duration = msg.Duration / time.Duration(len(m.Packages))
					}
				}
			}
			m.Installed = make([]int, len(m.Packages))
			for i := range m.Installed {
				m.Installed[i] = i
			}
		} else {
			m.Phase = models.PhaseFailed
			m.CompletedDuration = msg.Duration  // Capture duration even on failure
			m.Error = fmt.Errorf("failed at step: %s", msg.FailedStep)

			// Mark the failed step and any after it
			failedFound := false
			for i, pkg := range m.Packages {
				if strings.ToLower(pkg.Name) == strings.ToLower(msg.FailedStep) ||
				   strings.Contains(strings.ToLower(pkg.Name), strings.ToLower(msg.FailedStep)) {
					m.Packages[i].Status = "failed"
					failedFound = true
				} else if failedFound {
					// Leave subsequent steps as pending
					m.Packages[i].Status = "pending"
				} else {
					// Steps before failure are complete
					m.Packages[i].Status = "done"
					m.Installed = append(m.Installed, i)
				}
			}
		}

		// Update progress to 100% if success, or proportional if failed
		var progressPercent float64
		if msg.Success {
			progressPercent = 1.0
		} else {
			progressPercent = float64(len(m.Installed)) / float64(len(m.Packages))
		}
		progressCmd := m.Progress.SetPercent(progressPercent)

		return m, progressCmd
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
	// Check if working tree is clean before starting release
	if !gitcleanup.IsWorkingTreeClean() {
		m.Phase = models.PhaseFailed
		m.Error = fmt.Errorf("working tree is not clean - commit or stash changes first")
		return m, nil
	}

	version := m.getSelectedVersion()
	if version == "" {
		return m, nil
	}

	m.Version = version
	m.Phase = models.PhasePreFlight
	m.StartTime = time.Now()
	m.Installing = 0  // Start with first phase
	m.Installed = []int{}

	// Mark first phase as installing
	if len(m.Packages) > 0 {
		m.Packages[0].Status = "installing"
	}

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

	// Start with the progress at 0
	progressCmd := m.Progress.SetPercent(0)

	return m, tea.Batch(
		m.Spinner.Tick,
		progressCmd,
		tickProgress(),  // Start the progress animation
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