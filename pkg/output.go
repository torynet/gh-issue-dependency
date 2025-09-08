// Package pkg provides output formatting for dependency data
package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/muesli/termenv"
)

// OutputFormat represents the different output formats supported
type OutputFormat int

const (
	FormatAuto  OutputFormat = iota // Auto-detect based on TTY
	FormatTTY                       // Rich TTY output with colors and emojis
	FormatPlain                     // Plain text output
	FormatJSON                      // JSON output
	FormatCSV                       // CSV output
)

// OutputOptions contains configuration for output formatting
type OutputOptions struct {
	Format       OutputFormat
	JSONFields   []string // Specific fields to include in JSON output
	Detailed     bool     // Include detailed information
	Writer       io.Writer
	StateFilter  string          // Applied state filter for context-aware messaging
	OriginalData *DependencyData // Original data before filtering for comparison
}

// DefaultOutputOptions returns sensible defaults for output options
func DefaultOutputOptions() *OutputOptions {
	return &OutputOptions{
		Format:       FormatAuto,
		JSONFields:   []string{},
		Detailed:     false,
		Writer:       os.Stdout,
		StateFilter:  "all",
		OriginalData: nil,
	}
}

// OutputFormatter handles all output formatting operations
type OutputFormatter struct {
	options *OutputOptions
	output  *termenv.Output
}

// NewOutputFormatter creates a new output formatter with the given options
func NewOutputFormatter(options *OutputOptions) *OutputFormatter {
	if options == nil {
		options = DefaultOutputOptions()
	}

	return &OutputFormatter{
		options: options,
		output:  termenv.NewOutput(options.Writer),
	}
}

// IsTerminal detects if the output is going to a terminal/TTY
func IsTerminal() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}

// determineFormat resolves the actual format to use based on options and environment
func (f *OutputFormatter) determineFormat() OutputFormat {
	switch f.options.Format {
	case FormatAuto:
		if IsTerminal() {
			return FormatTTY
		}
		return FormatPlain
	default:
		return f.options.Format
	}
}

// FormatOutput formats dependency data according to the configured output format
func (f *OutputFormatter) FormatOutput(data *DependencyData) error {
	format := f.determineFormat()

	// Add filter context if applicable
	f.addFilterContext(data)

	switch format {
	case FormatTTY:
		return f.formatTTYOutput(data)
	case FormatPlain:
		return f.formatPlainOutput(data)
	case FormatJSON:
		return f.formatJSONOutput(data)
	case FormatCSV:
		return f.formatCSVOutput(data)
	default:
		return f.formatPlainOutput(data)
	}
}

