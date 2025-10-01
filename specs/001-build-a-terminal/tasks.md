# Implementation Tasks: distui Enhancements

**Feature**: Smart Commit Preferences, Repo Cleanup, Branch Selection, UI Notifications
**Branch**: `001-build-a-terminal`
**Plan**: [plan.md](./plan.md)
**Status**: Ready for Implementation

## Task Execution Guide

### Parallel Execution
Tasks marked with [P] can be executed in parallel. Run them concurrently for faster implementation:

```bash
# Example: Run 4 parallel tasks
# Terminal 1
claude-code "Execute T001"

# Terminal 2
claude-code "Execute T002"

# Terminal 3
claude-code "Execute T003"

# Terminal 4
claude-code "Execute T004"
```

### Dependencies
Tasks must be completed in order within each phase, but [P] tasks within the same phase can run concurrently.

---

## Phase 0: Infrastructure Setup (4 tasks)

### T001: Add new entity types to internal/models/types.go [P]
**File**: `internal/models/types.go`
**Description**: Add Go struct definitions for new data model entities
**Dependencies**: None
**Parallel**: Yes

Add these struct types to types.go:

```go
// Smart Commit Preferences
type SmartCommitPreferences struct {
    Enabled     bool               `yaml:"enabled"`
    CustomRules []FileCategoryRule `yaml:"custom_rules,omitempty"`
}

type FileCategoryRule struct {
    Pattern  string `yaml:"pattern"`   // "*.proto" or "**/test/**"
    Category string `yaml:"category"`  // "config", "code", "docs", etc.
    Priority int    `yaml:"priority"`  // Higher = applied first
}

// Repository Cleanup
type FlaggedFile struct {
    Path            string        `yaml:"path"`
    IssueType       string        `yaml:"issue_type"`        // "media", "excess-docs", "dev-artifact"
    SizeBytes       int64         `yaml:"size_bytes"`
    SuggestedAction string        `yaml:"suggested_action"`  // "delete", "ignore", "archive"
    FlaggedAt       time.Time     `yaml:"flagged_at"`
}

type CleanupScanResult struct {
    MediaFiles     []FlaggedFile     `yaml:"media_files"`
    ExcessDocs     []FlaggedFile     `yaml:"excess_docs"`
    DevArtifacts   []FlaggedFile     `yaml:"dev_artifacts"`
    TotalSizeBytes int64             `yaml:"total_size_bytes"`
    ScanDuration   time.Duration     `yaml:"scan_duration"`
    ScannedAt      time.Time         `yaml:"scanned_at"`
}

// Branch Selection
type BranchInfo struct {
    Name           string `yaml:"name"`             // "main", "origin/develop"
    IsCurrent      bool   `yaml:"is_current"`
    TrackingBranch string `yaml:"tracking_branch"`  // "" if no tracking
    AheadCount     int    `yaml:"ahead_count"`
    BehindCount    int    `yaml:"behind_count"`
}

type BranchSelectionModal struct {
    Branches      []BranchInfo `yaml:"branches"`
    SelectedIndex int          `yaml:"selected_index"`
    FilterQuery   string       `yaml:"filter_query"`
    Width         int          `yaml:"width"`
    Height        int          `yaml:"height"`
}

// UI Notifications
type UINotification struct {
    Message   string    `yaml:"message"`
    ShowUntil time.Time `yaml:"show_until"`
    Style     string    `yaml:"style"`  // "info", "success", "warning", "error"
}
```

Update ProjectConfig to include SmartCommitPreferences:
```go
type ProjectConfig struct {
    Project     *ProjectInfo            `yaml:"project"`
    Config      *Config                 `yaml:"config"`
    History     *History                `yaml:"history"`
    SmartCommit *SmartCommitPreferences `yaml:"smart_commit,omitempty"`  // NEW
}
```

**Acceptance**: All new types compile without errors, follow existing YAML tag patterns

---

### T002: Extend ProjectConfig YAML schema [P]
**File**: `internal/config/loader.go`
**Description**: Update LoadProject and SaveProject to handle SmartCommitPreferences field
**Dependencies**: None
**Parallel**: Yes

Ensure LoadProject unmarshals smart_commit section:
- Test with missing smart_commit (should be nil, not error)
- Test with empty smart_commit (should create empty struct)
- Test with populated custom_rules

Ensure SaveProject marshals smart_commit section:
- Omit if nil (omitempty tag)
- Preserve existing format
- Atomic file write (temp + rename)

**Acceptance**: Can load/save projects with smart_commit section without data loss

---

### T003: Create filescanner package skeleton [P]
**Directory**: `internal/filescanner/`
**Files**:
- `scanner.go` (file scanning logic)
- `categorizer.go` (issue type detection)
- `actions.go` (delete/ignore/archive operations)

**Description**: Create package structure with exported function signatures
**Dependencies**: None
**Parallel**: Yes

**scanner.go**:
```go
package filescanner

import (
    "io/fs"
    "path/filepath"
    "time"
    "distui/internal/models"
)

// ScanRepository walks directory and flags problematic files
func ScanRepository(root string) (*models.CleanupScanResult, error) {
    // TODO: Implement in T009
    return nil, nil
}
```

**categorizer.go**:
```go
package filescanner

import "io/fs"

// CategorizeFile determines issue type for a file
func CategorizeFile(path string, entry fs.DirEntry) (issueType string, shouldFlag bool) {
    // TODO: Implement in T009
    return "", false
}
```

**actions.go**:
```go
package filescanner

// DeleteFile removes file with confirmation
func DeleteFile(path string) error {
    // TODO: Implement in T010
    return nil
}

// AddToGitignore appends path to .gitignore
func AddToGitignore(path string) error {
    // TODO: Implement in T010
    return nil
}

// ArchiveFile moves file to .distui-archive/
func ArchiveFile(path string) error {
    // TODO: Implement in T010
    return nil
}
```

**Acceptance**: Package compiles, functions return placeholder values

---

### T004: Create gitops package skeleton [P]
**Directory**: `internal/gitops/`
**File**: `branches.go`
**Description**: Create package for git branch operations
**Dependencies**: None
**Parallel**: Yes

