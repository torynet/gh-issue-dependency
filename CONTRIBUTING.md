# Contributing to gh-issue-dependency

Thank you for your interest in contributing to `gh-issue-dependency`! This guide will help you get started with development and understand our contribution process.

## Table of Contents

- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Code Style Guidelines](#code-style-guidelines)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Release Process](#release-process)

## Development Setup

### Prerequisites

- **Go 1.19 or later**: [Install Go](https://golang.org/doc/install)
- **GitHub CLI**: [Install GitHub CLI](https://cli.github.com/) 
- **Git**: For version control
- **Make** (optional): For using the Makefile commands

### Clone and Setup

```bash
# Fork the repository on GitHub first, then:
git clone https://github.com/YOUR_USERNAME/gh-issue-dependency
cd gh-issue-dependency

# Install dependencies
go mod download

# Verify setup by building
go build -o gh-issue-dependency

# Run tests to ensure everything works
go test ./...
```

### Quick Development Testing

For fast, repeatable testing during development:

```bash
# Build and test the current version
go run main.go --help              # Test help system
go run main.go --version           # Test version display
go run main.go list --help         # Test command help

# Test with real repository (requires auth)
go run main.go list 123             # Test list command (will show "not implemented")
go run main.go add 123 --blocks 124 # Test add command (will show "not implemented") 
go run main.go remove 123 --blocks 124 # Test remove command (will show "not implemented")

# Quick build and install for testing
go build -o /tmp/gh-issue-dependency && /tmp/gh-issue-dependency --help

# Test error handling
go run main.go invalid-command      # Should show error and help
go run main.go list                 # Should show "missing issue number"
go run main.go add 123              # Should show "missing relationship flag"
```

**Note**: Commands will show "not implemented yet" until the actual dependency management logic is built. This tests the CLI framework, argument parsing, help system, and error handling.

### Development Dependencies

```bash
# Install additional development tools
go install golang.org/x/tools/cmd/goimports@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
```

## Project Structure

```
gh-issue-dependency/
├── main.go                      # Application entry point
├── cmd/                         # Command implementations (Cobra CLI)
│   ├── root.go                  # Root command and global flags
│   ├── list.go                  # List dependencies command
│   ├── add.go                   # Add dependency command
│   ├── remove.go                # Remove dependency command
│   └── *_test.go                # Command tests
├── pkg/                         # Shared utilities and libraries
│   ├── github.go                # GitHub API client and operations
│   ├── errors.go                # Error handling and formatting
│   └── *_test.go                # Package tests
├── tests/                       # Integration and end-to-end tests
│   ├── integration_test.sh      # Shell-based integration tests
│   └── fixtures/                # Test data and fixtures
├── .github/                     # GitHub workflows and templates
│   └── workflows/               # CI/CD workflows
├── docs/                        # Additional documentation
├── Makefile                     # Build and development commands
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── README.md                    # User documentation
└── CONTRIBUTING.md              # This file
```

### Key Components

- **`main.go`**: Entry point that calls the root command
- **`cmd/`**: Contains all CLI command implementations using Cobra
- **`pkg/github.go`**: GitHub API integration and repository operations
- **`pkg/errors.go`**: Structured error handling with user-friendly messages
- **`tests/`**: Integration tests that verify end-to-end functionality

## Development Workflow

### Making Changes

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the code style guidelines

3. **Add tests** for new functionality:
   ```bash
   # Add unit tests in the same package
   # Add integration tests in tests/ directory
   ```

4. **Test your changes**:
   ```bash
   # Run unit tests
   go test ./...
   
   # Run integration tests
   ./tests/integration_test.sh
   
   # Test manually with your GitHub account
   go run . list 123 --repo your/test-repo
   ```

5. **Update documentation** if needed:
   - Update help text in command files
   - Update README.md for user-facing changes
   - Update this CONTRIBUTING.md for developer-facing changes

### Using the Makefile

The project includes a Makefile with common development tasks:

```bash
# Build the binary
make build

# Run all tests
make test

# Run linting and formatting
make lint

# Clean build artifacts
make clean

# Install the binary locally
make install
```

## Code Style Guidelines

### Go Style

We follow standard Go conventions:

- **Formatting**: Use `gofmt` (or `goimports`) for automatic formatting
- **Naming**: Use Go naming conventions (PascalCase for exported, camelCase for unexported)
- **Comments**: Document all exported functions and types
- **Error handling**: Always handle errors explicitly; use structured errors from `pkg/errors.go`

### Example Code Style

```go
// Good: Properly documented exported function
// AddDependency creates a dependency relationship between two issues.
// It returns an error if the issues don't exist or the user lacks permissions.
func AddDependency(issueNum int, blockedBy []int) error {
    if issueNum <= 0 {
        return pkg.NewAppError(
            pkg.ErrorTypeValidation,
            "issue number must be positive",
            nil,
        ).WithContext("issue", issueNum)
    }
    
    // Implementation...
    return nil
}

// Good: Use structured error handling
func validateIssueNumber(num string) (int, error) {
    issueNum, err := strconv.Atoi(num)
    if err != nil {
        return 0, pkg.NewAppError(
            pkg.ErrorTypeValidation,
            fmt.Sprintf("invalid issue number: %s", num),
            err,
        ).WithSuggestion("Issue numbers must be positive integers")
    }
    return issueNum, nil
}
```

### Command Implementation Guidelines

When adding new commands:

1. **Follow Cobra patterns**: Use the same structure as existing commands
2. **Comprehensive help**: Include detailed help text with examples
3. **Input validation**: Validate all inputs before making API calls
4. **Error handling**: Use structured errors with helpful suggestions
5. **Testing**: Add both unit and integration tests

### Help Text Standards

Follow GitHub CLI conventions for help text:

```go
Long: `Command description with proper formatting.

SECTION HEADERS IN CAPS
  Details about the section with proper indentation
  • Use bullets for lists
  • Be specific and actionable

FLAGS
  --flag string   Description with default value shown

EXAMPLES
  # Comment explaining the example
  gh issue-dependency command arg --flag value`,
```

## Testing

### Unit Tests

Write unit tests for all public functions:

```bash
# Run unit tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

Integration tests verify the complete functionality:

```bash
# Run integration tests (requires GitHub authentication)
./tests/integration_test.sh

# Run specific test scenarios
./tests/integration_test.sh test_list_command
```

### Test Guidelines

1. **Unit tests**: Test individual functions in isolation
2. **Integration tests**: Test complete user workflows
3. **Error cases**: Test error handling and validation
4. **Mock when appropriate**: Mock GitHub API for unit tests
5. **Use real API sparingly**: Integration tests can use real API with test repositories

### Test Repository Setup

For testing, you can use a dedicated test repository:

```bash
# Create a test repository for development
gh repo create gh-issue-dependency-test --private
cd gh-issue-dependency-test

# Create some test issues
gh issue create --title "Test issue 1" --body "For testing dependencies"
gh issue create --title "Test issue 2" --body "For testing dependencies"
```

## Submitting Changes

### Pull Request Process

1. **Ensure all tests pass**:
   ```bash
   go test ./...
   ./tests/integration_test.sh
   ```

2. **Update documentation** as needed

3. **Commit with clear messages**:
   ```bash
   git commit -m "feat: add new dependency validation

   - Add cycle detection for dependency chains
   - Improve error messages for invalid relationships
   - Add comprehensive tests for validation logic
   
   Fixes #123"
   ```

4. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

5. **Create a pull request** with:
   - Clear title and description
   - Reference any related issues
   - Include testing instructions
   - Add screenshots for UI changes (if applicable)

### Pull Request Template

When creating a PR, include:

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix (non-breaking change)
- [ ] New feature (non-breaking change)
- [ ] Breaking change (fix or feature causing existing functionality to change)
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added for new functionality
```

### Code Review Process

1. **Automated checks**: CI must pass (tests, linting, build)
2. **Peer review**: At least one maintainer approval required
3. **Manual testing**: Reviewer tests the changes locally
4. **Documentation review**: Ensure docs are accurate and complete

## Release Process

### Versioning

We use [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features (backwards compatible)
- **PATCH**: Bug fixes (backwards compatible)

### Release Steps

1. **Update version** in relevant files
2. **Update CHANGELOG.md** with changes since last release
3. **Create git tag**: `git tag v1.2.3`
4. **Push tag**: `git push origin v1.2.3`
5. **GitHub Actions** will automatically build and release

### Pre-release Testing

Before major releases:

1. Test with multiple GitHub repositories
2. Test cross-repository functionality
3. Verify all examples in documentation work
4. Test installation methods

## Getting Help

### Communication Channels

- **Issues**: [GitHub Issues](https://github.com/torynet/gh-issue-dependency/issues) for bugs and feature requests
- **Discussions**: [GitHub Discussions](https://github.com/torynet/gh-issue-dependency/discussions) for questions and ideas
- **Pull Requests**: For code review and technical discussion

### Development Questions

When asking for help:

1. **Search existing issues** first
2. **Provide context**: What are you trying to accomplish?
3. **Include details**: OS, Go version, error messages
4. **Share code**: Minimal reproducible example when possible

### Reporting Bugs

Include in bug reports:

- **Steps to reproduce**
- **Expected behavior**
- **Actual behavior**  
- **Environment details** (OS, Go version, etc.)
- **Error messages** (full output)
- **Test repository** (if applicable)

## Recognition

Contributors will be:

- Listed in the project's contributors section
- Credited in release notes for their contributions
- Invited to join the maintainer team for significant ongoing contributions

Thank you for contributing to `gh-issue-dependency`!