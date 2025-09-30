# Current State Analysis - distui Release Workflow Implementation

**Date**: 2025-09-29
**Purpose**: Understand current state before implementing release workflow

## Architecture Overview

### ✅ CORRECT: Layered Architecture Is Solid
```
views/         -> Pure rendering (lipgloss formatting)
handlers/      -> State management + Update() logic
internal/      -> Business logic (git, detection, config, execution)
app.go         -> Router + global state
```

**This is GOOD** - no violations found.

## Current File State

### Handlers (State + Update Logic)
```
configure_handler.go     989 lines  ⚠️  Large but acceptable (complex tabs/git/spinner)
settings_handler.go      304 lines  ✓   Acceptable
repo_browser.go          209 lines  ✓   Acceptable
global_handler.go         84 lines  ✓   Good
cleanup_handler.go       107 lines  ✓   Good
commit_handler.go        110 lines  ✓   Good
github_handler.go        110 lines  ✓   Good
project_handler.go        29 lines  ✓   Simple (just navigation)
release_handler.go        28 lines  ⚠️  STUB - needs implementation
new_project_handler.go    28 lines  ⚠️  STUB - not used
```

### Views (Pure Rendering)
```
configure_view.go        389 lines  ⚠️  Large but acceptable (4 tabs + modals)
cleanup_view.go          318 lines  ⚠️  Large but acceptable (complex status display)
repo_browser_view.go     167 lines  ✓   Acceptable
settings_view.go         140 lines  ✓   Acceptable
commit_view.go           141 lines  ✓   Acceptable
project_view.go          129 lines  ✓   Acceptable
global_view.go           115 lines  ✓   Acceptable
smart_commit_view.go     111 lines  ✓   Good
github_view.go            87 lines  ✓   Good
release_view.go           70 lines  ⚠️  STUB - hardcoded data
```

### Internal (Business Logic)
```
internal/executor/release.go    88 lines  ⚠️  STUB - basic structure only
internal/gitcleanup/           All files ✓  Complete (commit, gh, repo, status, categorize)
internal/detection/            All files ✓  Complete (project detection)
internal/config/               All files ✓  Complete (YAML load/save)
internal/models/types.go      134 lines  ✓  Complete (all types defined)
```

## What's Done (Git Management ✅)

1. **Configure View with 4 Tabs** - Complete
   - Cleanup tab with file categorization
   - Distributions, Build, Advanced tabs
   - Async loading with spinner
   - Window resizing support

2. **Git Cleanup** - Complete
   - File categorization (config/code/docs/build/assets/data/other)
   - Smart commit with AI messages
   - File-by-file commit workflow
   - Repository browser for navigation
   - Push detection (with fallbacks)
   - Status display ("All synced!" indicator)

3. **GitHub Integration** - Complete
   - Create GitHub repos from TUI
   - Account/org selection
   - Visibility toggle
   - Remote detection with fallbacks

4. **Project Detection** - Complete
   - go.mod parsing (handles Go 1.24+)
   - Git remote detection
   - GitHub info via gh CLI
   - Graceful degradation without git/gh

5. **Settings & Config** - Complete
   - Interactive settings editor
   - Auto-detection of GitHub username
   - YAML persistence to ~/.distui
   - Atomic file writes

6. **Global View** - Complete
   - Project list with navigation
   - Add/delete/select functionality
   - Status indicators

## What Needs Work (Release Workflow 🔄)

### 1. Release Handler (handlers/release_handler.go)
**Current**: 28 lines, just navigation stub
**Needs**: Full state machine for release workflow

