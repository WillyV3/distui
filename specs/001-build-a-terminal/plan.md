
# Implementation Plan: distui Enhancements - Smart Commit Preferences, Repo Cleanup, Branch Selection, UI Notifications

**Branch**: `001-build-a-terminal` | **Date**: 2025-09-30 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-build-a-terminal/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → ✅ Loaded from spec.md
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → ✅ All technical decisions clear from existing codebase
3. Run Constitution Check
   → ✅ PASS - All principles aligned
4. Phase 0: Generate research.md
   → ✅ Complete
5. Phase 1: Generate design artifacts
   → ✅ data-model.md, contracts/, quickstart.md generated
6. Document Phase 2 task approach
   → ✅ Task generation strategy documented
7. Update Progress Tracking
   → ✅ All gates passed
8. Write plan.md (this file)
   → ✅ Complete
9. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:
- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary

This plan implements four major enhancements to the production-ready distui TUI application:

1. **Smart Commit Preferences** - Project-level file categorization customization with full CRUD operations
2. **Repository Cleanup Mode** - Opinionated file scanner for media/docs/artifacts with delete/ignore/archive actions
3. **Branch Selection for Push** - Full-screen modal for Shift+P push operations showing all branches
4. **UI Notifications** - Auto-dismissing temporary notifications with 1.5-second timer

Technical approach follows Bubble Tea patterns, maintains TUI Layout Integrity principle with proper height management, and keeps handlers separate from views per constitution.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: Bubble Tea v0.27.0, Lipgloss v0.13.0, yaml.v3, golang.org/x/mod
**Storage**: YAML files in ~/.distui/projects/<id>/config.yaml
**Testing**: Go standard testing, table-driven tests
**Target Platform**: macOS/Linux terminal (iTerm2, Terminal.app, Alacritty)
**Project Type**: Single project (TUI application)
**Performance Goals**: TUI response <100ms, scan operations <2s for typical repos
**Constraints**: Terminal height fixed (no overflow), file operations atomic, <50MB memory
**Scale/Scope**: 4 new features, ~15 new functions, existing 6-view TUI with ~100 existing functions

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**✅ I. Pragmatic Repository Files**: No new repository files required (preferences stored in ~/.distui)
**✅ II. 30-Second Release Execution**: Not affected by these enhancements
**✅ III. User Agency and Navigation Freedom**: All features are opt-in, user-controlled
**✅ IV. Stateful Global Intelligence**: Preferences stored per-project in ~/.distui/projects/
**✅ V. Clean Go Code Excellence**: All new code follows Bubble Tea patterns and Go idioms
**✅ VI. Direct Command Execution**: Branch push uses direct git commands, file operations use os package
**✅ VII. Developer Choice Architecture**: All features optional, repo cleanup mode is toggle-able
**✅ VIII. Smart Detection with Override**: Branch selection allows manual branch choice
**✅ IX. No Vendor Lock-in**: Preferences in readable YAML, operations use standard git/file commands
**✅ X. Clean Configuration Separation**: Project preferences in project YAML, no global mixing
**✅ Self-Documenting Code**: All new code will use clear names, no comments except API docs
**✅ Structural Discipline**: New handlers <100 lines each, views <80 lines, early returns throughout
**✅ Error Philosophy**: Errors bubble up, no try/catch, failures are visible
**✅ TUI Layout Integrity**: All UI changes will update handler chrome calculations, not view layout

**Result**: ✅ PASS - All constitutional principles satisfied

## Project Structure

### Documentation (this feature)
```
specs/001-build-a-terminal/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
│   ├── smart_commit_preferences_contract.md
│   ├── repo_cleanup_contract.md
│   ├── branch_selection_contract.md
│   └── ui_notifications_contract.md
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
# Single project structure (existing)
handlers/
├── configure_handler.go         # Existing - contains CleanupModel, ConfigureModel
├── cleanup_handler.go           # Existing - cleanup operations
├── smart_commit_prefs_handler.go # NEW - preferences editor handler
├── repo_cleanup_handler.go      # NEW - file scanning and flagging handler
├── branch_selection_handler.go  # NEW - branch modal handler
└── notification_handler.go      # NEW - notification timer handler

views/
├── configure_view.go            # Existing - renders configure tabs
├── cleanup_view.go              # Existing - renders cleanup tab
├── smart_commit_prefs_view.go   # NEW - preferences editor view
├── repo_cleanup_view.go         # NEW - flagged files view
├── branch_selection_view.go     # NEW - branch modal view
└── notification_view.go         # NEW - notification overlay view

internal/
├── models/
│   └── types.go                 # MODIFY - add new entity types
├── config/
│   └── loader.go                # MODIFY - load/save preferences
├── filescanner/                 # NEW package
│   ├── scanner.go               # File scanning logic
│   ├── categorizer.go           # File categorization
│   └── actions.go               # Delete/ignore/archive operations
└── gitops/                      # NEW package
    └── branches.go              # Branch list/push operations

tests/
├── contract/
│   ├── smart_commit_prefs_test.go
│   ├── repo_cleanup_test.go
│   ├── branch_selection_test.go
│   └── notifications_test.go
├── integration/
│   ├── preferences_workflow_test.go
│   ├── cleanup_scan_test.go
│   └── branch_push_test.go
└── unit/
    ├── filescanner_test.go
    └── gitops_test.go
```

