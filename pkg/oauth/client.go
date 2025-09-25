package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// logDebug logs debug messages if LINCTL_DEBUG environment variable is set
func logDebug(format string, args ...interface{}) {
	if os.Getenv("LINCTL_DEBUG") != "" {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// OAuthClient handles OAuth client credentials flow for Linear
type OAuthClient struct {
	clientID     string
	clientSecret string
	baseURL      string
	httpClient   *http.Client
	tokenStore   *TokenStore
	config       *Config
}

// NewOAuthClient creates a new OAuth client for Linear
func NewOAuthClient(clientID, clientSecret, baseURL string) *OAuthClient {
	if baseURL == "" {
		baseURL = "https://api.linear.app"
	}

	tokenStore, _ := NewTokenStore() // Ignore error, will handle gracefully

	return &OAuthClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		baseURL:      baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		tokenStore: tokenStore,
	}
}

// NewOAuthClientFromConfig creates a new OAuth client from configuration
func NewOAuthClientFromConfig(config *Config) (*OAuthClient, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid OAuth config: %w", err)
	}

	tokenStore, err := NewTokenStore()
	if err != nil {
		return nil, fmt.Errorf("failed to create token store: %w", err)
	}

	return &OAuthClient{
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		baseURL:      config.BaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		tokenStore: tokenStore,
		config:     config,
	}, nil
}

// GetAccessToken implements OAuth client credentials flow
// This is used for server-to-server authentication with Linear
func (c *OAuthClient) GetAccessToken(ctx context.Context, scopes []string) (*TokenResponse, error) {
	tokenURL := c.baseURL + "/oauth/token"
	scopeString := strings.Join(scopes, " ")

	// Prepare form data for client credentials flow
	data := url.Values{
		"grant_type": {"client_credentials"},
		"scope":      {scopeString},
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.clientID, c.clientSecret)

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request access token: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			if errorDesc, ok := errorResp["error_description"].(string); ok {
				return nil, fmt.Errorf("OAuth request failed (%d): %s", resp.StatusCode, errorDesc)
			}
			if errorType, ok := errorResp["error"].(string); ok {
				return nil, fmt.Errorf("OAuth request failed (%d): %s", resp.StatusCode, errorType)
			}
		}
		return nil, fmt.Errorf("OAuth request failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	// Validate response
	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("received empty access token")
	}

	if tokenResp.TokenType == "" {
		tokenResp.TokenType = "Bearer"
	}

	return &tokenResp, nil
}

