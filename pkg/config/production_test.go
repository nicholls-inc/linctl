package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dorkitude/linctl/pkg/logging"
)

func TestLoadProductionConfig(t *testing.T) {
	// Clear environment variables first
	clearTestEnvVars()
	defer clearTestEnvVars()
	
	config, err := LoadProductionConfig()
	if err != nil {
		t.Fatalf("LoadProductionConfig failed: %v", err)
	}
	
	if config == nil {
		t.Fatal("LoadProductionConfig returned nil config")
	}
	
	// Check that defaults are loaded
	if config.Retry.MaxAttempts <= 0 {
		t.Error("Default retry max attempts should be positive")
	}
	
	if config.RateLimit.RequestsPerSecond <= 0 {
		t.Error("Default rate limit RPS should be positive")
	}
	
	if config.Logging.Level == "" {
		t.Error("Default logging level should not be empty")
	}
}

func TestLoadProductionConfigWithEnvironment(t *testing.T) {
	// Clear environment variables first
	clearTestEnvVars()
	defer clearTestEnvVars()
	
	// Set test environment variables
	os.Setenv("LINCTL_RETRY_MAX_ATTEMPTS", "5")
	os.Setenv("LINCTL_RETRY_INITIAL_DELAY", "2s")
	os.Setenv("LINCTL_RETRY_MAX_DELAY", "60s")
	os.Setenv("LINCTL_RETRY_MULTIPLIER", "3.0")
	os.Setenv("LINCTL_RETRY_JITTER", "false")
	
	os.Setenv("LINCTL_RATE_LIMIT_RPS", "20.5")
	os.Setenv("LINCTL_RATE_LIMIT_BURST", "50")
	os.Setenv("LINCTL_RATE_LIMIT_ENABLED", "false")
	os.Setenv("LINCTL_RATE_LIMIT_ADAPTIVE", "false")
	os.Setenv("LINCTL_RATE_LIMIT_BACKOFF", "10s")
	
	os.Setenv("LINCTL_LOG_LEVEL", "debug")
	os.Setenv("LINCTL_LOG_FORMAT", "json")
	
	os.Setenv("LINCTL_ENCRYPT_TOKENS", "true")
	os.Setenv("LINCTL_AUDIT_LOG", "false")
	os.Setenv("LINCTL_VALIDATE_INPUT", "false")
	
	os.Setenv("LINCTL_METRICS_ENABLED", "true")
	os.Setenv("LINCTL_METRICS_EXPORT_PATH", "/custom/path/metrics.json")
	
	config, err := LoadProductionConfig()
	if err != nil {
		t.Fatalf("LoadProductionConfig failed: %v", err)
	}
	
	// Check retry config
	if config.Retry.MaxAttempts != 5 {
		t.Errorf("Expected retry max attempts 5, got %d", config.Retry.MaxAttempts)
	}
	
	if config.Retry.InitialDelay != 2*time.Second {
		t.Errorf("Expected retry initial delay 2s, got %v", config.Retry.InitialDelay)
	}
	
	if config.Retry.MaxDelay != 60*time.Second {
		t.Errorf("Expected retry max delay 60s, got %v", config.Retry.MaxDelay)
	}
	
	if config.Retry.Multiplier != 3.0 {
		t.Errorf("Expected retry multiplier 3.0, got %f", config.Retry.Multiplier)
	}
	
	if config.Retry.Jitter != false {
		t.Errorf("Expected retry jitter false, got %v", config.Retry.Jitter)
	}
	
	// Check rate limit config
	if config.RateLimit.RequestsPerSecond != 20.5 {
		t.Errorf("Expected rate limit RPS 20.5, got %f", config.RateLimit.RequestsPerSecond)
	}
	
	if config.RateLimit.Burst != 50 {
		t.Errorf("Expected rate limit burst 50, got %d", config.RateLimit.Burst)
	}
	
	if config.RateLimit.Enabled != false {
		t.Errorf("Expected rate limit enabled false, got %v", config.RateLimit.Enabled)
	}
	
	if config.RateLimit.AdaptiveMode != false {
		t.Errorf("Expected rate limit adaptive false, got %v", config.RateLimit.AdaptiveMode)
	}
	
	if config.RateLimit.BackoffDelay != 10*time.Second {
		t.Errorf("Expected rate limit backoff 10s, got %v", config.RateLimit.BackoffDelay)
	}
	
	// Check logging config
	if config.Logging.Level != "debug" {
		t.Errorf("Expected logging level debug, got %s", config.Logging.Level)
	}
	
	if config.Logging.Format != "json" {
		t.Errorf("Expected logging format json, got %s", config.Logging.Format)
	}
	
	// Check security config
	if config.Security.EncryptTokens != true {
		t.Errorf("Expected encrypt tokens true, got %v", config.Security.EncryptTokens)
	}
	
	if config.Security.AuditLog != false {
		t.Errorf("Expected audit log false, got %v", config.Security.AuditLog)
	}
	
	if config.Security.ValidateInput != false {
		t.Errorf("Expected validate input false, got %v", config.Security.ValidateInput)
	}
	
	// Check metrics config
	if config.Metrics.Enabled != true {
		t.Errorf("Expected metrics enabled true, got %v", config.Metrics.Enabled)
	}
	
	if config.Metrics.ExportPath != "/custom/path/metrics.json" {
		t.Errorf("Expected metrics export path /custom/path/metrics.json, got %s", config.Metrics.ExportPath)
	}
}

func TestProductionConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *ProductionConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &ProductionConfig{
				Retry: loadRetryConfig(),
				RateLimit: loadRateLimitConfig(),
				Logging: LoggingConfig{Level: "info", Format: "text"},
				Security: SecurityConfig{},
				Metrics: MetricsConfig{},
			},
			expectError: false,
		},
		{
			name: "invalid retry max attempts",
			config: &ProductionConfig{
				Retry: loadRetryConfig(),
				RateLimit: loadRateLimitConfig(),
				Logging: LoggingConfig{Level: "info", Format: "text"},
				Security: SecurityConfig{},
				Metrics: MetricsConfig{},
			},
			expectError: true,
			errorMsg:    "retry max_attempts must be positive",
		},
		{
			name: "invalid logging level",
			config: &ProductionConfig{
				Retry: loadRetryConfig(),
				RateLimit: loadRateLimitConfig(),
				Logging: LoggingConfig{Level: "invalid", Format: "text"},
				Security: SecurityConfig{},
				Metrics: MetricsConfig{},
			},
			expectError: true,
			errorMsg:    "logging level must be one of",
		},
		{
			name: "invalid logging format",
			config: &ProductionConfig{
				Retry: loadRetryConfig(),
				RateLimit: loadRateLimitConfig(),
				Logging: LoggingConfig{Level: "info", Format: "invalid"},
				Security: SecurityConfig{},
				Metrics: MetricsConfig{},
			},
			expectError: true,
			errorMsg:    "logging format must be one of",
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Modify config for specific test cases
			if test.name == "invalid retry max attempts" {
				test.config.Retry.MaxAttempts = 0
			}
			
			err := test.config.Validate()
			
			if test.expectError {
				if err == nil {
					t.Errorf("Expected validation error for %s", test.name)
				} else if !strings.Contains(err.Error(), test.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got: %v", test.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error for %s: %v", test.name, err)
				}
			}
		})
	}
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		level    string
		expected logging.LogLevel
	}{
		{"debug", logging.DebugLevel},
		{"info", logging.InfoLevel},
		{"warn", logging.WarnLevel},
		{"warning", logging.WarnLevel},
		{"error", logging.ErrorLevel},
		{"invalid", logging.InfoLevel}, // Default fallback
		{"", logging.InfoLevel},        // Default fallback
	}
	
	for _, test := range tests {
		t.Run(test.level, func(t *testing.T) {
			config := &ProductionConfig{
				Logging: LoggingConfig{Level: test.level},
			}
			
			result := config.GetLogLevel()
			if result != test.expected {
				t.Errorf("Expected log level %v for input '%s', got %v", test.expected, test.level, result)
			}
		})
	}
}

func TestPrintConfig(t *testing.T) {
	config, err := LoadProductionConfig()
	if err != nil {
		t.Fatalf("LoadProductionConfig failed: %v", err)
	}
	
	logger := logging.NewNoOpLogger()
	
	// This should not panic
	config.PrintConfig(logger)
}

