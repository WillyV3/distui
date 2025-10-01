package handlers

import (
	tea "github.com/charmbracelet/bubbletea"

	"distui/internal/detection"
	"distui/internal/models"
)

// StartDistributionDetectionCmd creates detection command from config
func StartDistributionDetectionCmd(detectedProject *models.ProjectInfo, globalConfig *models.GlobalConfig) tea.Cmd {
	if detectedProject == nil {
		return nil
	}

	homebrewTap := ""
	homebrewFormula := ""
	homebrewFromFile := false
	npmPackage := ""
	npmScopedPackage := ""
	npmFromFile := false

	// First check for existing .goreleaser.yaml and package.json
	if detectedProject.Path != "" {
		goreleaserConfig, _ := detection.DetectGoReleaserConfig(detectedProject.Path)
		packageJSON, _ := detection.DetectPackageJSON(detectedProject.Path)

		if goreleaserConfig != nil && goreleaserConfig.HasHomebrew {
			homebrewTap = goreleaserConfig.HomebrewTap
			homebrewFormula = goreleaserConfig.FormulaName
			homebrewFromFile = true
		}

		if packageJSON != nil && packageJSON.Name != "" {
			npmPackage = packageJSON.Name
			npmScopedPackage = packageJSON.Name
			npmFromFile = true
		}
	}

	// Fall back to global config if not found in files
	if homebrewTap == "" && globalConfig != nil && globalConfig.User.DefaultHomebrewTap != "" {
		homebrewTap = globalConfig.User.DefaultHomebrewTap
	}

	if homebrewFormula == "" && detectedProject.Binary != nil {
		homebrewFormula = detectedProject.Binary.Name
	}

	if npmPackage == "" && detectedProject.Module != nil {
		npmPackage = detectedProject.Module.Name
		if globalConfig != nil && globalConfig.User.NPMScope != "" {
			npmScopedPackage = "@" + globalConfig.User.NPMScope + "/" + detectedProject.Module.Name
		}
	}

	return DetectDistributionsCmd(homebrewTap, homebrewFormula, homebrewFromFile, npmPackage, npmScopedPackage, npmFromFile)
}

// DetectDistributionsCmd auto-detects if project exists in user's tap/npm
func DetectDistributionsCmd(homebrewTap, homebrewFormula string, homebrewFromFile bool, npmPackage, npmScopedPackage string, npmFromFile bool) tea.Cmd {
	return func() tea.Msg {
		result := distributionDetectedMsg{
			homebrewFromFile: homebrewFromFile,
			npmFromFile:      npmFromFile,
		}

		// Try Homebrew tap if configured
		if homebrewTap != "" && homebrewFormula != "" {
			info, err := detection.VerifyHomebrewFormula(homebrewTap, homebrewFormula)
			if err == nil && info.Exists {
				result.homebrewTap = homebrewTap
				result.homebrewFormula = homebrewFormula
				result.homebrewVersion = info.Version
				result.homebrewExists = true
			} else if homebrewFromFile {
				// Even if not published yet, if from file, include the config
				result.homebrewTap = homebrewTap
				result.homebrewFormula = homebrewFormula
				result.homebrewExists = false
			}
		}

		// Try NPM package (with scope first, then without)
		if npmScopedPackage != "" {
			info, err := detection.VerifyNPMPackage(npmScopedPackage)
			if err == nil && info.Exists {
				result.npmPackage = npmScopedPackage
				result.npmVersion = info.Version
				result.npmExists = true
			} else if npmFromFile {
				// Even if not published yet, if from file, include the config
				result.npmPackage = npmScopedPackage
				result.npmExists = false
			}
		}

		if !result.npmExists && npmPackage != "" {
			info, err := detection.VerifyNPMPackage(npmPackage)
			if err == nil && info.Exists {
				result.npmPackage = npmPackage
				result.npmVersion = info.Version
				result.npmExists = true
			} else if npmFromFile {
				// Even if not published yet, if from file, include the config
				result.npmPackage = npmPackage
				result.npmExists = false
			}
		}

		return result
	}
}

// VerifyDistributionsCmd verifies existing homebrew/npm distributions
func VerifyDistributionsCmd(checkHomebrew bool, homebrewTap, homebrewFormula string, checkNPM bool, npmPackage string) tea.Cmd {
	return func() tea.Msg {
		result := distributionVerifiedMsg{}

		if checkHomebrew && homebrewTap != "" && homebrewFormula != "" {
			info, err := detection.VerifyHomebrewFormula(homebrewTap, homebrewFormula)
			if err != nil {
				result.err = err
				return result
			}
			result.homebrewExists = info.Exists
			result.homebrewVersion = info.Version
		}

		if checkNPM && npmPackage != "" {
			info, err := detection.VerifyNPMPackage(npmPackage)
			if err != nil {
				result.err = err
				return result
			}
			result.npmExists = info.Exists
			result.npmVersion = info.Version
		}

		return result
	}
}