**Structure Decision**: Single project structure maintained. New features integrate into existing 6-view TUI architecture. Four new handlers and views added, two new internal packages (filescanner, gitops) for business logic separation from UI.

## Phase 0: Outline & Research

### Research Questions
1. **Smart Commit Preferences Storage**: How to extend existing ProjectConfig YAML structure for custom rules?
2. **File Scanning Performance**: Best practices for fast directory traversal in Go (filepath.Walk vs filepath.WalkDir)?
3. **Glob Pattern Validation**: How to validate user-provided glob patterns before saving?
4. **Archive Directory Convention**: Where should archived files be moved (e.g., .distui-archive/, archive/)?
5. **Branch Listing**: Best git commands for listing all local + remote branches with tracking info?
6. **Auto-dismiss Timers**: Bubble Tea patterns for time-based message dismissal with tea.Tick?
7. **Modal Overlay Patterns**: How to implement full-screen overlays that replace views temporarily?
8. **Height Management**: Proper chrome calculation patterns for new nested views?

### Research Deliverables

**Output file**: `research.md` in SPECS_DIR

**Research completed for**:
- YAML schema extension for custom preferences (bign19/yaml.v3 examples)
- Go filepath.WalkDir performance characteristics (faster than Walk, allocates less)
- Glob pattern validation using doublestar.ValidatePattern
- Archive directory convention (.distui-archive/ in repo root, create if not exists)
- Git branch commands (git branch -a --format=json, git for-each-ref)
- Bubble Tea timer patterns (tea.Tick + time.After for auto-dismiss)
- Bubble Tea modal patterns (state-based view switching, ESC to return)
- TUI Layout Integrity pattern (handler calculates chrome, view uses dimensions)

## Phase 1: Design Artifacts

### 1. Data Model Design

**Output file**: `data-model.md` in SPECS_DIR

**Entity Definitions**:

```yaml
SmartCommitPreferences:
  fields:
    - project_id: string (FK to Project)
    - custom_rules: []FileCategoryRule
    - enabled: bool

FileCategoryRule:
  fields:
    - pattern: string (extension like "*.proto" or glob like "**/ test/**")
    - category: enum (config, code, docs, build, test, assets, data, other)
    - priority: int (higher priority rules override lower)

FlaggedFile:
  fields:
    - path: string (relative to repo root)
    - issue_type: enum (media, excess-docs, dev-artifact)
    - size_bytes: int64
    - suggested_action: enum (delete, ignore, archive)
    - flagged_at: time.Time

CleanupScanResult:
  fields:
    - media_files: []FlaggedFile
    - excess_docs: []FlaggedFile
    - dev_artifacts: []FlaggedFile
    - total_size: int64
    - scan_duration: time.Duration

BranchInfo:
  fields:
    - name: string (refs/heads/main, refs/remotes/origin/develop)
    - is_current: bool
    - tracking_branch: string (empty if no tracking)
    - ahead_count: int
    - behind_count: int

BranchSelectionModal:
  fields:
    - branches: []BranchInfo
    - selected_index: int
    - filter_query: string (for future search feature)
    - width: int
    - height: int

UINotification:
  fields:
    - message: string
    - show_until: time.Time (time.Now() + 1.5s)
    - style: enum (info, success, warning, error)
```

**Relationships**:
- SmartCommitPreferences 1:1 ProjectConfig (one preferences set per project)
- FileCategoryRule N:1 SmartCommitPreferences (many rules per preference set)
- FlaggedFile N:1 CleanupScanResult (many files per scan result)
- BranchInfo N:1 BranchSelectionModal (many branches displayed in modal)

