package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data helpers
func createTestDependencyData() *DependencyData {
	return &DependencyData{
		SourceIssue: Issue{
			Number:     123,
			Title:      "Main Feature Implementation",
			State:      "open",
			Repository: "testowner/testrepo",
			HTMLURL:    "https://github.com/testowner/testrepo/issues/123",
			Assignees: []User{
				{Login: "alice", HTMLURL: "https://github.com/alice"},
			},
			Labels: []Label{
				{Name: "feature", Color: "0e8a16", Description: "New feature"},
			},
		},
		BlockedBy: []DependencyRelation{
			{
				Issue: Issue{
					Number:     45,
					Title:      "Setup Database Schema",
					State:      "open",
					Repository: "testowner/testrepo",
					HTMLURL:    "https://github.com/testowner/testrepo/issues/45",
					Assignees: []User{
						{Login: "bob", HTMLURL: "https://github.com/bob"},
					},
				},
				Type:       "blocked_by",
				Repository: "testowner/testrepo",
			},
			{
				Issue: Issue{
					Number:     67,
					Title:      "API Endpoint Creation",
					State:      "closed",
					Repository: "testowner/testrepo",
					HTMLURL:    "https://github.com/testowner/testrepo/issues/67",
				},
				Type:       "blocked_by",
				Repository: "testowner/testrepo",
			},
		},
		Blocking: []DependencyRelation{
			{
				Issue: Issue{
					Number:     89,
					Title:      "Frontend Integration",
					State:      "open",
					Repository: "testowner/frontend",
					HTMLURL:    "https://github.com/testowner/frontend/issues/89",
					Assignees: []User{
						{Login: "charlie", HTMLURL: "https://github.com/charlie"},
						{Login: "diana", HTMLURL: "https://github.com/diana"},
					},
					Labels: []Label{
						{Name: "frontend", Color: "1d76db", Description: "Frontend work"},
						{Name: "urgent", Color: "d93f0b"},
					},
				},
				Type:       "blocks",
				Repository: "testowner/frontend",
			},
		},
		FetchedAt:  time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		TotalCount: 3,
	}
}

func createEmptyDependencyData() *DependencyData {
	return &DependencyData{
		SourceIssue: Issue{
			Number:     456,
			Title:      "Standalone Task",
			State:      "open",
			Repository: "testowner/testrepo",
			HTMLURL:    "https://github.com/testowner/testrepo/issues/456",
		},
		BlockedBy:  []DependencyRelation{},
		Blocking:   []DependencyRelation{},
		FetchedAt:  time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		TotalCount: 0,
	}
}

// Test DefaultOutputOptions
func TestDefaultOutputOptions(t *testing.T) {
	options := DefaultOutputOptions()
	
	assert.Equal(t, FormatAuto, options.Format)
	assert.Empty(t, options.JSONFields)
	assert.False(t, options.Detailed)
	assert.NotNil(t, options.Writer)
	assert.Equal(t, "all", options.StateFilter)
	assert.Nil(t, options.OriginalData)
}

// Test NewOutputFormatter
func TestNewOutputFormatter(t *testing.T) {
	t.Run("with nil options uses defaults", func(t *testing.T) {
		formatter := NewOutputFormatter(nil)
		assert.NotNil(t, formatter)
		assert.NotNil(t, formatter.options)
		assert.Equal(t, FormatAuto, formatter.options.Format)
	})
	
	t.Run("with custom options", func(t *testing.T) {
		options := &OutputOptions{
			Format:     FormatJSON,
			JSONFields: []string{"blocked_by"},
			Detailed:   true,
		}
		formatter := NewOutputFormatter(options)
		assert.NotNil(t, formatter)
		assert.Equal(t, FormatJSON, formatter.options.Format)
		assert.True(t, formatter.options.Detailed)
		assert.Equal(t, []string{"blocked_by"}, formatter.options.JSONFields)
	})
}

