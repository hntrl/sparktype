# Configuration

sparktype is configured using a `typegen.jsonc` file in your project root. This file defines which OpenAPI specifications to read, what types to generate, and where to write the output.

## File Format

The configuration file uses [JSONC](https://code.visualstudio.com/docs/languages/json#_json-with-comments) (JSON with Comments), allowing you to add inline documentation to your config:

```jsonc
{
  // JSON Schema for IDE autocompletion
  "$schema": "https://hntrl.github.io/sparktype/schema.json",
  
  // Your OpenAPI spec sources
  "specs": { /* ... */ },
  
  // Output file definitions
  "outputs": [ /* ... */ ],
  
  // Global options (optional)
  "options": { /* ... */ }
}
```

## Schema Validation

Add the `$schema` property to enable IDE autocompletion and validation:

```jsonc
{
  "$schema": "https://hntrl.github.io/sparktype/schema.json"
}
```

Most modern editors will provide:
- Property autocompletion
- Inline documentation
- Validation errors

## Structure Overview

A configuration file has three main sections:

### specs

A map of named OpenAPI specification sources. Each spec can be:
- A **local file path** to a YAML or JSON spec
- A **remote URL** to fetch a spec from
- **Inline schemas** defined directly in the config

```jsonc
"specs": {
  "users": { "path": "./specs/users.yaml" },
  "products": { "url": "https://api.example.com/openapi.json" },
  "utils": { "schemas": { "Pagination": { /* ... */ } } }
}
```

[Learn more about specs →](./specs)

### outputs

An array of output file configurations. Each output specifies:
- **path**: Where to write the generated file
- **format**: The output format (`typescript`, `zod`, `python`, `go`)
- **contents**: Which schemas to include

```jsonc
"outputs": [
  {
    "path": "./src/types/api.ts",
    "format": "typescript",
    "contents": ["users:*", "products:Product"]
  }
]
```

[Learn more about outputs →](./outputs)

### options

Global generation settings that apply to all outputs:

```jsonc
"options": {
  "dereferenceRefs": false,
  "generateEnums": true,
  "nullableHandling": "optional"
}
```

[Learn more about options →](./options)

## Full Example

Here's a complete configuration file demonstrating common patterns:

```jsonc
{
  "$schema": "https://hntrl.github.io/sparktype/schema.json",
  
  "specs": {
    // Local OpenAPI spec
    "api": {
      "path": "./openapi.yaml"
    },
    
    // Remote spec with authentication
    "external": {
      "url": "https://api.partner.com/v2/openapi.json",
      "headers": {
        "Authorization": "Bearer ${API_TOKEN}"
      }
    },
    
    // Inline utility types
    "utils": {
      "schemas": {
        "Pagination": {
          "type": "object",
          "properties": {
            "page": { "type": "integer" },
            "pageSize": { "type": "integer" },
            "total": { "type": "integer" }
          },
          "required": ["page", "pageSize", "total"]
        }
      }
    }
  },
  
  "outputs": [
    // TypeScript types
    {
      "path": "./src/types/api.ts",
      "format": "typescript",
      "contents": [
        "api:*",
        "utils:*"
      ],
      "options": {
        "exportType": "interface"
      }
    },
    
    // Zod schemas with inferred types
    {
      "path": "./src/schemas/api.ts",
      "format": "zod",
      "contents": ["api:*"],
      "options": {
        "inferTypes": true
      }
    },
    
    // Organized by namespace
    {
      "path": "./src/types/namespaced.ts",
      "format": "typescript",
      "contents": [
        {
          "namespace": "API",
          "contents": ["api:*"]
        },
        {
          "namespace": "External",
          "contents": ["external:*"]
        }
      ]
    }
  ],
  
  "options": {
    "generateEnums": true,
    "nullableHandling": "optional"
  }
}
```

## Config File Location

By default, sparktype looks for `typegen.jsonc` in the current working directory. You can specify a different path with the `--config` flag:

```sh
sparktype generate --config ./config/typegen.jsonc
```

## Next Steps

- [Specs](./specs) - Learn about spec sources
- [Outputs](./outputs) - Configure output files
- [Contents](./contents) - Select and organize schemas
- [Options](./options) - Global and format-specific settings

