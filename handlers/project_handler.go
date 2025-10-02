package handlers

import (
	tea "github.com/charmbracelet/bubbletea"
	"distui/internal/detection"
	"distui/internal/models"
)

// UpdateProjectView handles project view updates and navigation
func UpdateProjectView(currentPage, previousPage int, msg tea.Msg, releaseModel *ReleaseModel, configureModel *ConfigureModel) (int, bool, tea.Cmd, *ReleaseModel) {
	if releaseModel != nil && releaseModel.Phase != models.PhaseVersionSelect {
		updatedModel, cmd := releaseModel.Update(msg)
		return currentPage, false, cmd, updatedModel
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if releaseModel != nil && releaseModel.Phase == models.PhaseVersionSelect {
			switch msg.String() {
			case "up", "k":
				if releaseModel.SelectedVersion > 0 {
					releaseModel.SelectedVersion--
				}
				return currentPage, false, nil, releaseModel
			case "down", "j":
				if releaseModel.SelectedVersion < 4 {
					releaseModel.SelectedVersion++
				}
				return currentPage, false, nil, releaseModel
			case "enter":
				// Check if "Configure Project" is selected (item 0)
				if releaseModel.SelectedVersion == 0 {
					return 3, false, tea.ClearScreen, releaseModel // Navigate to configureView
				}

				updatedModel, cmd := releaseModel.startRelease()
				return currentPage, false, cmd, updatedModel
			case "esc":
				releaseModel.Phase = models.PhaseVersionSelect
				releaseModel.SelectedVersion = 0
				return currentPage, false, nil, releaseModel
			}

			if releaseModel.SelectedVersion == 4 {
				var cmd tea.Cmd
				releaseModel.VersionInput, cmd = releaseModel.VersionInput.Update(msg)
				return currentPage, false, cmd, releaseModel
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return currentPage, true, tea.Quit, releaseModel
		case "g":
			return 1, false, tea.ClearScreen, releaseModel // globalView
		case "s":
			return 2, false, tea.ClearScreen, releaseModel // settingsView
		case "r":
			// Only block release if files are MISSING, not if they're custom
			// Check if required files exist (custom or distui-generated)
			hasGoreleaser := false
			if configureModel != nil && configureModel.DetectedProject != nil {
				goreleaserPaths := []string{
					configureModel.DetectedProject.Path + "/.goreleaser.yaml",
					configureModel.DetectedProject.Path + "/.goreleaser.yml",
				}
				for _, p := range goreleaserPaths {
					if detection.FileExists(p) {
						hasGoreleaser = true
						break
					}
				}
			}

			// Block only if .goreleaser.yaml is completely missing
			if !hasGoreleaser {
				// Missing required files - can't release
				return currentPage, false, nil, releaseModel
			}

			// Files exist (custom or distui-generated) - allow release
			if releaseModel != nil {
				releaseModel.Phase = models.PhaseVersionSelect
				releaseModel.SelectedVersion = 0
			}
			return currentPage, false, nil, releaseModel
		case "c":
			return 3, false, tea.ClearScreen, releaseModel // configureView
		case "n":
			return 4, false, tea.ClearScreen, releaseModel // newProjectView
		case "esc":
			if previousPage >= 0 {
				return previousPage, false, nil, releaseModel
			}
			return currentPage, false, nil, releaseModel
		}

	case models.ReleasePhaseMsg, models.ReleaseCompleteMsg, models.CommandCompleteMsg:
		if releaseModel != nil {
			updatedModel, cmd := releaseModel.Update(msg)
			return currentPage, false, cmd, updatedModel
		}
	}
	return currentPage, false, nil, releaseModel
}