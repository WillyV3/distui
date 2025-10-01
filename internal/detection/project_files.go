package detection

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type GoReleaserConfig struct {
	HasHomebrew bool
	HomebrewTap string
	FormulaName string
	HasNPM      bool
	NPMPackage  string
}

type PackageJSON struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func DetectGoReleaserConfig(projectPath string) (*GoReleaserConfig, error) {
	config := &GoReleaserConfig{}

	goreleaserPath := filepath.Join(projectPath, ".goreleaser.yaml")
	if _, err := os.Stat(goreleaserPath); err != nil {
		goreleaserPath = filepath.Join(projectPath, ".goreleaser.yml")
		if _, err := os.Stat(goreleaserPath); err != nil {
			return config, nil
		}
	}

	data, err := os.ReadFile(goreleaserPath)
	if err != nil {
		return config, err
	}

	var goreleaserConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &goreleaserConfig); err != nil {
		return config, err
	}

	// Check for brews section (Homebrew)
	if brews, ok := goreleaserConfig["brews"].([]interface{}); ok && len(brews) > 0 {
		config.HasHomebrew = true

		if brew, ok := brews[0].(map[string]interface{}); ok {
			if repository, ok := brew["repository"].(map[string]interface{}); ok {
				if owner, ok := repository["owner"].(string); ok {
					if name, ok := repository["name"].(string); ok {
						config.HomebrewTap = owner + "/" + name
					}
				}
			}

			if name, ok := brew["name"].(string); ok {
				config.FormulaName = name
			}
		}
	}

	// Check for publishers section (NPM)
	if publishers, ok := goreleaserConfig["publishers"].([]interface{}); ok {
		for _, pub := range publishers {
			if pubMap, ok := pub.(map[string]interface{}); ok {
				if cmd, ok := pubMap["cmd"].(string); ok {
					if strings.Contains(cmd, "npm publish") {
						config.HasNPM = true
						break
					}
				}
			}
		}
	}

	return config, nil
}

func DetectPackageJSON(projectPath string) (*PackageJSON, error) {
	pkgPath := filepath.Join(projectPath, "package.json")
	if _, err := os.Stat(pkgPath); err != nil {
		return nil, nil
	}

	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil, err
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	return &pkg, nil
}
