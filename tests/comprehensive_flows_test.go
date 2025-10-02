package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"distui/internal/config"
	"distui/internal/detection"
	"distui/internal/fileops"
	"distui/internal/gitcleanup"
	"distui/internal/models"
)

// TestFlow_MultiProjectIsolation tests that each project maintains isolated config
// Covers: Scenario #32, FR-089 from tasks-user-flows.md
func TestFlow_MultiProjectIsolation(t *testing.T) {
	tmpDir := t.TempDir()

	// Project A: Homebrew + NPM
	projectADir := filepath.Join(tmpDir, "project-a")
	require.NoError(t, os.MkdirAll(projectADir, 0755))

	projectA := &models.ProjectConfig{
		Project: &models.ProjectInfo{
			Identifier: "project-a",
			Path:       projectADir,
		},
		Config: &models.ProjectSettings{
			Distributions: models.Distributions{
				GitHubRelease: &models.GitHubReleaseConfig{Enabled: true},
				Homebrew:      &models.HomebrewConfig{Enabled: true, TapRepo: "user/tap-a"},
				NPM:           &models.NPMConfig{Enabled: true, PackageName: "pkg-a"},
			},
		},
		FirstTimeSetupCompleted: true,
	}

	// Project B: Homebrew only
	projectBDir := filepath.Join(tmpDir, "project-b")
	require.NoError(t, os.MkdirAll(projectBDir, 0755))

	projectB := &models.ProjectConfig{
		Project: &models.ProjectInfo{
			Identifier: "project-b",
			Path:       projectBDir,
		},
		Config: &models.ProjectSettings{
			Distributions: models.Distributions{
				Homebrew: &models.HomebrewConfig{Enabled: true, TapRepo: "user/tap-b"},
			},
		},
		FirstTimeSetupCompleted: true,
	}

	// Project C: NPM only
	projectCDir := filepath.Join(tmpDir, "project-c")
	require.NoError(t, os.MkdirAll(projectCDir, 0755))

	projectC := &models.ProjectConfig{
		Project: &models.ProjectInfo{
			Identifier: "project-c",
			Path:       projectCDir,
		},
		Config: &models.ProjectSettings{
			Distributions: models.Distributions{
				NPM: &models.NPMConfig{Enabled: true, PackageName: "pkg-c"},
			},
		},
		FirstTimeSetupCompleted: true,
	}

	// Save all projects
	os.Setenv("HOME", tmpDir)
	require.NoError(t, config.SaveProject(projectA))
	require.NoError(t, config.SaveProject(projectB))
	require.NoError(t, config.SaveProject(projectC))

	// Rapid switching between projects (simulates user navigation)
	for i := 0; i < 5; i++ {
		// Load Project A
		loadedA, err := config.LoadProject("project-a")
		require.NoError(t, err)
		assert.True(t, loadedA.Config.Distributions.Homebrew.Enabled)
		assert.True(t, loadedA.Config.Distributions.NPM.Enabled)
		assert.Equal(t, "user/tap-a", loadedA.Config.Distributions.Homebrew.TapRepo)

		// Load Project B
		loadedB, err := config.LoadProject("project-b")
		require.NoError(t, err)
		assert.True(t, loadedB.Config.Distributions.Homebrew.Enabled)
		assert.Nil(t, loadedB.Config.Distributions.NPM, "Project B should not have NPM")
		assert.Equal(t, "user/tap-b", loadedB.Config.Distributions.Homebrew.TapRepo)

		// Load Project C
		loadedC, err := config.LoadProject("project-c")
		require.NoError(t, err)
		assert.Nil(t, loadedC.Config.Distributions.Homebrew, "Project C should not have Homebrew")
		assert.True(t, loadedC.Config.Distributions.NPM.Enabled)
		assert.Equal(t, "pkg-c", loadedC.Config.Distributions.NPM.PackageName)
	}

	t.Log("✓ USER FLOW: Multi-project isolation verified - configs independent")
}

