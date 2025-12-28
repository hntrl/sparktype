# Monorepo Example

This example demonstrates how to use sparktype in a monorepo to share types across multiple packages and services.

## What This Example Shows

- **Single source of truth** - One OpenAPI spec generates types for all packages
- **Multi-language support** - TypeScript, Go, and Python from the same spec
- **Domain organization** - Types organized by namespace for better structure
- **Cross-package sharing** - Shared types package for frontend use
- **CI integration** - Single `sparktype check` validates all outputs

## Project Structure

```
monorepo/
├── openapi.yaml              # Single source of truth for API types
├── typegen.jsonc             # sparktype configuration
├── package.json              # Root package with workspaces
├── turbo.json                # Turborepo configuration
│
├── packages/
│   ├── frontend/             # React frontend
│   │   ├── src/
│   │   │   ├── types/
│   │   │   │   └── api.ts    # Generated TypeScript types
│   │   │   ├── schemas/
│   │   │   │   └── api.ts    # Generated Zod schemas
│   │   │   └── api/
│   │   │       └── client.ts # API client using types
│   │   └── package.json
│   │
│   └── shared/               # Shared utilities package
│       └── src/
│           ├── index.ts      # Re-exports and utilities
│           └── types/
│               └── domains.ts # Generated domain-namespaced types
│
└── services/
    ├── api/                  # Go API service
    │   ├── cmd/server/
    │   ├── internal/types/
    │   │   └── api.go        # Generated Go types
    │   └── go.mod
    │
    └── worker/               # Python worker service
        └── src/
            ├── types/
            │   └── api.py    # Generated Python types
            └── tasks.py      # Tasks using shared types
```

## Getting Started

### 1. Install dependencies

```bash
npm install
```

### 2. Generate all types

```bash
npm run types
```

This generates types in all configured outputs:
- `packages/frontend/src/types/api.ts` - TypeScript interfaces
- `packages/frontend/src/schemas/api.ts` - Zod schemas
- `packages/shared/src/types/domains.ts` - Namespaced TypeScript types
- `services/api/internal/types/api.go` - Go structs
- `services/worker/src/types/api.py` - Python TypedDicts

### 3. Build all packages

```bash
npm run build
```

## Configuration

The `typegen.jsonc` demonstrates generating multiple outputs from a single spec:

```jsonc
{
  "specs": {
    "api": { "path": "./openapi.yaml" }
  },
  "outputs": [
    // Frontend types
    { "path": "./packages/frontend/src/types/api.ts", "format": "typescript" },
    { "path": "./packages/frontend/src/schemas/api.ts", "format": "zod" },
    
    // Backend services
    { "path": "./services/api/internal/types/api.go", "format": "go" },
    { "path": "./services/worker/src/types/api.py", "format": "python" },
    
    // Domain-organized types for shared package
    {
      "path": "./packages/shared/src/types/domains.ts",
      "format": "typescript",
      "contents": [
        { "namespace": "Users", "contents": ["api:User", "api:UserRole", ...] },
        { "namespace": "Workspaces", "contents": [...] },
        { "namespace": "Documents", "contents": [...] }
      ]
    }
  ]
}
```

## Using Shared Types

### Frontend (TypeScript)

Import from the generated types or the shared package:

```typescript
// Direct import from generated types
import type { User, Document } from "../types/api";
import { userSchema, documentSchema } from "../schemas/api";

// Or from the shared package with domain organization
import { Users, Documents } from "@example/shared";

function renderUser(user: Users.User) {
  // ...
}
```

### API Service (Go)

```go
import "github.com/example/monorepo/services/api/internal/types"

func getUser(id string) (*types.User, error) {
    // types.User matches the TypeScript User type
    return &types.User{
        ID:    id,
        Email: "user@example.com",
        Name:  "Example User",
    }, nil
}
```

### Worker Service (Python)

```python
from .types.api import Document, DocumentStatus

def process_document(doc: Document) -> None:
    if doc["status"] == DocumentStatus.PUBLISHED:
        # Process published document
        pass
```

## Domain Namespacing

For larger projects, organize types by domain using namespaces:

```typescript
// Generated in packages/shared/src/types/domains.ts
export namespace Users {
  export interface User { ... }
  export type UserRole = "admin" | "member" | "guest";
  export interface CreateUserRequest { ... }
}

export namespace Documents {
  export interface Document { ... }
  export type DocumentStatus = "draft" | "published" | "archived";
}
```

This keeps related types together and avoids naming conflicts.

## CI Integration

Add a single check for all generated types:

```yaml
# .github/workflows/ci.yml
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
      - run: npm ci
      
      # Single command validates all outputs
      - name: Check generated types
        run: npm run types:check

  build-frontend:
    needs: validate
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: npm ci
      - run: npm run build -w @example/frontend

  build-api:
    needs: validate
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - run: cd services/api && go build ./...

  build-worker:
    needs: validate
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
      - run: cd services/worker && pip install -e ".[dev]" && mypy src
```

## Workflow

1. **Edit the OpenAPI spec** - Make changes to `openapi.yaml`
2. **Generate types** - Run `npm run types` to update all outputs
3. **Use in code** - Import types in frontend, backend, and workers
4. **Verify in CI** - `sparktype check` ensures types are in sync

## Key Benefits

### Type Safety Across Stack

The same User type definition is used everywhere:
- Frontend form validation (Zod)
- API request/response handling (Go)
- Background job processing (Python)

### Single Source of Truth

One OpenAPI spec defines all types. Changes propagate to all services automatically.

### Consistent API Contracts

Frontend and backend always agree on data shapes because they use the same generated types.

## Turborepo Integration

Use Turborepo for efficient builds:

```json
// turbo.json
{
  "tasks": {
    "build": {
      "dependsOn": ["^build"]
    },
    "types:check": {
      "outputs": []
    }
  }
}
```

Run type checking across all packages:

```bash
turbo run types:check
```

## Next Steps

- Add more services (e.g., notification service)
- Set up shared API clients
- Add end-to-end type tests
- Configure branch-based type generation for feature flags