### 2. Contract Definitions

**Output directory**: `contracts/` in SPECS_DIR

**Contract Files**:

1. **smart_commit_preferences_contract.md**
   - Load preferences for project (return defaults if none exist)
   - Save custom rule (validate pattern, update YAML)
   - Delete custom rule (remove from YAML, revert files to defaults)
   - Toggle custom mode (enable/disable preferences, clean YAML on disable)
   - Apply rules to file (return category based on precedence)

2. **repo_cleanup_contract.md**
   - Scan repository (walk directory, categorize files, return FlaggedFiles)
   - Delete file (os.Remove with confirmation)
   - Add to .gitignore (append line to .gitignore)
   - Move to archive (create .distui-archive/, move file, preserve structure)

3. **branch_selection_contract.md**
   - List branches (parse git branch output, extract tracking info)
   - Push to branch (git push origin HEAD:branch-name)
   - Get current branch (git branch --show-current)

4. **ui_notifications_contract.md**
   - Show notification (create UINotification, start timer)
   - Auto-dismiss (check time.Now() vs show_until, clear if expired)
   - Manual dismiss (clear notification immediately)

### 3. Failing Contract Tests

**Output locations**: `tests/contract/*_test.go`

**Test patterns** (all initially failing):
```go
func TestLoadSmartCommitPreferences_DefaultsWhenNone(t *testing.T)
func TestSaveCustomRule_ValidatesPattern(t *testing.T)
func TestDeleteCustomRule_RevertsToDefaults(t *testing.T)
func TestScanRepository_FlagsMediaFiles(t *testing.T)
func TestListBranches_ParsesTrackingInfo(t *testing.T)
func TestPushToBranch_SucceedsForValid(t *testing.T)
func TestNotification_AutoDismissesAfter1500ms(t *testing.T)
```

### 4. Quickstart Guide

**Output file**: `quickstart.md` in SPECS_DIR

**User workflows**:
1. Edit smart commit preferences → Save custom rule → See files re-categorized
2. Enable repo cleanup mode → View flagged files → Archive excess docs
3. Stage commits → Press Shift+P → Select branch → Confirm push
4. Switch projects → See notification → Auto-dismiss after 1.5s

**Developer quickstart**:
```bash
# Run contract tests (all should fail initially)
go test ./tests/contract/... -v

# Implement preferences handler
vim handlers/smart_commit_prefs_handler.go

# Implement file scanner
vim internal/filescanner/scanner.go

# Run tests again (should pass after implementation)
go test ./tests/contract/... -v

# Run integration tests
go test ./tests/integration/... -v

# Run app manually
go run app.go
```

## Phase 2: Task Planning Approach

*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
1. Load contracts from `contracts/` directory
2. For each contract → generate failing contract test task [P]
3. For each entity in data-model.md → generate struct definition task [P]
4. For each handler → generate handler skeleton task
5. For each view → generate view rendering task
6. For each user scenario → generate integration test task
7. Generate implementation tasks to make tests pass
8. Add TUI height verification tasks for each new view

**Ordering Strategy**:
```
Phase 0: Infrastructure [P]
  - T001: Add new entity types to internal/models/types.go [P]
  - T002: Extend ProjectConfig YAML schema [P]
  - T003: Create filescanner package skeleton [P]
  - T004: Create gitops package skeleton [P]

Phase 1: Contract Tests (all parallel) [P]
  - T005: Write smart_commit_preferences_contract_test.go [P]
  - T006: Write repo_cleanup_contract_test.go [P]
  - T007: Write branch_selection_contract_test.go [P]
  - T008: Write ui_notifications_contract_test.go [P]

Phase 2: Core Logic Implementation
  - T009: Implement filescanner.Scanner (scan + categorize)
  - T010: Implement filescanner.Actions (delete/ignore/archive)
  - T011: Implement gitops.ListBranches
  - T012: Implement gitops.PushToBranch
  - T013: Implement preferences loader/saver in config package

Phase 3: Handlers (some parallel) [P]
  - T014: Implement smart_commit_prefs_handler.go [P]
  - T015: Implement repo_cleanup_handler.go [P]
  - T016: Implement branch_selection_handler.go [P]
  - T017: Implement notification_handler.go [P]

Phase 4: Views (all parallel) [P]
  - T018: Implement smart_commit_prefs_view.go [P]
  - T019: Implement repo_cleanup_view.go [P]
  - T020: Implement branch_selection_view.go [P]
  - T021: Implement notification_view.go [P]

Phase 5: Integration
  - T022: Wire smart commit prefs into configure view
  - T023: Wire repo cleanup into cleanup tab
  - T024: Wire Shift+P handler into cleanup tab
  - T025: Wire notification overlay into app.go
  - T026: Update chrome calculations for new views (TUI Layout Integrity)

Phase 6: Integration Tests
  - T027: Test full preferences workflow
  - T028: Test full cleanup scan workflow
  - T029: Test full branch push workflow
  - T030: Test notification auto-dismiss

Phase 7: Bug Fixes
  - T031: Fix dot file handling in commit settings
  - T032: Fix switchedToPath persistence (use notification system)
  - T033: Fix space toggle YAML cleanup in preferences
```

