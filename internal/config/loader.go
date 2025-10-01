package config

import (
	"fmt"
	"os"
	"path/filepath"

	"distui/internal/models"
	"gopkg.in/yaml.v3"
)

func expandHome(path string) string {
	if path == "" {
		return ""
	}

	if path[0] != '~' {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("cannot determine home directory: %v", err))
	}

	if len(path) == 1 {
		return homeDir
	}

	return filepath.Join(homeDir, path[1:])
}

func LoadGlobalConfig() (*models.GlobalConfig, error) {
	configPath := expandHome("~/.distui/config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading global config: %w", err)
	}

	var config models.GlobalConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing global config: %w", err)
	}

	return &config, nil
}

func LoadProject(identifier string) (*models.ProjectConfig, error) {
	projectPath := expandHome(fmt.Sprintf("~/.distui/projects/%s.yaml", identifier))

	data, err := os.ReadFile(projectPath)
	if err != nil {
		return nil, fmt.Errorf("reading project %s: %w", identifier, err)
	}

	var project models.ProjectConfig
	if err := yaml.Unmarshal(data, &project); err != nil {
		return nil, fmt.Errorf("parsing project %s: %w", identifier, err)
	}

	if project.Config != nil && project.Config.SmartCommit == nil {
		project.Config.SmartCommit = getDefaultSmartCommitPrefs()
	}

	if project.Config != nil && project.Config.CICD != nil && project.Config.CICD.GitHubActions == nil {
		project.Config.CICD.GitHubActions = &models.GitHubActionsConfig{
			Enabled:        false,
			WorkflowPath:   ".github/workflows/release.yml",
			IncludeTests:   true,
			AutoRegenerate: false,
		}
	}

	return &project, nil
}

func getDefaultSmartCommitPrefs() *models.SmartCommitPrefs {
	return &models.SmartCommitPrefs{
		Enabled:        true,
		UseCustomRules: false,
		Categories:     make(map[string]models.CategoryRules),
	}
}

func SaveGlobalConfig(config *models.GlobalConfig) error {
	configDir := expandHome("~/.distui")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshaling global config: %w", err)
	}

	tempFile := configPath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0600); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}

	if err := os.Rename(tempFile, configPath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("atomic rename failed: %w", err)
	}

	return nil
}

func SaveProject(project *models.ProjectConfig) error {
	if project.Project == nil || project.Project.Identifier == "" {
		return fmt.Errorf("project missing identifier")
	}

	projectsDir := expandHome("~/.distui/projects")
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		return fmt.Errorf("creating projects directory: %w", err)
	}

	projectPath := filepath.Join(projectsDir, project.Project.Identifier+".yaml")

	data, err := yaml.Marshal(project)
	if err != nil {
		return fmt.Errorf("marshaling project: %w", err)
	}

	tempFile := projectPath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0600); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}

	if err := os.Rename(tempFile, projectPath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("atomic rename failed: %w", err)
	}

	return nil
}