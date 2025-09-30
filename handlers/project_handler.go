package handlers

import (
	tea "github.com/charmbracelet/bubbletea"
	"distui/internal/models"
)

// UpdateProjectView handles project view updates and navigation
func UpdateProjectView(currentPage, previousPage int, msg tea.Msg, releaseModel *ReleaseModel) (int, bool, tea.Cmd, *ReleaseModel) {
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
				if releaseModel.SelectedVersion < 3 {
					releaseModel.SelectedVersion++
				}
				return currentPage, false, nil, releaseModel
			case "enter":
				updatedModel, cmd := releaseModel.startRelease()
				return currentPage, false, cmd, updatedModel
			case "esc":
				releaseModel.Phase = models.PhaseVersionSelect
				releaseModel.SelectedVersion = 0
				return currentPage, false, nil, releaseModel
			}

			if releaseModel.SelectedVersion == 3 {
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