---
name: dependency-list
status: backlog
created: 2025-09-06T21:19:16Z
progress: 0%
prd: .claude/prds/dependency-list.md
github: https://github.com/torynet/gh-issue-dependency/issues/9
---

# Epic: dependency-list

## Overview

Implement the `list` command for gh-issue-dependency that displays GitHub issue dependencies in a clear, organized tabular format. This command will show both "blocked by" and "blocks" relationships with visual indicators, state information, and support for multiple output formats including TTY, plain text, and JSON.

## Architecture Decisions

- **GitHub API Integration**: Use GitHub's native issue dependencies API endpoints via go-gh/v2
- **Output Formatting**: Support TTY (colored), plain text, and JSON output modes with automatic detection
- **Data Model**: Fetch dependency data in parallel for optimal performance
- **Visual Design**: Follow gh-sub-issue patterns with emoji indicators and clear section headers
- **Error Handling**: Leverage cli-foundation's layered error system for user-friendly messages

## Technical Approach

### Core Components

**Command Implementation (cmd/list.go)**:
- Cobra command with `list <issue-number>` signature
- Argument validation for issue numbers and URLs
- Flag support: --repo, --state, --json
- Integration with repository context detection

**GitHub API Integration**:
- Parallel API calls for blocked_by and blocking relationships
- Issue details fetching for titles, states, assignees
- Cross-repository dependency support
- Rate limiting and error handling

**Output Formatting**:
- TTY mode with emoji indicators (🔵 open, ✅ closed)
- Sectioned output: "BLOCKED BY" and "BLOCKS" with counts
- JSON output with selectable fields
- Empty state handling with helpful messages

### Data Flow

1. **Input Processing**: Parse issue number/URL and repository context
2. **API Requests**: Parallel fetch of dependency relationships and issue details
3. **Data Assembly**: Combine dependency data with issue metadata
4. **Output Generation**: Format according to output mode and display

### Performance Optimizations

- **Parallel API Calls**: Fetch blocked_by and blocking data simultaneously
- **GraphQL Usage**: Use GraphQL for efficient data retrieval where possible
- **Caching Strategy**: Cache issue details during single command execution
- **Error Recovery**: Graceful degradation when some data unavailable

## Implementation Strategy

### Development Phases

**Phase 1: Core Command Structure** (2-3 hours)
- Implement list command with argument parsing
- Basic GitHub API integration for dependency endpoints
- Simple text output format

**Phase 2: Output Formatting** (2-3 hours)
- TTY mode with colored output and emoji indicators
- Sectioned display with counts and formatting
- JSON output support with field selection

**Phase 3: Enhanced Features** (1-2 hours)
- State filtering (--state open|closed|all)
- Cross-repository dependency display
- Empty state handling and error messages

**Phase 4: Testing & Polish** (1 hour)
- Unit tests for formatting and API integration
- Integration tests with real GitHub data
- Performance optimization and error handling

### Risk Mitigation

- **API Rate Limits**: Implement exponential backoff and informative error messages
- **Authentication Issues**: Clear guidance when gh auth is required
- **Cross-Repo Access**: Handle permission errors gracefully with context

## Task Breakdown Preview

High-level task categories that will be created:
- [ ] **Command Structure**: Basic list command implementation with argument parsing
- [ ] **GitHub API Integration**: Dependency data fetching and issue details retrieval
- [ ] **Output Formatting**: TTY, plain text, and JSON output modes
- [ ] **Cross-Repository Support**: Handle dependencies across multiple repositories
- [ ] **Testing & Validation**: Unit tests and integration validation

## Dependencies

### External Dependencies
- cli-foundation epic completed (command framework, error handling, repository detection)
- GitHub CLI (gh) for authentication and API access
- GitHub API issue dependencies endpoints
- go-gh/v2 library for GitHub API integration

### Internal Dependencies
- Root command structure from cli-foundation
- Repository context detection utilities
- Error handling and user message systems
- Shared output formatting patterns

### Prerequisite Work
- cli-foundation epic must be completed first
- GitHub API endpoints must be accessible and authenticated

## Success Criteria (Technical)

### Performance Benchmarks
- Command execution time < 2 seconds for typical repositories
- Efficient API usage with minimal redundant requests
- Graceful handling of large dependency lists (up to 50 per issue)

### Quality Gates
- All output formats render correctly and consistently
- Error messages provide actionable guidance
- Cross-repository dependencies display with proper context
- Empty states show helpful information

### Acceptance Criteria
- `gh issue-dependency list 123` shows all dependencies for issue 123
- TTY output includes emoji indicators and colored text
- JSON output supports field selection: `--json number,title,state`
- Cross-repository dependencies show repository context
- State filtering works: `--state open` shows only open dependencies
- Error handling provides clear guidance for authentication and permission issues

## Estimated Effort

### Overall Timeline
- **Total Implementation**: 1 day (6-9 hours)
- **Critical Path**: Sequential development with some parallel testing

### Resource Requirements
- **Single Developer**: Go development experience with CLI tools
- **Testing Environment**: Access to GitHub repository with issue dependencies

### Critical Path Items
1. Command structure and argument parsing (blocks API integration)
2. GitHub API integration (blocks output formatting)
3. Output formatting (blocks advanced features)
4. Testing and validation (final quality assurance)

### Delivery Milestones
- **Hour 2**: Basic command with simple text output
- **Hour 4**: Full TTY formatting with emoji indicators  
- **Hour 6**: JSON output and cross-repository support
- **Hour 8**: Complete testing and error handling
- **Hour 9**: Final polish and documentation

## Tasks Created
- [ ] #10 - Implement basic list command structure with argument parsing (parallel: false)
- [ ] #11 - Implement GitHub API integration for dependency data retrieval (parallel: false)
- [ ] #12 - Implement output formatting for TTY, plain text, and JSON modes (parallel: false)
- [ ] #13 - Add enhanced features for state filtering and cross-repository support (parallel: true)
- [ ] #14 - Create comprehensive testing and validation suite (parallel: true)

Total tasks: 5
Parallel tasks: 2
Sequential tasks: 3
Estimated total effort: 9-14 hours
