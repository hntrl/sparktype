# init

Create a new `typegen.jsonc` configuration file with a starter template.

## Usage

```sh
sparktype init
```

## Options

The `init` command has no options. It always creates `typegen.jsonc` in the current directory.

## Behavior

Running `sparktype init`:

1. Checks if `typegen.jsonc` already exists
2. If not, creates a new file with a starter template
3. Prints the path to the created file

**Output:**

```
Created: typegen.jsonc
```

## Generated Template

```jsonc
{
  "$schema": "https://hntrl.github.io/sparktype/schema.json",

  // Named spec sources - supports multiple specs
  "specs": {
    "api": {
      "path": "./openapi.yaml"
    }
  },

  // Output configurations
  "outputs": [
    {
      "path": "./src/types/api.ts",
      "format": "typescript",
      "spec": "api"
    }
  ],

  // Global options
  "options": {
    "dereferenceRefs": true,
    "generateEnums": true,
    "nullableHandling": "optional"
  }
}
```

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | File created successfully |
| `1` | File already exists or write failed |

## After Initialization

After running `init`:

1. **Update the spec path**: Change `./openapi.yaml` to your actual OpenAPI spec location

2. **Update the output path**: Change `./src/types/api.ts` to where you want generated types

3. **Adjust options**: Review the global options and modify as needed

4. **Run generate**:
   ```sh
   sparktype generate
   ```

## Alternative: Manual Creation

If you prefer to create the config manually, see the [Configuration Overview](/configuration/) for the full schema documentation.

Minimum required config:

```jsonc
{
  "specs": {
    "api": { "path": "./openapi.yaml" }
  },
  "outputs": [
    {
      "path": "./types/api.ts",
      "format": "typescript",
      "contents": ["api:*"]
    }
  ]
}
```

## See Also

- [Getting Started](/getting-started)
- [Configuration Overview](/configuration/)
- [generate command](./generate)

