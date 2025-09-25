package resilience

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"syscall"
	"time"

	"github.com/dorkitude/linctl/pkg/logging"
)

// RetryConfig defines the configuration for retry behavior
type RetryConfig struct {
	MaxAttempts  int           `json:"max_attempts"`
	InitialDelay time.Duration `json:"initial_delay"`
	MaxDelay     time.Duration `json:"max_delay"`
	Multiplier   float64       `json:"multiplier"`
	Jitter       bool          `json:"jitter"`
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

// RetryableClient wraps an HTTP client with retry logic
type RetryableClient struct {
	client *http.Client
	config RetryConfig
	logger logging.Logger
}

// NewRetryableClient creates a new retryable HTTP client
func NewRetryableClient(client *http.Client, config RetryConfig, logger logging.Logger) *RetryableClient {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	if logger == nil {
		logger = logging.NewNoOpLogger()
	}

	return &RetryableClient{
		client: client,
		config: config,
		logger: logger,
	}
}

// DoWithRetry executes an HTTP request with retry logic
func (r *RetryableClient) DoWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		// Clone the request for each attempt
		reqClone := req.Clone(ctx)

		r.logger.Debug("Attempting HTTP request",
			logging.String("method", req.Method),
			logging.String("url", req.URL.String()),
			logging.Int("attempt", attempt),
			logging.Int("max_attempts", r.config.MaxAttempts),
		)

		start := time.Now()
		resp, err := r.client.Do(reqClone)
		duration := time.Since(start)

		if err != nil {
			lastErr = err

			r.logger.Warn("HTTP request failed",
				logging.String("method", req.Method),
				logging.String("url", req.URL.String()),
				logging.Int("attempt", attempt),
				logging.Duration("duration", duration),
				logging.Error(err),
			)

			// Check if we should retry based on the error type
			if !r.shouldRetryError(err) {
				r.logger.Debug("Error is not retryable, giving up",
					logging.Error(err),
				)
				return nil, err
			}

			// Don't sleep after the last attempt
			if attempt < r.config.MaxAttempts {
				delay := r.calculateDelay(attempt)
				r.logger.Debug("Retrying after delay",
					logging.Duration("delay", delay),
					logging.Int("next_attempt", attempt+1),
				)

				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(delay):
					// Continue to next attempt
				}
			}
			continue
		}

		r.logger.Debug("HTTP request completed",
			logging.String("method", req.Method),
			logging.String("url", req.URL.String()),
			logging.Int("attempt", attempt),
			logging.Int("status_code", resp.StatusCode),
			logging.Duration("duration", duration),
		)

		// Check if we should retry based on the status code
		if r.shouldRetryStatus(resp.StatusCode) {
			r.logger.Warn("HTTP request returned retryable status",
				logging.String("method", req.Method),
				logging.String("url", req.URL.String()),
				logging.Int("attempt", attempt),
				logging.Int("status_code", resp.StatusCode),
			)

			// If this is the last attempt, return the response as-is
			if attempt >= r.config.MaxAttempts {
				return resp, nil
			}

			resp.Body.Close() // Close the body before retrying

			delay := r.calculateDelay(attempt)
			r.logger.Debug("Retrying after delay",
				logging.Duration("delay", delay),
				logging.Int("next_attempt", attempt+1),
			)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				// Continue to next attempt
			}
			continue
		}

		// Success or non-retryable error
		return resp, nil
	}

	// All attempts exhausted
	r.logger.Error("All retry attempts exhausted",
		logging.String("method", req.Method),
		logging.String("url", req.URL.String()),
		logging.Int("attempts", r.config.MaxAttempts),
		logging.Error(lastErr),
	)

	return nil, fmt.Errorf("request failed after %d attempts: %w", r.config.MaxAttempts, lastErr)
}

// shouldRetryError determines if an error is retryable
func (r *RetryableClient) shouldRetryError(err error) bool {
	// Context cancellation is not retryable
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}

	// URL errors (like parse errors) are not retryable
	if _, ok := err.(*url.Error); ok {
		// But network errors within URL errors might be retryable
		if urlErr, ok := err.(*url.Error); ok {
			return r.isNetworkError(urlErr.Err)
		}
	}

	// Network errors are retryable
	return r.isNetworkError(err)
}

// isNetworkError checks if an error is a network-related error
func (r *RetryableClient) isNetworkError(err error) bool {
	// DNS errors
	if _, ok := err.(*net.DNSError); ok {
		return true
	}

	// Connection errors
	if _, ok := err.(*net.OpError); ok {
		return true
	}

	// Timeout errors
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}

	// Connection refused, connection reset, etc.
	if err == syscall.ECONNREFUSED || err == syscall.ECONNRESET || err == syscall.ETIMEDOUT {
		return true
	}

	return false
}

// shouldRetryStatus determines if an HTTP status code is retryable
func (r *RetryableClient) shouldRetryStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests: // 429
		return true
	case http.StatusBadGateway: // 502
		return true
	case http.StatusServiceUnavailable: // 503
		return true
	case http.StatusGatewayTimeout: // 504
		return true
	default:
		return false
	}
}

// calculateDelay calculates the delay for the next retry attempt
func (r *RetryableClient) calculateDelay(attempt int) time.Duration {
	// Calculate exponential backoff
	delay := float64(r.config.InitialDelay) * math.Pow(r.config.Multiplier, float64(attempt-1))

	// Apply maximum delay
	if delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}

	// Apply jitter if enabled
	if r.config.Jitter {
		// Add random jitter of ±25%
		jitter := delay * 0.25 * (rand.Float64()*2 - 1)
		delay += jitter

		// Ensure delay is not negative
		if delay < 0 {
			delay = float64(r.config.InitialDelay)
		}
	}

	return time.Duration(delay)
}

// GetClient returns the underlying HTTP client
func (r *RetryableClient) GetClient() *http.Client {
	return r.client
}

// GetConfig returns the retry configuration
func (r *RetryableClient) GetConfig() RetryConfig {
	return r.config
}
