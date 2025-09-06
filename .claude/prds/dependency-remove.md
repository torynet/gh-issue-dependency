---
name: dependency-remove
description: Remove dependency relationships between GitHub issues with validation and confirmation
status: backlog
created: 2025-09-03T05:34:40Z
---

# PRD: dependency-remove

## Executive Summary

The `dependency-remove` feature enables users to delete dependency relationships between GitHub issues directly from the command line. It provides a safe, validated way to remove "blocks" and "blocked-by" relationships with confirmation prompts and comprehensive error handling to prevent accidental deletions.

## Problem Statement

### What problem are we solving?

Once dependency relationships are established, developers need a way to remove them when dependencies are resolved, plans change, or relationships were created in error. The GitHub web interface requires multiple clicks and page navigations to remove dependencies, disrupting CLI-focused workflows.

### Why is this important now?

- Dependency relationships need lifecycle management - they're not permanent once created
- Developers resolve blocking issues and need to clean up completed dependencies
- Project plans evolve and previously valid dependencies become obsolete
- Incorrectly created dependencies (human error) need quick correction
- CLI workflow completeness requires both creation and removal capabilities

## User Stories

### Primary User Personas

1. **CLI-First Developer** - Manages dependency lifecycle through terminal
2. **Project Manager** - Adjusts dependencies as project plans evolve
3. **Feature Lead** - Removes resolved dependencies during feature development

### Detailed User Journeys

#### Story 1: Dependency Resolution

```text
As a developer who completed a blocking issue
I want to remove the dependency relationship that's no longer relevant
So that the dependency graph accurately reflects current project state

Acceptance Criteria:
- Removes specific "blocked-by" or "blocks" relationship
- Shows confirmation of which relationship was removed
- Validates the relationship exists before attempting removal
- Provides clear error messages if removal fails
```

#### Story 2: Error Correction

```text
As a developer who accidentally created a wrong dependency
I want to quickly remove the incorrect relationship
So that I can fix my mistake without disrupting the team

Acceptance Criteria:
- Confirms relationship details before removal
- Provides dry-run option to preview removal
- Shows clear success/failure feedback
- Handles case where relationship doesn't exist gracefully
```

#### Story 3: Project Evolution

```text
As a project manager adjusting sprint plans
I want to remove dependencies that are no longer valid
So that the dependency graph reflects current project priorities

Acceptance Criteria:
- Validates user has permission to modify relationships
- Provides confirmation prompt to prevent accidental deletions
- Shows impact of removal (what issues are affected)
- Maintains relationship integrity after removal
```

### Pain Points Being Addressed

- **Context Switching**: Eliminates need to open browser for dependency removal
- **Manual Process**: Provides faster alternative to multi-click web interface
- **Error Recovery**: Quick correction of mistakenly created dependencies
- **Workflow Completeness**: Complements dependency creation with removal capabilities

## Requirements

### Functional Requirements

#### Core Removal Features

- **Relationship Types**: Support removal of both "blocks" and "blocked-by" relationships
- **Input Flexibility**: Accept issue numbers and GitHub URLs
- **Validation Engine**: Verify relationships exist before attempting removal
- **Confirmation System**: Prevent accidental deletions with user confirmation

#### Command Syntax

```bash
# Basic usage - remove specific relationship
gh issue-dependency remove <source-issue> --blocked-by <target-issue>
gh issue-dependency remove <source-issue> --blocks <target-issue>

# URL support
gh issue-dependency remove https://github.com/owner/repo/issues/123 --blocks 124

# Force removal without confirmation
gh issue-dependency remove 123 --blocks 124 --force

# Dry run mode
gh issue-dependency remove 123 --blocks 124 --dry-run
```

#### Validation Requirements

- **Relationship Existence**: Verify the dependency relationship actually exists
- **Permission Check**: Confirm user has write access to modify relationships
- **Issue Access**: Ensure user can access both source and target issues
- **Input Validation**: Sanitize and validate all user inputs

#### Output and Feedback

- **Confirmation Prompt**: Ask user to confirm removal unless --force flag used
- **Success Confirmation**: Clear confirmation of removed relationships
- **Validation Errors**: Detailed explanations when removal cannot proceed
- **Dry Run Preview**: Show what would be removed without making changes

#### Visual Design (Following project patterns)

```bash
# Confirmation prompt
Remove dependency relationship?
  Source: #123 - Feature: User Authentication System
  Target: #45 - Database migration setup
  Type: blocked-by

This will remove the "blocked-by" relationship between these issues.
Continue? (y/N): y

✅ Removed blocked-by relationship: #123 ← #45 (Database migration setup)

Dependency removed successfully.
```

