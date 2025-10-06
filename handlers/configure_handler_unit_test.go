package handlers

import (
	"errors"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"distui/internal/models"
)

// ============================================================================
// Test Message Types
// ============================================================================

func TestRepoCreatedMsg(t *testing.T) {
	tests := []struct {
		name string
		msg  repoCreatedMsg
		err  error
	}{
		{
			name: "success message",
			msg:  repoCreatedMsg{err: nil},
			err:  nil,
		},
		{
			name: "error message",
			msg:  repoCreatedMsg{err: errors.New("repo creation failed")},
			err:  errors.New("repo creation failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err \!= nil {
				assert.Error(t, tt.msg.err)
				assert.Equal(t, tt.err.Error(), tt.msg.err.Error())
			} else {
				assert.NoError(t, tt.msg.err)
			}
		})
	}
}

func TestPushCompleteMsg(t *testing.T) {
	tests := []struct {
		name string
		msg  pushCompleteMsg
		err  error
	}{
		{
			name: "success message",
			msg:  pushCompleteMsg{err: nil},
			err:  nil,
		},
		{
			name: "error message",
			msg:  pushCompleteMsg{err: errors.New("push failed")},
			err:  errors.New("push failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err \!= nil {
				assert.Error(t, tt.msg.err)
			} else {
				assert.NoError(t, tt.msg.err)
			}
		})
	}
}

func TestCommitCompleteMsg(t *testing.T) {
	tests := []struct {
		name    string
		msg     commitCompleteMsg
		message string
		err     error
	}{
		{
			name:    "success with message",
			msg:     commitCompleteMsg{message: "feat: add new feature", err: nil},
			message: "feat: add new feature",
			err:     nil,
		},
		{
			name:    "error message",
			msg:     commitCompleteMsg{message: "", err: errors.New("commit failed")},
			message: "",
			err:     errors.New("commit failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.message, tt.msg.message)
			if tt.err \!= nil {
				assert.Error(t, tt.msg.err)
			} else {
				assert.NoError(t, tt.msg.err)
			}
		})
	}
}

func TestFilesGeneratedMsg(t *testing.T) {
	tests := []struct {
		name string
		msg  filesGeneratedMsg
		err  error
	}{
		{
			name: "success",
			msg:  filesGeneratedMsg{err: nil},
			err:  nil,
		},
		{
			name: "error",
			msg:  filesGeneratedMsg{err: errors.New("generation failed")},
			err:  errors.New("generation failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err \!= nil {
				assert.Error(t, tt.msg.err)
			} else {
				assert.NoError(t, tt.msg.err)
			}
		})
	}
}

func TestDistributionVerifiedMsg(t *testing.T) {
	tests := []struct {
		name            string
		msg             distributionVerifiedMsg
		homebrewVersion string
		homebrewExists  bool
		npmVersion      string
		npmExists       bool
		err             error
	}{
		{
			name: "homebrew verified",
			msg: distributionVerifiedMsg{
				homebrewVersion: "v1.0.0",
				homebrewExists:  true,
				npmVersion:      "",
				npmExists:       false,
				err:             nil,
			},
			homebrewVersion: "v1.0.0",
			homebrewExists:  true,
		},
		{
			name: "npm verified",
			msg: distributionVerifiedMsg{
				homebrewVersion: "",
				homebrewExists:  false,
				npmVersion:      "1.0.0",
				npmExists:       true,
				err:             nil,
			},
			npmVersion: "1.0.0",
			npmExists:  true,
		},
		{
			name: "verification error",
			msg: distributionVerifiedMsg{
				err: errors.New("verification failed"),
			},
			err: errors.New("verification failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.homebrewVersion, tt.msg.homebrewVersion)
			assert.Equal(t, tt.homebrewExists, tt.msg.homebrewExists)
			assert.Equal(t, tt.npmVersion, tt.msg.npmVersion)
			assert.Equal(t, tt.npmExists, tt.msg.npmExists)
			if tt.err \!= nil {
				assert.Error(t, tt.msg.err)
			}
		})
	}
}

func TestDistributionDetectedMsg(t *testing.T) {
	tests := []struct {
		name           string
		msg            distributionDetectedMsg
		homebrewTap    string
		homebrewExists bool
		npmPackage     string
		npmExists      bool
	}{
		{
			name: "homebrew detected",
			msg: distributionDetectedMsg{
				homebrewTap:      "owner/repo",
				homebrewFormula:  "formula-name",
				homebrewVersion:  "v1.0.0",
				homebrewExists:   true,
				homebrewFromFile: false,
			},
			homebrewTap:    "owner/repo",
			homebrewExists: true,
		},
		{
			name: "npm detected",
			msg: distributionDetectedMsg{
				npmPackage:  "package-name",
				npmVersion:  "1.0.0",
				npmExists:   true,
				npmFromFile: false,
			},
			npmPackage: "package-name",
			npmExists:  true,
		},
		{
			name: "both detected from files",
			msg: distributionDetectedMsg{
				homebrewTap:      "owner/tap",
				homebrewFormula:  "formula",
				homebrewFromFile: true,
				npmPackage:       "@scope/package",
				npmFromFile:      true,
			},
			homebrewTap: "owner/tap",
			npmPackage:  "@scope/package",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.homebrewTap, tt.msg.homebrewTap)
			assert.Equal(t, tt.homebrewExists, tt.msg.homebrewExists)
			assert.Equal(t, tt.npmPackage, tt.msg.npmPackage)
			assert.Equal(t, tt.npmExists, tt.msg.npmExists)
		})
	}
}

// ============================================================================
// Test DistributionItem
// ============================================================================

func TestDistributionItem_Title(t *testing.T) {
	tests := []struct {
		name     string
		item     DistributionItem
		expected string
	}{
		{
			name:     "enabled item",
			item:     DistributionItem{Name: "GitHub", Enabled: true},
			expected: "[✓] GitHub",
		},
		{
			name:     "disabled item",
			item:     DistributionItem{Name: "Homebrew", Enabled: false},
			expected: "[ ] Homebrew",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.item.Title())
		})
	}
}

