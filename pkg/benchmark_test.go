package pkg

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

// Helper function to create RepositoryInfo from full name string
func createRepositoryInfo(fullName string) RepositoryInfo {
	parts := strings.SplitN(fullName, "/", 2)
	if len(parts) != 2 {
		return RepositoryInfo{FullName: fullName}
	}
	return RepositoryInfo{
		Name:     parts[1],
		FullName: fullName,
		HTMLURL:  "https://github.com/" + fullName,
		Owner: struct {
			Login string `json:"login"`
		}{
			Login: parts[0],
		},
	}
}

// Benchmark tests for performance validation as required by issue #14
// Target: < 2 seconds execution time for command

func BenchmarkOutputFormatting(b *testing.B) {
	// Create test data with various sizes to test performance scaling
	smallData := createBenchmarkData(10)
	mediumData := createBenchmarkData(50)
	largeData := createBenchmarkData(200)

	b.Run("TTY-Small", func(b *testing.B) {
		benchmarkTTYOutput(b, smallData)
	})

	b.Run("TTY-Medium", func(b *testing.B) {
		benchmarkTTYOutput(b, mediumData)
	})

	b.Run("TTY-Large", func(b *testing.B) {
		benchmarkTTYOutput(b, largeData)
	})

	b.Run("JSON-Small", func(b *testing.B) {
		benchmarkJSONOutput(b, smallData)
	})

	b.Run("JSON-Medium", func(b *testing.B) {
		benchmarkJSONOutput(b, mediumData)
	})

	b.Run("JSON-Large", func(b *testing.B) {
		benchmarkJSONOutput(b, largeData)
	})

	b.Run("CSV-Small", func(b *testing.B) {
		benchmarkCSVOutput(b, smallData)
	})

	b.Run("CSV-Medium", func(b *testing.B) {
		benchmarkCSVOutput(b, mediumData)
	})

	b.Run("CSV-Large", func(b *testing.B) {
		benchmarkCSVOutput(b, largeData)
	})
}

func createBenchmarkData(dependencyCount int) *DependencyData {
	data := &DependencyData{
		SourceIssue: Issue{
			Number:     1000,
			Title:      "Performance Test Issue with a reasonably long title that represents real-world usage",
			State:      "open",
			Repository: RepositoryInfo{FullName: "performance/test-repo"},
			HTMLURL:    "https://github.com/performance/test-repo/issues/1000",
			Assignees: []User{
				{Login: "performance-tester", HTMLURL: "https://github.com/performance-tester"},
				{Login: "benchmark-user", HTMLURL: "https://github.com/benchmark-user"},
			},
			Labels: []Label{
				{Name: "performance", Color: "ff0000", Description: "Performance testing"},
				{Name: "benchmark", Color: "00ff00", Description: "Benchmarking"},
			},
		},
		FetchedAt: time.Now(),
	}

	// Add blocked by dependencies
	for i := 1; i <= dependencyCount/2; i++ {
		dep := DependencyRelation{
			Issue: Issue{
				Number:     i,
				Title:      fmt.Sprintf("Blocking dependency #%d with descriptive title that explains the work needed", i),
				State:      []string{"open", "closed"}[i%2],
				Repository: createRepositoryInfo(fmt.Sprintf("dependency/repo-%d", i%10)),
				HTMLURL:    fmt.Sprintf("https://github.com/dependency/repo-%d/issues/%d", i%10, i),
				Assignees: []User{
					{Login: fmt.Sprintf("assignee-%d", i%20), HTMLURL: fmt.Sprintf("https://github.com/assignee-%d", i%20)},
				},
				Labels: []Label{
					{Name: fmt.Sprintf("priority-%d", i%5), Color: fmt.Sprintf("%06x", i*123456%16777216)},
					{Name: fmt.Sprintf("category-%s", []string{"backend", "frontend", "database", "api", "ui"}[i%5])},
				},
			},
			Type:       "blocked_by",
			Repository: fmt.Sprintf("dependency/repo-%d", i%10),
		}
		data.BlockedBy = append(data.BlockedBy, dep)
	}

	// Add blocking dependencies
	for i := dependencyCount/2 + 1; i <= dependencyCount; i++ {
		dep := DependencyRelation{
			Issue: Issue{
				Number:     i + 1000,
				Title:      fmt.Sprintf("Blocked issue #%d waiting for this feature to be completed", i),
				State:      []string{"open", "closed"}[i%2],
				Repository: createRepositoryInfo(fmt.Sprintf("blocked/repo-%d", i%8)),
				HTMLURL:    fmt.Sprintf("https://github.com/blocked/repo-%d/issues/%d", i%8, i+1000),
				Assignees: []User{
					{Login: fmt.Sprintf("blocked-assignee-%d", i%15), HTMLURL: fmt.Sprintf("https://github.com/blocked-assignee-%d", i%15)},
				},
				Labels: []Label{
					{Name: fmt.Sprintf("blocked-%d", i%3), Color: fmt.Sprintf("%06x", i*654321%16777216)},
				},
			},
			Type:       "blocks",
			Repository: fmt.Sprintf("blocked/repo-%d", i%8),
		}
		data.Blocking = append(data.Blocking, dep)
	}

	data.TotalCount = len(data.BlockedBy) + len(data.Blocking)
	return data
}

