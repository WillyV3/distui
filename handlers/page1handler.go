

package handlers

import tea "github.com/charmbracelet/bubbletea"

// UpdatePage1 handles page1 updates
func UpdatePage1(currentPage, homePage int, msg tea.Msg) (int, bool, tea.Cmd) {
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

