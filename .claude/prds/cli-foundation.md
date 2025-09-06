---
name: cli-foundation
description: Core CLI framework and infrastructure for the gh-issue-dependency GitHub extension following gh-sub-issue patterns
status: backlog
created: 2025-09-06T20:57:03Z
---

# PRD: cli-foundation

## Executive Summary

The `cli-foundation` provides the essential framework and infrastructure for building the gh-issue-dependency GitHub CLI extension. By replicating the proven patterns from gh-sub-issue, it establishes a reliable, extensible foundation that handles argument parsing, error management, repository detection, and command routing using Go and the Cobra CLI framework.

## Problem Statement

### What problem are we solving?

Building a GitHub CLI extension requires foundational components for argument parsing, command routing, error handling, and GitHub API integration. Rather than building from scratch, we need a proven foundation that follows established patterns and ensures reliability from day one.

### Why is this important now?

- GitHub CLI extensions need consistent, reliable foundations to meet user expectations
- gh-sub-issue demonstrates excellent CLI patterns that should be replicated, not reinvented
- All dependency management commands (list, add, remove) depend on this foundational infrastructure
- Early establishment of solid foundations prevents architectural debt and refactoring later
- CLI users expect consistent behavior and error handling across GitHub CLI extensions

## User Stories

### Primary User Personas

1. **CLI Extension User** - Expects consistent, reliable behavior following gh CLI conventions
2. **Developer** - Needs robust foundation to build dependency management commands upon
3. **Future Contributors** - Requires clear, extensible architecture for adding new features

### Detailed User Journeys

#### Story 1: Consistent CLI Experience

```text
As a GitHub CLI user familiar with gh extensions
I want gh-issue-dependency to follow standard gh CLI patterns
So that I can use it intuitively without learning new conventions

Acceptance Criteria:
- Help output follows standard gh format and conventions
- Flags and arguments follow established gh CLI patterns
- Error messages are clear and actionable like other gh extensions
- Exit codes follow standard conventions (0 success, 1 error)
```

#### Story 2: Reliable Error Handling

```text
As a developer using CLI tools in scripts and automation
I want clear, consistent error handling and messaging
So that I can reliably detect and respond to different failure scenarios

Acceptance Criteria:
- Distinguishes between user errors, permission errors, and system errors
- Provides actionable guidance for fixing common errors
- Uses consistent error message format across all commands
- Handles network failures and API rate limits gracefully
```

#### Story 3: Repository Context Detection

```text
As a developer working in GitHub repositories
I want the extension to automatically detect my current repository context
So that I don't have to manually specify repository information

Acceptance Criteria:
- Automatically detects current repository using `gh repo view --json`
- Supports explicit repository specification with --repo flag
- Handles cases where no repository context is available
- Works from any subdirectory within a repository
```

### Pain Points Being Addressed

- **Inconsistent UX**: Eliminates risk of creating non-standard CLI interfaces
- **Unreliable Foundation**: Prevents common CLI bugs through proven patterns
- **Development Delays**: Avoids time spent debugging basic infrastructure issues
- **Maintenance Burden**: Reduces ongoing maintenance by using battle-tested patterns

## Requirements

### Functional Requirements

#### Core Framework Components

- **Go Module Setup**: Proper go.mod with required dependencies
- **Cobra CLI Framework**: Root command structure and command registration
- **Command Routing**: Clean system for adding new commands (list, add, remove)
- **Argument Parsing**: Strict validation following gh-sub-issue patterns
- **Flag Management**: Standard flags (--repo, --help, etc.) with proper validation

#### Command Infrastructure

```go
// Root command structure following gh-sub-issue patterns
var rootCmd = &cobra.Command{
    Use:   "gh-issue-dependency",
    Short: "Manage GitHub issue dependencies from the command line",
    Long:  `Detailed description with examples...`,
}

// Standard command pattern for each subcommand
var listCmd = &cobra.Command{
    Use:   "list <issue-number>",
    Short: "List issue dependencies",
    Args:  cobra.ExactArgs(1),
    RunE:  runList,
}
```

#### Error Handling System

- **Layered Error Context**: Wrap errors with context at each level
- **User-Friendly Messages**: Convert technical errors to actionable guidance  
- **Error Categories**: Distinguish auth, permission, network, and validation errors
- **Recovery Suggestions**: Provide specific steps for error resolution

#### Repository Detection

- **Current Context**: Use `gh repo view --json owner,name` for current repository
- **URL Parsing**: Support GitHub issue URLs with repository extraction
- **Explicit Override**: Support --repo OWNER/REPO flag for cross-repository operations
- **Validation**: Verify repository exists and user has appropriate access

#### Help System

- **Standard Format**: Follow gh CLI help conventions exactly
- **Usage Examples**: Provide clear examples for each command
- **Flag Documentation**: Comprehensive flag descriptions with defaults
- **Error Guidance**: Include common error scenarios in help text

### Non-Functional Requirements

#### Performance Expectations

- **Startup Time**: < 500ms for help and basic commands
- **Repository Detection**: < 1 second for context resolution
- **Framework Overhead**: Minimal performance impact on actual commands

#### Security Considerations

- **Authentication**: Leverage existing gh CLI authentication seamlessly
- **Input Validation**: Sanitize and validate all user inputs
- **API Integration**: Use github.com/cli/go-gh/v2 for secure API access
- **Permission Handling**: Graceful handling of insufficient permissions

#### Reliability Requirements

