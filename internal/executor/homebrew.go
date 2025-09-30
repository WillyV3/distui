package executor

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// HomebrewUpdateResult is kept for backwards compatibility
type HomebrewUpdateResult struct {
	Success      bool
	FormulaPath  string
	TapPath      string
	CommitHash   string
	Error        error
}

// UpdateHomebrewTap is deprecated - GoReleaser handles this automatically
// through its brews configuration in .goreleaser.yaml
func UpdateHomebrewTap(ctx context.Context, projectName string, version string, tapRepo string, repoOwner string, repoName string) tea.Cmd {
	return func() tea.Msg {
		// GoReleaser handles all Homebrew updates automatically
		// This function is kept for backwards compatibility only
		return HomebrewUpdateResult{
			Success: true,
			Error:   nil,
		}
	}
}

// CreateInitialFormula is deprecated - formulas are managed by GoReleaser
func CreateInitialFormula(projectName string, description string, repoOwner string, repoName string, version string, tapPath string) error {
	return fmt.Errorf("homebrew formula creation is handled by GoReleaser - configure in .goreleaser.yaml")
}

// CreateInitialFormulaWithSHA is deprecated - formulas are managed by GoReleaser
func CreateInitialFormulaWithSHA(projectName string, description string, repoOwner string, repoName string, version string, sha256sum string, tapPath string) error {
	return fmt.Errorf("homebrew formula creation is handled by GoReleaser - configure in .goreleaser.yaml")
}