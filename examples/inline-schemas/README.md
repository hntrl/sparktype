# Inline Schemas Example

Define schemas directly in the config and organize outputs with namespaces.

## Overview

This example demonstrates two advanced features:

1. **Inline schemas** - Define utility types without an OpenAPI file
2. **Namespaces** - Organize generated types into logical groups

```
inline-schemas/
├── typegen.jsonc     # Config with inline schemas
└── output/
    └── types.ts      # Generated with namespaced types
```

## Configuration

```jsonc
{
  "specs": {
    // External OpenAPI spec
    "api": { "path": "../basic/openapi.yaml" },
    
    // Inline utility schemas
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
        },
        "SortOrder": {
          "type": "string",
          "enum": ["asc", "desc"]
        },
        "ApiResponse": {
          "type": "object",
          "properties": {
            "success": { "type": "boolean" },
            "data": {},
            "error": { "type": "string" }
          },
          "required": ["success"]
        }
      }
    }
  },
  "outputs": [
    {
      "path": "./output/types.ts",
      "format": "typescript",
      "contents": [
        // Root-level utility types
        "utils:Pagination",
        "utils:SortOrder",
        "utils:ApiResponse",
        
        // API types in a namespace
        {
          "namespace": "API",
          "contents": [
            "api:User",
            "api:UserRole",
            "api:CreateUserRequest",
            "api:UpdateUserRequest"
          ]
        }
      ]
    }
  ]
}
```

## Generated Output

```typescript
// Root-level utility types
export interface Pagination {
  page: number;
  pageSize: number;
  total: number;
}

export type SortOrder = "asc" | "desc";

export interface ApiResponse {
  success: boolean;
  data?: unknown;
  error?: string;
}

// Namespaced API types
export namespace API {
  export interface User {
    id: string;
    email: string;
    // ...
  }
  
  export enum UserRole {
    Admin = "admin",
    User = "user",
  }
}
```

## Usage

```typescript
import { Pagination, SortOrder, API } from "./types";

// Use utility types at root level
const pagination: Pagination = { page: 1, pageSize: 20, total: 100 };

// Use API types with namespace
const user: API.User = { id: "...", email: "..." };
const role: API.UserRole = API.UserRole.Admin;
```

## When to Use

### Inline Schemas

- **Utility types** not in your OpenAPI spec (pagination, sorting, wrappers)
- **Cross-cutting concerns** shared across multiple specs
- **Quick additions** without modifying the OpenAPI file

### Namespaces

- **Large APIs** with many types to organize
- **Multiple domains** (Users, Products, Orders)
- **Avoiding conflicts** when combining multiple specs

## Next Steps

- See [multi-spec](../multi-spec) for working with multiple OpenAPI files
- See [monorepo](../monorepo) for domain-based organization at scale

