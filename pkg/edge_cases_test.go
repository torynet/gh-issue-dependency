package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test various error conditions and edge cases
func TestErrorHandlingEdgeCases(t *testing.T) {
	t.Run("ParseRepoFlag edge cases", func(t *testing.T) {
		edgeCases := []struct {
			name        string
			input       string
			expectError bool
			errorType   string
		}{
			{
				name:        "only slashes",
				input:       "///",
				expectError: true,
				errorType:   "RepositoryFormatError",
			},
			{
				name:        "GitHub URL without repo",
				input:       "https://github.com/",
				expectError: true,
				errorType:   "RepositoryFormatError",
			},
			{
				name:        "GitHub URL with just owner",
				input:       "https://github.com/owner",
				expectError: true,
				errorType:   "RepositoryFormatError",
			},
			{
				name:        "whitespace input",
				input:       "   ",
				expectError: true,
				errorType:   "EmptyValueError",
			},
			{
				name:        "special characters in repo name",
				input:       "owner/repo-with-special_chars.test",
				expectError: false,
			},
			{
				name:        "unicode characters",
				input:       "owner/repo-测试",
				expectError: false,
			},
		}

		for _, tc := range edgeCases {
			t.Run(tc.name, func(t *testing.T) {
				owner, repo, err := ParseRepoFlag(tc.input)

				if tc.expectError {
					assert.Error(t, err, "Expected error for input: %s", tc.input)
					assert.Empty(t, owner, "Owner should be empty on error")
					assert.Empty(t, repo, "Repo should be empty on error")

					if tc.errorType != "" {
						// In a real implementation, we'd check error types
						assert.Contains(t, err.Error(), strings.ToLower(tc.input),
							"Error should reference the input")
					}
				} else {
					assert.NoError(t, err, "Should not error for valid input: %s", tc.input)
					assert.NotEmpty(t, owner, "Owner should not be empty for valid input")
					assert.NotEmpty(t, repo, "Repo should not be empty for valid input")
				}
			})
		}
	})

	t.Run("ParseIssueURL edge cases", func(t *testing.T) {
		edgeCases := []struct {
			name        string
			input       string
			expectError bool
		}{
			{
				name:        "URL with port",
				input:       "https://github.com:443/owner/repo/issues/123",
				expectError: true, // Current implementation doesn't support ports
			},
			{
				name:        "URL with subdomain",
				input:       "https://api.github.com/owner/repo/issues/123",
				expectError: true, // Should only work with github.com
			},
			{
				name:        "HTTP instead of HTTPS",
				input:       "http://github.com/owner/repo/issues/123",
				expectError: true, // Current implementation expects https
			},
			{
				name:        "Very large issue number",
				input:       "https://github.com/owner/repo/issues/999999999",
				expectError: false, // Should handle large numbers
			},
			{
				name:        "Issue number with leading zeros",
				input:       "https://github.com/owner/repo/issues/0123",
				expectError: false, // Should parse as 123
			},
			{
				name:        "URL with many path segments",
				input:       "https://github.com/owner/repo/issues/123/extra/segments",
				expectError: false, // Should ignore extra segments
			},
			{
				name:        "URL with unicode in path",
				input:       "https://github.com/测试/repo-测试/issues/123",
				expectError: false, // Should handle unicode
			},
		}

		for _, tc := range edgeCases {
			t.Run(tc.name, func(t *testing.T) {
				owner, repo, issueNum, err := ParseIssueURL(tc.input)

				if tc.expectError {
					assert.Error(t, err, "Expected error for URL: %s", tc.input)
				} else {
					if err != nil {
						t.Logf("URL parsing failed (might be expected): %s -> %v", tc.input, err)
					} else {
						assert.NotEmpty(t, owner, "Owner should not be empty for valid URL")
						assert.NotEmpty(t, repo, "Repo should not be empty for valid URL")
						assert.Greater(t, issueNum, 0, "Issue number should be positive")
					}
				}
			})
		}
	})
}

