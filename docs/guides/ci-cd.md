# CI/CD Integration

This guide covers how to integrate sparktype into your CI/CD pipeline to ensure generated types stay in sync with your OpenAPI specifications.

## Overview

The recommended workflow:

1. Developers run `sparktype generate` locally
2. Generated files are committed to version control
3. CI runs `sparktype check` to verify files are in sync
4. If check fails, the pipeline fails

This ensures:
- Generated code is reviewed in PRs
- Specs and types never drift
- No auto-generation magic in CI

## The Check Command

`sparktype check` compares existing generated files against what would be generated:

```sh
sparktype check
```

- **Exit code 0**: All files match
- **Exit code 1**: Files are out of sync or missing

## CI Platforms

::: details GitHub Actions

**Basic workflow:**

```yaml
# .github/workflows/check-types.yml
name: Check Types

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  check-types:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          
      - run: npm ci
      
      - name: Check generated types
        run: npx sparktype check
```

**With matrix strategy** (multiple config files):

```yaml
jobs:
  check-types:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        config:
          - ./frontend/typegen.jsonc
          - ./backend/typegen.jsonc
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - run: npm ci
      - run: npx sparktype check --config ${{ matrix.config }}
```

:::

::: details GitLab CI

```yaml
# .gitlab-ci.yml
stages:
  - validate
  - test

check-types:
  stage: validate
  image: node:20
  cache:
    paths:
      - node_modules/
  script:
    - npm ci
    - npx sparktype check
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
    - if: $CI_COMMIT_BRANCH == "main"

test:
  stage: test
  needs: [check-types]
  image: node:20
  script:
    - npm ci
    - npm test
```

:::

::: details CircleCI

```yaml
# .circleci/config.yml
version: 2.1

jobs:
  check-types:
    docker:
      - image: cimg/node:20.0
    steps:
      - checkout
      - restore_cache:
          keys:
            - npm-deps-{{ checksum "package-lock.json" }}
      - run: npm ci
      - save_cache:
          key: npm-deps-{{ checksum "package-lock.json" }}
          paths:
            - node_modules
      - run: npx sparktype check

workflows:
  main:
    jobs:
      - check-types
```

:::

::: details Azure Pipelines

```yaml
# azure-pipelines.yml
trigger:
  - main

pr:
  - main

pool:
  vmImage: 'ubuntu-latest'

steps:
  - task: NodeTool@0
    inputs:
      versionSpec: '20.x'
    displayName: 'Install Node.js'

  - script: npm ci
    displayName: 'Install dependencies'

  - script: npx sparktype check
    displayName: 'Check generated types'
```

:::

::: details Jenkins

```groovy
// Jenkinsfile
pipeline {
    agent {
        docker {
            image 'node:20'
        }
    }
    stages {
        stage('Install') {
            steps {
                sh 'npm ci'
            }
        }
        stage('Check Types') {
            steps {
                sh 'npx sparktype check'
            }
        }
        stage('Test') {
            steps {
                sh 'npm test'
            }
        }
    }
}
```

:::

## Pre-commit Hooks

::: details Husky (Node.js)

**Setup:**

```sh
npm install -D husky
npx husky init
```

**Pre-commit hook:**

```sh
# .husky/pre-commit
npx sparktype check
```

**Pre-push hook** (for faster commits):

```sh
# .husky/pre-push
npx sparktype check
```

:::

::: details pre-commit (Python)

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: sparktype-check
        name: Check generated types
        entry: sparktype check
        language: system
        pass_filenames: false
```

:::

## Handling Remote Specs

When using remote specs with authentication, pass secrets as environment variables:

::: details GitHub Actions

```yaml
- name: Check generated types
  env:
    API_TOKEN: ${{ secrets.API_TOKEN }}
  run: npx sparktype check
```

:::

::: details GitLab CI

```yaml
check-types:
  variables:
    API_TOKEN: $API_TOKEN  # Set in GitLab CI/CD settings
  script:
    - npx sparktype check
```

:::

::: details Config Example

```jsonc
{
  "specs": {
    "external": {
      "url": "https://api.partner.com/openapi.json",
      "headers": {
        "Authorization": "Bearer ${API_TOKEN}"
      }
    }
  }
}
```

:::

## Monorepo Setup

For monorepos with multiple packages:

::: details Turborepo

```json
// turbo.json
{
  "pipeline": {
    "check-types": {
      "dependsOn": ["^build"],
      "outputs": []
    }
  }
}
```

```sh
turbo run check-types
```

:::

::: details Nx

```json
// project.json
{
  "targets": {
    "check-types": {
      "executor": "nx:run-commands",
      "options": {
        "command": "npx sparktype check"
      }
    }
  }
}
```

:::

::: details GitHub Actions (matrix strategy)

```yaml
jobs:
  check-types:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        package: [frontend, backend, shared]
    defaults:
      run:
        working-directory: packages/${{ matrix.package }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
      - run: npm ci
      - run: npx sparktype check
```

:::

## Troubleshooting

### Check Passes Locally But Fails in CI

1. **Different versions**: Ensure sparktype version matches
   ```yaml
   - run: npx sparktype@1.0.0 check
   ```

2. **Line endings**: Configure Git in CI
   ```yaml
   - run: git config --global core.autocrlf input
   ```

3. **Missing dependencies**: Ensure all packages are installed

### Remote Spec Timeouts

Add retries for flaky remote specs:

```yaml
- name: Check generated types
  run: npx sparktype check
  timeout-minutes: 5
  continue-on-error: false
```

### Caching

Speed up CI with caching:

```yaml
- uses: actions/cache@v4
  with:
    path: ~/.npm
    key: npm-${{ hashFiles('**/package-lock.json') }}
```

## Best Practices

### 1. Fail Fast

Run type checks early in the pipeline before expensive operations.

### 2. Don't Auto-Generate in CI

Avoid this anti-pattern:

```yaml
# DON'T DO THIS
- run: npx sparktype generate
- run: git add . && git commit -m "Auto-update types"
```

This hides spec changes from code review.

### 3. Clear Error Messages

When check fails, the output shows what changed:

```
Checking ./src/types/api.ts... MISMATCH
  User:
    + email: string (added)

1 file(s) out of sync. Run 'sparktype generate' to update.
```

Include this in your CI output for easy debugging.

### 4. Document the Workflow

Add to your README:

```markdown
## Updating API Types

1. Update the OpenAPI spec
2. Run `npm run types` (sparktype generate)
3. Review changes
4. Commit both spec and generated files
```

## See Also

- [check command](/cli/check)
- [Remote Specs Guide](./remote-specs)
- [Configuration](/configuration/)

