# Release Files

## The Deal

distui can generate release configs for you, or you can keep your own. Your choice.

## Generated Files

Press `c` to configure, distui generates:
- `.goreleaser.yaml` - goreleaser config with distui markers
- `package.json` - NPM package config (if NPM distribution enabled)

These get created in YOUR project directory. Not hidden.

## Keep Your Own

Already have a `.goreleaser.yaml`? Cool. We detect it and use it.

distui shows a "custom" indicator when you're using your own files. We won't overwrite them unless you explicitly regenerate.

## Regeneration

If you fuck up your configs or want our latest:
1. Go to Configure view (`c`)
2. Hit the regenerate option
3. Confirm (we'll backup your old ones)

## What We Generate

Our configs are opinionated:
- Multi-platform builds (darwin/linux/windows, amd64/arm64)
- GitHub releases
- Homebrew tap support (if configured)
- Archive formats that make sense

Don't like it? Edit the files or use your own.

## Version Strategy

We support:
- Patch (0.0.1 → 0.0.2)
- Minor (0.1.0 → 0.2.0)
- Major (1.0.0 → 2.0.0)
- Custom (type whatever you want)