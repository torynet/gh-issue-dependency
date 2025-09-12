# GitHub Repository Rulesets Configuration

## Overview

This document describes the required GitHub repository rulesets to enforce trunk-based development and protect the main branch. Rulesets are the modern replacement for branch protection rules and provide more granular control.

## Required Rulesets

### 1. Main Branch Protection Ruleset

**Purpose**: Prevent direct commits to main, require PRs and status checks

#### Configuration Steps

1. Go to **Settings** → **Rules** → **Rulesets** → **New ruleset**
2. **Ruleset name**: `Main Branch Protection`
3. **Enforcement status**: ✅ Active
4. **Bypass permissions**: Repository administrators (optional)

#### Target Configuration
- **Target type**: Branch
- **Include by pattern**: `main`

#### Rules Configuration

**Access Restrictions:**
- ✅ **Restrict pushes** 
  - Nobody can push directly to main
- ✅ **Restrict force pushes**
  - Prevent force pushes that rewrite history

**Pull Request Requirements:**
- ✅ **Require pull requests**
  - Required for all changes to main
- ✅ **Require review from code owners** (if CODEOWNERS file exists)
- ✅ **Dismiss stale reviews when new commits are pushed**
- ✅ **Require conversation resolution before merging**
- **Required number of reviewers**: 1 (adjust based on team size)

**Status Check Requirements:**
- ✅ **Require status checks to pass**
- **Required status checks**:
  - `CI / validate-branch` (from ci.yml)
  - `CI / validate-pr-title` (from ci.yml)
  - `CI / test` (from ci.yml)  
  - `CI / lint` (from ci.yml)
  - `CI / security-scan` (from ci.yml)
  - `Auto-Squash Conventional Commits / generate-conventional-commit` (from auto-squash.yml)
- ✅ **Require branches to be up to date before merging**

**Additional Restrictions:**
- ✅ **Restrict deletions**
  - Prevent accidental branch deletion
- ✅ **Block force pushes**
  - Maintain clean git history

### 2. Feature Branch Validation Ruleset

**Purpose**: Enforce branch naming conventions for all feature branches

#### Configuration Steps

1. **Ruleset name**: `Feature Branch Validation`
2. **Enforcement status**: ✅ Active

#### Target Configuration
- **Target type**: Branch
- **Include by pattern**: `feature/*`, `hotfix/*`, `epic/*`

#### Rules Configuration

**Branch Naming Validation:**
- Enforced through CI workflow (automatic validation)
- Pattern: `(feature|hotfix|epic)/{issue-number}-{description}`

**Access Controls:**
- ✅ Allow force pushes (developers need this for rebasing)
- ✅ Allow deletions (cleanup after merge)

## Status Check Configuration

To make the rulesets work properly, ensure these GitHub Actions workflows are configured:

### Required Status Checks
```yaml
# From ci.yml workflow
- CI / validate-branch
- CI / validate-pr-title  
- CI / test
- CI / lint
- CI / security-scan
- CI / build

# From auto-squash.yml workflow  
- Auto-Squash Conventional Commits / generate-conventional-commit
```

### GitHub Actions Status Check Names

The status check names in rulesets must match exactly what appears in the GitHub Actions UI. Check the **Actions** tab after a PR to see the exact names.

## Environment Configuration

### Release Approval Environment

1. Go to **Settings** → **Environments** → **New environment**
2. **Environment name**: `release-approval`
3. **Deployment protection rules**:
   - ✅ **Required reviewers**: Add repository maintainers
   - **Wait timer**: 0 minutes (optional: add delay)
   - **Deployment branches**: Only `main` branch

#### Required Reviewers
Add users who should approve production releases:
- Repository owner
- Lead developer(s)
- Release manager(s)

## Verification Steps

After configuring rulesets, verify they work:

### 1. Test Direct Push Prevention
```bash
# This should fail
git checkout main
echo "test" >> README.md
git commit -m "test direct commit"
git push origin main
# Expected: Push rejected by remote
```

### 2. Test PR Requirement
```bash
# This should work
git checkout -b feature/999-test-ruleset
echo "test" >> README.md  
git commit -m "test PR requirement"
git push origin feature/999-test-ruleset
# Create PR through GitHub UI - should work
```

### 3. Test Branch Naming
```bash
# This should fail CI validation
git checkout -b invalid-branch-name
# Create PR - should fail branch validation check
```

### 4. Test Status Check Requirements
- Create PR with failing tests
- Verify merge is blocked until tests pass
- Verify all required status checks are green before merge allowed

## Troubleshooting

### Status Checks Not Required
**Problem**: PR can be merged despite failing tests

**Solution**: 
1. Check status check names in rulesets match exactly
2. Verify workflows are running and reporting status
3. Ensure rulesets are active and targeting correct branches

### Can't Push to Feature Branches  
**Problem**: Feature branches also blocked

**Solution**:
1. Verify feature branch ruleset allows pushes
2. Check branch pattern matching in rulesets
3. Ensure no conflicting branch protection rules exist

### Ruleset Bypass Not Working
**Problem**: Administrators can't bypass rules when needed

**Solution**:
1. Check bypass permissions in ruleset configuration
2. Verify user has admin role on repository
3. Consider emergency bypass procedures

### Status Check Names Don't Match
**Problem**: Required status checks don't appear

**Solution**:
1. Run a PR to see exact status check names in Actions tab
2. Update ruleset with exact names from GitHub Actions
3. Wait for workflow runs to register status checks

## Migration from Branch Protection Rules

If you have existing branch protection rules:

1. **Document existing rules**: Note current protection settings
2. **Create equivalent rulesets**: Transfer all rules to new rulesets
3. **Test thoroughly**: Verify rulesets work as expected
4. **Delete old rules**: Remove legacy branch protection rules
5. **Update documentation**: Inform team of new process

### Branch Protection vs Rulesets

**Branch Protection Rules (Legacy)**:
- Repository-specific
- Limited targeting options
- Basic rule types

**Repository Rulesets (Modern)**:
- More granular control
- Better pattern matching
- Advanced rule types
- Organization-wide capabilities

## Team Communication

### Developer Instructions

**For Developers**:
1. **Never push directly to main** - always use PRs
2. **Follow branch naming**: `feature/123-description`
3. **Complete PR template**: Ensure all status checks pass
4. **Keep branches updated**: Rebase or merge main regularly

**For Reviewers**:
1. **Check all status checks**: Don't approve if CI failing
2. **Review conventional commit format**: Ensure squash merge will work
3. **Verify breaking changes**: Check template completion for breaking changes

**For Release Managers**:
1. **Monitor trunk-release workflow**: Approve RC promotions
2. **Test RC builds**: Download and validate before production promotion
3. **Emergency procedures**: Know how to bypass rules if needed

## Security Considerations

### Bypass Permissions
- **Minimize bypass users**: Only essential administrators
- **Audit bypass usage**: Monitor when rules are bypassed
- **Emergency procedures**: Document when bypass is acceptable

### Status Check Security
- **Workflow security**: Ensure workflows can't be manipulated
- **Required checks**: Don't allow bypassing critical security scans
- **Third-party actions**: Verify security of all GitHub Actions used

## Maintenance

### Regular Tasks
- **Review bypass usage**: Monthly audit of rule bypasses
- **Update required checks**: When adding new CI workflows
- **Team training**: Ensure new team members understand process
- **Ruleset updates**: Keep rulesets aligned with development process

### When to Update Rulesets
- Adding new CI workflows
- Changing branch naming conventions  
- Modifying review requirements
- Updating security requirements

---

This ruleset configuration enforces professional development practices while maintaining development velocity through proper automation and clear processes.