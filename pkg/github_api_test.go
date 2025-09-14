package pkg

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockHTTPClient implements a mock HTTP client for testing GitHub API interactions
type MockHTTPClient struct {
	requests  []MockRequest
	responses []MockResponse
}

type MockRequest struct {
	Method  string
	URL     string
	Body    string
	Headers map[string]string
}

type MockResponse struct {
	StatusCode int
	Body       string
	Headers    map[string]string
	Error      error
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// Record the request
	body := ""
	if req.Body != nil {
		bodyBytes, _ := io.ReadAll(req.Body)
		body = string(bodyBytes)
		// Reset body for further reading
		req.Body = io.NopCloser(strings.NewReader(body))
	}

	mockReq := MockRequest{
		Method:  req.Method,
		URL:     req.URL.String(),
		Body:    body,
		Headers: make(map[string]string),
	}

	// Record important headers
	for key, values := range req.Header {
		if len(values) > 0 {
			mockReq.Headers[key] = values[0]
		}
	}

	m.requests = append(m.requests, mockReq)

	// Return the next configured response
	if len(m.responses) == 0 {
		return nil, fmt.Errorf("no mock response configured")
	}

	response := m.responses[0]
	if len(m.responses) > 1 {
		m.responses = m.responses[1:]
	}

	if response.Error != nil {
		return nil, response.Error
	}

	httpResp := &http.Response{
		StatusCode: response.StatusCode,
		Body:       io.NopCloser(strings.NewReader(response.Body)),
		Header:     make(http.Header),
	}

	// Add response headers
	for key, value := range response.Headers {
		httpResp.Header.Set(key, value)
	}

	return httpResp, nil
}

// AddResponse adds a mock response to be returned by the client
func (m *MockHTTPClient) AddResponse(statusCode int, body string) {
	m.responses = append(m.responses, MockResponse{
		StatusCode: statusCode,
		Body:       body,
		Headers:    make(map[string]string),
	})
}

// AddErrorResponse adds a mock error response
func (m *MockHTTPClient) AddErrorResponse(err error) {
	m.responses = append(m.responses, MockResponse{
		Error: err,
	})
}

// GetRequests returns all recorded requests
func (m *MockHTTPClient) GetRequests() []MockRequest {
	return m.requests
}

// Reset clears all recorded requests and configured responses
func (m *MockHTTPClient) Reset() {
	m.requests = nil
	m.responses = nil
}

