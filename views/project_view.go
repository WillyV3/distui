package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"distui/handlers"
	"distui/internal/gitcleanup"
	"distui/internal/models"
	"github.com/charmbracelet/lipgloss"
)

func RenderProjectContent(project *models.ProjectInfo, config *models.ProjectConfig, globalConfig *models.GlobalConfig, releaseModel *handlers.ReleaseModel, configureModel *handlers.ConfigureModel, switchedToPath string, asciiArt string) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true).
		Padding(0, 1)

	infoStyle := lipgloss.NewStyle().
		Padding(0, 1).
		MarginLeft(2)

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true).
		Padding(0, 1).
		MarginLeft(2)

	subtleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))

	var content strings.Builder

	// GitHub status (skip for first-time users - shown in welcome screen instead)
	isFirstTimeUser := globalConfig == nil || globalConfig.User.GitHubUsername == ""
	if !isFirstTimeUser && globalConfig.User.GitHubUsername != "" {
		content.WriteString(successStyle.Render(fmt.Sprintf("✓ GitHub: %s", globalConfig.User.GitHubUsername)) + "\n\n")
	}

	// Custom mode indicator
	if config != nil && config.CustomFilesMode {
		customIndicator := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Render("[Custom Config]")
		content.WriteString(customIndicator + " Using your own .goreleaser.yaml/package.json\n\n")
	}

	// Project switch notification
	if switchedToPath != "" {
		content.WriteString(successStyle.Render(fmt.Sprintf("→ Switched to: %s", switchedToPath)) + "\n\n")
	}

	// Regeneration warning
	if configureModel != nil && configureModel.NeedsRegeneration {
		content.WriteString(warningStyle.Render("⚠ Configuration changed - Press [c] then [R] to regenerate release files before releasing") + "\n\n")
	}

	// Check if .goreleaser.yaml exists
	hasGoReleaserConfig := false
	if project != nil {
		goreleaserPaths := []string{
			filepath.Join(project.Path, ".goreleaser.yaml"),
			filepath.Join(project.Path, ".goreleaser.yml"),
			filepath.Join(project.Path, "goreleaser.yaml"),
			filepath.Join(project.Path, "goreleaser.yml"),
		}
		for _, path := range goreleaserPaths {
			if _, err := os.Stat(path); err == nil {
				hasGoReleaserConfig = true
				break
			}
		}
	}

	// UNCONFIGURED project - minimal view (no config OR no .goreleaser.yaml)
	if project != nil && (config == nil || !hasGoReleaserConfig) {
		// Detect first-time user (no global config set up)
		isFirstTimeUser := globalConfig == nil || globalConfig.User.GitHubUsername == ""

		if isFirstTimeUser {
			// First-time user flow - guide to settings first
			// Left-aligned styles for welcome screen
			welcomeHeader := lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
			welcomeInfo := lipgloss.NewStyle()
			welcomeSubtle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
			welcomeWarning := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

			if asciiArt != "" {
				content.WriteString(asciiArt)
			}
			content.WriteString(welcomeHeader.Render("WELCOME TO DISTUI") + "\n\n")
			content.WriteString(welcomeWarning.Render("⚠ GitHub not configured") + "\n\n")
			content.WriteString(welcomeInfo.Render("Get started in 3 steps:") + "\n\n")
			content.WriteString(welcomeInfo.Render("1. Press [s] to configure global settings") + "\n")
			content.WriteString(welcomeInfo.Render("   → Add your GitHub username, Homebrew tap, NPM scope") + "\n\n")
			content.WriteString(welcomeInfo.Render("2. Press [g] to discover and manage projects") + "\n")
			content.WriteString(welcomeInfo.Render("   → distui will scan for your Go projects") + "\n\n")
			content.WriteString(welcomeInfo.Render("3. Return here and press [c] to configure this project") + "\n")
			content.WriteString(welcomeInfo.Render("   → Set up distributions and release settings") + "\n\n")
			content.WriteString(welcomeSubtle.Render("s: settings • g: global • q: quit"))
		} else {
			// Show project info
			content.WriteString(infoStyle.Render(fmt.Sprintf("%s", project.Module.Name)) + "\n")
			content.WriteString(subtleStyle.Render(fmt.Sprintf("%s", project.Path)) + "\n\n")

			// Status indicators
			if config != nil {
				// Config exists in distui
				content.WriteString(successStyle.Render("✓ Project configured in distui") + "\n")

				// Show what's enabled
				enabledDists := []string{}
				if config.Config != nil && config.Config.Distributions.GitHubRelease != nil && config.Config.Distributions.GitHubRelease.Enabled {
					enabledDists = append(enabledDists, "GitHub")
				}
				if config.Config != nil && config.Config.Distributions.Homebrew != nil && config.Config.Distributions.Homebrew.Enabled {
					enabledDists = append(enabledDists, "Homebrew")
				}
				if config.Config != nil && config.Config.Distributions.NPM != nil && config.Config.Distributions.NPM.Enabled {
					enabledDists = append(enabledDists, "NPM")
				}
				if len(enabledDists) > 0 {
					content.WriteString(subtleStyle.Render(fmt.Sprintf("  Distributions: %s", strings.Join(enabledDists, ", "))) + "\n")
				}
			} else {
				content.WriteString(warningStyle.Render("⚠ No distui configuration found") + "\n")
			}

			// Check for release files
			if !hasGoReleaserConfig {
				content.WriteString(warningStyle.Render("⚠ Release files not generated yet") + "\n")
			}

			content.WriteString("\n")

			// Next steps based on state
			if config != nil && !hasGoReleaserConfig {
				// Config exists but no release files
				content.WriteString(infoStyle.Render("Generate release files:") + "\n\n")
				content.WriteString(infoStyle.Render("1. Press [c] to open configure view") + "\n")
				content.WriteString(infoStyle.Render("1. Press [TAB] to navigate to Distribution Tab") + "\n")

				content.WriteString(infoStyle.Render("2. Press [R] to generate .goreleaser.yaml in your repo") + "\n")
				content.WriteString(infoStyle.Render("3. Commit the generated file to your repository") + "\n")
				content.WriteString(infoStyle.Render("4. Return here and press [r] to release") + "\n\n")
			} else {
				// No config at all
				content.WriteString(infoStyle.Render("Configure this project:") + "\n\n")
				content.WriteString(infoStyle.Render("1. Press [c] to configure distributions (Homebrew, NPM, etc.)") + "\n")
				content.WriteString(infoStyle.Render("2. Save your configuration") + "\n")
				content.WriteString(infoStyle.Render("3. Press [R] to generate .goreleaser.yaml in your repo") + "\n")
				content.WriteString(infoStyle.Render("4. Commit the config file and return here to release") + "\n\n")
			}

			content.WriteString(subtleStyle.Render("c: configure • g: global • s: settings • q: quit"))
		}
		return content.String()
	}

	// NO project detected
	if project == nil {
		isFirstTimeUser := globalConfig == nil || globalConfig.User.GitHubUsername == ""

		if isFirstTimeUser {
			// Left-aligned styles for welcome screen
			welcomeHeader := lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
			welcomeInfo := lipgloss.NewStyle()
			welcomeSubtle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
			welcomeWarning := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

			if asciiArt != "" {
				content.WriteString(asciiArt)
			}
			content.WriteString(welcomeHeader.Render("WELCOME TO DISTUI") + "\n\n")
			content.WriteString(welcomeWarning.Render("⚠ GitHub not configured") + "\n\n")
			content.WriteString(welcomeInfo.Render("No Go project detected in current directory") + "\n\n")
			content.WriteString(welcomeInfo.Render("Get started:") + "\n\n")
			content.WriteString(welcomeInfo.Render("1. Press [s] to configure global settings") + "\n")
			content.WriteString(welcomeInfo.Render("   → Add your GitHub username, Homebrew tap, NPM scope") + "\n\n")
			content.WriteString(welcomeInfo.Render("2. Press [g] to discover and manage your Go projects") + "\n\n")
			content.WriteString(welcomeSubtle.Render("s: settings • g: global • q: quit"))
		} else {
			content.WriteString(headerStyle.Render("NO PROJECT") + "\n\n")
			content.WriteString(infoStyle.Render("No Go project detected in current directory") + "\n\n")
			content.WriteString(infoStyle.Render("Press [g] to view and select from your configured projects") + "\n\n")
			content.WriteString(subtleStyle.Render("g: global • s: settings • q: quit"))
		}
		return content.String()
	}

	// Check if release is in progress (not just version selection)
	if releaseModel != nil && releaseModel.Phase != models.PhaseVersionSelect {
		// During release, ONLY show the release progress, not project info
		return renderInlineReleaseSection(releaseModel)
	}

	// CONFIGURED project - check if working tree is clean (only when not releasing)
	isClean := gitcleanup.IsWorkingTreeClean()

	if !isClean {
		content.WriteString(headerStyle.Render("WORKING TREE NOT CLEAN") + "\n\n")
		content.WriteString(infoStyle.Render(fmt.Sprintf("%s", project.Module.Name)) + "\n")
		content.WriteString(subtleStyle.Render(fmt.Sprintf("%s", project.Path)) + "\n\n")
		content.WriteString(warningStyle.Render("⚠ You have uncommitted changes") + "\n\n")
		content.WriteString(infoStyle.Render("Before releasing, you must clean your working tree:") + "\n\n")
		content.WriteString(infoStyle.Render("1. Press [c] to open the configuration view") + "\n")
		content.WriteString(infoStyle.Render("2. Use the Cleanup tab to commit/push changes") + "\n")
		content.WriteString(infoStyle.Render("3. Return here to release") + "\n\n")
		content.WriteString(subtleStyle.Render("c: configure • g: global • s: settings • q: quit"))
		return content.String()
	}

	// CONFIGURED project with clean working tree - full view
	content.WriteString(headerStyle.Render(project.Module.Name) + "\n\n")
	content.WriteString(infoStyle.Render(fmt.Sprintf("Version: %s", project.Module.Version)) + "\n")

	if project.Repository != nil {
		content.WriteString(infoStyle.Render(fmt.Sprintf("Repo: %s/%s",
			project.Repository.Owner, project.Repository.Name)) + "\n")
	}

	// Distribution info (only show when not in release mode)
	if releaseModel == nil || releaseModel.Phase == models.PhaseVersionSelect {
		if config != nil && config.Config != nil {
			hasDistributions := false
			if config.Config.Distributions.NPM != nil && config.Config.Distributions.NPM.Enabled {
				if !hasDistributions {
					content.WriteString("\n")
					hasDistributions = true
				}
				npmName := config.Config.Distributions.NPM.PackageName
				if npmName == "" && project != nil {
					npmName = project.Module.Name
				}
				content.WriteString(infoStyle.Render(fmt.Sprintf("NPM: %s", npmName)) + "\n")
			}
			if config.Config.Distributions.Homebrew != nil && config.Config.Distributions.Homebrew.Enabled {
				if !hasDistributions {
					content.WriteString("\n")
					hasDistributions = true
				}
				tapRepo := config.Config.Distributions.Homebrew.TapRepo
				if tapRepo == "" && project != nil && project.Repository != nil {
					tapRepo = fmt.Sprintf("%s/homebrew-tap", project.Repository.Owner)
				}
				content.WriteString(infoStyle.Render(fmt.Sprintf("Homebrew: %s", tapRepo)) + "\n")
			}
		}
	}

	// Version selection appears inline when [r] pressed
	if releaseModel != nil && releaseModel.Phase == models.PhaseVersionSelect {
		content.WriteString("\n")
		content.WriteString(renderInlineReleaseSection(releaseModel))
		content.WriteString("\n")
	}

	// Recent releases (only if history exists and no release in progress)
	if releaseModel == nil && config != nil && config.History != nil && len(config.History.Releases) > 0 {
		content.WriteString("\n" + headerStyle.Render("RECENT RELEASES") + "\n\n")
		for i, release := range config.History.Releases[:min(3, len(config.History.Releases))] {
			if i > 2 {
				break
			}
			status := "✓"
			if release.Status == "failed" {
				status = "✗"
			}
			content.WriteString(infoStyle.Render(fmt.Sprintf("%s %s (%s)",
				status, release.Version, release.Duration)) + "\n")
		}
		content.WriteString("\n")
	}

	content.WriteString(subtleStyle.Render("r: release • c: configure • g: global • s: settings • q: quit"))

	return content.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func renderInlineReleaseSection(m *handlers.ReleaseModel) string {
	if m == nil {
		return ""
	}

	switch m.Phase {
	case models.PhaseVersionSelect:
		return renderCompactVersionSelect(m)
	case models.PhaseComplete:
		return RenderSuccess(m)
	case models.PhaseFailed:
		return RenderFailure(m)
	default:
		return RenderProgress(m)
	}
}

