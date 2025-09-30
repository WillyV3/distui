package handlers

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"distui/internal/config"
	"distui/internal/gitcleanup"
	"distui/internal/models"
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

type loadCompleteMsg struct {
	cleanupModel *CleanupModel
}

type filesGeneratedMsg struct {
	err error
}

func generateFilesCmd(detectedProject *models.ProjectInfo, projectConfig *models.ProjectConfig, filesToGenerate []string, filesToDelete []string) tea.Cmd {
	return func() tea.Msg {
		// Delete files first
		if len(filesToDelete) > 0 {
			if err := DeleteConfigFiles(detectedProject.Path, filesToDelete); err != nil {
				return filesGeneratedMsg{err: err}
			}
		}
		// Then generate new files
		if len(filesToGenerate) > 0 {
			if err := GenerateConfigFiles(detectedProject, projectConfig, filesToGenerate); err != nil {
				return filesGeneratedMsg{err: err}
			}
		}
		return filesGeneratedMsg{err: nil}
	}
}

// ViewType for the configure screen
type ViewType uint

const (
	TabView ViewType = iota
	GitHubView
	CommitView
	SmartCommitConfirm
	GenerateConfigConsent
)

// ConfigureModel holds the state for the configure view
type ConfigureModel struct {
	ActiveTab       int
	Lists           [4]list.Model
	Width           int
	Height          int
	Initialized     bool
	CurrentView     ViewType
	Loading         bool

	// Project config for persistence
	ProjectConfig    *models.ProjectConfig
	ProjectIdentifier string

	// Sub-models for composable views
	CleanupModel    *CleanupModel
	GitHubModel     *GitHubModel
	CommitModel     *CommitModel

	// Config generation consent
	PendingGenerateFiles []string // Files that need to be generated
	PendingDeleteFiles   []string // Files that need to be deleted
	DetectedProject      *models.ProjectInfo
	GeneratingFiles      bool   // Currently generating files
	GenerateStatus       string // Status message for generation
	NeedsRegeneration    bool   // Config changed, files need regeneration

	// Legacy fields (to be removed)
	CreatingRepo       bool
	RepoNameInput      textinput.Model
	RepoDescInput      textinput.Model
	RepoInputFocus     int  // 0=name, 1=description, 2=private toggle, 3=account selection
	RepoIsPrivate      bool                   // true=private, false=public
	SelectedAccountIdx int                    // Index of selected GitHub account for repo creation
	GitHubAccounts     []models.GitHubAccount // List of available accounts/orgs
	// Spinner for repo creation
	IsCreating      bool
	CreateSpinner   spinner.Model
	CreateStatus    string
	// Cached git status to avoid expensive calls on every render
	GitHubRepoExists bool
	GitHubOwner      string
	GitHubRepo       string
	HasGitRemote     bool

	// NPM package name validation
	NPMNameStatus      string   // available, unavailable, checking, error
	NPMNameSuggestions []string // Alternative names if unavailable
	NPMNameError       string   // Error message if check failed

	// NPM package name editing
	NPMEditMode   bool
	NPMNameInput  textinput.Model
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

func (m *ConfigureModel) saveConfig() error {
	if m.ProjectConfig == nil || m.ProjectIdentifier == "" {
		return fmt.Errorf("no project config to save")
	}

	// Update config from current list states
	if m.ProjectConfig.Config == nil {
		m.ProjectConfig.Config = &models.ProjectSettings{}
	}

	// Update distributions (tab 1)
	items := m.Lists[1].Items()
	for _, item := range items {
		if dist, ok := item.(DistributionItem); ok {
			switch dist.Key {
			case "github":
				if m.ProjectConfig.Config.Distributions.GitHubRelease == nil {
					m.ProjectConfig.Config.Distributions.GitHubRelease = &models.GitHubReleaseConfig{}
				}
				m.ProjectConfig.Config.Distributions.GitHubRelease.Enabled = dist.Enabled
			case "homebrew":
				if m.ProjectConfig.Config.Distributions.Homebrew == nil {
					m.ProjectConfig.Config.Distributions.Homebrew = &models.HomebrewConfig{}
				}
				m.ProjectConfig.Config.Distributions.Homebrew.Enabled = dist.Enabled
			case "npm":
				if m.ProjectConfig.Config.Distributions.NPM == nil {
					m.ProjectConfig.Config.Distributions.NPM = &models.NPMConfig{}
				}
				m.ProjectConfig.Config.Distributions.NPM.Enabled = dist.Enabled
			case "go_install":
				if m.ProjectConfig.Config.Distributions.GoModule == nil {
					m.ProjectConfig.Config.Distributions.GoModule = &models.GoModuleConfig{}
				}
				m.ProjectConfig.Config.Distributions.GoModule.Enabled = dist.Enabled
			}
		}
	}

	// Update build settings (tab 2)
	buildItems := m.Lists[2].Items()
	for i, item := range buildItems {
		if build, ok := item.(BuildItem); ok {
			if i == 0 { // Run tests before release
				if m.ProjectConfig.Config.Release == nil {
					m.ProjectConfig.Config.Release = &models.ReleaseSettings{}
				}
				m.ProjectConfig.Config.Release.SkipTests = !build.Enabled
			}
		}
	}

	// Update advanced settings (tab 3)
	advItems := m.Lists[3].Items()
	for i, item := range advItems {
		if adv, ok := item.(BuildItem); ok {
			if m.ProjectConfig.Config.Release == nil {
				m.ProjectConfig.Config.Release = &models.ReleaseSettings{}
			}
			switch i {
			case 0: // Create draft releases
				m.ProjectConfig.Config.Release.CreateDraft = adv.Enabled
			case 1: // Mark as pre-release
				m.ProjectConfig.Config.Release.PreRelease = adv.Enabled
			case 2: // Generate changelog
				m.ProjectConfig.Config.Release.GenerateChangelog = adv.Enabled
			case 3: // Sign commits
				m.ProjectConfig.Config.Release.SignCommits = adv.Enabled
			}
		}
	}

	// Mark that regeneration is needed when config changes
	m.NeedsRegeneration = true

	// Save to disk
	return config.SaveProject(m.ProjectConfig)
}

// Initialize the configure model
func NewConfigureModel(width, height int, githubAccounts []models.GitHubAccount, projectConfig *models.ProjectConfig, detectedProject *models.ProjectInfo) *ConfigureModel {
	// Use provided dimensions or defaults
	if width <= 0 {
		width = 100
	}
	if height <= 0 {
		height = 30
	}

	// If no config exists, create initial structure from detected project
	if projectConfig == nil && detectedProject != nil {
		projectConfig = &models.ProjectConfig{
			Project: detectedProject,
			Config:  &models.ProjectSettings{},
			History: &models.ReleaseHistory{},
		}
	}

	m := &ConfigureModel{
		ActiveTab:         0,
		Width:             width,
		Height:            height,
		Initialized:       false,
		Loading:           true,
		GitHubAccounts:    githubAccounts,
		ProjectConfig:     projectConfig,
		DetectedProject:   detectedProject,
		ProjectIdentifier: "",
	}

	if projectConfig != nil && projectConfig.Project != nil {
		m.ProjectIdentifier = projectConfig.Project.Identifier
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

	// Initialize NPM package name input
	npmInput := textinput.New()
	npmInput.Placeholder = "package-name"
	npmInput.CharLimit = 214 // npm package name limit
	npmInput.Width = width - 8
	m.NPMNameInput = npmInput

	// Calculate list height more precisely
	// Account for UI elements:
	// - Header: 1 line
	// - Status: 2 lines (status + blank)
	// - Tabs: 3 lines (tabs + 2 blanks)
	// - Content box border: 2 lines (top + bottom)
	// - Content padding: 2 lines (vertical padding restored)
	// - Controls: 3 lines (2 blanks + control line)
	// Total: 13 lines of chrome, +1 if warning shown, +3 to 7 for NPM status
	chromeLines := 13
	if m.NeedsRegeneration {
		chromeLines = 14
	}
	// Add NPM status lines when on Distributions tab and status exists
	// NPM status: 2 blank lines + status line = 3 lines minimum
	// With suggestions: 2 blanks + status + 2 blanks + header + 3 suggestions + help = 10 lines
	if m.ActiveTab == 1 && m.NPMNameStatus == "unavailable" && len(m.NPMNameSuggestions) > 0 {
		chromeLines += 10 // 2 blanks + status + 2 blanks + header + 3 suggestions + help text
	} else if m.ActiveTab == 1 && m.NPMNameStatus != "" {
		chromeLines += 3 // 2 blanks + status line
	}
	listHeight := m.Height - chromeLines
	if listHeight < 5 {
		listHeight = 5
	}

	// Content box has no horizontal padding, just border (2 chars)
	listWidth := m.Width - 2
	if listWidth < 40 {
		listWidth = 40
	}

	// Don't load CleanupModel yet - will load async
	m.GitHubModel = NewGitHubModel(listWidth, listHeight)
	m.CurrentView = TabView

	// Empty cleanup list - will be populated when load completes
	cleanupItems := []list.Item{}
	cleanupList := list.New(cleanupItems, list.NewDefaultDelegate(), listWidth, listHeight)
	cleanupList.SetShowTitle(false)
	cleanupList.SetShowStatusBar(false)
	cleanupList.SetFilteringEnabled(false)
	cleanupList.SetShowHelp(false)
	m.Lists[0] = cleanupList

	// Initialize distributions list (tab 1) - using centralized builder
	distItems := BuildDistributionsList(projectConfig, detectedProject)
	distributions := make([]list.Item, len(distItems))
	for i, item := range distItems {
		distributions[i] = item
	}

	distList := list.New(distributions, list.NewDefaultDelegate(), listWidth, listHeight)
	distList.SetShowTitle(false)
	distList.SetShowStatusBar(false)
	distList.SetFilteringEnabled(false)
	distList.SetShowHelp(false)
	m.Lists[1] = distList

	// Initialize build settings list (tab 2) - load from config
	runTests := true
	cleanBuild := true
	allPlatforms := false
	arm64Builds := false

	if projectConfig != nil && projectConfig.Config != nil && projectConfig.Config.Release != nil {
		runTests = !projectConfig.Config.Release.SkipTests
	}

	buildItems := []list.Item{
		BuildItem{Name: "Run tests before release", Value: "go test ./...", Enabled: runTests},
		BuildItem{Name: "Clean build directory", Value: "", Enabled: cleanBuild},
		BuildItem{Name: "Build for all platforms", Value: "darwin, linux, windows", Enabled: allPlatforms},
		BuildItem{Name: "Include ARM64 builds", Value: "", Enabled: arm64Builds},
	}

	buildList := list.New(buildItems, list.NewDefaultDelegate(), listWidth, listHeight)
	buildList.SetShowTitle(false)
	buildList.SetShowStatusBar(false)
	buildList.SetFilteringEnabled(false)
	buildList.SetShowHelp(false)
	m.Lists[2] = buildList

	// Initialize advanced list (tab 3) - load from config
	createDraft := false
	preRelease := false
	generateChangelog := true
	signCommits := true

	if projectConfig != nil && projectConfig.Config != nil && projectConfig.Config.Release != nil {
		createDraft = projectConfig.Config.Release.CreateDraft
		preRelease = projectConfig.Config.Release.PreRelease
		generateChangelog = projectConfig.Config.Release.GenerateChangelog
		signCommits = projectConfig.Config.Release.SignCommits
	}

	advancedItems := []list.Item{
		BuildItem{Name: "Create draft releases", Value: "", Enabled: createDraft},
		BuildItem{Name: "Mark as pre-release", Value: "", Enabled: preRelease},
		BuildItem{Name: "Generate changelog", Value: "", Enabled: generateChangelog},
		BuildItem{Name: "Sign commits", Value: "", Enabled: signCommits},
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

// LoadCleanupCmd loads the cleanup model asynchronously
func LoadCleanupCmd(width, height int) tea.Cmd {
	return func() tea.Msg {
		cleanupModel := NewCleanupModel(width, height)
		return loadCompleteMsg{cleanupModel: cleanupModel}
	}
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
		// Use -u to set upstream tracking on first push
		cmd := exec.Command("git", "push", "-u", "origin", "HEAD")
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
			owner, repo, err := gitcleanup.GetRepoInfo()
			if err == nil && owner != "" && repo != "" {
				items = append(items, CleanupItem{
					Path:     fmt.Sprintf("Push to github.com/%s/%s", owner, repo),
					Status:   "↑",
					Category: "github-push",
					Action:   "skip",
				})
			}
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
		// Same calculation as in NewConfigureModel - Total UI chrome: 13 lines, +1 if warning, +3 to 10 for NPM
		chromeLines := 13
		if m.NeedsRegeneration {
			chromeLines = 14
		}
		// Add NPM status lines when on Distributions tab and status exists
		if m.ActiveTab == 1 && m.NPMNameStatus == "unavailable" && len(m.NPMNameSuggestions) > 0 {
			chromeLines += 10 // 2 blanks + status + 2 blanks + header + 3 suggestions + help text
		} else if m.ActiveTab == 1 && m.NPMNameStatus != "" {
			chromeLines += 3 // 2 blanks + status line
		}
		listHeight := m.Height - chromeLines
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
		var cmd tea.Cmd
		m.CreateSpinner, cmd = m.CreateSpinner.Update(msg)
		// Only continue ticking if we're showing the spinner
		if m.IsCreating || m.Loading || m.GeneratingFiles {
			return m, cmd
		}
		return m, nil
	case loadCompleteMsg:
		m.Loading = false
		m.Initialized = true
		m.CleanupModel = msg.cleanupModel
		m.Lists[0].SetItems(m.loadGitStatus())

		// Create project config file if it doesn't exist
		if m.ProjectConfig != nil && m.ProjectConfig.Project != nil {
			m.saveConfig() // This will create the file if needed
		}

		return m, nil
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
			if m.CleanupModel != nil {
				m.CleanupModel.Refresh()
			}
			m.Lists[0].SetItems(m.loadGitStatus())
			m.CreateStatus = "✓ Repository created successfully!"
			// Clear status after 1 second
			return m, tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
				return struct{}{}
			})
		} else {
			m.CreateStatus = fmt.Sprintf("✗ Failed: %v", msg.err)
			// Clear status after 1 second
			return m, tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
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
		// Clear status after 1 second
		return m, tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
			return struct{}{}
		})
	case filesGeneratedMsg:
		m.GeneratingFiles = false
		if msg.err == nil {
			m.GenerateStatus = "✓ Release files updated successfully!"
			m.CurrentView = TabView
			m.PendingGenerateFiles = nil
			m.PendingDeleteFiles = nil
			m.NeedsRegeneration = false
			// Reload git status to show the newly generated files
			m.Lists[0].SetItems(m.loadGitStatus())
		} else {
			m.GenerateStatus = fmt.Sprintf("✗ Generation failed: %v", msg.err)
		}
		// Clear status after 1 second
		return m, tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
			return struct{}{}
		})
	case npmNameCheckMsg:
		m.NPMNameStatus = string(msg.result.Status)
		m.NPMNameError = msg.result.Error
		m.NPMNameSuggestions = msg.result.Suggestions

		// No need to rebuild list - status shows below content box
		// List items stay clean, status displayed separately like cleanup tab

		return m, nil
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
		// Clear status after 1 second
		return m, tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
			return struct{}{}
		})
	case tea.WindowSizeMsg:
		// Note: app.go will update m.Width and m.Height after this handler returns
		// So we should use the current m.Width/m.Height which are already adjusted
		// by app.go (minus border and padding), not msg.Width/msg.Height

		// If model doesn't have dimensions yet, use msg dimensions minus app.go's chrome
		width := m.Width
		height := m.Height
		if width == 0 || height == 0 {
			width = msg.Width - 4   // border (2) + padding (2) from app.go View()
			height = msg.Height - 4
		}

		// Update list sizes with same calculation as NewConfigureModel
		// Total UI chrome: 13 lines, +1 if warning, +3 to 10 for NPM
		chromeLines := 13
		if m.NeedsRegeneration {
			chromeLines = 14
		}
		// Add NPM status lines when on Distributions tab and status exists
		if m.ActiveTab == 1 && m.NPMNameStatus == "unavailable" && len(m.NPMNameSuggestions) > 0 {
			chromeLines += 10 // 2 blanks + status + 2 blanks + header + 3 suggestions + help text
		} else if m.ActiveTab == 1 && m.NPMNameStatus != "" {
			chromeLines += 3 // 2 blanks + status line
		}
		listHeight := height - chromeLines
		if listHeight < 5 {
			listHeight = 5
		}
		// Content box has just border, no horizontal padding
		listWidth := width - 2
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
		if m.CommitModel != nil {
			m.CommitModel.SetSize(listWidth, listHeight)
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
		// Handle NPM name editing mode first
		if m.NPMEditMode {
			switch msg.String() {
			case "enter":
				// Save the new package name
				newName := m.NPMNameInput.Value()
				if newName != "" {
					if m.ProjectConfig.Config.Distributions.NPM == nil {
						m.ProjectConfig.Config.Distributions.NPM = &models.NPMConfig{}
					}
					m.ProjectConfig.Config.Distributions.NPM.PackageName = newName
					m.saveConfig()

					// Rebuild distributions list with new package name
					distItems := BuildDistributionsList(m.ProjectConfig, m.DetectedProject)
					listItems := make([]list.Item, len(distItems))
					for i, item := range distItems {
						listItems[i] = item
					}
					m.Lists[1].SetItems(listItems)

					// Trigger name check
					username := ""
					if m.DetectedProject != nil && m.DetectedProject.Repository != nil {
						username = m.DetectedProject.Repository.Owner
					}
					m.NPMNameStatus = "checking"
					m.NPMEditMode = false
					m.NPMNameInput.Blur()
					return m, checkNPMNameCmd(newName, username)
				}
				m.NPMEditMode = false
				m.NPMNameInput.Blur()
				return m, nil
			case "esc":
				// Cancel editing
				m.NPMEditMode = false
				m.NPMNameInput.Blur()
				return m, nil
			default:
				// Pass to text input
				var cmd tea.Cmd
				m.NPMNameInput, cmd = m.NPMNameInput.Update(msg)
				return m, cmd
			}
		}

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
			oldTab := m.ActiveTab
			m.ActiveTab = (m.ActiveTab + 1) % 4

			// Check NPM name when entering Distributions tab
			if m.ActiveTab == 1 && oldTab != 1 && m.NPMNameStatus == "" {
				if m.ProjectConfig != nil && m.ProjectConfig.Config != nil &&
					m.ProjectConfig.Config.Distributions.NPM != nil &&
					m.ProjectConfig.Config.Distributions.NPM.Enabled {

					packageName := m.ProjectConfig.Config.Distributions.NPM.PackageName
					if packageName == "" && m.DetectedProject != nil {
						if m.DetectedProject.Binary.Name != "" {
							packageName = m.DetectedProject.Binary.Name
						} else {
							packageName = m.DetectedProject.Module.Name
						}
					}

					username := ""
					if m.DetectedProject != nil && m.DetectedProject.Repository != nil {
						username = m.DetectedProject.Repository.Owner
					}

					m.NPMNameStatus = "checking"
					return m, checkNPMNameCmd(packageName, username)
				}
			}

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
				// Save config after toggle
				m.saveConfig()

				// If NPM was just enabled, trigger name check immediately
				if i.Key == "npm" && i.Enabled {
					packageName := ""
					if m.ProjectConfig != nil && m.ProjectConfig.Config != nil &&
						m.ProjectConfig.Config.Distributions.NPM != nil {
						packageName = m.ProjectConfig.Config.Distributions.NPM.PackageName
					}
					// If no package name yet, use project name
					if packageName == "" && m.DetectedProject != nil {
						if m.DetectedProject.Binary.Name != "" {
							packageName = m.DetectedProject.Binary.Name
						} else {
							packageName = m.DetectedProject.Module.Name
						}
					}

					username := ""
					if m.DetectedProject != nil && m.DetectedProject.Repository != nil {
						username = m.DetectedProject.Repository.Owner
					}

					m.NPMNameStatus = "checking"
					return m, checkNPMNameCmd(packageName, username)
				} else if i.Key == "npm" && !i.Enabled {
					// NPM was disabled, clear status
					m.NPMNameStatus = ""
					m.NPMNameError = ""
					m.NPMNameSuggestions = nil
				}
			} else if i, ok := currentList.SelectedItem().(BuildItem); ok {
				i.Enabled = !i.Enabled
				items := currentList.Items()
				items[currentList.Index()] = i
				currentList.SetItems(items)
				// Save config after toggle
				m.saveConfig()
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
		case "e":
			// Edit package name when on NPM item in Distributions tab
			if m.ActiveTab == 1 {
				selectedItem := m.Lists[1].SelectedItem()
				if dist, ok := selectedItem.(DistributionItem); ok && dist.Key == "npm" {
					// Enter edit mode
					m.NPMEditMode = true
					currentName := ""
					if m.ProjectConfig != nil && m.ProjectConfig.Config != nil &&
						m.ProjectConfig.Config.Distributions.NPM != nil {
						currentName = m.ProjectConfig.Config.Distributions.NPM.PackageName
					}
					m.NPMNameInput.SetValue(currentName)
					m.NPMNameInput.Focus()
					return m, nil
				}
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
	case repoCreatedMsg, pushCompleteMsg, commitCompleteMsg, spinner.TickMsg, filesGeneratedMsg:
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
		} else if configModel.CurrentView == GenerateConfigConsent {
			switch msg.String() {
			case "y", "Y":
				// Start async file generation with spinner
				configModel.GeneratingFiles = true
				configModel.GenerateStatus = "Generating release files..."
				return currentPage, false, tea.Batch(
					configModel.CreateSpinner.Tick,
					generateFilesCmd(configModel.DetectedProject, configModel.ProjectConfig, configModel.PendingGenerateFiles, configModel.PendingDeleteFiles),
				), configModel
			case "n", "N", "esc":
				configModel.CurrentView = TabView
				configModel.PendingGenerateFiles = nil
				configModel.PendingDeleteFiles = nil
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


		// Handle 'R' key to confirm and generate/regenerate release files (only in TabView, not on Cleanup tab)
		if (msg.String() == "r" || msg.String() == "R") && configModel.CurrentView == TabView && configModel.ActiveTab != 0 {
			missing := CheckMissingConfigFiles(configModel.DetectedProject, configModel.ProjectConfig)
			if len(missing) > 0 {
				// Files don't exist - show consent for creation
				configModel.PendingGenerateFiles = missing
				configModel.PendingDeleteFiles = nil
			} else {
				// Files exist - check what needs to be regenerated/deleted
				changes := GetConfigFileChanges(configModel.DetectedProject, configModel.ProjectConfig)
				configModel.PendingGenerateFiles = changes.FilesToGenerate
				configModel.PendingDeleteFiles = changes.FilesToDelete
			}
			configModel.CurrentView = GenerateConfigConsent
			return currentPage, false, nil, configModel
		}

		// Handle 'P' key to push to remote (only in TabView, Cleanup tab, and only if there are unpushed commits)
		if msg.String() == "P" && configModel.CurrentView == TabView && configModel.ActiveTab == 0 {
			// Push to remote only if there are unpushed commits
			if configModel.CleanupModel != nil && configModel.CleanupModel.RepoInfo != nil &&
				configModel.CleanupModel.RepoInfo.RemoteExists &&
				configModel.CleanupModel.RepoInfo.UnpushedCommits > 0 &&
				!configModel.IsCreating {
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

		// Handle 'G' key to create GitHub repository (only in TabView, Cleanup tab)
		if msg.String() == "G" && configModel != nil && !configModel.CreatingRepo && configModel.CurrentView == TabView && configModel.ActiveTab == 0 {
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
				// Cycle through name, description, private toggle, and account
				maxFields := 3
				if len(configModel.GitHubAccounts) > 0 {
					maxFields = 4
				}

				if configModel.RepoInputFocus == 0 {
					configModel.RepoInputFocus = 1
					configModel.RepoNameInput.Blur()
					configModel.RepoDescInput.Focus()
				} else if configModel.RepoInputFocus == 1 {
					configModel.RepoInputFocus = 2
					configModel.RepoDescInput.Blur()
				} else if configModel.RepoInputFocus == 2 && maxFields == 4 {
					configModel.RepoInputFocus = 3
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
				owner := ""
				if len(configModel.GitHubAccounts) > 0 && configModel.SelectedAccountIdx < len(configModel.GitHubAccounts) {
					owner = configModel.GitHubAccounts[configModel.SelectedAccountIdx].Username
				}

				// Return commands for both spinner and repo creation
				return currentPage, false, tea.Batch(
					configModel.CreateSpinner.Tick,
					createRepoCmd(configModel.RepoIsPrivate, repoName, repoDesc, owner),
				), configModel
			case " ":
				// Toggle based on focus
				if configModel.RepoInputFocus == 2 {
					// Toggle private/public
					configModel.RepoIsPrivate = !configModel.RepoIsPrivate
					return currentPage, false, nil, configModel
				}
				if configModel.RepoInputFocus == 3 && len(configModel.GitHubAccounts) > 0 {
					// Cycle through accounts
					configModel.SelectedAccountIdx++
					if configModel.SelectedAccountIdx >= len(configModel.GitHubAccounts) {
						configModel.SelectedAccountIdx = 0
					}
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
			// If in NPM edit mode, delegate to model's Update to handle cancellation
			if configModel != nil && configModel.NPMEditMode {
				newModel, cmd := configModel.Update(msg)
				return currentPage, false, cmd, newModel
			}
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