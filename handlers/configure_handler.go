package handlers

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"distui/internal/gitcleanup"
)

// ConfigureModel holds the state for the configure view
type ConfigureModel struct {
	ActiveTab     int
	Lists         [4]list.Model
	Width         int
	Height        int
	Initialized   bool
}

// Distribution item for the list
type DistributionItem struct {
	Name    string
	Desc    string
	Enabled bool
	Key     string
}

func (i DistributionItem) Title() string       {
	checkbox := "[ ]"
	if i.Enabled {
		checkbox = "[✓]"
	}
	return checkbox + " " + i.Name
}
func (i DistributionItem) Description() string { return i.Desc }
func (i DistributionItem) FilterValue() string { return i.Name }

// Build setting item
type BuildItem struct {
	Name    string
	Value   string
	Enabled bool
}

func (i BuildItem) Title() string {
	if i.Enabled {
		return "[✓] " + i.Name
	}
	return "[ ] " + i.Name
}
func (i BuildItem) Description() string { return i.Value }
func (i BuildItem) FilterValue() string { return i.Name }

// Cleanup item for git file management
type CleanupItem struct {
	Path     string
	Status   string // M=modified, A=added, D=deleted, ??=untracked
	Category string // auto, docs, ignore, other
	Action   string // commit, skip, ignore
}

func (i CleanupItem) Title() string {
	statusSymbol := "?"
	switch i.Status {
	case "M":
		statusSymbol = "M"
	case "A":
		statusSymbol = "+"
	case "D":
		statusSymbol = "-"
	case "??":
		statusSymbol = "?"
	}
	return fmt.Sprintf("[%s] %s", statusSymbol, i.Path)
}

func (i CleanupItem) Description() string {
	actionText := ""
	switch i.Action {
	case "commit":
		actionText = "→ Will commit"
	case "skip":
		actionText = "→ Skip"
	case "ignore":
		actionText = "→ Add to .gitignore"
	default:
		actionText = fmt.Sprintf("→ %s file", i.Category)
	}
	return actionText
}

func (i CleanupItem) FilterValue() string { return i.Path }

// Initialize the configure model
func NewConfigureModel(width, height int) *ConfigureModel {
	// Use provided dimensions or defaults
	if width <= 0 {
		width = 100
	}
	if height <= 0 {
		height = 30
	}

	m := &ConfigureModel{
		ActiveTab:   0,
		Width:       width,
		Height:      height,
		Initialized: true,
	}

	// Initialize distributions list
	distributions := []list.Item{
		DistributionItem{
			Name:    "GitHub Releases",
			Desc:    "Create releases on GitHub",
			Enabled: true,
			Key:     "github",
		},
		DistributionItem{
			Name:    "Homebrew",
			Desc:    "Tap: willyv3/homebrew-tap",
			Enabled: true,
			Key:     "homebrew",
		},
		DistributionItem{
			Name:    "NPM Package",
			Desc:    "Scope: @williavs",
			Enabled: false,
			Key:     "npm",
		},
		DistributionItem{
			Name:    "Go Install",
			Desc:    "Enable 'go install' support",
			Enabled: true,
			Key:     "go_install",
		},
	}

	// Calculate initial list size
	// Account for: header (1), tabs (2), controls (4) = 7 lines
	listHeight := m.Height - 7
	if listHeight < 5 {
		listHeight = 5
	}

	// Create distributions list
	distList := list.New(distributions, list.NewDefaultDelegate(), m.Width, listHeight)
	distList.Title = "Distributions"
	distList.SetShowStatusBar(false)
	distList.SetFilteringEnabled(false)
	distList.SetShowHelp(false)
	distList.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true)
	m.Lists[0] = distList

	// Initialize build settings list
	buildItems := []list.Item{
		BuildItem{Name: "Run tests before release", Value: "go test ./...", Enabled: true},
		BuildItem{Name: "Clean build directory", Value: "", Enabled: true},
		BuildItem{Name: "Build for all platforms", Value: "darwin, linux, windows", Enabled: false},
		BuildItem{Name: "Include ARM64 builds", Value: "", Enabled: false},
	}

	buildList := list.New(buildItems, list.NewDefaultDelegate(), m.Width, listHeight)
	buildList.Title = "Build Settings"
	buildList.SetShowStatusBar(false)
	buildList.SetFilteringEnabled(false)
	buildList.SetShowHelp(false)
	m.Lists[1] = buildList

	// Initialize advanced list
	advancedItems := []list.Item{
		BuildItem{Name: "Create draft releases", Value: "", Enabled: false},
		BuildItem{Name: "Mark as pre-release", Value: "", Enabled: false},
		BuildItem{Name: "Generate changelog", Value: "", Enabled: true},
		BuildItem{Name: "Sign commits", Value: "", Enabled: true},
	}

	advList := list.New(advancedItems, list.NewDefaultDelegate(), m.Width, listHeight)
	advList.Title = "Advanced"
	advList.SetShowStatusBar(false)
	advList.SetFilteringEnabled(false)
	advList.SetShowHelp(false)
	m.Lists[2] = advList

	// Initialize cleanup list with real git status
	cleanupItems := m.loadGitStatus()
	cleanupList := list.New(cleanupItems, list.NewDefaultDelegate(), m.Width, listHeight)
	cleanupList.Title = "Smart Git Cleanup"
	cleanupList.SetShowStatusBar(false)
	cleanupList.SetFilteringEnabled(false)
	cleanupList.SetShowHelp(false)
	m.Lists[3] = cleanupList

	return m
}

