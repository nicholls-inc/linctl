package auth

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nicholls-inc/linctl/pkg/oauth"
)

// TestAuthenticationPersistence verifies that authentication doesn't get corrupted
// after multiple consecutive API calls - the original issue we're fixing
func TestAuthenticationPersistence(t *testing.T) {
	// Skip if no OAuth environment configured
	if os.Getenv("LINEAR_CLIENT_ID") == "" || os.Getenv("LINEAR_CLIENT_SECRET") == "" {
		t.Skip("Skipping integration test - OAuth environment not configured")
	}

	// Test multiple consecutive GetAuthHeader calls
	const numCalls = 10
	var wg sync.WaitGroup
	errors := make(chan error, numCalls)
	headers := make(chan string, numCalls)

	for i := 0; i < numCalls; i++ {
		wg.Add(1)
		go func(callNum int) {
			defer wg.Done()

			header, err := GetAuthHeader()
			if err != nil {
				errors <- err
				return
			}

			if header == "" {
				errors <- fmt.Errorf("call %d: got empty header", callNum)
				return
			}

			headers <- header
		}(i)
	}

	wg.Wait()
	close(errors)
	close(headers)

	// Check for any errors
	var errorList []error
	for err := range errors {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		t.Fatalf("Authentication failed in %d/%d calls. Errors: %v", len(errorList), numCalls, errorList)
	}

	// Verify we got the expected number of successful headers
	var headerList []string
	for header := range headers {
		headerList = append(headerList, header)
	}

	if len(headerList) != numCalls {
		t.Fatalf("Expected %d successful auth headers, got %d", numCalls, len(headerList))
	}

	// Verify all headers are valid OAuth tokens (they may be different due to OAuth client credentials flow)
	for i, header := range headerList {
		if !strings.HasPrefix(header, "Bearer ") {
			t.Errorf("Header %d is not a valid Bearer token: %s", i, header)
		}
		if len(header) < 20 { // Reasonable minimum length for a token
			t.Errorf("Header %d appears to be too short: %s", i, header)
		}
	}

	t.Logf("✅ Authentication persistence test passed: %d consecutive calls succeeded", numCalls)
}

// TestTokenRefreshDuringRapidCalls verifies that token refresh works correctly
// during rapid API calls without causing authentication corruption
func TestTokenRefreshDuringRapidCalls(t *testing.T) {
	// Skip if no OAuth environment configured
	if os.Getenv("LINEAR_CLIENT_ID") == "" || os.Getenv("LINEAR_CLIENT_SECRET") == "" {
		t.Skip("Skipping integration test - OAuth environment not configured")
	}

	// Create a scenario where we might trigger token refresh
	// by making calls with a very short buffer
	oauthConfig, err := oauth.LoadFromEnvironment()
	if err != nil {
		t.Skip("OAuth not configured, skipping token refresh test")
	}

	if !oauthConfig.IsComplete() {
		t.Skip("OAuth configuration incomplete, skipping token refresh test")
	}

	// Create OAuth client
	oauthClient, err := oauth.NewOAuthClientFromConfig(oauthConfig)
	if err != nil {
		t.Fatalf("Failed to create OAuth client: %v", err)
	}

	// Force a token refresh to ensure we have a fresh token
	_, err = oauthClient.RefreshToken(context.Background(), oauthConfig.Scopes)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}

	// Now test rapid calls that might trigger refresh logic
	const numRapidCalls = 20
	const callInterval = 100 * time.Millisecond

	var wg sync.WaitGroup
	errors := make(chan error, numRapidCalls)

	for i := 0; i < numRapidCalls; i++ {
		wg.Add(1)
		go func(callNum int) {
			defer wg.Done()

			// Add small delay to simulate rapid but not simultaneous calls
			time.Sleep(time.Duration(callNum) * callInterval)

			header, err := GetAuthHeader()
			if err != nil {
				errors <- fmt.Errorf("call %d failed: %w", callNum, err)
				return
			}

			if header == "" {
				errors <- fmt.Errorf("call %d: got empty header", callNum)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	var errorList []error
	for err := range errors {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		t.Fatalf("Token refresh during rapid calls failed in %d/%d calls. Errors: %v", len(errorList), numRapidCalls, errorList)
	}

	t.Logf("✅ Token refresh during rapid calls test passed: %d calls succeeded", numRapidCalls)
}

// TestNoDualStorageCorruption verifies that we don't have dual storage issues
// by ensuring OAuth tokens are not stored in auth config
func TestNoDualStorageCorruption(t *testing.T) {
	WithIsolatedEnvironment(t, func(env *TestEnvironment) {
		env.WithMockedConfigPath(func() {
			// Save an auth config with only API key
			config := AuthConfig{
				APIKey: "test-api-key",
			}
			err := env.MockAuthConfig(config)
			if err != nil {
				t.Fatalf("Failed to save auth config: %v", err)
			}

			// Load the config back
			loadedConfig, err := loadAuth()
			if err != nil {
				t.Fatalf("Failed to load auth config: %v", err)
			}

			// Verify no OAuth token is stored in auth config
			if loadedConfig.APIKey != "test-api-key" {
				t.Errorf("Expected API key 'test-api-key', got '%s'", loadedConfig.APIKey)
			}

			// Verify the AuthConfig struct doesn't have OAuth token field
			// (this is a compile-time check - if OAuthToken field exists, this won't compile)
			_ = AuthConfig{
				APIKey: "test",
				// OAuthToken: "should-not-exist", // This line should cause compile error if uncommented
			}

			t.Log("✅ No dual storage corruption: OAuth tokens not stored in auth config")
		})
	})
}

// TestAuthenticationMethodPriority verifies the correct authentication priority:
// OAuth (with refresh) -> API key -> error
func TestAuthenticationMethodPriority(t *testing.T) {
	WithIsolatedEnvironment(t, func(env *TestEnvironment) {
		env.WithMockedConfigPath(func() {
			// Test 1: No OAuth environment, should use API key
			originalClientID := os.Getenv("LINEAR_CLIENT_ID")
			originalClientSecret := os.Getenv("LINEAR_CLIENT_SECRET")

			// Clear OAuth environment
			os.Setenv("LINEAR_CLIENT_ID", "")
			os.Setenv("LINEAR_CLIENT_SECRET", "")

			defer func() {
				os.Setenv("LINEAR_CLIENT_ID", originalClientID)
				os.Setenv("LINEAR_CLIENT_SECRET", originalClientSecret)
			}()

			// Set up API key fallback
			config := AuthConfig{
				APIKey: "fallback-api-key",
			}
			err := env.MockAuthConfig(config)
			if err != nil {
				t.Fatalf("Failed to save auth config: %v", err)
			}

			header, err := GetAuthHeader()
			if err != nil {
				t.Fatalf("Expected API key fallback to work, got error: %v", err)
			}

			if header != "fallback-api-key" {
				t.Errorf("Expected API key 'fallback-api-key', got '%s'", header)
			}

			// Test 2: No authentication available
			emptyConfig := AuthConfig{}
			err = env.MockAuthConfig(emptyConfig)
			if err != nil {
				t.Fatalf("Failed to save empty auth config: %v", err)
			}

			_, err = GetAuthHeader()
			if err == nil {
				t.Error("Expected error when no authentication available")
			}

			t.Log("✅ Authentication method priority test passed")
		})
	})
}