// formatTTYOutput formats output for TTY with colors, emojis, and visual enhancements
func (f *OutputFormatter) formatTTYOutput(data *DependencyData) error {
	// Color functions
	title := f.colorize(termenv.ANSIBrightBlue)
	header := f.colorize(termenv.ANSIYellow)
	separator := f.colorize(termenv.ANSIBrightBlack)
	info := f.colorize(termenv.ANSIBlue)
	muted := f.colorize(termenv.ANSIBrightBlack)

	// Issue title and header
	fmt.Fprintf(f.options.Writer, "%s\nDependencies for: #%d - %s%s\n",
		title(""), data.SourceIssue.Number, data.SourceIssue.Title, termenv.CSI+termenv.ResetSeq)

	// Repository context
	if !data.SourceIssue.Repository.IsEmpty() {
		fmt.Fprintf(f.options.Writer, "%sRepository: %s%s\n",
			muted(""), data.SourceIssue.Repository, termenv.CSI+termenv.ResetSeq)
	}

	fmt.Fprintf(f.options.Writer, "\n")

	// BLOCKED BY section
	if len(data.BlockedBy) > 0 {
		fmt.Fprintf(f.options.Writer, "%sBLOCKED BY (%d issues)%s\n",
			header(""), len(data.BlockedBy), termenv.CSI+termenv.ResetSeq)
		fmt.Fprintf(f.options.Writer, "%s────────────────────%s\n",
			separator(""), termenv.CSI+termenv.ResetSeq)

		for _, dep := range data.BlockedBy {
			emoji := f.getStateEmoji(dep.Issue.State)
			stateColor := f.getStateColor(dep.Issue.State)

			fmt.Fprintf(f.options.Writer, "%s #%-6d %s %s[%s]%s",
				emoji, dep.Issue.Number, dep.Issue.Title,
				stateColor(""), dep.Issue.State, termenv.CSI+termenv.ResetSeq)

			// Show assignees if available
			if len(dep.Issue.Assignees) > 0 {
				assigneeNames := make([]string, len(dep.Issue.Assignees))
				for i, assignee := range dep.Issue.Assignees {
					assigneeNames[i] = "@" + assignee.Login
				}
				fmt.Fprintf(f.options.Writer, " %s%s%s",
					info(""), strings.Join(assigneeNames, ", "), termenv.CSI+termenv.ResetSeq)
			}

			// Show repository context for cross-repo dependencies
			if dep.Repository != data.SourceIssue.Repository.String() {
				fmt.Fprintf(f.options.Writer, "\n         %s%s%s",
					muted(""), dep.Repository, termenv.CSI+termenv.ResetSeq)
			}

			// Show URL for easy navigation
			if dep.Issue.HTMLURL != "" {
				fmt.Fprintf(f.options.Writer, "\n         %s%s%s",
					muted(""), dep.Issue.HTMLURL, termenv.CSI+termenv.ResetSeq)
			}

			fmt.Fprintf(f.options.Writer, "\n")
		}
		fmt.Fprintf(f.options.Writer, "\n")
	}

	// BLOCKS section
	if len(data.Blocking) > 0 {
		fmt.Fprintf(f.options.Writer, "%sBLOCKS (%d issues)%s\n",
			header(""), len(data.Blocking), termenv.CSI+termenv.ResetSeq)
		fmt.Fprintf(f.options.Writer, "%s──────────────────%s\n",
			separator(""), termenv.CSI+termenv.ResetSeq)

		for _, dep := range data.Blocking {
			emoji := f.getStateEmoji(dep.Issue.State)
			stateColor := f.getStateColor(dep.Issue.State)

			fmt.Fprintf(f.options.Writer, "%s #%-6d %s %s[%s]%s",
				emoji, dep.Issue.Number, dep.Issue.Title,
				stateColor(""), dep.Issue.State, termenv.CSI+termenv.ResetSeq)

			// Show assignees if available
			if len(dep.Issue.Assignees) > 0 {
				assigneeNames := make([]string, len(dep.Issue.Assignees))
				for i, assignee := range dep.Issue.Assignees {
					assigneeNames[i] = "@" + assignee.Login
				}
				fmt.Fprintf(f.options.Writer, " %s%s%s",
					info(""), strings.Join(assigneeNames, ", "), termenv.CSI+termenv.ResetSeq)
			}

			// Show repository context for cross-repo dependencies
			if dep.Repository != data.SourceIssue.Repository.String() {
				fmt.Fprintf(f.options.Writer, "\n         %s%s%s",
					muted(""), dep.Repository, termenv.CSI+termenv.ResetSeq)
			}

			// Show URL for easy navigation
			if dep.Issue.HTMLURL != "" {
				fmt.Fprintf(f.options.Writer, "\n         %s%s%s",
					muted(""), dep.Issue.HTMLURL, termenv.CSI+termenv.ResetSeq)
			}

			fmt.Fprintf(f.options.Writer, "\n")
		}
		fmt.Fprintf(f.options.Writer, "\n")
	}

	// Empty state handling
	if data.TotalCount == 0 {
		mainMsg, tipMsg := f.getEmptyStateMessage(data)
		fmt.Fprintf(f.options.Writer, "%s💡 %s%s\n\n",
			info(""), mainMsg, termenv.CSI+termenv.ResetSeq)
		fmt.Fprintf(f.options.Writer, "%s%s%s\n",
			muted(""), tipMsg, termenv.CSI+termenv.ResetSeq)
	}

	// Footer with metadata
	if f.options.Detailed {
		fmt.Fprintf(f.options.Writer, "%sFetched at: %s%s\n",
			muted(""), data.FetchedAt.Format(time.RFC3339), termenv.CSI+termenv.ResetSeq)
	}

	return nil
}

