# distui

TUI for releasing Go applications. 30-second releases, zero repo pollution.

## Install

```bash
brew install willyv/tap/distui
```

or

```bash
go install github.com/willyv/distui@latest
```

## Usage

```bash
cd your-go-project
distui
```

Press `?` for help.

## What It Does

- Detects Go projects
- Generates `.goreleaser.yaml`
- Manages GitHub releases
- Cleans up old branches/tags
- Everything stored in `~/.distui/`

## Requirements

- Go 1.21+
- git
- gh CLI (authenticated)
- goreleaser

## Navigation

- `r` - Release
- `c` - Configure
- `g` - Global projects
- `?` - Help
- `q` - Quit

## First Time

uto-detects your Homebrew tap, generates configs. That's it.

## Limitations

- Go only
- Changelogs broken
- Basic Git ops only

Built because I was tired of reconfiguring release workflows for every TUI project.