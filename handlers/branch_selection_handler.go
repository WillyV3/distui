package handlers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"

	"distui/internal/models"
	"distui/internal/gitops"
)

type BranchSelectionModel struct {
	Branches      []models.BranchInfo
	SelectedIndex int
	Loading       bool
	LoadSpinner   spinner.Model
	Error         string
	Width         int
	Height        int
	ListHeight    int
}

type branchesLoadedMsg struct {
	branches []models.BranchInfo
	err      error
}

type pushResultMsg struct {
	success bool
	err     error
}

func NewBranchSelectionModel(width, height int) BranchSelectionModel {
	s := spinner.New()
	s.Spinner = spinner.Dot

	chrome := 4
	listHeight := height - chrome

	return BranchSelectionModel{
		Loading:     true,
		LoadSpinner: s,
		Width:       width,
		Height:      height,
		ListHeight:  listHeight,
	}
}

func (m BranchSelectionModel) Init() tea.Cmd {
	return tea.Batch(
		m.LoadSpinner.Tick,
		loadBranchesCmd,
	)
}

func (m BranchSelectionModel) Update(msg tea.Msg) (BranchSelectionModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.Loading {
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.SelectedIndex > 0 {
				m.SelectedIndex--
			}
		case "down", "j":
			if m.SelectedIndex < len(m.Branches)-1 {
				m.SelectedIndex++
			}
		case "enter":
			if len(m.Branches) > 0 && m.SelectedIndex < len(m.Branches) {
				branch := m.Branches[m.SelectedIndex].Name
				return m, pushToBranchCmd(branch)
			}
		}

	case branchesLoadedMsg:
		m.Loading = false
		if msg.err != nil {
			m.Error = msg.err.Error()
		} else {
			m.Branches = msg.branches
			if len(m.Branches) > 0 {
				m.SelectedIndex = 0
			}
		}
		return m, nil

	case pushResultMsg:
		if msg.err != nil {
			m.Error = msg.err.Error()
			return m, nil
		}
		return m, nil

	case spinner.TickMsg:
		if m.Loading {
			var cmd tea.Cmd
			m.LoadSpinner, cmd = m.LoadSpinner.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		chrome := 4
		m.ListHeight = m.Height - chrome
	}

	return m, nil
}

func loadBranchesCmd() tea.Msg {
	branches, err := gitops.ListBranches()
	return branchesLoadedMsg{branches: branches, err: err}
}

func pushToBranchCmd(branch string) tea.Cmd {
	return func() tea.Msg {
		err := gitops.PushToBranch(branch)
		return pushResultMsg{success: err == nil, err: err}
	}
}
