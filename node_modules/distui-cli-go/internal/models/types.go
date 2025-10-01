package models

import "time"

type GlobalConfig struct {
	Version     string      `yaml:"version"`
	User        UserConfig  `yaml:"user"`
	Preferences Preferences `yaml:"preferences"`
	UI          UIConfig    `yaml:"ui"`
	Paths       PathsConfig `yaml:"paths"`
}

type GitHubAccount struct {
	Username string `yaml:"username"`
	IsOrg    bool   `yaml:"is_org,omitempty"`
	Default  bool   `yaml:"default,omitempty"`
}

type UserConfig struct {
	GitHubUsername    string          `yaml:"github_username"` // Primary account (backwards compat)
	GitHubAccounts    []GitHubAccount `yaml:"github_accounts,omitempty"` // Multiple accounts/orgs
	DefaultHomebrewTap string         `yaml:"default_homebrew_tap,omitempty"`
	NPMScope          string         `yaml:"npm_scope,omitempty"`
}

type Preferences struct {
	ConfirmBeforeRelease bool   `yaml:"confirm_before_release"`
	DefaultVersionBump   string `yaml:"default_version_bump"`
	ShowCommandOutput    bool   `yaml:"show_command_output"`
	AutoDetectProjects   bool   `yaml:"auto_detect_projects"`
}

type UIConfig struct {
	Theme       string `yaml:"theme"`
	CompactMode bool   `yaml:"compact_mode"`
	ShowHints   bool   `yaml:"show_hints"`
}

type PathsConfig struct {
	HomebrewTapLocation string `yaml:"homebrew_tap_location"`
	GoreleaserConfig    string `yaml:"goreleaser_config"`
}

type ProjectConfig struct {
	Project *ProjectInfo     `yaml:"project"`
	Config  *ProjectSettings `yaml:"config"`
	History *ReleaseHistory  `yaml:"history,omitempty"`
}

type ProjectInfo struct {
	Identifier   string          `yaml:"identifier"`
	Path         string          `yaml:"path"`
	LastAccessed *time.Time      `yaml:"last_accessed,omitempty"`
	DetectedAt   *time.Time      `yaml:"detected_at,omitempty"`
	Repository   *RepositoryInfo `yaml:"repository"`
	Module       *ModuleInfo     `yaml:"module"`
	Binary       *BinaryInfo     `yaml:"binary,omitempty"`
}

type RepositoryInfo struct {
	Owner         string `yaml:"owner"`
	Name          string `yaml:"name"`
	DefaultBranch string `yaml:"default_branch,omitempty"`
}

type ModuleInfo struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version,omitempty"`
}

type BinaryInfo struct {
	Name       string   `yaml:"name"`
	BuildFlags []string `yaml:"build_flags,omitempty"`
}

type ProjectSettings struct {
	Distributions Distributions    `yaml:"distributions"`
	Build         *BuildSettings   `yaml:"build,omitempty"`
	Release       *ReleaseSettings `yaml:"release,omitempty"`
	CICD          *CICDSettings    `yaml:"ci_cd,omitempty"`
}

type Distributions struct {
	GitHubRelease *GitHubReleaseConfig `yaml:"github_release,omitempty"`
	Homebrew      *HomebrewConfig      `yaml:"homebrew,omitempty"`
	NPM           *NPMConfig           `yaml:"npm,omitempty"`
	GoModule      *GoModuleConfig      `yaml:"go_module,omitempty"`
}

type GitHubReleaseConfig struct {
	Enabled    bool `yaml:"enabled"`
	Draft      bool `yaml:"draft,omitempty"`
	Prerelease bool `yaml:"prerelease,omitempty"`
}

type HomebrewConfig struct {
	Enabled     bool   `yaml:"enabled"`
	TapRepo     string `yaml:"tap_repo,omitempty"`
	TapPath     string `yaml:"tap_path,omitempty"`
	FormulaName string `yaml:"formula_name,omitempty"`
	FormulaPath string `yaml:"formula_path,omitempty"`
}

type NPMConfig struct {
	Enabled     bool   `yaml:"enabled"`
	PackageName string `yaml:"package_name,omitempty"`
	Registry    string `yaml:"registry,omitempty"`
	Access      string `yaml:"access,omitempty"`
}

type GoModuleConfig struct {
	Enabled bool   `yaml:"enabled"`
	Proxy   string `yaml:"proxy,omitempty"`
}

type BuildSettings struct {
	GoreleaserConfig string `yaml:"goreleaser_config,omitempty"`
	TestCommand      string `yaml:"test_command,omitempty"`
}

type ReleaseSettings struct {
	SkipTests         bool `yaml:"skip_tests"`
	CreateDraft       bool `yaml:"create_draft"`
	PreRelease        bool `yaml:"pre_release"`
	GenerateChangelog bool `yaml:"generate_changelog"`
	SignCommits       bool `yaml:"sign_commits"`
}

type CICDSettings struct {
	GitHubActions *GitHubActionsConfig `yaml:"github_actions,omitempty"`
}

type GitHubActionsConfig struct {
	Enabled      bool   `yaml:"enabled"`
	WorkflowPath string `yaml:"workflow_path,omitempty"`
}

type ReleaseHistory struct {
	Releases []ReleaseRecord `yaml:"releases,omitempty"`
}

type ReleaseRecord struct {
	Version  string                 `yaml:"version"`
	Date     time.Time              `yaml:"date"`
	Method   string                 `yaml:"method,omitempty"`
	Duration string                 `yaml:"duration,omitempty"`
	Status   string                 `yaml:"status"`
	Channels map[string]bool        `yaml:"channels,omitempty"`
	Error    string                 `yaml:"error,omitempty"`
}