package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"tuitemplate/views"
	"tuitemplate/handlers"
)

// pageState tracks which page we're currently viewing
type pageState uint

const (
	homePage pageState = iota
	page1
	page2
	page3
)

// Global spinner styling
var (
	spinnerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true)
)

// MenuItem implements list.Item interface
type menuItem struct {
	title       string
	description string
	pageIndex   int
}

func (i menuItem) FilterValue() string { return i.title }
func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.description }

// Main application model
type appModel struct {
	currentPage pageState
	choice      int // For homepage menu selection
	width       int
	height      int
	quitting    bool
	spinner     spinner.Model
	startTime   time.Time
	menuList    list.Model
}

func initialAppModel() appModel {
	s := spinner.New()
	// Spinner options - uncomment one to try it:
	// s.Spinner = spinner.Line
	// s.Spinner = spinner.Dot
	// s.Spinner = spinner.MiniDot
	// s.Spinner = spinner.Jump
	// s.Spinner = spinner.Pulse
	// s.Spinner = spinner.Points
	// s.Spinner = spinner.Globe
	// s.Spinner = spinner.Moon
	// s.Spinner = spinner.Monkey
	s.Spinner = spinner.Meter
	// s.Spinner = spinner.Hamburger
	s.Style = spinnerStyle

	// Create menu items
	items := []list.Item{
		menuItem{title: "📄 Page One", description: "First sample page", pageIndex: 0},
		menuItem{title: "📄 Page Two", description: "Second sample page", pageIndex: 1},
		menuItem{title: "📄 Page Three", description: "Third sample page", pageIndex: 2},
	}

	// Create list with custom delegate
	delegate := list.NewDefaultDelegate()

	// Customize colors using your base scheme
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("117")).
		BorderLeftForeground(lipgloss.Color("117"))

	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("244"))

	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color("255"))

	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.
		Foreground(lipgloss.Color("244"))

	menuList := list.New(items, delegate, 50, 10)
	menuList.SetShowHelp(false)
	menuList.SetShowTitle(false)
	menuList.SetShowStatusBar(false)
	menuList.SetFilteringEnabled(false)

	return appModel{
		currentPage: homePage,
		choice:      0,
		width:       80,
		height:      24,
		spinner:     s,
		startTime:   time.Now(),
		menuList:    menuList,
	}
}

func (m appModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tea.SetWindowTitle("TUI Template App"),
	)
}

// Main update function - routes messages to appropriate page handlers
func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Global quit keys
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	}

	// Handle window resize globally
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
	}

	// Update spinner animation
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)

	// Route to appropriate page handler
	switch m.currentPage {
	case homePage:
		var pageCmd tea.Cmd
		m, pageCmd = m.updateHomePage(msg)
		return m, tea.Batch(cmd, pageCmd)
	case page1:
		newPage, quitting, pageCmd := handlers.UpdatePage1(int(m.currentPage), int(homePage), msg)
		m.currentPage = pageState(newPage)
		m.quitting = quitting
		return m, tea.Batch(cmd, pageCmd)
	case page2:
		newPage, quitting, pageCmd := handlers.UpdatePage2(int(m.currentPage), int(homePage), msg)
		m.currentPage = pageState(newPage)
		m.quitting = quitting
		return m, tea.Batch(cmd, pageCmd)
	case page3:
		newPage, quitting, pageCmd := handlers.UpdatePage3(int(m.currentPage), int(homePage), msg)
		m.currentPage = pageState(newPage)
		m.quitting = quitting
		return m, tea.Batch(cmd, pageCmd)
	default:
		return m, cmd
	}
}

// Homepage update handler
func (m appModel) updateHomePage(msg tea.Msg) (appModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "j", "down":
			if m.choice < 2 {
				m.choice++
			}
		case "k", "up":
			if m.choice > 0 {
				m.choice--
			}
		case "enter":
			// Navigate to selected page
			switch m.choice {
			case 0:
				m.currentPage = page1
			case 1:
				m.currentPage = page2
			case 2:
				m.currentPage = page3
			}
			return m, nil
		}
	}
	return m, nil
}



