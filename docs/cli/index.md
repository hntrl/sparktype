# CLI Reference

sparktype provides a command-line interface for generating, validating, and checking type definitions.

## Installation

::: code-group

```sh [npm]
npm install -D sparktype
```

```sh [pip]
pip install sparktype
```

```sh [Homebrew]
brew install hntrl/tap/sparktype
```

```sh [Go]
go install github.com/hntrl/sparktype/cmd/sparktype@latest
```

:::

### Direct Download

Pre-built binaries are available from [GitHub Releases](https://github.com/hntrl/sparktype/releases):

- macOS (Apple Silicon & Intel)
- Linux (x64 & ARM64)
- Windows (x64)

## Commands

| Command | Description |
|---------|-------------|
| [`generate`](./generate) | Generate types from OpenAPI specs |
| [`check`](./check) | Verify generated files match current specs |
| [`validate`](./validate) | Validate configuration file |
| [`init`](./init) | Create a new configuration file |

## Global Options

These options are available on all commands:

```
--help, -h     Show help message
--version, -v  Show version number
```

## Quick Reference

```sh
# Generate types
sparktype generate

# Generate with custom config path
sparktype generate --config ./config/typegen.jsonc

# Generate with watch mode
sparktype generate --watch

# Check if types are in sync (for CI)
sparktype check

# Validate configuration
sparktype validate

# Create new config file
sparktype init
```

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Error (invalid config, generation failure, or drift detected) |

## Configuration File Discovery

By default, all commands look for `typegen.jsonc` in the current directory. Override with `--config`:

```sh
sparktype generate --config ./path/to/typegen.jsonc
```

## npm Scripts

Add sparktype to your `package.json` scripts:

```json
{
  "scripts": {
    "types": "sparktype generate",
    "types:watch": "sparktype generate --watch",
    "types:check": "sparktype check"
  }
}
```

Then run:

```sh
npm run types
npm run types:watch
npm run types:check
```

## Next Steps

- [generate](./generate) - Full generate command documentation
- [check](./check) - CI/CD integration with check
- [validate](./validate) - Config validation
- [init](./init) - Initialize new projects