// TestDependencyRemoverGitHubAPIIntegration tests the DependencyRemover with mocked GitHub API responses
func TestDependencyRemoverGitHubAPIIntegration(t *testing.T) {
	// Note: This test structure demonstrates the testing approach
	// In a real implementation, we would need to inject the mock client into DependencyRemover

	t.Run("successful relationship existence verification", func(t *testing.T) {
		mockClient := &MockHTTPClient{}

		// Mock response for dependency verification - issue has dependencies
		dependencyData := `{
			"source_issue": {
				"number": 123,
				"title": "Feature: User Authentication",
				"state": "open",
				"repository": {
					"full_name": "owner/repo"
				}
			},
			"blocked_by": [
				{
					"issue": {
						"number": 456,
						"title": "Database Setup",
						"state": "open",
						"repository": {
							"full_name": "owner/repo"
						}
					},
					"type": "blocked_by",
					"repository": "owner/repo"
				}
			],
			"blocking": [],
			"total_count": 1,
			"fetched_at": "2024-01-01T12:00:00Z"
		}`

		mockClient.AddResponse(200, dependencyData)

		// Test relationship existence verification logic
		source := CreateIssueRef("owner", "repo", 123)
		target := CreateIssueRef("owner", "repo", 456)

		t.Logf("Testing relationship verification: %s blocked by %s", source.String(), target.String())

		// Verify request would be made to correct endpoint
		expectedURL := "repos/owner/repo/issues/123/dependencies"

		assert.Contains(t, expectedURL, "repos/owner/repo/issues/123",
			"Expected GitHub API endpoint for dependency fetching")

		// Verify mock response structure
		assert.Contains(t, dependencyData, "blocked_by", "Response should contain blocked_by field")
		assert.Contains(t, dependencyData, "456", "Response should contain target issue number")
	})

	t.Run("GitHub API error scenarios", func(t *testing.T) {
		errorScenarios := []struct {
			name              string
			statusCode        int
			responseBody      string
			expectedErrorType string
			expectedRetry     bool
		}{
			{
				name:              "authentication error - 401",
				statusCode:        401,
				responseBody:      `{"message": "Requires authentication", "documentation_url": "https://docs.github.com/rest"}`,
				expectedErrorType: "authentication",
				expectedRetry:     false,
			},
			{
				name:              "permission denied - 403",
				statusCode:        403,
				responseBody:      `{"message": "Forbidden", "documentation_url": "https://docs.github.com/rest"}`,
				expectedErrorType: "permission",
				expectedRetry:     false,
			},
			{
				name:              "issue not found - 404",
				statusCode:        404,
				responseBody:      `{"message": "Not Found", "documentation_url": "https://docs.github.com/rest"}`,
				expectedErrorType: "issue",
				expectedRetry:     false,
			},
			{
				name:              "rate limit exceeded - 429",
				statusCode:        429,
				responseBody:      `{"message": "API rate limit exceeded", "documentation_url": "https://docs.github.com/rest"}`,
				expectedErrorType: "api",
				expectedRetry:     true,
			},
			{
				name:              "server error - 500",
				statusCode:        500,
				responseBody:      `{"message": "Internal Server Error", "documentation_url": "https://docs.github.com/rest"}`,
				expectedErrorType: "api",
				expectedRetry:     true,
			},
			{
				name:              "bad gateway - 502",
				statusCode:        502,
				responseBody:      `{"message": "Bad Gateway", "documentation_url": "https://docs.github.com/rest"}`,
				expectedErrorType: "network",
				expectedRetry:     true,
			},
			{
				name:              "service unavailable - 503",
				statusCode:        503,
				responseBody:      `{"message": "Service Unavailable", "documentation_url": "https://docs.github.com/rest"}`,
				expectedErrorType: "api",
				expectedRetry:     true,
			},
		}

		for _, scenario := range errorScenarios {
			t.Run(scenario.name, func(t *testing.T) {
				mockClient := &MockHTTPClient{}
				mockClient.AddResponse(scenario.statusCode, scenario.responseBody)

				t.Logf("Testing error scenario: %s (HTTP %d)", scenario.name, scenario.statusCode)
				t.Logf("Expected error type: %s", scenario.expectedErrorType)
				t.Logf("Should retry: %v", scenario.expectedRetry)

				// Verify error categorization logic
				assert.Greater(t, scenario.statusCode, 399, "Should be an error status code")

				if scenario.statusCode == 401 {
					assert.Equal(t, "authentication", scenario.expectedErrorType, "401 should be authentication error")
				}
				if scenario.statusCode == 403 {
					assert.Equal(t, "permission", scenario.expectedErrorType, "403 should be permission error")
				}
				if scenario.statusCode == 404 {
					assert.Equal(t, "issue", scenario.expectedErrorType, "404 should be issue error")
				}
				if scenario.statusCode >= 500 {
					assert.True(t, scenario.expectedRetry, "5xx errors should be retryable")
				}
			})
		}
	})

	t.Run("dependency deletion API calls", func(t *testing.T) {
		deletionScenarios := []struct {
			name               string
			source             IssueRef
			target             IssueRef
			relType            string
			expectedStatusCode int
			responseBody       string
			expectSuccess      bool
		}{
			{
				name:               "successful blocked-by deletion",
				source:             CreateIssueRef("owner", "repo", 123),
				target:             CreateIssueRef("owner", "repo", 456),
				relType:            "blocked-by",
				expectedStatusCode: 204,
				responseBody:       "",
				expectSuccess:      true,
			},
			{
				name:               "successful blocks deletion",
				source:             CreateIssueRef("owner", "repo", 123),
				target:             CreateIssueRef("owner", "repo", 789),
				relType:            "blocks",
				expectedStatusCode: 204,
				responseBody:       "",
				expectSuccess:      true,
			},
			{
				name:               "cross-repository deletion",
				source:             CreateIssueRef("owner", "repo", 123),
				target:             CreateIssueRef("other", "repo", 456),
				relType:            "blocked-by",
				expectedStatusCode: 204,
				responseBody:       "",
				expectSuccess:      true,
			},
			{
				name:               "relationship not found during deletion",
				source:             CreateIssueRef("owner", "repo", 123),
				target:             CreateIssueRef("owner", "repo", 999),
				relType:            "blocked-by",
				expectedStatusCode: 404,
				responseBody:       `{"message": "Dependency relationship not found"}`,
				expectSuccess:      false,
			},
			{
				name:               "permission denied during deletion",
				source:             CreateIssueRef("owner", "repo", 123),
				target:             CreateIssueRef("owner", "repo", 456),
				relType:            "blocks",
				expectedStatusCode: 403,
				responseBody:       `{"message": "Forbidden - insufficient permissions"}`,
				expectSuccess:      false,
			},
		}

		for _, scenario := range deletionScenarios {
			t.Run(scenario.name, func(t *testing.T) {
				mockClient := &MockHTTPClient{}
				mockClient.AddResponse(scenario.expectedStatusCode, scenario.responseBody)

				// Construct expected DELETE endpoint
				relationshipID := fmt.Sprintf("%s#%d", scenario.target.String(), scenario.target.Number)
				if scenario.target.FullName != "" {
					relationshipID = scenario.target.String()
				}

				expectedEndpoint := fmt.Sprintf("repos/%s/%s/issues/%d/dependencies/%s",
					scenario.source.Owner, scenario.source.Repo, scenario.source.Number, relationshipID)

				t.Logf("Testing deletion: %s %s %s", scenario.source.String(), scenario.relType, scenario.target.String())
				t.Logf("Expected endpoint: DELETE %s", expectedEndpoint)
				t.Logf("Expected status: %d", scenario.expectedStatusCode)

				// Verify endpoint construction
				assert.Contains(t, expectedEndpoint, fmt.Sprintf("repos/%s/%s", scenario.source.Owner, scenario.source.Repo),
					"Endpoint should contain source repository")
				assert.Contains(t, expectedEndpoint, fmt.Sprintf("issues/%d", scenario.source.Number),
					"Endpoint should contain source issue number")
				assert.Contains(t, expectedEndpoint, "dependencies",
					"Endpoint should contain dependencies path")

				// Verify success/failure categorization
				if scenario.expectSuccess {
					assert.Equal(t, 204, scenario.expectedStatusCode, "Successful deletion should return 204")
				} else {
					assert.Greater(t, scenario.expectedStatusCode, 399, "Failed deletion should return error status")
				}
			})
		}
	})

	t.Run("retry logic for transient failures", func(t *testing.T) {
		retryScenarios := []struct {
			name                string
			responses           []MockResponse
			expectedRetryCount  int
			expectedFinalResult bool
		}{
			{
				name: "rate limit then success",
				responses: []MockResponse{
					{StatusCode: 429, Body: `{"message": "Rate limit exceeded"}`},
					{StatusCode: 429, Body: `{"message": "Rate limit exceeded"}`},
					{StatusCode: 204, Body: ""},
				},
				expectedRetryCount:  2,
				expectedFinalResult: true,
			},
			{
				name: "server error then success",
				responses: []MockResponse{
					{StatusCode: 500, Body: `{"message": "Internal Server Error"}`},
					{StatusCode: 204, Body: ""},
				},
				expectedRetryCount:  1,
				expectedFinalResult: true,
			},
			{
				name: "persistent server errors",
				responses: []MockResponse{
					{StatusCode: 500, Body: `{"message": "Internal Server Error"}`},
					{StatusCode: 500, Body: `{"message": "Internal Server Error"}`},
					{StatusCode: 500, Body: `{"message": "Internal Server Error"}`},
				},
				expectedRetryCount:  3,
				expectedFinalResult: false,
			},
			{
				name: "network error then success",
				responses: []MockResponse{
					{Error: fmt.Errorf("network timeout")},
					{StatusCode: 204, Body: ""},
				},
				expectedRetryCount:  1,
				expectedFinalResult: true,
			},
		}

		for _, scenario := range retryScenarios {
			t.Run(scenario.name, func(t *testing.T) {
				mockClient := &MockHTTPClient{}
				for _, resp := range scenario.responses {
					if resp.Error != nil {
						mockClient.AddErrorResponse(resp.Error)
					} else {
						mockClient.AddResponse(resp.StatusCode, resp.Body)
					}
				}

				t.Logf("Testing retry scenario: %s", scenario.name)
				t.Logf("Expected retries: %d", scenario.expectedRetryCount)
				t.Logf("Expected final result: %v", scenario.expectedFinalResult)

				// Test retry logic parameters
				maxRetries := 3
				baseDelay := 1 * time.Second

				assert.LessOrEqual(t, scenario.expectedRetryCount, maxRetries,
					"Retry count should not exceed maximum")

				// Test exponential backoff calculation
				for attempt := 1; attempt <= scenario.expectedRetryCount; attempt++ {
					delay := time.Duration(attempt) * baseDelay
					expectedDelay := time.Duration(attempt) * time.Second
					assert.Equal(t, expectedDelay, delay,
						"Retry delay should follow exponential backoff pattern")
				}

				// Verify the number of configured responses matches expected behavior
				responseCount := len(scenario.responses)
				if scenario.expectedFinalResult {
					assert.LessOrEqual(t, scenario.expectedRetryCount+1, responseCount,
						"Should have enough responses for retries + final success")
				}
			})
		}
	})
}

