---
name: dependency-remove
status: completed
updated: 2025-09-06T22:20:27Z
completed: 2025-09-07T19:47:28Z
progress: 100%
prd: .claude/prds/dependency-remove.md
github: https://github.com/torynet/gh-issue-dependency/issues/21
---

# Epic: dependency-remove

## Overview

Implement the `remove` command for gh-issue-dependency that safely deletes GitHub issue dependency relationships with comprehensive validation, confirmation prompts, and error prevention. This command provides lifecycle management for dependencies by allowing safe removal of "blocks" and "blocked-by" relationships while preventing accidental deletions through confirmation workflows and dry-run capabilities.

## Architecture Decisions

- **Safety-First Approach**: Default to confirmation prompts with explicit --force override for automation
- **Validation-Before-Action**: Verify relationship existence before attempting removal operations
- **Shared Validation Logic**: Leverage existing validation utilities from dependency-add and dependency-list
- **Consistent Error Handling**: Follow established error handling patterns from other dependency commands
- **API Integration**: Use GitHub's dependency deletion endpoints with proper error recovery

## Technical Approach

### Core Components

**Command Implementation (cmd/remove.go)**:
- Cobra command with `remove <issue-number> --blocked-by|--blocks <target-issue>` signature
- Mutual exclusion validation for --blocked-by and --blocks flags
- Issue number and URL parsing with repository context resolution
- Confirmation prompts with --force override and --dry-run preview mode

**Validation Engine**:
- Relationship existence verification before removal attempts
- Permission checking for repository write access (reuse from dependency-add)
- Issue accessibility validation for both source and target
- Input sanitization and format validation

**GitHub API Integration**:
- DELETE operations for dependency removal
- GET operations for relationship existence verification
- Error handling for permissions, rate limits, and network issues
- Integration with existing repository context detection patterns

### Removal Strategy

**Multi-Layer Validation Process**:
1. **Input Validation**: Issue format, flag combinations, repository context
2. **Permission Validation**: Write access to source repository
3. **Existence Validation**: Verify relationship actually exists
4. **Confirmation Flow**: User confirmation unless --force or --dry-run
5. **Deletion Execution**: API call with error handling and retry logic

**Relationship Removal Logic**:
```go
func (r *DependencyRemover) RemoveRelationship(source, target IssueRef, relType string, opts RemoveOptions) error {
    // Validate inputs and permissions
    if err := r.validator.ValidateRemoval(source, target, relType); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    // Verify relationship exists
    exists, err := r.verifyRelationshipExists(source, target, relType)
    if err != nil {
        return fmt.Errorf("failed to verify relationship: %w", err)
    }
    if !exists {
        return fmt.Errorf("relationship does not exist: %s %s %s", source, relType, target)
    }
    
    // Handle dry run
    if opts.DryRun {
        return r.outputDryRunPreview(source, target, relType)
    }
    
    // Get user confirmation
    if !opts.Force {
        confirmed, err := r.promptConfirmation(source, target, relType)
        if err != nil || !confirmed {
            return fmt.Errorf("removal cancelled")
        }
    }
    
    // Execute removal
    return r.deleteRelationship(source, target, relType)
}
```

### Output and User Experience

**Confirmation Flow**:
```text
Remove dependency relationship?
  Source: #123 - Feature: User Authentication System
  Target: #45 - Database migration setup  
  Type: blocked-by

This will remove the "blocked-by" relationship between these issues.
Continue? (y/N): y

✅ Removed blocked-by relationship: #123 ← #45 (Database migration setup)

Dependency removed successfully.
```

**Dry Run Preview**:
```text
Dry run: dependency removal preview

Would remove:
  ❌ blocked-by relationship: #123 ← #45 (Database migration setup)

No changes made. Use --force to skip confirmation or remove --dry-run to execute.
```

**Error Scenarios**:
```text
❌ Cannot remove dependency: relationship does not exist
   No blocked-by relationship found between #123 and #45

Use 'gh issue-dependency list #123' to see current dependencies.
```

## Implementation Strategy

### Development Phases

**Phase 1: Command Structure** (1-2 hours)
- Implement remove command with flag parsing and validation
- Basic GitHub API integration for relationship verification
- Input validation and error handling structure

