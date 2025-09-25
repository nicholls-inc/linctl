package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIssueCreateInput(t *testing.T) {
	tests := []struct {
		name     string
		input    IssueCreateInput
		expected map[string]interface{}
	}{
		{
			name: "basic issue input",
			input: IssueCreateInput{
				Title:  "Test Issue",
				TeamID: "team-123",
			},
			expected: map[string]interface{}{
				"title":  "Test Issue",
				"teamId": "team-123",
			},
		},
		{
			name: "issue with actor attribution",
			input: IssueCreateInput{
				Title:          "Test Issue",
				TeamID:         "team-123",
				CreateAsUser:   stringPtr("AI Agent"),
				DisplayIconURL: stringPtr("https://example.com/agent.png"),
			},
			expected: map[string]interface{}{
				"title":          "Test Issue",
				"teamId":         "team-123",
				"createAsUser":   "AI Agent",
				"displayIconUrl": "https://example.com/agent.png",
			},
		},
		{
			name: "issue with optional fields",
			input: IssueCreateInput{
				Title:       "Test Issue",
				TeamID:      "team-123",
				Description: stringPtr("Test description"),
				Priority:    intPtr(2),
				AssigneeID:  stringPtr("user-456"),
			},
			expected: map[string]interface{}{
				"title":       "Test Issue",
				"teamId":      "team-123",
				"description": "Test description",
				"priority":    float64(2), // JSON unmarshals numbers as float64
				"assigneeId":  "user-456",
			},
		},
		{
			name: "issue with all fields",
			input: IssueCreateInput{
				Title:          "Complete Issue",
				TeamID:         "team-123",
				Description:    stringPtr("Complete description"),
				Priority:       intPtr(1),
				AssigneeID:     stringPtr("user-456"),
				StateID:        stringPtr("state-789"),
				LabelIDs:       []string{"label-1", "label-2"},
				ProjectID:      stringPtr("project-abc"),
				CycleID:        stringPtr("cycle-def"),
				Estimate:       float64Ptr(5.0),
				DueDate:        stringPtr("2024-12-31"),
				CreateAsUser:   stringPtr("AI Agent"),
				DisplayIconURL: stringPtr("https://example.com/agent.png"),
			},
			expected: map[string]interface{}{
				"title":          "Complete Issue",
				"teamId":         "team-123",
				"description":    "Complete description",
				"priority":       float64(1),
				"assigneeId":     "user-456",
				"stateId":        "state-789",
				"labelIds":       []interface{}{"label-1", "label-2"},
				"projectId":      "project-abc",
				"cycleId":        "cycle-def",
				"estimate":       5.0,
				"dueDate":        "2024-12-31",
				"createAsUser":   "AI Agent",
				"displayIconUrl": "https://example.com/agent.png",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON and back to verify structure
			jsonData, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			var result map[string]interface{}
			err = json.Unmarshal(jsonData, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			// Compare expected fields
			for key, expectedValue := range tt.expected {
				actualValue, exists := result[key]
				if !exists {
					t.Errorf("Expected field '%s' not found in result", key)
					continue
				}

				// Handle slice comparison
				if expectedSlice, ok := expectedValue.([]string); ok {
					actualSlice, ok := actualValue.([]interface{})
					if !ok {
						t.Errorf("Expected field '%s' to be slice, got %T", key, actualValue)
						continue
					}
					if len(expectedSlice) != len(actualSlice) {
						t.Errorf("Expected field '%s' slice length %d, got %d", key, len(expectedSlice), len(actualSlice))
						continue
					}
					for i, expected := range expectedSlice {
						if actual, ok := actualSlice[i].(string); !ok || actual != expected {
							t.Errorf("Expected field '%s'[%d] to be '%s', got '%v'", key, i, expected, actualSlice[i])
						}
					}
				} else if expectedInterfaceSlice, ok := expectedValue.([]interface{}); ok {
					// Handle []interface{} comparison
					actualSlice, ok := actualValue.([]interface{})
					if !ok {
						t.Errorf("Expected field '%s' to be slice, got %T", key, actualValue)
						continue
					}
					if len(expectedInterfaceSlice) != len(actualSlice) {
						t.Errorf("Expected field '%s' slice length %d, got %d", key, len(expectedInterfaceSlice), len(actualSlice))
						continue
					}
					for i, expected := range expectedInterfaceSlice {
						if actualSlice[i] != expected {
							t.Errorf("Expected field '%s'[%d] to be '%v', got '%v'", key, i, expected, actualSlice[i])
						}
					}
				} else if actualValue != expectedValue {
					t.Errorf("Expected field '%s' to be %v, got %v", key, expectedValue, actualValue)
				}
			}

			// Verify no unexpected fields for basic cases
			if tt.name == "basic issue input" {
				for key := range result {
					if _, expected := tt.expected[key]; !expected {
						t.Errorf("Unexpected field '%s' in result", key)
					}
				}
			}
		})
	}
}

func TestCommentCreateInput(t *testing.T) {
	tests := []struct {
		name     string
		input    CommentCreateInput
		expected map[string]interface{}
	}{
		{
			name: "basic comment input",
			input: CommentCreateInput{
				IssueID: "issue-123",
				Body:    "Test comment",
			},
			expected: map[string]interface{}{
				"issueId": "issue-123",
				"body":    "Test comment",
			},
		},
		{
			name: "comment with actor attribution",
			input: CommentCreateInput{
				IssueID:        "issue-123",
				Body:           "Test comment",
				CreateAsUser:   stringPtr("AI Agent"),
				DisplayIconURL: stringPtr("https://example.com/agent.png"),
			},
			expected: map[string]interface{}{
				"issueId":        "issue-123",
				"body":           "Test comment",
				"createAsUser":   "AI Agent",
				"displayIconUrl": "https://example.com/agent.png",
			},
		},
		{
			name: "comment with only actor name",
			input: CommentCreateInput{
				IssueID:      "issue-123",
				Body:         "Test comment",
				CreateAsUser: stringPtr("AI Agent"),
			},
			expected: map[string]interface{}{
				"issueId":      "issue-123",
				"body":         "Test comment",
				"createAsUser": "AI Agent",
			},
		},
		{
			name: "comment with only avatar URL",
			input: CommentCreateInput{
				IssueID:        "issue-123",
				Body:           "Test comment",
				DisplayIconURL: stringPtr("https://example.com/agent.png"),
			},
			expected: map[string]interface{}{
				"issueId":        "issue-123",
				"body":           "Test comment",
				"displayIconUrl": "https://example.com/agent.png",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON and back to verify structure
			jsonData, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			var result map[string]interface{}
			err = json.Unmarshal(jsonData, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			// Compare expected fields
			for key, expectedValue := range tt.expected {
				actualValue, exists := result[key]
				if !exists {
					t.Errorf("Expected field '%s' not found in result", key)
					continue
				}

				if actualValue != expectedValue {
					t.Errorf("Expected field '%s' to be %v, got %v", key, expectedValue, actualValue)
				}
			}

			// Verify no unexpected fields for basic cases
			if tt.name == "basic comment input" {
				for key := range result {
					if _, expected := tt.expected[key]; !expected {
						t.Errorf("Unexpected field '%s' in result", key)
					}
				}
			}
		})
	}
}

func TestCreateIssue(t *testing.T) {
	tests := []struct {
		name           string
		input          IssueCreateInput
		serverResponse map[string]interface{}
		expectError    bool
		errorMsg       string
	}{
		{
			name: "successful issue creation",
			input: IssueCreateInput{
				Title:  "Test Issue",
				TeamID: "team-123",
			},
			serverResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"issueCreate": map[string]interface{}{
						"issue": map[string]interface{}{
							"id":         "issue-456",
							"identifier": "TEST-123",
							"title":      "Test Issue",
							"teamId":     "team-123",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "issue creation with actor",
			input: IssueCreateInput{
				Title:          "Test Issue",
				TeamID:         "team-123",
				CreateAsUser:   stringPtr("AI Agent"),
				DisplayIconURL: stringPtr("https://example.com/agent.png"),
			},
			serverResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"issueCreate": map[string]interface{}{
						"issue": map[string]interface{}{
							"id":         "issue-456",
							"identifier": "TEST-123",
							"title":      "Test Issue",
							"teamId":     "team-123",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "server error",
			input: IssueCreateInput{
				Title:  "Test Issue",
				TeamID: "team-123",
			},
			serverResponse: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"message": "Team not found",
					},
				},
			},
			expectError: true,
			errorMsg:    "GraphQL errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and headers
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}

				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
				}

				if r.Header.Get("Authorization") != "test-auth-header" {
					t.Errorf("Expected Authorization test-auth-header, got %s", r.Header.Get("Authorization"))
				}

				// Parse request body to verify input
				var requestBody map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&requestBody)
				if err != nil {
					t.Fatalf("Failed to decode request body: %v", err)
				}

				// Verify query contains CreateIssue mutation
				query, ok := requestBody["query"].(string)
				if !ok || !strings.Contains(query, "mutation CreateIssue") {
					t.Errorf("Expected CreateIssue mutation in query, got: %s", query)
				}

				// Verify variables contain input
				variables, ok := requestBody["variables"].(map[string]interface{})
				if !ok {
					t.Error("Expected variables in request")
				} else {
					input, ok := variables["input"].(map[string]interface{})
					if !ok {
						t.Error("Expected input in variables")
					} else {
						// Verify basic fields
						if input["title"] != tt.input.Title {
							t.Errorf("Expected title %s, got %v", tt.input.Title, input["title"])
						}
						if input["teamId"] != tt.input.TeamID {
							t.Errorf("Expected teamId %s, got %v", tt.input.TeamID, input["teamId"])
						}

						// Verify actor fields if present
						if tt.input.CreateAsUser != nil {
							if input["createAsUser"] != *tt.input.CreateAsUser {
								t.Errorf("Expected createAsUser %s, got %v", *tt.input.CreateAsUser, input["createAsUser"])
							}
						}
						if tt.input.DisplayIconURL != nil {
							if input["displayIconUrl"] != *tt.input.DisplayIconURL {
								t.Errorf("Expected displayIconUrl %s, got %v", *tt.input.DisplayIconURL, input["displayIconUrl"])
							}
						}
					}
				}

				// Send response
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			// Create client
			client := NewClientWithURL(server.URL, "test-auth-header")

			// Test CreateIssue
			issue, err := client.CreateIssue(context.Background(), tt.input)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if issue == nil {
				t.Fatal("Expected issue to be returned")
			}

			// Verify issue fields
			expectedIssue := tt.serverResponse["data"].(map[string]interface{})["issueCreate"].(map[string]interface{})["issue"].(map[string]interface{})
			if issue.ID != expectedIssue["id"] {
				t.Errorf("Expected issue ID %s, got %s", expectedIssue["id"], issue.ID)
			}
			if issue.Identifier != expectedIssue["identifier"] {
				t.Errorf("Expected issue identifier %s, got %s", expectedIssue["identifier"], issue.Identifier)
			}
			if issue.Title != expectedIssue["title"] {
				t.Errorf("Expected issue title %s, got %s", expectedIssue["title"], issue.Title)
			}
		})
	}
}