```go
package gitops

import (
    "distui/internal/models"
)

// ListBranches returns all local branches with tracking info
func ListBranches() ([]models.BranchInfo, error) {
    // TODO: Implement in T011
    return nil, nil
}

// GetCurrentBranch returns name of current branch
func GetCurrentBranch() (string, error) {
    // TODO: Implement in T011
    return "", nil
}

// PushToBranch pushes HEAD to specified branch
func PushToBranch(branch string) error {
    // TODO: Implement in T012
    return nil
}
```

**Acceptance**: Package compiles, functions return placeholder values

---

## Phase 1: Contract Tests (4 tasks - ALL PARALLEL)

### T005: Write smart_commit_preferences_contract_test.go [P]
**File**: `tests/contract/smart_commit_preferences_test.go`
**Description**: Contract tests for preferences loading, saving, validation
**Dependencies**: T001, T002
**Parallel**: Yes

Test cases (all should FAIL initially):
```go
func TestLoadSmartCommitPreferences_DefaultsWhenNone(t *testing.T) {
    // GIVEN: Project with no smart_commit section in YAML
    // WHEN: LoadProject is called
    // THEN: ProjectConfig.SmartCommit should be nil (not error)
}

func TestSaveCustomRule_ValidatesPattern(t *testing.T) {
    // GIVEN: FileCategoryRule with invalid glob pattern
    // WHEN: Validating pattern
    // THEN: Should return validation error
}

func TestDeleteCustomRule_RevertsToDefaults(t *testing.T) {
    // GIVEN: Files categorized using custom rule
    // WHEN: Custom rule is deleted
    // THEN: Files should immediately use default categorization
}

func TestToggleCustomMode_CleansYAML(t *testing.T) {
    // GIVEN: Project with custom_rules enabled
    // WHEN: Toggling use_custom_rules to false
    // THEN: custom_rules should be removed from YAML file
}

func TestApplyRules_PriorityOrder(t *testing.T) {
    // GIVEN: Multiple rules matching same file
    // WHEN: Applying categorization
    // THEN: Rule with highest priority should win
}
```

**Acceptance**: All tests compile, all FAIL with clear error messages

---

### T006: Write repo_cleanup_contract_test.go [P]
**File**: `tests/contract/repo_cleanup_test.go`
**Description**: Contract tests for file scanning and cleanup actions
**Dependencies**: T003
**Parallel**: Yes

Test cases (all should FAIL initially):
```go
func TestScanRepository_FlagsMediaFiles(t *testing.T) {
    // GIVEN: Directory with .mp4, .mov, .wav files
    // WHEN: ScanRepository is called
    // THEN: Files flagged as "media" issue type
}

func TestScanRepository_FlagsExcessDocs(t *testing.T) {
    // GIVEN: Directory with multiple .md files (not README)
    // WHEN: ScanRepository is called
    // THEN: Files flagged as "excess-docs" issue type
}

func TestScanRepository_FlagsDevArtifacts(t *testing.T) {
    // GIVEN: Directory with .DS_Store, .log files
    // WHEN: ScanRepository is called
    // THEN: Files flagged as "dev-artifact" issue type
}

func TestScanRepository_SkipsGitDirectory(t *testing.T) {
    // GIVEN: Repository with .git/ directory
    // WHEN: ScanRepository is called
    // THEN: .git/ contents should NOT be flagged
}

func TestArchiveFile_PreservesStructure(t *testing.T) {
    // GIVEN: File at path/to/file.txt
    // WHEN: ArchiveFile is called
    // THEN: File moved to .distui-archive/TIMESTAMP/path/to/file.txt
}

func TestAddToGitignore_AppendsLine(t *testing.T) {
    // GIVEN: Existing .gitignore file
    // WHEN: AddToGitignore("*.log") is called
    // THEN: "*.log" appended to .gitignore
}
```

**Acceptance**: All tests compile, all FAIL with clear error messages

---

### T007: Write branch_selection_contract_test.go [P]
**File**: `tests/contract/branch_selection_test.go`
**Description**: Contract tests for git branch operations
**Dependencies**: T004
**Parallel**: Yes

Test cases (all should FAIL initially):
```go
func TestListBranches_ParsesTrackingInfo(t *testing.T) {
    // GIVEN: Repository with branches having tracking info
    // WHEN: ListBranches is called
    // THEN: BranchInfo includes tracking_branch, ahead/behind counts
}

func TestListBranches_IdentifiesCurrent(t *testing.T) {
    // GIVEN: Repository on 'main' branch
    // WHEN: ListBranches is called
    // THEN: BranchInfo for 'main' has is_current = true
}

func TestGetCurrentBranch_ReturnsActiveBranch(t *testing.T) {
    // GIVEN: Repository on any branch
    // WHEN: GetCurrentBranch is called
    // THEN: Returns current branch name
}

func TestPushToBranch_SucceedsForValid(t *testing.T) {
    // GIVEN: Repository with unpushed commits
    // WHEN: PushToBranch("main") is called
    // THEN: Commits pushed to origin/main successfully
}

func TestPushToBranch_FailsForInvalid(t *testing.T) {
    // GIVEN: Repository with invalid remote
    // WHEN: PushToBranch("nonexistent") is called
    // THEN: Returns error with clear message
}
```

**Acceptance**: All tests compile, all FAIL with clear error messages

---

### T008: Write ui_notifications_contract_test.go [P]
**File**: `tests/contract/ui_notifications_test.go`
**Description**: Contract tests for notification timer behavior
**Dependencies**: T001
**Parallel**: Yes

Test cases (all should FAIL initially):
```go
func TestNotification_AutoDismissesAfter1500ms(t *testing.T) {
    // GIVEN: UINotification with ShowUntil = now + 1.5s
    // WHEN: 1.5 seconds elapse
    // THEN: Notification should be cleared automatically
}

func TestNotification_PersistsBeforeTimeout(t *testing.T) {
    // GIVEN: UINotification with ShowUntil = now + 1.5s
    // WHEN: 1.0 seconds elapse
    // THEN: Notification should still be visible
}

func TestNotification_ManualDismiss(t *testing.T) {
    // GIVEN: Active UINotification
    // WHEN: User dismisses manually
    // THEN: Notification cleared immediately, timer stopped
}

func TestNotification_StyleRendering(t *testing.T) {
    // GIVEN: UINotification with style "success"
    // WHEN: Rendering notification
    // THEN: Uses success color scheme (green)
}
```

