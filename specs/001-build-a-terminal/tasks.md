# Implementation Tasks: distui - Go Release Distribution Manager

**Feature**: Build a Terminal UI Application for Go Release Management
**Branch**: 001-build-a-terminal
**Created**: 2025-09-28

## Overview

This document contains all implementation tasks for the distui feature. Tasks are ordered by dependencies and marked with [P] when they can be executed in parallel.

### Quick Stats
- Total Tasks: ~35
- Completed: 100% FEATURE COMPLETE
- Status: **PRODUCTION READY** (v0.0.31)
- Latest Release: v0.0.31 (NPM published and working)

### ✅ FULLY WORKING FEATURES (2025-09-30 Update 3)

**Core Functionality:**
- ✅ Full TUI with 4 views (Project, Global, Settings, Configure)
- ✅ Project detection from go.mod and git
- ✅ Configuration persistence to ~/.distui/projects/{identifier}.yaml
- ✅ Terminal layout integrity (no overflow, dynamic height management)

**Release Workflow:**
- ✅ Version bumping (patch/minor/major/custom)
- ✅ Pre-release tests (go test ./...)
- ✅ Git tag creation and push
- ✅ GoReleaser integration with streaming output
- ✅ GitHub Releases (binary uploads, release notes)
- ✅ Homebrew formula updates (via GoReleaser brews config)
- ✅ NPM publishing (with golang-npm, post-GoReleaser)
- ✅ Multi-channel releases (GitHub + Homebrew + NPM simultaneously)

