package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"distui/handlers"
)

var (
	titleStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
	selectedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	normalStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	enabledStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	disabledStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	confirmStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
)

func RenderSmartCommitPrefs(model *handlers.SmartCommitPrefsModel) string {
	if model == nil {
		return "Loading preferences..."
	}

	var content strings.Builder

	content.WriteString(titleStyle.Render("SMART COMMIT PREFERENCES"))
	content.WriteString("\n\n")

	if model.ShowConfirm {
		return renderResetConfirm(&content)
	}

	useCustom := model.ProjectConfig.Config.SmartCommit != nil &&
		model.ProjectConfig.Config.SmartCommit.UseCustomRules

	// Show toggle
	if useCustom {
		content.WriteString(enabledStyle.Render("[✓] Use Custom Rules"))
	} else {
		content.WriteString(disabledStyle.Render("[ ] Use Custom Rules"))
	}
	content.WriteString("\n\n")

	if !useCustom {
		return renderCustomDisabled(&content)
	}

	switch model.EditMode {
	case handlers.ModeEditCategory:
		return renderEditCategory(&content, model)
	case handlers.ModeAddExtension:
		return renderAddExtension(&content, model)
	case handlers.ModeAddPattern:
		return renderAddPattern(&content, model)
	default:
		return renderBrowseCategories(&content, model)
	}
}

func renderResetConfirm(content *strings.Builder) string {
	content.WriteString(confirmStyle.Render("Reset custom rules to defaults?"))
	content.WriteString("\n\n")
	content.WriteString(normalStyle.Render("This will disable custom rules and revert to built-in categorization."))
	content.WriteString("\n\n")
	content.WriteString(dimStyle.Render("[y] Yes, reset  [n] No, keep custom  [ESC] Cancel"))
	return content.String()
}

func renderCustomDisabled(content *strings.Builder) string {
	content.WriteString(normalStyle.Render("What is Smart Commit?"))
	content.WriteString("\n")
	content.WriteString(dimStyle.Render("Smart commit automatically groups changed files by type (code, config, docs, etc.)"))
	content.WriteString("\n")
	content.WriteString(dimStyle.Render("and creates separate commits with descriptive messages for each category."))
	content.WriteString("\n\n")
	content.WriteString(normalStyle.Render("Custom Rules: Disabled"))
	content.WriteString("\n")
	content.WriteString(dimStyle.Render("Using built-in categorization:"))
	content.WriteString("\n")
	content.WriteString(dimStyle.Render("  • code: .go, .js, .ts, .py, .rb, .java, .c, .cpp, .h, .rs"))
	content.WriteString("\n")
	content.WriteString(dimStyle.Render("  • config: .yaml, .yml, .json, .toml, .ini, .conf, .env"))
	content.WriteString("\n")
	content.WriteString(dimStyle.Render("  • docs: .md, .txt, .rst, .adoc"))
	content.WriteString("\n")
	content.WriteString(dimStyle.Render("  • build: .mod, .sum, .lock, Makefile, Dockerfile, .goreleaser*"))
	content.WriteString("\n")
	content.WriteString(dimStyle.Render("  • test: *_test.go, .test, .spec.js, .spec.ts"))
	content.WriteString("\n")
	content.WriteString(dimStyle.Render("  • assets: .png, .jpg, .svg, .ico, .gif, .woff, .ttf, .css"))
	content.WriteString("\n")
	content.WriteString(dimStyle.Render("  • data: .sql, .db, .csv, .xml"))
	content.WriteString("\n\n")
	content.WriteString(normalStyle.Render("Enable custom rules to override these defaults."))
	content.WriteString("\n\n")
	content.WriteString(dimStyle.Render("[space] Enable custom rules  [s] Save  [ESC] Back"))
	return content.String()
}