```go
// What it needs:
type ReleaseModel struct {
    Phase          ReleasePhase      // Current phase
    Packages       []Package         // Steps as "packages"
    Installing     int               // Current step
    Installed      []int             // Completed steps
    Progress       progress.Model    // Bubble Tea progress bar
    Spinner        spinner.Model     // Spinner for current step
    Output         []string          // Command output buffer
    Version        string            // Selected version
    StartTime      time.Time         // For duration tracking
    Error          error             // If failed
}

type ReleasePhase int
const (
    PhaseVersionSelect ReleasePhase = iota  // Pick version
    PhasePreFlight                          // Validate config
    PhaseTests                              // Run tests
    PhaseTag                                // Create git tag
    PhaseGoReleaser                         // Run goreleaser
    PhaseHomebrew                           // Update tap
    PhaseNPM                                // Publish npm
    PhaseComplete                           // Done!
)

type Package struct {
    Name        string
    Status      string  // "pending", "installing", "done", "failed"
    Output      []string
    StartTime   time.Time
    Duration    time.Duration
}
```

**Pattern**: Use package-manager example
- Each phase is a "package" being "installed"
- Spinner shows current phase
- Progress bar shows overall completion
- tea.Printf() prints completed steps above

### 2. Release View (views/release_view.go)
**Current**: 70 lines, hardcoded mock data
**Needs**: Dynamic rendering based on ReleaseModel state

```go
// Needs to render:
// 1. Version selection UI (when Phase == PhaseVersionSelect)
// 2. Progress display (spinner + progress bar + package list)
// 3. Real-time output streaming
// 4. Success/failure summary
// 5. Error display with recovery options
```

### 3. Release Executor (internal/executor/release.go)
**Current**: 88 lines, basic structure with runTests/buildRelease/createTag/pushTag
**Needs**: Full orchestration with streaming output

**IMPORTANT**: Keep business logic HERE, not in handlers!

```go
// What exists (good foundation):
type ReleaseExecutor struct {
    projectPath string
    config      ReleaseConfig
}

// What needs expansion:
- Add GoReleaser execution (goreleaser release --clean)
- Add Homebrew tap update (download tarball, calc SHA256, update formula)
- Add NPM publishing (npm publish)
- Add output streaming (send messages to TUI via tea.Cmd)
- Add phase tracking (send progress updates)
- Add rollback on failure
```

### 4. Command Runner (NEW: internal/executor/command.go)
**Needs**: Stream command output to TUI

```go
type CommandOutput struct {
    Line   string
    IsErr  bool
}

type CommandComplete struct {
    ExitCode int
    Error    error
}

func RunCommandStreaming(name string, args []string, dir string) tea.Cmd {
    return func() tea.Msg {
        cmd := exec.Command(name, args...)
        cmd.Dir = dir

        stdout, _ := cmd.StdoutPipe()
        stderr, _ := cmd.StderrPipe()

        // Start command
        cmd.Start()

        // Stream output via goroutines
        // Send CommandOutput messages for each line
        // Send CommandComplete when done
    }
}
```

### 5. Homebrew Detection (NEW: internal/detection/homebrew.go)
**Task**: T012 from tasks.md
**Needs**: Detect Homebrew tap location

```go
func DetectHomebrewTap(username string) (*TapInfo, error) {
    // Check common locations:
    // 1. ~/homebrew-tap
    // 2. ~/repos/homebrew-tap
    // 3. Ask gh CLI for repos matching "homebrew-*"

    return &TapInfo{
        Path:     path,
        RepoURL:  url,
        Formulas: formulaNames,
    }, nil
}
```

## Package Manager Pattern for Release View

**Reference**: `/Users/williamvansickleiii/charmtuitemplate/distui/distui-app/examples/package-manager_main.go`

**Key Concepts**:
1. **Sequential Steps**: Each phase is a "package" being installed
2. **Progress Bar**: Shows overall completion (index/total)
3. **Spinner**: Shows current step is active
4. **tea.Printf()**: Prints completed steps above the progress line
5. **installedPkgMsg**: Message type sent when each step completes

