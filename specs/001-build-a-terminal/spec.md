# Feature Specification: distui - Go Release Distribution Manager

**Feature Branch**: `001-build-a-terminal`
**Created**: 2025-09-28
**Updated**: 2025-09-29
**Status**: In Progress - Git Management Complete, Release Workflow Pending
**Input**: User description update: "Complete the release workflow implementation including GoReleaser integration, Homebrew tap updates, NPM publishing, and animated progress display. Build release execution flow using package-manager example from charm-examples-inventory for animated progress. Support both interactive TUI releases and GitHub Actions workflow generation for CI/CD. Preserve all completed git management features (smart commit, file categorization, GitHub repo creation, cleanup view) while implementing release execution."

## Execution Flow (main)
```
1. Parse user description from Input
   → Feature description provided and parsed
2. Extract key concepts from description
   → Actors: Go developers managing releases
   → Actions: detect projects, configure distributions, execute releases
   → Data: project configs, distribution settings, release history
   → Constraints: 30-second releases, no repo pollution, user control
3. For each unclear aspect:
   → All core aspects specified in description
4. Fill User Scenarios & Testing section
   → Clear user flows for project detection, configuration, and releases
5. Generate Functional Requirements
   → Each requirement is testable and measurable
6. Identify Key Entities
   → Projects, Configurations, Distribution Channels, Release History
7. Run Review Checklist
   → All items addressed
8. Return: SUCCESS (spec ready for planning)
```

---

## ⚡ Quick Guidelines
- ✅ Focus on WHAT users need and WHY
- ❌ Avoid HOW to implement (no tech stack, APIs, code structure)
- 👥 Written for business stakeholders, not developers

### Section Requirements
- **Mandatory sections**: Must be completed for every feature
- **Optional sections**: Include only when relevant to the feature
- When a section doesn't apply, remove it entirely (don't leave as "N/A")

---

## User Scenarios & Testing

### Primary User Story
As a Go developer with multiple projects, I want to manage all my release distributions from a single TUI application that remembers my project configurations globally, so I can release any project in under 30 seconds without polluting my repositories with configuration files.

### Acceptance Scenarios

#### Completed Features ✅
1. **Given** I'm in a Go project directory, **When** I launch distui, **Then** the app detects the project from git remotes and go.mod and loads any existing configuration from ~/.distui ✅
2. **Given** I have multiple Go projects configured, **When** I press TAB in the TUI, **Then** I can cycle between Project view, Global view (all projects), and Settings view ✅
3. **Given** I'm setting up a new project, **When** the app detects repository info, **Then** I can override any detected values before saving the configuration ✅
4. **Given** I have uncommitted changes, **When** I view the cleanup tab, **Then** I can see file categorization (config, code, docs, etc.) with smart suggestions for commit/ignore ✅
5. **Given** I want to commit changes, **When** I use smart commit, **Then** files are categorized automatically and committed with AI-generated messages ✅
6. **Given** I need a GitHub repository, **When** I create it from the TUI, **Then** I can select personal/org account, set visibility, and have the remote added automatically ✅
7. **Given** I have unpushed commits, **When** I view status, **Then** I see commit count and can push with [P] key ✅
8. **Given** the repository is clean and synced, **When** I view status, **Then** I see "All synced!" message ✅

#### Pending Features 🔄
9. **Given** I'm in Project view with a configured project, **When** I initiate a release, **Then** the app executes tests, creates tags, runs GoReleaser, updates Homebrew tap, and optionally publishes to NPM within 30 seconds
10. **Given** a release is in progress, **When** commands execute, **Then** I see animated progress with package-manager style UI showing each step
11. **Given** a release is in progress, **When** a command fails, **Then** I see the error output in real-time and can choose how to proceed interactively
12. **Given** I have projects with different distribution needs, **When** I configure each project, **Then** each maintains its own distribution channel settings (GitHub only, Homebrew+GitHub, NPM+GitHub, etc.)
13. **Given** I prefer CI/CD over local releases, **When** I configure the project, **Then** I can generate GitHub Actions workflows for automated releases

### Edge Cases
- What happens when gh CLI is not installed or not authenticated?
- How does system handle when homebrew tap location doesn't exist?
- What occurs if project has no go.mod file?
- How does app respond when ~/.distui directory is not writable?
- What happens during concurrent releases of different projects?
- How does system handle when detected values conflict with saved configuration?

## Requirements

### Functional Requirements

#### Core Infrastructure (Completed ✅)
- **FR-001**: System MUST detect Go projects by reading git remote origin and go.mod files in the current directory ✅
- **FR-002**: System MUST store all configuration data in ~/.distui directory, never adding files to user repositories ✅
- **FR-003**: System MUST provide three navigable views: Project (current project), Global (all projects list), and Settings ✅
- **FR-004**: Users MUST be able to switch between views using TAB key or dedicated keyboard shortcuts ✅
- **FR-007**: System MUST remember project configurations globally, allowing instant access when revisiting projects ✅
- **FR-009**: System MUST use gh CLI for smart detection of repositories, taps, and authentication status ✅
- **FR-010**: Users MUST be able to override any automatically detected values ✅

