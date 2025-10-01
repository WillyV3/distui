package handlers

import (
	tea "github.com/charmbracelet/bubbletea"

	"distui/internal/gitcleanup"
)

type SelectableFile struct {
	Path     string
	Status   string
	Category gitcleanup.FileCategory
	Selected bool
	IsAuto   bool // Auto-commit file (can't be deselected)
}

type FileSelectionModel struct {
	Files         []SelectableFile
	SelectedIndex int
	Width         int
	Height        int
	CustomRules   bool
}

func NewFileSelectionModel(changes []gitcleanup.FileChange, customRules bool, projectConfig interface{}) *FileSelectionModel {
	files := []SelectableFile{}

	for _, change := range changes {
		category := gitcleanup.CategorizeFile(change.Path)

		isAuto := category == gitcleanup.CategoryAuto
		shouldSelect := false

		if customRules {
			// Custom rules: select everything except ignore
			shouldSelect = category != gitcleanup.CategoryIgnore
		} else {
			// Default: only auto files are selected
			shouldSelect = isAuto
		}

		files = append(files, SelectableFile{
			Path:     change.Path,
			Status:   change.Status,
			Category: category,
			Selected: shouldSelect,
			IsAuto:   isAuto,
		})
	}

	return &FileSelectionModel{
		Files:         files,
		SelectedIndex: 0,
		CustomRules:   customRules,
	}
}

func (m *FileSelectionModel) Update(msg tea.Msg) (*FileSelectionModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.SelectedIndex > 0 {
				m.SelectedIndex--
			}
		case "down", "j":
			if m.SelectedIndex < len(m.Files)-1 {
				m.SelectedIndex++
			}
		case " ", "space":
			// Toggle selection (only for non-auto files in default mode)
			if m.SelectedIndex < len(m.Files) {
				file := &m.Files[m.SelectedIndex]
				// In default mode, can only toggle non-auto files
				// In custom mode, can toggle anything
				if m.CustomRules || !file.IsAuto {
					file.Selected = !file.Selected
				}
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	return m, nil
}

func (m *FileSelectionModel) GetSelectedFiles() []gitcleanup.CleanupItem {
	items := []gitcleanup.CleanupItem{}

	for _, file := range m.Files {
		if file.Selected {
			items = append(items, gitcleanup.CleanupItem{
				Path:     file.Path,
				Status:   file.Status,
				Category: string(file.Category),
				Action:   "commit",
			})
		}
	}

	return items
}

func (m *FileSelectionModel) HasSelectedFiles() bool {
	for _, file := range m.Files {
		if file.Selected {
			return true
		}
	}
	return false
}
