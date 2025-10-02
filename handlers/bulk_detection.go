package handlers

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"distui/internal/config"
	"distui/internal/detection"
	"distui/internal/models"
)

type bulkDetectionResultMsg struct {
	homebrew []detection.DetectedDistribution
	npm      []detection.DetectedDistribution
	err      error
}

type ProjectsReloadedMsg struct {
	Projects []models.ProjectConfig
}

type ProjectSwitchedMsg struct {
	DetectedProject *models.ProjectInfo
	ProjectConfig   *models.ProjectConfig
}

func BulkDetectDistributionsCmd(globalConfig *models.GlobalConfig) tea.Cmd {
	return func() tea.Msg {
		result := bulkDetectionResultMsg{}

		homebrewTap := ""
		npmScope := ""

		if globalConfig != nil {
			homebrewTap = globalConfig.User.DefaultHomebrewTap
			npmScope = globalConfig.User.NPMScope
		}

		if homebrewTap != "" {
			distributions, err := detection.DetectAllHomebrewFormulas(homebrewTap)
			if err == nil {
				result.homebrew = distributions
			}
		}

		if npmScope != "" {
			distributions, err := detection.DetectAllNPMPackages(npmScope)
			if err == nil {
				result.npm = distributions
			}
		}

		if len(result.homebrew) == 0 && len(result.npm) == 0 {
			result.err = fmt.Errorf("no distributions found")
		}

		return result
	}
}

func ImportDetectedDistributions(distributions []detection.DetectedDistribution, globalConfig *models.GlobalConfig) error {
	now := time.Now()

	for _, dist := range distributions {
		identifier := sanitizeIdentifier(dist.Name)

		existing, _ := config.LoadProject(identifier)
		if existing != nil {
			continue
		}

		projectConfig := &models.ProjectConfig{
			Project: &models.ProjectInfo{
				Identifier:   identifier,
				LastAccessed: &now,
				DetectedAt:   &now,
				Module: &models.ModuleInfo{
					Name:    dist.Name,
					Version: dist.Version,
				},
			},
			Config: &models.ProjectSettings{
				Distributions: models.Distributions{},
			},
			History: &models.ReleaseHistory{
				Releases: []models.ReleaseRecord{},
			},
		}

		if dist.Type == "homebrew" && globalConfig != nil {
			projectConfig.Config.Distributions.Homebrew = &models.HomebrewConfig{
				Enabled:     true,
				TapRepo:     globalConfig.User.DefaultHomebrewTap,
				FormulaName: dist.Name,
			}
		}

		if dist.Type == "npm" {
			projectConfig.Config.Distributions.NPM = &models.NPMConfig{
				Enabled:     true,
				PackageName: dist.Name,
			}
		}

		if err := config.SaveProject(projectConfig); err != nil {
			return fmt.Errorf("saving project %s: %w", identifier, err)
		}
	}

	return nil
}

func sanitizeIdentifier(name string) string {
	identifier := strings.ReplaceAll(name, "/", "-")
	identifier = strings.ReplaceAll(identifier, "@", "")
	identifier = strings.ReplaceAll(identifier, ".", "-")
	identifier = strings.ReplaceAll(identifier, "_", "-")
	return identifier
}

func ReloadProjectsCmd() tea.Cmd {
	return func() tea.Msg {
		projects, err := config.LoadAllProjects()
		if err != nil {
			return ProjectsReloadedMsg{Projects: []models.ProjectConfig{}}
		}
		return ProjectsReloadedMsg{Projects: projects}
	}
}

func SwitchProjectCmd(projectConfig *models.ProjectConfig) tea.Cmd {
	return func() tea.Msg {
		if projectConfig == nil || projectConfig.Project == nil || projectConfig.Project.Path == "" {
			return ProjectSwitchedMsg{
				DetectedProject: nil,
				ProjectConfig:   projectConfig,
			}
		}

		if err := os.Chdir(projectConfig.Project.Path); err != nil {
			return ProjectSwitchedMsg{
				DetectedProject: nil,
				ProjectConfig:   projectConfig,
			}
		}

		detectedProject, err := detection.DetectProject(".")
		if err != nil {
			detectedProject = projectConfig.Project
		}

		// CRITICAL FIX: Reload config from disk to get fresh state
		// The projectConfig passed in might be stale from the global list
		freshConfig, err := config.LoadProject(projectConfig.Project.Identifier)
		if err != nil {
			// If no saved config exists, use the in-memory version
			freshConfig = projectConfig
		}

		// Ensure DetectedProject is synced with fresh config
		if freshConfig.Project == nil {
			freshConfig.Project = detectedProject
		}

		return ProjectSwitchedMsg{
			DetectedProject: detectedProject,
			ProjectConfig:   freshConfig,
		}
	}
}
