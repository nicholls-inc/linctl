package security

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Input validation patterns
var (
	// Linear issue ID pattern: TEAM-123
	issueIDPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]{1,10}-\d{1,6}$`)

	// Team key pattern: 2-10 uppercase letters/numbers
	teamKeyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]{1,9}$`)

	// URL pattern for avatar URLs
	urlPattern = regexp.MustCompile(`^https?://[^\s<>"{}|\\^` + "`" + `\[\]]+$`)
)

// ValidationError represents an input validation error
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// SanitizeInput sanitizes general text input by removing potentially dangerous characters
func SanitizeInput(input string) string {
	if input == "" {
		return input
	}

	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except common whitespace
	var result strings.Builder
	for _, r := range input {
		if r == '\n' || r == '\r' || r == '\t' || r == ' ' {
			result.WriteRune(r)
		} else if !unicode.IsControl(r) {
			result.WriteRune(r)
		}
	}

	// Trim excessive whitespace
	sanitized := strings.TrimSpace(result.String())

	// Collapse multiple consecutive spaces
	sanitized = regexp.MustCompile(`\s+`).ReplaceAllString(sanitized, " ")

	return sanitized
}

// ValidateIssueID validates a Linear issue ID format
func ValidateIssueID(id string) error {
	if id == "" {
		return ValidationError{
			Field:   "issue_id",
			Value:   id,
			Message: "issue ID cannot be empty",
		}
	}

	// Sanitize first
	sanitized := SanitizeInput(id)
	if sanitized != id {
		return ValidationError{
			Field:   "issue_id",
			Value:   id,
			Message: "issue ID contains invalid characters",
		}
	}

	// Check format
	if !issueIDPattern.MatchString(sanitized) {
		return ValidationError{
			Field:   "issue_id",
			Value:   id,
			Message: "issue ID must be in format TEAM-123 (e.g., ENG-456)",
		}
	}

	// Check length constraints
	if len(sanitized) > 20 {
		return ValidationError{
			Field:   "issue_id",
			Value:   id,
			Message: "issue ID is too long (maximum 20 characters)",
		}
	}

	return nil
}

// ValidateTeamKey validates a Linear team key format
func ValidateTeamKey(key string) error {
	if key == "" {
		return ValidationError{
			Field:   "team_key",
			Value:   key,
			Message: "team key cannot be empty",
		}
	}

	// Sanitize first
	sanitized := SanitizeInput(key)
	if sanitized != key {
		return ValidationError{
			Field:   "team_key",
			Value:   key,
			Message: "team key contains invalid characters",
		}
	}

	// Check format
	if !teamKeyPattern.MatchString(sanitized) {
		return ValidationError{
			Field:   "team_key",
			Value:   key,
			Message: "team key must be 2-10 uppercase letters/numbers starting with a letter (e.g., ENG, DESIGN)",
		}
	}

	return nil
}

// ValidateTitle validates issue/comment titles
func ValidateTitle(title string) error {
	if title == "" {
		return ValidationError{
			Field:   "title",
			Value:   title,
			Message: "title cannot be empty",
		}
	}

	// Sanitize
	sanitized := SanitizeInput(title)

	// Check length
	if len(sanitized) > 255 {
		return ValidationError{
			Field:   "title",
			Value:   title,
			Message: "title is too long (maximum 255 characters)",
		}
	}

	if len(sanitized) < 3 {
		return ValidationError{
			Field:   "title",
			Value:   title,
			Message: "title is too short (minimum 3 characters)",
		}
	}

	// Check for reasonable content
	if strings.TrimSpace(sanitized) == "" {
		return ValidationError{
			Field:   "title",
			Value:   title,
			Message: "title cannot be only whitespace",
		}
	}

	return nil
}

// ValidateDescription validates issue/comment descriptions
func ValidateDescription(description string) error {
	if description == "" {
		return nil // Description is optional
	}

	// Sanitize
	sanitized := SanitizeInput(description)

	// Check length (Linear has a limit)
	if len(sanitized) > 50000 {
		return ValidationError{
			Field:   "description",
			Value:   description,
			Message: "description is too long (maximum 50,000 characters)",
		}
	}

	return nil
}