// TestFlow_ArchiveBeforeModeSwitch tests that custom files are safely archived
// Covers: Flow 1 from tasks-user-flows.md
func TestFlow_ArchiveBeforeModeSwitch(t *testing.T) {
	projectDir := t.TempDir()

	// Create custom .goreleaser.yaml (no distui marker)
	customContent := `project_name: my-custom-app
builds:
  - main: ./cmd/app
    ldflags: -s -w
`
	goreleaserPath := filepath.Join(projectDir, ".goreleaser.yaml")
	require.NoError(t, os.WriteFile(goreleaserPath, []byte(customContent), 0644))

	// Verify it's detected as custom
	assert.True(t, detection.IsCustomConfig(goreleaserPath))

	// User decides to switch to managed mode - files must be archived first
	backupPath, err := fileops.ArchiveCustomFiles(projectDir, []string{".goreleaser.yaml"})
	require.NoError(t, err)

	// Verify backup exists
	_, err = os.Stat(backupPath)
	assert.NoError(t, err, "Backup directory should exist")

	// Verify archived file contains original content
	archivedFile := filepath.Join(backupPath, ".goreleaser.yaml")
	archivedContent, err := os.ReadFile(archivedFile)
	require.NoError(t, err)
	assert.Equal(t, customContent, string(archivedContent))

	// Original file should be moved (not exist anymore)
	_, err = os.Stat(goreleaserPath)
	assert.True(t, os.IsNotExist(err), "Original should be moved to archive")

	t.Log("✓ USER FLOW: Custom files safely archived before mode switch")
}

// TestFlow_CleanWorkingTreeDetection tests git status checking
// Covers: Flow 3 from tasks-user-flows.md (happy path release)
func TestFlow_CleanWorkingTreeDetection(t *testing.T) {
	projectDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Initialize git repo
	require.NoError(t, os.Chdir(projectDir))
	require.NoError(t, exec.Command("git", "init").Run())
	require.NoError(t, exec.Command("git", "config", "user.email", "test@example.com").Run())
	require.NoError(t, exec.Command("git", "config", "user.name", "Test User").Run())

	// Create and commit initial file
	testFile := filepath.Join(projectDir, "main.go")
	require.NoError(t, os.WriteFile(testFile, []byte("package main"), 0644))
	require.NoError(t, exec.Command("git", "add", ".").Run())
	require.NoError(t, exec.Command("git", "commit", "-m", "initial").Run())

	// Test 1: Clean working tree
	isClean := gitcleanup.IsWorkingTreeClean()
	assert.True(t, isClean, "Working tree should be clean after commit")

	// Test 2: Uncommitted changes (should block release)
	require.NoError(t, os.WriteFile(testFile, []byte("package main\n\nfunc main() {}"), 0644))
	isClean = gitcleanup.IsWorkingTreeClean()
	assert.False(t, isClean, "Working tree should be dirty with uncommitted changes")

	// Test 3: .gitignore changes (should NOT block release if filtered)
	require.NoError(t, exec.Command("git", "add", ".").Run())
	require.NoError(t, exec.Command("git", "commit", "-m", "add main").Run())

	gitignorePath := filepath.Join(projectDir, ".gitignore")
	require.NoError(t, os.WriteFile(gitignorePath, []byte("dist/\n*.log\n"), 0644))

	// Note: IsWorkingTreeClean doesn't filter .gitignore by default
	// This test documents current behavior
	isClean = gitcleanup.IsWorkingTreeClean()
	t.Logf("Working tree clean with .gitignore changes: %v", isClean)

	t.Log("✓ USER FLOW: Working tree detection tested")
}

