# gh-issue-dependency Project Plan

A GitHub CLI extension for managing issue dependencies using GitHub's native blocking/blocked-by relationships.

## Project Overview

**Goal**: Create a GitHub CLI extension similar to `gh-sub-issue` but for managing issue dependencies using the new GitHub issue dependencies API.

**Template**: Based on [gh-sub-issue](https://github.com/yahsan2/gh-sub-issue) architecture and patterns
**Estimated Timeline**: Short-term project
**Language**: Go (with Cobra CLI framework)
**Architecture**: Go binary following gh-sub-issue patterns and GitHub CLI extension conventions

## Core Features

### Commands

1. **`gh issue-dependency list <issue-number>`**
   - List issues that block the specified issue
   - List issues that are blocked by the specified issue
   - Display in tabular format with issue numbers, titles, labels, and status

2. **`gh issue-dependency add <issue-number> --blocked-by <blocking-issue>`**
   - Make an issue blocked by another issue
   - Optional `--replace` flag to replace existing blocking relationships

3. **`gh issue-dependency add <issue-number> --blocks <blocked-issue>`**
   - Make an issue block another issue
   - Optional `--replace` flag to replace existing relationships

4. **`gh issue-dependency remove <issue-number> --blocked-by <blocking-issue>`**
   - Remove a blocked-by relationship

5. **`gh issue-dependency remove <issue-number> --blocks <blocked-issue>`**
   - Remove a blocks relationship

## Technical Implementation

### API Endpoints

- `GET /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocked_by` - List blocking issues
- `POST /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocked_by` - Add blocking dependency
- `DELETE /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocked_by/{issue_id}` - Remove blocking dependency
- `GET /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocking` - List blocked issues

### File Structure

```text
gh-issue-dependency/
├── main.go                      # Entry point (delegates to cmd.Execute())
├── cmd/
│   ├── root.go                  # Root command and global flags
│   ├── list.go                  # List command implementation
│   ├── add.go                   # Add command implementation
│   └── remove.go                # Remove command implementation
├── pkg/                         # Shared utilities and types
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
├── README.md                    # Installation and usage docs
├── LICENSE                      # MIT license
├── project-plan.md             # This file
└── tests/
    ├── integration_test.sh      # Shell-based integration tests
    └── *_test.go                # Go unit tests
```

## Development Phases

### Phase 1: CLI Foundation (Core Infrastructure)

- [x] Repository setup and cloning
- [ ] Create Go CLI foundation with:
  - Go module initialization (go.mod)
  - Cobra CLI framework setup
  - Root command with help/usage
  - Argument parsing and validation framework
  - Layered error handling (following gh-sub-issue patterns)
  - Repository context detection using `gh repo view --json`
  - Testing infrastructure (unit tests + integration script)
  - Standard gh CLI conventions (flags, help format, exit codes)

### Phase 2: List Command

- [ ] Implement `list` command:
  - API call to get blocked-by dependencies
  - API call to get blocking dependencies
  - Format output in readable table
  - Handle empty results gracefully

### Phase 3: Add Commands

- [ ] Implement `add` command with `--blocked-by` option:
  - Validate issue numbers exist
  - Convert issue numbers to issue IDs via API
  - Make POST request to create dependency
  - Success/error messaging
- [ ] Implement `add` command with `--blocks` option:
  - Similar logic but reverse relationship
  - Handle both relationship directions

### Phase 4: Remove Commands

- [ ] Implement `remove` command with `--blocked-by` option:
  - Validate relationships exist
  - Make DELETE request to remove dependency
- [ ] Implement `remove` command with `--blocks` option:
  - Handle reverse relationship removal

### Phase 5: Polish & Testing

- [ ] Enhanced error handling:
  - Network errors
  - Permission errors
  - Invalid issue numbers
  - Rate limiting
- [ ] Input validation:
  - Issue number format validation
  - Repository context validation
- [ ] Output formatting improvements:
  - Color coding for different states
  - Better table formatting
  - Progress indicators for API calls
- [ ] Documentation:
  - Complete README with examples
  - Usage help text
  - Installation instructions

### Phase 6: Advanced Features (Optional)

- [ ] Bulk operations support
- [ ] Interactive mode for relationship management
- [ ] Integration with GitHub Projects
- [ ] Dependency graph visualization (ASCII art)

## Testing Strategy

### Manual Testing

- Create test issues in a test repository
- Test all commands with various scenarios:
  - Single dependencies
  - Multiple dependencies
  - Circular dependency detection
  - Non-existent issues
  - Permission scenarios

### Automated Testing

- Go unit tests for all parsing and validation functions (following gh-sub-issue patterns)
- Integration shell script that creates/cleans up test issues
- Table-driven tests with comprehensive error scenarios
- CI/CD integration with GitHub Actions

## Installation & Distribution

### Installation Command

```bash
gh extension install torynet/gh-issue-dependency
```

### Requirements

- Go 1.19+ (for building from source)
- GitHub CLI (`gh`) installed and authenticated
- Repository with Issues enabled
- "Issues" repository permissions (read for list, write for add/remove)

## Success Criteria

1. **Functionality**: All commands work with real GitHub repositories
2. **Usability**: Intuitive command syntax similar to existing gh extensions
3. **Reliability**: Proper error handling and graceful failure modes
4. **Documentation**: Clear README with examples and troubleshooting
5. **Performance**: Fast response times for typical operations

## Risk Mitigation

### Technical Risks

- **API Rate Limits**: Implement retry logic and informative error messages
- **Authentication Issues**: Clear error messages for token problems
- **Network Failures**: Graceful handling of connectivity issues

### User Experience Risks

- **Complex Syntax**: Follow established gh CLI patterns
- **Confusing Output**: Clear, consistent formatting across commands
- **Missing Context**: Always show repository and issue context

## Future Enhancements

- Integration with GitHub Projects for dependency tracking
- Export dependency graphs to visualization tools
- Webhook integration for dependency notifications
- Bulk import/export of dependency relationships
- Integration with issue templates for automatic dependency setup

## Implementation Strategy

### gh-sub-issue Template Usage

**Architectural Patterns to Duplicate:**
- Cobra CLI framework with clean command structure
- Layered error handling with user-friendly messages
- Repository context detection and flexible issue parsing
- Standard gh CLI conventions (flags, help, output formats)
- Comprehensive testing strategy (unit tests + integration)
- GraphQL API integration patterns

**Key Dependencies (mirror gh-sub-issue):**
- `github.com/spf13/cobra` - CLI framework
- `github.com/cli/go-gh/v2` - GitHub API integration
- `github.com/stretchr/testify` - Testing framework

**Development Approach:**
1. Start with exact gh-sub-issue foundation structure
2. Adapt their command patterns for dependency operations
3. Reuse their error handling and validation approaches
4. Follow their testing patterns and integration script structure

---

*This plan leverages the proven, reliable architecture of gh-sub-issue while adapting it for GitHub's native issue dependencies API.*