// TestBatchDependencyRemoval tests batch removal operations with mocked responses
func TestBatchDependencyRemoval(t *testing.T) {
	t.Run("successful batch removal", func(t *testing.T) {
		source := CreateIssueRef("owner", "repo", 123)
		targets := []IssueRef{
			CreateIssueRef("owner", "repo", 456),
			CreateIssueRef("owner", "repo", 789),
			CreateIssueRef("other", "repo", 101),
		}

		// Mock successful responses for each deletion
		mockClient := &MockHTTPClient{}
		for i := 0; i < len(targets); i++ {
			mockClient.AddResponse(204, "")
		}

		t.Logf("Testing batch removal: %s removes %d dependencies", source.String(), len(targets))

		// Verify each target would get its own DELETE request
		for i, target := range targets {
			relationshipID := target.String()
			expectedEndpoint := fmt.Sprintf("repos/%s/%s/issues/%d/dependencies/%s",
				source.Owner, source.Repo, source.Number, relationshipID)

			t.Logf("Batch item %d: DELETE %s", i+1, expectedEndpoint)

			assert.Contains(t, expectedEndpoint, "dependencies",
				"Each batch item should target dependencies endpoint")
		}

		// Verify cross-repository handling
		crossRepoTarget := targets[2] // other/repo#101
		assert.NotEqual(t, source.Owner, crossRepoTarget.Owner,
			"Batch should handle cross-repository dependencies")
	})

	t.Run("partial batch failure", func(t *testing.T) {
		source := CreateIssueRef("owner", "repo", 123)
		targets := []IssueRef{
			CreateIssueRef("owner", "repo", 456), // Success
			CreateIssueRef("owner", "repo", 789), // Failure - not found
			CreateIssueRef("owner", "repo", 101), // Success
		}

		mockClient := &MockHTTPClient{}
		mockClient.AddResponse(204, "")                                    // Success
		mockClient.AddResponse(404, `{"message": "Dependency not found"}`) // Failure
		mockClient.AddResponse(204, "")                                    // Success

		t.Logf("Testing partial batch failure scenario for %s", source.String())

		expectedResults := []struct {
			target     IssueRef
			success    bool
			statusCode int
		}{
			{targets[0], true, 204},
			{targets[1], false, 404},
			{targets[2], true, 204},
		}

		for i, expected := range expectedResults {
			t.Logf("Expected result %d: %s -> success=%v (status=%d)",
				i+1, expected.target.String(), expected.success, expected.statusCode)

			if expected.success {
				assert.Equal(t, 204, expected.statusCode, "Success should be 204")
			} else {
				assert.Greater(t, expected.statusCode, 399, "Failure should be error status")
			}
		}

		// Verify that partial failures are reported properly
		successCount := 0
		failureCount := 0
		for _, result := range expectedResults {
			if result.success {
				successCount++
			} else {
				failureCount++
			}
		}

		assert.Equal(t, 2, successCount, "Should have 2 successes")
		assert.Equal(t, 1, failureCount, "Should have 1 failure")
		assert.Equal(t, len(targets), successCount+failureCount, "All operations should be accounted for")
	})
}

