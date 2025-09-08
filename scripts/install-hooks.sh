#!/bin/bash
set -e

echo "Installing git hooks for trunk-based development..."

# Create hooks directory if it doesn't exist
mkdir -p .git/hooks

# Copy commit-msg hook from resources
echo "Installing commit-msg hook..."
if [ -f "D:\resources\commit-msg-hook.ps1" ]; then
    cp "D:\resources\commit-msg-hook.ps1" .git/hooks/commit-msg
    chmod +x .git/hooks/commit-msg
    echo "✅ PowerShell commit-msg hook installed"
elif [ -f "D:\resources\commit-msg-hook.sh" ]; then
    cp "D:\resources\commit-msg-hook.sh" .git/hooks/commit-msg
    chmod +x .git/hooks/commit-msg
    echo "✅ Bash commit-msg hook installed"
else
    echo "❌ Could not find commit hook in D:\resources\"
    exit 1
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