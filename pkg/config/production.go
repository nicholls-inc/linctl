package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dorkitude/linctl/pkg/logging"
	"github.com/dorkitude/linctl/pkg/ratelimit"
	"github.com/dorkitude/linctl/pkg/resilience"
)

// ProductionConfig holds all production-ready configuration
type ProductionConfig struct {
	Retry     resilience.RetryConfig    `json:"retry"`
	RateLimit ratelimit.RateLimitConfig `json:"rate_limit"`
	Logging   LoggingConfig             `json:"logging"`
	Security  SecurityConfig            `json:"security"`
	Metrics   MetricsConfig             `json:"metrics"`
}

// LoggingConfig configures logging behavior
type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

// SecurityConfig configures security features
type SecurityConfig struct {
	EncryptTokens bool `json:"encrypt_tokens"`
	AuditLog      bool `json:"audit_log"`
	ValidateInput bool `json:"validate_input"`
}

// MetricsConfig configures metrics collection
type MetricsConfig struct {
	Enabled    bool   `json:"enabled"`
	ExportPath string `json:"export_path"`
}

// LoadProductionConfig loads configuration from environment variables
func LoadProductionConfig() (*ProductionConfig, error) {
	config := &ProductionConfig{
		Retry:     loadRetryConfig(),
		RateLimit: loadRateLimitConfig(),
		Logging:   loadLoggingConfig(),
		Security:  loadSecurityConfig(),
		Metrics:   loadMetricsConfig(),
	}

	return config, nil
}

// loadRetryConfig loads retry configuration from environment
func loadRetryConfig() resilience.RetryConfig {
	config := resilience.DefaultRetryConfig()

	if maxAttempts := getEnvInt("LINCTL_RETRY_MAX_ATTEMPTS", config.MaxAttempts); maxAttempts > 0 {
		config.MaxAttempts = maxAttempts
	}

	if initialDelay := getEnvDuration("LINCTL_RETRY_INITIAL_DELAY", config.InitialDelay); initialDelay > 0 {
		config.InitialDelay = initialDelay
	}

	if maxDelay := getEnvDuration("LINCTL_RETRY_MAX_DELAY", config.MaxDelay); maxDelay > 0 {
		config.MaxDelay = maxDelay
	}

	if multiplier := getEnvFloat("LINCTL_RETRY_MULTIPLIER", config.Multiplier); multiplier > 1.0 {
		config.Multiplier = multiplier
	}

	if jitter := getEnvBool("LINCTL_RETRY_JITTER", config.Jitter); jitter != config.Jitter {
		config.Jitter = jitter
	}

	return config
}

// loadRateLimitConfig loads rate limiting configuration from environment
func loadRateLimitConfig() ratelimit.RateLimitConfig {
	config := ratelimit.DefaultRateLimitConfig()

	if rps := getEnvFloat("LINCTL_RATE_LIMIT_RPS", config.RequestsPerSecond); rps > 0 {
		config.RequestsPerSecond = rps
	}

	if burst := getEnvInt("LINCTL_RATE_LIMIT_BURST", config.Burst); burst > 0 {
		config.Burst = burst
	}

	if enabled := getEnvBool("LINCTL_RATE_LIMIT_ENABLED", config.Enabled); enabled != config.Enabled {
		config.Enabled = enabled
	}

	if adaptive := getEnvBool("LINCTL_RATE_LIMIT_ADAPTIVE", config.AdaptiveMode); adaptive != config.AdaptiveMode {
		config.AdaptiveMode = adaptive
	}

	if backoff := getEnvDuration("LINCTL_RATE_LIMIT_BACKOFF", config.BackoffDelay); backoff > 0 {
		config.BackoffDelay = backoff
	}

	return config
}

// loadLoggingConfig loads logging configuration from environment
func loadLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:  getEnvString("LINCTL_LOG_LEVEL", "info"),
		Format: getEnvString("LINCTL_LOG_FORMAT", "text"),
	}
}

// loadSecurityConfig loads security configuration from environment
func loadSecurityConfig() SecurityConfig {
	return SecurityConfig{
		EncryptTokens: getEnvBool("LINCTL_ENCRYPT_TOKENS", false),
		AuditLog:      getEnvBool("LINCTL_AUDIT_LOG", true),
		ValidateInput: getEnvBool("LINCTL_VALIDATE_INPUT", true),
	}
}

// loadMetricsConfig loads metrics configuration from environment
func loadMetricsConfig() MetricsConfig {
	return MetricsConfig{
		Enabled:    getEnvBool("LINCTL_METRICS_ENABLED", false),
		ExportPath: getEnvString("LINCTL_METRICS_EXPORT_PATH", "/tmp/linctl-metrics.json"),
	}
}

