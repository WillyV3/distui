# Cleanup Tab

## What It Actually Does

The Cleanup tab manages your **local Git repository status**, not GitHub cleanup.

Access: Press `c` from project view → lands on Cleanup tab (first tab)

## Features

### Repository Status
Shows:
- Git initialization state
- GitHub remote configuration
- Repository existence on GitHub
- Uncommitted file changes
- Current branch
- Unpushed commits warning
- Smart commit mode status

### File Management
- View modified/added/deleted/untracked files
- Scan repository files (`f` key)
- Stage and commit changes

### Commit Operations
Two modes:

**Smart Commit** (`s` key):
- Auto-categorizes files by type
- Generates semantic commit messages
- Configure rules with `p` key

**Regular Commit** (`C` key):
- Standard git commit flow
- Write custom message
- Full control

### GitHub Operations
- Create GitHub repository (`G` key)
- Push to remote branches (`P` key → branch modal)
- Create pull requests

## Keybindings

**Main tab:**
- `s` - Smart commit
- `C` - Regular commit
- `p` - Smart commit preferences
- `f` - Scan repository files
- `G` - Create GitHub repository
- `P` - Branch/push modal
- `r` - Refresh status
- `space` - Cycle through options

**Branch modal:**
- Arrow keys - Select branch
- Enter - Push to selected branch or create PR
- Esc - Cancel

## Auto-Refresh

Refreshes GitHub repo status every time you navigate to this tab.

## What This ISN'T

**NOT a Git branch/tag manager.** Can't:
- Delete branches
- Manage tags
- Clean up releases
- View diffs
- Resolve conflicts

Use actual `git` or `gh` CLI for those operations.