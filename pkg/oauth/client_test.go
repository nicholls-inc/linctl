package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewOAuthClient(t *testing.T) {
	tests := []struct {
		name         string
		clientID     string
		clientSecret string
		baseURL      string
		expectedURL  string
	}{
		{
			name:         "with custom base URL",
			clientID:     "test-id",
			clientSecret: "test-secret",
			baseURL:      "https://custom.linear.app",
			expectedURL:  "https://custom.linear.app",
		},
		{
			name:         "with empty base URL defaults to Linear",
			clientID:     "test-id",
			clientSecret: "test-secret",
			baseURL:      "",
			expectedURL:  "https://api.linear.app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOAuthClient(tt.clientID, tt.clientSecret, tt.baseURL)
			
			if client.clientID != tt.clientID {
				t.Errorf("Expected clientID %s, got %s", tt.clientID, client.clientID)
			}
			if client.clientSecret != tt.clientSecret {
				t.Errorf("Expected clientSecret %s, got %s", tt.clientSecret, client.clientSecret)
			}
			if client.baseURL != tt.expectedURL {
				t.Errorf("Expected baseURL %s, got %s", tt.expectedURL, client.baseURL)
			}
			if client.httpClient == nil {
				t.Error("Expected httpClient to be initialized")
			}
			if client.httpClient.Timeout != 30*time.Second {
				t.Errorf("Expected timeout 30s, got %v", client.httpClient.Timeout)
			}
		})
	}
}

func TestOAuthClient_GetAccessToken_Success(t *testing.T) {
	// Mock server that returns a successful OAuth response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/oauth/token" {
			t.Errorf("Expected /oauth/token path, got %s", r.URL.Path)
		}

		// Verify headers
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("Expected application/x-www-form-urlencoded content type")
		}

		// Verify basic auth
		username, password, ok := r.BasicAuth()
		if !ok {
			t.Error("Expected basic auth to be present")
		}
		if username != "test-client-id" || password != "test-client-secret" {
			t.Errorf("Expected basic auth test-client-id:test-client-secret, got %s:%s", username, password)
		}

		// Verify form data
		err := r.ParseForm()
		if err != nil {
			t.Errorf("Failed to parse form: %v", err)
		}
		if r.Form.Get("grant_type") != "client_credentials" {
			t.Errorf("Expected grant_type=client_credentials, got %s", r.Form.Get("grant_type"))
		}
		expectedScope := "read write"
		if r.Form.Get("scope") != expectedScope {
			t.Errorf("Expected scope=%s, got %s", expectedScope, r.Form.Get("scope"))
		}

		// Return successful response
		response := TokenResponse{
			AccessToken: "test-access-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
			Scope:       "read write",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewOAuthClient("test-client-id", "test-client-secret", server.URL)
	scopes := []string{"read", "write"}

	token, err := client.GetAccessToken(context.Background(), scopes)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token.AccessToken != "test-access-token" {
		t.Errorf("Expected access token 'test-access-token', got %s", token.AccessToken)
	}
	if token.TokenType != "Bearer" {
		t.Errorf("Expected token type 'Bearer', got %s", token.TokenType)
	}
	if token.ExpiresIn != 3600 {
		t.Errorf("Expected expires in 3600, got %d", token.ExpiresIn)
	}
	if token.Scope != "read write" {
		t.Errorf("Expected scope 'read write', got %s", token.Scope)
	}
}

func TestOAuthClient_GetAccessToken_DefaultTokenType(t *testing.T) {
	// Mock server that returns response without token_type
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := TokenResponse{
			AccessToken: "test-access-token",
			// TokenType omitted
			ExpiresIn: 3600,
			Scope:     "read write",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewOAuthClient("test-client-id", "test-client-secret", server.URL)
	scopes := []string{"read", "write"}

	token, err := client.GetAccessToken(context.Background(), scopes)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token.TokenType != "Bearer" {
		t.Errorf("Expected default token type 'Bearer', got %s", token.TokenType)
	}
}

func TestOAuthClient_GetAccessToken_HTTPError(t *testing.T) {
	// Mock server that returns HTTP 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "invalid_client",
			"error_description": "Invalid client: client is invalid",
		})
	}))
	defer server.Close()

	client := NewOAuthClient("test-client-id", "test-client-secret", server.URL)
	scopes := []string{"read", "write"}

	_, err := client.GetAccessToken(context.Background(), scopes)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "OAuth request failed (401)") {
		t.Errorf("Expected error to contain OAuth request failed (401), got %s", err.Error())
	}
}