// Test TTY Output Formatting
func TestFormatTTYOutput(t *testing.T) {
	tests := []struct {
		name     string
		data     *DependencyData
		detailed bool
		contains []string
	}{
		{
			name:     "single blocked by dependency",
			data:     createTestDependencyData(),
			detailed: false,
			contains: []string{
				"Dependencies for: #123 - Main Feature Implementation",
				"BLOCKED BY (2 issues)",
				"🔵 #45",
				"Setup Database Schema",
				"[open]",
				"✅ #67",
				"API Endpoint Creation",
				"[closed]",
				"BLOCKS (1 issues)",
				"🔵 #89",
				"Frontend Integration",
				"testowner/frontend",
			},
		},
		{
			name:     "detailed output with metadata",
			data:     createTestDependencyData(),
			detailed: true,
			contains: []string{
				"Dependencies for: #123 - Main Feature Implementation",
				"@bob",
				"@charlie, @diana",
				"Fetched at: 2024-01-01T12:00:00Z",
				"https://github.com/testowner/testrepo/issues/45",
			},
		},
		{
			name:     "empty dependencies",
			data:     createEmptyDependencyData(),
			detailed: false,
			contains: []string{
				"Dependencies for: #456 - Standalone Task",
				"💡 No dependencies found",
				"gh issue-dependency add",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer
			options := &OutputOptions{
				Format:   FormatTTY,
				Writer:   &buffer,
				Detailed: tt.detailed,
			}
			
			formatter := NewOutputFormatter(options)
			err := formatter.FormatOutput(tt.data)
			
			assert.NoError(t, err)
			output := buffer.String()
			
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}
		})
	}
}

// Test Plain Output Formatting
func TestFormatPlainOutput(t *testing.T) {
	tests := []struct {
		name     string
		data     *DependencyData
		detailed bool
		contains []string
		notContains []string
	}{
		{
			name:     "basic plain output",
			data:     createTestDependencyData(),
			detailed: false,
			contains: []string{
				"Dependencies for: #123 - Main Feature Implementation",
				"Repository: testowner/testrepo",
				"BLOCKED BY (2 issues)",
				"#45 Setup Database Schema [open]",
				"#67 API Endpoint Creation [closed]",
				"BLOCKS (1 issues)",
				"#89 Frontend Integration [open]",
			},
			notContains: []string{
				"🔵", "✅", "💡", // No emojis in plain output
			},
		},
		{
			name:     "plain output with assignees",
			data:     createTestDependencyData(),
			detailed: false,
			contains: []string{
				"@bob",
				"@charlie, @diana",
				"Repository: testowner/frontend",
			},
		},
		{
			name:     "detailed plain output",
			data:     createTestDependencyData(),
			detailed: true,
			contains: []string{
				"URL: https://github.com/testowner/testrepo/issues/45",
				"Fetched at: 2024-01-01T12:00:00Z",
			},
		},
		{
			name:     "empty plain output",
			data:     createEmptyDependencyData(),
			detailed: false,
			contains: []string{
				"No dependencies found for issue #456",
				"gh issue-dependency add",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer
			options := &OutputOptions{
				Format:   FormatPlain,
				Writer:   &buffer,
				Detailed: tt.detailed,
			}
			
			formatter := NewOutputFormatter(options)
			err := formatter.FormatOutput(tt.data)
			
			assert.NoError(t, err)
			output := buffer.String()
			
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}
			for _, unexpected := range tt.notContains {
				assert.NotContains(t, output, unexpected, "Output should not contain: %s", unexpected)
			}
		})
	}
}

