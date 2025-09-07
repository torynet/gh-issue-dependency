package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/torynet/gh-issue-dependency/pkg"
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

		// Parse and validate the main issue number
		_, issueNum, err := pkg.ParseIssueReference(issueNumber)
		if err != nil {
			return err
		}

		// Validate that exactly one of --blocked-by or --blocks is specified
		if addBlockedBy == "" && addBlocks == "" {
			return pkg.NewAppError(
				pkg.ErrorTypeValidation,
				"Must specify either --blocked-by or --blocks",
				nil,
			).WithSuggestion("Use --blocked-by to specify issues that block this one").
				WithSuggestion("Use --blocks to specify issues that this one blocks")
		}
		
		if addBlockedBy != "" && addBlocks != "" {
			return pkg.NewAppError(
				pkg.ErrorTypeValidation,
				"Cannot specify both --blocked-by and --blocks at the same time",
				nil,
			).WithSuggestion("Choose either --blocked-by or --blocks, not both")
		}

		// Parse dependency issue references
		var dependencyRefs []string
		if addBlockedBy != "" {
			dependencyRefs = strings.Split(addBlockedBy, ",")
		} else {
			dependencyRefs = strings.Split(addBlocks, ",")
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

		// TODO: Implement add functionality
		if addBlockedBy != "" {
			return pkg.WrapInternalError(
				"adding blocked-by dependencies", 
				fmt.Errorf("add command not implemented yet: issue %d blocked by %s", issueNum, addBlockedBy),
			).WithSuggestion("This feature is currently under development")
		} else {
			return pkg.WrapInternalError(
				"adding blocks dependencies",
				fmt.Errorf("add command not implemented yet: issue %d blocks %s", issueNum, addBlocks),
			).WithSuggestion("This feature is currently under development")
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