# Contents

The `contents` array determines which schemas from your specs are included in an output file and how they're organized. It supports two item types: **patterns** and **namespaces**.

## Pattern Syntax

Patterns select schemas using the format `specName:glob`:

```jsonc
"contents": [
  "api:*",           // All schemas from 'api' spec
  "users:User",      // Just 'User' schema from 'users' spec
  "products:*Item"   // Schemas ending with 'Item' from 'products'
]
```

### Pattern Components

| Component | Description |
|-----------|-------------|
| `specName` | The name of a spec defined in your `specs` section |
| `:` | Separator between spec name and glob |
| `glob` | Pattern to match schema names |

### Glob Patterns

The glob portion supports these wildcards:

| Pattern | Matches |
|---------|---------|
| `*` | Any characters (e.g., `*Request` matches `CreateUserRequest`) |
| `?` | Single character |
| `[abc]` | Character class |
| `[a-z]` | Character range |

Examples:

```jsonc
"contents": [
  "api:*",              // All schemas
  "api:User*",          // User, UserProfile, UserSettings, etc.
  "api:*Request",       // CreateUserRequest, UpdateUserRequest, etc.
  "api:*User*",         // Anything containing 'User'
  "api:???Error",       // Three-letter prefix + Error (e.g., ApiError)
]
```

### Explicit Schema Lists

For precise control, list schemas explicitly:

```jsonc
"contents": [
  "api:User",
  "api:UserProfile",
  "api:CreateUserRequest",
  "api:UpdateUserRequest"
]
```

This gives you:
- Exact control over which schemas are included
- Explicit ordering in the output
- Protection against accidentally including new schemas

## Namespaces

Namespaces group schemas into nested structures. They're supported by `typescript` and `zod` formats.

::: code-group

```jsonc [Config]
"contents": [
  {
    "namespace": "Models",
    "contents": [
      "api:User",
      "api:Product"
    ]
  }
]
```

```typescript [TypeScript Output]
export namespace Models {
  export interface User {
    id: string;
    name: string;
  }
  
  export interface Product {
    id: string;
    title: string;
  }
}
```

```typescript [Zod Output]
import * as z from "zod";

// Schemas at top level
export const userSchema = z.object({
  id: z.string(),
  name: z.string()
});

export const productSchema = z.object({
  id: z.string(),
  title: z.string()
});

// Inferred types in namespace
export namespace Models {
  export type User = z.infer<typeof userSchema>;
  export type Product = z.infer<typeof productSchema>;
}
```

:::

### Nested Namespaces

Namespaces can be nested arbitrarily deep:

::: code-group

```jsonc [Config]
"contents": [
  {
    "namespace": "API",
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

```typescript [TypeScript Output]
export namespace API {
  export namespace Users {
    export interface User {
      id: string;
      email: string;
    }
  }
  
  export namespace Products {
    export interface Product {
      id: string;
      name: string;
    }
  }
}
```

```typescript [Zod Output]
import * as z from "zod";

// Schemas at top level
export const userSchema = z.object({
  id: z.string(),
  email: z.string()
});

export const productSchema = z.object({
  id: z.string(),
  name: z.string()
});

// Inferred types in nested namespaces
export namespace API {
  export namespace Users {
    export type User = z.infer<typeof userSchema>;
  }
  
  export namespace Products {
    export type Product = z.infer<typeof productSchema>;
  }
}
```

:::

### Mixed Contents

Combine top-level schemas with namespaces:

::: code-group

```jsonc [Config]
"contents": [
  // Top-level utility types
  "utils:Pagination",
  "utils:SortOrder",
  
  // Namespaced API types
  {
    "namespace": "API",
    "contents": ["api:*"]
  }
]
```

```typescript [TypeScript Output]
// Top-level types
export interface Pagination {
  page: number;
  pageSize: number;
  total: number;
}

export type SortOrder = "asc" | "desc";

