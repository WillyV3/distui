package handlers

import (
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

func GetConfigFilesForRegeneration(detectedProject *models.ProjectInfo, projectConfig *models.ProjectConfig) []string {
	if detectedProject == nil || projectConfig == nil {
		return nil
	}

	var files []string

	// Always include .goreleaser.yaml for regeneration
	files = append(files, ".goreleaser.yaml")

	// Include package.json if NPM is enabled
	if projectConfig.Config != nil &&
		projectConfig.Config.Distributions.NPM != nil &&
		projectConfig.Config.Distributions.NPM.Enabled {
		files = append(files, "package.json")
	}

	return files
}