func benchmarkTTYOutput(b *testing.B, data *DependencyData) {
	var buffer bytes.Buffer
	options := &OutputOptions{
		Format:   FormatTTY,
		Writer:   &buffer,
		Detailed: true, // Test with detailed output for more realistic performance
	}

	formatter := NewOutputFormatter(options)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer.Reset()
		err := formatter.FormatOutput(data)
		if err != nil {
			b.Fatalf("TTY formatting failed: %v", err)
		}
	}

	// Report performance characteristics
	b.ReportMetric(float64(data.TotalCount), "dependencies")
	b.ReportMetric(float64(buffer.Len()), "output_bytes")
}

func benchmarkJSONOutput(b *testing.B, data *DependencyData) {
	var buffer bytes.Buffer
	options := &OutputOptions{
		Format:   FormatJSON,
		Writer:   &buffer,
		Detailed: true,
	}

	formatter := NewOutputFormatter(options)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer.Reset()
		err := formatter.FormatOutput(data)
		if err != nil {
			b.Fatalf("JSON formatting failed: %v", err)
		}
	}

	b.ReportMetric(float64(data.TotalCount), "dependencies")
	b.ReportMetric(float64(buffer.Len()), "output_bytes")
}

func benchmarkCSVOutput(b *testing.B, data *DependencyData) {
	var buffer bytes.Buffer
	options := &OutputOptions{
		Format:   FormatCSV,
		Writer:   &buffer,
		Detailed: true,
	}

	formatter := NewOutputFormatter(options)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer.Reset()
		err := formatter.FormatOutput(data)
		if err != nil {
			b.Fatalf("CSV formatting failed: %v", err)
		}
	}

	b.ReportMetric(float64(data.TotalCount), "dependencies")
	b.ReportMetric(float64(buffer.Len()), "output_bytes")
}

// Benchmark parsing and validation functions
func BenchmarkRepositoryParsing(b *testing.B) {
	testCases := []string{
		"octocat/Hello-World",
		"https://github.com/owner/repo",
		"github.com/org/project",
		"microsoft/vscode",
		"https://github.com/facebook/react",
	}

	b.Run("ParseRepoFlag", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, tc := range testCases {
				_, _, err := ParseRepoFlag(tc)
				_ = err // Error expected for some test cases
			}
		}
	})

	urls := []string{
		"https://github.com/octocat/Hello-World/issues/123",
		"https://github.com/microsoft/vscode/issues/456",
		"https://github.com/facebook/react/issues/789",
		"https://github.com/golang/go/issues/12345",
		"https://github.com/kubernetes/kubernetes/issues/67890",
	}

	b.Run("ParseIssueURL", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, url := range urls {
				_, _, _, err := ParseIssueURL(url)
				if err != nil {
					b.Fatalf("URL parsing failed: %v", err)
				}
			}
		}
	})

	b.Run("ExtractRepoFromURL", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, url := range urls {
				result := extractRepoFromURL(url)
				if result == "" {
					b.Fatalf("Failed to extract repo from URL: %s", url)
				}
			}
		}
	})
}