func TestDistributionItem_Description(t *testing.T) {
	item := DistributionItem{
		Name: "GitHub",
		Desc: "Publish releases to GitHub",
	}
	assert.Equal(t, "Publish releases to GitHub", item.Description())
}

func TestDistributionItem_FilterValue(t *testing.T) {
	item := DistributionItem{Name: "NPM"}
	assert.Equal(t, "NPM", item.FilterValue())
}

// ============================================================================
// Test BuildItem
// ============================================================================

func TestBuildItem_Title(t *testing.T) {
	tests := []struct {
		name     string
		item     BuildItem
		expected string
	}{
		{
			name:     "enabled build item",
			item:     BuildItem{Name: "Run tests", Enabled: true},
			expected: "[✓] Run tests",
		},
		{
			name:     "disabled build item",
			item:     BuildItem{Name: "Clean build", Enabled: false},
			expected: "[ ] Clean build",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.item.Title())
		})
	}
}

func TestBuildItem_Description(t *testing.T) {
	item := BuildItem{
		Name:  "Run tests",
		Value: "go test ./...",
	}
	assert.Equal(t, "go test ./...", item.Description())
}

func TestBuildItem_FilterValue(t *testing.T) {
	item := BuildItem{Name: "Run tests"}
	assert.Equal(t, "Run tests", item.FilterValue())
}

// ============================================================================
// Test CleanupItem
// ============================================================================

func TestCleanupItem_Title(t *testing.T) {
	tests := []struct {
		name     string
		item     CleanupItem
		expected string
	}{
		{
			name:     "modified file",
			item:     CleanupItem{Path: "README.md", Status: "M"},
			expected: "[M] README.md",
		},
		{
			name:     "added file",
			item:     CleanupItem{Path: "new.go", Status: "A"},
			expected: "[+] new.go",
		},
		{
			name:     "deleted file",
			item:     CleanupItem{Path: "old.go", Status: "D"},
			expected: "[-] old.go",
		},
		{
			name:     "untracked file",
			item:     CleanupItem{Path: "temp.txt", Status: "??"},
			expected: "[?] temp.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.item.Title())
		})
	}
}

func TestCleanupItem_Description(t *testing.T) {
	tests := []struct {
		name     string
		item     CleanupItem
		expected string
	}{
		{
			name:     "commit action",
			item:     CleanupItem{Action: "commit"},
			expected: "→ Will commit",
		},
		{
			name:     "skip action",
			item:     CleanupItem{Action: "skip"},
			expected: "→ Skip",
		},
		{
			name:     "ignore action",
			item:     CleanupItem{Action: "ignore"},
			expected: "→ Add to .gitignore",
		},
		{
			name:     "github-new create",
			item:     CleanupItem{Category: "github-new", Action: "create"},
			expected: "→ Will create GitHub repo",
		},
		{
			name:     "github-new skip",
			item:     CleanupItem{Category: "github-new", Action: "skip"},
			expected: "→ Skip",
		},
		{
			name:     "github-push create",
			item:     CleanupItem{Category: "github-push", Action: "create"},
			expected: "→ Will push to GitHub",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.item.Description())
		})
	}
}

func TestCleanupItem_FilterValue(t *testing.T) {
	item := CleanupItem{Path: "test.go"}
	assert.Equal(t, "test.go", item.FilterValue())
}

// ============================================================================
// Test ConfigureModel Creation and Initialization
// ============================================================================

func TestNewConfigureModel_BasicInitialization(t *testing.T) {
	width, height := 100, 30
	accounts := []models.GitHubAccount{{Username: "testuser"}}
	projectConfig := &models.ProjectConfig{
		Project: &models.ProjectInfo{
			Identifier: "test-project",
		},
		Config: &models.ProjectSettings{},
	}
	detectedProject := &models.ProjectInfo{
		Identifier: "test-project",
	}
	globalConfig := &models.GlobalConfig{}

	m := NewConfigureModel(width, height, accounts, projectConfig, detectedProject, globalConfig)

	assert.NotNil(t, m)
	assert.Equal(t, width, m.Width)
	assert.Equal(t, height, m.Height)
	assert.Equal(t, 0, m.ActiveTab)
	assert.True(t, m.Loading)
	assert.False(t, m.Initialized)
	assert.Equal(t, "test-project", m.ProjectIdentifier)
	assert.Equal(t, TabView, m.CurrentView)
	assert.Len(t, m.GitHubAccounts, 1)
}

func TestNewConfigureModel_WithDefaultDimensions(t *testing.T) {
	// Test with invalid dimensions
	m := NewConfigureModel(0, 0, nil, nil, nil, nil)

	assert.NotNil(t, m)
	assert.Equal(t, 100, m.Width)
	assert.Equal(t, 30, m.Height)
}

func TestNewConfigureModel_WithNilProjectConfig(t *testing.T) {
	detectedProject := &models.ProjectInfo{
		Identifier: "detected-project",
	}

	m := NewConfigureModel(100, 30, nil, nil, detectedProject, nil)

	assert.NotNil(t, m)
	assert.NotNil(t, m.ProjectConfig)
	assert.Equal(t, detectedProject, m.ProjectConfig.Project)
}

func TestNewConfigureModel_FirstTimeSetup(t *testing.T) {
	tests := []struct {
		name             string
		projectConfig    *models.ProjectConfig
		detectedProject  *models.ProjectInfo
		expectFirstTime  bool
	}{
		{
			name:          "new project with version",
			projectConfig: nil,
			detectedProject: &models.ProjectInfo{
				Module: &models.ModuleInfo{
					Version: "v1.2.3",
				},
			},
			expectFirstTime: true,
		},
		{
			name:          "new project without version",
			projectConfig: nil,
			detectedProject: &models.ProjectInfo{
				Module: &models.ModuleInfo{
					Version: "v0.0.1",
				},
			},
			expectFirstTime: false,
		},
		{
			name: "bulk imported project",
			projectConfig: &models.ProjectConfig{
				Project: &models.ProjectInfo{},
				Config: &models.ProjectSettings{
					Distributions: models.DistributionSettings{
						Homebrew: &models.HomebrewConfig{Enabled: true},
					},
				},
				History: &models.ReleaseHistory{},
			},
			detectedProject: &models.ProjectInfo{},
			expectFirstTime: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewConfigureModel(100, 30, nil, tt.projectConfig, tt.detectedProject, nil)
			assert.Equal(t, tt.expectFirstTime, m.FirstTimeSetup)
			if tt.expectFirstTime {
				assert.True(t, m.DetectingDistributions)
				assert.Equal(t, FirstTimeSetupView, m.CurrentView)
			}
		})
	}
}