// Test JSON Output Formatting
func TestFormatJSONOutput(t *testing.T) {
	tests := []struct {
		name       string
		data       *DependencyData
		detailed   bool
		jsonFields []string
		validate   func(t *testing.T, output map[string]interface{})
	}{
		{
			name:     "basic JSON output",
			data:     createTestDependencyData(),
			detailed: false,
			validate: func(t *testing.T, output map[string]interface{}) {
				assert.Contains(t, output, "source_issue")
				assert.Contains(t, output, "blocked_by")
				assert.Contains(t, output, "blocks")
				assert.Contains(t, output, "summary")
				
				sourceIssue := output["source_issue"].(map[string]interface{})
				assert.Equal(t, float64(123), sourceIssue["number"])
				assert.Equal(t, "Main Feature Implementation", sourceIssue["title"])
				assert.Equal(t, "open", sourceIssue["state"])
				assert.Equal(t, "testowner/testrepo", sourceIssue["repository"])
				
				blockedBy := output["blocked_by"].([]interface{})
				assert.Len(t, blockedBy, 2)
				
				blocks := output["blocks"].([]interface{})
				assert.Len(t, blocks, 1)
				
				summary := output["summary"].(map[string]interface{})
				assert.Equal(t, float64(3), summary["total_count"])
				assert.Equal(t, float64(2), summary["blocked_by_count"])
				assert.Equal(t, float64(1), summary["blocks_count"])
			},
		},
		{
			name:     "detailed JSON output",
			data:     createTestDependencyData(),
			detailed: true,
			validate: func(t *testing.T, output map[string]interface{}) {
				sourceIssue := output["source_issue"].(map[string]interface{})
				assert.Contains(t, sourceIssue, "assignees")
				assert.Contains(t, sourceIssue, "labels")
				assert.Contains(t, sourceIssue, "html_url")
				
				assignees := sourceIssue["assignees"].([]interface{})
				assert.Len(t, assignees, 1)
				assignee := assignees[0].(map[string]interface{})
				assert.Equal(t, "alice", assignee["login"])
			},
		},
		{
			name:       "JSON field selection",
			data:       createTestDependencyData(),
			detailed:   false,
			jsonFields: []string{"blocked_by", "summary"},
			validate: func(t *testing.T, output map[string]interface{}) {
				assert.Contains(t, output, "blocked_by")
				assert.Contains(t, output, "summary")
				assert.NotContains(t, output, "source_issue")
				assert.NotContains(t, output, "blocks")
			},
		},
		{
			name:     "empty JSON output",
			data:     createEmptyDependencyData(),
			detailed: false,
			validate: func(t *testing.T, output map[string]interface{}) {
				// Handle the case where blocked_by and blocks might be nil or empty arrays
				if blockedBy, exists := output["blocked_by"]; exists && blockedBy != nil {
					blockedBySlice := blockedBy.([]interface{})
					assert.Len(t, blockedBySlice, 0)
				}
				
				if blocks, exists := output["blocks"]; exists && blocks != nil {
					blocksSlice := blocks.([]interface{})
					assert.Len(t, blocksSlice, 0)
				}
				
				summary := output["summary"].(map[string]interface{})
				assert.Equal(t, float64(0), summary["total_count"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer
			options := &OutputOptions{
				Format:     FormatJSON,
				Writer:     &buffer,
				Detailed:   tt.detailed,
				JSONFields: tt.jsonFields,
			}
			
			formatter := NewOutputFormatter(options)
			err := formatter.FormatOutput(tt.data)
			
			assert.NoError(t, err)
			
			var output map[string]interface{}
			err = json.Unmarshal(buffer.Bytes(), &output)
			require.NoError(t, err, "JSON output should be valid")
			
			tt.validate(t, output)
		})
	}
}

// Test CSV Output Formatting
func TestFormatCSVOutput(t *testing.T) {
	tests := []struct {
		name     string
		data     *DependencyData
		detailed bool
		validate func(t *testing.T, lines []string)
	}{
		{
			name:     "basic CSV output",
			data:     createTestDependencyData(),
			detailed: false,
			validate: func(t *testing.T, lines []string) {
				assert.Len(t, lines, 5) // header + source + 2 blocked_by + 1 blocks + empty line
				
				// Header
				assert.Equal(t, "type,repository,number,title,state", lines[0])
				
				// Source issue
				assert.Contains(t, lines[1], "source,testowner/testrepo,123,Main Feature Implementation,open")
				
				// Dependencies
				assert.Contains(t, lines[2], "blocked_by,testowner/testrepo,45,Setup Database Schema,open")
				assert.Contains(t, lines[3], "blocked_by,testowner/testrepo,67,API Endpoint Creation,closed")
				assert.Contains(t, lines[4], "blocking,testowner/frontend,89,Frontend Integration,open")
			},
		},
		{
			name:     "detailed CSV output",
			data:     createTestDependencyData(),
			detailed: true,
			validate: func(t *testing.T, lines []string) {
				// Header should include additional fields
				assert.Equal(t, "type,repository,number,title,state,assignees,labels,html_url", lines[0])
				
				// Check assignees and labels are included
				assert.Contains(t, lines[1], "@alice")
				assert.Contains(t, lines[2], "@bob")
				assert.Contains(t, lines[4], "@charlie; @diana")
				assert.Contains(t, lines[4], "frontend; urgent")
			},
		},
		{
			name:     "CSV with special characters",
			data: &DependencyData{
				SourceIssue: Issue{
					Number:     1,
					Title:      "Issue with \"quotes\" and, commas",
					State:      "open",
					Repository: "test/repo",
				},
				BlockedBy:  []DependencyRelation{},
				Blocking:   []DependencyRelation{},
				FetchedAt:  time.Now(),
				TotalCount: 0,
			},
			detailed: false,
			validate: func(t *testing.T, lines []string) {
				// Check CSV escaping
				assert.Contains(t, lines[1], "\"Issue with \"\"quotes\"\" and, commas\"")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer
			options := &OutputOptions{
				Format:   FormatCSV,
				Writer:   &buffer,
				Detailed: tt.detailed,
			}
			
			formatter := NewOutputFormatter(options)
			err := formatter.FormatOutput(tt.data)
			
			assert.NoError(t, err)
			
			lines := strings.Split(strings.TrimSpace(buffer.String()), "\n")
			tt.validate(t, lines)
		})
	}
}

// Test Auto Format Detection
func TestDetermineFormat(t *testing.T) {
	tests := []struct {
		name           string
		format         OutputFormat
		expectedFormat OutputFormat
	}{
		{
			name:           "explicit TTY format",
			format:         FormatTTY,
			expectedFormat: FormatTTY,
		},
		{
			name:           "explicit plain format",
			format:         FormatPlain,
			expectedFormat: FormatPlain,
		},
		{
			name:           "explicit JSON format",
			format:         FormatJSON,
			expectedFormat: FormatJSON,
		},
		{
			name:           "explicit CSV format",
			format:         FormatCSV,
			expectedFormat: FormatCSV,
		},
		{
			name:           "auto format defaults to plain (non-TTY)",
			format:         FormatAuto,
			expectedFormat: FormatPlain, // Assuming non-TTY environment in tests
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer
			options := &OutputOptions{
				Format: tt.format,
				Writer: &buffer,
			}
			
			formatter := NewOutputFormatter(options)
			actualFormat := formatter.determineFormat()
			
			assert.Equal(t, tt.expectedFormat, actualFormat)
		})
	}
}

// Test Empty State Messages
func TestGetEmptyStateMessage(t *testing.T) {
	tests := []struct {
		name        string
		stateFilter string
		data        *DependencyData
		originalData *DependencyData
		expectMain  string
		expectTip   string
	}{
		{
			name:        "all state - no dependencies",
			stateFilter: "all",
			data:        createEmptyDependencyData(),
			expectMain:  "No dependencies found for issue #456",
			expectTip:   "Use 'gh issue-dependency add' to create dependency relationships",
		},
		{
			name:        "open state - no open dependencies",
			stateFilter: "open",
			data:        createEmptyDependencyData(),
			expectMain:  "No open dependencies found for issue #456",
			expectTip:   "Use --state all to see closed dependencies",
		},
		{
			name:        "closed state - no closed dependencies",
			stateFilter: "closed",
			data:        createEmptyDependencyData(),
			expectMain:  "No closed dependencies found for issue #456",
			expectTip:   "Use --state all to see open dependencies",
		},
		{
			name:         "open state - with closed dependencies available",
			stateFilter:  "open",
			data:         createEmptyDependencyData(),
			originalData: createTestDependencyData(),
			expectMain:   "No open dependencies found for issue #456",
			expectTip:    "3 closed dependencies found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer
			options := &OutputOptions{
				Format:       FormatTTY,
				Writer:       &buffer,
				StateFilter:  tt.stateFilter,
				OriginalData: tt.originalData,
			}
			
			formatter := NewOutputFormatter(options)
			mainMsg, tipMsg := formatter.getEmptyStateMessage(tt.data)
			
			assert.Contains(t, mainMsg, tt.expectMain)
			assert.Contains(t, tipMsg, tt.expectTip)
		})
	}
}