#### Git Management (Completed ✅)
- **FR-019**: System MUST categorize uncommitted files by type (config, code, docs, build, assets, data, other) ✅
- **FR-020**: System MUST allow per-file actions (commit, skip, ignore) in cleanup view ✅
- **FR-021**: System MUST provide smart commit with AI-generated commit messages based on file changes ✅
- **FR-022**: System MUST support creating GitHub repositories from TUI with account/org selection ✅
- **FR-023**: System MUST detect unpushed commits and provide push action ✅
- **FR-024**: System MUST show "All synced!" when repository is clean and pushed ✅
- **FR-025**: System MUST support file-by-file commit workflow with custom messages ✅
- **FR-026**: System MUST provide repository browser for navigating project files ✅

#### Release Workflow (Pending 🔄)
- **FR-005**: System MUST execute release processes directly from the TUI including: running tests, creating git tags, executing GoReleaser, updating Homebrew formulas, and publishing to NPM
- **FR-006**: System MUST complete typical release execution within 30 seconds
- **FR-008**: System MUST support configuration of distribution channels per project (GitHub releases, Homebrew tap settings, NPM package settings)
- **FR-011**: System MUST display real-time command output during release execution with package-manager style animated progress
- **FR-012**: System MUST provide interactive error handling when commands fail
- **FR-013**: System MUST maintain release history for each project
- **FR-014**: System MUST support generating GitHub Actions workflows for CI/CD as an alternative to local execution
- **FR-015**: System MUST allow configuration of Homebrew tap location and formula names per project
- **FR-016**: System MUST allow configuration of NPM package names and settings per project
- **FR-017**: System MUST work with projects that have some but not all distribution channels enabled
- **FR-018**: System MUST provide visual progress indicators during long-running operations using package-manager example pattern
- **FR-027**: System MUST bump version numbers automatically or allow user to specify version
- **FR-028**: System MUST validate GoReleaser configuration before executing release
- **FR-029**: System MUST support rollback if release fails mid-process
- **FR-030**: System MUST show live streaming output for each release step (tests, build, publish)

### Key Entities
- **Project**: Represents a Go application with its repository information, module path, and current version
- **Configuration**: Project-specific settings including distribution channels, build preferences, and release options
- **Distribution Channel**: A method of distributing releases (GitHub Releases, Homebrew, NPM) with its specific settings
- **Release History**: Record of past releases including version, date, method (local/CI), and status
- **Global Settings**: User-wide preferences including default Homebrew tap location, NPM scope, and UI preferences
- **File Change**: Represents uncommitted file with path, status (modified/added/deleted/untracked), and category
- **Cleanup Item**: UI representation of file with action (commit/skip/ignore)
- **Commit Model**: State for commit view including selected files and commit message

### Reference Materials
- **Charm Examples**: `/Users/williamvansickleiii/charmtuitemplate/distui/docs/charm-examples-inventory/` - Source of truth for TUI patterns
- **Package Manager Example**: `/Users/williamvansickleiii/charmtuitemplate/distui/docs/charm-examples-inventory/bubbletea/examples/package-manager` - Pattern for animated release progress
- **GoReleaser Examples**: `/Users/williamvansickleiii/charmtuitemplate/distui/goreleaser-examples` - Reference configurations
- **Brew Docs**: `/Users/williamvansickleiii/charmtuitemplate/distui/docs/brew-docs.md` - Homebrew tap management
- **NPM Release**: `/Users/williamvansickleiii/charmtuitemplate/distui/docs/npm-release.md` - NPM publishing workflow

---

## Review & Acceptance Checklist

### Content Quality
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

### Requirement Completeness
- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

---

## Execution Status

- [x] User description parsed (original + update)
- [x] Key concepts extracted
- [x] Ambiguities marked (none found)
- [x] User scenarios defined (split into completed/pending)
- [x] Requirements generated (categorized by status)
- [x] Entities identified (including new git management entities)
- [x] Review checklist passed
- [x] Reference materials documented

## Implementation Progress

### Completed (2025-09-28 to 2025-09-29)
- ✅ Full TUI infrastructure with 6 views
- ✅ Configuration management with ~/.distui storage
- ✅ Project detection from git/go.mod
- ✅ Git cleanup view with file categorization
- ✅ Smart commit with AI-generated messages
- ✅ GitHub repository creation from TUI
- ✅ File-by-file commit workflow
- ✅ Repository browser for file navigation
- ✅ Push detection and execution
- ✅ Repository status display with sync indicator
- ✅ Async loading with spinner for configure view

### Remaining (Release Workflow & UI Polish)
- 🔄 **Project view redesign** - Two states (unconfigured/configured) showing dashboard with git status, distributions, last release
- 🔄 Version bumping logic
- 🔄 GoReleaser integration and execution
- 🔄 Homebrew tap updates
- 🔄 NPM publishing
- 🔄 Animated progress UI (package-manager pattern)
- 🔄 Release history tracking and display
- 🔄 GitHub Actions workflow generation
- 🔄 Interactive error handling during releases
- 🔄 Rollback on release failure

---