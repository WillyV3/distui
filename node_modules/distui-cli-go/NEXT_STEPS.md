# Next Steps: Release Configuration Implementation

**Date**: 2025-09-29
**Current State**: Release workflow core complete, but config disconnected

## What We Just Built ✅

**10 Release Workflow Tasks Complete:**
1. T012 - Homebrew detection
2. T030 - Message types
3. T022 - Command streaming
4. T023 - Test executor
5. T024 - GoReleaser executor
6. T025 - Homebrew updater
7. T021 - Release executor
8. T020 - Release handler
9. T019 - Release view
10. T031 - Wiring to app

**All compilation errors fixed!**

## The Problem 🔴

**The release UI works, but it's disconnected from configuration:**

```
User toggles checkbox: [✓] Homebrew
    ↓
NOWHERE - lost in memory
    ↓
Release executes with: EnableHomebrew = false (hardcoded)
    ↓
Result: Only GitHub release created, Homebrew skipped
```

**Root cause:** Configure UI is 100% cosmetic. No persistence, no loading, no wiring to release.

## What Needs to Happen Next 🔥

**9 Critical Configuration Tasks (T-CFG-1 to T-CFG-9):**

### Phase 1: Foundation (2 tasks, ~4 points)
**T-CFG-1**: Add ReleaseSettings struct to ProjectConfig
- New struct with all distribution/build settings
- YAML persistence ready

**T-CFG-3**: Load config on configure view init
- Read saved settings
- Populate checkboxes from disk

### Phase 2: Persistence (2 tasks, ~4 points)
**T-CFG-2**: Auto-save on every toggle
- Space key → save to ~/.distui/projects/{project}.yaml
- Modern editor pattern (save on every change)

**T-CFG-6**: Fix backwards SkipTests logic
- "Run tests" enabled = SkipTests FALSE

### Phase 3: Detection (2 tasks, ~6 points)
**T-CFG-4**: Detect actual Homebrew tap
- Call DetectHomebrewTap() we already built
- Show real path, not hardcoded string
- Allow user to edit

**T-CFG-5**: Configure NPM settings
- Detect from package.json
- Allow manual configuration
- Persist to config

### Phase 4: Wiring (3 tasks, ~6 points)
**T-CFG-7**: Wire config to ReleaseModel
- Pass settings from loaded project config
- Release uses real EnableHomebrew, HomebrewTap values

**T-CFG-8**: Show detection status in UI
- "✓ Detected: ~/homebrew-tap" (green)
- "⚠ Not configured" (yellow)

**T-CFG-9**: Add config validation
- Validate tap path exists
- Check gh CLI installed
- NPM package format valid

## Recommended Execution Order

**Do these in sequence (dependencies matter):**

1. **T-CFG-1** (2pts) - Add data structure
2. **T-CFG-3** (2pts) - Load existing config
3. **T-CFG-2** (3pts) - Save on toggle
4. **T-CFG-6** (1pt) - Fix logic bug
5. **T-CFG-4** (3pts) - Homebrew detection
6. **T-CFG-5** (3pts) - NPM configuration
7. **T-CFG-7** (2pts) - Wire to release
8. **T-CFG-8** (2pts) - Status display
9. **T-CFG-9** (2pts) - Validation

**Total effort:** ~20 points (~1-2 days)

## Files to Touch

**Must modify:**
1. `internal/models/types.go` - Add ReleaseSettings
2. `handlers/configure_handler.go` - Save/load logic
3. `app.go` - Pass config to ReleaseModel
4. `views/configure_view.go` - Status display

**Already have (no changes needed):**
- `internal/detection/homebrew.go` ✅
- `internal/executor/release.go` ✅
- All release execution logic ✅

## Testing After Implementation

**Manual test flow:**
1. Launch distui in project
2. Press [c] for configure
3. Toggle [✓] Homebrew
4. Verify saved to ~/.distui/projects/{project}.yaml
5. Restart distui
6. Verify checkbox still checked (loaded from disk)
7. Press [r] for release
8. Verify Homebrew step actually executes

**Expected behavior:**
- Config persists across restarts
- Release uses actual user configuration
- Homebrew publishes to correct tap
- NPM publishes if enabled

## What Works Without Config

**Still functional (using GitHub only):**
- Version selection UI
- Release execution flow
- Progress display
- Success/failure screens
- GitHub release creation

**These need config to work:**
- Homebrew tap updates (needs tap path)
- NPM publishing (needs scope/package)
- Test skipping (needs boolean)

## Summary

**Status:** 🟢 Foundation complete, 🔴 Configuration missing

**Blockers:** None technical - just need to implement the 9 config tasks

**Architecture:** All correct - just need to connect the dots

**Next action:** Start with T-CFG-1 (add data structure)