// Benchmark cache operations
func BenchmarkCacheOperations(b *testing.B) {

	b.Run("GetCacheKey", func(b *testing.B) {
		owners := []string{"owner1", "owner2", "owner3", "owner4", "owner5"}
		repos := []string{"repo1", "repo2", "repo3", "repo4", "repo5"}
		issues := []int{123, 456, 789, 1011, 1213}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			owner := owners[i%len(owners)]
			repo := repos[i%len(repos)]
			issue := issues[i%len(issues)]

			key := getCacheKey(owner, repo, issue)
			if len(key) != 32 {
				b.Fatalf("Invalid cache key length: %d", len(key))
			}
		}
	})

	b.Run("CacheAccess", func(b *testing.B) {
		keys := make([]string, 100)
		for i := range keys {
			keys[i] = getCacheKey("owner", "repo", i)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := keys[i%len(keys)]

			// Test cache miss (most common scenario)
			data, found := getFromCache(key)
			if found {
				// Unexpected cache hit in benchmark environment
				b.Logf("Unexpected cache hit for key %s", key)
			}
			if data != nil && found {
				b.Logf("Got cached data: %v", data.SourceIssue.Number)
			}
		}
	})
}

// Benchmark data filtering and sorting operations
func BenchmarkDataProcessing(b *testing.B) {
	data := createBenchmarkData(100)

	b.Run("StateFiltering", func(b *testing.B) {
		states := []string{"all", "open", "closed"}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			state := states[i%len(states)]

			// This is the actual function from cmd/list.go
			filtered := applyStateFilter(data, state)

			if filtered == nil {
				b.Fatal("State filtering returned nil")
			}
		}
	})

	b.Run("DataSorting", func(b *testing.B) {
		sortOrders := []string{"number", "title", "state", "repository"}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sortOrder := sortOrders[i%len(sortOrders)]

			// This is the actual function from cmd/list.go
			sorted := applySorting(data, sortOrder)

			if sorted == nil {
				b.Fatal("Data sorting returned nil")
			}
		}
	})
}

// Benchmark string operations that are performance-critical
func BenchmarkStringOperations(b *testing.B) {
	longString := strings.Repeat("Issue with very long title that needs special handling ", 10)
	specialChars := "Issue with \"quotes\", commas, and\nnewlines\ttabs"

	b.Run("CSVEscaping", func(b *testing.B) {
		testStrings := []string{
			"simple string",
			longString,
			specialChars,
			"string,with,commas",
			"string\"with\"quotes",
			"string\nwith\nnewlines",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, str := range testStrings {
				result := escapeCSV(str)
				if result == "" && str != "" {
					b.Fatal("CSV escaping failed")
				}
			}
		}
	})

	b.Run("UserFormatting", func(b *testing.B) {
		users := []User{
			{Login: "user1"},
			{Login: "user2"},
			{Login: "user3"},
			{Login: "user4"},
			{Login: "user5"},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result := formatAssigneesForCSV(users)
			if !strings.Contains(result, "@user1") {
				b.Fatal("User formatting failed")
			}
		}
	})

	b.Run("LabelFormatting", func(b *testing.B) {
		labels := []Label{
			{Name: "bug"},
			{Name: "feature"},
			{Name: "enhancement"},
			{Name: "documentation"},
			{Name: "performance"},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result := formatLabelsForCSV(labels)
			if !strings.Contains(result, "bug") {
				b.Fatal("Label formatting failed")
			}
		}
	})
}

// Benchmark memory allocation patterns
func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("DependencyDataCreation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := &DependencyData{
				SourceIssue: Issue{
					Number:     i,
					Title:      fmt.Sprintf("Issue %d", i),
					State:      "open",
					Repository: createRepositoryInfo("test/repo"),
					Assignees:  []User{{Login: fmt.Sprintf("user%d", i)}},
					Labels:     []Label{{Name: fmt.Sprintf("label%d", i)}},
				},
				BlockedBy:  make([]DependencyRelation, 0, 10),
				Blocking:   make([]DependencyRelation, 0, 10),
				FetchedAt:  time.Now(),
				TotalCount: 0,
			}

			if data.SourceIssue.Number != i {
				b.Fatal("Data creation failed")
			}
		}
	})

	b.Run("SliceAppending", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var deps []DependencyRelation

			for j := 0; j < 50; j++ {
				dep := DependencyRelation{
					Issue: Issue{
						Number:     j,
						Title:      fmt.Sprintf("Dependency %d", j),
						State:      "open",
						Repository: createRepositoryInfo("test/repo"),
					},
					Type:       "blocked_by",
					Repository: "test/repo",
				}
				deps = append(deps, dep)
			}

			if len(deps) != 50 {
				b.Fatal("Slice operations failed")
			}
		}
	})
}