**How to adapt**:
```
"Installing chalk"           -> "Running tests"
"Installing react"           -> "Creating git tag"
"Installing typescript"      -> "Executing GoReleaser"
"Installing webpack"         -> "Updating Homebrew tap"

Progress: [=====>     ] 2/4  -> Progress: [=====>     ] 2/4
Spinner keeps spinning       -> Spinner keeps spinning
Checkmark on completion      -> Checkmark on completion
```

## Project View Issue

**Current**: Shows same UI for configured/unconfigured projects
**ui-states.md says**: Two different states needed

```
STATE 1: Unconfigured Project
- Show: "Press [c] to configure" (primary action)
- Disable: [r] Release action

STATE 2: Configured Project
- Show: Dashboard (git status, distributions, last release)
- Enable: [r] Release (primary action)
```

**Fix**: Update `views/project_view.go` to render differently based on `config == nil`

**Already partially done**: Lines 45-53 show unconfigured state, but could be more prominent.

## Implementation Order (Recommended)

### Phase 1: Foundation (Business Logic)
1. ✅ Review internal/executor/release.go structure
2. 🔄 Add internal/executor/command.go (streaming output)
3. 🔄 Add internal/detection/homebrew.go (tap detection)
4. 🔄 Expand internal/executor/release.go with GoReleaser/Homebrew/NPM

### Phase 2: TUI Integration (Handlers)
5. 🔄 Create ReleaseModel in handlers/release_handler.go
6. 🔄 Implement version selection UI
7. 🔄 Wire up package-manager pattern (progress + spinner)
8. 🔄 Handle phase transitions

### Phase 3: View Rendering (Views)
9. 🔄 Update views/release_view.go for version selection
10. 🔄 Add progress display rendering
11. 🔄 Add output streaming display
12. 🔄 Add success/error summary views

### Phase 4: Polish
13. 🔄 Update project_view.go for configured/unconfigured states
14. 🔄 Add error recovery (retry/skip/cancel)
15. 🔄 Add release history tracking
16. 🔄 Add CI/CD workflow generation option

## Architecture Violations to Avoid

### ❌ DON'T DO THIS:
```go
// In handler:
func (m ReleaseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // DON'T: Execute goreleaser directly here
    cmd := exec.Command("goreleaser", "release")
    cmd.Run()  // ❌ Business logic in handler!
}
```

### ✅ DO THIS INSTEAD:
```go
// In handler:
func (m ReleaseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if msg.String() == "enter" {
        // DO: Delegate to internal package
        return m, executeReleaseCmd(m.project, m.version)
    }
}

// In internal/executor/release.go:
func executeReleaseCmd(project *models.ProjectInfo, version string) tea.Cmd {
    return func() tea.Msg {
        executor := NewReleaseExecutor(project.Path, ReleaseConfig{
            Version: version,
        })
        result, err := executor.Execute(context.Background())
        return releaseCompleteMsg{result: result, err: err}
    }
}
```

## File Size Concerns

**Constitution v1.1.0**: 100 lines ideal, >300 lines = refactoring candidate

**Current Violations**:
- configure_handler.go: 989 lines ⚠️
- configure_view.go: 389 lines ⚠️
- cleanup_view.go: 318 lines ⚠️

**Assessment**: Acceptable per constitution
- All contain essential, non-redundant logic
- Natural cohesion (tabs/git management)
- No arbitrary splits needed
- Complex UI requires more code

**For Release Implementation**:
- Keep release_handler.go under 300 lines
- If exceeds, split by phase (version_select, execution, summary)
- Keep internal/executor files focused and small

## Summary

**Architecture**: ✅ Solid, no violations
**Git Management**: ✅ Complete and working
**Release Workflow**: 🔄 Needs full implementation
**Project View**: ⚠️ Needs configured/unconfigured states
**File Sizes**: ⚠️ Some large files, but acceptable per constitution

**Ready to Code**: YES
**Start With**: internal/executor package (business logic first)
**Then**: handlers/release_handler.go (state machine)
**Finally**: views/release_view.go (rendering)