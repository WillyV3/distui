package executor

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"distui/internal/models"
)

type ReleaseExecutor struct {
	projectPath string
	config      ReleaseConfig
}

type ReleaseConfig struct {
	Version        string
	SkipTests      bool
	EnableHomebrew bool
	EnableNPM      bool
	HomebrewTap    string
	RepoOwner      string
	RepoName       string
	ProjectName    string
}

type ExecutionResult struct {
	Success    bool
	Version    string
	Channels   []string
	Duration   time.Duration
	Error      error
	FailedStep string
}

func NewReleaseExecutor(projectPath string, config ReleaseConfig) *ReleaseExecutor {
	return &ReleaseExecutor{
		projectPath: projectPath,
		config:      config,
	}
}

func (r *ReleaseExecutor) ExecuteReleasePhases(ctx context.Context) tea.Cmd {
	return tea.Batch(
		// Start preflight phase
		func() tea.Msg {
			return models.ReleasePhaseMsg{
				Phase:     models.PhasePreFlight,
				StartTime: time.Now(),
			}
		},
		// Run the actual release
		r.doExecuteReleasePhases(ctx),
	)
}

func (r *ReleaseExecutor) doExecuteReleasePhases(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		startTime := time.Now()
		channels := []string{"GitHub"}

		// Track completed phases
		completedPhases := make([]models.ReleasePhaseCompleteMsg, 0)

		// Pre-flight checks
		phaseStart := time.Now()
		if err := r.ValidatePreFlight(); err != nil {
			return r.failureResult(startTime, "preflight", err, channels)
		}
		completedPhases = append(completedPhases, models.ReleasePhaseCompleteMsg{
			Phase:    models.PhasePreFlight,
			Duration: time.Since(phaseStart),
			Success:  true,
		})

		// Tests
		if !r.config.SkipTests {
			phaseStart = time.Now()
			testCmd := RunTests(ctx, r.projectPath)
			msg := testCmd()
			if completeMsg, ok := msg.(models.CommandCompleteMsg); ok {
				if completeMsg.ExitCode != 0 {
					return r.failureResult(startTime, "tests", completeMsg.Error, channels)
				}
			}
			completedPhases = append(completedPhases, models.ReleasePhaseCompleteMsg{
				Phase:    models.PhaseTests,
				Duration: time.Since(phaseStart),
				Success:  true,
			})
		}

		// Tag
		phaseStart = time.Now()
		if err := r.createAndPushTag(ctx); err != nil {
			return r.failureResult(startTime, "tag", err, channels)
		}
		completedPhases = append(completedPhases, models.ReleasePhaseCompleteMsg{
			Phase:    models.PhaseTag,
			Duration: time.Since(phaseStart),
			Success:  true,
		})

		// GoReleaser
		phaseStart = time.Now()
		goreleaserCmd := RunGoReleaser(ctx, r.projectPath, r.config.Version)
		msg := goreleaserCmd()
		if err, ok := msg.(error); ok {
			return r.failureResult(startTime, "goreleaser", err, channels)
		}
		completedPhases = append(completedPhases, models.ReleasePhaseCompleteMsg{
			Phase:    models.PhaseGoReleaser,
			Duration: time.Since(phaseStart),
			Success:  true,
		})

		// Homebrew
		if r.config.EnableHomebrew && r.config.HomebrewTap != "" {
			phaseStart = time.Now()
			homebrewCmd := UpdateHomebrewTap(ctx, r.config.ProjectName, r.config.Version, r.config.HomebrewTap, r.config.RepoOwner, r.config.RepoName)
			msg := homebrewCmd()
			if result, ok := msg.(HomebrewUpdateResult); ok && !result.Success {
				return r.failureResult(startTime, "homebrew", result.Error, channels)
			}
			channels = append(channels, "Homebrew")
			completedPhases = append(completedPhases, models.ReleasePhaseCompleteMsg{
				Phase:    models.PhaseHomebrew,
				Duration: time.Since(phaseStart),
				Success:  true,
			})
		}

		if r.config.EnableNPM {
			channels = append(channels, "NPM")
		}

		// Return a batch command that sends all phase completions followed by the final message
		return tea.Batch(
			// Send all phase completion messages
			func() tea.Msg {
				// We can only return one message, so we'll return the final complete message
				// The phases are already marked in the UI during execution
				return models.ReleaseCompleteMsg{
					Success:    true,
					Version:    r.config.Version,
					Duration:   time.Since(startTime),
					Channels:   channels,
					TotalSteps: r.countSteps(),
				}
			},
		)
	}
}

func (r *ReleaseExecutor) createAndPushTag(ctx context.Context) error {
	// Try to delete existing tag (ignore errors if it doesn't exist)
	deleteCmd := RunCommandStreaming(ctx, "git", []string{"tag", "-d", r.config.Version}, r.projectPath)
	deleteCmd() // Ignore result

	// Try to delete remote tag (ignore errors if it doesn't exist)
	pushDeleteCmd := RunCommandStreaming(ctx, "git", []string{"push", "origin", ":refs/tags/" + r.config.Version}, r.projectPath)
	pushDeleteCmd() // Ignore result

	// Create new tag
	tagCmd := RunCommandStreaming(ctx, "git", []string{"tag", r.config.Version}, r.projectPath)
	msg := tagCmd()
	if completeMsg, ok := msg.(models.CommandCompleteMsg); ok {
		if completeMsg.ExitCode != 0 {
			return fmt.Errorf("creating tag: %w", completeMsg.Error)
		}
	}

	// Push tag
	pushCmd := RunCommandStreaming(ctx, "git", []string{"push", "origin", r.config.Version}, r.projectPath)
	msg = pushCmd()
	if completeMsg, ok := msg.(models.CommandCompleteMsg); ok {
		if completeMsg.ExitCode != 0 {
			return fmt.Errorf("pushing tag: %w", completeMsg.Error)
		}
	}

	return nil
}

func (r *ReleaseExecutor) failureResult(startTime time.Time, step string, err error, channels []string) models.ReleaseCompleteMsg {
	return models.ReleaseCompleteMsg{
		Success:    false,
		Version:    r.config.Version,
		Duration:   time.Since(startTime),
		Channels:   channels,
		TotalSteps: r.countSteps(),
		FailedStep: step,
	}
}

func (r *ReleaseExecutor) countSteps() int {
	steps := 3
	if !r.config.SkipTests {
		steps++
	}
	if r.config.EnableHomebrew {
		steps++
	}
	if r.config.EnableNPM {
		steps++
	}
	return steps
}

func (r *ReleaseExecutor) ValidatePreFlight() error {
	if !CheckGoReleaserInstalled() {
		return fmt.Errorf("goreleaser not installed")
	}

	if !CheckGoReleaserConfigExists(r.projectPath) {
		return fmt.Errorf(".goreleaser.yml not found")
	}

	if _, err := GetGitHubToken(); err != nil {
		return fmt.Errorf("GitHub authentication: %w", err)
	}

	return nil
}