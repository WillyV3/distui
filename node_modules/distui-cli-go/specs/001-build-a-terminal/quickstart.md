# distui Quick Start Guide

## Installation

### Via Homebrew
```bash
brew install yourusername/tap/distui
```

### Via Go Install
```bash
go install github.com/yourusername/distui@latest
```

### From Source
```bash
git clone https://github.com/yourusername/distui.git
cd distui
go build -o distui cmd/distui/main.go
mv distui /usr/local/bin/
```

## Prerequisites

Ensure you have these tools installed:
- `git` - Version control
- `gh` - GitHub CLI (authenticated)
- `goreleaser` - For building releases
- `brew` - If using Homebrew distribution

## First Launch

1. **Start distui in your Go project directory:**
```bash
cd ~/code/my-go-project
distui
```

2. **Initial setup (first time only):**
   - distui detects your project automatically
   - Confirms GitHub repository from git remotes
   - Finds your Homebrew tap (if exists)
   - Saves configuration to `~/.distui/`

## Basic Usage

### Navigation
- **TAB** - Cycle through views (Project → Global → Settings)
- **P** - Jump to Project view
- **G** - Jump to Global view (all projects)
- **S** - Jump to Settings

### Release Your Project

1. **In Project View, press `r` to start a release**

2. **Select version bump type:**
   - Patch (1.2.3 → 1.2.4)
   - Minor (1.2.3 → 1.3.0)
   - Major (1.2.3 → 2.0.0)
   - Custom version

3. **Watch the release execute:**
   ```
   ✓ Running tests
   ✓ Creating git tag v1.2.4
   ✓ Running GoReleaser
   ✓ Updating Homebrew tap
   ⠋ Publishing to NPM...
   ```

4. **Release complete in < 30 seconds!**

## Managing Multiple Projects

### View All Projects
1. Press **G** to open Global view
2. See all your configured projects
3. Press **Enter** to switch to a project
4. Press **N** to add a new project

### Quick Project Switching
```bash
# Run distui from any project directory
cd ~/code/another-project
distui
# Automatically loads this project's configuration
```

## Configuration

### Configure Current Project
1. In Project view, press **C**
2. Navigate tabs with **Tab**:
   - **Distributions**: Enable/disable GitHub, Homebrew, NPM
   - **Build**: GoReleaser settings, test commands
   - **CI/CD**: GitHub Actions generation

3. Edit fields with **Enter**
4. Save with **S**

### Global Settings
1. Press **S** for Settings view
2. Configure:
   - Default Homebrew tap location
   - NPM scope
   - UI preferences
   - Release defaults

## Distribution Channels

### GitHub Releases (Default)
- Automatically enabled for all projects
- Creates releases with assets from GoReleaser
- Generates changelogs from commits

### Homebrew Tap
1. Enable in project configuration
2. Specify your tap repository
3. distui updates formula automatically

### NPM Packages
1. Enable for projects with JavaScript bindings
2. Configure package name and scope
3. Publishes after GitHub release

## Advanced Features

### Custom Test Commands
```yaml
# In project configuration
build:
  test_command: "make test"
```

### GitHub Actions Generation
1. In Configure view, go to CI/CD tab
2. Enable GitHub Actions
3. distui generates `.github/workflows/release.yml`
4. Push to repository for automated releases

### Release History
- Press **H** in Project view
- See last 10 releases
- View success/failure status
- Check release duration

## Keyboard Shortcuts

### Global
- `q` - Quit
- `?` - Help
- `TAB` - Next view
- `/` - Search (in lists)

### Project View
- `r` - New release
- `c` - Configure
- `h` - History
- `t` - Test build

### Global View
- `↑/↓` - Navigate projects
- `Enter` - Open project
- `n` - New project
- `d` - Delete project

## Troubleshooting

### "gh CLI not authenticated"
```bash
gh auth login
```

### "Homebrew tap not found"
1. Create tap repository on GitHub
2. Clone locally:
   ```bash
   git clone git@github.com:yourusername/homebrew-tap.git ~/homebrew-tap
   ```
3. Update path in distui settings

### "Release failed"
- Check error message in output
- Ensure all prerequisites installed
- Verify GitHub permissions
- Check network connectivity

## Configuration Files

distui stores all configuration in `~/.distui/`:
```
~/.distui/
├── config.yaml                    # Global settings
├── projects/
│   └── github-com-user-*.yaml    # Per-project configs
└── cache/                         # Temporary files
```

## Tips & Tricks

### Fast Release
```bash
# Combine navigation and action
distui && r && Enter
# Opens distui, starts release, uses defaults
```

### Dry Run
- Hold **Shift** while pressing **R** for dry-run mode
- See what would happen without executing

### Verbose Output
- Press **V** during release for detailed command output
- Helpful for debugging issues

## Getting Help

### In-App Help
Press **?** at any time for context-sensitive help

### Documentation
Visit https://github.com/yourusername/distui/docs

### Report Issues
https://github.com/yourusername/distui/issues

## Next Steps

1. Configure your first project
2. Try a test release
3. Add more projects
4. Customize your workflow
5. Share with your team!

---

**Remember**: distui never adds files to your repositories. All configuration is stored globally in `~/.distui/`.