// loadGitStatus loads current git status and categorizes files
func (m *ConfigureModel) loadGitStatus() []list.Item {
	items := []list.Item{}

	gitFiles, err := gitcleanup.GetGitStatus()
	if err != nil {
		// Return empty list if not in git repo
		items = append(items, CleanupItem{
			Path:     "Not in a git repository",
			Status:   "??",
			Category: "other",
			Action:   "skip",
		})
		return items
	}

	if len(gitFiles) == 0 {
		items = append(items, CleanupItem{
			Path:     "Working directory is clean",
			Status:   "✓",
			Category: "other",
			Action:   "skip",
		})
		return items
	}

	// Convert git files to cleanup items with smart defaults
	for _, gf := range gitFiles {
		action := "skip"

		// Set default action based on category
		switch gf.Category {
		case gitcleanup.CategoryAuto:
			action = "commit"
		case gitcleanup.CategoryIgnore:
			action = "ignore"
		case gitcleanup.CategoryDocs:
			action = "skip" // Ask user
		default:
			action = "skip"
		}

		items = append(items, CleanupItem{
			Path:     gf.Path,
			Status:   gf.Status,
			Category: string(gf.Category),
			Action:   action,
		})
	}

	return items
}

// Update the configure model
func (m *ConfigureModel) Update(msg tea.Msg) (*ConfigureModel, tea.Cmd) {
	// Update list sizes based on current dimensions
	if m.Width > 0 && m.Height > 0 {
		// Account for: header (1), tabs (2), controls (4) = 7 lines
		listHeight := m.Height - 7
		if listHeight < 5 {
			listHeight = 5
		}
		for i := range m.Lists {
			m.Lists[i].SetWidth(m.Width)
			m.Lists[i].SetHeight(listHeight)
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		// Update list sizes
		listHeight := msg.Height - 7
		if listHeight < 5 {
			listHeight = 5
		}
		for i := range m.Lists {
			m.Lists[i].SetWidth(msg.Width)
			m.Lists[i].SetHeight(listHeight)
		}
		m.Initialized = true

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.ActiveTab = (m.ActiveTab + 1) % 4
			return m, nil
		case "shift+tab":
			m.ActiveTab = (m.ActiveTab + 3) % 4
			return m, nil
		case "space":
			// Toggle the selected item
			currentList := &m.Lists[m.ActiveTab]
			if i, ok := currentList.SelectedItem().(DistributionItem); ok {
				i.Enabled = !i.Enabled
				items := currentList.Items()
				items[currentList.Index()] = i
				currentList.SetItems(items)
			} else if i, ok := currentList.SelectedItem().(BuildItem); ok {
				i.Enabled = !i.Enabled
				items := currentList.Items()
				items[currentList.Index()] = i
				currentList.SetItems(items)
			} else if i, ok := currentList.SelectedItem().(CleanupItem); ok {
				// Cycle through actions: commit -> skip -> ignore -> commit
				switch i.Action {
				case "commit":
					i.Action = "skip"
				case "skip":
					i.Action = "ignore"
				case "ignore":
					i.Action = "commit"
				default:
					i.Action = "commit"
				}
				items := currentList.Items()
				items[currentList.Index()] = i
				currentList.SetItems(items)
			}
			return m, nil
		case "a":
			// Check/uncheck all in current tab
			if m.ActiveTab == 0 {
				items := m.Lists[0].Items()
				allChecked := true
				for _, item := range items {
					if dist, ok := item.(DistributionItem); ok && !dist.Enabled {
						allChecked = false
						break
					}
				}
				// Toggle all
				for i, item := range items {
					if dist, ok := item.(DistributionItem); ok {
						dist.Enabled = !allChecked
						items[i] = dist
					}
				}
				m.Lists[0].SetItems(items)
			}
			return m, nil
		default:
			// Pass through to the active list
			var cmd tea.Cmd
			m.Lists[m.ActiveTab], cmd = m.Lists[m.ActiveTab].Update(msg)
			return m, cmd
		}
	}

	// Update the active list
	var cmd tea.Cmd
	m.Lists[m.ActiveTab], cmd = m.Lists[m.ActiveTab].Update(msg)
	return m, cmd
}

// UpdateConfigureView handles configure view updates and navigation
func UpdateConfigureView(currentPage, previousPage int, msg tea.Msg, configModel *ConfigureModel) (int, bool, tea.Cmd, *ConfigureModel) {
	// Model will be created in app.go with proper dimensions

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return currentPage, true, tea.Quit, configModel
		case "esc":
			return 0, false, nil, configModel // back to projectView
		case "r":
			// Refresh git status in cleanup tab
			if configModel != nil && configModel.ActiveTab == 3 {
				configModel.Lists[3].SetItems(configModel.loadGitStatus())
			}
			return currentPage, false, nil, configModel
		case "s":
			// Save configuration or execute smart commit
			if configModel != nil && configModel.ActiveTab == 3 {
				// Execute smart commit for cleanup tab
				items := []gitcleanup.CleanupItem{}
				for _, listItem := range configModel.Lists[3].Items() {
					if ci, ok := listItem.(CleanupItem); ok {
						items = append(items, gitcleanup.CleanupItem{
							Path:     ci.Path,
							Status:   ci.Status,
							Category: ci.Category,
							Action:   ci.Action,
						})
					}
				}

				if _, err := gitcleanup.ExecuteSmartCommit(items); err == nil {
					// Refresh the cleanup list after commit
					configModel.Lists[3].SetItems(configModel.loadGitStatus())
					// TODO: Show success message
				}
			} else {
				// TODO: Implement save logic for other tabs
			}
			return 0, false, nil, configModel // back to projectView after save
		}
	}

	// Update the model
	newModel, cmd := configModel.Update(msg)
	return currentPage, false, cmd, newModel
}