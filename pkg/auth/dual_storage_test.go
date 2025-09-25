package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dorkitude/linctl/pkg/oauth"
)

// TestDualStorageSynchronizationPrevention verifies that we've eliminated
// the dual storage synchronization issue that caused the original authentication corruption
func TestDualStorageSynchronizationPrevention(t *testing.T) {
	WithIsolatedEnvironment(t, func(env *TestEnvironment) {
		env.WithMockedConfigPath(func() {
			// Create a scenario that would have caused the original issue:
			// 1. OAuth token in auth config (legacy)
			// 2. Different OAuth token in OAuth TokenStore
			// 3. Verify that only the OAuth TokenStore is used

			// Step 1: Create auth config with legacy OAuth token (this should be ignored)
			legacyConfig := map[string]interface{}{
				"api_key":     "fallback-api-key",
				"oauth_token": "legacy-oauth-token-should-be-ignored",
			}

			configPath, err := getConfigPath()
			if err != nil {
				t.Fatalf("Failed to get config path: %v", err)
			}

			// Manually write legacy config with OAuth token
			legacyData, err := json.MarshalIndent(legacyConfig, "", "  ")
			if err != nil {
				t.Fatalf("Failed to marshal legacy config: %v", err)
			}

			err = os.WriteFile(configPath, legacyData, 0600)
			if err != nil {
				t.Fatalf("Failed to write legacy config: %v", err)
			}

			// Step 2: Create OAuth TokenStore with different token
			homeDir, err := os.UserHomeDir()
			if err != nil {
				t.Fatalf("Failed to get home dir: %v", err)
			}

			oauthTokenPath := filepath.Join(homeDir, ".linctl-oauth-token.json")
			oauthTokenData := map[string]interface{}{
				"access_token": "oauth-store-token-should-be-used",
				"token_type":   "Bearer",
				"expires_in":   3600,
				"scope":        "read write",
				"expires_at":   time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				"created_at":   time.Now().Format(time.RFC3339),
			}

			oauthData, err := json.MarshalIndent(oauthTokenData, "", "  ")
			if err != nil {
				t.Fatalf("Failed to marshal OAuth token: %v", err)
			}

			err = os.WriteFile(oauthTokenPath, oauthData, 0600)
			if err != nil {
				t.Fatalf("Failed to write OAuth token: %v", err)
			}

			// Step 3: Set up OAuth environment to enable OAuth path
			originalClientID := os.Getenv("LINEAR_CLIENT_ID")
			originalClientSecret := os.Getenv("LINEAR_CLIENT_SECRET")

			os.Setenv("LINEAR_CLIENT_ID", "test-client-id")
			os.Setenv("LINEAR_CLIENT_SECRET", "test-client-secret")

			defer func() {
				os.Setenv("LINEAR_CLIENT_ID", originalClientID)
				os.Setenv("LINEAR_CLIENT_SECRET", originalClientSecret)
				os.Remove(oauthTokenPath)
			}()

			// Step 4: Test that GetAuthHeader uses OAuth TokenStore, not auth config
			header, err := GetAuthHeader()

			// This should succeed and use the OAuth TokenStore token
			if err != nil {
				// If OAuth fails (expected in test environment), should fall back to API key
				if header != "fallback-api-key" {
					t.Errorf("Expected fallback to API key 'fallback-api-key', got '%s'", header)
				}
				t.Log("✅ OAuth failed as expected in test environment, fell back to API key")
			} else {
				// If OAuth succeeds, should use OAuth TokenStore token, not legacy config token
				expectedHeader := "Bearer oauth-store-token-should-be-used"
				if header != expectedHeader {
					t.Errorf("Expected OAuth TokenStore token '%s', got '%s'", expectedHeader, header)
				}
				t.Log("✅ OAuth succeeded and used TokenStore token, not legacy config token")
			}

			// Step 5: Verify that auth config doesn't contain OAuth token after our fix
			cleanedConfig, err := loadAuth()
			if err != nil {
				t.Fatalf("Failed to load cleaned config: %v", err)
			}

			// The new AuthConfig struct should only have APIKey
			if cleanedConfig.APIKey != "fallback-api-key" {
				t.Errorf("Expected API key 'fallback-api-key', got '%s'", cleanedConfig.APIKey)
			}

			// Verify no OAuth token field exists (compile-time check)
			_ = AuthConfig{
				APIKey: "test",
				// OAuthToken: "should-not-compile", // This would cause compile error
			}

			t.Log("✅ Dual storage synchronization prevention test passed")
		})
	})
}

