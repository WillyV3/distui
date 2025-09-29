package views

import (
	"fmt"
	"strings"

	"distui/handlers"
	"github.com/charmbracelet/lipgloss"
)

// Tab styles (from tabs example)
func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true)
	controlStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

// RenderConfigureContent returns the content for the project configuration view
func RenderConfigureContent(project string, configModel *handlers.ConfigureModel) string {
	if configModel == nil {
		return "Loading configuration..."
	}

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true)

	var content strings.Builder

	content.WriteString(headerStyle.Render(fmt.Sprintf("CONFIGURE PROJECT: %s", project)) + "\n")

	// Render tabs
	tabs := []string{"Distributions", "Build Settings", "Advanced", "Cleanup"}
	var renderedTabs []string

	for i, t := range tabs {
		style := inactiveTabStyle
		if i == configModel.ActiveTab {
			style = activeTabStyle
		}
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	content.WriteString(row + "\n")

	// Render the active list
	if configModel.Initialized {
		content.WriteString(configModel.Lists[configModel.ActiveTab].View())
	} else {
		content.WriteString("Initializing...")
	}

	// Controls
	if configModel.Width > 0 {
		divider := strings.Repeat("─", configModel.Width)
		content.WriteString("\n" + controlStyle.Render(divider))
	} else {
		content.WriteString("\n" + controlStyle.Render("──────────────────────────────────────────"))
	}

	// Different controls for cleanup tab
	if configModel.ActiveTab == 3 {
		content.WriteString("\n" + controlStyle.Render("[Space] Cycle Action  [s] Smart Commit  [r] Refresh"))
		content.WriteString("\n" + controlStyle.Render("[Tab] Next Tab  [ESC] Cancel  [↑/↓] Navigate"))
	} else {
		content.WriteString("\n" + controlStyle.Render("[Space] Toggle  [a] Check All  [Tab] Next Tab"))
		content.WriteString("\n" + controlStyle.Render("[s] Save  [ESC] Cancel  [↑/↓] Navigate"))
	}

	return content.String()
}

// Compatibility struct for backward compatibility
type ConfigureState struct {
	ActiveTab      int
	Distributions  map[string]bool
	HomebrewTap    string
	NPMScope       string
	TestCommand    string
	UseGoReleaser  bool
	SelectedRow    int
}

// Legacy render function - redirects to new implementation
func RenderConfigureContentLegacy(project string, state *ConfigureState) string {
	return RenderConfigureContent(project, nil)
}