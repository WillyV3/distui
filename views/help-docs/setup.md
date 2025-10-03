# First-Time Setup

## What Happens When You Run distui

### Initial Detection

When you run `distui` for the first time, it:

1. **Checks for ~/.distui/config.yaml**
   - No config? You're a first-time user
   - No GitHub username? Also first-time


2. **Creates Directory Structure**
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

1. **Checks existing config files first**:
   - Looks for .goreleaser.yaml (extracts Homebrew tap)
   - Looks for package.json (extracts NPM package)

2. **Falls back to global config**:
   - Uses default Homebrew tap from ~/.distui/config.yaml
   - No NPM unless package.json exists

3. **Verifies with commands**:
   - `brew info tap/formula` (if Homebrew config found)
   - `npm view package version` (if package.json exists)

4. **Shows confirmation screen**:
   - Displays detected distributions
   - Press Enter to confirm, Esc to edit

5. **Generates files**:
   - .goreleaser.yaml in YOUR project
   - Config metadata in ~/.distui

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

### Detection Priority

distui checks in this order:

1. **Existing project files**:
   - .goreleaser.yaml → Homebrew tap/formula
   - package.json → NPM package name

2. **Global config fallback**:
   - ~/.distui/config.yaml → default Homebrew tap

3. **Verification commands**:
   ```bash
   brew info tap/formula --json=v2
   npm view package version
   git remote get-url origin
   ```

No GitHub API searches. Reads local files first.

### Files Created in Your Project

Only if you choose distui management:

```
your-project/
├── .goreleaser.yaml     # Build configuration
└── package.json         # NPM package (if NPM enabled)
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