// Helper functions for environment variable parsing

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch strings.ToLower(value) {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// Validate validates the production configuration
func (c *ProductionConfig) Validate() error {
	// Validate retry config
	if c.Retry.MaxAttempts <= 0 {
		return fmt.Errorf("retry max_attempts must be positive")
	}
	if c.Retry.InitialDelay <= 0 {
		return fmt.Errorf("retry initial_delay must be positive")
	}
	if c.Retry.MaxDelay <= c.Retry.InitialDelay {
		return fmt.Errorf("retry max_delay must be greater than initial_delay")
	}
	if c.Retry.Multiplier <= 1.0 {
		return fmt.Errorf("retry multiplier must be greater than 1.0")
	}

	// Validate rate limit config
	if c.RateLimit.RequestsPerSecond <= 0 {
		return fmt.Errorf("rate_limit requests_per_second must be positive")
	}
	if c.RateLimit.Burst <= 0 {
		return fmt.Errorf("rate_limit burst must be positive")
	}
	if c.RateLimit.BackoffDelay <= 0 {
		return fmt.Errorf("rate_limit backoff_delay must be positive")
	}

	// Validate logging config
	validLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLevels, strings.ToLower(c.Logging.Level)) {
		return fmt.Errorf("logging level must be one of: %v", validLevels)
	}

	validFormats := []string{"text", "json"}
	if !contains(validFormats, strings.ToLower(c.Logging.Format)) {
		return fmt.Errorf("logging format must be one of: %v", validFormats)
	}

	return nil
}

// GetLogLevel returns the logging level as a logging.LogLevel
func (c *ProductionConfig) GetLogLevel() logging.LogLevel {
	switch strings.ToLower(c.Logging.Level) {
	case "debug":
		return logging.DebugLevel
	case "info":
		return logging.InfoLevel
	case "warn", "warning":
		return logging.WarnLevel
	case "error":
		return logging.ErrorLevel
	default:
		return logging.InfoLevel
	}
}

// PrintConfig prints the current configuration (for debugging)
func (c *ProductionConfig) PrintConfig(logger logging.Logger) {
	logger.Info("Production configuration loaded",
		// Retry config
		logging.Int("retry_max_attempts", c.Retry.MaxAttempts),
		logging.Duration("retry_initial_delay", c.Retry.InitialDelay),
		logging.Duration("retry_max_delay", c.Retry.MaxDelay),
		logging.String("retry_multiplier", fmt.Sprintf("%.1f", c.Retry.Multiplier)),
		logging.Bool("retry_jitter", c.Retry.Jitter),

		// Rate limit config
		logging.String("rate_limit_rps", fmt.Sprintf("%.1f", c.RateLimit.RequestsPerSecond)),
		logging.Int("rate_limit_burst", c.RateLimit.Burst),
		logging.Bool("rate_limit_enabled", c.RateLimit.Enabled),
		logging.Bool("rate_limit_adaptive", c.RateLimit.AdaptiveMode),
		logging.Duration("rate_limit_backoff", c.RateLimit.BackoffDelay),

		// Logging config
		logging.String("log_level", c.Logging.Level),
		logging.String("log_format", c.Logging.Format),

		// Security config
		logging.Bool("encrypt_tokens", c.Security.EncryptTokens),
		logging.Bool("audit_log", c.Security.AuditLog),
		logging.Bool("validate_input", c.Security.ValidateInput),

		// Metrics config
		logging.Bool("metrics_enabled", c.Metrics.Enabled),
		logging.String("metrics_export_path", c.Metrics.ExportPath),
	)
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetEnvironmentVariablesHelp returns help text for environment variables
func GetEnvironmentVariablesHelp() string {
	return `
Production Configuration Environment Variables:

Retry Configuration:
  LINCTL_RETRY_MAX_ATTEMPTS=3        # Maximum retry attempts
  LINCTL_RETRY_INITIAL_DELAY=1s      # Initial delay between retries
  LINCTL_RETRY_MAX_DELAY=30s         # Maximum delay between retries
  LINCTL_RETRY_MULTIPLIER=2.0        # Delay multiplier for exponential backoff
  LINCTL_RETRY_JITTER=true           # Add random jitter to delays

Rate Limiting Configuration:
  LINCTL_RATE_LIMIT_RPS=10.0         # Requests per second limit
  LINCTL_RATE_LIMIT_BURST=20         # Burst capacity
  LINCTL_RATE_LIMIT_ENABLED=true     # Enable rate limiting
  LINCTL_RATE_LIMIT_ADAPTIVE=true    # Enable adaptive rate limiting
  LINCTL_RATE_LIMIT_BACKOFF=5s       # Backoff delay for rate limit hits

Logging Configuration:
  LINCTL_LOG_LEVEL=info              # Log level (debug, info, warn, error)
  LINCTL_LOG_FORMAT=text             # Log format (text, json)

Security Configuration:
  LINCTL_ENCRYPT_TOKENS=false        # Encrypt tokens at rest
  LINCTL_AUDIT_LOG=true              # Enable audit logging
  LINCTL_VALIDATE_INPUT=true         # Enable input validation

Metrics Configuration:
  LINCTL_METRICS_ENABLED=false       # Enable metrics collection
  LINCTL_METRICS_EXPORT_PATH=/tmp/linctl-metrics.json  # Metrics export path

OAuth Configuration (from previous phases):
  LINEAR_CLIENT_ID=your-client-id    # OAuth client ID
  LINEAR_CLIENT_SECRET=your-secret   # OAuth client secret
  LINEAR_BASE_URL=https://api.linear.app  # Linear API base URL
  LINEAR_SCOPES=read,write           # OAuth scopes
  LINEAR_DEFAULT_ACTOR=Agent Name    # Default actor for attribution
  LINEAR_DEFAULT_AVATAR_URL=https://example.com/avatar.png  # Default avatar URL
`
}