// handleFirstTimeSetupKeys handles keyboard input for first-time setup wizard
func (m *ConfigureModel) handleFirstTimeSetupKeys(msg tea.KeyMsg) (*ConfigureModel, tea.Cmd) {
	// Handle confirmation screen separately
	if m.FirstTimeSetupConfirmation {
		switch msg.String() {
		case "esc":
			// Go back to editing
			m.FirstTimeSetupConfirmation = false
			return m, nil
		case "enter":
			// Proceed with verification
			m.VerifyingDistributions = true
			m.DistributionVerifyError = ""

			return m, tea.Batch(
				m.CreateSpinner.Tick,
				VerifyDistributionsCmd(
					m.HomebrewCheckEnabled,
					m.HomebrewTapInput.Value(),
					m.HomebrewFormulaInput.Value(),
					m.NPMCheckEnabled,
					m.NPMPackageInput.Value(),
				),
			)
		}
		return m, nil
	}

	switch msg.String() {
	case "esc":
		// Skip setup and go to normal view
		m.FirstTimeSetup = false
		m.CurrentView = TabView
		// Mark first-time setup as completed (user skipped it)
		if m.ProjectConfig != nil {
			m.ProjectConfig.FirstTimeSetupCompleted = true
			m.saveConfig()
		}
		return m, nil

	case "up", "k":
		if m.FirstTimeSetupFocus > 0 {
			m.FirstTimeSetupFocus--
			// Skip tap/formula fields if homebrew unchecked
			if !m.HomebrewCheckEnabled && m.FirstTimeSetupFocus == 2 {
				m.FirstTimeSetupFocus = 0
			}
			if !m.HomebrewCheckEnabled && m.FirstTimeSetupFocus == 1 {
				m.FirstTimeSetupFocus = 0
			}
			// Skip package field if npm unchecked
			if !m.NPMCheckEnabled && m.FirstTimeSetupFocus == 4 {
				m.FirstTimeSetupFocus = 3
			}
		}
		return m, nil

	case "down", "j":
		maxFocus := 4
		if m.FirstTimeSetupFocus < maxFocus {
			m.FirstTimeSetupFocus++
			// Skip tap/formula fields if homebrew unchecked
			if !m.HomebrewCheckEnabled && m.FirstTimeSetupFocus == 1 {
				m.FirstTimeSetupFocus = 3
			}
			if !m.HomebrewCheckEnabled && m.FirstTimeSetupFocus == 2 {
				m.FirstTimeSetupFocus = 3
			}
			// Skip package field if npm unchecked
			if !m.NPMCheckEnabled && m.FirstTimeSetupFocus == 4 {
				m.FirstTimeSetupFocus = 0
			}
		}
		return m, nil

	case "tab":
		// Cycle through all fields
		m.FirstTimeSetupFocus = (m.FirstTimeSetupFocus + 1) % 5
		// Skip disabled fields
		if !m.HomebrewCheckEnabled && (m.FirstTimeSetupFocus == 1 || m.FirstTimeSetupFocus == 2) {
			m.FirstTimeSetupFocus = 3
		}
		if !m.NPMCheckEnabled && m.FirstTimeSetupFocus == 4 {
			m.FirstTimeSetupFocus = 0
		}
		return m, nil

	case " ", "space":
		// Toggle checkboxes
		if m.FirstTimeSetupFocus == 0 {
			m.HomebrewCheckEnabled = !m.HomebrewCheckEnabled
			if m.HomebrewCheckEnabled {
				m.HomebrewTapInput.Focus()
			}
		} else if m.FirstTimeSetupFocus == 3 {
			m.NPMCheckEnabled = !m.NPMCheckEnabled
			if m.NPMCheckEnabled {
				m.NPMPackageInput.Focus()
			}
		}
		return m, nil

	case  "S":
		// Show confirmation screen if at least one is checked
		if !m.HomebrewCheckEnabled && !m.NPMCheckEnabled {
			return m, nil
		}
		m.FirstTimeSetupConfirmation = true
		return m, nil

	default:
		// Handle text input
		var cmd tea.Cmd
		if m.FirstTimeSetupFocus == 1 {
			m.HomebrewTapInput, cmd = m.HomebrewTapInput.Update(msg)
		} else if m.FirstTimeSetupFocus == 2 {
			m.HomebrewFormulaInput, cmd = m.HomebrewFormulaInput.Update(msg)
		} else if m.FirstTimeSetupFocus == 4 {
			m.NPMPackageInput, cmd = m.NPMPackageInput.Update(msg)
		}
		return m, cmd
	}
}
