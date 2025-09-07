package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add <issue-number>",
	Short: "Add a dependency relationship between issues",
	Long: `Add a dependency relationship between two issues. You must specify either:
- --blocked-by to indicate the issue is blocked by another issue
- --blocks to indicate the issue blocks another issue

This creates a dependency link using GitHub's native dependency API.
Dependencies can be within the same repository or cross-repository.`,
	Example: `  # Add issue 123 as blocked by issue 456
  gh issue-dependency add 123 --blocked-by 456

  # Add issue 123 as blocking issue 789
  gh issue-dependency add 123 --blocks 789

  # Add cross-repository dependency
  gh issue-dependency add 123 --blocked-by owner/other-repo#456

  # Add multiple dependencies
  gh issue-dependency add 123 --blocked-by 456,789`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueNumber := args[0]

		// Validate issue number format
		if issueNumber == "" {
			return fmt.Errorf("issue number cannot be empty")
		}

		// Validate that exactly one of --blocked-by or --blocks is specified
		if addBlockedBy == "" && addBlocks == "" {
			return fmt.Errorf("must specify either --blocked-by or --blocks")
		}
		
		if addBlockedBy != "" && addBlocks != "" {
			return fmt.Errorf("cannot specify both --blocked-by and --blocks")
		}

		// TODO: Implement add functionality
		if addBlockedBy != "" {
			return fmt.Errorf("add command not implemented yet: issue %s blocked by %s", issueNumber, addBlockedBy)
		} else {
			return fmt.Errorf("add command not implemented yet: issue %s blocks %s", issueNumber, addBlocks)
		}
	},
}

// Flags for add command
var (
	addBlockedBy string
	addBlocks    string
)

func init() {
	rootCmd.AddCommand(addCmd)

	// Local flags for add command
	addCmd.Flags().StringVar(&addBlockedBy, "blocked-by", "", "Issue number(s) that block this issue (comma-separated)")
	addCmd.Flags().StringVar(&addBlocks, "blocks", "", "Issue number(s) that this issue blocks (comma-separated)")
}