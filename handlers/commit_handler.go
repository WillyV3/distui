package handlers

import (
	"distui/internal/gitcleanup"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type CommitModel struct {
	FileChanges   []gitcleanup.FileChange
	CurrentIndex  int
	CommitMessage textinput.Model
	Width         int
	Height        int
	Decisions     map[string]string // filepath -> "stage", "skip", or "ignore"
}

func NewCommitModel(width, height int) *CommitModel {
	msgInput := textinput.New()
	msgInput.Placeholder = "Enter commit message..."
	msgInput.Focus()
	msgInput.CharLimit = 200
	msgInput.Width = width - 4

	m := &CommitModel{
		Width:         width,
		Height:        height,
		CommitMessage: msgInput,
		CurrentIndex:  0,
		Decisions:     make(map[string]string),
	}

	// Load file changes
	changes, _ := gitcleanup.GetFileChanges()
	m.FileChanges = changes

	// Initialize all files as "skip" by default
	for _, change := range changes {
		m.Decisions[change.Path] = "skip"
	}

	return m
}

func (m *CommitModel) Update(msg tea.Msg) (*CommitModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If we're past all files, we're on the commit message input
		if m.CurrentIndex >= len(m.FileChanges) {
			var cmd tea.Cmd
			m.CommitMessage, cmd = m.CommitMessage.Update(msg)
			return m, cmd
		}

		// Otherwise handle file decisions
		switch msg.String() {
		case "a": // Add/stage this file
			if m.CurrentIndex < len(m.FileChanges) {
				file := m.FileChanges[m.CurrentIndex]
				m.Decisions[file.Path] = "stage"
				m.CurrentIndex++
			}
		case "s": // Skip this file
			if m.CurrentIndex < len(m.FileChanges) {
				file := m.FileChanges[m.CurrentIndex]
				m.Decisions[file.Path] = "skip"
				m.CurrentIndex++
			}
		case "i": // Add to .gitignore
			if m.CurrentIndex < len(m.FileChanges) {
				file := m.FileChanges[m.CurrentIndex]
				m.Decisions[file.Path] = "ignore"
				m.CurrentIndex++
			}
		case "p": // Go to previous file
			if m.CurrentIndex > 0 {
				m.CurrentIndex--
			}
		}
	}
	return m, nil
}

func (m *CommitModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
	m.CommitMessage.Width = width - 4
}

func (m *CommitModel) GetStagedFiles() []string {
	var staged []string
	for path, decision := range m.Decisions {
		if decision == "stage" {
			staged = append(staged, path)
		}
	}
	return staged
}

func (m *CommitModel) HasStagedFiles() bool {
	for _, decision := range m.Decisions {
		if decision == "stage" {
			return true
		}
	}
	return false
}

func (m *CommitModel) IsComplete() bool {
	return m.CurrentIndex >= len(m.FileChanges)
}