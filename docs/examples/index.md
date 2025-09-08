# Examples

Real-world usage scenarios and patterns for gh-issue-dependency.

## Sprint Planning

### Setting Up Epic Dependencies

```bash
# Epic issue that depends on multiple features
gh issue-dependency add 100 --blocked-by 101,102,103,104

# Features have their own dependencies
gh issue-dependency add 101 --blocked-by 105,106    # API endpoints needed first
gh issue-dependency add 102 --blocked-by 105,107    # Database and UI components
gh issue-dependency add 103 --blocked-by 101,102    # Integration depends on features
gh issue-dependency add 104 --blocked-by 103        # Testing depends on integration
```

### Reviewing Sprint Progress

```bash
# Check epic status
gh issue-dependency list 100

# Check individual feature progress  
for issue in 101 102 103 104; do
  echo "Feature $issue dependencies:"
  gh issue-dependency list $issue
  echo "---"
done
```

## Release Management

### Feature Release Train

```bash
# Release depends on all features being complete
gh issue-dependency add 200 --blocked-by 201,202,203,204

# Testing can start when features are ready
gh issue-dependency add 300 --blocked-by 201,202,203,204

# Documentation can be written in parallel with development
gh issue-dependency add 400 --blocked-by 201,202,203,204

# Deployment depends on testing
gh issue-dependency add 500 --blocked-by 300
```

### Hot Fix Dependencies

```bash
# Hot fix must be tested before deployment  
gh issue-dependency add 600 --blocked-by 601  # Deploy depends on testing
gh issue-dependency add 601 --blocked-by 602  # Testing depends on fix
```

## Cross-Team Coordination

### Frontend/Backend Dependencies

```bash
# Frontend features depend on backend APIs
gh issue-dependency add frontend/app#123 --blocked-by backend/api#456
gh issue-dependency add frontend/app#124 --blocked-by backend/api#456
gh issue-dependency add frontend/app#125 --blocked-by backend/api#457

# Mobile app depends on the same backend APIs
gh issue-dependency add mobile/ios#200 --blocked-by backend/api#456
gh issue-dependency add mobile/android#201 --blocked-by backend/api#456
```

### Infrastructure Dependencies

```bash
# Application deployment depends on infrastructure
gh issue-dependency add app/deploy#300 --blocked-by infra/setup#400

# Multiple applications depend on shared infrastructure
gh issue-dependency add app1/deploy#301 --blocked-by infra/database#401
gh issue-dependency add app2/deploy#302 --blocked-by infra/database#401
gh issue-dependency add app3/deploy#303 --blocked-by infra/database#401
```

## Database Migration Workflows

### Sequential Migration Dependencies

```bash
# Migrations must happen in order
gh issue-dependency add db/migration-002#100 --blocked-by db/migration-001#99
gh issue-dependency add db/migration-003#101 --blocked-by db/migration-002#100

# Application updates depend on migrations
gh issue-dependency add app/v2.1#200 --blocked-by db/migration-002#100
gh issue-dependency add app/v2.2#201 --blocked-by db/migration-003#101
```

### Migration Rollback Planning

```bash
# Rollback procedures block deployment
gh issue-dependency add app/deploy-v2#300 --blocked-by db/rollback-plan#400

# Testing includes rollback scenarios  
gh issue-dependency add test/integration#500 --blocked-by db/rollback-plan#400
```

## Security and Compliance

### Security Review Dependencies

```bash
# Features must pass security review before release
gh issue-dependency add release/v1.5#100 --blocked-by security/review#200

# Multiple features need security review
gh issue-dependency add security/review#200 --blocked-by feature/auth#101,feature/payment#102,feature/data#103

# Penetration testing depends on feature completion
gh issue-dependency add security/pentest#300 --blocked-by feature/auth#101,feature/payment#102
```

### Compliance Workflows

```bash
# Audit documentation depends on feature implementation
gh issue-dependency add compliance/audit#400 --blocked-by feature/logging#101,feature/encryption#102

# Legal review blocks public release
gh issue-dependency add release/public#500 --blocked-by legal/review#600
```

## Bug Fix Workflows

### Critical Bug Dependencies

```bash
# Hot fix deployment depends on testing
gh issue-dependency add hotfix/deploy#100 --blocked-by hotfix/test#101

# Testing depends on the actual fix
gh issue-dependency add hotfix/test#101 --blocked-by hotfix/implement#102

# Multiple areas need testing
gh issue-dependency add hotfix/deploy#100 --blocked-by hotfix/test-api#103,hotfix/test-ui#104,hotfix/test-db#105
```

### Regression Testing

```bash
# Regression tests block release
gh issue-dependency add release/v1.2.1#200 --blocked-by test/regression#201

# Regression tests depend on fix completion
gh issue-dependency add test/regression#201 --blocked-by bugfix/critical#202,bugfix/high#203
```

## Documentation Workflows

### Documentation Dependencies

```bash
# User documentation depends on feature completion
gh issue-dependency add docs/user-guide#100 --blocked-by feature/new-ui#200

# API documentation blocks integration work
gh issue-dependency add integration/partners#300 --blocked-by docs/api-reference#101

# Training materials depend on documentation
gh issue-dependency add training/materials#400 --blocked-by docs/user-guide#100,docs/api-reference#101
```

## Performance and Testing

### Performance Testing Dependencies

```bash
# Performance tests depend on feature implementation
gh issue-dependency add test/performance#100 --blocked-by feature/search#200,feature/filters#201

# Optimization depends on performance test results
gh issue-dependency add feature/optimization#300 --blocked-by test/performance#100

# Load testing blocks production deployment
gh issue-dependency add deploy/production#400 --blocked-by test/load#101
```

