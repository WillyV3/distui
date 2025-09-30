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
	// Start with the first phase
	return r.executePhase(ctx, models.PhasePreFlight, time.Now(), []string{"GitHub"}, 0)
}

func (r *ReleaseExecutor) executePhase(ctx context.Context, phase models.ReleasePhase, startTime time.Time, channels []string, completedCount int) tea.Cmd {
	return func() tea.Msg {

		var err error
		phaseStart := time.Now()

		// Execute the current phase
		switch phase {
		case models.PhasePreFlight:
			err = r.ValidatePreFlight()

		case models.PhaseTests:
			if !r.config.SkipTests {
				testCmd := RunTests(ctx, r.projectPath)
				msg := testCmd()
				if completeMsg, ok := msg.(models.CommandCompleteMsg); ok {
					if completeMsg.ExitCode != 0 {
						err = completeMsg.Error
					}
				}
			}

		case models.PhaseTag:
			err = r.createAndPushTag(ctx)

		case models.PhaseGoReleaser:
			goreleaserCmd := RunGoReleaser(ctx, r.projectPath, r.config.Version)
			msg := goreleaserCmd()
			if errMsg, ok := msg.(error); ok {
				err = errMsg
			}

		case models.PhaseHomebrew:
			if r.config.EnableHomebrew && r.config.HomebrewTap != "" {
				homebrewCmd := UpdateHomebrewTap(ctx, r.config.ProjectName, r.config.Version, r.config.HomebrewTap, r.config.RepoOwner, r.config.RepoName)
				msg := homebrewCmd()
				if result, ok := msg.(HomebrewUpdateResult); ok && !result.Success {
					err = result.Error
				} else {
					channels = append(channels, "Homebrew")
				}
			}
		}

		// If there was an error, return failure
		if err != nil {
			return r.failureResult(startTime, phase.String(), err, channels)
		}

		// Mark this phase as complete
		completedCount++

		// Determine next phase
		var nextPhase models.ReleasePhase
		var hasNext bool

		switch phase {
		case models.PhasePreFlight:
			if r.config.SkipTests {
				nextPhase = models.PhaseTag
			} else {
				nextPhase = models.PhaseTests
			}
			hasNext = true

		case models.PhaseTests:
			nextPhase = models.PhaseTag
			hasNext = true

		case models.PhaseTag:
			nextPhase = models.PhaseGoReleaser
			hasNext = true

		case models.PhaseGoReleaser:
			if r.config.EnableHomebrew && r.config.HomebrewTap != "" {
				nextPhase = models.PhaseHomebrew
				hasNext = true
			}

		case models.PhaseHomebrew:
			// No more phases
			hasNext = false
		}

		// If no more phases, we're done
		if !hasNext {
			return models.ReleaseCompleteMsg{
				Success:    true,
				Version:    r.config.Version,
				Duration:   time.Since(startTime),
				Channels:   channels,
				TotalSteps: r.countSteps(),
			}
		}

		// Continue to next phase
		return tea.Batch(
			// Mark current phase as complete
			func() tea.Msg {
				return models.ReleasePhaseCompleteMsg{
					Phase:    phase,
					Duration: time.Since(phaseStart),
					Success:  true,
				}
			},
			// Start next phase
			func() tea.Msg {
				return models.ReleasePhaseMsg{
					Phase:     nextPhase,
					StartTime: time.Now(),
				}
			},
			// Continue with next phase execution
			r.executePhase(ctx, nextPhase, startTime, channels, completedCount),
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