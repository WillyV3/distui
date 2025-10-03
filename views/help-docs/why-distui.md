# Why distui Exists

## The Problem

I've been building TUIs like I'm running a marathon. Every project needs:
- goreleaser config
- GitHub Actions setup
- Release workflows
- Branch management

Every. Single. Time.

Either I copy-paste from old projects or have Claude regenerate it. Both suck.

## The Solution

One tool that:
- Detects my Go projects
- Generates configs once
- Manages releases
- Cleans up GitHub mess

Built for myself. Shared because why not.

## Design Choices

**Minimal repo footprint**: Project metadata in `~/.distui/`. Release configs (.goreleaser.yaml, workflows) in your repo where they belong.

**Fast releases**: Quick version bump, build, push. Timing varies by project size.

**Direct command execution**: No shell scripts. Uses goreleaser and gh CLI directly.

**Opinionated**: It works how I work. Don't like it? Fork it.

## What's Next

Honestly? I'll probably rebuild this after learning what works and what doesn't.

But for now, it ships. It works. I use it daily.

## The Philosophy

Tools should:
- Do one thing well
- Get out of your way
- Not require documentation
- Just fucking work

distui tries to follow that.

## Credits

Built with:
- Bubble Tea (TUI framework)
- Goreleaser (the actual release tool)
- Too much coffee

--

*If you're reading this, you're probably as tired of release configs as I am.*