// formatPlainOutput formats output for plain text without colors or emojis
func (f *OutputFormatter) formatPlainOutput(data *DependencyData) error {
	// Header
	fmt.Fprintf(f.options.Writer, "Dependencies for: #%d - %s\n",
		data.SourceIssue.Number, data.SourceIssue.Title)

	if !data.SourceIssue.Repository.IsEmpty() {
		fmt.Fprintf(f.options.Writer, "Repository: %s\n", data.SourceIssue.Repository)
	}
	fmt.Fprintf(f.options.Writer, "\n")

	// BLOCKED BY section
	if len(data.BlockedBy) > 0 {
		fmt.Fprintf(f.options.Writer, "BLOCKED BY (%d issues)\n", len(data.BlockedBy))
		fmt.Fprintf(f.options.Writer, "========================\n")

		for _, dep := range data.BlockedBy {
			fmt.Fprintf(f.options.Writer, "#%d %s [%s]",
				dep.Issue.Number, dep.Issue.Title, dep.Issue.State)

			// Show assignees if available
			if len(dep.Issue.Assignees) > 0 {
				assigneeNames := make([]string, len(dep.Issue.Assignees))
				for i, assignee := range dep.Issue.Assignees {
					assigneeNames[i] = "@" + assignee.Login
				}
				fmt.Fprintf(f.options.Writer, " %s", strings.Join(assigneeNames, ", "))
			}

			// Show repository context for cross-repo dependencies
			if dep.Repository != data.SourceIssue.Repository.String() {
				fmt.Fprintf(f.options.Writer, "\n       Repository: %s", dep.Repository)
			}

			// Show URL for easy navigation
			if dep.Issue.HTMLURL != "" {
				fmt.Fprintf(f.options.Writer, "\n       URL: %s", dep.Issue.HTMLURL)
			}

			fmt.Fprintf(f.options.Writer, "\n")
		}
		fmt.Fprintf(f.options.Writer, "\n")
	}

	// BLOCKS section
	if len(data.Blocking) > 0 {
		fmt.Fprintf(f.options.Writer, "BLOCKS (%d issues)\n", len(data.Blocking))
		fmt.Fprintf(f.options.Writer, "==================\n")

		for _, dep := range data.Blocking {
			fmt.Fprintf(f.options.Writer, "#%d %s [%s]",
				dep.Issue.Number, dep.Issue.Title, dep.Issue.State)

			// Show assignees if available
			if len(dep.Issue.Assignees) > 0 {
				assigneeNames := make([]string, len(dep.Issue.Assignees))
				for i, assignee := range dep.Issue.Assignees {
					assigneeNames[i] = "@" + assignee.Login
				}
				fmt.Fprintf(f.options.Writer, " %s", strings.Join(assigneeNames, ", "))
			}

			// Show repository context for cross-repo dependencies
			if dep.Repository != data.SourceIssue.Repository.String() {
				fmt.Fprintf(f.options.Writer, "\n       Repository: %s", dep.Repository)
			}

			// Show URL for easy navigation
			if dep.Issue.HTMLURL != "" {
				fmt.Fprintf(f.options.Writer, "\n       URL: %s", dep.Issue.HTMLURL)
			}

			fmt.Fprintf(f.options.Writer, "\n")
		}
		fmt.Fprintf(f.options.Writer, "\n")
	}

	// Empty state handling
	if data.TotalCount == 0 {
		mainMsg, tipMsg := f.getEmptyStateMessage(data)
		fmt.Fprintf(f.options.Writer, "%s\n\n", mainMsg)
		fmt.Fprintf(f.options.Writer, "%s\n", tipMsg)
	}

	// Footer with metadata
	if f.options.Detailed {
		fmt.Fprintf(f.options.Writer, "Fetched at: %s\n", data.FetchedAt.Format(time.RFC3339))
	}

	return nil
}

