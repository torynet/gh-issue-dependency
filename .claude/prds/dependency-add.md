---
name: dependency-add
description: Add dependency relationships between GitHub issues with validation and error handling
status: backlog
created: 2025-09-03T05:23:08Z
---

# PRD: dependency-add

## Executive Summary

The `dependency-add` feature enables users to create dependency relationships between GitHub issues directly from the command line. It provides a secure, validated way to establish "blocks" and "blocked-by" relationships, with comprehensive error handling and user feedback to prevent invalid dependency states.

## Problem Statement

### What problem are we solving?

While GitHub's web interface allows creating issue dependencies, CLI-focused developers need a way to establish these relationships without leaving their terminal workflow. Additionally, the web interface lacks advanced validation that could prevent circular dependencies or other problematic relationship states.

### Why is this important now?

- GitHub's issue dependencies are newly available (August 2025) but lack CLI tooling
- Developers need to establish dependencies during feature planning and issue triage
- Manual dependency creation through web interface disrupts terminal-focused workflows
- Validation and error prevention reduce invalid dependency states that can confuse project tracking

## User Stories

### Primary User Personas

1. **Feature Lead** - Plans complex features with multiple interdependent issues
2. **CLI-First Developer** - Manages all GitHub interactions through terminal
3. **Project Manager** - Establishes dependency relationships during sprint planning

### Detailed User Journeys

#### Story 1: Simple Dependency Creation

```text
As a developer planning a feature implementation
I want to mark issue #123 as blocked by #45
So that the dependency is clearly documented for the team

Acceptance Criteria:
- Creates a "blocked-by" relationship from #123 to #45
- Shows confirmation message with relationship details
- Validates both issues exist before creating relationship
- Fails gracefully with clear error messages
```

#### Story 2: Validation and Error Prevention

```text
As a developer adding dependencies to a complex feature
I want to be warned if I'm creating a circular dependency
So that I don't create invalid dependency states

Acceptance Criteria:
- Detects and prevents circular dependency creation
- Prevents self-referential dependencies (issue blocking itself)
- Warns if dependency relationship already exists
- Provides clear guidance on how to resolve validation errors
```

### Pain Points Being Addressed

- **Context Switching**: Eliminates need to open browser for dependency creation
- **Manual Process**: Automates validation that's missing from web interface
- **Error Prevention**: Validates relationships before creation to prevent invalid states

## Requirements

### Functional Requirements

#### Core Creation Features

- **Relationship Types**: Support both "blocks" and "blocked-by" relationship creation
- **Input Flexibility**: Accept issue numbers and basic GitHub URLs
- **Validation Engine**: Comprehensive validation before relationship creation

#### Command Syntax

```bash
# Basic usage - create "blocked-by" relationship
gh issue-dependency add <source-issue> --blocked-by <target-issue>

# Create "blocks" relationship  
gh issue-dependency add <source-issue> --blocks <target-issue>

# URL support
gh issue-dependency add https://github.com/owner/repo/issues/123 --blocks 124

# Dry run mode
gh issue-dependency add 123 --blocks 124 --dry-run
```

#### Validation Requirements

- **Issue Existence**: Verify all referenced issues exist and are accessible
- **Permission Check**: Confirm user has write access to issues/repositories
- **Circular Detection**: Prevent creation of circular dependency chains
- **Self-Reference**: Block attempts to make issue depend on itself
- **Duplicate Detection**: Warn if relationship already exists (with override option)

#### Output and Feedback

- **Success Confirmation**: Clear confirmation of created relationships
- **Validation Warnings**: Detailed explanations of validation failures
- **Dry Run Preview**: Show what would be created without making changes

#### Visual Design (Following project patterns)

```text
Creating dependency for: #123 - Feature: User Authentication System

✅ Added blocked-by relationship: #123 ← #45 (Database migration setup)

Dependency created successfully.
```

### Non-Functional Requirements

#### Performance Expectations

- **Response Time**: < 2 seconds for single relationship creation
- **API Efficiency**: Minimize API calls for validation and creation
- **Validation Speed**: Real-time validation without blocking user experience

#### Security Considerations

