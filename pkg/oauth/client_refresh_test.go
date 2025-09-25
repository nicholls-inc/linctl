package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestGetValidTokenWithRefresh_ConcurrentCalls verifies that concurrent calls
// to GetValidTokenWithRefresh don't cause race conditions or token corruption
func TestGetValidTokenWithRefresh_ConcurrentCalls(t *testing.T) {
	tempDir := t.TempDir()

	// Mock server that tracks request count
	var requestCount int
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		currentCount := requestCount
		mu.Unlock()

		if r.URL.Path == "/oauth/token" {
			response := TokenResponse{
				AccessToken: "fresh-token-" + string(rune(currentCount)),
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				Scope:       "read write",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else if r.URL.Path == "/graphql" {
			// Mock GraphQL validation response
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"data":{"viewer":{"id":"test"}}}`))
		}
	}))
	defer server.Close()

	// Create client with custom token store
	client := NewOAuthClient("test-client-id", "test-client-secret", server.URL)
	client.tokenStore = NewTokenStoreWithPath(filepath.Join(tempDir, "test-token.json"))

	// Test concurrent calls
	const numConcurrentCalls = 10
	var wg sync.WaitGroup
	results := make(chan *TokenResponse, numConcurrentCalls)
	errors := make(chan error, numConcurrentCalls)

	for i := 0; i < numConcurrentCalls; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			token, err := client.GetValidTokenWithRefresh(context.Background(), []string{"read", "write"})
			if err != nil {
				errors <- err
				return
			}
			results <- token
		}()
	}

	wg.Wait()
	close(results)
	close(errors)

	// Check for errors
	var errorList []error
	for err := range errors {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		t.Fatalf("Concurrent token refresh failed: %v", errorList)
	}

	// Verify we got results
	var tokens []*TokenResponse
	for token := range results {
		tokens = append(tokens, token)
	}

	if len(tokens) != numConcurrentCalls {
		t.Fatalf("Expected %d tokens, got %d", numConcurrentCalls, len(tokens))
	}

	// All tokens should be the same (cached) or at least valid
	for i, token := range tokens {
		if token.AccessToken == "" {
			t.Errorf("Token %d has empty access token", i)
		}
		if token.TokenType != "Bearer" {
			t.Errorf("Token %d has wrong type: %s", i, token.TokenType)
		}
	}

	t.Logf("✅ Concurrent token refresh test passed: %d calls succeeded", numConcurrentCalls)
}

// TestTokenRefreshRetryLogic verifies that token refresh retry logic works correctly
func TestTokenRefreshRetryLogic(t *testing.T) {
	tempDir := t.TempDir()

	// Mock server that fails first few requests then succeeds
	var requestCount int
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		currentCount := requestCount
		mu.Unlock()

		if r.URL.Path == "/oauth/token" {
			// Fail first 2 requests, succeed on 3rd
			if currentCount <= 2 {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"server_error","error_description":"Temporary server error"}`))
				return
			}

			response := TokenResponse{
				AccessToken: "retry-success-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				Scope:       "read write",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	// Create client with custom token store
	client := NewOAuthClient("test-client-id", "test-client-secret", server.URL)
	client.tokenStore = NewTokenStoreWithPath(filepath.Join(tempDir, "test-token-retry.json"))

	// Test retry logic
	token, err := client.GetValidTokenWithRefresh(context.Background(), []string{"read", "write"})
	if err != nil {
		t.Fatalf("Expected retry logic to succeed, got error: %v", err)
	}

	if token.AccessToken != "retry-success-token" {
		t.Errorf("Expected 'retry-success-token', got '%s'", token.AccessToken)
	}

	// Verify we made the expected number of requests (3: 2 failures + 1 success)
	mu.Lock()
	finalCount := requestCount
	mu.Unlock()

	if finalCount != 3 {
		t.Errorf("Expected 3 requests (2 failures + 1 success), got %d", finalCount)
	}

	t.Log("✅ Token refresh retry logic test passed")
}

// TestTokenExpiryBufferBehavior verifies the new 2-minute buffer behavior
func TestTokenExpiryBufferBehavior(t *testing.T) {
	tempDir := t.TempDir()
	tokenPath := filepath.Join(tempDir, "test-token-buffer.json")
	store := NewTokenStoreWithPath(tokenPath)

	// Create token that expires in 3 minutes
	token := &TokenResponse{
		AccessToken: "buffer-test-token",
		TokenType:   "Bearer",
		ExpiresIn:   180, // 3 minutes
		Scope:       "read write",
	}

	err := store.SaveToken(token)
	if err != nil {
		t.Fatalf("Failed to save token: %v", err)
	}

	// Test with 2-minute buffer (should be valid)
	validToken, err := store.GetValidTokenWithBuffer(2 * time.Minute)
	if err != nil {
		t.Errorf("Token should be valid with 2-minute buffer: %v", err)
	}
	if validToken == nil {
		t.Error("Expected valid token, got nil")
	}

	// Test with 4-minute buffer (should be invalid)
	_, err = store.GetValidTokenWithBuffer(4 * time.Minute)
	if err == nil {
		t.Error("Token should be invalid with 4-minute buffer")
	}

	// Test the buffer calculation
	storedToken, _ := store.LoadToken()
	if !store.IsTokenExpiredWithBuffer(storedToken, 4*time.Minute) {
		t.Error("Token should be considered expired with 4-minute buffer")
	}
	if store.IsTokenExpiredWithBuffer(storedToken, 2*time.Minute) {
		t.Error("Token should not be considered expired with 2-minute buffer")
	}

	t.Log("✅ Token expiry buffer behavior test passed")
}

// TestAuthenticationPersistenceAcrossTokenExpiry simulates the original issue:
// authentication working initially but failing after token expires
func TestAuthenticationPersistenceAcrossTokenExpiry(t *testing.T) {
	tempDir := t.TempDir()

	// Mock server that provides tokens with very short expiry
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			response := TokenResponse{
				AccessToken: "short-lived-token-" + time.Now().Format("15:04:05"),
				TokenType:   "Bearer",
				ExpiresIn:   1, // 1 second expiry for testing
				Scope:       "read write",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else if r.URL.Path == "/graphql" {
			// Mock GraphQL validation response
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"data":{"viewer":{"id":"test"}}}`))
		}
	}))
	defer server.Close()

	// Create client with custom token store
	client := NewOAuthClient("test-client-id", "test-client-secret", server.URL)
	client.tokenStore = NewTokenStoreWithPath(filepath.Join(tempDir, "test-token-expiry.json"))

	// Get initial token
	token1, err := client.GetValidTokenWithRefresh(context.Background(), []string{"read", "write"})
	if err != nil {
		t.Fatalf("Failed to get initial token: %v", err)
	}

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	// Get token again - should automatically refresh
	token2, err := client.GetValidTokenWithRefresh(context.Background(), []string{"read", "write"})
	if err != nil {
		t.Fatalf("Failed to get token after expiry: %v", err)
	}

	// Tokens should be different (new token was issued)
	if token1.AccessToken == token2.AccessToken {
		t.Error("Expected different tokens after expiry, but got the same token")
	}

	// Both tokens should be valid format
	if !strings.HasPrefix(token1.AccessToken, "short-lived-token-") {
		t.Errorf("First token has unexpected format: %s", token1.AccessToken)
	}
	if !strings.HasPrefix(token2.AccessToken, "short-lived-token-") {
		t.Errorf("Second token has unexpected format: %s", token2.AccessToken)
	}

	t.Log("✅ Authentication persistence across token expiry test passed")
}
