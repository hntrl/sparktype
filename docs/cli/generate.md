# generate

Generate type definitions from OpenAPI specifications.

## Usage

```sh
sparktype generate [options]
```

## Options

| Option | Alias | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Path to configuration file |
| `--watch` | `-w` | Watch for changes and regenerate |

## Examples

### Basic Generation

```sh
sparktype generate
```

Looks for `typegen.jsonc` in the current directory and generates all configured outputs.

**Output:**

```
Generated: ./src/types/api.ts
Generated: ./src/schemas/api.ts
```

### Custom Config Path

```sh
sparktype generate --config ./config/typegen.jsonc
```

Use a configuration file from a different location.

### Watch Mode

```sh
sparktype generate --watch
```

Watches OpenAPI spec files and the config file for changes, regenerating automatically when modifications are detected.

**Output:**

```
Generated: ./src/types/api.ts

Watching for changes... (Press Ctrl+C to stop)

Detected change in: openapi.yaml
Regenerating...
Generated: ./src/types/api.ts
```

Watch mode monitors:
- All local spec files referenced in your config
- The `typegen.jsonc` config file itself

::: tip
Remote specs (loaded via `url`) are not watched. They're fetched fresh on each generation.
:::

### Combined Options

```sh
sparktype generate -c ./config/typegen.jsonc -w
```

Watch mode with a custom config path.

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | All outputs generated successfully |
| `1` | One or more outputs failed |

## Integration with Build Tools

::: details npm / Node.js

```json
{
  "scripts": {
    "prebuild": "sparktype generate",
    "types": "sparktype generate",
    "types:watch": "sparktype generate --watch",
    "types:check": "sparktype check"
  }
}
```

Run with `npm run types` or automatically before builds with `prebuild`.

:::

::: details Go

Use `go:generate` to integrate with Go's build tooling:

```go
//go:generate sparktype generate

package main
```

Then run:

```sh
go generate ./...
```

:::

::: details Python / uv

Using [uv](https://github.com/astral-sh/uv) with sparktype installed via pip:

```sh
# Generate types
uv run sparktype generate

# Watch mode
uv run sparktype generate --watch
```

Or in a `pyproject.toml` script:

```toml
[project.scripts]
types = "sparktype generate"
types-check = "sparktype check"
```

:::

::: details Makefile

```makefile
.PHONY: types types-watch types-check

types:
	sparktype generate

types-watch:
	sparktype generate --watch

types-check:
	sparktype check
```

:::
## See Also

- [Configuration Overview](/configuration/)
- [check command](./check) - Verify generated files
- [CI/CD Integration](/guides/ci-cd)

