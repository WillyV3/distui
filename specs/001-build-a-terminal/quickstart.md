# Quickstart: Testing distui v0.1.0 Features

**Feature**: distui v0.1.0 enhancements
**Date**: 2025-09-30
**Purpose**: Validate new features work end-to-end

## Prerequisites

- distui v0.0.32+ installed and working
- Go project with git repository
- GitHub CLI (`gh`) authenticated
- gitcha library: `go get github.com/muesli/gitcha`

## Test Scenarios

### Feature 1: Smart Commit Preferences (Already Implemented)

#### Scenario 1.1: Custom Smart Commit Preferences

**Goal**: Verify users can customize file categorization rules

**Steps**:
1. Launch `distui` in a Go project directory
2. Press `c` to open Configure view
3. Navigate to Advanced tab (press TAB until reached)
4. Find "Smart Commit Preferences" section
5. Toggle "Use Custom Rules" to enabled
6. Select "code" category
7. Add custom extension: `.proto`
8. Add custom pattern: `**/internal/**`
9. Press `s` to save preferences
10. Return to Configure > Cleanup tab
11. Verify `.proto` files show as "code" category
12. Verify files in `internal/` show as "code" category

**Expected Outcome**:
- Custom rules toggle works
- Extensions can be added/removed
- Patterns can be added/removed
- Files categorized according to custom rules
- Changes persist after restart
- Config saved to `~/.distui/projects/<id>/config.yaml`

**Validation**:
```bash
# Check config file
cat ~/.distui/projects/github-com-user-repo.yaml | grep -A 20 "smart_commit:"

# Should show:
# smart_commit:
#   enabled: true
#   use_custom_rules: true
#   categories:
#     code:
#       extensions: [".go", ".proto"]
#       patterns: ["**/src/**", "**/internal/**"]
```

#### Scenario 1.2: Reset to Default Rules

**Goal**: Verify users can revert custom rules

**Steps**:
1. With custom rules enabled (from Scenario 1)
2. Navigate to Smart Commit Preferences
3. Press `r` to reset to defaults
4. Confirm reset action
5. Verify categories show default extensions/patterns
6. Return to Cleanup tab
7. Verify files categorized with default rules

**Expected Outcome**:
- Reset button works
- Defaults restored
- Custom rules cleared from config
- Files re-categorized immediately

#### Scenario 1.3: GitHub Workflow Generation (Opt-In)

**Goal**: Verify workflow can be generated with user consent

**Steps**:
1. Navigate to Configure > Advanced tab
2. Find "GitHub Workflow Generation" section
3. Toggle "Enable Workflow Generation" to enabled
4. Configure options:
   - Include tests: Yes
   - Distribution channels: (auto-detected from project)
5. Press `p` to preview workflow
6. Review YAML in preview modal
7. Press ESC to close preview
8. Press `g` to generate workflow
9. Confirm file creation
10. Check `.github/workflows/release.yml` was created

**Expected Outcome**:
- Toggle works
- Preview shows valid GitHub Actions YAML
- File created only after confirmation
- Workflow includes correct distribution channels
- Required secrets documented in workflow comments

**Validation**:
```bash
# Check workflow file exists
ls -la .github/workflows/release.yml

# Check workflow content
cat .github/workflows/release.yml | grep "name: Release"

# Verify includes GoReleaser step
cat .github/workflows/release.yml | grep "goreleaser"

# If NPM enabled, verify NPM publish step
cat .github/workflows/release.yml | grep "npm publish"
```

#### Scenario 1.4: Workflow Regeneration

**Goal**: Verify workflow updates when config changes

**Steps**:
1. With workflow generation enabled
2. Navigate to Distributions tab
3. Enable NPM distribution (if not already)
4. Save configuration
5. Return to Advanced tab
6. Verify "Regenerate Workflow" warning appears
7. Press `g` to regenerate
8. Check `.github/workflows/release.yml` updated

**Expected Outcome**:
- Warning shown when distribution config changes
- Regeneration preserves user comments
- New distribution steps added to workflow

#### Scenario 1.5: Disable Workflow Generation

**Goal**: Verify workflow generation can be disabled (respects user agency)

**Steps**:
1. With workflow generation enabled and file created
2. Navigate to Advanced tab
3. Toggle "Enable Workflow Generation" to disabled
4. Verify workflow file NOT deleted (user owns the file)
5. Make distribution config changes
6. Verify no regeneration warning (generation disabled)

**Expected Outcome**:
- Disabling generation doesn't delete existing file
- No auto-regeneration when disabled
- User maintains full control

#### Scenario 1.6: Dot File Bug Fix

**Goal**: Verify dot files can be committed

