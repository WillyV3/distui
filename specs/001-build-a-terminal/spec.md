# Feature Specification: distui - Go Release Distribution Manager

**Feature Branch**: `001-build-a-terminal`
**Created**: 2025-09-28
**Status**: Draft
**Input**: User description: "Build a Terminal User Interface (TUI) application called distui that manages Go application releases and distributions. The app detects Go projects by reading git remotes and go.mod files, stores all configuration globally in ~/.distui (never polluting user repos), and provides a seamless interface for managing releases across multiple projects. Users can navigate between Project view (current project actions), Global view (all projects list), and Settings view using TAB or keyboard shortcuts. The TUI executes release processes directly including running tests, creating git tags, running GoReleaser, updating Homebrew taps, and optionally publishing to NPM. Each project's configuration is remembered globally, including distribution channels (GitHub releases, Homebrew tap location and formula names, NPM package settings), build preferences, and release history. The app supports both rapid local releases (30 seconds) and optional GitHub Actions workflow generation for CI/CD. Smart detection uses gh CLI to find repositories, homebrew taps, and authentication status, but always allows user overrides. The interface shows real-time progress during releases with live command output and interactive error handling."

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
1. **Given** I'm in a Go project directory, **When** I launch distui, **Then** the app detects the project from git remotes and go.mod and loads any existing configuration from ~/.distui
2. **Given** I have multiple Go projects configured, **When** I press TAB in the TUI, **Then** I can cycle between Project view, Global view (all projects), and Settings view
3. **Given** I'm in Project view with a configured project, **When** I initiate a release, **Then** the app executes tests, creates tags, runs GoReleaser, updates Homebrew tap, and optionally publishes to NPM within 30 seconds
4. **Given** I'm setting up a new project, **When** the app detects repository info, **Then** I can override any detected values before saving the configuration
5. **Given** a release is in progress, **When** a command fails, **Then** I see the error output in real-time and can choose how to proceed interactively
6. **Given** I have projects with different distribution needs, **When** I configure each project, **Then** each maintains its own distribution channel settings (GitHub only, Homebrew+GitHub, NPM+GitHub, etc.)

### Edge Cases
- What happens when gh CLI is not installed or not authenticated?
- How does system handle when homebrew tap location doesn't exist?
- What occurs if project has no go.mod file?
- How does app respond when ~/.distui directory is not writable?
- What happens during concurrent releases of different projects?
- How does system handle when detected values conflict with saved configuration?

## Requirements

### Functional Requirements
- **FR-001**: System MUST detect Go projects by reading git remote origin and go.mod files in the current directory
- **FR-002**: System MUST store all configuration data in ~/.distui directory, never adding files to user repositories
- **FR-003**: System MUST provide three navigable views: Project (current project), Global (all projects list), and Settings
- **FR-004**: Users MUST be able to switch between views using TAB key or dedicated keyboard shortcuts
- **FR-005**: System MUST execute release processes directly from the TUI including: running tests, creating git tags, executing GoReleaser, updating Homebrew formulas, and publishing to NPM
- **FR-006**: System MUST complete typical release execution within 30 seconds
- **FR-007**: System MUST remember project configurations globally, allowing instant access when revisiting projects
- **FR-008**: System MUST support configuration of distribution channels per project (GitHub releases, Homebrew tap settings, NPM package settings)
- **FR-009**: System MUST use gh CLI for smart detection of repositories, taps, and authentication status
- **FR-010**: Users MUST be able to override any automatically detected values
- **FR-011**: System MUST display real-time command output during release execution
- **FR-012**: System MUST provide interactive error handling when commands fail
- **FR-013**: System MUST maintain release history for each project
- **FR-014**: System MUST support generating GitHub Actions workflows for CI/CD as an alternative to local execution
- **FR-015**: System MUST allow configuration of Homebrew tap location and formula names per project
- **FR-016**: System MUST allow configuration of NPM package names and settings per project
- **FR-017**: System MUST work with projects that have some but not all distribution channels enabled
- **FR-018**: System MUST provide visual progress indicators during long-running operations

### Key Entities
- **Project**: Represents a Go application with its repository information, module path, and current version
- **Configuration**: Project-specific settings including distribution channels, build preferences, and release options
- **Distribution Channel**: A method of distributing releases (GitHub Releases, Homebrew, NPM) with its specific settings
- **Release History**: Record of past releases including version, date, method (local/CI), and status
- **Global Settings**: User-wide preferences including default Homebrew tap location, NPM scope, and UI preferences

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

- [x] User description parsed
- [x] Key concepts extracted
- [x] Ambiguities marked (none found)
- [x] User scenarios defined
- [x] Requirements generated
- [x] Entities identified
- [x] Review checklist passed

---