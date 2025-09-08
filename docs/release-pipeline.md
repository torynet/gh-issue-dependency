# Release Pipeline

## Overview

This project uses a trunk-based development workflow with automated release management through conventional commits and manual approval gates.

## Workflow

### Development Process

1. **Create Feature Branch**
   ```bash
   git checkout -b feature/123-add-new-feature
   ```

2. **Make Commits** (automatic issue number injection)
   ```bash
   git commit -m "Add user authentication"
   # Hook transforms to: "#123: Add user authentication"
   ```

3. **Create Pull Request**
   - Use PR template to indicate breaking changes
   - Branch and PR title validation runs automatically

4. **Review and Merge**
   - Squash merge to main with conventional commit format
   - Example: `feat: add user authentication` or `feat!: add breaking API changes`

### Release Process

#### Automatic RC Creation
- Push to `main` triggers conventional commit analysis
- Version bump determined by commit type:
  - `fix:` → patch (1.0.0 → 1.0.1)
  - `feat:` → minor (1.0.0 → 1.1.0)  
  - `feat!:` or `BREAKING CHANGE:` → major (1.0.0 → 2.0.0)
- RC tag created: `v1.1.0-rc1`
- Beta build artifacts generated

#### Manual Release Gate
- RC builds await approval in `release-approval` environment
- Reviewers test beta artifacts
- Manual approval promotes RC to production release
- Release tag created: `v1.1.0`
- Production build artifacts generated

## Branch Naming

### Required Format
```
(feature|hotfix|epic)/{issue-number}-{description}
```

### Examples
- `feature/123-add-user-auth`
- `hotfix/456-fix-memory-leak`
- `epic/789-redesign-ui`

## Conventional Commits

### Format
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types
- `feat:` - New feature (minor version bump)
- `fix:` - Bug fix (patch version bump)
- `perf:` - Performance improvement (patch version bump)
- `docs:` - Documentation changes (no version bump)
- `style:` - Code style changes (no version bump)
- `refactor:` - Code refactoring (no version bump)
- `test:` - Test additions/changes (no version bump)
- `build:` - Build system changes (no version bump)
- `ci:` - CI configuration changes (no version bump)
- `chore:` - Maintenance tasks (no version bump)

### Breaking Changes
Add `!` after type or include `BREAKING CHANGE:` in footer:
```
feat!: remove deprecated API endpoints
```

## GitHub Actions Workflows

### CI Pipeline (`ci.yml`)
**Triggers:** Push to feature branches, PRs to main

**Jobs:**
- `validate-branch` - Ensures branch naming convention
- `validate-pr-title` - Ensures PR title includes issue number
- `test` - Unit and integration tests across Go versions
- `lint` - Code linting and formatting checks
- `security` - Security scanning with Gosec
- `build-cross-platform` - Cross-platform build verification

### Trunk Release Pipeline (`trunk-release.yml`)
**Triggers:** Push to main, manual workflow dispatch

**Jobs:**
- `create-rc` - Analyzes commits and creates RC tags
- `build-rc` - Builds beta artifacts for RC
- `promote-gate` - Manual approval gate (release-approval environment)
- `create-release` - Creates production release tag
- `notify-success` - Success notifications

### Production Release (`release.yml`)
**Triggers:** Release tags (excluding RC tags)

**Jobs:**
- `build` - Cross-platform production builds
- `release` - Creates GitHub release with artifacts
- `update-homebrew` - Updates Homebrew formula (planned)
- `notify` - Release notifications

## Build Artifacts

### RC Builds (Beta)
- Binary naming: `gh-issue-dependency-v1.0.0-rc1-{os}-{arch}`
- Build flags: `-X cmd.BuildType=beta`
- Purpose: Pre-release testing and validation

### Release Builds (Production)  
- Binary naming: `gh-issue-dependency-v1.0.0-{os}-{arch}`
- Build flags: `-X cmd.BuildType=release`
- Purpose: End-user distribution

## Environment Setup

### Required GitHub Configuration

1. **Environment: `release-approval`**
   - Required reviewers: Release managers
   - Deployment protection: `main` branch only
   - Manual approval required

2. **Branch Protection: `main`**
   - Require status checks: `validate-branch`, `validate-pr-title`, `test`, `lint`
   - Require branches up to date
   - No direct pushes (PR only)

### Developer Setup

1. **Install Git Hooks**
   ```bash
   ./scripts/install-hooks.sh
   ```

2. **Configure Git** (if needed)
   ```bash
   git config --global user.name "Your Name"
   git config --global user.email "your.email@example.com"
   ```

## Release Promotion

### Normal Release
1. RC builds and tests complete successfully
2. Navigate to Actions → Environments → `release-approval`
3. Review RC artifacts and test results
4. Click "Approve deployment"
5. Production release created automatically

### Manual Promotion
```bash
# Promote existing RC to release
gh workflow run trunk-release.yml -f promote_rc_to_release=v1.0.0-rc1
```

### Hotfix Process
1. Create hotfix branch: `hotfix/999-critical-fix`
2. Follow normal PR process with `fix:` conventional commit
3. Results in patch version bump
4. Same RC → release promotion process

## Monitoring

### Release Status
- **GitHub Actions**: Monitor workflow runs
- **GitHub Releases**: Track version history
- **Environments**: View approval history

### Build Verification
- Cross-platform compatibility testing
- Security scanning results
- Code coverage reports
- Integration test results

## Troubleshooting

### Common Issues

**RC not created after main push:**
- Check commit message follows conventional format
- Verify commit type triggers release (`feat`, `fix`, `perf`)
- Review workflow logs for analysis errors

**Release gate not triggered:**
- Verify `release-approval` environment exists
- Check required reviewers are configured
- Ensure user has approval permissions

**Build failures:**
- Check Go version compatibility (1.21+)
- Verify cross-platform build matrix
- Review dependency updates

### Support
- **Workflow Issues**: Check GitHub Actions logs
- **Branch Issues**: Review CI validation output  
- **Hook Issues**: Re-run `./scripts/install-hooks.sh`