**Steps**:
1. Create test dot file: `touch .github/workflows/test.yml`
2. Add content: `echo "test" > .github/workflows/test.yml`
3. Launch distui, press `c` for Configure > Cleanup tab
4. Verify `.github/workflows/test.yml` appears in file list
5. Select file and choose "commit" action
6. Verify commit succeeds
7. Check git log shows commit with dot file

**Expected Outcome**:
- Dot files visible in cleanup tab
- Can select and commit dot files
- No errors or special handling needed

**Validation**:
```bash
git log -1 --name-only | grep ".github"
# Should show .github/workflows/test.yml
```

### Feature 2: Repository Cleanup Mode (Git-Aware File Scanning)

#### Scenario 2.1: Auto-Detect Repository Issues

**Goal**: Verify cleanup mode detects problematic files using git-aware scanning

**Steps**:
1. Create test repository with problematic files:
   ```bash
   touch video.mp4 audio.wav image.png
   touch .DS_Store Thumbs.db desktop.ini
   git add video.mp4 audio.wav
   ```
2. Launch distui, navigate to Configure > Cleanup tab
3. Press `f` to trigger full repository scan
4. Wait for scan completion (spinner shows progress)
5. Review scan results showing:
   - Media files (video.mp4, audio.wav) flagged as "Tracked - Delete"
   - System files (.DS_Store, etc.) flagged as "Untracked - Ignore"
   - Excess docs and dev artifacts categorized

**Expected Outcome**:
- gitcha detects tracked problematic files (video.mp4, audio.wav)
- filepath.WalkDir detects untracked artifacts (.DS_Store, Thumbs.db)
- Different suggested actions based on git status
- Scan completes in <5 seconds for typical projects
- Respects .gitignore automatically

**Validation**:
```bash
# Check scan results shown in TUI
# Media Files: 2 items (video.mp4, audio.wav) - "delete"
# Dev Artifacts: 3 items (.DS_Store, Thumbs.db, desktop.ini) - "ignore"
# Total Size: 15.2 MB
# Scan Duration: 1.3s
```

#### Scenario 2.2: Bulk Archive Tracked Media

**Goal**: Verify bulk operations on tracked media files

**Steps**:
1. With scan results from Scenario 2.1
2. Navigate to "Media Files" category
3. Select multiple files using Space
4. Press `a` to archive selected files
5. Confirm bulk archive operation
6. Verify files moved to `.distui-archive/media/`
7. Verify git status shows files removed

**Expected Outcome**:
- Multi-select works with visual indicators
- Bulk archive moves all selected files
- Archive preserves directory structure
- Git status updated automatically
- Notifications shown for each operation

#### Scenario 2.3: Bulk Ignore Untracked Artifacts

**Goal**: Verify bulk ignore operations for untracked files

**Steps**:
1. With scan results showing dev artifacts
2. Navigate to "Dev Artifacts" category
3. Select all .DS_Store and Thumbs.db files
4. Press `i` to add to .gitignore
5. Confirm operation
6. Verify .gitignore updated with patterns
7. Verify files no longer shown in git status

**Expected Outcome**:
- Patterns added to .gitignore (not individual files)
- .gitignore created if doesn't exist
- Patterns use glob syntax (*.DS_Store)
- Duplicate patterns not added

#### Scenario 2.4: Undo Last Operation

**Goal**: Verify users can undo cleanup actions

**Steps**:
1. Archive a file (from Scenario 2.2)
2. Press `u` to undo last operation
3. Verify file restored from archive
4. Verify git status restored

**Expected Outcome**:
- Last operation undone successfully
- File moved back to original location
- Git status matches pre-operation state
- Notification confirms undo

**Validation**:
```bash
# After archive
ls -la .distui-archive/media/video.mp4  # exists
git status | grep "deleted: video.mp4"  # shown

# After undo
ls -la video.mp4                         # exists
git status | grep "video.mp4"            # back to original state
```

### Feature 3: Branch Selection Modal for Push

#### Scenario 3.1: Push with Branch Selection

**Goal**: Verify Shift+B shows branch selection modal

**Steps**:
1. Create multiple local branches:
   ```bash
   git checkout -b feature/test
   git checkout main
   ```
2. Launch distui, navigate to Configure > Cleanup tab
3. Commit a file
4. Press `Shift+B` to open branch selection modal
5. Review list of available branches
6. Select branch with arrow keys
7. Press Enter to confirm
8. Verify push executes to selected branch

**Expected Outcome**:
- Shift+B opens full-screen modal
- All local branches listed
- Current branch highlighted
- Selection updates highlighted branch
- ESC closes modal without action
- Enter pushes to selected branch

#### Scenario 3.2: Quick Push with Shift+P

**Goal**: Verify existing Shift+P quick push still works

