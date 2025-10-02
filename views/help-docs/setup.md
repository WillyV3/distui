# Setup

distui is a TUI for releasing Go applications. That's it.

## Requirements

- Go 1.21+
- git
- GitHub CLI (`gh`) authenticated
- goreleaser (we'll install it if you don't have it)

## Installation

```bash
go install github.com/yourusername/distui@latest
```

Or clone and build:
```bash
git clone https://github.com/yourusername/distui
cd distui
go build
```

## First Run

Just run `distui` in your Go project directory. It'll detect your project and create config files in `~/.distui/`.

No setup wizard. No config prompts. It just works.

## What It Does

1. Detects your Go project
2. Creates release configs if needed
3. Manages GitHub branches/tags
4. Runs goreleaser for you

That's literally it. If you need more complex shit, use the actual tools directly.