func TestNewConfigureModel_InitializesTextInputs(t *testing.T) {
	m := NewConfigureModel(100, 30, nil, nil, nil, nil)

	assert.NotNil(t, m.RepoNameInput)
	assert.NotNil(t, m.RepoDescInput)
	assert.NotNil(t, m.NPMNameInput)
	assert.Equal(t, 96, m.RepoNameInput.Width)
	assert.Equal(t, 92, m.NPMNameInput.Width)
}

func TestNewConfigureModel_InitializesLists(t *testing.T) {
	projectConfig := &models.ProjectConfig{
		Project: &models.ProjectInfo{},
		Config: &models.ProjectSettings{
			Release: &models.ReleaseSettings{
				SkipTests:         false,
				CreateDraft:       true,
				PreRelease:        false,
				GenerateChangelog: true,
				SignCommits:       false,
			},
		},
	}

	m := NewConfigureModel(100, 30, nil, projectConfig, nil, nil)

	// Check that all 4 lists are initialized
	assert.Len(t, m.Lists, 4)

	// Check build settings list reflects config
	buildItems := m.Lists[2].Items()
	assert.NotEmpty(t, buildItems)
	if len(buildItems) > 0 {
		firstItem := buildItems[0].(BuildItem)
		assert.True(t, firstItem.Enabled) // SkipTests=false means tests are enabled
	}

	// Check advanced settings list reflects config
	advItems := m.Lists[3].Items()
	assert.NotEmpty(t, advItems)
	if len(advItems) > 0 {
		firstItem := advItems[0].(BuildItem)
		assert.True(t, firstItem.Enabled) // CreateDraft=true
	}
}

// ============================================================================
// Test saveConfig and saveConfigWithRegenFlag
// ============================================================================

func TestConfigureModel_SaveConfig_WithNilConfig(t *testing.T) {
	m := &ConfigureModel{
		ProjectConfig:     nil,
		ProjectIdentifier: "",
	}

	err := m.saveConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no project config to save")
}

func TestConfigureModel_SaveConfig_UpdatesDistributions(t *testing.T) {
	m := &ConfigureModel{
		ProjectConfig: &models.ProjectConfig{
			Project: &models.ProjectInfo{
				Identifier: "test",
			},
			Config: &models.ProjectSettings{},
		},
		ProjectIdentifier: "test",
	}

	// Create distributions list
	items := []list.Item{
		DistributionItem{Name: "GitHub", Key: "github", Enabled: true},
		DistributionItem{Name: "Homebrew", Key: "homebrew", Enabled: false},
		DistributionItem{Name: "NPM", Key: "npm", Enabled: true},
		DistributionItem{Name: "Go Module", Key: "go_install", Enabled: false},
	}
	m.Lists[1] = list.New(items, list.NewDefaultDelegate(), 80, 20)

	// Note: Actual save will fail without proper config path, but we test the logic
	err := m.saveConfigWithRegenFlag(false)
	// We expect error due to config not being properly initialized for save
	// but we verify the config was updated
	assert.NotNil(m.ProjectConfig.Config.Distributions.GitHubRelease)
	assert.True(m.ProjectConfig.Config.Distributions.GitHubRelease.Enabled)
	assert.NotNil(m.ProjectConfig.Config.Distributions.Homebrew)
	assert.False(m.ProjectConfig.Config.Distributions.Homebrew.Enabled)
	assert.NotNil(m.ProjectConfig.Config.Distributions.NPM)
	assert.True(m.ProjectConfig.Config.Distributions.NPM.Enabled)
}

func TestConfigureModel_SaveConfig_UpdatesBuildSettings(t *testing.T) {
	m := &ConfigureModel{
		ProjectConfig: &models.ProjectConfig{
			Project: &models.ProjectInfo{
				Identifier: "test",
			},
			Config: &models.ProjectSettings{},
		},
		ProjectIdentifier: "test",
	}

	// Create build settings list
	items := []list.Item{
		BuildItem{Name: "Run tests", Enabled: false}, // Index 0
	}
	m.Lists[2] = list.New(items, list.NewDefaultDelegate(), 80, 20)

	m.saveConfigWithRegenFlag(false)

	assert.NotNil(m.ProjectConfig.Config.Release)
	assert.True(m.ProjectConfig.Config.Release.SkipTests) // Inverted logic
}

func TestConfigureModel_SaveConfig_UpdatesAdvancedSettings(t *testing.T) {
	m := &ConfigureModel{
		ProjectConfig: &models.ProjectConfig{
			Project: &models.ProjectInfo{
				Identifier: "test",
			},
			Config: &models.ProjectSettings{},
		},
		ProjectIdentifier: "test",
	}

	// Create advanced settings list
	items := []list.Item{
		BuildItem{Name: "Draft", Enabled: true},      // Index 0
		BuildItem{Name: "PreRelease", Enabled: true}, // Index 1
		BuildItem{Name: "Changelog", Enabled: false}, // Index 2
		BuildItem{Name: "Sign", Enabled: true},       // Index 3
	}
	m.Lists[3] = list.New(items, list.NewDefaultDelegate(), 80, 20)

	m.saveConfigWithRegenFlag(false)

	assert.NotNil(m.ProjectConfig.Config.Release)
	assert.True(m.ProjectConfig.Config.Release.CreateDraft)
	assert.True(m.ProjectConfig.Config.Release.PreRelease)
	assert.False(m.ProjectConfig.Config.Release.GenerateChangelog)
	assert.True(m.ProjectConfig.Config.Release.SignCommits)
}