- **Authentication**: Use existing gh CLI authentication tokens
- **Permissions**: Respect GitHub's issue write permissions
- **Rate Limiting**: Implement exponential backoff for API rate limits
- **Input Validation**: Sanitize all user inputs to prevent injection issues

#### Reliability Requirements

- **Error Recovery**: Graceful handling of network failures and API errors
- **Idempotency**: Safe to retry operations without creating duplicates
- **State Consistency**: Ensure dependency graph remains valid after operations

## Success Criteria

### Measurable Outcomes

- **Adoption Rate**: 70% of dependency-list users also use dependency-add within 2 weeks
- **Error Rate**: < 2% of dependency creation attempts fail due to bugs
- **Validation Effectiveness**: 90% reduction in invalid dependencies vs web-only creation
- **User Productivity**: Average 50% faster dependency creation vs web interface

### Key Metrics and KPIs

- Command execution time (target: < 2s for single relationship)
- Validation accuracy (target: 100% circular dependency prevention)

## Constraints & Assumptions

### Technical Limitations

- **GitHub API Rate Limits**: 5000 requests/hour for authenticated users
- **Dependency Limits**: GitHub allows max 50 dependencies per issue
- **Permission Requirements**: Users must have "Issues" write access
- **Network Dependency**: Requires internet connection for GitHub API access
- **API Coverage**: Limited by GitHub's dependency API capabilities

### Timeline Constraints

- **Phase 1 Implementation**: 1 day (part of 2-3 day total project timeline)
- **Testing Window**: Same day as implementation
- **Integration**: Must work with existing dependency-list functionality

### Resource Limitations

- **Single Developer**: One developer implementing entire feature
- **Go Implementation**: Using Go with Cobra framework following gh-sub-issue patterns
- **Existing Infrastructure**: Must leverage shared utilities from dependency-list

## Out of Scope

### What we're explicitly NOT building (MVP)

- **Bulk Operations**: Multiple dependency creation in single command (future enhancement)
- **Cross-Repository Dependencies**: Dependencies between different repositories (future enhancement)
- **Bidirectional Creation**: Creating both directions simultaneously (future enhancement)
- **Dependency Modification**: Editing existing relationship types (use remove + add)
- **Dependency Visualization**: ASCII graphs or dependency trees
- **Template-Based Creation**: Pre-defined dependency patterns or templates
- **Automated Dependency Inference**: AI-based dependency suggestions
- **Batch Import**: CSV or JSON file-based bulk imports
- **Workflow Integration**: Git hooks or CI/CD pipeline integration
- **Notification System**: Automated alerts when dependencies are created
- **Advanced Querying**: Complex dependency search or filtering capabilities

## Dependencies

### External Dependencies

- **GitHub CLI (gh)**: Required for authentication and API access
- **GitHub Issues API**: Dependency creation endpoints
- **Internet Connection**: Required for API communication
- **Repository Permissions**: "Issues" write access on target repositories

### Internal Team Dependencies

- **dependency-list**: Shares validation logic and API patterns
- **CLI Infrastructure**: Uses shared argument parsing and error handling
- **Output Formatting**: Leverages common formatting utilities
- **Error Handling**: Builds on shared error reporting patterns

### API Dependencies

- `POST /repos/{owner}/{repo}/issues/{issue_number}/dependencies` - Create dependency relationships
- `GET /repos/{owner}/{repo}/issues/{issue_number}` - Issue validation
- `GET /repos/{owner}/{repo}/issues/{issue_number}/dependencies/*` - Circular dependency detection

### Validation Dependencies

- **Issue Existence Check**: Must validate all referenced issues exist
- **Permission Verification**: Check write access before attempting creation
- **Circular Dependency Detection**: Requires traversing existing dependency graph

---

## Implementation Notes

This PRD builds on the foundation established by dependency-list, focusing on secure and validated relationship creation. The emphasis on validation and error prevention differentiates this from basic web interface functionality, providing additional value through CLI automation and safety checks.

The design prioritizes user safety through comprehensive validation while maintaining the speed and efficiency expected in CLI tools. The MVP focuses on single relationship creation within the current repository, with bulk operations and cross-repository support planned for future enhancements.
