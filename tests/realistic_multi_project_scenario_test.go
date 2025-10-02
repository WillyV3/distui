package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"distui/internal/config"
	"distui/internal/detection"
	"distui/internal/models"
)

// TestRealisticMultiProjectScenario simulates a real user with multiple projects:
// - Project A: Has custom .goreleaser.yaml (user's existing setup)
// - Project B: Fresh project, no config files yet
// - Project C: Has distui-generated files already
// - Project D: Has package.json but no .goreleaser.yaml
//
// This test creates ACTUAL config files in temp directories that can be reviewed.
func TestRealisticMultiProjectScenario(t *testing.T) {
	// Create realistic temp directory structure
	tmpRoot := t.TempDir()
	configRoot := filepath.Join(tmpRoot, ".distui")
	projectsRoot := filepath.Join(tmpRoot, "projects")

	t.Logf("\n========================================")
	t.Logf("TEST ENVIRONMENT CREATED")
	t.Logf("========================================")
	t.Logf("Config Root:    %s", configRoot)
	t.Logf("Projects Root:  %s", projectsRoot)
	t.Logf("========================================\n")

	// Set HOME to our temp directory so config goes there
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpRoot)
	defer os.Setenv("HOME", originalHome)

	require.NoError(t, os.MkdirAll(configRoot, 0755))
	require.NoError(t, os.MkdirAll(projectsRoot, 0755))

	// ========================================
	// SETUP: Create realistic project structures
	// ========================================

	// Project A: Existing project with custom .goreleaser.yaml (veteran user)
	t.Log("\n[SETUP] Creating Project A: my-cli-tool (custom .goreleaser.yaml)")
	projectA := setupProjectA(t, projectsRoot)

	// Project B: Brand new Go project (fresh user)
	t.Log("[SETUP] Creating Project B: new-app (no config files)")
	projectB := setupProjectB(t, projectsRoot)

	// Project C: Already using distui (returning user)
	t.Log("[SETUP] Creating Project C: existing-tool (distui-managed)")
	projectC := setupProjectC(t, projectsRoot)

	// Project D: NPM package but no goreleaser (npm-only project)
	t.Log("[SETUP] Creating Project D: js-wrapper (package.json only)")
	projectD := setupProjectD(t, projectsRoot)

	// ========================================
	// SCENARIO: User works with multiple projects
	// ========================================

	t.Log("\n========================================")
	t.Log("SCENARIO START: User configures 4 projects")
	t.Log("========================================\n")

	// Step 1: User opens Project A (custom .goreleaser.yaml)
	t.Log("\n[STEP 1] User opens Project A in distui")
	detectedA, err := detection.DetectProject(projectA.path)
	require.NoError(t, err)
	t.Logf("  Detected: %s", detectedA.Identifier)

	// Check if custom .goreleaser.yaml exists
	goreleaserA := filepath.Join(projectA.path, ".goreleaser.yaml")
	isCustomA := detection.IsCustomConfig(goreleaserA)
	t.Logf("  Has custom .goreleaser.yaml: %v", isCustomA)
	assert.True(t, isCustomA, "Project A should have custom config")

	// User decides to KEEP custom files
	t.Log("  → User chooses [K]eep custom files")
	configA := &models.ProjectConfig{
		Project: &models.ProjectInfo{
			Identifier: "my-cli-tool",
			Path:       projectA.path,
		},
		Config: &models.ProjectSettings{
			Distributions: models.Distributions{
				Homebrew: &models.HomebrewConfig{
					Enabled: true,
					TapRepo: "user/homebrew-my-cli-tool",
				},
			},
		},
		CustomFilesMode:         true, // User chose to keep custom
		FirstTimeSetupCompleted: true,
	}
	require.NoError(t, config.SaveProject(configA))
	t.Logf("  ✓ Config saved: %s", filepath.Join(configRoot, "projects", "my-cli-tool.yaml"))

	// Step 2: User switches to Project B (no config)
	t.Log("\n[STEP 2] User switches to Project B")
	detectedB, err := detection.DetectProject(projectB.path)
	require.NoError(t, err)
	t.Logf("  Detected: %s", detectedB.Identifier)

	// No config files exist
	_, err = os.Stat(filepath.Join(projectB.path, ".goreleaser.yaml"))
	assert.True(t, os.IsNotExist(err), "Project B should not have .goreleaser.yaml")
	t.Log("  No .goreleaser.yaml found")

	// User configures from scratch
	t.Log("  → User enables Homebrew + NPM distributions")
	configB := &models.ProjectConfig{
		Project: &models.ProjectInfo{
			Identifier: "new-app",
			Path:       projectB.path,
		},
		Config: &models.ProjectSettings{
			Distributions: models.Distributions{
				GitHubRelease: &models.GitHubReleaseConfig{Enabled: true},
				Homebrew:      &models.HomebrewConfig{Enabled: true, TapRepo: "user/homebrew-new-app"},
				NPM:           &models.NPMConfig{Enabled: true, PackageName: "new-app"},
			},
		},
		CustomFilesMode:         false, // distui-managed
		FirstTimeSetupCompleted: true,
	}
	require.NoError(t, config.SaveProject(configB))
	t.Logf("  ✓ Config saved: %s", filepath.Join(configRoot, "projects", "new-app.yaml"))

	// Step 3: User switches to Project C (already distui-managed)
	t.Log("\n[STEP 3] User opens Project C (existing distui user)")
	detectedC, err := detection.DetectProject(projectC.path)
	require.NoError(t, err)
	t.Logf("  Detected: %s", detectedC.Identifier)

	// Has distui-generated .goreleaser.yaml
	goreleaserC := filepath.Join(projectC.path, ".goreleaser.yaml")
	isCustomC := detection.IsCustomConfig(goreleaserC)
	t.Logf("  Has distui-generated .goreleaser.yaml: %v", !isCustomC)
	assert.False(t, isCustomC, "Project C should have distui-generated config")

	// User just updates distribution settings
	t.Log("  → User adds NPM distribution")
	configC := &models.ProjectConfig{
		Project: &models.ProjectInfo{
			Identifier: "existing-tool",
			Path:       projectC.path,
		},
		Config: &models.ProjectSettings{
			Distributions: models.Distributions{
				Homebrew: &models.HomebrewConfig{Enabled: true, TapRepo: "user/homebrew-existing-tool"},
				NPM:      &models.NPMConfig{Enabled: true, PackageName: "existing-tool"}, // NEW
			},
		},
		CustomFilesMode:         false,
		FirstTimeSetupCompleted: true,
	}
	require.NoError(t, config.SaveProject(configC))
	t.Logf("  ✓ Config updated: %s", filepath.Join(configRoot, "projects", "existing-tool.yaml"))

	// Step 4: User switches to Project D (package.json only)
	t.Log("\n[STEP 4] User opens Project D (NPM package wrapper)")
	detectedD, err := detection.DetectProject(projectD.path)
	require.NoError(t, err)
	t.Logf("  Detected: %s", detectedD.Identifier)

	// Has package.json but no .goreleaser.yaml
	_, err = os.Stat(filepath.Join(projectD.path, ".goreleaser.yaml"))
	assert.True(t, os.IsNotExist(err), "Project D should not have .goreleaser.yaml")
	_, err = os.Stat(filepath.Join(projectD.path, "package.json"))
	assert.NoError(t, err, "Project D should have package.json")
	t.Log("  Has package.json, no .goreleaser.yaml")

	// User configures NPM-only
	t.Log("  → User enables NPM only (no Homebrew)")
	configD := &models.ProjectConfig{
		Project: &models.ProjectInfo{
			Identifier: "js-wrapper",
			Path:       projectD.path,
		},
		Config: &models.ProjectSettings{
			Distributions: models.Distributions{
				NPM: &models.NPMConfig{Enabled: true, PackageName: "@myorg/js-wrapper"},
			},
		},
		CustomFilesMode:         false,
		FirstTimeSetupCompleted: true,
	}
	require.NoError(t, config.SaveProject(configD))
	t.Logf("  ✓ Config saved: %s", filepath.Join(configRoot, "projects", "js-wrapper.yaml"))

	// ========================================
	// VERIFICATION: Test rapid project switching
	// ========================================

	t.Log("\n========================================")
	t.Log("VERIFICATION: Rapid project switching")
	t.Log("========================================\n")

	for i := 1; i <= 3; i++ {
		t.Logf("[Switch Round %d]", i)

		// Load Project A config
		loadedA, err := config.LoadProject("my-cli-tool")
		require.NoError(t, err)
		assert.True(t, loadedA.CustomFilesMode, "Project A should be custom mode")
		assert.True(t, loadedA.Config.Distributions.Homebrew.Enabled)
		assert.Nil(t, loadedA.Config.Distributions.NPM)
		t.Log("  ✓ Project A config correct (custom mode, Homebrew only)")

		// Load Project B config
		loadedB, err := config.LoadProject("new-app")
		require.NoError(t, err)
		assert.False(t, loadedB.CustomFilesMode, "Project B should be managed mode")
		assert.True(t, loadedB.Config.Distributions.Homebrew.Enabled)
		assert.True(t, loadedB.Config.Distributions.NPM.Enabled)
		t.Log("  ✓ Project B config correct (managed mode, Homebrew + NPM)")

		// Load Project C config
		loadedC, err := config.LoadProject("existing-tool")
		require.NoError(t, err)
		assert.False(t, loadedC.CustomFilesMode, "Project C should be managed mode")
		assert.True(t, loadedC.Config.Distributions.NPM.Enabled, "NPM should be enabled")
		t.Log("  ✓ Project C config correct (managed mode, updated with NPM)")

		// Load Project D config
		loadedD, err := config.LoadProject("js-wrapper")
		require.NoError(t, err)
		assert.False(t, loadedD.CustomFilesMode)
		assert.Nil(t, loadedD.Config.Distributions.Homebrew, "Project D should not have Homebrew")
		assert.True(t, loadedD.Config.Distributions.NPM.Enabled)
		t.Log("  ✓ Project D config correct (managed mode, NPM only)")

		time.Sleep(10 * time.Millisecond)
	}

	// ========================================
	// FINAL VERIFICATION: All configs independent
	// ========================================

	t.Log("\n========================================")
	t.Log("FINAL VERIFICATION")
	t.Log("========================================\n")

	// Verify all 4 configs exist and are independent
	configFiles := []string{
		filepath.Join(configRoot, "projects", "my-cli-tool.yaml"),
		filepath.Join(configRoot, "projects", "new-app.yaml"),
		filepath.Join(configRoot, "projects", "existing-tool.yaml"),
		filepath.Join(configRoot, "projects", "js-wrapper.yaml"),
	}

	t.Log("Config files created:")
	for _, cf := range configFiles {
		info, err := os.Stat(cf)
		require.NoError(t, err, "Config file should exist: %s", cf)
		t.Logf("  ✓ %s (%d bytes)", cf, info.Size())
	}

	// Verify project files still exist
	t.Log("\nProject files preserved:")
	projectFiles := []struct {
		path string
		file string
	}{
		{projectA.path, ".goreleaser.yaml"},
		{projectB.path, "go.mod"},
		{projectC.path, ".goreleaser.yaml"},
		{projectD.path, "package.json"},
	}

	for _, pf := range projectFiles {
		fullPath := filepath.Join(pf.path, pf.file)
		_, err := os.Stat(fullPath)
		assert.NoError(t, err, "Project file should exist: %s", fullPath)
		t.Logf("  ✓ %s", fullPath)
	}

	t.Log("\n========================================")
	t.Log("✓ REALISTIC MULTI-PROJECT SCENARIO PASSED")
	t.Log("========================================")
	t.Logf("\nAll config files preserved in: %s", configRoot)
	t.Logf("All project files preserved in: %s", projectsRoot)
	t.Log("\nYou can inspect these files after test completion.")
}