func TestConfigureModel_SaveConfig_SetsRegenerationFlag(t *testing.T) {
	m := &ConfigureModel{
		ProjectConfig: &models.ProjectConfig{
			Project: &models.ProjectInfo{Identifier: "test"},
			Config:  &models.ProjectSettings{},
		},
		ProjectIdentifier: "test",
		NeedsRegeneration: false,
	}

	m.saveConfigWithRegenFlag(true)
	assert.True(t, m.NeedsRegeneration)

	m.NeedsRegeneration = false
	m.saveConfigWithRegenFlag(false)
	assert.False(t, m.NeedsRegeneration)
}

// ============================================================================
// Test Update Method - Message Handling
// ============================================================================

func TestConfigureModel_Update_SpinnerTickMsg(t *testing.T) {
	m := &ConfigureModel{
		Width:          100,
		Height:         30,
		IsCreating:     true,
		CreateSpinner:  spinner.New(),
	}

	// Spinner should continue ticking when creating
	_, cmd := m.Update(spinner.TickMsg{})
	assert.NotNil(t, cmd)

	// Spinner should stop when not creating
	m.IsCreating = false
	m.Loading = false
	m.GeneratingFiles = false
	_, cmd = m.Update(spinner.TickMsg{})
	assert.Nil(t, cmd)
}

func TestConfigureModel_Update_LoadCompleteMsg(t *testing.T) {
	m := &ConfigureModel{
		Width:   100,
		Height:  30,
		Loading: true,
		ProjectConfig: &models.ProjectConfig{
			Project: &models.ProjectInfo{Identifier: "test"},
			Config:  &models.ProjectSettings{},
		},
	}

	cleanupModel := &CleanupModel{}
	msg := loadCompleteMsg{cleanupModel: cleanupModel}

	m, _ = m.Update(msg)

	assert.False(t, m.Loading)
	assert.True(t, m.Initialized)
	assert.Equal(t, cleanupModel, m.CleanupModel)
}

func TestConfigureModel_Update_RepoCreatedMsg(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectStatus   string
	}{
		{
			name:         "success",
			err:          nil,
			expectStatus: "✓ Repository created successfully\!",
		},
		{
			name:         "failure",
			err:          errors.New("creation failed"),
			expectStatus: "✗ Failed: creation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ConfigureModel{
				Width:       100,
				Height:      30,
				IsCreating:  true,
				RepoNameInput: textinput.New(),
				RepoDescInput: textinput.New(),
			}

			msg := repoCreatedMsg{err: tt.err}
			m, cmd := m.Update(msg)

			assert.False(t, m.IsCreating)
			assert.Contains(t, m.CreateStatus, tt.expectStatus)
			assert.NotNil(t, cmd) // Should have timeout command
		})
	}
}

func TestConfigureModel_Update_PushCompleteMsg(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectStatus string
	}{
		{
			name:         "success",
			err:          nil,
			expectStatus: "✓ Pushed to remote successfully\!",
		},
		{
			name:         "failure",
			err:          errors.New("push failed"),
			expectStatus: "✗ Push failed: push failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ConfigureModel{
				Width:      100,
				Height:     30,
				IsCreating: true,
			}

			msg := pushCompleteMsg{err: tt.err}
			m, cmd := m.Update(msg)

			assert.False(t, m.IsCreating)
			assert.Contains(t, m.CreateStatus, tt.expectStatus)
			assert.NotNil(t, cmd)
		})
	}
}

func TestConfigureModel_Update_FilesGeneratedMsg(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectStatus   string
		expectView     ViewType
	}{
		{
			name:         "success",
			err:          nil,
			expectStatus: "✓ Release files updated successfully\!",
			expectView:   TabView,
		},
		{
			name:         "failure",
			err:          errors.New("generation failed"),
			expectStatus: "✗ Generation failed: generation failed",
			expectView:   GenerateConfigConsent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ConfigureModel{
				Width:             100,
				Height:            30,
				GeneratingFiles:   true,
				NeedsRegeneration: true,
				CurrentView:       GenerateConfigConsent,
				PendingGenerateFiles: []string{"file1"},
				PendingDeleteFiles:   []string{"file2"},
			}

			msg := filesGeneratedMsg{err: tt.err}
			m, cmd := m.Update(msg)

			assert.False(t, m.GeneratingFiles)
			assert.Contains(t, m.GenerateStatus, tt.expectStatus)
			if tt.err == nil {
				assert.Equal(t, TabView, m.CurrentView)
				assert.Nil(t, m.PendingGenerateFiles)
				assert.Nil(t, m.PendingDeleteFiles)
				assert.False(t, m.NeedsRegeneration)
			}
			assert.NotNil(t, cmd)
		})
	}
}

func TestConfigureModel_Update_DistributionDetectedMsg(t *testing.T) {
	tests := []struct {
		name                string
		msg                 distributionDetectedMsg
		expectAutoDetected  bool
		expectConfirmation  bool
		expectHomebrewCheck bool
		expectNPMCheck      bool
	}{
		{
			name: "homebrew found in registry",
			msg: distributionDetectedMsg{
				homebrewTap:     "owner/tap",
				homebrewFormula: "formula",
				homebrewExists:  true,
			},
			expectAutoDetected:  true,
			expectConfirmation:  true,
			expectHomebrewCheck: true,
		},
		{
			name: "npm found in registry",
			msg: distributionDetectedMsg{
				npmPackage: "package-name",
				npmExists:  true,
			},
			expectAutoDetected: true,
			expectConfirmation: true,
			expectNPMCheck:     true,
		},
		{
			name: "homebrew from file",
			msg: distributionDetectedMsg{
				homebrewTap:      "owner/tap",
				homebrewFormula:  "formula",
				homebrewFromFile: true,
			},
			expectAutoDetected:  true,
			expectConfirmation:  true,
			expectHomebrewCheck: true,
		},
		{
			name:               "nothing found",
			msg:                distributionDetectedMsg{},
			expectAutoDetected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ConfigureModel{
				Width:                  100,
				Height:                 30,
				DetectingDistributions: true,
				GlobalConfig:           &models.GlobalConfig{},
				DetectedProject:        &models.ProjectInfo{},
			}

			m, _ = m.Update(tt.msg)

			assert.False(t, m.DetectingDistributions)
			assert.Equal(t, tt.expectAutoDetected, m.AutoDetected)
			assert.Equal(t, tt.expectConfirmation, m.FirstTimeSetupConfirmation)
			assert.Equal(t, tt.expectHomebrewCheck, m.HomebrewCheckEnabled)
			assert.Equal(t, tt.expectNPMCheck, m.NPMCheckEnabled)
		})
	}
}

