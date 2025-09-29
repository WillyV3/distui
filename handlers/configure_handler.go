package handlers

import (
	"fmt"
	"os/exec"
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

type pushCompleteMsg struct {
	err error
}

type commitCompleteMsg struct {
	message string
	err     error
}

// ViewType for the configure screen
type ViewType uint

const (
	TabView ViewType = iota
	GitHubView
	CommitView
	SmartCommitConfirm
)

// ConfigureModel holds the state for the configure view
type ConfigureModel struct {
	ActiveTab       int
	Lists           [4]list.Model
	Width           int
	Height          int
	Initialized     bool
	CurrentView     ViewType

	// Sub-models for composable views
	CleanupModel    *CleanupModel
	GitHubModel     *GitHubModel
	CommitModel     *CommitModel

	// Legacy fields (to be removed)
	CreatingRepo    bool
	RepoNameInput   textinput.Model
	RepoDescInput   textinput.Model
	RepoInputFocus  int  // 0=name, 1=description, 2=private toggle, 3=account selection
	RepoIsPrivate   bool // true=private, false=public
	SelectedAccountIdx int // Index of selected GitHub account for repo creation
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
	nameInput.CharLimit = 100
	nameInput.Width = width - 4
	nameInput.SetValue("")  // Explicitly set empty value
	nameInput.SetCursor(0) // Reset cursor position
	m.RepoNameInput = nameInput

	descInput := textinput.New()
	descInput.Placeholder = "Repository description (optional)"
	descInput.CharLimit = 200
	descInput.Width = width - 4
	descInput.SetValue("")  // Explicitly set empty value
	descInput.SetCursor(0) // Reset cursor position
	m.RepoDescInput = descInput

	// Initialize spinner for repo creation
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	m.CreateSpinner = s

	// Calculate list height more precisely
	// Account for UI elements:
	// - Header: 1 line
	// - Status: 2 lines (status + blank)
	// - Tabs: 3 lines (tabs + 2 blanks)
	// - Content box border: 2 lines (top + bottom)
	// - Content padding: 2 lines (vertical padding restored)
	// - Controls: 3 lines (2 blanks + control line)
	// Total: 13 lines of chrome
	listHeight := m.Height - 13
	if listHeight < 5 {
		listHeight = 5
	}

	// Content box has no horizontal padding, just border (2 chars)
	listWidth := m.Width - 2
	if listWidth < 40 {
		listWidth = 40
	}

	// Initialize sub-models with content dimensions
	m.CleanupModel = NewCleanupModel(listWidth, listHeight)
	m.GitHubModel = NewGitHubModel(listWidth, listHeight)
	m.CurrentView = TabView

	// Initialize cleanup list first (tab 0)
	cleanupItems := m.loadGitStatus()
	cleanupList := list.New(cleanupItems, list.NewDefaultDelegate(), listWidth, listHeight)
	cleanupList.SetShowTitle(false)
	cleanupList.SetShowStatusBar(false)
	cleanupList.SetFilteringEnabled(false)
	cleanupList.SetShowHelp(false)
	m.Lists[0] = cleanupList

	// Initialize distributions list (tab 1)
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

	distList := list.New(distributions, list.NewDefaultDelegate(), listWidth, listHeight)
	distList.SetShowTitle(false)
	distList.SetShowStatusBar(false)
	distList.SetFilteringEnabled(false)
	distList.SetShowHelp(false)
	m.Lists[1] = distList

	// Initialize build settings list (tab 2)
	buildItems := []list.Item{
		BuildItem{Name: "Run tests before release", Value: "go test ./...", Enabled: true},
		BuildItem{Name: "Clean build directory", Value: "", Enabled: true},
		BuildItem{Name: "Build for all platforms", Value: "darwin, linux, windows", Enabled: false},
		BuildItem{Name: "Include ARM64 builds", Value: "", Enabled: false},
	}

	buildList := list.New(buildItems, list.NewDefaultDelegate(), listWidth, listHeight)
	buildList.SetShowTitle(false)
	buildList.SetShowStatusBar(false)
	buildList.SetFilteringEnabled(false)
	buildList.SetShowHelp(false)
	m.Lists[2] = buildList

	// Initialize advanced list (tab 3)
	advancedItems := []list.Item{
		BuildItem{Name: "Create draft releases", Value: "", Enabled: false},
		BuildItem{Name: "Mark as pre-release", Value: "", Enabled: false},
		BuildItem{Name: "Generate changelog", Value: "", Enabled: true},
		BuildItem{Name: "Sign commits", Value: "", Enabled: true},
	}

	advList := list.New(advancedItems, list.NewDefaultDelegate(), listWidth, listHeight)
	advList.SetShowTitle(false)
	advList.SetShowStatusBar(false)
	advList.SetFilteringEnabled(false)
	advList.SetShowHelp(false)
	m.Lists[3] = advList

	// Cache GitHub status on initialization
	m.refreshGitHubStatus()

	return m
}

// createRepoCmd creates a GitHub repo asynchronously
func createRepoCmd(isPrivate bool, name, description, owner string) tea.Cmd {
	return func() tea.Msg {
		err := gitcleanup.CreateGitHubRepo(isPrivate, name, description, owner)
		return repoCreatedMsg{err: err}
	}
}

// pushCmd pushes to remote asynchronously
func pushCmd() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("git", "push")
		err := cmd.Run()
		return pushCompleteMsg{err: err}
	}
}

