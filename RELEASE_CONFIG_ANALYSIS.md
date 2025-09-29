# Release Configuration Analysis

**Date**: 2025-09-29
**Purpose**: Inventory what's needed to make releases actually work

## What Release Executor Needs

From `internal/executor/release.go`:
```go
type ReleaseConfig struct {
    Version        string    // Selected at release time ✅
    SkipTests      bool      // From build settings ⚠️
    EnableHomebrew bool      // From distributions ⚠️
    EnableNPM      bool      // From distributions ⚠️
    HomebrewTap    string    // NOT IMPLEMENTED ❌
    RepoOwner      string    // Auto-detected ✅
    RepoName       string    // Auto-detected ✅
    ProjectName    string    // Auto-detected ✅
}
```

## Current Configure UI State

### Tab 1: Distributions
**What exists:**
```go
- [✓] GitHub Releases (always enabled - hardcoded)
- [✓] Homebrew (Tap: willyv3/homebrew-tap) - HARDCODED STRING
- [ ] NPM Package (Scope: @williavs) - HARDCODED STRING
- [✓] Go Install (enabled)
```

**Problems:**
1. **Checkboxes don't do anything** - toggle in memory only, not persisted
2. **Homebrew tap is hardcoded** - doesn't use DetectHomebrewTap()
3. **NPM scope is hardcoded** - no detection or configuration
4. **No save mechanism** - changes lost on exit
5. **No load mechanism** - doesn't read from ~/.distui/projects/{project}.yaml

### Tab 2: Build Settings
**What exists:**
```go
- [✓] Run tests before release
- [✓] Clean build directory
- [ ] Build for all platforms
- [ ] Include ARM64 builds
```

**Problems:**
1. **Logic is backwards** - "Run tests" enabled = SkipTests should be FALSE
2. **Not persisted** - settings lost on exit
3. **Not wired to ReleaseConfig** - doesn't pass to executor

### Tab 3: Advanced
**What exists:**
```go
- [ ] Create draft releases
- [ ] Mark as pre-release
- [✓] Generate changelog
- [✓] Sign commits
```

**Status:** Not used by release executor yet

## Critical Missing Pieces

### 1. ProjectConfig Structure ❌
**Problem:** No data structure to store release configuration

**Need:**
```go
// In internal/models/types.go
type ProjectConfig struct {
    // Existing fields...
    ReleaseSettings ReleaseSettings `yaml:"release_settings"`
}

type ReleaseSettings struct {
    EnableHomebrew bool   `yaml:"enable_homebrew"`
    HomebrewTap    string `yaml:"homebrew_tap"`
    EnableNPM      bool   `yaml:"enable_npm"`
    NPMScope       string `yaml:"npm_scope"`
    NPMPackage     string `yaml:"npm_package"`
    SkipTests      bool   `yaml:"skip_tests"`

    // Advanced
    CreateDraft    bool `yaml:"create_draft"`
    PreRelease     bool `yaml:"pre_release"`
    GenerateLog    bool `yaml:"generate_changelog"`
    SignCommits    bool `yaml:"sign_commits"`
}
```

### 2. Homebrew Tap Detection/Configuration ❌
**Problem:** Hardcoded string, doesn't use our DetectHomebrewTap function

**Need:**
- Call `DetectHomebrewTap(username)` on configure view init
- If found: populate with detected path
- If not found: show "Not configured" + [Edit] button
- Allow user to manually set tap path
- Save to project config

### 3. NPM Package Configuration ❌
**Problem:** Hardcoded scope, no actual configuration

**Need:**
- Detect from package.json if exists
- Allow user to configure scope + package name
- Save to project config

### 4. Config Persistence ❌
**Problem:** No save/load for distribution settings

**Need:**
- Save button handler in configure view
- `config.SaveProject(projectConfig)` call
- Load existing config on init
- Update checkboxes from loaded config

### 5. Wire Configure → Release ❌
**Problem:** ReleaseModel doesn't get config from ConfigureModel

