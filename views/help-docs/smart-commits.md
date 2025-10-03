# Smart Commits

## What It Does

Smart commits categorize your changes by file type and generate semantic commit messages.

Example: Changes to `*.go` files → "code: update authentication logic"

## Accessing Smart Commit Preferences

1. Press `c` to open Configure view
2. You're on Cleanup tab by default
3. Press `p` to enter Smart Commit Preferences

## The UI

### Two Modes

**Default Mode (Disabled):**
- distui uses built-in conventional commit format
- No customization needed

**Custom Rules Mode (Enabled):**
- Define file patterns per category
- Categories: code, config, docs, build, test, assets, data
- Each category maps to a commit prefix

### Keybindings

**Normal Mode:**
- `space` - Toggle custom rules on/off
- `up/down` or `k/j` - Navigate categories
- `e` - Edit selected category
- `r` - Reset to defaults (with confirmation)
- `s` - Save configuration
- `esc` - Back to configure view

**Edit Category Mode:**
- `e` - Add file extension (.go, .md, etc.)
- `p` - Add glob pattern (**/*.test.go)
- `d` - Delete selected rule
- `esc` - Back to category list

**Adding Extension/Pattern:**
- Type the value
- `enter` - Save
- `esc` - Cancel

## How Categories Work

When you commit, distui scans changed files and matches them to categories:

```
code:     *.go, *.js, *.py
config:   *.yaml, *.json, *.toml
docs:     *.md, *.txt
build:    Makefile, Dockerfile, *.sh
test:     *_test.go, *.test.js
assets:   *.png, *.svg, *.css
data:     *.sql, *.db
```

Commit message format: `<category>: <description>`

## Storage

Rules saved in:
`~/.distui/projects/YOUR_PROJECT.yaml` under `smart_commit` section

## Default Rules

Pressing `r` resets to distui's built-in patterns.

## During Release

If smart commits enabled, distui:
1. Scans your changes
2. Groups files by category
3. Generates commit with dominant category
4. Shows preview before committing

You can still commit manually before running distui.