**Steps**:
1. Commit a file
2. Press `Shift+P` (existing key binding)
3. Verify push executes to current branch immediately
4. No modal shown

**Expected Outcome**:
- Shift+P remains unchanged (push current branch)
- Shift+B provides choice of target branch
- Both methods coexist without conflict

**Validation**:
```bash
# Check push succeeded
git log origin/feature/test --oneline | head -n1
# Should show latest commit
```

### Feature 4: UI Notifications System

#### Scenario 4.1: Auto-Dismiss Success Notifications

**Goal**: Verify success notifications auto-dismiss after 3 seconds

**Steps**:
1. Perform any successful operation (commit, archive, ignore)
2. Observe green success notification appears
3. Wait 3 seconds without interaction
4. Verify notification disappears automatically

**Expected Outcome**:
- Notification appears immediately after operation
- Auto-dismisses after exactly 3 seconds
- Doesn't block UI interaction
- Can be manually dismissed with ESC

#### Scenario 4.2: Persistent Error Notifications

**Goal**: Verify error notifications remain until dismissed

**Steps**:
1. Trigger an error (try to commit with no files selected)
2. Observe red error notification appears
3. Wait >3 seconds
4. Verify notification remains visible
5. Press ESC to dismiss

**Expected Outcome**:
- Error notifications stay until manually dismissed
- Red color distinguishes from success
- ESC dismisses notification
- Error message clear and actionable

#### Scenario 4.3: Warning Notifications

**Goal**: Verify warning notifications shown for reversible operations

**Steps**:
1. Archive a large file (>10MB)
2. Observe yellow warning notification
3. Wait 3 seconds
4. Verify auto-dismisses like success

**Expected Outcome**:
- Warnings shown for bulk operations
- Auto-dismiss after 3 seconds
- Yellow color distinguishes from success/error

**Validation**:
```bash
# No validation needed - purely visual test
# Check notification colors match spec:
# - Success: Green (#00ff00)
# - Warning: Yellow (#ffff00)
# - Error: Red (#ff0000)
```

## Integration Test: Full Workflow

**Goal**: Test all features together

**Steps**:
1. Start with clean Go project
2. Create test files with issues:
   ```bash
   touch video.mp4 .DS_Store README2.md
   git add video.mp4
   ```
3. Configure custom smart commit rules
4. Run repository scan (press `f` in Cleanup tab)
5. Archive tracked media files (video.mp4)
6. Ignore untracked artifacts (.DS_Store)
7. Commit remaining files using smart commit
8. Press `Shift+B` to select branch for push
9. Choose target branch and push
10. Verify success notification appears and auto-dismisses
11. Verify GitHub shows commit on selected branch

**Expected Outcome**:
- All features work together seamlessly
- Repository cleaned of problematic files
- .gitignore updated appropriately
- Files committed with correct categorization
- Push succeeds to selected branch
- Notifications guide user through workflow
- No repository pollution (all config in ~/.distui)

## Performance Validation

### Test 1: Pattern Matching Performance

**Steps**:
1. Create project with 200+ files
2. Add custom patterns with wildcards
3. Open Cleanup tab
4. Measure categorization time

**Expected Outcome**:
- Categorization completes in <100ms
- UI remains responsive
- No lag when switching tabs

### Test 2: Repository Scan Performance

**Steps**:
1. Create project with 1000+ files
2. Trigger full repository scan (press `f`)
3. Measure scan duration

**Expected Outcome**:
- gitcha scan completes in <3 seconds
- Untracked scan completes in <2 seconds
- Total scan time <5 seconds for typical projects
- Spinner shows progress during scan
- Results appear immediately after scan

**Measurement**:
```bash
# Check scan results in TUI shows duration
# Scan Duration: 2.8s (acceptable)
# Scan Duration: >10s (investigate gitcha performance)
```

## Error Scenarios

### Error 1: Invalid Pattern

**Steps**:
1. Try to add invalid glob pattern: `[[[invalid`
2. Press save

**Expected Outcome**:
- Error notification shown (red)
- Pattern not saved
- Config remains valid
- Notification persists until dismissed

### Error 2: Missing Workflow Directory

**Steps**:
1. Enable workflow generation
2. Delete `.github/` directory
3. Generate workflow

**Expected Outcome**:
- Directory created automatically
- Workflow file generated successfully
- Success notification shown
- No errors

### Error 3: No Write Permission

**Steps**:
1. Make `~/.distui/projects/` read-only
2. Try to save custom rules

**Expected Outcome**:
- Error notification shown clearly
- User informed of permission issue
- Graceful degradation (read-only mode)
- Notification remains until dismissed

### Error 4: Archive Directory Creation Fails

**Steps**:
1. Make project directory read-only
2. Try to archive a file

**Expected Outcome**:
- Error notification explains issue
- Original file not deleted
- Operation fully rolled back
- User can retry or skip