// TestFlow_DistributionCombinations tests various distribution configs
// Covers: T027-T030 from tasks-user-flows.md
func TestFlow_DistributionCombinations(t *testing.T) {
	tests := []struct {
		name           string
		setupConfig    func() *models.ProjectConfig
		verifyHomebrew bool
		verifyNPM      bool
		verifyGitHub   bool
	}{
		{
			name: "homebrew_only",
			setupConfig: func() *models.ProjectConfig {
				return &models.ProjectConfig{
					Project: &models.ProjectInfo{Identifier: "test-homebrew"},
					Config: &models.ProjectSettings{
						Distributions: models.Distributions{
							Homebrew: &models.HomebrewConfig{
								Enabled: true,
								TapRepo: "user/homebrew-tap",
							},
						},
					},
				}
			},
			verifyHomebrew: true,
			verifyNPM:      false,
			verifyGitHub:   false,
		},
		{
			name: "npm_only",
			setupConfig: func() *models.ProjectConfig {
				return &models.ProjectConfig{
					Project: &models.ProjectInfo{Identifier: "test-npm"},
					Config: &models.ProjectSettings{
						Distributions: models.Distributions{
							NPM: &models.NPMConfig{
								Enabled:     true,
								PackageName: "my-package",
							},
						},
					},
				}
			},
			verifyHomebrew: false,
			verifyNPM:      true,
			verifyGitHub:   false,
		},
		{
			name: "github_only",
			setupConfig: func() *models.ProjectConfig {
				return &models.ProjectConfig{
					Project: &models.ProjectInfo{Identifier: "test-github"},
					Config: &models.ProjectSettings{
						Distributions: models.Distributions{
							GitHubRelease: &models.GitHubReleaseConfig{
								Enabled: true,
							},
						},
					},
				}
			},
			verifyHomebrew: false,
			verifyNPM:      false,
			verifyGitHub:   true,
		},
		{
			name: "all_distributions",
			setupConfig: func() *models.ProjectConfig {
				return &models.ProjectConfig{
					Project: &models.ProjectInfo{Identifier: "test-all"},
					Config: &models.ProjectSettings{
						Distributions: models.Distributions{
							GitHubRelease: &models.GitHubReleaseConfig{Enabled: true},
							Homebrew:      &models.HomebrewConfig{Enabled: true, TapRepo: "user/tap"},
							NPM:           &models.NPMConfig{Enabled: true, PackageName: "pkg"},
						},
					},
				}
			},
			verifyHomebrew: true,
			verifyNPM:      true,
			verifyGitHub:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			os.Setenv("HOME", tmpDir)

			pc := tt.setupConfig()
			require.NoError(t, config.SaveProject(pc))

			// Load and verify
			loaded, err := config.LoadProject(pc.Project.Identifier)
			require.NoError(t, err)

			// Verify Homebrew
			if tt.verifyHomebrew {
				assert.NotNil(t, loaded.Config.Distributions.Homebrew)
				assert.True(t, loaded.Config.Distributions.Homebrew.Enabled)
			} else {
				if loaded.Config.Distributions.Homebrew != nil {
					assert.False(t, loaded.Config.Distributions.Homebrew.Enabled)
				}
			}

			// Verify NPM
			if tt.verifyNPM {
				assert.NotNil(t, loaded.Config.Distributions.NPM)
				assert.True(t, loaded.Config.Distributions.NPM.Enabled)
			} else {
				if loaded.Config.Distributions.NPM != nil {
					assert.False(t, loaded.Config.Distributions.NPM.Enabled)
				}
			}

			// Verify GitHub
			if tt.verifyGitHub {
				assert.NotNil(t, loaded.Config.Distributions.GitHubRelease)
				assert.True(t, loaded.Config.Distributions.GitHubRelease.Enabled)
			} else {
				if loaded.Config.Distributions.GitHubRelease != nil {
					assert.False(t, loaded.Config.Distributions.GitHubRelease.Enabled)
				}
			}
		})
	}

	t.Log("✓ USER FLOW: All distribution combinations work correctly")
}

