---
name: dependency-add
status: backlog
created: 2025-09-06T21:33:37Z
progress: 0%
prd: .claude/prds/dependency-add.md
github: https://github.com/torynet/gh-issue-dependency/issues/15
---

# Epic: dependency-add

## Overview

Implement the `add` command for gh-issue-dependency that creates GitHub issue dependency relationships with comprehensive validation and error prevention. This command will support both "blocks" and "blocked-by" relationships while preventing circular dependencies, validating permissions, and providing clear user feedback throughout the process.

## Architecture Decisions

- **Validation-First Approach**: Comprehensive validation before any API calls to prevent invalid states
- **Circular Dependency Detection**: Graph traversal algorithm to detect cycles before creation
- **Dry Run Support**: Preview functionality to show intended changes without execution
- **Error Handling**: Layered validation with user-friendly error messages and recovery guidance
- **API Integration**: Leverage dependency-list patterns for consistent GitHub API interaction

## Technical Approach

### Core Components

**Command Implementation (cmd/add.go)**:
- Cobra command with `add <issue-number> --blocked-by|--blocks <target-issue>` signature
- Mutual exclusion validation for --blocked-by and --blocks flags
- Issue number and URL parsing with repository context resolution
- Dry run mode with --dry-run flag

**Validation Engine**:
- Issue existence verification for both source and target
- Permission checking for repository write access
- Circular dependency detection using depth-first search
- Self-reference prevention (issue cannot depend on itself)
- Duplicate relationship detection with override options

**GitHub API Integration**:
- POST operations for dependency creation
- GET operations for validation and circular dependency checking
- Error handling for rate limits, permissions, and network issues
- Integration with existing repository context detection

### Validation Strategy

**Multi-Layer Validation Process**:
1. **Input Validation**: Issue format, flag combinations, repository context
2. **Permission Validation**: Write access to source and target repositories
3. **Existence Validation**: Verify both issues exist and are accessible
4. **Business Logic Validation**: Circular dependencies, self-reference, duplicates
5. **Pre-flight Checks**: Final validation before API calls

**Circular Dependency Detection**:
```go
func detectCircularDependency(source, target IssueRef, client *github.Client) error {
    visited := make(map[string]bool)
    return dfsCircularCheck(target, source, visited, client)
}
```

### Output and User Experience

**Success Flow**:
```text
Creating dependency for: #123 - Feature: User Authentication System

✅ Added blocked-by relationship: #123 ← #45 (Database migration setup)

Dependency created successfully.
```

**Validation Error Flow**:
```text
❌ Cannot create dependency: circular dependency detected
   #123 → #45 → #67 → #123

This would create a circular dependency. Consider breaking the cycle by:
- Removing the dependency from #67 to #123, or
- Restructuring the dependency relationships
```

**Dry Run Flow**:
```text
Dry run: dependency creation preview

Would create:
  ✓ blocked-by relationship: #123 ← #45 (Database migration setup)

Validation checks:
  ✓ Issues exist and are accessible
  ✓ User has write permissions
  ✓ No circular dependency detected
  ✓ Relationship does not already exist

Use --force to skip confirmation or remove --dry-run to execute.
```

## Implementation Strategy

### Development Phases

**Phase 1: Command Structure** (2-3 hours)
- Implement add command with flag parsing and validation
- Basic GitHub API integration for dependency creation
- Input validation and error handling structure

**Phase 2: Validation Engine** (3-4 hours)
- Circular dependency detection algorithm
- Permission and existence validation
- Duplicate detection and prevention

**Phase 3: User Experience** (1-2 hours)
- Dry run functionality and output formatting
- Enhanced error messages with recovery guidance
- Success confirmation and feedback

**Phase 4: Testing & Integration** (1 hour)
- Unit tests for validation logic
- Integration tests with dependency-list patterns
- Error scenario testing and validation

### Risk Mitigation

- **Complex Validation Logic**: Break down into discrete, testable functions
- **GitHub API Limitations**: Implement proper error handling and rate limiting
- **User Experience**: Provide clear guidance for all error scenarios

## Task Breakdown Preview

High-level task categories that will be created:
- [ ] **Command Structure**: Add command implementation with argument and flag parsing
- [ ] **Validation Engine**: Comprehensive validation including circular dependency detection
- [ ] **GitHub API Integration**: Dependency creation with error handling and retry logic
- [ ] **User Experience**: Dry run mode, output formatting, and error guidance
- [ ] **Testing & Validation**: Unit tests and integration validation

## Dependencies

### External Dependencies
- cli-foundation epic completed (command framework, error handling, repository detection)
- dependency-list epic completed (shared GitHub API patterns and validation utilities)
- GitHub CLI (gh) for authentication and API access
- GitHub API dependency creation endpoints

### Internal Dependencies
- Repository context detection from cli-foundation
- Error handling patterns from cli-foundation
- GitHub API client patterns from dependency-list
- Output formatting utilities (can share with dependency-list)

### Prerequisite Work
- cli-foundation epic must be completed
- dependency-list implementation provides patterns for GitHub API integration
- Understanding of GitHub's dependency API endpoints and limitations

## Success Criteria (Technical)

### Performance Benchmarks
- Command execution time < 2 seconds for single dependency creation
- Validation checks complete within 1 second
- Efficient API usage with minimal redundant calls

### Quality Gates
- 100% circular dependency prevention in validation tests
- All error scenarios provide actionable recovery guidance
- Dry run mode accurately previews all changes
- Integration with existing CLI patterns and user experience

### Acceptance Criteria
- `gh issue-dependency add 123 --blocked-by 45` creates dependency relationship
- `gh issue-dependency add 123 --blocks 45` creates reverse relationship
- Circular dependency attempts are blocked with clear explanations
- Dry run mode shows exactly what would be created
- Error messages provide specific guidance for resolution
- Command integrates seamlessly with existing gh-issue-dependency workflow

## Estimated Effort

### Overall Timeline
- **Total Implementation**: 1 day (7-10 hours)
- **Critical Path**: Sequential development with validation complexity

### Resource Requirements
- **Single Developer**: Go development experience with CLI tools and graph algorithms
- **Testing Environment**: Access to GitHub repository with issues for testing

### Critical Path Items
1. Command structure and flag parsing (blocks validation implementation)
2. Validation engine with circular dependency detection (blocks API integration)
3. GitHub API integration (blocks user experience features)
4. User experience and error handling (final integration)

### Delivery Milestones
- **Hour 3**: Basic command with simple dependency creation
- **Hour 6**: Complete validation engine with circular dependency detection
- **Hour 8**: Full user experience with dry run and enhanced error messages
- **Hour 10**: Complete testing and integration validation

## Tasks Created
- [ ] #16 - Implement add command structure with argument and flag parsing (parallel: false)
- [ ] #17 - Implement comprehensive validation engine with circular dependency detection (parallel: false)
- [ ] #18 - Implement GitHub API integration for dependency creation (parallel: false)
- [ ] #19 - Implement user experience features with dry run and enhanced error messaging (parallel: true)
- [ ] #20 - Create comprehensive testing and validation suite (parallel: true)

Total tasks: 5
Parallel tasks: 2
Sequential tasks: 3
Estimated total effort: 9-14 hours