// ValidateToken validates an access token by making a simple API call
func (c *OAuthClient) ValidateToken(ctx context.Context, accessToken string) error {
	// Make a simple GraphQL query to validate the token
	query := `query { viewer { id name } }`
	payload := map[string]interface{}{
		"query": query,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal validation query: %w", err)
	}

	// Create request
	graphqlURL := c.baseURL + "/graphql"
	req, err := http.NewRequestWithContext(ctx, "POST", graphqlURL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("access token is invalid or expired")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token validation failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetValidToken returns a valid access token, refreshing if necessary
func (c *OAuthClient) GetValidToken(ctx context.Context, scopes []string) (*TokenResponse, error) {
	if c.tokenStore == nil {
		// Fallback to direct token request if no token store
		return c.GetAccessToken(ctx, scopes)
	}

	// Try to load existing valid token
	storedToken, err := c.tokenStore.GetValidToken()
	if err == nil && storedToken != nil {
		// Token is valid, return it
		return storedToken.ToTokenResponse(), nil
	}

	// Token is missing or expired, get a new one
	newToken, err := c.GetAccessToken(ctx, scopes)
	if err != nil {
		return nil, fmt.Errorf("failed to get new access token: %w", err)
	}

	// Save the new token
	if saveErr := c.tokenStore.SaveToken(newToken); saveErr != nil {
		// Log the error but don't fail the request
		// The token is still valid for immediate use
		logDebug("Warning: failed to save OAuth token to store: %v", saveErr)
	}

	return newToken, nil
}

// GetValidTokenWithRefresh returns a valid access token with enhanced refresh logic and retry
func (c *OAuthClient) GetValidTokenWithRefresh(ctx context.Context, scopes []string) (*TokenResponse, error) {
	if c.tokenStore == nil {
		// Fallback to direct token request if no token store
		return c.GetAccessToken(ctx, scopes)
	}

	// Try to load existing valid token with reduced buffer (2 minutes instead of 5)
	storedToken, err := c.tokenStore.GetValidTokenWithBuffer(2 * time.Minute)
	if err == nil && storedToken != nil {
		// Token is valid with buffer, return it
		return storedToken.ToTokenResponse(), nil
	}

	// Token is missing, expired, or will expire soon - get a new one with retry logic
	const maxRetries = 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		newToken, err := c.GetAccessToken(ctx, scopes)
		if err == nil {
			// Successfully got new token, save it
			if saveErr := c.tokenStore.SaveToken(newToken); saveErr != nil {
				// Log the error but don't fail the request
				logDebug("Warning: failed to save OAuth token on attempt %d: %v", attempt, saveErr)
			}
			return newToken, nil
		}

		lastErr = err
		if attempt < maxRetries {
			// Wait before retry with exponential backoff
			waitTime := time.Duration(attempt) * time.Second
			time.Sleep(waitTime)
		}
	}

	return nil, fmt.Errorf("failed to get new access token after %d attempts: %w", maxRetries, lastErr)
}

// RefreshToken forces a token refresh and saves the new token with retry logic
func (c *OAuthClient) RefreshToken(ctx context.Context, scopes []string) (*TokenResponse, error) {
	const maxRetries = 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Get a fresh token
		newToken, err := c.GetAccessToken(ctx, scopes)
		if err == nil {
			// Successfully got new token, save it if we have a token store
			if c.tokenStore != nil {
				if saveErr := c.tokenStore.SaveToken(newToken); saveErr != nil {
					// Log warning but don't fail - token is still valid for immediate use
					logDebug("Warning: failed to save refreshed token on attempt %d: %v", attempt, saveErr)
				}
			}
			return newToken, nil
		}

		lastErr = err
		if attempt < maxRetries {
			// Wait before retry with exponential backoff
			waitTime := time.Duration(attempt) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(waitTime):
				// Continue to next attempt
			}
		}
	}

	return nil, fmt.Errorf("failed to refresh token after %d attempts: %w", maxRetries, lastErr)
}

// GetStoredTokenInfo returns information about the currently stored token
func (c *OAuthClient) GetStoredTokenInfo() map[string]interface{} {
	if c.tokenStore == nil {
		return map[string]interface{}{
			"error": "no token store available",
		}
	}

	storedToken, err := c.tokenStore.LoadToken()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
			"valid": false,
		}
	}

	return storedToken.GetTokenInfo()
}

// ClearStoredToken removes any stored token
func (c *OAuthClient) ClearStoredToken() error {
	if c.tokenStore == nil {
		return fmt.Errorf("no token store available")
	}

	return c.tokenStore.ClearToken()
}

// HasValidStoredToken checks if there's a valid token stored
func (c *OAuthClient) HasValidStoredToken() bool {
	if c.tokenStore == nil {
		return false
	}

	return c.tokenStore.IsTokenValid()
}

// ValidateAndRefreshToken validates a token and refreshes if invalid
func (c *OAuthClient) ValidateAndRefreshToken(ctx context.Context, scopes []string) (*TokenResponse, error) {
	// First try to get a valid token (may use cached)
	token, err := c.GetValidTokenWithRefresh(ctx, scopes)
	if err != nil {
		return nil, err
	}

	// Validate the token by making a test API call
	if err := c.ValidateToken(ctx, token.AccessToken); err != nil {
		// Token validation failed, force refresh
		return c.RefreshToken(ctx, scopes)
	}

	return token, nil
}

// IsTokenError checks if an error indicates token-related issues
func IsTokenError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "authentication") ||
		strings.Contains(errStr, "token")
}
