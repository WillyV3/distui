# Limitations

## Shit That Doesn't Work

Let's be real about what distui can't do.

## Changelogs

**Changelogs are disabled by default** in .goreleaser.yaml.

You can enable them:
1. Configure view (`c`)
2. Advanced Options tab
3. Toggle "Generate changelog"
4. Enter release notes during release flow

The feature works. Just turned off to keep releases fast.

## Git Operations

This is NOT a Git UI. Can't:
- View diffs
- Resolve merge conflicts (probably)
- Interactive rebase
- Cherry-pick commits
- Stash management

You can create PRs to main via gh CLI. That's about it.

## Language Support

**Go only.** Period.

This was hard enough for one language. Not adding Python, Rust, or whatever.

## Complex Workflows

distui assumes simple workflow:
- Work on branch
- Merge to main
- Tag and release

If you need:
- Multiple release branches
- Complex CI/CD pipelines
- Custom deployment targets
- Docker orchestration

Use the actual tools.

## Platform Limits

- Requires `gh` CLI for GitHub
- No GitLab/Bitbucket support
- No self-hosted GitHub support
- macOS and Linux only (Windows might work, untested)

## File Size

Can't handle:
- Repos over 1GB
- Binary files in releases (use LFS)
- Thousands of tags/branches (UI will lag)

## The Truth

I built this for myself. It works for my workflow. Your mileage may vary.