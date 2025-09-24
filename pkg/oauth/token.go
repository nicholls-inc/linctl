package oauth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TokenStore manages OAuth token persistence
type TokenStore struct {
	configPath string
}

// StoredToken represents a token with metadata for persistence
type StoredToken struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"`
	Scope       string    `json:"scope"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// NewTokenStore creates a new token store with the default config path
func NewTokenStore() (*TokenStore, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	
	configPath := filepath.Join(homeDir, ".linctl-oauth-token.json")
	return &TokenStore{configPath: configPath}, nil
}

// NewTokenStoreWithPath creates a new token store with a custom config path
func NewTokenStoreWithPath(configPath string) *TokenStore {
	return &TokenStore{configPath: configPath}
}

// SaveToken saves a token response to persistent storage
func (ts *TokenStore) SaveToken(token *TokenResponse) error {
	if token == nil {
		return fmt.Errorf("token cannot be nil")
	}

	now := time.Now()
	storedToken := StoredToken{
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
		ExpiresIn:   token.ExpiresIn,
		Scope:       token.Scope,
		ExpiresAt:   now.Add(time.Duration(token.ExpiresIn) * time.Second),
		CreatedAt:   now,
	}

	data, err := json.MarshalIndent(storedToken, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(ts.configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write with secure permissions (readable only by owner)
	if err := os.WriteFile(ts.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return nil
}

// LoadToken loads a token from persistent storage
func (ts *TokenStore) LoadToken() (*StoredToken, error) {
	data, err := os.ReadFile(ts.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no stored token found")
		}
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var token StoredToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to parse stored token: %w", err)
	}

	return &token, nil
}

// ClearToken removes the stored token
func (ts *TokenStore) ClearToken() error {
	err := os.Remove(ts.configPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear token: %w", err)
	}
	return nil
}

// IsTokenExpired checks if a token is expired or will expire soon
// Uses a 5-minute buffer to ensure token doesn't expire during use
func (ts *TokenStore) IsTokenExpired(token *StoredToken) bool {
	if token == nil {
		return true
	}
	
	// Consider token expired if it expires within 5 minutes
	buffer := 5 * time.Minute
	return time.Now().Add(buffer).After(token.ExpiresAt)
}

// IsTokenValid checks if a token exists and is not expired
func (ts *TokenStore) IsTokenValid() bool {
	token, err := ts.LoadToken()
	if err != nil {
		return false
	}
	
	return !ts.IsTokenExpired(token)
}

// GetValidToken returns a valid token if available, nil if expired or missing
func (ts *TokenStore) GetValidToken() (*StoredToken, error) {
	token, err := ts.LoadToken()
	if err != nil {
		return nil, err
	}
	
	if ts.IsTokenExpired(token) {
		return nil, fmt.Errorf("stored token is expired")
	}
	
	return token, nil
}

// ToTokenResponse converts a StoredToken back to a TokenResponse
func (st *StoredToken) ToTokenResponse() *TokenResponse {
	if st == nil {
		return nil
	}
	
	// Calculate remaining seconds until expiry
	remainingSeconds := int(time.Until(st.ExpiresAt).Seconds())
	if remainingSeconds < 0 {
		remainingSeconds = 0
	}
	
	return &TokenResponse{
		AccessToken: st.AccessToken,
		TokenType:   st.TokenType,
		ExpiresIn:   remainingSeconds,
		Scope:       st.Scope,
	}
}

// GetTokenInfo returns human-readable token information
func (st *StoredToken) GetTokenInfo() map[string]interface{} {
	if st == nil {
		return map[string]interface{}{
			"valid": false,
		}
	}
	
	now := time.Now()
	isExpired := now.After(st.ExpiresAt)
	timeUntilExpiry := st.ExpiresAt.Sub(now)
	
	info := map[string]interface{}{
		"valid":           !isExpired,
		"expires_at":      st.ExpiresAt.Format(time.RFC3339),
		"created_at":      st.CreatedAt.Format(time.RFC3339),
		"scope":           st.Scope,
		"token_type":      st.TokenType,
	}
	
	if !isExpired {
		info["expires_in_seconds"] = int(timeUntilExpiry.Seconds())
		info["expires_in_human"] = formatDuration(timeUntilExpiry)
	} else {
		info["expired_ago_seconds"] = int(-timeUntilExpiry.Seconds())
		info["expired_ago_human"] = formatDuration(-timeUntilExpiry)
	}
	
	return info
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < 0 {
		return formatDuration(-d) + " ago"
	}
	
	if d < time.Minute {
		return fmt.Sprintf("%.0f seconds", d.Seconds())
	}
	
	if d < time.Hour {
		return fmt.Sprintf("%.0f minutes", d.Minutes())
	}
	
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1f hours", d.Hours())
	}
	
	return fmt.Sprintf("%.1f days", d.Hours()/24)
}