// Main view function - routes to appropriate page renderer
func (m appModel) View() string {
	if m.quitting {
		return "\n  Thanks for visiting TUI Template App! 👋\n\n"
	}

	switch m.currentPage {
	case homePage:
		return m.homePageView()
	case page1:
		return m.page1View()
	case page2:
		return m.page2View()
	case page3:
		return m.page3View()
	default:
		return "Unknown page"
	}
}

// Homepage view renderer
func (m appModel) homePageView() string {
	// Load ASCII title from file
	asciiContent, err := os.ReadFile("ascii-art-txt")
	var asciiTitle string
	if err == nil {
		asciiTitle = string(asciiContent)
	} else {
		// Fallback if file can't be read
		asciiTitle = "🖥️  TUI TEMPLATE APP  🖥️"
	}

	// Create responsive layout styles with outer border
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("117")).
		Padding(1).
		Width(m.width - 2).
		Height(m.height - 2).
		Align(lipgloss.Center, lipgloss.Center)

	// Static ASCII title - no animation
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Align(lipgloss.Center)


	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Italic(true).
		Align(lipgloss.Center).
		Margin(1, 0)

	// Build the animated subtitle with spinners on both sides
	animatedSubtitle := fmt.Sprintf("%s %s %s",
		m.spinner.View(),
		"A template for building terminal applications",
		m.spinner.View())

	// Simple menu items - all visible at once
	menuItems := []struct {
		title string
		desc  string
		index int
	}{
		{"📄 Page One", "First sample page", 0},
		{"📄 Page Two", "Second sample page", 1},
		{"📄 Page Three", "Third sample page", 2},
	}

	// Build simple menu display
	var menuContent strings.Builder
	menuContent.WriteString("Welcome to the TUI Template! Choose a page to explore:\n\n")

	for i, item := range menuItems {
		cursor := "   "
		itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

		if m.choice == i {
			cursor = "▶  "
			itemStyle = itemStyle.Foreground(lipgloss.Color("117")).Bold(true)
			descStyle = descStyle.Foreground(lipgloss.Color("117"))
		}

		menuContent.WriteString(itemStyle.Render(cursor + item.title) + "\n")
		menuContent.WriteString(descStyle.Render("    " + item.desc) + "\n\n")
	}

	menuContent.WriteString("↑/↓: navigate • enter: select • q: quit")

	// Create a styled container for the menu
	menuContainer := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("117")).
		Padding(2, 4).
		Margin(1, 0).
		Align(lipgloss.Center)

	// Combine everything with responsive layout
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render(asciiTitle),
		subtitleStyle.Render(animatedSubtitle),
		menuContainer.Render(menuContent.String()),
	)

	return containerStyle.Render(content)
}

// Standard page layout wrapper - locked pattern for all pages
func (m appModel) renderPage(title, content string) string {
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("117")).
		Padding(1).
		Width(m.width - 2).
		Height(m.height - 2).
		Align(lipgloss.Center, lipgloss.Top)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true).
		Align(lipgloss.Center).
		Margin(0, 0, 1, 0)

	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("117")).
		Padding(2, 4).
		Width(m.width - 8)

	pageContent := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render(title),
		contentStyle.Render(content),
	)

	return containerStyle.Render(pageContent)
}

// Page1 view renderer
func (m appModel) page1View() string {
	content := views.RenderPage1Content()
	return m.renderPage("📄 Page One", content)
}

// Page2 view renderer
func (m appModel) page2View() string {
	content := views.RenderPage2Content()
	return m.renderPage("📄 Page Two", content)
}

// Page3 view renderer
func (m appModel) page3View() string {
	content := views.RenderPage3Content()
	return m.renderPage("📄 Page Three", content)
}