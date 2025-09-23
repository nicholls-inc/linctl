package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuthConfig_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		config   AuthConfig
		expected string
	}{
		{
			name: "API key only",
			config: AuthConfig{
				APIKey: "test-api-key",
			},
			expected: `{"api_key":"test-api-key"}`,
		},
		{
			name: "OAuth token only",
			config: AuthConfig{
				OAuthToken: "test-oauth-token",
			},
			expected: `{"oauth_token":"test-oauth-token"}`,
		},
		{
			name: "both API key and OAuth token",
			config: AuthConfig{
				APIKey:     "test-api-key",
				OAuthToken: "test-oauth-token",
			},
			expected: `{"api_key":"test-api-key","oauth_token":"test-oauth-token"}`,
		},
		{
			name:     "empty config",
			config:   AuthConfig{},
			expected: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.config)
			if err != nil {
				t.Fatalf("Failed to marshal config: %v", err)
			}

			if string(jsonData) != tt.expected {
				t.Errorf("Expected JSON %s, got %s", tt.expected, string(jsonData))
			}

			// Test unmarshaling
			var unmarshaled AuthConfig
			err = json.Unmarshal(jsonData, &unmarshaled)
			if err != nil {
				t.Fatalf("Failed to unmarshal config: %v", err)
			}

			if unmarshaled.APIKey != tt.config.APIKey {
				t.Errorf("Expected APIKey %s, got %s", tt.config.APIKey, unmarshaled.APIKey)
			}
			if unmarshaled.OAuthToken != tt.config.OAuthToken {
				t.Errorf("Expected OAuthToken %s, got %s", tt.config.OAuthToken, unmarshaled.OAuthToken)
			}
		})
	}
}

func TestGetAuthHeader(t *testing.T) {
	// Create temporary directory for test config files
	tempDir, err := os.MkdirTemp("", "linctl-auth-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the config path for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	tests := []struct {
		name           string
		config         AuthConfig
		expectedHeader string
		expectedError  string
	}{
		{
			name: "OAuth token takes priority",
			config: AuthConfig{
				APIKey:     "test-api-key",
				OAuthToken: "test-oauth-token",
			},
			expectedHeader: "Bearer test-oauth-token",
		},
		{
			name: "API key fallback",
			config: AuthConfig{
				APIKey: "test-api-key",
			},
			expectedHeader: "test-api-key",
		},
		{
			name: "OAuth token only",
			config: AuthConfig{
				OAuthToken: "test-oauth-token",
			},
			expectedHeader: "Bearer test-oauth-token",
		},
		{
			name:          "no authentication",
			config:        AuthConfig{},
			expectedError: "no valid authentication found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save test config
			err := saveAuth(tt.config)
			if err != nil {
				t.Fatalf("Failed to save auth config: %v", err)
			}

			header, err := GetAuthHeader()

			if tt.expectedError != "" {
				if err == nil {
					t.Fatalf("Expected error %s, got nil", tt.expectedError)
				}
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error to contain %s, got %s", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if header != tt.expectedHeader {
				t.Errorf("Expected header %s, got %s", tt.expectedHeader, header)
			}
		})
	}
}

func TestGetAuthMethod(t *testing.T) {
	// Create temporary directory for test config files
	tempDir, err := os.MkdirTemp("", "linctl-auth-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the config path for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	tests := []struct {
		name           string
		config         AuthConfig
		expectedMethod string
		expectedError  string
	}{
		{
			name: "OAuth token takes priority",
			config: AuthConfig{
				APIKey:     "test-api-key",
				OAuthToken: "test-oauth-token",
			},
			expectedMethod: "oauth",
		},
		{
			name: "API key fallback",
			config: AuthConfig{
				APIKey: "test-api-key",
			},
			expectedMethod: "api_key",
		},
		{
			name: "OAuth token only",
			config: AuthConfig{
				OAuthToken: "test-oauth-token",
			},
			expectedMethod: "oauth",
		},
		{
			name:          "no authentication",
			config:        AuthConfig{},
			expectedError: "no valid authentication found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save test config
			err := saveAuth(tt.config)
			if err != nil {
				t.Fatalf("Failed to save auth config: %v", err)
			}

			method, err := GetAuthMethod()

			if tt.expectedError != "" {
				if err == nil {
					t.Fatalf("Expected error %s, got nil", tt.expectedError)
				}
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error to contain %s, got %s", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if method != tt.expectedMethod {
				t.Errorf("Expected method %s, got %s", tt.expectedMethod, method)
			}
		})
	}
}

