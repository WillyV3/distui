# GitHub Cleanup

## What It Does

The Cleanup view (`c` → Cleanup tab) manages your GitHub mess:
- Old branches
- Duplicate tags
- Failed releases

## Branch Management

Shows all your remote branches with:
- Last commit date
- Whether it's merged
- Quick delete option

Select branches with space, delete with `d`. We won't let you delete main/master.

## Tag Cleanup

Lists all tags with:
- Version
- Release status
- Creation date

Delete old or broken tags before re-releasing.

## Auto-Refresh

The cleanup view refreshes GitHub status every time you navigate to it. No stale data.

## Safety

We always:
- Show what will be deleted
- Require confirmation
- Use `gh` CLI for authentication
- Never force-delete

## Quick Actions

- `r` - Refresh all data
- `space` - Select/deselect
- `d` - Delete selected
- `a` - Select all (except protected)
- `n` - Deselect all

## Limitations

This isn't a full Git UI. Can't:
- View diffs
- Resolve conflicts
- Cherry-pick
- Rebase

Use actual git for complex operations.