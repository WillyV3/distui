# Feature Specification: distui - Go Release Distribution Manager

**Feature Branch**: `001-build-a-terminal`
**Created**: 2025-09-28
**Updated**: 2025-09-30
**Status**: ✅ PRODUCTION READY - v0.0.31 - Core features complete, enhancements planned
**Input**: User description update: "Complete the release workflow implementation including GoReleaser integration, Homebrew tap updates, NPM publishing, and animated progress display. Build release execution flow using package-manager example from charm-examples-inventory for animated progress. Support both interactive TUI releases and GitHub Actions workflow generation for CI/CD. Preserve all completed git management features (smart commit, file categorization, GitHub repo creation, cleanup view) while implementing release execution."

**Planned Enhancements**: "Project-level smart commit preferences customization - users want to customize file categorization rules per project with full freedom over all file types. Settings should be stored in project YAML, keeping code out of configure_handler as much as possible. GitHub Actions workflow generation for CI/CD - optional/opt-in to respect developer autonomy, helping beginners without forcing opinions on experienced developers. All features wired following code conventions and constitution."

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

#### Release Workflow ✅
9. **Given** I'm in Project view with a configured project, **When** I initiate a release, **Then** the app executes tests, creates tags, runs GoReleaser, updates Homebrew tap, and optionally publishes to NPM within 30 seconds ✅
10. **Given** a release is in progress, **When** commands execute, **Then** I see animated progress with package-manager style UI showing each step ✅
11. **Given** I'm viewing NPM configuration, **When** I check a package name, **Then** the system detects similar packages (not just exact matches) and shows availability status with suggestions if conflicts exist ✅
12. **Given** NPM detects a package conflict, **When** I press 'e' on the NPM item, **Then** I can edit the package name inline without leaving the distributions tab ✅
13. **Given** I have projects with different distribution needs, **When** I configure each project, **Then** each maintains its own distribution channel settings (GitHub only, Homebrew+GitHub, NPM+GitHub, etc.) ✅
14. **Given** I've made configuration changes, **When** I attempt to release without regenerating files, **Then** the system blocks the release and shows a warning ✅
15. **Given** I publish to NPM, **When** the release completes, **Then** package.json version bump is automatically committed and pushed ✅
16. **Given** I own an NPM package, **When** the system checks availability, **Then** it recognizes my ownership and marks the package as available ✅

#### Planned Enhancements 🔄
17. **Given** I have specific file categorization preferences, **When** I configure smart commit settings for my project, **Then** I can customize which file extensions map to which categories (config, code, docs, etc.) and these preferences are stored in my project configuration
18. **Given** I want to customize smart commit behavior, **When** I edit smart commit preferences, **Then** I can define custom patterns for all file types without modifying global settings
19. **Given** I'm a beginner setting up CI/CD, **When** I enable GitHub Actions workflow generation, **Then** the system generates appropriate workflow files for releases and distributions based on my project configuration
20. **Given** I'm an experienced developer, **When** I view workflow generation options, **Then** I can opt-out of automatic workflow generation and maintain full control over my CI/CD setup

### Edge Cases (Handled ✅)
- NPM package names with similar variants (hyphens/underscores) detected and flagged ✅
- User owns existing NPM package - recognized and marked as available ✅
- Configuration changes require file regeneration - releases blocked until regenerated ✅
- Package.json modified during NPM publish - auto-committed and pushed ✅
- Working tree checks don't flash during NPM publish (check order optimized) ✅
- Cleanup tab auto-refreshes when switching tabs after config changes ✅
- ESC during NPM name editing cancels edit mode, not entire view ✅

### Known Issues 🐛
- Directories beginning with "." (dot files/folders like .github, .goreleaser) cannot be modified in commit settings interface

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

#### Release Workflow (Completed ✅)
- **FR-005**: System MUST execute release processes directly from the TUI including: running tests, creating git tags, executing GoReleaser, updating Homebrew formulas, and publishing to NPM ✅
- **FR-006**: System MUST complete typical release execution within 30 seconds ✅
- **FR-008**: System MUST support configuration of distribution channels per project (GitHub releases, Homebrew tap settings, NPM package settings) ✅
- **FR-011**: System MUST display real-time command output during release execution with package-manager style animated progress ✅
- **FR-013**: System MUST maintain release history for each project ✅
- **FR-015**: System MUST allow configuration of Homebrew tap location and formula names per project ✅
- **FR-016**: System MUST allow configuration of NPM package names and settings per project ✅
- **FR-017**: System MUST work with projects that have some but not all distribution channels enabled ✅
- **FR-018**: System MUST provide visual progress indicators during long-running operations using package-manager example pattern ✅
- **FR-027**: System MUST bump version numbers automatically or allow user to specify version ✅
- **FR-028**: System MUST validate GoReleaser configuration before executing release ✅
- **FR-030**: System MUST show live streaming output for each release step (tests, build, publish) ✅
- **FR-031**: System MUST detect NPM package name similarity conflicts (not just exact matches) and provide scoped/suffixed suggestions ✅
- **FR-032**: System MUST allow inline editing of NPM package names without leaving distributions tab ✅
- **FR-033**: System MUST auto-trigger NPM name validation when entering distributions tab with NPM enabled ✅
- **FR-034**: System MUST block releases when configuration changes require file regeneration ✅
- **FR-035**: System MUST auto-commit and push package.json version bumps after NPM publish ✅
- **FR-036**: System MUST recognize when user owns an NPM package and mark it as available ✅
- **FR-037**: System MUST auto-refresh cleanup tab when switching tabs after configuration changes ✅
- **FR-038**: System MUST display current version and distribution info in project view ✅
- **FR-039**: System MUST allow dismissing release success screen with ESC/Enter/Space ✅

