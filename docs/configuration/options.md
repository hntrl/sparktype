# Options

sparktype supports both global options (applied to all outputs) and format-specific options (per output).

## Global Options

Set in the top-level `options` object:

```jsonc
{
  "specs": { /* ... */ },
  "outputs": [ /* ... */ ],
  "options": {
    "dereferenceRefs": false,
    "generateEnums": true,
    "nullableHandling": "optional"
  }
}
```

### dereferenceRefs

**Type:** `boolean`  
**Default:** `false`

When `true`, `$ref` references are inlined instead of kept as type references.

```jsonc
"options": {
  "dereferenceRefs": true
}
```

**Example with `dereferenceRefs: false` (default):**

```typescript
export interface User {
  address: Address;  // Reference to Address type
}
```

**Example with `dereferenceRefs: true`:**

```typescript
export interface User {
  address: {         // Inlined
    street: string;
    city: string;
  };
}
```

Use dereferencing when:
- You want fully self-contained types
- The target language doesn't support references well
- You're generating documentation

### generateEnums

**Type:** `boolean`  
**Default:** `false`

When `true`, string enums produce native enum types instead of union types.

```jsonc
"options": {
  "generateEnums": true
}
```

**Example with `generateEnums: false` (default):**

```typescript
export type Status = "active" | "inactive" | "pending";
```

**Example with `generateEnums: true`:**

```typescript
export enum Status {
  Active = "active",
  Inactive = "inactive",
  Pending = "pending"
}
```

::: tip
Union types are generally preferred in TypeScript for better tree-shaking and type inference. Use native enums when you need runtime enum features or for consistency with other languages.
:::

### nullableHandling

**Type:** `"optional" | "nullable" | "both"`  
**Default:** varies by format

Controls how nullable fields are represented in generated output.

| Value | Description |
|-------|-------------|
| `optional` | Nullable fields become optional (`field?: Type`) |
| `nullable` | Nullable fields use null union (`field: Type \| null`) |
| `both` | Fields are both optional and nullable (`field?: Type \| null`) |

```jsonc
"options": {
  "nullableHandling": "nullable"
}
```

**Example with `nullableHandling: "optional"`:**

```typescript
export interface User {
  name?: string;  // Optional
}
```

**Example with `nullableHandling: "nullable"`:**

```typescript
export interface User {
  name: string | null;  // Nullable
}
```

**Example with `nullableHandling: "both"`:**

```typescript
export interface User {
  name?: string | null;  // Optional and nullable
}
```

## Format-Specific Options

Set per-output in the `options` object:

```jsonc
"outputs": [
  {
    "path": "./types/api.ts",
    "format": "typescript",
    "contents": ["api:*"],
    "options": {
      "exportType": "interface"
    }
  }
]
```

Each format supports different options. See the format documentation for details:

- [TypeScript Options](/formats/typescript#options) - `exportType`, `readonlyProperties`
- [Zod Options](/formats/zod#options) - `inferTypes`
- [Python Options](/formats/python#options) - (global options only)
- [Go Options](/formats/go#options) - `package`

## Option Precedence

Output-specific options override global options:

```jsonc
{
  "options": {
    "generateEnums": true  // Global default
  },
  "outputs": [
    {
      "path": "./types/with-enums.ts",
      "format": "typescript",
      "contents": ["api:*"]
      // Uses generateEnums: true from global
    },
    {
      "path": "./types/with-unions.ts",
      "format": "typescript",
      "contents": ["api:*"],
      "options": {
        "generateEnums": false  // Overrides global
      }
    }
  ]
}
```

## Quick Reference

### Global Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `dereferenceRefs` | `boolean` | `false` | Inline `$ref` references |
| `generateEnums` | `boolean` | `false` | Use native enums for string enums |
| `nullableHandling` | `string` | varies | How to handle nullable fields |

### Format-Specific Options

| Format | Options |
|--------|---------|
| [TypeScript](/formats/typescript#options) | `exportType`, `readonlyProperties` |
| [Zod](/formats/zod#options) | `inferTypes` |
| [Python](/formats/python) | (global options only) |
| [Go](/formats/go#options) | `package` |

