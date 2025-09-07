package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

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
  --format string  Output format: table, json, csv (default "table")
  --state string   Filter dependencies by issue state: all, open, closed (default "all")
  --json string    Output JSON with specific fields (e.g., "blocked_by,blocks")`,
	Example: `  # List all dependencies for issue #123
  gh issue-dependency list 123

  # List dependencies for issue in a different repository  
  gh issue-dependency list 456 --repo owner/other-repo

  # Show detailed dependency information
  gh issue-dependency list 789 --detailed

  # Output dependencies as JSON for scripting
  gh issue-dependency list 123 --format json

  # Output specific JSON fields
  gh issue-dependency list 123 --json blocked_by,summary

  # Export dependencies to CSV for analysis
  gh issue-dependency list 456 --format csv > dependencies.csv

  # Show only open dependencies
  gh issue-dependency list 123 --state open

  # List closed dependencies in JSON format
  gh issue-dependency list 456 --state closed --format json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueNumber := args[0]
		
		// Resolve repository context using GitHub repository detection.
		// This handles both explicit --repo flags and automatic detection
		// from the current working directory's git remote.
		owner, repo, err := pkg.ResolveRepository(repoFlag, issueNumber)
		if err != nil {
			return err
		}
		
		// Parse and validate the issue number from the user input.
		// This supports both simple numbers (123) and full references (owner/repo#123).
		_, issueNum, err := pkg.ParseIssueReference(issueNumber)
		if err != nil {
			return err
		}

		// Validate the output format option against supported formats.
		// We support table (default), JSON, and CSV formats for different use cases.
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

		// Validate the state filter option against supported states.
		// We support all (default), open, and closed states for filtering dependencies.
		validStates := []string{"all", "open", "closed"}
		isValidState := false
		for _, state := range validStates {
			if listState == state {
				isValidState = true
				break
			}
		}
		if !isValidState {
			return pkg.NewAppError(
				pkg.ErrorTypeValidation,
				fmt.Sprintf("Invalid state: %s", listState),
				nil,
			).WithContext("state", listState).
				WithSuggestion("Use one of: all, open, closed")
		}

		// Fetch dependency data from GitHub API and display results
		// This replaces the placeholder output with real GitHub API integration
		return fetchAndDisplayDependencies(owner, repo, issueNum, listFormat, listState, listDetailed)
	},
}

// Flags for list command
var (
	// listDetailed controls whether to show detailed dependency information
	// including creation dates, users who created relationships, etc.
	listDetailed bool
	
	// listFormat specifies the output format for dependency information.
	// Supported formats: table (default), json, csv
	listFormat string
	
	// listState filters dependencies by issue state.
	// Supported states: all (default), open, closed
	listState string
	
	// listJSON specifies JSON fields for selective output
	// When set, overrides listFormat to use JSON with specific fields
	listJSON string
)

// fetchAndDisplayDependencies fetches real dependency data from GitHub API and displays it
// This function replaces the placeholder output with actual GitHub API integration
func fetchAndDisplayDependencies(owner, repo string, issueNum int, format, state string, detailed bool) error {
	// Create context with timeout for API calls
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	// Fetch dependency data from GitHub API
	originalData, err := pkg.FetchIssueDependencies(ctx, owner, repo, issueNum)
	if err != nil {
		return err
	}
	
	// Apply state filtering, keeping reference to original data
	filteredData := applyStateFilter(originalData, state)
	
	// Determine output format and create formatter
	outputOptions := pkg.DefaultOutputOptions()
	outputOptions.Detailed = detailed
	outputOptions.StateFilter = state
	outputOptions.OriginalData = originalData
	
	// Handle JSON field selection
	if listJSON != "" {
		outputOptions.Format = pkg.FormatJSON
		if listJSON != "true" { // If not just --json, parse fields
			fields := parseJSONFields(listJSON)
			outputOptions.JSONFields = fields
		}
	} else {
		// Use regular format flag
		switch format {
		case "json":
			outputOptions.Format = pkg.FormatJSON
		case "csv":
			outputOptions.Format = pkg.FormatCSV
		case "table":
			outputOptions.Format = pkg.FormatAuto // Auto-detect TTY vs plain
		default:
			outputOptions.Format = pkg.FormatAuto
		}
	}
	
	// Create formatter and display results
	formatter := pkg.NewOutputFormatter(outputOptions)
	return formatter.FormatOutput(filteredData)
}

// parseJSONFields parses the JSON fields specification
func parseJSONFields(fieldsStr string) []string {
	if fieldsStr == "" {
		return []string{}
	}
	
	// Split by comma and trim whitespace
	var fields []string
	for _, field := range strings.Split(fieldsStr, ",") {
		field = strings.TrimSpace(field)
		if field != "" {
			fields = append(fields, field)
		}
	}
	return fields
}

// applyStateFilter filters dependencies based on issue state
func applyStateFilter(data *pkg.DependencyData, state string) *pkg.DependencyData {
	if state == "all" {
		return data
	}
	
	// Create a copy to avoid modifying the original
	filtered := &pkg.DependencyData{
		SourceIssue: data.SourceIssue,
		BlockedBy:   []pkg.DependencyRelation{},
		Blocking:    []pkg.DependencyRelation{},
		FetchedAt:   data.FetchedAt,
	}
	
	// Filter blocked_by relationships
	for _, dep := range data.BlockedBy {
		if state == "all" || dep.Issue.State == state {
			filtered.BlockedBy = append(filtered.BlockedBy, dep)
		}
	}
	
	// Filter blocking relationships
	for _, dep := range data.Blocking {
		if state == "all" || dep.Issue.State == state {
			filtered.Blocking = append(filtered.Blocking, dep)
		}
	}
	
	// Update total count
	filtered.TotalCount = len(filtered.BlockedBy) + len(filtered.Blocking)
	
	return filtered
}

// init registers the list command with the root command and sets up its flags.
func init() {
	rootCmd.AddCommand(listCmd)

	// Local flags specific to the list command
	listCmd.Flags().BoolVar(&listDetailed, "detailed", false, "Show detailed dependency information including dates and users")
	listCmd.Flags().StringVar(&listFormat, "format", "table", "Output format: table (default), json, csv")
	listCmd.Flags().StringVar(&listState, "state", "all", "Filter dependencies by issue state: all (default), open, closed")
	listCmd.Flags().StringVar(&listJSON, "json", "", "Output JSON with specific fields: e.g. 'blocked_by,blocks' or 'summary'")
}