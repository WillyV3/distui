package detection

import (
	"path/filepath"
)

// DetectProjectMode determines if project is using custom config files
// and whether first-time setup is needed.
// Returns (customMode, needsSetup, error).
func DetectProjectMode(projectPath string) (bool, bool, error) {
	goreleaserPath := filepath.Join(projectPath, ".goreleaser.yaml")
	goreleaserYmlPath := filepath.Join(projectPath, ".goreleaser.yml")
	packageJSONPath := filepath.Join(projectPath, "package.json")

	hasGoreleaser := FileExists(goreleaserPath) || FileExists(goreleaserYmlPath)
	hasPackageJSON := FileExists(packageJSONPath)

	// No files = fresh project, needs setup
	if !hasGoreleaser && !hasPackageJSON {
		return false, true, nil
	}

	// Check if files are custom (not distui-generated)
	customGoreleaser := IsCustomConfig(goreleaserPath) || IsCustomConfig(goreleaserYmlPath)
	customPackageJSON := IsCustomConfig(packageJSONPath)

	// If any custom files exist, enter custom mode
	if customGoreleaser || customPackageJSON {
		return true, true, nil // Custom mode, needs setup (to set flags)
	}

	// Distui-managed files exist, no setup needed
	return false, false, nil
}
