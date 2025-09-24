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

// Tests for enhanced OAuth client functionality

func TestNewOAuthClientFromConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &Config{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				BaseURL:      "https://api.linear.app",
				Scopes:       []string{"read", "write"},
			},
			expectError: false,
		},
		{
			name: "invalid config - missing client ID",
			config: &Config{
				ClientSecret: "test-client-secret",
				BaseURL:      "https://api.linear.app",
				Scopes:       []string{"read", "write"},
			},
			expectError: true,
			errorMsg:    "invalid OAuth config",
		},
		{
			name: "invalid config - missing client secret",
			config: &Config{
				ClientID: "test-client-id",
				BaseURL:  "https://api.linear.app",
				Scopes:   []string{"read", "write"},
			},
			expectError: true,
			errorMsg:    "invalid OAuth config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewOAuthClientFromConfig(tt.config)
			
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
			
			if client == nil {
				t.Fatal("Expected client to be created")
			}
			
			if client.clientID != tt.config.ClientID {
				t.Errorf("Expected client ID '%s', got '%s'", tt.config.ClientID, client.clientID)
			}
			
			if client.clientSecret != tt.config.ClientSecret {
				t.Errorf("Expected client secret '%s', got '%s'", tt.config.ClientSecret, client.clientSecret)
			}
			
			if client.baseURL != tt.config.BaseURL {
				t.Errorf("Expected base URL '%s', got '%s'", tt.config.BaseURL, client.baseURL)
			}
			
			if client.tokenStore == nil {
				t.Error("Expected token store to be initialized")
			}
			
			if client.config == nil {
				t.Error("Expected config to be stored")
			}
		})
	}
}

func TestOAuthClient_GetValidToken(t *testing.T) {
	// Create a temporary directory for token storage
	tempDir := t.TempDir()
	
	tests := []struct {
		name           string
		setupToken     *TokenResponse
		tokenExpired   bool
		serverResponse *TokenResponse
		expectError    bool
		errorMsg       string
	}{
		{
			name: "valid stored token",
			setupToken: &TokenResponse{
				AccessToken: "stored-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				Scope:       "read write",
			},
			tokenExpired: false,
			expectError:  false,
		},
		{
			name: "expired token - refresh successful",
			setupToken: &TokenResponse{
				AccessToken: "expired-token",
				TokenType:   "Bearer",
				ExpiresIn:   1, // Will be expired due to buffer
				Scope:       "read write",
			},
			tokenExpired: true,
			serverResponse: &TokenResponse{
				AccessToken: "new-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				Scope:       "read write",
			},
			expectError: false,
		},
		{
			name:        "no stored token - get new token",
			setupToken:  nil,
			tokenExpired: false,
			serverResponse: &TokenResponse{
				AccessToken: "fresh-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				Scope:       "read write",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.serverResponse != nil {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(tt.serverResponse)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}))
			defer server.Close()

			// Create client with custom token store path
			testTokenPath := tempDir + "/test-token-" + strings.ReplaceAll(tt.name, " ", "-") + ".json"
			client := NewOAuthClient("test-client-id", "test-client-secret", server.URL)
			client.tokenStore = NewTokenStoreWithPath(testTokenPath)

			// Setup stored token if provided
			if tt.setupToken != nil {
				// Adjust expiry time based on test case
				if tt.tokenExpired {
					// Make token expire soon (within buffer)
					time.Sleep(10 * time.Millisecond) // Ensure some time passes
				}
				err := client.tokenStore.SaveToken(tt.setupToken)
				if err != nil {
					t.Fatalf("Failed to setup test token: %v", err)
				}
				
				if tt.tokenExpired {
					// Wait a bit more to ensure expiry
					time.Sleep(10 * time.Millisecond)
				}
			}

			// Test GetValidToken
			token, err := client.GetValidToken(context.Background(), []string{"read", "write"})

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

			if token == nil {
				t.Fatal("Expected token to be returned")
			}

			// Verify token content
			if tt.setupToken != nil && !tt.tokenExpired {
				// Should return stored token
				if token.AccessToken != tt.setupToken.AccessToken {
					t.Errorf("Expected stored token '%s', got '%s'", tt.setupToken.AccessToken, token.AccessToken)
				}
			} else if tt.serverResponse != nil {
				// Should return new token from server
				if token.AccessToken != tt.serverResponse.AccessToken {
					t.Errorf("Expected new token '%s', got '%s'", tt.serverResponse.AccessToken, token.AccessToken)
				}
			}
		})
	}
}

func TestOAuthClient_RefreshToken(t *testing.T) {
	tempDir := t.TempDir()
	
	tests := []struct {
		name           string
		serverResponse *TokenResponse
		expectError    bool
		errorMsg       string
	}{
		{
			name: "successful refresh",
			serverResponse: &TokenResponse{
				AccessToken: "refreshed-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				Scope:       "read write",
			},
			expectError: false,
		},
		{
			name:        "server error",
			serverResponse: nil,
			expectError: true,
			errorMsg:    "failed to refresh token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.serverResponse != nil {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(tt.serverResponse)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}))
			defer server.Close()

			// Create client with custom token store path
			testTokenPath := tempDir + "/test-token-" + strings.ReplaceAll(tt.name, " ", "-") + ".json"
			client := NewOAuthClient("test-client-id", "test-client-secret", server.URL)
			client.tokenStore = NewTokenStoreWithPath(testTokenPath)

			// Test RefreshToken
			token, err := client.RefreshToken(context.Background(), []string{"read", "write"})

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

			if token == nil {
				t.Fatal("Expected token to be returned")
			}

			if token.AccessToken != tt.serverResponse.AccessToken {
				t.Errorf("Expected token '%s', got '%s'", tt.serverResponse.AccessToken, token.AccessToken)
			}

			// Verify token was saved
			storedToken, err := client.tokenStore.LoadToken()
			if err != nil {
				t.Fatalf("Failed to load stored token: %v", err)
			}

			if storedToken.AccessToken != tt.serverResponse.AccessToken {
				t.Errorf("Expected stored token '%s', got '%s'", tt.serverResponse.AccessToken, storedToken.AccessToken)
			}
		})
	}
}