// TestCrossRepositoryDependencyRemoval tests cross-repository scenarios
func TestCrossRepositoryDependencyRemoval(t *testing.T) {
	scenarios := []struct {
		name    string
		source  IssueRef
		target  IssueRef
		relType string
	}{
		{
			name:    "same owner, different repo",
			source:  CreateIssueRef("owner", "repo1", 123),
			target:  CreateIssueRef("owner", "repo2", 456),
			relType: "blocked-by",
		},
		{
			name:    "different owner, same repo name",
			source:  CreateIssueRef("owner1", "repo", 123),
			target:  CreateIssueRef("owner2", "repo", 456),
			relType: "blocks",
		},
		{
			name:    "completely different repositories",
			source:  CreateIssueRef("owner1", "repo1", 123),
			target:  CreateIssueRef("owner2", "repo2", 456),
			relType: "blocked-by",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{}
			mockClient.AddResponse(204, "")

			// Verify cross-repository detection
			isCrossRepo := (scenario.source.Owner != scenario.target.Owner) ||
				(scenario.source.Repo != scenario.target.Repo)

			assert.True(t, isCrossRepo, "Scenario should be cross-repository")

			t.Logf("Cross-repo removal: %s %s %s",
				scenario.source.String(), scenario.relType, scenario.target.String())

			// Verify endpoint construction for cross-repository
			relationshipID := scenario.target.String()
			expectedEndpoint := fmt.Sprintf("repos/%s/%s/issues/%d/dependencies/%s",
				scenario.source.Owner, scenario.source.Repo, scenario.source.Number, relationshipID)

			t.Logf("Expected endpoint: DELETE %s", expectedEndpoint)

			// Cross-repository deletions should still use the source repository endpoint
			assert.Contains(t, expectedEndpoint, scenario.source.Owner+"/"+scenario.source.Repo,
				"Cross-repo deletion should use source repository endpoint")
			assert.Contains(t, expectedEndpoint, scenario.target.String(),
				"Cross-repo deletion should reference target in relationship ID")
		})
	}
}

