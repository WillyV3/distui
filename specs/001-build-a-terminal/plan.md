# Implementation Plan: distui - Go Release Distribution Manager

**Branch**: `001-build-a-terminal` | **Date**: 2025-09-28 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-build-a-terminal/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → ✓ Spec loaded successfully
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → ✓ No NEEDS CLARIFICATION found
   → Detected Project Type: CLI/TUI application
   → Set Structure Decision: Single project structure
3. Fill the Constitution Check section
   → ✓ Aligned with all 10 core principles
4. Evaluate Constitution Check section
   → ✓ No violations found
   → Update Progress Tracking: Initial Constitution Check
5. Execute Phase 0 → research.md
   → ✓ Research completed
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, CLAUDE.md
   → ✓ All artifacts generated
7. Re-evaluate Constitution Check section
   → ✓ No new violations
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Task generation approach documented
9. STOP - Ready for /tasks command
```

## Summary
Build a TUI application for managing Go application releases that stores all configuration globally in ~/.distui, provides multi-view navigation (Project/Global/Settings), executes releases directly from the TUI in under 30 seconds, and supports multiple distribution channels (GitHub, Homebrew, NPM). Technical approach uses Bubble Tea framework for the TUI with direct command execution via os/exec, YAML-based configuration storage, and gh CLI for smart detection.

## Technical Context
**Language/Version**: Go 1.24+
**Primary Dependencies**: Bubble Tea (TUI framework), Lipgloss (styling), gh CLI (GitHub operations), GoReleaser
**Storage**: File-based YAML in ~/.distui (config.yaml + projects/*.yaml)
**Testing**: Go standard testing package with testify for assertions
**Target Platform**: macOS, Linux (terminal environments)
**Project Type**: single - CLI/TUI application
**Performance Goals**: < 100ms UI response, < 30s release execution, < 1s startup
**Constraints**: < 50MB memory usage, no repository pollution, terminal-only interface
**Scale/Scope**: Manage unlimited projects, typical user has 5-20 Go projects

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Initial Check (Pre-Phase 0)
- [x] **Zero Repository Pollution**: All config in ~/.distui ✓
- [x] **30-Second Release**: Direct command execution design ✓
- [x] **User Agency**: TAB navigation, no forced modes ✓
- [x] **Stateful Intelligence**: Global project memory design ✓
- [x] **Clean Go Code**: Bubble Tea + Lipgloss specified ✓
- [x] **Direct Execution**: No script generation, direct os/exec ✓
- [x] **Developer Choice**: Local + optional CI/CD support ✓
- [x] **Smart Detection**: gh CLI with overrides ✓
- [x] **No Lock-in**: YAML configs work standalone ✓
- [x] **Clean Separation**: Global vs project configs ✓
- [x] **Code Quality**: 100-line files, minimal nesting planned ✓
- [x] **Error Philosophy**: Errors bubble up, visible failures ✓

### Post-Design Check (After Phase 1)
- [x] All principles maintained in detailed design
- [x] No constitutional violations introduced
- [x] Complexity justified by user requirements

## Project Structure

### Documentation (this feature)
```
specs/001-build-a-terminal/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (completed)
├── data-model.md        # Phase 1 output (completed)
├── quickstart.md        # Phase 1 output (completed)
├── contracts/           # Phase 1 output (completed)
│   ├── config.yaml      # Configuration contract
│   ├── project.yaml     # Project config contract
│   └── ui-states.md     # UI state contracts
├── CLAUDE.md            # Claude-specific guidance (completed)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
cmd/
├── distui/
│   └── main.go          # Entry point, < 100 lines

