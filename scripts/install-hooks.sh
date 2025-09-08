#!/bin/bash
set -e

echo "Installing git hooks for trunk-based development..."

# Create hooks directory if it doesn't exist
mkdir -p .git/hooks

# Detect platform and available tools
echo "Detecting platform: $OSTYPE"

# Check if PowerShell Core is available
if command -v pwsh >/dev/null 2>&1; then
    echo "PowerShell Core detected - you can choose implementation:"
    echo "1) Bash (universal compatibility)"
    echo "2) PowerShell Core (better Windows integration)"
    read -p "Choose [1] or 2: " choice
    choice=${choice:-1}
else
    echo "PowerShell Core not found - using bash implementation"
    choice=1
fi

if [ "$choice" = "2" ] && [ -f "D:/resources/commit-msg-hook.ps1" ]; then
    echo "Installing PowerShell commit-msg hook..."
    cp "D:/resources/commit-msg-hook.ps1" .git/hooks/commit-msg
    chmod +x .git/hooks/commit-msg
    echo "✅ PowerShell commit-msg hook installed"
else
    echo "Installing bash commit-msg hook..."
cat > .git/hooks/commit-msg << 'EOF'
#!/bin/bash
# Cross-platform commit-msg hook for trunk-based development
COMMIT_MSG_FILE="$1"

# Read the commit message  
commit_message=$(head -1 "$COMMIT_MSG_FILE")

# Skip if empty or already has issue format
if [[ -z "$commit_message" ]] || [[ "$commit_message" == \#* ]]; then
    exit 0
fi

# Skip for merge commits
if [[ "$commit_message" =~ ^(Merge|Revert|fixup!|squash!) ]]; then
    exit 0
fi

# Get current branch
branch_name=$(git branch --show-current)

# Extract issue number from feature/123-description format
if [[ "$branch_name" =~ ^(feature|hotfix|epic)/([0-9]+)- ]]; then
    issue_id="${BASH_REMATCH[2]}"
    # Prepend issue ID to commit message (cross-platform sed)
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS requires backup extension for -i flag
        sed -i '' "1s/^/#$issue_id: /" "$COMMIT_MSG_FILE"
    else
        # Linux/Windows Git Bash
        sed -i "1s/^/#$issue_id: /" "$COMMIT_MSG_FILE"
    fi
    echo "✅ Added issue #$issue_id to commit message (platform: $OSTYPE)"
fi

exit 0
EOF

chmod +x .git/hooks/commit-msg
    echo "✅ Cross-platform bash commit-msg hook installed"
fi

echo ""
echo "🎉 Git hooks installed successfully!"
echo ""
echo "The commit-msg hook will automatically:"
echo "  - Extract issue numbers from branch names (feature/123-description)"
echo "  - Prepend '#123: ' to commit messages"
echo "  - Skip main branch and merge commits"
echo ""
echo "Branch naming convention:"
echo "  feature/123-add-new-feature"
echo "  hotfix/456-fix-critical-bug"  
echo "  epic/789-redesign-dashboard"
echo ""
echo "Example workflow:"
echo "  1. Create branch: git checkout -b feature/123-add-auth"
echo "  2. Make commit: git commit -m 'Add user authentication'"
echo "  3. Hook transforms to: git commit -m '#123: Add user authentication'"