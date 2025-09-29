package handlers

import tea "github.com/charmbracelet/bubbletea"

// UpdateReleaseView handles release view updates and navigation
func UpdateReleaseView(currentPage, previousPage int, msg tea.Msg) (int, bool, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return currentPage, true, tea.Quit
		case "tab":
			return 4, false, nil // configureView
		case "p":
			return 0, false, nil // projectView
		case "g":
			return 1, false, nil // globalView
		case "s":
			return 2, false, nil // settingsView
		case "c":
			return 4, false, nil // configureView
		case "n":
			return 5, false, nil // newProjectView
		case "esc":
			return 0, false, nil // projectView
		}
	}
	return currentPage, false, nil
}