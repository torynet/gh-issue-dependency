package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
  --state string   Filter dependencies by issue state: all, open, closed (default "all")`,
	Example: `  # List all dependencies for issue #123
  gh issue-dependency list 123

  # List dependencies for issue in a different repository  
  gh issue-dependency list 456 --repo owner/other-repo

  # Show detailed dependency information
  gh issue-dependency list 789 --detailed

  # Output dependencies as JSON for scripting
  gh issue-dependency list 123 --format json

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
)

// fetchAndDisplayDependencies fetches real dependency data from GitHub API and displays it
// This function replaces the placeholder output with actual GitHub API integration
func fetchAndDisplayDependencies(owner, repo string, issueNum int, format, state string, detailed bool) error {
	// Create context with timeout for API calls
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	// Fetch dependency data from GitHub API
	data, err := pkg.FetchIssueDependencies(ctx, owner, repo, issueNum)
	if err != nil {
		return err
	}
	
	// Apply state filtering
	data = applyStateFilter(data, state)
	
	// Display in the requested format
	switch format {
	case "json":
		return displayJSONDependencies(data, detailed)
	case "csv":
		return displayCSVDependencies(data, detailed)
	default: // table
		return displayTableDependencies(data, detailed)
	}
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

// displayTableDependencies displays dependencies in human-readable table format
func displayTableDependencies(data *pkg.DependencyData, detailed bool) error {
	fmt.Printf("Dependencies for issue #%d: %s\n", 
		data.SourceIssue.Number, data.SourceIssue.Title)
	fmt.Printf("Repository: %s\n", data.SourceIssue.Repository)
	fmt.Printf("State: %s\n", data.SourceIssue.State)
	fmt.Printf("Total dependencies: %d\n\n", data.TotalCount)
	
	if len(data.BlockedBy) > 0 {
		fmt.Printf("BLOCKED BY (%d issues):\n", len(data.BlockedBy))
		for i, dep := range data.BlockedBy {
			fmt.Printf("  %d. #%d: %s [%s]", 
				i+1, dep.Issue.Number, dep.Issue.Title, dep.Issue.State)
			if dep.Repository != data.SourceIssue.Repository {
				fmt.Printf(" (%s)", dep.Repository)
			}
			fmt.Printf("\n")
			
			if detailed {
				displayDetailedIssueInfo(&dep.Issue, "     ")
			}
		}
		fmt.Printf("\n")
	}
	
	if len(data.Blocking) > 0 {
		fmt.Printf("BLOCKING (%d issues):\n", len(data.Blocking))
		for i, dep := range data.Blocking {
			fmt.Printf("  %d. #%d: %s [%s]", 
				i+1, dep.Issue.Number, dep.Issue.Title, dep.Issue.State)
			if dep.Repository != data.SourceIssue.Repository {
				fmt.Printf(" (%s)", dep.Repository)
			}
			fmt.Printf("\n")
			
			if detailed {
				displayDetailedIssueInfo(&dep.Issue, "     ")
			}
		}
		fmt.Printf("\n")
	}
	
	if data.TotalCount == 0 {
		fmt.Printf("No dependencies found for this issue.\n")
	}
	
	if detailed {
		fmt.Printf("Fetched at: %s\n", data.FetchedAt.Format(time.RFC3339))
	}
	
	return nil
}

// displayDetailedIssueInfo shows additional issue details for detailed view
func displayDetailedIssueInfo(issue *pkg.Issue, indent string) {
	if len(issue.Assignees) > 0 {
		fmt.Printf("%sAssignees: ", indent)
		for i, assignee := range issue.Assignees {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("@%s", assignee.Login)
		}
		fmt.Printf("\n")
	}
	
	if len(issue.Labels) > 0 {
		fmt.Printf("%sLabels: ", indent)
		for i, label := range issue.Labels {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", label.Name)
		}
		fmt.Printf("\n")
	}
	
	if issue.HTMLURL != "" {
		fmt.Printf("%sURL: %s\n", indent, issue.HTMLURL)
	}
}

// displayJSONDependencies outputs dependencies in JSON format
func displayJSONDependencies(data *pkg.DependencyData, detailed bool) error {
	// Create output structure
	output := map[string]interface{}{
		"source_issue": map[string]interface{}{
			"number":     data.SourceIssue.Number,
			"title":      data.SourceIssue.Title,
			"state":      data.SourceIssue.State,
			"repository": data.SourceIssue.Repository,
		},
		"blocked_by": formatDependenciesForJSON(data.BlockedBy, detailed),
		"blocking":   formatDependenciesForJSON(data.Blocking, detailed),
		"summary": map[string]interface{}{
			"total_count":     data.TotalCount,
			"blocked_by_count": len(data.BlockedBy),
			"blocking_count":   len(data.Blocking),
			"fetched_at":      data.FetchedAt.Format(time.RFC3339),
		},
	}
	
	if detailed {
		if sourceAssignees := formatAssigneesForJSON(data.SourceIssue.Assignees); len(sourceAssignees) > 0 {
			output["source_issue"].(map[string]interface{})["assignees"] = sourceAssignees
		}
		if sourceLabels := formatLabelsForJSON(data.SourceIssue.Labels); len(sourceLabels) > 0 {
			output["source_issue"].(map[string]interface{})["labels"] = sourceLabels
		}
		if data.SourceIssue.HTMLURL != "" {
			output["source_issue"].(map[string]interface{})["html_url"] = data.SourceIssue.HTMLURL
		}
	}
	
	// Use Go's JSON encoder for consistent formatting
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// formatDependenciesForJSON converts dependencies to JSON-friendly format
func formatDependenciesForJSON(deps []pkg.DependencyRelation, detailed bool) []map[string]interface{} {
	var result []map[string]interface{}
	
	for _, dep := range deps {
		item := map[string]interface{}{
			"number":     dep.Issue.Number,
			"title":      dep.Issue.Title,
			"state":      dep.Issue.State,
			"repository": dep.Repository,
		}
		
		if detailed {
			if assignees := formatAssigneesForJSON(dep.Issue.Assignees); len(assignees) > 0 {
				item["assignees"] = assignees
			}
			if labels := formatLabelsForJSON(dep.Issue.Labels); len(labels) > 0 {
				item["labels"] = labels
			}
			if dep.Issue.HTMLURL != "" {
				item["html_url"] = dep.Issue.HTMLURL
			}
		}
		
		result = append(result, item)
	}
	
	return result
}

// formatAssigneesForJSON converts assignees to JSON-friendly format
func formatAssigneesForJSON(assignees []pkg.User) []map[string]interface{} {
	var result []map[string]interface{}
	
	for _, user := range assignees {
		result = append(result, map[string]interface{}{
			"login": user.Login,
			"html_url": user.HTMLURL,
		})
	}
	
	return result
}

// formatLabelsForJSON converts labels to JSON-friendly format
func formatLabelsForJSON(labels []pkg.Label) []map[string]interface{} {
	var result []map[string]interface{}
	
	for _, label := range labels {
		item := map[string]interface{}{
			"name":  label.Name,
			"color": label.Color,
		}
		if label.Description != "" {
			item["description"] = label.Description
		}
		result = append(result, item)
	}
	
	return result
}

// displayCSVDependencies outputs dependencies in CSV format
func displayCSVDependencies(data *pkg.DependencyData, detailed bool) error {
	// CSV header
	if detailed {
		fmt.Printf("type,repository,number,title,state,assignees,labels,html_url\n")
	} else {
		fmt.Printf("type,repository,number,title,state\n")
	}
	
	// Source issue
	fmt.Printf("source,%s,%d,%s,%s", 
		escapeCSV(data.SourceIssue.Repository),
		data.SourceIssue.Number,
		escapeCSV(data.SourceIssue.Title),
		data.SourceIssue.State)
	
	if detailed {
		fmt.Printf(",%s,%s,%s",
			escapeCSV(formatAssigneesForCSV(data.SourceIssue.Assignees)),
			escapeCSV(formatLabelsForCSV(data.SourceIssue.Labels)),
			escapeCSV(data.SourceIssue.HTMLURL))
	}
	fmt.Printf("\n")
	
	// Blocked by dependencies
	for _, dep := range data.BlockedBy {
		fmt.Printf("blocked_by,%s,%d,%s,%s", 
			escapeCSV(dep.Repository),
			dep.Issue.Number,
			escapeCSV(dep.Issue.Title),
			dep.Issue.State)
		
		if detailed {
			fmt.Printf(",%s,%s,%s",
				escapeCSV(formatAssigneesForCSV(dep.Issue.Assignees)),
				escapeCSV(formatLabelsForCSV(dep.Issue.Labels)),
				escapeCSV(dep.Issue.HTMLURL))
		}
		fmt.Printf("\n")
	}
	
	// Blocking dependencies
	for _, dep := range data.Blocking {
		fmt.Printf("blocking,%s,%d,%s,%s", 
			escapeCSV(dep.Repository),
			dep.Issue.Number,
			escapeCSV(dep.Issue.Title),
			dep.Issue.State)
		
		if detailed {
			fmt.Printf(",%s,%s,%s",
				escapeCSV(formatAssigneesForCSV(dep.Issue.Assignees)),
				escapeCSV(formatLabelsForCSV(dep.Issue.Labels)),
				escapeCSV(dep.Issue.HTMLURL))
		}
		fmt.Printf("\n")
	}
	
	return nil
}

// formatAssigneesForCSV converts assignees to CSV-friendly string
func formatAssigneesForCSV(assignees []pkg.User) string {
	if len(assignees) == 0 {
		return ""
	}
	
	var names []string
	for _, user := range assignees {
		names = append(names, "@"+user.Login)
	}
	return joinStrings(names, "; ")
}

// formatLabelsForCSV converts labels to CSV-friendly string
func formatLabelsForCSV(labels []pkg.Label) string {
	if len(labels) == 0 {
		return ""
	}
	
	var names []string
	for _, label := range labels {
		names = append(names, label.Name)
	}
	return joinStrings(names, "; ")
}

// escapeCSV escapes special characters in CSV values
func escapeCSV(value string) string {
	// If value contains comma, quote, or newline, wrap in quotes and escape internal quotes
	needsQuoting := false
	for _, char := range value {
		if char == ',' || char == '"' || char == '\n' || char == '\r' {
			needsQuoting = true
			break
		}
	}
	
	if needsQuoting {
		// Replace internal quotes with double quotes
		escaped := ""
		for _, char := range value {
			if char == '"' {
				escaped += "\"\""
			} else {
				escaped += string(char)
			}
		}
		return "\"" + escaped + "\""
	}
	
	return value
}

// joinStrings joins a slice of strings with a separator
func joinStrings(items []string, separator string) string {
	if len(items) == 0 {
		return ""
	}
	
	result := items[0]
	for i := 1; i < len(items); i++ {
		result += separator + items[i]
	}
	return result
}

// init registers the list command with the root command and sets up its flags.
func init() {
	rootCmd.AddCommand(listCmd)

	// Local flags specific to the list command
	listCmd.Flags().BoolVar(&listDetailed, "detailed", false, "Show detailed dependency information including dates and users")
	listCmd.Flags().StringVar(&listFormat, "format", "table", "Output format: table (default), json, csv")
	listCmd.Flags().StringVar(&listState, "state", "all", "Filter dependencies by issue state: all (default), open, closed")
}