func TestCreateComment(t *testing.T) {
	tests := []struct {
		name           string
		input          CommentCreateInput
		serverResponse map[string]interface{}
		expectError    bool
		errorMsg       string
	}{
		{
			name: "successful comment creation",
			input: CommentCreateInput{
				IssueID: "issue-123",
				Body:    "Test comment",
			},
			serverResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"commentCreate": map[string]interface{}{
						"comment": map[string]interface{}{
							"id":   "comment-456",
							"body": "Test comment",
							"user": map[string]interface{}{
								"id":   "user-789",
								"name": "Test User",
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "comment creation with actor",
			input: CommentCreateInput{
				IssueID:        "issue-123",
				Body:           "Test comment",
				CreateAsUser:   stringPtr("AI Agent"),
				DisplayIconURL: stringPtr("https://example.com/agent.png"),
			},
			serverResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"commentCreate": map[string]interface{}{
						"comment": map[string]interface{}{
							"id":   "comment-456",
							"body": "Test comment",
							"user": map[string]interface{}{
								"id":   "user-789",
								"name": "AI Agent",
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "server error",
			input: CommentCreateInput{
				IssueID: "issue-123",
				Body:    "Test comment",
			},
			serverResponse: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"message": "Issue not found",
					},
				},
			},
			expectError: true,
			errorMsg:    "GraphQL errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and headers
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}

				// Parse request body to verify input
				var requestBody map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&requestBody)
				if err != nil {
					t.Fatalf("Failed to decode request body: %v", err)
				}

				// Verify query contains CreateComment mutation
				query, ok := requestBody["query"].(string)
				if !ok || !strings.Contains(query, "mutation CreateComment") {
					t.Errorf("Expected CreateComment mutation in query, got: %s", query)
				}

				// Verify variables contain input
				variables, ok := requestBody["variables"].(map[string]interface{})
				if !ok {
					t.Error("Expected variables in request")
				} else {
					input, ok := variables["input"].(map[string]interface{})
					if !ok {
						t.Error("Expected input in variables")
					} else {
						// Verify basic fields
						if input["issueId"] != tt.input.IssueID {
							t.Errorf("Expected issueId %s, got %v", tt.input.IssueID, input["issueId"])
						}
						if input["body"] != tt.input.Body {
							t.Errorf("Expected body %s, got %v", tt.input.Body, input["body"])
						}

						// Verify actor fields if present
						if tt.input.CreateAsUser != nil {
							if input["createAsUser"] != *tt.input.CreateAsUser {
								t.Errorf("Expected createAsUser %s, got %v", *tt.input.CreateAsUser, input["createAsUser"])
							}
						}
						if tt.input.DisplayIconURL != nil {
							if input["displayIconUrl"] != *tt.input.DisplayIconURL {
								t.Errorf("Expected displayIconUrl %s, got %v", *tt.input.DisplayIconURL, input["displayIconUrl"])
							}
						}
					}
				}

				// Send response
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			// Create client
			client := NewClientWithURL(server.URL, "test-auth-header")

			// Test CreateComment
			comment, err := client.CreateComment(context.Background(), tt.input)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if comment == nil {
				t.Fatal("Expected comment to be returned")
			}

			// Verify comment fields
			expectedComment := tt.serverResponse["data"].(map[string]interface{})["commentCreate"].(map[string]interface{})["comment"].(map[string]interface{})
			if comment.ID != expectedComment["id"] {
				t.Errorf("Expected comment ID %s, got %s", expectedComment["id"], comment.ID)
			}
			if comment.Body != expectedComment["body"] {
				t.Errorf("Expected comment body %s, got %s", expectedComment["body"], comment.Body)
			}
		})
	}
}

func TestCreateCommentSimple(t *testing.T) {
	// Test backward compatibility method
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"commentCreate": map[string]interface{}{
					"comment": map[string]interface{}{
						"id":   "comment-456",
						"body": "Simple comment",
						"user": map[string]interface{}{
							"id":   "user-789",
							"name": "Test User",
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL, "test-auth-header")

	comment, err := client.CreateCommentSimple(context.Background(), "issue-123", "Simple comment")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if comment == nil {
		t.Fatal("Expected comment to be returned")
	}

	if comment.ID != "comment-456" {
		t.Errorf("Expected comment ID comment-456, got %s", comment.ID)
	}

	if comment.Body != "Simple comment" {
		t.Errorf("Expected comment body 'Simple comment', got %s", comment.Body)
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}
