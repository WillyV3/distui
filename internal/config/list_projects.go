package config

import (
	"os"
	"strings"

	"distui/internal/models"
)

func LoadAllProjects() ([]models.ProjectConfig, error) {
	projectsDir := expandHome("~/.distui/projects")

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, err
	}

	var projects []models.ProjectConfig
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		identifier := strings.TrimSuffix(entry.Name(), ".yaml")
		project, err := LoadProject(identifier)
		if err != nil {
			continue
		}

		projects = append(projects, *project)
	}

	return projects, nil
}
