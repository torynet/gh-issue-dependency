---
name: cli-foundation
status: backlog
created: 2025-09-06T21:00:16Z
progress: 0%
prd: .claude/prds/cli-foundation.md
github: https://github.com/torynet/gh-issue-dependency/issues/1
---

# Epic: cli-foundation

## Overview

Implement the core CLI framework for gh-issue-dependency by replicating the proven patterns from gh-sub-issue. This foundation provides Cobra-based command routing, layered error handling, repository context detection, and testing infrastructure that all dependency management commands will build upon.

## Architecture Decisions

- **Go + Cobra Framework**: Use exact dependency versions from gh-sub-issue for proven reliability
- **Pattern Replication**: Mirror gh-sub-issue's architecture exactly rather than innovating
- **Layered Error Handling**: Implement context-wrapping error patterns for user-friendly messages
- **GitHub CLI Integration**: Leverage github.com/cli/go-gh/v2 for authentication and API access
- **Testing Strategy**: Unit tests paired with integration shell script following gh-sub-issue model

## Technical Approach

### Core Framework Components

**Go Module Setup**:
- Initialize `go.mod` with exact versions: cobra v1.7.0, go-gh/v2 v2.4.0, testify v1.8.4
- Go 1.19+ requirement for modern features and security

**Entry Point (main.go)**:
- Clean delegation to cmd.Execute() with proper exit code handling
- No business logic in main function, pure orchestration

**Root Command (cmd/root.go)**:
- Root command setup with help text and global flags
- Command registration system for list, add, remove commands
- Standard --repo flag for cross-repository operations

### Command Infrastructure

**Command Stubs**:
- cmd/list.go, cmd/add.go, cmd/remove.go with Cobra structure
- Consistent init() pattern for command registration
- Proper argument validation using cobra.ExactArgs()

**Repository Context Detection**:
- Use `gh repo view --json owner,name` for current repository detection
- GitHub URL parsing for cross-repository issue references
- --repo flag validation and override functionality

### Error Handling System

**Layered Error Context**:
- Wrap errors with context at each level (API, validation, business logic)
- Convert technical errors to actionable user guidance
- Categorize errors: auth, permission, network, validation

**User-Friendly Messages**:
- "Run 'gh auth login' first" for authentication errors
- "Insufficient permissions" for access errors with specific guidance
- Clear validation error messages with correction suggestions

### Testing Infrastructure

**Unit Tests**:
- *_test.go files paired with each command file
- Table-driven tests with comprehensive error scenarios
- Testify framework for assertions and mocking

**Integration Testing**:
- Shell script for build validation and help text verification
- Error case validation with expected message matching
- Integration with actual GitHub repositories for validation

## Implementation Strategy

### Development Phases

**Phase 1: Project Structure** (2-3 hours)
- Initialize Go module with exact gh-sub-issue dependencies
- Create directory structure mirroring gh-sub-issue layout
- Set up basic main.go and cmd/root.go files

**Phase 2: Command Framework** (3-4 hours)
- Implement root command with help system
- Create command stubs for list, add, remove with proper Cobra structure
- Add global --repo flag with validation

**Phase 3: Error Handling** (2-3 hours)
- Implement layered error wrapping patterns
- Add user-friendly error message conversion
- Handle authentication and permission error scenarios

**Phase 4: Repository Detection** (2-3 hours)
- Implement `gh repo view --json` integration
- Add GitHub URL parsing for cross-repository support
- Validate repository access and permissions

**Phase 5: Testing & Validation** (1-2 hours)
- Create unit test structure and basic tests
- Implement integration test shell script
- Validate against gh-sub-issue patterns

### Risk Mitigation

- **Pattern Deviation**: Strict adherence to gh-sub-issue structure prevents architectural mistakes
- **Dependency Issues**: Use exact versions to avoid compatibility problems
- **Performance**: Minimize framework overhead through efficient command registration

## Task Breakdown Preview

High-level task categories that will be created:
- [ ] **Project Setup**: Initialize Go module and directory structure
- [ ] **Root Command Infrastructure**: main.go and cmd/root.go implementation
- [ ] **Command Framework**: Stub implementations for list, add, remove commands
- [ ] **Error Handling System**: Layered error patterns and user-friendly messaging
- [ ] **Repository Context**: GitHub repository detection and validation
- [ ] **Testing Infrastructure**: Unit tests and integration validation
- [ ] **Documentation**: Help system and usage examples following gh CLI patterns

## Dependencies

### External Dependencies
- Go 1.19+ development environment
- github.com/spf13/cobra v1.7.0 (CLI framework)
- github.com/cli/go-gh/v2 v2.4.0 (GitHub API integration)
- github.com/stretchr/testify v1.8.4 (testing framework)
- GitHub CLI (gh) installed and authenticated

### Internal Dependencies
- Complete analysis of gh-sub-issue patterns (completed)
- Updated project plan with Go implementation approach (completed)
- Access to GitHub repository for testing and validation

### Prerequisite Work
- None - this is the foundational component all other work depends on

## Success Criteria (Technical)

### Performance Benchmarks
- Framework startup time < 500ms for help and basic commands
- Repository context resolution < 1 second
- Zero performance overhead for command implementation

### Quality Gates
- All unit tests pass with comprehensive error scenario coverage
- Integration tests validate help system and basic command routing
- Error messages provide actionable guidance in all failure scenarios
- Code follows exact patterns from gh-sub-issue for consistency

### Acceptance Criteria
- All dependency commands (list, add, remove) can build on foundation without modifications
- Help output follows standard gh CLI format and conventions
- Error handling provides clear, categorized feedback for all failure types
- Repository detection works from any subdirectory within a GitHub repository

## Estimated Effort

### Overall Timeline
- **Total Implementation**: 1 day (10-15 hours)
- **Critical Path**: Sequential implementation required, no parallelization possible

### Resource Requirements
- **Single Developer**: Full-stack Go development with CLI experience
- **Development Environment**: Go 1.19+, GitHub CLI, test repository access

### Critical Path Items
1. Project structure setup (blocks all other work)
2. Root command implementation (blocks command development)
3. Error handling system (blocks reliable command implementation)
4. Repository detection (blocks practical usage)
5. Testing infrastructure (blocks quality validation)

### Delivery Milestones
- **Hour 3**: Basic project structure with buildable main.go
- **Hour 7**: Complete command framework with help system
- **Hour 10**: Error handling and repository detection functional
- **Hour 13**: Testing infrastructure complete
- **Hour 15**: Full validation and documentation complete

## Tasks Created
- [ ] #2 - Initialize Go module and project structure (parallel: false)
- [ ] #3 - Implement main.go and root command structure (parallel: false)
- [ ] #4 - Create command stubs for list, add, remove (parallel: false)
- [ ] #5 - Implement layered error handling system (parallel: true)
- [ ] #6 - Implement GitHub repository context detection (parallel: true)
- [ ] #7 - Create testing infrastructure and basic tests (parallel: false)
- [ ] #8 - Implement comprehensive help system and documentation (parallel: false)

Total tasks: 7
Parallel tasks: 2
Sequential tasks: 5
Estimated total effort: 14-19 hours