func TestOAuthClient_GetAccessToken_EmptyAccessToken(t *testing.T) {
	// Mock server that returns empty access token
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := TokenResponse{
			AccessToken: "", // Empty access token
			TokenType:   "Bearer",
			ExpiresIn:   3600,
			Scope:       "read write",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewOAuthClient("test-client-id", "test-client-secret", server.URL)
	scopes := []string{"read", "write"}

	_, err := client.GetAccessToken(context.Background(), scopes)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedError := "received empty access token"
	if err.Error() != expectedError {
		t.Errorf("Expected error %s, got %s", expectedError, err.Error())
	}
}

func TestOAuthClient_GetAccessToken_InvalidJSON(t *testing.T) {
	// Mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewOAuthClient("test-client-id", "test-client-secret", server.URL)
	scopes := []string{"read", "write"}

	_, err := client.GetAccessToken(context.Background(), scopes)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to decode token response") {
		t.Errorf("Expected error to contain 'failed to decode token response', got %s", err.Error())
	}
}

func TestOAuthClient_ValidateToken_Success(t *testing.T) {
	// Mock server that returns successful GraphQL response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/graphql" {
			t.Errorf("Expected /graphql path, got %s", r.URL.Path)
		}

		// Verify headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected application/json content type")
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Bearer test-token authorization, got %s", r.Header.Get("Authorization"))
		}

		// Return successful GraphQL response
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data": {"viewer": {"id": "user-123", "name": "Test User"}}}`))
	}))
	defer server.Close()

	client := NewOAuthClient("test-client-id", "test-client-secret", server.URL)

	err := client.ValidateToken(context.Background(), "test-token")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestOAuthClient_ValidateToken_Unauthorized(t *testing.T) {
	// Mock server that returns 401 Unauthorized
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"errors": [{"message": "Invalid token"}]}`))
	}))
	defer server.Close()

	client := NewOAuthClient("test-client-id", "test-client-secret", server.URL)

	err := client.ValidateToken(context.Background(), "invalid-token")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedError := "access token is invalid or expired"
	if err.Error() != expectedError {
		t.Errorf("Expected error %s, got %s", expectedError, err.Error())
	}
}

func TestOAuthClient_ValidateToken_ServerError(t *testing.T) {
	// Mock server that returns 500 Internal Server Error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewOAuthClient("test-client-id", "test-client-secret", server.URL)

	err := client.ValidateToken(context.Background(), "test-token")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "token validation failed with status: 500") {
		t.Errorf("Expected error to contain status 500, got %s", err.Error())
	}
}

func TestOAuthClient_ValidateToken_InvalidJSON(t *testing.T) {
	// Mock server that returns invalid JSON for validation query
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewOAuthClient("test-client-id", "test-client-secret", server.URL)

	// This should still succeed because we only check the HTTP status code
	err := client.ValidateToken(context.Background(), "test-token")
	if err != nil {
		t.Errorf("Expected no error for invalid JSON response, got %v", err)
	}
}

func TestOAuthClient_ContextCancellation(t *testing.T) {
	// Mock server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		response := TokenResponse{
			AccessToken: "test-access-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
			Scope:       "read write",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewOAuthClient("test-client-id", "test-client-secret", server.URL)
	
	// Create context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GetAccessToken(ctx, []string{"read", "write"})
	if err == nil {
		t.Fatal("Expected context cancellation error, got nil")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Expected context canceled error, got %s", err.Error())
	}
}