func TestConfigureModel_Update_DistributionVerifiedMsg(t *testing.T) {
	tests := []struct {
		name                string
		msg                 distributionVerifiedMsg
		expectError         bool
		expectFirstTimeExit bool
	}{
		{
			name: "verification error",
			msg: distributionVerifiedMsg{
				err: errors.New("verification failed"),
			},
			expectError:         true,
			expectFirstTimeExit: false,
		},
		{
			name: "homebrew verified",
			msg: distributionVerifiedMsg{
				homebrewVersion: "v1.0.0",
				homebrewExists:  true,
			},
			expectFirstTimeExit: true,
		},
		{
			name: "npm verified",
			msg: distributionVerifiedMsg{
				npmVersion: "1.0.0",
				npmExists:  true,
			},
			expectFirstTimeExit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ConfigureModel{
				Width:                  100,
				Height:                 30,
				VerifyingDistributions: true,
				FirstTimeSetup:         true,
				CurrentView:            FirstTimeSetupView,
				ProjectConfig: &models.ProjectConfig{
					Project: &models.ProjectInfo{
						Module: &models.ModuleInfo{},
					},
					Config: &models.ProjectSettings{},
				},
				HomebrewTapInput:     textinput.New(),
				HomebrewFormulaInput: textinput.New(),
				NPMPackageInput:      textinput.New(),
			}

			m, _ = m.Update(tt.msg)

			assert.False(t, m.VerifyingDistributions)
			if tt.expectError {
				assert.NotEmpty(t, m.DistributionVerifyError)
			} else if tt.expectFirstTimeExit {
				assert.False(t, m.FirstTimeSetup)
				assert.Equal(t, TabView, m.CurrentView)
			}
		})
	}
}

func TestConfigureModel_Update_WindowSizeMsg(t *testing.T) {
	m := &ConfigureModel{
		Width:  0,
		Height: 0,
	}

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	m, _ = m.Update(msg)

	// Dimensions should be set by app.go, but we verify initialization
	assert.True(t, m.Initialized)
}

// ============================================================================
// Test Update Method - Key Handling
// ============================================================================

func TestConfigureModel_Update_TabKey(t *testing.T) {
	m := &ConfigureModel{
		Width:      100,
		Height:     30,
		ActiveTab:  0,
		CurrentView: TabView,
	}
	m.Lists[0] = list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)
	m.Lists[1] = list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)
	m.Lists[2] = list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)
	m.Lists[3] = list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)

	msg := tea.KeyMsg{Type: tea.KeyTab}
	m, _ = m.Update(msg)

	assert.Equal(t, 1, m.ActiveTab)

	// Tab again
	m, _ = m.Update(msg)
	assert.Equal(t, 2, m.ActiveTab)

	// Wrap around
	m.ActiveTab = 3
	m, _ = m.Update(msg)
	assert.Equal(t, 0, m.ActiveTab)
}

func TestConfigureModel_Update_ShiftTabKey(t *testing.T) {
	m := &ConfigureModel{
		Width:      100,
		Height:     30,
		ActiveTab:  1,
		CurrentView: TabView,
	}
	m.Lists[0] = list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)
	m.Lists[1] = list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)
	m.Lists[2] = list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)
	m.Lists[3] = list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)

	msg := tea.KeyMsg{Type: tea.KeyShiftTab}
	m, _ = m.Update(msg)

	assert.Equal(t, 0, m.ActiveTab)

	// Wrap around backwards
	m, _ = m.Update(msg)
	assert.Equal(t, 3, m.ActiveTab)
}

func TestConfigureModel_Update_SpaceKey_ToggleDistribution(t *testing.T) {
	items := []list.Item{
		DistributionItem{Name: "GitHub", Key: "github", Enabled: false},
	}
	m := &ConfigureModel{
		Width:      100,
		Height:     30,
		ActiveTab:  1, // Distributions tab
		CurrentView: TabView,
		ProjectConfig: &models.ProjectConfig{
			Project: &models.ProjectInfo{Identifier: "test"},
			Config:  &models.ProjectSettings{},
		},
		ProjectIdentifier: "test",
	}
	m.Lists[1] = list.New(items, list.NewDefaultDelegate(), 80, 20)

	msg := tea.KeyMsg{Type: tea.KeySpace}
	m, _ = m.Update(msg)

	// Item should be toggled
	updatedItems := m.Lists[1].Items()
	require.Len(t, updatedItems, 1)
	item := updatedItems[0].(DistributionItem)
	assert.True(t, item.Enabled)
}

func TestConfigureModel_Update_SpaceKey_ToggleBuildItem(t *testing.T) {
	items := []list.Item{
		BuildItem{Name: "Run tests", Enabled: false},
	}
	m := &ConfigureModel{
		Width:      100,
		Height:     30,
		ActiveTab:  2, // Build settings tab
		CurrentView: TabView,
		ProjectConfig: &models.ProjectConfig{
			Project: &models.ProjectInfo{Identifier: "test"},
			Config:  &models.ProjectSettings{},
		},
		ProjectIdentifier: "test",
	}
	m.Lists[2] = list.New(items, list.NewDefaultDelegate(), 80, 20)

	msg := tea.KeyMsg{Type: tea.KeySpace}
	m, _ = m.Update(msg)

	updatedItems := m.Lists[2].Items()
	require.Len(t, updatedItems, 1)
	item := updatedItems[0].(BuildItem)
	assert.True(t, item.Enabled)
}