func TestGetEnvironmentVariablesHelp(t *testing.T) {
	help := GetEnvironmentVariablesHelp()
	
	if help == "" {
		t.Error("Environment variables help should not be empty")
	}
	
	// Check that it contains key sections
	expectedSections := []string{
		"Retry Configuration",
		"Rate Limiting Configuration",
		"Logging Configuration",
		"Security Configuration",
		"Metrics Configuration",
		"OAuth Configuration",
	}
	
	for _, section := range expectedSections {
		if !strings.Contains(help, section) {
			t.Errorf("Help text should contain section: %s", section)
		}
	}
	
	// Check that it contains key environment variables
	expectedVars := []string{
		"LINCTL_RETRY_MAX_ATTEMPTS",
		"LINCTL_RATE_LIMIT_RPS",
		"LINCTL_LOG_LEVEL",
		"LINCTL_ENCRYPT_TOKENS",
		"LINCTL_METRICS_ENABLED",
		"LINEAR_CLIENT_ID",
	}
	
	for _, envVar := range expectedVars {
		if !strings.Contains(help, envVar) {
			t.Errorf("Help text should contain environment variable: %s", envVar)
		}
	}
}

func TestGetEnvString(t *testing.T) {
	clearTestEnvVars()
	defer clearTestEnvVars()
	
	// Test with no environment variable
	result := getEnvString("TEST_VAR", "default")
	if result != "default" {
		t.Errorf("Expected default value 'default', got '%s'", result)
	}
	
	// Test with environment variable set
	os.Setenv("TEST_VAR", "custom")
	result = getEnvString("TEST_VAR", "default")
	if result != "custom" {
		t.Errorf("Expected custom value 'custom', got '%s'", result)
	}
}

func TestGetEnvInt(t *testing.T) {
	clearTestEnvVars()
	defer clearTestEnvVars()
	
	// Test with no environment variable
	result := getEnvInt("TEST_INT", 42)
	if result != 42 {
		t.Errorf("Expected default value 42, got %d", result)
	}
	
	// Test with valid environment variable
	os.Setenv("TEST_INT", "123")
	result = getEnvInt("TEST_INT", 42)
	if result != 123 {
		t.Errorf("Expected custom value 123, got %d", result)
	}
	
	// Test with invalid environment variable
	os.Setenv("TEST_INT", "invalid")
	result = getEnvInt("TEST_INT", 42)
	if result != 42 {
		t.Errorf("Expected default value 42 for invalid input, got %d", result)
	}
}

func TestGetEnvFloat(t *testing.T) {
	clearTestEnvVars()
	defer clearTestEnvVars()
	
	// Test with no environment variable
	result := getEnvFloat("TEST_FLOAT", 3.14)
	if result != 3.14 {
		t.Errorf("Expected default value 3.14, got %f", result)
	}
	
	// Test with valid environment variable
	os.Setenv("TEST_FLOAT", "2.71")
	result = getEnvFloat("TEST_FLOAT", 3.14)
	if result != 2.71 {
		t.Errorf("Expected custom value 2.71, got %f", result)
	}
	
	// Test with invalid environment variable
	os.Setenv("TEST_FLOAT", "invalid")
	result = getEnvFloat("TEST_FLOAT", 3.14)
	if result != 3.14 {
		t.Errorf("Expected default value 3.14 for invalid input, got %f", result)
	}
}

func TestGetEnvBool(t *testing.T) {
	clearTestEnvVars()
	defer clearTestEnvVars()
	
	tests := []struct {
		name         string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{"no env var", "", true, true},
		{"no env var false default", "", false, false},
		{"true", "true", false, true},
		{"1", "1", false, true},
		{"yes", "yes", false, true},
		{"on", "on", false, true},
		{"false", "false", true, false},
		{"0", "0", true, false},
		{"no", "no", true, false},
		{"off", "off", true, false},
		{"invalid", "invalid", true, true}, // Should use default
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			os.Unsetenv("TEST_BOOL")
			if test.envValue != "" {
				os.Setenv("TEST_BOOL", test.envValue)
			}
			
			result := getEnvBool("TEST_BOOL", test.defaultValue)
			if result != test.expected {
				t.Errorf("Expected %v for env='%s' default=%v, got %v", 
					test.expected, test.envValue, test.defaultValue, result)
			}
		})
	}
}

