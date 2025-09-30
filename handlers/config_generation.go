package handlers

import (
	"fmt"
	"os"

	"distui/internal/generator"
	"distui/internal/models"
)

func CheckMissingConfigFiles(detectedProject *models.ProjectInfo, projectConfig *models.ProjectConfig) []string {
	if detectedProject == nil {
		return nil
	}

	var missing []string
	projectPath := detectedProject.Path

	// Check for .goreleaser.yaml
	goreleaserPaths := []string{
		projectPath + "/.goreleaser.yaml",
		projectPath + "/.goreleaser.yml",
		projectPath + "/goreleaser.yaml",
		projectPath + "/goreleaser.yml",
	}
	hasGoreleaser := false
	for _, p := range goreleaserPaths {
		if _, err := os.Stat(p); err == nil {
			hasGoreleaser = true
			break
		}
	}
	if !hasGoreleaser {
		missing = append(missing, ".goreleaser.yaml")
	}

	// Check for package.json if NPM enabled
	if projectConfig != nil && projectConfig.Config != nil &&
		projectConfig.Config.Distributions.NPM != nil &&
		projectConfig.Config.Distributions.NPM.Enabled {
		if _, err := os.Stat(projectPath + "/package.json"); err != nil {
			missing = append(missing, "package.json")
		}
	}

	return missing
}

func GenerateConfigFiles(detectedProject *models.ProjectInfo, projectConfig *models.ProjectConfig, files []string) error {
	if detectedProject == nil || projectConfig == nil {
		return nil
	}

	for _, fileName := range files {
		if fileName == ".goreleaser.yaml" {
			content, err := generator.GenerateGoReleaserConfig(detectedProject, projectConfig)
			if err != nil {
				return err
			}
			// Allow overwrite for regeneration
			if err := generator.WriteGoReleaserConfigForce(detectedProject.Path, content); err != nil {
				return err
			}
		} else if fileName == "package.json" {
			content, err := generator.GeneratePackageJSON(detectedProject, projectConfig)
			if err != nil {
				return err
			}
			// Allow overwrite for regeneration
			if err := generator.WritePackageJSONForce(detectedProject.Path, content); err != nil {
				return err
			}
		}
	}

	return nil
}

func DeleteConfigFiles(projectPath string, files []string) error {
	for _, fileName := range files {
		var fullPath string
		if fileName == ".goreleaser.yaml" {
			// Try both .yaml and .yml variants
			yamlPath := projectPath + "/.goreleaser.yaml"
			ymlPath := projectPath + "/.goreleaser.yml"

			if _, err := os.Stat(yamlPath); err == nil {
				fullPath = yamlPath
			} else if _, err := os.Stat(ymlPath); err == nil {
				fullPath = ymlPath
			}
		} else {
			fullPath = projectPath + "/" + fileName
		}

		if fullPath != "" {
			if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("deleting %s: %w", fileName, err)
			}
		}
	}
	return nil
}

type ConfigFileChanges struct {
	FilesToGenerate []string
	FilesToDelete   []string
}

func GetConfigFilesForRegeneration(detectedProject *models.ProjectInfo, projectConfig *models.ProjectConfig) []string {
	changes := GetConfigFileChanges(detectedProject, projectConfig)
	return changes.FilesToGenerate
}

func GetConfigFileChanges(detectedProject *models.ProjectInfo, projectConfig *models.ProjectConfig) ConfigFileChanges {
	if detectedProject == nil || projectConfig == nil {
		return ConfigFileChanges{}
	}

	var changes ConfigFileChanges
	projectPath := detectedProject.Path

	// Check if GoReleaser is needed (for GitHub Releases or Homebrew)
	needsGoreleaser := false
	if projectConfig.Config != nil {
		if projectConfig.Config.Distributions.GitHubRelease != nil &&
			projectConfig.Config.Distributions.GitHubRelease.Enabled {
			needsGoreleaser = true
		}
		if projectConfig.Config.Distributions.Homebrew != nil &&
			projectConfig.Config.Distributions.Homebrew.Enabled {
			needsGoreleaser = true
		}
	}

	// Check if .goreleaser.yaml exists
	goreleaserExists := false
	goreleaserPaths := []string{
		projectPath + "/.goreleaser.yaml",
		projectPath + "/.goreleaser.yml",
	}
	for _, p := range goreleaserPaths {
		if _, err := os.Stat(p); err == nil {
			goreleaserExists = true
			break
		}
	}

	// Determine GoReleaser action
	if needsGoreleaser {
		changes.FilesToGenerate = append(changes.FilesToGenerate, ".goreleaser.yaml")
	} else if goreleaserExists {
		changes.FilesToDelete = append(changes.FilesToDelete, ".goreleaser.yaml")
	}

	// Check if NPM is enabled
	npmEnabled := false
	if projectConfig.Config != nil &&
		projectConfig.Config.Distributions.NPM != nil &&
		projectConfig.Config.Distributions.NPM.Enabled {
		npmEnabled = true
	}

	// Check if package.json exists
	packageJsonExists := false
	if _, err := os.Stat(projectPath + "/package.json"); err == nil {
		packageJsonExists = true
	}

	// Determine package.json action
	if npmEnabled {
		changes.FilesToGenerate = append(changes.FilesToGenerate, "package.json")
	} else if packageJsonExists {
		changes.FilesToDelete = append(changes.FilesToDelete, "package.json")
	}

	return changes
}