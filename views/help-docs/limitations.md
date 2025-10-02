# Limitations

## Shit That Doesn't Work

Let's be real about what distui can't do.

## Changelogs

**Changelog generation is broken.** Sorry.

If you need changelogs, release without distui for now. We're working on it.

## Git Operations

This is NOT a Git UI. Can't:
- View diffs
- Resolve merge conflicts (probably)
- Interactive rebase
- Cherry-pick commits
- Stash management

You can merge current branch to main. That's about it.

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