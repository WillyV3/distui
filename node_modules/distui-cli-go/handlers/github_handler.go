package handlers

import (
	"distui/internal/gitcleanup"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type githubState uint

const (
	githubOverview githubState = iota
	githubCreate
)

type GitHubModel struct {
	State       githubState
	RepoName    textinput.Model
	RepoDesc    textinput.Model
	IsPrivate   bool
	FocusIndex  int
	Width       int
	Height      int
	RepoInfo    *gitcleanup.RepoInfo
	RepoBrowser *RepoBrowserModel
}

func NewGitHubModel(width, height int) *GitHubModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "Repository name"
	nameInput.CharLimit = 100
	nameInput.Width = width - 4
	nameInput.Focus()

	descInput := textinput.New()
	descInput.Placeholder = "Description (optional)"
	descInput.CharLimit = 200
	descInput.Width = width - 4

	m := &GitHubModel{
		State:       githubOverview,
		RepoName:    nameInput,
		RepoDesc:    descInput,
		IsPrivate:   false,
		FocusIndex:  0,
		Width:       width,
		Height:      height,
		RepoBrowser: NewRepoBrowserModel(width, height),
	}

	info, _ := gitcleanup.CheckRepoState()
	m.RepoInfo = info

	return m
}

func (m *GitHubModel) Update(msg tea.Msg) (*GitHubModel, tea.Cmd) {
	var cmd tea.Cmd

	// Always delegate to repo browser if in overview state
	if m.State == githubOverview && m.RepoBrowser != nil {
		m.RepoBrowser, cmd = m.RepoBrowser.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if m.State == githubCreate {
				m.FocusIndex = (m.FocusIndex + 1) % 3
				if m.FocusIndex == 0 {
					m.RepoName.Focus()
					m.RepoDesc.Blur()
				} else if m.FocusIndex == 1 {
					m.RepoName.Blur()
					m.RepoDesc.Focus()
				} else {
					m.RepoDesc.Blur()
				}
			}
		case " ":
			if m.FocusIndex == 2 {
				m.IsPrivate = !m.IsPrivate
			}
		case "enter":
			if m.State == githubCreate {
				// Execute creation (handled by parent)
				return m, nil
			}
		}

		if m.FocusIndex == 0 {
			m.RepoName, cmd = m.RepoName.Update(msg)
		} else if m.FocusIndex == 1 {
			m.RepoDesc, cmd = m.RepoDesc.Update(msg)
		}
	}

	return m, cmd
}

func (m *GitHubModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
	m.RepoName.Width = width - 4
	m.RepoDesc.Width = width - 4
	if m.RepoBrowser != nil {
		m.RepoBrowser.SetSize(width, height)
	}
}