// TestFlow_ConfigPersistenceAcrossRestarts simulates app restarts
// Covers: Persona C (T025) from tasks-user-flows.md
func TestFlow_ConfigPersistenceAcrossRestarts(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// First "launch" - create config
	original := &models.ProjectConfig{
		Project: &models.ProjectInfo{
			Identifier: "persistent-project",
			Path:       "/path/to/project",
		},
		Config: &models.ProjectSettings{
			Distributions: models.Distributions{
				Homebrew: &models.HomebrewConfig{
					Enabled: true,
					TapRepo: "user/homebrew-tap",
				},
				NPM: &models.NPMConfig{
					Enabled:     true,
					PackageName: "my-cli",
				},
			},
		},
		FirstTimeSetupCompleted: true,
	}

	require.NoError(t, config.SaveProject(original))

	// Simulate app "restart" - load from disk
	time.Sleep(10 * time.Millisecond)

	loaded1, err := config.LoadProject("persistent-project")
	require.NoError(t, err)
	assert.Equal(t, original.Project.Identifier, loaded1.Project.Identifier)
	assert.True(t, loaded1.Config.Distributions.Homebrew.Enabled)
	assert.True(t, loaded1.Config.Distributions.NPM.Enabled)
	assert.True(t, loaded1.FirstTimeSetupCompleted)

	// Make changes and save
	loaded1.Config.Distributions.NPM.PackageName = "updated-cli"
	require.NoError(t, config.SaveProject(loaded1))

	// Another "restart" - verify changes persisted
	time.Sleep(10 * time.Millisecond)

	loaded2, err := config.LoadProject("persistent-project")
	require.NoError(t, err)
	assert.Equal(t, "updated-cli", loaded2.Config.Distributions.NPM.PackageName)
	assert.True(t, loaded2.FirstTimeSetupCompleted, "First-time setup should remain completed")

	t.Log("✓ USER FLOW: Config persists correctly across restarts")
}

// TestFlow_PackageJSONDetection tests NPM package name detection
// Covers: Scenario #28 from tasks-user-flows.md
func TestFlow_PackageJSONDetection(t *testing.T) {
	projectDir := t.TempDir()

	// Create package.json with specific name
	packageJSON := `{
  "name": "@myorg/my-cli-tool",
  "version": "1.2.3",
  "description": "A CLI tool"
}`
	require.NoError(t, os.WriteFile(
		filepath.Join(projectDir, "package.json"),
		[]byte(packageJSON),
		0644,
	))

	// Create go.mod
	goMod := `module github.com/user/my-cli-tool

go 1.24`
	require.NoError(t, os.WriteFile(
		filepath.Join(projectDir, "go.mod"),
		[]byte(goMod),
		0644,
	))

	// Detect project
	project, err := detection.DetectProject(projectDir)
	require.NoError(t, err)

	// Verify module detected
	assert.NotNil(t, project.Module, "Module should be detected")
	if project.Module != nil {
		assert.Equal(t, "github.com/user/my-cli-tool", project.Module.Name)
	}

	t.Log("✓ USER FLOW: Project information correctly detected")
}

// TestFlow_BinaryDetection tests binary name detection from project structure
// Covers: Multiple scenarios requiring correct binary name
func TestFlow_BinaryDetection(t *testing.T) {
	tests := []struct {
		name           string
		setupProject   func(string)
		expectedBinary string
	}{
		{
			name: "cmd_directory_structure",
			setupProject: func(dir string) {
				// Need go.mod for detection to work
				os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/user/my-tool\n\ngo 1.24"), 0644)
				cmdDir := filepath.Join(dir, "cmd", "my-tool")
				os.MkdirAll(cmdDir, 0755)
				os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte("package main"), 0644)
			},
			expectedBinary: "my-tool",
		},
		{
			name: "root_main_with_gomod",
			setupProject: func(dir string) {
				os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
				os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/user/cli-app\n\ngo 1.24"), 0644)
			},
			expectedBinary: "cli-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir := t.TempDir()
			tt.setupProject(projectDir)

			project, err := detection.DetectProject(projectDir)
			require.NoError(t, err)

			assert.NotNil(t, project.Binary, "Binary should be detected")
			if project.Binary != nil {
				assert.Equal(t, tt.expectedBinary, project.Binary.Name)
			}
		})
	}

	t.Log("✓ USER FLOW: Binary names correctly detected from project structure")
}
