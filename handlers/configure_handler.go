package handlers

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"distui/internal/config"
	"distui/internal/detection"
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

type distributionVerifiedMsg struct {
	homebrewVersion string
	homebrewExists  bool
	npmVersion      string
	npmExists       bool
	err             error
}

type distributionDetectedMsg struct {
	homebrewTap       string
	homebrewFormula   string
	homebrewVersion   string
	homebrewExists    bool
	homebrewFromFile  bool // Detected from .goreleaser.yaml
	npmPackage        string
	npmVersion        string
	npmExists         bool
	npmFromFile       bool // Detected from package.json
}

type githubStatusMsg struct {
	hasRemote  bool
	owner      string
	repo       string
	repoExists bool
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
	SmartCommitFileSelection
	GenerateConfigConsent
	SmartCommitPrefsView
	RepoCleanupView
	FirstTimeSetupView
	ModeSwitchWarning
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
	GlobalConfig      *models.GlobalConfig

	// Sub-models for composable views
	CleanupModel          *CleanupModel
	GitHubModel           *GitHubModel
	CommitModel           *CommitModel
	SmartCommitPrefsModel *SmartCommitPrefsModel
	RepoCleanupModel      *RepoCleanupModel
	BranchModal           *BranchSelectionModel
	FileSelectionModel    *FileSelectionModel
	ScanningRepo          bool
	ShowingBranchModal    bool

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

	// Background git polling
	GitWatcherActive bool // True if background watcher is running

	// NPM package name validation
	NPMNameStatus      string   // available, unavailable, checking, error
	NPMNameSuggestions []string // Alternative names if unavailable
	NPMNameError       string   // Error message if check failed

	// NPM package name editing
	NPMEditMode   bool
	NPMNameInput  textinput.Model

	// First-time setup for existing distributions
	FirstTimeSetup             bool
	FirstTimeSetupConfirmation bool   // Show confirmation screen before verifying
	DetectingDistributions     bool   // Auto-detecting from tap/npm
	AutoDetected               bool   // True if distributions were auto-detected
	HomebrewDetectedFromFile   bool   // Detected from .goreleaser.yaml
	NPMDetectedFromFile        bool   // Detected from package.json
	HomebrewCheckEnabled       bool
	NPMCheckEnabled            bool
	HomebrewTapInput           textinput.Model
	HomebrewFormulaInput       textinput.Model
	NPMPackageInput            textinput.Model
	FirstTimeSetupFocus        int    // 0=homebrew checkbox, 1=tap, 2=formula, 3=npm checkbox, 4=package
	VerifyingDistributions     bool
	DistributionVerifyError    string

	// Custom config overwrite warning
	ShowOverwriteWarning bool

	// First-time setup custom file detection
	FirstTimeSetupCustomChoice bool
	CustomFilesDetected        []string
	FilesToOverwrite     []string
	PendingSaveConfig    *models.ProjectConfig
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

func (m *ConfigureModel) detectFilesToOverwrite() []string {
	var files []string
	if m.DetectedProject == nil {
		return files
	}

	projectPath := m.DetectedProject.Path

	// Check .goreleaser.yaml
	goreleaserPaths := []string{
		projectPath + "/.goreleaser.yaml",
		projectPath + "/.goreleaser.yml",
	}
	for _, p := range goreleaserPaths {
		if detection.IsCustomConfig(p) {
			files = append(files, ".goreleaser.yaml")
			break
		}
	}

	// Check package.json
	packagePath := projectPath + "/package.json"
	if detection.IsCustomConfig(packagePath) {
		files = append(files, "package.json")
	}

	return files
}

func (m *ConfigureModel) saveConfig() error {
	return m.saveConfigWithRegenFlag(true)
}

func (m *ConfigureModel) saveConfigWithRegenFlag(needsRegen bool) error {
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
	if needsRegen {
		m.NeedsRegeneration = true
	}

	// Save to disk
	return config.SaveProject(m.ProjectConfig)
}

// Initialize the configure model
func NewConfigureModel(width, height int, githubAccounts []models.GitHubAccount, projectConfig *models.ProjectConfig, detectedProject *models.ProjectInfo, globalConfig *models.GlobalConfig) *ConfigureModel {
	// Use provided dimensions or defaults
	if width <= 0 {
		width = 100
	}
	if height <= 0 {
		height = 30
	}

	// Track if this is first-time (no saved config exists)
	hadSavedConfig := projectConfig != nil

	// If no config exists, create initial structure from detected project
	if projectConfig == nil && detectedProject != nil {
		projectConfig = &models.ProjectConfig{
			Project: detectedProject,
			Config:  &models.ProjectSettings{},
			History: &models.ReleaseHistory{},
		}
	}

	// Detect custom mode if project path available
	if detectedProject != nil && detectedProject.Path != "" {
		customMode, _, err := detection.DetectProjectMode(detectedProject.Path)
		if err == nil && projectConfig != nil {
			projectConfig.CustomFilesMode = customMode
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
		GlobalConfig:      globalConfig,
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

	// Check if this is first-time setup
	// Trigger if:
	// 1. No saved config + has versions (normal first-time)
	// 2. Bulk-imported: has distributions enabled but empty release history
	hasDistributionsEnabled := projectConfig != nil && projectConfig.Config != nil &&
		((projectConfig.Config.Distributions.Homebrew != nil && projectConfig.Config.Distributions.Homebrew.Enabled) ||
			(projectConfig.Config.Distributions.NPM != nil && projectConfig.Config.Distributions.NPM.Enabled))

	hasEmptyHistory := projectConfig != nil &&
		(projectConfig.History == nil || len(projectConfig.History.Releases) == 0)

	isBulkImported := hadSavedConfig && hasDistributionsEnabled && hasEmptyHistory

	// Check if user has already completed or skipped first-time setup
	alreadyCompletedSetup := projectConfig != nil && projectConfig.FirstTimeSetupCompleted

	isFirstTime := !alreadyCompletedSetup &&
		((!hadSavedConfig && detectedProject != nil &&
			detectedProject.Module != nil && detectedProject.Module.Version != "" &&
			detectedProject.Module.Version != "v0.0.1") ||
			isBulkImported)

	if isFirstTime {
		m.FirstTimeSetup = true
		m.DetectingDistributions = true
		m.CurrentView = FirstTimeSetupView
	}

	// Don't cache GitHub status synchronously - will be loaded async
	// m.refreshGitHubStatus()  // REMOVED: causes 500ms lag on view switch

	return m
}

// LoadCleanupCmd loads the cleanup model asynchronously
func LoadCleanupCmd(width, height int) tea.Cmd {
	return func() tea.Msg {
		cleanupModel := NewCleanupModel(width, height)
		return loadCompleteMsg{cleanupModel: cleanupModel}
	}
}

// RefreshGitHubStatusCmd loads GitHub status asynchronously (non-blocking)
func RefreshGitHubStatusCmd() tea.Cmd {
	return func() tea.Msg {
		msg := githubStatusMsg{}

		if gitcleanup.HasGitRepo() && gitcleanup.HasGitHubRemote() {
			msg.hasRemote = true
			owner, repo, err := gitcleanup.GetRepoInfo()
			if err == nil {
				msg.owner = owner
				msg.repo = repo
				msg.repoExists = gitcleanup.CheckGitHubRepoExists()
			}
		}

		return msg
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
	// Use CACHED data to avoid expensive git/API calls every 2 seconds
	if gitcleanup.HasGitRepo() {
		if !m.HasGitRemote {
			items = append(items, CleanupItem{
				Path:     "Create GitHub repository",
				Status:   "+",
				Category: "github-new",
				Action:   "skip",
			})
		} else if !m.GitHubRepoExists && m.GitHubOwner != "" && m.GitHubRepo != "" {
			items = append(items, CleanupItem{
				Path:     fmt.Sprintf("Push to github.com/%s/%s", m.GitHubOwner, m.GitHubRepo),
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

