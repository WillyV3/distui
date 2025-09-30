# Data Model: distui

## Core Entities

### GlobalConfig
**Purpose**: User-wide settings and preferences
```yaml
user:
  github_username: string
  default_homebrew_tap: string
  npm_scope: string (optional)

preferences:
  confirm_before_release: boolean
  default_version_bump: "patch" | "minor" | "major"
  show_command_output: boolean

ui:
  theme: "default" | "dark" | "light"
  compact_mode: boolean
```

### Project
**Purpose**: Represents a Go project managed by distui
```yaml
identifier: string          # github.com/user/repo
path: string                # /absolute/path/to/project
last_accessed: timestamp
detected_at: timestamp

repository:
  owner: string             # GitHub owner
  name: string              # Repository name
  default_branch: string    # main, master, etc.

module:
  name: string              # From go.mod
  version: string           # Current version

binary:
  name: string              # Output binary name
  build_flags: []string     # Custom build flags
```

### ProjectConfig
**Purpose**: Project-specific distribution settings
```yaml
distributions:
  github_release:
    enabled: boolean
    draft: boolean
    prerelease: boolean

  homebrew:
    enabled: boolean
    tap_repo: string        # username/homebrew-tap
    tap_path: string        # ~/homebrew-tap
    formula_name: string    # tool.rb
    formula_path: string    # Formula/tool.rb

  npm:
    enabled: boolean
    package_name: string    # @username/tool
    registry: string        # https://registry.npmjs.org
    access: "public" | "private"

  go_module:
    enabled: boolean
    proxy: string           # proxy.golang.org

build:
  goreleaser_config: string # Path to .goreleaser.yaml
  test_command: string      # go test ./...

smart_commit:
  enabled: boolean
  use_custom_rules: boolean
  categories:
    config:
      extensions: []string  # [".yaml", ".yml", ".json", ...]
      patterns: []string    # ["**/config/**", ...]
    code:
      extensions: []string
      patterns: []string
    docs:
      extensions: []string
      patterns: []string
    build:
      extensions: []string
      patterns: []string
    test:
      extensions: []string
      patterns: []string
    assets:
      extensions: []string
      patterns: []string
    data:
      extensions: []string
      patterns: []string

ci_cd:
  github_actions:
    enabled: boolean
    workflow_path: string        # .github/workflows/release.yml
    auto_regenerate: boolean     # Regenerate on config changes
    include_tests: boolean       # Run tests in workflow
    environments: []string       # ["production", "staging"]
    secrets_required: []string   # ["NPM_TOKEN", ...]
```

### ReleaseHistory
**Purpose**: Track releases for each project
```yaml
releases:
  - version: string         # v1.2.3
    date: timestamp
    method: "local" | "ci"
    duration: duration      # 28s
    status: "success" | "failed" | "partial"
    channels:              # Which distributions succeeded
      - github: boolean
      - homebrew: boolean
      - npm: boolean
    error: string (optional)
```

### DistributionChannel
**Purpose**: Abstract interface for distribution methods
```go
type DistributionChannel interface {
    Name() string
    Enabled() bool
    Validate() error
    Execute(version string) error
    Rollback(version string) error
}
```

## Relationships

```
GlobalConfig (1) ─────────────┐
                              │
Project (1) ──────> (1) ProjectConfig
    │                         │
    └──> (many) ReleaseHistory
                              │
DistributionChannel (abstract) <──┤
    ├── GitHubRelease          │
    ├── HomebrewTap            │
    ├── NPMPackage             │
    └── GoModule               │
```

## State Management

### Application State
```go
type AppState struct {
    CurrentView    ViewType    // Project, Global, Settings
    CurrentProject *Project    // nil if in global view
    Projects       []Project   // All known projects
    GlobalConfig   GlobalConfig

    // UI State
    SelectedIndex  int         // For list navigation
    SearchQuery    string      // For filtering
    Modal          *Modal      // Active modal if any

    // Release State
    ReleaseInProgress bool
    ReleaseOutput     []string  // Command output buffer
    ReleaseProgress   float64   // 0.0 to 1.0
}
```

### View Types
```go
type ViewType int

const (
    ProjectView ViewType = iota
    GlobalView
    SettingsView
    ReleaseView    // Active during release
)
```

## Storage Schema

### File Structure
```
~/.distui/
├── config.yaml                      # GlobalConfig
├── projects/
│   ├── github-com-user-repo1.yaml  # Project + ProjectConfig
│   ├── github-com-user-repo2.yaml
│   └── ...
├── history/
│   ├── github-com-user-repo1.yaml  # ReleaseHistory
│   └── ...
└── cache/
    ├── detections.json              # Cached detection results
    └── tokens.enc                   # Encrypted tokens (optional)
```

### Project File Schema
```yaml
# ~/.distui/projects/github-com-user-repo.yaml
project:
  identifier: github.com/user/repo
  path: /Users/me/code/repo
  last_accessed: 2025-09-28T10:00:00Z
  # ... rest of Project fields

config:
  distributions:
    # ... ProjectConfig fields

history:
  releases:
    # ... Last 10 releases only
```

## Data Operations

### CRUD Operations
```go
// Config operations
LoadGlobalConfig() (*GlobalConfig, error)
SaveGlobalConfig(config *GlobalConfig) error

// Project operations
LoadProject(identifier string) (*Project, error)
SaveProject(project *Project) error
ListProjects() ([]*Project, error)
DeleteProject(identifier string) error

// History operations
AddRelease(projectID string, release *Release) error
GetReleaseHistory(projectID string, limit int) ([]*Release, error)
```

### Atomic Operations
- File writes use temp file + rename pattern
- Exclusive file locks during updates
- Transactional updates for multi-file changes

## Validation Rules

### Project Validation
- `identifier` must match `github.com/owner/repo` pattern
- `path` must exist and contain go.mod
- `binary.name` must be valid filename

### Config Validation
- `homebrew.tap_repo` must match `owner/repo` pattern
- `homebrew.tap_path` must be absolute path
- `npm.package_name` must follow NPM naming rules
- At least one distribution channel must be enabled

### Version Validation
- Must follow semantic versioning (v1.2.3)
- Must be greater than previous version
- Must not already exist as tag

## Migration Strategy

### Version 1.0 Schema
- Initial schema as defined above
- No migration needed for new installations

### Future Migrations
- Schema version in config.yaml
- Automatic backup before migration
- Rollback capability on migration failure
- Migration applied on first run after upgrade