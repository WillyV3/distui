package handlers

import (
	tea "github.com/charmbracelet/bubbletea"
	"distui/internal/models"
)

type GlobalModel struct {
	Projects      []models.ProjectConfig
	SelectedIndex int
	DeletingMode  bool
}

func NewGlobalModel(projects []models.ProjectConfig) *GlobalModel {
	return &GlobalModel{
		Projects:      projects,
		SelectedIndex: 0,
	}
}

// UpdateGlobalView handles global view updates and navigation
func UpdateGlobalView(currentPage, previousPage int, msg tea.Msg, model *GlobalModel) (int, bool, tea.Cmd, *GlobalModel) {
	if model == nil {
		model = NewGlobalModel(nil)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return currentPage, true, tea.Quit, model
		case "p":
			return 0, false, nil, model // projectView
		case "r":
			return 3, false, nil, model // releaseView
		case "c":
			return 4, false, nil, model // configureView
		case "n":
			return 5, false, nil, model // newProjectView
		case "esc":
			if model.DeletingMode {
				model.DeletingMode = false
				return currentPage, false, nil, model
			}
			return 0, false, nil, model // projectView
		case "up", "k":
			if model.SelectedIndex > 0 {
				model.SelectedIndex--
			}
			return currentPage, false, nil, model
		case "down", "j":
			if model.SelectedIndex < len(model.Projects)-1 {
				model.SelectedIndex++
			}
			return currentPage, false, nil, model
		case "enter":
			if model.DeletingMode && len(model.Projects) > 0 {
				model.removeSelectedProject()
				model.DeletingMode = false
			} else if len(model.Projects) > 0 {
				// Open selected project in project view
				return 0, false, nil, model
			}
			return currentPage, false, nil, model
		case "a":
			// Add new project
			return 5, false, nil, model // newProjectView
		case "d":
			if !model.DeletingMode && len(model.Projects) > 0 {
				model.DeletingMode = true
			}
			return currentPage, false, nil, model
		}
	}
	return currentPage, false, nil, model
}

func (m *GlobalModel) removeSelectedProject() {
	if m.SelectedIndex >= 0 && m.SelectedIndex < len(m.Projects) {
		m.Projects = append(m.Projects[:m.SelectedIndex], m.Projects[m.SelectedIndex+1:]...)
		if m.SelectedIndex >= len(m.Projects) && m.SelectedIndex > 0 {
			m.SelectedIndex--
		}
	}
}