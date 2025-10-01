package handlers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"

	"distui/internal/models"
	"distui/internal/gitops"
)

type BranchSelectionModel struct {
	Branches       []models.BranchInfo
	SelectedIndex  int
	Loading        bool
	LoadSpinner    spinner.Model
	Error          string
	Width          int
	Height         int
	ListHeight     int
	CurrentBranch  string
	Pushing        bool
	PushStatus     string
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
		loadBranchesCmd(),
	)
}

func (m BranchSelectionModel) Update(msg tea.Msg) (BranchSelectionModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Allow Esc even during loading
		if msg.String() == "esc" {
			return m, nil
		}

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
				selectedBranch := m.Branches[m.SelectedIndex]

				m.Pushing = true
				m.LoadSpinner = spinner.New()
				m.LoadSpinner.Spinner = spinner.Dot

				// Determine action based on selected branch
				if selectedBranch.IsCurrent {
					// Pushing to current branch origin
					m.PushStatus = "Pushing to origin/" + selectedBranch.Name + "..."
					return m, tea.Batch(m.LoadSpinner.Tick, pushCurrentBranchCmd())
				} else if selectedBranch.Name == "main" || selectedBranch.Name == "master" {
					// Creating PR to main/master
					m.PushStatus = "Creating PR to " + selectedBranch.Name + "..."
					return m, tea.Batch(m.LoadSpinner.Tick, createPRCmd(selectedBranch.Name))
				} else {
					// Pushing to different branch
					m.PushStatus = "Pushing to " + selectedBranch.Name + "..."
					return m, tea.Batch(m.LoadSpinner.Tick, pushToBranchCmd(selectedBranch.Name))
				}
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
		m.Pushing = false
		if msg.err != nil {
			m.Error = msg.err.Error()
			return m, nil
		}
		// Success - modal will be closed by configure handler
		return m, nil

	case spinner.TickMsg:
		if m.Loading || m.Pushing {
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

func loadBranchesCmd() tea.Cmd {
	return func() tea.Msg {
		branches, err := gitops.ListBranches()
		return branchesLoadedMsg{branches: branches, err: err}
	}
}

func pushToBranchCmd(branch string) tea.Cmd {
	return func() tea.Msg {
		err := gitops.PushToBranch(branch)
		return pushResultMsg{success: err == nil, err: err}
	}
}

func pushCurrentBranchCmd() tea.Cmd {
	return func() tea.Msg {
		err := gitops.PushCurrentBranch()
		return pushResultMsg{success: err == nil, err: err}
	}
}

func createPRCmd(targetBranch string) tea.Cmd {
	return func() tea.Msg {
		err := gitops.CreatePullRequest(targetBranch)
		return pushResultMsg{success: err == nil, err: err}
	}
}
