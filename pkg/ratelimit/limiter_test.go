package ratelimit

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/dorkitude/linctl/pkg/logging"
	"strings"
)

// Helper function to check if error is a context deadline error
func isContextDeadlineError(err error) bool {
	return strings.Contains(err.Error(), "context deadline") || err == context.DeadlineExceeded
}

func TestNewRateLimiter(t *testing.T) {
	config := DefaultRateLimitConfig()
	logger := logging.NewNoOpLogger()

	limiter := NewRateLimiter(config, logger)

	if limiter == nil {
		t.Fatal("NewRateLimiter returned nil")
	}

	if limiter.config.RequestsPerSecond != config.RequestsPerSecond {
		t.Errorf("Expected RPS %f, got %f", config.RequestsPerSecond, limiter.config.RequestsPerSecond)
	}

	if limiter.config.Burst != config.Burst {
		t.Errorf("Expected burst %d, got %d", config.Burst, limiter.config.Burst)
	}
}

func TestNewRateLimiterWithNilLogger(t *testing.T) {
	config := DefaultRateLimitConfig()

	limiter := NewRateLimiter(config, nil)

	if limiter == nil {
		t.Fatal("NewRateLimiter with nil logger returned nil")
	}

	// Should not panic when using logger
	err := limiter.Wait(context.Background())
	if err != nil {
		t.Errorf("Wait failed: %v", err)
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	config := RateLimitConfig{
		RequestsPerSecond: 100.0, // High rate for fast testing
		Burst:             10,
		Enabled:           true,
		AdaptiveMode:      false,
		BackoffDelay:      1 * time.Second,
	}

	limiter := NewRateLimiter(config, logging.NewNoOpLogger())

	// First request should be immediate
	start := time.Now()
	err := limiter.Wait(context.Background())
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Wait failed: %v", err)
	}

	if duration > 10*time.Millisecond {
		t.Errorf("First request took too long: %v", duration)
	}
}

func TestRateLimiter_WaitDisabled(t *testing.T) {
	config := RateLimitConfig{
		RequestsPerSecond: 1.0, // Very low rate
		Burst:             1,
		Enabled:           false, // Disabled
		AdaptiveMode:      false,
		BackoffDelay:      1 * time.Second,
	}

	limiter := NewRateLimiter(config, logging.NewNoOpLogger())

	// Multiple requests should all be immediate when disabled
	for i := 0; i < 5; i++ {
		start := time.Now()
		err := limiter.Wait(context.Background())
		duration := time.Since(start)

		if err != nil {
			t.Errorf("Wait %d failed: %v", i, err)
		}

		if duration > 10*time.Millisecond {
			t.Errorf("Request %d took too long when disabled: %v", i, duration)
		}
	}
}

