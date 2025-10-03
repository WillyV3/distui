# Projects & Navigation

## Global Project List

Press `g` from anywhere to see all configured projects.

## Where Projects Come From

distui shows projects from **three sources**:

1. **Current directory** - Auto-detected when you launch distui in a Go project
2. **Saved projects** - Previously configured projects in `~/.distui/projects/*.yaml`
3. **Imported distributions** - From Homebrew taps or NPM registries (press `D`)

**NO automatic directory scanning.** Projects must be:
- Manually navigated to (cd your-project; distui)
- Saved via configuration
- Or imported from Homebrew/NPM

## Adding Projects

### Method 1: Navigate and Launch
```bash
cd ~/your-go-project
distui
# Auto-detects go.mod, saves to ~/.distui/projects/
```

### Method 2: Import Distributions
1. Press `g` for global view
2. Press `D` to detect/import
3. Searches your Homebrew tap and NPM scope
4. Creates project configs for found distributions
5. Set working directory for each

## Project Switching

When you select a project:
1. distui changes to that project's directory (`os.Chdir`)
2. Loads configuration from `~/.distui/projects/PROJECT.yaml`
3. Detects current Git state
4. Shows project view

You're literally managing the project from where it lives. No symlinks.

## Project View

Shows:
- Project name and version (from go.mod)
- GitHub info (if repo detected)
- Quick release options (press `r`)
- Distribution status (Homebrew/NPM)

## Detection Requirements

A valid Go project needs:
- `go.mod` file (required)
- `.git` directory (optional, recommended for releases)

No check for actual .go files - just go.mod existence.

## Global View Actions

- Arrow keys - Navigate projects
- Enter - Switch to selected project
- `D` - Detect/import distributions from Homebrew/NPM
- `r` - Refresh project list
- Esc - Back to current project

## If Project Not Found

If your project doesn't appear in global view:
1. Navigate to the directory in terminal
2. Run `distui` to detect it
3. Configure it (press `c`)
4. It'll be saved to ~/.distui/projects/

Projects aren't automatically discovered. You have to tell distui about them.