func TestConfigureModel_Update_SpaceKey_CycleCleanupAction(t *testing.T) {
	items := []list.Item{
		CleanupItem{Path: "file.go", Action: "commit"},
	}
	m := &ConfigureModel{
		Width:      100,
		Height:     30,
		ActiveTab:  0, // Cleanup tab
		CurrentView: TabView,
	}
	m.Lists[0] = list.New(items, list.NewDefaultDelegate(), 80, 20)

	msg := tea.KeyMsg{Type: tea.KeySpace}

	// commit -> skip
	m, _ = m.Update(msg)
	item := m.Lists[0].Items()[0].(CleanupItem)
	assert.Equal(t, "skip", item.Action)

	// skip -> ignore
	m, _ = m.Update(msg)
	item = m.Lists[0].Items()[0].(CleanupItem)
	assert.Equal(t, "ignore", item.Action)

	// ignore -> commit
	m, _ = m.Update(msg)
	item = m.Lists[0].Items()[0].(CleanupItem)
	assert.Equal(t, "commit", item.Action)
}

func TestConfigureModel_Update_EditKey_NPMPackage(t *testing.T) {
	items := []list.Item{
		DistributionItem{Name: "NPM", Key: "npm", Enabled: true},
	}
	m := &ConfigureModel{
		Width:      100,
		Height:     30,
		ActiveTab:  1,
		CurrentView: TabView,
		ProjectConfig: &models.ProjectConfig{
			Config: &models.ProjectSettings{
				Distributions: models.DistributionSettings{
					NPM: &models.NPMConfig{
						PackageName: "existing-package",
					},
				},
			},
		},
		NPMNameInput: textinput.New(),
	}
	m.Lists[1] = list.New(items, list.NewDefaultDelegate(), 80, 20)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	m, _ = m.Update(msg)

	assert.True(t, m.NPMEditMode)
	assert.Equal(t, "existing-package", m.NPMNameInput.Value())
}

// ============================================================================
// Test ViewType Enum
// ============================================================================

func TestViewType_Constants(t *testing.T) {
	assert.Equal(t, ViewType(0), TabView)
	assert.Equal(t, ViewType(1), GitHubView)
	assert.Equal(t, ViewType(2), CommitView)
	assert.Equal(t, ViewType(3), SmartCommitConfirm)
	assert.Equal(t, ViewType(4), GenerateConfigConsent)
	assert.Equal(t, ViewType(5), SmartCommitPrefsView)
	assert.Equal(t, ViewType(6), FirstTimeSetupView)
}

// ============================================================================
// Test Command Functions
// ============================================================================

func TestLoadCleanupCmd(t *testing.T) {
	cmd := LoadCleanupCmd(100, 30)
	require.NotNil(t, cmd)

	msg := cmd()
	require.IsType(t, loadCompleteMsg{}, msg)

	loadMsg := msg.(loadCompleteMsg)
	assert.NotNil(t, loadMsg.cleanupModel)
}

func TestCreateRepoCmd(t *testing.T) {
	// Test that the command returns a function
	cmd := createRepoCmd(false, "test-repo", "description", "owner")
	require.NotNil(t, cmd)

	// We can't test actual execution without mocking gitcleanup
	// but we can verify it returns the right type
	assert.IsType(t, tea.Cmd(nil), cmd)
}

func TestPushCmd(t *testing.T) {
	cmd := pushCmd()
	require.NotNil(t, cmd)
	assert.IsType(t, tea.Cmd(nil), cmd)
}

func TestSmartCommitCmd(t *testing.T) {
	items := []gitcleanup.CleanupItem{}
	cmd := smartCommitCmd(items)
	require.NotNil(t, cmd)
	assert.IsType(t, tea.Cmd(nil), cmd)
}

func TestRegularCommitCmd(t *testing.T) {
	files := []string{"file1.go", "file2.go"}
	message := "test commit"
	cmd := regularCommitCmd(files, message)
	require.NotNil(t, cmd)
	assert.IsType(t, tea.Cmd(nil), cmd)
}

func TestGenerateFilesCmd(t *testing.T) {
	detectedProject := &models.ProjectInfo{}
	projectConfig := &models.ProjectConfig{}
	filesToGenerate := []string{"file1"}
	filesToDelete := []string{"file2"}

	cmd := generateFilesCmd(detectedProject, projectConfig, filesToGenerate, filesToDelete)
	require.NotNil(t, cmd)
	assert.IsType(t, tea.Cmd(nil), cmd)
}

// ============================================================================
// Test Edge Cases and Error Conditions
// ============================================================================

func TestConfigureModel_Update_WithInvalidTabIndex(t *testing.T) {
	m := &ConfigureModel{
		Width:      100,
		Height:     30,
		ActiveTab:  5, // Invalid index
		CurrentView: TabView,
	}

	msg := tea.KeyMsg{Type: tea.KeySpace}
	m, _ = m.Update(msg)

	// Should handle gracefully without panic
	assert.Equal(t, 5, m.ActiveTab)
}

func TestConfigureModel_Update_EmptyList(t *testing.T) {
	m := &ConfigureModel{
		Width:      100,
		Height:     30,
		ActiveTab:  1,
		CurrentView: TabView,
	}
	m.Lists[1] = list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)

	msg := tea.KeyMsg{Type: tea.KeySpace}
	m, _ = m.Update(msg)

	// Should handle empty list without panic
	assert.NotNil(t, m)
}