func renderBrowseCategories(content *strings.Builder, model *handlers.SmartCommitPrefsModel) string {
	content.WriteString(normalStyle.Render("File Categories"))
	content.WriteString("\n")
	content.WriteString(dimStyle.Render("Smart commit groups files into these categories for separate commits."))
	content.WriteString("\n")
	content.WriteString(dimStyle.Render("Select a category to add custom extensions or glob patterns."))
	content.WriteString("\n\n")

	categoryDescriptions := map[string]string{
		"code":   "Source code files",
		"config": "Configuration files",
		"docs":   "Documentation files",
		"build":  "Build system files",
		"test":   "Test files",
		"assets": "Images, fonts, CSS",
		"data":   "Database, CSV, XML",
	}

	for i, category := range model.Categories {
		desc := categoryDescriptions[category]
		if i == model.SelectedCategory {
			content.WriteString(selectedStyle.Render("→ " + category))
			content.WriteString(dimStyle.Render(" - " + desc))
		} else {
			content.WriteString(dimStyle.Render("  " + category))
			content.WriteString(dimStyle.Render(" - " + desc))
		}
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(dimStyle.Render("[↑↓] Navigate  [e] Edit selected  [space] Disable custom  [r] Reset"))
	content.WriteString("\n")
	content.WriteString(dimStyle.Render("[s] Save  [ESC] Back"))
	return content.String()
}

func renderEditCategory(content *strings.Builder, model *handlers.SmartCommitPrefsModel) string {
	category := model.Categories[model.SelectedCategory]
	rules := model.ProjectConfig.Config.SmartCommit.Categories[category]

	categoryDescriptions := map[string]string{
		"code":   "Source code files",
		"config": "Configuration files",
		"docs":   "Documentation files",
		"build":  "Build system files",
		"test":   "Test files",
		"assets": "Images, fonts, CSS",
		"data":   "Database, CSV, XML",
	}

	content.WriteString(selectedStyle.Render("Editing: " + category))
	content.WriteString(dimStyle.Render(" (" + categoryDescriptions[category] + ")"))
	content.WriteString("\n")
	content.WriteString(dimStyle.Render("Files matching these rules will be committed under this category."))
	content.WriteString("\n\n")

	content.WriteString(normalStyle.Render("File Extensions:"))
	content.WriteString(dimStyle.Render(" (e.g., .proto, .rs, .vue)"))
	content.WriteString("\n")
	if len(rules.Extensions) > 0 {
		for _, ext := range rules.Extensions {
			content.WriteString(dimStyle.Render("  • " + ext))
			content.WriteString("\n")
		}
	} else {
		content.WriteString(dimStyle.Render("  (none added)"))
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(normalStyle.Render("Glob Patterns:"))
	content.WriteString(dimStyle.Render(" (e.g., **/migrations/**, **/*.proto)"))
	content.WriteString("\n")
	if len(rules.Patterns) > 0 {
		for _, pat := range rules.Patterns {
			content.WriteString(dimStyle.Render("  • " + pat))
			content.WriteString("\n")
		}
	} else {
		content.WriteString(dimStyle.Render("  (none added)"))
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(dimStyle.Render("[e] Add extension  [p] Add pattern  [d] Delete (TODO)  [ESC] Back"))
	return content.String()
}

func renderAddExtension(content *strings.Builder, model *handlers.SmartCommitPrefsModel) string {
	category := model.Categories[model.SelectedCategory]

	content.WriteString(selectedStyle.Render("Add Extension to: " + category))
	content.WriteString("\n\n")
	content.WriteString(normalStyle.Render("Enter file extension (e.g., .rs, .proto, .yml):"))
	content.WriteString("\n\n")
	content.WriteString(model.ExtensionInput.View())
	content.WriteString("\n\n")
	content.WriteString(dimStyle.Render("[Enter] Add extension  [ESC] Cancel"))
	return content.String()
}

func renderAddPattern(content *strings.Builder, model *handlers.SmartCommitPrefsModel) string {
	category := model.Categories[model.SelectedCategory]

	content.WriteString(selectedStyle.Render("Add Pattern to: " + category))
	content.WriteString("\n\n")
	content.WriteString(normalStyle.Render("Enter glob pattern (e.g., **/src/**, **/*_test.go):"))
	content.WriteString("\n\n")
	content.WriteString(model.PatternInput.View())
	content.WriteString("\n\n")
	content.WriteString(dimStyle.Render("[Enter] Add pattern  [ESC] Cancel"))
	return content.String()
}
