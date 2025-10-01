package handlers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"

	"distui/internal/models"
	"distui/internal/filescanner"
)

type RepoCleanupModel struct {
	ScanResult    *models.CleanupScanResult
	FlaggedFiles  []models.FlaggedFile
	SelectedIndex int
	Scanning      bool
	ScanSpinner   spinner.Model
	Width         int
	Height        int
}

type scanCompleteMsg struct {
	result *models.CleanupScanResult
	err    error
}

type fileActionResultMsg struct {
	success bool
	err     error
}

func NewRepoCleanupModel(width, height int) RepoCleanupModel {
	s := spinner.New()
	s.Spinner = spinner.Dot

	return RepoCleanupModel{
		Scanning:    true,
		ScanSpinner: s,
		Width:       width,
		Height:      height,
	}
}

func (m RepoCleanupModel) Init() tea.Cmd {
	return tea.Batch(
		m.ScanSpinner.Tick,
		scanRepositoryCmd,
	)
}

func (m RepoCleanupModel) Update(msg tea.Msg) (RepoCleanupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.Scanning {
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.SelectedIndex > 0 {
				m.SelectedIndex--
			}
		case "down", "j":
			if m.SelectedIndex < len(m.FlaggedFiles)-1 {
				m.SelectedIndex++
			}
		case "d":
			if len(m.FlaggedFiles) > 0 && m.SelectedIndex < len(m.FlaggedFiles) {
				return m, deleteFileCmd(m.FlaggedFiles[m.SelectedIndex].Path)
			}
		case "i":
			if len(m.FlaggedFiles) > 0 && m.SelectedIndex < len(m.FlaggedFiles) {
				return m, addToGitignoreCmd(m.FlaggedFiles[m.SelectedIndex].Path)
			}
		case "a":
			if len(m.FlaggedFiles) > 0 && m.SelectedIndex < len(m.FlaggedFiles) {
				return m, archiveFileCmd(m.FlaggedFiles[m.SelectedIndex].Path)
			}
		case "r":
			m.Scanning = true
			return m, tea.Batch(m.ScanSpinner.Tick, scanRepositoryCmd)
		}

	case scanCompleteMsg:
		m.Scanning = false
		if msg.err == nil {
			m.ScanResult = msg.result
			m.FlaggedFiles = flattenFlaggedFiles(msg.result)
			if len(m.FlaggedFiles) > 0 {
				m.SelectedIndex = 0
			}
		}
		return m, nil

	case fileActionResultMsg:
		if msg.success {
			m.Scanning = true
			return m, tea.Batch(m.ScanSpinner.Tick, scanRepositoryCmd)
		}
		return m, nil

	case spinner.TickMsg:
		if m.Scanning {
			var cmd tea.Cmd
			m.ScanSpinner, cmd = m.ScanSpinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func scanRepositoryCmd() tea.Msg {
	result, err := filescanner.ScanRepository(".")
	return scanCompleteMsg{result: result, err: err}
}

func deleteFileCmd(path string) tea.Cmd {
	return func() tea.Msg {
		err := filescanner.DeleteFile(path)
		return fileActionResultMsg{success: err == nil, err: err}
	}
}

func archiveFileCmd(path string) tea.Cmd {
	return func() tea.Msg {
		err := filescanner.ArchiveFile(path)
		return fileActionResultMsg{success: err == nil, err: err}
	}
}

func addToGitignoreCmd(path string) tea.Cmd {
	return func() tea.Msg {
		err := filescanner.AddToGitignore(path)
		return fileActionResultMsg{success: err == nil, err: err}
	}
}

func flattenFlaggedFiles(result *models.CleanupScanResult) []models.FlaggedFile {
	var files []models.FlaggedFile
	files = append(files, result.MediaFiles...)
	files = append(files, result.ExcessDocs...)
	files = append(files, result.DevArtifacts...)
	return files
}
