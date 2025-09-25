package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/time/rate"

	"github.com/nicholls-inc/linctl/pkg/logging"
)

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond float64       `json:"requests_per_second"`
	Burst             int           `json:"burst"`
	Enabled           bool          `json:"enabled"`
	AdaptiveMode      bool          `json:"adaptive_mode"`
	BackoffDelay      time.Duration `json:"backoff_delay"`
}

// DefaultRateLimitConfig returns a sensible default rate limit configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerSecond: 10.0, // Conservative default
		Burst:             20,
		Enabled:           true,
		AdaptiveMode:      true,
		BackoffDelay:      5 * time.Second,
	}
}

// RateLimiter manages request rate limiting
type RateLimiter struct {
	limiter      *rate.Limiter
	config       RateLimitConfig
	logger       logging.Logger
	lastRateInfo *LinearRateInfo
}

// LinearRateInfo represents rate limit information from Linear's API
type LinearRateInfo struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset"`
	Used      int       `json:"used"`
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimitConfig, logger logging.Logger) *RateLimiter {
	if logger == nil {
		logger = logging.NewNoOpLogger()
	}

	limiter := rate.NewLimiter(rate.Limit(config.RequestsPerSecond), config.Burst)

	return &RateLimiter{
		limiter: limiter,
		config:  config,
		logger:  logger,
	}
}

// Wait waits for permission to make a request
func (rl *RateLimiter) Wait(ctx context.Context) error {
	if !rl.config.Enabled {
		return nil
	}

	start := time.Now()
	err := rl.limiter.Wait(ctx)
	waitTime := time.Since(start)

	if err != nil {
		rl.logger.Error("Rate limiter wait failed",
			logging.Error(err),
			logging.Duration("wait_time", waitTime),
		)
		return err
	}

	if waitTime > 100*time.Millisecond {
		rl.logger.Debug("Rate limiter applied delay",
			logging.Duration("wait_time", waitTime),
		)
	}

	return nil
}

// Allow checks if a request is allowed without waiting
func (rl *RateLimiter) Allow() bool {
	if !rl.config.Enabled {
		return true
	}

	allowed := rl.limiter.Allow()

	if !allowed {
		rl.logger.Debug("Request denied by rate limiter")
	}

	return allowed
}

// UpdateFromResponse updates the rate limiter based on Linear's response headers
func (rl *RateLimiter) UpdateFromResponse(resp *http.Response) {
	if !rl.config.AdaptiveMode {
		return
	}

	rateInfo := rl.parseRateHeaders(resp)
	if rateInfo == nil {
		return
	}

	rl.lastRateInfo = rateInfo

	// Adaptive rate limiting based on remaining quota
	if rateInfo.Remaining > 0 {
		// Calculate time until reset
		timeUntilReset := time.Until(rateInfo.Reset)
		if timeUntilReset > 0 {
			// Calculate safe rate to avoid hitting the limit
			safeRate := float64(rateInfo.Remaining) / timeUntilReset.Seconds()

			// Apply a safety margin (use 80% of calculated rate)
			safeRate *= 0.8

			// Don't go below a minimum rate
			minRate := 1.0
			if safeRate < minRate {
				safeRate = minRate
			}

			// Don't exceed configured maximum
			if safeRate > rl.config.RequestsPerSecond {
				safeRate = rl.config.RequestsPerSecond
			}

			// Update the limiter if the rate changed significantly
			currentRate := float64(rl.limiter.Limit())
			if abs(safeRate-currentRate)/currentRate > 0.1 { // 10% change threshold
				rl.limiter.SetLimit(rate.Limit(safeRate))

				rl.logger.Debug("Adaptive rate limit updated",
					logging.Int("remaining", rateInfo.Remaining),
					logging.Int("limit", rateInfo.Limit),
					logging.Duration("time_until_reset", timeUntilReset),
					logging.String("old_rate", fmt.Sprintf("%.2f", currentRate)),
					logging.String("new_rate", fmt.Sprintf("%.2f", safeRate)),
				)
			}
		}
	}

	// Log rate limit status
	rl.logger.Debug("Rate limit status",
		logging.Int("limit", rateInfo.Limit),
		logging.Int("remaining", rateInfo.Remaining),
		logging.Int("used", rateInfo.Used),
		logging.String("reset", rateInfo.Reset.Format(time.RFC3339)),
	)
}

