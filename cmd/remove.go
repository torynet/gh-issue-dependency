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
	Short: "Remove dependency relationships between issues",
	Long: `Remove existing dependency relationships between issues using GitHub's native dependency API.

RELATIONSHIP TYPES TO REMOVE
You must specify exactly one of the following relationship types:

  --blocked-by   Remove issues that are blocking the specified issue
                 (removes the "blocked by" relationship)
                 
  --blocks       Remove issues that are blocked by the specified issue
                 (removes the "blocks" relationship)

ISSUE REFERENCES
Issues can be referenced in multiple ways:
  • Simple number: 123 (same repository)
  • Full reference: owner/repo#123 (cross-repository)  
  • Multiple issues: 123,456,789 (comma-separated, no spaces)

SAFETY CONSIDERATIONS
The command will:
  • Validate that the specified relationships exist before attempting removal
  • Show which relationships will be removed before making changes
  • Fail gracefully if any referenced issues don't exist or aren't accessible
  • Not require confirmation by default (relationships can be easily re-added)

Note: Removing a dependency relationship does not affect the issues themselves,
only the dependency links between them.

FLAGS
  --blocked-by string   Issue number(s) to remove from blocking this issue (comma-separated)
  --blocks string       Issue number(s) to remove from being blocked by this issue (comma-separated)`,
	Example: `  # Remove issue #456 from blocking issue #123
  gh issue-dependency remove 123 --blocked-by 456

  # Remove issue #789 from being blocked by issue #123  
  gh issue-dependency remove 123 --blocks 789

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

		// Validate that exactly one of --blocked-by or --blocks is specified
		if removeBlockedBy == "" && removeBlocks == "" {
			return pkg.NewAppError(
				pkg.ErrorTypeValidation,
				"Must specify either --blocked-by or --blocks",
				nil,
			).WithSuggestion("Use --blocked-by to remove issues that block this one").
				WithSuggestion("Use --blocks to remove issues that this one blocks")
		}
		
		if removeBlockedBy != "" && removeBlocks != "" {
			return pkg.NewAppError(
				pkg.ErrorTypeValidation,
				"Cannot specify both --blocked-by and --blocks at the same time",
				nil,
			).WithSuggestion("Choose either --blocked-by or --blocks, not both")
		}

		// Parse dependency issue references
		var dependencyRefs []string
		if removeBlockedBy != "" {
			dependencyRefs = strings.Split(removeBlockedBy, ",")
		} else {
			dependencyRefs = strings.Split(removeBlocks, ",")
		}

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

		// TODO: Implement remove functionality
		if removeBlockedBy != "" {
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
)

// init registers the remove command with the root command and sets up its flags.
func init() {
	rootCmd.AddCommand(removeCmd)

	// Local flags specific to the remove command
	// Note: These flags are mutually exclusive - validation happens in the command logic
	removeCmd.Flags().StringVar(&removeBlockedBy, "blocked-by", "", "Issue number(s) to remove from blocking this issue (comma-separated)")
	removeCmd.Flags().StringVar(&removeBlocks, "blocks", "", "Issue number(s) to remove from being blocked by this issue (comma-separated)")
}