// Test Helper Functions
func TestHelperFunctions(t *testing.T) {
	t.Run("escapeCSV", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"simple", "simple"},
			{"with,comma", "\"with,comma\""},
			{"with\"quote", "\"with\"\"quote\""},
			{"with\nnewline", "\"with\nnewline\""},
			{"with,comma\"and quote", "\"with,comma\"\"and quote\""},
		}
		
		for _, tt := range tests {
			result := escapeCSV(tt.input)
			assert.Equal(t, tt.expected, result, "escapeCSV(%q)", tt.input)
		}
	})
	
	t.Run("formatAssigneesForCSV", func(t *testing.T) {
		users := []User{
			{Login: "alice"},
			{Login: "bob"},
		}
		result := formatAssigneesForCSV(users)
		assert.Equal(t, "@alice; @bob", result)
		
		empty := formatAssigneesForCSV([]User{})
		assert.Equal(t, "", empty)
	})
	
	t.Run("formatLabelsForCSV", func(t *testing.T) {
		labels := []Label{
			{Name: "bug"},
			{Name: "urgent"},
		}
		result := formatLabelsForCSV(labels)
		assert.Equal(t, "bug; urgent", result)
		
		empty := formatLabelsForCSV([]Label{})
		assert.Equal(t, "", empty)
	})
}