func TestGetEnvDuration(t *testing.T) {
	clearTestEnvVars()
	defer clearTestEnvVars()
	
	// Test with no environment variable
	result := getEnvDuration("TEST_DURATION", 5*time.Second)
	if result != 5*time.Second {
		t.Errorf("Expected default value 5s, got %v", result)
	}
	
	// Test with valid environment variable
	os.Setenv("TEST_DURATION", "10s")
	result = getEnvDuration("TEST_DURATION", 5*time.Second)
	if result != 10*time.Second {
		t.Errorf("Expected custom value 10s, got %v", result)
	}
	
	// Test with invalid environment variable
	os.Setenv("TEST_DURATION", "invalid")
	result = getEnvDuration("TEST_DURATION", 5*time.Second)
	if result != 5*time.Second {
		t.Errorf("Expected default value 5s for invalid input, got %v", result)
	}
}

func TestLoadRetryConfig(t *testing.T) {
	clearTestEnvVars()
	defer clearTestEnvVars()
	
	// Test with defaults
	config := loadRetryConfig()
	if config.MaxAttempts <= 0 {
		t.Error("Default max attempts should be positive")
	}
	if config.InitialDelay <= 0 {
		t.Error("Default initial delay should be positive")
	}
	if config.MaxDelay <= config.InitialDelay {
		t.Error("Default max delay should be greater than initial delay")
	}
	if config.Multiplier <= 1.0 {
		t.Error("Default multiplier should be greater than 1.0")
	}
	
	// Test with environment variables
	os.Setenv("LINCTL_RETRY_MAX_ATTEMPTS", "7")
	os.Setenv("LINCTL_RETRY_INITIAL_DELAY", "3s")
	os.Setenv("LINCTL_RETRY_MAX_DELAY", "90s")
	os.Setenv("LINCTL_RETRY_MULTIPLIER", "2.5")
	os.Setenv("LINCTL_RETRY_JITTER", "false")
	
	config = loadRetryConfig()
	if config.MaxAttempts != 7 {
		t.Errorf("Expected max attempts 7, got %d", config.MaxAttempts)
	}
	if config.InitialDelay != 3*time.Second {
		t.Errorf("Expected initial delay 3s, got %v", config.InitialDelay)
	}
	if config.MaxDelay != 90*time.Second {
		t.Errorf("Expected max delay 90s, got %v", config.MaxDelay)
	}
	if config.Multiplier != 2.5 {
		t.Errorf("Expected multiplier 2.5, got %f", config.Multiplier)
	}
	if config.Jitter != false {
		t.Errorf("Expected jitter false, got %v", config.Jitter)
	}
}

func TestLoadRateLimitConfig(t *testing.T) {
	clearTestEnvVars()
	defer clearTestEnvVars()
	
	// Test with defaults
	config := loadRateLimitConfig()
	if config.RequestsPerSecond <= 0 {
		t.Error("Default RPS should be positive")
	}
	if config.Burst <= 0 {
		t.Error("Default burst should be positive")
	}
	if config.BackoffDelay <= 0 {
		t.Error("Default backoff delay should be positive")
	}
	
	// Test with environment variables
	os.Setenv("LINCTL_RATE_LIMIT_RPS", "25.0")
	os.Setenv("LINCTL_RATE_LIMIT_BURST", "75")
	os.Setenv("LINCTL_RATE_LIMIT_ENABLED", "false")
	os.Setenv("LINCTL_RATE_LIMIT_ADAPTIVE", "false")
	os.Setenv("LINCTL_RATE_LIMIT_BACKOFF", "15s")
	
	config = loadRateLimitConfig()
	if config.RequestsPerSecond != 25.0 {
		t.Errorf("Expected RPS 25.0, got %f", config.RequestsPerSecond)
	}
	if config.Burst != 75 {
		t.Errorf("Expected burst 75, got %d", config.Burst)
	}
	if config.Enabled != false {
		t.Errorf("Expected enabled false, got %v", config.Enabled)
	}
	if config.AdaptiveMode != false {
		t.Errorf("Expected adaptive false, got %v", config.AdaptiveMode)
	}
	if config.BackoffDelay != 15*time.Second {
		t.Errorf("Expected backoff 15s, got %v", config.BackoffDelay)
	}
}

