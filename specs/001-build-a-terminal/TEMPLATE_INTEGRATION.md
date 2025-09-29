# Distui Template Integration Plan

## Template Structure Analysis

The existing template uses a clean separation of concerns:
- **app.go**: Main application model and routing logic
- **handlers/**: Update logic for each page (state transitions)
- **views/**: Rendering logic for each page (visual output)

Key patterns:
1. **Enum-based page state** (`pageState` type with const values)
2. **Router pattern** in Update() and View() methods
3. **Handler functions** return `(newPage int, quitting bool, cmd tea.Cmd)`
4. **View functions** return formatted strings
5. **100-line file limit** already enforced

## Distui Page Mapping

### Page State Enum (app.go)
```go
type pageState uint

const (
    projectView pageState = iota  // Replaces homePage
    globalView                    // Replaces page1
    settingsView                  // Replaces page2
    releaseView                   // Replaces page3
    configureView                 // New page4
    newProjectView                // New page5
)
```

### File Structure Mapping

```
distui-app/
├── app.go                       # Main router (modified from template)
├── internal/
│   ├── config/
│   │   ├── loader.go           # Load/save YAML configs
│   │   └── paths.go            # ~/.distui path management
│   ├── detection/
│   │   ├── project.go          # Detect Go project info
│   │   └── repository.go       # Git/GitHub detection
│   ├── executor/
│   │   ├── command.go          # Direct command execution
│   │   └── release.go          # Release workflow
│   └── models/
│       ├── project.go          # Project data structures
│       └── config.go           # Config data structures
├── handlers/
│   ├── project_handler.go      # Project view logic
│   ├── global_handler.go       # Global view logic
│   ├── settings_handler.go     # Settings view logic
│   ├── release_handler.go      # Release execution logic
│   ├── configure_handler.go    # Project config logic
│   └── newproject_handler.go   # New project setup logic
└── views/
    ├── project_view.go          # Project dashboard
    ├── global_view.go           # All projects list
    ├── settings_view.go         # Global settings
    ├── release_view.go          # Release progress
    ├── configure_view.go        # Project configuration
    └── newproject_view.go       # New project wizard

```

## Navigation Flow

### Primary Views (TAB cycle)
1. **Project View** (default when project detected)
   - Shows current project info
   - Quick stats and last release
   - Action buttons: [r]elease, [c]onfigure, [h]istory

2. **Global View** (TAB or 'g')
   - Lists all configured projects
   - Navigate with up/down
   - Actions: [Enter] select, [n]ew, [d]elete

3. **Settings View** (TAB or 's')
   - Global configuration
   - User preferences
   - Default paths

### Modal Views (context-triggered)
4. **Release View** (from Project via 'r')
   - Version selection
   - Real-time command output
   - Progress indicators
   - Returns to Project on complete/cancel

5. **Configure View** (from Project via 'c')
   - Project-specific settings
   - Distribution channels
   - Build configuration
   - Returns to Project on save/cancel

6. **New Project View** (from Global via 'n')
   - Detection wizard
   - Initial configuration
   - Returns to Project on save

## Handler Pattern Adaptation

### Template Pattern
```go
func UpdatePage1(currentPage, homePage int, msg tea.Msg) (int, bool, tea.Cmd)
```

### Distui Pattern (keeping same signature)
```go
func UpdateProjectView(currentPage, previousPage int, msg tea.Msg) (int, bool, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "r":
            return int(releaseView), false, startReleaseCmd()
        case "c":
            return int(configureView), false, loadConfigCmd()
        case "tab":
            return int(globalView), false, nil
        case "g":
            return int(globalView), false, nil
        case "s":
            return int(settingsView), false, nil
        case "q":
            return currentPage, true, tea.Quit
        }
    case projectLoadedMsg:
        // Handle project data updates
    }
    return currentPage, false, nil
}
```

## Model Extension

### Template appModel
```go
type appModel struct {
    currentPage pageState
    choice      int
    width       int
    height      int
    quitting    bool
    spinner     spinner.Model
    startTime   time.Time
    menuList    list.Model
}
```

### Distui appModel
```go
type appModel struct {
    // Template fields
    currentPage pageState
    width       int
    height      int
    quitting    bool
    spinner     spinner.Model

    // Distui-specific fields
    currentProject  *models.Project
    projects        []models.Project
    globalConfig    *models.GlobalConfig

    // View-specific state
    projectState    projectViewState
    globalState     globalViewState
    settingsState   settingsViewState
    releaseState    releaseViewState
    configureState  configureViewState
    newProjectState newProjectViewState

    // Shared UI components
    projectList     list.Model
    errorModal      *errorModalState
}
```

## View State Management

Each view maintains its own state struct:

```go
type projectViewState struct {
    selectedAction  int
    lastRelease    *models.Release
    quickStats     models.Stats
}

type globalViewState struct {
    selectedIndex  int
    sortBy        string
    filterQuery   string
}

type releaseViewState struct {
    version       string
    versionType   string
    phase         ReleasePhase
    currentStep   string
    progress      float64
    output        []string
    status        string
}
```

## Command Patterns

### Detection Commands
```go
func detectProjectCmd(path string) tea.Cmd {
    return func() tea.Msg {
        project, err := detection.DetectProject(path)
        if err != nil {
            return errorMsg{err}
        }
        return projectDetectedMsg{project}
    }
}
```

### Release Execution
```go
func executeReleaseCmd(project *models.Project, version string) tea.Cmd {
    return func() tea.Msg {
        return executor.StreamRelease(project, version)
    }
}
```

## Implementation Phases

### Phase 1: Core Structure
1. Modify app.go with new pageState enum
2. Create internal/ packages (config, detection, models)
3. Create handler stubs for all views
4. Create basic view renderers

### Phase 2: Project Detection
1. Implement detection.DetectProject()
2. Create projectDetectedMsg handling
3. Load/save project configs

### Phase 3: View Implementation
1. Project view with stats display
2. Global view with project list
3. Settings view with form fields

### Phase 4: Release Execution
1. Release handler with version selection
2. Command executor with streaming output
3. Progress tracking and error handling

### Phase 5: Configuration
1. Configure view with tabs
2. New project wizard
3. Validation and persistence

## Key Differences from Original Plan

1. **No separate TUI package** - using template structure
2. **Handlers return int** instead of tea.Model
3. **Views are functions** not methods on Model
4. **Page state is enum** not string constants
5. **Navigation via switch** not dynamic routing

## Constitution Compliance

✅ Files under 100 lines (template enforces this)
✅ No nested conditionals (using early returns)
✅ No comments except API docs
✅ Self-documenting names
✅ Errors bubble up via tea.Msg

## Next Steps

1. Update app.go with distui page states
2. Create internal/ package structure
3. Implement handlers following template pattern
4. Build views using template's renderPage wrapper
5. Add detection and execution logic