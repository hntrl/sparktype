# Outputs

The `outputs` array defines what files sparktype generates. Each output specifies a file path, format, and which schemas to include.

```jsonc
"outputs": [
  {
    "path": "./src/types/api.ts",
    "format": "typescript",
    "contents": ["api:*"]
  }
]
```

## Output Properties

### path

The file path where generated output will be written.

```jsonc
{
  "path": "./src/types/api.ts"
}
```

- Relative paths are resolved from the current working directory
- Parent directories are created automatically if they don't exist
- Existing files are overwritten

### format

The output format determines which generator is used:
| Format | Description | File Extension |
|--------|-------------|----------------|
| [`typescript`](/formats/typescript) | TypeScript interfaces | `.ts` |
| [`zod`](/formats/zod) | Zod schemas | `.ts` |
| [`python`](/formats/python) | Python TypedDicts | `.py` |
| [`go`](/formats/go) | Go structs | `.go` |

See [Output Formats](/formats/) for detailed documentation on each format.

### contents

An array defining which schemas to include and how to organize them. Items can be:
- **Pattern strings**: Select schemas by glob pattern
- **Namespace objects**: Group schemas into namespaces

```jsonc
{
  "contents": [
    "api:*",                    // Pattern: all schemas from 'api'
    "users:*Request",           // Pattern: request types from 'users'
    {                           // Namespace: grouped schemas
      "namespace": "Models",
      "contents": ["api:User", "api:Product"]
    }
  ]
}
```

See [Contents](./contents) for full pattern and namespace documentation.

### options

Format-specific options that override global settings:

```jsonc
{
  "format": "typescript",
  "options": {
    "exportType": "interface",
    "readonlyProperties": true
  }
}
```

Available options depend on the format. See [Options](./options) for details.

## Multiple Outputs

Generate multiple files from the same spec:

```jsonc
"outputs": [
  // TypeScript interfaces
  {
    "path": "./src/types/api.ts",
    "format": "typescript",
    "contents": ["api:*"]
  },
  
  // Zod schemas for runtime validation
  {
    "path": "./src/schemas/api.ts",
    "format": "zod",
    "contents": ["api:*"]
  },
  
  // Python types for backend
  {
    "path": "./backend/types/api.py",
    "format": "python",
    "contents": ["api:*"]
  }
]
```

## Splitting Output by Pattern

Generate separate files for different schema categories:

```jsonc
"outputs": [
  // Request/response types
  {
    "path": "./src/types/requests.ts",
    "format": "typescript",
    "contents": ["api:*Request", "api:*Response"]
  },
  
  // Model types
  {
    "path": "./src/types/models.ts",
    "format": "typescript",
    "contents": ["api:User", "api:Product", "api:Order"]
  },
  
  // Enum types
  {
    "path": "./src/types/enums.ts",
    "format": "typescript",
    "contents": ["api:*Status", "api:*Type"]
  }
]
```

## Combining Multiple Specs

Merge schemas from different specs into one file:

```jsonc
"outputs": [
  {
    "path": "./src/types/all.ts",
    "format": "typescript",
    "contents": [
      "users:*",
      "products:*",
      "orders:*"
    ]
  }
]
```

Or organize them with namespaces:

```jsonc
"outputs": [
  {
    "path": "./src/types/api.ts",
    "format": "typescript",
    "contents": [
      {
        "namespace": "Users",
        "contents": ["users:*"]
      },
      {
        "namespace": "Products",
        "contents": ["products:*"]
      }
    ]
  }
]
```

## Output Order

Schemas appear in the output in the order they're listed in `contents`. This matters when:
- Schemas reference each other (references must be defined first in some languages)
- You want a specific visual organization

sparktype automatically handles reference ordering within each content item, but the order of content items themselves is preserved.

```jsonc
"contents": [
  "api:BaseEntity",     // Appears first
  "api:User",           // References BaseEntity
  "api:*Request"        // Request types at the end
]
```

