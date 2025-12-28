---
layout: home

hero:
  name: "sparktype ⚡"
  text: "Type-safe OpenAPI"
  tagline: Generate static type definitions from OpenAPI specifications for TypeScript, Python, and Go.
  actions:
    - theme: brand
      text: Get Started
      link: /getting-started
    - theme: alt
      text: View on GitHub
      link: https://github.com/hntrl/sparktype

features:
  - icon: 📦
    title: Multiple Formats
    details: Output schemas as TypeScript interfaces, Zod schemas, Python TypedDicts, or Go structs from a single source of truth.
  - icon: 🔋
    title: CI-Friendly
    details: Detect drift between your OpenAPI specs and generated types with the check command. Perfect for CI/CD pipelines.
  - icon: 🔍
    title: Filtering & Organizing
    details: Include or exclude schemas by glob patterns. Organize output with namespaces to keep your types clean and maintainable.
  - icon: 🌐
    title: Cross-Platform
    details: Available via npm, pip, Homebrew, or direct binary download. Works on macOS, Linux, and Windows.
---

## Quick Start

Install sparktype using your preferred package manager:

::: code-group

```sh [npm]
npm install -D sparktype
```

```sh [pip]
pip install sparktype
```

```sh [Homebrew]
brew install hntrl/tap/sparktype
```

```sh [Go]
go install github.com/hntrl/sparktype/cmd/sparktype@latest
```

:::

Create a `typegen.jsonc` configuration file:

```jsonc
{
  "$schema": "https://hntrl.github.io/sparktype/schema.json",
  "specs": {
    "api": {
      "path": "./openapi.yaml"
    }
  },
  "outputs": [
    {
      "path": "./src/types/api.ts",
      "format": "typescript",
      "contents": ["api:*"]
    }
  ]
}
```

Generate your types:

::: code-group

```sh [CLI]
sparktype generate
```

```json [package.json]
{
  "scripts": {
    "types": "sparktype generate",
    "types:check": "sparktype check"
  }
}
```

```makefile [Makefile]
.PHONY: types
types:
	sparktype generate

.PHONY: types-check
types-check:
	sparktype check
```

```go [go:generate]
//go:generate sparktype generate
package main
```

:::

That's it! Your types are now generated and ready to use. See the [Getting Started](/getting-started) guide for a complete walkthrough.

