# TypeScript + React Example

This example demonstrates how to use sparktype in a React application with TypeScript for end-to-end type safety.

## What This Example Shows

- **Type-safe API client** - Fetch wrapper using generated TypeScript types
- **Runtime validation** - Zod schemas to validate API responses
- **React hooks** - Custom hooks with typed state and return values
- **Form handling** - Type-safe form state using generated request types
- **Enum usage** - Category filtering using generated enum types

## Project Structure

```
typescript-react/
├── openapi.yaml          # E-commerce API spec
├── typegen.jsonc         # sparktype configuration
├── package.json
├── src/
│   ├── types/
│   │   └── api.ts        # Generated TypeScript interfaces
│   ├── schemas/
│   │   └── api.ts        # Generated Zod schemas
│   ├── api/
│   │   └── client.ts     # Type-safe API client
│   ├── hooks/
│   │   ├── useProducts.ts
│   │   └── useCart.ts
│   └── components/
│       ├── ProductCard.tsx
│       ├── ProductList.tsx
│       └── CheckoutForm.tsx
```

## Getting Started

### 1. Install dependencies

```bash
npm install
```

### 2. Generate types

```bash
npm run types
```

This generates two files:
- `src/types/api.ts` - TypeScript interfaces for type-only imports
- `src/schemas/api.ts` - Zod schemas for runtime validation

### 3. Start development

```bash
npm run dev
```

## How It Works

### Generated TypeScript Types

The generated types provide compile-time type checking:

```typescript
import type { Product, ProductCategory } from "../types/api";

// TypeScript knows all properties and their types
function formatProduct(product: Product): string {
  return `${product.name} - ${product.price.formatted}`;
}
```

### Generated Zod Schemas

The Zod schemas validate data at runtime, catching API contract violations:

```typescript
import { productSchema } from "../schemas/api";

const response = await fetch("/api/products/123");
const data = await response.json();

// Throws if the response doesn't match the expected schema
const product = productSchema.parse(data);
```

### Type-Safe API Client

The API client combines both for maximum safety:

```typescript
// Request body is type-checked at compile time
await createProduct({
  name: "Widget",
  price: { amount: 999, currency: "USD" },
  category: ProductCategory.Electronics,
});

// Response is validated at runtime
const product = productSchema.parse(await response.json());
```

### React Hooks with Full Type Inference

Custom hooks return properly typed data:

```typescript
const { products, loading, error } = useProducts({
  category: ProductCategory.Electronics,
});

// products is typed as Product[]
products.map(p => p.name); // ✓ Type safe
products.map(p => p.foo);  // ✗ Type error
```

## CI Integration

Add type checking to your CI pipeline:

```yaml
# .github/workflows/ci.yml
- name: Check generated types
  run: npm run types:check

- name: TypeScript check
  run: npm run typecheck
```

## Key Patterns Demonstrated

### 1. Dual Output Strategy

Generate both TypeScript types and Zod schemas:
- **TypeScript** (`format: typescript`) - Zero runtime overhead, type-only imports
- **Zod** (`format: zod`) - Runtime validation for API responses

### 2. Error Handling

Use the generated `ApiError` type for consistent error handling:

```typescript
import type { ApiError } from "../types/api";
import { apiErrorSchema } from "../schemas/api";

if (!response.ok) {
  const error: ApiError = apiErrorSchema.parse(await response.json());
  throw new ApiRequestError(response.status, error);
}
```

### 3. Form Validation

Use request types for form state and Zod for validation:

```typescript
import type { CreateOrderRequest } from "../types/api";
import { createOrderRequestSchema } from "../schemas/api";

const formData: CreateOrderRequest = { /* ... */ };
const result = createOrderRequestSchema.safeParse(formData);
```

## Next Steps

- Add authentication (see the [remote specs guide](https://hntrl.github.io/sparktype/guides/remote-specs))
- Set up pre-commit hooks with `sparktype check`
- Add API mocking for tests using the generated schemas