func renderCompactVersionSelect(m *handlers.ReleaseModel) string {
	var content strings.Builder

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true).
		Padding(0, 1)

	fieldStyle := lipgloss.NewStyle().
		Padding(0, 1).
		MarginLeft(2)

	actionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true).
		Padding(0, 1).
		MarginLeft(2)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true).
		Padding(0, 1).
		MarginLeft(2)

	subtleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	configureStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117")) // Teal/cyan color for configure
	configureSelectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)

	content.WriteString(headerStyle.Render("SELECT RELEASE VERSION") + "\n\n")
	content.WriteString(fieldStyle.Render(fmt.Sprintf("Current: %s", m.CurrentVersion)) + "\n\n")

	versions := []string{
		"Configure Project",
		"Patch (bug fixes)",
		"Minor (new features)",
		"Major (breaking changes)",
		"Custom version",
	}

	for i, ver := range versions {
		prefix := "  "
		style := actionStyle

		// Special styling for Configure Project (item 0)
		if i == 0 {
			style = configureStyle
			if i == m.SelectedVersion {
				prefix = "> "
				style = configureSelectedStyle
			}
		} else {
			if i == m.SelectedVersion {
				prefix = "> "
				style = selectedStyle
			}
		}

		content.WriteString(style.Render(prefix+ver) + "\n")
	}

	if m.SelectedVersion == 4 {
		content.WriteString("\n" + fieldStyle.Render("Enter version: ") + m.VersionInput.View() + "\n")
	}

	// Show changelog input if enabled and a release version is selected (not Configure Project, not Custom)
	needsChangelog := false
	if m.ProjectConfig != nil && m.ProjectConfig.Config != nil && m.ProjectConfig.Config.Release != nil {
		needsChangelog = m.ProjectConfig.Config.Release.GenerateChangelog
	}
	if needsChangelog && m.SelectedVersion > 0 && m.SelectedVersion < 4 {
		content.WriteString("\n" + fieldStyle.Render("Changelog: ") + m.ChangelogInput.View() + "\n")
	}

	content.WriteString("\n" + subtleStyle.Render("↑/↓: navigate • enter: start • esc: cancel"))

	return content.String()
}