func TestOAuthClient_GetStoredTokenInfo(t *testing.T) {
	tempDir := t.TempDir()
	
	tests := []struct {
		name        string
		setupToken  *TokenResponse
		expectValid bool
	}{
		{
			name: "valid stored token",
			setupToken: &TokenResponse{
				AccessToken: "valid-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				Scope:       "read write",
			},
			expectValid: true,
		},
		{
			name:        "no stored token",
			setupToken:  nil,
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOAuthClient("test-client-id", "test-client-secret", "https://api.linear.app")
			client.tokenStore = NewTokenStoreWithPath(tempDir + "/test-token-" + tt.name + ".json")

			// Setup stored token if provided
			if tt.setupToken != nil {
				err := client.tokenStore.SaveToken(tt.setupToken)
				if err != nil {
					t.Fatalf("Failed to setup test token: %v", err)
				}
			}

			// Test GetStoredTokenInfo
			info := client.GetStoredTokenInfo()

			if info == nil {
				t.Fatal("Expected token info to be returned")
			}

			if tt.expectValid {
				if valid, ok := info["valid"].(bool); !ok || !valid {
					t.Error("Expected token to be reported as valid")
				}
				
				if _, ok := info["expires_at"]; !ok {
					t.Error("Expected expires_at field in token info")
				}
				
				if scope, ok := info["scope"].(string); !ok || scope != tt.setupToken.Scope {
					t.Errorf("Expected scope '%s', got '%v'", tt.setupToken.Scope, scope)
				}
			} else {
				if _, ok := info["error"]; !ok {
					t.Error("Expected error field when no token stored")
				}
			}
		})
	}
}

func TestOAuthClient_ClearStoredToken(t *testing.T) {
	tempDir := t.TempDir()
	
	client := NewOAuthClient("test-client-id", "test-client-secret", "https://api.linear.app")
	client.tokenStore = NewTokenStoreWithPath(tempDir + "/test-token.json")

	// Setup a token first
	token := &TokenResponse{
		AccessToken: "test-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		Scope:       "read write",
	}
	
	err := client.tokenStore.SaveToken(token)
	if err != nil {
		t.Fatalf("Failed to setup test token: %v", err)
	}

	// Verify token exists
	if !client.HasValidStoredToken() {
		t.Error("Expected token to exist before clearing")
	}

	// Clear token
	err = client.ClearStoredToken()
	if err != nil {
		t.Fatalf("Failed to clear token: %v", err)
	}

	// Verify token is cleared
	if client.HasValidStoredToken() {
		t.Error("Expected token to be cleared")
	}
}

func TestOAuthClient_HasValidStoredToken(t *testing.T) {
	tempDir := t.TempDir()
	
	tests := []struct {
		name        string
		setupToken  *TokenResponse
		expectValid bool
	}{
		{
			name: "valid stored token",
			setupToken: &TokenResponse{
				AccessToken: "valid-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				Scope:       "read write",
			},
			expectValid: true,
		},
		{
			name: "expired stored token",
			setupToken: &TokenResponse{
				AccessToken: "expired-token",
				TokenType:   "Bearer",
				ExpiresIn:   1, // Will be expired due to buffer
				Scope:       "read write",
			},
			expectValid: false,
		},
		{
			name:        "no stored token",
			setupToken:  nil,
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOAuthClient("test-client-id", "test-client-secret", "https://api.linear.app")
			client.tokenStore = NewTokenStoreWithPath(tempDir + "/test-token-" + tt.name + ".json")

			// Setup stored token if provided
			if tt.setupToken != nil {
				err := client.tokenStore.SaveToken(tt.setupToken)
				if err != nil {
					t.Fatalf("Failed to setup test token: %v", err)
				}
				
				if tt.setupToken.ExpiresIn == 1 {
					// Wait for token to expire
					time.Sleep(10 * time.Millisecond)
				}
			}

			// Test HasValidStoredToken
			hasValid := client.HasValidStoredToken()

			if hasValid != tt.expectValid {
				t.Errorf("Expected HasValidStoredToken to return %v, got %v", tt.expectValid, hasValid)
			}
		})
	}
}

func TestOAuthClient_NoTokenStore(t *testing.T) {
	// Test client behavior when token store is nil
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := TokenResponse{
			AccessToken: "fallback-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
			Scope:       "read write",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewOAuthClient("test-client-id", "test-client-secret", server.URL)
	client.tokenStore = nil // Simulate no token store

	// GetValidToken should fallback to direct token request
	token, err := client.GetValidToken(context.Background(), []string{"read", "write"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if token.AccessToken != "fallback-token" {
		t.Errorf("Expected fallback token, got '%s'", token.AccessToken)
	}

	// HasValidStoredToken should return false
	if client.HasValidStoredToken() {
		t.Error("Expected HasValidStoredToken to return false when no token store")
	}

	// ClearStoredToken should return error
	err = client.ClearStoredToken()
	if err == nil {
		t.Error("Expected error when clearing token with no token store")
	}

	// GetStoredTokenInfo should return error info
	info := client.GetStoredTokenInfo()
	if errorMsg, ok := info["error"].(string); !ok || errorMsg == "" {
		t.Error("Expected error message when no token store available")
	}
}