**Phase 2: Validation & Safety** (2-3 hours)
- Relationship existence verification
- Permission and accessibility validation  
- Confirmation prompt system with --force override

**Phase 3: User Experience** (1-2 hours)
- Dry run functionality and output formatting
- Enhanced error messages with recovery guidance
- Success confirmation and feedback

**Phase 4: Testing & Integration** (1 hour)
- Unit tests for validation and removal logic
- Integration tests with dependency-add/list patterns
- Error scenario testing and edge case validation

### Risk Mitigation

- **Accidental Deletion**: Confirmation prompts prevent mistakes, --dry-run allows testing
- **API Limitations**: Proper error handling for GitHub API constraints and rate limits
- **Complex Validation**: Reuse proven validation patterns from dependency-add
- **User Experience**: Clear feedback for all scenarios with actionable guidance

## Task Breakdown Preview

High-level task categories that will be created:
- [ ] **Command Structure**: Remove command implementation with argument and flag parsing
- [ ] **Validation Engine**: Relationship existence verification and permission checking
- [ ] **User Safety Features**: Confirmation prompts, dry run mode, and force override
- [ ] **GitHub API Integration**: DELETE operations with error handling and retry logic
- [ ] **Testing & Validation**: Unit tests and integration validation

## Dependencies

### External Dependencies
- cli-foundation epic completed (command framework, error handling, repository detection)
- dependency-add epic completed (shared validation patterns and API integration)
- dependency-list epic completed (relationship verification utilities)
- GitHub CLI (gh) for authentication and API access
- GitHub API dependency deletion endpoints

### Internal Dependencies
- Repository context detection from cli-foundation
- Error handling patterns from cli-foundation
- Validation utilities from dependency-add (permission checks, input validation)
- GitHub API client patterns from dependency-list
- Output formatting utilities (shared with other dependency commands)

### Prerequisite Work
- cli-foundation epic must be completed for command infrastructure
- dependency-add implementation provides validation and API patterns to reuse
- dependency-list provides relationship verification utilities

## Success Criteria (Technical)

### Performance Benchmarks
- Command execution time < 2 seconds for single dependency removal
- Relationship existence verification within 1 second
- Efficient API usage with minimal redundant calls

### Quality Gates
- 100% prevention of non-existent relationship removal attempts
- All error scenarios provide actionable recovery guidance
- Confirmation prompts prevent accidental deletions effectively
- Dry run mode accurately previews all removal operations

### Acceptance Criteria
- `gh issue-dependency remove 123 --blocked-by 45` removes specific relationship
- `gh issue-dependency remove 123 --blocks 45` removes reverse relationship
- Non-existent relationship attempts provide clear error messages
- Confirmation prompts work correctly with --force override
- Dry run mode shows exactly what would be removed without side effects
- Command integrates seamlessly with existing gh-issue-dependency workflow

## Estimated Effort

### Overall Timeline
- **Total Implementation**: 1 day (5-8 hours)
- **Critical Path**: Sequential development with validation complexity

### Resource Requirements
- **Single Developer**: Go development experience with CLI tools and GitHub API
- **Testing Environment**: Access to GitHub repository with existing dependencies for testing

### Critical Path Items
1. Command structure and flag parsing (enables validation implementation)
2. Relationship existence verification (blocks removal operations)  
3. Confirmation and safety features (blocks user experience)
4. GitHub API integration (final implementation component)

### Delivery Milestones
- **Hour 2**: Basic command with relationship existence verification
- **Hour 4**: Complete validation with confirmation prompts and safety features
- **Hour 6**: Full user experience with dry run and enhanced error messages
- **Hour 8**: Complete testing and integration validation

## Tasks Created
- [ ] #22 - Implement remove command structure with argument and flag parsing (parallel: false)
- [ ] #23 - Implement relationship existence validation and permission checking (parallel: false)
- [ ] #24 - Implement user safety features with confirmation prompts and dry run mode (parallel: false)
- [ ] #25 - Implement GitHub API integration for dependency deletion (parallel: false)
- [ ] #26 - Create comprehensive testing and validation suite (parallel: true)

Total tasks: 5
Parallel tasks: 1
Sequential tasks: 4
Estimated total effort: 5-8 hours