// formatJSONOutput formats output as JSON with optional field selection
func (f *OutputFormatter) formatJSONOutput(data *DependencyData) error {
	// Create output structure
	output := map[string]interface{}{
		"source_issue": f.formatIssueForJSON(&data.SourceIssue),
		"blocked_by":   f.formatDependenciesForJSON(data.BlockedBy),
		"blocks":       f.formatDependenciesForJSON(data.Blocking),
		"summary": map[string]interface{}{
			"total_count":      data.TotalCount,
			"blocked_by_count": len(data.BlockedBy),
			"blocks_count":     len(data.Blocking),
			"fetched_at":       data.FetchedAt.Format(time.RFC3339),
		},
	}

	// Apply field selection if specified
	if len(f.options.JSONFields) > 0 {
		filtered := make(map[string]interface{})
		for _, field := range f.options.JSONFields {
			if value, exists := output[field]; exists {
				filtered[field] = value
			}
		}
		output = filtered
	}

	// Use Go's JSON encoder for consistent formatting
	encoder := json.NewEncoder(f.options.Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// formatCSVOutput formats output as CSV (reuse existing implementation)
func (f *OutputFormatter) formatCSVOutput(data *DependencyData) error {
	// CSV header
	if f.options.Detailed {
		fmt.Fprintf(f.options.Writer, "type,repository,number,title,state,assignees,labels,html_url\n")
	} else {
		fmt.Fprintf(f.options.Writer, "type,repository,number,title,state\n")
	}

	// Source issue
	fmt.Fprintf(f.options.Writer, "source,%s,%d,%s,%s",
		escapeCSV(data.SourceIssue.Repository.String()),
		data.SourceIssue.Number,
		escapeCSV(data.SourceIssue.Title),
		data.SourceIssue.State)

	if f.options.Detailed {
		fmt.Fprintf(f.options.Writer, ",%s,%s,%s",
			escapeCSV(formatAssigneesForCSV(data.SourceIssue.Assignees)),
			escapeCSV(formatLabelsForCSV(data.SourceIssue.Labels)),
			escapeCSV(data.SourceIssue.HTMLURL))
	}
	fmt.Fprintf(f.options.Writer, "\n")

	// Blocked by dependencies
	for _, dep := range data.BlockedBy {
		fmt.Fprintf(f.options.Writer, "blocked_by,%s,%d,%s,%s",
			escapeCSV(dep.Repository),
			dep.Issue.Number,
			escapeCSV(dep.Issue.Title),
			dep.Issue.State)

		if f.options.Detailed {
			fmt.Fprintf(f.options.Writer, ",%s,%s,%s",
				escapeCSV(formatAssigneesForCSV(dep.Issue.Assignees)),
				escapeCSV(formatLabelsForCSV(dep.Issue.Labels)),
				escapeCSV(dep.Issue.HTMLURL))
		}
		fmt.Fprintf(f.options.Writer, "\n")
	}

	// Blocking dependencies
	for _, dep := range data.Blocking {
		fmt.Fprintf(f.options.Writer, "blocking,%s,%d,%s,%s",
			escapeCSV(dep.Repository),
			dep.Issue.Number,
			escapeCSV(dep.Issue.Title),
			dep.Issue.State)

		if f.options.Detailed {
			fmt.Fprintf(f.options.Writer, ",%s,%s,%s",
				escapeCSV(formatAssigneesForCSV(dep.Issue.Assignees)),
				escapeCSV(formatLabelsForCSV(dep.Issue.Labels)),
				escapeCSV(dep.Issue.HTMLURL))
		}
		fmt.Fprintf(f.options.Writer, "\n")
	}

	return nil
}

// Helper functions for TTY output

// colorize returns a function that applies the given color
func (f *OutputFormatter) colorize(color termenv.Color) func(string) string {
	return func(s string) string {
		return termenv.String(s).Foreground(color).String()
	}
}

// getStateEmoji returns the appropriate emoji for the issue state
func (f *OutputFormatter) getStateEmoji(state string) string {
	switch strings.ToLower(state) {
	case "open":
		return "🔵"
	case "closed":
		return "✅"
	default:
		return "⚪"
	}
}

// getStateColor returns the appropriate color function for the issue state
func (f *OutputFormatter) getStateColor(state string) func(string) string {
	switch strings.ToLower(state) {
	case "open":
		return f.colorize(termenv.ANSIGreen)
	case "closed":
		return f.colorize(termenv.ANSIBrightBlack)
	default:
		return f.colorize(termenv.ANSIWhite)
	}
}

// Helper functions for JSON formatting

// formatIssueForJSON formats a single issue for JSON output
func (f *OutputFormatter) formatIssueForJSON(issue *Issue) map[string]interface{} {
	result := map[string]interface{}{
		"number":     issue.Number,
		"title":      issue.Title,
		"state":      issue.State,
		"repository": issue.Repository,
	}

	if f.options.Detailed {
		if len(issue.Assignees) > 0 {
			result["assignees"] = formatAssigneesForJSON(issue.Assignees)
		}
		if len(issue.Labels) > 0 {
			result["labels"] = formatLabelsForJSON(issue.Labels)
		}
		if issue.HTMLURL != "" {
			result["html_url"] = issue.HTMLURL
		}
	}

	return result
}

// formatDependenciesForJSON formats dependency relations for JSON output
func (f *OutputFormatter) formatDependenciesForJSON(deps []DependencyRelation) []map[string]interface{} {
	var result []map[string]interface{}

	for _, dep := range deps {
		item := map[string]interface{}{
			"number":     dep.Issue.Number,
			"title":      dep.Issue.Title,
			"state":      dep.Issue.State,
			"repository": dep.Repository,
		}

		if f.options.Detailed {
			if len(dep.Issue.Assignees) > 0 {
				item["assignees"] = formatAssigneesForJSON(dep.Issue.Assignees)
			}
			if len(dep.Issue.Labels) > 0 {
				item["labels"] = formatLabelsForJSON(dep.Issue.Labels)
			}
			if dep.Issue.HTMLURL != "" {
				item["html_url"] = dep.Issue.HTMLURL
			}
		}

		result = append(result, item)
	}

	return result
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

// formatAssigneesForCSV converts assignees to CSV-friendly string
func formatAssigneesForCSV(assignees []User) string {
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
func formatLabelsForCSV(labels []Label) string {
	if len(labels) == 0 {
		return ""
	}

	var names []string
	for _, label := range labels {
		names = append(names, label.Name)
	}
	return joinStrings(names, "; ")
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

// formatAssigneesForJSON converts assignees to JSON-friendly format
func formatAssigneesForJSON(assignees []User) []map[string]interface{} {
	var result []map[string]interface{}

	for _, user := range assignees {
		result = append(result, map[string]interface{}{
			"login":    user.Login,
			"html_url": user.HTMLURL,
		})
	}

	return result
}

// formatLabelsForJSON converts labels to JSON-friendly format
func formatLabelsForJSON(labels []Label) []map[string]interface{} {
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

// addFilterContext adds filter context information to help with empty state messaging
func (f *OutputFormatter) addFilterContext(data *DependencyData) {
	// Store original data counts if we have it for comparison
	if f.options.OriginalData != nil {
		data.OriginalBlockedByCount = len(f.options.OriginalData.BlockedBy)
		data.OriginalBlockingCount = len(f.options.OriginalData.Blocking)
	}
}

// getEmptyStateMessage returns context-aware empty state message
func (f *OutputFormatter) getEmptyStateMessage(data *DependencyData) (string, string) {
	var mainMsg, tipMsg string

	// Check if we have original data to compare against
	hasOriginalData := f.options.OriginalData != nil &&
		(f.options.OriginalData.TotalCount > data.TotalCount)

	switch f.options.StateFilter {
	case "open":
		if hasOriginalData {
			closedCount := f.options.OriginalData.TotalCount - data.TotalCount
			mainMsg = fmt.Sprintf("No open dependencies found for issue #%d.", data.SourceIssue.Number)
			if closedCount > 0 {
				tipMsg = fmt.Sprintf("Note: %d closed dependencies found. Use --state all to see all dependencies.", closedCount)
			} else {
				tipMsg = "No dependencies exist for this issue. Use 'gh issue-dependency add' to create relationships."
			}
		} else {
			mainMsg = fmt.Sprintf("No open dependencies found for issue #%d.", data.SourceIssue.Number)
			tipMsg = "Use --state all to see closed dependencies, or --state closed for closed only."
		}

	case "closed":
		if hasOriginalData {
			openCount := f.options.OriginalData.TotalCount - data.TotalCount
			mainMsg = fmt.Sprintf("No closed dependencies found for issue #%d.", data.SourceIssue.Number)
			if openCount > 0 {
				tipMsg = fmt.Sprintf("Note: %d open dependencies found. Use --state all to see all dependencies.", openCount)
			} else {
				tipMsg = "No dependencies exist for this issue. Use 'gh issue-dependency add' to create relationships."
			}
		} else {
			mainMsg = fmt.Sprintf("No closed dependencies found for issue #%d.", data.SourceIssue.Number)
			tipMsg = "Use --state all to see open dependencies, or --state open for open only."
		}

	default: // "all"
		mainMsg = fmt.Sprintf("No dependencies found for issue #%d.", data.SourceIssue.Number)
		tipMsg = "Use 'gh issue-dependency add' to create dependency relationships."
		if !data.SourceIssue.Repository.IsEmpty() {
			tipMsg += "\nNote: Some dependencies may exist in repositories you don't have access to."
		}
	}

	return mainMsg, tipMsg
}