#### Smart Commit Customization (Planned 🔄)
- **FR-040**: System MUST allow users to configure custom file categorization rules per project
- **FR-041**: System MUST store smart commit preferences in project configuration file (~/.distui/projects/<id>/config.yaml)
- **FR-042**: Users MUST be able to define custom file extension to category mappings (e.g., .proto → code, .sql → data)
- **FR-043**: System MUST support custom glob patterns for file categorization (e.g., **/test/** → test files)
- **FR-044**: Users MUST be able to override default categorization rules without affecting other projects
- **FR-045**: System MUST provide a UI for editing smart commit preferences separate from main configure handler
- **FR-046**: Smart commit preferences MUST be optional - defaults apply if not customized
- **FR-047**: System MUST validate custom categorization rules before saving to prevent conflicts

#### GitHub Actions Workflow Generation (Planned 🔄)
- **FR-048**: System MUST provide optional GitHub Actions workflow generation for release automation
- **FR-049**: Workflow generation MUST be opt-in to respect developer preferences and avoid forcing opinions
- **FR-050**: System MUST generate workflows that match the project's enabled distribution channels (Homebrew, NPM, GitHub Releases)
- **FR-051**: Generated workflows MUST follow GitHub Actions best practices and community standards
- **FR-052**: System MUST allow users to preview generated workflow files before creation
- **FR-053**: Users MUST be able to disable workflow generation entirely in project settings
- **FR-054**: Generated workflows MUST be added to .github/workflows/ in user's repository (with explicit consent)
- **FR-055**: System MUST support regenerating workflows when distribution configuration changes
- **FR-056**: Workflow generation MUST include proper error handling and notification steps

### Key Entities
- **Project**: Represents a Go application with its repository information, module path, and current version
- **Configuration**: Project-specific settings including distribution channels, build preferences, and release options
- **Distribution Channel**: A method of distributing releases (GitHub Releases, Homebrew, NPM) with its specific settings
- **Release History**: Record of past releases including version, date, method (local/CI), and status
- **Global Settings**: User-wide preferences including default Homebrew tap location, NPM scope, and UI preferences
- **File Change**: Represents uncommitted file with path, status (modified/added/deleted/untracked), and category
- **Cleanup Item**: UI representation of file with action (commit/skip/ignore)
- **Commit Model**: State for commit view including selected files and commit message
- **NPM Name Check**: Result of NPM package name validation including status (available/unavailable/checking/error), ownership info, and suggestions
- **Release Phase**: State machine for release workflow (version select, validation, building, publishing, complete, failed)
- **Version Bump**: Type of version increment (patch, minor, major, custom) with validation logic
- **Smart Commit Preferences**: Project-level customization of file categorization rules including custom extensions and glob patterns per category
- **File Category Rule**: Extension or pattern mapping that determines how files are categorized (config, code, docs, build, test, assets, data)
- **GitHub Workflow Config**: Settings for GitHub Actions workflow generation including enabled channels, secrets, and environments

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

### Completed (2025-09-28 to 2025-09-30)

#### Core Infrastructure ✅
- Full TUI infrastructure with 6 views
- Configuration management with ~/.distui storage
- Project detection from git/go.mod
- Async loading with spinner for configure view

#### Git Management ✅
- Git cleanup view with file categorization
- Smart commit with AI-generated messages
- GitHub repository creation from TUI
- File-by-file commit workflow
- Repository browser for file navigation
- Push detection and execution
- Repository status display with sync indicator

#### Release Workflow ✅
- Version bumping logic (patch/minor/major/custom)
- GoReleaser integration and execution
- Homebrew tap updates via GoReleaser
- NPM publishing with pre-flight checks
- Animated progress UI (package-manager pattern)
- Release history tracking and display
- Project view redesign with git status and distribution info

#### NPM Publishing Features ✅ (v0.0.28-0.0.31)
- **Similarity Detection**: Detects conflicts like "distui" vs "dist-ui" using variation generation
- **Ownership Detection**: Recognizes user's own packages via `npm whoami`
- **Inline Package Name Editing**: Edit NPM package name directly in distributions tab with 'e' key
- **Auto-trigger Validation**: Automatically checks NPM name when entering distributions tab
- **Scoped Package Suggestions**: Provides @username/package and suffixed alternatives
- **Auto-commit Package.json**: Commits and pushes version bump after successful NPM publish
- **Release Blocking**: Prevents releases when configuration changes require file regeneration
- **Status Display**: Shows availability with ownership info below content box

#### UI/UX Improvements ✅
- ESC handler to dismiss release success screen
- Working tree check order optimized (no flash during NPM publish)
- Tab auto-refresh with loading spinner (cleanup tab)
- ESC cancels NPM edit mode without exiting entire view
- Distribution info display in project view (NPM package name, Homebrew tap)
- Regeneration warning in project view when config changes require file updates
- NeedsRegeneration flag only set on actual config changes, not initial load

### Outstanding Features 🔄

#### Smart Commit Customization (Planned)
- Project-level file categorization preferences editor
- Custom extension to category mappings
- Custom glob pattern support for file categorization
- Live preview of how files would be categorized
- Validation of custom patterns before saving
- Reset to defaults functionality
- Separate handler to keep logic out of configure_handler.go

#### GitHub Actions Workflow Generation (Planned)
- Optional workflow generation for automated releases (opt-in)
- Preview generated workflow before creating files
- Support for multiple distribution channels in workflow
- Configuration of workflow triggers and secrets
- Auto-regeneration when distribution config changes
- Respect for developer autonomy - easy opt-out, no forced opinions

#### Bug Fixes Needed
- Fix handling of dot files/directories (e.g., .github, .goreleaser) in commit settings interface

---