### Integration Testing

```bash
# Integration tests depend on multiple components
gh issue-dependency add test/integration#100 --blocked-by component/a#200,component/b#201,component/c#202

# End-to-end tests depend on integration tests
gh issue-dependency add test/e2e#300 --blocked-by test/integration#100

# User acceptance testing is the final gate
gh issue-dependency add release/staging#400 --blocked-by test/uat#301
```

## Batch Operations and Management

### Bulk Dependency Creation

```bash
# Create multiple dependencies efficiently
gh issue-dependency add 100 --blocked-by 101,102,103,104,105,106,107,108,109,110

# Preview large changes first
gh issue-dependency add 200 --blocks 201,202,203,204,205 --dry-run
```

### Dependency Cleanup

```bash
# Remove completed dependencies
gh issue-dependency remove 100 --blocked-by 101,102 # These are done

# Clean up entire issue dependencies
gh issue-dependency remove 100 --all --dry-run  # Preview first
gh issue-dependency remove 100 --all            # Then execute
```

### Restructuring Dependencies

```bash
# Remove old structure
gh issue-dependency remove 100 --all

# Add new structure
gh issue-dependency add 100 --blocked-by 200,201,202  # New dependencies
gh issue-dependency add 100 --blocks 300,301,302     # What this blocks
```

## Monitoring and Reporting

### Dependency Health Checks

```bash
#!/bin/bash
# Script to check dependency health for a milestone

milestone_issues=(100 101 102 103 104 105)

echo "Dependency Health Report"
echo "======================="

for issue in "${milestone_issues[@]}"; do
    echo "Issue #$issue:"
    gh issue-dependency list $issue --format json | jq -r '
        "  Blocked by: " + (.blocked_by | length | tostring) + " issues",
        "  Blocks: " + (.blocks | length | tostring) + " issues"
    '
    echo ""
done
```

### Progress Tracking

```bash
#!/bin/bash
# Track progress on epic dependencies

epic_issue=100

echo "Epic Progress Report"
echo "==================="

# Get all dependencies
dependencies=$(gh issue-dependency list $epic_issue --format json)

# Check status of blocking issues
echo "$dependencies" | jq -r '.blocked_by[] | .number' | while read issue; do
    status=$(gh issue view $issue --json state | jq -r '.state')
    title=$(gh issue view $issue --json title | jq -r '.title')
    echo "Issue #$issue: $status - $title"
done
```

## Integration with Other Tools

### GitHub Actions Integration

```yaml
# .github/workflows/dependency-check.yml
name: Dependency Check

on:
  issues:
    types: [closed]

jobs:
  check-dependencies:
    runs-on: ubuntu-latest
    steps:
      - name: Check if issue unblocks others
        run: |
          # Install gh-issue-dependency
          go install github.com/torynet/gh-issue-dependency@latest
          
          # Find issues blocked by this one
          blocked_issues=$(gh issue-dependency list ${{ github.event.issue.number }} --format json | jq -r '.blocks[]?.number')
          
          if [ ! -z "$blocked_issues" ]; then
            echo "Issue #${{ github.event.issue.number }} was blocking: $blocked_issues"
            # Notify teams or update project boards
          fi
```

### Project Board Automation

```bash
#!/bin/bash
# Update project board based on dependency status

project_id="PVT_kwDOABCD123"
epic_issue=100

# Get all issues blocked by the epic
blocked_issues=$(gh issue-dependency list $epic_issue --format json | jq -r '.blocks[]?.number')

# Move to "Ready" column when dependencies are resolved
for issue in $blocked_issues; do
    blocking_count=$(gh issue-dependency list $issue --format json | jq '.blocked_by | length')
    
    if [ "$blocking_count" -eq 0 ]; then
        echo "Issue #$issue is ready to start"
        # Move to appropriate project board column
        gh project item-edit --id $project_id --field Status --value "Ready"
    fi
done
```

## Best Practices Examples

### Dependency Documentation

```bash
# Document why dependencies exist in issue comments
gh issue comment 123 --body "
This issue depends on #456 because:
- Database schema changes are required first
- API contract needs to be established
- Security review must be completed

Created dependency with: \`gh issue-dependency add 123 --blocked-by 456\`
"
```

### Team Coordination

```bash
# Create dependencies with team notification
gh issue-dependency add 123 --blocked-by 456 --dry-run  # Preview first

# Notify teams about the dependency structure
gh issue comment 123 --body "
@frontend-team This issue now depends on backend API work (#456).
@backend-team Please prioritize #456 as it's blocking frontend work.

Current dependencies: \`gh issue-dependency list 123\`
"
```

### Release Planning

```bash
# Set up release dependencies with documentation
milestone="v2.1.0"

# Create release issue
gh issue create --title "Release $milestone" --body "
## Release Dependencies

This release depends on the following completed work:
- [ ] Feature A (#101)
- [ ] Feature B (#102) 
- [ ] Bug fixes (#103, #104)
- [ ] Testing (#105)
- [ ] Documentation (#106)

Dependencies managed with gh-issue-dependency.
"

# Set up dependencies
release_issue=$(gh issue list --search "Release $milestone" --json number | jq -r '.[0].number')
gh issue-dependency add $release_issue --blocked-by 101,102,103,104,105,106
```

These examples demonstrate the flexibility and power of gh-issue-dependency for managing complex project workflows and team coordination.