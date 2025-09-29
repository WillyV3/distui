package handlers

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"distui/internal/gitcleanup"
)

// Message types for async operations
type repoCreatedMsg struct {
	err error
}

// ConfigureModel holds the state for the configure view
type ConfigureModel struct {
	ActiveTab       int
	Lists           [4]list.Model
	Width           int
	Height          int
	Initialized     bool
	CreatingRepo    bool
	RepoNameInput   textinput.Model
	RepoDescInput   textinput.Model
	RepoInputFocus  int  // 0=name, 1=description, 2=private toggle
	RepoIsPrivate   bool // true=private, false=public
	// Spinner for repo creation
	IsCreating      bool
	CreateSpinner   spinner.Model
	CreateStatus    string
	// Cached git status to avoid expensive calls on every render
	GitHubRepoExists bool
	GitHubOwner      string
	GitHubRepo       string
	HasGitRemote     bool
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

	// Special handling for GitHub repo
	if i.Category == "github-new" {
		if i.Action == "create" {
			return "→ Will create GitHub repo"
		}
		return "→ Skip"
	}
	if i.Category == "github-push" {
		if i.Action == "create" {
			return "→ Will push to GitHub"
		}
		return "→ Skip"
	}

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

	// Initialize repo creation inputs
	nameInput := textinput.New()
	nameInput.Placeholder = "Repository name (e.g., my-awesome-project)"
	nameInput.Focus()
	nameInput.CharLimit = 100
	nameInput.Width = width - 4
	m.RepoNameInput = nameInput

	descInput := textinput.New()
	descInput.Placeholder = "Repository description (optional)"
	descInput.CharLimit = 200
	descInput.Width = width - 4
	m.RepoDescInput = descInput

	// Initialize spinner for repo creation
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	m.CreateSpinner = s

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
	// Account for: header(1) + status(1) + newline(1) + tabs with border(3) + divider(1) + controls(2) + padding(2) = 11
	listHeight := m.Height - 11
	if listHeight < 5 {
		listHeight = 5
	}

	// Create distributions list
	distList := list.New(distributions, list.NewDefaultDelegate(), m.Width, listHeight)
	distList.SetShowTitle(false)
	distList.SetShowStatusBar(false)
	distList.SetFilteringEnabled(false)
	distList.SetShowHelp(false)
	m.Lists[0] = distList

	// Initialize build settings list
	buildItems := []list.Item{
		BuildItem{Name: "Run tests before release", Value: "go test ./...", Enabled: true},
		BuildItem{Name: "Clean build directory", Value: "", Enabled: true},
		BuildItem{Name: "Build for all platforms", Value: "darwin, linux, windows", Enabled: false},
		BuildItem{Name: "Include ARM64 builds", Value: "", Enabled: false},
	}

	buildList := list.New(buildItems, list.NewDefaultDelegate(), m.Width, listHeight)
	buildList.SetShowTitle(false)
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
	advList.SetShowTitle(false)
	advList.SetShowStatusBar(false)
	advList.SetFilteringEnabled(false)
	advList.SetShowHelp(false)
	m.Lists[2] = advList

	// Initialize cleanup list with real git status
	cleanupItems := m.loadGitStatus()
	cleanupList := list.New(cleanupItems, list.NewDefaultDelegate(), m.Width, listHeight)
	cleanupList.SetShowTitle(false)
	cleanupList.SetShowStatusBar(false)
	cleanupList.SetFilteringEnabled(false)
	cleanupList.SetShowHelp(false)
	m.Lists[3] = cleanupList

	// Cache GitHub status on initialization
	m.refreshGitHubStatus()

	return m
}

// createRepoCmd creates a GitHub repo asynchronously
func createRepoCmd(isPrivate bool, name, description string) tea.Cmd {
	return func() tea.Msg {
		err := gitcleanup.CreateGitHubRepo(isPrivate, name, description)
		return repoCreatedMsg{err: err}
	}
}

// refreshGitHubStatus updates cached GitHub repo status
func (m *ConfigureModel) refreshGitHubStatus() {
	if gitcleanup.HasGitRepo() && gitcleanup.HasGitHubRemote() {
		m.HasGitRemote = true
		owner, repo, err := gitcleanup.GetRepoInfo()
		if err == nil {
			m.GitHubOwner = owner
			m.GitHubRepo = repo
			m.GitHubRepoExists = gitcleanup.CheckGitHubRepoExists()
		}
	} else {
		m.HasGitRemote = false
		m.GitHubRepoExists = false
	}
}

