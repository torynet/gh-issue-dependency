# Trunk-Based Release System Setup

## Overview

This repository uses a trunk-based development workflow with automated RC creation and manual release promotion gates. This document describes the setup steps and usage.

## GitHub Environment Setup

### 1. Create Release Approval Environment

You need to create a `release-approval` environment in your GitHub repository settings:

1. Go to **Settings** → **Environments** → **New environment**
2. Name: `release-approval`
3. Configure deployment protection rules:
   - **Required reviewers**: Add repository maintainers/release managers
   - **Deployment branches**: Restrict to `main` branch only
   - **Wait timer**: Optional (e.g., 5 minutes minimum wait)

### 2. Branch Protection Rules (Optional but Recommended)

Configure branch protection for `main`:
1. Go to **Settings** → **Branches** → **Add rule**
2. Branch name: `main`
3. Configure:
   - ✅ Require pull request reviews before merging
   - ✅ Require status checks to pass before merging
   - ✅ Require branches to be up to date before merging
   - ✅ Include administrators

## Installation and Usage

### 1. Install Commit Hooks

```bash
# Install git hooks for automatic issue ID insertion
./scripts/install-hooks.sh
```

This installs a commit-msg hook that:
- Extracts issue numbers from branch names
- Automatically prepends `#123: ` to commit messages
- Works cross-platform (Bash/PowerShell)

### 2. Development Workflow

#### Create Feature Branch
```bash
# Branch naming convention: (feature|hotfix|epic)/{issue-number}-{description}
git checkout -b feature/123-add-user-authentication
```

#### Make Commits
```bash
# Your commit message
git commit -m "Add login form validation"

# Hook automatically transforms to:
# "#123: Add login form validation"
```

#### Create Pull Request
1. **PR Title Format**: `123: Add user authentication`
2. **Complete PR Template**: Check appropriate boxes for change type
3. **Breaking Changes**: If applicable, fill out breaking change section

#### Squash Merge to Main
- When PR is approved, use "Squash and merge"
- The squash commit message should follow conventional commit format
- Example: `feat: Add user authentication` or `feat!: Add user authentication` (breaking)

### 3. Release Process

#### Automatic RC Creation
After squash merge to main:
1. **Conventional commit analysis** determines if release is needed
2. **RC tag created** automatically (e.g., `v1.2.0-rc1`)
3. **Beta binaries built** and uploaded as workflow artifacts
4. **Release approval gate** waits for manual approval

#### Manual Release Promotion
To promote RC to production:
1. Go to **Actions** → **Trunk-Based Release Pipeline**
2. Click **"Run workflow"**
3. Select **"Promote RC to production"**
4. Enter RC tag (e.g., `v1.2.0-rc1`)
5. **Release approval required** - designated reviewers must approve
6. **Production release created** automatically after approval

## Workflow Files

### Core Workflows
- **`ci.yml`**: Branch validation, testing, cross-platform builds
- **`trunk-release.yml`**: RC creation, manual gate, release promotion
- **`auto-squash.yml`**: Conventional commit generation from PR metadata

### Supporting Files
- **`.github/pull_request_template.md`**: PR template with conventional commit guidance
- **`scripts/install-hooks.sh`**: Git hook installation script

## Conventional Commit Format

### Commit Types and Version Impact
- **`feat:`** → Minor version bump (new feature)
- **`fix:`** → Patch version bump (bug fix)
- **`perf:`** → Patch version bump (performance improvement)
- **`feat!:`** → Major version bump (breaking feature)
- **`fix!:`** → Major version bump (breaking fix)
- **`docs:`** → No version bump
- **`refactor:`** → No version bump
- **`test:`** → No version bump
- **`ci:`** → No version bump

### Breaking Changes
Include `BREAKING CHANGE:` in commit body:
```
feat!: redesign authentication API

BREAKING CHANGE: The authenticate() function now returns a Promise instead of a callback.
Migration: Change authenticate(callback) to await authenticate().
```

## Troubleshooting

### Branch Validation Fails
**Error**: Branch name doesn't follow convention

**Solution**: Use pattern `(feature|hotfix|epic)/{number}-{description}`
- ✅ `feature/123-add-auth`
- ❌ `feature/add-auth`
- ❌ `123-add-auth`

### PR Title Validation Fails
**Error**: PR title must start with issue number

**Solution**: Ensure PR title starts with issue number from branch
- Branch: `feature/123-add-auth`
- ✅ PR Title: `123: Add user authentication`
- ❌ PR Title: `Add user authentication`

### RC Not Created After Merge
**Possible Causes**:
1. **Non-conventional commit**: Squash commit doesn't use `feat:`, `fix:`, or `perf:`
2. **Template not completed**: PR template checkboxes not selected
3. **Internal change**: Changes like `docs:`, `refactor:`, `test:` don't trigger releases

**Solution**: Check squash commit message follows conventional format

### Release Gate Stuck
**Possible Causes**:
1. **Missing reviewers**: No one assigned to `release-approval` environment
2. **Wrong branch**: Release approval restricted to `main` branch only
3. **Pending approvals**: Required reviewers haven't approved yet

**Solution**: Check environment configuration and get required approvals

### Installation Error
**Error**: `gh extension install` fails with "no usable release artifact"

**Cause**: No production releases created yet (only RCs exist)

**Solution**: Promote an RC to production release first, then extension will be installable

## Examples

### Complete Feature Development
```bash
# 1. Create feature branch
git checkout -b feature/456-improve-performance

# 2. Make changes and commit
git add .
git commit -m "Optimize dependency lookup algorithm"
# Hook transforms to: "#456: Optimize dependency lookup algorithm"

# 3. Push and create PR
git push -u origin feature/456-improve-performance

# 4. In PR template, check:
# - [x] ⚡ Performance improvement (patch version bump)

# 5. After approval, squash merge creates:
# "perf: improve dependency lookup performance"

# 6. Automatic RC creation: v1.2.1-rc1
# 7. Manual promotion to: v1.2.1
```

### Breaking Change Process
```bash
# 1. Create feature branch  
git checkout -b feature/789-redesign-api

# 2. Make breaking changes
# ... implement new API ...

# 3. In PR template:
# - [x] ✨ New feature (minor version bump)
# - [x] ⚠️ This PR contains breaking changes (major version bump)
# 
# Breaking Change Description:
# The list() method now returns a Promise instead of taking a callback
#
# Migration Guide:
# 1. Change list(callback) to await list()
# 2. Handle Promise rejection for error cases

# 4. Squash merge creates:
# "feat!: redesign list API for better async support
# 
# BREAKING CHANGE: The list() method now returns a Promise instead of taking a callback"

# 5. Results in major version bump: v1.0.0 → v2.0.0
```

## Best Practices

### For Developers
1. **Always use issue-linked branches**: `feature/123-description`
2. **Complete PR template**: Select appropriate change type
3. **Write clear commit messages**: Hook will add issue ID automatically  
4. **Test thoroughly**: RCs should be production-ready

### For Release Managers
1. **Review RC artifacts**: Download and test beta binaries before promotion
2. **Validate breaking changes**: Ensure migration guides are complete
3. **Check version bumps**: Verify semantic versioning is correct
4. **Monitor releases**: Watch for installation/upgrade issues post-release

---

This trunk-based system provides controlled, automated releases while maintaining development velocity through proper branching conventions and conventional commits.