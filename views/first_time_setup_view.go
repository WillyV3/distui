package views

import (
	"fmt"
	"strings"

	"distui/handlers"
	"github.com/charmbracelet/lipgloss"
)

func RenderFirstTimeSetup(model *handlers.ConfigureModel) string {
	if model == nil {
		return "Loading..."
	}

	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)

	var content strings.Builder

	content.WriteString(headerStyle.Render("FIRST-TIME SETUP"))
	content.WriteString("\n\n")

	// Show detection spinner
	if model.DetectingDistributions {
		content.WriteString(spinnerStyle.Render(model.CreateSpinner.View()))
		content.WriteString(" Detecting existing distributions...")
		content.WriteString("\n\n")
		content.WriteString(dimStyle.Render("Checking your Homebrew tap and NPM packages..."))
		return content.String()
	}

	// Show confirmation screen
	if model.FirstTimeSetupConfirmation {
		if model.AutoDetected {
			content.WriteString(successStyle.Render("AUTO-DETECTED DISTRIBUTIONS"))
		} else {
			content.WriteString(warningStyle.Render("CONFIRM DISTRIBUTION SETTINGS"))
		}
		content.WriteString("\n\n")

		if model.AutoDetected {
			content.WriteString(labelStyle.Render("We found this project in the following distributions:"))
		} else {
			content.WriteString(labelStyle.Render("The following distributions will be verified:"))
		}
		content.WriteString("\n\n")

		if model.HomebrewCheckEnabled {
			if model.AutoDetected {
				if model.HomebrewDetectedFromFile {
					content.WriteString(successStyle.Render("✓ Found in .goreleaser.yaml"))
				} else {
					content.WriteString(successStyle.Render("✓ Found in Homebrew"))
				}
			} else {
				content.WriteString(successStyle.Render("✓ Homebrew Formula"))
			}
			content.WriteString("\n")
			content.WriteString(dimStyle.Render("  Tap:     " + model.HomebrewTapInput.Value()))
			content.WriteString("\n")
			content.WriteString(dimStyle.Render("  Formula: " + model.HomebrewFormulaInput.Value()))
			content.WriteString("\n\n")
		}

		if model.NPMCheckEnabled {
			if model.AutoDetected {
				if model.NPMDetectedFromFile {
					content.WriteString(successStyle.Render("✓ Found in package.json"))
				} else {
					content.WriteString(successStyle.Render("✓ Found in NPM"))
				}
			} else {
				content.WriteString(successStyle.Render("✓ NPM Package"))
			}
			content.WriteString("\n")
			content.WriteString(dimStyle.Render("  Package: " + model.NPMPackageInput.Value()))
			content.WriteString("\n\n")
		}

		content.WriteString("\n")

		if model.AutoDetected {
			content.WriteString(labelStyle.Render("Current versions will be imported from:"))
		} else {
			content.WriteString(labelStyle.Render("This will run:"))
		}
		content.WriteString("\n")
		if model.HomebrewCheckEnabled {
			content.WriteString(dimStyle.Render("  brew info " + model.HomebrewTapInput.Value() + "/" + model.HomebrewFormulaInput.Value() + " --json=v2"))
			content.WriteString("\n")
		}
		if model.NPMCheckEnabled {
			content.WriteString(dimStyle.Render("  npm view " + model.NPMPackageInput.Value() + " version"))
			content.WriteString("\n")
		}
		content.WriteString("\n")

		content.WriteString(successStyle.Render("[Enter] Confirm & Import"))
		content.WriteString("  ")
		content.WriteString(dimStyle.Render("[Esc] Go Back"))

		return content.String()
	}

	if model.VerifyingDistributions {
		content.WriteString(spinnerStyle.Render(model.CreateSpinner.View()))
		content.WriteString(" Verifying distributions...")
		content.WriteString("\n\n")
		content.WriteString(dimStyle.Render("This may take a moment..."))
		return content.String()
	}

	if model.DistributionVerifyError != "" {
		content.WriteString(errorStyle.Render("Error: " + model.DistributionVerifyError))
		content.WriteString("\n\n")
		content.WriteString(dimStyle.Render("[Enter] Try again  [Esc] Skip setup"))
		return content.String()
	}

	versionInfo := ""
	if model.ProjectConfig != nil && model.ProjectConfig.Project != nil && model.ProjectConfig.Project.Module != nil {
		versionInfo = model.ProjectConfig.Project.Module.Version
	}

	content.WriteString(labelStyle.Render(fmt.Sprintf("We detected this project has versions (current: %s)", versionInfo)))
	content.WriteString("\n")
	content.WriteString(dimStyle.Render("Is this project already distributed on Homebrew or NPM?"))
	content.WriteString("\n\n")

	homebrewCheckbox := "[ ]"
	if model.HomebrewCheckEnabled {
		homebrewCheckbox = "[✓]"
	}

	npmCheckbox := "[ ]"
	if model.NPMCheckEnabled {
		npmCheckbox = "[✓]"
	}

	focus := model.FirstTimeSetupFocus

	if focus == 0 {
		content.WriteString(selectedStyle.Render("→ " + homebrewCheckbox + " Homebrew"))
	} else {
		content.WriteString(labelStyle.Render("  " + homebrewCheckbox + " Homebrew"))
	}
	content.WriteString("\n")

	if model.HomebrewCheckEnabled {
		content.WriteString("\n")

		if focus == 1 {
			content.WriteString(selectedStyle.Render("    Tap:     "))
		} else {
			content.WriteString(labelStyle.Render("    Tap:     "))
		}
		content.WriteString(model.HomebrewTapInput.View())
		content.WriteString("\n")

		if focus == 2 {
			content.WriteString(selectedStyle.Render("    Formula: "))
		} else {
			content.WriteString(labelStyle.Render("    Formula: "))
		}
		content.WriteString(model.HomebrewFormulaInput.View())
		content.WriteString("\n")
	}

	content.WriteString("\n")

	if focus == 3 {
		content.WriteString(selectedStyle.Render("→ " + npmCheckbox + " NPM"))
	} else {
		content.WriteString(labelStyle.Render("  " + npmCheckbox + " NPM"))
	}
	content.WriteString("\n")

	if model.NPMCheckEnabled {
		content.WriteString("\n")

		if focus == 4 {
			content.WriteString(selectedStyle.Render("    Package: "))
		} else {
			content.WriteString(labelStyle.Render("    Package: "))
		}
		content.WriteString(model.NPMPackageInput.View())
		content.WriteString("\n")
	}

	content.WriteString("\n\n")

	if model.HomebrewCheckEnabled || model.NPMCheckEnabled {
		content.WriteString(successStyle.Render("[S] Save & Continue"))
	} else {
		content.WriteString(dimStyle.Render("[S] Save & Continue (select at least one)"))
	}

	content.WriteString("\n")
	content.WriteString(dimStyle.Render("[Esc] Skip - Configure Manually"))
	content.WriteString("\n\n")
	content.WriteString(dimStyle.Render("Navigation: [↑↓] Move  [Space] Toggle  [Tab] Next field"))

	return content.String()
}