**Current:** `NewReleaseModel()` has these hardcoded:
```go
EnableHomebrew: false,  // WRONG - should come from config
EnableNPM:      false,  // WRONG - should come from config
HomebrewTap:    "",     // WRONG - should come from config
```

**Need:** Pass project config to ReleaseModel:
```go
// In app.go when creating release model
if m.currentProject != nil && m.currentProject.ReleaseSettings.EnableHomebrew {
    m.releaseModel.EnableHomebrew = true
    m.releaseModel.HomebrewTap = m.currentProject.ReleaseSettings.HomebrewTap
}
```

## Implementation Architecture

### Data Flow (Current - BROKEN):
```
Configure UI Checkboxes
    ↓
  [NOWHERE] ← Lost in memory
    ↓
  [NOTHING]
    ↓
ReleaseModel (uses hardcoded false)
```

### Data Flow (Target - CORRECT):
```
Configure UI Checkboxes
    ↓
Space key toggles item
    ↓
[s] Save → internal/config.SaveProject()
    ↓
~/.distui/projects/{project}.yaml (persisted)
    ↓
App loads project config
    ↓
ReleaseModel.EnableHomebrew = config.ReleaseSettings.EnableHomebrew
ReleaseModel.HomebrewTap = config.ReleaseSettings.HomebrewTap
    ↓
ReleaseExecutor gets correct config
```

## Tasks Required

### High Priority (Blockers)
1. **Update ProjectConfig struct** - Add ReleaseSettings
2. **Implement config save** - Wire Space key to save toggle state
3. **Implement config load** - Load settings on configure view init
4. **Detect Homebrew tap** - Use DetectHomebrewTap() function
5. **Wire config to release** - Pass settings from project config to ReleaseModel

### Medium Priority (Important)
6. **Add tap configuration UI** - Allow editing homebrew tap path
7. **Add NPM configuration UI** - Allow editing scope/package
8. **Implement SkipTests logic** - Fix backwards logic
9. **Add save indicator** - Show when config is saved

### Low Priority (Polish)
10. **Validate configuration** - Check tap exists, npm package valid
11. **Show detection status** - "Detected: ~/homebrew-tap" vs "Not found"
12. **Add reset to defaults** - Clear configuration option

## File Changes Required

### 1. internal/models/types.go
- Add ReleaseSettings struct
- Add to ProjectConfig

### 2. handlers/configure_handler.go
- Add SaveProjectConfig() function
- Wire Space key to toggle + auto-save
- Load config on init
- Populate lists from loaded config
- Call DetectHomebrewTap() on init

### 3. internal/config/loader.go
- Ensure SaveProject() handles new ReleaseSettings field

### 4. app.go
- Pass project config to NewReleaseModel()
- Wire EnableHomebrew, HomebrewTap, EnableNPM from config

### 5. views/configure_view.go
- Show save indicator
- Show "Detected" vs "Not configured" status

## Breaking Issues Right Now

**If you press [r] for release today:**
1. ✅ Version selection works
2. ✅ Release execution runs
3. ❌ **Homebrew is DISABLED** (EnableHomebrew = false hardcoded)
4. ❌ **NPM is DISABLED** (EnableNPM = false hardcoded)
5. ❌ **Tap path is EMPTY** (HomebrewTap = "" hardcoded)
6. ❌ **Tests might be skipped incorrectly** (logic backwards)

**Result:** Release only creates GitHub release, skips all other channels

## Summary

**What works:**
- UI exists (tabs, checkboxes, lists)
- Distribution items display
- Can toggle checkboxes (in memory)

**What's broken:**
- Nothing is saved
- Nothing is loaded
- ReleaseModel doesn't get the config
- Homebrew tap not detected
- NPM not configured
- Logic backwards for tests

**Bottom line:** Configure UI is 100% cosmetic right now. Need to wire it to:
1. Persistent storage (YAML)
2. ReleaseModel initialization
3. Actual detection functions