# Smart Commits

## For Opinionated Go Devs

distui can create commits for you. Or don't use it. We don't care.

## How It Works

During release setup:
1. Choose "Smart Commit" option
2. distui stages your changes
3. Creates a semantic commit message
4. Pushes to your branch

## Commit Format

We follow conventional commits:
```
feat: add new terminal UI
fix: resolve goreleaser config issue
chore: update dependencies
```

## Creating Your Own

Don't like our format? Make your own:

1. Edit `~/.distui/commit-templates.yaml`
2. Add your patterns:
```yaml
templates:
  release: "release: v{version}"
  feature: "feat: {description}"
```

## Current Branch Workflow

Smart commits work with your current branch:
1. Commit changes
2. Push to origin
3. Optionally merge to main
4. Create release

We handle the git gymnastics.

## Manual Override

Press `m` in release view to manually:
- Stage specific files
- Write custom commit message
- Push to different branch

## Warning

Smart commits will:
- Stage ALL changes
- Create opinionated messages
- Push immediately

Not comfortable with that? Commit manually first, then release.