internal/
├── config/
│   ├── loader.go        # Config loading, < 100 lines
│   ├── writer.go        # Config persistence, < 100 lines
│   └── types.go         # Config types, < 100 lines
├── detection/
│   ├── project.go       # Project detection logic
│   ├── github.go        # GitHub/gh CLI detection
│   └── homebrew.go      # Homebrew tap detection
├── executor/
│   ├── release.go       # Release orchestration
│   ├── test.go          # Test execution
│   ├── goreleaser.go    # GoReleaser execution
│   ├── homebrew.go      # Homebrew tap updates
│   └── npm.go           # NPM publishing
├── tui/
│   ├── app.go           # Main Bubble Tea app
│   ├── keys.go          # Keyboard handling
│   └── styles.go        # Lipgloss styles
├── views/
│   ├── project.go       # Project view
│   ├── global.go        # All projects view
│   └── settings.go      # Settings view
└── models/
    ├── project.go       # Project domain model
    ├── release.go       # Release domain model
    └── distribution.go  # Distribution channel model

tests/
├── integration/
│   ├── release_test.go
│   └── detection_test.go
└── unit/
    ├── config_test.go
    ├── executor_test.go
    └── views_test.go
```

## Phase 0: Research & Clarification

### Key Technical Decisions
1. **Bubble Tea for TUI**: Mature, well-supported Go TUI framework
2. **Direct os/exec**: No shell scripts, direct command execution
3. **YAML storage**: Human-readable, git-friendly configuration
4. **gh CLI dependency**: Leverage existing GitHub authentication

### Dependencies Analysis
- **charmbracelet/bubbletea**: v0.27.0 - TUI framework
- **charmbracelet/lipgloss**: v0.13.0 - Terminal styling
- **gopkg.in/yaml.v3**: v3.0.1 - YAML parsing
- **stretchr/testify**: v1.9.0 - Test assertions

### Risk Mitigation
- **gh CLI not installed**: Graceful degradation with manual input
- **Homebrew tap conflicts**: User override for all detected values
- **Concurrent releases**: File locking on project configs
- **Terminal compatibility**: ANSI escape sequence detection

## Phase 1: Design & Contracts

### Core Contracts
- **Configuration Schema**: Defined in contracts/config.yaml
- **Project Schema**: Defined in contracts/project.yaml
- **UI State Machine**: Defined in contracts/ui-states.md

### Data Model
- **Project**: Repository info, module path, version tracking
- **Configuration**: Distribution channels, build settings, preferences
- **ReleaseHistory**: Version, timestamp, method, status
- **DistributionChannel**: Type, settings, enabled state

### API Design
- **Command Execution**: Synchronous with real-time output streaming
- **Configuration API**: Load/Save with atomic writes
- **Detection API**: Best-effort with user confirmation

## Phase 2: Task Planning (To be executed by /tasks)

### Task Generation Approach
Tasks will be organized by architectural layer following clean architecture:
1. **Foundation**: Config management, types, interfaces
2. **Core Business**: Models, detection logic, executor logic
3. **TUI Layer**: Views, navigation, keyboard handling
4. **Integration**: Command execution, file I/O, gh CLI
5. **Testing**: Unit tests per module, integration tests

Each task will be:
- Self-contained (< 100 lines per file)
- Testable in isolation
- Aligned with constitutional principles
- Estimated in story points (1, 2, 3, 5)

## Progress Tracking

### Phase 0: Research & Clarification
- [x] Analyze feature specification
- [x] Identify technical stack
- [x] Document dependencies
- [x] Risk assessment
- [x] Generate research.md

### Phase 1: Design & Architecture
- [x] Define contracts
- [x] Create data model
- [x] Document quickstart guide
- [x] Generate Claude-specific guidance
- [x] Architecture validation

### Phase 2: Task Generation
- [ ] Ready for /tasks command
- [ ] Tasks will follow clean architecture
- [ ] Each task under 100 lines

### Phase 3-4: Implementation
- [ ] To be executed after /tasks
- [ ] Following generated task list

## Complexity Tracking

### Justified Complexity
None required - design aligns with all constitutional principles.

### Simplifications Made
1. File-based storage instead of database (simpler, portable)
2. Direct command execution instead of script generation
3. Single binary distribution (no plugins)
4. Terminal-only interface (no web UI)

## Next Steps
Run `/tasks` command to generate the detailed task list for implementation.