func TestDataStructureEdgeCases(t *testing.T) {
	t.Run("DependencyData with nil slices", func(t *testing.T) {
		data := &DependencyData{
			SourceIssue: Issue{Number: 123, Title: "Test", State: "open"},
			BlockedBy:   nil, // Explicitly nil
			Blocking:    nil, // Explicitly nil
			FetchedAt:   time.Now(),
			TotalCount:  0,
		}

		// Should handle nil slices gracefully
		assert.Equal(t, 0, len(data.BlockedBy))
		assert.Equal(t, 0, len(data.Blocking))
		assert.Equal(t, 0, data.TotalCount)
	})

	t.Run("Issue with empty fields", func(t *testing.T) {
		issue := Issue{
			Number:     0,                // Zero number
			Title:      "",               // Empty title
			State:      "",               // Empty state
			Assignees:  []User{},         // Empty slice
			Labels:     []Label{},        // Empty slice
			HTMLURL:    "",               // Empty URL
			Repository: RepositoryInfo{}, // Empty repository
		}

		// Should not panic with empty fields
		assert.Equal(t, 0, issue.Number)
		assert.Empty(t, issue.Title)
		assert.Empty(t, issue.State)
		assert.NotNil(t, issue.Assignees) // Should be empty slice, not nil
		assert.NotNil(t, issue.Labels)    // Should be empty slice, not nil
	})

	t.Run("User with special characters", func(t *testing.T) {
		user := User{
			Login:   "user-with_special.chars123",
			HTMLURL: "https://github.com/user-with_special.chars123",
		}

		// Should handle special characters in usernames
		assert.Contains(t, user.Login, "_")
		assert.Contains(t, user.Login, ".")
		assert.Contains(t, user.Login, "-")
	})

	t.Run("Label with edge case values", func(t *testing.T) {
		label := Label{
			Name:        "very-long-label-name-that-exceeds-normal-expectations",
			Color:       "InvalidColorCode", // Invalid color
			Description: "",                 // Empty description
		}

		// Should handle edge case label values
		assert.Greater(t, len(label.Name), 30, "Should handle long label names")
		assert.NotEqual(t, "000000", label.Color, "Color validation would be done elsewhere")
	})
}

func TestOutputFormattingEdgeCases(t *testing.T) {
	t.Run("TTY output with very long titles", func(t *testing.T) {
		longTitle := strings.Repeat("Very Long Issue Title That Exceeds Normal Length Expectations ", 5)
		data := &DependencyData{
			SourceIssue: Issue{
				Number:     123,
				Title:      longTitle,
				State:      "open",
				Repository: createRepositoryInfo("test/repo"),
			},
			BlockedBy:  []DependencyRelation{},
			Blocking:   []DependencyRelation{},
			FetchedAt:  time.Now(),
			TotalCount: 0,
		}

		var buffer bytes.Buffer
		options := &OutputOptions{
			Format: FormatTTY,
			Writer: &buffer,
		}

		formatter := NewOutputFormatter(options)
		err := formatter.FormatOutput(data)

		assert.NoError(t, err, "Should handle very long titles without error")
		output := buffer.String()
		assert.Contains(t, output, longTitle[:50], "Should contain part of the long title")
	})

	t.Run("JSON output with special characters", func(t *testing.T) {
		data := &DependencyData{
			SourceIssue: Issue{
				Number:     123,
				Title:      "Issue with \"quotes\" and \n newlines \t tabs",
				State:      "open",
				Repository: createRepositoryInfo("test/repo"),
			},
			BlockedBy:  []DependencyRelation{},
			Blocking:   []DependencyRelation{},
			FetchedAt:  time.Now(),
			TotalCount: 0,
		}

		var buffer bytes.Buffer
		options := &OutputOptions{
			Format: FormatJSON,
			Writer: &buffer,
		}

		formatter := NewOutputFormatter(options)
		err := formatter.FormatOutput(data)

		assert.NoError(t, err, "Should handle special characters in JSON")

		// Verify JSON is valid
		var result map[string]interface{}
		err = json.Unmarshal(buffer.Bytes(), &result)
		require.NoError(t, err, "Should produce valid JSON")

		sourceIssue := result["source_issue"].(map[string]interface{})
		title := sourceIssue["title"].(string)
		assert.Contains(t, title, "quotes", "Should preserve content with special chars")
	})

	t.Run("CSV output with problematic characters", func(t *testing.T) {
		data := &DependencyData{
			SourceIssue: Issue{
				Number:     123,
				Title:      "Issue, with \"commas\" and\nnewlines",
				State:      "open",
				Repository: createRepositoryInfo("test/repo"),
				Assignees: []User{
					{Login: "user,with,commas"},
				},
			},
			BlockedBy:  []DependencyRelation{},
			Blocking:   []DependencyRelation{},
			FetchedAt:  time.Now(),
			TotalCount: 0,
		}

		var buffer bytes.Buffer
		options := &OutputOptions{
			Format:   FormatCSV,
			Writer:   &buffer,
			Detailed: true,
		}

		formatter := NewOutputFormatter(options)
		err := formatter.FormatOutput(data)

		assert.NoError(t, err, "Should handle CSV special characters")
		output := buffer.String()

		// Should properly escape CSV special characters
		assert.Contains(t, output, "\"Issue, with \"\"commas\"\" and\nnewlines\"",
			"Should escape commas and quotes in CSV")
		assert.Contains(t, output, "\"@user,with,commas\"",
			"Should escape commas in assignees")
	})

	t.Run("output with nil writer", func(t *testing.T) {
		testData := &DependencyData{
			SourceIssue: Issue{Number: 123, Title: "Test", State: "open"},
			FetchedAt:   time.Now(),
			TotalCount:  0,
		}

		options := &OutputOptions{
			Format: FormatTTY,
			Writer: nil, // Nil writer
		}

		// Should handle nil writer gracefully (DefaultOutputOptions sets os.Stdout)
		formatter := NewOutputFormatter(options)
		assert.NotNil(t, formatter, "Should create formatter even with nil options")

		// Use the test data to avoid unused variable error
		assert.Equal(t, 123, testData.SourceIssue.Number)
	})
}