// smartCommitCmd executes smart commit asynchronously
func smartCommitCmd(items []gitcleanup.CleanupItem) tea.Cmd {
	return func() tea.Msg {
		message, err := gitcleanup.ExecuteSmartCommit(items)
		return commitCompleteMsg{message: message, err: err}
	}
}

func regularCommitCmd(files []string, message string) tea.Cmd {
	return func() tea.Msg {
		// Convert file paths to GitFile structs
		var gitFiles []gitcleanup.GitFile
		for _, path := range files {
			gitFiles = append(gitFiles, gitcleanup.GitFile{
				Path: path,
			})
		}

		// Commit the files
		err := gitcleanup.CommitFiles(gitFiles, message)
		if err != nil {
			return commitCompleteMsg{err: err}
		}

		return commitCompleteMsg{message: message}
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
		// Same calculation as in NewConfigureModel - Total UI chrome: 13 lines
		listHeight := m.Height - 13
		if listHeight < 5 {
			listHeight = 5
		}
		// Content box has just border, no horizontal padding
		listWidth := m.Width - 2
		if listWidth < 40 {
			listWidth = 40
		}
		for i := range m.Lists {
			m.Lists[i].SetWidth(listWidth)
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
			m.Lists[0].SetItems(m.loadGitStatus())
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
	case pushCompleteMsg:
		m.IsCreating = false
		if msg.err == nil {
			m.CreateStatus = "✓ Pushed to remote successfully!"
			if m.CleanupModel != nil {
				m.CleanupModel.Refresh()
			}
		} else {
			m.CreateStatus = fmt.Sprintf("✗ Push failed: %v", msg.err)
		}
		// Clear status after 3 seconds
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return struct{}{}
		})
	case commitCompleteMsg:
		m.IsCreating = false
		if msg.err == nil {
			m.CreateStatus = fmt.Sprintf("✓ Committed: %s", msg.message)
			if m.CleanupModel != nil {
				m.CleanupModel.Refresh()
			}
			m.Lists[0].SetItems(m.loadGitStatus())
			// Return to main view
			m.CurrentView = TabView
			m.CommitModel = nil // Clean up
		} else {
			m.CreateStatus = fmt.Sprintf("✗ Commit failed: %v", msg.err)
		}
		// Clear status after 3 seconds
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return struct{}{}
		})
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		// Update list sizes with same calculation as NewConfigureModel
		// Total UI chrome: 13 lines
		listHeight := msg.Height - 13
		if listHeight < 5 {
			listHeight = 5
		}
		// Content box has just border, no horizontal padding
		listWidth := msg.Width - 2
		if listWidth < 40 {
			listWidth = 40
		}

		// Update sub-models with CONTENT dimensions, not window dimensions
		if m.CleanupModel != nil {
			m.CleanupModel.Update(listWidth, listHeight)
		}
		if m.GitHubModel != nil {
			m.GitHubModel.SetSize(listWidth, listHeight)
		}
		for i := range m.Lists {
			m.Lists[i].SetWidth(listWidth)
			m.Lists[i].SetHeight(listHeight)
		}
		// Update text input widths for repo creation
		m.RepoNameInput.Width = msg.Width - 4
		m.RepoDescInput.Width = msg.Width - 4
		m.Initialized = true

	case tea.KeyMsg:
		// If we're on the cleanup tab and there are no changes, delegate navigation to the repo browser
		if m.ActiveTab == 0 && m.CleanupModel != nil && !m.CleanupModel.HasChanges() {
			// Check if this is a navigation key that should go to the repo browser
			switch msg.String() {
			case "j", "down", "k", "up", "g", "G", "h", "left", "backspace", "l", "right", "enter":
				var cmd tea.Cmd
				m.CleanupModel, cmd = m.CleanupModel.HandleKey(msg)
				return m, cmd
			}
		}

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
			if m.ActiveTab == 1 {
				items := m.Lists[1].Items()
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
				m.Lists[1].SetItems(items)
			}
			return m, nil
		default:
			// Don't pass navigation to the list if we're showing the repo browser
			if m.ActiveTab == 0 && m.CleanupModel != nil && !m.CleanupModel.HasChanges() {
				// Already handled above
				return m, nil
			}
			// Pass through to the active list
			var cmd tea.Cmd
			m.Lists[m.ActiveTab], cmd = m.Lists[m.ActiveTab].Update(msg)
			return m, cmd
		}
	}

	// Don't update the list if we're on cleanup tab with repo browser
	if m.ActiveTab == 0 && m.CleanupModel != nil && !m.CleanupModel.HasChanges() {
		return m, nil
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
	case repoCreatedMsg, pushCompleteMsg, commitCompleteMsg, spinner.TickMsg:
		// Pass these messages directly to the model's Update
		if configModel != nil {
			newModel, cmd := configModel.Update(msg)
			return currentPage, false, cmd, newModel
		}
	case tea.KeyMsg:
		// Handle view switching
		if configModel.CurrentView == CommitView {
			switch msg.String() {
			case "esc":
				configModel.CurrentView = TabView
				configModel.CommitModel = nil // Reset
				configModel.CleanupModel.Refresh()
				return currentPage, false, nil, configModel
			case "enter":
				// Only process if we're at the commit message stage and have files staged
				if configModel.CommitModel != nil && configModel.CommitModel.IsComplete() && configModel.CommitModel.HasStagedFiles() {
					// Execute the commit
					message := configModel.CommitModel.CommitMessage.Value()
					if message != "" {
						configModel.IsCreating = true
						configModel.CreateStatus = "Committing files..."
						// Stage the selected files and commit
						stagedFiles := configModel.CommitModel.GetStagedFiles()
						return currentPage, false, tea.Batch(
							configModel.CreateSpinner.Tick,
							regularCommitCmd(stagedFiles, message),
						), configModel
					}
				}
			default:
				// Pass key to commit model
				if configModel.CommitModel != nil {
					var cmd tea.Cmd
					configModel.CommitModel, cmd = configModel.CommitModel.Update(msg)
					return currentPage, false, cmd, configModel
				}
			}
		} else if configModel.CurrentView == SmartCommitConfirm {
			switch msg.String() {
			case "y", "Y":
				// User confirmed, execute smart commit
				changes, _ := gitcleanup.GetFileChanges()
				items := []gitcleanup.CleanupItem{}

				for _, change := range changes {
					items = append(items, gitcleanup.CleanupItem{
						Path:     change.Path,
						Status:   change.Status,
						Category: "auto",
						Action:   "commit",
					})
				}

				if len(items) > 0 {
					configModel.CurrentView = TabView
					configModel.IsCreating = true
					configModel.CreateStatus = "Committing changes..."
					return currentPage, false, tea.Batch(
						configModel.CreateSpinner.Tick,
						smartCommitCmd(items),
					), configModel
				}
			case "n", "N", "esc":
				// User cancelled
				configModel.CurrentView = TabView
				return currentPage, false, nil, configModel
			}
		} else if configModel.CurrentView == GitHubView {
			switch msg.String() {
			case "esc":
				configModel.CurrentView = TabView
				configModel.CleanupModel.Refresh()
				return currentPage, false, nil, configModel
			default:
				newModel, cmd := configModel.GitHubModel.Update(msg)
				configModel.GitHubModel = newModel
				return currentPage, false, cmd, configModel
			}
		}

		// G key handler removed - repo browser is now embedded in cleanup view

		// Handle 'C' key to switch to Commit view (only in TabView)
		if msg.String() == "C" && configModel.CurrentView == TabView && configModel.ActiveTab == 0 {
			if configModel.CleanupModel != nil && configModel.CleanupModel.HasChanges() {
				// Initialize commit model if needed
				if configModel.CommitModel == nil {
					configModel.CommitModel = NewCommitModel(configModel.Width-2, configModel.Height-13)
				}
				configModel.CurrentView = CommitView
			}
			return currentPage, false, nil, configModel
		}

		// Handle 'P' key to push to remote (only in TabView, Cleanup tab)
		if msg.String() == "P" && configModel.CurrentView == TabView && configModel.ActiveTab == 0 {
			// Push to remote
			if configModel.CleanupModel != nil && configModel.CleanupModel.RepoInfo != nil &&
				configModel.CleanupModel.RepoInfo.RemoteExists && !configModel.IsCreating {
				// Start spinner and execute async push
				configModel.IsCreating = true
				configModel.CreateStatus = "Pushing to remote..."
				return currentPage, false, tea.Batch(
					configModel.CreateSpinner.Tick,
					pushCmd(),
				), configModel
			}
			return currentPage, false, nil, configModel
		}

		// Legacy handler for old 'G' key behavior (to be removed)
		if false && msg.String() == "G" && configModel != nil && !configModel.CreatingRepo {
			// Check if we need to create a GitHub repo
			if gitcleanup.HasGitRepo() {
				if !gitcleanup.HasGitHubRemote() || !gitcleanup.CheckGitHubRepoExists() {
					// Enter repo creation mode
					configModel.CreatingRepo = true
					configModel.RepoInputFocus = 0

					// Create fresh inputs to avoid cursor issues
					nameInput := textinput.New()
					defaultName := gitcleanup.GetDefaultRepoName()
					nameInput.Placeholder = fmt.Sprintf("Repository name (default: %s)", defaultName)
					nameInput.CharLimit = 100
					if configModel.Width > 0 {
						nameInput.Width = configModel.Width - 4
					}
					// Explicitly clear value and reset before focusing
					nameInput.SetValue("")
					nameInput.Focus()
					// Clear again after focusing in case something got buffered
					nameInput.SetValue("")
					nameInput.CursorStart()
					configModel.RepoNameInput = nameInput

					descInput := textinput.New()
					descInput.Placeholder = "Repository description (optional)"
					descInput.CharLimit = 200
					if configModel.Width > 0 {
						descInput.Width = configModel.Width - 4
					}
					configModel.RepoDescInput = descInput
				}
			}
			// Always return here to consume the 'G' key
			return currentPage, false, nil, configModel
		}

		// Handle repo creation mode inputs
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

				// Get repo details (use default if empty)
				repoName := configModel.RepoNameInput.Value()
				if repoName == "" {
					repoName = gitcleanup.GetDefaultRepoName()
				}
				repoDesc := configModel.RepoDescInput.Value()

				// Start creating with spinner
				configModel.IsCreating = true
				configModel.CreateStatus = "Creating repository..."

				// Get the owner (account) to create under
				// TODO: This needs to be populated from global config accounts
				owner := "" // Will use default account for now

				// Return commands for both spinner and repo creation
				return currentPage, false, tea.Batch(
					configModel.CreateSpinner.Tick,
					createRepoCmd(configModel.RepoIsPrivate, repoName, repoDesc, owner),
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
			if configModel != nil && configModel.ActiveTab == 0 {
				configModel.refreshGitHubStatus()
				configModel.Lists[0].SetItems(configModel.loadGitStatus())
			}
			return currentPage, false, nil, configModel
		case "s":
			// Show smart commit confirmation
			if configModel != nil && configModel.ActiveTab == 0 && !configModel.IsCreating {
				if configModel.CleanupModel != nil && configModel.CleanupModel.HasChanges() {
					// Switch to confirmation view
					configModel.CurrentView = SmartCommitConfirm
					return currentPage, false, nil, configModel
				}
			}
			return currentPage, false, nil, configModel // stay in configure view
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