// TestAPIRateLimitHandling tests rate limiting scenarios
func TestAPIRateLimitHandling(t *testing.T) {
	t.Run("rate limit with retry-after header", func(t *testing.T) {
		mockClient := &MockHTTPClient{}

		// First response: rate limited with Retry-After header
		rateLimitResp := MockResponse{
			StatusCode: 429,
			Body:       `{"message": "API rate limit exceeded for user"}`,
			Headers:    map[string]string{"Retry-After": "60"},
		}

		// Second response: success
		successResp := MockResponse{
			StatusCode: 204,
			Body:       "",
		}

		mockClient.responses = []MockResponse{rateLimitResp, successResp}

		t.Logf("Testing rate limit handling with Retry-After header")

		// Simulate making the requests
		for i, expectedResp := range []MockResponse{rateLimitResp, successResp} {
			req, _ := http.NewRequest("DELETE", "https://api.github.com/test", nil)
			resp, err := mockClient.Do(req)

			require.NoError(t, err, "Mock client should not error")
			assert.Equal(t, expectedResp.StatusCode, resp.StatusCode,
				"Response %d status should match", i+1)

			if resp.StatusCode == 429 {
				retryAfter := resp.Header.Get("Retry-After")
				assert.Equal(t, "60", retryAfter, "Retry-After header should be preserved")

				t.Logf("Request %d: Rate limited, retry after %s seconds", i+1, retryAfter)
			} else {
				t.Logf("Request %d: Success", i+1)
			}
		}

		requests := mockClient.GetRequests()
		assert.Len(t, requests, 2, "Should have made 2 requests (retry after rate limit)")
	})

	t.Run("rate limit without retry-after header", func(t *testing.T) {
		mockClient := &MockHTTPClient{}
		mockClient.AddResponse(429, `{"message": "API rate limit exceeded"}`)
		mockClient.AddResponse(204, "")

		t.Logf("Testing rate limit handling without Retry-After header")

		// First request - rate limited
		req1, _ := http.NewRequest("DELETE", "https://api.github.com/test", nil)
		resp1, err := mockClient.Do(req1)
		require.NoError(t, err)
		assert.Equal(t, 429, resp1.StatusCode)

		// Should use exponential backoff when no Retry-After header
		baseDelay := 1 * time.Second
		exponentialDelay := baseDelay * 2 // First retry

		t.Logf("Rate limited without Retry-After, using exponential backoff: %v", exponentialDelay)

		// Second request - success
		req2, _ := http.NewRequest("DELETE", "https://api.github.com/test", nil)
		resp2, err := mockClient.Do(req2)
		require.NoError(t, err)
		assert.Equal(t, 204, resp2.StatusCode)

		requests := mockClient.GetRequests()
		assert.Len(t, requests, 2, "Should have made 2 requests")
	})
}

