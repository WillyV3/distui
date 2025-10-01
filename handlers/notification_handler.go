package handlers

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"distui/internal/models"
)

type NotificationModel struct {
	Notification *models.UINotification
	Ticking      bool
}

type tickMsg time.Time

func ShowNotification(message string, style string) (*models.UINotification, tea.Cmd) {
	notification := &models.UINotification{
		Message:   message,
		ShowUntil: time.Now().Add(1500 * time.Millisecond),
		Style:     style,
	}

	return notification, tickCmd()
}

func (m NotificationModel) Update(msg tea.Msg) (NotificationModel, tea.Cmd) {
	switch msg.(type) {
	case tickMsg:
		if m.Notification == nil {
			m.Ticking = false
			return m, nil
		}

		if time.Now().After(m.Notification.ShowUntil) {
			m.Notification = nil
			m.Ticking = false
			return m, nil
		}

		return m, tickCmd()
	}

	return m, nil
}

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func DismissNotification(model *NotificationModel) {
	model.Notification = nil
	model.Ticking = false
}

func CreateNotification(message string, style string) (*models.UINotification, tea.Cmd) {
	return ShowNotification(message, style)
}
