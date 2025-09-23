package handlers

import tea "github.com/charmbracelet/bubbletea"

// UpdatePage3 handles page3 updates
func UpdatePage3(currentPage, homePage int, msg tea.Msg) (int, bool, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return homePage, false, nil
		case "q":
			return currentPage, true, tea.Quit
		}
	}
	return currentPage, false, nil
}