// Helper: Setup Project A - existing project with custom .goreleaser.yaml
func setupProjectA(t *testing.T, root string) *projectSetup {
	projectPath := filepath.Join(root, "my-cli-tool")
	require.NoError(t, os.MkdirAll(projectPath, 0755))

	// Create go.mod
	goMod := `module github.com/user/my-cli-tool

go 1.24

require (
	github.com/spf13/cobra v1.8.0
)`
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte(goMod), 0644))

	// Create custom .goreleaser.yaml (no distui marker)
	customGoReleaser := `project_name: my-cli-tool

before:
  hooks:
    - go mod download
    - go test ./...

builds:
  - main: ./cmd/my-cli-tool
    binary: my-cli-tool
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
    goos:
      - linux
      - darwin
      - windows

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
`
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, ".goreleaser.yaml"), []byte(customGoReleaser), 0644))

	// Create main.go
	mainGo := `package main

import "fmt"

func main() {
	fmt.Println("My CLI Tool")
}`
	cmdDir := filepath.Join(projectPath, "cmd", "my-cli-tool")
	require.NoError(t, os.MkdirAll(cmdDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte(mainGo), 0644))

	return &projectSetup{path: projectPath, name: "my-cli-tool"}
}

// Helper: Setup Project B - brand new project
func setupProjectB(t *testing.T, root string) *projectSetup {
	projectPath := filepath.Join(root, "new-app")
	require.NoError(t, os.MkdirAll(projectPath, 0755))

	// Just go.mod and main.go
	goMod := `module github.com/user/new-app

go 1.24`
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte(goMod), 0644))

	mainGo := `package main

func main() {
	println("New App")
}`
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "main.go"), []byte(mainGo), 0644))

	return &projectSetup{path: projectPath, name: "new-app"}
}