// Namespaced types
export namespace API {
  export interface User {
    id: string;
    name: string;
  }
}
```

:::

## Pattern Examples

### Include All Types

```jsonc
"contents": ["api:*"]
```

### Request/Response Types Only

```jsonc
"contents": [
  "api:*Request",
  "api:*Response"
]
```

### Exclude Internal Types

If you have internal types prefixed with `_` or `Internal`, list only public types:

```jsonc
"contents": [
  "api:User",
  "api:Product",
  "api:Order",
  "api:*Request",
  "api:*Response"
]
```

### Multiple Specs Combined

```jsonc
"contents": [
  "users:*",
  "products:*",
  "shared:*"
]
```

### Organized by Domain

```jsonc
"contents": [
  {
    "namespace": "Users",
    "contents": [
      "users:User",
      "users:UserProfile",
      "users:UserSettings"
    ]
  },
  {
    "namespace": "Products",
    "contents": [
      "products:Product",
      "products:Category",
      "products:Price"
    ]
  },
  {
    "namespace": "Requests",
    "contents": [
      "users:*Request",
      "products:*Request"
    ]
  }
]
```

## Automatic Dependencies

When you include a schema that references other schemas, sparktype automatically includes those dependencies. This ensures your generated output is always complete and valid.

```jsonc
"contents": [
  "api:Order"  // If Order references User and Product, they're auto-included
]
```

### How It Works

1. **Direct references** - If `Order` has a `user` property of type `User`, the `User` schema is automatically included
2. **Transitive dependencies** - If `User` references `Address`, that's included too
3. **Same namespace** - Auto-included schemas are placed in the same namespace as the schema that first references them

### Example

If your spec has these schemas:

```yaml
Order:
  properties:
    user:
      $ref: '#/components/schemas/User'
    items:
      type: array
      items:
        $ref: '#/components/schemas/OrderItem'

User:
  properties:
    address:
      $ref: '#/components/schemas/Address'
```

And your config only explicitly includes `Order`:

```jsonc
"contents": [
  {
    "namespace": "API",
    "contents": ["api:Order"]
  }
]
```

The output will include `Order`, `User`, `OrderItem`, and `Address` - all in the `API` namespace.

### Explicit Control

If you want to organize dependencies into different namespaces, include them explicitly:

```jsonc
"contents": [
  {
    "namespace": "Models",
    "contents": ["api:User", "api:Address"]
  },
  {
    "namespace": "Orders",
    "contents": ["api:Order", "api:OrderItem"]
  }
]
```

When a schema is explicitly included, it won't be auto-included elsewhere - the explicit placement takes precedence.

## Cross-References

When schemas reference each other across namespaces, sparktype automatically resolves the correct reference path:

::: code-group

```jsonc [Config]
"contents": [
  {
    "namespace": "Models",
    "contents": ["api:User", "api:Address"]
  },
  {
    "namespace": "Requests",
    "contents": ["api:CreateUserRequest"]  // References User
  }
]
```

```typescript [TypeScript Output]
export namespace Models {
  export interface User {
    id: string;
    address: Address;
  }
  
  export interface Address {
    street: string;
    city: string;
  }
}

export namespace Requests {
  export interface CreateUserRequest {
    user: Models.User;  // Cross-namespace reference
  }
}
```

```typescript [Zod Output]
import * as z from "zod";

// Schemas at top level
export const addressSchema = z.object({
  street: z.string(),
  city: z.string()
});

export const userSchema = z.object({
  id: z.string(),
  address: addressSchema
});

export const createUserRequestSchema = z.object({
  user: userSchema
});

// Inferred types in namespaces
export namespace Models {
  export type Address = z.infer<typeof addressSchema>;
  export type User = z.infer<typeof userSchema>;
}

export namespace Requests {
  export type CreateUserRequest = z.infer<typeof createUserRequestSchema>;
}
```

:::

If `CreateUserRequest` has a `user` property referencing `User`, the generated code will use `Models.User` as the type reference.

## Namespace Support by Format

| Format | Namespace Support |
|--------|-------------------|
| `typescript` | Full support |
| `zod` | Full support |
| `python` | Not supported (error if used) |
| `go` | Not supported (error if used) |

For formats that don't support namespaces, use multiple output files instead:

```jsonc
"outputs": [
  {
    "path": "./types/users.py",
    "format": "python",
    "contents": ["users:*"]
  },
  {
    "path": "./types/products.py",
    "format": "python",
    "contents": ["products:*"]
  }
]
```