func TestRateLimiter_WaitContextCancellation(t *testing.T) {
	config := RateLimitConfig{
		RequestsPerSecond: 0.1, // Very slow rate to force waiting
		Burst:             1,
		Enabled:           true,
		AdaptiveMode:      false,
		BackoffDelay:      1 * time.Second,
	}

	limiter := NewRateLimiter(config, logging.NewNoOpLogger())

	// Use up the burst capacity
	err := limiter.Wait(context.Background())
	if err != nil {
		t.Fatalf("Initial wait failed: %v", err)
	}

	// Create a context that will be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// This should fail due to context cancellation
	err = limiter.Wait(ctx)
	if err == nil {
		t.Error("Expected context cancellation error")
	}

	// Check that it's a context deadline error (the exact error type may vary)
	if err == nil || !isContextDeadlineError(err) {
		t.Errorf("Expected context deadline error, got %v", err)
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	config := RateLimitConfig{
		RequestsPerSecond: 100.0,
		Burst:             2,
		Enabled:           true,
		AdaptiveMode:      false,
		BackoffDelay:      1 * time.Second,
	}

	limiter := NewRateLimiter(config, logging.NewNoOpLogger())

	// First two requests should be allowed (burst capacity)
	if !limiter.Allow() {
		t.Error("First request should be allowed")
	}

	if !limiter.Allow() {
		t.Error("Second request should be allowed")
	}

	// Third request might be denied depending on timing
	// We'll just check that Allow() doesn't panic
	_ = limiter.Allow()
}

func TestRateLimiter_AllowDisabled(t *testing.T) {
	config := RateLimitConfig{
		RequestsPerSecond: 0.1, // Very low rate
		Burst:             1,
		Enabled:           false, // Disabled
		AdaptiveMode:      false,
		BackoffDelay:      1 * time.Second,
	}

	limiter := NewRateLimiter(config, logging.NewNoOpLogger())

	// All requests should be allowed when disabled
	for i := 0; i < 10; i++ {
		if !limiter.Allow() {
			t.Errorf("Request %d should be allowed when rate limiting is disabled", i)
		}
	}
}

func TestRateLimiter_ParseRateHeaders(t *testing.T) {
	limiter := NewRateLimiter(DefaultRateLimitConfig(), logging.NewNoOpLogger())

	tests := []struct {
		name     string
		headers  map[string]string
		expected *LinearRateInfo
	}{
		{
			name: "complete headers",
			headers: map[string]string{
				"X-RateLimit-Limit":     "5000",
				"X-RateLimit-Remaining": "4999",
				"X-RateLimit-Reset":     "1640995200",
				"X-RateLimit-Used":      "1",
			},
			expected: &LinearRateInfo{
				Limit:     5000,
				Remaining: 4999,
				Reset:     time.Unix(1640995200, 0),
				Used:      1,
			},
		},
		{
			name: "alternative header names",
			headers: map[string]string{
				"RateLimit-Limit":     "1000",
				"RateLimit-Remaining": "999",
			},
			expected: &LinearRateInfo{
				Limit:     1000,
				Remaining: 999,
				Reset:     time.Time{},
				Used:      0,
			},
		},
		{
			name: "missing headers",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			expected: nil,
		},
		{
			name: "invalid limit header",
			headers: map[string]string{
				"X-RateLimit-Limit":     "invalid",
				"X-RateLimit-Remaining": "100",
			},
			expected: nil,
		},
		{
			name: "invalid remaining header",
			headers: map[string]string{
				"X-RateLimit-Limit":     "1000",
				"X-RateLimit-Remaining": "invalid",
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a mock response with headers
			resp := &http.Response{
				Header: make(http.Header),
			}

			for key, value := range test.headers {
				resp.Header.Set(key, value)
			}

			result := limiter.parseRateHeaders(resp)

			if test.expected == nil {
				if result != nil {
					t.Errorf("Expected nil result, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Fatalf("Expected result, got nil")
			}

			if result.Limit != test.expected.Limit {
				t.Errorf("Expected limit %d, got %d", test.expected.Limit, result.Limit)
			}

			if result.Remaining != test.expected.Remaining {
				t.Errorf("Expected remaining %d, got %d", test.expected.Remaining, result.Remaining)
			}

			if result.Used != test.expected.Used {
				t.Errorf("Expected used %d, got %d", test.expected.Used, result.Used)
			}

			if !test.expected.Reset.IsZero() && !result.Reset.Equal(test.expected.Reset) {
				t.Errorf("Expected reset %v, got %v", test.expected.Reset, result.Reset)
			}
		})
	}
}

func TestRateLimiter_UpdateFromResponse(t *testing.T) {
	config := RateLimitConfig{
		RequestsPerSecond: 10.0,
		Burst:             20,
		Enabled:           true,
		AdaptiveMode:      true, // Enable adaptive mode
		BackoffDelay:      1 * time.Second,
	}

	limiter := NewRateLimiter(config, logging.NewNoOpLogger())

	// Create a mock response with rate limit headers
	resp := &http.Response{
		Header: make(http.Header),
	}
	resp.Header.Set("X-RateLimit-Limit", "1000")
	resp.Header.Set("X-RateLimit-Remaining", "500")
	resp.Header.Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10))
	resp.Header.Set("X-RateLimit-Used", "500")

	// Update from response
	limiter.UpdateFromResponse(resp)

	// Check that rate info was stored
	if limiter.lastRateInfo == nil {
		t.Fatal("Rate info should be stored after UpdateFromResponse")
	}

	if limiter.lastRateInfo.Limit != 1000 {
		t.Errorf("Expected limit 1000, got %d", limiter.lastRateInfo.Limit)
	}

	if limiter.lastRateInfo.Remaining != 500 {
		t.Errorf("Expected remaining 500, got %d", limiter.lastRateInfo.Remaining)
	}
}

func TestRateLimiter_UpdateFromResponseNonAdaptive(t *testing.T) {
	config := RateLimitConfig{
		RequestsPerSecond: 10.0,
		Burst:             20,
		Enabled:           true,
		AdaptiveMode:      false, // Disable adaptive mode
		BackoffDelay:      1 * time.Second,
	}

	limiter := NewRateLimiter(config, logging.NewNoOpLogger())

	// Create a mock response with rate limit headers
	resp := &http.Response{
		Header: make(http.Header),
	}
	resp.Header.Set("X-RateLimit-Limit", "1000")
	resp.Header.Set("X-RateLimit-Remaining", "500")

	// Update from response (should be ignored in non-adaptive mode)
	limiter.UpdateFromResponse(resp)

	// Rate info should not be stored in non-adaptive mode
	if limiter.lastRateInfo != nil {
		t.Error("Rate info should not be stored in non-adaptive mode")
	}
}

func TestRateLimiter_GetStatus(t *testing.T) {
	config := RateLimitConfig{
		RequestsPerSecond: 15.5,
		Burst:             25,
		Enabled:           true,
		AdaptiveMode:      true,
		BackoffDelay:      2 * time.Second,
	}

	limiter := NewRateLimiter(config, logging.NewNoOpLogger())

	status := limiter.GetStatus()

	// Check basic configuration
	if status["enabled"] != true {
		t.Errorf("Expected enabled=true, got %v", status["enabled"])
	}

	if status["requests_per_second"] != 15.5 {
		t.Errorf("Expected RPS=15.5, got %v", status["requests_per_second"])
	}

	if status["burst"] != 25 {
		t.Errorf("Expected burst=25, got %v", status["burst"])
	}

	if status["adaptive_mode"] != true {
		t.Errorf("Expected adaptive_mode=true, got %v", status["adaptive_mode"])
	}

	// Linear rate info should not be present initially
	if _, exists := status["linear_limit"]; exists {
		t.Error("Linear rate info should not be present initially")
	}
}

func TestRateLimiter_GetStatusWithRateInfo(t *testing.T) {
	limiter := NewRateLimiter(DefaultRateLimitConfig(), logging.NewNoOpLogger())

	// Set some rate info
	limiter.lastRateInfo = &LinearRateInfo{
		Limit:     2000,
		Remaining: 1500,
		Used:      500,
		Reset:     time.Unix(1640995200, 0),
	}

	status := limiter.GetStatus()

	// Check Linear rate info
	if status["linear_limit"] != 2000 {
		t.Errorf("Expected linear_limit=2000, got %v", status["linear_limit"])
	}

	if status["linear_remaining"] != 1500 {
		t.Errorf("Expected linear_remaining=1500, got %v", status["linear_remaining"])
	}

	if status["linear_used"] != 500 {
		t.Errorf("Expected linear_used=500, got %v", status["linear_used"])
	}

	if status["linear_reset"] != "2022-01-01T00:00:00Z" {
		t.Errorf("Expected linear_reset=2022-01-01T00:00:00Z, got %v", status["linear_reset"])
	}
}

func TestRateLimiter_HandleRateLimitResponse(t *testing.T) {
	limiter := NewRateLimiter(DefaultRateLimitConfig(), logging.NewNoOpLogger())

	tests := []struct {
		name           string
		retryAfter     string
		expectedDelay  time.Duration
		shouldParseInt bool
	}{
		{
			name:           "with retry-after header",
			retryAfter:     "30",
			expectedDelay:  30 * time.Second,
			shouldParseInt: true,
		},
		{
			name:           "without retry-after header",
			retryAfter:     "",
			expectedDelay:  DefaultRateLimitConfig().BackoffDelay,
			shouldParseInt: false,
		},
		{
			name:           "invalid retry-after header",
			retryAfter:     "invalid",
			expectedDelay:  DefaultRateLimitConfig().BackoffDelay,
			shouldParseInt: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a mock 429 response
			resp := &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     make(http.Header),
			}

			if test.retryAfter != "" {
				resp.Header.Set("Retry-After", test.retryAfter)
			}

			delay := limiter.HandleRateLimitResponse(resp)

			if delay != test.expectedDelay {
				t.Errorf("Expected delay %v, got %v", test.expectedDelay, delay)
			}
		})
	}
}

