package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list <issue-number>",
	Short: "List dependencies for an issue",
	Long: `List all dependencies for the specified issue, showing both blocking and blocked-by relationships.

This command displays:
- Issues that are blocking the specified issue (dependencies)
- Issues that are blocked by the specified issue (dependents)
- Cross-repository dependencies when applicable

The output includes issue numbers, titles, and current status.`,
	Example: `  # List dependencies for issue 123
  gh issue-dependency list 123

  # List dependencies for issue in a specific repository  
  gh issue-dependency list 456 --repo owner/repo

  # List dependencies with detailed output
  gh issue-dependency list 789 --detailed`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueNumber := args[0]
		
		// Validate issue number format
		if issueNumber == "" {
			return fmt.Errorf("issue number cannot be empty")
		}

		// TODO: Implement list functionality
		return fmt.Errorf("list command not implemented yet for issue %s", issueNumber)
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