**Acceptance**: All tests compile, all FAIL with clear error messages

---

## Phase 2: Core Logic Implementation (5 tasks)

### T009: Implement filescanner.Scanner and Categorizer
**Files**:
- `internal/filescanner/scanner.go`
- `internal/filescanner/categorizer.go`

**Description**: Implement repository file scanning and categorization using hybrid approach
**Dependencies**: T003, T006
**Parallel**: No (sequential after tests)
**New Dependency**: `github.com/muesli/gitcha` for git-aware file scanning

**Implementation Strategy**:
Use **gitcha** for tracked problematic files (respects .gitignore) + **filepath.WalkDir** for untracked files that should be ignored.

**Implementation Requirements**:

1. **Add gitcha dependency**:
```bash
go get github.com/muesli/gitcha
```

2. **ScanRepository function - Hybrid Approach**:

```go
package filescanner

import (
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "strings"
    "time"

    "distui/internal/models"
    "github.com/muesli/gitcha"
)

func ScanRepository(root string) (*models.CleanupScanResult, error) {
    startTime := time.Now()
    result := &models.CleanupScanResult{
        MediaFiles:   []models.FlaggedFile{},
        ExcessDocs:   []models.FlaggedFile{},
        DevArtifacts: []models.FlaggedFile{},
        ScannedAt:    startTime,
    }

    // PHASE 1: Use gitcha to find TRACKED problematic files
    // These files are in git but shouldn't be (suggest delete/archive)
    repo, err := gitcha.GitRepoForPath(root)
    if err == nil {
        // Find tracked media files (videos, audio, large images)
        if err := scanTrackedMedia(root, result); err != nil {
            return nil, fmt.Errorf("scanning tracked media: %w", err)
        }

        // Find tracked excess documentation
        if err := scanTrackedDocs(root, result); err != nil {
            return nil, fmt.Errorf("scanning tracked docs: %w", err)
        }
    }

    // PHASE 2: Use filepath.WalkDir to find UNTRACKED problematic files
    // These files should be in .gitignore but aren't (suggest ignore)
    if err := scanUntrackedArtifacts(root, result); err != nil {
        return nil, fmt.Errorf("scanning untracked artifacts: %w", err)
    }

    // Calculate totals
    result.ScanDuration = time.Since(startTime)
    result.TotalSizeBytes = calculateTotalSize(result)

    return result, nil
}

// scanTrackedMedia finds media files that are tracked in git
func scanTrackedMedia(root string, result *models.CleanupScanResult) error {
    mediaPatterns := []string{
        "*.mp4", "*.mov", "*.avi", "*.mkv", "*.flv", "*.wmv", // video
        "*.wav", "*.mp3", "*.flac", "*.aac", "*.ogg",         // audio
        "*.jpg", "*.jpeg", "*.png", "*.gif", "*.bmp", "*.svg", // images
    }

    ch, err := gitcha.FindFiles(root, mediaPatterns)
    if err != nil {
        return err
    }

    for file := range ch {
        // Skip common icon/logo files
        basename := filepath.Base(file.Path)
        if isIconOrLogo(basename) {
            continue
        }

        info, _ := os.Stat(file.Path)
        result.MediaFiles = append(result.MediaFiles, models.FlaggedFile{
            Path:            file.Path,
            IssueType:       "media",
            SizeBytes:       info.Size(),
            SuggestedAction: "delete",
            FlaggedAt:       time.Now(),
        })
    }

    return nil
}

// scanTrackedDocs finds excess documentation files tracked in git
func scanTrackedDocs(root string, result *models.CleanupScanResult) error {
    // Find all markdown files, excluding README
    ch, err := gitcha.FindFilesExcept(root,
        []string{"*.md", "*.markdown"},
        []string{"README.md", "README.markdown", "readme.md"},
    )
    if err != nil {
        return err
    }

    for file := range ch {
        info, _ := os.Stat(file.Path)
        result.ExcessDocs = append(result.ExcessDocs, models.FlaggedFile{
            Path:            file.Path,
            IssueType:       "excess-docs",
            SizeBytes:       info.Size(),
            SuggestedAction: "archive",
            FlaggedAt:       time.Now(),
        })
    }

    // Find other document types
    docPatterns := []string{"*.pdf", "*.doc", "*.docx", "*.ppt", "*.pptx"}
    ch, err = gitcha.FindFiles(root, docPatterns)
    if err != nil {
        return err
    }

    for file := range ch {
        info, _ := os.Stat(file.Path)
        result.ExcessDocs = append(result.ExcessDocs, models.FlaggedFile{
            Path:            file.Path,
            IssueType:       "excess-docs",
            SizeBytes:       info.Size(),
            SuggestedAction: "archive",
            FlaggedAt:       time.Now(),
        })
    }

    return nil
}

// scanUntrackedArtifacts finds untracked files that should be ignored
func scanUntrackedArtifacts(root string, result *models.CleanupScanResult) error {
    return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        // Skip .git and .distui-archive
        if d.IsDir() {
            name := d.Name()
            if name == ".git" || name == ".distui-archive" {
                return filepath.SkipDir
            }
            return nil
        }

        // Check if file is a dev artifact
        basename := d.Name()
        ext := filepath.Ext(basename)

        // System files
        if basename == ".DS_Store" || basename == "Thumbs.db" || basename == "desktop.ini" {
            info, _ := d.Info()
            result.DevArtifacts = append(result.DevArtifacts, models.FlaggedFile{
                Path:            path,
                IssueType:       "dev-artifact",
                SizeBytes:       info.Size(),
                SuggestedAction: "ignore",
                FlaggedAt:       time.Now(),
            })
            return nil
        }

        // Temp/log files
        if ext == ".log" || ext == ".tmp" || ext == ".temp" ||
           ext == ".swp" || ext == ".swo" {
            info, _ := d.Info()
            result.DevArtifacts = append(result.DevArtifacts, models.FlaggedFile{
                Path:            path,
                IssueType:       "dev-artifact",
                SizeBytes:       info.Size(),
                SuggestedAction: "ignore",
                FlaggedAt:       time.Now(),
            })
        }

        return nil
    })
}

func isIconOrLogo(filename string) bool {
    lower := strings.ToLower(filename)
    return strings.Contains(lower, "icon") ||
           strings.Contains(lower, "logo") ||
           lower == "favicon.ico"
}

func calculateTotalSize(result *models.CleanupScanResult) int64 {
    var total int64
    for _, f := range result.MediaFiles {
        total += f.SizeBytes
    }
    for _, f := range result.ExcessDocs {
        total += f.SizeBytes
    }
    for _, f := range result.DevArtifacts {
        total += f.SizeBytes
    }
    return total
}
```

