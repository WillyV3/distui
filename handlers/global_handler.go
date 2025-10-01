package handlers

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"distui/internal/config"
	"distui/internal/fsearch"
	"distui/internal/models"
)

type GlobalModel struct {
	Projects           []models.ProjectConfig
	SelectedIndex      int
	Detecting          bool
	DetectStatus       string
	GlobalConfig       *models.GlobalConfig
	DetectSpinner      spinner.Model
	SettingWorkingDir  bool
	WorkingDirInput    textinput.Model
	WorkingDirResults  []string
	WorkingDirSelected int
	TargetProject      *models.ProjectConfig
}

func NewGlobalModel(projects []models.ProjectConfig, globalConfig *models.GlobalConfig) *GlobalModel {
	s := spinner.New()
	s.Spinner = spinner.Jump
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("117"))

	dirInput := textinput.New()
	dirInput.Placeholder = "Type to search directories..."
	dirInput.CharLimit = 200
	dirInput.Width = 60

	return &GlobalModel{
		Projects:        projects,
		SelectedIndex:   0,
		GlobalConfig:    globalConfig,
		DetectSpinner:   s,
		WorkingDirInput: dirInput,
	}
}

// UpdateGlobalView handles global view updates and navigation
func UpdateGlobalView(currentPage, previousPage int, msg tea.Msg, model *GlobalModel, globalConfig *models.GlobalConfig) (int, bool, tea.Cmd, *GlobalModel) {
	if model == nil {
		model = NewGlobalModel(nil, globalConfig)
	}
	if model.GlobalConfig == nil {
		model.GlobalConfig = globalConfig
	}

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		model.DetectSpinner, cmd = model.DetectSpinner.Update(msg)
		return currentPage, false, cmd, model

	case bulkDetectionResultMsg:
		model.Detecting = false
		if msg.err != nil {
			model.DetectStatus = "Error: " + msg.err.Error()
			return currentPage, false, nil, model
		}

		allDistributions := append(msg.homebrew, msg.npm...)
		if err := ImportDetectedDistributions(allDistributions, model.GlobalConfig); err != nil {
			model.DetectStatus = "Import failed: " + err.Error()
			return currentPage, false, nil, model
		}

		model.DetectStatus = fmt.Sprintf("Imported %d distributions", len(allDistributions))
		return currentPage, false, ReloadProjectsCmd(), model

	case tea.KeyMsg:
		if model.SettingWorkingDir {
			return handleWorkingDirKeys(msg, model, currentPage)
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return currentPage, true, tea.Quit, model
		case "esc":
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
			if len(model.Projects) > 0 && model.SelectedIndex < len(model.Projects) {
				selectedProject := model.Projects[model.SelectedIndex]

				if selectedProject.Project == nil || selectedProject.Project.Path == "" {
					model.SettingWorkingDir = true
					model.TargetProject = &selectedProject
					model.WorkingDirInput.Focus()
					model.WorkingDirInput.SetValue("")
					model.WorkingDirResults = []string{}
					model.WorkingDirSelected = 0
					return currentPage, false, nil, model
				}

				return 0, false, SwitchProjectCmd(&selectedProject), model
			}
			return currentPage, false, nil, model
		case "D":
			if !model.Detecting {
				model.Detecting = true
				model.DetectStatus = ""
				return currentPage, false, tea.Batch(
					model.DetectSpinner.Tick,
					BulkDetectDistributionsCmd(model.GlobalConfig),
				), model
			}
			return currentPage, false, nil, model
		}
	}
	return currentPage, false, nil, model
}

func handleWorkingDirKeys(msg tea.KeyMsg, model *GlobalModel, currentPage int) (int, bool, tea.Cmd, *GlobalModel) {
	switch msg.String() {
	case "esc":
		model.SettingWorkingDir = false
		model.TargetProject = nil
		model.WorkingDirInput.Blur()
		return currentPage, false, nil, model

	case "up":
		if model.WorkingDirSelected > 0 {
			model.WorkingDirSelected--
		}
		return currentPage, false, nil, model

	case "down":
		if model.WorkingDirSelected < len(model.WorkingDirResults)-1 {
			model.WorkingDirSelected++
		}
		return currentPage, false, nil, model

	case "enter":
		if len(model.WorkingDirResults) > 0 && model.WorkingDirSelected < len(model.WorkingDirResults) {
			selectedPath := model.WorkingDirResults[model.WorkingDirSelected]

			if model.TargetProject != nil && model.TargetProject.Project != nil {
				model.TargetProject.Project.Path = selectedPath
				if err := config.SaveProject(model.TargetProject); err == nil {
					projectCopy := *model.TargetProject
					model.SettingWorkingDir = false
					model.TargetProject = nil
					model.WorkingDirInput.Blur()
					return 0, false, tea.Batch(
						ReloadProjectsCmd(),
						SwitchProjectCmd(&projectCopy),
					), model
				}
			}
		}
		return currentPage, false, nil, model

	default:
		var cmd tea.Cmd
		model.WorkingDirInput, cmd = model.WorkingDirInput.Update(msg)

		query := model.WorkingDirInput.Value()
		if query != "" {
			model.WorkingDirResults = fsearch.FuzzySearchDirectories(query, 3)
			model.WorkingDirSelected = 0
		} else {
			model.WorkingDirResults = []string{}
		}

		return currentPage, false, cmd, model
	}
}