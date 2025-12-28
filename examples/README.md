# sparktype Examples

This directory contains examples demonstrating how to use sparktype in different environments and use cases.

## End-to-End Examples

These examples show complete project setups with working code that uses the generated types:

| Example | Description | Languages |
|---------|-------------|-----------|
| [typescript-react](./typescript-react) | React app with type-safe API client and Zod validation | TypeScript |
| [python-fastapi](./python-fastapi) | FastAPI backend with TypedDict request/response handling | Python |
| [go-service](./go-service) | Go HTTP service with Chi router and generated structs | Go |
| [monorepo](./monorepo) | Multi-package/service setup sharing types across stack | TypeScript, Go, Python |

## Configuration Examples

These examples demonstrate specific sparktype configuration patterns:

| Example | Description |
|---------|-------------|
| [basic](./basic) | Minimal setup with a single spec and output |
| [all-formats](./all-formats) | Generating all supported formats from one spec |
| [inline-schemas](./inline-schemas) | Using inline schema definitions and namespaces |
| [multi-spec](./multi-spec) | Working with multiple OpenAPI specs |

## Quick Start

Each example can be run independently. Start with the example that matches your stack:

### TypeScript/React

```bash
cd typescript-react
npm install
npm run types    # Generate types
npm run dev      # Start dev server
```

### Python/FastAPI

```bash
cd python-fastapi
python -m venv venv && source venv/bin/activate
pip install -e ".[dev]"
sparktype generate   # Generate types
uvicorn app.main:app --reload
```

### Go

```bash
cd go-service
sparktype generate   # Generate types
go run ./cmd/server
```

### Monorepo

```bash
cd monorepo
npm install
npm run types    # Generate types for all packages
npm run build    # Build all packages
```

## What Each Example Demonstrates

### typescript-react

- **API Client**: Type-safe fetch wrapper using generated types
- **Runtime Validation**: Zod schemas to validate API responses
- **React Hooks**: Custom hooks with typed state and returns
- **Form Handling**: Type-safe forms using request types
- **Enum Usage**: Category filtering with generated enums

### python-fastapi

- **TypedDict Usage**: Static type hints for request/response bodies
- **Pydantic Integration**: Using TypedDicts alongside Pydantic models
- **Enum Types**: Type-safe status values with Python enums
- **Database Layer**: Types flowing through your data layer
- **mypy Compatibility**: Full type checking support

### go-service

- **Go Structs**: Generated types with JSON tags
- **HTTP Handlers**: Request/response handling with type safety
- **Enum Handling**: String constants for enum values
- **Error Responses**: Type-safe error structures
- **Optional Fields**: Pointer types for nullable fields

### monorepo

- **Multi-Output**: Generate types for multiple packages at once
- **Cross-Language**: Same spec generates TypeScript, Go, Python
- **Domain Namespacing**: Organize types by domain/feature
- **Shared Types**: Types package consumed by multiple apps
- **CI Integration**: Single command validates all outputs

## CI Integration

All examples follow the same CI pattern:

```yaml
# .github/workflows/ci.yml
- name: Check generated types
  run: sparktype check  # or: npm run types:check

# This ensures generated files are committed and up-to-date
```

See the [CI/CD Guide](https://hntrl.github.io/sparktype/guides/ci-cd) for platform-specific examples.

## Learning Path

1. **New to sparktype?** Start with [basic](./basic) to understand the config format
2. **Building a frontend?** Check out [typescript-react](./typescript-react)
3. **Building a backend?** See [python-fastapi](./python-fastapi) or [go-service](./go-service)
4. **Complex project?** Look at [monorepo](./monorepo) for multi-package setups
5. **Advanced config?** See [multi-spec](./multi-spec) and [inline-schemas](./inline-schemas)

## Related Documentation

- [Getting Started](https://hntrl.github.io/sparktype/getting-started)
- [Configuration](https://hntrl.github.io/sparktype/configuration/)
- [Output Formats](https://hntrl.github.io/sparktype/formats/)
- [CI/CD Integration](https://hntrl.github.io/sparktype/guides/ci-cd)