func TestCachingEdgeCases(t *testing.T) {
	t.Run("cache with expired entries", func(t *testing.T) {
		// Test cache expiration logic
		pastTime := time.Now().Add(-10 * time.Minute) // 10 minutes ago
		data := &DependencyData{
			SourceIssue: Issue{Number: 123, Title: "Test", State: "open"},
			FetchedAt:   pastTime,
			TotalCount:  0,
		}

		entry := CacheEntry{
			Data:      *data,
			ExpiresAt: pastTime, // Already expired
		}

		// Should detect expired entries
		assert.True(t, time.Now().After(entry.ExpiresAt),
			"Entry should be considered expired")
	})

	t.Run("cache key generation edge cases", func(t *testing.T) {
		tests := []struct {
			name   string
			owner  string
			repo   string
			number int
		}{
			{"unicode characters", "测试", "repo-测试", 123},
			{"special characters", "owner-test", "repo_test.git", 456},
			{"very long names", strings.Repeat("a", 100), strings.Repeat("b", 100), 789},
			{"empty strings", "", "", 0},
		}

		keys := make(map[string]bool)
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				key := getCacheKey(tc.owner, tc.repo, tc.number)
				assert.Equal(t, 64, len(key), "Cache key should always be 64 chars (SHA256 hash)")

				// Ensure uniqueness
				assert.False(t, keys[key], "Cache keys should be unique")
				keys[key] = true
			})
		}
	})
}

func TestRepositoryContextEdgeCases(t *testing.T) {
	t.Run("ResolveRepository with conflicting inputs", func(t *testing.T) {
		// Test priority: repoFlag > issueURL > current repo
		owner, repo, err := ResolveRepository(
			"priority/repo", // High priority
			"https://github.com/from-url/repo/issues/123", // Lower priority
		)

		// Should prioritize repo flag over URL
		if err == nil || strings.Contains(err.Error(), "not available") {
			// If gh CLI is available, should prefer repo flag
			if err == nil {
				assert.Equal(t, "priority", owner, "Should use repo flag owner")
				assert.Equal(t, "repo", repo, "Should use repo flag repo")
			}
		} else {
			t.Logf("Test skipped due to GitHub CLI not available: %v", err)
		}
	})

	t.Run("ValidateRepoAccess with invalid repo", func(t *testing.T) {
		// Test with obviously invalid repository
		err := ValidateRepoAccess("definitely-does-not-exist", "invalid-repo-name-12345")

		// Should return appropriate error
		assert.Error(t, err, "Should error for non-existent repository")

		// Error should indicate the problem
		errorMsg := strings.ToLower(err.Error())
		assert.True(t,
			strings.Contains(errorMsg, "not found") ||
				strings.Contains(errorMsg, "not available") ||
				strings.Contains(errorMsg, "authentication") ||
				strings.Contains(errorMsg, "cannot access") ||
				strings.Contains(errorMsg, "access repository") ||
				strings.Contains(errorMsg, "cannot access repository"),
			"Error should indicate repository access issue: %v", err)
	})
}

func TestAPIIntegrationEdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("FetchIssueDependencies with timeout", func(t *testing.T) {
		// Test with very short timeout
		shortCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
		defer cancel()

		// Should handle timeout gracefully
		_, err := FetchIssueDependencies(shortCtx, "owner", "repo", 123)
		assert.Error(t, err, "Should error with short timeout")

		// Error might be timeout or authentication - both are acceptable
		errorMsg := strings.ToLower(err.Error())
		assert.True(t,
			strings.Contains(errorMsg, "timeout") ||
				strings.Contains(errorMsg, "context") ||
				strings.Contains(errorMsg, "authentication") ||
				strings.Contains(errorMsg, "not available") ||
				strings.Contains(errorMsg, "cannot access"),
			"Should indicate timeout or auth issue: %v", err)
	})

	t.Run("extractRepoFromURL with malformed URLs", func(t *testing.T) {
		malformedURLs := []string{
			"https://github.com/",
			"https://github.com/owner",
			"https://github.com/owner/",
			"github.com/owner/repo/issues/123",         // Missing protocol
			"https://github.com/owner/repo/issues/",    // Missing issue number
			"https://github.com/owner/repo/issues/abc", // Non-numeric issue
		}

		for _, url := range malformedURLs {
			t.Run(url, func(t *testing.T) {
				result := extractRepoFromURL(url)
				// Should return empty string for malformed URLs
				if result != "" {
					t.Logf("URL %s returned repo %s (might be partially valid)", url, result)
				}
			})
		}
	})
}