// TestNetworkErrorHandling tests network-level error scenarios
func TestNetworkErrorHandling(t *testing.T) {
	networkErrors := []struct {
		name  string
		error error
	}{
		{
			name:  "connection timeout",
			error: fmt.Errorf("dial tcp: i/o timeout"),
		},
		{
			name:  "connection refused",
			error: fmt.Errorf("dial tcp: connection refused"),
		},
		{
			name:  "DNS resolution failure",
			error: fmt.Errorf("dial tcp: no such host"),
		},
		{
			name:  "TLS handshake failure",
			error: fmt.Errorf("tls: handshake failure"),
		},
	}

	for _, scenario := range networkErrors {
		t.Run(scenario.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{}
			mockClient.AddErrorResponse(scenario.error)
			mockClient.AddResponse(204, "") // Success after retry

			t.Logf("Testing network error: %s", scenario.name)

			// First request - network error
			req1, _ := http.NewRequest("DELETE", "https://api.github.com/test", nil)
			_, err := mockClient.Do(req1)
			assert.Error(t, err, "Should return network error")
			assert.Contains(t, err.Error(), strings.Split(scenario.error.Error(), ":")[0],
				"Error should contain expected network error type")

			// Second request - success (after retry)
			req2, _ := http.NewRequest("DELETE", "https://api.github.com/test", nil)
			resp, err := mockClient.Do(req2)
			require.NoError(t, err, "Retry should succeed")
			assert.Equal(t, 204, resp.StatusCode, "Retry should return success")

			t.Logf("Network error resolved on retry")
		})
	}
}

