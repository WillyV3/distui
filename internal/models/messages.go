package models

import "time"

type ReleasePhase int

const (
	PhaseVersionSelect ReleasePhase = iota
	PhasePreFlight
	PhaseTests
	PhaseTag
	PhaseGoReleaser
	PhaseHomebrew
	PhaseNPM
	PhaseComplete
	PhaseFailed
)

func (p ReleasePhase) String() string {
	switch p {
	case PhaseVersionSelect:
		return "Select Version"
	case PhasePreFlight:
		return "Pre-flight Checks"
	case PhaseTests:
		return "Running Tests"
	case PhaseTag:
		return "Creating Tag"
	case PhaseGoReleaser:
		return "GoReleaser"
	case PhaseHomebrew:
		return "Homebrew Tap"
	case PhaseNPM:
		return "NPM Publish"
	case PhaseComplete:
		return "Complete"
	case PhaseFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

type ReleasePhaseMsg struct {
	Phase     ReleasePhase
	StartTime time.Time
}

type ReleasePhaseCompleteMsg struct {
	Phase    ReleasePhase
	Duration time.Duration
	Success  bool
}

type CommandOutputMsg struct {
	Line  string
	IsErr bool
}

type CommandCompleteMsg struct {
	ExitCode int
	Error    error
	Duration time.Duration
}

type ReleaseCompleteMsg struct {
	Success      bool
	Version      string
	Duration     time.Duration
	Channels     []string
	TotalSteps   int
	FailedStep   string
	Error        error
}

type ReleaseErrorMsg struct {
	Phase        ReleasePhase
	Error        error
	CanRetry     bool
	CanSkip      bool
	CanRollback  bool
}