// loadGitStatus loads current git status and categorizes files
func (m *ConfigureModel) loadGitStatus() []list.Item {
	items := []list.Item{}

	// Add GitHub repo creation option if needed
	if gitcleanup.HasGitRepo() {
		if !gitcleanup.HasGitHubRemote() {
			items = append(items, CleanupItem{
				Path:     "Create GitHub repository",
				Status:   "+",
				Category: "github-new",
				Action:   "skip",
			})
		} else if !gitcleanup.CheckGitHubRepoExists() {
			owner, repo, _ := gitcleanup.GetRepoInfo()
			items = append(items, CleanupItem{
				Path:     fmt.Sprintf("Push to github.com/%s/%s", owner, repo),
				Status:   "↑",
				Category: "github-push",
				Action:   "skip",
			})
		}
	}

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

		// Truncate path if too long
		path := gf.Path
		if m.Width > 0 && len(path) > m.Width-15 {
			path = "..." + path[len(path)-(m.Width-18):]
		}

		items = append(items, CleanupItem{
			Path:     path,
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
		// Account for: header(1) + status(1) + newline(1) + tabs with border(3) + divider(1) + controls(2) + padding(2) = 11
		listHeight := m.Height - 11
		if listHeight < 5 {
			listHeight = 5
		}
		for i := range m.Lists {
			m.Lists[i].SetWidth(m.Width)
			m.Lists[i].SetHeight(listHeight)
		}
	}

	switch msg := msg.(type) {
	case struct{}:
		// Clear status message after timeout
		m.CreateStatus = ""
		return m, nil
	case spinner.TickMsg:
		if m.IsCreating {
			var cmd tea.Cmd
			m.CreateSpinner, cmd = m.CreateSpinner.Update(msg)
			return m, cmd
		}
	case repoCreatedMsg:
		m.IsCreating = false
		if msg.err == nil {
			// Success - refresh and clear inputs
			m.CreatingRepo = false
			m.RepoNameInput.SetValue("")
			m.RepoDescInput.SetValue("")
			m.RepoIsPrivate = false
			m.RepoInputFocus = 0
			m.refreshGitHubStatus()
			m.Lists[3].SetItems(m.loadGitStatus())
			m.CreateStatus = "✓ Repository created successfully!"
			// Clear status after 3 seconds
			return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return struct{}{}
			})
		} else {
			m.CreateStatus = fmt.Sprintf("✗ Failed: %v", msg.err)
			// Clear status after 3 seconds
			return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return struct{}{}
			})
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		// Update list sizes (account for border around tabs)
		listHeight := msg.Height - 11
		if listHeight < 5 {
			listHeight = 5
		}
		for i := range m.Lists {
			m.Lists[i].SetWidth(msg.Width)
			m.Lists[i].SetHeight(listHeight)
		}
		// Update text input widths for repo creation
		m.RepoNameInput.Width = msg.Width - 4
		m.RepoDescInput.Width = msg.Width - 4
		m.Initialized = true

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.ActiveTab = (m.ActiveTab + 1) % 4
			return m, nil
		case "shift+tab":
			m.ActiveTab = (m.ActiveTab + 3) % 4
			return m, nil
		case " ", "space":
			// Toggle the selected item
			if m.ActiveTab < 0 || m.ActiveTab >= len(m.Lists) {
				return m, nil
			}
			currentList := &m.Lists[m.ActiveTab]
			selectedItem := currentList.SelectedItem()
			if selectedItem == nil {
				return m, nil
			}

			if i, ok := selectedItem.(DistributionItem); ok {
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
				// Special handling for GitHub repo creation
				if i.Category == "github-new" || i.Category == "github-push" {
					if i.Action == "create" {
						i.Action = "skip"
					} else {
						i.Action = "create"
					}
				} else {
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
	case repoCreatedMsg, spinner.TickMsg:
		// Pass these messages directly to the model's Update
		if configModel != nil {
			newModel, cmd := configModel.Update(msg)
			return currentPage, false, cmd, newModel
		}
	case tea.KeyMsg:
		// Handle repo creation mode inputs first
		if configModel != nil && configModel.CreatingRepo {
			switch msg.String() {
			case "esc":
				// Cancel repo creation
				configModel.CreatingRepo = false
				return currentPage, false, nil, configModel
			case "tab":
				// Cycle through name, description, and private toggle
				if configModel.RepoInputFocus == 0 {
					configModel.RepoInputFocus = 1
					configModel.RepoNameInput.Blur()
					configModel.RepoDescInput.Focus()
				} else if configModel.RepoInputFocus == 1 {
					configModel.RepoInputFocus = 2
					configModel.RepoDescInput.Blur()
				} else {
					configModel.RepoInputFocus = 0
					configModel.RepoNameInput.Focus()
				}
				return currentPage, false, nil, configModel
			case "enter":
				// Don't allow creation if already creating
				if configModel.IsCreating {
					return currentPage, false, nil, configModel
				}

				// Get repo details
				repoName := configModel.RepoNameInput.Value()
				if repoName == "" {
					repoName = gitcleanup.GetDefaultRepoName()
				}
				repoDesc := configModel.RepoDescInput.Value()

				// Start creating with spinner
				configModel.IsCreating = true
				configModel.CreateStatus = "Creating repository..."

				// Return commands for both spinner and repo creation
				return currentPage, false, tea.Batch(
					configModel.CreateSpinner.Tick,
					createRepoCmd(configModel.RepoIsPrivate, repoName, repoDesc),
				), configModel
			case " ":
				// Toggle private/public when on that option
				if configModel.RepoInputFocus == 2 {
					configModel.RepoIsPrivate = !configModel.RepoIsPrivate
					return currentPage, false, nil, configModel
				}
				// For text inputs, fall through to default to handle space as text
				fallthrough
			default:
				// Update the focused input
				if configModel.RepoInputFocus == 0 {
					configModel.RepoNameInput, _ = configModel.RepoNameInput.Update(msg)
				} else if configModel.RepoInputFocus == 1 {
					configModel.RepoDescInput, _ = configModel.RepoDescInput.Update(msg)
				}
				return currentPage, false, nil, configModel
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return currentPage, true, tea.Quit, configModel
		case "esc":
			return 0, false, nil, configModel // back to projectView
		case "r":
			// Refresh git status in cleanup tab
			if configModel != nil && configModel.ActiveTab == 3 {
				configModel.refreshGitHubStatus()
				configModel.Lists[3].SetItems(configModel.loadGitStatus())
			}
			return currentPage, false, nil, configModel
		case "G":
			// Start GitHub repo creation (uppercase G for GitHub) - available in all tabs
			if configModel != nil && !configModel.CreatingRepo {
				// Check if we need to create a GitHub repo
				if gitcleanup.HasGitRepo() {
					if !gitcleanup.HasGitHubRemote() || !gitcleanup.CheckGitHubRepoExists() {
						// Enter repo creation mode
						configModel.CreatingRepo = true
						configModel.RepoInputFocus = 0
						configModel.RepoNameInput.Focus()
						configModel.RepoDescInput.Blur()
						// Set default repo name from directory
						if configModel.RepoNameInput.Value() == "" {
							configModel.RepoNameInput.SetValue(gitcleanup.GetDefaultRepoName())
						}
					}
				}
			}
			return currentPage, false, nil, configModel
		case "s":
			// Save configuration or execute smart commit
			if configModel != nil && configModel.ActiveTab == 3 {
				// Handle GitHub repo creation first if needed
				for _, listItem := range configModel.Lists[3].Items() {
					if ci, ok := listItem.(CleanupItem); ok {
						if (ci.Category == "github-new" || ci.Category == "github-push") && ci.Action == "create" {
							// Create GitHub repo
							gitcleanup.CreateGitHubRepo(false, "", "") // public by default
							// Refresh list after creation
							configModel.Lists[3].SetItems(configModel.loadGitStatus())
							return currentPage, false, nil, configModel
						}
					}
				}

				// Execute smart commit for cleanup tab
				items := []gitcleanup.CleanupItem{}
				for _, listItem := range configModel.Lists[3].Items() {
					if ci, ok := listItem.(CleanupItem); ok {
						// Skip GitHub repo items
						if ci.Category == "github-new" || ci.Category == "github-push" {
							continue
						}
						items = append(items, gitcleanup.CleanupItem{
							Path:     ci.Path,
							Status:   ci.Status,
							Category: ci.Category,
							Action:   ci.Action,
						})
					}
				}

				if len(items) > 0 {
					if _, err := gitcleanup.ExecuteSmartCommit(items); err == nil {
						// Refresh the cleanup list after commit
						configModel.Lists[3].SetItems(configModel.loadGitStatus())
						// TODO: Show success message
					}
				}
			} else {
				// TODO: Implement save logic for other tabs
			}
			return 0, false, nil, configModel // back to projectView after save
		default:
			// Let unhandled keys fall through to the model's Update
		}
	}

	// Update the model if it exists
	if configModel != nil {
		newModel, cmd := configModel.Update(msg)
		return currentPage, false, cmd, newModel
	}
	return currentPage, false, nil, configModel
}