package handlers

import (
	"distui/internal/models"
)

func BuildDistributionsList(projectConfig *models.ProjectConfig, detectedProject *models.ProjectInfo, npmStatus string) []DistributionItem {
	items := []DistributionItem{}

	if projectConfig == nil || projectConfig.Config == nil {
		return items
	}

	// If custom mode, show distui defaults (all disabled)
	if projectConfig.CustomFilesMode {
		return []DistributionItem{
			{Name: "GitHub Releases", Desc: "Create GitHub releases with GoReleaser", Enabled: false, Key: "github"},
			{Name: "Homebrew", Desc: "Publish to Homebrew tap", Enabled: false, Key: "homebrew"},
			{Name: "NPM", Desc: "Publish to NPM registry", Enabled: false, Key: "npm", Status: npmStatus},
			{Name: "Go Install", Desc: "Installable via go install", Enabled: false, Key: "go_install"},
		}
	}

	// GitHub Releases
	githubEnabled := false
	githubDesc := "Upload binaries to GitHub releases"
	if projectConfig.Config.Distributions.GitHubRelease != nil {
		githubEnabled = projectConfig.Config.Distributions.GitHubRelease.Enabled
	}
	items = append(items, DistributionItem{
		Name:    "GitHub Releases",
		Desc:    githubDesc,
		Enabled: githubEnabled,
		Key:     "github",
	})

	// Homebrew
	homebrewEnabled := false
	homebrewDesc := "Publish to Homebrew tap"
	if projectConfig.Config.Distributions.Homebrew != nil {
		homebrewEnabled = projectConfig.Config.Distributions.Homebrew.Enabled
		if projectConfig.Config.Distributions.Homebrew.TapRepo != "" {
			homebrewDesc = "Tap: " + projectConfig.Config.Distributions.Homebrew.TapRepo
		}
	}
	items = append(items, DistributionItem{
		Name:    "Homebrew",
		Desc:    homebrewDesc,
		Enabled: homebrewEnabled,
		Key:     "homebrew",
	})

	// NPM - clean description without status clutter
	npmEnabled := false
	npmDesc := "Publish to NPM registry"
	if projectConfig.Config.Distributions.NPM != nil {
		npmEnabled = projectConfig.Config.Distributions.NPM.Enabled
		if projectConfig.Config.Distributions.NPM.PackageName != "" {
			npmDesc = "Package: " + projectConfig.Config.Distributions.NPM.PackageName
		} else if detectedProject != nil && detectedProject.Binary != nil && detectedProject.Binary.Name != "" {
			npmDesc = "Package: " + detectedProject.Binary.Name
		}
	}
	items = append(items, DistributionItem{
		Name:    "NPM",
		Desc:    npmDesc,
		Enabled: npmEnabled,
		Key:     "npm",
		Status:  npmStatus,
	})

	// Go Module
	goModuleEnabled := false
	goModuleDesc := "Install via go install (automatic with git tags)"
	if projectConfig.Config.Distributions.GoModule != nil {
		goModuleEnabled = projectConfig.Config.Distributions.GoModule.Enabled
	}
	items = append(items, DistributionItem{
		Name:    "Go Module",
		Desc:    goModuleDesc,
		Enabled: goModuleEnabled,
		Key:     "go_install",
	})

	return items
}