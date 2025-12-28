# Multi-Spec Example

Work with multiple OpenAPI specifications and organize outputs flexibly.

## Overview

This example demonstrates how to:

- Load multiple OpenAPI specs
- Generate separate files for each spec
- Combine specs into a single output
- Filter types using glob patterns
- Organize with namespaces

```
multi-spec/
├── users-api.yaml       # Users service spec
├── products-api.yaml    # Products service spec
├── typegen.jsonc        # Configuration
└── types/
    ├── users.ts         # All users types
    ├── products.ts      # All products types
    ├── users-requests.ts # Only request types from users
    ├── products-public.ts # Selected public types
    └── all-namespaced.ts  # Both specs, namespaced
```

## Configuration

```jsonc
{
  "specs": {
    "users": { "path": "./users-api.yaml" },
    "products": { "path": "./products-api.yaml" }
  },
  "outputs": [
    // Separate files per spec
    { "path": "./types/users.ts", "contents": ["users:*"] },
    { "path": "./types/products.ts", "contents": ["products:*"] },
    
    // Pattern matching - only *Request types
    { "path": "./types/users-requests.ts", "contents": ["users:*Request"] },
    
    // Explicit selection
    { 
      "path": "./types/products-public.ts", 
      "contents": ["products:Product", "products:Price", "products:Category"] 
    },
    
    // Combined with namespaces
    {
      "path": "./types/all-namespaced.ts",
      "contents": [
        { "namespace": "Users", "contents": ["users:*"] },
        { "namespace": "Products", "contents": ["products:*"] }
      ]
    }
  ]
}
```

## Content Patterns

sparktype supports glob patterns for flexible type selection:

| Pattern | Matches |
|---------|---------|
| `api:*` | All types from "api" spec |
| `api:User` | Only the User type |
| `api:*Request` | Types ending in "Request" |
| `api:Create*` | Types starting with "Create" |
| `api:*User*` | Types containing "User" |

## Generated Output

### Separate Files

`users.ts`:
```typescript
export interface User { ... }
export interface UserProfile { ... }
export interface CreateUserRequest { ... }
```

`products.ts`:
```typescript
export interface Product { ... }
export interface Price { ... }
export interface Category { ... }
```

### Pattern-Filtered

`users-requests.ts`:
```typescript
// Only types matching *Request
export interface CreateUserRequest { ... }
export interface UpdateUserRequest { ... }
```

### Namespaced

`all-namespaced.ts`:
```typescript
export namespace Users {
  export interface User { ... }
  export interface CreateUserRequest { ... }
}

export namespace Products {
  export interface Product { ... }
  export interface Price { ... }
}
```

## Usage

```typescript
// Import from separate files
import { User, CreateUserRequest } from "./types/users";
import { Product } from "./types/products";

// Or from namespaced file
import { Users, Products } from "./types/all-namespaced";

const user: Users.User = { ... };
const product: Products.Product = { ... };
```

## Use Cases

### Microservices

Each service has its own OpenAPI spec but shares common types:

```jsonc
"specs": {
  "users": { "url": "https://users-service/openapi.json" },
  "products": { "url": "https://products-service/openapi.json" },
  "common": { "path": "./common-schemas.yaml" }
}
```

### API Versioning

Generate types for multiple API versions:

```jsonc
"specs": {
  "v1": { "path": "./api-v1.yaml" },
  "v2": { "path": "./api-v2.yaml" }
},
"outputs": [
  { "path": "./types/v1.ts", "contents": ["v1:*"] },
  { "path": "./types/v2.ts", "contents": ["v2:*"] }
]
```

### Domain Organization

Split by domain regardless of source spec:

```jsonc
"outputs": [
  {
    "path": "./types/auth.ts",
    "contents": ["users:User", "users:Session", "users:*Auth*"]
  },
  {
    "path": "./types/commerce.ts", 
    "contents": ["products:*", "orders:*"]
  }
]
```

## Next Steps

- See [monorepo](../monorepo) for organizing types across a large project
- See [inline-schemas](../inline-schemas) for adding utility types

