package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dorkitude/linctl/pkg/logging"
	"github.com/dorkitude/linctl/pkg/ratelimit"
	"github.com/dorkitude/linctl/pkg/resilience"
)

// EnhancedClient is a production-ready API client with retry logic and rate limiting
type EnhancedClient struct {
	baseClient  *Client
	retryClient *resilience.RetryableClient
	rateLimiter *ratelimit.RateLimiter
	logger      logging.Logger
	requestID   string
	metrics     *ClientMetrics
}

// ClientMetrics tracks client performance metrics
type ClientMetrics struct {
	RequestCount    int64         `json:"request_count"`
	ErrorCount      int64         `json:"error_count"`
	RateLimitHits   int64         `json:"rate_limit_hits"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
}

// EnhancedClientConfig configures the enhanced client
type EnhancedClientConfig struct {
	RetryConfig     resilience.RetryConfig    `json:"retry_config"`
	RateLimitConfig ratelimit.RateLimitConfig `json:"rate_limit_config"`
	Logger          logging.Logger            `json:"-"`
	BaseURL         string                    `json:"base_url"`
	Timeout         time.Duration             `json:"timeout"`
}

// DefaultEnhancedClientConfig returns a production-ready configuration
func DefaultEnhancedClientConfig() EnhancedClientConfig {
	return EnhancedClientConfig{
		RetryConfig:     resilience.DefaultRetryConfig(),
		RateLimitConfig: ratelimit.DefaultRateLimitConfig(),
		Logger:          logging.NewLogger(),
		BaseURL:         BaseURL,
		Timeout:         30 * time.Second,
	}
}

// NewEnhancedClient creates a new enhanced API client
func NewEnhancedClient(authHeader string, config EnhancedClientConfig) *EnhancedClient {
	if config.Logger == nil {
		config.Logger = logging.NewLogger()
	}

	// Create base HTTP client
	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	// Create retryable client
	retryClient := resilience.NewRetryableClient(httpClient, config.RetryConfig, config.Logger)

	// Create rate limiter
	rateLimiter := ratelimit.NewRateLimiter(config.RateLimitConfig, config.Logger)

	// Create base client
	baseClient := NewClientWithURL(config.BaseURL, authHeader)

	return &EnhancedClient{
		baseClient:  baseClient,
		retryClient: retryClient,
		rateLimiter: rateLimiter,
		logger:      config.Logger,
		requestID:   generateRequestID(),
		metrics:     &ClientMetrics{},
	}
}

// Execute performs a GraphQL request with retry logic and rate limiting
func (c *EnhancedClient) Execute(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	start := time.Now()

	// Generate request ID for tracing
	requestID := generateRequestID()
	logger := c.logger.With(logging.String("request_id", requestID))

	logger.Debug("Starting GraphQL request",
		logging.String("query_type", extractQueryType(query)),
	)

	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		c.recordError()
		logger.Error("Rate limiter wait failed", logging.Error(err))
		return fmt.Errorf("rate limit error: %w", err)
	}

	// Prepare request
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		c.recordError()
		logger.Error("Failed to marshal request", logging.Error(err))
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseClient.baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		c.recordError()
		logger.Error("Failed to create request", logging.Error(err))
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.baseClient.authHeader)
	req.Header.Set("User-Agent", "linctl/1.0.0")
	req.Header.Set("X-Request-ID", requestID)

	// Execute with retry logic
	resp, err := c.retryClient.DoWithRetry(ctx, req)
	if err != nil {
		c.recordError()
		duration := time.Since(start)
		logger.Error("Request failed after retries",
			logging.Error(err),
			logging.Duration("total_duration", duration),
		)
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Update rate limiter with response headers
	c.rateLimiter.UpdateFromResponse(resp)

	// Handle rate limiting
	if resp.StatusCode == http.StatusTooManyRequests {
		c.recordRateLimit()
		delay := c.rateLimiter.HandleRateLimitResponse(resp)

		logger.Warn("Rate limited by server",
			logging.Int("status_code", resp.StatusCode),
			logging.Duration("retry_delay", delay),
		)

		// Wait for the specified delay
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Retry the request
			return c.Execute(ctx, query, variables, result)
		}
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.recordError()
		logger.Error("Failed to read response", logging.Error(err))
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		c.recordError()
		logger.Error("API request failed",
			logging.Int("status_code", resp.StatusCode),
			logging.String("response_body", string(body)),
		)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse GraphQL response
	var gqlResp GraphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		c.recordError()
		logger.Error("Failed to parse response", logging.Error(err))
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for GraphQL errors
	if len(gqlResp.Errors) > 0 {
		c.recordError()
		logger.Error("GraphQL errors in response",
			logging.Int("error_count", len(gqlResp.Errors)),
		)

		for i, gqlErr := range gqlResp.Errors {
			logger.Error("GraphQL error",
				logging.Int("error_index", i),
				logging.String("message", gqlErr.Message),
			)
		}

		return fmt.Errorf("GraphQL errors: %v", gqlResp.Errors)
	}

	// Unmarshal result
	if result != nil {
		if err := json.Unmarshal(gqlResp.Data, result); err != nil {
			c.recordError()
			logger.Error("Failed to unmarshal data", logging.Error(err))
			return fmt.Errorf("failed to unmarshal data: %w", err)
		}
	}

	// Record successful request
	duration := time.Since(start)
	c.recordSuccess(duration)

	logger.Debug("GraphQL request completed successfully",
		logging.Duration("duration", duration),
		logging.Int("status_code", resp.StatusCode),
	)

	return nil
}

// GetMetrics returns current client metrics
func (c *EnhancedClient) GetMetrics() ClientMetrics {
	metrics := *c.metrics
	if metrics.RequestCount > 0 {
		metrics.AverageDuration = time.Duration(int64(metrics.TotalDuration) / metrics.RequestCount)
	}
	return metrics
}

// GetRateLimitStatus returns current rate limit status
func (c *EnhancedClient) GetRateLimitStatus() map[string]interface{} {
	return c.rateLimiter.GetStatus()
}

// recordSuccess records a successful request
func (c *EnhancedClient) recordSuccess(duration time.Duration) {
	c.metrics.RequestCount++
	c.metrics.TotalDuration += duration
}

// recordError records a failed request
func (c *EnhancedClient) recordError() {
	c.metrics.RequestCount++
	c.metrics.ErrorCount++
}

// recordRateLimit records a rate limit hit
func (c *EnhancedClient) recordRateLimit() {
	c.metrics.RateLimitHits++
}

// generateRequestID generates a unique request ID for tracing
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// extractQueryType extracts the operation type from a GraphQL query
func extractQueryType(query string) string {
	// Simple heuristic to determine query type
	if len(query) > 20 {
		prefix := query[:20]
		if contains(prefix, "mutation") {
			return "mutation"
		} else if contains(prefix, "subscription") {
			return "subscription"
		}
	}
	return "query"
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s[:len(substr)] == substr ||
			s[:len(substr)] == capitalizeFirst(substr))
}

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}
