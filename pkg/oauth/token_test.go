package oauth

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTokenStore(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "linctl-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create token store with custom path
	tokenPath := filepath.Join(tempDir, "test-token.json")
	store := NewTokenStoreWithPath(tokenPath)

	// Test saving and loading token
	originalToken := &TokenResponse{
		AccessToken: "test-access-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		Scope:       "read write",
	}

	// Save token
	err = store.SaveToken(originalToken)
	if err != nil {
		t.Fatalf("Failed to save token: %v", err)
	}

	// Load token
	storedToken, err := store.LoadToken()
	if err != nil {
		t.Fatalf("Failed to load token: %v", err)
	}

	// Verify token data
	if storedToken.AccessToken != originalToken.AccessToken {
		t.Errorf("Expected access token %s, got %s", originalToken.AccessToken, storedToken.AccessToken)
	}

	if storedToken.TokenType != originalToken.TokenType {
		t.Errorf("Expected token type %s, got %s", originalToken.TokenType, storedToken.TokenType)
	}

	if storedToken.Scope != originalToken.Scope {
		t.Errorf("Expected scope %s, got %s", originalToken.Scope, storedToken.Scope)
	}

	// Test token expiration
	if store.IsTokenExpired(storedToken) {
		t.Error("Token should not be expired immediately after creation")
	}

	// Test valid token retrieval
	validToken, err := store.GetValidToken()
	if err != nil {
		t.Fatalf("Failed to get valid token: %v", err)
	}

	if validToken.AccessToken != originalToken.AccessToken {
		t.Errorf("Valid token access token mismatch")
	}

	// Test token conversion
	tokenResp := storedToken.ToTokenResponse()
	if tokenResp.AccessToken != originalToken.AccessToken {
		t.Errorf("Token conversion failed")
	}

	// Test clearing token
	err = store.ClearToken()
	if err != nil {
		t.Fatalf("Failed to clear token: %v", err)
	}

	// Verify token is cleared
	_, err = store.LoadToken()
	if err == nil {
		t.Error("Expected error when loading cleared token")
	}
}

func TestTokenExpiration(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "linctl-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tokenPath := filepath.Join(tempDir, "test-token.json")
	store := NewTokenStoreWithPath(tokenPath)

	// Create expired token
	expiredToken := &StoredToken{
		AccessToken: "expired-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		Scope:       "read write",
		ExpiresAt:   time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		CreatedAt:   time.Now().Add(-2 * time.Hour),
	}

	// Test expiration check
	if !store.IsTokenExpired(expiredToken) {
		t.Error("Token should be detected as expired")
	}

	// Create token that expires soon (within buffer)
	soonExpiredToken := &StoredToken{
		AccessToken: "soon-expired-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		Scope:       "read write",
		ExpiresAt:   time.Now().Add(2 * time.Minute), // Expires in 2 minutes (within 5-minute buffer)
		CreatedAt:   time.Now().Add(-58 * time.Minute),
	}

	// Should be considered expired due to buffer
	if !store.IsTokenExpired(soonExpiredToken) {
		t.Error("Token should be detected as expired due to buffer time")
	}

	// Create valid token
	validToken := &StoredToken{
		AccessToken: "valid-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		Scope:       "read write",
		ExpiresAt:   time.Now().Add(30 * time.Minute), // Expires in 30 minutes
		CreatedAt:   time.Now().Add(-30 * time.Minute),
	}

	// Should not be expired
	if store.IsTokenExpired(validToken) {
		t.Error("Token should not be detected as expired")
	}
}

func TestTokenInfo(t *testing.T) {
	// Test valid token info
	validToken := &StoredToken{
		AccessToken: "valid-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		Scope:       "read write",
		ExpiresAt:   time.Now().Add(30 * time.Minute),
		CreatedAt:   time.Now().Add(-30 * time.Minute),
	}

	info := validToken.GetTokenInfo()
	
	if !info["valid"].(bool) {
		t.Error("Token should be reported as valid")
	}

	if _, ok := info["expires_in_seconds"]; !ok {
		t.Error("Should include expires_in_seconds for valid token")
	}

	if _, ok := info["expires_in_human"]; !ok {
		t.Error("Should include expires_in_human for valid token")
	}

	// Test expired token info
	expiredToken := &StoredToken{
		AccessToken: "expired-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		Scope:       "read write",
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
		CreatedAt:   time.Now().Add(-2 * time.Hour),
	}

	expiredInfo := expiredToken.GetTokenInfo()
	
	if expiredInfo["valid"].(bool) {
		t.Error("Expired token should be reported as invalid")
	}

	if _, ok := expiredInfo["expired_ago_seconds"]; !ok {
		t.Error("Should include expired_ago_seconds for expired token")
	}

	// Test nil token info
	nilInfo := (*StoredToken)(nil).GetTokenInfo()
	if nilInfo["valid"].(bool) {
		t.Error("Nil token should be reported as invalid")
	}
}

func TestTokenStore_GetValidTokenWithBuffer(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "linctl-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create token store with custom path
	tokenPath := filepath.Join(tempDir, "test-token-buffer.json")
	store := NewTokenStoreWithPath(tokenPath)

	// Test with token that expires in 3 minutes (should be valid with 2-minute buffer)
	validToken := &TokenResponse{
		AccessToken: "valid-token",
		TokenType:   "Bearer",
		ExpiresIn:   180, // 3 minutes
		Scope:       "read write",
	}

	err = store.SaveToken(validToken)
	if err != nil {
		t.Fatalf("Failed to save token: %v", err)
	}

	// Should be valid with 2-minute buffer
	token, err := store.GetValidTokenWithBuffer(2 * time.Minute)
	if err != nil {
		t.Errorf("Token should be valid with 2-minute buffer: %v", err)
	}
	if token == nil {
		t.Error("Expected valid token, got nil")
	}

	// Should be invalid with 4-minute buffer
	_, err = store.GetValidTokenWithBuffer(4 * time.Minute)
	if err == nil {
		t.Error("Token should be invalid with 4-minute buffer")
	}

	// Test IsTokenExpiredWithBuffer
	storedToken, _ := store.LoadToken()
	if store.IsTokenExpiredWithBuffer(storedToken, 2*time.Minute) {
		t.Error("Token should not be expired with 2-minute buffer")
	}
	if !store.IsTokenExpiredWithBuffer(storedToken, 4*time.Minute) {
		t.Error("Token should be expired with 4-minute buffer")
	}
}