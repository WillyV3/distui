# Quickstart: Testing Smart Commit Preferences & Workflow Generation

**Feature**: distui v0.0.32 enhancements
**Date**: 2025-09-30
**Purpose**: Validate new features work end-to-end

## Prerequisites

- distui v0.0.31 installed and working
- Go project with git repository
- GitHub CLI (`gh`) authenticated

## Test Scenarios

### Scenario 1: Custom Smart Commit Preferences

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

### Scenario 2: Reset to Default Rules

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

### Scenario 3: GitHub Workflow Generation (Opt-In)

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

### Scenario 4: Workflow Regeneration

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

### Scenario 5: Disable Workflow Generation

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

### Scenario 6: Dot File Bug Fix

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

## Integration Test: Full Workflow

**Goal**: Test all features together

**Steps**:
1. Start with clean Go project
2. Configure custom smart commit rules
3. Create test files matching custom patterns
4. Enable workflow generation
5. Generate workflow file
6. Commit workflow file using smart commit (uses custom rules)
7. Verify workflow file committed with correct category
8. Push to GitHub
9. Create tag to trigger workflow
10. Verify GitHub Actions runs successfully

**Expected Outcome**:
- All features work together seamlessly
- No conflicts or errors
- Configuration persists correctly
- Workflow executes successfully on GitHub

## Performance Validation

**Test**: Pattern matching performance

**Steps**:
1. Create project with 200+ files
2. Add custom patterns with wildcards
3. Open Cleanup tab
4. Measure categorization time

**Expected Outcome**:
- Categorization completes in <100ms
- UI remains responsive
- No lag when switching tabs

**Measurement**:
```bash
# Add timing to categorize function (development only)
time distui
# Should show instant UI response
```

## Error Scenarios

### Invalid Pattern

**Steps**:
1. Try to add invalid glob pattern: `[[[invalid`
2. Press save

**Expected Outcome**:
- Error message shown
- Pattern not saved
- Config remains valid

### Missing Workflow Directory

**Steps**:
1. Enable workflow generation
2. Delete `.github/` directory
3. Generate workflow

**Expected Outcome**:
- Directory created automatically
- Workflow file generated successfully
- No errors

### No Write Permission

**Steps**:
1. Make `~/.distui/projects/` read-only
2. Try to save custom rules

**Expected Outcome**:
- Error message shown clearly
- User informed of permission issue
- Graceful degradation (read-only mode)

## Cleanup

After testing:
```bash
# Remove test workflow file
rm -rf .github/workflows/release.yml

# Reset custom rules (via TUI)
# OR manually:
# Edit ~/.distui/projects/<id>/config.yaml
# Set use_custom_rules: false

# Remove test commits if desired
git reset --soft HEAD~N  # N = number of test commits
```

## Success Criteria

All scenarios pass when:
- ✅ Custom rules can be added/edited/removed
- ✅ Rules persist across restarts
- ✅ Workflow generation is opt-in
- ✅ Preview works before file creation
- ✅ Workflows can be disabled
- ✅ Dot files handled correctly
- ✅ Performance meets targets (<100ms)
- ✅ Error cases handled gracefully
- ✅ All features follow constitution (user agency, no forced changes)

## Known Limitations

- Custom patterns use doublestar syntax (same as .gitignore)
- Workflow generation requires .github directory
- Pattern validation may reject complex but valid patterns (conservative validation)

## Troubleshooting

**Issue**: Custom rules not applied
- Check `use_custom_rules: true` in config
- Verify patterns use correct syntax
- Check file paths match patterns

**Issue**: Workflow not generated
- Verify `.github/` directory exists
- Check file permissions
- Review preview for errors

**Issue**: Dot files still not visible
- Check git status shows them as modified
- Verify not in .gitignore
- Check file actually exists in working directory