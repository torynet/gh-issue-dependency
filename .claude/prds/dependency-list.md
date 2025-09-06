---
name: dependency-list
description: Display issue dependency relationships in a tabular format showing blocking and blocked-by issues
status: backlog
created: 2025-09-03T05:06:44Z
---

# PRD: dependency-list

## Executive Summary

The `dependency-list` feature provides users with a clear, organized view of GitHub issue dependencies through the command line. It displays both blocking and blocked-by relationships for any issue, formatted in an intuitive table that helps users understand project dependencies and potential bottlenecks.

## Problem Statement

### What problem are we solving?

GitHub's new issue dependencies feature (launched August 2025) is only accessible through the web interface, limiting developers who prefer command-line workflows. Users need a quick way to visualize issue relationships without context-switching to the browser, especially when managing complex project dependencies.

### Why is this important now?

- GitHub has just made issue dependencies generally available, creating demand for CLI integration
- Project managers and developers need better dependency visibility in their existing CLI workflows
- Issue blocking relationships are critical for sprint planning and identifying bottlenecks
- Current web-only access disrupts terminal-focused development workflows

## User Stories

### Primary User Personas

1. **CLI-First Developer** - Prefers terminal-based workflows, uses gh CLI extensively
2. **Project Manager** - Needs quick dependency overviews for sprint planning
3. **DevOps Engineer** - Manages release dependencies and deployment blockers

### Detailed User Journeys

#### Story 1: Quick Dependency Check

```text
As a developer working on issue #123
I want to see what issues are blocking my work
So that I can understand dependencies before starting implementation

Acceptance Criteria:
- Shows all issues that block #123
- Shows all issues that #123 blocks
- Displays issue numbers, titles, and states
- Runs in under 2 seconds for typical repositories
```

#### Story 2: Project Planning Overview

```text
As a project manager reviewing sprint dependencies
I want to see the full dependency tree for a feature issue
So that I can identify potential bottlenecks and plan accordingly

Acceptance Criteria:
- Clear visual distinction between blocking and blocked relationships
- Shows issue assignees for dependency planning
- Indicates open vs closed status with visual cues
- Supports both TTY (colored) and plain text output
```

#### Story 3: Cross-Repository Dependencies

```text
As a DevOps engineer managing microservice dependencies
I want to see dependencies that span multiple repositories
So that I can coordinate releases across services

Acceptance Criteria:
- Shows repository context for cross-repo dependencies
- Handles authentication across different repositories
- Provides URLs for easy navigation to dependency issues
```

### Pain Points Being Addressed

- **Context Switching**: Eliminates need to open browser for dependency checks
- **Slow Feedback**: Provides instant dependency visibility vs web interface loading
- **Limited Visibility**: Makes hidden dependencies visible in CLI workflows
- **Manual Tracking**: Reduces need for manual dependency documentation

## Requirements

### Functional Requirements

#### Core Display Features

- **Dual Relationship View**: Show both "blocked by" and "blocks" relationships
- **Tabular Format**: Present information in organized, scannable columns
- **Issue Details**: Display issue number, title, state (open/closed), labels
- **Repository Context**: Show repository information for cross-repo dependencies
- **Empty State Handling**: Graceful display when no dependencies exist

#### Command Syntax

```bash
# Basic usage
gh issue-dependency list <issue-number>

# Cross-repository
gh issue-dependency list <issue-number> --repo owner/repo

# URL support
gh issue-dependency list https://github.com/owner/repo/issues/123

# State filtering
gh issue-dependency list 123 --state all|open|closed

# JSON output for scripting
gh issue-dependency list 123 --json number,title,state
```

#### Output Formats

- **TTY Mode**: Colored output with visual indicators (✅ closed, 🔵 open)
- **Plain Text**: Non-colored output for scripts and logs
- **JSON Mode**: Structured data with selectable fields

#### Visual Design (Following gh-sub-issue patterns)

