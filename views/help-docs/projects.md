# Projects & Navigation

## Global Detection

Press `g` from anywhere to see all your Go projects. distui scans common directories:
- ~/go/src
- ~/projects
- ~/code
- ~/dev
- Any parent directories with go.mod files

## How It Works

When you select a project in the global view, distui:
1. Changes to that project's directory
2. Loads its configuration
3. Shows you the project view

You're literally managing the project from where it lives. No symlinks, no copying files.

## Project View

Shows:
- Project name and version
- GitHub info (if connected)
- Quick release options
- Current branch status

## Switching Projects

1. Hit `g` for global view
2. Arrow keys to select
3. Enter to switch
4. You're now in that project's directory

## Detection Rules

A valid Go project needs:
- `go.mod` file
- `.git` directory (optional but recommended)
- Actual Go code files

If your project isn't detected, check those basics.