package views

import (
	"strings"

	"distui/handlers"
)

func RenderWorkflowGen(model *handlers.WorkflowGenModel) string {
	if model == nil {
		return "Loading workflow generator..."
	}

	var content strings.Builder

	content.WriteString(titleStyle.Render("GITHUB ACTIONS WORKFLOW"))
	content.WriteString("\n\n")

	if model.PreviewMode {
		return renderPreview(model)
	}

	if model.ShowConfirm {
		return renderConfirm(model)
	}

	enabled := model.ProjectConfig.Config.CICD.GitHubActions.Enabled
	includeTests := model.ProjectConfig.Config.CICD.GitHubActions.IncludeTests
	autoRegen := model.ProjectConfig.Config.CICD.GitHubActions.AutoRegenerate

	if enabled {
		content.WriteString(enabledStyle.Render("[✓] Enable Workflow Generation"))
	} else {
		content.WriteString(disabledStyle.Render("[ ] Enable Workflow Generation"))
	}
	content.WriteString("  ")
	content.WriteString(dimStyle.Render("(press [space] to toggle)"))
	content.WriteString("\n\n")

	if !enabled {
		content.WriteString(normalStyle.Render("Workflow generation is disabled"))
		content.WriteString("\n\n")
		content.WriteString(dimStyle.Render("[space] Enable • [ESC] Back • [s] Save"))
		return content.String()
	}

	content.WriteString(normalStyle.Render("Options:"))
	content.WriteString("\n")

	if includeTests {
		content.WriteString(enabledStyle.Render("  [✓] Include Tests"))
	} else {
		content.WriteString(disabledStyle.Render("  [ ] Include Tests"))
	}
	content.WriteString("  ")
	content.WriteString(dimStyle.Render("(press [t] to toggle)"))
	content.WriteString("\n")

	if autoRegen {
		content.WriteString(enabledStyle.Render("  [✓] Auto-regenerate on config change"))
	} else {
		content.WriteString(disabledStyle.Render("  [ ] Auto-regenerate on config change"))
	}
	content.WriteString("  ")
	content.WriteString(dimStyle.Render("(press [a] to toggle)"))
	content.WriteString("\n\n")

	content.WriteString(normalStyle.Render("Required Secrets:"))
	content.WriteString("\n")
	for _, secret := range model.RequiredSecrets {
		content.WriteString("  • ")
		content.WriteString(normalStyle.Render(secret))
		content.WriteString("\n")
	}
	content.WriteString("\n")

	if model.Success {
		content.WriteString(enabledStyle.Render("✓ Workflow file generated successfully!"))
		content.WriteString("\n\n")
	}

	if model.Error != "" {
		content.WriteString(disabledStyle.Render("✗ Error: " + model.Error))
		content.WriteString("\n\n")
	}

	content.WriteString(dimStyle.Render("[p] Preview • [g] Generate • [s] Save • [ESC] Back"))

	return content.String()
}

func renderPreview(model *handlers.WorkflowGenModel) string {
	var content strings.Builder

	content.WriteString(titleStyle.Render("WORKFLOW PREVIEW"))
	content.WriteString("\n\n")

	if model.PreviewContent == "" {
		content.WriteString(normalStyle.Render("No preview available"))
	} else {
		lines := strings.Split(model.PreviewContent, "\n")
		maxLines := 20
		if len(lines) > maxLines {
			lines = lines[:maxLines]
			lines = append(lines, "...")
		}
		for _, line := range lines {
			content.WriteString(normalStyle.Render(line))
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")
	content.WriteString(dimStyle.Render("[ESC] Close"))

	return content.String()
}

func renderConfirm(model *handlers.WorkflowGenModel) string {
	var content strings.Builder

	if model.ConfirmOverwrite {
		content.WriteString(confirmStyle.Render("Workflow file already exists. Overwrite?"))
	} else {
		content.WriteString(confirmStyle.Render("Create workflow file?"))
	}
	content.WriteString("\n\n")

	content.WriteString(normalStyle.Render("File: .github/workflows/release.yml"))
	content.WriteString("\n\n")

	content.WriteString(dimStyle.Render("[y] Yes  [n] No  [ESC] Cancel"))

	return content.String()
}
