# validate

Validate a `typegen.jsonc` configuration file for syntax and semantic errors.

## Usage

```sh
sparktype validate [options]
```

## Options

| Option | Alias | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Path to configuration file |

## Purpose

The `validate` command checks your configuration file for:

- **Syntax errors**: Invalid JSON/JSONC syntax
- **Schema errors**: Missing required fields, invalid values
- **Semantic errors**: References to undefined specs, invalid patterns

It does **not** load or parse OpenAPI specs—only the config file itself is validated.

## Examples

### Valid Configuration

```sh
sparktype validate
```

**Output:**

```
Configuration is valid: typegen.jsonc
```

### Invalid Configuration

```sh
sparktype validate
```

**Output:**

```
configuration validation failed: output[0]: unknown format "ts" (valid: typescript, zod, python, go)
```

### Custom Config Path

```sh
sparktype validate --config ./config/typegen.jsonc
```

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Configuration is valid |
| `1` | Configuration has errors |

## When to Use

### During Development

Run `validate` after editing your config to catch errors before running `generate`:

```sh
sparktype validate && sparktype generate
```

### In CI/CD

Add validation as a separate step:

```yaml
- name: Validate config
  run: npx sparktype validate

- name: Check types
  run: npx sparktype check
```

## Validation vs Generation

| Check | `validate` | `generate` |
|-------|------------|------------|
| Config syntax | Yes | Yes |
| Config schema | Yes | Yes |
| Spec file exists | No | Yes |
| Spec is valid OpenAPI | No | Yes |
| Schema patterns match | No | Yes |

Use `validate` for quick config checks. Use `generate` or `check` for full validation including spec files.

## Common Errors

### "unexpected end of JSON input"

Your config file has a syntax error. Common causes:
- Missing closing brace or bracket
- Trailing comma in arrays/objects
- Unclosed string

### "missing required field"

Add the required field to your config. Required fields are:
- `specs` (object)
- `outputs` (array)
- For each output: `path`, `format`, `contents`

### "unknown format"

Use one of: `typescript`, `zod`, `python`, `go`

### "spec must have exactly one source"

Each spec needs exactly one of:
- `path` (local file)
- `url` (remote URL)
- `schemas` (inline definitions)

## See Also

- [Configuration Overview](/configuration/)
- [init command](./init) - Create a new config
- [generate command](./generate)

