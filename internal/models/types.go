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
	Project                 *ProjectInfo     `yaml:"project"`
	Config                  *ProjectSettings `yaml:"config"`
	History                 *ReleaseHistory  `yaml:"history,omitempty"`
	FirstTimeSetupCompleted bool             `yaml:"first_time_setup_completed,omitempty"`
	CustomFilesMode         bool             `yaml:"custom_files_mode,omitempty"`
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
	Distributions Distributions      `yaml:"distributions"`
	Build         *BuildSettings     `yaml:"build,omitempty"`
	Release       *ReleaseSettings   `yaml:"release,omitempty"`
	SmartCommit   *SmartCommitPrefs  `yaml:"smart_commit,omitempty"`
	CICD          *CICDSettings      `yaml:"ci_cd,omitempty"`
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
	Enabled          bool     `yaml:"enabled"`
	WorkflowPath     string   `yaml:"workflow_path,omitempty"`
	AutoRegenerate   bool     `yaml:"auto_regenerate,omitempty"`
	IncludeTests     bool     `yaml:"include_tests,omitempty"`
	Environments     []string `yaml:"environments,omitempty"`
	SecretsRequired  []string `yaml:"secrets_required,omitempty"`
}

type CategoryRules struct {
	Extensions []string `yaml:"extensions"`
	Patterns   []string `yaml:"patterns"`
}

type SmartCommitPrefs struct {
	Enabled        bool                     `yaml:"enabled"`
	UseCustomRules bool                     `yaml:"use_custom_rules"`
	Categories     map[string]CategoryRules `yaml:"categories"`
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

type FileCategoryRule struct {
	Pattern  string `yaml:"pattern"`
	Category string `yaml:"category"`
	Priority int    `yaml:"priority"`
}

type SmartCommitPreferences struct {
	Enabled     bool               `yaml:"enabled"`
	CustomRules []FileCategoryRule `yaml:"custom_rules,omitempty"`
}

type FlaggedFile struct {
	Path            string    `yaml:"path"`
	IssueType       string    `yaml:"issue_type"`
	SizeBytes       int64     `yaml:"size_bytes"`
	SuggestedAction string    `yaml:"suggested_action"`
	FlaggedAt       time.Time `yaml:"flagged_at"`
}

type CleanupScanResult struct {
	MediaFiles     []FlaggedFile `yaml:"media_files"`
	ExcessDocs     []FlaggedFile `yaml:"excess_docs"`
	DevArtifacts   []FlaggedFile `yaml:"dev_artifacts"`
	TotalSizeBytes int64         `yaml:"total_size_bytes"`
	ScanDuration   time.Duration `yaml:"scan_duration"`
	ScannedAt      time.Time     `yaml:"scanned_at"`
}

type BranchInfo struct {
	Name           string `yaml:"name"`
	IsCurrent      bool   `yaml:"is_current"`
	TrackingBranch string `yaml:"tracking_branch"`
	AheadCount     int    `yaml:"ahead_count"`
	BehindCount    int    `yaml:"behind_count"`
}

type BranchSelectionModal struct {
	Branches      []BranchInfo `yaml:"branches"`
	SelectedIndex int          `yaml:"selected_index"`
	FilterQuery   string       `yaml:"filter_query"`
	Width         int          `yaml:"width"`
	Height        int          `yaml:"height"`
}

type UINotification struct {
	Message   string    `yaml:"message"`
	ShowUntil time.Time `yaml:"show_until"`
	Style     string    `yaml:"style"`
}