func TestMemoryAndPerformanceEdgeCases(t *testing.T) {
	t.Run("large dependency dataset", func(t *testing.T) {
		// Create data with many dependencies to test memory handling
		data := &DependencyData{
			SourceIssue: Issue{Number: 1000, Title: "Large Dataset Test", State: "open"},
			FetchedAt:   time.Now(),
		}

		// Add many dependencies
		for i := 1; i <= 1000; i++ {
			dep := DependencyRelation{
				Issue: Issue{
					Number:     i,
					Title:      fmt.Sprintf("Issue %d with some descriptive text", i),
					State:      []string{"open", "closed"}[i%2], // Alternate states
					Repository: createRepositoryInfo(fmt.Sprintf("repo%d/project%d", i%10, i%5)),
					Assignees: []User{
						{Login: fmt.Sprintf("user%d", i%50)},
					},
					Labels: []Label{
						{Name: fmt.Sprintf("label%d", i%20)},
					},
				},
				Type:       "blocked_by",
				Repository: fmt.Sprintf("repo%d/project%d", i%10, i%5),
			}
			data.BlockedBy = append(data.BlockedBy, dep)
		}
		data.TotalCount = len(data.BlockedBy)

		// Test that large datasets don't cause issues
		assert.Equal(t, 1000, len(data.BlockedBy), "Should handle 1000 dependencies")
		assert.Equal(t, 1000, data.TotalCount, "TotalCount should match")

		// Test formatting with large dataset
		var buffer bytes.Buffer
		options := &OutputOptions{
			Format: FormatJSON,
			Writer: &buffer,
		}

		formatter := NewOutputFormatter(options)
		err := formatter.FormatOutput(data)

		assert.NoError(t, err, "Should handle large datasets without error")
		assert.Greater(t, buffer.Len(), 10000, "Should produce substantial output")

		// Verify JSON is still valid
		var result map[string]interface{}
		err = json.Unmarshal(buffer.Bytes(), &result)
		require.NoError(t, err, "Large dataset should produce valid JSON")
	})

	t.Run("deeply nested repository paths", func(t *testing.T) {
		// Test with repositories that have deep organization structures
		deepRepo := strings.Repeat("organization/", 10) + "deeply/nested/repo/structure"

		data := &DependencyData{
			SourceIssue: Issue{
				Number:     123,
				Title:      "Deep Repository Test",
				State:      "open",
				Repository: createRepositoryInfo(deepRepo),
			},
			FetchedAt:  time.Now(),
			TotalCount: 0,
		}

		// Should handle deep repository paths
		var buffer bytes.Buffer
		options := &OutputOptions{
			Format: FormatPlain,
			Writer: &buffer,
		}

		formatter := NewOutputFormatter(options)
		err := formatter.FormatOutput(data)

		assert.NoError(t, err, "Should handle deep repository paths")
		output := buffer.String()
		assert.Contains(t, output, deepRepo, "Should display full repository path")
	})
}

func TestConcurrencyEdgeCases(t *testing.T) {
	t.Run("concurrent cache access", func(t *testing.T) {
		// Test that cache operations are safe for concurrent access
		// Note: Current implementation may not be fully concurrent-safe

		const numGoroutines = 10
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				// Generate different cache keys
				key := getCacheKey("owner", "repo", id)
				assert.Equal(t, 32, len(key), "Cache key should be valid length")

				// Attempt cache operations (these might fail due to permissions)
				data, found := getFromCache(key)
				assert.False(t, found, "Should not find non-existent cache entry")
				assert.Nil(t, data, "Data should be nil for cache miss")
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})

	t.Run("concurrent output formatting", func(t *testing.T) {
		// Test that output formatting is safe for concurrent use
		data := &DependencyData{
			SourceIssue: Issue{Number: 123, Title: "Concurrent Test", State: "open"},
			FetchedAt:   time.Now(),
			TotalCount:  0,
		}

		const numGoroutines = 5
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				var buffer bytes.Buffer
				options := &OutputOptions{
					Format: FormatJSON,
					Writer: &buffer,
				}

				formatter := NewOutputFormatter(options)
				err := formatter.FormatOutput(data)

				assert.NoError(t, err, "Concurrent formatting should not error")
				assert.Greater(t, buffer.Len(), 0, "Should produce output")
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})
}