// parseRateHeaders extracts rate limit information from HTTP response headers
func (rl *RateLimiter) parseRateHeaders(resp *http.Response) *LinearRateInfo {
	// Linear uses X-RateLimit-* headers (common pattern)
	limitStr := resp.Header.Get("X-RateLimit-Limit")
	remainingStr := resp.Header.Get("X-RateLimit-Remaining")
	resetStr := resp.Header.Get("X-RateLimit-Reset")
	usedStr := resp.Header.Get("X-RateLimit-Used")

	if limitStr == "" || remainingStr == "" {
		// Try alternative header names
		limitStr = resp.Header.Get("RateLimit-Limit")
		remainingStr = resp.Header.Get("RateLimit-Remaining")
		resetStr = resp.Header.Get("RateLimit-Reset")
	}

	if limitStr == "" || remainingStr == "" {
		return nil
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		rl.logger.Warn("Failed to parse rate limit header",
			logging.String("header", "X-RateLimit-Limit"),
			logging.String("value", limitStr),
			logging.Error(err),
		)
		return nil
	}

	remaining, err := strconv.Atoi(remainingStr)
	if err != nil {
		rl.logger.Warn("Failed to parse rate limit header",
			logging.String("header", "X-RateLimit-Remaining"),
			logging.String("value", remainingStr),
			logging.Error(err),
		)
		return nil
	}

	var reset time.Time
	if resetStr != "" {
		if resetUnix, err := strconv.ParseInt(resetStr, 10, 64); err == nil {
			reset = time.Unix(resetUnix, 0)
		} else {
			// Try parsing as RFC3339
			if resetTime, err := time.Parse(time.RFC3339, resetStr); err == nil {
				reset = resetTime
			}
		}
	}

	var used int
	if usedStr != "" {
		if parsedUsed, err := strconv.Atoi(usedStr); err == nil {
			used = parsedUsed
		}
	}

	return &LinearRateInfo{
		Limit:     limit,
		Remaining: remaining,
		Reset:     reset,
		Used:      used,
	}
}

// GetStatus returns the current rate limit status
func (rl *RateLimiter) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"enabled":             rl.config.Enabled,
		"requests_per_second": float64(rl.limiter.Limit()),
		"burst":               rl.limiter.Burst(),
		"adaptive_mode":       rl.config.AdaptiveMode,
	}

	if rl.lastRateInfo != nil {
		status["linear_limit"] = rl.lastRateInfo.Limit
		status["linear_remaining"] = rl.lastRateInfo.Remaining
		status["linear_used"] = rl.lastRateInfo.Used
		if !rl.lastRateInfo.Reset.IsZero() {
			status["linear_reset"] = rl.lastRateInfo.Reset.Format(time.RFC3339)
		}
	}

	return status
}

// HandleRateLimitResponse handles a 429 Too Many Requests response
func (rl *RateLimiter) HandleRateLimitResponse(resp *http.Response) time.Duration {
	// Update rate info from headers
	rl.UpdateFromResponse(resp)

	// Check for Retry-After header
	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			delay := time.Duration(seconds) * time.Second
			rl.logger.Warn("Rate limited by server",
				logging.Duration("retry_after", delay),
			)
			return delay
		}
	}

	// Use configured backoff delay
	rl.logger.Warn("Rate limited by server, using default backoff",
		logging.Duration("backoff_delay", rl.config.BackoffDelay),
	)
	return rl.config.BackoffDelay
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
