package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	return r.ExecuteReleasePhasesWithOutput(ctx, nil)
}

func (r *ReleaseExecutor) ExecuteReleasePhasesWithOutput(ctx context.Context, outputChan chan<- string) tea.Cmd {
	return func() tea.Msg {
		defer func() {
			if outputChan != nil {
				close(outputChan)
			}
		}()

		startTime := time.Now()
		channels := []string{"GitHub"}

		// Helper to send output if channel is available
		sendOutput := func(msg string) {
			if outputChan != nil {
				select {
				case outputChan <- msg:
				default:
					// Don't block if channel is full
				}
			}
		}

		// We'll track completed phases in a slice to send back
		type phaseResult struct {
			phase    models.ReleasePhase
			duration time.Duration
			success  bool
		}
		completedPhases := []phaseResult{}

		// Pre-flight checks
		sendOutput("Starting pre-flight checks...")
		phaseStart := time.Now()
		if err := r.ValidatePreFlight(); err != nil {
			sendOutput("✗ Pre-flight checks failed: " + err.Error())
			return r.failureResult(startTime, "preflight", err, channels)
		}
		sendOutput("✓ Pre-flight checks passed")
		completedPhases = append(completedPhases, phaseResult{
			phase:    models.PhasePreFlight,
			duration: time.Since(phaseStart),
			success:  true,
		})

		// Tests
		if !r.config.SkipTests {
			sendOutput("Running tests...")
			phaseStart = time.Now()
			testCmd := RunTests(ctx, r.projectPath)
			msg := testCmd()
			if completeMsg, ok := msg.(models.CommandCompleteMsg); ok {
				if completeMsg.ExitCode != 0 {
					sendOutput("✗ Tests failed")
					return r.failureResult(startTime, "tests", completeMsg.Error, channels)
				}
			}
			sendOutput("✓ All tests passed")
			completedPhases = append(completedPhases, phaseResult{
				phase:    models.PhaseTests,
				duration: time.Since(phaseStart),
				success:  true,
			})
		}

		// Create and push tag
		sendOutput("Creating and pushing tag " + r.config.Version + "...")
		phaseStart = time.Now()
		if err := r.createAndPushTag(ctx); err != nil {
			sendOutput("✗ Tag creation failed: " + err.Error())
			return r.failureResult(startTime, "tag", err, channels)
		}
		sendOutput("✓ Tag created and pushed: " + r.config.Version)
		completedPhases = append(completedPhases, phaseResult{
			phase:    models.PhaseTag,
			duration: time.Since(phaseStart),
			success:  true,
		})

		// Run GoReleaser
		sendOutput("Running GoReleaser...")
		phaseStart = time.Now()
		goreleaserCmd := RunGoReleaserWithOutput(ctx, r.projectPath, r.config.Version, outputChan)
		msg := goreleaserCmd()
		if err, ok := msg.(error); ok {
			sendOutput("✗ GoReleaser failed: " + err.Error())
			return r.failureResult(startTime, "goreleaser", err, channels)
		}
		sendOutput("✓ GoReleaser completed successfully")
		completedPhases = append(completedPhases, phaseResult{
			phase:    models.PhaseGoReleaser,
			duration: time.Since(phaseStart),
			success:  true,
		})

		// Homebrew is handled by GoReleaser's brews configuration
		if r.config.EnableHomebrew {
			channels = append(channels, "Homebrew")
		}

		// NPM publish runs AFTER GoReleaser (needs the GitHub release to exist)
		if r.config.EnableNPM {
			sendOutput("Publishing to NPM...")
			phaseStart = time.Now()

			// Get package name from package.json
			pkgPath := filepath.Join(r.projectPath, "package.json")
			pkgData, err := os.ReadFile(pkgPath)
			if err != nil {
				sendOutput("✗ NPM publish failed: " + err.Error())
				// Don't fail the entire release, just log the error
			} else {
				var pkg map[string]interface{}
				if err := json.Unmarshal(pkgData, &pkg); err == nil {
					if pkgName, ok := pkg["name"].(string); ok {
						// Run NPM publish
						if err := PublishToNPM(r.projectPath, r.config.Version, pkgName, outputChan); err != nil {
							sendOutput("✗ NPM publish failed: " + err.Error())
							// Don't fail the entire release, NPM is optional
						} else {
							sendOutput("✓ NPM publish completed successfully")
							channels = append(channels, "NPM")
						}
					}
				}
			}

			completedPhases = append(completedPhases, phaseResult{
				phase:    models.PhaseNPM,
				duration: time.Since(phaseStart),
				success:  true, // Mark as success even if NPM failed (non-blocking)
			})
		}

		// Return success with all phases marked complete
		return models.ReleaseCompleteMsg{
			Success:    true,
			Version:    r.config.Version,
			Duration:   time.Since(startTime),
			Channels:   channels,
			TotalSteps: r.countSteps(),
		}
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
		Error:      err,
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