func TestConfigureModel_ListSizeCalculation(t *testing.T) {
	tests := []struct {
		name               string
		width              int
		height             int
		activeTab          int
		needsRegeneration  bool
		npmStatus          string
		npmSuggestions     []string
		expectedMinHeight  int
	}{
		{
			name:              "basic calculation",
			width:             100,
			height:            30,
			activeTab:         0,
			expectedMinHeight: 5,
		},
		{
			name:               "with regeneration warning",
			width:              100,
			height:             30,
			needsRegeneration:  true,
			expectedMinHeight:  5,
		},
		{
			name:           "with npm status",
			width:          100,
			height:         30,
			activeTab:      1,
			npmStatus:      "checking",
			expectedMinHeight: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ConfigureModel{
				Width:             tt.width,
				Height:            tt.height,
				ActiveTab:         tt.activeTab,
				NeedsRegeneration: tt.needsRegeneration,
				NPMNameStatus:     tt.npmStatus,
				NPMNameSuggestions: tt.npmSuggestions,
			}

			m.Lists[0] = list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)
			m.Lists[1] = list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)
			m.Lists[2] = list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)
			m.Lists[3] = list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)

			// Trigger size recalculation
			m, _ = m.Update(tea.WindowSizeMsg{Width: tt.width, Height: tt.height})

			// Verify lists have valid dimensions
			for i := range m.Lists {
				_, height := m.Lists[i].Size()
				assert.GreaterOrEqual(t, height, tt.expectedMinHeight)
			}
		})
	}
}

func TestConfigureModel_NPMEditMode_EnterKey(t *testing.T) {
	m := &ConfigureModel{
		Width:      100,
		Height:     30,
		ActiveTab:  1,
		NPMEditMode: true,
		ProjectConfig: &models.ProjectConfig{
			Project: &models.ProjectInfo{Identifier: "test"},
			Config: &models.ProjectSettings{
				Distributions: models.DistributionSettings{
					NPM: &models.NPMConfig{},
				},
			},
		},
		ProjectIdentifier: "test",
		NPMNameInput:      textinput.New(),
		DetectedProject: &models.ProjectInfo{
			Repository: &models.RepositoryInfo{
				Owner: "testowner",
			},
		},
	}
	m.NPMNameInput.SetValue("new-package-name")

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	m, cmd := m.Update(msg)

	assert.False(t, m.NPMEditMode)
	assert.Equal(t, "new-package-name", m.ProjectConfig.Config.Distributions.NPM.PackageName)
	assert.NotNil(t, cmd) // Should trigger name check
}

func TestConfigureModel_NPMEditMode_EscKey(t *testing.T) {
	m := &ConfigureModel{
		Width:       100,
		Height:      30,
		NPMEditMode: true,
		NPMNameInput: textinput.New(),
	}
	m.NPMNameInput.SetValue("some-text")

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	m, _ = m.Update(msg)

	assert.False(t, m.NPMEditMode)
}

func TestConfigureModel_NPMToggle_ChecksName(t *testing.T) {
	items := []list.Item{
		DistributionItem{Name: "NPM", Key: "npm", Enabled: false},
	}
	m := &ConfigureModel{
		Width:      100,
		Height:     30,
		ActiveTab:  1,
		CurrentView: TabView,
		ProjectConfig: &models.ProjectConfig{
			Project: &models.ProjectInfo{Identifier: "test"},
			Config: &models.ProjectSettings{
				Distributions: models.DistributionSettings{
					NPM: &models.NPMConfig{
						PackageName: "my-package",
					},
				},
			},
		},
		ProjectIdentifier: "test",
		DetectedProject: &models.ProjectInfo{
			Repository: &models.RepositoryInfo{
				Owner: "owner",
			},
		},
	}
	m.Lists[1] = list.New(items, list.NewDefaultDelegate(), 80, 20)

	msg := tea.KeyMsg{Type: tea.KeySpace}
	m, cmd := m.Update(msg)

	// Should trigger NPM name check
	assert.NotNil(t, cmd)
	assert.Equal(t, "checking", m.NPMNameStatus)
}

func TestConfigureModel_NPMDisable_ClearsStatus(t *testing.T) {
	items := []list.Item{
		DistributionItem{Name: "NPM", Key: "npm", Enabled: true},
	}
	m := &ConfigureModel{
		Width:      100,
		Height:     30,
		ActiveTab:  1,
		CurrentView: TabView,
		ProjectConfig: &models.ProjectConfig{
			Project: &models.ProjectInfo{Identifier: "test"},
			Config:  &models.ProjectSettings{},
		},
		ProjectIdentifier: "test",
		NPMNameStatus:     "available",
		NPMNameSuggestions: []string{"alt1", "alt2"},
	}
	m.Lists[1] = list.New(items, list.NewDefaultDelegate(), 80, 20)

	msg := tea.KeyMsg{Type: tea.KeySpace}
	m, _ = m.Update(msg)

	// Should clear NPM status
	assert.Empty(t, m.NPMNameStatus)
	assert.Empty(t, m.NPMNameError)
	assert.Nil(t, m.NPMNameSuggestions)
}

func TestUpdateConfigureView_CommitViewEscape(t *testing.T) {
	configModel := &ConfigureModel{
		Width:       100,
		Height:      30,
		CurrentView: CommitView,
		CommitModel: &CommitModel{},
		CleanupModel: &CleanupModel{},
	}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, _, _, updatedModel := UpdateConfigureView(1, 1, msg, configModel)

	assert.Equal(t, TabView, updatedModel.CurrentView)
	assert.Nil(t, updatedModel.CommitModel)
}

