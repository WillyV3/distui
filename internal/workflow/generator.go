package workflow

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"distui/internal/models"
)

type WorkflowData struct {
	IncludeTests bool
	NPMEnabled   bool
}

func GenerateWorkflow(config *models.ProjectConfig) (string, error) {
	tmpl, err := template.New("workflow").Parse(workflowTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	data := WorkflowData{
		IncludeTests: true,
		NPMEnabled:   false,
	}

	if config.Config != nil {
		if config.Config.CICD != nil && config.Config.CICD.GitHubActions != nil {
			data.IncludeTests = config.Config.CICD.GitHubActions.IncludeTests
		}
		if config.Config.Distributions.NPM != nil && config.Config.Distributions.NPM.Enabled {
			data.NPMEnabled = true
		}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}

func GetRequiredSecrets(config *models.ProjectConfig) []string {
	secrets := []string{"GITHUB_TOKEN (automatic)"}

	if config.Config != nil {
		if config.Config.Distributions.NPM != nil && config.Config.Distributions.NPM.Enabled {
			secrets = append(secrets, "NPM_TOKEN")
		}
	}

	return secrets
}

func WriteWorkflowFile(projectPath, yamlContent string) error {
	workflowDir := filepath.Join(projectPath, ".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		return fmt.Errorf("creating workflow directory: %w", err)
	}

	workflowPath := filepath.Join(workflowDir, "release.yml")

	tempFile := workflowPath + ".tmp"
	if err := os.WriteFile(tempFile, []byte(yamlContent), 0644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}

	if err := os.Rename(tempFile, workflowPath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("atomic rename failed: %w", err)
	}

	return nil
}

func WorkflowExists(projectPath string) bool {
	workflowPath := filepath.Join(projectPath, ".github", "workflows", "release.yml")
	_, err := os.Stat(workflowPath)
	return err == nil
}
