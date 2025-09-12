# Release Pipeline Documentation

## Overview

This project uses a **4-workflow architecture** with trunk-based development, automated RC creation, beta testing, manual promotion gates, and protected production deployments.

## Pipeline Architecture

### 4-Workflow System

#### 1. **CI Workflow** (`ci.yml`)
- **Triggers**: Feature/hotfix/epic branch pushes + PRs to main
- **Purpose**: Code validation during development  
- **No approval gates** - automated validation only

#### 2. **RC Workflow** (`rc.yml`)
- **Triggers**: Main branch pushes (squash merges)
- **Purpose**: Automatic RC tag creation based on conventional commits
- **No approval gates** - fully automated
- **Outputs**: Creates RC tags like `v1.0.0-rc1`, `v1.0.0-rc2`

#### 3. **Beta Workflow** (`beta.yml`)
- **Triggers**: RC tags (`v*-rc*`)
- **Two phases**:
  1. **Beta Deploy**: Auto-deploy to `beta` environment → create pre-release
  2. **Promotion Gate**: Manual approval in `beta-approval` environment → create production tag
- **This is where RC→Release promotion happens**

#### 4. **Release Workflow** (`release.yml`)
- **Triggers**: Production tags (non-RC: `v*` but not `v*-rc*`)
- **Purpose**: Production deployment with `release` environment protection
- **Creates**: Final production releases with binaries

### Complete Release Flow

1. **Development**:
   ```bash
   git checkout -b feature/123-new-feature
   # CI validates on push, PR validation on merge
   ```

2. **RC Creation** (Automatic):
   ```
   Main push → RC workflow → Creates v1.0.0-rc1 → Notifies
   ```

3. **Beta Testing** (Automatic + Manual):
   ```
   RC tag → Beta workflow → Deploys to beta → Creates pre-release
   → Manual approval → Creates v1.0.0 production tag
   ```

4. **Production Release** (Automatic with approval):
   ```
   Release tag → Release workflow → Requires approval → Production release
   ```

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

### GitHub Environments

1. **`beta`** Environment
   - **Purpose**: RC tag testing and validation
   - **Deployment tags**: `v[0-9]*.[0-9]*.[0-9]*-rc[0-9]*` (RC tags only)
   - **Protection**: Optional - can be auto-deploy or minimal approval

2. **`beta-approval`** Environment 
   - **Purpose**: Manual gate for RC→Release promotion
   - **Deployment branches**: `main` (where beta workflow runs)
   - **Protection**: **Required reviewers** for production promotion
   - **Critical**: This is the main approval gate for production releases

3. **`release`** Environment
   - **Purpose**: Production release deployment protection  
   - **Deployment tags**: `v[0-9]*.[0-9]*.[0-9]` (release tags only)
   - **Protection**: **Required reviewers** for production deployment

### Branch Protection (via Rulesets)
- **Main branch**: Require PR reviews, status checks, no direct pushes
- **Required status checks**: CI validation, tests, linting
- **Squash merge only**: Maintains clean commit history

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