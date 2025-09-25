package resilience

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"syscall"
	"testing"
	"time"

	"github.com/dorkitude/linctl/pkg/logging"
)

func TestRetryableClient_Success(t *testing.T) {
	// Create a test server that always succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	config := DefaultRetryConfig()
	client := NewRetryableClient(nil, config, logging.NewNoOpLogger())

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.DoWithRetry(context.Background(), req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestRetryableClient_RetryOnNetworkError(t *testing.T) {
	// Skip this test as it's difficult to simulate network errors reliably in tests
	t.Skip("Network error simulation is unreliable in test environment")
}

func TestRetryableClient_RetryOnRetryableStatus(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	client := NewRetryableClient(nil, config, logging.NewNoOpLogger())

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.DoWithRetry(context.Background(), req)
	if err != nil {
		t.Fatalf("Request failed after retries: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetryableClient_NoRetryOnNonRetryableStatus(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest) // 400 is not retryable
	}))
	defer server.Close()

	config := DefaultRetryConfig()
	client := NewRetryableClient(nil, config, logging.NewNoOpLogger())

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.DoWithRetry(context.Background(), req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt for non-retryable status, got %d", attempts)
	}
}

func TestRetryableClient_ExhaustRetries(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	config := RetryConfig{
		MaxAttempts:  2,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	client := NewRetryableClient(nil, config, logging.NewNoOpLogger())

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.DoWithRetry(context.Background(), req)
	if err != nil {
		t.Fatalf("Request should succeed but return error status: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected final status 503, got %d", resp.StatusCode)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetryableClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := DefaultRetryConfig()
	client := NewRetryableClient(nil, config, logging.NewNoOpLogger())

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = client.DoWithRetry(ctx, req)
	if err == nil {
		t.Fatal("Expected context timeout error")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}

func TestShouldRetryError(t *testing.T) {
	client := NewRetryableClient(nil, DefaultRetryConfig(), logging.NewNoOpLogger())

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "DNS error",
			err:      &net.DNSError{},
			expected: true,
		},
		{
			name:     "Connection refused",
			err:      syscall.ECONNREFUSED,
			expected: true,
		},
		{
			name:     "Context canceled",
			err:      context.Canceled,
			expected: false,
		},
		{
			name:     "Context deadline exceeded",
			err:      context.DeadlineExceeded,
			expected: false,
		},
		{
			name:     "Generic error",
			err:      fmt.Errorf("generic error"),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := client.shouldRetryError(test.err)
			if result != test.expected {
				t.Errorf("shouldRetryError(%v) = %v, expected %v", test.err, result, test.expected)
			}
		})
	}
}

func TestShouldRetryStatus(t *testing.T) {
	client := NewRetryableClient(nil, DefaultRetryConfig(), logging.NewNoOpLogger())

	tests := []struct {
		status   int
		expected bool
	}{
		{http.StatusOK, false},
		{http.StatusBadRequest, false},
		{http.StatusUnauthorized, false},
		{http.StatusForbidden, false},
		{http.StatusNotFound, false},
		{http.StatusTooManyRequests, true},
		{http.StatusInternalServerError, false},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
		{http.StatusGatewayTimeout, true},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("status_%d", test.status), func(t *testing.T) {
			result := client.shouldRetryStatus(test.status)
			if result != test.expected {
				t.Errorf("shouldRetryStatus(%d) = %v, expected %v", test.status, result, test.expected)
			}
		})
	}
}

func TestCalculateDelay(t *testing.T) {
	config := RetryConfig{
		InitialDelay: 1 * time.Second,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
	}

	client := NewRetryableClient(nil, config, logging.NewNoOpLogger())

	// Test exponential backoff
	delay1 := client.calculateDelay(1)
	delay2 := client.calculateDelay(2)
	delay3 := client.calculateDelay(3)

	if delay1 != 1*time.Second {
		t.Errorf("First delay should be 1s, got %v", delay1)
	}

	if delay2 != 2*time.Second {
		t.Errorf("Second delay should be 2s, got %v", delay2)
	}

	if delay3 != 4*time.Second {
		t.Errorf("Third delay should be 4s, got %v", delay3)
	}

	// Test max delay cap
	delay10 := client.calculateDelay(10)
	if delay10 != 10*time.Second {
		t.Errorf("Delay should be capped at max delay (10s), got %v", delay10)
	}
}

func TestCalculateDelayWithJitter(t *testing.T) {
	config := RetryConfig{
		InitialDelay: 1 * time.Second,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}

	client := NewRetryableClient(nil, config, logging.NewNoOpLogger())

	// Test that jitter produces different values
	delays := make([]time.Duration, 10)
	for i := 0; i < 10; i++ {
		delays[i] = client.calculateDelay(1)
	}

	// Check that we got some variation (not all delays are identical)
	allSame := true
	for i := 1; i < len(delays); i++ {
		if delays[i] != delays[0] {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("Jitter should produce different delay values")
	}

	// Check that all delays are reasonable (within expected range)
	baseDelay := 1 * time.Second
	for i, delay := range delays {
		if delay < 0 {
			t.Errorf("Delay %d should not be negative: %v", i, delay)
		}

		// With 25% jitter, delay should be roughly between 0.75s and 1.25s
		if delay < 500*time.Millisecond || delay > 2*time.Second {
			t.Errorf("Delay %d seems out of reasonable range: %v (base: %v)", i, delay, baseDelay)
		}
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxAttempts <= 0 {
		t.Error("MaxAttempts should be positive")
	}

	if config.InitialDelay <= 0 {
		t.Error("InitialDelay should be positive")
	}

	if config.MaxDelay <= config.InitialDelay {
		t.Error("MaxDelay should be greater than InitialDelay")
	}

	if config.Multiplier <= 1.0 {
		t.Error("Multiplier should be greater than 1.0")
	}
}
