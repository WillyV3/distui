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
	SkipTests      bool  // From config: Run tests before release

	// Changelog
	ChangelogInput textinput.Model

	// Project config to check settings at runtime
	ProjectConfig *models.ProjectConfig

	// Channel for receiving output
	outputChan chan string
}

type Package struct {
	Name     string
	Status   string
	Output   []string
	Duration time.Duration
}

// Messages for progress updates
type ProgressTickMsg struct{}
type ReleaseOutputMsg struct {
	Line string
}

func tickProgress() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return ProgressTickMsg{}
	})
}

// Wait for output on the channel
func waitForOutput(sub chan string) tea.Cmd {
	return func() tea.Msg {
		line := <-sub
		return ReleaseOutputMsg{Line: line}
	}
}

func NewReleaseModel(width, height int, projectPath, projectName, currentVersion, repoOwner, repoName string, projectConfig *models.ProjectConfig) *ReleaseModel {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	p := progress.New(
		progress.WithScaledGradient("#00CED1", "#9370DB"), // Teal to Purple
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
	skipTests := false  // Default: run tests

	if projectConfig != nil && projectConfig.Config != nil {
		if projectConfig.Config.Distributions.Homebrew != nil {
			enableHomebrew = projectConfig.Config.Distributions.Homebrew.Enabled
			homebrewTap = projectConfig.Config.Distributions.Homebrew.TapRepo
		}
		if projectConfig.Config.Distributions.NPM != nil {
			enableNPM = projectConfig.Config.Distributions.NPM.Enabled
		}
		if projectConfig.Config.Release != nil {
			skipTests = projectConfig.Config.Release.SkipTests
		}
	}

	// Initialize changelog input
	changelogInput := textinput.New()
	changelogInput.Placeholder = "What's changed in this release?"
	changelogInput.CharLimit = 500
	changelogInput.Width = 60

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
		ChangelogInput:  changelogInput,
		SelectedVersion: 0,
		CurrentVersion:  currentVersion,
		ProjectPath:     projectPath,
		ProjectName:     projectName,
		RepoOwner:       repoOwner,
		RepoName:        repoName,
		EnableHomebrew:    enableHomebrew,
		EnableNPM:         enableNPM,
		HomebrewTap:       homebrewTap,
		SkipTests:         skipTests,
		ProjectConfig:     projectConfig,
	}
}

func (m *ReleaseModel) Update(msg tea.Msg) (*ReleaseModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		updatedModel, cmd := m.handleKeyPress(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		m = updatedModel

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case progress.FrameMsg:
		progressModel, cmd := m.Progress.Update(msg)
		m.Progress = progressModel.(progress.Model)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case ProgressTickMsg:
		// Gradually increment progress while release is running
		if m.Phase != models.PhaseComplete && m.Phase != models.PhaseFailed && m.Phase != models.PhaseVersionSelect {
			// Animate progress smoothly, but cap at 97%
			if m.Progress.Percent() < 0.97 {
				cmd := m.Progress.IncrPercent(0.006) // Faster increment (was 0.002)
				return m, tea.Batch(cmd, tickProgress())
			}
			// At 97%, keep ticking but don't increment
			return m, tickProgress()
		}
		return m, nil

	case ReleaseOutputMsg:
		// Add output line
		m.Output = append(m.Output, msg.Line)
		if len(m.Output) > 100 {
			m.Output = m.Output[1:]
		}

		// Detect phase changes based on output patterns
		currentPhase := m.detectPhaseFromOutput(msg.Line)
		if currentPhase >= 0 && currentPhase != m.Installing {
			// Mark previous as done
			if m.Installing >= 0 && m.Installing < len(m.Packages) {
				m.Packages[m.Installing].Status = "done"
				m.Installed = append(m.Installed, m.Installing)
			}
			// Mark new one as installing
			m.Installing = currentPhase
			if currentPhase < len(m.Packages) {
				m.Packages[currentPhase].Status = "installing"
			}
		}

		// Continue waiting for output
		return m, waitForOutput(m.outputChan)

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
			if msg.Error != nil {
				m.Error = msg.Error
			} else {
				m.Error = fmt.Errorf("failed at step: %s", msg.FailedStep)
			}

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
			// Manage input focus based on selection
			m.updateInputFocus()
		case "down", "j":
			if m.SelectedVersion < 4 {
				m.SelectedVersion++
			}
			// Manage input focus based on selection
			m.updateInputFocus()
		case "enter":
			return m.startRelease()
		}

		// Update custom version input if selected
		if m.SelectedVersion == 4 {
			var cmd tea.Cmd
			m.VersionInput, cmd = m.VersionInput.Update(msg)
			return m, cmd
		}

		// Update changelog input if changelog is enabled and a version is selected (not Configure Project)
		needsChangelog := false
		if m.ProjectConfig != nil && m.ProjectConfig.Config != nil && m.ProjectConfig.Config.Release != nil {
			needsChangelog = m.ProjectConfig.Config.Release.GenerateChangelog
		}
		if needsChangelog && m.SelectedVersion > 0 && m.SelectedVersion < 4 {
			var cmd tea.Cmd
			m.ChangelogInput, cmd = m.ChangelogInput.Update(msg)
			return m, cmd
		}
	}

	// Handle completion - ESC to dismiss
	if m.Phase == models.PhaseComplete {
		switch msg.String() {
		case "esc", "enter", " ":
			// Reset to initial state - user will return to project view
			m.Phase = models.PhaseVersionSelect
			m.Output = []string{}
			m.Error = nil
			m.Installing = -1
			m.Installed = []int{}
			m.SelectedVersion = 0
			for i := range m.Packages {
				m.Packages[i].Status = "pending"
			}
			return m, nil
		}
	}

	// Handle retry on failure
	if m.Phase == models.PhaseFailed {
		switch msg.String() {
		case "r", "R":
			// Reset to version selection for retry
			m.Phase = models.PhaseVersionSelect
			m.Output = []string{}
			m.Error = nil
			m.Installing = -1
			m.Installed = []int{}
			// Reset packages status
			for i := range m.Packages {
				m.Packages[i].Status = "pending"
			}
			return m, nil
		case "esc", "enter", " ":
			// Return to version selection on ESC/Enter/Space
			m.Phase = models.PhaseVersionSelect
			m.Output = []string{}
			m.Error = nil
			m.Installing = -1
			m.Installed = []int{}
			m.SelectedVersion = 0
			for i := range m.Packages {
				m.Packages[i].Status = "pending"
			}
			return m, nil
		}
	}

	return m, nil
}