**Configuration Management:**
- ✅ All 4 configure tabs (Cleanup, Distributions, Build, Advanced)
- ✅ Smart file generation (.goreleaser.yaml, package.json)
- ✅ File deletion when distributions disabled
- ✅ Consent screen showing generate/delete changes
- ✅ Auto-regeneration indicator when config changes
- ✅ Stable JSON field order (no git diffs on regeneration)
- ✅ Regex-based version updates (preserves formatting)
- ✅ NPM package name validation with availability checking
  - ✅ Ownership detection (distinguishes user's packages from others)
  - ✅ Similarity detection (checks variations with hyphens/underscores)
  - ✅ Alternative name suggestions (scoped packages, suffixes)
  - ✅ Inline package name editing in Distributions tab
  - ✅ Auto-trigger checking when tab opens or NPM enabled
- ✅ Release blocking when regeneration needed
- ✅ Tab refresh with loading spinner (Cleanup tab auto-refreshes)

**Git Management:**
- ✅ Git cleanup UI with intelligent categorization
- ✅ GitHub repository creation/connection
- ✅ Smart commit with auto-categorized files
- ✅ Push detection and remote sync
- ✅ Binary and build artifact exclusion

**Distribution Channels:**
- ✅ GitHub Releases - GoReleaser handles binary builds and uploads
- ✅ Homebrew - GoReleaser pushes to tap with correct formula
- ✅ NPM - Separate publish using golang-npm for binary distribution
  - ✅ Real-time package name availability checking
  - ✅ Ownership detection (recognizes user's existing packages)
  - ✅ Conflict detection (e.g., "distui" vs "dist-ui", "distui-cli" vs "distui_cli")
  - ✅ Scoped package suggestions (@username/package)
  - ✅ Alternative name generation (package-cli, package-tool, etc.)
  - ✅ Automatic package.json version bump on publish
  - ✅ Auto-commit and push package.json changes post-publish
  - ✅ Verified working: `npm install -g distui-cli-go` installs and runs successfully
- ✅ Go Module - Via git tags (no special handling needed)

**UI/UX Improvements:**
- ✅ Release success screen with ESC to dismiss
- ✅ Project view shows NPM and Homebrew distribution info
- ✅ Distribution info hidden during active release
- ✅ Working tree check moved after release check (prevents flash during NPM publish)
- ✅ Clean project view after successful release (no dirty tree warnings)
- ✅ All warnings preserved (regeneration, working tree, GitHub auth, config missing)

**Recent Bug Fixes (v0.0.28-0.0.31):**
- ✅ Fixed NPM checker incorrectly flagging user's own packages as unavailable
- ✅ Fixed ESC not canceling NPM package name edit mode
- ✅ Fixed cleanup tab not refreshing after config changes in other tabs
- ✅ Fixed "WORKING TREE NOT CLEAN" flashing during NPM publish
- ✅ Fixed release blocking not working when regeneration needed
- ✅ Fixed NPM variation checker to detect underscore/hyphen swaps
- ✅ Added loading spinner when switching to cleanup tab
- ✅ Removed all debug statements from NPM publisher

### What's Left for MVP
1. **Testing** (T032-T039) - Optional, can ship without
2. **Polish** (T044-T047) - Nice to have (spinners already work, help screen optional)
3. **Integration cleanup** (T040-T043) - Mostly done, just cleanup

### REMOVED Tasks (Not Needed)
- ~~T028-T029: New Project Wizard~~ - Configure view IS the project setup
- ~~T-CFG-4,5: Homebrew/NPM detection~~ - Not needed for MVP, users can toggle manually
- ~~T-CFG-8,9: Status display, validation~~ - Current UI is sufficient

### Execution Guide
Tasks marked with [P] can be run in parallel. For example:
```bash
# Run parallel tasks together
Task general-purpose "Complete T003, T004, and T005 in parallel"
```

## Setup Tasks (Dependencies First)

### ✅ T001: Adapt Template Structure for Distui [COMPLETED]
**File**: app.go
- Replace pageState enum with distui views (projectView, globalView, settingsView, releaseView, configureView, newProjectView)
- Update main menu items to show Project/Global/Settings options
- Modify router in Update() to handle new page states
- Update View() switch to render distui views
**Status**: Complete - app.go updated with all distui views and navigation
**Estimate**: 2 points

### ✅ T002: Initialize Go Module and Dependencies [COMPLETED]
**File**: go.mod
- Create go.mod with module name "distui"
- Add Bubble Tea v0.27.0 dependency
- Add Lipgloss v0.13.0 dependency
- Add yaml.v3 dependency
- Run go mod tidy
**Status**: Complete - go.mod created with all dependencies
**Estimate**: 1 point

### ✅ T003: [P] Create Internal Package Structure [COMPLETED]
**Files**: internal/*/
- Create internal/config directory with loader.go (99 lines)
- Create internal/detection directory with project.go (97 lines)
- Create internal/executor directory with release.go (87 lines)
- Create internal/models directory with types.go (81 lines)
- Add package declarations to each file
**Status**: Complete - all packages created with proper structure
**Estimate**: 1 point

### ✅ T004: [P] Create Handler Stubs [COMPLETED]
**Files**: handlers/*.go
- Created handlers/project_handler.go with UpdateProjectView function
- Created handlers/global_handler.go with UpdateGlobalView function
- Created handlers/settings_handler.go with UpdateSettingsView function
- Created handlers/release_handler.go with UpdateReleaseView function
- Created handlers/configure_handler.go with UpdateConfigureView function
- Created handlers/newproject_handler.go with UpdateNewProjectView function
- All handlers follow pattern: func UpdateXxx(currentPage, previousPage int, msg tea.Msg) (int, bool, tea.Cmd)
**Status**: Complete - all handlers created (replaced page*handler.go files)
**Estimate**: 2 points

### ✅ T005: [P] Create View Stubs [COMPLETED]
**Files**: views/*.go
- Created views/project_view.go with RenderProjectContent function (61 lines)
- Created views/global_view.go with RenderGlobalContent function (80 lines)
- Created views/settings_view.go with RenderSettingsContent function (66 lines)
- Created views/release_view.go with RenderReleaseContent function (70 lines)
- Created views/configure_view.go with RenderConfigureContent function (73 lines)
- All views return string content with proper formatting
**Status**: Complete - all views created with placeholder content
**Estimate**: 2 points

## Configuration Management Tasks

### ✅ T006: Implement Config Loader [COMPLETED]
**File**: internal/config/loader.go
- Implemented LoadGlobalConfig() function to read ~/.distui/config.yaml
- Implemented LoadProject(identifier string) to read project YAML
- NO fallbacks - errors bubble up as per constitution
- Parse YAML into GlobalConfig struct using proper types
- 119 lines (slightly over due to essential code)
**Status**: Complete - tested with real config files
**Estimate**: 3 points

### ✅ T007: Implement Config Writer [COMPLETED]
**File**: internal/config/loader.go (combined with loader)
- Implemented SaveGlobalConfig(config *GlobalConfig) with atomic writes
- Implemented SaveProject(project *Project) with atomic writes
- Use temp file + rename pattern for safety
- Create directories if needed
- Handle write permissions properly
**Status**: Complete - save functions in loader.go (constitution: avoid unnecessary abstraction)
**Estimate**: 3 points

### ✅ T008: [P] Define Configuration Types [COMPLETED]
**File**: internal/models/types.go
- Defined GlobalConfig struct matching contracts/config.yaml
- Defined ProjectConfig struct matching contracts/project.yaml
- Defined ProjectSettings struct with distributions
- Defined ReleaseHistory struct
- Added YAML tags to all fields
**Status**: Complete - all types defined (134 lines, acceptable for completeness)
**Estimate**: 2 points

### ✅ T009: [P] Implement Path Management [COMPLETED]
**File**: internal/config/loader.go (integrated)
- expandHome() function handles ~ expansion
- Path management integrated into load/save functions
- Directories created as needed in Save functions
- Home directory expansion working
**Status**: Complete - path management in loader.go (constitution: avoid unnecessary abstraction)
**Estimate**: 1 point

## Detection Tasks

### ✅ T010: Implement Project Detection [COMPLETED]
**File**: internal/detection/project.go
- Implemented DetectProject(path string) function
- Parses go.mod with fallback for new Go versions (1.24.0+)
- Optional Git/GitHub detection (won't fail without .git)
- Extracts binary name from module path
- Added sanitizeIdentifier() for safe file names
- Returns ProjectInfo struct with all detected values
**Status**: Complete - works with Go 1.25.1 and projects without Git
**Estimate**: 3 points

### ✅ T011: Implement GitHub Detection [COMPLETED]
**File**: internal/detection/project.go (integrated)
- Implemented DetectGitHubUsingGH() function
- Uses gh CLI to get repo info
- Gracefully handles gh CLI not installed
- Returns RepositoryInfo with owner/name/branch
**Status**: Complete - integrated into project.go
**Estimate**: 2 points

### T012: [P] Implement Homebrew Detection
**File**: internal/detection/homebrew.go (TO CREATE)
- Implement DetectHomebrewTap(username string)
- Check common tap locations
- Look for existing formula files
- Return tap repository path
- Handle taps not found gracefully
**Estimate**: 2 points

## Core View Implementation

### ✅ T013: Implement Project View [COMPLETED]
**File**: views/project_view.go
- ✅ RenderProjectContent accepts real ProjectInfo and ProjectConfig
- ✅ Displays actual module name, version, path from detection
- ✅ Shows repository info if available
- ✅ Displays release history from config
- ✅ Has keyboard hints and action buttons
- ✅ Shows GitHub user status indicator (green when configured, warning when not)
**Status**: Complete with GitHub status indicator
**Estimate**: 3 points

### ✅ T014: Implement Project Handler [COMPLETED]
**File**: handlers/project_handler.go
- ✅ Handle 'r' key to start release (return releaseView)
- ✅ Handle 'c' key for configuration (return configureView)
- ✅ Handle 'tab' to cycle to global view
- ✅ Handle 'g' shortcut to global view
- ✅ Handle 's' shortcut to settings view (fixed page index mapping)
**Integration**: app.go calls UpdateProjectView() correctly
**Status**: Complete - all navigation working
**Estimate**: 2 points

### ✅ T015: Implement Global View [COMPLETE]
**File**: views/global_view.go
- ✅ RenderGlobalContent accepts projects list and selectedIndex
- ✅ Displays project table with name/version/status
- ✅ Shows selected project with arrow indicator
- ✅ Has action buttons for add/scan/delete
- ✅ Shows navigation hints
- ✅ Supports delete mode and scan mode indicators
**Handler**: handlers/global_handler.go
- ✅ GlobalModel manages project list and selection
- ✅ Arrow key navigation through projects
- ✅ Delete mode with confirmation
- ✅ Scan mode placeholder
- ✅ Action handling (add, delete, select)
**Status**: COMPLETE
**Estimate**: 3 points

### ✅ T016: Implement Global Handler [COMPLETE]
**File**: handlers/global_handler.go
- ✅ Handle up/down arrow navigation
- ✅ Handle Enter to switch to selected project
- ✅ Handle 'n' to add new project
- ✅ Handle 'd' to delete project
- ✅ Handle tab cycling and shortcuts
**Status**: Fully implemented with GlobalModel
**Estimate**: 2 points

### ✅ T017: Implement Settings View [COMPLETED]
**File**: views/settings_view.go
- ✅ RenderSettingsContent with interactive SettingsModel
- ✅ Shows current configuration values
- ✅ Interactive edit mode with textinput components
- ✅ Visual feedback for focused fields
- ✅ Save confirmation message
**Status**: Complete with Bubble Tea interactive components
**Estimate**: 3 points

### ✅ T018: Implement Settings Handler [COMPLETED]
**File**: handlers/settings_handler.go
- ✅ Handle 'e' key to enter edit mode
- ✅ Interactive textinput fields for all settings
- ✅ Auto-detection of GitHub username from gh CLI
- ✅ Tab/Shift+Tab navigation between fields
- ✅ Enter to save, Escape to cancel
- ✅ Persists to ~/.distui/config.yaml
- ✅ Smart pre-population with detected values
**Status**: Complete with auto-detection
**Estimate**: 3 points

## Release Execution Tasks

### ✅ T012: Implement Homebrew Detection [COMPLETED]
**File**: internal/detection/homebrew.go (CREATED)
**Status**: COMPLETE
- Implement DetectHomebrewTap(username string)
- Check common tap locations (~/homebrew-tap, ~/repos/homebrew-tap)
- Use gh CLI to find repos matching "homebrew-*" pattern
- Return tap repository path and existing formulas
- Handle taps not found gracefully
**Estimate**: 2 points
**Architecture**: Business logic in internal/detection (NOT in handlers)

### ✅ T022: Implement Command Runner [COMPLETED]
**File**: internal/executor/command.go (CREATED)
**Status**: COMPLETE
- Implement RunCommandStreaming(name, args, dir) function
- Create CommandOutput message type for line-by-line streaming
- Create CommandComplete message type for completion status
- Handle stdout and stderr separately with goroutines
- Send tea.Msg for each output line
- Return exit code and error
**Estimate**: 3 points
**Pattern**: Use goroutines + channels to stream, send tea.Msg to TUI
**Architecture**: Business logic in internal/executor, messages to handlers

### ✅ T021: Expand Release Executor [COMPLETED]
**File**: internal/executor/release.go (EXPANDED)
**Status**: COMPLETE
**Current State**: Has basic runTests/buildRelease/createTag/pushTag stubs
**Needs**:
- Keep existing ReleaseExecutor/ReleaseConfig structure
- Add ExecuteReleasePhases(ctx, phases) function
- Add RunGoReleaser(ctx, version) function
- Add UpdateHomebrewTap(ctx, tapPath, version) function
- Add PublishNPM(ctx, packageName) function
- Integrate with RunCommandStreaming for output
- Send phase completion messages to TUI
- Handle rollback on failure
**Estimate**: 5 points
**Architecture**: ALL execution logic stays here, handlers only manage state

### ✅ T023: [P] Implement Test Executor [COMPLETED]
**File**: internal/executor/test.go (CREATED)
**Status**: COMPLETE
- Implement RunTests(project) function
- Execute "go test ./..." command
- Use RunCommandStreaming for output
- Return success/failure status
- Keep under 100 lines
**Estimate**: 2 points

### ✅ T024: [P] Implement GoReleaser Executor [COMPLETED]
**File**: internal/executor/goreleaser.go (CREATED)
**Status**: COMPLETE
- Implement RunGoReleaser(project, version) function
- Get GITHUB_TOKEN from gh auth token
- Execute "goreleaser release --clean" command
- Use RunCommandStreaming for output
- Check if goreleaser is installed first
- Handle goreleaser not installed gracefully
**Estimate**: 3 points

### ✅ T025: [P] Implement Homebrew Updater [COMPLETED]
**File**: internal/executor/homebrew.go (CREATED)
**Status**: COMPLETE
- Implement UpdateHomebrewTap(project, version, tapPath) function
- Download release tarball from GitHub
- Calculate SHA256 checksum
- Update formula file with new version + SHA256
- Commit changes to tap repository
- Push to remote
- Use RunCommandStreaming for git commands
**Estimate**: 5 points

### ✅ T030: [P] Define Message Types [COMPLETED]
**File**: internal/models/messages.go (CREATED)
**Status**: COMPLETE
- Define releasePhaseMsg (phase started)
- Define releasePhaseCompleteMsg (phase done)
- Define commandOutputMsg (streaming output line)
- Define commandCompleteMsg (command finished)
- Define releaseCompleteMsg (all phases done)
- Define releaseErrorMsg (failure with recovery options)
**Estimate**: 1 point

### ✅ T020: Implement Release Handler [COMPLETED]
**File**: handlers/release_handler.go (EXPANDED)
**Status**: COMPLETE
**Current State**: Just navigation, no state management
**Needs**:
- Create ReleaseModel struct with:
  - Phase (ReleasePhase enum)
  - Packages ([]Package for package-manager pattern)
  - Installing (int, current step index)
  - Installed ([]int, completed steps)
  - Progress (progress.Model from bubbles)
  - Spinner (spinner.Model from bubbles)
  - Output ([]string buffer)
  - Version (string)
  - StartTime (time.Time)
  - Error (error)
- Implement version selection state (patch/minor/major/custom)
- Handle Enter to launch executeReleaseCmd
- Handle phase completion messages
- Update progress and spinner
- Handle command output messages (append to buffer)
- Return to project view on completion
**Pattern**: Use package-manager example (each phase = "package")
**Architecture**: State management ONLY, no execution logic
**Estimate**: 5 points

### ✅ T019: Implement Release View [COMPLETED]
**File**: views/release_view.go (EXPANDED)
**Status**: COMPLETE
**Current State**: Hardcoded mock data, no dynamic rendering
**Needs**:
- Accept ReleaseModel as parameter (not empty function)
- Render version selection UI when Phase == PhaseVersionSelect
  - Show current version
  - Show patch/minor/major options
  - Show custom input option
  - Highlight selected option
- Render progress display when Phase >= PhaseTests
  - Spinner for current phase
  - Progress bar (bubbles/progress)
  - List of packages (phases) with status
  - Checkmarks for completed phases
  - Elapsed time counter
- Render output streaming area (scrollable)
  - Last 20 lines of command output
  - Auto-scroll to bottom
  - Dim color for older lines
- Render success/failure summary when Phase == PhaseComplete
  - Duration, channels published, next steps
- Render error display with recovery options on failure
**Pattern**: Use package-manager example for progress rendering
**Architecture**: Pure rendering, no business logic
**Estimate**: 5 points

### ✅ T031: Wire Release to App [COMPLETED]
**File**: app.go (EXISTS - needs update)
**Status**: COMPLETE
- Add releaseModel *handlers.ReleaseModel to model struct
- Initialize releaseModel when navigating to releaseView (like configureModel)
- Pass project info and version to NewReleaseModel
- Route releaseView case to handlers.UpdateReleaseView
- Handle window sizing for releaseModel
- Pass releaseModel to views.RenderReleaseContent
**Estimate**: 2 points

## View Architecture Refactor [CRITICAL]

### ✅ T-CFG-10: Merge Release View into Project View [COMPLETED] [3 points]
**Files**:
- views/project_view.go (modify) ✅
- views/release_view.go (refactor) ✅
- handlers/project_handler.go (modify) ✅
- app.go (modify) ✅

**Current State**:
- Project view shows static project info + "Press [r] to release"
- Separate releaseView page (pageState = 3) for version selection
- Extra navigation step: project → [r] → release → select version → enter
- release_view.go renders 4 phases: PhaseVersionSelect, PhaseComplete, PhaseFailed, progress

**Target State**:
- Project view includes inline release version selector
- No separate release page needed
- Direct interaction: arrow keys select version → enter starts release
- Release progress shown in same view (replaces project info during execution)

**Implementation**:

1. **views/project_view.go** - Add release section:
```go
func RenderProjectContent(p *ProjectState, releaseModel *handlers.ReleaseModel) string {
    sections := []string{
        renderProjectHeader(p),
        renderInlineReleaseSection(releaseModel), // NEW
        renderQuickActions(p),
    }
    return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func renderInlineReleaseSection(m *handlers.ReleaseModel) string {
    if m == nil {
        return ""
    }

    switch m.Phase {
    case models.PhaseVersionSelect:
        return renderCompactVersionSelect(m) // Inline version, not full screen
    case models.PhaseComplete:
        return views.RenderSuccess(m)
    case models.PhaseFailed:
        return views.RenderFailure(m)
    default:
        return views.RenderProgress(m)
    }
}

func renderCompactVersionSelect(m *handlers.ReleaseModel) string {
    // Simplified version of renderVersionSelection from release_view.go
    // Shows "SELECT RELEASE VERSION" + 4 options + keyboard hints
    // Fits in ~10 lines instead of full screen
}
```

2. **views/release_view.go** - Refactor to export helpers:
```go
// Keep these as exported functions for reuse:
// - RenderProgress()
// - RenderSuccess()
// - RenderFailure()

// Remove or make internal:
// - RenderReleaseContent() (replaced by inline in project view)
// - renderVersionSelection() (replaced by compact version)
```

3. **handlers/project_handler.go** - Handle version selection:
```go
func UpdateProjectView(currentPage, previousPage int, msg tea.Msg,
                       projectState *ProjectState,
                       releaseModel *handlers.ReleaseModel) (int, bool, tea.Cmd, *handlers.ReleaseModel) {

    switch msg := msg.(type) {
    case tea.KeyMsg:
        // If release phase is version select, handle arrow keys + enter
        if releaseModel != nil && releaseModel.Phase == models.PhaseVersionSelect {
            switch msg.String() {
            case "up", "k":
                releaseModel.SelectedVersion--
                // clamp to bounds
            case "down", "j":
                releaseModel.SelectedVersion++
                // clamp to bounds
            case "enter":
                return currentPage, false, releaseModel.StartRelease(), releaseModel
            case "esc":
                releaseModel.Phase = models.PhaseVersionSelect // reset
                releaseModel.SelectedVersion = 0
            }
        }

        // Normal project view navigation
        switch msg.String() {
        case "g":
            return int(globalView), false, nil, releaseModel
        case "s":
            return int(settingsView), false, nil, releaseModel
        case "c":
            return int(configureView), false, nil, releaseModel
        case "r":
            // Don't navigate away - just activate version selector
            releaseModel.Phase = models.PhaseVersionSelect
            return currentPage, false, nil, releaseModel
        }

    case models.ReleasePhaseMsg, models.ReleaseCompleteMsg:
        // Forward to releaseModel.Update()
        updatedModel, cmd := releaseModel.Update(msg)
        return currentPage, false, cmd, updatedModel
    }

    return currentPage, false, nil, releaseModel
}
```

4. **app.go** - Remove releaseView page state:
```go
type pageState int

const (
    projectView pageState = iota
    globalView
    settingsView
    configureView
    newProjectView
    // REMOVE: releaseView
)

// Initialize releaseModel when project is detected, not on navigation
if m.detectedProject != nil && m.releaseModel == nil && m.width > 0 && m.height > 0 {
    m.releaseModel = handlers.NewReleaseModel(...)
}

// In Update(), remove releaseView case entirely
// Project view now handles everything

// In View(), update project rendering:
case projectView:
    return m.renderProjectView(m.projectState, m.releaseModel)
```

**Architecture Notes**:
- Eliminates redundant page state (5 total instead of 6)
- Follows "30-second release" goal (fewer keypresses)
- Maintains handler/view separation
- ReleaseModel still manages state, just rendered inline
- Progressive disclosure: version selector only shows when [r] pressed
- During release execution, progress takes over the whole view

**Dependencies**: None (standalone refactor)

**Validation**:
- User opens app → sees project info with "Press [r]" hint
- Press [r] → version selector appears inline (4 options)
- Arrow keys move selection
- Enter starts release → progress view takes over screen
- On completion → shows success, ESC returns to project view with version selector hidden
- No [r] navigation needed, no separate page

**Estimate**: 3 points (moderate complexity, multiple file changes, careful state management)

---

## Release Configuration Tasks [COMPLETE]

### ✅ T-CFG-1: Add ReleaseSettings to ProjectConfig [COMPLETED]
**Status**: COMPLETE - Added ReleaseSettings struct to internal/models/types.go

### ✅ T-CFG-2: Implement Config Save on Toggle [COMPLETED]
**Status**: COMPLETE - Auto-save on every toggle via saveConfig() in configure_handler.go

### ✅ T-CFG-3: Implement Config Load on Init [COMPLETED]
**Status**: COMPLETE - Loads all 4 tabs from config in NewConfigureModel

### ✅ T-CFG-6: Fix SkipTests Logic [COMPLETED]
**Status**: COMPLETE - Boolean inversion verified correct (checkbox enabled → SkipTests=false)

### ✅ T-CFG-7: Wire Config to ReleaseModel [COMPLETED]
**Status**: COMPLETE - ReleaseModel loads EnableHomebrew, EnableNPM, HomebrewTap from config

### T-CFG-8: Add Configuration Status Display [OPTIONAL]
**Status**: DEFERRED - Not needed for MVP, current checkboxes are sufficient

### T-CFG-9: Add Config Validation [OPTIONAL]
**Status**: DEFERRED - Save works without validation, can be added later if needed

## User Environment & Onboarding Tasks [NEW]

### ✅ T-EXTRA-1: Implement User Environment Detection [COMPLETED]
**File**: internal/detection/project.go
- ✅ DetectUserEnvironment() function
- ✅ Parse git config for name/email
- ✅ Parse gh CLI status for GitHub username
- ✅ Fixed parsing to handle "✓ Logged in to github.com account USERNAME" format
- ✅ Returns UserEnvironment struct
**Status**: Complete with proper gh CLI parsing
**Estimate**: 3 points

### ✅ T-EXTRA-2: Implement Smart Onboarding [COMPLETED]
**Files**: handlers/onboarding_handler.go, views/onboarding_view.go
- ✅ Created onboarding handler and view
- ✅ Auto-detect user configuration
- ✅ Minimal user input required
- ✅ Replaced with simple status indicator in main view
**Status**: Complete - simplified to status indicator
**Estimate**: 2 points

### ✅ T-EXTRA-3: GitHub Status Indicator [COMPLETED]
**File**: views/project_view.go
- ✅ Show GitHub username in green when configured
- ✅ Show warning with instructions when not configured
- ✅ Non-intrusive - part of main view
- ✅ Direct user to Settings > Edit to fix
**Status**: Complete
**Estimate**: 1 point

## GitHub Management Refactor Tasks [HIGH PRIORITY]

### ✅ T-GH-1: Create Cleanup Status Model [COMPLETED]
**File**: handlers/cleanup_handler.go (NEW)
- Create CleanupModel struct for status overview
- Implement loadRepoStatus() to check git/GitHub state
- Define RepoStatus enum (NoRepo, NoRemote, Unpushed, Clean)
- Add file counting logic (modified, new, deleted)
- Keep under 100 lines
**Estimate**: 2 points

### ✅ T-GH-2: Create Cleanup Status View [COMPLETED]
**File**: views/cleanup_view.go (NEW)
- Implement RenderCleanupStatus() function
- Display repository status with clear icons
- Show file change summary
- Add action hints ([G] GitHub, [C] Commit)
- Format with lipgloss styles
**Estimate**: 2 points

### ✅ T-GH-3: Create GitHub Management Model [COMPLETED]
**File**: handlers/github_handler.go (NEW)
- Create GitHubModel with state management
- Implement githubState enum (overview, create, connect, push)
- Add textinput fields for repo creation
- Handle GitHub CLI interactions
- Implement state transitions
**Estimate**: 3 points

### ✅ T-GH-4: Create GitHub Management View [COMPLETED]
**File**: views/github_view.go (NEW)
- Implement RenderGitHubManagement() function
- Show different UI based on githubState
- Create repo form with name/desc/visibility
- Connect existing repo form
- Style with focused/unfocused states
**Estimate**: 3 points

### ✅ T-GH-5: Create Commit Management Model [COMPLETED]
**File**: handlers/commit_handler.go (NEW)
**Status**: COMPLETE - CommitModel manages file selection and commit execution

### ✅ T-GH-6: Create Commit Management View [COMPLETED]
**File**: views/commit_view.go (NEW)
**Status**: COMPLETE - Renders commit interface with file checkboxes and message input

### ✅ T-GH-7: Update Configure Handler for Composition [COMPLETED]
**File**: handlers/configure_handler.go
- Add ViewType enum (TabView, GitHubView, CommitView)
- Integrate CleanupModel, GitHubModel, CommitModel
- Handle 'G' key to switch to GitHub view
- Handle 'C' key to switch to Commit view
- Route Update() calls to active sub-model
**Estimate**: 3 points

### ✅ T-GH-8: Update Configure View for Sub-Views [COMPLETED]
**File**: views/configure_view.go
- Check currentView type in RenderConfigureContent
- Delegate to appropriate sub-view renderer
- Maintain tab display for TabView
- Full-screen for GitHub/Commit views
- Handle view transitions smoothly
**Estimate**: 2 points

### ✅ T-GH-9: Simplify Git Status Logic [COMPLETED]
**File**: internal/gitcleanup/status.go
- Remove complex categorization
- Add GetRepoStatus() for simple state check
- Implement GetFileChanges() for clean file list
- Return user-friendly status strings
- Keep functions focused and simple
**Estimate**: 2 points

### ✅ T-GH-10: Add Repository State Detection [COMPLETED]
**File**: internal/gitcleanup/repo.go (NEW)
- Create CheckRepoState() function
- Detect git initialization
- Check for remote configuration
- Verify GitHub repo exists
- Return structured RepoInfo
**Estimate**: 2 points

## Project Management Tasks

### ✅ T026: Implement Configure View [COMPLETED]
**File**: views/configure_view.go
**Status**: Complete with tabbed interface and interactive lists

### ✅ T027: Implement Configure Handler [COMPLETED]
**File**: handlers/configure_handler.go
**Status**: Complete with full interactivity and config persistence

### ~~T028-T029: New Project Wizard [REMOVED]~~
**Reason**: Configure view IS the project setup. Once user configures distributions/build settings, they just run releases from project page. No separate wizard needed.



## Testing Tasks

### T032: [P] Test Config Management
**File**: internal/config/loader_test.go
- Test LoadGlobalConfig with valid YAML
- Test LoadGlobalConfig with missing file
- Test LoadProject with valid data
- Test SaveGlobalConfig creates file
- Use testify assertions
**Estimate**: 2 points

### T033: [P] Test Project Detection
**File**: internal/detection/project_test.go
- Test DetectProject with valid Go module
- Test DetectProject without go.mod
- Test GitHub detection parsing
- Mock gh CLI output
- Table-driven tests
**Estimate**: 2 points

### T034: [P] Test Command Executor
**File**: internal/executor/command_test.go
- Test RunCommand with echo
- Test command timeout handling
- Test output streaming
- Test error propagation
- Mock command execution
**Estimate**: 2 points

### T035: [P] Test View Rendering
**File**: views/project_view_test.go
- Test RenderProjectContent output
- Test with nil project
- Test with valid project data
- Verify keyboard hints present
- Check formatting consistency
**Estimate**: 1 point

### T036: [P] Test Handler Logic
**File**: handlers/project_handler_test.go
- Test navigation key handling
- Test action key triggering
- Test state transitions
- Verify returned page values
- Test quit behavior
**Estimate**: 2 points

### T037: Integration Test - Project Detection Flow
**File**: tests/integration/detection_test.go
- Create temp directory with go.mod
- Initialize git repository
- Run full detection
- Verify all fields populated
- Clean up test files
**Estimate**: 3 points

### T038: Integration Test - Release Flow
**File**: tests/integration/release_test.go
- Mock goreleaser command
- Execute full release flow
- Verify command sequence
- Check output messages
- Validate final state
**Estimate**: 3 points

### T039: Integration Test - Config Persistence
**File**: tests/integration/config_test.go
- Create and save config
- Reload and verify
- Modify and save again
- Test concurrent access
- Verify atomic writes
**Estimate**: 2 points



### T047: [P] Add Help Screen
**File**: views/help.go (create)
- Create help modal view
- List all keyboard shortcuts
- Explain navigation
- Show command descriptions
- Toggle with '?' key
**Estimate**: 1 point

## Execution Examples

### Parallel Execution Groups

**Group 1: Initial Setup (T003-T005)**
```bash
Task general-purpose "Complete T003, T004, and T005 in parallel - create all package structures, handlers, and view stubs"
```

**Group 2: Type Definitions (T008-T009, T030-T031)**
```bash
Task general-purpose "Complete T008, T009, T030, and T031 in parallel - define all types, messages, and commands"
```

**Group 3: Testing Suite (T032-T036)**
```bash
Task general-purpose "Complete T032 through T036 in parallel - implement all unit tests"
```

**Group 4: Polish Features (T044-T047)**
```bash
Task general-purpose "Complete T044 through T047 in parallel - add all UI polish features"
```

### Sequential Critical Path

1. T001-T002 (Setup template and module)
2. T003-T005 (Create structure) [P]
3. T006-T009 (Config management)
4. T010-T012 (Detection)
5. T013-T018 (Core views)
6. T019-T025 (Release execution)
7. T040-T041 (Wire up application)
8. T032-T039 (Testing)
9. T044-T047 (Polish) [P]

## Recent Additions (v0.0.21)

### NPM Package Name Validation Feature

**Files Created:**
- `internal/executor/npm_check.go` (76 lines) - NPM registry checking and name suggestions
- `handlers/npm_check_handler.go` (14 lines) - Bubble Tea async command handler

**Files Modified:**
- `handlers/configure_handler.go` - Added NPM validation state and message handling
- `views/configure_view.go` - Added NPM status display with suggestions
- Chrome calculations updated in 4 places to account for NPM status UI (3-7 lines)

**Functionality:**
1. **Automatic Validation**: When user enables NPM in Distributions tab, package name is checked against npm registry
2. **Visual Feedback**:
   - ⏳ Checking status (blue)
   - ✓ Available (green)
   - ✗ Unavailable (yellow) with suggestions
   - ✗ Error (red) with error details
3. **Smart Suggestions**:
   - Scoped package using GitHub username: `@username/package`
   - Common suffixes: `-cli`, `-tool`, `-release`, `-dist`
   - Shows up to 3 suggestions
4. **Terminal Layout Integrity**:
   - Handler calculates chrome including NPM status (3 lines for simple, 7 for with suggestions)
   - View uses same calculation to prevent overflow
   - Follows constitution principle for fixed terminal height

**UX Pattern**: Similar to regeneration warning - appears/disappears based on state, proper chrome accounting prevents layout issues.

## Notes

- All tasks must maintain < 100 lines per file (pragmatic: essential files may exceed if non-redundant)
- Use early returns to avoid nested conditionals
- No comments except API documentation
- Follow template's handler pattern exactly
- Test each component in isolation
- Terminal layout integrity: chrome calculations MUST be updated when adding UI lines
- Ensure atomic operations for all file I/O
---

# v0.0.32 Enhancement Tasks

**Target Version**: v0.0.32
**Date Added**: 2025-09-30
**Status**: Ready for Implementation
**Features**: Smart Commit Preferences, GitHub Workflow Generation, Dot File Bug Fix

## Overview

This section adds tasks for three enhancements to the production-ready v0.0.31 release:
1. Project-level smart commit file categorization customization
2. Optional GitHub Actions workflow generation (opt-in)
3. Bug fix for dot file handling in commit settings

### Task Count
- Setup: 3 tasks
- Bug Fix: 2 tasks  
- Smart Commit Preferences: 10 tasks
- Workflow Generation: 10 tasks
- Integration: 5 tasks
- Polish: 4 tasks
**Total**: 34 new tasks

---

## Phase 1: Setup Tasks (v0.0.32)

### ☑ T-V32-001: Add doublestar Dependency to go.mod [COMPLETED]
**File**: go.mod
**Description**: Add github.com/bmatcuk/doublestar/v4 for glob pattern matching
**Actions**:
- Run `go get github.com/bmatcuk/doublestar/v4`
- Run `go mod tidy`
- Verify import works with `go build`
**Estimate**: 1 point
**Dependencies**: None

### ☑ T-V32-002: Create internal/workflow Package Structure [COMPLETED]
**Files**: internal/workflow/
**Description**: Create new package directory for workflow generation logic
**Actions**:
- Create `internal/workflow/` directory
- Add package declaration placeholder
- No implementation yet, just structure
**Estimate**: 1 point
**Dependencies**: None

### ☑ T-V32-003: Update internal/models/types.go with New Structs [COMPLETED]
**File**: internal/models/types.go
**Description**: Add CategoryRules, SmartCommitPrefs, WorkflowConfig struct definitions
**Actions**:
```go
type CategoryRules struct {
    Extensions []string `yaml:"extensions"`
    Patterns   []string `yaml:"patterns"`
}

type SmartCommitPrefs struct {
    Enabled        bool                        `yaml:"enabled"`
    UseCustomRules bool                        `yaml:"use_custom_rules"`
    Categories     map[string]CategoryRules    `yaml:"categories"`
}

type WorkflowConfig struct {
    Enabled          bool     `yaml:"enabled"`
    WorkflowPath     string   `yaml:"workflow_path"`
    AutoRegenerate   bool     `yaml:"auto_regenerate"`
    IncludeTests     bool     `yaml:"include_tests"`
    Environments     []string `yaml:"environments"`
    SecretsRequired  []string `yaml:"secrets_required"`
}
```
- Add to existing ProjectConfig struct fields
- Follow existing YAML tag patterns
**Estimate**: 2 points
**Dependencies**: None

---

## Phase 2: Bug Fix Tasks (High Priority)

### ☑ T-V32-004: Fix Dot File Handling in Git Cleanup [COMPLETED]
**File**: internal/gitcleanup/categorize.go
**Description**: Fix bug where files/directories starting with "." cannot be modified in commit settings
**Root Cause**: Likely over-aggressive hidden file filtering
**Actions**:
- Review file listing logic
- Ensure dot files included (except .git/ itself)
- Test with .github/, .goreleaser.yaml, .env files
- Verify categorization works for dot files
**Acceptance**:
- .github/workflows/test.yml can be committed
- .goreleaser.yaml shows in cleanup tab
- .git/ directory still excluded
**Estimate**: 2 points
**Dependencies**: None

### ☐ T-V32-005: [P] Add Test for Dot File Categorization
**File**: internal/gitcleanup/categorize_test.go
**Description**: Add table-driven test covering dot file scenarios
**Test Cases**:
- `.github/workflows/release.yml` → build category
- `.goreleaser.yaml` → build category
- `.env` → config category
- `.gitignore` → config category
- `.git/config` → should be excluded
**Estimate**: 2 points
**Dependencies**: T-V32-004

---

## Phase 3: Smart Commit Preferences Tasks

### ☑ T-V32-006: [P] Add smart_commit Section Parsing to Config Loader [COMPLETED]
**File**: internal/config/loader.go
**Description**: Update LoadProject() to parse smart_commit YAML section
**Actions**:
- Add parsing for smart_commit section
- Set defaults if section missing:
  ```go
  if config.Config.SmartCommit == nil {
      config.Config.SmartCommit = getDefaultSmartCommitPrefs()
  }
  ```
- Implement getDefaultSmartCommitPrefs() with hardcoded rules
- Update SaveProject() to serialize smart_commit
**Estimate**: 3 points
**Dependencies**: T-V32-003

### ☑ T-V32-007: [P] Create Pattern Matching Logic with doublestar [COMPLETED]
**File**: internal/gitcleanup/matcher.go (NEW)
**Description**: Implement pattern matching functions using doublestar library
**Functions Needed**:
```go
// MatchesPattern checks if path matches any pattern in list
func MatchesPattern(path string, patterns []string) (bool, error)

// MatchesExtension checks if file has extension in list
func MatchesExtension(path string, extensions []string) bool

// CategorizeWithRules applies custom or default rules
func CategorizeWithRules(path string, rules map[string]CategoryRules) string
```
**Actions**:
- Import doublestar: `github.com/bmatcuk/doublestar/v4`
- Handle errors from glob matching
- Use case-insensitive extension matching
- Return "other" if no category matches
**Estimate**: 3 points
**Dependencies**: T-V32-001, T-V32-003

### ☑ T-V32-008: Update Categorize.go to Use Custom Rules [COMPLETED]
**File**: internal/gitcleanup/categorize.go
**Description**: Modify CategorizeFile() to check for custom rules and use them
**Logic**:
```go
func CategorizeFile(path string, projectConfig *models.ProjectConfig) string {
    if projectConfig.Config.SmartCommit != nil && 
       projectConfig.Config.SmartCommit.UseCustomRules {
        return CategorizeWithRules(path, projectConfig.Config.SmartCommit.Categories)
    }
    return categorizeWithDefaults(path) // existing logic
}
```
**Actions**:
- Add projectConfig parameter to CategorizeFile
- Check UseCustomRules flag
- Fall back to defaults if custom disabled
- Maintain backward compatibility
**Estimate**: 2 points
**Dependencies**: T-V32-004, T-V32-006, T-V32-007

### ☑ T-V32-009: Create handlers/smart_commit_prefs_handler.go [COMPLETED]
**File**: handlers/smart_commit_prefs_handler.go (NEW)
**Description**: Create Bubble Tea handler for smart commit preferences editing
**Model Struct**:
```go
type SmartCommitPrefsModel struct {
    ProjectConfig     *models.ProjectConfig
    SelectedCategory  int          // Index in category list
    EditMode          bool         // Editing extensions/patterns
    ExtensionInput    textinput.Model
    PatternInput      textinput.Model
    Width             int
    Height            int
}
```
**Key Functions**:
- `NewSmartCommitPrefsModel()` - Initialize with project config
- `Update()` - Handle key presses (↑/↓ navigate, e edit, a add, d delete, r reset, s save)
- `ToggleCustomRules()` - Enable/disable custom rules
- `AddExtension()`, `RemoveExtension()` - Modify extension list
- `AddPattern()`, `RemovePattern()` - Modify pattern list
- `ResetToDefaults()` - Clear custom rules, use defaults
**Actions**:
- Keep file under 300 lines (constitution guideline)
- Use early returns, no nested conditionals
- Self-documenting names
**Estimate**: 5 points
**Dependencies**: T-V32-006, T-V32-007

### ☑ T-V32-010: Create views/smart_commit_prefs_view.go [COMPLETED]
**File**: views/smart_commit_prefs_view.go (NEW)
**Description**: Render UI for smart commit preferences
**Layout**:
```
┌─ Smart Commit Preferences ─────────┐
│ [✓] Use Custom Rules               │
│                                     │
│ Categories:                         │
│ > code                              │
│   config                            │
│   docs                              │
│   ...                               │
│                                     │
│ code Category:                      │
│ Extensions: .go, .js, .ts          │
│ Patterns: **/src/**, **/lib/**     │
│                                     │
│ [r] Reset  [s] Save  [ESC] Cancel  │
└─────────────────────────────────────┘
```
**Actions**:
- Use lipgloss for styling
- Show selected category highlighted
- Display extensions and patterns for selected category
- Show keyboard hints at bottom
- Keep under 200 lines
**Estimate**: 4 points
**Dependencies**: T-V32-009

### ☑ T-V32-011: Integrate Smart Commit Prefs into Configure View Cleanup Tab [COMPLETED]
**Files**: handlers/configure_handler.go, views/configure_view.go, views/cleanup_view.go, handlers/smart_commit_prefs_handler.go
**Description**: Added smart commit preferences as nested view in Cleanup tab (project-level settings)
**Implementation**:
- Added `SmartCommitPrefsView` to ViewType enum
- Added `SmartCommitPrefsModel` field to ConfigureModel
- Added 'p' key handler in Cleanup tab to open preferences
- Added view delegation case in UpdateConfigureView
- Updated configure_view.go to render SmartCommitPrefsView
- Updated cleanup_view.go actions to show [p] Preferences hint
- Added SetSize method for proper dimension handling
- Updated WindowSizeMsg handling to resize SmartCommitPrefsModel
**Pattern**: Follows exact pattern of GitHubView and CommitView integration
**Estimate**: 3 points
**Dependencies**: T-V32-009, T-V32-010

### ☑ T-V32-012: Add Default Rules Reset Functionality [COMPLETED]
**File**: handlers/smart_commit_prefs_handler.go
**Description**: Implemented reset to defaults with confirmation
**Implementation**:
- Added ShowConfirm bool field to model
- Added handleConfirm method that handles 'y' and 'n' keys
- Added resetToDefaults method: sets UseCustomRules = false, clears Categories map, saves config
- Pressing 'r' in normal mode triggers ShowConfirm = true
- View renders confirmation modal: "Reset custom rules to defaults?" with [y] Yes [n] No
- After reset, returns to preferences view showing defaults
**Estimate**: 2 points
**Dependencies**: T-V32-009

### ☐ T-V32-013: [P] Write Unit Tests for Pattern Matching
**File**: internal/gitcleanup/matcher_test.go (NEW)
**Description**: Table-driven tests for pattern matching logic
**Test Cases**:
- Extension matching: ".go" matches "file.go"
- Pattern matching: "**/src/**" matches "project/src/main.go"
- Globstar: "**/test/**" matches "a/b/c/test/d/file.go"
- Case insensitive: ".GO" matches ".go"
- No match: returns "other"
- Multiple patterns: first match wins
**Estimate**: 3 points
**Dependencies**: T-V32-007

### ☐ T-V32-014: [P] Write Integration Tests for Preferences UI
**File**: handlers/smart_commit_prefs_handler_test.go (NEW)
**Description**: Test handler update logic with Bubble Tea test messages
**Test Scenarios**:
- Toggle custom rules on/off
- Navigate categories with arrow keys
- Add extension: press 'a', type ".proto", press enter
- Remove extension: select, press 'd'
- Reset to defaults: press 'r', confirm
- Save preferences: press 's'
**Estimate**: 3 points
**Dependencies**: T-V32-009

---

## Phase 4: Workflow Generation Tasks

### ☑ T-V32-015: [P] Create internal/workflow/template.go with Embedded YAML [COMPLETED]
**File**: internal/workflow/template.go (NEW)
**Description**: Define GitHub Actions workflow template as Go text/template
**Template Structure**:
```yaml
name: Release
on:
  push:
    tags: ['v*']
  workflow_dispatch:

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      {{if .IncludeTests}}
      - name: Run tests
        run: go test ./...
      {{end}}
      - uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          {{if .NPMEnabled}}
          NPM_TOKEN: ${{ secrets.NPM_TOKEN }}
          {{end}}
```
**Actions**:
- Use `text/template` package
- Define template variables: IncludeTests, NPMEnabled, HomebrewEnabled
- Add comments documenting required secrets
- Keep template readable and maintainable
**Estimate**: 4 points
**Dependencies**: T-V32-002

### ☑ T-V32-016: [P] Create internal/workflow/generator.go [COMPLETED]
**File**: internal/workflow/generator.go (NEW)
**Description**: Implement workflow generation logic
**Functions Needed**:
```go
// GenerateWorkflow creates YAML from template and config
func GenerateWorkflow(config *models.ProjectConfig) (string, error)

// ValidateWorkflow checks generated YAML is valid
func ValidateWorkflow(yamlContent string) error

// GetRequiredSecrets returns list of secrets needed
func GetRequiredSecrets(config *models.ProjectConfig) []string

// WriteWorkflowFile writes YAML to .github/workflows/release.yml
func WriteWorkflowFile(projectPath, yamlContent string) error
```
**Actions**:
- Execute template with config data
- Validate YAML syntax (basic check)
- Create .github/workflows/ directory if needed
- Atomic write (temp file + rename)
- Return list of required secrets
**Estimate**: 4 points
**Dependencies**: T-V32-015

### ☑ T-V32-017: Add workflow_generation Section Parsing to Config Loader [COMPLETED]
**File**: internal/config/loader.go
**Description**: Parse ci_cd.github_actions section from YAML
**Actions**:
- Add parsing for ci_cd.github_actions section
- Set defaults if missing:
  ```go
  if config.Config.CICD.GitHubActions == nil {
      config.Config.CICD.GitHubActions = &models.WorkflowConfig{
          Enabled: false,
          WorkflowPath: ".github/workflows/release.yml",
          IncludeTests: true,
      }
  }
  ```
- Update SaveProject() to serialize workflow config
**Estimate**: 2 points
**Dependencies**: T-V32-003, T-V32-006

### ☑ T-V32-018: Create handlers/workflow_gen_handler.go [COMPLETED]
**File**: handlers/workflow_gen_handler.go (NEW)
**Description**: Handler for workflow generation UI
**Model Struct**:
```go
type WorkflowGenModel struct {
    ProjectConfig     *models.ProjectConfig
    PreviewMode       bool
    PreviewContent    string
    GeneratedYAML     string
    RequiredSecrets   []string
    ShowConfirmation  bool
    Width             int
    Height            int
}
```
**Key Functions**:
- `NewWorkflowGenModel()` - Initialize
- `Update()` - Handle key presses (space toggle, p preview, g generate, ESC cancel)
- `ToggleEnabled()` - Enable/disable workflow generation
- `PreviewWorkflow()` - Generate and show YAML preview
- `GenerateWorkflow()` - Create file with user confirmation
- `CheckIfWorkflowExists()` - Detect existing workflow file
**Actions**:
- Check if file exists before overwriting
- Show required secrets list
- Add confirmation modal before file creation
- Keep under 300 lines
**Estimate**: 5 points
**Dependencies**: T-V32-016, T-V32-017

### ☑ T-V32-019: Create views/workflow_gen_view.go [COMPLETED]
**File**: views/workflow_gen_view.go (NEW)
**Description**: Render workflow generation UI
**Layout**:
```
┌─ GitHub Actions Workflow ──────────┐
│ [✓] Enable Workflow Generation     │
│                                     │
│ Options:                            │
│ [✓] Include Tests                  │
│ [✓] Auto-regenerate on config      │
│                                     │
│ Required Secrets:                   │
│  • GITHUB_TOKEN (automatic)        │
│  • NPM_TOKEN (if NPM enabled)      │
│                                     │
│ [p] Preview  [g] Generate          │
│ [ESC] Cancel                        │
└─────────────────────────────────────┘
```
**Preview Modal**:
```
┌─ Workflow Preview ──────────────────┐
│ name: Release                       │
│ on:                                 │
│   push:                             │
│     tags: ['v*']                    │
│ ...                                 │
│                                     │
│ [ESC] Close                         │
└─────────────────────────────────────┘
```
**Actions**:
- Use lipgloss for styling
- Syntax highlighting for YAML (basic, optional)
- Scrollable preview if content too long
- Keep under 200 lines
**Estimate**: 4 points
**Dependencies**: T-V32-018

### ☐ T-V32-020: Integrate Workflow Gen into Configure View Advanced Tab
**Files**: handlers/configure_handler.go, views/configure_view.go
**Description**: Add workflow generation section to Advanced tab
**Changes in configure_handler.go**:
- Add `WorkflowGenModel` field to ConfigureModel
- Initialize in NewConfigureModel()
- Add key binding 'w' for workflow in Advanced tab
- Route to workflow model when 'w' pressed
**Changes in configure_view.go**:
- Add "GitHub Workflow Generation" section in Advanced tab
- Show enabled state and shortcut '[w] Configure Workflow'
- Display warning if secrets missing
**Estimate**: 3 points
**Dependencies**: T-V32-018, T-V32-019

### ☐ T-V32-021: Add Preview Modal for Workflow YAML
**File**: views/workflow_gen_view.go
**Description**: Implement scrollable preview modal showing generated YAML
**Features**:
- Full-screen modal overlay
- Scrollable content (↑/↓ or j/k)
- Line numbers optional
- ESC to close
- Proper word wrapping for long lines
**Actions**:
- Use lipgloss.Place() for centering
- Add scroll state to WorkflowGenModel
- Handle WindowSizeMsg for responsive sizing
**Estimate**: 3 points
**Dependencies**: T-V32-019

### ☐ T-V32-022: Add File Generation with User Consent
**File**: handlers/workflow_gen_handler.go
**Description**: Implement workflow file creation with confirmation modal
**Confirmation Flow**:
1. User presses 'g' to generate
2. Check if .github/workflows/release.yml exists
3. If exists: "File exists. Overwrite?" [y/n]
4. If new: "Create workflow file?" [y/n]
5. On confirm: call WriteWorkflowFile()
6. Show success message with file path
**Actions**:
- Add confirmation modal state
- Handle yes/no key presses
- Create directory if needed
- Show error if write fails (permissions, etc.)
**Estimate**: 3 points
**Dependencies**: T-V32-018, T-V32-021

### ☐ T-V32-023: [P] Write Tests for Template Generation
**File**: internal/workflow/template_test.go (NEW)
**Description**: Test workflow template with various configs
**Test Cases**:
- NPM enabled: includes NPM_TOKEN secret
- NPM disabled: no NPM_TOKEN
- Tests enabled: includes test step
- Tests disabled: no test step
- Homebrew: handled by GoReleaser (no special workflow changes)
- All combinations: NPM+Tests, NPM only, Tests only, neither
**Estimate**: 3 points
**Dependencies**: T-V32-015, T-V32-016

### ☐ T-V32-024: [P] Write Tests for Workflow Validation
**File**: internal/workflow/generator_test.go (NEW)
**Description**: Test workflow generation and validation
**Test Cases**:
- Generated YAML is valid
- Required secrets detected correctly
- File write creates .github/workflows/ directory
- Overwrites existing file correctly
- Handles permission errors gracefully
**Estimate**: 3 points
**Dependencies**: T-V32-016

---

## Phase 5: Integration Tasks

### ☐ T-V32-025: Update Configure View to Handle Advanced Tab Expansion
**File**: handlers/configure_handler.go
**Description**: Expand Advanced tab to accommodate two new feature sections
**Actions**:
- Increase Advanced tab height allocation if needed
- Add routing logic for 'p' (prefs) and 'w' (workflow) keys
- Update chrome calculations for new UI lines
- Ensure scrolling works if content exceeds screen
**Chrome Update**:
- Smart Commit Prefs section: +3 lines
- Workflow Generation section: +4 lines
- Total Advanced tab increase: +7 lines
**Estimate**: 2 points
**Dependencies**: T-V32-011, T-V32-020

### ☐ T-V32-026: Add Navigation Between Smart Commit Prefs and Workflow Gen
**File**: handlers/configure_handler.go
**Description**: Allow seamless navigation between the two features
**Navigation**:
- From Advanced tab: 'p' for prefs, 'w' for workflow
- From Prefs: ESC back to Advanced tab
- From Workflow: ESC back to Advanced tab
- From Prefs: 'w' to switch to Workflow (optional)
- From Workflow: 'p' to switch to Prefs (optional)
**Actions**:
- Add view state tracking (advanced/prefs/workflow)
- Route key presses to appropriate sub-model
- Maintain scroll position when navigating back
**Estimate**: 2 points
**Dependencies**: T-V32-011, T-V32-020

### ☐ T-V32-027: Wire Up Save/Load for Both Feature Configs
**File**: internal/config/loader.go
**Description**: Ensure both smart_commit and ci_cd sections persist correctly
**Actions**:
- Test save after modifying smart commit prefs
- Test save after enabling workflow generation
- Verify YAML serialization is correct
- Test load on app restart
- Ensure atomic writes work for both sections
**Test Manually**:
```bash
# Edit prefs, save, restart
distui # verify prefs loaded

# Enable workflow, save, restart
distui # verify workflow config loaded
```
**Estimate**: 2 points
**Dependencies**: T-V32-006, T-V32-017

### ☐ T-V32-028: Test Full Integration with Existing Features
**Description**: Manual integration testing of new features with v0.0.31 features
**Test Scenarios**:
1. **Smart Commit with Custom Rules**:
   - Configure custom rules for .proto → code
   - Create .proto file
   - Open Cleanup tab
   - Verify .proto categorized as "code"
   - Commit file using smart commit

2. **Workflow Generation with NPM**:
   - Enable NPM distribution
   - Enable workflow generation
   - Preview workflow
   - Verify NPM_TOKEN in required secrets
   - Generate workflow file
   - Verify .github/workflows/release.yml created

3. **Dot File Handling**:
   - Create .github/workflows/test.yml
   - Open Cleanup tab
   - Verify file visible and committable

4. **Configuration Persistence**:
   - Set custom rules + enable workflow
   - Save, exit distui
   - Restart distui
   - Verify settings loaded correctly

**Acceptance**: All scenarios pass without errors
**Estimate**: 3 points
**Dependencies**: T-V32-011, T-V32-020, T-V32-027

### ☐ T-V32-029: Update Quickstart.md Validation
**File**: specs/001-build-a-terminal/quickstart.md
**Description**: Validate all quickstart scenarios work with implementation
**Actions**:
- Run through all 6 test scenarios
- Verify expected outcomes match actual behavior
- Update any steps that changed during implementation
- Add troubleshooting notes if needed
- Mark all scenarios as validated
**Estimate**: 2 points
**Dependencies**: T-V32-028

---

## Phase 6: Polish Tasks

### ☐ T-V32-030: [P] Add Error Handling for Invalid Patterns
**Files**: internal/gitcleanup/matcher.go, handlers/smart_commit_prefs_handler.go
**Description**: Validate glob patterns before saving
**Validation Rules**:
- Reject patterns with unmatched brackets: `[[[invalid`
- Reject empty patterns
- Test pattern with doublestar before saving
- Show error message to user if invalid
**Error Messages**:
- "Invalid pattern: [pattern]"
- "Pattern must not be empty"
- "Pattern syntax error: [details]"
**Actions**:
- Add ValidatePattern() function
- Call before adding pattern to list
- Display error in UI (red text)
**Estimate**: 2 points
**Dependencies**: T-V32-007, T-V32-009

### ☐ T-V32-031: [P] Add Loading States for Async Operations
**Files**: handlers/smart_commit_prefs_handler.go, handlers/workflow_gen_handler.go
**Description**: Show spinner during file I/O operations
**Operations to Cover**:
- Loading project config
- Saving preferences
- Generating workflow file
- Validating workflow YAML
**Actions**:
- Add spinner model to handlers
- Show "Saving..." with spinner
- Show "Generating..." with spinner
- Hide spinner on completion or error
**Estimate**: 2 points
**Dependencies**: T-V32-009, T-V32-018

### ☐ T-V32-032: [P] Add Keyboard Shortcuts Documentation
**File**: README.md or docs/SHORTCUTS.md (if exists)
**Description**: Document new keyboard shortcuts for v0.0.32
**New Shortcuts**:
- Advanced Tab:
  - `p` - Edit Smart Commit Preferences
  - `w` - Configure GitHub Workflow
- Smart Commit Prefs:
  - `↑/↓` - Navigate categories
  - `e` - Edit selected category
  - `a` - Add extension/pattern
  - `d` - Delete extension/pattern
  - `r` - Reset to defaults
  - `s` - Save preferences
- Workflow Gen:
  - `space` - Toggle enabled
  - `p` - Preview workflow
  - `g` - Generate file
**Estimate**: 1 point
**Dependencies**: None (documentation only)

### ☐ T-V32-033: [P] Performance Testing for Pattern Matching
**File**: internal/gitcleanup/matcher_bench_test.go (NEW)
**Description**: Benchmark pattern matching performance
**Benchmarks**:
```go
func BenchmarkMatchesPattern(b *testing.B)
func BenchmarkMatchesExtension(b *testing.B)
func BenchmarkCategorizeWithRules(b *testing.B)
```
**Performance Targets** (from research.md):
- Pattern matching: <1ms for 100 files
- Extension matching: <0.1ms per file
- Full categorization: <100ms for typical project
**Actions**:
- Run benchmarks: `go test -bench=. ./internal/gitcleanup`
- Verify meets performance targets
- Optimize if needed (unlikely)
**Estimate**: 2 points
**Dependencies**: T-V32-007

---

## Execution Guide for v0.0.32

### Sequential Task Order
1. Run Setup tasks (T-V32-001 to T-V32-003) first
2. Run Bug Fix tasks (T-V32-004 to T-V32-005)
3. Run Smart Commit Preferences tasks (T-V32-006 to T-V32-014)
4. Run Workflow Generation tasks (T-V32-015 to T-V32-024)
5. Run Integration tasks (T-V32-025 to T-V32-029)
6. Run Polish tasks (T-V32-030 to T-V32-033)

### Parallel Execution Opportunities

**Can Run in Parallel After Setup**:
```bash
# After T-V32-003, run these together:
- T-V32-005 (dot file test)
- T-V32-006 (config parsing)
- T-V32-007 (pattern matching)
```

**Smart Commit Tests (Parallel)**:
```bash
# After T-V32-012, run these together:
- T-V32-013 (pattern matching tests)
- T-V32-014 (UI integration tests)
```

**Workflow Tasks (Parallel)**:
```bash
# After T-V32-002, run these together:
- T-V32-015 (template)
- T-V32-016 (generator)
```

**Workflow Tests (Parallel)**:
```bash
# After T-V32-022, run these together:
- T-V32-023 (template tests)
- T-V32-024 (generator tests)
```

**Polish Tasks (All Parallel After Integration)**:
```bash
# After T-V32-029, run these together:
- T-V32-030 (error handling)
- T-V32-031 (loading states)
- T-V32-032 (documentation)
- T-V32-033 (performance testing)
```

### Testing Checkpoints
- After T-V32-005: Run dot file tests
- After T-V32-013: Run pattern matching tests
- After T-V32-014: Run preferences UI tests
- After T-V32-023: Run workflow template tests
- After T-V32-024: Run workflow generator tests
- After T-V32-028: Full integration testing
- After T-V32-033: Performance validation

### Success Criteria
- All 34 tasks completed
- All tests passing (unit + integration)
- Performance targets met (<100ms categorization)
- Quickstart scenarios validated
- No regressions in v0.0.31 features
- Ready for v0.0.32 release

---

## Implementation Notes for v0.0.32

### Code Quality Reminders
- Keep files under 300 lines (strong refactoring target)
- Use early returns, avoid nested conditionals
- No comments except API docs
- Self-documenting names
- Test each component in isolation

### Constitutional Compliance
- Smart commit prefs in ~/.distui (not in repo)
- Workflow generation requires explicit user consent
- Easy to disable all new features
- User maintains full control
- No forced behaviors

### Terminal Layout Integrity
When adding UI elements:
- **Smart Commit Prefs section in Advanced**: +3 lines
- **Workflow Generation section in Advanced**: +4 lines
- Update chrome calculations in configure_handler.go
- Test with small terminal size (80x24)
- Ensure no overflow/scrolling issues

### Integration with Existing Code
- Smart commit prefs integrates with Cleanup tab
- Workflow generation reads from Distributions config
- Both use existing config loader patterns
- No breaking changes to existing APIs

---
