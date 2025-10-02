# First-Time Setup

## What Happens When You Run distui

### Initial Detection

When you run `distui` for the first time, it:

1. **Checks for ~/.distui/config.yaml**
   - No config? You're a first-time user
   - No GitHub username? Also first-time

2. **Shows ASCII Art Animation**
   - distui logo animates line-by-line
   - Color cycles: teal → orange → purple
   - Only shows on first run

3. **Creates Directory Structure**
   ```
   ~/.distui/
   ├── config.yaml          # Global settings
   ├── projects/            # Per-project configs
   └── templates/           # Release templates
   ```

### Project Detection

distui scans your current directory for:

1. **go.mod file** - Required. No Go module = no project
2. **.git directory** - Optional but recommended
3. **Existing files**:
   - .goreleaser.yaml
   - .goreleaser.yml
   - .github/workflows/
   - Any custom release configs

### Configuration Flow

#### New Projects (No Existing Files)

1. **Auto-detects distributions**:
   - Searches GitHub for Homebrew taps
   - Checks npm registry for packages
   - Matches by project name

2. **Shows what it found**:
   - Homebrew: yourname/homebrew-tap
   - NPM: @yourname/package

3. **Generates files**:
   - .goreleaser.yaml in YOUR project
   - Release metadata in ~/.distui

#### Projects With Existing Configs

1. **Detects custom files**:
   - Shows list of found configs
   - Asks: Keep custom or use distui?

2. **Custom Mode**:
   - Your files stay untouched
   - distui becomes a wrapper
   - Shows "custom" indicator

3. **distui Mode**:
   - Backs up your files
   - Generates new configs
   - Full management enabled

### Auto-Detection Magic

distui checks:

1. **GitHub via gh CLI**:
   ```bash
   gh repo list --json name
   gh api repos/USER/homebrew-TAP
   ```

2. **NPM Registry**:
   ```bash
   npm view @scope/package
   ```

3. **Local Git Remote**:
   ```bash
   git remote get-url origin
   ```

### Files Created in Your Project

Only if you choose distui management:

```
your-project/
├── .goreleaser.yaml     # Build configuration
├── .release.yaml        # Release metadata
└── scripts/            # Optional build scripts
```

### Skipping Setup

Press `Esc` anytime to skip and go to manual config.

Setup only runs once per project. Flag stored in:
`~/.distui/projects/YOUR_PROJECT.yaml` → `first_time_setup_completed: true`

### Manual Override

Don't like auto-detection? Press `m` to manually enter:
- Homebrew tap name
- NPM package name
- Distribution channels

### What Gets Saved

In ~/.distui/config.yaml:
```yaml
user:
  github_username: detected_from_git
  github_email: from_git_config
```

In ~/.distui/projects/YOUR_PROJECT.yaml:
```yaml
first_time_setup_completed: true
custom_files_mode: false  # or true
distributions:
  homebrew:
    enabled: true
    tap: yourname/homebrew-tap
```