3. **Key Benefits of Hybrid Approach**:
   - ✅ **gitcha respects .gitignore**: Won't flag files already properly ignored
   - ✅ **Distinguishes tracked vs untracked**: Different suggested actions
   - ✅ **Tracked problematic files** → suggest delete/archive (remove from git)
   - ✅ **Untracked problematic files** → suggest ignore (add to .gitignore)
   - ✅ **Simpler logic**: No need to manually parse .gitignore patterns

4. **Performance targets**:
   - < 2 seconds for repos with <10k files
   - gitcha uses efficient git commands internally
   - Early return on errors
   - No nested conditionals (switch on extension where needed)

**Reference**:
- gitcha library: https://github.com/muesli/gitcha
- Research findings in research.md (filepath.WalkDir pattern)
- Plan.md Phase 0 section 2 (file scanning best practices)

**Acceptance**:
- T006 contract tests pass
- Scan completes in <2s for typical repo
- Correctly distinguishes tracked vs untracked files
- Respects .gitignore automatically via gitcha

---

### T010: Implement filescanner.Actions
**File**: `internal/filescanner/actions.go`
**Description**: Implement delete/ignore/archive file operations
**Dependencies**: T003, T006
**Parallel**: No (after T009)

**Implementation Requirements**:

1. **DeleteFile function**:
   - Use os.Remove
   - Return error if file doesn't exist
   - Atomic operation (no confirmation needed - confirmation in UI layer)

2. **AddToGitignore function**:
   - Check if line already exists (avoid duplicates)
   - Append to .gitignore with newline
   - Create .gitignore if doesn't exist
   - Use atomic write pattern (temp file + rename)

3. **ArchiveFile function**:
   - Archive directory: .distui-archive/YYYY-MM-DD-HHMMSS/
   - Preserve directory structure (e.g., docs/file.pdf → .distui-archive/2025-09-30-143021/docs/file.pdf)
   - Create archive directory if not exists (os.MkdirAll)
   - Use os.Rename for move (atomic on same filesystem)
   - Fallback to copy+delete if cross-filesystem

**Reference**:
- Research findings in research.md section 4 (archive directory convention)
- Plan.md Phase 1 contract definitions

**Acceptance**: T006 contract tests pass, archived files preserve structure

---

### T011: Implement gitops.ListBranches and GetCurrentBranch
**File**: `internal/gitops/branches.go`
**Description**: Parse git branch output into structured data
**Dependencies**: T004, T007
**Parallel**: No

**Implementation Requirements**:

1. **ListBranches function**:
   - Execute: `git for-each-ref --format='%(refname:short)|%(upstream:short)|%(HEAD)' refs/heads refs/remotes`
   - Parse output into BranchInfo structs
   - Split by "|" delimiter
   - Filter out remote refs (only show local branches)
   - Set IsCurrent based on %(HEAD) = "*"
   - Set TrackingBranch from %(upstream:short)
   - AheadCount/BehindCount = 0 (leave for future enhancement)

2. **GetCurrentBranch function**:
   - Execute: `git branch --show-current`
   - Return trimmed output
   - Return error if not in git repo

**Reference**:
- Research findings in research.md section 5 (git branch commands)
- Plan.md Phase 0 outlines git for-each-ref pattern

**Acceptance**: T007 contract tests pass, branches parsed correctly

---

### T012: Implement gitops.PushToBranch
**File**: `internal/gitops/branches.go`
**Description**: Execute git push to specified branch
**Dependencies**: T004, T007
**Parallel**: No (after T011)

**Implementation Requirements**:

1. **PushToBranch function**:
   - Execute: `git push origin HEAD:refs/heads/{branch}`
   - Stream output to capture progress
   - Return error with output if push fails
   - Handle "no upstream" gracefully
   - Handle authentication failures with clear message

2. **Error handling**:
   - Wrap errors with context (fmt.Errorf)
   - Preserve git error messages
   - Detect common failures (auth, network, force push required)

**Reference**:
- Constitution principle VI (direct command execution)
- Existing push logic in handlers/cleanup_handler.go

**Acceptance**: T007 contract tests pass, successful push verified

---

### T013: Implement preferences loader/saver in config package
**File**: `internal/config/loader.go`
**Description**: Add functions for smart commit preferences CRUD
**Dependencies**: T001, T002, T005
**Parallel**: No

**Implementation Requirements**:

1. **LoadSmartCommitPreferences function**:
   - Load from ProjectConfig.SmartCommit
   - Return defaults if nil (empty CustomRules, Enabled=false)
   - Never error on missing section

2. **SaveSmartCommitPreferences function**:
   - Update ProjectConfig.SmartCommit
   - Call SaveProject with atomic write
   - Validate rules before saving (use doublestar.ValidatePattern)

3. **DeleteCustomRule function**:
   - Remove rule from CustomRules by index
   - Save updated ProjectConfig
   - Return error if index out of bounds

4. **ToggleCustomMode function**:
   - Set Enabled boolean
   - If disabling: clear CustomRules array (clean YAML)
   - Save updated ProjectConfig

