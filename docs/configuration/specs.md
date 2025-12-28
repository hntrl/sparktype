# Specs

The `specs` section defines your OpenAPI specification sources. Each spec is given a name that you'll reference when selecting schemas in your outputs.

```jsonc
"specs": {
  "api": { /* spec configuration */ },
  "users": { /* spec configuration */ },
  "products": { /* spec configuration */ }
}
```

## Spec Types

Each spec must use exactly one source type: `path`, `url`, or `schemas`.

### Local File (`path`)

Load a spec from a local file:

```jsonc
"specs": {
  "api": {
    "path": "./openapi.yaml"
  }
}
```

- Supports both YAML (`.yaml`, `.yml`) and JSON (`.json`) formats
- Relative paths are resolved from the directory containing `typegen.jsonc`
- Absolute paths are also supported

::: tip
Local specs are watched automatically when using `sparktype generate --watch`.
:::

### Remote URL (`url`)

Fetch a spec from a remote server:

```jsonc
"specs": {
  "external": {
    "url": "https://api.example.com/openapi.json"
  }
}
```

The content type is auto-detected from:
1. The `Content-Type` response header
2. The URL file extension

#### Authentication

Add custom headers for authenticated endpoints:

```jsonc
"specs": {
  "private-api": {
    "url": "https://api.internal.com/openapi.json",
    "headers": {
      "Authorization": "Bearer ${API_TOKEN}",
      "X-API-Key": "${API_KEY}"
    }
  }
}
```

Header values support environment variable interpolation using `${VAR_NAME}` syntax.

::: warning
Never commit secrets directly in your config file. Always use environment variables for sensitive values.
:::

### Inline Schemas (`schemas`)

Define schemas directly in the config without a full OpenAPI spec:

```jsonc
"specs": {
  "utils": {
    "schemas": {
      "Pagination": {
        "type": "object",
        "description": "Pagination parameters",
        "properties": {
          "page": { "type": "integer" },
          "pageSize": { "type": "integer" },
          "total": { "type": "integer" }
        },
        "required": ["page", "pageSize", "total"]
      },
      "SortOrder": {
        "type": "string",
        "enum": ["asc", "desc"]
      }
    }
  }
}
```

Inline schemas use the same format as OpenAPI Schema Objects. This is useful for:
- Utility types shared across projects
- Simple schemas that don't need full OpenAPI
- Testing and prototyping

## Environment Variables

Remote spec headers support environment variable expansion:

```jsonc
"headers": {
  "Authorization": "Bearer ${GITHUB_TOKEN}"
}
```

Variables are expanded at runtime. If a variable is not set, the literal `${VAR_NAME}` string is kept (which will likely cause authentication to fail).

Set variables before running sparktype:

```sh
export API_TOKEN="your-token-here"
sparktype generate
```

Or inline:

```sh
API_TOKEN="your-token" sparktype generate
```

## Multiple Specs

You can define multiple specs and mix source types:

```jsonc
"specs": {
  // Local development API
  "api": {
    "path": "./openapi.yaml"
  },
  
  // Partner API fetched remotely
  "partner": {
    "url": "https://partner.example.com/api/v2/openapi.json",
    "headers": {
      "X-Partner-Key": "${PARTNER_API_KEY}"
    }
  },
  
  // Shared utility types
  "common": {
    "schemas": {
      "UUID": { "type": "string", "format": "uuid" },
      "Timestamp": { "type": "string", "format": "date-time" }
    }
  }
}
```

Each spec's schemas are accessed using the `specName:pattern` syntax in your outputs:

```jsonc
"outputs": [
  {
    "path": "./types/all.ts",
    "format": "typescript",
    "contents": [
      "api:*",        // All schemas from 'api' spec
      "partner:User", // Just 'User' from 'partner' spec
      "common:*"      // All utility types
    ]
  }
]
```

## Best Practices

### Naming Conventions

Use descriptive, lowercase names for your specs:

```jsonc
// Good
"specs": {
  "users-api": { /* ... */ },
  "products-api": { /* ... */ }
}

// Avoid
"specs": {
  "spec1": { /* ... */ },
  "MySpec": { /* ... */ }
}
```

### Organizing Large Projects

For projects with many APIs, consider:

1. **One spec per service**: Keep specs focused
2. **Use namespaces**: Organize output by service (see [Contents](./contents))
3. **Separate config files**: For very large monorepos, use multiple `typegen.jsonc` files

### Remote Spec Caching

Remote specs are fetched fresh on every run. For faster builds:
- Cache specs locally in CI
- Use a spec management tool
- Consider switching to local paths for production builds

