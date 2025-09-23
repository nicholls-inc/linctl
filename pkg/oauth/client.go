package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OAuthClient handles OAuth client credentials flow for Linear
type OAuthClient struct {
	clientID     string
	clientSecret string
	baseURL      string
	httpClient   *http.Client
}

// NewOAuthClient creates a new OAuth client for Linear
func NewOAuthClient(clientID, clientSecret, baseURL string) *OAuthClient {
	if baseURL == "" {
		baseURL = "https://api.linear.app"
	}
	
	return &OAuthClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		baseURL:      baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
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