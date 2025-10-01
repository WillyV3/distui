package views

import (
	"github.com/charmbracelet/lipgloss"

	"distui/internal/models"
)

var (
	notifInfoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Background(lipgloss.Color("0")).Padding(0, 1)
	notifSuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("2")).Padding(0, 1).Bold(true)
	notifWarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("3")).Padding(0, 1)
	notifErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("1")).Padding(0, 1).Bold(true)
)

func RenderNotification(notification *models.UINotification) string {
	if notification == nil {
		return ""
	}

	message := notification.Message
	if len(message) > 60 {
		message = message[:57] + "..."
	}

	var styledMessage string
	switch notification.Style {
	case "success":
		styledMessage = notifSuccessStyle.Render("✓ " + message)
	case "warning":
		styledMessage = notifWarningStyle.Render("⚠ " + message)
	case "error":
		styledMessage = notifErrorStyle.Render("✗ " + message)
	default:
		styledMessage = notifInfoStyle.Render("ℹ " + message)
	}

	return "\n" + styledMessage + "\n"
}