**Reference**:
- Research findings in research.md section 3 (glob validation)
- Data model in data-model.md (smart_commit schema)

**Acceptance**: T005 contract tests pass, preferences persist correctly

---

## Phase 3: Handlers (4 tasks - SOME PARALLEL)

### T014: ✅ Implement smart_commit_prefs_handler.go [P] - COMPLETED
**File**: `handlers/smart_commit_prefs_handler.go`
**Description**: Handler for smart commit preferences editor UI
**Dependencies**: T013
**Parallel**: Yes (separate file)
**Status**: ✅ Already implemented - Feature complete, minor bug fix needed in T033

**Implementation Requirements**:

1. **SmartCommitPrefsModel struct**:
   - Categories list (config, code, docs, build, test, assets, data)
   - SelectedCategory int
   - EditingExtension bool
   - ExtensionInput textinput.Model
   - Rules []FileCategoryRule
   - Width, Height int
   - ProjectConfig *models.ProjectConfig

2. **NewSmartCommitPrefsModel function**:
   - Initialize with project config
   - Load existing preferences or defaults
   - Create text input for editing

3. **Update function**:
   - Handle category navigation (up/down arrows)
   - Handle extension editing (enter, esc)
   - Handle add/delete operations
   - Handle save (validate, call SaveSmartCommitPreferences)
   - Handle reset to defaults

4. **Key bindings**:
   - [↑/↓] Navigate categories
   - [→] Edit selected category
   - [a] Add extension/pattern
   - [d] Delete extension/pattern
   - [r] Reset to defaults
   - [s] Save preferences
   - [Esc] Cancel

**File size target**: <100 lines (justified if 100-120 for essential CRUD logic)

**Reference**:
- Existing handler patterns in handlers/cleanup_handler.go
- Bubble Tea input handling examples

**Acceptance**: Compiles, key bindings work, saves to YAML

---

### T015: Implement repo_cleanup_handler.go [P]
**File**: `handlers/repo_cleanup_handler.go`
**Description**: Handler for repository cleanup mode UI state
**Dependencies**: T009, T010
**Parallel**: Yes (separate file)

**Implementation Requirements**:

1. **RepoCleanupModel struct**:
   - ScanResult *models.CleanupScanResult
   - FlaggedFiles []models.FlaggedFile (flattened view)
   - SelectedIndex int
   - Scanning bool
   - ScanSpinner spinner.Model
   - Width, Height int

2. **NewRepoCleanupModel function**:
   - Initialize spinner
   - Start scan asynchronously (return tea.Cmd)

3. **Update function**:
   - Handle scan complete message
   - Handle file navigation (up/down)
   - Handle action keys (delete, ignore, archive)
   - Handle confirmation modals
   - Update spinner while scanning

4. **Commands**:
   - ScanRepositoryCmd: Async scan, returns scanCompleteMsg
   - DeleteFileCmd: Execute delete, return resultMsg
   - ArchiveFileCmd: Execute archive, return resultMsg
   - AddToGitignoreCmd: Execute ignore, return resultMsg

5. **Key bindings**:
   - [↑/↓] Navigate files
   - [d] Delete file (show confirmation)
   - [i] Add to .gitignore
   - [a] Archive file
   - [r] Re-scan
   - [Esc] Cancel/back

**File size target**: <100 lines

**Reference**:
- Async patterns in handlers/configure_handler.go (LoadCleanupCmd)
- Spinner usage in handlers/global_handler.go

**Acceptance**: Compiles, scan works, actions execute correctly

---

### T016: Implement branch_selection_handler.go [P]
**File**: `handlers/branch_selection_handler.go`
**Description**: Handler for branch selection modal state
**Dependencies**: T011, T012
**Parallel**: Yes (separate file)

**Implementation Requirements**:

1. **BranchSelectionModel struct**:
   - Branches []models.BranchInfo
   - SelectedIndex int
   - Loading bool
   - LoadSpinner spinner.Model
   - Error string
   - Width, Height int

2. **NewBranchSelectionModel function**:
   - Start loading branches (return tea.Cmd)
   - Initialize spinner
   - Calculate content dimensions (handle chrome)

3. **Update function**:
   - Handle branches loaded message
   - Handle navigation (up/down arrows)
   - Handle selection (enter key)
   - Handle cancel (esc key)
   - Update spinner while loading

4. **Commands**:
   - LoadBranchesCmd: Async load, returns branchesLoadedMsg
   - PushToBranchCmd: Execute push, returns pushResultMsg

5. **Key bindings**:
   - [↑/↓] Navigate branches
   - [Enter] Push to selected branch
   - [Esc] Cancel

**Height management**:
- Chrome: header(1) + blank(1) + instructions(1) + controls(1) = 4 lines
- ListHeight = Height - 4
- Pass ListHeight to view, NOT Height

**File size target**: <100 lines

**Reference**:
- Modal patterns in contracts/ui-states.md (BranchSelectionModal section)
- TUI Layout Integrity principle in constitution

**Acceptance**: Compiles, branch list loads, push executes

---

### T017: Implement notification_handler.go [P]
**File**: `handlers/notification_handler.go`
**Description**: Handler for auto-dismissing notification overlay
**Dependencies**: T001
**Parallel**: Yes (separate file)

**Implementation Requirements**:

1. **NotificationModel struct**:
   - Notification *models.UINotification
   - Ticking bool

2. **ShowNotification function**:
   - Create UINotification with ShowUntil = now + 1.5s
   - Return notification and tickCmd

3. **Update function**:
   - Handle tickMsg
   - Check if time.Now() > ShowUntil
   - If expired: clear notification, return nil cmd
   - If active: return tickCmd to continue

4. **tickCmd function**:
   - Return tea.Tick(100ms, ...) wrapped in tickMsg

5. **Helper functions**:
   - CreateNotification(message string, style string) (*models.UINotification, tea.Cmd)
   - DismissNotification(model *NotificationModel)

**File size target**: <80 lines (very simple)

**Reference**:
- Research findings in research.md section 6 (timer patterns)
- Plan.md Phase 1 contract definitions