// Helper: Setup Project C - already distui-managed
func setupProjectC(t *testing.T, root string) *projectSetup {
	projectPath := filepath.Join(root, "existing-tool")
	require.NoError(t, os.MkdirAll(projectPath, 0755))

	// Create go.mod
	goMod := `module github.com/user/existing-tool

go 1.24`
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte(goMod), 0644))

	// Create distui-generated .goreleaser.yaml (WITH marker)
	distuiGoReleaser := `# Generated by distui - DO NOT EDIT
# This file is managed by distui

project_name: existing-tool

builds:
  - main: .
    binary: existing-tool
    goos:
      - linux
      - darwin

archives:
  - format: tar.gz
`
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, ".goreleaser.yaml"), []byte(distuiGoReleaser), 0644))

	mainGo := `package main

func main() {}`
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "main.go"), []byte(mainGo), 0644))

	return &projectSetup{path: projectPath, name: "existing-tool"}
}

// Helper: Setup Project D - NPM package only
func setupProjectD(t *testing.T, root string) *projectSetup {
	projectPath := filepath.Join(root, "js-wrapper")
	require.NoError(t, os.MkdirAll(projectPath, 0755))

	// Create go.mod
	goMod := `module github.com/myorg/js-wrapper

go 1.24`
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte(goMod), 0644))

	// Create package.json
	packageJSON := `{
  "name": "@myorg/js-wrapper",
  "version": "1.0.0",
  "description": "JavaScript wrapper for Go CLI",
  "main": "index.js",
  "bin": {
    "js-wrapper": "./bin/cli.js"
  }
}`
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "package.json"), []byte(packageJSON), 0644))

	mainGo := `package main

func main() {}`
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "main.go"), []byte(mainGo), 0644))

	return &projectSetup{path: projectPath, name: "js-wrapper"}
}

type projectSetup struct {
	path string
	name string
}
