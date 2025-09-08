# Development Workflow

## Quick Start

### New Feature Development

1. **Create Issue** (if not exists)
   - Create GitHub issue for the feature/bug
   - Note the issue number (e.g., #123)

2. **Create Branch**
   ```bash
   git checkout -b feature/123-short-description
   ```

3. **Develop and Commit**
   ```bash
   # Make changes
   git add .
   git commit -m "Add authentication middleware"
   # Hook automatically transforms to: "#123: Add authentication middleware"
   ```

4. **Push and Create PR**
   ```bash
   git push origin feature/123-short-description
   # Create PR with title: "123: Add authentication middleware"
   ```

5. **Review and Merge**
   - Fill out PR template (especially breaking changes section)
   - Get approval
   - Squash merge with conventional commit message

## Branch Types

### Feature Branches
- **Format**: `feature/{issue-number}-{description}`
- **Purpose**: New features and enhancements
- **Example**: `feature/123-add-user-auth`
- **Merge**: Squash merge to `main` with `feat:` commit

### Hotfix Branches  
- **Format**: `hotfix/{issue-number}-{description}`
- **Purpose**: Critical bug fixes for production
- **Example**: `hotfix/456-fix-memory-leak`
- **Merge**: Squash merge to `main` with `fix:` commit

### Epic Branches
- **Format**: `epic/{issue-number}-{description}` 
- **Purpose**: Large features spanning multiple PRs
- **Example**: `epic/789-redesign-dashboard`
- **Process**: May have multiple feature branches merge into it first

## Commit Message Guidelines

### Automatic Issue Injection
The commit hook automatically prepends issue numbers:
```bash
# You write:
git commit -m "Fix validation logic"

# Hook transforms to:
git commit -m "#123: Fix validation logic"
```

### Conventional Commit Format (for Squash Merges)
When squash merging PRs, use conventional commit format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer]
```

#### Types and Version Impact
- `feat:` → Minor version (1.0.0 → 1.1.0)
- `fix:` → Patch version (1.0.0 → 1.0.1)  
- `perf:` → Patch version (1.0.0 → 1.0.1)
- `docs:`, `style:`, `refactor:`, `test:`, `build:`, `ci:`, `chore:` → No version bump

#### Breaking Changes
For breaking changes, add `!` after type:
```
feat!: remove legacy API endpoints

BREAKING CHANGE: The /v1/users endpoint has been removed. Use /v2/users instead.
```

## Pull Request Process

### PR Template
The PR template guides you through:
- Change type selection
- Breaking change declaration
- Testing confirmation
- Conventional commit preview

### Breaking Changes
If your PR contains breaking changes:
1. ✅ Check "This PR contains breaking changes"
2. Fill out breaking change description
3. Provide migration guide
4. List affected APIs/functions

### Review Requirements
- All CI checks must pass
- At least one approval required
- Branch naming validation
- PR title validation (must include issue number)

## Testing Requirements

### Before Creating PR
- [ ] Unit tests pass: `go test ./...`
- [ ] Integration tests pass: `./tests/integration_test.sh`
- [ ] Code formatting: `go fmt ./...`
- [ ] Linting: `golangci-lint run`

### Automated Testing
CI automatically runs:
- Unit tests across Go versions (1.21, 1.22)
- Integration tests
- Cross-platform builds
- Security scanning
- Code coverage analysis

## Local Development Setup

### First-Time Setup
```bash
# Clone repository
git clone <repository-url>
cd gh-issue-dependency

# Install dependencies
go mod download

# Install git hooks
./scripts/install-hooks.sh

# Verify setup
make test
```

### Development Commands
```bash
# Run tests
make test
make test-unit
make test-integration

# Build
make build
make build-all

# Linting and formatting
make fmt
make lint
make vet

# Complete CI check
make ci
```

## Code Style Guidelines

### Go Conventions
- Follow standard Go formatting (`gofmt`)
- Use meaningful variable names
- Add comments for exported functions
- Follow error handling best practices

### Project Structure
```
gh-issue-dependency/
├── cmd/           # Command-line interface
├── internal/      # Private application code
├── pkg/           # Public library code
├── tests/         # Integration tests
├── docs/          # Documentation
├── scripts/       # Build and utility scripts
└── .github/       # GitHub workflows and templates
```

## Release Flow

### Version Lifecycle
1. **Development** → Feature branches
2. **Integration** → Squash merge to `main`
3. **RC Creation** → `v1.1.0-rc1` tag and beta build
4. **Testing** → Manual validation of RC artifacts
5. **Release** → Manual approval → `v1.1.0` tag and production build

### Version Numbers
- **Major** (2.0.0): Breaking changes
- **Minor** (1.1.0): New features, backward compatible
- **Patch** (1.0.1): Bug fixes, backward compatible

### Release Artifacts
- Cross-platform binaries (Linux, macOS, Windows)
- SHA256 checksums
- GitHub Release with changelog
- Future: Homebrew formula, package managers

## Best Practices

### Commit Practices
- Make atomic commits (one logical change per commit)
- Write clear commit messages
- Don't commit secrets, credentials, or large files
- Use the git hooks to ensure issue tracking

### Branch Management
- Keep branches short-lived
- Rebase or merge main regularly to stay current
- Delete feature branches after merge
- Use descriptive branch names

### PR Practices
- Keep PRs focused and reviewable
- Include tests for new functionality
- Update documentation as needed
- Respond to review feedback promptly

### Issue Tracking
- Every branch must reference a GitHub issue
- Use issue numbers in branch names and PR titles
- Link related issues in PR descriptions
- Close issues when work is complete

## Troubleshooting

### Common Developer Issues

**Git hook not working:**
```bash
# Reinstall hooks
./scripts/install-hooks.sh

# Check hook permissions
ls -la .git/hooks/commit-msg
chmod +x .git/hooks/commit-msg
```

**Branch validation failing:**
- Ensure branch name follows: `(feature|hotfix|epic)/{number}-{description}`
- Check that issue number exists in GitHub
- Verify PR title starts with issue number

**CI checks failing:**
- Run tests locally: `make test`
- Check code formatting: `make fmt`
- Review linting: `make lint`
- Verify cross-platform build: `make build-all`

**Version not bumping:**
- Ensure squash merge uses conventional commit format
- Check that commit type triggers releases (`feat`, `fix`, `perf`)
- Review trunk-release workflow logs