**Acceptance**: Compiles, notification dismisses after 1.5s

---

## Phase 4: Views (4 tasks - ALL PARALLEL)

### T018: ✅ Implement smart_commit_prefs_view.go [P] - COMPLETED
**File**: `views/smart_commit_prefs_view.go`
**Description**: Render smart commit preferences editor UI
**Dependencies**: T014
**Parallel**: Yes
**Status**: ✅ Already implemented - Feature complete

**Implementation Requirements**:

1. **RenderSmartCommitPrefs function**:
   - Takes SmartCommitPrefsModel as input
   - Returns formatted string
   - Header: "SMART COMMIT PREFERENCES"
   - Left panel: Category list with selection indicator
   - Right panel: Extensions and patterns for selected category
   - Bottom: Control hints ([s] Save, [r] Reset, [Esc] Cancel)

2. **Visual layout**:
   ```
   ┌─ SMART COMMIT PREFERENCES ─────────────────┐
   │                                             │
   │  Categories        Selected: code          │
   │  > config          Extensions:             │
   │    code              *.go                   │
   │    docs              *.proto                │
   │    build           Patterns:               │
   │    test              **/cmd/**              │
   │    assets                                   │
   │    data            [a] Add  [d] Delete     │
   │                                             │
   │  [s] Save  [r] Reset  [Esc] Cancel         │
   └─────────────────────────────────────────────┘
   ```

3. **Styling**:
   - Use lipgloss for colors
   - Selected category: bold + color
   - Editing mode: highlight input field

4. **Height management**:
   - Use model.Height directly (handler already subtracted chrome)
   - Do NOT subtract additional chrome in view

**File size target**: <80 lines

**Reference**:
- Existing view patterns in views/cleanup_view.go
- TUI Layout Integrity in constitution

**Acceptance**: Renders correctly, styles apply, no overflow

---

### T019: Implement repo_cleanup_view.go [P]
**File**: `views/repo_cleanup_view.go`
**Description**: Render repository cleanup scan results
**Dependencies**: T015
**Parallel**: Yes

**Implementation Requirements**:

1. **RenderRepoCleanup function**:
   - Takes RepoCleanupModel as input
   - Returns formatted string
   - Show spinner if scanning
   - Show flagged files grouped by issue type
   - Highlight selected file
   - Show file details (size, suggested action)

2. **Visual layout**:
   ```
   ┌─ REPOSITORY CLEANUP ────────────────────────┐
   │                                             │
   │  Media Files (3 files, 45.2 MB)            │
   │  > video/demo.mp4 (25.1 MB) - Delete       │
   │    images/large.png (15.0 MB) - Delete     │
   │    audio/song.mp3 (5.1 MB) - Delete        │
   │                                             │
   │  Excess Docs (2 files, 1.2 MB)             │
   │    docs/old-spec.md (800 KB) - Archive     │
   │    guide.pdf (400 KB) - Archive            │
   │                                             │
   │  Dev Artifacts (5 files, 2.3 MB)           │
   │    .DS_Store (6 KB) - Delete               │
   │    debug.log (1.5 MB) - Ignore             │
   │    ...                                      │
   │                                             │
   │  [d] Delete  [i] Ignore  [a] Archive       │
   │  [r] Re-scan  [Esc] Cancel                 │
   └─────────────────────────────────────────────┘
   ```

3. **Styling**:
   - Issue types color-coded (media=red, docs=yellow, artifacts=gray)
   - Selected file: bold + background highlight
   - Size formatting: humanize (KB, MB, GB)

4. **Height management**:
   - Use model.Height directly
   - Scrollable list if more files than fit

**File size target**: <80 lines

**Reference**:
- List rendering in views/cleanup_view.go
- Grouped display patterns

**Acceptance**: Renders correctly, groups visible, no overflow, safely confirms with user before deteling files

---

### T020: Implement branch_selection_view.go [P]
**File**: `views/branch_selection_view.go`
**Description**: Render full-screen branch selection modal
**Dependencies**: T016
**Parallel**: Yes

**Implementation Requirements**:

1. **RenderBranchSelection function**:
   - Takes BranchSelectionModel as input
   - Returns formatted string
   - Full-screen overlay (uses entire terminal)
   - Header: "SELECT BRANCH TO PUSH"
   - Branch list with tracking info
   - Highlight current branch
   - Show selection indicator
   - Bottom: Control hints

2. **Visual layout**:
   ```
   ┌─ SELECT BRANCH TO PUSH ─────────────────────┐
   │                                              │
   │  > main (current) ← origin/main             │
   │    develop → origin/develop (ahead 2)       │
   │    feature/new-api (no tracking)            │
   │    feature/bugfix → origin/feature/bugfix   │
   │                                              │
   │  ↑/↓: navigate • enter: push • esc: cancel  │
   └──────────────────────────────────────────────┘
   ```

3. **Branch display format**:
   - Current branch: "(current)" label
   - Tracking: "→ remote/branch" or "← remote/branch" with ahead/behind
   - No tracking: "(no tracking)" label
   - Selected: "> " prefix + bold

4. **Height management**:
   - Use model.ListHeight (already calculated in handler)
   - Scrollable if more branches than fit

**File size target**: <80 lines

**Reference**:
- Modal overlay in contracts/ui-states.md
- Research findings in research.md section 7

**Acceptance**: Renders correctly, full-screen modal, no overflow

---

### T021: Implement notification_view.go [P]
**File**: `views/notification_view.go`
**Description**: Render auto-dismissing notification overlay
**Dependencies**: T017
**Parallel**: Yes

**Implementation Requirements**:

1. **RenderNotification function**:
   - Takes UINotification as input
   - Returns formatted string (single line)
   - Style based on notification.Style
   - Positioned at top of screen (overlay)

2. **Visual layout**:
   ```
   ┌─────────────────────────────────────────┐
   │ ✓ Switched to: /Users/me/project       │  <- Success notification
   └─────────────────────────────────────────┘
   ```

