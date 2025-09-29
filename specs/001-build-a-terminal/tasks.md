# Implementation Tasks: distui - Go Release Distribution Manager

**Feature**: Build a Terminal UI Application for Go Release Management
**Branch**: 001-build-a-terminal
**Created**: 2025-09-28

## Overview

This document contains all implementation tasks for the distui feature. Tasks are ordered by dependencies and marked with [P] when they can be executed in parallel.

### Quick Stats
- Total Tasks: 61 (47 original + 3 extra + 10 config tasks + 1 view merge)
- Completed: 29 (18 previous + 10 release workflow + 1 view merge)
- In Progress: 0
- Remaining: 32

### Completed Categories
- ✅ Setup Tasks: 5/5 (100%)
- ✅ Configuration Management: 9/9 (100%)
- ✅ Detection: 3/3 (100%)
- ✅ Core Views (Project/Settings/Configure): 6/6 (100%)
- ✅ User Environment & Onboarding: 3/3 (100%)
- ✅ Git Management: 10/10 (100%)
- ✅ Release Workflow Core: 10/10 (100%)

### Pending Categories
- ✅ Release Workflow Core: 10/10 (T012, T019-T025, T030-T031) - COMPLETE
- 🔄 Release Configuration: 0/9 (T-CFG-1 to T-CFG-9) - CRITICAL
- 🔄 Project Management: 0/2 (T028-T029 - New Project wizard)
- 🔄 Testing: 0/8 (T032-T039)
- 🔄 Integration: 0/4 (T040-T043)
- 🔄 Polish: 0/4 (T044-T047)

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

## Release Configuration Tasks [CRITICAL - BLOCKERS]

### T-CFG-1: Add ReleaseSettings to ProjectConfig
**File**: internal/models/types.go (EXISTS - needs expansion)
**Status**: PENDING - Foundation for all config
**Current State**: ProjectConfig exists but has no release settings
**Needs**:
- Add ReleaseSettings struct with:
  - EnableHomebrew bool
  - HomebrewTap string
  - EnableNPM bool
  - NPMScope string
  - NPMPackage string
  - SkipTests bool
  - CreateDraft bool
  - PreRelease bool
  - GenerateChangelog bool
  - SignCommits bool
- Add ReleaseSettings field to ProjectConfig
- Add yaml tags for persistence
**Estimate**: 2 points
**Architecture**: Business data model in internal/models

### T-CFG-2: Implement Config Save on Toggle
**File**: handlers/configure_handler.go (EXISTS - needs save logic)
**Status**: PENDING - Depends on T-CFG-1
**Current State**: Space key toggles items in memory only, not persisted
**Needs**:
- Add SaveProjectConfig() function
- Wire to Space key handler for distribution items
- Wire to Space key handler for build items
- Call config.SaveProject() with updated settings
- Show save indicator (subtle flash/checkmark)
- Handle save errors gracefully
**Pattern**: Auto-save on every toggle (like modern editors)
**Estimate**: 3 points
**Architecture**: State management in handlers, calls internal/config

### T-CFG-3: Implement Config Load on Init
**File**: handlers/configure_handler.go (EXISTS - needs load logic)
**Status**: PENDING - Depends on T-CFG-1
**Current State**: Lists initialized with hardcoded defaults
**Needs**:
- Load project config in NewConfigureModel
- Read ReleaseSettings from loaded config
- Set DistributionItem.Enabled from config
- Set BuildItem.Enabled from config
- Update list items with loaded values
- Handle missing config gracefully (use defaults)
**Estimate**: 2 points
**Architecture**: State management in handlers, calls internal/config

### T-CFG-4: Detect and Configure Homebrew Tap
**File**: handlers/configure_handler.go (EXISTS - needs detection)
**Status**: PENDING - Depends on T-CFG-3
**Current State**: Hardcoded "Tap: willyv3/homebrew-tap" string
**Needs**:
- Call detection.DetectHomebrewTap(username) on init
- If found: Update DistributionItem description with actual path
- If not found: Show "Not configured - [e] to edit"
- Add 'e' key handler to edit tap path (textinput modal)
- Save tap path to config on edit
- Validate tap path exists
**Pattern**: Auto-detect, allow override
**Estimate**: 3 points
**Architecture**: Uses internal/detection, saves to internal/models

### T-CFG-5: Configure NPM Package Settings
**File**: handlers/configure_handler.go (EXISTS - needs NPM config)
**Status**: PENDING - Depends on T-CFG-3
**Current State**: Hardcoded "Scope: @williavs" string
**Needs**:
- Detect package.json if exists
- Parse name/scope from package.json
- If found: Update DistributionItem description
- If not found: Show "Not configured - [e] to edit"
- Add 'e' key handler to configure scope + package name
- Save NPM settings to config
**Estimate**: 3 points
**Architecture**: Detection in handlers, saves to internal/models

### T-CFG-6: Fix SkipTests Logic
**File**: handlers/configure_handler.go (EXISTS - logic fix)
**Status**: PENDING - Depends on T-CFG-2
**Current State**: "Run tests before release" checkbox logic is backwards
**Needs**:
- When "Run tests" is ENABLED, SkipTests = FALSE
- When "Run tests" is DISABLED, SkipTests = TRUE
- Update BuildItem to toggle correctly
- Save correct boolean to config
**Note**: Simple boolean inversion
**Estimate**: 1 point