```bash
# Dry run output
Dry run: dependency removal preview

Would remove:
  ❌ blocked-by relationship: #123 ← #45 (Database migration setup)

No changes made. Use --force to skip confirmation or remove --dry-run to execute.
```

### Non-Functional Requirements

#### Performance Expectations

- **Response Time**: < 2 seconds for single relationship removal
- **API Efficiency**: Minimize API calls for validation and removal
- **Validation Speed**: Quick relationship existence checks

#### Security Considerations

- **Authentication**: Use existing gh CLI authentication tokens
- **Permissions**: Respect GitHub's issue write permissions
- **Confirmation**: Default to safe behavior requiring explicit confirmation
- **Input Validation**: Sanitize all user inputs to prevent injection issues

#### Reliability Requirements

- **Error Recovery**: Graceful handling of network failures and API errors
- **Idempotency**: Safe to retry removal operations
- **State Consistency**: Ensure dependency graph remains valid after removal
- **Rollback Safety**: No unintended side effects from failed removal attempts

## Success Criteria

### Measurable Outcomes

- **Adoption Rate**: 60% of dependency-add users also use dependency-remove within 2 weeks
- **Error Rate**: < 2% of dependency removal attempts fail due to bugs
- **Safety Effectiveness**: Zero accidental deletions reported with default confirmation
- **User Productivity**: Average 40% faster dependency removal vs web interface

### Key Metrics and KPIs

- Command execution time (target: < 2s for single removal)
- Confirmation bypass rate (--force flag usage)
- Dry-run usage percentage for testing
- Error prevention rate (attempted removal of non-existent relationships)

## Constraints & Assumptions

### Technical Limitations

- **GitHub API Rate Limits**: 5000 requests/hour for authenticated users
- **Permission Requirements**: Users must have "Issues" write access
- **Network Dependency**: Requires internet connection for GitHub API access
- **API Coverage**: Limited by GitHub's dependency API deletion capabilities

### Timeline Constraints

- **Phase 1 Implementation**: 1 day (part of 2-3 day total project timeline)
- **Testing Window**: Same day as implementation
- **Integration**: Must work with existing dependency-list and dependency-add functionality

### Resource Limitations

- **Single Developer**: One developer implementing entire feature
- **Go Implementation**: Using Go with Cobra framework following gh-sub-issue patterns
- **Existing Infrastructure**: Must leverage shared utilities from other dependency commands

## Out of Scope

### What we're explicitly NOT building (MVP)

- **Bulk Removal**: Multiple dependency removal in single command (future enhancement)
- **Cross-Repository Removal**: Dependencies between different repositories (future enhancement)
- **Cascade Removal**: Automatically removing related dependencies (future enhancement)
- **Dependency History**: Tracking when/why dependencies were removed
- **Undo Functionality**: Restoring accidentally removed dependencies
- **Advanced Filtering**: Removing dependencies based on complex criteria
- **Batch Operations**: CSV or file-based bulk removals
- **Workflow Integration**: Git hooks or CI/CD pipeline integration
- **Notification System**: Automated alerts when dependencies are removed
- **Dependency Analytics**: Metrics on removal patterns

## Dependencies

### External Dependencies

- **GitHub CLI (gh)**: Required for authentication and API access
- **GitHub Issues API**: Dependency deletion endpoints
- **Internet Connection**: Required for API communication
- **Repository Permissions**: "Issues" write access on target repositories

### Internal Team Dependencies

- **dependency-list**: Shares validation logic and relationship checking
- **dependency-add**: Uses similar API patterns and error handling
- **CLI Infrastructure**: Uses shared argument parsing and error handling
- **Output Formatting**: Leverages common formatting utilities

### API Dependencies

- `DELETE /repos/{owner}/{repo}/issues/{issue_number}/dependencies/{relationship_id}` - Remove dependency relationships
- `GET /repos/{owner}/{repo}/issues/{issue_number}` - Issue validation
- `GET /repos/{owner}/{repo}/issues/{issue_number}/dependencies/*` - Relationship existence verification

### Validation Dependencies

- **Relationship Existence Check**: Must validate relationships exist before removal
- **Permission Verification**: Check write access before attempting removal
- **Issue Access Check**: Ensure user can access both issues in the relationship

---

## Implementation Notes

This PRD complements the dependency-add and dependency-list features to provide complete dependency lifecycle management. The emphasis on safety through confirmation prompts and validation prevents common mistakes while maintaining CLI efficiency.

The MVP focuses on single relationship removal within the current repository, with bulk operations and cross-repository support planned for future enhancements. The design prioritizes user safety and clear feedback over advanced features.
