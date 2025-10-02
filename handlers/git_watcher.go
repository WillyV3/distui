package handlers

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// gitWatchTickMsg is sent periodically to refresh git status in background
type gitWatchTickMsg struct{}

// StartGitWatcherCmd starts background git status polling
func StartGitWatcherCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return gitWatchTickMsg{}
	})
}

// HandleGitWatchTick handles background git refresh and schedules next tick
func (m *ConfigureModel) HandleGitWatchTick() (*ConfigureModel, tea.Cmd) {
	// Only refresh if CleanupModel exists (tab has been visited at least once)
	if m.CleanupModel != nil && m.Width > 0 && m.Height > 0 {
		// Refresh git status (lightweight - just reads git state, no network calls)
		m.CleanupModel.Refresh()

		// Update git status items in the list
		m.Lists[0].SetItems(m.loadGitStatus())

		// Update dimensions in case terminal was resized
		// Total: 4 (app wrapper) + 11 (view chrome) = 15
		chromeLines := 15
		if m.NeedsRegeneration {
			chromeLines = 16
		}
		listHeight := m.Height - chromeLines
		if listHeight < 5 {
			listHeight = 5
		}
		listWidth := m.Width - 2
		if listWidth < 40 {
			listWidth = 40
		}
		m.CleanupModel.Update(listWidth, listHeight)
	}

	// Schedule next tick (continues polling every 2 seconds)
	return m, StartGitWatcherCmd()
}