func TestDefaultRateLimitConfig(t *testing.T) {
	config := DefaultRateLimitConfig()

	if config.RequestsPerSecond <= 0 {
		t.Error("RequestsPerSecond should be positive")
	}

	if config.Burst <= 0 {
		t.Error("Burst should be positive")
	}

	if !config.Enabled {
		t.Error("Rate limiting should be enabled by default")
	}

	if !config.AdaptiveMode {
		t.Error("Adaptive mode should be enabled by default")
	}

	if config.BackoffDelay <= 0 {
		t.Error("BackoffDelay should be positive")
	}
}

func TestLinearRateInfo(t *testing.T) {
	info := &LinearRateInfo{
		Limit:     1000,
		Remaining: 500,
		Reset:     time.Now().Add(time.Hour),
		Used:      500,
	}

	if info.Limit != 1000 {
		t.Errorf("Expected limit 1000, got %d", info.Limit)
	}

	if info.Remaining != 500 {
		t.Errorf("Expected remaining 500, got %d", info.Remaining)
	}

	if info.Used != 500 {
		t.Errorf("Expected used 500, got %d", info.Used)
	}

	if info.Reset.Before(time.Now()) {
		t.Error("Reset time should be in the future")
	}
}

func TestRateLimiter_AdaptiveRateAdjustment(t *testing.T) {
	config := RateLimitConfig{
		RequestsPerSecond: 10.0,
		Burst:             20,
		Enabled:           true,
		AdaptiveMode:      true,
		BackoffDelay:      1 * time.Second,
	}

	limiter := NewRateLimiter(config, logging.NewNoOpLogger())

	// Get initial rate
	_ = float64(limiter.limiter.Limit())

	// Create a response indicating we have plenty of quota left
	resp := &http.Response{
		Header: make(http.Header),
	}
	resp.Header.Set("X-RateLimit-Limit", "1000")
	resp.Header.Set("X-RateLimit-Remaining", "900") // Lots remaining
	resp.Header.Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10))

	// Update from response
	limiter.UpdateFromResponse(resp)

	// The rate might be adjusted, but we can't predict exactly how
	// Just verify the limiter is still functional
	err := limiter.Wait(context.Background())
	if err != nil {
		t.Errorf("Wait should still work after adaptive adjustment: %v", err)
	}

	// Verify rate info was stored
	if limiter.lastRateInfo == nil {
		t.Error("Rate info should be stored")
	}

	if limiter.lastRateInfo.Remaining != 900 {
		t.Errorf("Expected remaining 900, got %d", limiter.lastRateInfo.Remaining)
	}
}

// Benchmark tests
func BenchmarkRateLimiter_Allow(b *testing.B) {
	limiter := NewRateLimiter(DefaultRateLimitConfig(), logging.NewNoOpLogger())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow()
	}
}

func BenchmarkRateLimiter_Wait(b *testing.B) {
	config := RateLimitConfig{
		RequestsPerSecond: 1000.0, // High rate for benchmarking
		Burst:             100,
		Enabled:           true,
		AdaptiveMode:      false,
		BackoffDelay:      1 * time.Second,
	}

	limiter := NewRateLimiter(config, logging.NewNoOpLogger())
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Wait(ctx)
	}
}
