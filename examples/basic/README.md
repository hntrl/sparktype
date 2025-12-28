# Basic Example

The simplest sparktype configuration - a single OpenAPI spec generating TypeScript types.

## Overview

This example demonstrates the minimal setup needed to use sparktype:

```
basic/
├── openapi.yaml      # OpenAPI specification
├── typegen.jsonc     # sparktype configuration
└── types/
    └── api.ts        # Generated TypeScript types
```

## Configuration

```jsonc
{
  "$schema": "https://hntrl.github.io/sparktype/schema.json",
  "specs": {
    "api": { "path": "./openapi.yaml" }
  },
  "outputs": [
    {
      "path": "./types/api.ts",
      "format": "typescript",
      "contents": ["api:*"]
    }
  ],
  "options": {
    "dereferenceRefs": true,
    "generateEnums": true,
    "nullableHandling": "optional"
  }
}
```

## Usage

Generate types:

```bash
sparktype generate
```

Check types are in sync:

```bash
sparktype check
```

## Key Concepts

- **specs**: Named map of OpenAPI sources (here just one called "api")
- **outputs**: Array of output file configurations
- **contents**: `["api:*"]` includes all schemas from the "api" spec
- **options**: Global settings affecting all outputs

## What Gets Generated

From the OpenAPI schemas like `User`, `Address`, `UserRole`, sparktype generates:

```typescript
export interface User {
  id: string;
  email: string;
  name?: string;
  role: UserRole;
  address?: Address;
}

export enum UserRole {
  Admin = "admin",
  User = "user",
  Guest = "guest",
}
```

## Next Steps

- See [all-formats](../all-formats) to generate Python, Go, and Zod outputs
- See [typescript-react](../typescript-react) for a full React application example

