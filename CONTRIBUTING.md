# Contributing to sparktype

Thanks for taking the time to contribute to sparktype!

When it comes to open source, there are many different ways to contribute, all of which are valuable. Here are a few guidelines that will help you contribute to sparktype.

## Did you find a bug?

- Make sure the bug wasn't already reported by searching on GitHub under [Issues](https://github.com/hntrl/sparktype/issues).
- If you're unable to find an open issue addressing the problem, [open a new one](https://github.com/hntrl/sparktype/issues/new). Be sure to include a title and clear description, as much relevant information as possible, and a code sample if applicable. If possible, use the relevant issue templates to create the issue.
- If you've encountered a security issue, please see our [SECURITY.md](SECURITY.md) guide for info on how to report it.

## Proposing new or changing existing features?

Please provide thoughtful comments and some sample code that show what you'd like to do with sparktype. It helps the conversation if you can show us how you're limited by the current API first before jumping to a conclusion about what needs to be changed and/or added.

## Issue not getting attention?

If you need a bug fixed and nobody is fixing it, your best bet is to provide a fix for it and make a [pull request](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/creating-a-pull-request). Open source code belongs to all of us, and it's all of our responsibility to push it forward.

## Making a Pull Request?

When creating the PR in GitHub, make sure that you set the base to the correct branch. You set the base in GitHub when authoring the PR with the dropdown below the "Compare changes" heading.

---

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```sh
   git clone https://github.com/YOUR_USERNAME/sparktype.git
   cd sparktype
   ```
3. **Add the upstream remote**:
   ```sh
   git remote add upstream https://github.com/hntrl/sparktype.git
   ```

## Development Setup

### Prerequisites

- **Go 1.22+** - [Installation guide](https://go.dev/doc/install)
- **Make** - Usually pre-installed on macOS/Linux

### Building

```sh
# Build the binary
make build

# Or install directly to $GOPATH/bin
make install
```

### Running

```sh
# Run via make
make run ARGS="generate --config typegen.jsonc"

# Or after building
./bin/sparktype generate --config typegen.jsonc
```

## Making Changes

### Branch Naming

Use descriptive branch names:
- `feat/add-rust-generator` - New features
- `fix/typescript-nullable-types` - Bug fixes
- `docs/update-getting-started` - Documentation changes
- `refactor/parser-cleanup` - Code refactoring
- `test/add-golang-tests` - Test additions

### Code Style

- Run `make fmt` before committing to format your code
- Run `make lint` to check for common issues
- Follow existing patterns and conventions in the codebase
- Add comments for complex logic
- Keep functions focused and reasonably sized

### Commit Messages

Write clear, descriptive commit messages:

```
feat: add support for Rust code generation

- Add rust generator package
- Implement struct generation
- Add comprehensive test coverage
```

Follow these conventions:
- Use present tense ("add feature" not "added feature")
- Use imperative mood ("move cursor to..." not "moves cursor to...")
- Start with a type: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`
- Keep the first line under 72 characters
- Add detailed description in the body if needed

## Testing

### Running Tests

```sh
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./internal/generators/typescript/...

# Run tests with race detection
go test -race ./...
```

### Golden Tests

The generators use golden file testing. If you're adding new features or fixing bugs in generators:

1. Run tests to see failures:
   ```sh
   go test -v ./internal/generators/typescript/...
   ```

2. Update golden files if the changes are intentional:
   ```sh
   go test ./internal/generators/typescript/... -update
   ```

3. Review the updated golden files before committing

### Writing Tests

- Add tests for new functionality
- Update tests when changing behavior
- Use table-driven tests where appropriate
- Test edge cases and error conditions

## Submitting a Pull Request

### Before Submitting

1. **Sync with upstream**:
   ```sh
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run the full test suite**:
   ```sh
   make test
   make lint
   ```

3. **Build and verify**:
   ```sh
   make build
   ./bin/sparktype --help
   ```

### PR Guidelines

1. **Create a focused PR** - One feature or fix per PR
2. **Fill out the PR template** - Provide context and testing notes
3. **Link related issues** - Use "Fixes #123" or "Relates to #456"
4. **Keep PRs reasonably sized** - Large PRs are harder to review
5. **Respond to feedback** - Address reviewer comments promptly

## CI Workflows

When you open a PR, several checks run automatically:

| Workflow | What it does |
|----------|--------------|
| **Test** | Runs tests on Linux, macOS, Windows with Go 1.21 & 1.22 |
| **Lint** | Runs golangci-lint to check code style |
| **Build** | Builds the binary and verifies it runs |

All checks must pass before a PR can be merged.

## Project Structure

```
sparktype/
├── cmd/sparktype/         # CLI entry point
├── internal/
│   ├── cli/               # CLI commands (generate, check, etc.)
│   ├── config/            # Configuration loading and validation
│   ├── contents/          # Content tree management
│   ├── generators/        # Code generators (typescript, go, python, zod)
│   ├── parser/            # Pattern parsing
│   └── spec/              # OpenAPI spec loading
├── distributions/         # Package distribution files (npm, pypi, homebrew)
├── docs/                  # Documentation (VitePress)
├── examples/              # Example configurations
└── schema/                # JSON schema for configuration
```

## Need Help?

- Check the [documentation](https://hntrl.github.io/sparktype)
- Search [existing issues](https://github.com/hntrl/sparktype/issues)
- Open a [new discussion](https://github.com/hntrl/sparktype/discussions) for questions

---

This project is a volunteer effort, and we encourage you to contribute in any way you can.

Thanks! ❤️ ❤️ ❤️
