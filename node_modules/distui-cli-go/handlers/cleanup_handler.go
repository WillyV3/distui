package handlers

import (
	"distui/internal/gitcleanup"
	tea "github.com/charmbracelet/bubbletea"
)

type CleanupModel struct {
	RepoInfo         *gitcleanup.RepoInfo
	FileChanges      []gitcleanup.FileChange
	StatusText       string
	Width            int
	Height           int
	RepoBrowser      *RepoBrowserModel
	GitHubRepoExists bool
}

func NewCleanupModel(width, height int) *CleanupModel {
	m := &CleanupModel{
		Width:       width,
		Height:      height,
		RepoBrowser: NewRepoBrowserModel(width, height),
	}
	m.loadRepoStatus()
	return m
}

func (m *CleanupModel) loadRepoStatus() error {
	info, err := gitcleanup.CheckRepoState()
	if err != nil {
		m.StatusText = "Error checking repository status"
		return err
	}
	m.RepoInfo = info

	changes, err := gitcleanup.GetFileChanges()
	if err == nil {
		m.FileChanges = changes
	}

	m.StatusText = gitcleanup.GetRepoStatus()

	// Cache GitHub repo existence check
	if m.RepoInfo != nil && m.RepoInfo.RemoteExists {
		m.GitHubRepoExists = gitcleanup.CheckGitHubRepoExists()
	} else {
		m.GitHubRepoExists = false
	}

	return nil
}

func (m *CleanupModel) GetFileSummary() (modified, added, deleted, untracked int) {
	if m.RepoInfo == nil {
		return 0, 0, 0, 0
	}
	return m.RepoInfo.FileStats.Modified,
		m.RepoInfo.FileStats.Added,
		m.RepoInfo.FileStats.Deleted,
		m.RepoInfo.FileStats.Untracked
}

func (m *CleanupModel) NeedsGitHub() bool {
	if m.RepoInfo == nil {
		return false
	}
	return m.RepoInfo.Status == gitcleanup.RepoStatusNoRemote
}

func (m *CleanupModel) HasChanges() bool {
	if m.RepoInfo == nil {
		return false
	}
	return m.RepoInfo.Status == gitcleanup.RepoStatusDirty
}

func (m *CleanupModel) IsClean() bool {
	if m.RepoInfo == nil {
		return false
	}
	return m.RepoInfo.Status == gitcleanup.RepoStatusClean
}

func (m *CleanupModel) Refresh() {
	m.loadRepoStatus()
	// Also refresh the repo browser
	if m.RepoBrowser != nil {
		m.RepoBrowser.LoadDirectory()
	}
}

func (m *CleanupModel) Update(width, height int) {
	m.Width = width
	m.Height = height
	if m.RepoBrowser != nil {
		m.RepoBrowser.SetSize(width, height)
	}
}

func (m *CleanupModel) HandleKey(msg tea.KeyMsg) (*CleanupModel, tea.Cmd) {
	// If there are no changes, delegate navigation to the repo browser
	if !m.HasChanges() && m.RepoBrowser != nil {
		var cmd tea.Cmd
		m.RepoBrowser, cmd = m.RepoBrowser.Update(msg)
		return m, cmd
	}
	return m, nil
}