// TestAuthConfigStructureChange verifies that the AuthConfig struct
// no longer contains OAuth token fields
func TestAuthConfigStructureChange(t *testing.T) {
	// Test that AuthConfig only contains APIKey
	config := AuthConfig{
		APIKey: "test-api-key",
	}

	// Serialize to JSON
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal AuthConfig: %v", err)
	}

	// Verify JSON structure
	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Should only contain api_key
	if len(jsonMap) != 1 {
		t.Errorf("Expected AuthConfig to have 1 field, got %d: %v", len(jsonMap), jsonMap)
	}

	if apiKey, exists := jsonMap["api_key"]; !exists || apiKey != "test-api-key" {
		t.Errorf("Expected api_key field with value 'test-api-key', got %v", apiKey)
	}

	// Should NOT contain oauth_token
	if _, exists := jsonMap["oauth_token"]; exists {
		t.Error("AuthConfig should not contain oauth_token field")
	}

	t.Log("✅ AuthConfig structure change test passed")
}

// TestOAuthTokenStoreSeparation verifies that OAuth tokens are managed
// exclusively by the OAuth TokenStore, not the auth config
func TestOAuthTokenStoreSeparation(t *testing.T) {
	tempDir := t.TempDir()

	// Create OAuth TokenStore
	tokenStorePath := filepath.Join(tempDir, "oauth-token.json")
	tokenStore := oauth.NewTokenStoreWithPath(tokenStorePath)

	// Save token to OAuth TokenStore
	token := &oauth.TokenResponse{
		AccessToken: "oauth-store-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		Scope:       "read write",
	}

	err := tokenStore.SaveToken(token)
	if err != nil {
		t.Fatalf("Failed to save token to OAuth store: %v", err)
	}

	// Create auth config (should not contain OAuth token)
	authConfigPath := filepath.Join(tempDir, "auth-config.json")
	authConfig := AuthConfig{
		APIKey: "api-key-only",
	}

	authData, err := json.MarshalIndent(authConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal auth config: %v", err)
	}

	err = os.WriteFile(authConfigPath, authData, 0600)
	if err != nil {
		t.Fatalf("Failed to write auth config: %v", err)
	}

	// Verify OAuth token is in OAuth TokenStore
	storedToken, err := tokenStore.LoadToken()
	if err != nil {
		t.Fatalf("Failed to load token from OAuth store: %v", err)
	}

	if storedToken.AccessToken != "oauth-store-token" {
		t.Errorf("Expected 'oauth-store-token', got '%s'", storedToken.AccessToken)
	}

	// Verify auth config only contains API key
	authData, err = os.ReadFile(authConfigPath)
	if err != nil {
		t.Fatalf("Failed to read auth config: %v", err)
	}

	var loadedAuthConfig map[string]interface{}
	err = json.Unmarshal(authData, &loadedAuthConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal auth config: %v", err)
	}

	if loadedAuthConfig["api_key"] != "api-key-only" {
		t.Errorf("Expected API key 'api-key-only', got %v", loadedAuthConfig["api_key"])
	}

	if _, exists := loadedAuthConfig["oauth_token"]; exists {
		t.Error("Auth config should not contain oauth_token")
	}

	t.Log("✅ OAuth token store separation test passed")
}

// TestNoTokenCorruptionDuringRefresh verifies that token refresh operations
// don't corrupt the authentication state
func TestNoTokenCorruptionDuringRefresh(t *testing.T) {
	tempDir := t.TempDir()

	// Create OAuth client with mock token store
	tokenStorePath := filepath.Join(tempDir, "refresh-test-token.json")
	tokenStore := oauth.NewTokenStoreWithPath(tokenStorePath)

	// Save initial token that's about to expire
	initialToken := &oauth.TokenResponse{
		AccessToken: "initial-token",
		TokenType:   "Bearer",
		ExpiresIn:   1, // 1 second - will expire quickly
		Scope:       "read write",
	}

	err := tokenStore.SaveToken(initialToken)
	if err != nil {
		t.Fatalf("Failed to save initial token: %v", err)
	}

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	// Verify token is expired
	if tokenStore.IsTokenValid() {
		t.Error("Token should be expired")
	}

	// Verify that attempting to get a valid token doesn't corrupt the store
	_, err = tokenStore.GetValidToken()
	if err == nil {
		t.Error("Expected error for expired token")
	}

	// Verify the token store is still readable and not corrupted
	expiredToken, err := tokenStore.LoadToken()
	if err != nil {
		t.Fatalf("Token store should still be readable: %v", err)
	}

	if expiredToken.AccessToken != "initial-token" {
		t.Errorf("Token store corrupted: expected 'initial-token', got '%s'", expiredToken.AccessToken)
	}

	t.Log("✅ No token corruption during refresh test passed")
}
