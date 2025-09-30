package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"distui/internal/models"
)

type PackageJSON struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description,omitempty"`
	Main        string            `json:"main,omitempty"`
	Bin         map[string]string `json:"bin,omitempty"`
	Repository  *Repository       `json:"repository,omitempty"`
	Keywords    []string          `json:"keywords,omitempty"`
	Author      string            `json:"author,omitempty"`
	License     string            `json:"license,omitempty"`
}

type Repository struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

func GeneratePackageJSON(project *models.ProjectInfo, config *models.ProjectConfig) (string, error) {
	if project == nil || config == nil {
		return "", fmt.Errorf("project and config required")
	}

	if config.Config == nil || config.Config.Distributions.NPM == nil {
		return "", fmt.Errorf("npm config not found")
	}

	packageName := config.Config.Distributions.NPM.PackageName
	if packageName == "" {
		packageName = project.Binary.Name
		if packageName == "" {
			packageName = project.Module.Name
		}
	}

	version := project.Module.Version
	if version == "" {
		version = "0.0.1"
	}
	if version[0] == 'v' {
		version = version[1:]
	}

	pkg := PackageJSON{
		Name:        packageName,
		Version:     version,
		Description: fmt.Sprintf("%s - distributed via distui", project.Module.Name),
		Bin: map[string]string{
			packageName: "./bin/" + project.Binary.Name,
		},
		Keywords: []string{"cli", "tool"},
		License:  "MIT",
	}

	if project.Repository != nil {
		pkg.Repository = &Repository{
			Type: "git",
			URL:  fmt.Sprintf("https://github.com/%s/%s.git", project.Repository.Owner, project.Repository.Name),
		}
	}

	jsonBytes, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling package.json: %w", err)
	}

	return string(jsonBytes) + "\n", nil
}

func WritePackageJSON(projectPath string, content string) error {
	pkgPath := filepath.Join(projectPath, "package.json")

	if _, err := os.Stat(pkgPath); err == nil {
		return fmt.Errorf("file already exists: %s", pkgPath)
	}

	if err := os.WriteFile(pkgPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing package.json: %w", err)
	}

	return nil
}

func WritePackageJSONForce(projectPath string, content string) error {
	pkgPath := filepath.Join(projectPath, "package.json")

	if err := os.WriteFile(pkgPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing package.json: %w", err)
	}

	return nil
}