func TestLoadLoggingConfig(t *testing.T) {
	clearTestEnvVars()
	defer clearTestEnvVars()
	
	// Test with defaults
	config := loadLoggingConfig()
	if config.Level == "" {
		t.Error("Default log level should not be empty")
	}
	if config.Format == "" {
		t.Error("Default log format should not be empty")
	}
	
	// Test with environment variables
	os.Setenv("LINCTL_LOG_LEVEL", "error")
	os.Setenv("LINCTL_LOG_FORMAT", "json")
	
	config = loadLoggingConfig()
	if config.Level != "error" {
		t.Errorf("Expected log level error, got %s", config.Level)
	}
	if config.Format != "json" {
		t.Errorf("Expected log format json, got %s", config.Format)
	}
}

func TestLoadSecurityConfig(t *testing.T) {
	clearTestEnvVars()
	defer clearTestEnvVars()
	
	// Test with defaults
	config := loadSecurityConfig()
	// Just verify it doesn't panic and returns a config
	
	// Test with environment variables
	os.Setenv("LINCTL_ENCRYPT_TOKENS", "true")
	os.Setenv("LINCTL_AUDIT_LOG", "false")
	os.Setenv("LINCTL_VALIDATE_INPUT", "false")
	
	config = loadSecurityConfig()
	if config.EncryptTokens != true {
		t.Errorf("Expected encrypt tokens true, got %v", config.EncryptTokens)
	}
	if config.AuditLog != false {
		t.Errorf("Expected audit log false, got %v", config.AuditLog)
	}
	if config.ValidateInput != false {
		t.Errorf("Expected validate input false, got %v", config.ValidateInput)
	}
}

func TestLoadMetricsConfig(t *testing.T) {
	clearTestEnvVars()
	defer clearTestEnvVars()
	
	// Test with defaults
	config := loadMetricsConfig()
	if config.ExportPath == "" {
		t.Error("Default export path should not be empty")
	}
	
	// Test with environment variables
	os.Setenv("LINCTL_METRICS_ENABLED", "true")
	os.Setenv("LINCTL_METRICS_EXPORT_PATH", "/test/metrics.json")
	
	config = loadMetricsConfig()
	if config.Enabled != true {
		t.Errorf("Expected metrics enabled true, got %v", config.Enabled)
	}
	if config.ExportPath != "/test/metrics.json" {
		t.Errorf("Expected export path /test/metrics.json, got %s", config.ExportPath)
	}
}

func TestContains(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}
	
	if !contains(slice, "banana") {
		t.Error("Should find 'banana' in slice")
	}
	
	if contains(slice, "grape") {
		t.Error("Should not find 'grape' in slice")
	}
	
	if contains([]string{}, "anything") {
		t.Error("Should not find anything in empty slice")
	}
}

// Helper function to clear test environment variables
func clearTestEnvVars() {
	envVars := []string{
		"LINCTL_RETRY_MAX_ATTEMPTS",
		"LINCTL_RETRY_INITIAL_DELAY",
		"LINCTL_RETRY_MAX_DELAY",
		"LINCTL_RETRY_MULTIPLIER",
		"LINCTL_RETRY_JITTER",
		"LINCTL_RATE_LIMIT_RPS",
		"LINCTL_RATE_LIMIT_BURST",
		"LINCTL_RATE_LIMIT_ENABLED",
		"LINCTL_RATE_LIMIT_ADAPTIVE",
		"LINCTL_RATE_LIMIT_BACKOFF",
		"LINCTL_LOG_LEVEL",
		"LINCTL_LOG_FORMAT",
		"LINCTL_ENCRYPT_TOKENS",
		"LINCTL_AUDIT_LOG",
		"LINCTL_VALIDATE_INPUT",
		"LINCTL_METRICS_ENABLED",
		"LINCTL_METRICS_EXPORT_PATH",
		"TEST_VAR",
		"TEST_INT",
		"TEST_FLOAT",
		"TEST_BOOL",
		"TEST_DURATION",
	}
	
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}