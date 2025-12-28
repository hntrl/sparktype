# Working with Multiple Specs

This guide covers strategies for working with multiple OpenAPI specifications in a single project, including microservices, third-party APIs, and domain separation.

## When to Use Multiple Specs

Common scenarios:
- **Microservices**: Each service has its own OpenAPI spec
- **Third-party APIs**: Integrating external APIs alongside your own
- **Domain separation**: Splitting a large API into logical domains
- **API versions**: Managing multiple API versions

## Basic Multi-Spec Setup

Define multiple specs in your configuration:

```jsonc
{
  "specs": {
    "users": {
      "path": "./specs/users-api.yaml"
    },
    "products": {
      "path": "./specs/products-api.yaml"
    },
    "orders": {
      "path": "./specs/orders-api.yaml"
    }
  },
  "outputs": [
    {
      "path": "./src/types/api.ts",
      "format": "typescript",
      "contents": [
        "users:*",
        "products:*",
        "orders:*"
      ]
    }
  ]
}
```

## Strategies

### Strategy 1: Single Combined File

Combine all types into one file:

```jsonc
"outputs": [
  {
    "path": "./src/types/api.ts",
    "format": "typescript",
    "contents": [
      "users:*",
      "products:*",
      "orders:*"
    ]
  }
]
```

### Strategy 2: Separate Files Per Spec

One output file per spec:

```jsonc
"outputs": [
  {
    "path": "./src/types/users.ts",
    "format": "typescript",
    "contents": ["users:*"]
  },
  {
    "path": "./src/types/products.ts",
    "format": "typescript",
    "contents": ["products:*"]
  },
  {
    "path": "./src/types/orders.ts",
    "format": "typescript",
    "contents": ["orders:*"]
  }
]
```

### Strategy 3: Namespaced Output

Use namespaces to organize in a single file:

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
      },
      {
        "namespace": "Orders",
        "contents": ["orders:*"]
      }
    ]
  }
]
```

**Generated:**

```typescript
export namespace Users {
  export interface User { /* ... */ }
  export interface UserProfile { /* ... */ }
}

export namespace Products {
  export interface Product { /* ... */ }
  export interface Category { /* ... */ }
}

export namespace Orders {
  export interface Order { /* ... */ }
  export interface OrderItem { /* ... */ }
}
```

**Usage:**

```typescript
import { Users, Products, Orders } from './types/api';

const user: Users.User = { /* ... */ };
const product: Products.Product = { /* ... */ };
```

### Strategy 4: Hybrid Approach

Combine strategies based on use case:

```jsonc
"outputs": [
  // Namespaced for internal APIs
  {
    "path": "./src/types/internal.ts",
    "format": "typescript",
    "contents": [
      { "namespace": "Users", "contents": ["users:*"] },
      { "namespace": "Products", "contents": ["products:*"] }
    ]
  },
  
  // Separate files for external APIs
  {
    "path": "./src/types/external/stripe.ts",
    "format": "typescript",
    "contents": ["stripe:*"]
  },
  {
    "path": "./src/types/external/sendgrid.ts",
    "format": "typescript",
    "contents": ["sendgrid:*"]
  }
]
```

## Cross-Spec References

When types from one spec reference types from another, sparktype handles the references automatically within namespaces.

### Same Output File

If both specs are in the same output with namespaces:

```jsonc
"contents": [
  { "namespace": "Users", "contents": ["users:User"] },
  { "namespace": "Orders", "contents": ["orders:Order"] }
]
```

If `Order` references `User`, the generated code uses:

```typescript
export namespace Orders {
  export interface Order {
    customer: Users.User;  // Cross-namespace reference
  }
}
```

### Different Output Files

For types split across files, you may need to:

1. **Include shared types**: Put common types in both outputs
2. **Use a shared spec**: Create a shared spec for common types
3. **Manual imports**: Add imports in consuming code

**Shared spec approach:**

```jsonc
{
  "specs": {
    "shared": {
      "schemas": {
        "Pagination": { /* ... */ },
        "ErrorResponse": { /* ... */ }
      }
    },
    "users": { "path": "./specs/users.yaml" },
    "products": { "path": "./specs/products.yaml" }
  },
  "outputs": [
    {
      "path": "./src/types/shared.ts",
      "contents": ["shared:*"]
    },
    {
      "path": "./src/types/users.ts",
      "contents": ["users:*"]
    },
    {
      "path": "./src/types/products.ts",
      "contents": ["products:*"]
    }
  ]
}
```

## Filtering Patterns

Use glob patterns to include specific types from each spec:

```jsonc
"contents": [
  // All types from users
  "users:*",
  
  // Only public types from products
  "products:Product",
  "products:Category",
  "products:Price",
  
  // Only request/response types from orders
  "orders:*Request",
  "orders:*Response"
]
```

### Pattern Examples

| Pattern | Matches |
|---------|---------|
| `spec:*` | All types |
| `spec:User` | Exact match |
| `spec:User*` | User, UserProfile, UserSettings |
| `spec:*Request` | CreateUserRequest, UpdateUserRequest |
| `spec:*User*` | User, UserProfile, CreateUserRequest |

## Third-Party API Integration

```jsonc
{
  "specs": {
    // Your API
    "api": { "path": "./openapi.yaml" },
    
    // Third-party APIs
    "stripe": { 
      "url": "https://raw.githubusercontent.com/stripe/openapi/master/openapi/spec3.json"
    },
    "twilio": {
      "url": "https://raw.githubusercontent.com/twilio/twilio-oai/main/spec/json/twilio_api_v2010.json"
    }
  },
  
  "outputs": [
    {
      "path": "./src/types/api.ts",
      "format": "typescript",
      "contents": [
        "api:*",
        
        // Only needed Stripe types
        "stripe:Customer",
        "stripe:PaymentIntent",
        "stripe:Subscription",
        
        // Only needed Twilio types
        "twilio:Message",
        "twilio:Call"
      ]
    }
  ]
}
```

## Best Practices

### 1. Consistent Naming

Use consistent spec names across your project:

```jsonc
// Good - clear, consistent
"specs": {
  "user-service": { /* ... */ },
  "product-service": { /* ... */ }
}

// Avoid - inconsistent
"specs": {
  "users": { /* ... */ },
  "ProductAPI": { /* ... */ }
}
```

### 2. Document Spec Sources

Add comments explaining each spec:

```jsonc
"specs": {
  // Internal user management service
  "users": { "path": "./specs/users.yaml" },
  
  // Partner API - contact partner@example.com for access
  "partner": { "url": "https://api.partner.com/openapi.json" }
}
```

### 3. Version Specs

Include version in spec names if managing multiple versions:

```jsonc
"specs": {
  "api-v1": { "path": "./specs/v1/openapi.yaml" },
  "api-v2": { "path": "./specs/v2/openapi.yaml" }
}
```

## See Also

- [Configuration: Specs](/configuration/specs)
- [Configuration: Contents](/configuration/contents)
- [Remote Specs Guide](./remote-specs)

