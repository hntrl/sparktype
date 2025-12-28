# All Formats Example

Generate TypeScript, Zod, Python, and Go types from a single OpenAPI spec.

## Overview

This example shows how sparktype can output multiple formats simultaneously:

```
all-formats/
├── typegen.jsonc     # Configuration with all formats
└── output/
    ├── types.ts      # TypeScript interfaces
    ├── schemas.ts    # Zod schemas (without types)
    ├── combined.ts   # Zod schemas with inferred types
    ├── types.py      # Python TypedDicts
    └── types.go      # Go structs
```

## Configuration

```jsonc
{
  "specs": {
    "api": { "path": "../basic/openapi.yaml" }
  },
  "outputs": [
    // TypeScript interfaces
    { "path": "./output/types.ts", "format": "typescript" },
    
    // Zod schemas only
    { "path": "./output/schemas.ts", "format": "zod" },
    
    // Zod with inferred TypeScript types
    { 
      "path": "./output/combined.ts", 
      "format": "zod",
      "options": { "inferTypes": true }
    },
    
    // Python TypedDicts
    { "path": "./output/types.py", "format": "python" },
    
    // Go structs
    { 
      "path": "./output/types.go", 
      "format": "go",
      "options": { "package": "api" }
    }
  ]
}
```

## Generated Output

### TypeScript (`types.ts`)

```typescript
export interface User {
  id: string;
  email: string;
  name?: string;
  role: UserRole;
}

export enum UserRole {
  Admin = "admin",
  User = "user",
  Guest = "guest",
}
```

### Zod (`schemas.ts`)

```typescript
import * as z from "zod";

export const userSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email(),
  name: z.string().optional(),
  role: z.lazy(() => userRoleSchema),
});

export const userRoleSchema = z.enum(["admin", "user", "guest"]);
```

### Zod with Types (`combined.ts`)

```typescript
export const userSchema = z.object({ ... });
export type User = z.infer<typeof userSchema>;
```

### Python (`types.py`)

```python
class User(TypedDict):
    id: str
    email: str
    name: NotRequired[str]
    role: "UserRole"

class UserRole(str, Enum):
    ADMIN = "admin"
    USER = "user"
    GUEST = "guest"
```

### Go (`types.go`)

```go
type User struct {
    ID    string   `json:"id"`
    Email string   `json:"email"`
    Name  *string  `json:"name,omitempty"`
    Role  UserRole `json:"role"`
}

type UserRole string

const (
    UserRoleAdmin UserRole = "admin"
    UserRoleUser  UserRole = "user"
    UserRoleGuest UserRole = "guest"
)
```

## Usage

```bash
sparktype generate
```

## When to Use Each Format

| Format | Best For |
|--------|----------|
| `typescript` | Type-only imports, maximum IDE support |
| `zod` | Runtime validation, form validation |
| `python` | FastAPI/Flask backends, mypy type checking |
| `go` | Go microservices, API clients |

## Next Steps

- See [typescript-react](../typescript-react) for using TypeScript + Zod together
- See [python-fastapi](../python-fastapi) for Python backend usage
- See [go-service](../go-service) for Go service usage

