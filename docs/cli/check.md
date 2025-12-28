# check

Compare generated files against current OpenAPI specifications to detect drift.

## Usage

```sh
sparktype check [options]
```

## Options

| Option | Alias | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Path to configuration file |

## Purpose

The `check` command verifies that your generated type files match what would be generated from the current OpenAPI specifications. It's designed for CI/CD pipelines to ensure generated types stay in sync.

**Key behaviors:**
- Does not modify any files
- Compares existing files against expected output
- Exits with code `1` if any files are out of sync
- Shows detailed diff of what changed

## Examples

### Basic Check

```sh
sparktype check
```

**Output when in sync:**

```
Checking ./src/types/api.ts... OK
Checking ./src/schemas/api.ts... OK

All files are in sync.
```

**Output when out of sync:**

```
Checking ./src/types/api.ts... MISMATCH
  User:
    + email: string (added)
    ~ name: string -> string | undefined
Checking ./src/schemas/api.ts... OK

1 file(s) out of sync. Run 'sparktype generate' to update.
```

### Missing Files

If a generated file doesn't exist:

```
Checking ./src/types/api.ts... MISSING

1 file(s) out of sync. Run 'sparktype generate' to update.
```

### Custom Config Path

```sh
sparktype check --config ./config/typegen.jsonc
```

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | All files are in sync |
| `1` | One or more files are out of sync or missing |

## Diff Output

When files differ, `check` shows a structured diff:

```
Checking ./src/types/api.ts... MISMATCH
  + NewType (added)
  - RemovedType (removed)
  ChangedType:
    + newField: string (added)
    - oldField: number (removed)
    ~ modifiedField: string -> number
```

### Diff Symbols

| Symbol | Meaning |
|--------|---------|
| `+` | Added (type or property) |
| `-` | Removed (type or property) |
| `~` | Changed (shows old → new value) |

## Workflow Recommendations

### Fail Early

Run `check` early in your CI pipeline, before tests or builds:

```yaml
jobs:
  lint:
    # ...
  
  check-types:
    # ...
  
  test:
    needs: [lint, check-types]
    # ...
```

### Require Regeneration

When `check` fails, developers should:

1. Run `sparktype generate` locally
2. Review the changes
3. Commit the updated files

### Don't Auto-Generate in CI

Avoid running `generate` in CI and committing automatically. This hides spec changes from code review.

**Instead:**
1. Run `check` in CI (fails if out of sync)
2. Require developers to run `generate` locally
3. Include generated files in code review

## Comparison vs Generation

| Aspect | `generate` | `check` |
|--------|------------|---------|
| Modifies files | Yes | No |
| Exit code on diff | N/A | `1` |
| Shows diff | No | Yes |
| Use case | Development | CI/CD |

## Troubleshooting

### False Positives

If `check` reports differences but `generate` produces the same file:

1. Ensure you're using the same sparktype version in CI and locally
2. Check for line ending differences (CRLF vs LF)
3. Verify the same config options are used

### Performance

For large specs, `check` may take a few seconds. Cache dependencies in CI:

```yaml
- uses: actions/cache@v4
  with:
    path: ~/.npm
    key: ${{ runner.os }}-npm-${{ hashFiles('**/package-lock.json') }}
```

## See Also

- [generate command](./generate)
- [CI/CD Integration Guide](/guides/ci-cd)
- [Configuration](/configuration/)