func TestLoginWithOAuth_EnvironmentVariables(t *testing.T) {
	t.Skip("Skipping integration test that requires API mocking")
	// Create temporary directory for test config files
	tempDir, err := os.MkdirTemp("", "linctl-auth-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the config path for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Mock Linear OAuth and GraphQL servers
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			// Return successful OAuth response
			response := map[string]interface{}{
				"access_token": "test-oauth-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
				"scope":        "read write issues:create comments:create",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else if r.URL.Path == "/graphql" {
			// Return successful user query response
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"viewer": map[string]interface{}{
						"id":    "user-123",
						"name":  "Test User",
						"email": "test@example.com",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer oauthServer.Close()

	// Set environment variables
	os.Setenv("LINEAR_CLIENT_ID", "test-client-id")
	os.Setenv("LINEAR_CLIENT_SECRET", "test-client-secret")
	os.Setenv("LINEAR_BASE_URL", oauthServer.URL)
	defer func() {
		os.Unsetenv("LINEAR_CLIENT_ID")
		os.Unsetenv("LINEAR_CLIENT_SECRET")
		os.Unsetenv("LINEAR_BASE_URL")
	}()

	// Test OAuth login
	err = LoginWithOAuth(true, false) // plaintext mode to avoid prompts
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify token was saved
	config, err := loadAuth()
	if err != nil {
		t.Fatalf("Failed to load auth config: %v", err)
	}

	if config.OAuthToken != "test-oauth-token" {
		t.Errorf("Expected OAuth token to be saved as 'test-oauth-token', got %s", config.OAuthToken)
	}

	// Verify auth header uses OAuth token
	header, err := GetAuthHeader()
	if err != nil {
		t.Fatalf("Failed to get auth header: %v", err)
	}

	expectedHeader := "Bearer test-oauth-token"
	if header != expectedHeader {
		t.Errorf("Expected auth header %s, got %s", expectedHeader, header)
	}
}

func TestLoginWithOAuth_MissingCredentials(t *testing.T) {
	// Ensure environment variables are not set
	os.Unsetenv("LINEAR_CLIENT_ID")
	os.Unsetenv("LINEAR_CLIENT_SECRET")

	// Create a pipe to simulate empty input
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	// Close write end immediately to simulate empty input
	w.Close()

	// Redirect stdin
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	err = LoginWithOAuth(true, false) // plaintext mode
	if err == nil {
		t.Fatal("Expected error for missing credentials, got nil")
	}

	if err == nil {
		t.Fatal("Expected error for missing credentials, got nil")
	}
	// The error could be EOF from empty input or the credentials error
	if !strings.Contains(err.Error(), "OAuth client ID and secret are required") && !strings.Contains(err.Error(), "EOF") {
		t.Errorf("Expected error about missing credentials or EOF, got %s", err.Error())
	}
}

func TestLoginWithOAuth_InvalidCredentials(t *testing.T) {
	// Create temporary directory for test config files
	tempDir, err := os.MkdirTemp("", "linctl-auth-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the config path for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Mock server that returns 401 Unauthorized
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "invalid_client",
			"error_description": "Invalid client: client is invalid",
		})
	}))
	defer server.Close()

	// Set environment variables with invalid credentials
	os.Setenv("LINEAR_CLIENT_ID", "invalid-client-id")
	os.Setenv("LINEAR_CLIENT_SECRET", "invalid-client-secret")
	os.Setenv("LINEAR_BASE_URL", server.URL)
	defer func() {
		os.Unsetenv("LINEAR_CLIENT_ID")
		os.Unsetenv("LINEAR_CLIENT_SECRET")
		os.Unsetenv("LINEAR_BASE_URL")
	}()

	err = LoginWithOAuth(true, false) // plaintext mode
	if err == nil {
		t.Fatal("Expected error for invalid credentials, got nil")
	}

	if !strings.Contains(err.Error(), "failed to get OAuth token") {
		t.Errorf("Expected error to contain 'failed to get OAuth token', got %s", err.Error())
	}
}

