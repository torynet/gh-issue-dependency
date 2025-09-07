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
	Short: "Remove a dependency relationship between issues",
	Long: `Remove an existing dependency relationship between two issues. You must specify either:
- --blocked-by to remove a "blocked by" relationship
- --blocks to remove a "blocks" relationship

This removes the dependency link using GitHub's native dependency API.
Works for both same-repository and cross-repository dependencies.`,
	Example: `  # Remove issue 123 being blocked by issue 456
  gh issue-dependency remove 123 --blocked-by 456

  # Remove issue 123 blocking issue 789
  gh issue-dependency remove 123 --blocks 789

  # Remove cross-repository dependency
  gh issue-dependency remove 123 --blocked-by owner/other-repo#456

  # Remove multiple dependencies
  gh issue-dependency remove 123 --blocked-by 456,789`,
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
	removeBlockedBy string
	removeBlocks    string
)

func init() {
	rootCmd.AddCommand(removeCmd)

	// Local flags for remove command
	removeCmd.Flags().StringVar(&removeBlockedBy, "blocked-by", "", "Issue number(s) to remove from blocking this issue (comma-separated)")
	removeCmd.Flags().StringVar(&removeBlocks, "blocks", "", "Issue number(s) to remove from being blocked by this issue (comma-separated)")
}