// Benchmark tests for performance validation
func BenchmarkFormatTTYOutput(b *testing.B) {
	data := createTestDependencyData()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buffer bytes.Buffer
		options := &OutputOptions{
			Format: FormatTTY,
			Writer: &buffer,
		}
		
		formatter := NewOutputFormatter(options)
		_ = formatter.FormatOutput(data)
	}
}

func BenchmarkFormatJSONOutput(b *testing.B) {
	data := createTestDependencyData()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buffer bytes.Buffer
		options := &OutputOptions{
			Format: FormatJSON,
			Writer: &buffer,
		}
		
		formatter := NewOutputFormatter(options)
		_ = formatter.FormatOutput(data)
	}
}

func BenchmarkFormatCSVOutput(b *testing.B) {
	data := createTestDependencyData()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buffer bytes.Buffer
		options := &OutputOptions{
			Format: FormatCSV,
			Writer: &buffer,
		}
		
		formatter := NewOutputFormatter(options)
		_ = formatter.FormatOutput(data)
	}
}

// Test edge cases for large datasets
func TestLargeDatasetOutput(t *testing.T) {
	// Create test data with many dependencies
	data := &DependencyData{
		SourceIssue: Issue{
			Number:     1000,
			Title:      "Large Feature",
			State:      "open",
			Repository: "test/repo",
		},
		FetchedAt: time.Now(),
	}
	
	// Add many dependencies
	for i := 1; i <= 100; i++ {
		dep := DependencyRelation{
			Issue: Issue{
				Number:     i,
				Title:      fmt.Sprintf("Dependency %d", i),
				State:      "open",
				Repository: "test/repo",
			},
			Type:       "blocked_by",
			Repository: "test/repo",
		}
		data.BlockedBy = append(data.BlockedBy, dep)
	}
	data.TotalCount = len(data.BlockedBy)
	
	t.Run("large dataset TTY output", func(t *testing.T) {
		var buffer bytes.Buffer
		options := &OutputOptions{
			Format: FormatTTY,
			Writer: &buffer,
		}
		
		formatter := NewOutputFormatter(options)
		err := formatter.FormatOutput(data)
		
		assert.NoError(t, err)
		assert.Contains(t, buffer.String(), "BLOCKED BY (100 issues)")
	})
	
	t.Run("large dataset JSON output", func(t *testing.T) {
		var buffer bytes.Buffer
		options := &OutputOptions{
			Format: FormatJSON,
			Writer: &buffer,
		}
		
		formatter := NewOutputFormatter(options)
		err := formatter.FormatOutput(data)
		
		assert.NoError(t, err)
		
		var output map[string]interface{}
		err = json.Unmarshal(buffer.Bytes(), &output)
		require.NoError(t, err)
		
		blockedBy := output["blocked_by"].([]interface{})
		assert.Len(t, blockedBy, 100)
	})
}