```text
Dependencies for: #123 - Feature: User Authentication System

BLOCKED BY (2 issues)
────────────────────
🔵 #45   Database migration setup     [open]   @alice
✅ #67   Security review completed    [closed] @security-team

BLOCKS (3 issues)
──────────────────
🔵 #124  Frontend login component    [open]   @bob
🔵 #125  API documentation update    [open]   @docs
✅ #126  Integration tests           [closed] @qa
```

### Non-Functional Requirements

#### Performance Expectations

- **Response Time**: < 2 seconds for typical repositories (< 50 dependencies)
- **API Efficiency**: Minimize GitHub API calls using GraphQL where possible
- **Concurrent Requests**: Handle multiple dependency API calls efficiently

#### Security Considerations

- **Authentication**: Use existing gh CLI authentication
- **Permissions**: Respect repository access permissions
- **Rate Limiting**: Handle GitHub API rate limits gracefully with retry logic

#### Scalability Needs

- **Large Dependency Lists**: Handle up to 50 dependencies per issue (GitHub's limit)
- **Cross-Repository**: Support dependencies across multiple repositories
- **Pagination**: Handle large result sets efficiently

## Success Criteria

### Measurable Outcomes

- **Adoption**: 80% of CLI-focused developers use list command within first week
- **Performance**: Average response time under 2 seconds for 95% of queries
- **Reliability**: 99%+ success rate for dependency retrieval
- **User Satisfaction**: Positive feedback on output format and usability

### Key Metrics and KPIs

- Command execution time (target: < 2s average)
- API error rate (target: < 1%)
- Cross-repository usage percentage
- JSON output usage for automation scripts

## Constraints & Assumptions

### Technical Limitations

- **GitHub API Rate Limits**: 5000 requests/hour for authenticated users
- **Dependency Limits**: GitHub allows max 50 dependencies per issue
- **Repository Permissions**: Users must have "Issues" read access
- **Network Dependency**: Requires internet connection for GitHub API access

### Timeline Constraints

- **Phase 1 Implementation**: 1 day (part of 2-3 day total project timeline)
- **Testing Window**: Same day as implementation
- **Documentation**: Concurrent with development

### Resource Limitations

- **Single Developer**: One developer implementing entire feature
- **Go Implementation**: Using Go with Cobra framework following gh-sub-issue patterns
- **Existing Tools**: Must leverage existing gh CLI API capabilities

## Out of Scope

### What we're explicitly NOT building

- **Dependency Editing**: Adding/removing dependencies (separate commands)
- **Dependency Visualization**: ASCII graphs or tree views
- **Real-time Updates**: Live monitoring of dependency changes
- **Bulk Export**: Mass export of dependency data across projects
- **Dependency Analytics**: Metrics on dependency patterns or bottlenecks
- **Integration Webhooks**: Automated dependency notifications
- **Custom Sorting**: Advanced sorting options beyond basic state filtering

## Dependencies

### External Dependencies

- **GitHub CLI (gh)**: Required for API access and authentication
- **GitHub API**: Issue dependencies endpoints (launched August 2025)
- **Internet Connection**: Required for API communication
- **Repository Permissions**: "Issues" read access on target repositories

### Internal Team Dependencies

- **Core Infrastructure**: Depends on main script framework (Phase 1)
- **Error Handling**: Relies on shared error handling patterns
- **Output Formatting**: Builds on common table formatting utilities
- **API Client**: Uses shared GitHub API interaction patterns

### API Dependencies

- `GET /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocked_by`
- `GET /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocking`
- `GET /repos/{owner}/{repo}/issues/{issue_number}` (for issue details)

---

## Implementation Notes

This PRD adapts the proven patterns from yahsan2/gh-sub-issue while leveraging GitHub's new native dependency API. The focus is on providing immediate value through clear, actionable dependency visualization that integrates seamlessly into existing CLI workflows.