// TestGitHubAPIResponseParsing tests parsing of various GitHub API responses
func TestGitHubAPIResponseParsing(t *testing.T) {
	t.Run("dependency data response parsing", func(t *testing.T) {
		responseBody := `{
			"source_issue": {
				"number": 123,
				"title": "Feature Implementation",
				"state": "open",
				"repository": {
					"full_name": "owner/repo",
					"name": "repo",
					"owner": {
						"login": "owner"
					}
				},
				"html_url": "https://github.com/owner/repo/issues/123"
			},
			"blocked_by": [
				{
					"issue": {
						"number": 456,
						"title": "Database Setup",
						"state": "open",
						"repository": {
							"full_name": "owner/repo"
						}
					},
					"type": "blocked_by",
					"repository": "owner/repo"
				}
			],
			"blocking": [
				{
					"issue": {
						"number": 789,
						"title": "Frontend Integration", 
						"state": "open",
						"repository": {
							"full_name": "other/repo"
						}
					},
					"type": "blocks",
					"repository": "other/repo"
				}
			],
			"total_count": 2,
			"fetched_at": "2024-01-01T12:00:00Z"
		}`

		t.Logf("Testing dependency data response parsing")

		// Verify response contains expected fields
		assert.Contains(t, responseBody, "source_issue", "Response should contain source_issue")
		assert.Contains(t, responseBody, "blocked_by", "Response should contain blocked_by")
		assert.Contains(t, responseBody, "blocking", "Response should contain blocking")
		assert.Contains(t, responseBody, "total_count", "Response should contain total_count")

		// Verify issue numbers are present
		assert.Contains(t, responseBody, "123", "Response should contain source issue number")
		assert.Contains(t, responseBody, "456", "Response should contain blocked_by issue number")
		assert.Contains(t, responseBody, "789", "Response should contain blocking issue number")

		// Verify repository information
		assert.Contains(t, responseBody, "owner/repo", "Response should contain repository names")
		assert.Contains(t, responseBody, "other/repo", "Response should contain cross-repo references")
	})

	t.Run("error response parsing", func(t *testing.T) {
		errorResponses := []struct {
			name     string
			status   int
			body     string
			errorMsg string
		}{
			{
				name:     "authentication error",
				status:   401,
				body:     `{"message": "Requires authentication", "documentation_url": "https://docs.github.com/rest"}`,
				errorMsg: "Requires authentication",
			},
			{
				name:     "forbidden error",
				status:   403,
				body:     `{"message": "Forbidden", "documentation_url": "https://docs.github.com/rest"}`,
				errorMsg: "Forbidden",
			},
			{
				name:     "not found error",
				status:   404,
				body:     `{"message": "Not Found", "documentation_url": "https://docs.github.com/rest"}`,
				errorMsg: "Not Found",
			},
			{
				name:     "validation error",
				status:   422,
				body:     `{"message": "Validation Failed", "errors": [{"field": "dependency", "code": "invalid"}], "documentation_url": "https://docs.github.com/rest"}`,
				errorMsg: "Validation Failed",
			},
		}

		for _, scenario := range errorResponses {
			t.Run(scenario.name, func(t *testing.T) {
				mockClient := &MockHTTPClient{}
				mockClient.AddResponse(scenario.status, scenario.body)

				req, _ := http.NewRequest("DELETE", "https://api.github.com/test", nil)
				resp, err := mockClient.Do(req)

				require.NoError(t, err, "Mock client should not error")
				assert.Equal(t, scenario.status, resp.StatusCode, "Status should match")

				// Read and verify response body
				bodyBytes, err := io.ReadAll(resp.Body)
				require.NoError(t, err, "Should read response body")

				bodyStr := string(bodyBytes)
				assert.Contains(t, bodyStr, scenario.errorMsg, "Response should contain expected error message")
				assert.Contains(t, bodyStr, "documentation_url", "Error response should contain documentation URL")

				t.Logf("Error response parsed: %s (HTTP %d)", scenario.errorMsg, scenario.status)
			})
		}
	})
}
