package handlers

import (
	tea "github.com/charmbracelet/bubbletea"

	"distui/internal/executor"
)

type npmNameCheckMsg struct {
	result executor.NPMNameCheckResult
}

func checkNPMNameCmd(packageName, username string) tea.Cmd {
	return func() tea.Msg {
		result := executor.CheckNPMName(packageName, username)
		return npmNameCheckMsg{result: result}
	}
}