3. **Styling**:
   - info: blue background, white text
   - success: green background, white text, checkmark icon
   - warning: yellow background, black text, warning icon
   - error: red background, white text, X icon

4. **Positioning**:
   - Top-center of screen
   - Auto-width based on message length
   - Max width: 60 characters (truncate longer messages)

**File size target**: <60 lines (very simple)

**Reference**:
- Notification overlay concept in plan.md
- Lipgloss positioning (Place function)

**Acceptance**: Renders correctly, styled properly, positioned correctly

---

## Phase 5: Integration (5 tasks)

### T022: ✅ Wire smart commit prefs into configure view - COMPLETED
**Files**:
- `handlers/configure_handler.go` (update)
- `app.go` (update)

**Description**: Add smart commit preferences as sub-view accessible from configure view
**Dependencies**: T014, T018
**Parallel**: No
**Status**: ✅ Already integrated - Feature accessible from configure view

**Implementation Requirements**:

1. **Update ConfigureModel**:
   - Add PrefsModel *SmartCommitPrefsModel field
   - Add EditingPrefs bool field

2. **Update ConfigureHandler**:
   - Add key binding: [P] Edit preferences (from Advanced tab)
   - When pressed: create PrefsModel, set EditingPrefs = true
   - Route messages to PrefsModel when EditingPrefs = true
   - On save: update ProjectConfig, set EditingPrefs = false

3. **Update ConfigureView**:
   - If EditingPrefs: render PrefsModel.View()
   - Else: render normal tab view

4. **Chrome calculation**:
   - PrefsModel chrome same as other tabs (13 lines)
   - Pass m.Height - 13 to NewSmartCommitPrefsModel

**Reference**:
- Existing tab structure in handlers/configure_handler.go
- Sub-view patterns in contracts/ui-states.md

**Acceptance**: [P] key opens preferences, saves work, ESC returns to configure

---

### T023: Wire repo cleanup into cleanup tab
**Files**:
- `handlers/configure_handler.go` (update)
- `views/cleanup_view.go` (update)

**Description**: Add repo cleanup mode toggle and UI to cleanup tab
**Dependencies**: T015, T019
**Parallel**: No

**Implementation Requirements**:

1. **Update CleanupModel**:
   - Add CleanupMode *RepoCleanupModel field
   - Add ScanningRepo bool field

2. **Update CleanupTab handling**:
   - Add key binding: [C] Scan repository (shift+c to distinguish from commit)
   - When pressed: create CleanupMode, set ScanningRepo = true
   - Route messages to CleanupMode when ScanningRepo = true
   - On action complete: refresh file list, set ScanningRepo = false

3. **Update CleanupView**:
   - If ScanningRepo: render CleanupMode.View()
   - Else: render normal git status view

4. **Chrome calculation**:
   - CleanupMode uses same listHeight as file list
   - Already calculated in handler

**Reference**:
- Nested view approach in constitution (TUI Layout Integrity)
- Cleanup tab structure in handlers/configure_handler.go

**Acceptance**: [C] key starts scan, actions work, returns to cleanup tab

---

### T024: Wire Shift+B handler into cleanup tab
**Files**:
- `handlers/cleanup_handler.go` (update)
- `app.go` (update)

**Description**: Add branch selection modal triggered by Shift+B
**Dependencies**: T016, T020
**Parallel**: No

**Implementation Requirements**:

1. **Update app.go Model**:
   - Add BranchModal *BranchSelectionModal field
   - Add ShowingBranchModal bool field

2. **Update cleanup key handling**:
   - Add case for "B" (shift+B)
   - Check if unpushed commits exist
   - If yes: create BranchModal, set ShowingBranchModal = true
   - Return to normal view after push complete or cancel

3. **Update app View**:
   - If ShowingBranchModal: render BranchModal.View() (full screen)
   - Else: render normal view

4. **Message routing**:
   - Route messages to BranchModal when ShowingBranchModal = true
   - On push complete: set ShowingBranchModal = false, refresh status

**Reference**:
- Modal overlay patterns in research.md section 7
- Full-screen replacement approach from plan

**Acceptance**: Shift+B shows modal, branch selection works, push executes

---

### T025: Wire notification overlay into app.go
**Files**:
- `app.go` (update)
- `handlers/notification_handler.go` (update)

**Description**: Add notification system for switchedToPath and other messages
**Dependencies**: T017, T021
**Parallel**: No

**Implementation Requirements**:

1. **Update app Model**:
   - Replace switchedToPath string with Notification *UINotification
   - Add NotificationModel *NotificationModel field

2. **Replace switchedToPath logic**:
   - When project switches: call ShowNotification("Switched to: " + path, "success")
   - Store notification + start timer cmd
   - Remove switchedToPath clearing on keypress (auto-dismiss handles it)

3. **Update app Update**:
   - Route tickMsg to NotificationModel
   - Clear notification when auto-dismissed

4. **Update app View**:
   - If notification active: render at top of screen (overlay)
   - Use lipgloss.Place to position at top-center

**Reference**:
- Notification timer pattern in plan.md
- Auto-dismiss implementation in handler

**Acceptance**: Project switch shows notification, auto-dismisses after 1.5s

---

### T026: Update chrome calculations for new views (TUI Layout Integrity)
**Files**:
- `handlers/smart_commit_prefs_handler.go` (verify)
- `handlers/repo_cleanup_handler.go` (verify)
- `handlers/branch_selection_handler.go` (verify)
- `handlers/notification_handler.go` (N/A - overlay)

**Description**: Verify all new handlers calculate chrome correctly, no overflow
**Dependencies**: T022, T023, T024, T025
**Parallel**: No

**Verification Checklist**:

1. **For each handler**:
   - Count chrome lines (headers, blanks, controls, dividers)
   - Verify listHeight = Height - chromeLines
   - Pass listHeight to view, not Height
   - View NEVER subtracts additional chrome

2. **Test scenarios**:
   - Small terminal (80x24): No overflow
   - Normal terminal (120x40): Content fits
   - Large terminal (200x60): Content scales