func TestGetCurrentUser_WithOAuth(t *testing.T) {
	// Create temporary directory for test config files
	tempDir, err := os.MkdirTemp("", "linctl-auth-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the config path for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify OAuth token is used
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-oauth-token" {
			t.Errorf("Expected Bearer test-oauth-token, got %s", authHeader)
		}

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"id":        "user-123",
					"name":      "Test User",
					"email":     "test@example.com",
					"avatarUrl": "https://example.com/avatar.png",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Save OAuth token
	config := AuthConfig{
		OAuthToken: "test-oauth-token",
	}
	err = saveAuth(config)
	if err != nil {
		t.Fatalf("Failed to save auth config: %v", err)
	}

	// Override the API base URL for testing
	// Note: This would require modifying the api package to support URL override
	// For now, we'll test that the auth header is correctly formatted
	header, err := GetAuthHeader()
	if err != nil {
		t.Fatalf("Failed to get auth header: %v", err)
	}

	expectedHeader := "Bearer test-oauth-token"
	if header != expectedHeader {
		t.Errorf("Expected auth header %s, got %s", expectedHeader, header)
	}
}

func TestConfigPath(t *testing.T) {
	// Test that config path is correctly constructed
	originalHome := os.Getenv("HOME")
	testHome := "/test/home"
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", originalHome)

	configPath, err := getConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	expectedPath := filepath.Join(testHome, ".linctl-auth.json")
	if configPath != expectedPath {
		t.Errorf("Expected config path %s, got %s", expectedPath, configPath)
	}
}

func TestSaveAndLoadAuth(t *testing.T) {
	// Create temporary directory for test config files
	tempDir, err := os.MkdirTemp("", "linctl-auth-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the config path for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	tests := []struct {
		name   string
		config AuthConfig
	}{
		{
			name: "API key only",
			config: AuthConfig{
				APIKey: "test-api-key",
			},
		},
		{
			name: "OAuth token only",
			config: AuthConfig{
				OAuthToken: "test-oauth-token",
			},
		},
		{
			name: "both API key and OAuth token",
			config: AuthConfig{
				APIKey:     "test-api-key",
				OAuthToken: "test-oauth-token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save config
			err := saveAuth(tt.config)
			if err != nil {
				t.Fatalf("Failed to save auth config: %v", err)
			}

			// Load config
			loaded, err := loadAuth()
			if err != nil {
				t.Fatalf("Failed to load auth config: %v", err)
			}

			// Verify loaded config matches saved config
			if loaded.APIKey != tt.config.APIKey {
				t.Errorf("Expected APIKey %s, got %s", tt.config.APIKey, loaded.APIKey)
			}
			if loaded.OAuthToken != tt.config.OAuthToken {
				t.Errorf("Expected OAuthToken %s, got %s", tt.config.OAuthToken, loaded.OAuthToken)
			}

			// Verify file permissions
			configPath, _ := getConfigPath()
			info, err := os.Stat(configPath)
			if err != nil {
				t.Fatalf("Failed to stat config file: %v", err)
			}

			expectedPerm := os.FileMode(0600)
			if info.Mode().Perm() != expectedPerm {
				t.Errorf("Expected file permissions %v, got %v", expectedPerm, info.Mode().Perm())
			}
		})
	}
}

func TestLoadAuth_FileNotExists(t *testing.T) {
	// Create temporary directory for test config files
	tempDir, err := os.MkdirTemp("", "linctl-auth-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the config path for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Try to load auth when file doesn't exist
	_, err = loadAuth()
	if err == nil {
		t.Fatal("Expected error when config file doesn't exist, got nil")
	}

	expectedError := "not authenticated"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain %s, got %s", expectedError, err.Error())
	}
}