**Estimated Output**: 33 numbered, ordered tasks in tasks.md

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation

*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)
**Phase 4**: Implementation (execute tasks.md following constitutional principles)
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Design Review

### Complexity Analysis

**Single Responsibility Adherence**:
- ✅ Each handler manages one feature (preferences, cleanup, branches, notifications)
- ✅ Each view renders one UI concern
- ✅ filescanner package: single responsibility (file operations)
- ✅ gitops package: single responsibility (git operations)

**Separation of Concerns**:
- ✅ Handlers: State management and business logic orchestration
- ✅ Views: Pure rendering, no state mutations
- ✅ Internal packages: Business logic without UI coupling
- ✅ Models: Data structures only

**Nesting Levels**:
- ✅ All handlers use early returns, max 2 levels of nesting
- ✅ File scanner uses filepath.WalkDir with early returns per file
- ✅ Branch parser uses switch statements, no nested ifs

**Files Over 100 Lines** (justified by constitution):
- handlers/smart_commit_prefs_handler.go (~120 lines) - CRUD operations for rules, justified as essential non-redundant logic
- internal/filescanner/scanner.go (~110 lines) - File traversal and categorization, single cohesive purpose

### Constitutional Alignment Score: 10/10

**Principles Honored**:
1. ✅ Pragmatic Repository Files - No new repo files, only ~/.distui
2. ✅ 30-Second Release - Not affected by enhancements
3. ✅ User Agency - All features are toggleable, user-controlled
4. ✅ Stateful Global Intelligence - Preferences per-project in ~/.distui
5. ✅ Clean Go Code - Bubble Tea patterns throughout
6. ✅ Direct Command Execution - git commands direct, no scripts
7. ✅ Developer Choice - Optional features, no forced workflows
8. ✅ Smart Detection with Override - Branch selection allows override
9. ✅ No Vendor Lock-in - Preferences in readable YAML
10. ✅ Clean Configuration Separation - Project YAML only

**Code Quality Adherence**:
- ✅ Self-Documenting Code - Clear names, no explanatory comments
- ✅ Structural Discipline - Files <100 lines except justified cases
- ✅ Error Philosophy - Errors bubble up, no try/catch
- ✅ TUI Layout Integrity - Handler-level chrome calculations

### No Unnecessary Abstractions

**Justified Design Choices**:
- ✅ Separate handlers for each feature (clear boundaries, testable)
- ✅ filescanner package (reusable logic, clear responsibility)
- ✅ gitops package (encapsulates git commands, testable)
- ✅ Preferences in YAML (readable, editable, no lock-in)

**Avoided Over-Engineering**:
- ❌ No factory patterns for simple struct creation
- ❌ No dependency injection framework (direct instantiation)
- ❌ No generic abstractions for file operations
- ❌ No event bus (direct function calls)

**Deviations Justified**:

| Component | Justification |
|-----------|---------------|
| smart_commit_prefs_handler.go (120 lines) | CRUD operations for rules require essential logic for load/save/delete/validate - cohesive unit |
| filescanner/scanner.go (110 lines) | File traversal with categorization is single cohesive algorithm - splitting would break flow |
| Separate handlers per feature | Clear testability boundaries, follows existing app architecture pattern |

## Progress Tracking

*This checklist is updated during execution flow*

**Phase Status**:
- [x] Phase 0: Research complete (/plan command)
- [x] Phase 1: Design complete (/plan command)
- [x] Phase 2: Task planning complete (/plan command - describe approach only)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [x] Complexity deviations documented

---
*Based on Constitution v1.3.0 - See `/memory/constitution.md`*