// updateInputFocus manages focus for version input and changelog input based on current selection
func (m *ReleaseModel) updateInputFocus() {
	// Check if changelog is enabled
	needsChangelog := false
	if m.ProjectConfig != nil && m.ProjectConfig.Config != nil && m.ProjectConfig.Config.Release != nil {
		needsChangelog = m.ProjectConfig.Config.Release.GenerateChangelog
	}

	// Blur all inputs first
	m.VersionInput.Blur()
	m.ChangelogInput.Blur()

	// Focus appropriate input based on selection
	if m.SelectedVersion == 4 {
		// Custom version selected - focus version input
		m.VersionInput.Focus()
	} else if needsChangelog && m.SelectedVersion > 0 && m.SelectedVersion < 4 {
		// Release option selected (not Configure Project, not Custom) and changelog enabled
		m.ChangelogInput.Focus()
	}
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
	m.Installing = -1  // Not started yet
	m.Installed = []int{}
	m.outputChan = make(chan string, 100)

	releaseConfig := executor.ReleaseConfig{
		Version:        version,
		SkipTests:      m.SkipTests,  // Use config value instead of hardcoded false
		EnableHomebrew: m.EnableHomebrew,
		EnableNPM:      m.EnableNPM,
		HomebrewTap:    m.HomebrewTap,
		RepoOwner:      m.RepoOwner,
		RepoName:       m.RepoName,
		ProjectName:    m.ProjectName,
		Changelog:      m.ChangelogInput.Value(),
	}

	// Start with the progress at 0
	progressCmd := m.Progress.SetPercent(0)

	return m, tea.Batch(
		m.Spinner.Tick,
		progressCmd,
		tickProgress(),  // Start the progress animation
		waitForOutput(m.outputChan), // Start waiting for output
		m.runReleaseWithOutput(releaseConfig), // Run release and stream output
	)
}

func (m *ReleaseModel) runReleaseWithOutput(config executor.ReleaseConfig) tea.Cmd {
	return func() tea.Msg {
		// Actually run the release with real output streaming
		releaseExecutor := executor.NewReleaseExecutor(m.ProjectPath, config)
		return releaseExecutor.ExecuteReleasePhasesWithOutput(context.Background(), m.outputChan)()
	}
}

func (m *ReleaseModel) detectPhaseFromOutput(line string) int {
	line = strings.ToLower(line)

	if strings.Contains(line, "pre-flight") || strings.Contains(line, "preflight") {
		return 0
	}
	if strings.Contains(line, "test") {
		return 1
	}
	if strings.Contains(line, "tag") {
		return 2
	}
	if strings.Contains(line, "goreleaser") {
		return 3
	}
	if strings.Contains(line, "homebrew") || strings.Contains(line, "tap") {
		return 4
	}

	return -1
}

func (m *ReleaseModel) getSelectedVersion() string {
	baseVersion := m.CurrentVersion
	if baseVersion == "" {
		baseVersion = "v0.1.0"
	}

	// Index 0 is "Configure Project" (handled separately, never reaches here)
	// Index 1 is Patch, Index 2 is Minor, Index 3 is Major, Index 4 is Custom
	switch m.SelectedVersion {
	case 1:
		return bumpPatch(baseVersion)
	case 2:
		return bumpMinor(baseVersion)
	case 3:
		return bumpMajor(baseVersion)
	case 4:
		return m.VersionInput.Value()
	}

	return ""
}

func bumpPatch(version string) string {
	// Parse version like v1.2.3 or 1.2.3
	v := strings.TrimPrefix(version, "v")
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return version // Return unchanged if not semantic
	}

	// Parse patch number
	patch := 0
	fmt.Sscanf(parts[2], "%d", &patch)

	return fmt.Sprintf("v%s.%s.%d", parts[0], parts[1], patch+1)
}

func bumpMinor(version string) string {
	// Parse version like v1.2.3 or 1.2.3
	v := strings.TrimPrefix(version, "v")
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return version // Return unchanged if not semantic
	}

	// Parse minor number
	minor := 0
	fmt.Sscanf(parts[1], "%d", &minor)

	return fmt.Sprintf("v%s.%d.0", parts[0], minor+1)
}

func bumpMajor(version string) string {
	// Parse version like v1.2.3 or 1.2.3
	v := strings.TrimPrefix(version, "v")
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return version // Return unchanged if not semantic
	}

	// Parse major number
	major := 0
	fmt.Sscanf(parts[0], "%d", &major)

	return fmt.Sprintf("v%d.0.0", major+1)
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