3. **Chrome calculations**:
   - SmartCommitPrefsModel: 13 lines (same as configure tabs)
   - RepoCleanupModel: 13 lines (same as cleanup tab)
   - BranchSelectionModal: 4 lines (header + blank + instructions + controls)
   - NotificationModel: 0 lines (overlay, doesn't affect layout)

**Reference**:
- Constitution TUI Layout Integrity principle
- Research findings in research.md section 8
- Existing chrome calculations in configure_handler.go

**Acceptance**: No terminal overflow in any view at any terminal size

---

## Phase 6: Integration Tests (4 tasks)

### T027: Test full preferences workflow
**File**: `tests/integration/preferences_workflow_test.go`
**Description**: End-to-end test of smart commit preferences feature
**Dependencies**: T022
**Parallel**: No

Test workflow:
1. Open configure view
2. Navigate to preferences
3. Add custom rule (*.proto → code)
4. Save preferences
5. Verify YAML file updated
6. Apply rule to test files
7. Verify categorization correct
8. Delete custom rule
9. Verify files revert to defaults

**Acceptance**: Full workflow executes without errors

---

### T028: Test full cleanup scan workflow
**File**: `tests/integration/cleanup_scan_workflow_test.go`
**Description**: End-to-end test of repository cleanup feature
**Dependencies**: T023
**Parallel**: No

Test workflow:
1. Create test repo with problematic files
2. Open cleanup tab
3. Trigger scan
4. Verify files flagged correctly
5. Delete a media file
6. Archive an excess doc
7. Add artifact to .gitignore
8. Verify actions completed
9. Re-scan and verify changes

**Acceptance**: Full workflow executes, actions work correctly

---

### T029: Test full branch push workflow
**File**: `tests/integration/branch_push_workflow_test.go`
**Description**: End-to-end test of branch selection and push
**Dependencies**: T024
**Parallel**: No

Test workflow:
1. Create test repo with unpushed commits
2. Open cleanup tab
3. Press Shift+B
4. Verify branch modal appears
5. Navigate branch list
6. Select non-current branch
7. Confirm push
8. Verify commits pushed
9. Verify modal dismissed

**Acceptance**: Full workflow executes, push succeeds

---

### T030: Test notification auto-dismiss
**File**: `tests/integration/notification_autodismiss_test.go`
**Description**: Test notification timer behavior
**Dependencies**: T025
**Parallel**: No

Test workflow:
1. Trigger notification (project switch)
2. Verify notification appears
3. Wait 1.0 seconds
4. Verify notification still visible
5. Wait 0.6 seconds (total 1.6s)
6. Verify notification dismissed
7. Verify timer stopped

**Acceptance**: Notification dismisses within 1.5s ± 100ms

---

## Phase 7: Bug Fixes (3 tasks)

### T031: Fix dot file handling in commit settings
**Files**:
- `internal/gitcleanup/categorize.go` (likely)
- Or: `handlers/cleanup_handler.go`

**Description**: Fix bug where files/dirs starting with "." cannot be modified
**Dependencies**: None (bug fix)
**Parallel**: No

**Investigation**:
1. Reproduce bug: Create .github directory, try to modify in commit settings
2. Find where dot files are filtered
3. Update filter to ONLY skip .git/ directory
4. Preserve .github, .goreleaser.yaml, etc.

**Test**:
- Create .github/workflows/test.yml
- Modify file in cleanup tab
- Verify can commit, skip, ignore

**Reference**:
- Bug noted in contracts/ui-states.md Known Issues

**Acceptance**: Dot files (except .git/) work in commit settings

---

### T032: Fix switchedToPath persistence (use notification system)
**Files**: `app.go`
**Description**: Replace manual keypress clearing with notification auto-dismiss
**Dependencies**: T025
**Parallel**: No

**Implementation**:
- Already done in T025 (notification integration)
- This task verifies fix works correctly
- Remove any remaining switchedToPath clearing logic

**Test**:
- Switch projects
- Verify notification shows
- Don't press any keys
- Wait 1.5 seconds
- Verify notification dismissed automatically

**Acceptance**: No manual keypress needed, auto-dismisses correctly

---

### T033: Fix space toggle YAML cleanup in preferences
**Files**: `handlers/smart_commit_prefs_handler.go`
**Description**: Ensure toggling off custom mode removes custom_rules from YAML
**Dependencies**: T014
**Parallel**: No

**Implementation**:
- In ToggleCustomMode function (T013):
  - When disabling: set CustomRules = nil (not empty slice)
  - This triggers omitempty YAML tag
  - Field excluded from saved YAML

**Test**:
- Enable custom rules
- Add several rules
- Save
- Toggle off custom mode
- Save
- Open YAML file
- Verify custom_rules field NOT present

**Acceptance**: Toggling off completely removes custom_rules from YAML file

---

## Summary

**Total Tasks**: 33
**Completed**: 3 (T014, T018, T022 - Smart Commit Preferences feature)
**Remaining**: 30
**Estimated Duration**:
- Phase 0-1: 2-3 days (setup + tests)
- Phase 2-3: 3-4 days (core logic + handlers)
- Phase 4: 1-2 days (views - parallel)
- Phase 5-7: 2-3 days (integration + bugs)
- **Total**: 8-12 days (reduced with 3 tasks complete)

**Parallel Opportunities**:
- Phase 0: All 4 tasks can run in parallel
- Phase 1: All 4 test tasks can run in parallel
- Phase 3: All 4 handler tasks can run in parallel (separate files)
- Phase 4: All 4 view tasks can run in parallel (separate files)

**Critical Path**:
T001 → T005 → T009 → T015 → T023 (Repo cleanup)
T001 → T007 → T011 → T016 → T024 (Branch selection)
T001 → T005 → T013 → T014 → T022 (Preferences)
T001 → T008 → T017 → T025 (Notifications)

**Key Principles**:
- TDD: Tests written before implementation
- Constitution compliance: TUI Layout Integrity, clean code, no nesting
- Atomic operations: File writes use temp + rename
- Direct execution: No script generation
- Early returns: Minimize nesting throughout

**Next Step**: Begin Phase 0 by executing T001-T004 in parallel terminals.
