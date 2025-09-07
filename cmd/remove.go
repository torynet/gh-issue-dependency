package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/torynet/gh-issue-dependency/pkg"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove <issue-number>",
	Short: "Remove dependency relationships between GitHub issues",
	Long: `Remove existing dependency relationships between GitHub issues with validation and confirmation.

RELATIONSHIP TYPES TO REMOVE
You must specify exactly one of the following relationship types:

  --blocked-by   Remove specific issues that are blocking the specified issue
                 (removes the "blocked by" relationship)
                 
  --blocks       Remove specific issues that are blocked by the specified issue
                 (removes the "blocks" relationship)
                 
  --all          Remove all dependency relationships for the specified issue
                 (removes both "blocked by" and "blocks" relationships)

ISSUE REFERENCES
Issues can be referenced in multiple ways:
  • Simple number: 123 (same repository)
  • Full reference: owner/repo#123 (cross-repository)  
  • GitHub issue URL: https://github.com/owner/repo/issues/123
  • Multiple issues: 123,456,789 (comma-separated, no spaces)

SAFETY AND CONFIRMATION
The command will:
  • Validate that the specified relationships exist before attempting removal
  • Show which relationships will be removed before making changes
  • Prompt for confirmation unless --force is specified
  • Support dry-run mode with --dry-run to preview changes without executing
  • Fail gracefully if any referenced issues don't exist or aren't accessible

Note: Removing a dependency relationship does not affect the issues themselves,
only the dependency links between them.

FLAGS
  --blocked-by string   Issue number(s) to remove from blocking this issue (comma-separated)
  --blocks string       Issue number(s) to remove from being blocked by this issue (comma-separated)  
  --all                 Remove all dependency relationships for this issue
  --dry-run            Show what would be removed without making changes
  --force              Skip confirmation prompts`,
	Example: `  # Remove issue #456 from blocking issue #123
  gh issue-dependency remove 123 --blocked-by 456

  # Remove issue #789 from being blocked by issue #123  
  gh issue-dependency remove 123 --blocks 789

  # Remove all dependency relationships for issue #123
  gh issue-dependency remove 123 --all

  # Preview what would be removed without making changes
  gh issue-dependency remove 123 --blocked-by 456 --dry-run

  # Remove dependencies without confirmation prompts
  gh issue-dependency remove 123 --blocks 789 --force

  # Remove cross-repository dependency
  gh issue-dependency remove 123 --blocked-by owner/other-repo#456

  # Remove multiple dependencies at once
  gh issue-dependency remove 123 --blocked-by 456,789,101

  # Work with issues in a different repository
  gh issue-dependency remove 123 --blocks 456 --repo owner/other-repo`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueNumber := args[0]

		// Parse and validate the main issue number
		_, issueNum, err := pkg.ParseIssueReference(issueNumber)
		if err != nil {
			return err
		}

		// Validate flag mutual exclusion
		flagCount := 0
		if removeBlockedBy != "" {
			flagCount++
		}
		if removeBlocks != "" {
			flagCount++
		}
		if removeAll {
			flagCount++
		}

		if flagCount == 0 {
			return pkg.NewAppError(
				pkg.ErrorTypeValidation,
				"Must specify exactly one of --blocked-by, --blocks, or --all",
				nil,
			).WithSuggestion("Use --blocked-by to remove issues that block this one").
				WithSuggestion("Use --blocks to remove issues that this one blocks").
				WithSuggestion("Use --all to remove all dependency relationships")
		}
		
		if flagCount > 1 {
			return pkg.NewAppError(
				pkg.ErrorTypeValidation,
				"Cannot specify multiple relationship flags at the same time",
				nil,
			).WithSuggestion("Choose exactly one of --blocked-by, --blocks, or --all")
		}

		// Parse and validate dependency references (only for specific removals)
		var dependencyRefs []string
		var relationType string

		if removeBlockedBy != "" {
			dependencyRefs = strings.Split(removeBlockedBy, ",")
			relationType = "blocked-by"
			// Validate all dependency references
			for _, ref := range dependencyRefs {
				ref = strings.TrimSpace(ref)
				if ref == "" {
					continue
				}
				_, _, err := pkg.ParseIssueReference(ref)
				if err != nil {
					return err
				}
			}
		} else if removeBlocks != "" {
			dependencyRefs = strings.Split(removeBlocks, ",")
			relationType = "blocks"
			// Validate all dependency references
			for _, ref := range dependencyRefs {
				ref = strings.TrimSpace(ref)
				if ref == "" {
					continue
				}
				_, _, err := pkg.ParseIssueReference(ref)
				if err != nil {
					return err
				}
			}
		} else if removeAll {
			relationType = "all"
		}

		// Display removal preview if dry-run mode
		if dryRun {
			return pkg.WrapInternalError(
				"dry-run mode",
				fmt.Errorf("dry-run mode not implemented yet: would remove %s relationships for issue %d", relationType, issueNum),
			).WithSuggestion("This feature is currently under development")
		}

		// TODO: Implement remove functionality based on relationType and force flag
		if removeAll {
			return pkg.WrapInternalError(
				"removing all dependencies",
				fmt.Errorf("remove command not implemented yet: removing all relationships for issue %d", issueNum),
			).WithSuggestion("This feature is currently under development")
		} else if removeBlockedBy != "" {
			return pkg.WrapInternalError(
				"removing blocked-by dependencies",
				fmt.Errorf("remove command not implemented yet: removing issue %d blocked by %s", issueNum, removeBlockedBy),
			).WithSuggestion("This feature is currently under development")
		} else {
			return pkg.WrapInternalError(
				"removing blocks dependencies",
				fmt.Errorf("remove command not implemented yet: removing issue %d blocks %s", issueNum, removeBlocks),
			).WithSuggestion("This feature is currently under development")
		}
	},
}

// Flags for remove command
var (
	// removeBlockedBy contains a comma-separated list of issue references to remove
	// from blocking the target issue. This will remove the "blocked by" relationship.
	removeBlockedBy string
	
	// removeBlocks contains a comma-separated list of issue references to remove
	// from being blocked by the target issue. This will remove the "blocks" relationship.
	removeBlocks string

	// removeAll indicates whether to remove all dependency relationships for the target issue
	removeAll bool

	// dryRun indicates whether to show what would be removed without making changes
	dryRun bool

	// force indicates whether to skip confirmation prompts during removal
	force bool
)

// init registers the remove command with the root command and sets up its flags.
func init() {
	rootCmd.AddCommand(removeCmd)

	// Local flags specific to the remove command
	// Note: --blocked-by, --blocks, and --all are mutually exclusive - validation happens in the command logic
	removeCmd.Flags().StringVar(&removeBlockedBy, "blocked-by", "", "Issue number(s) to remove from blocking this issue (comma-separated)")
	removeCmd.Flags().StringVar(&removeBlocks, "blocks", "", "Issue number(s) to remove from being blocked by this issue (comma-separated)")
	removeCmd.Flags().BoolVar(&removeAll, "all", false, "Remove all dependency relationships for this issue")
	removeCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be removed without making changes")
	removeCmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompts")
}