### Error 5: Git Remote Not Configured

**Steps**:
1. Create local-only git repository
2. Try to push with Shift+B or Shift+P

**Expected Outcome**:
- Error notification: "No remote configured"
- Suggests running `git remote add origin <url>`
- No push attempted
- Notification dismissible

## Cleanup

After testing:
```bash
# Remove test workflow file
rm -rf .github/workflows/release.yml

# Remove test files
rm -f video.mp4 audio.wav image.png .DS_Store Thumbs.db desktop.ini

# Remove archive directory
rm -rf .distui-archive/

# Reset custom rules (via TUI)
# OR manually:
# Edit ~/.distui/projects/<id>/config.yaml
# Set use_custom_rules: false

# Remove test branches
git branch -D feature/test

# Remove test commits if desired
git reset --soft HEAD~N  # N = number of test commits

# Clean .gitignore additions
# Edit .gitignore and remove test patterns
```

## Success Criteria

All scenarios pass when:
- ✅ Custom rules can be added/edited/removed
- ✅ Rules persist across restarts
- ✅ Workflow generation is opt-in
- ✅ Preview works before file creation
- ✅ Workflows can be disabled
- ✅ Dot files handled correctly
- ✅ Repository scan uses gitcha for git-aware detection
- ✅ Bulk operations work on tracked and untracked files
- ✅ Archive/ignore operations fully reversible with undo
- ✅ Branch selection modal shows all local branches
- ✅ Shift+B for branch selection, Shift+P for quick push
- ✅ Success notifications auto-dismiss after 3 seconds
- ✅ Error notifications persist until dismissed
- ✅ Performance meets targets (scan <5s, categorization <100ms)
- ✅ Error cases handled gracefully with clear notifications
- ✅ All features follow constitution (user agency, no forced changes, no repository pollution)

## Known Limitations

- Custom patterns use doublestar syntax (same as .gitignore)
- Workflow generation requires .github directory
- Pattern validation may reject complex but valid patterns (conservative validation)
- gitcha library requires git repository (non-git directories not scanned)
- Undo only supports last operation (no multi-level undo history)
- Branch selection shows local branches only (not remote branches)
- Archive directory (.distui-archive) created in project root (not in ~/.distui)

## Troubleshooting

### Smart Commit Preferences

**Issue**: Custom rules not applied
- Check `use_custom_rules: true` in config
- Verify patterns use correct syntax
- Check file paths match patterns
- Try pressing `r` to refresh cleanup tab

**Issue**: Workflow not generated
- Verify `.github/` directory exists
- Check file permissions
- Review preview for errors
- Check write access to project directory

**Issue**: Dot files still not visible
- Check git status shows them as modified
- Verify not in .gitignore
- Check file actually exists in working directory
- Press `r` to refresh file list

### Repository Cleanup

**Issue**: Scan takes too long (>10 seconds)
- Check project size (>10k files may be slow)
- Verify gitcha library installed correctly
- Check for symlink loops in directory structure
- Try excluding large vendor directories

**Issue**: Files not detected by scan
- Verify files are tracked in git (use `git status`)
- Check .gitignore doesn't exclude them
- Verify file extensions match patterns
- Press `f` to trigger fresh scan

**Issue**: Archive operation fails
- Check write permissions in project directory
- Verify sufficient disk space
- Check file not open in another application
- Try archiving single file first

**Issue**: Undo doesn't restore file
- Check .distui-archive directory exists
- Verify archive file not manually deleted
- Only last operation can be undone
- Check file wasn't committed after archive

### Branch Selection

**Issue**: Branch selection modal empty
- Verify git repository has branches
- Run `git branch` to check local branches
- Ensure not in detached HEAD state
- Check git repository initialized correctly

**Issue**: Push fails after branch selection
- Verify remote configured: `git remote -v`
- Check GitHub CLI authenticated: `gh auth status`
- Verify branch exists on remote
- Check network connectivity

### UI Notifications

**Issue**: Notifications not appearing
- Check terminal supports color output
- Verify TUI not in error state
- Try performing operation again
- Check notification area not hidden by modal

**Issue**: Notifications don't auto-dismiss
- Success notifications should dismiss after 3s
- Error notifications require manual dismiss (ESC)
- Check system time not frozen (rare VM issue)
- Verify tea.Tick command executing

## Installation Dependencies

If features don't work, verify dependencies:

```bash
# Check gitcha library installed
go list -m github.com/muesli/gitcha
# Should show: github.com/muesli/gitcha v0.2.0 (or later)

# Install if missing
go get github.com/muesli/gitcha

# Rebuild distui
go build -o distui app.go

# Verify version
./distui --version
# Should show: distui v0.1.0 or later
```