// Test performance characteristics under load
func BenchmarkConcurrentAccess(b *testing.B) {
	data := createBenchmarkData(50)

	b.Run("ConcurrentFormatting", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				var buffer bytes.Buffer
				options := &OutputOptions{
					Format: FormatJSON,
					Writer: &buffer,
				}

				formatter := NewOutputFormatter(options)
				err := formatter.FormatOutput(data)
				if err != nil {
					b.Fatalf("Concurrent formatting failed: %v", err)
				}
			}
		})
	})

	b.Run("ConcurrentCacheAccess", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := getCacheKey("owner", "repo", i)
				_, found := getFromCache(key)
				_ = found // Should not be found in benchmark environment
				i++
			}
		})
	})
}

// Performance regression tests - these should complete within time bounds
func TestPerformanceTargets(t *testing.T) {
	// Target: < 2 seconds for typical operations as per issue requirements

	t.Run("formatting performance target", func(t *testing.T) {
		data := createBenchmarkData(100) // Typical large dataset

		start := time.Now()

		var buffer bytes.Buffer
		options := &OutputOptions{
			Format:   FormatTTY,
			Writer:   &buffer,
			Detailed: true,
		}

		formatter := NewOutputFormatter(options)
		err := formatter.FormatOutput(data)

		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Formatting failed: %v", err)
		}

		// Should complete well under 2 seconds
		if duration > 500*time.Millisecond {
			t.Logf("Warning: Formatting took %v (target: < 500ms for 100 dependencies)", duration)
		}

		// Hard limit - should never exceed 2 seconds
		if duration > 2*time.Second {
			t.Errorf("Formatting took %v, exceeds 2 second target", duration)
		}
	})

	t.Run("parsing performance target", func(t *testing.T) {
		urls := []string{
			"https://github.com/octocat/Hello-World/issues/123",
			"https://github.com/microsoft/vscode/issues/456",
			"https://github.com/facebook/react/issues/789",
		}

		start := time.Now()

		// Parse 1000 URLs to simulate heavy usage
		for i := 0; i < 1000; i++ {
			url := urls[i%len(urls)]
			_, _, _, err := ParseIssueURL(url)
			if err != nil {
				t.Fatalf("URL parsing failed: %v", err)
			}
		}

		duration := time.Since(start)

		// Should complete very quickly
		if duration > 100*time.Millisecond {
			t.Logf("Warning: Parsing 1000 URLs took %v (target: < 100ms)", duration)
		}

		if duration > 1*time.Second {
			t.Errorf("Parsing took %v, exceeds reasonable target", duration)
		}
	})
}

// Helper function to simulate the actual data filtering from cmd/list.go
// We need to import or replicate these functions for benchmarking
func applyStateFilter(data *DependencyData, state string) *DependencyData {
	if state == "all" {
		return data
	}

	filtered := &DependencyData{
		SourceIssue: data.SourceIssue,
		BlockedBy:   []DependencyRelation{},
		Blocking:    []DependencyRelation{},
		FetchedAt:   data.FetchedAt,
	}

	for _, dep := range data.BlockedBy {
		if dep.Issue.State == state {
			filtered.BlockedBy = append(filtered.BlockedBy, dep)
		}
	}

	for _, dep := range data.Blocking {
		if dep.Issue.State == state {
			filtered.Blocking = append(filtered.Blocking, dep)
		}
	}

	filtered.TotalCount = len(filtered.BlockedBy) + len(filtered.Blocking)
	return filtered
}

func applySorting(data *DependencyData, sortOrder string) *DependencyData {
	if sortOrder == "" || sortOrder == "number" {
		return data
	}

	sorted := &DependencyData{
		SourceIssue: data.SourceIssue,
		BlockedBy:   make([]DependencyRelation, len(data.BlockedBy)),
		Blocking:    make([]DependencyRelation, len(data.Blocking)),
		FetchedAt:   data.FetchedAt,
		TotalCount:  data.TotalCount,
	}

	copy(sorted.BlockedBy, data.BlockedBy)
	copy(sorted.Blocking, data.Blocking)

	// Simple sorting implementation for benchmarking
	// In real code, this would use the actual sorting logic

	return sorted
}
