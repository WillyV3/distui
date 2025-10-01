package handlers

import tea "github.com/charmbracelet/bubbletea"

// UpdateNewProjectView handles new project view updates and navigation
func UpdateNewProjectView(currentPage, previousPage int, msg tea.Msg) (int, bool, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return currentPage, true, tea.Quit
		case "tab":
			return 0, false, nil // projectView
		case "p":
			return 0, false, nil // projectView
		case "g":
			return 1, false, nil // globalView
		case "s":
			return 2, false, nil // settingsView
		case "r":
			return 3, false, nil // releaseView
		case "c":
			return 4, false, nil // configureView
		case "esc":
			return 1, false, nil // globalView
		}
	}
	return currentPage, false, nil
}