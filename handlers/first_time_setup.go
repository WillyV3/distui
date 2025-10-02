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
		goreleaserConfig, err := detection.DetectGoReleaserConfig(detectedProject.Path)
		// Debug: Log detection results
		// fmt.Printf("DEBUG: GoReleaser detection - err: %v, config: %+v\n", err, goreleaserConfig)
		if err == nil && goreleaserConfig != nil && goreleaserConfig.HasHomebrew {
			homebrewTap = goreleaserConfig.HomebrewTap
			homebrewFormula = goreleaserConfig.FormulaName
			homebrewFromFile = true
			// Debug: Log detected values
			// fmt.Printf("DEBUG: Detected Homebrew - Tap: %s, Formula: %s\n", homebrewTap, homebrewFormula)
		}

		packageJSON, err := detection.DetectPackageJSON(detectedProject.Path)
		if err == nil && packageJSON != nil && packageJSON.Name != "" {
			npmPackage = packageJSON.Name
			npmScopedPackage = packageJSON.Name
			npmFromFile = true
		}
	}

	// Fall back to global config ONLY for Homebrew (not found in .goreleaser.yaml)
	// NPM: If package.json doesn't exist, assume NO npm distribution
	if homebrewTap == "" && globalConfig != nil && globalConfig.User.DefaultHomebrewTap != "" {
		homebrewTap = globalConfig.User.DefaultHomebrewTap
	}

	if homebrewFormula == "" && detectedProject.Binary != nil {
		homebrewFormula = detectedProject.Binary.Name
	}

	// NPM fallback logic REMOVED - if package.json doesn't exist, no npm distribution
	// npmPackage and npmScopedPackage stay empty if npmFromFile is false

	return DetectDistributionsCmd(homebrewTap, homebrewFormula, homebrewFromFile, npmPackage, npmScopedPackage, npmFromFile)
}

// DetectDistributionsCmd auto-detects if project exists in user's tap/npm
func DetectDistributionsCmd(homebrewTap, homebrewFormula string, homebrewFromFile bool, npmPackage, npmScopedPackage string, npmFromFile bool) tea.Cmd {
	return func() tea.Msg {
		result := distributionDetectedMsg{
			homebrewFromFile: homebrewFromFile,
			npmFromFile:      npmFromFile,
		}

		// If detected from file, always include the config
		if homebrewFromFile && homebrewTap != "" && homebrewFormula != "" {
			result.homebrewTap = homebrewTap
			result.homebrewFormula = homebrewFormula
			result.homebrewExists = false

			// Try to verify if it actually exists in registry
			info, err := detection.VerifyHomebrewFormula(homebrewTap, homebrewFormula)
			if err == nil && info.Exists {
				result.homebrewVersion = info.Version
				result.homebrewExists = true
			}
		} else if homebrewTap != "" && homebrewFormula != "" {
			// Not from file, try to verify in registry
			info, err := detection.VerifyHomebrewFormula(homebrewTap, homebrewFormula)
			if err == nil && info.Exists {
				result.homebrewTap = homebrewTap
				result.homebrewFormula = homebrewFormula
				result.homebrewVersion = info.Version
				result.homebrewExists = true
			}
		}

		// Handle NPM package detection
		// If detected from package.json, assume it's theirs (don't check registry)
		// Registry check is only for name conflict detection in Distributions tab
		if npmFromFile && npmScopedPackage != "" {
			result.npmPackage = npmScopedPackage
			result.npmExists = false // Don't check registry - it's their package.json
		} else if npmFromFile && npmPackage != "" {
			result.npmPackage = npmPackage
			result.npmExists = false // Don't check registry - it's their package.json
		} else {
			// Not from file, try to verify in registry (fallback only)
			if npmScopedPackage != "" {
				info, err := detection.VerifyNPMPackage(npmScopedPackage)
				if err == nil && info.Exists {
					result.npmPackage = npmScopedPackage
					result.npmVersion = info.Version
					result.npmExists = true
				}
			}

			if !result.npmExists && npmPackage != "" {
				info, err := detection.VerifyNPMPackage(npmPackage)
				if err == nil && info.Exists {
					result.npmPackage = npmPackage
					result.npmVersion = info.Version
					result.npmExists = true
				}
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
	// Handle custom file choice dialog first
	if m.FirstTimeSetupCustomChoice {
		switch msg.String() {
		case "k", "K":
			// Keep custom files - enter custom mode permanently
			m.FirstTimeSetupCustomChoice = false
			m.FirstTimeSetup = false
			m.CurrentView = TabView
			if m.ProjectConfig != nil {
				m.ProjectConfig.CustomFilesMode = true
				m.ProjectConfig.FirstTimeSetupCompleted = true
				m.saveConfig()
			}
			m.CustomFilesDetected = nil
			return m, nil

		case "o", "O":
			// Overwrite - switch to managed mode, continue with wizard
			m.FirstTimeSetupCustomChoice = false
			m.CustomFilesDetected = nil
			// Continue to normal first-time setup confirmation
			m.FirstTimeSetupConfirmation = true
			return m, nil

		case "esc":
			// Cancel setup
			m.FirstTimeSetupCustomChoice = false
			m.FirstTimeSetup = false
			m.CurrentView = TabView
			m.CustomFilesDetected = nil
			return m, nil
		}
		return m, nil // Consume all other inputs during choice
	}

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

	case "i", "I":
		// Import custom config (only if custom files detected)
		if m.ProjectConfig != nil && m.ProjectConfig.CustomFilesMode {
			// Set flags for custom mode
			m.ProjectConfig.CustomFilesMode = true
			m.ProjectConfig.FirstTimeSetupCompleted = true
			m.saveConfig()

			// Exit wizard, go to normal configure view
			m.CurrentView = TabView
			m.FirstTimeSetup = false
			return m, nil
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
