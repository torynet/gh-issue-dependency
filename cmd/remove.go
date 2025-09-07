package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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

		// Validate issue number format
		if issueNumber == "" {
			return fmt.Errorf("issue number cannot be empty")
		}

		// Validate that exactly one of --blocked-by or --blocks is specified
		if removeBlockedBy == "" && removeBlocks == "" {
			return fmt.Errorf("must specify either --blocked-by or --blocks")
		}
		
		if removeBlockedBy != "" && removeBlocks != "" {
			return fmt.Errorf("cannot specify both --blocked-by and --blocks")
		}

		// TODO: Implement remove functionality
		if removeBlockedBy != "" {
			return fmt.Errorf("remove command not implemented yet: removing issue %s blocked by %s", issueNumber, removeBlockedBy)
		} else {
			return fmt.Errorf("remove command not implemented yet: removing issue %s blocks %s", issueNumber, removeBlocks)
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