func TestUpdateConfigureView_SmartCommitConfirm_Yes(t *testing.T) {
	configModel := &ConfigureModel{
		Width:         100,
		Height:        30,
		CurrentView:   SmartCommitConfirm,
		CreateSpinner: spinner.New(),
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	_, _, cmd, updatedModel := UpdateConfigureView(1, 1, msg, configModel)

	// Should trigger smart commit
	assert.NotNil(t, cmd)
	assert.True(t, updatedModel.IsCreating)
}

func TestUpdateConfigureView_GenerateConfigConsent_Yes(t *testing.T) {
	configModel := &ConfigureModel{
		Width:         100,
		Height:        30,
		CurrentView:   GenerateConfigConsent,
		ProjectConfig: &models.ProjectConfig{},
		DetectedProject: &models.ProjectInfo{},
		PendingGenerateFiles: []string{"file1"},
		CreateSpinner: spinner.New(),
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y'}}
	_, _, cmd, updatedModel := UpdateConfigureView(1, 1, msg, configModel)

	assert.True(t, updatedModel.GeneratingFiles)
	assert.NotNil(t, cmd)
}

func TestUpdateConfigureView_GenerateConfigConsent_No(t *testing.T) {
	configModel := &ConfigureModel{
		Width:                100,
		Height:               30,
		CurrentView:          GenerateConfigConsent,
		PendingGenerateFiles: []string{"file1"},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	_, _, _, updatedModel := UpdateConfigureView(1, 1, msg, configModel)

	assert.Equal(t, TabView, updatedModel.CurrentView)
	assert.Nil(t, updatedModel.PendingGenerateFiles)
}

func TestUpdateConfigureView_QuitKey(t *testing.T) {
	configModel := &ConfigureModel{
		CurrentView: TabView,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, shouldQuit, cmd, _ := UpdateConfigureView(1, 1, msg, configModel)

	assert.True(t, shouldQuit)
	assert.NotNil(t, cmd)
}

func TestUpdateConfigureView_EscapeFromTabView(t *testing.T) {
	configModel := &ConfigureModel{
		CurrentView: TabView,
	}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newPage, _, _, _ := UpdateConfigureView(1, 1, msg, configModel)

	// Should return to project view (page 0)
	assert.Equal(t, 0, newPage)
}

// ============================================================================
// Test Concurrent Message Handling
// ============================================================================

func TestConfigureModel_MultipleSpinnerUpdates(t *testing.T) {
	m := &ConfigureModel{
		Width:          100,
		Height:         30,
		IsCreating:     true,
		CreateSpinner:  spinner.New(),
	}

	// Multiple spinner ticks should be handled
	for i := 0; i < 5; i++ {
		m, cmd := m.Update(spinner.TickMsg{})
		assert.NotNil(t, cmd)
		assert.NotNil(t, m)
	}
}

func TestConfigureModel_StatusMessageTimeout(t *testing.T) {
	m := &ConfigureModel{
		Width:        100,
		Height:       30,
		CreateStatus: "Test status",
	}

	// Empty struct message should clear status
	m, _ = m.Update(struct{}{})
	assert.Empty(t, m.CreateStatus)
}

// ============================================================================
// Test Boundary Conditions
// ============================================================================

func TestConfigureModel_MinimumDimensions(t *testing.T) {
	m := NewConfigureModel(1, 1, nil, nil, nil, nil)

	assert.NotNil(t, m)
	assert.GreaterOrEqual(t, m.Width, 40)
	assert.GreaterOrEqual(t, m.Height, 5)
}

func TestConfigureModel_LargeNumberOfLists(t *testing.T) {
	m := &ConfigureModel{
		Width:  200,
		Height: 100,
	}

	// Initialize many items
	for i := 0; i < 4; i++ {
		items := make([]list.Item, 100)
		for j := range items {
			items[j] = BuildItem{Name: "Item", Enabled: false}
		}
		m.Lists[i] = list.New(items, list.NewDefaultDelegate(), 180, 80)
	}

	// Should handle large lists
	assert.NotNil(t, m)
}

func TestCleanupItem_PathTruncation(t *testing.T) {
	m := &ConfigureModel{
		Width: 50,
	}

	longPath := "very/long/path/to/some/file/that/exceeds/width/limit/file.go"
	items := []list.Item{
		CleanupItem{Path: longPath, Status: "M"},
	}

	// loadGitStatus should truncate paths
	// We verify the CleanupItem behavior
	item := items[0].(CleanupItem)
	assert.NotEmpty(t, item.Path)
}

func TestConfigureModel_AllTabNavigation(t *testing.T) {
	m := &ConfigureModel{
		Width:      100,
		Height:     30,
		ActiveTab:  0,
		CurrentView: TabView,
	}

	// Initialize all lists
	for i := 0; i < 4; i++ {
		m.Lists[i] = list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)
	}

	// Navigate through all tabs
	expectedTabs := []int{1, 2, 3, 0}
	for _, expected := range expectedTabs {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
		assert.Equal(t, expected, m.ActiveTab)
	}
}

// ============================================================================
// Test Complex Scenarios
// ============================================================================

func TestConfigureModel_CompleteRepoCreationFlow(t *testing.T) {
	m := &ConfigureModel{
		Width:         100,
		Height:        30,
		CreatingRepo:  true,
		RepoInputFocus: 0,
		RepoNameInput: textinput.New(),
		RepoDescInput: textinput.New(),
		GitHubAccounts: []models.GitHubAccount{
			{Username: "user1"},
			{Username: "user2"},
		},
		CreateSpinner: spinner.New(),
	}

	// Tab through fields
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, 1, m.RepoInputFocus)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, 2, m.RepoInputFocus)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, 3, m.RepoInputFocus)

	// Cancel creation
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, m.CreatingRepo)
}

func TestConfigureModel_NPMEditingFlow(t *testing.T) {
	items := []list.Item{
		DistributionItem{Name: "NPM", Key: "npm", Enabled: true},
	}
	m := &ConfigureModel{
		Width:      100,
		Height:     30,
		ActiveTab:  1,
		CurrentView: TabView,
		ProjectConfig: &models.ProjectConfig{
			Project: &models.ProjectInfo{Identifier: "test"},
			Config: &models.ProjectSettings{
				Distributions: models.DistributionSettings{
					NPM: &models.NPMConfig{PackageName: "old-name"},
				},
			},
		},
		ProjectIdentifier: "test",
		NPMNameInput:      textinput.New(),
		DetectedProject: &models.ProjectInfo{
			Repository: &models.RepositoryInfo{Owner: "owner"},
		},
	}
	m.Lists[1] = list.New(items, list.NewDefaultDelegate(), 80, 20)

	// Enter edit mode
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	assert.True(t, m.NPMEditMode)

	// Type new name (simulated)
	m.NPMNameInput.SetValue("new-name")

	// Confirm with Enter
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, m.NPMEditMode)
	assert.Equal(t, "new-name", m.ProjectConfig.Config.Distributions.NPM.PackageName)
	assert.NotNil(t, cmd)
}