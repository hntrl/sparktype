# sparktype

Generate static types from OpenAPI specifications.

## Installation

```bash
pip install sparktype
# or
pipx install sparktype
```

## Usage

```bash
# Generate types using typegen.jsonc in current directory
sparktype generate

# Specify config file
sparktype generate --config ./path/to/typegen.jsonc

# Initialize a new config file
sparktype init

# Validate config file
sparktype validate
```

## Configuration

Create a `typegen.jsonc` file:

```jsonc
{
  "specs": {
    "api": {
      "path": "./openapi.yaml"
    }
  },
  "outputs": [
    {
      "path": "./types.py",
      "format": "python",
      "spec": "api"
    }
  ]
}
```

## Supported Output Formats

- `typescript` - TypeScript interfaces
- `zod` - Zod schemas with inferred types
- `zod-typescript` - Both Zod schemas and TypeScript interfaces
- `python` - Python TypedDict classes
- `go` - Go structs with JSON tags

## Documentation

For full documentation, visit: https://github.com/hntrl/sparktype