- **Consistent Behavior**: Identical patterns across all commands
- **Error Recovery**: Graceful handling of all failure scenarios
- **State Consistency**: No partial states or corruption from framework issues
- **Cross-Platform**: Works identically on Windows, macOS, and Linux

## Success Criteria

### Measurable Outcomes

- **Foundation Completeness**: All dependency commands can build on this foundation without modifications
- **Error Handling Quality**: Zero framework-related bugs in dependent commands
- **Development Velocity**: New commands can be added in 1 day using established patterns
- **User Experience**: Help system and error messages receive positive feedback

### Key Metrics and KPIs

- Framework startup time (target: < 500ms)
- Repository context resolution time (target: < 1s)
- Error message clarity rating from user feedback
- Code reuse percentage across commands (target: 80%+)

## Constraints & Assumptions

### Technical Limitations

- **Go Version**: Requires Go 1.19+ for modern features and security
- **Dependency Management**: Must use exact versions from gh-sub-issue for compatibility
- **GitHub API**: Limited by GitHub CLI authentication and API capabilities
- **Platform Support**: Must work on all platforms supported by GitHub CLI

### Timeline Constraints

- **Phase 1 Foundation**: 1 day for complete CLI foundation implementation
- **Testing Window**: Same day as implementation with comprehensive validation
- **Command Dependencies**: All other commands block on foundation completion

### Resource Limitations

- **Single Developer**: One developer implementing entire foundation
- **Pattern Replication**: Must exactly follow gh-sub-issue patterns, not innovate
- **Existing Tools**: Must leverage github.com/cli/go-gh/v2 for all GitHub integration

## Out of Scope

### What we're explicitly NOT building

- **Command Implementation**: Actual list, add, remove command logic (separate PRDs)
- **Advanced CLI Features**: Custom completion, shell integration, plugins
- **Configuration Management**: User preferences, settings files, or configuration
- **Logging System**: Advanced logging, debug modes, or telemetry
- **Custom Output Formats**: Beyond standard table and JSON formats
- **Caching Layer**: API response caching or optimization
- **Testing Framework**: Beyond basic unit tests and integration validation
- **Documentation Generation**: Automated help or man page generation
- **Interactive Modes**: Prompts, wizards, or interactive workflows
- **Plugin Architecture**: Extensibility beyond command addition

## Dependencies

### External Dependencies

- **Go 1.19+**: Required for modern language features and security updates
- **github.com/spf13/cobra**: CLI framework (exact version from gh-sub-issue)
- **github.com/cli/go-gh/v2**: GitHub API integration and authentication
- **github.com/stretchr/testify**: Testing framework for unit tests
- **GitHub CLI (gh)**: Must be installed and authenticated for runtime

### Internal Team Dependencies

- **gh-sub-issue Analysis**: Complete understanding of patterns to replicate
- **Project Plan**: Updated project plan reflecting Go implementation approach
- **API Documentation**: Understanding of GitHub issue dependencies API

### Development Dependencies

- **Go Development Environment**: Properly configured Go workspace
- **GitHub Repository**: Access to create and test GitHub CLI extensions
- **Testing Repository**: Access to repository for integration testing

## Implementation Strategy

### gh-sub-issue Pattern Replication

**Core Patterns to Duplicate:**

1. **main.go Entry Point**: Clean delegation to cmd.Execute() with exit code handling
2. **cmd/root.go Structure**: Root command setup with global flags and command registration
3. **Command Organization**: Each command in separate file with consistent init() registration
4. **Error Handling**: Layered error wrapping with user-friendly message conversion
5. **Testing Structure**: Unit tests paired with each command file plus integration script

### File Structure (Mirroring gh-sub-issue)

```text
gh-issue-dependency/
├── main.go                      # Entry point, delegates to cmd.Execute()
├── cmd/
│   ├── root.go                  # Root command, global flags, command registration
│   ├── list.go                  # List command (empty implementation initially)
│   ├── add.go                   # Add command (empty implementation initially)
│   └── remove.go                # Remove command (empty implementation initially)
├── pkg/                         # Shared utilities (GitHub API, validation, formatting)
├── go.mod                       # Go module with exact dependencies from gh-sub-issue
├── go.sum                       # Dependency checksums
└── tests/
    ├── integration_test.sh      # Shell-based integration tests
    └── *_test.go                # Go unit tests for each component
```

### Dependencies (Exact Versions from gh-sub-issue)

```go
module github.com/torynet/gh-issue-dependency

go 1.19

require (
    github.com/spf13/cobra v1.7.0
    github.com/cli/go-gh/v2 v2.4.0
    github.com/stretchr/testify v1.8.4
)
```

### Implementation Phases

1. **Project Setup**: Initialize Go module and directory structure
2. **Root Command**: Create main.go and cmd/root.go following gh-sub-issue exactly
3. **Command Stubs**: Create empty command files with proper Cobra structure
4. **Error Handling**: Implement layered error handling patterns
5. **Repository Detection**: Add GitHub repository context detection
6. **Testing Infrastructure**: Unit tests and integration test script
7. **Validation**: Comprehensive testing against gh-sub-issue patterns

---

## Implementation Notes

This PRD establishes the architectural foundation that all subsequent commands will build upon. Success is measured by how seamlessly dependency commands can be implemented on top of this foundation without requiring any changes to the core framework.

The emphasis on exact pattern replication from gh-sub-issue ensures we inherit their reliability and user experience quality while avoiding common CLI development pitfalls. The foundation should be so solid that command implementation becomes straightforward pattern application rather than complex integration work.