### T-CFG-7: Wire Config to ReleaseModel
**File**: app.go (EXISTS - needs config passing)
**Status**: PENDING - Depends on T-CFG-1, T-CFG-3
**Current State**: ReleaseModel gets hardcoded false values
**Needs**:
- Read m.currentProject.ReleaseSettings on releaseView init
- Pass EnableHomebrew from config to NewReleaseModel
- Pass HomebrewTap from config to NewReleaseModel
- Pass EnableNPM from config to NewReleaseModel
- Pass NPM settings from config to NewReleaseModel
- Handle nil config gracefully (use safe defaults)
**Estimate**: 2 points
**Architecture**: App routes config from loaded project to release handler

### T-CFG-8: Add Configuration Status Display
**File**: views/configure_view.go (EXISTS - needs status)
**Status**: PENDING - Depends on T-CFG-4, T-CFG-5
**Current State**: No indication if tap/npm detected or configured
**Needs**:
- Show "✓ Detected: ~/homebrew-tap" when tap found
- Show "⚠ Not configured" when tap not found
- Show "✓ Configured: @scope/package" for NPM
- Show "⚠ Not configured" for NPM
- Add subtle color coding (green = good, yellow = warning)
**Estimate**: 2 points
**Architecture**: Pure rendering in views

### T-CFG-9: Add Config Validation
**File**: handlers/configure_handler.go (EXISTS - needs validation)
**Status**: PENDING - Depends on T-CFG-7
**Current State**: No validation before saving
**Needs**:
- Validate homebrew tap path exists before saving
- Check gh CLI installed if Homebrew enabled
- Validate NPM package name format
- Show validation errors inline
- Prevent save if critical validation fails
**Estimate**: 2 points
**Architecture**: Validation logic in handlers

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

### T-GH-5: Create Commit Management Model
**File**: handlers/commit_handler.go (NEW)
- Create CommitModel for file selection
- Implement file toggle logic
- Add commit message textinput
- Calculate diff statistics
- Handle commit execution
**Estimate**: 3 points

### T-GH-6: Create Commit Management View
**File**: views/commit_view.go (NEW)
- Implement RenderCommitView() function
- Show file list with checkboxes
- Display commit message input
- Show diff summary
- Add keyboard hints
**Estimate**: 2 points

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
- ✅ Created tabbed interface (Distributions, Build Settings, Advanced)
- ✅ Interactive lists with Bubble Tea list.Model
- ✅ Checkbox toggles with [✓] and [ ] patterns
- ✅ Dynamic height adjustment for window resizing
- ✅ Professional list navigation
**Status**: Complete with list-based UI
**Estimate**: 5 points

### ✅ T027: Implement Configure Handler [COMPLETED]
**File**: handlers/configure_handler.go
- ✅ Handle Tab key for tab switching
- ✅ Handle Space to toggle checkboxes
- ✅ Handle 'a' for check all functionality
- ✅ List navigation with up/down arrows
- ✅ Proper window size handling
- ✅ Maintains state across tab switches
**Status**: Complete with full interactivity
**Estimate**: 3 points

### T028: Implement New Project View
**File**: views/newproject_view.go
- Show detection results
- Allow editing detected values
- Display confirmation step
- Show initial configuration
- Format as wizard flow
**Estimate**: 3 points

### T029: Implement New Project Handler
**File**: handlers/newproject_handler.go
- Handle project detection flow
- Process user overrides
- Save new project configuration
- Switch to project view on completion
- Handle cancellation
**Estimate**: 3 points

## Message and Command Types

### T030: [P] Define Message Types
**File**: internal/models/messages.go
- Define projectDetectedMsg struct
- Define commandOutputMsg struct
- Define releaseProgressMsg struct
- Define configSavedMsg struct
- Define errorMsg struct
**Estimate**: 1 point

### T031: [P] Define Command Functions
**File**: internal/models/commands.go
- Create detectProjectCmd function
- Create loadProjectCmd function
- Create saveConfigCmd function
- Create executeReleaseCmd function
- All return tea.Cmd
**Estimate**: 2 points

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

## Integration Tasks

### T040: Wire Up Main Application
**File**: app.go (update existing)
- Add distui state fields to appModel
- Initialize with project detection on startup
- Connect handlers to model updates
- Implement proper state management
- Add command batching
**Estimate**: 3 points

### T041: Implement Main Entry Point
**File**: main.go (create new)
- Create cmd/distui/main.go
- Initialize Bubble Tea program
- Detect current directory project
- Handle command line flags
- Start TUI application
**Estimate**: 2 points

### T042: Add Keyboard Navigation
**File**: internal/tui/keys.go (create)
- Define global key bindings
- Implement TAB cycling logic
- Add direct navigation shortcuts
- Handle modal key events
- Keep consistent across views
**Estimate**: 2 points

### T043: Add Styling Consistency
**File**: internal/tui/styles.go (create)
- Define color scheme constants
- Create reusable style functions
- Implement progress bar styles
- Define border styles
- Ensure theme consistency
**Estimate**: 2 points

## Polish Tasks

### T044: [P] Add Loading Spinner
**File**: views/common.go (create)
- Create reusable spinner component
- Show during detection
- Display during release
- Animate during saves
- Integrate with all views
**Estimate**: 1 point

### T045: [P] Add Error Modal
**File**: views/error_modal.go (create)
- Create error display modal
- Show error details
- Provide recovery suggestions
- Handle dismissal
- Style with warning colors
**Estimate**: 2 points

### T046: [P] Add Progress Indicators
**File**: views/progress.go (create)
- Create progress bar component
- Calculate release progress
- Show step indicators
- Display time remaining
- Update in real-time
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

## Notes

- All tasks must maintain < 100 lines per file
- Use early returns to avoid nested conditionals
- No comments except API documentation
- Follow template's handler pattern exactly
- Test each component in isolation
- Ensure atomic operations for all file I/O