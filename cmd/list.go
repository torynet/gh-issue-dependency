package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/torynet/gh-issue-dependency/pkg"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list <issue-number>",
	Short: "List issue dependencies and relationships",
	Long: `List all dependencies for the specified issue, showing both blocking and blocked-by relationships.

DEPENDENCIES SHOWN
  • Blocking issues: Issues that must be resolved before this issue can be completed
  • Blocked issues: Issues that are waiting for this issue to be completed  
  • Cross-repository dependencies when applicable

OUTPUT FORMATS
  • table (default): Human-readable table format with issue titles and states
  • json: Machine-readable JSON for scripting and integration
  • csv: Comma-separated values for spreadsheet import

The output includes issue numbers, repository information, titles, current status,
and relationship type (blocking vs blocked).

FLAGS
  --detailed       Show detailed dependency information including dates and users
  --format string  Output format: table, json, csv (default "table")`,
	Example: `  # List all dependencies for issue #123
  gh issue-dependency list 123

  # List dependencies for issue in a different repository  
  gh issue-dependency list 456 --repo owner/other-repo

  # Show detailed dependency information
  gh issue-dependency list 789 --detailed

  # Output dependencies as JSON for scripting
  gh issue-dependency list 123 --format json

  # Export dependencies to CSV for analysis
  gh issue-dependency list 456 --format csv > dependencies.csv`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueNumber := args[0]
		
		// Resolve repository context using the new github detection
		owner, repo, err := pkg.ResolveRepository(repoFlag, issueNumber)
		if err != nil {
			return err
		}
		
		// Parse and validate the issue number
		_, issueNum, err := pkg.ParseIssueReference(issueNumber)
		if err != nil {
			return err
		}

		// Validate format option
		validFormats := []string{"table", "json", "csv"}
		isValidFormat := false
		for _, format := range validFormats {
			if listFormat == format {
				isValidFormat = true
				break
			}
		}
		if !isValidFormat {
			return pkg.NewAppError(
				pkg.ErrorTypeValidation,
				fmt.Sprintf("Invalid format: %s", listFormat),
				nil,
			).WithContext("format", listFormat).
				WithSuggestion("Use one of: table, json, csv")
		}

		// TODO: Implement list functionality
		return pkg.WrapInternalError(
			"listing dependencies",
			fmt.Errorf("list command not implemented yet for issue #%d in %s/%s", issueNum, owner, repo),
		).WithSuggestion("This feature is currently under development").
			WithContext("repository", fmt.Sprintf("%s/%s", owner, repo)).
			WithContext("issue", fmt.Sprintf("#%d", issueNum))
	},
}

// Flags for list command
var (
	listDetailed bool
	listFormat   string
)

func init() {
	rootCmd.AddCommand(listCmd)

	// Local flags for list command
	listCmd.Flags().BoolVar(&listDetailed, "detailed", false, "Show detailed dependency information")
	listCmd.Flags().StringVar(&listFormat, "format", "table", "Output format (table, json, csv)")
}