// ValidateActorName validates actor names for attribution
func ValidateActorName(name string) error {
	if name == "" {
		return nil // Actor name is optional
	}

	// Sanitize
	sanitized := SanitizeInput(name)

	// Check length
	if len(sanitized) > 100 {
		return ValidationError{
			Field:   "actor_name",
			Value:   name,
			Message: "actor name is too long (maximum 100 characters)",
		}
	}

	if len(sanitized) < 1 {
		return ValidationError{
			Field:   "actor_name",
			Value:   name,
			Message: "actor name cannot be empty if provided",
		}
	}

	// Check for reasonable content
	if strings.TrimSpace(sanitized) == "" {
		return ValidationError{
			Field:   "actor_name",
			Value:   name,
			Message: "actor name cannot be only whitespace",
		}
	}

	return nil
}

// ValidateAvatarURL validates avatar URLs
func ValidateAvatarURL(url string) error {
	if url == "" {
		return nil // Avatar URL is optional
	}

	// Basic URL format check
	if !urlPattern.MatchString(url) {
		return ValidationError{
			Field:   "avatar_url",
			Value:   url,
			Message: "avatar URL must be a valid HTTP/HTTPS URL",
		}
	}

	// Check length
	if len(url) > 2048 {
		return ValidationError{
			Field:   "avatar_url",
			Value:   url,
			Message: "avatar URL is too long (maximum 2048 characters)",
		}
	}

	// Ensure HTTPS for security
	if !strings.HasPrefix(url, "https://") {
		return ValidationError{
			Field:   "avatar_url",
			Value:   url,
			Message: "avatar URL must use HTTPS for security",
		}
	}

	return nil
}

// ValidatePriority validates issue priority values
func ValidatePriority(priority int) error {
	if priority < 0 || priority > 4 {
		return ValidationError{
			Field:   "priority",
			Value:   fmt.Sprintf("%d", priority),
			Message: "priority must be between 0 (None) and 4 (Low)",
		}
	}
	return nil
}

// SanitizeAndValidateAll performs comprehensive validation on common input fields
func SanitizeAndValidateAll(fields map[string]interface{}) (map[string]interface{}, []ValidationError) {
	var errors []ValidationError
	sanitized := make(map[string]interface{})

	for key, value := range fields {
		strValue, ok := value.(string)
		if !ok {
			sanitized[key] = value
			continue
		}

		switch key {
		case "issue_id":
			if err := ValidateIssueID(strValue); err != nil {
				if valErr, ok := err.(ValidationError); ok {
					errors = append(errors, valErr)
				}
			}
			sanitized[key] = SanitizeInput(strValue)

		case "team_key", "team":
			if err := ValidateTeamKey(strValue); err != nil {
				if valErr, ok := err.(ValidationError); ok {
					errors = append(errors, valErr)
				}
			}
			sanitized[key] = SanitizeInput(strValue)

		case "title":
			if err := ValidateTitle(strValue); err != nil {
				if valErr, ok := err.(ValidationError); ok {
					errors = append(errors, valErr)
				}
			}
			sanitized[key] = SanitizeInput(strValue)

		case "description", "body":
			if err := ValidateDescription(strValue); err != nil {
				if valErr, ok := err.(ValidationError); ok {
					errors = append(errors, valErr)
				}
			}
			sanitized[key] = SanitizeInput(strValue)

		case "actor", "actor_name":
			if err := ValidateActorName(strValue); err != nil {
				if valErr, ok := err.(ValidationError); ok {
					errors = append(errors, valErr)
				}
			}
			sanitized[key] = SanitizeInput(strValue)

		case "avatar_url":
			if err := ValidateAvatarURL(strValue); err != nil {
				if valErr, ok := err.(ValidationError); ok {
					errors = append(errors, valErr)
				}
			}
			sanitized[key] = strValue // Don't sanitize URLs

		default:
			// Generic sanitization for other string fields
			sanitized[key] = SanitizeInput(strValue)
		}
	}

	// Handle non-string fields
	if priority, ok := fields["priority"].(int); ok {
		if err := ValidatePriority(priority); err != nil {
			if valErr, ok := err.(ValidationError); ok {
				errors = append(errors, valErr)
			}
		}
		sanitized["priority"] = priority
	}

	return sanitized, errors
}

// IsValidInput performs a quick check if input contains only safe characters
func IsValidInput(input string) bool {
	if input == "" {
		return true
	}

	// Check for null bytes
	if strings.Contains(input, "\x00") {
		return false
	}

	// Check for excessive control characters
	controlCount := 0
	for _, r := range input {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			controlCount++
			if controlCount > 5 { // Allow some control chars but not too many
				return false
			}
		}
	}

	return true
}
